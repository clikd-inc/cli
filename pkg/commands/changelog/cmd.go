package changelog

import (
	"os"

	"clikd/pkg/internal/changelog"
	"clikd/pkg/internal/changelog/initializer"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"

	"github.com/tsuyoshiwada/go-gitcmd"
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
	cmd.Flags().StringVarP(&configFlag, "config", "c", ".chglog/config.yml", "specifies a different configuration file to pick up")
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

	wd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory: %v", err)
		return err
	}

	initCtx := &initializer.InitContext{
		WorkingDir: wd,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}

	fs := initializer.NewFileSystem()
	gitClient := gitcmd.New(&gitcmd.Config{Bin: "git"})
	questioner := initializer.NewQuestioner(gitClient, fs)
	configBuilder := initializer.NewConfigBuilder()
	templateBuilder := initializer.NewTemplateBuilderFactory()

	init := initializer.NewInitializer(initCtx, fs, questioner, configBuilder, templateBuilder)
	exitCode := init.Run()

	if exitCode != initializer.ExitCodeOK {
		return changelog.ErrNotSpecifiedCLIContext
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

	configPath := resolveConfigPath(configFlag)

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

	// Konfiguration laden
	loader := initializer.NewConfigLoader()
	config, err := loader.Load(ctx)
	if err != nil {
		logger.Error("Failed to load config: %v", err)
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
