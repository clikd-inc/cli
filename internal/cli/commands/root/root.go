package root

import (
	"fmt"
	"os"

	"clikd/internal/config"
	"clikd/internal/utils"

	"github.com/spf13/cobra"
)

var (
	// Used for flags
	cfgFile     string
	logLevel    string
	colorOutput bool
	appConfig   *config.ConfigData
	logger      utils.Logger
)

// NewRootCmd creates the root command for the CLI application
func NewRootCmd() *cobra.Command {
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

			// Override color output flag if provided
			if cmd.Flags().Changed("color") {
				// Note: Global color config removed - each service manages its own colors
				// This flag is kept for backward compatibility but not stored in config
				colorOutput = cmd.Flag("color").Value.String() == "true"
			}

			// Initialize logger with hardcoded color support
			logger = utils.NewLogger(appConfig.General.LogLevel, true)

			configPath, err := config.GetConfigFilePath()
			if err != nil {
				logger.Debug("Using default configuration")
			} else if configPath != "" {
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
	rootCmd.PersistentFlags().BoolVar(&colorOutput, "color", true, "Enable colorized output")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")

	return rootCmd
}

// Execute executes the root command
func Execute(version string) {
	rootCmd := NewRootCmd()
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("clikd version {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
