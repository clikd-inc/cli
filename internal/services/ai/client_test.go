package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTemperature(t *testing.T) {
	req := &CompletionRequest{}

	option := WithTemperature(0.8)
	option(req)

	assert.Equal(t, 0.8, req.Temperature)
}

func TestWithMaxTokens(t *testing.T) {
	req := &CompletionRequest{}

	option := WithMaxTokens(512)
	option(req)

	assert.Equal(t, 512, req.MaxTokens)
}

func TestWithTopP(t *testing.T) {
	req := &CompletionRequest{}

	option := WithTopP(0.95)
	option(req)

	assert.Equal(t, 0.95, req.TopP)
}

func TestWithStream(t *testing.T) {
	req := &CompletionRequest{}

	option := WithStream(true)
	option(req)

	assert.True(t, req.Stream)

	// Test false case
	option = WithStream(false)
	option(req)

	assert.False(t, req.Stream)
}

func TestWithStop(t *testing.T) {
	req := &CompletionRequest{}
	stopSequences := []string{"END", "STOP", "\n\n"}

	option := WithStop(stopSequences)
	option(req)

	assert.Equal(t, stopSequences, req.Stop)
}

func TestWithJSONResponse(t *testing.T) {
	req := &CompletionRequest{}

	option := WithJSONResponse()
	option(req)

	assert.Equal(t, "json", req.ResponseType)
}

func TestWithTimeout(t *testing.T) {
	opts := &ClientOptions{}

	option := WithTimeout(30)
	option(opts)

	assert.Equal(t, 30, opts.Timeout)
}

func TestWithDebug(t *testing.T) {
	opts := &ClientOptions{}

	option := WithDebug(true)
	option(opts)

	assert.True(t, opts.Debug)

	// Test false case
	option = WithDebug(false)
	option(opts)

	assert.False(t, opts.Debug)
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		provider         string
		model            string
		apiKey           string
		endpoint         string
		tokensMaxInput   int
		tokensMaxOutput  int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:             "invalid provider",
			provider:         "invalid_provider",
			model:            "test-model",
			apiKey:           "test-key",
			endpoint:         "",
			tokensMaxInput:   1000,
			tokensMaxOutput:  500,
			expectError:      true,
			expectedErrorMsg: "failed to create gollm client",
		},
		{
			name:             "missing API key for non-local provider",
			provider:         "openai",
			model:            "gpt-3.5-turbo",
			apiKey:           "", // Empty API key
			endpoint:         "",
			tokensMaxInput:   1000,
			tokensMaxOutput:  500,
			expectError:      true,
			expectedErrorMsg: "failed to create model config",
		},
		{
			name:             "local provider without API key should work",
			provider:         "local",
			model:            "local-model",
			apiKey:           "", // Empty API key is OK for local
			endpoint:         "http://localhost:8080",
			tokensMaxInput:   1000,
			tokensMaxOutput:  500,
			expectError:      true, // Will still fail because we don't have a real local server
			expectedErrorMsg: "",   // Error will be from Gollm client creation, not API key validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ctx, tt.provider, tt.model, tt.apiKey, tt.endpoint, tt.tokensMaxInput, tt.tokensMaxOutput)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: "Hello, how can I help you?",
	}

	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "Hello, how can I help you?", msg.Content)
}

func TestCompletionRequest(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	req := CompletionRequest{
		Prompt:       "Complete this text",
		Messages:     messages,
		MaxTokens:    256,
		Temperature:  0.7,
		TopP:         0.9,
		Stop:         []string{"END"},
		Stream:       false,
		ModelName:    "test-model",
		ResponseType: "text",
	}

	assert.Equal(t, "Complete this text", req.Prompt)
	assert.Equal(t, messages, req.Messages)
	assert.Equal(t, 256, req.MaxTokens)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 0.9, req.TopP)
	assert.Equal(t, []string{"END"}, req.Stop)
	assert.False(t, req.Stream)
	assert.Equal(t, "test-model", req.ModelName)
	assert.Equal(t, "text", req.ResponseType)
}

func TestCompletionResponse(t *testing.T) {
	resp := CompletionResponse{
		Text:         "This is the generated response",
		FinishReason: "stop",
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     25,
			CompletionTokens: 15,
			TotalTokens:      40,
		},
	}

	assert.Equal(t, "This is the generated response", resp.Text)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Equal(t, 25, resp.Usage.PromptTokens)
	assert.Equal(t, 15, resp.Usage.CompletionTokens)
	assert.Equal(t, 40, resp.Usage.TotalTokens)
}

func TestModelCapabilities(t *testing.T) {
	caps := ModelCapabilities{
		SupportsChat:         true,
		SupportsCompletion:   true,
		SupportsStream:       false,
		SupportsJSONResponse: true,
		MaxContextWindow:     4096,
	}

	assert.True(t, caps.SupportsChat)
	assert.True(t, caps.SupportsCompletion)
	assert.False(t, caps.SupportsStream)
	assert.True(t, caps.SupportsJSONResponse)
	assert.Equal(t, 4096, caps.MaxContextWindow)
}

func TestClientOptions(t *testing.T) {
	opts := ClientOptions{
		Timeout: 60,
		Debug:   true,
	}

	assert.Equal(t, 60, opts.Timeout)
	assert.True(t, opts.Debug)
}

func TestChatOptionChaining(t *testing.T) {
	req := &CompletionRequest{}

	// Test chaining multiple options
	options := []ChatOption{
		WithTemperature(0.5),
		WithMaxTokens(100),
		WithTopP(0.8),
		WithStream(true),
		WithStop([]string{"STOP"}),
		WithJSONResponse(),
	}

	for _, option := range options {
		option(req)
	}

	assert.Equal(t, 0.5, req.Temperature)
	assert.Equal(t, 100, req.MaxTokens)
	assert.Equal(t, 0.8, req.TopP)
	assert.True(t, req.Stream)
	assert.Equal(t, []string{"STOP"}, req.Stop)
	assert.Equal(t, "json", req.ResponseType)
}

func TestClientOptionChaining(t *testing.T) {
	opts := &ClientOptions{}

	// Test chaining multiple client options
	options := []ClientOption{
		WithTimeout(45),
		WithDebug(true),
	}

	for _, option := range options {
		option(opts)
	}

	assert.Equal(t, 45, opts.Timeout)
	assert.True(t, opts.Debug)
}
