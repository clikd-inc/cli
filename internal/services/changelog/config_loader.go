package changelog

import (
	"clikd/internal/utils"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigLoader ...
type ConfigLoader interface {
	Load(string) (*Config, error)
}

type configLoaderImpl struct {
}

// NewConfigLoader ...
func NewConfigLoader() ConfigLoader {
	return &configLoaderImpl{}
}

func (loader *configLoaderImpl) Load(configPath string) (*Config, error) {
	// Direkt die YAML-Changelog-Konfiguration laden
	config, err := loader.loadYAMLConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML config from %s: %w", configPath, err)
	}

	// Working Directory setzen (relativ zur Konfigurationsdatei)
	config.WorkingDir = filepath.Dir(configPath)

	return config, nil
}

func (loader *configLoaderImpl) loadYAMLConfig(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Load into ChangelogConfig first (YAML structure)
	yamlConfig := &ChangelogConfig{}
	err = yaml.Unmarshal(bytes, yamlConfig)
	if err != nil {
		return nil, err
	}

	// Convert to internal Config structure
	config := &Config{
		Bin:        yamlConfig.Bin,
		Template:   yamlConfig.Template,
		WorkingDir: "", // Will be set by caller
		Info: &Info{
			Title:         yamlConfig.Info.Title,
			RepositoryURL: yamlConfig.Info.RepositoryURL,
		},
		Options: &Options{
			TagFilterPattern:            yamlConfig.Options.TagFilterPattern,
			Sort:                        yamlConfig.Options.Sort,
			CommitFilters:               yamlConfig.Options.Commits.Filters,
			CommitSortBy:                yamlConfig.Options.Commits.SortBy,
			CommitGroupBy:               yamlConfig.Options.CommitGroups.GroupBy,
			CommitGroupSortBy:           yamlConfig.Options.CommitGroups.SortBy,
			CommitGroupTitleOrder:       yamlConfig.Options.CommitGroups.TitleOrder,
			CommitGroupTitleMaps:        yamlConfig.Options.CommitGroups.TitleMaps,
			HeaderPattern:               yamlConfig.Options.Header.Pattern,
			HeaderPatternMaps:           yamlConfig.Options.Header.PatternMaps,
			IssuePrefix:                 yamlConfig.Options.Issues.Prefix,
			RefActions:                  yamlConfig.Options.Refs.Actions,
			MergePattern:                yamlConfig.Options.Merges.Pattern,
			MergePatternMaps:            yamlConfig.Options.Merges.PatternMaps,
			RevertPattern:               yamlConfig.Options.Reverts.Pattern,
			RevertPatternMaps:           yamlConfig.Options.Reverts.PatternMaps,
			NoteKeywords:                yamlConfig.Options.Notes.Keywords,
			JiraUsername:                yamlConfig.Options.Jira.ClintInfo.Username,
			JiraToken:                   yamlConfig.Options.Jira.ClintInfo.Token,
			JiraURL:                     yamlConfig.Options.Jira.ClintInfo.URL,
			JiraTypeMaps:                yamlConfig.Options.Jira.Issue.TypeMaps,
			JiraIssueDescriptionPattern: yamlConfig.Options.Jira.Issue.DescriptionPattern,
		},
	}

	// Set defaults if not specified
	if config.Bin == "" {
		config.Bin = "git"
	}

	return config, nil
}

// CommandConfig enthält die Konfiguration für den Changelog-Befehl
type CommandConfig struct {
	// Basis-Konfiguration
	WorkingDir    string
	ConfigPath    string
	Template      string
	RepositoryURL string
	OutputPath    string

	// Filter und Optionen
	Silent           bool
	NoColor          bool
	NoEmoji          bool
	NoCaseSensitive  bool
	Query            string
	NextTag          string
	TagFilterPattern string
	Sort             string
	Processor        string

	// Jira-Integration
	JiraUsername string
	JiraToken    string
	JiraURL      string

	// Commit-Filter
	Paths []string
}

