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
	CategorizeCommit(commitMessage string) (string, error)
	EnhanceCommitMessage(commitMessage string) (string, error)
	ExtractCommitInfo(commitMessage string) (map[string]interface{}, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	client Client
	config *Config
}

// NewService creates a new AI service
func NewService(ctx context.Context, config *Config) (Service, error) {
	// Config is required for AI service
	if config == nil {
		return nil, fmt.Errorf("AI configuration is required")
	}

	// Create AI client
	logService.Debug("Creating AI client with provider %s and model %s",
		config.Provider, config.Model)
	client, err := NewClient(ctx, config, "")
	if err != nil {
		logService.Error("Failed to create AI client: %v", err)
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	logService.Info("AI service initialized successfully with provider %s and model %s",
		config.Provider, config.Model)
	return &ServiceImpl{
		client: client,
		config: config,
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

// CategorizeCommit implements the Service interface
func (s *ServiceImpl) CategorizeCommit(commitMessage string) (string, error) {
	logService.Debug("Categorizing commit with AI")

	// Create client adapter for usecase
	clientAdapter := &usecaseClientAdapter{client: s.client}

	// Delegate to the usecase implementation
	ctx := context.Background()
	category, err := usecases.CategorizeCommit(clientAdapter, ctx, commitMessage)
	if err != nil {
		logService.Error("Failed to categorize commit: %v", err)
		return "other", err
	}

	logService.Debug("Commit categorized successfully as: %s", category)
	return string(category), nil
}

// EnhanceCommitMessage implements the Service interface
func (s *ServiceImpl) EnhanceCommitMessage(commitMessage string) (string, error) {
	logService.Debug("Enhancing commit message with AI")

	// Create client adapter for usecase
	clientAdapter := &usecaseClientAdapter{client: s.client}

	// Delegate to the usecase implementation
	ctx := context.Background()
	enhancedMessage, err := usecases.EnhanceCommitMessage(clientAdapter, ctx, commitMessage)
	if err != nil {
		logService.Error("Failed to enhance commit message: %v", err)
		return commitMessage, err
	}

	logService.Debug("Commit message enhanced successfully")
	return enhancedMessage, nil
}

// ExtractCommitInfo implements the Service interface
func (s *ServiceImpl) ExtractCommitInfo(commitMessage string) (map[string]interface{}, error) {
	logService.Debug("Extracting commit info with AI")

	// Create client adapter for usecase
	clientAdapter := &usecaseClientAdapter{client: s.client}

	// Delegate to the usecase implementation
	ctx := context.Background()
	// ExtractCommitInfo requires author, date, and hash parameters
	// For now, we'll use empty values since they're not provided in the interface
	info, err := usecases.ExtractCommitInfo(clientAdapter, ctx, commitMessage, "", "", "")
	if err != nil {
		logService.Error("Failed to extract commit info: %v", err)
		return nil, err
	}

	// Convert CommitInfo to map[string]interface{}
	result := map[string]interface{}{
		"message":    info.Message,
		"author":     info.Author,
		"date":       info.Date,
		"hash":       info.Hash,
		"category":   string(info.Category),
		"scope":      info.Scope,
		"summary":    info.Summary,
		"issue_refs": info.IssueRefs,
	}

	logService.Debug("Commit info extracted successfully")
	return result, nil
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
