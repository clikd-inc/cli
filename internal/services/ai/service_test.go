package ai

import (
	"context"
	"testing"
)

func TestService_IsEnabled(t *testing.T) {
	// Test with AI enabled
	enabledConfig := &Config{
		EnableAI: true,
		Provider: ProviderMistral,
		Model:    "mistral-medium",
		APIKey:   "test-key",
	}

	enabledService := &ServiceImpl{
		isEnabled: true,
		config:    enabledConfig,
	}

	if !enabledService.IsEnabled() {
		t.Error("Expected IsEnabled() to return true for enabled service")
	}

	// Test with AI disabled
	disabledConfig := &Config{
		EnableAI: false,
	}

	disabledService := &ServiceImpl{
		isEnabled: false,
		config:    disabledConfig,
	}

	if disabledService.IsEnabled() {
		t.Error("Expected IsEnabled() to return false for disabled service")
	}
}

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
		client:    mockClient,
		isEnabled: true,
		config:    &Config{EnableAI: true},
	}

	// Test enhancing changelog
	ctx := context.Background()
	input := "## [1.0.0] - 2023-01-01\n- Fixed bugs\n- Added features"
	expected := "Enhanced changelog"

	result, err := service.EnhanceChangelog(ctx, input)
	if err != nil {
		t.Errorf("EnhanceChangelog() returned error: %v", err)
	}

	if result != expected {
		t.Errorf("EnhanceChangelog() = %q, want %q", result, expected)
	}

	// Test with disabled service
	disabledService := &ServiceImpl{
		isEnabled: false,
		config:    &Config{EnableAI: false},
	}

	result, err = disabledService.EnhanceChangelog(ctx, input)
	if err != nil {
		t.Errorf("EnhanceChangelog() with disabled service returned error: %v", err)
	}

	if result != input {
		t.Errorf("EnhanceChangelog() with disabled service should return input unchanged, got %q, want %q", result, input)
	}
}
