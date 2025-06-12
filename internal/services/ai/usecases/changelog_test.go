package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CompletionResponse), args.Error(1)
}

func (m *MockClient) Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error) {
	args := m.Called(ctx, messages, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CompletionResponse), args.Error(1)
}

func TestWithJSONResponse(t *testing.T) {
	req := &CompletionRequest{}

	option := WithJSONResponse()
	option(req)

	assert.Equal(t, "json", req.ResponseType)
}

func TestEnhanceChangelog(t *testing.T) {
	tests := []struct {
		name           string
		changelog      string
		mockResponse   *CompletionResponse
		mockError      error
		expectedResult string
		expectedError  string
	}{
		{
			name:      "successful enhancement",
			changelog: "# Changelog\n\n## v1.0.0\n- Added feature X",
			mockResponse: &CompletionResponse{
				Text:         "# Changelog\n\n## v1.0.0\n- **Added**: Feature X with improved functionality",
				FinishReason: "stop",
				Usage: struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				}{
					PromptTokens:     50,
					CompletionTokens: 30,
					TotalTokens:      80,
				},
			},
			mockError:      nil,
			expectedResult: "# Changelog\n\n## v1.0.0\n- **Added**: Feature X with improved functionality",
			expectedError:  "",
		},
		{
			name:           "client error returns original changelog",
			changelog:      "# Original Changelog\n\n## v1.0.0\n- Basic feature",
			mockResponse:   nil,
			mockError:      errors.New("API rate limit exceeded"),
			expectedResult: "# Original Changelog\n\n## v1.0.0\n- Basic feature",
			expectedError:  "failed to enhance changelog: API rate limit exceeded",
		},
		{
			name:      "empty changelog",
			changelog: "",
			mockResponse: &CompletionResponse{
				Text:         "# Changelog\n\nNo changes recorded.",
				FinishReason: "stop",
			},
			mockError:      nil,
			expectedResult: "# Changelog\n\nNo changes recorded.",
			expectedError:  "",
		},
		{
			name:      "changelog with special characters",
			changelog: "# Changelog\n\n## v1.0.0\n- Fixed bug with $special & characters",
			mockResponse: &CompletionResponse{
				Text:         "# Changelog\n\n## v1.0.0\n- **Fixed**: Bug with $special & characters resolved",
				FinishReason: "stop",
			},
			mockError:      nil,
			expectedResult: "# Changelog\n\n## v1.0.0\n- **Fixed**: Bug with $special & characters resolved",
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			ctx := context.Background()

			// Set up mock expectations
			mockClient.On("Complete", ctx, mock.MatchedBy(func(req *CompletionRequest) bool {
				// Verify the request contains our changelog
				return req.MaxTokens == 1024 &&
					req.Temperature == 0.7 &&
					req.TopP == 0.9 &&
					len(req.Prompt) > 0
			})).Return(tt.mockResponse, tt.mockError)

			// Call the function
			result, err := EnhanceChangelog(mockClient, ctx, tt.changelog)

			// Verify results
			assert.Equal(t, tt.expectedResult, result)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEnhanceChangelogPromptFormat(t *testing.T) {
	mockClient := new(MockClient)
	ctx := context.Background()
	changelog := "Test changelog content"

	mockClient.On("Complete", ctx, mock.MatchedBy(func(req *CompletionRequest) bool {
		// Verify the prompt contains the expected structure
		expectedPromptStart := "You are an expert in writing clear, concise, and informative changelogs."
		expectedChangelogSection := "CHANGELOG:\nTest changelog content"
		expectedEndSection := "ENHANCED CHANGELOG:"

		return req.MaxTokens == 1024 &&
			req.Temperature == 0.7 &&
			req.TopP == 0.9 &&
			len(req.Prompt) > 0 &&
			// Check that prompt contains expected sections
			assert.Contains(t, req.Prompt, expectedPromptStart) &&
			assert.Contains(t, req.Prompt, expectedChangelogSection) &&
			assert.Contains(t, req.Prompt, expectedEndSection)
	})).Return(&CompletionResponse{
		Text: "Enhanced changelog",
	}, nil)

	_, err := EnhanceChangelog(mockClient, ctx, changelog)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Test message content",
	}

	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Test message content", msg.Content)
}

func TestCompletionRequest(t *testing.T) {
	req := CompletionRequest{
		Prompt:       "Test prompt",
		MaxTokens:    100,
		Temperature:  0.5,
		TopP:         0.8,
		Stop:         []string{"END"},
		Stream:       true,
		ModelName:    "test-model",
		ResponseType: "json",
	}

	assert.Equal(t, "Test prompt", req.Prompt)
	assert.Equal(t, 100, req.MaxTokens)
	assert.Equal(t, 0.5, req.Temperature)
	assert.Equal(t, 0.8, req.TopP)
	assert.Equal(t, []string{"END"}, req.Stop)
	assert.True(t, req.Stream)
	assert.Equal(t, "test-model", req.ModelName)
	assert.Equal(t, "json", req.ResponseType)
}

func TestCompletionResponse(t *testing.T) {
	resp := CompletionResponse{
		Text:         "Test response",
		FinishReason: "stop",
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	assert.Equal(t, "Test response", resp.Text)
	assert.Equal(t, "stop", resp.FinishReason)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 20, resp.Usage.CompletionTokens)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}
