package usecases

import (
	"context"
	"encoding/json"
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

// CommitCategory represents a category of commit messages
type CommitCategory string

const (
	// CommitCategoryFeature represents a feature commit
	CommitCategoryFeature CommitCategory = "feature"
	// CommitCategoryFix represents a bug fix commit
	CommitCategoryFix CommitCategory = "fix"
	// CommitCategoryDocs represents a documentation commit
	CommitCategoryDocs CommitCategory = "docs"
	// CommitCategoryStyle represents a style commit
	CommitCategoryStyle CommitCategory = "style"
	// CommitCategoryRefactor represents a refactor commit
	CommitCategoryRefactor CommitCategory = "refactor"
	// CommitCategoryTest represents a test commit
	CommitCategoryTest CommitCategory = "test"
	// CommitCategoryChore represents a chore commit
	CommitCategoryChore CommitCategory = "chore"
	// CommitCategoryPerf represents a performance commit
	CommitCategoryPerf CommitCategory = "perf"
	// CommitCategoryOther represents other commit types
	CommitCategoryOther CommitCategory = "other"
)

// CommitInfo represents information about a commit
type CommitInfo struct {
	Message   string         `json:"message"`
	Author    string         `json:"author"`
	Date      string         `json:"date"`
	Hash      string         `json:"hash"`
	Category  CommitCategory `json:"category"`
	Scope     string         `json:"scope"`
	Summary   string         `json:"summary"`
	IssueRefs []string       `json:"issue_refs"`
}

// EnhanceChangelog improves a generated changelog
func EnhanceChangelog(client Client, ctx context.Context, changelog string) (string, error) {
	prompt := `You are an expert in writing clear, concise, and informative changelogs.
Please enhance the following changelog to make it more professional and readable.
Maintain the same structure and format, but improve clarity, fix grammar issues,
and ensure consistent style throughout.

CHANGELOG:
%s

ENHANCED CHANGELOG:`

	req := &CompletionRequest{
		Prompt:      fmt.Sprintf(prompt, changelog),
		MaxTokens:   1024,
		Temperature: 0.7,
		TopP:        0.9,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		return changelog, fmt.Errorf("failed to enhance changelog: %w", err)
	}

	return resp.Text, nil
}

// CategorizeCommit uses AI to categorize a commit message
func CategorizeCommit(client Client, ctx context.Context, message string) (CommitCategory, error) {
	prompt := `Analyze the following git commit message and categorize it into one of these categories:
- feat: A new feature or enhancement
- fix: A bug fix
- docs: Documentation changes
- style: Code style changes (formatting, missing semicolons, etc)
- refactor: Code refactoring without changing functionality
- test: Adding or modifying tests
- chore: Maintenance tasks, dependency updates, etc
- perf: Performance improvements
- other: None of the above

Respond ONLY with the category name, nothing else.

Commit message: %s`

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that categorizes git commit messages accurately.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompt, message),
		},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		return CommitCategoryOther, fmt.Errorf("failed to categorize commit: %w", err)
	}

	category := CommitCategory(strings.TrimSpace(strings.ToLower(resp.Text)))

	// Validate category
	switch category {
	case CommitCategoryFeature, CommitCategoryFix, CommitCategoryDocs,
		CommitCategoryStyle, CommitCategoryRefactor, CommitCategoryTest,
		CommitCategoryChore, CommitCategoryPerf:
		return category, nil
	default:
		return CommitCategoryOther, nil
	}
}

// EnhanceCommitMessage uses AI to improve a commit message for better readability
func EnhanceCommitMessage(client Client, ctx context.Context, message string) (string, error) {
	prompt := `Improve the following git commit message to make it more clear, concise, and descriptive for a changelog entry.
Maintain the original meaning but make it more readable and professional.
If the message is already well-written, you can return it unchanged.

Original commit message: %s

Improved message (keep it under 80 characters if possible):`

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that improves git commit messages for changelog entries.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompt, message),
		},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		return message, fmt.Errorf("failed to enhance commit message: %w", err)
	}

	// If the AI didn't change much, return the original
	if len(resp.Text) < 3 || strings.TrimSpace(resp.Text) == strings.TrimSpace(message) {
		return message, nil
	}

	return strings.TrimSpace(resp.Text), nil
}

