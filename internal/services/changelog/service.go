package changelog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"clikd/internal/utils"
)

// Service provides functions for changelog management
type Service struct {
	ConfigPath string
	factory    ServiceFactoryInterface // Optional factory for AI service injection
}

// ServiceFactoryInterface defines the interface for service creation
// This allows for dependency injection while avoiding circular imports
type ServiceFactoryInterface interface {
	CreateAIServiceForChangelog() (AIServiceInterface, error)
	GetConfigForChangelog() ConfigInterface
}

// AIServiceInterface defines the interface for AI services
type AIServiceInterface interface {
	EnhanceChangelog(changelog string) (string, error)
}

// ConfigInterface defines the interface for configuration
type ConfigInterface interface {
	GetAIConfig() (provider, model, apiKey, apiURL string, tokensMaxInput, tokensMaxOutput int)
}

// GenerationOptions contains all options for changelog generation
type GenerationOptions struct {
	// Core options
	ConfigPath    string
	Template      string
	RepositoryURL string
	OutputPath    string
	Query         string
	NextTag       string

	// Filtering options
	TagFilterPattern string
	Paths            []string

	// Display options
	Silent          bool
	NoColor         bool
	NoEmoji         bool
	NoCaseSensitive bool

	// Integration options
	JiraURL      string
	JiraUsername string
	JiraToken    string

	// Sorting options
	Sort      string
	Processor string
}

// GenerationResult contains the result of changelog generation
type GenerationResult struct {
	Content       string
	CommandConfig *CommandConfig
	ShouldUseUI   bool
}

// PrepareGeneration prepares all the configuration and determines the generation strategy
func (s *Service) PrepareGeneration(ctx context.Context, options *GenerationOptions) (*GenerationResult, error) {
	// Create logger with appropriate level
	logLevel := "error"
	if options.OutputPath != "" {
		logLevel = "info" // Show progress for file output
	}
	logger := utils.NewLogger(logLevel, !options.NoColor)

	// Determine working directory
	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory: %v", err)
		return nil, err
	}

	// Resolve and validate config path
	configPath := utils.ResolveConfigPath(options.ConfigPath)
	configFileInfo, err := os.Stat(configPath)
	if err != nil {
		logger.Error("Configuration file not found at %s: %v", configPath, err)
		return nil, fmt.Errorf("configuration file not found at %s: %v", configPath, err)
	}
	if configFileInfo.IsDir() {
		configPath = filepath.Join(configPath, "config.yml")
	}

	// Create command configuration
	cmdConfig := &CommandConfig{
		WorkingDir:       wd,
		ConfigPath:       configPath,
		Template:         options.Template,
		RepositoryURL:    options.RepositoryURL,
		OutputPath:       options.OutputPath,
		Silent:           options.Silent,
		NoColor:          options.NoColor,
		NoEmoji:          options.NoEmoji,
		NoCaseSensitive:  options.NoCaseSensitive,
		Query:            options.Query,
		NextTag:          options.NextTag,
		TagFilterPattern: options.TagFilterPattern,
		JiraUsername:     options.JiraUsername,
		JiraToken:        options.JiraToken,
		JiraURL:          options.JiraURL,
		Paths:            options.Paths,
		Sort:             options.Sort,
		Processor:        options.Processor,
	}

	// Determine if UI should be used
	shouldUseUI := options.OutputPath == "" && !options.NoColor

	result := &GenerationResult{
		CommandConfig: cmdConfig,
		ShouldUseUI:   shouldUseUI,
	}

	// If not using UI, generate content directly
	if !shouldUseUI {
		content, err := s.generateDirect(logger, cmdConfig, options.Query)
		if err != nil {
			return nil, err
		}
		result.Content = content
	}

	return result, nil
}

// GenerateChangelog is the high-level method that handles all changelog generation logic
func (s *Service) GenerateChangelog(ctx context.Context, options *GenerationOptions) error {
	result, err := s.PrepareGeneration(ctx, options)
	if err != nil {
		return err
	}

	// If content was generated directly, handle output
	if !result.ShouldUseUI {
		if options.OutputPath == "" {
			// Terminal output
			fmt.Print(result.Content)
		} else {
			// File output - content was already written to file in generateDirect
		}
	}

	return nil
}

// aiServiceToGeneratorAdapter adapts changelog.AIServiceInterface to ai.Service
type aiServiceToGeneratorAdapter struct {
	aiService AIServiceInterface
}

func (a *aiServiceToGeneratorAdapter) EnhanceChangelog(changelog string) (string, error) {
	return a.aiService.EnhanceChangelog(changelog)
}

