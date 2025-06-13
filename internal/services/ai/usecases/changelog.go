package usecases

import (
	"context"
	"fmt"
	"strings"
)

// Client interface needed for the changelog usecase
type Client interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Chat generates a response for a conversation
	Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error)
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Prompt       string    `json:"prompt"`
	Messages     []Message `json:"messages"`
	MaxTokens    int       `json:"max_tokens"`
	Temperature  float64   `json:"temperature"`
	TopP         float64   `json:"top_p"`
	Stop         []string  `json:"stop,omitempty"`
	Stream       bool      `json:"stream"`
	ModelName    string    `json:"model_name,omitempty"`
	ResponseType string    `json:"response_type,omitempty"` // text, json, etc.
}

// CompletionResponse represents a response from text completion
type CompletionResponse struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"`
	Usage        struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatOption represents an option for chat completion
type ChatOption func(*CompletionRequest)

// WithJSONResponse sets the response type to JSON
func WithJSONResponse() ChatOption {
	return func(r *CompletionRequest) {
		r.ResponseType = "json"
	}
}

// EnhanceChangelogOptions contains configuration for changelog enhancement
type EnhanceChangelogOptions struct {
	MaxTokens   int
	Temperature float64
	TopP        float64
}

// EnhanceCommitMessagesBatch improves multiple commit messages in a single AI call for better performance
func EnhanceCommitMessagesBatch(client Client, ctx context.Context, commitMessages []string, options EnhanceChangelogOptions) (map[string][]string, error) {
	if len(commitMessages) == 0 {
		return make(map[string][]string), nil
	}

	// Filter out very short messages and build the batch
	var validMessages []string
	var messageMap = make(map[string]bool)

	for _, msg := range commitMessages {
		if len(msg) >= 30 && !messageMap[msg] { // Skip duplicates and short messages
			validMessages = append(validMessages, msg)
			messageMap[msg] = true
		}
	}

	if len(validMessages) == 0 {
		// Return original messages if none are suitable for enhancement
		result := make(map[string][]string)
		for _, msg := range commitMessages {
			result[msg] = []string{msg}
		}
		return result, nil
	}

	// Build batch prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`You are a technical documentation expert specializing in changelog generation.

Your task: Process multiple commit messages and split complex ones into clear, individual changelog entries.

STRICT REQUIREMENTS:
1. PRESERVE ALL TECHNICAL DETAILS - do not lose any information
2. If multiple changes are described in one commit, create separate entries for each
3. Keep the original technical language and specificity
4. Each entry should be 1 clear, complete sentence
5. Maintain conventional commit format when present
6. If it's already a single clear change, return it unchanged
7. Process each commit separately and maintain the order

RESPONSE FORMAT:
For each input commit, return the enhanced entries on separate lines, followed by "---" as a separator.

INPUT COMMITS:
`)

	for i, msg := range validMessages {
		promptBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, msg))
	}

	promptBuilder.WriteString(`
EXAMPLE OUTPUT FORMAT:
feat(auth): add JWT authentication with refresh tokens
---
refactor(config): replace ConfigData with Config structure to simplify configuration and improve clarity
refactor(config): remove unnecessary conversion functions and streamline configuration handling
---

Now process the input commits:`)

	response, err := client.Complete(ctx, &CompletionRequest{
		Prompt:      promptBuilder.String(),
		MaxTokens:   options.MaxTokens * 2, // Increase token limit for batch processing
		Temperature: options.Temperature,
		TopP:        options.TopP,
	})
	if err != nil {
		// Fallback: return original messages on batch failure
		result := make(map[string][]string)
		for _, msg := range commitMessages {
			result[msg] = []string{msg}
		}
		return result, nil
	}

	// Parse batch response
	result := make(map[string][]string)
	sections := strings.Split(response.Text, "---")

	for i, msg := range validMessages {
		if i < len(sections) {
			section := strings.TrimSpace(sections[i])
			lines := strings.Split(section, "\n")
			var enhancedMessages []string

			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Remove bullet points, dashes, or numbering
				line = strings.TrimPrefix(line, "- ")
				line = strings.TrimPrefix(line, "* ")
				line = strings.TrimPrefix(line, "• ")
				// Remove numbering like "1. ", "2. ", etc.
				if len(line) > 3 && line[1] == '.' && line[2] == ' ' {
					line = line[3:]
				}
				line = strings.TrimSpace(line)

				if line != "" && len(line) > 10 {
					enhancedMessages = append(enhancedMessages, line)
				}
			}

			if len(enhancedMessages) > 0 {
				result[msg] = enhancedMessages
			} else {
				result[msg] = []string{msg}
			}
		} else {
			result[msg] = []string{msg}
		}
	}

	// Add back short messages that were skipped
	for _, msg := range commitMessages {
		if _, exists := result[msg]; !exists {
			result[msg] = []string{msg}
		}
	}

	return result, nil
}
