package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	mistralAPIURL = "https://api.mistral.ai/v1"
)

// MistralClient implements the Client interface for Mistral AI
type MistralClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	modelName  string
	config     ModelConfig
}

// mistralChatRequest represents a request to the Mistral chat API
type mistralChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Safe        bool      `json:"safe_prompt,omitempty"`
	RandomSeed  *int      `json:"random_seed,omitempty"`
}

// mistralChatResponse represents a response from the Mistral chat API
type mistralChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewMistralClient creates a new Mistral AI client
func NewMistralClient(ctx context.Context, config ModelConfig) (Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Mistral API key is required")
	}

	baseURL := mistralAPIURL
	if config.Endpoint != "" {
		baseURL = config.Endpoint
	}

	return &MistralClient{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		modelName:  config.ModelID,
		config:     config,
	}, nil
}

// Complete implements the Client interface
func (c *MistralClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Mistral's API doesn't have a dedicated completion endpoint, so we'll use the chat API
	// with a single system message
	messages := []Message{
		{
			Role:    "user",
			Content: req.Prompt,
		},
	}

	return c.Chat(ctx, messages,
		WithMaxTokens(req.MaxTokens),
		WithTemperature(req.Temperature),
		WithTopP(req.TopP),
		WithStream(req.Stream),
	)
}

// Chat implements the Client interface
func (c *MistralClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
	req := &CompletionRequest{
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		TopP:        c.config.TopP,
		Stream:      c.config.StreamResponse,
		ModelName:   c.modelName,
	}

	// Apply options
	for _, opt := range options {
		opt(req)
	}

	// Create Mistral-specific request
	mistralReq := mistralChatRequest{
		Model:       c.modelName,
		Messages:    messages,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Safe:        true,
	}

	// Serialize request
	reqBody, err := json.Marshal(mistralReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var mistralResp mistralChatResponse
	if err := json.Unmarshal(respBody, &mistralResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if choices are empty
	if len(mistralResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Convert to CompletionResponse
	return &CompletionResponse{
		Text:         mistralResp.Choices[0].Message.Content,
		FinishReason: mistralResp.Choices[0].FinishReason,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     mistralResp.Usage.PromptTokens,
			CompletionTokens: mistralResp.Usage.CompletionTokens,
			TotalTokens:      mistralResp.Usage.TotalTokens,
		},
	}, nil
}

// GetProvider implements the Client interface
func (c *MistralClient) GetProvider() Provider {
	return ProviderMistral
}

// GetModelName implements the Client interface
func (c *MistralClient) GetModelName() string {
	return c.modelName
}

// GetCapabilities implements the Client interface
func (c *MistralClient) GetCapabilities() ModelCapabilities {
	return ModelCapabilities{
		SupportsChat:         true,
		SupportsCompletion:   true, // via chat API
		SupportsStream:       true,
		SupportsJSONResponse: true,
		MaxContextWindow:     c.config.ContextWindow,
	}
}
