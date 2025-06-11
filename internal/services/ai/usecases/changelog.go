package usecases

import (
	"context"
	"fmt"
)

// Client interface needed for the changelog usecase
type Client interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Chat generates a response for a conversation
	Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error)
}

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

// ChatOption represents an option for chat completion
type ChatOption func(*CompletionRequest)

// WithJSONResponse sets the response type to JSON
func WithJSONResponse() ChatOption {
	return func(r *CompletionRequest) {
		r.ResponseType = "json"
	}
}

// EnhanceChangelog improves a generated changelog
func EnhanceChangelog(client Client, ctx context.Context, changelog string) (string, error) {
	prompt := `You are an expert in writing clear, concise, and informative changelogs.
Please enhance the following changelog to make it more professional and readable.
Maintain the same structure and format, but improve clarity, fix grammar issues,
and ensure consistent style throughout.

CHANGELOG:
%s

ENHANCED CHANGELOG:`

	req := &CompletionRequest{
		Prompt:      fmt.Sprintf(prompt, changelog),
		MaxTokens:   1024,
		Temperature: 0.7,
		TopP:        0.9,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		return changelog, fmt.Errorf("failed to enhance changelog: %w", err)
	}

	return resp.Text, nil
}
