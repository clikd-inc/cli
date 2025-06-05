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
	openaiAPIURL = "https://api.openai.com/v1"
)

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	modelName  string
	config     ModelConfig
}

// openaiChatRequest represents a request to the OpenAI chat API
type openaiChatRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Temperature    float64   `json:"temperature,omitempty"`
	TopP           float64   `json:"top_p,omitempty"`
	MaxTokens      int       `json:"max_tokens,omitempty"`
	Stream         bool      `json:"stream,omitempty"`
	Stop           []string  `json:"stop,omitempty"`
	Tools          []any     `json:"tools,omitempty"`
	ToolChoice     any       `json:"tool_choice,omitempty"`
	User           string    `json:"user,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

// openaiChatResponse represents a response from the OpenAI chat API
type openaiChatResponse struct {
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

// openaiCompletionRequest represents a request to the OpenAI completion API
type openaiCompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	User        string   `json:"user,omitempty"`
}

// openaiCompletionResponse represents a response from the OpenAI completion API
type openaiCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(ctx context.Context, config ModelConfig) (Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := openaiAPIURL
	if config.Endpoint != "" {
		baseURL = config.Endpoint
	}

	return &OpenAIClient{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		modelName:  config.ModelID,
		config:     config,
	}, nil
}

// Complete implements the Client interface
func (c *OpenAIClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Create OpenAI-specific request
	openaiReq := openaiCompletionRequest{
		Model:       c.modelName,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Stop:        req.Stop,
	}

	// Serialize request
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/completions", c.baseURL),
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
	var openaiResp openaiCompletionResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if choices are empty
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Convert to CompletionResponse
	return &CompletionResponse{
		Text:         openaiResp.Choices[0].Text,
		FinishReason: openaiResp.Choices[0].FinishReason,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// Chat implements the Client interface
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
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

	// Create OpenAI-specific request
	openaiReq := openaiChatRequest{
		Model:       c.modelName,
		Messages:    messages,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Stop:        req.Stop,
	}

	// Check if JSON response is requested
	if req.ResponseType == "json" {
		openaiReq.ResponseFormat = &struct {
			Type string `json:"type"`
		}{
			Type: "json_object",
		}
	}

	// Serialize request
	reqBody, err := json.Marshal(openaiReq)
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
	var openaiResp openaiChatResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if choices are empty
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Convert to CompletionResponse
	return &CompletionResponse{
		Text:         openaiResp.Choices[0].Message.Content,
		FinishReason: openaiResp.Choices[0].FinishReason,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// GetProvider implements the Client interface
func (c *OpenAIClient) GetProvider() Provider {
	return ProviderOpenAI
}

// GetModelName implements the Client interface
func (c *OpenAIClient) GetModelName() string {
	return c.modelName
}

// GetCapabilities implements the Client interface
func (c *OpenAIClient) GetCapabilities() ModelCapabilities {
	return ModelCapabilities{
		SupportsChat:         true,
		SupportsCompletion:   true,
		SupportsStream:       true,
		SupportsJSONResponse: true,
		MaxContextWindow:     c.config.ContextWindow,
	}
}
