package root

import (
	"fmt"
	"os"

	"clikd/pkg/config"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	// Used for flags
	cfgFile   string
	logLevel  string
	appConfig *config.Config
	logger    *utils.Logger
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
			// Load configuration before executing any command
			var err error
			appConfig, err = config.LoadConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			// Override config with command line flags if provided
			if cmd.Flags().Changed("log-level") {
				appConfig.LogLevel = logLevel
			}

			// Initialize logger
			logger = utils.NewLogger(appConfig.LogLevel, true)
			logger.Debug("Configuration loaded from: %s", config.GetConfigFilePath())

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is provided, print help
			cmd.Help()
		},
	}

	// Add persistent flags that will be available to all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.clikd/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error, fatal)")

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
