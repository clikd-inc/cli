package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	"clikd/internal/utils"

	gitcmd "github.com/tsuyoshiwada/go-gitcmd"
)

// CommitOptions defines the options for retrieving commits
type CommitOptions struct {
	Revision string   // Git revision (e.g. "HEAD", "master", tag)
	Paths    []string // Paths for which commits should be retrieved
	// Pattern configuration
	HeaderPattern     string   // A regular expression for parsing the commit header
	HeaderPatternMaps []string // A rule for mapping the result of `HeaderPattern` to the property of `Commit`
	MergePattern      string   // A regular expression for parsing merge commits
	MergePatternMaps  []string // Similar to `HeaderPatternMaps`
	RevertPattern     string   // A regular expression for parsing revert commits
	RevertPatternMaps []string // Similar to `HeaderPatternMaps`
	RefActions        []string // Word list of `Ref.Action`
	IssuePrefix       []string // Prefix for issues (e.g. `#`, `gh-`)
	NoteKeywords      []string // Keyword list for finding `Note`
}

// Client interface defines the minimal interface needed by tagReader
type Client interface {
	// Exec executes a Git command directly
	Exec(subcmd string, args ...string) (string, error)

	// CanExec checks if Git can be executed
	CanExec() error

	// InsideWorkTree checks if we are inside a Git work tree
	InsideWorkTree() error

	// GetRepositoryRoot returns the absolute path to the root of the Git repository
	GetRepositoryRoot() (string, error)

	// IsGitRepository checks if the current directory is a Git repository
	IsGitRepository() (bool, error)

	// GetCommits retrieves all commits in the repository based on filter options
	GetCommits(options CommitOptions) ([]*Commit, error)

	// GetCommitsWithOptions retrieves all commits with full pattern options
	GetCommitsWithOptions(options CommitOptions) ([]*Commit, error)

	// GetLatestTag returns the latest tag in the repository
	GetLatestTag() (string, error)

	// GetTags returns all tags in the repository
	GetTags() ([]string, error)

	// GetTagsWithPattern returns tags that match a specific pattern
	GetTagsWithPattern(pattern string) ([]string, error)

	// GetCommitsBetweenTags returns all commits between two tags
	GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error)

	// GetStagedChanges returns the staged changes
	GetStagedChanges() (string, error)

	// GetDiffBetweenTags returns the diff between two tags
	GetDiffBetweenTags(fromTag, toTag string) (string, error)

	// CreateTag creates a new tag
	CreateTag(tag string, message string) error

	// HasRemote checks if the repository has a remote
	HasRemote() (bool, error)

	// GetCurrentBranch returns the current branch
	GetCurrentBranch() (string, error)
}

// Service definiert die Hauptschnittstelle für Git-Funktionalitäten
type Service interface {
	// Core repository operations
	GetRepositoryRoot() (string, error)
	IsGitRepository() (bool, error)
	HasRemote() (bool, error)
	GetCurrentBranch() (string, error)

	// Commit operations
	GetCommits(revision string, paths []string) ([]*Commit, error)
	GetCommitsWithOptions(options CommitOptions) ([]*Commit, error)
	GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error)
	GetStagedChanges() (string, error)

	// Tag operations
	GetLatestTag() (string, error)
	GetTags() ([]string, error)
	GetTagsWithPattern(pattern string) ([]string, error)
	GetAllTagsWithDetails() ([]*Tag, error)
	CreateTag(tag string, message string) error

	// High-level operations
	SelectTagRange(query string) (string, string, error)
	SelectTagsWithQuery(tags []*Tag, query string) ([]*Tag, string, error)
	ExtractCommits(commits []*Commit, opts *Options) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup)
	GetDiffBetweenTags(fromTag, toTag string) (string, error)

	// New high-level methods
	AnalyzeRepository(ctx context.Context) (*RepositoryInfo, error)
	GetChangelogData(ctx context.Context, options *ChangelogOptions) (*ChangelogData, error)
}

// RepositoryInfo contains comprehensive repository information
type RepositoryInfo struct {
	Root          string
	IsGitRepo     bool
	HasRemote     bool
	CurrentBranch string
	LatestTag     string
	TotalTags     int
	TotalCommits  int
}

// ChangelogOptions contains options for changelog data extraction
type ChangelogOptions struct {
	Query            string
	TagFilterPattern string
	TagSortBy        string
	Paths            []string
	FromTag          string
	ToTag            string
}

