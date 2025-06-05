package changelog

import (
	"context"
	"fmt"
	"os"
	"sync"

	"clikd/pkg/ai"
	"clikd/pkg/config"
)

var (
	aiServiceSingleton     *ai.ChangelogService
	aiServiceSingletonLock sync.Mutex
	aiEnabled              = false
	aiConfig               *ai.Config
)

// AIOptions represents options for AI-enhanced changelog functionality
type AIOptions struct {
	// EnableAI enables or disables AI-enhanced features
	EnableAI bool
	// ModelName specifies which AI model to use (if blank, uses default from config)
	ModelName string
	// Features to enable
	EnhanceCommitMessages bool
	GenerateSummaries     bool
	CategorizeCommits     bool
	SuggestVersionBump    bool
}

// InitAI initializes the AI subsystem for changelog generation
func InitAI(modelConfig config.ModelConfig, opts AIOptions) error {
	aiServiceSingletonLock.Lock()
	defer aiServiceSingletonLock.Unlock()

	// Don't initialize if AI is disabled
	if !opts.EnableAI {
		aiEnabled = false
		return nil
	}

	// Create AI configuration based on our new config structure
	aiCfg := &ai.Config{
		DefaultProvider: ai.Provider(modelConfig.Provider),
		DefaultModel:    opts.ModelName,
		EnableAI:        opts.EnableAI,
		Verbose:         false,
		Models: map[string]ai.ModelConfig{
			opts.ModelName: {
				Provider:       ai.Provider(modelConfig.Provider),
				ModelID:        modelConfig.ModelID,
				APIKey:         modelConfig.APIKey,
				Endpoint:       modelConfig.Endpoint,
				MaxTokens:      modelConfig.MaxTokens,
				Temperature:    modelConfig.Temperature,
				TopP:           modelConfig.TopP,
				ContextWindow:  modelConfig.ContextWindow,
				StreamResponse: modelConfig.StreamResponse,
			},
		},
	}

	// Store configuration
	aiConfig = aiCfg
	aiEnabled = opts.EnableAI

	return nil
}

// getAIService returns the singleton instance of the AI service
func getAIService(modelName string) (*ai.ChangelogService, error) {
	aiServiceSingletonLock.Lock()
	defer aiServiceSingletonLock.Unlock()

	if !aiEnabled || aiConfig == nil {
		return nil, fmt.Errorf("AI functionality is not enabled or configured")
	}

	if aiServiceSingleton != nil {
		return aiServiceSingleton, nil
	}

	// Create a new service
	service, err := ai.NewChangelogService(context.Background(), aiConfig, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI service: %w", err)
	}

	aiServiceSingleton = service
	return service, nil
}

// IsAIEnabled returns whether AI functionality is enabled
func IsAIEnabled() bool {
	return aiEnabled
}

// EnhanceCommitMessage improves a commit message using AI if enabled
func EnhanceCommitMessage(message string, modelName string) (string, error) {
	if !aiEnabled {
		return message, nil
	}

	service, err := getAIService(modelName)
	if err != nil {
		return message, err
	}

	return service.EnhanceCommitMessage(context.Background(), message)
}

// CategorizeCommit uses AI to categorize a commit message if enabled
func CategorizeCommit(message string, modelName string) (string, error) {
	if !aiEnabled {
		return "other", nil
	}

	service, err := getAIService(modelName)
	if err != nil {
		return "other", err
	}

	category, err := service.CategorizeCommit(context.Background(), message)
	if err != nil {
		return "other", err
	}

	return string(category), nil
}

// ExtractCommitInfo uses AI to extract structured information from a commit message if enabled
func ExtractCommitInfo(message, author, date, hash string, modelName string) (*ai.CommitInfo, error) {
	if !aiEnabled {
		return &ai.CommitInfo{
			Message:   message,
			Author:    author,
			Date:      date,
			Hash:      hash,
			Category:  ai.CommitCategoryOther,
			Summary:   message,
			IssueRefs: []string{},
		}, nil
	}

	service, err := getAIService(modelName)
	if err != nil {
		return nil, err
	}

	return service.ExtractCommitInfo(context.Background(), message, author, date, hash)
}

// GenerateChangeSummary generates a summary for a set of commits using AI if enabled
func GenerateChangeSummary(commits []ai.CommitInfo, modelName string) (string, error) {
	if !aiEnabled || len(commits) == 0 {
		return "No summary available.", nil
	}

	service, err := getAIService(modelName)
	if err != nil {
		return "No summary available.", err
	}

	return service.GenerateSummary(context.Background(), commits)
}

// SuggestVersionBump suggests whether a version should be bumped as major, minor, or patch
func SuggestVersionBump(commits []ai.CommitInfo, currentVersion, modelName string) (string, string, error) {
	if !aiEnabled || len(commits) == 0 {
		return "patch", "No significant changes detected.", nil
	}

	service, err := getAIService(modelName)
	if err != nil {
		return "patch", "Failed to get AI service.", err
	}

	return service.SuggestVersionBump(context.Background(), commits, currentVersion)
}

// LoadAIFromEnv initializes AI configuration from environment variables
func LoadAIFromEnv() {
	// Check if AI functionality is enabled via environment
	aiEnv := os.Getenv("CLIKD_CHANGELOG_AI_ENABLED")
	if aiEnv == "true" || aiEnv == "1" || aiEnv == "yes" {
		// Get our config
		cfg, err := config.EnsureInitialized()
		if err != nil {
			return
		}

		// Find a suitable model
		var modelConfig config.ModelConfig
		if cfg.AI.Models != nil && cfg.AI.DefaultModel != "" {
			if model, ok := cfg.AI.Models[cfg.AI.DefaultModel]; ok {
				modelConfig = model
			}
		}

		// Initialize AI with basic options
		InitAI(modelConfig, AIOptions{
			EnableAI:              true,
			EnhanceCommitMessages: true,
			GenerateSummaries:     true,
			CategorizeCommits:     true,
			SuggestVersionBump:    true,
		})
	}
}
