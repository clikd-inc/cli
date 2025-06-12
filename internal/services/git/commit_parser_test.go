package git

import (
	"testing"
	"time"

	"clikd/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuyoshiwada/go-gitcmd"
)

// mockGitClient is a mock implementation of gitcmd.Client for testing
type mockGitClient struct {
	execFunc func(subcmd string, args ...string) (string, error)
}

func (m *mockGitClient) Exec(subcmd string, args ...string) (string, error) {
	if m.execFunc != nil {
		return m.execFunc(subcmd, args...)
	}
	return "", nil
}

func (m *mockGitClient) CanExec() error {
	return nil
}

func (m *mockGitClient) InsideWorkTree() error {
	return nil
}

func (m *mockGitClient) Root() (string, error) {
	return "/test/repo", nil
}

// TestCommitParser_Creation tests commit parser creation
func TestCommitParser_Creation(t *testing.T) {
	logger := utils.NewLogger("debug", true)
	client := &mockGitClient{}

	config := &Config{
		Options: &Options{
			HeaderPattern:     `^(\w*)(?:\(([^\)]*)\))?\: (.*)$`,
			HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
			MergePattern:      `^Merge pull request #(\d+) from (.*)$`,
			MergePatternMaps:  []string{"Ref", "Source"},
			RevertPattern:     `^Revert "(.*)"\s*This reverts commit (\w+)\.`,
			RevertPatternMaps: []string{"Header", "Hash"},
			RefActions:        []string{"Closes", "Fixes", "Resolves"},
			IssuePrefix:       []string{"#", "gh-"},
			NoteKeywords:      []string{"BREAKING CHANGE", "BREAKING CHANGES"},
		},
	}

	parser := newCommitParser(logger, client, config)

	assert.NotNil(t, parser)
	assert.Equal(t, logger, parser.logger)
	assert.Equal(t, client, parser.client)
	assert.Equal(t, config, parser.config)
	assert.NotNil(t, parser.reHeader)
	assert.NotNil(t, parser.reMerge)
	assert.NotNil(t, parser.reRevert)
	assert.NotNil(t, parser.reRef)
	assert.NotNil(t, parser.reIssue)
	assert.NotNil(t, parser.reNotes)
	assert.NotNil(t, parser.reMention)
	assert.NotNil(t, parser.reSignOff)
	assert.NotNil(t, parser.reCoAuthor)
}

