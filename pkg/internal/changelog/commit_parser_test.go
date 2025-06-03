package changelog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	agjira "github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
)

func TestCommitParserParse(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)

	mock := &mockClient{
		ReturnExec: func(subcmd string, args ...string) (string, error) {
			if subcmd != "log" {
				return "", errors.New("")
			}

			bytes, _ := os.ReadFile(filepath.Join("testdata", "gitlog.txt"))

			return string(bytes), nil
		},
	}

	parser := newCommitParser(NewLogger(os.Stdout, os.Stderr, false, true),
		mock, nil, &Config{
			Options: &Options{
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
						"fix",
						"perf",
						"refactor",
					},
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
				IssuePrefix: []string{
					"#",
					"gh-",
				},
				RefActions: []string{
					"close",
					"closes",
					"closed",
					"fix",
					"fixes",
					"fixed",
					"resolve",
					"resolves",
					"resolved",
				},
				MergePattern: "^Merge pull request #(\\d+) from (.*)$",
				MergePatternMaps: []string{
					"Ref",
					"Source",
				},
				RevertPattern: "^Revert \"([\\s\\S]*)\"$",
				RevertPatternMaps: []string{
					"Header",
				},
				NoteKeywords: []string{
					"BREAKING CHANGE",
				},
			},
		})

	commits, err := parser.Parse("HEAD")
	assert.Nil(err)
	assert.Len(commits, 5)

	assert.Equal("65cf1add9735dcc4810dda3312b0792236c97c4e", commits[0].Hash.Long)
	assert.Equal("feat", commits[0].Type)
	assert.Equal("*", commits[0].Scope)
	assert.Equal("Add new feature #123", commits[0].Subject)

	assert.Equal("809a8280ffd0dadb0f4e7ba9fc835e63c37d6af6", commits[2].Hash.Long)
	assert.Equal("fix", commits[2].Type)
	assert.Equal("controller", commits[2].Scope)
	assert.Equal("Fix cors configure", commits[2].Subject)

	assert.Equal("74824d6bd1470b901ec7123d13a76a1b8938d8d0", commits[3].Hash.Long)
	assert.Equal("fix(model): Remove hoge attributes", commits[3].Header)
}

type mockJiraClient struct {
}

func (jira mockJiraClient) GetJiraIssue(id string) (*agjira.Issue, error) {
	summary := fmt.Sprintf("This is a jira task: %s", id)
	description := fmt.Sprintf("The description: %s-description", id)

	if id == "TEST-invalid" {
		return nil, fmt.Errorf("Not Found error ID: %s", id)
	}

	issueFields := agjira.IssueFields{
		Type: agjira.IssueType{
			Name: "Task",
		},
		Summary:     summary,
		Description: description,
		Labels:      []string{"label1", "label2"},
	}

	if id == "TEST-123" {
		issueFields.Type.Name = "Bug"
	}

	return &agjira.Issue{
		Key:    id,
		Fields: &issueFields,
	}, nil
}

func TestCommitParserParseWithJira(t *testing.T) {
	assert := assert.New(t)

	mock := &mockClient{
		ReturnExec: func(subcmd string, args ...string) (string, error) {
			if subcmd != "log" {
				return "", errors.New("")
			}

			bytes, _ := os.ReadFile(filepath.Join("testdata", "gitlog_jira.txt"))

			return string(bytes), nil
		},
	}

	jira := &mockJiraClient{}

	parser := newCommitParser(NewLogger(os.Stdout, os.Stderr, false, true),
		mock, jira, &Config{
			Options: &Options{
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
						"fix",
						"perf",
						"refactor",
					},
				},
				HeaderPattern: "^(?:\\[([A-Z0-9]+-[0-9]+)\\]\\s?)?(?:(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s)?(.*)$",
				HeaderPatternMaps: []string{
					"JiraIssueID",
					"Type",
					"Scope",
					"Subject",
				},
				IssuePrefix: []string{
					"#",
					"gh-",
				},
				RefActions: []string{
					"close",
					"closes",
					"closed",
					"fix",
					"fixes",
					"fixed",
					"resolve",
					"resolves",
					"resolved",
				},
				MergePattern: "^Merge pull request #(\\d+) from (.*)$",
				MergePatternMaps: []string{
					"Ref",
					"Source",
				},
				RevertPattern: "^Revert \"([\\s\\S]*)\"$",
				RevertPatternMaps: []string{
					"Header",
				},
				NoteKeywords: []string{
					"BREAKING CHANGE",
				},
			},
		})

	commits, err := parser.Parse("HEAD")
	assert.Nil(err)

	assert.Len(commits, 1)

	assert.Equal("[JIRA-1111]: Add new feature #123", commits[0].Header)
	assert.Equal("This is body message.", commits[0].Body)
}