// ChangelogData contains all data needed for changelog generation
type ChangelogData struct {
	Repository *RepositoryInfo
	Tags       []*Tag
	Commits    []*Commit
	FromTag    string
	ToTag      string
}

// ServiceImpl ist die konkrete Implementierung des Service-Interfaces
type ServiceImpl struct {
	gitCmd      gitcmd.Client
	repoDir     string
	tagReader   *tagReader
	tagSelector *tagSelector
	logger      utils.Logger
}

// NewService erstellt eine neue Instanz des Git-Services mit Standard-Konfiguration
func NewService() (Service, error) {
	// Use current working directory as default repository directory
	repoDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error determining current directory: %w", err)
	}

	return NewServiceWithRepoDir(repoDir)
}

// NewServiceWithRepoDir erstellt eine neue Instanz des Git-Services mit einem benutzerdefinierten Repository-Pfad
func NewServiceWithRepoDir(repoDir string) (Service, error) {
	return NewServiceWithOptions(repoDir, "", "date", nil)
}

// NewServiceWithOptions erstellt eine neue Instanz des Git-Services mit benutzerdefinierten Optionen
func NewServiceWithOptions(repoDir, tagFilterPattern string, tagSortBy string, logger utils.Logger) (Service, error) {
	// Use provided logger or create default
	if logger == nil {
		logger = utils.DefaultLogger.WithFields(map[string]interface{}{"module": "git"})
	}

	// Create Git client
	gitCmd := gitcmd.New(&gitcmd.Config{
		Bin: "git",
	})

	if repoDir == "" {
		// Use current working directory as default repository directory
		var err error
		repoDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error determining current directory: %w", err)
		}
	}

	service := &ServiceImpl{
		gitCmd:  gitCmd,
		repoDir: repoDir,
		logger:  logger,
	}

	// Create tag reader and selector with the service as client
	service.tagReader = newTagReader(&gitExecutor{service}, tagFilterPattern, tagSortBy)
	service.tagSelector = newTagSelector(logger)

	return service, nil
}

// gitExecutor is a wrapper that implements the Client interface for tagReader
type gitExecutor struct {
	service *ServiceImpl
}

// Exec implements the Client interface for gitExecutor
func (g *gitExecutor) Exec(subcmd string, args ...string) (string, error) {
	return g.service.Exec(subcmd, args...)
}

// CanExec implements the Client interface for gitExecutor
func (g *gitExecutor) CanExec() error {
	return g.service.CanExec()
}

// InsideWorkTree implements the Client interface for gitExecutor
func (g *gitExecutor) InsideWorkTree() error {
	return g.service.InsideWorkTree()
}

// GetRepositoryRoot implements the Client interface for gitExecutor
func (g *gitExecutor) GetRepositoryRoot() (string, error) {
	return g.service.GetRepositoryRoot()
}

// IsGitRepository implements the Client interface for gitExecutor
func (g *gitExecutor) IsGitRepository() (bool, error) {
	return g.service.IsGitRepository()
}

// GetCommits implements the Client interface for gitExecutor
func (g *gitExecutor) GetCommits(options CommitOptions) ([]*Commit, error) {
	return g.service.GetCommitsWithOptions(options)
}

// GetCommitsWithOptions implements the Client interface for gitExecutor
func (g *gitExecutor) GetCommitsWithOptions(options CommitOptions) ([]*Commit, error) {
	return g.service.GetCommitsWithOptions(options)
}

// GetLatestTag implements the Client interface for gitExecutor
func (g *gitExecutor) GetLatestTag() (string, error) {
	return g.service.GetLatestTag()
}

// GetTags implements the Client interface for gitExecutor
func (g *gitExecutor) GetTags() ([]string, error) {
	return g.service.GetTags()
}

// GetTagsWithPattern implements the Client interface for gitExecutor
func (g *gitExecutor) GetTagsWithPattern(pattern string) ([]string, error) {
	return g.service.GetTagsWithPattern(pattern)
}

// GetCommitsBetweenTags implements the Client interface for gitExecutor
func (g *gitExecutor) GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error) {
	return g.service.GetCommitsBetweenTags(fromTag, toTag)
}

// GetStagedChanges implements the Client interface for gitExecutor
func (g *gitExecutor) GetStagedChanges() (string, error) {
	return g.service.GetStagedChanges()
}

// GetDiffBetweenTags implements the Client interface for gitExecutor
func (g *gitExecutor) GetDiffBetweenTags(fromTag, toTag string) (string, error) {
	return g.service.GetDiffBetweenTags(fromTag, toTag)
}

