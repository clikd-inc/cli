package changelog

import (
	"clikd/internal/services/git"
	"clikd/internal/utils"
	"io"
)

// ChangelogGenerator ist die Schnittstelle für den Changelog-Generator
type ChangelogGenerator interface {
	Generate(logger utils.Logger, w io.Writer, query string, config *Config) error
}

// DefaultGenerator ist die Standard-Implementierung des Generators
type DefaultGenerator struct{}

// NewChangelogGenerator erstellt eine neue Generator-Instanz
func NewChangelogGenerator() ChangelogGenerator {
	return &DefaultGenerator{}
}

// Generate generiert den Changelog basierend auf der Konfiguration
func (g *DefaultGenerator) Generate(logger utils.Logger, w io.Writer, query string, config *Config) error {
	// Repository-Verzeichnis
	repoDir := config.WorkingDir
	if repoDir == "" {
		repoDir = "." // Verwende das aktuelle Verzeichnis, wenn keines angegeben ist
	}

	// Git-Service erstellen
	gitService, err := git.NewServiceWithRepoDir(repoDir)
	if err != nil {
		logger.Error("Failed to create Git service", "error", err)
		return err
	}

	// Verwende die Git-Optionen direkt aus der Konfiguration
	gitConfig := &git.Options{
		CommitGroupBy:        config.Options.CommitGroupBy,
		CommitGroupSortBy:    config.Options.CommitGroupSortBy,
		CommitGroupTitleMaps: config.Options.CommitGroupTitleMaps,
	}

	// Prozessor-Factory erstellen
	processorFactory := NewProcessorFactory()
	processor, err := processorFactory.Create(config)
	if err != nil {
		logger.Error("Failed to create processor", "error", err)
		return err
	}

	// Jira-Client erstellen für Issue-Anreicherung
	// Wenn wir den Jira-Client später benötigen, können wir ihn hier aktivieren
	// jiraClient := NewJiraClient(config)

	// Extrahiere Commits basierend auf der Query
	commits, err := gitService.GetCommits(query, config.Options.Paths)
	if err != nil {
		logger.Error("Failed to get commits", "error", err)
		return err
	}

	// Wrapper für Commits erstellen
	wrappedCommits := make([]*ChangelogCommit, len(commits))
	for i, commit := range commits {
		wrappedCommits[i] = &ChangelogCommit{Commit: commit}

		// Verarbeite Commits mit dem Processor, wenn vorhanden
		if processor != nil {
			// Wir verwenden direkt den Git-Commit, da die Processor-Schnittstelle nur mit Git-Commits arbeitet
			if p, ok := processor.(CommitProcessor); ok {
				p.ProcessCommit(wrappedCommits[i])
			}
		}
	}

	// Extrahiere Commit-Gruppen und andere Daten
	commitGroups, mergeCommits, revertCommits, _ := gitService.ExtractCommits(commits, gitConfig)

	// Wrapped Commit-Gruppen erstellen
	wrappedCommitGroups := make([]*ChangelogCommitGroup, len(commitGroups))
	for i, group := range commitGroups {
		wrappedGroup := &ChangelogCommitGroup{
			RawTitle: group.RawTitle,
			Title:    group.Title,
			Commits:  make([]*ChangelogCommit, len(group.Commits)),
		}

		for j, commit := range group.Commits {
			wrappedGroup.Commits[j] = &ChangelogCommit{Commit: commit}
		}

		wrappedCommitGroups[i] = wrappedGroup
	}

	// Wrapped Merge-Commits erstellen
	wrappedMergeCommits := make([]*ChangelogCommit, len(mergeCommits))
	for i, commit := range mergeCommits {
		wrappedMergeCommits[i] = &ChangelogCommit{Commit: commit}
	}

	// Wrapped Revert-Commits erstellen
	wrappedRevertCommits := make([]*ChangelogCommit, len(revertCommits))
	for i, commit := range revertCommits {
		wrappedRevertCommits[i] = &ChangelogCommit{Commit: commit}
	}

	// TODO: Rendern des Changelogs basierend auf den verarbeiteten Commits
	// Hier würden wir den Changelog in die io.Writer-Instanz schreiben
	// unter Verwendung der gesammelten Daten

	logger.Info("Changelog generation completed successfully")
	return nil
}

// CommitProcessor ist eine erweiterte Schnittstelle für Prozessoren,
// die einzelne Commits verarbeiten können
type CommitProcessor interface {
	Processor
}
