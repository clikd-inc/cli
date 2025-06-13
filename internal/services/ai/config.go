package ai

import (
	"fmt"

	"clikd/internal/utils"
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
// This is the ONLY ModelConfig structure used by the AI service
type ModelConfig struct {
	Provider       Provider `json:"provider" toml:"provider"`
	ModelID        string   `json:"model_id" toml:"model_id"`
	APIKey         string   `json:"api_key,omitempty" toml:"api_key,omitempty"`
	Endpoint       string   `json:"endpoint,omitempty" toml:"endpoint,omitempty"`
	MaxTokens      int      `json:"max_tokens" toml:"max_tokens"`
	Temperature    float64  `json:"temperature" toml:"temperature"`
	TopP           float64  `json:"top_p" toml:"top_p"`
	ContextWindow  int      `json:"context_window" toml:"context_window"`
	StreamResponse bool     `json:"stream_response" toml:"stream_response"`
}

// CreateModelConfig creates a ModelConfig from global configuration values
// This replaces the old Config struct and GetModelConfig method
func CreateModelConfig(provider, model, apiKey, endpoint string, tokensMaxInput, tokensMaxOutput int) (ModelConfig, error) {
	if provider == "" || model == "" {
		return ModelConfig{}, fmt.Errorf("provider and model must not be empty")
	}

	// Convert string provider to Provider type
	providerType := Provider(provider)

	// Create a ModelConfig based on the provider and model
	modelConfig := ModelConfig{
		Provider: providerType,
		ModelID:  model,
		APIKey:   apiKey,
		Endpoint: endpoint,
		// Use token limits from global configuration
		MaxTokens:      tokensMaxInput, // Use input tokens as max tokens for generation
		Temperature:    0.7,
		TopP:           0.9,
		StreamResponse: false,
	}

	// Set context window based on provider and model
	switch providerType {
	case ProviderMistral:
		modelConfig.ContextWindow = 32000 // Maximum for Mistral
	case ProviderOpenAI:
		if model == "gpt-4o" {
			modelConfig.ContextWindow = 128000
		} else if model == "gpt-4" {
			modelConfig.ContextWindow = 8192
		} else {
			modelConfig.ContextWindow = 4096 // For older models like gpt-3.5-turbo
		}
	case ProviderAnthropic:
		modelConfig.ContextWindow = 100000
	default:
		modelConfig.ContextWindow = 8192 // Default value for other providers
	}

	// If API key is not set, try to get it from environment
	if modelConfig.APIKey == "" && providerType != ProviderLocal {
		logConfig.Debug("API key not provided, attempting to get from environment")
		apiKey, err := utils.GetAPIKey()
		if err != nil {
			logConfig.Error("Failed to get API key: %v", err)
			return modelConfig, fmt.Errorf("API key for %s not configured. %s",
				providerType, getAPIKeySetupInstructions(providerType))
		}
		modelConfig.APIKey = apiKey
		logConfig.Debug("API key successfully retrieved from environment")
	}

	return modelConfig, nil
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