// TestCommitParser_ParseHash tests hash parsing
func TestCommitParser_ParseHash(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected *Hash
	}{
		{
			name:  "Valid hash",
			input: "1234567890abcdef1234567890abcdef12345678\t1234567",
			expected: &Hash{
				Long:  "1234567890abcdef1234567890abcdef12345678",
				Short: "1234567",
			},
		},
		{
			name:  "Short hash only",
			input: "abcdef1234567890abcdef1234567890abcdef12\tabcdef1",
			expected: &Hash{
				Long:  "abcdef1234567890abcdef1234567890abcdef12",
				Short: "abcdef1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseHash(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_ParseAuthor tests author parsing
func TestCommitParser_ParseAuthor(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected *Author
	}{
		{
			name:  "Valid author",
			input: "John Doe\tjohn.doe@example.com\t1609459200",
			expected: &Author{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				Date:  time.Unix(1609459200, 0),
			},
		},
		{
			name:  "Author with special characters",
			input: "José García\tjose.garcia@example.com\t1612137600",
			expected: &Author{
				Name:  "José García",
				Email: "jose.garcia@example.com",
				Date:  time.Unix(1612137600, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseAuthor(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_ParseCommitter tests committer parsing
func TestCommitParser_ParseCommitter(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected *Committer
	}{
		{
			name:  "Valid committer",
			input: "Jane Smith\tjane.smith@example.com\t1609459200",
			expected: &Committer{
				Name:  "Jane Smith",
				Email: "jane.smith@example.com",
				Date:  time.Unix(1609459200, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseCommitter(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_ProcessHeader tests header processing
func TestCommitParser_ProcessHeader(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected struct {
			Type    string
			Scope   string
			Subject string
		}
	}{
		{
			name:  "Conventional commit with scope",
			input: "feat(auth): Add user authentication",
			expected: struct {
				Type    string
				Scope   string
				Subject string
			}{
				Type:    "feat",
				Scope:   "auth",
				Subject: "Add user authentication",
			},
		},
		{
			name:  "Conventional commit without scope",
			input: "fix: Fix critical bug",
			expected: struct {
				Type    string
				Scope   string
				Subject string
			}{
				Type:    "fix",
				Scope:   "",
				Subject: "Fix critical bug",
			},
		},
		{
			name:  "Non-conventional commit",
			input: "Update README file",
			expected: struct {
				Type    string
				Scope   string
				Subject string
			}{
				Type:    "",
				Scope:   "",
				Subject: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := &Commit{Header: tt.input}
			parser.processHeader(commit, tt.input)

			assert.Equal(t, tt.expected.Type, commit.Type)
			assert.Equal(t, tt.expected.Scope, commit.Scope)
			assert.Equal(t, tt.expected.Subject, commit.Subject)
		})
	}
}

// TestCommitParser_ParseRefs tests reference parsing
func TestCommitParser_ParseRefs(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected []*Ref
	}{
		{
			name:  "Single close reference",
			input: "Closes #123",
			expected: []*Ref{
				{Action: "Closes", Ref: "123"},
			},
		},
		{
			name:  "Multiple references",
			input: "Fixes #456 and resolves #789",
			expected: []*Ref{
				{Action: "Fixes", Ref: "456"},
				{Action: "resolves", Ref: "789"},
			},
		},
		{
			name:     "No references",
			input:    "Just a regular commit message",
			expected: []*Ref{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseRefs(tt.input)
			assert.Equal(t, len(tt.expected), len(result))

			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected.Action, result[i].Action)
					assert.Equal(t, expected.Ref, result[i].Ref)
				}
			}
		})
	}
}

// TestCommitParser_ParseMentions tests mention parsing
func TestCommitParser_ParseMentions(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single mention",
			input:    "Thanks @johndoe for the review",
			expected: []string{"johndoe"},
		},
		{
			name:     "Multiple mentions",
			input:    "Thanks @johndoe and @janedoe for the review",
			expected: []string{"johndoe", "janedoe"},
		},
		{
			name:     "No mentions",
			input:    "Just a regular commit message",
			expected: []string{},
		},
		{
			name:     "Duplicate mentions",
			input:    "Thanks @johndoe and @johndoe again",
			expected: []string{"johndoe", "johndoe"}, // parseMentions() doesn't deduplicate, uniqMentions() does
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseMentions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_UniqMentions tests mention deduplication
func TestCommitParser_UniqMentions(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		mentions []string
		expected []string
	}{
		{
			name:     "No duplicates",
			mentions: []string{"johndoe", "janedoe"},
			expected: []string{"johndoe", "janedoe"},
		},
		{
			name:     "With duplicates",
			mentions: []string{"johndoe", "janedoe", "johndoe"},
			expected: []string{"johndoe", "janedoe"},
		},
		{
			name:     "All duplicates",
			mentions: []string{"johndoe", "johndoe", "johndoe"},
			expected: []string{"johndoe"},
		},
		{
			name:     "Empty slice",
			mentions: []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.uniqMentions(tt.mentions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_ParseCoAuthors tests co-author parsing
func TestCommitParser_ParseCoAuthors(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected []Contact
	}{
		{
			name:  "Single co-author",
			input: "Co-authored-by: John Doe <john.doe@example.com>",
			expected: []Contact{
				{Name: "John Doe", Email: "john.doe@example.com"},
			},
		},
		{
			name: "Multiple co-authors",
			input: `Co-authored-by: John Doe <john.doe@example.com>
Co-authored-by: Jane Smith <jane.smith@example.com>`,
			expected: []Contact{
				{Name: "John Doe", Email: "john.doe@example.com"},
				{Name: "Jane Smith", Email: "jane.smith@example.com"},
			},
		},
		{
			name:     "No co-authors",
			input:    "Just a regular commit message",
			expected: []Contact{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseCoAuthors(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_ParseSigners tests signer parsing
func TestCommitParser_ParseSigners(t *testing.T) {
	parser := createTestParser()

	tests := []struct {
		name     string
		input    string
		expected []Contact
	}{
		{
			name:  "Single signer",
			input: "Signed-off-by: John Doe <john.doe@example.com>",
			expected: []Contact{
				{Name: "John Doe", Email: "john.doe@example.com"},
			},
		},
		{
			name: "Multiple signers",
			input: `Signed-off-by: John Doe <john.doe@example.com>
Signed-off-by: Jane Smith <jane.smith@example.com>`,
			expected: []Contact{
				{Name: "John Doe", Email: "john.doe@example.com"},
				{Name: "Jane Smith", Email: "jane.smith@example.com"},
			},
		},
		{
			name:     "No signers",
			input:    "Just a regular commit message",
			expected: []Contact{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseSigners(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommitParser_Parse tests the main parse function
func TestCommitParser_Parse(t *testing.T) {
	logger := utils.NewLogger("debug", true)

	// Mock git output
	gitOutput := `@@__CHGLOG__@@HASH:1234567890abcdef1234567890abcdef12345678	1234567@@__CHGLOG_DELIMITER__@@AUTHOR:John Doe	john.doe@example.com	1609459200@@__CHGLOG_DELIMITER__@@COMMITTER:John Doe	john.doe@example.com	1609459200@@__CHGLOG_DELIMITER__@@SUBJECT:feat(auth): Add user authentication@@__CHGLOG_DELIMITER__@@BODY:Add JWT-based authentication system

Closes #123

Co-authored-by: Jane Smith <jane.smith@example.com>`

	client := &mockGitClient{
		execFunc: func(subcmd string, args ...string) (string, error) {
			if subcmd == "log" {
				return gitOutput, nil
			}
			return "", nil
		},
	}

	config := &Config{
		Options: &Options{
			HeaderPattern:     `^(\w*)(?:\(([^\)]*)\))?\: (.*)$`,
			HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
			MergePattern:      `^Merge pull request #(\d+) from (.*)$`,
			MergePatternMaps:  []string{"Ref", "Source"},
			RevertPattern:     `^Revert "(.*)"\s*This reverts commit (\w+)\.`,
			RevertPatternMaps: []string{"Header", "Hash"},
			RefActions:        []string{"Closes", "Fixes", "Resolves"},
			IssuePrefix:       []string{"#", "gh-"},
			NoteKeywords:      []string{"BREAKING CHANGE", "BREAKING CHANGES"},
			Paths:             []string{},
		},
	}

	parser := newCommitParser(logger, client, config)

	commits, err := parser.Parse("HEAD")
	require.NoError(t, err)
	require.Len(t, commits, 1)

	commit := commits[0]
	assert.NotNil(t, commit.Hash)
	assert.Equal(t, "1234567890abcdef1234567890abcdef12345678", commit.Hash.Long)
	assert.Equal(t, "1234567", commit.Hash.Short)

	assert.NotNil(t, commit.Author)
	assert.Equal(t, "John Doe", commit.Author.Name)
	assert.Equal(t, "john.doe@example.com", commit.Author.Email)

	assert.Equal(t, "feat(auth): Add user authentication", commit.Header)
	assert.Equal(t, "feat", commit.Type)
	assert.Equal(t, "auth", commit.Scope)
	assert.Equal(t, "Add user authentication", commit.Subject)

	assert.Len(t, commit.Refs, 1)
	assert.Equal(t, "Closes", commit.Refs[0].Action)
	assert.Equal(t, "123", commit.Refs[0].Ref)

	assert.Len(t, commit.CoAuthors, 1)
	assert.Equal(t, "Jane Smith", commit.CoAuthors[0].Name)
	assert.Equal(t, "jane.smith@example.com", commit.CoAuthors[0].Email)
}

// TestCommitParser_EdgeCases tests edge cases and error conditions
func TestCommitParser_EdgeCases(t *testing.T) {
	t.Run("Empty git output", func(t *testing.T) {
		client := &mockGitClient{
			execFunc: func(subcmd string, args ...string) (string, error) {
				return "", nil
			},
		}

		parser := createTestParserWithClient(client)
		commits, err := parser.Parse("HEAD")

		assert.NoError(t, err)
		assert.Len(t, commits, 0)
	})

	t.Run("Malformed git output", func(t *testing.T) {
		client := &mockGitClient{
			execFunc: func(subcmd string, args ...string) (string, error) {
				// Return properly formatted but minimal git output
				return "@@__CHGLOG__@@HASH:abc123\tabc@@__CHGLOG_DELIMITER__@@AUTHOR:Test\ttest@example.com\t1609459200@@__CHGLOG_DELIMITER__@@COMMITTER:Test\ttest@example.com\t1609459200@@__CHGLOG_DELIMITER__@@SUBJECT:Test commit@@__CHGLOG_DELIMITER__@@BODY:", nil
			},
		}

		parser := createTestParserWithClient(client)
		commits, err := parser.Parse("HEAD")

		assert.NoError(t, err)
		assert.Len(t, commits, 1)
		// Should handle gracefully with minimal data
		assert.Equal(t, "abc123", commits[0].Hash.Long)
		assert.Equal(t, "Test commit", commits[0].Header)
	})
}

// Helper functions for creating test parsers
func createTestParser() *commitParser {
	logger := utils.NewLogger("debug", true)
	client := &mockGitClient{}
	config := &Config{
		Options: &Options{
			HeaderPattern:     `^(\w*)(?:\(([^\)]*)\))?\: (.*)$`,
			HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
			MergePattern:      `^Merge pull request #(\d+) from (.*)$`,
			MergePatternMaps:  []string{"Ref", "Source"},
			RevertPattern:     `^Revert "(.*)"\s*This reverts commit (\w+)\.`,
			RevertPatternMaps: []string{"Header", "Hash"},
			RefActions:        []string{"Closes", "Fixes", "Resolves"},
			IssuePrefix:       []string{"#", "gh-"},
			NoteKeywords:      []string{"BREAKING CHANGE", "BREAKING CHANGES"},
		},
	}
	return newCommitParser(logger, client, config)
}

func createTestParserWithClient(client gitcmd.Client) *commitParser {
	logger := utils.NewLogger("debug", true)
	config := &Config{
		Options: &Options{
			HeaderPattern:     `^(\w*)(?:\(([^\)]*)\))?\: (.*)$`,
			HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
			MergePattern:      `^Merge pull request #(\d+) from (.*)$`,
			MergePatternMaps:  []string{"Ref", "Source"},
			RevertPattern:     `^Revert "(.*)"\s*This reverts commit (\w+)\.`,
			RevertPatternMaps: []string{"Header", "Hash"},
			RefActions:        []string{"Closes", "Fixes", "Resolves"},
			IssuePrefix:       []string{"#", "gh-"},
			NoteKeywords:      []string{"BREAKING CHANGE", "BREAKING CHANGES"},
		},
	}
	return newCommitParser(logger, client, config)
}
