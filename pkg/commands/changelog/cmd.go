package changelog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"clikd/pkg/config"
	"clikd/pkg/internal/changelog"
	"clikd/pkg/internal/changelog/initializer"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
)

// Variablen für Flags
var (
	// initFlag wurde entfernt, da die Funktionalität in den init-Befehl integriert wurde
	configFlag           string
	templateFlag         string
	repositoryURLFlag    string
	outputFlag           string
	nextTagFlag          string
	silentFlag           bool
	noColorFlag          bool
	noEmojiFlag          bool
	noCaseFlag           bool
	tagFilterPatternFlag string
	jiraURLFlag          string
	jiraUsernameFlag     string
	jiraTokenFlag        string
	sortFlag             string
	pathsFlag            []string
	// AI-Flags für bestimmte Funktionen bleiben erhalten für fine-grained control
)

// NewChangelogCmd erstellt ein neues Changelog-Kommando
func NewChangelogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog [options] <tag query>",
		Short: "Generate a changelog from git history",
		Long: `Generate a changelog from git history using conventional commits.

This command generates a changelog based on conventional commit messages
in your git history. It supports various customization options and integrations.

Tag Query Formats:
  1. <old>..<new>  - Commits between <old> tag and <new> tag (e.g., v1.0.0..v2.0.0)
  2. <tag>..       - Commits from <tag> to the latest (e.g., v1.0.0..)
  3. ..<tag>       - Commits from the oldest tag to <tag> (e.g., ..v1.0.0)
  4. <tag>         - Commits contained in <tag> only (e.g., v1.0.0)

Special Features:
  - Jira Integration: Automatically fetches ticket information when Jira IDs are 
    present in commit messages. Configure via --jira-* flags or environment variables.
  - Path Filtering: Filter commits by specific files or directories with --path.
  - Tag Filtering: Filter tags using regular expressions with --tag-filter-pattern.
  - Semver Sorting: Sort tags by semantic version instead of date with --sort=semver.
  - AI Features: Enable or disable AI features via global configuration ('ai.enable=true').
    The --ai flag can be used to override this setting for a single command.

AI Configuration:
  AI features can be configured in the global configuration file:
  - Enable/disable AI: 'ai.enable=true/false'
  - Set default model: 'ai.default_model=mistral-medium'
  
  For local projects, create a .env file with the required API keys.
  For global usage, add API keys to the global configuration.

Examples:
  # Initialize changelog configuration with the global init command
  clikd init

  # Generate changelog for all tags to stdout (uses configuration setting for AI)
  clikd changelog

  # Generate changelog to file (uses configuration setting for AI)
  clikd changelog -o CHANGELOG.md

  # Override configuration and enable AI for this command only
  clikd changelog --ai -o CHANGELOG.md

  # Override configuration and disable AI for this command only
  clikd changelog --ai=false -o CHANGELOG.md

  # Generate changelog for specific tag range
  clikd changelog v1.0.0..v2.0.0 -o CHANGELOG.md

  # Generate changelog with "unreleased" commits as next version
  clikd changelog --next-tag v2.0.0 -o CHANGELOG.md

  # Filter commits by path
  clikd changelog --path="pkg/,cmd/" -o CHANGELOG.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger("info", true)

			// KI-Funktionalität initialisieren
			if err := InitializeAI(); err != nil {
				logger.Error("Fehler bei der KI-Initialisierung: %v", err)
				logger.Info("Changelog wird ohne KI-Funktionen generiert")
				// Wir gehen weiter, auch wenn die KI-Initialisierung fehlschlägt
			} else if changelog.IsAIEnabled() {
				// Zeige KI-Status, wenn KI aktiviert ist
				ShowAIStatus()
			} else {
				logger.Info("Changelog wird ohne KI-Funktionen generiert")
				logger.Info("Um KI zu aktivieren, verwenden Sie --ai oder setzen Sie ai.enable=true in der Konfiguration")
			}

			// Sonst den normalen Changelog-Generator ausführen
			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			return runGenerator(query)
		},
	}

	// Flags hinzufügen - initFlag wurde entfernt
	cmd.Flags().StringSliceVar(&pathsFlag, "path", []string{}, "Filter commits by path(s). Can use multiple times.")
	cmd.Flags().StringVarP(&configFlag, "config", "c", "clikd/config.toml", "specifies a different configuration file to pick up")
	cmd.Flags().StringVarP(&templateFlag, "template", "t", "", "specifies a template file to pick up. If not specified, use the one in config")
	cmd.Flags().StringVar(&repositoryURLFlag, "repository-url", "", "specifies git repo URL. If not specified, use 'repository_url' in config")
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output path and filename for the changelogs. If not specified, output to stdout")
	cmd.Flags().StringVar(&nextTagFlag, "next-tag", "", "treat unreleased commits as specified tags (EXPERIMENTAL)")
	cmd.Flags().BoolVar(&silentFlag, "silent", false, "disable stdout output")
	cmd.Flags().BoolVar(&noColorFlag, "no-color", false, "disable color output")
	cmd.Flags().BoolVar(&noEmojiFlag, "no-emoji", false, "disable emoji output")
	cmd.Flags().BoolVar(&noCaseFlag, "no-case", false, "disable case sensitive filters")
	cmd.Flags().StringVar(&tagFilterPatternFlag, "tag-filter-pattern", "", "Regular expression of tag filter. Is specified, only matched tags will be picked")
	cmd.Flags().StringVar(&jiraURLFlag, "jira-url", "", "Jira URL")
	cmd.Flags().StringVar(&jiraUsernameFlag, "jira-username", "", "Jira username")
	cmd.Flags().StringVar(&jiraTokenFlag, "jira-token", "", "Jira token")
	cmd.Flags().StringVar(&sortFlag, "sort", "date", "Specify how to sort tags; currently supports \"date\" or by \"semver\"")

	// KI-bezogene Flags hinzufügen (für fine-grained control)
	AddAIFlags(cmd)

	return cmd
}

// runInitializer-Funktion wurde entfernt, da die Funktionalität in den init-Befehl integriert wurde

// runGenerator führt den Changelog-Generator aus
func runGenerator(query string) error {
	logger := utils.NewLogger("debug", true) // Log-Level auf debug setzen für detailliertere Ausgaben
	logger.Info("Generating changelog for query: %s", query)

	// Eigene Debug-Ausgabe für Git-Informationen
	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory: %v", err)
		return err
	}
	logger.Debug("Working directory: %s", wd)

	// Für Testzwecke: Explizites Setzen des Test-Repository-Pfads
	testRepoDir := "/Users/nyxb/Projects/nyxb/cli/clikd/test_repo"
	logger.Debug("Test repository directory: %s", testRepoDir)

	// Git-Tags auflisten (Debug)
	if out, err := exec.Command("git", "-C", testRepoDir, "tag", "-l").Output(); err == nil {
		logger.Debug("Available git tags: %s", string(out))
	} else {
		logger.Debug("Failed to list git tags: %v", err)
	}

	// Git-Log anzeigen (Debug)
	if out, err := exec.Command("git", "-C", testRepoDir, "log", "--pretty=format:%h - %s", "-n", "5").Output(); err == nil {
		logger.Debug("Recent git commits: %s", string(out))
	} else {
		logger.Debug("Failed to list git commits: %v", err)
	}

	// Konfigurationspfad auflösen
	configPath := resolveConfigPath(configFlag)
	logger.Debug("Resolved config path: %s", configPath)

	// Überprüfen, ob configPath ein Verzeichnis oder eine Datei ist
	configFileInfo, err := os.Stat(configPath)
	if err != nil {
		logger.Error("Configuration file not found at %s: %v", configPath, err)
		return fmt.Errorf("configuration file not found at %s: %v", configPath, err)
	}
	if configFileInfo.IsDir() {
		// Wenn es ein Verzeichnis ist, nehmen wir an, dass die Konfiguration in config.toml ist
		configPath = filepath.Join(configPath, "config.toml")
		logger.Debug("Config path is a directory, using %s instead", configPath)
	}

	templatePath := ""

	// Prüfen, ob ein hardcoded Template-Pfad funktioniert
	directTemplatePath := filepath.Join(wd, "clikd", "templates", "changelog.md")
	if _, err := os.Stat(directTemplatePath); err == nil {
		logger.Debug("Found template at hardcoded path: %s", directTemplatePath)
		templatePath = directTemplatePath
	} else {
		logger.Debug("Hardcoded template path not found: %s", directTemplatePath)
	}

	// Laden der clikd-Konfiguration, um die Template-Einstellungen zu erhalten
	if _, err := os.Stat(configPath); err == nil {
		// Die TOML-Konfiguration existiert
		logger.Debug("Found configuration at %s", configPath)

		// Global configuration manager
		cfg, err := config.Get()
		if err != nil {
			logger.Debug("Error getting config: %v", err)
		} else {
			// Get template path from clikd config
			tmplRelPath := cfg.Changelog.Template
			logger.Debug("Template path from config: %s", tmplRelPath)

			// Wenn der Template-Pfad in der Konfiguration leer ist, verwenden wir unseren direkten Pfad
			if tmplRelPath == "" {
				logger.Debug("Template path in config is empty, using hardcoded path")
				// Wir verwenden den zuvor gefundenen hardcoded Pfad
				if templatePath != "" {
					logger.Debug("Using previously found template path: %s", templatePath)
				} else {
					logger.Error("No template path found")
					return fmt.Errorf("no template path found")
				}
			} else {
				// Handle both absolute and relative paths
				if filepath.IsAbs(tmplRelPath) {
					logger.Debug("Template path is absolute")
					templatePath = tmplRelPath
				} else {
					// If relative path, it's relative to the clikd directory
					clikdDir := filepath.Dir(configPath)
					logger.Debug("clikdDir: %s", clikdDir)

					calculatedPath := filepath.Join(clikdDir, tmplRelPath)
					logger.Debug("Calculated template path: %s", calculatedPath)

					// Überprüfen, ob die berechnete Datei existiert
					if _, err := os.Stat(calculatedPath); err == nil {
						logger.Debug("Template file exists at: %s", calculatedPath)
						templatePath = calculatedPath
					} else {
						logger.Debug("Template file not found at calculated path: %s", calculatedPath)

						// Verschiedene Alternativen durchprobieren
						alternatives := []string{
							filepath.Join(wd, "clikd", "templates", "changelog.md"),
							filepath.Join(clikdDir, "templates", "changelog.md"),
							filepath.Join(wd, tmplRelPath),
						}

						for _, altPath := range alternatives {
							logger.Debug("Trying alternative path: %s", altPath)
							if _, err := os.Stat(altPath); err == nil {
								logger.Debug("Found template at alternative path: %s", altPath)
								templatePath = altPath
								break
							}
						}
					}
				}
			}

			logger.Debug("Final template path: %s", templatePath)

			// Prüfen, ob die Template-Datei existiert
			if templatePath == "" {
				logger.Error("No valid template path found")
				return fmt.Errorf("no valid template path found")
			}

			if fileInfo, err := os.Stat(templatePath); err != nil {
				logger.Error("Template file not found at: %s", templatePath)
				return err
			} else if fileInfo.IsDir() {
				// Es ist ein Verzeichnis, keine Datei
				logger.Error("Template path is a directory, not a file: %s", templatePath)
				return fmt.Errorf("template path is a directory, not a file: %s", templatePath)
			}

			// Override template flag if it was not explicitly set
			if templateFlag == "" {
				templateFlag = templatePath
				logger.Debug("Setting template flag to: %s", templateFlag)
			} else {
				logger.Debug("Template flag already set to: %s", templateFlag)
			}
		}
	} else {
		// Wenn die Konfigurationsdatei nicht existiert, einen Fehler zurückgeben
		logger.Error("Configuration file not found at %s", configPath)
		return fmt.Errorf("configuration file not found at %s", configPath)
	}

	// CLI-Kontext erstellen
	ctx := &initializer.CLIContext{
		WorkingDir:       testRepoDir, // Explizit das Test-Repository-Verzeichnis verwenden
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		ConfigPath:       configPath,
		Template:         templateFlag,
		RepositoryURL:    repositoryURLFlag,
		OutputPath:       outputFlag,
		Silent:           silentFlag,
		NoColor:          noColorFlag,
		NoEmoji:          noEmojiFlag,
		NoCaseSensitive:  noCaseFlag,
		Query:            query,
		NextTag:          nextTagFlag,
		TagFilterPattern: tagFilterPatternFlag,
		JiraUsername:     jiraUsernameFlag,
		JiraToken:        jiraTokenFlag,
		JiraURL:          jiraURLFlag,
		Paths:            pathsFlag,
		Sort:             sortFlag,
	}

	logger.Debug("CLI Context - WorkingDir: %s", ctx.WorkingDir)
	logger.Debug("CLI Context - ConfigPath: %s", ctx.ConfigPath)
	logger.Debug("CLI Context - Template: %s", ctx.Template)

	// Explizit das Template im Kontext setzen, um die Template-Pfad-Normalisierung zu überspringen
	if templatePath != "" {
		ctx.Template = templatePath
		logger.Debug("Overriding CLI Context Template to: %s", ctx.Template)
	}

	// Konfiguration laden
	loader := initializer.NewConfigLoader()

	config, err := loader.Load(ctx)
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return err
	}

	// Explizit das Arbeitsverzeichnis auf das Test-Repository setzen
	config.WorkingDir = testRepoDir
	logger.Debug("Set config WorkingDir to: %s", config.WorkingDir)

	logger.Debug("Loaded config - Template: %s", config.Template)

	// Sicherstellen, dass der Template-Pfad korrekt in der Konfiguration gesetzt ist
	if templatePath != "" {
		config.Template = templatePath
		logger.Debug("Set final template path in config to: %s", config.Template)
	}

	logger.Debug("Final template path used: %s", config.Template)

	// Prüfen, ob diese Datei tatsächlich existiert
	if _, err := os.Stat(config.Template); err != nil {
		logger.Error("Final template file not found at: %s", config.Template)
		return err
	}

	// Changelog-Generator erstellen und ausführen
	chglogLogger := changelog.NewLogger(os.Stdout, os.Stderr, ctx.Silent, ctx.NoEmoji)
	generator := changelog.NewGenerator(chglogLogger, config)

	// Output-Writer erstellen
	writer, err := createOutputWriter(ctx.OutputPath)
	if err != nil {
		logger.Error("Failed to create output writer: %v", err)
		return err
	}
	defer closeOutputWriter(writer)

	// Changelog generieren
	err = generator.Generate(writer, ctx.Query)
	if err != nil {
		logger.Error("Failed to generate changelog: %v", err)
		return err
	}

	return nil
}

// Die resolveConfigPath-Funktion wird nicht mehr benötigt, da sie bereits in helpers.go definiert ist
// resolveConfigPath löst den Pfad zur Konfigurationsdatei auf
// func resolveConfigPath(configPath string) string {
// 	// Wenn der Pfad absolut ist, direkt zurückgeben
// 	if filepath.IsAbs(configPath) {
// 		return configPath
// 	}
//
// 	// Relativen Pfad auflösen
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		return configPath
// 	}
//
// 	return filepath.Join(wd, configPath)
// }
