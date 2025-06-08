package changelog

import (
	agjira "github.com/andygrunwald/go-jira"
)

// JiraClient is an HTTP client for Jira
type JiraClient interface {
	GetJiraIssue(id string) (*agjira.Issue, error)
}

type jiraClient struct {
	username string
	token    string
	url      string
}

// NewJiraClient returns an instance of JiraClient
func NewJiraClient(config *Config) JiraClient {
	// Hole Jira-Informationen aus der Konfiguration
	username := config.Options.Jira.ClintInfo.Username
	token := config.Options.Jira.ClintInfo.Token
	url := config.Options.Jira.ClintInfo.URL

	return jiraClient{
		username: username,
		token:    token,
		url:      url,
	}
}

func (jira jiraClient) GetJiraIssue(id string) (*agjira.Issue, error) {
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
