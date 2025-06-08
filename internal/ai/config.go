package ai

import (
	"fmt"
	"os"

	"clikd/internal/utils"

	"github.com/spf13/viper"
)

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
	EnableAI bool     `json:"enable_ai" yaml:"enable_ai"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Provider: ProviderMistral,
		Model:    "mistral-medium",
		APIKey:   "",
		APIURL:   "",
		EnableAI: true,
	}
}

// LoadConfig loads the AI configuration from viper
func LoadConfig(v *viper.Viper) (*Config, error) {
	config := DefaultConfig()

	// Load AI configuration from viper
	if v.IsSet("ai") {
		config.EnableAI = v.GetBool("ai.enable")

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

	// If AI is not enabled, return the config as is
	if !config.EnableAI {
		return config, nil
	}

	// Set API key from environment variables if not set
	if config.APIKey == "" {
		apiKey, err := utils.GetAPIKey(mapProviderToKeyInfo(config.Provider), true)
		if err != nil {
			// Wenn der API-Schlüssel nicht gefunden werden konnte, geben wir eine
			// benutzerfreundliche Fehlermeldung zurück.
			return config, fmt.Errorf("API-Schlüssel für %s nicht gefunden. %s",
				config.Provider, getAPIKeySetupInstructions(config.Provider))
		}
		config.APIKey = apiKey
	}

	// Set API URL from environment variables if not set
	if config.APIURL == "" {
		envVarName := getEndpointEnvVar(config.Provider)
		if envVarName != "" {
			if envValue := os.Getenv(envVarName); envValue != "" {
				config.APIURL = envValue
			}
		}
	}

	return config, nil
}

// GetModelConfig converts the unified config to a ModelConfig
func (c *Config) GetModelConfig(modelName string) (ModelConfig, error) {
	// If no model name is provided, use the default
	if modelName == "" {
		modelName = c.Model
	}

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
		return modelConfig, fmt.Errorf("API-Schlüssel für %s nicht konfiguriert. %s",
			c.Provider, getAPIKeySetupInstructions(c.Provider))
	}

	return modelConfig, nil
}

// Helper functions

// mapProviderToKeyInfo maps our provider enum to utils.ProviderKeyInfo
func mapProviderToKeyInfo(provider Provider) utils.ProviderKeyInfo {
	switch provider {
	case ProviderMistral:
		return utils.MistralProvider
	case ProviderOpenAI:
		return utils.OpenAIProvider
	case ProviderAnthropic:
		return utils.AnthropicProvider
	case ProviderGroq:
		return utils.GroqProvider
	case ProviderOpenRouter:
		return utils.OpenRouterProvider
	default:
		// Generische Struktur für andere Provider
		return utils.ProviderKeyInfo{
			Name:            string(provider),
			ConfigKey:       fmt.Sprintf("ai.api_key"),
			EnvVarName:      fmt.Sprintf("CLIKD_%s_API_KEY", provider),
			EnvVarNameShort: fmt.Sprintf("%s_API_KEY", provider),
			Required:        true,
		}
	}
}

// getAPIKeyEnvVar returns the environment variable name for the API key
func getAPIKeyEnvVar(provider Provider) string {
	switch provider {
	case ProviderMistral:
		return "CLIKD_MISTRAL_API_KEY"
	case ProviderOpenAI:
		return "CLIKD_OPENAI_API_KEY"
	case ProviderAnthropic:
		return "CLIKD_ANTHROPIC_API_KEY"
	case ProviderGroq:
		return "CLIKD_GROQ_API_KEY"
	case ProviderOpenRouter:
		return "CLIKD_OPENROUTER_API_KEY"
	default:
		return fmt.Sprintf("CLIKD_%s_API_KEY", provider)
	}
}

// getEndpointEnvVar returns the environment variable name for the endpoint
func getEndpointEnvVar(provider Provider) string {
	switch provider {
	case ProviderLocal:
		return "CLIKD_OLLAMA_BASE_URL"
	default:
		return "CLIKD_API_URL"
	}
}

// requiresEndpoint returns whether the provider requires an endpoint
func requiresEndpoint(provider Provider) bool {
	return provider == ProviderLocal
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
	envVar := getAPIKeyEnvVar(provider)

	return fmt.Sprintf(`
Sie können den Schlüssel auf folgende Weise hinzufügen:

1. Erstellen Sie eine .env-Datei im Projektverzeichnis und fügen Sie hinzu:
   %s=ihr_api_schlüssel

2. Oder fügen Sie den Schlüssel zu Ihrer globalen Konfiguration hinzu:
   clikd init config set %s=ihr_api_schlüssel

Um einen API-Schlüssel zu erhalten, besuchen Sie die Website des Anbieters: %s`,
		envVar, envVar, website)
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
