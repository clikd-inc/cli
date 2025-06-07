package changelog

import (
	"os"

	"clikd/pkg/config"
	"clikd/pkg/internal/changelog"
	"clikd/pkg/utils"

	"github.com/spf13/cobra"
)

// AI-related Flags
var (
	aiEnableFlag             bool
	aiModelFlag              string
	aiEnhanceMessagesFlag    bool
	aiGenerateSummariesFlag  bool
	aiCategorizeCommitsFlag  bool
	aiSuggestVersionBumpFlag bool
)

// AddAIFlags adds the AI-related flags to a command
func AddAIFlags(cmd *cobra.Command) {
	// This flag remains for command-specific overrides
	cmd.Flags().BoolVar(&aiEnableFlag, "ai", false, "Override global AI setting for this command only")

	// Model selection
	cmd.Flags().StringVar(&aiModelFlag, "ai-model", "", "Specify AI model to use (overrides config default)")

	// Detailed control of AI functions
	cmd.Flags().BoolVar(&aiEnhanceMessagesFlag, "ai-enhance-messages", true, "Use AI to enhance commit messages (requires AI enabled)")
	cmd.Flags().BoolVar(&aiGenerateSummariesFlag, "ai-generate-summaries", true, "Use AI to generate summaries for changes (requires AI enabled)")
	cmd.Flags().BoolVar(&aiCategorizeCommitsFlag, "ai-categorize-commits", true, "Use AI to categorize commit messages (requires AI enabled)")
	cmd.Flags().BoolVar(&aiSuggestVersionBumpFlag, "ai-suggest-version", true, "Use AI to suggest version bump type (requires AI enabled)")

	// Mark flags as hidden to keep the CLI interface clean
	// They can still be used but don't appear in the basic help
	cmd.Flags().MarkHidden("ai-model")
	cmd.Flags().MarkHidden("ai-enhance-messages")
	cmd.Flags().MarkHidden("ai-generate-summaries")
	cmd.Flags().MarkHidden("ai-categorize-commits")
	cmd.Flags().MarkHidden("ai-suggest-version")
}

// InitializeAI initializes the AI functionality based on the flags
func InitializeAI() error {
	logger := utils.NewLogger("info", true)

	// Get global configuration
	cfg, err := config.EnsureInitialized()
	if err != nil {
		return err
	}

	// Show configuration path
	configPath, _ := config.GetConfigFilePath()
	if configPath != "" {
		logger.Info("Configuration loaded from: %s", configPath)
	} else {
		logger.Info("Using default configuration (no configuration file found)")
	}

	// Debug output of the configuration
	logger.Info("Loaded AI configuration: enable=%v, provider=%s, model=%s",
		cfg.AI.Enable, cfg.AI.Provider, cfg.AI.Model)

	// Check for global AI activation from configuration or environment variable
	globalAIEnabled := cfg.AI.Enable
	if envEnabled := os.Getenv("CLIKD_AI_ENABLE"); envEnabled != "" {
		if envEnabled == "true" {
			globalAIEnabled = true
			logger.Info("AI overridden by environment variable CLIKD_AI_ENABLE: %v", globalAIEnabled)
		} else if envEnabled == "false" {
			globalAIEnabled = false
			logger.Info("AI overridden by environment variable CLIKD_AI_ENABLE: %v", globalAIEnabled)
		}
	}

	// Flag status for the current execution
	// IMPORTANT: We use the configuration as the default setting here and
	// only override it if the flag was explicitly set
	currentAIEnabled := globalAIEnabled

	// Determine if AI is enabled (priority: Flag > Env > Config)
	flagExplicitlySet := false

	// This check requires access to the flag via cobra
	// Since we don't have direct access, we use the variable values
	// from aiEnableFlag, which is set by the flag
	// If aiEnableFlag != globalAIEnabled, then the flag was explicitly set
	// We could also implement command-level tracking
	if os.Getenv("CLIKD_AI_FLAG_SET") == "true" {
		// This environment variable can be set by the main application
		// when it detects that the flag was explicitly set
		flagExplicitlySet = true
		logger.Debug("AI flag was explicitly set via command line")
	}

	// Based on the determined values, now enable or disable AI
	if flagExplicitlySet {
		// Flag has highest priority
		currentAIEnabled = aiEnableFlag
		logger.Info("AI functionality was %s: Flag=%v, Configuration=%v",
			humanizeEnabled(currentAIEnabled), aiEnableFlag, globalAIEnabled)
	} else {
		// Configuration or environment variable determines the status
		logger.Info("AI functionality according to configuration: %v", currentAIEnabled)
	}

	// Only perform initialization if AI is enabled
	if !currentAIEnabled {
		logger.Info("AI functionality is disabled")
		return nil
	}

	logger.Info("Initializing AI subsystem with model: %s", cfg.AI.Model)

	// Create AI configuration
	modelConfig, err := config.GetAIModelConfig(aiModelFlag)
	if err != nil {
		return err
	}

	// Create AI options
	aiOpts := changelog.AIOptions{
		EnableAI:              currentAIEnabled,
		ModelName:             modelConfig.ModelID,
		EnhanceCommitMessages: aiEnhanceMessagesFlag,
		GenerateSummaries:     aiGenerateSummariesFlag,
		CategorizeCommits:     aiCategorizeCommitsFlag,
		SuggestVersionBump:    aiSuggestVersionBumpFlag,
	}

	// Initialize AI subsystem
	if err := changelog.InitAI(modelConfig, aiOpts); err != nil {
		return err
	}

	// Successful initialization
	ShowAIStatus()

	return nil
}