// Normalize applies CLI context values to the config
func (config *Config) Normalize(ctx *CLIContext) error {
	// Override config with CLI context values
	if ctx.Template != "" {
		config.Template = ctx.Template
	}

	if ctx.WorkingDir != "" {
		config.WorkingDir = ctx.WorkingDir
	}

	if ctx.RepositoryURL != "" {
		config.Info.RepositoryURL = ctx.RepositoryURL
	}

	if ctx.NoCaseSensitive {
		config.Options.NoCaseSensitive = true
	}

	if ctx.NextTag != "" {
		config.Options.NextTag = ctx.NextTag
	}

	if ctx.TagFilterPattern != "" {
		config.Options.TagFilterPattern = ctx.TagFilterPattern
	}

	if ctx.Sort != "" {
		config.Options.Sort = ctx.Sort
	}

	if len(ctx.Paths) > 0 {
		config.Options.Paths = ctx.Paths
	}

	// Jira integration
	if ctx.JiraUsername != "" {
		config.Options.JiraUsername = ctx.JiraUsername
	}

	if ctx.JiraToken != "" {
		config.Options.JiraToken = ctx.JiraToken
	}

	if ctx.JiraURL != "" {
		config.Options.JiraURL = ctx.JiraURL
	}

	// Template-Pfad auflösen: Wenn der Template-Pfad relativ ist,
	// löse ihn relativ zum Konfigurationsverzeichnis auf (wie im Original)
	if config.Template != "" && !filepath.IsAbs(config.Template) {
		configDir := filepath.Dir(ctx.ConfigPath)
		config.Template = filepath.Join(configDir, config.Template)
	}

	// Processor konfigurieren
	if ctx.Processor != "" {
		factory := NewProcessorFactory()
		processor, err := factory.CreateProcessorFromString(ctx.Processor)
		if err != nil {
			return fmt.Errorf("failed to create processor: %w", err)
		}
		config.Options.Processor = processor
	}

	return nil
}

// LoadConfigFromCommand lädt die Konfiguration aus dem CommandConfig-Objekt
func LoadConfigFromCommand(cmdConfig *CommandConfig) (*Config, error) {
	logger := utils.NewLogger("debug", !cmdConfig.NoColor)
	logger.Debug("Loading configuration from command")

	// Überprüfen, ob die Konfigurationsdatei existiert
	if _, err := os.Stat(cmdConfig.ConfigPath); err != nil {
		return nil, fmt.Errorf("config file not found: %s", cmdConfig.ConfigPath)
	}

	// Template-Existenz prüfen
	if cmdConfig.Template != "" {
		if _, err := os.Stat(cmdConfig.Template); err != nil {
			return nil, fmt.Errorf("template file not found: %s", cmdConfig.Template)
		}
	}

	// CLI-Kontext erstellen
	ctx := &CLIContext{
		WorkingDir:       cmdConfig.WorkingDir,
		ConfigPath:       cmdConfig.ConfigPath,
		Template:         cmdConfig.Template,
		RepositoryURL:    cmdConfig.RepositoryURL,
		OutputPath:       cmdConfig.OutputPath,
		Silent:           cmdConfig.Silent,
		NoColor:          cmdConfig.NoColor,
		NoEmoji:          cmdConfig.NoEmoji,
		NoCaseSensitive:  cmdConfig.NoCaseSensitive,
		Query:            cmdConfig.Query,
		NextTag:          cmdConfig.NextTag,
		TagFilterPattern: cmdConfig.TagFilterPattern,
		JiraUsername:     cmdConfig.JiraUsername,
		JiraToken:        cmdConfig.JiraToken,
		JiraURL:          cmdConfig.JiraURL,
		Paths:            cmdConfig.Paths,
		Sort:             cmdConfig.Sort,
		Processor:        cmdConfig.Processor,
	}

	// Konfiguration laden (direkt YAML, kein TOML)
	config, err := LoadConfig(cmdConfig.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Konfiguration normalisieren
	if err := config.Normalize(ctx); err != nil {
		return nil, err
	}

	// Fertige Konfiguration zurückgeben
	return config, nil
}

// LoadConfig lädt die Konfiguration direkt aus der YAML-Datei
func LoadConfig(path string) (*Config, error) {
	loader := NewConfigLoader()
	return loader.Load(path)
}
