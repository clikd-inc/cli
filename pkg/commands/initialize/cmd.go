package initialize

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"clikd/pkg/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	forceFlag  bool
	globalFlag bool
	yesFlag    bool // Nicht-interaktiver Modus
	modernFlag bool // Moderne UI mit Bubble Tea
)

// NewInitCmd creates a new init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new clikd configuration",
		Long: `Initialize a new clikd configuration file.
This will create a new configuration file in the current directory or globally in your home directory.
If a configuration file already exists, it will not be overwritten unless the --force flag is used.

The initialization process can configure:
- General clikd settings
- AI integration configuration 
- Changelog generation settings

If run in a Git repository, it will automatically detect it and offer repository-specific options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Wenn modern Flag gesetzt ist, verwende Bubble Tea UI
			if modernFlag {
				return runInitWithBubbleTea(globalFlag, forceFlag)
			}

			// Ansonsten klassische UI verwenden
			return runInit(globalFlag, forceFlag, yesFlag)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&globalFlag, "global", "g", false, "Create a global configuration in ~/.clikd")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Non-interactive mode, use defaults")
	cmd.Flags().BoolVarP(&modernFlag, "modern", "m", true, "Use modern UI with Bubble Tea")

	// Add config subcommand
	cmd.AddCommand(newConfigCmd())

	return cmd
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [set|get|list] KEY=VALUE",
		Short: "Configure clikd settings",
		Long: `Configure clikd settings including API keys.
Set a configuration value with 'set' or view a value with 'get'.

Examples:
  # Set an API key in the global configuration
  clikd init config set ai.api_key=your_api_key_here

  # Get a configuration value
  clikd init config get ai.api_key
  
  # Set a nested configuration value
  clikd init config set ai.model=gpt-4
  
  # List all configuration keys
  clikd init config list
  
  # Show supported providers and models
  clikd init config list providers

