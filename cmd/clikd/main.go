package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"clikd/internal/cli/commands/changelog"
	"clikd/internal/cli/commands/initialize"
	"clikd/internal/cli/commands/version"
	"clikd/internal/config"
	"clikd/internal/services/ai"
	"clikd/internal/utils"

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
	logger      utils.Logger
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

				// Set environment variable so subcommands can detect that the flag was explicitly set
				os.Setenv("CLIKD_AI_FLAG_SET", "true")
			}

			// Override color output flag if provided
			if cmd.Flags().Changed("color") {
				// Note: Global color config removed - each service manages its own colors
				// This flag is kept for backward compatibility but not stored in config
				colorOutput = cmd.Flag("color").Value.String() == "true"
			}

			// Initialize logger with hardcoded color support
			logger = utils.NewLogger(appConfig.General.LogLevel, true)

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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	rootCmd.PersistentFlags().StringVarP(&level, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "V", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&colorize, "no-color", true, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&aiEnabled, "ai", false, "Enable AI-powered features globally for all commands")
	rootCmd.PersistentFlags().BoolVar(&colorOutput, "color", true, "Enable colorized output")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")

	// Add commands
	rootCmd.AddCommand(version.NewVersionCmd(Version))
	rootCmd.AddCommand(changelog.NewChangelogCmd())
	rootCmd.AddCommand(initialize.NewInitCmd())

	// Add AI test command
	aiTestCmd := &cobra.Command{
		Use:   "ai-test [prompt]",
		Short: "Test the AI integration with gollm",
		Long:  `Test the AI integration with the gollm library by sending a prompt to the configured model.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Join all arguments as prompt
			prompt := strings.Join(args, " ")

			// Initialize AI configuration
			_, err := config.EnsureInitialized()
			if err != nil {
				return fmt.Errorf("Error initializing configuration: %w", err)
			}

			// Create logger for debugging output
			logger := utils.NewLogger(level, colorize)
			logger.Info("Starting AI test with gollm...")

			// Use model from flag or default model
			modelName, _ := cmd.Flags().GetString("model")

			// Create client
			ctx := context.Background()

			// Load configuration in Viper
			v := viper.New()
			v.SetConfigType("yaml")

			// Load AI configuration from global configuration
			modelConfig, err := config.GetAIModelConfig(modelName)
			if err != nil {
				return fmt.Errorf("Error loading AI configuration: %w", err)
			}

			// Convert ModelConfig to ai.Config
			aiConfig := &ai.Config{
				Provider: ai.Provider(modelConfig.Provider),
				Model:    modelConfig.ModelID,
				APIKey:   modelConfig.APIKey,
				APIURL:   modelConfig.Endpoint,
				EnableAI: true,
			}

			// Create client
			client, err := ai.NewClient(ctx, aiConfig, modelName)
			if err != nil {
				return fmt.Errorf("Error creating AI client: %w", err)
			}

			logger.Info("Using provider: %s, model: %s", client.GetProvider(), client.GetModelName())

			// Send prompt
			messages := []ai.Message{
				{
					Role:    "system",
					Content: "You are a helpful assistant.",
				},
				{
					Role:    "user",
					Content: prompt,
				},
			}

			resp, err := client.Chat(ctx, messages)
			if err != nil {
				return fmt.Errorf("Error in AI response: %w", err)
			}

			// Output response
			fmt.Println("\n--- AI Response ---")
			fmt.Println(resp.Text)
			fmt.Println("------------------")

			return nil
		},
	}

	// Add model flag for AI test command
	aiTestCmd.Flags().String("model", "", "Use model for the request")

	// Add AI test command to root command
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
