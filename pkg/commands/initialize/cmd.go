package initialize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clikd/pkg/config"

	"github.com/spf13/cobra"
)

var (
	forceFlag  bool
	globalFlag bool
)

// NewInitCmd creates a new init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new clikd configuration",
		Long: `Initialize a new clikd configuration file.
This will create a new configuration file in the current directory or globally in your home directory.
If a configuration file already exists, it will not be overwritten unless the --force flag is used.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(globalFlag, forceFlag)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&globalFlag, "global", "g", false, "Create a global configuration in ~/.clikd")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing configuration")

	// Add config subcommand
	cmd.AddCommand(newConfigCmd())

	return cmd
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [set|get] KEY=VALUE",
		Short: "Configure clikd settings",
		Long: `Configure clikd settings including API keys.
Set a configuration value with 'set' or view a value with 'get'.

Examples:
  # Set an API key in the global configuration
  clikd init config set CLIKD_MISTRAL_API_KEY=your_api_key_here

  # Get a configuration value
  clikd init config get CLIKD_MISTRAL_API_KEY
  
  # Set a nested configuration value
  clikd init config set ai.default_model=gpt-4
  
Available API key settings:
  CLIKD_MISTRAL_API_KEY      - API key for Mistral AI models (stored in global config only)
  CLIKD_OPENAI_API_KEY       - API key for OpenAI models (stored in global config only)
  CLIKD_ANTHROPIC_API_KEY    - API key for Anthropic models (stored in global config only)
  CLIKD_AZURE_OPENAI_API_KEY - API key for Azure OpenAI models (stored in global config only)
  CLIKD_JIRA_API_KEY         - API key for JIRA integration (stored in global config only)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires at least one argument: set or get")
			}

			// For API keys, always use the global config
			isAPIKey := false
			if len(args) > 1 && args[0] == "set" {
				keyValue := args[1]
				for _, prefix := range []string{"CLIKD_MISTRAL_API_KEY", "CLIKD_OPENAI_API_KEY", "CLIKD_ANTHROPIC_API_KEY", "CLIKD_AZURE_OPENAI_API_KEY", "CLIKD_JIRA_API_KEY"} {
					if strings.HasPrefix(keyValue, prefix+"=") {
						isAPIKey = true
						break
					}
				}
			} else if len(args) > 1 && args[0] == "get" {
				key := args[1]
				for _, prefix := range []string{"CLIKD_MISTRAL_API_KEY", "CLIKD_OPENAI_API_KEY", "CLIKD_ANTHROPIC_API_KEY", "CLIKD_AZURE_OPENAI_API_KEY", "CLIKD_JIRA_API_KEY"} {
					if key == prefix {
						isAPIKey = true
						break
					}
				}
			}

			// Get config file path, for API keys always use global config
			useGlobal := isAPIKey
			configPath, err := getConfigPath(useGlobal)
			if err != nil {
				return err
			}

			// Create a new config manager
			manager := config.NewManager()

			// Try to load existing config if it exists
			if _, err := os.Stat(configPath); err == nil {
				if err := manager.InitConfig(configPath); err != nil {
					return fmt.Errorf("error loading existing config: %w", err)
				}
			} else {
				// Initialize with default config
				if err := manager.InitConfig(""); err != nil {
					return fmt.Errorf("error initializing config: %w", err)
				}
			}

			// Handle set/get commands
			switch args[0] {
			case "set":
				if len(args) < 2 {
					return fmt.Errorf("'set' requires a KEY=VALUE argument")
				}

				// Process the KEY=VALUE format
				keyValue := args[1]
				var key, value string

				for i, c := range keyValue {
					if c == '=' {
						key = keyValue[:i]
						value = keyValue[i+1:]
						break
					}
				}

				if key == "" || value == "" {
					return fmt.Errorf("invalid format, use KEY=VALUE")
				}

				// Handle API key special cases
				switch key {
				case "CLIKD_MISTRAL_API_KEY", "CLIKD_OPENAI_API_KEY", "CLIKD_ANTHROPIC_API_KEY", "CLIKD_AZURE_OPENAI_API_KEY":
					// Map the environment variable name to the provider name
					var providerName string
					switch key {
					case "CLIKD_MISTRAL_API_KEY":
						providerName = "mistral"
					case "CLIKD_OPENAI_API_KEY":
						providerName = "openai"
					case "CLIKD_ANTHROPIC_API_KEY":
						providerName = "anthropic"
					case "CLIKD_AZURE_OPENAI_API_KEY":
						providerName = "azure-openai"
					}

					// Get models for this provider and set the API key for each
					cfg := manager.GetConfig()
					modelsUpdated := false

					for modelName, model := range cfg.AI.Models {
						if model.Provider == providerName {
							model.APIKey = value
							cfg.AI.Models[modelName] = model
							modelsUpdated = true
						}
					}

					if !modelsUpdated {
						return fmt.Errorf("no models found for provider %s", providerName)
					}

				case "CLIKD_JIRA_API_KEY":
					// Set the JIRA API key
					cfg := manager.GetConfig()
					cfg.Changelog.Jira.APIKey = value

				default:
					// For regular config settings, use SetConfigValue
					if err := manager.SetConfigValue(key, value); err != nil {
						return fmt.Errorf("error setting config value: %w", err)
					}
				}

				// Save the configuration
				if err := manager.SaveConfig(configPath); err != nil {
					return fmt.Errorf("error saving config: %w", err)
				}

				if isAPIKey {
					fmt.Printf("API key '%s' set in global config at %s\n", key, configPath)
				} else {
					fmt.Printf("Configuration value '%s' set in %s\n", key, configPath)
				}
				return nil

			case "get":
				if len(args) < 2 {
					return fmt.Errorf("'get' requires a KEY argument")
				}

				key := args[1]

				// Handle API key special cases
				switch key {
				case "CLIKD_MISTRAL_API_KEY", "CLIKD_OPENAI_API_KEY", "CLIKD_ANTHROPIC_API_KEY", "CLIKD_AZURE_OPENAI_API_KEY":
					// Map the environment variable name to the provider name
					var providerName string
					switch key {
					case "CLIKD_MISTRAL_API_KEY":
						providerName = "mistral"
					case "CLIKD_OPENAI_API_KEY":
						providerName = "openai"
					case "CLIKD_ANTHROPIC_API_KEY":
						providerName = "anthropic"
					case "CLIKD_AZURE_OPENAI_API_KEY":
						providerName = "azure-openai"
					}

					// Find a model that uses this provider
					cfg := manager.GetConfig()
					var apiKey string

					for _, model := range cfg.AI.Models {
						if model.Provider == providerName && model.APIKey != "" {
							apiKey = model.APIKey
							break
						}
					}

					if apiKey == "" {
						fmt.Printf("API key '%s' is not set\n", key)
					} else {
						// Only show the first few characters for security
						maskedKey := maskAPIKey(apiKey)
						fmt.Printf("%s=%s\n", key, maskedKey)
					}
					return nil

				case "CLIKD_JIRA_API_KEY":
					// Get the JIRA API key
					cfg := manager.GetConfig()
					apiKey := cfg.Changelog.Jira.APIKey

					if apiKey == "" {
						fmt.Printf("API key '%s' is not set\n", key)
					} else {
						// Only show the first few characters for security
						maskedKey := maskAPIKey(apiKey)
						fmt.Printf("%s=%s\n", key, maskedKey)
					}
					return nil

				default:
					// Get a regular config value
					cfg := manager.GetConfig()

					// Basic dot notation support for nested values
					parts := strings.Split(key, ".")
					if len(parts) > 1 {
						// Special handling for common nested values
						switch parts[0] {
						case "ai":
							if len(parts) > 1 {
								switch parts[1] {
								case "enable":
									fmt.Printf("%s=%v\n", key, cfg.AI.Enable)
								case "default_model":
									fmt.Printf("%s=%s\n", key, cfg.AI.DefaultModel)
								case "default_provider":
									fmt.Printf("%s=%s\n", key, cfg.AI.DefaultProvider)
								case "verbose":
									fmt.Printf("%s=%v\n", key, cfg.AI.Verbose)
								default:
									fmt.Printf("Nested key '%s' not found or not supported\n", key)
								}
								return nil
							}
						case "general":
							if len(parts) > 1 {
								switch parts[1] {
								case "log_level":
									fmt.Printf("%s=%s\n", key, cfg.General.LogLevel)
								case "color":
									fmt.Printf("%s=%v\n", key, cfg.General.Color)
								default:
									fmt.Printf("Nested key '%s' not found or not supported\n", key)
								}
								return nil
							}
						case "changelog":
							if len(parts) > 1 {
								switch parts[1] {
								case "style":
									fmt.Printf("%s=%s\n", key, cfg.Changelog.Style)
								case "template":
									fmt.Printf("%s=%s\n", key, cfg.Changelog.Template)
								default:
									fmt.Printf("Nested key '%s' not found or not supported\n", key)
								}
								return nil
							}
						}
					}

					fmt.Printf("Getting complex config values is not fully supported yet. Try a specific path like 'ai.default_model'\n")
					return nil
				}

			case "list":
				// List available configuration keys
				fmt.Println("Available configuration keys:")
				fmt.Println("\nAPI Keys (stored in global config only):")
				fmt.Println("  CLIKD_MISTRAL_API_KEY")
				fmt.Println("  CLIKD_OPENAI_API_KEY")
				fmt.Println("  CLIKD_ANTHROPIC_API_KEY")
				fmt.Println("  CLIKD_AZURE_OPENAI_API_KEY")
				fmt.Println("  CLIKD_JIRA_API_KEY")

				fmt.Println("\nGeneral Settings:")
				fmt.Println("  general.log_level")
				fmt.Println("  general.color")

				fmt.Println("\nAI Settings:")
				fmt.Println("  ai.enable")
				fmt.Println("  ai.default_model")
				fmt.Println("  ai.default_provider")
				fmt.Println("  ai.verbose")

				fmt.Println("\nChangelog Settings:")
				fmt.Println("  changelog.style")
				fmt.Println("  changelog.template")
				fmt.Println("  changelog.jira_integration")
				fmt.Println("  changelog.sort")

				return nil

			default:
				return fmt.Errorf("unknown command: %s (use 'set', 'get', or 'list')", args[0])
			}
		},
	}

	return cmd
}

