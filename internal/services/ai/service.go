package ai

import (
	"context"
	"fmt"

	"clikd/internal/services/ai/usecases"
	"clikd/internal/utils"
)

// Use the same logger instance as in other AI package files
var logService = utils.NewLogger("info", true)

// Service defines the interface for AI operations
type Service interface {
	EnhanceChangelog(changelog string) (string, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	client   Client
	provider string
	model    string
}

// NewService creates a new AI service
func NewService(ctx context.Context, provider, model, apiKey, endpoint string, tokensMaxInput, tokensMaxOutput int) (Service, error) {
	// Provider and model are required for AI service
	if provider == "" || model == "" {
		return nil, fmt.Errorf("AI provider and model are required")
	}

	// Create AI client
	logService.Debug("Creating AI client with provider %s and model %s", provider, model)
	client, err := NewClient(ctx, provider, model, apiKey, endpoint, tokensMaxInput, tokensMaxOutput)
	if err != nil {
		logService.Error("Failed to create AI client: %v", err)
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	logService.Info("AI service initialized successfully with provider %s and model %s", provider, model)
	return &ServiceImpl{
		client:   client,
		provider: provider,
		model:    model,
	}, nil
}

// EnhanceChangelog implements the Service interface
func (s *ServiceImpl) EnhanceChangelog(changelog string) (string, error) {
	logService.Debug("Enhancing changelog with AI")

	// Create client adapter for usecase
	clientAdapter := &usecaseClientAdapter{client: s.client}

	// Delegate to the usecase implementation
	ctx := context.Background()
	enhancedChangelog, err := usecases.EnhanceChangelog(clientAdapter, ctx, changelog)
	if err != nil {
		logService.Error("Failed to enhance changelog: %v", err)
		return changelog, err
	}

	logService.Debug("Changelog enhanced successfully")
	return enhancedChangelog, nil
}

// usecaseClientAdapter adapts our Client to the usecases.Client interface
type usecaseClientAdapter struct {
	client Client
}

// Complete implements the usecases.Client interface
func (a *usecaseClientAdapter) Complete(ctx context.Context, req *usecases.CompletionRequest) (*usecases.CompletionResponse, error) {
	// Convert request
	aiReq := &CompletionRequest{
		Prompt:       req.Prompt,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		TopP:         req.TopP,
		Stop:         req.Stop,
		Stream:       req.Stream,
		ModelName:    req.ModelName,
		ResponseType: req.ResponseType,
	}

	// Call the actual client
	logService.Debug("Calling AI completion with prompt length %d", len(req.Prompt))
	resp, err := a.client.Complete(ctx, aiReq)
	if err != nil {
		logService.Error("AI completion failed: %v", err)
		return nil, err
	}

	logService.Debug("AI completion successful, response length: %d", len(resp.Text))
	// Convert response
	return &usecases.CompletionResponse{
		Text:         resp.Text,
		FinishReason: resp.FinishReason,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// Chat implements the usecases.Client interface
func (a *usecaseClientAdapter) Chat(ctx context.Context, messages []usecases.Message, options ...usecases.ChatOption) (*usecases.CompletionResponse, error) {
	// Convert messages
	aiMessages := make([]Message, len(messages))
	for i, msg := range messages {
		aiMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	logService.Debug("Calling AI chat with %d messages", len(messages))

	// Convert options
	aiOptions := make([]ChatOption, len(options))
	for i, opt := range options {
		aiOptions[i] = func(r *CompletionRequest) {
			// Create a temporary request to apply the usecase option
			tempReq := &usecases.CompletionRequest{}
			opt(tempReq)

			// Copy the changes to our request
			r.ResponseType = tempReq.ResponseType
			// Add other fields as needed
		}
	}

	// Call the actual client
	resp, err := a.client.Chat(ctx, aiMessages, aiOptions...)
	if err != nil {
		logService.Error("AI chat failed: %v", err)
		return nil, err
	}

	logService.Debug("AI chat successful, response length: %d", len(resp.Text))
	// Convert response
	return &usecases.CompletionResponse{
		Text:         resp.Text,
		FinishReason: resp.FinishReason,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}
