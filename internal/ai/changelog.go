package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

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

// ChangelogService provides AI-powered features for changelog generation
type ChangelogService struct {
	client Client
	config *Config
}

// NewChangelogService creates a new changelog service
func NewChangelogService(ctx context.Context, config *Config, modelName string) (*ChangelogService, error) {
	if !config.EnableAI {
		return nil, fmt.Errorf("AI is disabled in configuration")
	}

	client, err := NewClient(ctx, config, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &ChangelogService{
		client: client,
		config: config,
	}, nil
}

// CategorizeCommit uses AI to categorize a commit message
func (s *ChangelogService) CategorizeCommit(ctx context.Context, message string) (CommitCategory, error) {
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

	resp, err := s.client.Chat(ctx, messages)
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
func (s *ChangelogService) EnhanceCommitMessage(ctx context.Context, message string) (string, error) {
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

	resp, err := s.client.Chat(ctx, messages)
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
func (s *ChangelogService) ExtractCommitInfo(ctx context.Context, message, author, date, hash string) (*CommitInfo, error) {
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

	resp, err := s.client.Chat(ctx, messages, WithJSONResponse())
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

	// Add the original fields
	info.Message = message
	info.Author = author
	info.Date = date
	info.Hash = hash

	return &info, nil
}

// GenerateSummary generates a summary for a set of commits
func (s *ChangelogService) GenerateSummary(ctx context.Context, commits []CommitInfo) (string, error) {
	if len(commits) == 0 {
		return "No changes in this release.", nil
	}

	// Format commits for the prompt
	commitStr := ""
	for i, commit := range commits {
		if i > 20 {
			commitStr += fmt.Sprintf("%d more commits...\n", len(commits)-20)
			break
		}
		commitStr += fmt.Sprintf("- [%s] %s\n", commit.Category, commit.Summary)
	}

	prompt := `Generate a concise summary of these changes for a changelog:

%s

Write a brief executive summary (2-3 sentences) highlighting the most important changes in this release.
Focus on user-facing changes and major improvements.`

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that generates concise, informative changelog summaries.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompt, commitStr),
		},
	}

	resp, err := s.client.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	return strings.TrimSpace(resp.Text), nil
}

// SuggestVersionBump suggests whether a version should be bumped as major, minor, or patch
func (s *ChangelogService) SuggestVersionBump(ctx context.Context, commits []CommitInfo, currentVersion string) (string, string, error) {
	if len(commits) == 0 {
		return "patch", "No significant changes detected.", nil
	}

	// Format commits for the prompt
	commitStr := ""
	for i, commit := range commits {
		if i > 30 {
			commitStr += fmt.Sprintf("%d more commits...\n", len(commits)-30)
			break
		}
		commitStr += fmt.Sprintf("- [%s] %s\n", commit.Category, commit.Summary)
	}

	prompt := `Based on semantic versioning (MAJOR.MINOR.PATCH), recommend a version bump type for these changes:

Current version: %s

Changes:
%s

Respond with ONE of these keywords followed by a brief explanation:
- "major" - For backwards incompatible changes
- "minor" - For new features that are backwards compatible
- "patch" - For backwards compatible bug fixes

Your response format should be:
BUMP_TYPE
Brief explanation of why you recommended this bump type.`

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that recommends semantic version bumps based on changes.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompt, currentVersion, commitStr),
		},
	}

	resp, err := s.client.Chat(ctx, messages)
	if err != nil {
		return "patch", "", fmt.Errorf("failed to suggest version bump: %w", err)
	}

	lines := strings.SplitN(resp.Text, "\n", 2)
	bumpType := strings.ToLower(strings.TrimSpace(lines[0]))

	explanation := ""
	if len(lines) > 1 {
		explanation = strings.TrimSpace(lines[1])
	}

	// Validate bump type
	switch bumpType {
	case "major", "minor", "patch":
		return bumpType, explanation, nil
	default:
		return "patch", "AI couldn't determine bump type, defaulting to patch.", nil
	}
}