// getConfigPath gibt den Pfad zur Konfigurationsdatei zurück
// Wenn useGlobal true ist, wird der Pfad zur globalen Konfiguration im Home-Verzeichnis zurückgegeben
// Ansonsten wird der Pfad zur lokalen Konfiguration im aktuellen Verzeichnis zurückgegeben
func getConfigPath(useGlobal bool) (string, error) {
	if useGlobal {
		// Globale Konfiguration im Home-Verzeichnis
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get home directory: %w", err)
		}

		// Stelle sicher, dass das .clikd-Verzeichnis existiert
		configDir := filepath.Join(homeDir, ".clikd")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("could not create config directory: %w", err)
		}

		return filepath.Join(configDir, "config.toml"), nil
	} else {
		// Lokale Konfiguration im aktuellen Verzeichnis
		// Stelle sicher, dass das .clikd-Verzeichnis existiert
		configDir := ".clikd"
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("could not create config directory: %w", err)
		}

		return filepath.Join(configDir, "config.toml"), nil
	}
}

// maskAPIKey maskiert einen API-Schlüssel für die Anzeige
func maskAPIKey(apiKey string) string {
	return config.MaskAPIKey(apiKey)
}

// runInit initializes a new config file
func runInit(global, force bool) error {
	var configPath string

	if global {
		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error finding home directory: %w", err)
		}

		// Create global config directory
		configDir := filepath.Join(home, ".clikd")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}

		configPath = filepath.Join(configDir, "config.toml")
	} else {
		// Create local config directory
		if err := os.MkdirAll("clikd", 0755); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}

		configPath = filepath.Join("clikd", "config.toml")
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite", configPath)
	}

	// Create a new config manager
	manager := config.NewManager()

	// Initialize with default config
	if err := manager.InitConfig(""); err != nil {
		return fmt.Errorf("error initializing config: %w", err)
	}

	// Create templates directory
	var templatesDir string
	if global {
		home, _ := os.UserHomeDir()
		templatesDir = filepath.Join(home, ".clikd", "templates")
	} else {
		templatesDir = filepath.Join("clikd", "templates")
	}

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("error creating templates directory: %w", err)
	}

	// Create cache directory
	var cacheDir string
	if global {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".clikd", "cache")
	} else {
		cacheDir = filepath.Join("clikd", "cache")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %w", err)
	}

	// Create default changelog template
	templatePath := filepath.Join(templatesDir, "changelog.md")
	templateContent := `# {{ .Info.Title }}

All notable changes to this project will be documented in this file.

{{ if .Versions -}}
{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }} ({{ datetime "2006-01-02" .Tag.Date }})

{{ range .CommitGroups -}}
### {{ .Title }}

{{ range .Commits -}}
* {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .RevertCommits -}}
### Reverts

{{ range .RevertCommits -}}
* {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
### Merges

{{ range .MergeCommits -}}
* {{ .Header }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}

{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("error creating changelog template: %w", err)
	}

	// Update template path in config
	if global {
		manager.SetConfigValue("changelog.template", filepath.Join("templates", "changelog.md"))
	} else {
		manager.SetConfigValue("changelog.template", filepath.Join("templates", "changelog.md"))
	}

	// Save the configuration
	if err := manager.SaveConfig(configPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	fmt.Printf("Configuration initialized at %s\n", configPath)
	fmt.Println("Project structure created:")
	if global {
		home, _ := os.UserHomeDir()
		basePath := filepath.Join(home, ".clikd")
		fmt.Printf("  ├── %s/\n", basePath)
	} else {
		fmt.Println("  ├── clikd/")
	}
	fmt.Println("  │   ├── config.toml")
	fmt.Println("  │   ├── templates/")
	fmt.Println("  │   │   └── changelog.md")
	fmt.Println("  │   └── cache/")

	return nil
}

// fileExists prüft, ob eine Datei existiert
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
