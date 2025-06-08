package ai

import (
	"context"

	"clikd/internal/services/ai/usecases"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	CompleteFunc    func(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	ChatFunc        func(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error)
	ProviderVal     Provider
	ModelNameVal    string
	CapabilitiesVal ModelCapabilities
}

// Complete implements the Client interface
func (m *MockClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, req)
	}
	return &CompletionResponse{
		Text: "Mock completion response",
	}, nil
}

// Chat implements the Client interface
func (m *MockClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
	if m.ChatFunc != nil {
		return m.ChatFunc(ctx, messages, options...)
	}
	return &CompletionResponse{
		Text: "Mock chat response",
	}, nil
}

// GetProvider implements the Client interface
func (m *MockClient) GetProvider() Provider {
	return m.ProviderVal
}

// GetModelName implements the Client interface
func (m *MockClient) GetModelName() string {
	return m.ModelNameVal
}

// GetCapabilities implements the Client interface
func (m *MockClient) GetCapabilities() ModelCapabilities {
	return m.CapabilitiesVal
}

// MockUsecaseClient implements the usecases.Client interface for testing
type MockUsecaseClient struct {
	CompleteFunc func(ctx context.Context, req *usecases.CompletionRequest) (*usecases.CompletionResponse, error)
	ChatFunc     func(ctx context.Context, messages []usecases.Message, options ...usecases.ChatOption) (*usecases.CompletionResponse, error)
}

// Complete implements the usecases.Client interface
func (m *MockUsecaseClient) Complete(ctx context.Context, req *usecases.CompletionRequest) (*usecases.CompletionResponse, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, req)
	}
	return &usecases.CompletionResponse{
		Text: "Mock usecase completion response",
	}, nil
}

// Chat implements the usecases.Client interface
func (m *MockUsecaseClient) Chat(ctx context.Context, messages []usecases.Message, options ...usecases.ChatOption) (*usecases.CompletionResponse, error) {
	if m.ChatFunc != nil {
		return m.ChatFunc(ctx, messages, options...)
	}
	return &usecases.CompletionResponse{
		Text: "Mock usecase chat response",
	}, nil
}
