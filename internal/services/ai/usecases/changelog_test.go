package usecases

import (
	"context"
	"testing"
)

// Mock client for testing
type mockClient struct {
	response string
	err      error
}

func (m *mockClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CompletionResponse{Text: m.response}, nil
}

func (m *mockClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CompletionResponse{Text: m.response}, nil
}

func TestEnhanceCommitMessagesBatch(t *testing.T) {
	tests := []struct {
		name           string
		commitMessages []string
		mockResponse   string
		expectedCount  int
		expectedFirst  string
	}{
		{
			name:           "simple commit message",
			commitMessages: []string{"fix(api): resolve timeout issue in authentication service"},
			mockResponse:   "fix(api): resolve timeout issue in authentication service\n---",
			expectedCount:  1,
			expectedFirst:  "fix(api): resolve timeout issue in authentication service",
		},
		{
			name:           "complex commit message",
			commitMessages: []string{"feat(auth): add login endpoint and fix validation bugs and update tests"},
			mockResponse:   "feat(auth): add login endpoint\nfix(auth): fix validation bugs\ntest(auth): update authentication tests\n---",
			expectedCount:  3,
			expectedFirst:  "feat(auth): add login endpoint",
		},
		{
			name:           "multiple commit messages",
			commitMessages: []string{"fix(api): resolve timeout issue in authentication service", "feat(ui): add new dashboard component with responsive design"},
			mockResponse:   "fix(api): resolve timeout issue in authentication service\n---\nfeat(ui): add new dashboard component with responsive design\n---",
			expectedCount:  1, // Testing first message only
			expectedFirst:  "fix(api): resolve timeout issue in authentication service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockClient{response: tt.mockResponse}
			ctx := context.Background()
			options := EnhanceChangelogOptions{
				MaxTokens:   1000,
				Temperature: 0.3,
				TopP:        0.9,
			}

			result, err := EnhanceCommitMessagesBatch(mockClient, ctx, tt.commitMessages, options)
			if err != nil {
				t.Errorf("EnhanceCommitMessagesBatch() returned error: %v", err)
				return
			}

			// Test the first commit message result
			firstMessage := tt.commitMessages[0]
			if messages, exists := result[firstMessage]; exists {
				if len(messages) != tt.expectedCount {
					t.Errorf("Expected %d messages for first commit, got %d", tt.expectedCount, len(messages))
					return
				}

				if messages[0] != tt.expectedFirst {
					t.Errorf("Expected first message '%s', got '%s'", tt.expectedFirst, messages[0])
				}
			} else {
				t.Errorf("Expected result for commit message '%s' not found", firstMessage)
			}
		})
	}
}
