package ai

import (
	"context"
	"fmt"

	"clikd/internal/services/ai/usecases"
	"clikd/internal/utils"
)

// Use the same logger instance as in other AI package files
var logService = utils.NewLogger("error", true)

// Service defines the interface for AI operations
type Service interface {
	// Commit-related operations - batch processing only
	EnhanceCommitMessagesBatch(commitMessages []string) (map[string][]string, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	client      Client
	provider    string
	model       string
	maxTokens   int
	temperature float64
	topP        float64
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
		client:      client,
		provider:    provider,
		model:       model,
		maxTokens:   tokensMaxInput, // Use input tokens as max tokens for generation
		temperature: 0.1,            // Very low temperature for fastest, most deterministic responses
		topP:        0.7,            // Lower topP for faster token selection and more focused output
	}, nil
}

// EnhanceCommitMessagesBatch implements the Service interface for batch processing
func (s *ServiceImpl) EnhanceCommitMessagesBatch(commitMessages []string) (map[string][]string, error) {
	logService.Debug("Enhancing %d commit messages in batch with AI using config values: maxTokens=%d, temperature=%.2f, topP=%.2f",
		len(commitMessages), s.maxTokens, s.temperature, s.topP)

	// Create client adapter for usecase
	clientAdapter := &usecaseClientAdapter{client: s.client}

	// Use configuration values from service
	options := usecases.EnhanceChangelogOptions{
		MaxTokens:   s.maxTokens,
		Temperature: s.temperature,
		TopP:        s.topP,
	}

	// Delegate to the usecase implementation with configuration
	ctx := context.Background()
	enhancedMap, err := usecases.EnhanceCommitMessagesBatch(clientAdapter, ctx, commitMessages, options)
	if err != nil {
		logService.Debug("Failed to enhance commit messages batch: %v", err)
		// Return original messages on error
		result := make(map[string][]string)
		for _, msg := range commitMessages {
			result[msg] = []string{msg}
		}
		return result, nil
	}

	logService.Debug("Successfully enhanced %d commit messages in batch", len(commitMessages))
	return enhancedMap, nil
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
