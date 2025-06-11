package ai

import (
	"fmt"
	"os"

	"clikd/internal/utils"

	"github.com/spf13/viper"
)

// Use the same logger instance as in client.go
var logConfig = utils.NewLogger("info", true)

// Provider represents an AI provider type
type Provider string

const (
	// ProviderMistral represents Mistral AI
	ProviderMistral Provider = "mistral"
	// ProviderOpenAI represents OpenAI
	ProviderOpenAI Provider = "openai"
	// ProviderLocal represents local models
	ProviderLocal Provider = "local"
	// ProviderAnthropic represents Anthropic
	ProviderAnthropic Provider = "anthropic"
	// ProviderGroq represents Groq
	ProviderGroq Provider = "groq"
	// ProviderOpenRouter represents OpenRouter
	ProviderOpenRouter Provider = "openrouter"
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
	Provider Provider `json:"provider" yaml:"provider"`
	Model    string   `json:"model" yaml:"model"`
	APIKey   string   `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	APIURL   string   `json:"api_url,omitempty" yaml:"api_url,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Provider: ProviderMistral,
		Model:    "mistral-medium",
		APIKey:   "",
		APIURL:   "",
	}
}

// LoadConfig loads the AI configuration from viper
func LoadConfig(v *viper.Viper) (*Config, error) {
	config := DefaultConfig()

	// Load AI configuration from viper
	if v.IsSet("ai") {
		// Load model configuration
		providerStr := v.GetString("ai.provider")
		if providerStr != "" {
			config.Provider = Provider(providerStr)
		}

		modelName := v.GetString("ai.model")
		if modelName != "" {
			config.Model = modelName
		}

		// Load other parameters
		if v.IsSet("ai.api_key") {
			config.APIKey = v.GetString("ai.api_key")
		}

		if v.IsSet("ai.api_url") {
			config.APIURL = v.GetString("ai.api_url")
		}
	}

	// Set API key from environment variables if not set
	if config.APIKey == "" {
		logConfig.Debug("API key not set in config, attempting to get from environment")
		apiKey, err := utils.GetAPIKey()
		if err != nil {
			// Log the error
			logConfig.Error("Failed to get API key: %v", err)
			// Provide provider-specific setup instructions
			return config, fmt.Errorf("API key not found for provider %s: %w\n\n%s",
				config.Provider, err, getAPIKeySetupInstructions(config.Provider))
		}
		config.APIKey = apiKey
		logConfig.Debug("API key successfully retrieved from environment")
	}

	// Set API URL from environment variables if not set
	if config.APIURL == "" {
		envVarName := getEndpointEnvVar(config.Provider)
		if envVarName != "" {
			if envValue := os.Getenv(envVarName); envValue != "" {
				logConfig.Debug("Using API URL from environment variable %s", envVarName)
				config.APIURL = envValue
			}
		}
	}

	logConfig.Debug("AI configuration loaded: provider=%s, model=%s",
		config.Provider, config.Model)
	return config, nil
}

// GetModelConfig converts the unified config to a ModelConfig
func (c *Config) GetModelConfig(modelName string) (ModelConfig, error) {
	// If no model name is provided, use the default
	if modelName == "" {
		modelName = c.Model
	}

	logConfig.Debug("Getting model config for %s with provider %s", modelName, c.Provider)

	// Create a ModelConfig based on the provider and model
	modelConfig := ModelConfig{
		Provider: c.Provider,
		ModelID:  modelName,
		APIKey:   c.APIKey,
		Endpoint: c.APIURL,
		// Verwende Standardwerte für die Parameter
		MaxTokens:      1024,
		Temperature:    0.7,
		TopP:           0.9,
		StreamResponse: false,
	}

	// Set context window based on provider and model
	switch c.Provider {
	case ProviderMistral:
		modelConfig.ContextWindow = 32000 // Maximum für Mistral
	case ProviderOpenAI:
		if modelName == "gpt-4o" {
			modelConfig.ContextWindow = 128000
		} else if modelName == "gpt-4" {
			modelConfig.ContextWindow = 8192
		} else {
			modelConfig.ContextWindow = 4096 // Für ältere Modelle wie gpt-3.5-turbo
		}
	case ProviderAnthropic:
		modelConfig.ContextWindow = 100000
	default:
		modelConfig.ContextWindow = 8192 // Standardwert für andere Provider
	}

	// If API key is not set, try to get it
	if modelConfig.APIKey == "" && c.Provider != ProviderLocal {
		logConfig.Error("API key not configured for %s", c.Provider)
		return modelConfig, fmt.Errorf("API-Schlüssel für %s nicht konfiguriert. %s",
			c.Provider, getAPIKeySetupInstructions(c.Provider))
	}

	return modelConfig, nil
}

// Helper functions

// getEndpointEnvVar returns the environment variable name for the endpoint
func getEndpointEnvVar(provider Provider) string {
	switch provider {
	case ProviderLocal:
		return "CLIKD_OLLAMA_BASE_URL"
	default:
		return "CLIKD_API_URL"
	}
}

// getProviderWebsite returns the website URL for the provider
func getProviderWebsite(provider Provider) string {
	switch provider {
	case ProviderMistral:
		return "https://console.mistral.ai/api-keys/"
	case ProviderOpenAI:
		return "https://platform.openai.com/account/api-keys"
	case ProviderAnthropic:
		return "https://console.anthropic.com/settings/keys"
	case ProviderGroq:
		return "https://console.groq.com/keys"
	case ProviderOpenRouter:
		return "https://openrouter.ai/keys"
	case ProviderLocal:
		return "https://github.com/ollama/ollama"
	default:
		return "https://clikd.dev/docs/ai-configuration"
	}
}

// getAPIKeySetupInstructions returns instructions for setting up the API key
func getAPIKeySetupInstructions(provider Provider) string {
	website := getProviderWebsite(provider)

	return fmt.Sprintf(`
You can add the API key in the following ways:

1. Create a .env file in the project directory and add:
   CLIKD_API_KEY=your_api_key

2. Or add it to your global configuration:
   clikd config set ai.api_key your_api_key

To obtain an API key for %s, visit: %s`,
		provider, website)
}

// IsAPIKeyConfigured checks if an API key is configured
func (c *Config) IsAPIKeyConfigured() bool {
	return c.APIKey != ""
}

// SetModel sets the model
func (c *Config) SetModel(modelName string) {
	c.Model = modelName
}

// SetProvider sets the provider
func (c *Config) SetProvider(provider Provider) {
	c.Provider = provider
}

// GetContext returns the context window size based on the provider and model
func (c *Config) GetContext() int {
	// Diese Methode ruft GetModelConfig auf und gibt den ContextWindow zurück
	modelConfig, err := c.GetModelConfig(c.Model)
	if err != nil {
		return 8192 // Default value if there's an error
	}
	return modelConfig.ContextWindow
}
