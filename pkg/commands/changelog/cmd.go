package changelog

import (
	"fmt"
	"os"
	"path/filepath"

	"clikd/pkg/config"
	"clikd/pkg/internal/changelog"
	"clikd/pkg/internal/changelog/initializer"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
)

// Variablen für Flags
var (
	initFlag             bool
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
  - AI Features: Enable AI-powered features with the global --ai flag or
    fine-tune behavior with specific AI flags.

Examples:
  # Generate changelog for all tags to stdout
  clikd changelog

  # Generate changelog to file
  clikd changelog -o CHANGELOG.md

  # Generate changelog for specific tag range
  clikd changelog v1.0.0..v2.0.0 -o CHANGELOG.md

  # Generate changelog with "unreleased" commits as next version
  clikd changelog --next-tag v2.0.0 -o CHANGELOG.md

  # Filter commits by path
  clikd changelog --path="pkg/,cmd/" -o CHANGELOG.md

  # Interactive initialization of config
  clikd changelog --init
  
  # Enable AI features globally
  clikd --ai changelog -o CHANGELOG.md
  
  # Override specific AI features when global AI is enabled
  clikd --ai changelog --ai-enhance-messages=false -o CHANGELOG.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Wenn --init Flag gesetzt ist, den Initializer ausführen
			if initFlag {
				return runInitializer()
			}

			// KI-Funktionalität initialisieren
			if err := InitializeAI(); err != nil {
				utils.NewLogger("error", true).Error("Failed to initialize AI: %v", err)
				// Wir gehen weiter, auch wenn die KI-Initialisierung fehlschlägt
			} else if changelog.IsAIEnabled() {
				// Zeige KI-Status, wenn KI aktiviert ist
				ShowAIStatus()
			}

			// Sonst den normalen Changelog-Generator ausführen
			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			return runGenerator(query)
		},
	}

	// Flags hinzufügen
	cmd.Flags().BoolVar(&initFlag, "init", false, "generate the git-chglog configuration file in interactive")
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

// runInitializer führt den interaktiven Initialisierungsprozess aus
func runInitializer() error {
	logger := utils.NewLogger("info", true)

	// Den `init` Befehl aufrufen, um die Konfiguration zu erstellen
	logger.Info("Initializing changelog configuration...")

	// Prüfen, ob bereits eine clikd/config.toml Datei existiert
	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory: %v", err)
		return err
	}

	// Konfiguration initialisieren
	initCmd := cobra.Command{}
	initCmd.SetOut(os.Stdout)
	initCmd.SetErr(os.Stderr)

	// `init` ausführen, aber nur wenn noch keine Konfiguration existiert
	clikdDir := filepath.Join(wd, "clikd")
	if _, err := os.Stat(clikdDir); os.IsNotExist(err) {
		logger.Info("Creating clikd configuration directory...")

		// Verzeichnisstruktur erstellen
		if err := os.MkdirAll(filepath.Join(clikdDir, "templates"), 0755); err != nil {
			logger.Error("Failed to create directories: %v", err)
			return err
		}

		// Standardkonfiguration erstellen
		manager := config.NewManager()
		manager.InitConfig("")

		// Standardwerte für Changelog anpassen
		manager.SetConfigValue("changelog.style", "github")
		manager.SetConfigValue("changelog.template", "templates/changelog.md")

		// Konfiguration speichern
		if err := manager.SaveConfig(filepath.Join(clikdDir, "config.toml")); err != nil {
			logger.Error("Failed to save configuration: %v", err)
			return err
		}

		// Changelog-Template erstellen
		templatePath := filepath.Join(clikdDir, "templates", "changelog.md")
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
			logger.Error("Failed to create changelog template: %v", err)
			return err
		}

		logger.Info("Changelog configuration initialized in %s", clikdDir)
	} else {
		logger.Info("Changelog configuration already exists in %s", clikdDir)
	}

	return nil
}

// runGenerator führt den Changelog-Generator aus
func runGenerator(query string) error {
	logger := utils.NewLogger("info", true)
	logger.Info("Generating changelog for query: %s", query)

	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory: %v", err)
		return err
	}

	logger.Debug("Working directory: %s", wd)

	// Konfigurationspfad auflösen
	configPath := resolveConfigPath(configFlag)
	logger.Debug("Resolved config path: %s", configPath)

	templatePath := ""

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

			// Handle both absolute and relative paths
			if filepath.IsAbs(tmplRelPath) {
				templatePath = tmplRelPath
			} else {
				// If relative path, it's relative to the clikd directory
				clikdDir := filepath.Dir(configPath)
				templatePath = filepath.Join(clikdDir, tmplRelPath)
			}

			logger.Debug("Using template from config: %s", templatePath)

			// Prüfen, ob die Template-Datei existiert
			if _, err := os.Stat(templatePath); err != nil {
				logger.Error("Template file not found at: %s", templatePath)
				return err
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
		WorkingDir:       wd,
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

	// Erweiterte Optionen aus der clikd-Konfiguration an den Loader übergeben
	// Hier müsste eigentlich eine direkte Übergabe der Konfiguration an den Loader erfolgen
	// Da dies eine größere Änderung wäre, beschränken wir uns vorerst auf die Anpassung
	// der Standard-Konfiguration und lassen den vorhandenen Mechanismus bestehen

	// In Zukunft könnte hier eine Funktion stehen, die die TOML-Konfiguration
	// direkt an den Loader übergibt, ohne dass dieser die Datei neu einlesen muss

	config, err := loader.Load(ctx)
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return err
	}

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
