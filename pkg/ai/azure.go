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

// AzureOpenAIClient implements the Client interface for Azure OpenAI
type AzureOpenAIClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	modelName  string
	config     ModelConfig
	apiVersion string
}

// NewAzureOpenAIClient creates a new Azure OpenAI client
func NewAzureOpenAIClient(ctx context.Context, config ModelConfig) (Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Azure OpenAI API key is required")
	}

	if config.Endpoint == "" {
		return nil, fmt.Errorf("Azure OpenAI endpoint is required")
	}

	return &AzureOpenAIClient{
		apiKey:     config.APIKey,
		endpoint:   config.Endpoint,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		modelName:  config.ModelID,
		config:     config,
		apiVersion: "2023-05-15", // Use a sensible default, could be made configurable
	}, nil
}

// Complete implements the Client interface
func (c *AzureOpenAIClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Azure OpenAI's API is similar to OpenAI but with different endpoint structure
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

	// Azure endpoint format: {endpoint}/openai/deployments/{deployment-id}/completions?api-version={api-version}
	url := fmt.Sprintf("%s/openai/deployments/%s/completions?api-version=%s",
		c.endpoint, c.modelName, c.apiVersion)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("api-key", c.apiKey)

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
func (c *AzureOpenAIClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
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

	// Create Azure-specific request (same format as OpenAI)
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

	// Azure endpoint format: {endpoint}/openai/deployments/{deployment-id}/chat/completions?api-version={api-version}
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		c.endpoint, c.modelName, c.apiVersion)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("api-key", c.apiKey)

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
func (c *AzureOpenAIClient) GetProvider() Provider {
	return ProviderAzureOpenAI
}

// GetModelName implements the Client interface
func (c *AzureOpenAIClient) GetModelName() string {
	return c.modelName
}

// GetCapabilities implements the Client interface
func (c *AzureOpenAIClient) GetCapabilities() ModelCapabilities {
	return ModelCapabilities{
		SupportsChat:         true,
		SupportsCompletion:   true,
		SupportsStream:       true,
		SupportsJSONResponse: true,
		MaxContextWindow:     c.config.ContextWindow,
	}
}