// humanizeEnabled returns a human-readable representation of the enable status
func humanizeEnabled(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

// ShowAIStatus displays the status of the AI functionality
func ShowAIStatus() {
	logger := utils.NewLogger("info", true)

	if changelog.IsAIEnabled() {
		logger.Info("AI functionality is enabled")

		// Show source of activation
		cfg, err := config.Get()
		if err == nil {
			// Check if enabled via flag, environment variable, or configuration
			if aiEnableFlag && !cfg.AI.Enable {
				logger.Info("AI enabled via command flag --ai (overrides configuration)")
			} else if cfg.AI.Enable {
				logger.Info("AI enabled via global configuration (ai.enable=true in config.toml)")
			} else if os.Getenv("CLIKD_AI_ENABLE") == "true" {
				logger.Info("AI enabled via environment variable CLIKD_AI_ENABLE")
			}
		}

		// Show model information
		modelName := aiModelFlag
		if modelName == "" {
			modelName = os.Getenv("CLIKD_MODEL")
			if modelName == "" {
				cfg, err := config.Get()
				if err == nil && cfg.AI.Model != "" {
					modelName = cfg.AI.Model
				} else {
					modelName = "mistral-medium" // Default model
				}
			}
		}

		logger.Info("Selected AI model: %s", modelName)

		// Show provider
		provider := os.Getenv("CLIKD_AI_PROVIDER")
		if provider == "" {
			cfg, err := config.Get()
			if err == nil && cfg.AI.Provider != "" {
				provider = cfg.AI.Provider
			} else {
				provider = "mistral" // Default provider
			}
		}
		logger.Info("Provider: %s", provider)

		// Show enabled functions
		logger.Info("Enabled AI functions:")
		if aiEnhanceMessagesFlag {
			logger.Info("- Enhance commit messages")
		}
		if aiGenerateSummariesFlag {
			logger.Info("- Generate summaries")
		}
		if aiCategorizeCommitsFlag {
			logger.Info("- Categorize commits")
		}
		if aiSuggestVersionBumpFlag {
			logger.Info("- Suggest version bump")
		}

		// Note on API key source
		if utils.IsLocalConfigPresent() {
			logger.Info("API key is loaded from the local .env file")
		} else {
			logger.Info("API key is loaded from the global configuration")
		}
	} else {
		logger.Info("AI functionality is disabled")
		logger.Info("To enable AI functions, set ai.enable=true in the configuration or use the --ai flag or set CLIKD_AI_ENABLE=true")
	}
}
