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
	MaxTokens   int     // Maximum tokens for the response (from config.toml)
	Temperature float64 // Temperature for AI generation (from config.toml)
	TopP        float64 // TopP for AI generation (from config.toml)
}

// EnhanceChangelog improves a generated changelog by splitting complex commits and enhancing readability
// Deprecated: Use EnhanceChangelogWithOptions instead for configurable parameters from config.toml
// This function uses hardcoded default values and should not be used in production
func EnhanceChangelog(client Client, ctx context.Context, changelog string) (string, error) {
	return EnhanceChangelogWithOptions(client, ctx, changelog, EnhanceChangelogOptions{
		MaxTokens:   3072, // Hardcoded fallback - use EnhanceChangelogWithOptions instead
		Temperature: 0.3,  // Hardcoded fallback - use EnhanceChangelogWithOptions instead
		TopP:        0.9,  // Hardcoded fallback - use EnhanceChangelogWithOptions instead
	})
}

// EnhanceChangelogWithOptions improves a generated changelog with configurable options
func EnhanceChangelogWithOptions(client Client, ctx context.Context, changelog string, options EnhanceChangelogOptions) (string, error) {
	prompt := `You are an expert in writing clear, professional changelogs following industry standards.

Your task is to enhance the following changelog to make it more readable and professional while maintaining the exact structure and format.

CRITICAL RULES:
1. PRESERVE the exact markdown structure (headers, sections, links, version numbers, dates)
2. SPLIT complex bullet points that contain multiple changes into separate, individual bullet points
3. Each bullet point should describe ONE specific change only
4. Use clear, concise language suitable for end users
5. Remove redundant information and technical jargon
6. Start each entry with an action verb (add, fix, improve, enhance, etc.)
7. Keep entries short and scannable
8. Maintain the same categorization (Features, Bug Fixes, Code Refactoring, etc.)
9. Remove internal references like "chore:", "feat:", "fix:" prefixes from the final output
10. Focus on user-facing benefits and changes

IMPORTANT: Return ONLY the enhanced changelog content. Do not add any explanatory text, introductions, or improvement notes.

EXAMPLE OF SPLITTING:
BAD: "- **ai:** remove unused methods from Service interface to simplify code and improve maintainability chore(ai): delete corresponding tests for removed methods to keep test suite clean refactor(usecases): remove unused types and functions related to commit categorization and enhancement to streamline codebase"

GOOD: 
"- **ai:** remove unused methods from Service interface to simplify code and improve maintainability
- **ai:** delete corresponding tests for removed methods to keep test suite clean  
- **ai:** remove unused types and functions related to commit categorization to streamline codebase"

CHANGELOG TO ENHANCE:
%s`

	req := &CompletionRequest{
		Prompt:      fmt.Sprintf(prompt, changelog),
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		return changelog, fmt.Errorf("failed to enhance changelog: %w", err)
	}

	// Clean up the response to remove any potential AI-added text
	result := resp.Text

	// Remove common AI response patterns
	if strings.HasPrefix(result, "Here's the enhanced changelog") {
		// Find the start of the actual changelog content
		lines := strings.Split(result, "\n")
		var cleanLines []string
		foundStart := false

		for _, line := range lines {
			// Look for the start of actual changelog content
			if !foundStart {
				if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "<a name=") || strings.HasPrefix(line, "##") {
					foundStart = true
					cleanLines = append(cleanLines, line)
				}
			} else {
				// Skip improvement notes at the end
				if strings.HasPrefix(line, "Key improvements made:") {
					break
				}
				cleanLines = append(cleanLines, line)
			}
		}

		if len(cleanLines) > 0 {
			result = strings.Join(cleanLines, "\n")
		}
	}

	// Remove trailing improvement notes if they exist
	if idx := strings.Index(result, "Key improvements made:"); idx != -1 {
		result = result[:idx]
	}

	// Clean up any trailing whitespace
	result = strings.TrimSpace(result)

	return result, nil
}
