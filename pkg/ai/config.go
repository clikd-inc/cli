package ai

import (
	"fmt"
	"os"

	"clikd/pkg/utils"

	"github.com/spf13/viper"
)

// Provider represents an AI provider type
type Provider string

const (
	// ProviderMistral represents Mistral AI
	ProviderMistral Provider = "mistral"
	// ProviderOpenAI represents OpenAI
	ProviderOpenAI Provider = "openai"
	// ProviderAzureOpenAI represents Azure OpenAI
	ProviderAzureOpenAI Provider = "azure-openai"
	// ProviderLocal represents local models
	ProviderLocal Provider = "local"
)

// ModelConfig represents configuration for a specific AI model
type ModelConfig struct {
	Provider       Provider `json:"provider" yaml:"provider"`
	ModelID        string   `json:"model_id" yaml:"model_id"`
	APIKey         string   `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	Endpoint       string   `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	MaxTokens      int      `json:"max_tokens" yaml:"max_tokens"`
	Temperature    float64  `json:"temperature" yaml:"temperature"`
	TopP           float64  `json:"top_p" yaml:"top_p"`
	ContextWindow  int      `json:"context_window" yaml:"context_window"`
	StreamResponse bool     `json:"stream_response" yaml:"stream_response"`
}

// Config represents the central AI configuration
type Config struct {
	DefaultProvider Provider               `json:"default_provider" yaml:"default_provider"`
	DefaultModel    string                 `json:"default_model" yaml:"default_model"`
	Models          map[string]ModelConfig `json:"models" yaml:"models"`
	EnableAI        bool                   `json:"enable_ai" yaml:"enable_ai"`
	Verbose         bool                   `json:"verbose" yaml:"verbose"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: ProviderMistral,
		DefaultModel:    "mistral-medium",
		EnableAI:        true,
		Verbose:         false,
		Models: map[string]ModelConfig{
			"mistral-medium": {
				Provider:       ProviderMistral,
				ModelID:        "mistral-medium",
				MaxTokens:      1024,
				Temperature:    0.7,
				TopP:           0.9,
				ContextWindow:  8192,
				StreamResponse: false,
			},
			"mistral-small": {
				Provider:       ProviderMistral,
				ModelID:        "mistral-small",
				MaxTokens:      1024,
				Temperature:    0.7,
				TopP:           0.9,
				ContextWindow:  8192,
				StreamResponse: false,
			},
			"gpt-3.5-turbo": {
				Provider:       ProviderOpenAI,
				ModelID:        "gpt-3.5-turbo",
				MaxTokens:      1024,
				Temperature:    0.7,
				TopP:           0.9,
				ContextWindow:  4096,
				StreamResponse: false,
			},
			"gpt-4": {
				Provider:       ProviderOpenAI,
				ModelID:        "gpt-4",
				MaxTokens:      1024,
				Temperature:    0.7,
				TopP:           0.9,
				ContextWindow:  8192,
				StreamResponse: false,
			},
		},
	}
}

// LoadConfig loads AI configuration from viper
func LoadConfig(v *viper.Viper) (*Config, error) {
	config := DefaultConfig()

	// First check if we have ai.* keys directly set
	// This typically happens in tests or when directly setting config values
	if v.IsSet("ai.default_model") {
		config.DefaultModel = v.GetString("ai.default_model")
	}

	if v.IsSet("ai.enable_ai") {
		config.EnableAI = v.GetBool("ai.enable_ai")
	}

	// Then check if we have an "ai" structure (typical when loading from config file)
	if v.IsSet("ai") {
		// Try to unmarshal the configuration
		if err := v.UnmarshalKey("ai", config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AI config: %w", err)
		}
	}

	// Check if we have a local configuration
	localConfigExists := utils.IsLocalConfigPresent()

	// Load API keys for each model using our new utility
	for name, model := range config.Models {
		// Create a provider info structure for the API key utility
		var providerInfo utils.ProviderKeyInfo

		switch model.Provider {
		case ProviderMistral:
			providerInfo = utils.MistralProvider
		case ProviderOpenAI:
			providerInfo = utils.OpenAIProvider
		case ProviderAzureOpenAI:
			// If not already defined in utils, we would add this
			providerInfo = utils.ProviderKeyInfo{
				Name:            "Azure OpenAI",
				ConfigKey:       "ai.models.azure-openai.api_key",
				EnvVarName:      "CLIKD_AZURE_OPENAI_API_KEY",
				EnvVarNameShort: "AZURE_OPENAI_API_KEY",
				Required:        false,
			}
		default:
			// Skip for providers that don't need API keys
			continue
		}

		// Set as not required for initial loading, will be checked when used
		providerInfo.Required = false

		// Get the API key following our hierarchy
		apiKey, _ := utils.GetAPIKey(providerInfo, localConfigExists)
		if apiKey != "" {
			model.APIKey = apiKey
			config.Models[name] = model
		}

		// Load endpoints from environment variables if applicable
		if model.Endpoint == "" && requiresEndpoint(model.Provider) {
			envEndpoint := getEndpointEnvVar(model.Provider)
			if envEndpoint != "" && os.Getenv(envEndpoint) != "" {
				model.Endpoint = os.Getenv(envEndpoint)
				config.Models[name] = model
			}
		}
	}

	return config, nil
}

// GetModelConfig returns the configuration for a specific model
func (c *Config) GetModelConfig(modelName string) (ModelConfig, error) {
	model, exists := c.Models[modelName]
	if !exists {
		// If the requested model doesn't exist, fall back to the default
		model, exists = c.Models[c.DefaultModel]
		if !exists {
			return ModelConfig{}, fmt.Errorf("model %s not found and default model %s is not configured",
				modelName, c.DefaultModel)
		}
		modelName = c.DefaultModel
	}

	// Check if the API key is set, if not try to get it
	if model.APIKey == "" && model.Provider != ProviderLocal {
		// Check if we have a local configuration
		localConfigExists := utils.IsLocalConfigPresent()

		// Create a provider info structure for the API key utility
		var providerInfo utils.ProviderKeyInfo

		switch model.Provider {
		case ProviderMistral:
			providerInfo = utils.MistralProvider
		case ProviderOpenAI:
			providerInfo = utils.OpenAIProvider
		case ProviderAzureOpenAI:
			providerInfo = utils.ProviderKeyInfo{
				Name:            "Azure OpenAI",
				ConfigKey:       "ai.models.azure-openai.api_key",
				EnvVarName:      "CLIKD_AZURE_OPENAI_API_KEY",
				EnvVarNameShort: "AZURE_OPENAI_API_KEY",
				Required:        true,
			}
		default:
			// Skip for providers that don't need API keys
			return model, nil
		}

		// When getting a specific model, the key is required
		providerInfo.Required = true

		// Get the API key following our hierarchy
		apiKey, err := utils.GetAPIKey(providerInfo, localConfigExists)
		if err != nil {
			return model, fmt.Errorf("failed to get API key for %s model: %w", modelName, err)
		}

		if apiKey != "" {
			model.APIKey = apiKey
			// Update the model in the config for future use
			c.Models[modelName] = model
		}
	}

	return model, nil
}

// getAPIKeyEnvVar returns the environment variable name for the API key
func getAPIKeyEnvVar(provider Provider) string {
	switch provider {
	case ProviderMistral:
		return "MISTRAL_API_KEY"
	case ProviderOpenAI:
		return "OPENAI_API_KEY"
	case ProviderAzureOpenAI:
		return "AZURE_OPENAI_API_KEY"
	case ProviderLocal:
		return ""
	default:
		return ""
	}
}

// getEndpointEnvVar returns the environment variable name for the endpoint
func getEndpointEnvVar(provider Provider) string {
	switch provider {
	case ProviderAzureOpenAI:
		return "AZURE_OPENAI_ENDPOINT"
	case ProviderLocal:
		return "LOCAL_AI_ENDPOINT"
	default:
		return ""
	}
}

// requiresEndpoint returns true if the provider requires an endpoint
func requiresEndpoint(provider Provider) bool {
	return provider == ProviderAzureOpenAI || provider == ProviderLocal
}

// IsAPIKeyConfigured checks if the API key is configured for a model
func (c *Config) IsAPIKeyConfigured(modelName string) bool {
	model, err := c.GetModelConfig(modelName)
	if err != nil {
		return false
	}

	return model.APIKey != ""
}

// AddModel adds or updates a model configuration
func (c *Config) AddModel(name string, config ModelConfig) {
	if c.Models == nil {
		c.Models = make(map[string]ModelConfig)
	}
	c.Models[name] = config
}

// GetAvailableModels returns a list of configured model names
func (c *Config) GetAvailableModels() []string {
	models := make([]string, 0, len(c.Models))
	for name := range c.Models {
		models = append(models, name)
	}
	return models
}

// SetDefaultModel sets the default model
func (c *Config) SetDefaultModel(modelName string) error {
	if _, exists := c.Models[modelName]; !exists {
		return fmt.Errorf("model %s not found in configuration", modelName)
	}
	c.DefaultModel = modelName
	c.DefaultProvider = c.Models[modelName].Provider
	return nil
}
