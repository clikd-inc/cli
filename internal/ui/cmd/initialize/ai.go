package initialize

import (
	"fmt"

	"clikd/internal/config"
	"clikd/internal/ui/bubble"
)

// selectAIProvider displays a selection dialog for AI providers
// and returns the selected provider
func selectAIProvider() string {
	// Get supported providers
	providerOptions := config.SupportedProviders

	// Filter for supported providers
	supportedGollmProviders := []string{"mistral", "openai", "anthropic", "groq"}
	filteredProviderOptions := []string{}

	for _, provider := range providerOptions {
		for _, supported := range supportedGollmProviders {
			if provider == supported {
				filteredProviderOptions = append(filteredProviderOptions, provider)
				break
			}
		}
	}

	// Create select items for providers
	providerItems := make([]bubble.SelectItem, len(filteredProviderOptions))
	for i, provider := range filteredProviderOptions {
		defaultModel, _ := config.GetDefaultModelForProvider(provider)
		description := fmt.Sprintf("Default model: %s", defaultModel)
		if provider == "mistral" {
			description = fmt.Sprintf("RECOMMENDED - Default model: %s", defaultModel)
		}
		providerItems[i] = bubble.SelectItem{
			Title:       provider,
			Description: description,
			Value:       provider,
		}
	}

	// Run selection dialog
	selected := bubble.RunSelect("Select AI Provider", providerItems)
	if selected == nil {
		return ""
	}

	return selected.Value.(string)
}

// selectAIModel displays a selection dialog for AI models
// based on the selected provider and returns the selected model
func selectAIModel(provider string) string {
	// Get models for the selected provider
	supportedModels, _ := config.GetSupportedModelsForProvider(provider)

	// Define recommended models for each provider
	recommendedModels := map[string]string{
		"mistral":   "mistral-medium",
		"openai":    "gpt-4o",
		"anthropic": "claude-3-sonnet",
		"groq":      "llama3-70b-8192",
	}

	defaultModel, _ := config.GetDefaultModelForProvider(provider)
	recommendedModel := defaultModel

	if recModel, exists := recommendedModels[provider]; exists {
		recommendedModel = recModel
	}

	// Create model items
	modelItems := make([]bubble.SelectItem, len(supportedModels))
	for i, model := range supportedModels {
		description := fmt.Sprintf("Model for %s", provider)

		if model == defaultModel && model == recommendedModel {
			description = "DEFAULT & RECOMMENDED - Best balance of performance and cost"
		} else if model == defaultModel {
			description = "DEFAULT - Standard model for this provider"
		} else if model == recommendedModel {
			description = "RECOMMENDED - Best balance of performance and cost"
		}

		// Provider-specific descriptions
		if provider == "mistral" {
			if model == "mistral-tiny" {
				description = "Fastest, most cost-effective option, but less capable"
			} else if model == "mistral-small" {
				description = "Good balance of speed and capability"
			} else if model == "mistral-medium" {
				description = "RECOMMENDED - Best overall value"
			} else if model == "mistral-large" {
				description = "Most capable Mistral model, but more expensive"
			}
		} else if provider == "openai" {
			if model == "gpt-3.5-turbo" {
				description = "Faster and cheaper, but less capable"
			} else if model == "gpt-4o" {
				description = "RECOMMENDED - Latest model with best performance"
			} else if model == "gpt-4" {
				description = "Older model, still powerful but slower than gpt-4o"
			}
		} else if provider == "anthropic" {
			if model == "claude-3-haiku" {
				description = "Faster and cheaper, good for simple tasks"
			} else if model == "claude-3-sonnet" {
				description = "RECOMMENDED - Good balance of capability and cost"
			} else if model == "claude-3-opus" {
				description = "Most capable Claude model, but more expensive"
			}
		}

		modelItems[i] = bubble.SelectItem{
			Title:       model,
			Description: description,
			Value:       model,
		}
	}

	// Run selection dialog
	selected := bubble.RunSelect(fmt.Sprintf("Select Model for %s", provider), modelItems)
	if selected == nil {
		return ""
	}

	return selected.Value.(string)
}

// configureAdvancedAIOptions handles the advanced AI options configuration
// and returns the configured options as a map
func configureAdvancedAIOptions() map[string]string {
	options := make(map[string]string)

	// Ask if user wants to configure advanced options
	configureAdvanced := bubble.RunConfirm(
		"Advanced AI Options",
		"Do you want to configure advanced AI options?",
	)

	if !configureAdvanced {
		// Use defaults
		options["tokens_max_input"] = "4096"
		options["tokens_max_output"] = "500"
		return options
	}

	// Configure advanced options - placeholder values will be used as defaults automatically
	options["tokens_max_input"] = bubble.RunInput(
		"Max Input Tokens",
		"Maximum number of input tokens (context size)",
		"4096",
	)

	options["tokens_max_output"] = bubble.RunInput(
		"Max Output Tokens",
		"Maximum number of output tokens (response length)",
		"500",
	)

	options["api_url"] = bubble.RunInput(
		"Custom API URL",
		"Custom API endpoint URL (leave empty to use official API)",
		"",
	)

	options["api_custom_headers"] = bubble.RunInput(
		"Custom API Headers",
		"Custom HTTP headers in JSON format (leave empty for standard authentication)",
		"",
	)

	return options
}

// configureAPIKey handles the API key configuration
// Returns the API key if provided, empty string otherwise
func configureAPIKey(provider string, isGlobal bool) string {
	if isGlobal {
		// For global config, ask for API key directly
		return bubble.RunPasswordInput(
			"API Key Configuration",
			fmt.Sprintf("Enter your %s API key (or leave empty to configure later)", provider),
			"Your API key",
		)
	} else {
		// For local config, ask if user wants to create/update .env file
		createEnv := bubble.RunConfirm(
			"API Key Configuration",
			"Do you want to create/update the .env file with your API key?",
		)

		if !createEnv {
			return ""
		}

		return bubble.RunPasswordInput(
			"API Key for .env file",
			fmt.Sprintf("Enter your %s API key for the .env file", provider),
			"Your API key",
		)
	}
}