Available key settings:
  ai.api_key      - API key for the selected AI provider (stored in global config only)
  ai.provider     - AI provider (e.g., "openai", "mistral", "anthropic")
  ai.model        - AI model to use with the provider (e.g., "gpt-4", "mistral-medium")
  general.log_level - Logging level (e.g., "info", "debug")`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires at least one argument: set or get")
			}

			// For API keys, always use the global config
			isAPIKey := false
			if len(args) > 1 && args[0] == "set" {
				keyValue := args[1]
				if strings.HasPrefix(keyValue, "ai.api_key=") {
					isAPIKey = true
				}
			} else if len(args) > 1 && args[0] == "get" {
				key := args[1]
				if key == "ai.api_key" {
					isAPIKey = true
				}
			}

			// Get config file path, for API keys always use global config
			useGlobal := isAPIKey || globalFlag
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
				if key == "ai.api_key" {
					// API key directly in AI configuration
					cfg := manager.GetConfig()
					cfg.AI.APIKey = value

					fmt.Printf("API key set for %s provider\n", cfg.AI.Provider)
				} else if key == "ai.provider" {
					// Provider ändern mit zusätzlichen Informationen
					currentProvider := manager.GetConfig().AI.Provider

					// Liste der unterstützten Modelle für den neuen Provider anzeigen
					supportedModels, err := config.GetSupportedModelsForProvider(value)
					if err == nil {
						fmt.Printf("Unterstützte Modelle für Provider '%s':\n", value)
						for _, model := range supportedModels {
							fmt.Printf("  - %s\n", model)
						}
					}

					// Standardmodell für den neuen Provider ermitteln
					defaultModel, err := config.GetDefaultModelForProvider(value)
					if err == nil {
						fmt.Printf("Standardmodell für Provider '%s': %s\n", value, defaultModel)
					}

					// Provider ändern
					if err := manager.SetConfigValue(key, value); err != nil {
						return fmt.Errorf("error setting provider: %w", err)
					}

					fmt.Printf("Provider von '%s' auf '%s' geändert.\n", currentProvider, value)
				} else if key == "ai.model" {
					// Modell ändern mit Validierung
					currentProvider := manager.GetConfig().AI.Provider

					// Prüfen, ob das Modell mit dem Provider kompatibel ist
					if err := config.ValidateProviderModel(currentProvider, value); err != nil {
						return fmt.Errorf("error setting model: %w", err)
					}

					// Modell ändern
					if err := manager.SetConfigValue(key, value); err != nil {
						return fmt.Errorf("error setting model: %w", err)
					}

					fmt.Printf("Modell auf '%s' geändert für Provider '%s'.\n", value, currentProvider)
				} else {
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

				// Handle API key special case
				if key == "ai.api_key" {
					cfg := manager.GetConfig()
					apiKey := cfg.AI.APIKey

					if apiKey == "" {
						fmt.Printf("API key is not set\n")
					} else {
						// Only show the first few characters for security
						maskedKey := maskAPIKey(apiKey)
						fmt.Printf("ai.api_key=%s\n", maskedKey)
					}
					return nil
				} else {
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
								case "model":
									fmt.Printf("%s=%s\n", key, cfg.AI.Model)
								case "provider":
									fmt.Printf("%s=%s\n", key, cfg.AI.Provider)
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

					fmt.Printf("Getting complex config values is not fully supported yet. Try a specific path like 'ai.model'\n")
					return nil
				}

			case "list":
				// List available configuration keys
				if len(args) > 1 && args[1] == "providers" {
					// Liste der unterstützten Provider anzeigen
					fmt.Println("Unterstützte AI-Provider:")
					for _, provider := range config.SupportedProviders {
						// Standardmodell für den Provider anzeigen
						defaultModel, _ := config.GetDefaultModelForProvider(provider)
						fmt.Printf("  %s (Standard-Modell: %s)\n", provider, defaultModel)

						// Modelle für den Provider anzeigen
						supportedModels, err := config.GetSupportedModelsForProvider(provider)
						if err == nil {
							fmt.Println("    Unterstützte Modelle:")
							for _, model := range supportedModels {
								fmt.Printf("    - %s\n", model)
							}
						}
						fmt.Println()
					}
					return nil
				}

				// Standard-Konfigurationsliste anzeigen
				fmt.Println("Verfügbare Konfigurationsschlüssel:")
				fmt.Println("\nAPI-Schlüssel (nur in globaler Konfiguration gespeichert):")
				fmt.Println("  ai.api_key")

				fmt.Println("\nAllgemeine Einstellungen:")
				fmt.Println("  general.log_level")
				fmt.Println("  general.color")

				fmt.Println("\nKI-Einstellungen:")
				fmt.Println("  ai.enable")
				fmt.Println("  ai.model")
				fmt.Println("  ai.provider")

				fmt.Println("\nChangelog-Einstellungen:")
				fmt.Println("  changelog.style")
				fmt.Println("  changelog.template")
				fmt.Println("  changelog.jira_integration")
				fmt.Println("  changelog.sort")

				fmt.Println("\nVerwenden Sie 'clikd init config list providers' für eine Liste unterstützter Provider und Modelle.")

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
		configDir := "clikd"

		// Prüfen, ob der Pfad existiert und ob es sich um einen Ordner handelt
		info, err := os.Stat(configDir)
		if err == nil {
			// Pfad existiert, prüfen ob es ein Ordner ist
			if !info.IsDir() {
				return "", fmt.Errorf("could not create config directory: %s already exists but is not a directory", configDir)
			}
		} else if os.IsNotExist(err) {
			// Pfad existiert nicht, erstellen
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return "", fmt.Errorf("could not create config directory: %w", err)
			}
		} else {
			// Anderer Fehler beim Zugriff
			return "", fmt.Errorf("could not check config directory: %w", err)
		}

		return filepath.Join(configDir, "config.toml"), nil
	}
}

// maskAPIKey maskiert einen API-Schlüssel für die Anzeige
func maskAPIKey(apiKey string) string {
	return config.MaskAPIKey(apiKey)
}

// printHeader gibt eine farbige Überschrift aus
func printHeader(text string) {
	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Printf("=== %s ===\n", text)
}

// printSuccess gibt eine Erfolgsmeldung aus
func printSuccess(text string) {
	color.New(color.FgGreen).Println(text)
}

// printWarning gibt eine Warnmeldung aus
func printWarning(text string) {
	color.New(color.FgYellow).Println(text)
}

// printError gibt eine Fehlermeldung aus
func printError(text string) {
	color.New(color.FgRed).Println(text)
}

// printInfo gibt eine Informationsmeldung aus
func printInfo(text string) {
	color.New(color.FgBlue).Println(text)
}

// promptUser fragt den Benutzer nach einer Eingabe mit farbig hervorgehobenem Standardwert
func promptUser(prompt string, defaultValue string) string {
	defaultText := ""
	if defaultValue != "" {
		if defaultValue == "J" || defaultValue == "j" || defaultValue == "Y" || defaultValue == "y" {
			defaultText = fmt.Sprintf(" [%s/%s]", color.HiGreenString("J"), "n")
		} else if defaultValue == "N" || defaultValue == "n" {
			defaultText = fmt.Sprintf(" [j/%s]", color.HiRedString("N"))
		} else {
			defaultText = fmt.Sprintf(" [%s]", color.HiGreenString(defaultValue))
		}
	}

	fmt.Printf("%s%s: ", prompt, defaultText)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// promptSelect zeigt eine nummerierte Auswahlliste an und gibt den ausgewählten Index zurück
func promptSelect(prompt string, options []string, defaultIndex int) int {
	fmt.Println()
	for i, option := range options {
		if i == defaultIndex {
			color.New(color.FgHiGreen).Printf("%d. %s (Standard)\n", i+1, option)
		} else {
			fmt.Printf("%d. %s\n", i+1, option)
		}
	}

	fmt.Println()
	defaultText := fmt.Sprintf("[Standard: %s]", color.HiGreenString("%d", defaultIndex+1))
	fmt.Printf("%s %s: ", prompt, defaultText)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultIndex
	}

	index := 0
	fmt.Sscanf(input, "%d", &index)
	if index > 0 && index <= len(options) {
		return index - 1
	}
	return defaultIndex
}

// runInit initializes a new config file
func runInit(global, force, nonInteractive bool) error {
	var configPath string
	var configDir string
	var isInGitRepo bool
	var repoURL string

	// ASCII Art-Logo für clikd
	color.New(color.FgHiMagenta, color.Bold).Println(`
   ______   __     __   __            __  
  / ____/  / /    /  | / /           / /  
 / /      / /    / /||/ / ____ ___  / /__ 
