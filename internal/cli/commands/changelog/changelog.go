package changelog

import (
	"clikd/internal/config"
	"clikd/internal/services/changelog"
	"clikd/internal/utils"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Variablen für Flags
var (
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

Examples:
  # Initialize changelog configuration with the global init command
  clikd init

  # Generate changelog for all tags to stdout
  clikd changelog

  # Generate changelog to file
  clikd changelog -o CHANGELOG.md

  # Generate changelog for specific tag range
  clikd changelog v1.0.0..v2.0.0 -o CHANGELOG.md

  # Generate changelog with "unreleased" commits as next version
  clikd changelog --next-tag v2.0.0 -o CHANGELOG.md

  # Filter commits by path
  clikd changelog --path="internal/,cmd/" -o CHANGELOG.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := utils.NewLogger("info", true)
			logger.Info("Starting changelog generation")

			// Sonst den normalen Changelog-Generator ausführen
			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			return runGenerator(query)
		},
	}

	// Flags hinzufügen
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

	return cmd
}

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
	configPath := utils.ResolveConfigPath(configFlag)
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

	// Immer "clikd" als Basisverzeichnis verwenden
	baseDir := "clikd"

	// Prüfen, ob ein Template-Pfad mit dem korrekten Verzeichnis funktioniert
	directTemplatePath := filepath.Join(wd, baseDir, "templates", "changelog.md")
	if _, err := os.Stat(directTemplatePath); err == nil {
		templatePath = directTemplatePath
	} else {
		logger.Debug("Template path not found: %s", directTemplatePath)
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
							filepath.Join(wd, tmplRelPath),
							filepath.Join(wd, baseDir, tmplRelPath),
							filepath.Join(wd, baseDir, "templates", filepath.Base(tmplRelPath)),
						}

						for _, alt := range alternatives {
							logger.Debug("Trying alternative template path: %s", alt)
							if _, err := os.Stat(alt); err == nil {
								logger.Debug("Template file found at alternative path: %s", alt)
								templatePath = alt
								break
							}
						}

						if templatePath == "" {
							logger.Error("Template file not found at any expected location")
							return fmt.Errorf("template file not found at any expected location")
						}
					}
				}
			}
		}
	} else {
		logger.Debug("Configuration file not found: %s", configPath)
		if templatePath == "" {
			logger.Error("No template path found")
			return fmt.Errorf("no template path found")
		}
	}

	logger.Info("Using template: %s", templatePath)

	// Command-Konfiguration erstellen
	cmdConfig := &changelog.CommandConfig{
		WorkingDir:       testRepoDir, // Für Tests explizit ein Test-Repository festlegen
		ConfigPath:       configPath,
		Template:         templatePath,
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

	// Konfiguration laden
	config, err := changelog.LoadConfigFromCommand(cmdConfig)
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return err
	}

	// Generator erstellen
	generator := changelog.NewGenerator(logger, config)

	// Output-Writer bestimmen
	var writer io.Writer
	if cmdConfig.OutputPath == "" {
		// Wenn kein Output-Pfad angegeben ist, wird auf stdout ausgegeben
		writer = os.Stdout
	} else {
		// Verzeichnis erstellen, falls es nicht existiert
		dir := filepath.Dir(cmdConfig.OutputPath)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		}

		// Datei öffnen
		file, err := os.Create(cmdConfig.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()
		writer = file
	}

	// Changelog generieren
	return generator.Generate(writer, query)
}
