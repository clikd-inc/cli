package git

import (
	"clikd/internal/utils"
	"fmt"
	"os"
	"strings"

	gitcmd "github.com/tsuyoshiwada/go-gitcmd"
)

// Client definiert die Schnittstelle für Git-Operationen auf niedriger Ebene
type Client interface {
	// GetRepositoryRoot gibt den absoluten Pfad zum Root des Git-Repositories zurück
	GetRepositoryRoot() (string, error)

	// IsGitRepository prüft, ob das aktuelle Verzeichnis ein Git-Repository ist
	IsGitRepository() (bool, error)

	// GetCommits holt alle Commits im Repository basierend auf den Filteroptionen
	GetCommits(options CommitOptions) ([]*Commit, error)

	// GetCommitsWithOptions holt alle Commits mit vollständigen Pattern-Optionen
	GetCommitsWithOptions(options CommitOptions) ([]*Commit, error)

	// GetLatestTag gibt das neueste Tag im Repository zurück
	GetLatestTag() (string, error)

	// GetTags gibt alle Tags im Repository zurück
	GetTags() ([]string, error)

	// GetTagsWithPattern gibt Tags zurück, die einem bestimmten Pattern entsprechen
	GetTagsWithPattern(pattern string) ([]string, error)

	// GetCommitsBetweenTags gibt alle Commits zwischen zwei Tags zurück
	GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error)

	// GetStagedChanges gibt die gestaged Änderungen zurück
	GetStagedChanges() (string, error)

	// GetDiffBetweenTags gibt die Diff zwischen zwei Tags zurück
	GetDiffBetweenTags(fromTag, toTag string) (string, error)

	// CreateTag erstellt ein neues Tag
	CreateTag(tag string, message string) error

	// HasRemote prüft, ob das Repository einen Remote hat
	HasRemote() (bool, error)

	// GetCurrentBranch gibt den aktuellen Branch zurück
	GetCurrentBranch() (string, error)

	// Exec führt einen Git-Befehl direkt aus
	Exec(subcmd string, args ...string) (string, error)

	// CanExec prüft, ob Git ausgeführt werden kann
	CanExec() error

	// InsideWorkTree prüft, ob wir uns in einem Git-Repository befinden
	InsideWorkTree() error
}

// CommitOptions definiert die Optionen für das Abrufen von Commits
type CommitOptions struct {
	Revision string   // Git-Revision (z.B. "HEAD", "master", Tag)
	Paths    []string // Pfade, für die Commits abgerufen werden sollen
	// Pattern-Konfiguration
	HeaderPattern     string   // Ein regulärer Ausdruck für das Parsen des Commit-Headers
	HeaderPatternMaps []string // Eine Regel für die Zuordnung des Ergebnisses von `HeaderPattern` zur Eigenschaft von `Commit`
	MergePattern      string   // Ein regulärer Ausdruck für das Parsen des Merge-Commits
	MergePatternMaps  []string // Ähnlich wie `HeaderPatternMaps`
	RevertPattern     string   // Ein regulärer Ausdruck für das Parsen des Revert-Commits
	RevertPatternMaps []string // Ähnlich wie `HeaderPatternMaps`
	RefActions        []string // Wortliste von `Ref.Action`
	IssuePrefix       []string // Präfix für Issues (z.B. `#`, `gh-`)
	NoteKeywords      []string // Schlüsselwortliste zum Finden von `Note`
}

// TagSelectionOptions enthält die Optionen für die Tag-Auswahl
type TagSelectionOptions struct {
	Query string // Format: <old>..<new>, <tag>.., ..<tag>, <tag>
}

// ClientImpl ist die konkrete Implementierung des Client-Interfaces
type ClientImpl struct {
	client  gitcmd.Client
	repoDir string
	logger  utils.Logger
}

// NewClient erstellt einen neuen Git-Client
func NewClient() (Client, error) {
	// Verwende den globalen Logger anstatt hardcoded "info"
	logger := utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})

	// Git-Client erstellen
	client := gitcmd.New(&gitcmd.Config{
		Bin: "git",
	})

	// Aktuelles Arbeitsverzeichnis als Standard-Repository-Verzeichnis verwenden
	repoDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Ermitteln des aktuellen Verzeichnisses: %w", err)
	}

	return &ClientImpl{
		client:  client,
		repoDir: repoDir,
		logger:  logger,
	}, nil
}

// NewClientWithRepoDir erstellt einen neuen Git-Client für ein bestimmtes Repository-Verzeichnis
func NewClientWithRepoDir(repoDir string) (Client, error) {
	// Verwende den globalen Logger anstatt hardcoded "info"
	logger := utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})

	// Git-Client erstellen
	client := gitcmd.New(&gitcmd.Config{
		Bin: "git",
	})

	if repoDir == "" {
		// Aktuelles Arbeitsverzeichnis als Standard-Repository-Verzeichnis verwenden
		var err error
		repoDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Fehler beim Ermitteln des aktuellen Verzeichnisses: %w", err)
		}
	}

	return &ClientImpl{
		client:  client,
		repoDir: repoDir,
		logger:  logger,
	}, nil
}

