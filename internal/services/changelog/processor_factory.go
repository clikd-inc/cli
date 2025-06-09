package changelog

import (
	"fmt"
	"net/url"
	"strings"
)

// ProcessorFactory erstellt Prozessoren für die Changelog-Generierung
type ProcessorFactory struct {
	hostRegistry map[string]string
}

// NewProcessorFactory erstellt eine neue ProcessorFactory-Instanz
func NewProcessorFactory() *ProcessorFactory {
	return &ProcessorFactory{
		hostRegistry: map[string]string{
			"github":    "github.com",
			"gitlab":    "gitlab.com",
			"bitbucket": "bitbucket.org",
		},
	}
}

// Create erstellt einen Prozessor basierend auf der Konfiguration
func (factory *ProcessorFactory) Create(config *Config) (Processor, error) {
	if config.Info.RepositoryURL == "" {
		return nil, nil
	}

	obj, err := url.Parse(config.Info.RepositoryURL)
	if err != nil {
		return nil, err
	}

	host := obj.Host

	// Wähle den passenden Prozessor basierend auf dem Host
	hostURL := fmt.Sprintf("%s://%s", obj.Scheme, obj.Host)

	switch {
	case strings.Contains(host, "github.com"):
		return &GitHubProcessorAdapter{
			ProcessorAdapter: ProcessorAdapter{
				Host:   hostURL,
				config: config,
			},
		}, nil
	case strings.Contains(host, "gitlab.com"):
		return &GitLabProcessorAdapter{
			ProcessorAdapter: ProcessorAdapter{
				Host:   hostURL,
				config: config,
			},
		}, nil
	case strings.Contains(host, "bitbucket.org"):
		return &BitbucketProcessorAdapter{
			ProcessorAdapter: ProcessorAdapter{
				Host:   hostURL,
				config: config,
			},
		}, nil
	default:
		return nil, nil
	}
}

// ProcessorAdapter ist eine Basisstruktur für Prozessor-Adapter
type ProcessorAdapter struct {
	Host   string
	config *Config
}

// GitHubProcessorAdapter adaptiert den GitHubProcessor für die Processor-Schnittstelle
type GitHubProcessorAdapter struct {
	ProcessorAdapter
}

// Bootstrap implementiert die Processor-Schnittstelle
func (p *GitHubProcessorAdapter) Bootstrap(config *Config) error {
	// Konfiguration für GitHub spezifisch anpassen
	return nil
}

// ProcessCommit implementiert die Processor-Schnittstelle
func (p *GitHubProcessorAdapter) ProcessCommit(commit *ChangelogCommit) *ChangelogCommit {
	// GitHub-spezifische Verarbeitung des Commits
	return commit
}

// GitLabProcessorAdapter adaptiert den GitLabProcessor für die Processor-Schnittstelle
type GitLabProcessorAdapter struct {
	ProcessorAdapter
}

// Bootstrap implementiert die Processor-Schnittstelle
func (p *GitLabProcessorAdapter) Bootstrap(config *Config) error {
	// Konfiguration für GitLab spezifisch anpassen
	return nil
}

// ProcessCommit implementiert die Processor-Schnittstelle
func (p *GitLabProcessorAdapter) ProcessCommit(commit *ChangelogCommit) *ChangelogCommit {
	// GitLab-spezifische Verarbeitung des Commits
	return commit
}

// BitbucketProcessorAdapter adaptiert den BitbucketProcessor für die Processor-Schnittstelle
type BitbucketProcessorAdapter struct {
	ProcessorAdapter
}

// Bootstrap implementiert die Processor-Schnittstelle
func (p *BitbucketProcessorAdapter) Bootstrap(config *Config) error {
	// Konfiguration für Bitbucket spezifisch anpassen
	return nil
}

// ProcessCommit implementiert die Processor-Schnittstelle
func (p *BitbucketProcessorAdapter) ProcessCommit(commit *ChangelogCommit) *ChangelogCommit {
	// Bitbucket-spezifische Verarbeitung des Commits
	return commit
}
