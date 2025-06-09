package git

import (
	"clikd/internal/utils"
)

// Service definiert die Hauptschnittstelle für Git-Funktionalitäten
type Service interface {
	// GetRepositoryRoot gibt den absoluten Pfad zum Root des Git-Repositories zurück
	GetRepositoryRoot() (string, error)

	// IsGitRepository prüft, ob das aktuelle Verzeichnis ein Git-Repository ist
	IsGitRepository() (bool, error)

	// GetCommits holt alle Commits im Repository basierend auf den Filteroptionen
	GetCommits(revision string, paths []string) ([]*Commit, error)

	// GetCommitsWithOptions holt alle Commits mit vollständigen Pattern-Optionen
	GetCommitsWithOptions(options CommitOptions) ([]*Commit, error)

	// GetLatestTag gibt das neueste Tag im Repository zurück
	GetLatestTag() (string, error)

	// GetTags gibt alle Tags im Repository zurück
	GetTags() ([]string, error)

	// GetTagsWithPattern gibt Tags zurück, die einem bestimmten Pattern entsprechen
	GetTagsWithPattern(pattern string) ([]string, error)

	// SelectTagRange wählt einen Bereich von Tags aus basierend auf den Optionen
	SelectTagRange(query string) (string, string, error)

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

	// ExtractCommits gruppiert Commits basierend auf den konfigurierten Optionen
	ExtractCommits(commits []*Commit, opts *Options) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup)

	// GetAllTagsWithDetails gibt alle Tags mit Details zurück
	GetAllTagsWithDetails() ([]*Tag, error)

	// SelectTagsWithQuery wählt Tags basierend auf einer Abfrage aus
	SelectTagsWithQuery(tags []*Tag, query string) ([]*Tag, string, error)
}

// ServiceImpl ist die konkrete Implementierung des Service-Interfaces
type ServiceImpl struct {
	client      Client
	tagReader   *tagReader
	tagSelector *tagSelector
	logger      utils.Logger
}

// NewService erstellt eine neue Instanz des Git-Services
func NewService() (Service, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	// Verwende den globalen Logger anstatt hardcoded "debug"
	logger := utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})

	// Tag-Reader und Selektor erstellen
	tagReader := newTagReader(client, "", "date")
	tagSelector := newTagSelector(logger)

	return &ServiceImpl{
		client:      client,
		tagReader:   tagReader,
		tagSelector: tagSelector,
		logger:      logger,
	}, nil
}

// NewServiceWithClient erstellt eine neue Instanz des Git-Services mit einem benutzerdefinierten Client
func NewServiceWithClient(client Client) Service {
	// Verwende den globalen Logger anstatt hardcoded "debug"
	logger := utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})

	// Tag-Reader und Selektor erstellen
	tagReader := newTagReader(client, "", "date")
	tagSelector := newTagSelector(logger)

	return &ServiceImpl{
		client:      client,
		tagReader:   tagReader,
		tagSelector: tagSelector,
		logger:      logger,
	}
}

// NewServiceWithRepoDir erstellt eine neue Instanz des Git-Services mit einem benutzerdefinierten Repository-Pfad
func NewServiceWithRepoDir(repoDir string) (Service, error) {
	client, err := NewClientWithRepoDir(repoDir)
	if err != nil {
		return nil, err
	}

	// Verwende den globalen Logger anstatt hardcoded "debug"
	logger := utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})

	// Tag-Reader und Selektor erstellen
	tagReader := newTagReader(client, "", "date")
	tagSelector := newTagSelector(logger)

	return &ServiceImpl{
		client:      client,
		tagReader:   tagReader,
		tagSelector: tagSelector,
		logger:      logger,
	}, nil
}

// NewServiceWithOptions erstellt eine neue Instanz des Git-Services mit benutzerdefinierten Optionen
func NewServiceWithOptions(client Client, tagFilterPattern string, tagSortBy string) Service {
	// Verwende den globalen Logger anstatt hardcoded "debug"
	logger := utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})

	// Tag-Reader mit Filter-Pattern und Sortierung erstellen
	tagReader := newTagReader(client, tagFilterPattern, tagSortBy)
	tagSelector := newTagSelector(logger)

	return &ServiceImpl{
		client:      client,
		tagReader:   tagReader,
		tagSelector: tagSelector,
		logger:      logger,
	}
}

// GetRepositoryRoot implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetRepositoryRoot() (string, error) {
	return s.client.GetRepositoryRoot()
}

// IsGitRepository implementiert die Service-Schnittstelle
func (s *ServiceImpl) IsGitRepository() (bool, error) {
	return s.client.IsGitRepository()
}

// GetCommits implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCommits(revision string, paths []string) ([]*Commit, error) {
	return s.client.GetCommits(CommitOptions{
		Revision: revision,
		Paths:    paths,
	})
}

// GetCommitsWithOptions implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCommitsWithOptions(options CommitOptions) ([]*Commit, error) {
	return s.client.GetCommitsWithOptions(options)
}

// GetLatestTag implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetLatestTag() (string, error) {
	return s.client.GetLatestTag()
}

// GetTags implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetTags() ([]string, error) {
	return s.client.GetTags()
}

// GetTagsWithPattern implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetTagsWithPattern(pattern string) ([]string, error) {
	return s.client.GetTagsWithPattern(pattern)
}

// SelectTagRange implementiert die Service-Schnittstelle
func (s *ServiceImpl) SelectTagRange(query string) (string, string, error) {
	tags, err := s.GetAllTagsWithDetails()
	if err != nil {
		return "", "", err
	}

	selectedTags, firstTag, err := s.SelectTagsWithQuery(tags, query)
	if err != nil {
		return "", "", err
	}

	if len(selectedTags) < 2 {
		return firstTag, "", nil
	}

	return firstTag, selectedTags[len(selectedTags)-1].Name, nil
}

// GetCommitsBetweenTags implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error) {
	return s.client.GetCommitsBetweenTags(fromTag, toTag)
}

// GetStagedChanges implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetStagedChanges() (string, error) {
	return s.client.GetStagedChanges()
}

// GetDiffBetweenTags implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetDiffBetweenTags(fromTag, toTag string) (string, error) {
	return s.client.GetDiffBetweenTags(fromTag, toTag)
}

// CreateTag implementiert die Service-Schnittstelle
func (s *ServiceImpl) CreateTag(tag string, message string) error {
	return s.client.CreateTag(tag, message)
}

// HasRemote implementiert die Service-Schnittstelle
func (s *ServiceImpl) HasRemote() (bool, error) {
	return s.client.HasRemote()
}

// GetCurrentBranch implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCurrentBranch() (string, error) {
	return s.client.GetCurrentBranch()
}

// ExtractCommits implementiert die Service-Schnittstelle
func (s *ServiceImpl) ExtractCommits(commits []*Commit, opts *Options) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup) {
	extractor := newCommitExtractor(opts, s.logger)
	return extractor.Extract(commits)
}

// GetAllTagsWithDetails implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetAllTagsWithDetails() ([]*Tag, error) {
	return s.tagReader.ReadAll()
}

// SelectTagsWithQuery implementiert die Service-Schnittstelle
func (s *ServiceImpl) SelectTagsWithQuery(tags []*Tag, query string) ([]*Tag, string, error) {
	return s.tagSelector.Select(tags, query)
}
