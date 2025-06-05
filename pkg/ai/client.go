package ai

import (
	"context"
	"fmt"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Prompt       string    `json:"prompt"`
	Messages     []Message `json:"messages"`
	MaxTokens    int       `json:"max_tokens"`
	Temperature  float64   `json:"temperature"`
	TopP         float64   `json:"top_p"`
	Stop         []string  `json:"stop,omitempty"`
	Stream       bool      `json:"stream"`
	ModelName    string    `json:"model_name,omitempty"`
	ResponseType string    `json:"response_type,omitempty"` // text, json, etc.
}

// CompletionResponse represents a response from text completion
type CompletionResponse struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"`
	Usage        struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Client defines the interface for AI providers
type Client interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Chat generates a response for a conversation
	Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error)

	// GetProvider returns the provider type
	GetProvider() Provider

	// GetModelName returns the model name
	GetModelName() string

	// GetCapabilities returns the capabilities of the model
	GetCapabilities() ModelCapabilities
}

// ModelCapabilities represents the capabilities of a model
type ModelCapabilities struct {
	SupportsChat         bool `json:"supports_chat"`
	SupportsCompletion   bool `json:"supports_completion"`
	SupportsStream       bool `json:"supports_stream"`
	SupportsJSONResponse bool `json:"supports_json_response"`
	MaxContextWindow     int  `json:"max_context_window"`
}

// ChatOption represents an option for chat completion
type ChatOption func(*CompletionRequest)

// WithTemperature sets the temperature for the request
func WithTemperature(temp float64) ChatOption {
	return func(r *CompletionRequest) {
		r.Temperature = temp
	}
}

// WithMaxTokens sets the max tokens for the request
func WithMaxTokens(tokens int) ChatOption {
	return func(r *CompletionRequest) {
		r.MaxTokens = tokens
	}
}

// WithTopP sets the top_p for the request
func WithTopP(topP float64) ChatOption {
	return func(r *CompletionRequest) {
		r.TopP = topP
	}
}

// WithStream sets whether to stream the response
func WithStream(stream bool) ChatOption {
	return func(r *CompletionRequest) {
		r.Stream = stream
	}
}

// WithStop sets the stop sequences for the request
func WithStop(stop []string) ChatOption {
	return func(r *CompletionRequest) {
		r.Stop = stop
	}
}

// WithJSONResponse sets the response type to JSON
func WithJSONResponse() ChatOption {
	return func(r *CompletionRequest) {
		r.ResponseType = "json"
	}
}

// NewClient creates a new AI client based on the configuration
func NewClient(ctx context.Context, config *Config, modelName string) (Client, error) {
	// If no model specified, use the default
	if modelName == "" {
		modelName = config.DefaultModel
	}

	// Get the model configuration
	modelConfig, err := config.GetModelConfig(modelName)
	if err != nil {
		// Wenn der Fehler von GetAPIKey kommt, diese Fehlermeldung direkt weitergeben
		// da sie bereits detaillierte Anweisungen enthält
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	// Check if API key is configured
	if modelConfig.APIKey == "" && modelConfig.Provider != ProviderLocal {
		// Wenn wir hier ankommen, dann ist etwas schief gelaufen.
		// GetModelConfig sollte normalerweise bereits einen Fehler mit detaillierten Anweisungen
		// zurückgeben, falls kein API-Schlüssel gefunden wurde.

		// Trotzdem wollen wir sicherstellen, dass der Benutzer hilfreiche Informationen erhält
		return nil, fmt.Errorf("API-Schlüssel für das Modell %s (%s) fehlt. Bitte führen Sie den Befehl erneut aus.",
			modelName, modelConfig.Provider)
	}

	// Create the appropriate client based on the provider
	switch modelConfig.Provider {
	case ProviderMistral:
		return NewMistralClient(ctx, modelConfig)
	case ProviderOpenAI:
		return NewOpenAIClient(ctx, modelConfig)
	case ProviderAzureOpenAI:
		return NewAzureOpenAIClient(ctx, modelConfig)
	case ProviderLocal:
		return NewLocalClient(ctx, modelConfig)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", modelConfig.Provider)
	}
}

// ClientOption represents an option for client creation
type ClientOption func(*ClientOptions)

// ClientOptions represents options for client creation
type ClientOptions struct {
	Timeout int
	Debug   bool
}

// WithTimeout sets the timeout for the client
func WithTimeout(timeout int) ClientOption {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// WithDebug enables debug mode for the client
func WithDebug(debug bool) ClientOption {
	return func(o *ClientOptions) {
		o.Debug = debug
	}
}
