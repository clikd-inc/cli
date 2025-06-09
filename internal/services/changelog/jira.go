package changelog

import (
	"clikd/internal/services/git"

	agjira "github.com/andygrunwald/go-jira"
)

// StandardJiraClient ist ein HTTP-Client für Jira
type StandardJiraClient struct {
	username string
	token    string
	url      string
}

// JiraConfig enthält die Konfiguration für den Jira-Client
type JiraConfig struct {
	Username string
	Token    string
	URL      string
}

// NewStandardJiraClient erstellt eine neue Instanz des Jira-Clients
func NewStandardJiraClient(config *JiraConfig) *StandardJiraClient {
	return &StandardJiraClient{
		username: config.Username,
		token:    config.Token,
		url:      config.URL,
	}
}

// GetJiraIssue holt ein Jira-Issue anhand seiner ID
func (jira *StandardJiraClient) GetJiraIssue(id string) (*agjira.Issue, error) {
	// Wenn keine Jira-URL konfiguriert ist, Abbrechen
	if jira.url == "" {
		return nil, nil
	}

	tp := agjira.BasicAuthTransport{
		Username: jira.username,
		Password: jira.token,
	}
	client, err := agjira.NewClient(tp.Client(), jira.url)
	if err != nil {
		return nil, err
	}
	issue, _, err := client.Issue.Get(id, nil)
	return issue, err
}

// JiraClientAdapter ist ein Adapter zwischen StandardJiraClient und JiraClientInterface
type JiraClientAdapter struct {
	client *StandardJiraClient
}

// NewJiraClient erstellt einen neuen Jira-Client für die Changelog-Generierung
func NewJiraClient(config *ChangelogConfig) *JiraClientAdapter {
	jiraConfig := &JiraConfig{
		Username: config.Options.Jira.ClintInfo.Username,
		Token:    config.Options.Jira.ClintInfo.Token,
		URL:      config.Options.Jira.ClintInfo.URL,
	}

	return &JiraClientAdapter{
		client: NewStandardJiraClient(jiraConfig),
	}
}

// FetchIssue implementiert die JiraClientInterface für JiraClientAdapter
func (j *JiraClientAdapter) FetchIssue(issueID string) (*git.JiraIssue, error) {
	issue, err := j.client.GetJiraIssue(issueID)
	if err != nil {
		return nil, err
	}

	if issue == nil {
		return nil, nil
	}

	// Konvertiere agjira.Issue zu git.JiraIssue
	return &git.JiraIssue{
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		Type:        issue.Fields.Type.Name,
	}, nil
}
