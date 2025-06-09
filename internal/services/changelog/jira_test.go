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

	"github.com/stretchr/testify/assert"
	gitcmd "github.com/tsuyoshiwada/go-gitcmd"

	"clikd/internal/services/git"
)

func TestJira(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Options: &Options{
			JiraUsername: "uuu",
			JiraToken:    "ppp",
			JiraURL:      "http://jira.com",
		},
	}

	// Create a standard JIRA client directly for testing
	jiraConfig := &JiraConfig{
		Username: config.Options.JiraUsername,
		Token:    config.Options.JiraToken,
		URL:      config.Options.JiraURL,
	}

	jira := NewStandardJiraClient(jiraConfig)
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
			JiraUsername: "test",
			JiraToken:    "token",
			JiraURL:      server.URL,
			JiraTypeMaps: map[string]string{"Story": "feat", "Bug": "fix"},
		},
	}

	// Create a standard JIRA client directly for testing
	jiraConfig := &JiraConfig{
		Username: config.Options.JiraUsername,
		Token:    config.Options.JiraToken,
		URL:      config.Options.JiraURL,
	}

	jira := NewStandardJiraClient(jiraConfig)
	issue, err := jira.GetJiraIssue("TEST-123")

	assert.NoError(err)
	assert.NotNil(issue)
	assert.Equal("TEST-123", issue.Key)
	assert.Equal("Test Jira Issue", issue.Fields.Summary)
	assert.Equal("This is a description for a test Jira issue.", issue.Fields.Description)
	assert.Equal("Story", issue.Fields.Type.Name)
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

	// Create generator
	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(cwd, testRepoRoot, testName),
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
		})

	// Mock version data
	version := &ChangelogVersion{
		Tag: &git.Tag{
			Name: "1.0.0",
			Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		CommitGroups: []*ChangelogCommitGroup{
			{
				Title: "Features",
				Commits: []*ChangelogCommit{
					{
						Commit: &git.Commit{
							Subject: "Add feature TEST-123",
							JiraIssue: &git.JiraIssue{
								Key:         "TEST-123",
								Type:        "Story",
								Summary:     "Test Jira Issue",
								Description: "This is a description for the test Jira issue.",
							},
						},
					},
				},
			},
		},
	}

	// Test rendering
	buf := &bytes.Buffer{}
	err := gen.render(buf, nil, []*ChangelogVersion{version})

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
	assert.Contains(result, "**Description:**")
}

// TestJiraClientMock testet die Mock-Implementierung des Jira-Clients
func TestJiraClientMock(t *testing.T) {
	assert := assert.New(t)

	mockClient := &JiraClientMock{}

	// Test erfolgreicher Fall
	issue, err := mockClient.FetchIssue("TEST-123")
	assert.NoError(err)
	assert.NotNil(issue)
	assert.Equal("TEST-123", issue.Key)
	assert.Equal("Mock issue description", issue.Description)
}

// TestJiraClientCreation testet die Erstellung des Jira-Clients
func TestJiraClientCreation(t *testing.T) {
	assert := assert.New(t)

	// Erstelle eine Konfiguration mit Jira-Einstellungen
	config := &Config{
		Options: &Options{
			JiraUsername: "testuser",
			JiraToken:    "testtoken",
			JiraURL:      "https://jira.example.com",
		},
	}

	// Erstelle einen Jira-Client
	jiraConfig := &JiraConfig{
		Username: config.Options.JiraUsername,
		Token:    config.Options.JiraToken,
		URL:      config.Options.JiraURL,
	}

	client := NewStandardJiraClient(jiraConfig)
	assert.NotNil(client)
}

// TestCustomJiraClient ist ein erweiterter Test mit einem benutzerdefinierten Jira-Client
func TestCustomJiraClient(t *testing.T) {
	assert := assert.New(t)

	// Erstelle einen Mock-Jira-Client
	mockClient := &CustomJiraClientMock{}

	// Teste den Mock-Client
	issue, err := mockClient.FetchIssue("CUSTOM-456")
	assert.NoError(err)
	assert.NotNil(issue)
	assert.Equal("CUSTOM-456", issue.Key)
	assert.Equal("Custom Jira Issue", issue.Summary)
}

// CustomJiraClientMock ist ein benutzerdefinierter Mock für Jira-Tests
type CustomJiraClientMock struct{}

// FetchIssue implementiert die JiraClientInterface-Schnittstelle
func (j *CustomJiraClientMock) FetchIssue(issueID string) (*git.JiraIssue, error) {
	if issueID == "CUSTOM-456" {
		return &git.JiraIssue{
			Key:         "CUSTOM-456",
			Type:        "Bug",
			Summary:     "Custom Jira Issue",
			Description: "This is a custom mock Jira issue for testing",
		}, nil
	}

	return nil, fmt.Errorf("issue %s not found", issueID)
}
