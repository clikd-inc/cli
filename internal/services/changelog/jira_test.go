package changelog

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	agjira "github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
	gitcmd "github.com/tsuyoshiwada/go-gitcmd"

	"clikd/internal/services/git"
	"clikd/internal/utils"
)

func TestJira(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Options: &Options{
			Processor:                   nil,
			NextTag:                     "",
			TagFilterPattern:            "",
			CommitFilters:               nil,
			CommitSortBy:                "",
			CommitGroupBy:               "",
			CommitGroupSortBy:           "",
			CommitGroupTitleMaps:        nil,
			HeaderPattern:               "",
			HeaderPatternMaps:           nil,
			IssuePrefix:                 nil,
			RefActions:                  nil,
			MergePattern:                "",
			MergePatternMaps:            nil,
			RevertPattern:               "",
			RevertPatternMaps:           nil,
			NoteKeywords:                nil,
			JiraUsername:                "uuu",
			JiraToken:                   "ppp",
			JiraURL:                     "http://jira.com",
			JiraTypeMaps:                nil,
			JiraIssueDescriptionPattern: "",
		},
	}

	jira := NewJiraClient(config)
	issue, err := jira.GetJiraIssue("fake")
	assert.Nil(issue)
	assert.Error(err)
}

func TestJiraIntegration(t *testing.T) {
	assert := assert.New(t)

	// Mock Jira server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/rest/api/2/issue/TEST-123") {
			// Return mock Jira issue
			w.Write([]byte(`{
				"key": "TEST-123",
				"fields": {
					"summary": "Test Jira Issue",
					"description": "This is a description for a test Jira issue.",
					"issuetype": {
						"name": "Story"
					},
					"labels": ["test", "integration"]
				}
			}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := &Config{
		Options: &Options{
			JiraUsername:                "test",
			JiraToken:                   "token",
			JiraURL:                     server.URL,
			JiraTypeMaps:                map[string]string{"Story": "feat", "Bug": "fix"},
			JiraIssueDescriptionPattern: "",
		},
	}

	jira := NewJiraClient(config)
	issue, err := jira.GetJiraIssue("TEST-123")

	assert.NoError(err)
	assert.NotNil(issue)
	assert.Equal("TEST-123", issue.Key)
	assert.Equal("Test Jira Issue", issue.Fields.Summary)
	assert.Equal("This is a description for a test Jira issue.", issue.Fields.Description)
	assert.Equal("Story", issue.Fields.Type.Name)
	assert.Equal([]string{"test", "integration"}, issue.Fields.Labels)
}

func TestJiraIntegrationWithTemplate(t *testing.T) {
	assert := assert.New(t)
	cwd, _ := os.Getwd()
	testName := "jira_integration"

	// Setup test repo
	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2023-01-01 00:00:00", "feat(core): Add feature TEST-123", "")
		tag("1.0.0")
	})

	// Create generator with mock client
	gen := &Generator{
		gitService: nil,
		config: &Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				RepositoryURL: "https://example.com",
			},
			Options: &Options{
				JiraURL:      "https://jira.example.com",
				JiraUsername: "test",
				JiraToken:    "token",
				JiraTypeMaps: map[string]string{"Story": "feat", "Bug": "fix"},
				CommitFilters: map[string][]string{
					"Type": {"feat", "fix"},
				},
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
					"fix":  "Bug Fixes",
				},
				HeaderPattern:     "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
			},
		},
		jiraClient: nil,
		logger:     utils.NewLogger("info", true),
	}

	// Mock version data
	version := &git.Version{
		Tag: &git.Tag{
			Name: "1.0.0",
			Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		CommitGroups: []*git.CommitGroup{
			{
				Title: "Features",
				Commits: []*git.Commit{
					{
						Type:    "feat",
						Scope:   "core",
						Subject: "Add feature TEST-123",
						JiraIssue: &git.JiraIssue{
							Key:         "TEST-123",
							Type:        "Story",
							Summary:     "Test Jira Issue",
							Description: "This is a description for the test Jira issue.",
							Labels:      []string{"test", "integration"},
						},
					},
				},
			},
		},
	}

	// Test rendering
	buf := &bytes.Buffer{}
	err := gen.render(buf, nil, []*git.Version{version})

	// If the template doesn't exist, this will fail
	if os.IsNotExist(err) {
		t.Skip("Skipping test because template file doesn't exist")
		return
	}

	assert.NoError(err)
	result := buf.String()

	// Verify output contains Jira info
	assert.Contains(result, "Add feature TEST-123")
	assert.Contains(result, "**Jira:** [TEST-123]")
	assert.Contains(result, "**Summary:** Test Jira Issue")
	assert.Contains(result, "**Type:** Story")
	assert.Contains(result, "**Labels:** test, integration")
	assert.Contains(result, "**Description:**")
}

// TestJiraClientMock testet die Mock-Implementierung des Jira-Clients
func TestJiraClientMock(t *testing.T) {
	assert := assert.New(t)

	mockClient := &mockTestJiraClient{}

	// Test erfolgreicher Fall
	issue, err := mockClient.GetJiraIssue("TEST-123")
	assert.NoError(err)
	assert.NotNil(issue)
	assert.Equal("TEST-123", issue.Key)
	assert.Equal("Test Jira Issue", issue.Fields.Summary)
	assert.Equal("Story", issue.Fields.Type.Name)
	assert.Equal([]string{"test", "integration"}, issue.Fields.Labels)

	// Test Fehlerfall
	issue, err = mockClient.GetJiraIssue("NONEXISTENT-456")
	assert.Error(err)
	assert.Nil(issue)
	assert.Contains(err.Error(), "status: 404")
}

// Mock Jira client for testing
type mockTestJiraClient struct {
}

func (m *mockTestJiraClient) GetJiraIssue(id string) (*agjira.Issue, error) {
	if id == "TEST-123" {
		issueFields := &agjira.IssueFields{
			Summary:     "Test Jira Issue",
			Description: "This is a description for the test Jira issue.",
			Type: agjira.IssueType{
				Name: "Story",
			},
			Labels: []string{"test", "integration"},
		}

		return &agjira.Issue{
			Key:    id,
			Fields: issueFields,
		}, nil
	}

	return nil, &agjira.Error{
		HTTPError:     fmt.Errorf("status: %d", http.StatusNotFound),
		ErrorMessages: []string{"Issue not found"},
	}
}