// CreateTag implements the Client interface for gitExecutor
func (g *gitExecutor) CreateTag(tag string, message string) error {
	return g.service.CreateTag(tag, message)
}

// HasRemote implements the Client interface for gitExecutor
func (g *gitExecutor) HasRemote() (bool, error) {
	return g.service.HasRemote()
}

// GetCurrentBranch implements the Client interface for gitExecutor
func (g *gitExecutor) GetCurrentBranch() (string, error) {
	return g.service.GetCurrentBranch()
}

// Exec executes a Git command in the repository directory
func (s *ServiceImpl) Exec(subcmd string, args ...string) (string, error) {
	s.logger.Debug("Executing Git command", "subcmd", subcmd, "args", strings.Join(args, " "))

	// Save original working directory
	origWd, err := os.Getwd()
	if err != nil {
		s.logger.Error("Error determining current directory", "error", err)
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to the repository directory
	if err := os.Chdir(s.repoDir); err != nil {
		s.logger.Error("Error changing to repository directory", "error", err)
		return "", fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Execute Git command
	out, err := s.gitCmd.Exec(subcmd, args...)
	if err != nil {
		s.logger.Error("Git command failed", "error", err)
	}

	// Return to the original directory
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		s.logger.Error("Error returning to original directory", "error", chdirErr)
		// If the Git command was successful, we don't return the directory change error
		if err == nil {
			err = chdirErr
		}
	}

	return out, err
}

// CanExec checks if the Git command can be executed
func (s *ServiceImpl) CanExec() error {
	return s.gitCmd.CanExec()
}

// InsideWorkTree checks if we are inside a Git work tree
func (s *ServiceImpl) InsideWorkTree() error {
	// Save original working directory
	origWd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to the repository directory
	if err := os.Chdir(s.repoDir); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Check if we are in a Git work tree
	result := s.gitCmd.InsideWorkTree()

	// Return to the original directory
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		// If the InsideWorkTree check was successful, we return the directory change error
		if result == nil {
			return chdirErr
		}
	}

	return result
}