// ExtractCommitInfo uses AI to extract structured information from a commit message
func ExtractCommitInfo(client Client, ctx context.Context, message, author, date, hash string) (*CommitInfo, error) {
	prompt := `Extract structured information from this git commit message:

Message: %s
Author: %s
Date: %s
Hash: %s

Return a JSON object with:
- category: The type of change (feature, fix, docs, style, refactor, test, chore, perf, or other)
- scope: The scope or module affected (if mentioned)
- summary: A clean one-line summary of the change
- issue_refs: Array of issue references (like "JIRA-123" or "#456")

Example response format:
{
  "category": "fix",
  "scope": "api",
  "summary": "Fix null pointer exception in user authentication",
  "issue_refs": ["JIRA-123", "#456"]
}

JSON response:`

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that extracts structured information from git commit messages.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompt, message, author, date, hash),
		},
	}

	resp, err := client.Chat(ctx, messages, WithJSONResponse())
	if err != nil {
		return nil, fmt.Errorf("failed to extract commit info: %w", err)
	}

	// Parse the response as JSON
	var info CommitInfo
	if err := json.Unmarshal([]byte(resp.Text), &info); err != nil {
		// Fallback: create a basic info object
		return &CommitInfo{
			Message:   message,
			Author:    author,
			Date:      date,
			Hash:      hash,
			Category:  CommitCategoryOther,
			Summary:   message,
			IssueRefs: []string{},
		}, nil
	}

	// Set the original message and metadata
	info.Message = message
	info.Author = author
	info.Date = date
	info.Hash = hash

	return &info, nil
}

// GenerateSummary uses AI to generate a summary of commits
func GenerateSummary(client Client, ctx context.Context, commits []CommitInfo) (string, error) {
	// Convert commits to a format suitable for the prompt
	commitSummaries := make([]string, len(commits))
	for i, commit := range commits {
		cat := string(commit.Category)
		if commit.Scope != "" {
			cat += "(" + commit.Scope + ")"
		}
		commitSummaries[i] = fmt.Sprintf("- %s: %s [%s]", cat, commit.Summary, commit.Hash[:7])
	}

	commitsText := strings.Join(commitSummaries, "\n")
	prompt := `Generate a concise summary of the following changes for a changelog.
Focus on the main themes and significant changes. Group related changes together.
Keep it clear, informative, and suitable for both technical and non-technical readers.

Changes:
%s

Summary (markdown format, approximately 3-5 paragraphs):`

	req := &CompletionRequest{
		Prompt:      fmt.Sprintf(prompt, commitsText),
		MaxTokens:   512,
		Temperature: 0.7,
		TopP:        0.9,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	return strings.TrimSpace(resp.Text), nil
}

// SuggestVersionBump uses AI to suggest a version bump type
func SuggestVersionBump(client Client, ctx context.Context, commits []CommitInfo, currentVersion string) (string, string, error) {
	// Convert commits to a format suitable for the prompt
	commitSummaries := make([]string, len(commits))
	for i, commit := range commits {
		cat := string(commit.Category)
		if commit.Scope != "" {
			cat += "(" + commit.Scope + ")"
		}
		commitSummaries[i] = fmt.Sprintf("- %s: %s [%s]", cat, commit.Summary, commit.Hash[:7])
	}

	commitsText := strings.Join(commitSummaries, "\n")
	prompt := `Based on the following commits, suggest a semantic version bump (major, minor, or patch) from the current version "%s".

Rules:
- major: Breaking changes (not backwards compatible)
- minor: New features (backwards compatible)
- patch: Bug fixes and minor improvements (backwards compatible)

Look for keywords in commit messages like "BREAKING CHANGE", "breaking", "feat", "feature", "fix", etc.

Commits:
%s

Respond in JSON format:
{
  "bump_type": "major|minor|patch",
  "reason": "Brief explanation of why this bump type was chosen"
}
`

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that suggests semantic version bumps based on commit messages.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompt, currentVersion, commitsText),
		},
	}

	resp, err := client.Chat(ctx, messages, WithJSONResponse())
	if err != nil {
		return "patch", "Default patch bump due to error", fmt.Errorf("failed to suggest version bump: %w", err)
	}

	// Parse the response as JSON
	var result struct {
		BumpType string `json:"bump_type"`
		Reason   string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return "patch", "Default patch bump due to parsing error", nil
	}

	// Normalize the bump type
	bumpType := strings.ToLower(strings.TrimSpace(result.BumpType))
	switch bumpType {
	case "major", "minor", "patch":
		return bumpType, result.Reason, nil
	default:
		return "patch", "Default patch bump (AI suggested invalid type)", nil
	}
}
