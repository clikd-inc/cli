package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/teilomillet/gollm"
)

// GollmClient implements the Client interface using gollm
type GollmClient struct {
	llm      gollm.LLM
	provider Provider
	modelID  string
	config   ModelConfig
}

// NewGollmClient creates a new gollm-based client
func NewGollmClient(ctx context.Context, config ModelConfig) (Client, error) {
	// Map our provider to gollm provider
	providerName := mapProviderToGollm(config.Provider)

	// Map model name to gollm-compatible model name
	modelName := mapModelToGollm(config.Provider, config.ModelID)

	// Create configuration options
	options := []gollm.ConfigOption{
		gollm.SetProvider(providerName),
		gollm.SetModel(modelName),
		gollm.SetMaxTokens(config.MaxTokens),
		gollm.SetTemperature(config.Temperature),
		gollm.SetTopP(config.TopP),
		gollm.SetAPIKey(config.APIKey),
	}

	// Set debug level to error to reduce log pollution
	options = append(options, gollm.SetLogLevel(gollm.LogLevelError))

	// Add support for memory if needed
	if config.ContextWindow > 0 {
		options = append(options, gollm.SetMemory(config.ContextWindow))
	}

	// Set max retries for API calls
	options = append(options, gollm.SetMaxRetries(3))

	// Create the LLM instance
	llm, err := gollm.NewLLM(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gollm client: %w", err)
	}

	// Set provider-specific configurations
	switch config.Provider {
	case ProviderLocal:
		// Set Ollama endpoint if provided
		if config.Endpoint != "" {
			llm.SetOption("ollama_base_url", config.Endpoint)
		}
	case ProviderOpenRouter:
		// Enable OpenRouter features if needed
		llm.SetOption("enable_prompt_caching", true)
		llm.SetOption("enable_reasoning", true)

		// Add fallback models for reliability
		llm.SetOption("fallback_models", []string{
			"openai/gpt-4o",
			"anthropic/claude-3-opus-20240229",
			"mistral/mistral-large-latest",
		})
	}

	return &GollmClient{
		llm:      llm,
		provider: config.Provider,
		modelID:  config.ModelID, // Keep original model ID for reference
		config:   config,
	}, nil
}

// Complete implements the Client interface
func (c *GollmClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	var prompt *gollm.Prompt

	// Erstelle Prompt basierend auf dem Request
	if len(req.Messages) > 0 {
		// Für Chat-Modelle, Baue einen Prompt mit System- und User-Nachrichten
		var systemPrompt, userPrompt string
		var assistantMessages []string

		for _, msg := range req.Messages {
			switch strings.ToLower(msg.Role) {
			case "system":
				systemPrompt += msg.Content + "\n"
			case "user":
				userPrompt += msg.Content + "\n"
			case "assistant":
				assistantMessages = append(assistantMessages, msg.Content)
			}
		}

		// Prompt-Optionen für NewPrompt
		options := []gollm.PromptOption{}

		// Füge Systemkontext hinzu, wenn vorhanden
		if systemPrompt != "" {
			options = append(options, gollm.WithContext(systemPrompt))
		}

		// Füge Beispiele hinzu, wenn vorhanden
		if len(assistantMessages) > 0 {
			options = append(options, gollm.WithExamples(assistantMessages...))
		}

		// Erstelle den Prompt
		prompt = gollm.NewPrompt(userPrompt, options...)
	} else {
		// Für Completion-Modelle, nutze den Prompt direkt
		prompt = gollm.NewPrompt(req.Prompt)
	}

	// Füge JSON-Validierung hinzu, wenn angefordert
	if req.ResponseType == "json" {
		// Spezielle Option für JSON-Antworten (falls unterstützt)
		prompt = gollm.NewPrompt(prompt.Input,
			gollm.WithOutput("Respond in valid JSON format."),
		)
	}

	// Generiere die Antwort
	response, err := c.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("gollm generation failed: %w", err)
	}

	// Erstelle die Completion-Antwort
	return &CompletionResponse{
		Text: response,
		// Da gollm keine Nutzungsstatistiken direkt bereitstellt, lassen wir sie leer
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     0, // Nicht von gollm bereitgestellt
			CompletionTokens: 0, // Nicht von gollm bereitgestellt
			TotalTokens:      0, // Nicht von gollm bereitgestellt
		},
	}, nil
}

// Chat implements the Client interface for chat models
func (c *GollmClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
	// Create a request from the messages
	req := &CompletionRequest{
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		TopP:        c.config.TopP,
	}

	// Apply options
	for _, opt := range options {
		opt(req)
	}

	// Use the Complete method which handles both chat and completion
	return c.Complete(ctx, req)
}

// GetProvider returns the provider type
func (c *GollmClient) GetProvider() Provider {
	return c.provider
}

// GetModelName returns the model name
func (c *GollmClient) GetModelName() string {
	return c.modelID
}

// GetCapabilities returns the capabilities of the model
func (c *GollmClient) GetCapabilities() ModelCapabilities {
	// Determine capabilities based on the provider
	supportsChat := true
	supportsJSON := true // gollm supports JSON schema validation for all providers
	maxContext := c.config.ContextWindow

	return ModelCapabilities{
		SupportsChat:         supportsChat,
		SupportsCompletion:   true,
		SupportsStream:       c.config.StreamResponse,
		SupportsJSONResponse: supportsJSON,
		MaxContextWindow:     maxContext,
	}
}

// Helper functions

// mapProviderToGollm maps our provider enum to gollm's provider string
func mapProviderToGollm(provider Provider) string {
	switch provider {
	case ProviderMistral:
		return "mistral"
	case ProviderOpenAI:
		return "openai"
	case ProviderLocal:
		return "ollama"
	case ProviderAnthropic:
		return "anthropic"
	case ProviderGroq:
		return "groq"
	case ProviderOpenRouter:
		return "openrouter"
	default:
		return string(provider)
	}
}

// mapModelToGollm maps our model enum to gollm's compatible model name
func mapModelToGollm(provider Provider, modelID string) string {
	switch provider {
	case ProviderMistral:
		// Map Mistral model names to Gollm-compatible names
		// Based on Gollm documentation and common Mistral API model names
		switch modelID {
		case "mistral-medium":
			// Use the original mapping that works despite the warning
			return "mistral-medium-latest"
		case "mistral-small":
			return "mistral-small-latest"
		case "mistral-large":
			return "mistral-large-latest"
		case "mistral-tiny":
			return "open-mistral-7b"
		case "mistral-7b":
			return "open-mistral-7b"
		case "mixtral-8x7b":
			return "open-mixtral-8x7b"
		default:
			// For unknown Mistral models, try the original name first
			return modelID
		}
	case ProviderOpenAI:
		// OpenAI models are usually correctly named
		return modelID
	case ProviderLocal:
		// Local models (Ollama) use their original names
		return modelID
	case ProviderAnthropic:
		// Anthropic models are usually correctly named
		return modelID
	case ProviderGroq:
		// Groq models are usually correctly named
		return modelID
	case ProviderOpenRouter:
		// OpenRouter models are usually correctly named
		return modelID
	default:
		return modelID
	}
}
