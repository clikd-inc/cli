package changelog

import (
	"clikd/internal/services/git"
	"clikd/internal/utils"
	"io"
)

// Generator ist die Schnittstelle für den Changelog-Generator
type Generator interface {
	Generate(io.Writer, string) error
}

// generatorImpl implementiert den Changelog-Generator
type generatorImpl struct {
	config     *Config
	gitService git.Service
	jiraClient JiraClient
	logger     utils.Logger
}

// NewGenerator erstellt einen neuen Changelog-Generator
func NewGenerator(logger utils.Logger, config *Config) Generator {
	// Repository-Verzeichnis
	repoDir := config.Options.WorkingDir
	if repoDir == "" {
		repoDir = "." // Verwende das aktuelle Verzeichnis, wenn keines angegeben ist
	}

	// Git-Service erstellen
	gitService, err := git.NewServiceWithRepoDir(repoDir)
	if err != nil {
		logger.Error("Failed to create Git service: %v", err)
		return nil
	}

	// Jira-Client erstellen
	jiraClient := NewJiraClient(config)

	return &generatorImpl{
		config:     config,
		gitService: gitService,
		jiraClient: jiraClient,
		logger:     logger,
	}
}

// Generate generiert den Changelog für die angegebene Tag-Abfrage
func (g *generatorImpl) Generate(w io.Writer, query string) error {
	g.logger.Debug("Generating changelog for query: %s", query)

	// Implementierung hier...
	// Verwende den Git-Service, um Commits zu holen
	// Parse und formatiere die Commits
	// Schreibe das Ergebnis in den Writer

	// Vorläufige Implementierung - muss vervollständigt werden
	g.logger.Info("Changelog generation not fully implemented yet")
	return nil
}
