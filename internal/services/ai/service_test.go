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
