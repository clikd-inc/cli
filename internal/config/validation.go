package config

import (
	"fmt"
	"strings"
)

var (
	// SupportedProviders enthält alle unterstützten AI-Provider
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

	// ProviderModels enthält für jeden Provider die unterstützten Modelle
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
		// Für OpenRouter und Ollama erlauben wir alle Modelle
		"openrouter": {"*"}, // OpenRouter unterstützt verschiedene Modelle
		"ollama":     {"*"}, // Ollama unterstützt benutzerdefinierte Modelle
	}

	// ProviderAliases enthält alternative Namen für Provider
	ProviderAliases = map[string]string{
		"azure":  "azure-openai",
		"claude": "anthropic",
		"gemini": "google",
	}

	// ModelAliases enthält alternative Namen für Modelle
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

// ValidateProviderModel prüft, ob ein Modell mit einem Provider kompatibel ist
func ValidateProviderModel(provider, model string) error {
	if provider == "" || model == "" {
		return fmt.Errorf("provider und model dürfen nicht leer sein")
	}

	// Provider-Alias auflösen, falls vorhanden
	if alias, exists := ProviderAliases[provider]; exists {
		provider = alias
	}

	// Prüfen, ob der Provider unterstützt wird
	providerSupported := false
	for _, p := range SupportedProviders {
		if p == provider {
			providerSupported = true
			break
		}
	}

	if !providerSupported {
		return fmt.Errorf("provider '%s' wird nicht unterstützt. Unterstützte Provider: %s",
			provider, strings.Join(SupportedProviders, ", "))
	}

	// Modell-Alias auflösen, falls vorhanden
	if aliases, exists := ModelAliases[provider]; exists {
		if alias, exists := aliases[model]; exists {
			model = alias
		}
	}

	// Für Openrouter und Ollama erlauben wir alle Modelle
	if provider == "openrouter" || provider == "ollama" {
		return nil
	}

	// Prüfen, ob das Modell vom Provider unterstützt wird
	modelsForProvider, exists := ProviderModels[provider]
	if !exists {
		return fmt.Errorf("keine Modelle für Provider '%s' definiert", provider)
	}

	for _, m := range modelsForProvider {
		if m == model {
			return nil
		}
	}

	return fmt.Errorf("modell '%s' wird nicht vom Provider '%s' unterstützt.\nUnterstützte Modelle: %s",
		model, provider, strings.Join(modelsForProvider, ", "))
}

// GetSupportedModelsForProvider gibt alle unterstützten Modelle für einen Provider zurück
func GetSupportedModelsForProvider(provider string) ([]string, error) {
	// Provider-Alias auflösen, falls vorhanden
	if alias, exists := ProviderAliases[provider]; exists {
		provider = alias
	}

	// Prüfen, ob der Provider unterstützt wird
	providerSupported := false
	for _, p := range SupportedProviders {
		if p == provider {
			providerSupported = true
			break
		}
	}

	if !providerSupported {
		return nil, fmt.Errorf("provider '%s' wird nicht unterstützt", provider)
	}

	// Für Openrouter und Ollama können wir keine spezifische Liste zurückgeben
	if provider == "openrouter" || provider == "ollama" {
		return []string{"[Alle Modelle werden unterstützt]"}, nil
	}

	// Modelle für den Provider zurückgeben
	modelsForProvider, exists := ProviderModels[provider]
	if !exists {
		return nil, fmt.Errorf("keine Modelle für Provider '%s' definiert", provider)
	}

	return modelsForProvider, nil
}

// GetDefaultModelForProvider gibt das empfohlene Standardmodell für einen Provider zurück
func GetDefaultModelForProvider(provider string) (string, error) {
	// Provider-Alias auflösen, falls vorhanden
	if alias, exists := ProviderAliases[provider]; exists {
		provider = alias
	}

	// Prüfen, ob der Provider unterstützt wird
	providerSupported := false
	for _, p := range SupportedProviders {
		if p == provider {
			providerSupported = true
			break
		}
	}

	if !providerSupported {
		return "", fmt.Errorf("provider '%s' wird nicht unterstützt", provider)
	}

	// Standardmodell je nach Provider zurückgeben
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
		return "", fmt.Errorf("kein Standardmodell für Provider '%s' definiert", provider)
	}
}
