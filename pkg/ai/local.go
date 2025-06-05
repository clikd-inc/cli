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
	defaultLocalEndpoint = "http://localhost:11434/api" // Default for Ollama
)

// LocalClient implements the Client interface for local models
type LocalClient struct {
	baseURL    string
	httpClient *http.Client
	modelName  string
	config     ModelConfig
}

// localChatRequest represents a request to the local chat API (Ollama format)
type localChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages,omitempty"`
	Prompt   string                 `json:"prompt,omitempty"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// localChatResponse represents a response from the local chat API
type localChatResponse struct {
	Model           string  `json:"model"`
	CreatedAt       string  `json:"created_at"`
	Message         Message `json:"message"`
	Done            bool    `json:"done"`
	TotalDuration   int64   `json:"total_duration,omitempty"`
	LoadDuration    int64   `json:"load_duration,omitempty"`
	PromptEvalCount int     `json:"prompt_eval_count,omitempty"`
	EvalCount       int     `json:"eval_count,omitempty"`
}

// localCompletionRequest represents a request to the local completion API
type localCompletionRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// localCompletionResponse represents a response from the local completion API
type localCompletionResponse struct {
	Model           string `json:"model"`
	CreatedAt       string `json:"created_at"`
	Response        string `json:"response"`
	Done            bool   `json:"done"`
	TotalDuration   int64  `json:"total_duration,omitempty"`
	LoadDuration    int64  `json:"load_duration,omitempty"`
	PromptEvalCount int    `json:"prompt_eval_count,omitempty"`
	EvalCount       int    `json:"eval_count,omitempty"`
}

// NewLocalClient creates a new client for local models
func NewLocalClient(ctx context.Context, config ModelConfig) (Client, error) {
	baseURL := defaultLocalEndpoint
	if config.Endpoint != "" {
		baseURL = config.Endpoint
	}

	return &LocalClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		modelName:  config.ModelID,
		config:     config,
	}, nil
}

// Complete implements the Client interface
func (c *LocalClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Create options map
	options := map[string]interface{}{
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}

	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}

	// Create local-specific request
	localReq := localCompletionRequest{
		Model:   c.modelName,
		Prompt:  req.Prompt,
		Stream:  req.Stream,
		Options: options,
	}

	// Serialize request
	reqBody, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/generate", c.baseURL),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

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
	var localResp localCompletionResponse
	if err := json.Unmarshal(respBody, &localResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to CompletionResponse
	return &CompletionResponse{
		Text:         localResp.Response,
		FinishReason: "stop", // Local models typically don't provide this
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     localResp.PromptEvalCount,
			CompletionTokens: localResp.EvalCount,
			TotalTokens:      localResp.PromptEvalCount + localResp.EvalCount,
		},
	}, nil
}

// Chat implements the Client interface
func (c *LocalClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
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

	// Create options map
	optionsMap := map[string]interface{}{
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}

	if req.MaxTokens > 0 {
		optionsMap["num_predict"] = req.MaxTokens
	}

	// Create local-specific request
	localReq := localChatRequest{
		Model:    c.modelName,
		Messages: messages,
		Stream:   req.Stream,
		Options:  optionsMap,
	}

	// Serialize request
	reqBody, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/chat", c.baseURL),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

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
	var localResp localChatResponse
	if err := json.Unmarshal(respBody, &localResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to CompletionResponse
	return &CompletionResponse{
		Text:         localResp.Message.Content,
		FinishReason: "stop", // Local models typically don't provide this
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     localResp.PromptEvalCount,
			CompletionTokens: localResp.EvalCount,
			TotalTokens:      localResp.PromptEvalCount + localResp.EvalCount,
		},
	}, nil
}

// GetProvider implements the Client interface
func (c *LocalClient) GetProvider() Provider {
	return ProviderLocal
}

// GetModelName implements the Client interface
func (c *LocalClient) GetModelName() string {
	return c.modelName
}

// GetCapabilities implements the Client interface
func (c *LocalClient) GetCapabilities() ModelCapabilities {
	return ModelCapabilities{
		SupportsChat:         true,
		SupportsCompletion:   true,
		SupportsStream:       true,
		SupportsJSONResponse: false, // Most local models don't support JSON mode
		MaxContextWindow:     c.config.ContextWindow,
	}
}
