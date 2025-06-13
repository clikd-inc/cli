package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"clikd/internal/cli/commands/changelog"
	"clikd/internal/cli/commands/initialize"
	"clikd/internal/cli/commands/version"
	cliversion "clikd/internal/cli/version"
	"clikd/internal/config"
	"clikd/internal/services"
	"clikd/internal/ui/bubble"
	"clikd/internal/utils"

	"github.com/spf13/cobra"
)

var (
	// Version is the version of the CLI - use the centralized version
	Version = cliversion.GetVersion()

	// Used for flags
	cfgFile     string
	aiEnabled   bool
	colorOutput bool
	appConfig   *config.Config
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
				if err := config.Set("general.log_level", level); err != nil {
					return fmt.Errorf("error setting log level: %w", err)
				}
				appConfig.General.LogLevel = level
			}

			// Set log level as environment variable so Bubble Tea UI can access it
			os.Setenv("CLIKD_LOG_LEVEL", appConfig.General.LogLevel)

			// AI is now always enabled, no need for flag override
			// Set environment variable so subcommands can detect that AI is enabled
			os.Setenv("CLIKD_AI_FLAG_SET", "true")

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
				logger.Debug("Configuration loaded", "path", configPath)
			} else {
				logger.Debug("Using default configuration")
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Skip update check for version and completion commands
			if cmd.Name() == "version" || cmd.Name() == "completion" {
				return nil
			}

			// Create a context with timeout for update check
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Create service factory for update service
			factory, err := services.NewServiceFactory(ctx)
			if err != nil {
				logger.Debug("Failed to create service factory for update check", "error", err)
				return nil // Silently ignore update check errors
			}

			// Create update service
			updateService := factory.CreateUpdateService()

			// Check for updates using the service
			hasUpdate, latestVersion, releaseURL, err := updateService.CheckForUpdates(ctx, Version)
			if err != nil {
				logger.Debug("Failed to check for updates", "error", err)
				return nil // Silently ignore update check errors
			}

			// If there's an update, print a message
			if hasUpdate {
				// Get terminal width (default to 80 if can't determine)
				width := 80

				// Render update notification
				notification := bubble.RenderUpdateNotification(Version, latestVersion, releaseURL, width)

				// Print a blank line for spacing and then the notification
				fmt.Println()
				fmt.Println(notification)
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
	rootCmd.PersistentFlags().BoolVar(&colorOutput, "color", true, "Enable colorized output")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number")

	// Add completion command
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script for your shell",
		Long: `To load completions:

Bash:
  $ source <(clikd completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ clikd completion bash > /etc/bash_completion.d/clikd
  # macOS:
  $ clikd completion bash > $(brew --prefix)/etc/bash_completion.d/clikd

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ clikd completion zsh > "${fpath[1]}/_clikd"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ clikd completion fish > ~/.config/fish/completions/clikd.fish

PowerShell:
  PS> clikd completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> clikd completion powershell > clikd.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	// Add commands
	rootCmd.AddCommand(completionCmd)
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

			// Create service factory
			ctx := context.Background()
			factory, err := services.NewServiceFactory(ctx)
			if err != nil {
				return fmt.Errorf("Error creating service factory: %w", err)
			}

			// Get logger from factory
			logger := factory.GetLogger()
			logger.Info("Starting AI test with gollm...")

			// Use model from flag or default model
			modelName, _ := cmd.Flags().GetString("model")
			if modelName != "" {
				// TODO: Support model override in factory
				logger.Warn("Model override not yet supported via factory, using configured model")
			}

			// Create AI service via factory
			aiService, err := factory.CreateAIService()
			if err != nil {
				return fmt.Errorf("Error creating AI service: %w", err)
			}

			// Test the service using batch processing (with single message)
			batchResult, err := aiService.EnhanceCommitMessagesBatch([]string{prompt})
			if err != nil {
				return fmt.Errorf("Error testing AI service: %w", err)
			}

			// Output result
			fmt.Println("AI Response:")
			if result, exists := batchResult[prompt]; exists {
				if len(result) > 1 {
					fmt.Println("Split into multiple messages:")
					for i, msg := range result {
						fmt.Printf("%d. %s\n", i+1, msg)
					}
				} else {
					fmt.Println(result[0])
				}
			} else {
				fmt.Println("No result returned")
			}

			return nil
		},
	}

	// Add flags to AI test command
	aiTestCmd.Flags().String("model", "", "Model to use (defaults to configured model)")

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
