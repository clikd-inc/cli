package config

import (
	"os"
)

// ModelConfig is maintained for API compatibility
type ModelConfig struct {
	// Provider is the AI provider (e.g., "openai", "mistral")
	Provider string `json:"provider" mapstructure:"provider"`
	// ModelID is the specific model ID from the provider (e.g., "gpt-4", "mistral-medium")
	ModelID string `json:"model_id" mapstructure:"model_id"`
	// APIKey is the API key for this provider (only in global configuration)
	APIKey string `json:"api_key,omitempty" mapstructure:"api_key,omitempty"`
	// Endpoint is an optional custom API endpoint
	Endpoint string `json:"endpoint,omitempty" mapstructure:"endpoint,omitempty"`
}

// ChangelogCommandConfig contains the configuration for the changelog command
// This structure replaces the old CLIContext of the initializer
type ChangelogCommandConfig struct {
	WorkingDir       string
	ConfigPath       string
	Template         string
	RepositoryURL    string
	OutputPath       string
	Silent           bool
	NoColor          bool
	NoEmoji          bool
	NoCaseSensitive  bool
	Query            string
	NextTag          string
	TagFilterPattern string
	JiraUsername     string
	JiraToken        string
	JiraURL          string
	Paths            []string
	Sort             string
}

// ChangelogConfig contains changelog-related settings
type ChangelogConfig struct {
	Template   string `mapstructure:"template"`
	ConfigFile string `mapstructure:"config_file"`
}

// ChangelogInfoConfig contains metadata for the changelog
type ChangelogInfoConfig struct {
	Title         string `mapstructure:"title"`
	RepositoryURL string `mapstructure:"repository_url"`
}

// GetModelConfig returns the configuration for a specific model
// This method is kept for backward compatibility but now uses the new config system
func GetModelConfig(modelName string) (ModelConfig, error) {
	// First read the provider from the environment variable
	provider := os.Getenv("CLIKD_AI_PROVIDER")
	if provider == "" {
		// Default value if not set
		provider = "mistral"
	}

	// Read the model from the environment variable if not explicitly specified
	if modelName == "" {
		modelName = os.Getenv("CLIKD_MODEL")
		if modelName == "" {
			// Default value if not set
			modelName = "mistral-medium"
		}
	}

	// Validate the requested model with the provider
	if err := ValidateProviderModel(provider, modelName); err != nil {
		// Return error for invalid combination
		return ModelConfig{}, err
	}

	// Read API URL from environment variable
	endpoint := os.Getenv("CLIKD_API_URL")

	// If we reach this point, the combination is valid
	model := ModelConfig{
		Provider: provider,
		ModelID:  modelName,
		APIKey:   "", // API key is retrieved via GetAPIKey
		Endpoint: endpoint,
	}

	return model, nil
}