/ /____  / /___ / / |  / / __// _ \/  __/ 
\____/ /_____//_/  |_/ /_/   \___/\__/   
                                          
`)
	color.New(color.FgHiYellow).Println("Willkommen beim clikd-Konfigurations-Assistenten!")
	fmt.Println("Dieser Assistent hilft Ihnen bei der Einrichtung von clikd für Ihr Projekt.")

	// Prüfe, ob wir uns in einem Git-Repository befinden
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err == nil {
		isInGitRepo = true

		// Versuche, die Repository-URL zu ermitteln
		remoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
		if remoteOutput, err := remoteCmd.Output(); err == nil {
			repoURL = strings.TrimSpace(string(remoteOutput))
		}
	}

	// Wenn wir in einem Git-Repository sind und nicht im globalen Modus sind,
	// frage den Benutzer, ob er eine lokale oder globale Konfiguration möchte
	if isInGitRepo && !global && !nonInteractive {
		printInfo("Git-Repository erkannt: " + repoURL)
		response := promptUser("Möchten Sie eine lokale Konfiguration für dieses Repository erstellen?", "J")

		if strings.ToLower(response) == "n" || strings.ToLower(response) == "nein" || strings.ToLower(response) == "no" {
			printInfo("Erstelle stattdessen eine globale Konfiguration...")
			global = true
		} else {
			printInfo("Erstelle lokale Konfiguration für dieses Repository...")
		}
	}

	// Bestimme Konfigurationspfad
	if global {
		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			printError("Fehler beim Ermitteln des Home-Verzeichnisses: " + err.Error())
			return fmt.Errorf("error finding home directory: %w", err)
		}

		// Create global config directory
		configDir = filepath.Join(home, ".clikd")
	} else {
		// Lokaler Konfigurationsordner
		configDir = "clikd"
	}

	// Prüfen, ob der Pfad existiert und ob es sich um einen Ordner handelt
	info, err := os.Stat(configDir)
	if err == nil {
		// Pfad existiert, prüfen ob es ein Ordner ist
		if !info.IsDir() {
			printError("Fehler beim Erstellen des Konfigurationsverzeichnisses: " + configDir + " existiert bereits, ist aber kein Verzeichnis")
			return fmt.Errorf("error creating config directory: %s already exists but is not a directory", configDir)
		}
	} else if os.IsNotExist(err) {
		// Pfad existiert nicht, erstellen
		if err := os.MkdirAll(configDir, 0755); err != nil {
			printError("Fehler beim Erstellen des Konfigurationsverzeichnisses: " + err.Error())
			return fmt.Errorf("error creating config directory: %w", err)
		}
	} else {
		// Anderer Fehler beim Zugriff
		printError("Fehler beim Prüfen des Konfigurationsverzeichnisses: " + err.Error())
		return fmt.Errorf("error checking config directory: %w", err)
	}

	configPath = filepath.Join(configDir, "config.toml")

	// Prüfe, ob Konfiguration bereits existiert
	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		if !force && !nonInteractive {
			printWarning("Konfigurationsdatei existiert bereits unter " + configPath)
			response := promptUser("Möchten Sie die bestehende Konfiguration überschreiben?", "N")

			if strings.ToLower(response) != "j" && strings.ToLower(response) != "ja" && strings.ToLower(response) != "yes" && strings.ToLower(response) != "y" {
				printInfo("Abbruch, bestehende Konfiguration wird nicht überschrieben.")
				return nil
			}
			force = true
		} else if !force {
			printError("Konfigurationsdatei existiert bereits unter " + configPath + ". Verwenden Sie --force zum Überschreiben.")
			return fmt.Errorf("config file already exists at %s. Use --force to overwrite", configPath)
		}
	}

	// Create a new config manager
	manager := config.NewManager()

	// Lade bestehende Konfiguration, wenn sie existiert und nicht überschrieben werden soll
	if configExists && !force {
		if err := manager.InitConfig(configPath); err != nil {
			printError("Fehler beim Laden der bestehenden Konfiguration: " + err.Error())
			return fmt.Errorf("error loading existing config: %w", err)
		}
		printSuccess("Bestehende Konfiguration geladen.")
	} else {
		// Initialize with default config
		if err := manager.InitConfig(""); err != nil {
			printError("Fehler beim Initialisieren der Konfiguration: " + err.Error())
			return fmt.Errorf("error initializing config: %w", err)
		}
	}

	// Interaktive Konfiguration, wenn nicht im non-interaktiven Modus
	if !nonInteractive {
		// AI-Konfiguration
		printHeader("KI-Konfiguration")
		response := promptUser("KI-Funktionen aktivieren?", "J")

		aiEnabled := true
		if strings.ToLower(response) == "n" || strings.ToLower(response) == "nein" || strings.ToLower(response) == "no" {
			aiEnabled = false
		}
		manager.SetConfigValue("ai.enable", fmt.Sprintf("%t", aiEnabled))

		if aiEnabled {
			// Provider auswählen
			providerOptions := config.SupportedProviders
			providerInfo := make([]string, len(providerOptions))

			for i, provider := range providerOptions {
				defaultModel, _ := config.GetDefaultModelForProvider(provider)
				providerInfo[i] = fmt.Sprintf("%s (Standardmodell: %s)", provider, defaultModel)
			}

			selectedProviderIndex := promptSelect("Wählen Sie einen Provider", providerInfo, 0)
			selectedProvider := providerOptions[selectedProviderIndex]

			manager.SetConfigValue("ai.provider", selectedProvider)

			// Modelle für den ausgewählten Provider anzeigen
			supportedModels, _ := config.GetSupportedModelsForProvider(selectedProvider)
			defaultModel, _ := config.GetDefaultModelForProvider(selectedProvider)

			defaultModelIndex := 0
			for i, model := range supportedModels {
				if model == defaultModel {
					defaultModelIndex = i
					break
				}
			}

			selectedModelIndex := promptSelect(fmt.Sprintf("Wählen Sie ein Modell für %s", color.HiCyanString(selectedProvider)), supportedModels, defaultModelIndex)
			selectedModel := supportedModels[selectedModelIndex]

			manager.SetConfigValue("ai.model", selectedModel)

			// API-Schlüssel-Hinweis
			apiKeyInfo := ""
			switch selectedProvider {
			case "openai":
				apiKeyInfo = "https://platform.openai.com/api-keys"
			case "mistral":
				apiKeyInfo = "https://console.mistral.ai/api-keys/"
			case "anthropic":
				apiKeyInfo = "https://console.anthropic.com/settings/keys"
			default:
				apiKeyInfo = "der Website des Anbieters"
			}

			printHeader("API-Schlüssel-Konfiguration")
			if global {
				printInfo("Für globale KI-Konfiguration können Sie den API-Schlüssel wie folgt setzen:")
				color.New(color.FgHiCyan).Printf("  clikd init config set ai.api_key=IHR_API_SCHLÜSSEL\n")
			} else {
				printInfo("Für lokale Projekte erstellen Sie eine .env-Datei im Projektverzeichnis mit:")
				color.New(color.FgHiCyan).Printf("  CLIKD_%s_API_KEY=IHR_API_SCHLÜSSEL\n", strings.ToUpper(selectedProvider))
			}
			printInfo("API-Schlüssel erhalten Sie auf: " + color.HiCyanString(apiKeyInfo))
		}

		// Changelog-Konfiguration
		printHeader("Changelog-Konfiguration")
		response = promptUser("Changelog-Funktionen konfigurieren?", "J")

		setupChangelog := true
		if strings.ToLower(response) == "n" || strings.ToLower(response) == "nein" || strings.ToLower(response) == "no" {
			setupChangelog = false
		}

		if setupChangelog {
			// Stil auswählen
			styleOptions := []string{"github", "gitlab", "bitbucket"}
			selectedStyleIndex := promptSelect("Wählen Sie einen Changelog-Stil", styleOptions, 0)
			selectedStyle := styleOptions[selectedStyleIndex]

			manager.SetConfigValue("changelog.style", selectedStyle)

			// Repository-URL, wenn in einem Git-Repository
			if isInGitRepo && repoURL != "" {
				manager.SetConfigValue("changelog.repository_url", repoURL)
				printSuccess("Repository-URL auf " + color.HiGreenString(repoURL) + " gesetzt.")
			} else if isInGitRepo {
				response = promptUser("Geben Sie die Repository-URL ein", "")

				if response != "" {
					manager.SetConfigValue("changelog.repository_url", response)
					printSuccess("Repository-URL auf " + color.HiGreenString(response) + " gesetzt.")
				}
			}

			// Erweiterte Changelog-Einstellungen
			printInfo("Konfiguriere erweiterte Changelog-Einstellungen...")

			// HeaderPattern und PatternMaps für die Commit-Nachrichtenanalyse
			manager.SetConfigValue("changelog.options.header.pattern", "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$")

			// PatternMaps als Array setzen
			patternMaps := []string{"Type", "Scope", "Subject"}
			for i, pattern := range patternMaps {
				manager.SetConfigValue(fmt.Sprintf("changelog.options.header.pattern_maps.%d", i), pattern)
			}

			// Commit-Filter konfigurieren - als Array für jeden Typ
			commitTypes := []string{"feat", "fix", "perf", "refactor", "chore"}
			for i, commitType := range commitTypes {
				manager.SetConfigValue(fmt.Sprintf("changelog.options.commits.filters.Type.%d", i), commitType)
			}

			// Gruppierung der Commits
			manager.SetConfigValue("changelog.options.commit_groups.group_by", "Type")
			manager.SetConfigValue("changelog.options.commit_groups.sort_by", "Title")

			// Titel-Mappings für Commit-Gruppen
			manager.SetConfigValue("changelog.options.commit_groups.title_maps.feat", "Features")
			manager.SetConfigValue("changelog.options.commit_groups.title_maps.fix", "Bug Fixes")
			manager.SetConfigValue("changelog.options.commit_groups.title_maps.perf", "Performance Improvements")
			manager.SetConfigValue("changelog.options.commit_groups.title_maps.refactor", "Code Refactoring")
			manager.SetConfigValue("changelog.options.commit_groups.title_maps.chore", "Chores")

			// Note Keywords für Breaking Changes usw.
			noteKeywords := []string{"BREAKING CHANGE", "SECURITY"}
			for i, keyword := range noteKeywords {
				manager.SetConfigValue(fmt.Sprintf("changelog.options.notes.keywords.%d", i), keyword)
			}
		}
	} else {
		// Non-interaktiver Modus: Standardwerte setzen
		printInfo("Verwende Standardwerte im nicht-interaktiven Modus...")

		// Standard-AI-Konfiguration
		manager.SetConfigValue("ai.enable", "true")
		manager.SetConfigValue("ai.provider", "mistral")
		manager.SetConfigValue("ai.model", "mistral-medium")

		// Standard-Changelog-Konfiguration
		manager.SetConfigValue("changelog.style", "github")

		// Wenn wir in einem Git-Repository sind, setze die Repository-URL
		if isInGitRepo && repoURL != "" {
			manager.SetConfigValue("changelog.repository_url", repoURL)
		}

		// HeaderPattern und PatternMaps für die Commit-Nachrichtenanalyse
		manager.SetConfigValue("changelog.options.header.pattern", "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$")

		// PatternMaps als Array setzen
		patternMaps := []string{"Type", "Scope", "Subject"}
		for i, pattern := range patternMaps {
			manager.SetConfigValue(fmt.Sprintf("changelog.options.header.pattern_maps.%d", i), pattern)
		}

		// Commit-Filter konfigurieren - als Array für jeden Typ
		commitTypes := []string{"feat", "fix", "perf", "refactor", "chore"}
		for i, commitType := range commitTypes {
			manager.SetConfigValue(fmt.Sprintf("changelog.options.commits.filters.Type.%d", i), commitType)
		}

		// Gruppierung der Commits
		manager.SetConfigValue("changelog.options.commit_groups.group_by", "Type")
		manager.SetConfigValue("changelog.options.commit_groups.sort_by", "Title")

		// Titel-Mappings für Commit-Gruppen
		manager.SetConfigValue("changelog.options.commit_groups.title_maps.feat", "Features")
		manager.SetConfigValue("changelog.options.commit_groups.title_maps.fix", "Bug Fixes")
		manager.SetConfigValue("changelog.options.commit_groups.title_maps.perf", "Performance Improvements")
		manager.SetConfigValue("changelog.options.commit_groups.title_maps.refactor", "Code Refactoring")
		manager.SetConfigValue("changelog.options.commit_groups.title_maps.chore", "Chores")

		// Note Keywords für Breaking Changes usw.
		noteKeywords := []string{"BREAKING CHANGE", "SECURITY"}
		for i, keyword := range noteKeywords {
			manager.SetConfigValue(fmt.Sprintf("changelog.options.notes.keywords.%d", i), keyword)
		}
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
		printError("Fehler beim Erstellen des Templates-Verzeichnisses: " + err.Error())
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
		printError("Fehler beim Erstellen des Cache-Verzeichnisses: " + err.Error())
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
		printError("Fehler beim Erstellen des Changelog-Templates: " + err.Error())
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
		printError("Fehler beim Speichern der Konfiguration: " + err.Error())
		return fmt.Errorf("error saving config: %w", err)
	}

	// Prüfen, ob KI aktiviert ist (für "Nächste Schritte")
	cfg := manager.GetConfig()
	aiEnabled := cfg.AI.Enable

	printHeader("Konfiguration abgeschlossen")
	printSuccess("Konfiguration erfolgreich initialisiert unter " + color.HiGreenString(configPath))
	color.New(color.FgHiYellow).Println("\nProjektstruktur erstellt:")
	if global {
		home, _ := os.UserHomeDir()
		basePath := filepath.Join(home, ".clikd")
		color.New(color.FgCyan).Printf("  ├── %s/\n", basePath)
	} else {
		color.New(color.FgCyan).Println("  ├── clikd/")
	}
	color.New(color.FgHiCyan).Println("  │   ├── config.toml")
	color.New(color.FgCyan).Println("  │   ├── templates/")
	color.New(color.FgHiCyan).Println("  │   │   └── changelog.md")
	color.New(color.FgCyan).Println("  │   └── cache/")

	printHeader("Nächste Schritte")
	if aiEnabled {
		printInfo("1. API-Schlüssel konfigurieren (falls noch nicht geschehen)")
		if global {
			color.New(color.FgHiCyan).Printf("   clikd init config set ai.api_key=IHR_API_SCHLÜSSEL\n")
		} else {
			color.New(color.FgHiCyan).Printf("   Erstellen Sie eine .env-Datei mit CLIKD_*_API_KEY\n")
		}
	}
	printInfo("2. Einen Changelog generieren")
	color.New(color.FgHiCyan).Printf("   clikd changelog -o CHANGELOG.md\n")

	fmt.Println()
	printSuccess("clikd ist jetzt einsatzbereit! Viel Spaß beim Verwenden!")

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