// GetRepositoryRoot implementiert die Client-Schnittstelle
func (c *ClientImpl) GetRepositoryRoot() (string, error) {
	out, err := c.Exec("rev-parse", "--show-toplevel")
	if err != nil {
		c.logger.Error("Fehler beim Ermitteln des Repository-Roots: %v", err)
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsGitRepository implementiert die Client-Schnittstelle
func (c *ClientImpl) IsGitRepository() (bool, error) {
	err := c.InsideWorkTree()
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetCommits implementiert die Client-Schnittstelle
func (c *ClientImpl) GetCommits(options CommitOptions) ([]*Commit, error) {
	c.logger.Debug("Hole Commits für Revision: %s", options.Revision)

	// Verwende Standardwerte für Pattern, falls nicht angegeben
	if options.HeaderPattern == "" {
		options.HeaderPattern = "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
		options.HeaderPatternMaps = []string{"Type", "Scope", "Subject"}
	}
	if options.MergePattern == "" {
		options.MergePattern = "^Merge pull request #(\\d+) from (.*)$"
		options.MergePatternMaps = []string{"Ref", "Source"}
	}
	if options.RevertPattern == "" {
		options.RevertPattern = "^Revert \"([\\s\\S]*)\"$"
		options.RevertPatternMaps = []string{"Header"}
	}
	if len(options.RefActions) == 0 {
		options.RefActions = []string{"close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved"}
	}
	if len(options.IssuePrefix) == 0 {
		options.IssuePrefix = []string{"#"}
	}
	if len(options.NoteKeywords) == 0 {
		options.NoteKeywords = []string{"BREAKING CHANGE"}
	}

	return c.GetCommitsWithOptions(options)
}

// GetCommitsWithOptions implementiert die Client-Schnittstelle
func (c *ClientImpl) GetCommitsWithOptions(options CommitOptions) ([]*Commit, error) {
	c.logger.Debug("Hole Commits für Revision: %s", options.Revision)

	// Verwende den vollständigen commitParser anstelle der einfachen parseCommit Funktion
	config := &Config{
		Options: &Options{
			HeaderPattern:     options.HeaderPattern,
			HeaderPatternMaps: options.HeaderPatternMaps,
			MergePattern:      options.MergePattern,
			MergePatternMaps:  options.MergePatternMaps,
			RevertPattern:     options.RevertPattern,
			RevertPatternMaps: options.RevertPatternMaps,
			RefActions:        options.RefActions,
			IssuePrefix:       options.IssuePrefix,
			NoteKeywords:      options.NoteKeywords,
			Paths:             options.Paths,
		},
	}

	parser := newCommitParser(c.logger, c.client, config)
	return parser.Parse(options.Revision)
}

// GetLatestTag implementiert die Client-Schnittstelle
func (c *ClientImpl) GetLatestTag() (string, error) {
	out, err := c.Exec("describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetTags implementiert die Client-Schnittstelle
func (c *ClientImpl) GetTags() ([]string, error) {
	out, err := c.Exec("tag", "--sort=-creatordate")
	if err != nil {
		return nil, err
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	// Leere Tags entfernen
	var result []string
	for _, tag := range tags {
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result, nil
}

// GetTagsWithPattern implementiert die Client-Schnittstelle
func (c *ClientImpl) GetTagsWithPattern(pattern string) ([]string, error) {
	out, err := c.Exec("tag", "-l", pattern, "--sort=-creatordate")
	if err != nil {
		return nil, err
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	// Leere Tags entfernen
	var result []string
	for _, tag := range tags {
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result, nil
}

// GetCommitsBetweenTags implementiert die Client-Schnittstelle
func (c *ClientImpl) GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error) {
	revRange := fmt.Sprintf("%s..%s", fromTag, toTag)
	return c.GetCommits(CommitOptions{Revision: revRange})
}

// GetStagedChanges implementiert die Client-Schnittstelle
func (c *ClientImpl) GetStagedChanges() (string, error) {
	return c.Exec("diff", "--cached")
}

// GetDiffBetweenTags implementiert die Client-Schnittstelle
func (c *ClientImpl) GetDiffBetweenTags(fromTag, toTag string) (string, error) {
	return c.Exec("diff", fromTag+".."+toTag)
}

// CreateTag implementiert die Client-Schnittstelle
func (c *ClientImpl) CreateTag(tag string, message string) error {
	_, err := c.Exec("tag", "-a", tag, "-m", message)
	return err
}

// HasRemote implementiert die Client-Schnittstelle
func (c *ClientImpl) HasRemote() (bool, error) {
	out, err := c.Exec("remote")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// GetCurrentBranch implementiert die Client-Schnittstelle
func (c *ClientImpl) GetCurrentBranch() (string, error) {
	out, err := c.Exec("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Exec führt den Git-Befehl im Repository-Verzeichnis aus
func (c *ClientImpl) Exec(subcmd string, args ...string) (string, error) {
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
	out, err := c.client.Exec(subcmd, args...)
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
func (c *ClientImpl) CanExec() error {
	return c.client.CanExec()
}

// InsideWorkTree prüft, ob wir uns innerhalb eines Git-Arbeitsbaumes befinden
func (c *ClientImpl) InsideWorkTree() error {
	// Ursprüngliches Arbeitsverzeichnis speichern
	origWd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// In das Repository-Verzeichnis wechseln
	if err := os.Chdir(c.repoDir); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Prüfen, ob wir uns in einem Git-Arbeitsbaum befinden
	result := c.client.InsideWorkTree()

	// Zum ursprünglichen Verzeichnis zurückkehren
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		// Falls die InsideWorkTree-Prüfung erfolgreich war, geben wir den Fehler beim Verzeichniswechsel zurück
		if result == nil {
			return chdirErr
		}
	}

	return result
}

// Diese Funktionen wurden durch den vollständigen commitParser ersetzt
