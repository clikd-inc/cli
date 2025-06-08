package ai

import (
	"context"
	"fmt"

	"clikd/internal/utils"
)

// Add logger
var log = utils.NewLogger("info", true)

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
		modelName = config.Model
	}

	// Get the model configuration
	modelConfig, err := config.GetModelConfig(modelName)
	if err != nil {
		// Log the error before returning it
		log.Error("Failed to get model config for %s: %v", modelName, err)
		// If the error comes from GetAPIKey, pass this error message directly
		// as it already contains detailed instructions
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	// Check if API key is configured for non-local providers
	if modelConfig.APIKey == "" && modelConfig.Provider != ProviderLocal {
		// Log the error
		log.Error("API key missing for non-local provider %s model %s",
			modelConfig.Provider, modelName)

		// If we arrive here, something went wrong.
		// GetModelConfig should normally already return an error with detailed instructions
		// if no API key was found.

		// Nevertheless, we want to ensure that the user receives helpful information
		return nil, fmt.Errorf("API key for model %s (%s) is missing, please run the command again",
			modelName, modelConfig.Provider)
	}

	// Gollm supports all providers, so we use it exclusively
	client, err := NewGollmClient(ctx, modelConfig)
	if err != nil {
		log.Error("Failed to create Gollm client for %s (%s): %v",
			modelName, modelConfig.Provider, err)
		return nil, err
	}

	log.Debug("AI client created successfully for %s (%s)",
		modelName, modelConfig.Provider)

	return client, nil
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
