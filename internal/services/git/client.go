package git

import (
	"clikd/internal/utils"
	"fmt"
	"os"
	"strings"

	gitcmd "github.com/tsuyoshiwada/go-gitcmd"
)

// Service bietet eine zentrale Schnittstelle für alle Git-Operationen
type Service struct {
	client      gitcmd.Client
	repoDir     string
	tagReader   *tagReader
	tagSelector *tagSelector
	logger      utils.Logger
}

// ServiceOptions enthält die Konfigurationsoptionen für den Git-Service
type ServiceOptions struct {
	RepoDir string
	Logger  utils.Logger
}

// NewService erstellt einen neuen Git-Service
func NewService(repoDir string) *Service {
	return NewServiceWithOptions(ServiceOptions{
		RepoDir: repoDir,
		Logger:  utils.NewLogger("info", true).WithFields(map[string]interface{}{"module": "git"}),
	})
}

// NewServiceWithOptions erstellt einen neuen Git-Service mit benutzerdefinierten Optionen
func NewServiceWithOptions(opts ServiceOptions) *Service {
	// Fallback-Logger erstellen, wenn keiner übergeben wurde
	logger := opts.Logger
	if logger == nil {
		logger = utils.NewLogger("info", true).WithFields(map[string]interface{}{"module": "git"})
	}

	// Git-Client erstellen
	client := &customGitClient{
		wrapped: gitcmd.New(&gitcmd.Config{
			Bin: "git",
		}),
		repoDir: opts.RepoDir,
		logger:  logger,
	}

	// Tag-Reader und Selektor erstellen
	tagReader := newTagReader(client, "", "date")
	tagSelector := newTagSelector(logger)

	return &Service{
		client:      client,
		repoDir:     opts.RepoDir,
		tagReader:   tagReader,
		tagSelector: tagSelector,
		logger:      logger,
	}
}

// GetTags gibt alle Tags im Repository zurück
func (s *Service) GetTags() ([]*Tag, error) {
	s.logger.Debug("Lese alle Git-Tags")
	return s.tagReader.ReadAll()
}

// SelectTags wählt Tags basierend auf einer Abfrage aus
// Format der Abfrage:
//
//	<old>..<new> - Commits zwischen <old> und <new> Tag
//	<tag>..      - Commits vom <tag> bis zum neuesten Tag
//	..<tag>      - Commits vom ältesten Tag bis zum <tag>
//	<tag>        - Commits nur für das <tag>
func (s *Service) SelectTags(tags []*Tag, query string) ([]*Tag, string, error) {
	s.logger.Debug("Wähle Tags mit Abfrage: %s", query)
	return s.tagSelector.Select(tags, query)
}

