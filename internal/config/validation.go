package config

import (
	"fmt"
	"strings"
)

var (
	// SupportedProviders contains all supported AI providers
	SupportedProviders = []string{
		"openai",
		"mistral",
		"anthropic",
		"azure-openai",
		"groq",
		"openrouter",
		"ollama",
		"google",
	}

	// ProviderModels contains the supported models for each provider
	ProviderModels = map[string][]string{
		"openai": {
			"gpt-4", "gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo",
			"gpt-3.5-turbo-0125", "gpt-4-1106-preview", "gpt-4-turbo-preview",
			"gpt-4-0125-preview", "gpt-4-32k", "gpt-3.5-turbo-16k",
		},
		"mistral": {
			"mistral-tiny", "mistral-small", "mistral-medium", "mistral-large",
			"mistral-small-latest", "mistral-medium-latest", "mistral-large-latest",
			"open-mistral-7b", "open-mixtral-8x7b",
		},
		"anthropic": {
			"claude-3-opus", "claude-3-sonnet", "claude-3-haiku",
			"claude-2.1", "claude-2.0", "claude-instant-1",
		},
		"azure-openai": {
			"gpt-4", "gpt-35-turbo", "gpt-4-turbo", "gpt-4-32k", "gpt-35-turbo-16k",
			"text-embedding-ada-002",
		},
		"groq": {
			"llama2-70b-4096", "mixtral-8x7b-32768", "gemma-7b-it",
			"llama3-8b-8192", "llama3-70b-8192", "mixtral-8x7b-instruct",
		},
		"google": {
			"gemini-pro", "gemini-ultra", "gemini-flash",
		},
		// For OpenRouter and Ollama we allow all models
		"openrouter": {"*"}, // OpenRouter supports various models
		"ollama":     {"*"}, // Ollama supports custom models
	}

	// ProviderAliases contains alternative names for providers
	ProviderAliases = map[string]string{
		"azure":  "azure-openai",
		"claude": "anthropic",
		"gemini": "google",
	}

	// ModelAliases contains alternative names for models
	ModelAliases = map[string]map[string]string{
		"openai": {
			"gpt4":  "gpt-4",
			"gpt4o": "gpt-4o",
			"gpt35": "gpt-3.5-turbo",
		},
		"mistral": {
			"tiny":   "mistral-tiny",
			"small":  "mistral-small",
			"medium": "mistral-medium",
			"large":  "mistral-large",
		},
		"anthropic": {
			"opus":   "claude-3-opus",
			"sonnet": "claude-3-sonnet",
			"haiku":  "claude-3-haiku",
		},
	}
)

// ValidateProviderModel checks if a model is compatible with a provider
func ValidateProviderModel(provider, model string) error {
	if provider == "" || model == "" {
		return fmt.Errorf("provider and model must not be empty")
	}

	// Resolve provider alias if present
	if alias, exists := ProviderAliases[provider]; exists {
		provider = alias
	}

	// Check if the provider is supported
	providerSupported := false
	for _, p := range SupportedProviders {
		if p == provider {
			providerSupported = true
			break
		}
	}

	if !providerSupported {
		return fmt.Errorf("provider '%s' is not supported. Supported providers: %s",
			provider, strings.Join(SupportedProviders, ", "))
	}

	// Resolve model alias if present
	if aliases, exists := ModelAliases[provider]; exists {
		if alias, exists := aliases[model]; exists {
			model = alias
		}
	}

	// For OpenRouter and Ollama we allow all models
	if provider == "openrouter" || provider == "ollama" {
		return nil
	}

	// Check if the model is supported by the provider
	modelsForProvider, exists := ProviderModels[provider]
	if !exists {
		return fmt.Errorf("no models defined for provider '%s'", provider)
	}

	for _, m := range modelsForProvider {
		if m == model {
			return nil
		}
	}

	return fmt.Errorf("model '%s' is not supported by provider '%s'.\nSupported models: %s",
		model, provider, strings.Join(modelsForProvider, ", "))
}

// GetSupportedModelsForProvider returns all supported models for a provider
func GetSupportedModelsForProvider(provider string) ([]string, error) {
	// Resolve provider alias if present
	if alias, exists := ProviderAliases[provider]; exists {
		provider = alias
	}

	// Check if the provider is supported
	providerSupported := false
	for _, p := range SupportedProviders {
		if p == provider {
			providerSupported = true
			break
		}
	}

	if !providerSupported {
		return nil, fmt.Errorf("provider '%s' is not supported", provider)
	}

	// For OpenRouter and Ollama we cannot return a specific list
	if provider == "openrouter" || provider == "ollama" {
		return []string{"[All models are supported]"}, nil
	}

	// Return models for the provider
	modelsForProvider, exists := ProviderModels[provider]
	if !exists {
		return nil, fmt.Errorf("no models defined for provider '%s'", provider)
	}

	return modelsForProvider, nil
}

// GetDefaultModelForProvider returns the recommended default model for a provider
func GetDefaultModelForProvider(provider string) (string, error) {
	// Resolve provider alias if present
	if alias, exists := ProviderAliases[provider]; exists {
		provider = alias
	}

	// Check if the provider is supported
	providerSupported := false
	for _, p := range SupportedProviders {
		if p == provider {
			providerSupported = true
			break
		}
	}

	if !providerSupported {
		return "", fmt.Errorf("provider '%s' is not supported", provider)
	}

	// Return default model according to provider
	switch provider {
	case "openai":
		return "gpt-4o", nil
	case "mistral":
		return "mistral-medium", nil
	case "anthropic":
		return "claude-3-sonnet", nil
	case "azure-openai":
		return "gpt-4", nil
	case "groq":
		return "llama3-70b-8192", nil
	case "google":
		return "gemini-pro", nil
	case "openrouter":
		return "openai/gpt-4o", nil
	case "ollama":
		return "llama3", nil
	default:
		return "", fmt.Errorf("no default model defined for provider '%s'", provider)
	}
}
