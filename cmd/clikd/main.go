package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"clikd/pkg/ai"
	"clikd/pkg/commands/changelog"
	"clikd/pkg/commands/hello"
	"clikd/pkg/commands/initialize"
	"clikd/pkg/commands/ui"
	"clikd/pkg/commands/version"
	"clikd/pkg/config"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is the version of the CLI
	Version = "0.1.0"

	// Used for flags
	cfgFile     string
	logLevel    string
	aiEnabled   bool
	colorOutput bool
	appConfig   *config.ConfigData
	logger      *utils.Logger
	configFile  string
	verboseFlag bool
	level       string
	colorize    bool
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "clikd",
		Short: "clikd - A powerful CLI tool",
		Long: `clikd is a flexible and powerful command line interface tool
that helps you accomplish various tasks efficiently.
Use it to automate workflows and enhance productivity.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize configuration before executing any command
			var err error
			if configFile != "" {
				if err := config.Initialize(configFile); err != nil {
					return err
				}
			} else {
				if err := config.Initialize(""); err != nil {
					return err
				}
			}

			// Load app configuration
			appConfig, err = config.EnsureInitialized()
			if err != nil {
				return fmt.Errorf("error loading configuration: %w", err)
			}

			// Override config with command line flags if provided
			if cmd.Flags().Changed("log-level") {
				if err := config.Set("general.log_level", logLevel); err != nil {
					return fmt.Errorf("error setting log level: %w", err)
				}
				appConfig.General.LogLevel = logLevel
			}

			// Override AI enabled flag if provided and set environment variable to mark flag as explicitly set
			if cmd.Flags().Changed("ai") {
				if err := config.Set("ai.enable", aiEnabled); err != nil {
					return fmt.Errorf("error setting AI enabled flag: %w", err)
				}
				appConfig.AI.Enable = aiEnabled

				// Umgebungsvariable setzen, damit Unterbefehle erkennen können, dass das Flag explizit gesetzt wurde
				os.Setenv("CLIKD_AI_FLAG_SET", "true")
			}

			// Override color output flag if provided
			if cmd.Flags().Changed("color") {
				if err := config.Set("general.color", colorOutput); err != nil {
					return fmt.Errorf("error setting color output flag: %w", err)
				}
				appConfig.General.Color = colorOutput
			}

			// Initialize logger
			logger = utils.NewLogger(appConfig.General.LogLevel, appConfig.General.Color)

			configPath, _ := config.GetConfigFilePath()
			if configPath != "" {
				logger.Debug("Configuration loaded from: %s", configPath)
			} else {
				logger.Debug("Using default configuration")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is provided, print help
			cmd.Help()
		},
	}

	// Add persistent flags that will be available to all commands
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Pfad zur Konfigurationsdatei")
	rootCmd.PersistentFlags().StringVarP(&level, "log-level", "l", "info", "Log-Level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "V", false, "Ausführliche Ausgabe aktivieren")
	rootCmd.PersistentFlags().BoolVar(&colorize, "no-color", true, "Farbige Ausgabe deaktivieren")
	rootCmd.PersistentFlags().BoolVar(&aiEnabled, "ai", false, "Enable AI-powered features globally for all commands")
	rootCmd.PersistentFlags().BoolVar(&colorOutput, "color", true, "Enable colorized output")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")

	// Add commands
	rootCmd.AddCommand(version.NewVersionCmd(Version))
	rootCmd.AddCommand(hello.NewHelloCmd())
	rootCmd.AddCommand(changelog.NewChangelogCmd())
	rootCmd.AddCommand(initialize.NewInitCmd())
	rootCmd.AddCommand(ui.NewUICmd())

	// AI-Test-Befehl hinzufügen
	aiTestCmd := &cobra.Command{
		Use:   "ai-test [prompt]",
		Short: "Test the AI integration with gollm",
		Long:  `Test the AI integration with the gollm library by sending a prompt to the configured model.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Alle Argumente als Prompt zusammenfügen
			prompt := strings.Join(args, " ")

			// AI-Konfiguration initialisieren
			_, err := config.EnsureInitialized()
			if err != nil {
				return fmt.Errorf("Fehler beim Initialisieren der Konfiguration: %w", err)
			}

			// Logger für Debugging-Ausgaben erstellen
			logger := utils.NewLogger(level, colorize)
			logger.Info("Starte AI-Test mit gollm...")

			// Modell aus Flag oder Standardmodell verwenden
			modelName, _ := cmd.Flags().GetString("model")

			// Client erstellen
			ctx := context.Background()

			// Konfiguration in Viper laden
			v := viper.New()
			v.SetConfigType("yaml")

			// AI-Konfiguration aus der globalen Konfiguration laden
			aiConfig, err := ai.LoadConfig(v)
			if err != nil {
				return fmt.Errorf("Fehler beim Laden der AI-Konfiguration: %w", err)
			}

			// Client erstellen
			client, err := ai.NewClient(ctx, aiConfig, modelName)
			if err != nil {
				return fmt.Errorf("Fehler beim Erstellen des AI-Clients: %w", err)
			}

			logger.Info("Verwende Anbieter: %s, Modell: %s", client.GetProvider(), client.GetModelName())

			// Prompt senden
			messages := []ai.Message{
				{
					Role:    "system",
					Content: "Du bist ein hilfreicher Assistent.",
				},
				{
					Role:    "user",
					Content: prompt,
				},
			}

			resp, err := client.Chat(ctx, messages)
			if err != nil {
				return fmt.Errorf("Fehler bei der KI-Antwort: %w", err)
			}

			// Antwort ausgeben
			fmt.Println("\n--- KI-Antwort ---")
			fmt.Println(resp.Text)
			fmt.Println("------------------")

			return nil
		},
	}

	// Modell-Flag für AI-Test-Befehl hinzufügen
	aiTestCmd.Flags().String("model", "", "Modell für die Anfrage verwenden")

	// AI-Test-Befehl zum Root-Befehl hinzufügen
	rootCmd.AddCommand(aiTestCmd)

	return rootCmd
}

func main() {
	rootCmd := newRootCmd()
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("clikd version {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
