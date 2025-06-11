package ai

import (
	"context"
	"testing"
)

func TestService_EnhanceChangelog(t *testing.T) {
	// Create mock client
	mockClient := &MockClient{
		CompleteFunc: func(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{
				Text: "Enhanced changelog",
			}, nil
		},
	}

	// Create service with mock client
	service := &ServiceImpl{
		client: mockClient,
		config: &Config{
			Provider: ProviderMistral,
			Model:    "mistral-medium",
			APIKey:   "test-key",
		},
	}

	// Test enhancing changelog
	input := "## [1.0.0] - 2023-01-01\n- Fixed bugs\n- Added features"

	result, err := service.EnhanceChangelog(input)
	if err != nil {
		t.Errorf("EnhanceChangelog() returned error: %v", err)
	}

	// Just check that we got some result back
	if result == "" {
		t.Error("EnhanceChangelog() returned empty result")
	}
}

func TestService_CategorizeCommit(t *testing.T) {
	// Create mock client
	mockClient := &MockClient{
		CompleteFunc: func(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{
				Text: "feat",
			}, nil
		},
	}

	// Create service with mock client
	service := &ServiceImpl{
		client: mockClient,
		config: &Config{
			Provider: ProviderMistral,
			Model:    "mistral-medium",
			APIKey:   "test-key",
		},
	}

	// Test categorizing commit
	input := "add new feature for user authentication"

	result, err := service.CategorizeCommit(input)
	if err != nil {
		t.Errorf("CategorizeCommit() returned error: %v", err)
	}

	// Just check that we got some result back (could be "other" as fallback)
	if result == "" {
		t.Error("CategorizeCommit() returned empty result")
	}
}

func TestService_EnhanceCommitMessage(t *testing.T) {
	// Create mock client
	mockClient := &MockClient{
		CompleteFunc: func(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{
				Text: "feat: add user authentication system",
			}, nil
		},
	}

	// Create service with mock client
	service := &ServiceImpl{
		client: mockClient,
		config: &Config{
			Provider: ProviderMistral,
			Model:    "mistral-medium",
			APIKey:   "test-key",
		},
	}

	// Test enhancing commit message
	input := "add auth"

	result, err := service.EnhanceCommitMessage(input)
	if err != nil {
		t.Errorf("EnhanceCommitMessage() returned error: %v", err)
	}

	// Just check that we got some result back
	if result == "" {
		t.Error("EnhanceCommitMessage() returned empty result")
	}
}