// generateDirect handles file output or no-color terminal output
func (s *Service) generateDirect(logger utils.Logger, cmdConfig *CommandConfig, query string) (string, error) {
	// Load configuration
	config, err := LoadConfigFromCommand(cmdConfig)
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return "", err
	}

	// Create generator
	generator := NewGenerator(logger, config)

	// Try to inject AI service if factory is available
	if s.factory != nil {
		aiService, err := s.factory.CreateAIServiceForChangelog()
		if err != nil {
			logger.Debug("Could not create AI service, generator will work without AI enhancement: %v", err)
		} else {
			logger.Debug("AI service created successfully, injecting into changelog generator")
			// Create adapter to convert changelog.AIServiceInterface to ai.Service
			adapter := &aiServiceToGeneratorAdapter{aiService: aiService}
			generator.SetAIService(adapter)
		}
	} else {
		logger.Debug("No service factory available, generator will work without AI enhancement")
	}

	// Determine output writer
	var writer io.Writer
	var isStdout bool

	if cmdConfig.OutputPath == "" {
		// Terminal output - use buffer
		isStdout = true
		writer = &bytes.Buffer{}
	} else {
		// File output - write directly to file
		isStdout = false

		// Create directory if it doesn't exist
		dir := filepath.Dir(cmdConfig.OutputPath)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %v", err)
			}
		}

		// Open file
		file, err := os.Create(cmdConfig.OutputPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()
		writer = file
	}

	// Generate changelog (with AI enhancement if available)
	err = generator.Generate(writer, query)
	if err != nil {
		return "", err
	}

	// Return content if stdout was used
	if isStdout {
		buffer := writer.(*bytes.Buffer)
		return buffer.String(), nil
	}

	return "", nil
}

// NewService creates a new Changelog service
func NewService(configPath string) *Service {
	return &Service{
		ConfigPath: configPath,
		factory:    nil, // No factory injection - graceful degradation
	}
}

// NewServiceWithFactory creates a new Changelog service with ServiceFactory injection
func NewServiceWithFactory(configPath string, factory ServiceFactoryInterface) *Service {
	return &Service{
		ConfigPath: configPath,
		factory:    factory,
	}
}

// InitializeTemplates creates template and configuration files
func (s *Service) InitializeTemplates(style string, configDir string) error {
	// Create directories
	templateDir := filepath.Join(configDir, "templates")
	configDir = filepath.Join(configDir, "config")

	dirs := []string{templateDir, configDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}

	// Write template and configuration files
	templatePath := filepath.Join(templateDir, style+".tpl.md")
	configPath := filepath.Join(configDir, style+".yml")

	// Create an Answer object (not used)
	// but can be useful in future implementations

	// Generate template and configuration
	templateContent := getDefaultTemplate(style)
	configContent := getDefaultConfig(style)

	// Write template
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("error writing template file: %w", err)
	}

	// Write configuration
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}

	return nil
}

// EnsureTemplateExists ensures that the template file exists
// If not, it will be restored from the embedded template
func (s *Service) EnsureTemplateExists(templatePath, style string) error {
	if templatePath == "" {
		return nil // No template file configured
	}

	// Check if file exists
	if _, err := os.Stat(templatePath); err == nil {
		return nil // File exists
	}

	// Ensure directory exists
	dir := filepath.Dir(templatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating template directory: %w", err)
	}

	// Write template
	templateContent := getDefaultTemplate(style)
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("error restoring template: %w", err)
	}

	fmt.Printf("Template file has been restored: %s\n", templatePath)
	return nil
}

// EnsureConfigExists ensures that the configuration file exists
// If not, it will be restored from the embedded configuration
func (s *Service) EnsureConfigExists(configPath, style string) error {
	if configPath == "" {
		return nil // No configuration file configured
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File exists
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating configuration directory: %w", err)
	}

	// Write configuration
	configContent := getDefaultConfig(style)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("error restoring configuration: %w", err)
	}

	fmt.Printf("Configuration file has been restored: %s\n", configPath)
	return nil
}

// getDefaultTemplate returns the default template for the specified style
func getDefaultTemplate(style string) string {
	switch style {
	case "github":
		return `# {{.Info.Title}}
{{range .Versions}}
<a name="{{.Tag.Name}}"></a>
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	case "gitlab":
		return `# {{.Info.Title}}
{{range .Versions}}
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	case "bitbucket":
		return `# {{.Info.Title}}
{{range .Versions}}
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	default:
		return `# {{.Info.Title}}
{{range .Versions}}
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	}
}

// getDefaultConfig returns the default configuration for the specified style
func getDefaultConfig(style string) string {
	switch style {
	case "github":
		return `style: github
template: templates/github.tpl.md
info:
  title: CHANGELOG
  repository_url: https://github.com/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	case "gitlab":
		return `style: gitlab
template: templates/gitlab.tpl.md
info:
  title: CHANGELOG
  repository_url: https://gitlab.com/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	case "bitbucket":
		return `style: bitbucket
template: templates/bitbucket.tpl.md
info:
  title: CHANGELOG
  repository_url: https://bitbucket.org/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	default:
		return `style: github
template: templates/github.tpl.md
info:
  title: CHANGELOG
  repository_url: https://github.com/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	}
}
