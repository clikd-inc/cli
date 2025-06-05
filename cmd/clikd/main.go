package main

import (
	"fmt"
	"os"

	"clikd/pkg/commands/changelog"
	"clikd/pkg/commands/hello"
	"clikd/pkg/commands/initialize"
	"clikd/pkg/commands/version"
	"clikd/pkg/config"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
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
			if err := config.Initialize(cfgFile); err != nil {
				return fmt.Errorf("error initializing config: %w", err)
			}

			var err error
			appConfig, err = config.Get()
			if err != nil {
				return fmt.Errorf("error retrieving config: %w", err)
			}

			// Override config with command line flags if provided
			if cmd.Flags().Changed("log-level") {
				if err := config.Set("general.log_level", logLevel); err != nil {
					return fmt.Errorf("error setting log level: %w", err)
				}
				appConfig.General.LogLevel = logLevel
			}

			// Override AI enabled flag if provided
			if cmd.Flags().Changed("ai") {
				if err := config.Set("ai.enable", aiEnabled); err != nil {
					return fmt.Errorf("error setting AI enabled flag: %w", err)
				}
				appConfig.AI.Enable = aiEnabled
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.clikd/config.toml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().BoolVar(&aiEnabled, "ai", false, "Enable AI-powered features globally for all commands")
	rootCmd.PersistentFlags().BoolVar(&colorOutput, "color", true, "Enable colorized output")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")

	// Add commands
	rootCmd.AddCommand(version.NewVersionCmd(Version))
	rootCmd.AddCommand(hello.NewHelloCmd())
	rootCmd.AddCommand(changelog.NewChangelogCmd())
	rootCmd.AddCommand(initialize.NewInitCmd())

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
