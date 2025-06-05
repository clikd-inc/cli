package ai

import (
	"context"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	EnhanceMessageResponse string
	EnhanceMessageError    error

	CategorizeCommitResponse CommitCategory
	CategorizeCommitError    error

	ExtractInfoResponse *CommitInfo
	ExtractInfoError    error

	GenerateSummaryResponse string
	GenerateSummaryError    error

	SuggestVersionResponse string
	SuggestReasonResponse  string
	SuggestVersionError    error
}

// NewMockClient creates a new mock client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		EnhanceMessageResponse:   "Enhanced commit message",
		CategorizeCommitResponse: CommitCategoryFeature,
		ExtractInfoResponse: &CommitInfo{
			Message:   "Test commit",
			Summary:   "Test summary",
			Category:  CommitCategoryFeature,
			IssueRefs: []string{"ISSUE-123"},
		},
		GenerateSummaryResponse: "Test summary of changes",
		SuggestVersionResponse:  "minor",
		SuggestReasonResponse:   "New features added",
	}
}

// EnhanceCommitMessage implements the Client interface
func (m *MockClient) EnhanceCommitMessage(ctx context.Context, message string) (string, error) {
	return m.EnhanceMessageResponse, m.EnhanceMessageError
}

// CategorizeCommit implements the Client interface
func (m *MockClient) CategorizeCommit(ctx context.Context, message string) (CommitCategory, error) {
	return m.CategorizeCommitResponse, m.CategorizeCommitError
}

// ExtractCommitInfo implements the Client interface
func (m *MockClient) ExtractCommitInfo(ctx context.Context, message, author, date, hash string) (*CommitInfo, error) {
	return m.ExtractInfoResponse, m.ExtractInfoError
}

// GenerateSummary implements the Client interface
func (m *MockClient) GenerateSummary(ctx context.Context, commits []CommitInfo) (string, error) {
	return m.GenerateSummaryResponse, m.GenerateSummaryError
}

// SuggestVersionBump implements the Client interface
func (m *MockClient) SuggestVersionBump(ctx context.Context, commits []CommitInfo, currentVersion string) (string, string, error) {
	return m.SuggestVersionResponse, m.SuggestReasonResponse, m.SuggestVersionError
}

// MockChangelogService is a mock implementation of ChangelogService for testing
type MockChangelogService struct {
	Client *MockClient
}

// NewMockChangelogService creates a new MockChangelogService
func NewMockChangelogService() *MockChangelogService {
	return &MockChangelogService{
		Client: NewMockClient(),
	}
}

// EnhanceCommitMessage implements the ChangelogService interface
func (m *MockChangelogService) EnhanceCommitMessage(ctx context.Context, message string) (string, error) {
	return m.Client.EnhanceCommitMessage(ctx, message)
}

// CategorizeCommit implements the ChangelogService interface
func (m *MockChangelogService) CategorizeCommit(ctx context.Context, message string) (CommitCategory, error) {
	return m.Client.CategorizeCommit(ctx, message)
}

// ExtractCommitInfo implements the ChangelogService interface
func (m *MockChangelogService) ExtractCommitInfo(ctx context.Context, message, author, date, hash string) (*CommitInfo, error) {
	return m.Client.ExtractCommitInfo(ctx, message, author, date, hash)
}

// GenerateSummary implements the ChangelogService interface
func (m *MockChangelogService) GenerateSummary(ctx context.Context, commits []CommitInfo) (string, error) {
	return m.Client.GenerateSummary(ctx, commits)
}

// SuggestVersionBump implements the ChangelogService interface
func (m *MockChangelogService) SuggestVersionBump(ctx context.Context, commits []CommitInfo, currentVersion string) (string, string, error) {
	return m.Client.SuggestVersionBump(ctx, commits, currentVersion)
}
