package changelog

import (
	"clikd/internal/services/changelog"
	"clikd/internal/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

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

// getEnvBool liest einen boolean Wert aus einer Umgebungsvariable
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvString liest einen String-Wert aus einer Umgebungsvariable
func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

Environment Variables:
  NO_COLOR         - Disable color output (same as --no-color)
  NO_EMOJI         - Disable emoji output (same as --no-emoji)
  JIRA_URL         - Jira URL (same as --jira-url)
  JIRA_USERNAME    - Jira username (same as --jira-username)
  JIRA_TOKEN       - Jira token (same as --jira-token)

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

			// Umgebungsvariablen auswerten und Flags überschreiben
			noColorFlag = noColorFlag || getEnvBool("NO_COLOR", false)
			noEmojiFlag = noEmojiFlag || getEnvBool("NO_EMOJI", false)

			// Jira-Umgebungsvariablen nur setzen, wenn Flags nicht bereits gesetzt sind
			if jiraURLFlag == "" {
				jiraURLFlag = getEnvString("JIRA_URL", "")
			}
			if jiraUsernameFlag == "" {
				jiraUsernameFlag = getEnvString("JIRA_USERNAME", "")
			}
			if jiraTokenFlag == "" {
				jiraTokenFlag = getEnvString("JIRA_TOKEN", "")
			}

			// Query aus Argumenten extrahieren
			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			return runGenerator(query)
		},
	}

	// Flags hinzufügen
	cmd.Flags().StringSliceVar(&pathsFlag, "path", []string{}, "Filter commits by path(s). Can use multiple times.")
	cmd.Flags().StringVarP(&configFlag, "config", "c", "clikd/changelog/config.yml", "specifies a different configuration file to pick up")
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
	logger := utils.NewLogger("debug", !noColorFlag) // Verwende noColorFlag für Logger
	logger.Info("Generating changelog for query: %s", query)

	// Working Directory bestimmen
	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory: %v", err)
		return err
	}
	logger.Debug("Working directory: %s", wd)

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
		// Wenn es ein Verzeichnis ist, nehmen wir an, dass die Konfiguration in config.yml ist
		configPath = filepath.Join(configPath, "config.yml")
		logger.Debug("Config path is a directory, using %s instead", configPath)
	}

	// Template-Pfad wird jetzt direkt aus der YAML-Konfiguration gelesen
	// Die Template-Pfad-Logik ist jetzt in der config.go Normalize-Methode implementiert

	// Command-Konfiguration erstellen
	cmdConfig := &changelog.CommandConfig{
		WorkingDir:       wd,
		ConfigPath:       configPath,
		Template:         templateFlag, // Nur setzen wenn explizit angegeben
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