// GetRepositoryRoot implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetRepositoryRoot() (string, error) {
	out, err := s.Exec("rev-parse", "--show-toplevel")
	if err != nil {
		s.logger.Error("Error determining repository root", "error", err)
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsGitRepository implementiert die Service-Schnittstelle
func (s *ServiceImpl) IsGitRepository() (bool, error) {
	err := s.InsideWorkTree()
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetCommits implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCommits(revision string, paths []string) ([]*Commit, error) {
	s.logger.Debug("Getting commits", "revision", revision)

	// Use default values for patterns
	options := CommitOptions{
		Revision:          revision,
		Paths:             paths,
		HeaderPattern:     "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
		HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
		MergePattern:      "^Merge pull request #(\\d+) from (.*)$",
		MergePatternMaps:  []string{"Ref", "Source"},
		RevertPattern:     "^Revert \"([\\s\\S]*)\"$",
		RevertPatternMaps: []string{"Header"},
		RefActions:        []string{"close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved"},
		IssuePrefix:       []string{"#"},
		NoteKeywords:      []string{"BREAKING CHANGE"},
	}

	return s.GetCommitsWithOptions(options)
}

// GetCommitsWithOptions implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCommitsWithOptions(options CommitOptions) ([]*Commit, error) {
	s.logger.Debug("Getting commits with options", "revision", options.Revision)

	// Use the full commitParser
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

	parser := newCommitParser(s.logger, s.gitCmd, config)
	return parser.Parse(options.Revision)
}

// GetLatestTag implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetLatestTag() (string, error) {
	out, err := s.Exec("describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetTags implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetTags() ([]string, error) {
	out, err := s.Exec("tag", "--sort=-creatordate")
	if err != nil {
		return nil, err
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	// Remove empty tags
	var result []string
	for _, tag := range tags {
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result, nil
}

// GetTagsWithPattern implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetTagsWithPattern(pattern string) ([]string, error) {
	out, err := s.Exec("tag", "-l", pattern, "--sort=-creatordate")
	if err != nil {
		return nil, err
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	// Remove empty tags
	var result []string
	for _, tag := range tags {
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result, nil
}

// GetCommitsBetweenTags implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error) {
	revRange := fmt.Sprintf("%s..%s", fromTag, toTag)
	return s.GetCommits(revRange, nil)
}

// GetStagedChanges implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetStagedChanges() (string, error) {
	return s.Exec("diff", "--cached")
}

// GetDiffBetweenTags implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetDiffBetweenTags(fromTag, toTag string) (string, error) {
	return s.Exec("diff", fromTag+".."+toTag)
}

// CreateTag implementiert die Service-Schnittstelle
func (s *ServiceImpl) CreateTag(tag string, message string) error {
	_, err := s.Exec("tag", "-a", tag, "-m", message)
	return err
}

// HasRemote implementiert die Service-Schnittstelle
func (s *ServiceImpl) HasRemote() (bool, error) {
	out, err := s.Exec("remote")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// GetCurrentBranch implementiert die Service-Schnittstelle
func (s *ServiceImpl) GetCurrentBranch() (string, error) {
	out, err := s.Exec("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// AnalyzeRepository provides comprehensive repository information
func (s *ServiceImpl) AnalyzeRepository(ctx context.Context) (*RepositoryInfo, error) {
	s.logger.Debug("Analyzing repository...")

	info := &RepositoryInfo{}

	// Get repository root
	root, err := s.GetRepositoryRoot()
	if err != nil {
		return nil, err
	}
	info.Root = root

	// Check if it's a git repository
	isGit, err := s.IsGitRepository()
	if err != nil {
		return nil, err
	}
	info.IsGitRepo = isGit

	if !isGit {
		return info, nil
	}

	// Get remote status
	hasRemote, err := s.HasRemote()
	if err != nil {
		s.logger.Warn("Failed to check remote status", "error", err)
	} else {
		info.HasRemote = hasRemote
	}

	// Get current branch
	branch, err := s.GetCurrentBranch()
	if err != nil {
		s.logger.Warn("Failed to get current branch", "error", err)
	} else {
		info.CurrentBranch = branch
	}

	// Get latest tag
	latestTag, err := s.GetLatestTag()
	if err != nil {
		s.logger.Debug("No tags found or failed to get latest tag", "error", err)
	} else {
		info.LatestTag = latestTag
	}

	// Get tag count
	tags, err := s.GetTags()
	if err != nil {
		s.logger.Debug("Failed to get tags", "error", err)
	} else {
		info.TotalTags = len(tags)
	}

	// Get commit count (approximate)
	commits, err := s.GetCommits("", nil)
	if err != nil {
		s.logger.Debug("Failed to get commits", "error", err)
	} else {
		info.TotalCommits = len(commits)
	}

	s.logger.Debug("Repository analysis complete", "tags", info.TotalTags, "commits", info.TotalCommits)
	return info, nil
}

// GetChangelogData extracts all data needed for changelog generation
func (s *ServiceImpl) GetChangelogData(ctx context.Context, options *ChangelogOptions) (*ChangelogData, error) {
	s.logger.Debug("Extracting changelog data", "options", fmt.Sprintf("%+v", options))

	data := &ChangelogData{}

	// Analyze repository first
	repoInfo, err := s.AnalyzeRepository(ctx)
	if err != nil {
		return nil, err
	}
	data.Repository = repoInfo

	if !repoInfo.IsGitRepo {
		return data, nil
	}

	// Get all tags with details
	tags, err := s.GetAllTagsWithDetails()
	if err != nil {
		return nil, err
	}
	data.Tags = tags

	// Select tag range if query provided
	if options.Query != "" {
		selectedTags, firstTag, err := s.SelectTagsWithQuery(tags, options.Query)
		if err != nil {
			return nil, err
		}

		if len(selectedTags) >= 2 {
			data.FromTag = selectedTags[len(selectedTags)-1].Name
			data.ToTag = selectedTags[0].Name
		} else if firstTag != "" {
			data.FromTag = firstTag
		}
	} else if options.FromTag != "" || options.ToTag != "" {
		data.FromTag = options.FromTag
		data.ToTag = options.ToTag
	}

	// Get commits based on tag range or all commits
	var commits []*Commit
	if data.FromTag != "" && data.ToTag != "" {
		commits, err = s.GetCommitsBetweenTags(data.FromTag, data.ToTag)
	} else if data.FromTag != "" {
		commits, err = s.GetCommits(data.FromTag+"..", options.Paths)
	} else {
		commits, err = s.GetCommits("", options.Paths)
	}

	if err != nil {
		return nil, err
	}
	data.Commits = commits

	s.logger.Debug("Changelog data extracted", "tags", len(data.Tags), "commits", len(data.Commits), "fromTag", data.FromTag, "toTag", data.ToTag)

	return data, nil
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

	return selectedTags[len(selectedTags)-1].Name, selectedTags[0].Name, nil
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