// GetCommits holt Commits basierend auf einer Revision
func (s *Service) GetCommits(rev string) ([]*Commit, error) {
	s.logger.Debug("Hole Commits für Revision: %s", rev)

	paths := []string{}

	args := []string{
		rev,
		"--no-decorate",
		"--pretty=" + logFormat,
	}

	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}

	out, err := s.client.Exec("log", args...)
	if err != nil {
		s.logger.Error("Fehler beim Ausführen von git log: %v", err)
		return nil, err
	}

	lines := strings.Split(out, separator)
	if len(lines) > 0 {
		lines = lines[1:]
	}

	commits := make([]*Commit, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		commit := parseCommit(line)
		if commit != nil {
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

// GetCurrentBranch gibt den aktuellen Branch-Namen zurück
func (s *Service) GetCurrentBranch() (string, error) {
	s.logger.Debug("Ermittle aktuellen Branch")
	out, err := s.client.Exec("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		s.logger.Error("Fehler beim Ermitteln des aktuellen Branches: %v", err)
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ExtractCommits gruppiert Commits basierend auf den konfigurierten Optionen
// und extrahiert Merge- und Revert-Commits sowie Notizen
func (s *Service) ExtractCommits(commits []*Commit, opts *Options) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup) {
	s.logger.Debug("Extrahiere und gruppiere %d Commits", len(commits))
	extractor := newCommitExtractor(opts, s.logger)
	return extractor.Extract(commits)
}

// customGitClient ist ein Wrapper um gitcmd.Client, der alle Befehle im korrekten Repository-Verzeichnis ausführt
type customGitClient struct {
	wrapped gitcmd.Client
	repoDir string
	logger  utils.Logger
}

// Exec führt den Git-Befehl im Repository-Verzeichnis aus
func (c *customGitClient) Exec(subcmd string, args ...string) (string, error) {
	c.logger.Debug("Führe Git-Befehl aus: %s %s", subcmd, strings.Join(args, " "))

	// Ursprüngliches Arbeitsverzeichnis speichern
	origWd, err := os.Getwd()
	if err != nil {
		c.logger.Error("Fehler beim Ermitteln des aktuellen Verzeichnisses: %v", err)
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// In das Repository-Verzeichnis wechseln
	if err := os.Chdir(c.repoDir); err != nil {
		c.logger.Error("Fehler beim Wechseln in das Repository-Verzeichnis: %v", err)
		return "", fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Git-Befehl ausführen
	out, err := c.wrapped.Exec(subcmd, args...)
	if err != nil {
		c.logger.Error("Git-Befehl fehlgeschlagen: %v", err)
	}

	// Zum ursprünglichen Verzeichnis zurückkehren
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		c.logger.Error("Fehler beim Zurückkehren zum ursprünglichen Verzeichnis: %v", chdirErr)
		// Falls der Git-Befehl erfolgreich war, geben wir den Fehler beim Verzeichniswechsel nicht zurück
		if err == nil {
			err = chdirErr
		}
	}

	return out, err
}

// CanExec prüft, ob der Git-Befehl ausgeführt werden kann
func (c *customGitClient) CanExec() error {
	c.logger.Debug("Prüfe, ob Git ausgeführt werden kann")
	return c.wrapped.CanExec()
}

// InsideWorkTree prüft, ob wir uns innerhalb eines Git-Arbeitsbaumes befinden
func (c *customGitClient) InsideWorkTree() error {
	c.logger.Debug("Prüfe, ob wir uns in einem Git-Repository befinden")

	// Ursprüngliches Arbeitsverzeichnis speichern
	origWd, err := os.Getwd()
	if err != nil {
		c.logger.Error("Fehler beim Ermitteln des aktuellen Verzeichnisses: %v", err)
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// In das Repository-Verzeichnis wechseln
	if err := os.Chdir(c.repoDir); err != nil {
		c.logger.Error("Fehler beim Wechseln in das Repository-Verzeichnis: %v", err)
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Prüfen, ob wir uns in einem Git-Arbeitsbaum befinden
	result := c.wrapped.InsideWorkTree()
	if result != nil {
		c.logger.Error("Nicht in einem Git-Repository: %v", result)
	}

	// Zum ursprünglichen Verzeichnis zurückkehren
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		c.logger.Error("Fehler beim Zurückkehren zum ursprünglichen Verzeichnis: %v", chdirErr)
		// Falls die InsideWorkTree-Prüfung erfolgreich war, geben wir den Fehler beim Verzeichniswechsel zurück
		if result == nil {
			return chdirErr
		}
	}

	return result
}

// parseCommit parst eine Commit-Zeile im Log-Format
func parseCommit(input string) *Commit {
	commit := &Commit{}
	tokens := strings.Split(input, delimiter)

	for _, token := range tokens {
		firstSep := strings.Index(token, ":")
		if firstSep <= 0 || firstSep >= len(token)-1 {
			continue
		}

		field := token[0:firstSep]
		value := strings.TrimSpace(token[firstSep+1:])

		switch field {
		case hashField:
			commit.Hash = parseHash(value)
		case authorField:
			commit.Author = parseAuthor(value)
		case committerField:
			commit.Committer = parseCommitter(value)
		case subjectField:
			// Basic parsing - only set subject
			commit.Subject = value
		}
	}

	return commit
}

// parseHash parst einen Hash-Wert
func parseHash(input string) *Hash {
	arr := strings.Split(input, "\t")
	if len(arr) != 2 {
		return &Hash{}
	}

	return &Hash{
		Long:  arr[0],
		Short: arr[1],
	}
}

// parseAuthor parst einen Autor
func parseAuthor(input string) *Author {
	// Einfache Implementation
	return &Author{
		Name: input,
	}
}

// parseCommitter parst einen Committer
func parseCommitter(input string) *Committer {
	// Einfache Implementation
	return &Committer{
		Name: input,
	}
}
