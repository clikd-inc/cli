// Package chglog implements main logic for the CHANGELOG generate.
package changelog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"

	"clikd/internal/services/git"
	"clikd/internal/utils"
)

// Options is an option used to process commits
type Options struct {
	Processor                   Processor
	NextTag                     string              // Treat unreleased commits as specified tags (EXPERIMENTAL)
	TagFilterPattern            string              // Filter tag by regexp
	Sort                        string              // Specify how to sort tags; currently supports "date" (default) or by "semver".
	NoCaseSensitive             bool                // Filter commits in a case insensitive way
	CommitFilters               map[string][]string // Filter by using `Commit` properties and values. Filtering is not done by specifying an empty value
	CommitSortBy                string              // Property name to use for sorting `Commit` (e.g. `Scope`)
	CommitGroupBy               string              // Property name of `Commit` to be grouped into `CommitGroup` (e.g. `Type`)
	CommitGroupSortBy           string              // Property name to use for sorting `CommitGroup` (e.g. `Title`)
	CommitGroupTitleOrder       []string            // Predefined sorted list of titles to use for sorting `CommitGroup`. Only if `CommitGroupSortBy` is `Custom`
	CommitGroupTitleMaps        map[string]string   // Map for `CommitGroup` title conversion
	HeaderPattern               string              // A regular expression to use for parsing the commit header
	HeaderPatternMaps           []string            // A rule for mapping the result of `HeaderPattern` to the property of `Commit`
	IssuePrefix                 []string            // Prefix used for issues (e.g. `#`, `gh-`)
	RefActions                  []string            // Word list of `Ref.Action`
	MergePattern                string              // A regular expression to use for parsing the merge commit
	MergePatternMaps            []string            // Similar to `HeaderPatternMaps`
	RevertPattern               string              // A regular expression to use for parsing the revert commit
	RevertPatternMaps           []string            // Similar to `HeaderPatternMaps`
	NoteKeywords                []string            // Keyword list to find `Note`. A semicolon is a separator, like `<keyword>:` (e.g. `BREAKING CHANGE`)
	JiraUsername                string
	JiraToken                   string
	JiraURL                     string
	JiraTypeMaps                map[string]string
	JiraIssueDescriptionPattern string
	Paths                       []string // Path filter
}

// Info is metadata related to CHANGELOG
type Info struct {
	Title         string // Title of CHANGELOG
	RepositoryURL string // URL of git repository
}

// RenderData is the data passed to the template
type RenderData struct {
	Info       *Info
	Unreleased *git.Unreleased
	Versions   []*git.Version
}

// Config for generating CHANGELOG
type Config struct {
	Bin        string // Git execution command
	WorkingDir string // Working directory
	Template   string // Path for template file. If a relative path is specified, it depends on the value of `WorkingDir`.
	Info       *Info
	Options    *Options
}

func normalizeConfig(config *Config) {
	opts := config.Options

	if opts.HeaderPattern == "" {
		opts.HeaderPattern = "^(.*)$"
		opts.HeaderPatternMaps = []string{
			"Subject",
		}
	}

	if opts.MergePattern == "" {
		opts.MergePattern = "^Merge branch '(\\w+)'$"
		opts.MergePatternMaps = []string{
			"Source",
		}
	}

	if opts.RevertPattern == "" {
		opts.RevertPattern = "^Revert \"([\\s\\S]*)\"$"
		opts.RevertPatternMaps = []string{
			"Header",
		}
	}

	config.Options = opts
}

// Generator of CHANGELOG
type Generator struct {
	config     *Config
	gitService git.Service
	jiraClient JiraClient
	logger     utils.Logger
}

// NewGenerator receives `Config` and create an new `Generator`
func NewGenerator(logger utils.Logger, config *Config) *Generator {
	// Repository directory
	repoDir := config.WorkingDir

	// Create Git service with the target repository
	gitService, err := git.NewServiceWithRepoDir(repoDir)
	if err != nil {
		logger.Error("Failed to create Git service: %v", err)
		return nil
	}

	jiraClient := NewJiraClient(config)

	if config.Options.Processor != nil {
		config.Options.Processor.Bootstrap(config)
	}

	normalizeConfig(config)

	return &Generator{
		config:     config,
		gitService: gitService,
		jiraClient: jiraClient,
		logger:     logger,
	}
}

// Generate gets the commit based on the specified tag `query` and writes the result to `io.Writer`
//
// tag `query` can be specified with the following rule
//
//	<old>..<new> - Commit contained in `<new>` tags from `<old>` (e.g. `1.0.0..2.0.0`)
//	<tagname>..  - Commit from the `<tagname>` to the latest tag (e.g. `1.0.0..`)
//	..<tagname>  - Commit from the oldest tag to `<tagname>` (e.g. `..1.0.0`)
//	<tagname>    - Commit contained in `<tagname>` (e.g. `1.0.0`)
func (gen *Generator) Generate(w io.Writer, query string) error {
	fmt.Printf("DEBUG: Generate called with query: %q\n", query)

	// Check current working directory
	if currWd, err := os.Getwd(); err == nil {
		fmt.Printf("DEBUG: Current working directory in Generate: %s\n", currWd)
	}

	// Get tags using Git service
	tags, first, err := gen.getTags(query)
	if err != nil {
		fmt.Printf("DEBUG: Error in getTags: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: getTags returned %d tags, first=%s\n", len(tags), first)

	// Debug: Show detailed information for each tag
	for i, tag := range tags {
		prevName := "nil"
		if tag.Previous != nil {
			prevName = tag.Previous.Name
		}
		fmt.Printf("DEBUG: tag[%d]: name=%s, date=%s, previous=%s\n",
			i, tag.Name, tag.Date.Format("2006-01-02 15:04:05"), prevName)
	}

	unreleased, err := gen.readUnreleased(tags)
	if err != nil {
		fmt.Printf("DEBUG: Error in readUnreleased: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: readUnreleased result: commitGroups=%d, commits=%d\n",
		len(unreleased.CommitGroups), len(unreleased.Commits))

	// Debug: Show details for unreleased commits
	for i, commit := range unreleased.Commits {
		fmt.Printf("DEBUG: unreleased commit[%d]: hash=%s, subject=%s\n",
			i, commit.Hash.Short, commit.Subject)
	}

	versions, err := gen.readVersions(tags, first)
	if err != nil {
		fmt.Printf("DEBUG: Error in readVersions: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: readVersions returned %d versions\n", len(versions))
	for i, v := range versions {
		fmt.Printf("DEBUG: version[%d]: tag=%s, commits=%d, commitGroups=%d\n",
			i, v.Tag.Name, len(v.Commits), len(v.CommitGroups))

		// Debug: Show details for version commits
		for j, commit := range v.Commits {
			fmt.Printf("DEBUG: version[%d] commit[%d]: hash=%s, subject=%s\n",
				i, j, commit.Hash.Short, commit.Subject)
		}
	}

	if len(versions) == 0 {
		fmt.Printf("DEBUG: No versions found for query %q\n", query)
		return fmt.Errorf("commits corresponding to \"%s\" was not found", query)
	}

	return gen.render(w, unreleased, versions)
}

func (gen *Generator) readVersions(tags []*git.Tag, first string) ([]*git.Version, error) {
	fmt.Printf("DEBUG: readVersions called with %d tags, first=%s\n", len(tags), first)

	next := gen.config.Options.NextTag
	versions := []*git.Version{}

	for i, tag := range tags {
		var (
			isNext = next == tag.Name
			rev    string
		)

		if isNext {
			if tag.Previous != nil {
				rev = tag.Previous.Name + "..HEAD"
			} else {
				rev = "HEAD"
			}
		} else {
			if i+1 < len(tags) {
				rev = tags[i+1].Name + ".." + tag.Name
			} else {
				if first != "" {
					rev = first + ".." + tag.Name
				} else {
					rev = tag.Name
				}
			}
		}

		fmt.Printf("DEBUG: Processing tag[%d] %s with rev=%s\n", i, tag.Name, rev)

		// Get commits for this revision using Git service
		commits, err := gen.gitService.GetCommits(rev, gen.config.Options.Paths)
		if err != nil {
			return nil, err
		}

		// Process commits
		commits = gen.processCommits(commits)

		// Extrahiere und gruppiere Commits mit dem Git-Service
		commitGroups, mergeCommits, revertCommits, noteGroups := gen.extractCommits(commits)

		version := &git.Version{
			Tag:           tag,
			Commits:       commits,
			CommitGroups:  commitGroups,
			MergeCommits:  mergeCommits,
			RevertCommits: revertCommits,
			NoteGroups:    noteGroups,
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func (gen *Generator) readUnreleased(tags []*git.Tag) (*git.Unreleased, error) {
	if len(tags) == 0 {
		return &git.Unreleased{}, nil
	}

	rev := tags[0].Name + "..HEAD"
	fmt.Printf("DEBUG: readUnreleased: Getting commits with rev=%s\n", rev)

	// Get commits for unreleased changes using Git service
	commits, err := gen.gitService.GetCommits(rev, gen.config.Options.Paths)
	if err != nil {
		return nil, err
	}

	// Process commits
	commits = gen.processCommits(commits)

	// Extrahiere und gruppiere Commits mit dem Git-Service
	commitGroups, mergeCommits, revertCommits, noteGroups := gen.extractCommits(commits)

	unreleased := &git.Unreleased{
		Commits:       commits,
		CommitGroups:  commitGroups,
		MergeCommits:  mergeCommits,
		RevertCommits: revertCommits,
		NoteGroups:    noteGroups,
	}

	return unreleased, nil
}

func (gen *Generator) getTags(query string) ([]*git.Tag, string, error) {
	// Get all tags using Git service
	tagsWithDetails, err := gen.gitService.GetAllTagsWithDetails()
	if err != nil {
		return nil, "", err
	}
	fmt.Printf("DEBUG: getTags: Found %d tags\n", len(tagsWithDetails))

	// Select tags based on query using Git service
	selectedTags, first, err := gen.gitService.SelectTagsWithQuery(tagsWithDetails, query)
	if err != nil {
		return nil, "", err
	}

	return selectedTags, first, nil
}

func (gen *Generator) workdir() (func() error, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = os.Chdir(gen.config.WorkingDir)
	if err != nil {
		return nil, err
	}

	return func() error {
		return os.Chdir(cwd)
	}, nil
}

func (gen *Generator) render(w io.Writer, unreleased *git.Unreleased, versions []*git.Version) error {
	fmap := template.FuncMap{
		"datetime": func(layout string, input time.Time) string {
			return input.Format(layout)
		},
	}

	// Combine with sprig functions
	for k, v := range sprig.TxtFuncMap() {
		fmap[k] = v
	}

	fname := gen.config.Template
	if !filepath.IsAbs(fname) {
		fname = filepath.Join(gen.config.WorkingDir, fname)
	}

	tmpl, err := template.New("").Funcs(fmap).ParseFiles(fname)
	if err != nil {
		return err
	}

	tname := filepath.Base(fname)
	data := &RenderData{
		Info:       gen.config.Info,
		Unreleased: unreleased,
		Versions:   versions,
	}

	return tmpl.ExecuteTemplate(w, tname, data)
}

// Process commits using the Processor from options
func (gen *Generator) processCommits(commits []*git.Commit) []*git.Commit {
	processor := gen.config.Options.Processor
	if processor == nil {
		return commits
	}

	// Process jira issues for each commit
	for _, commit := range commits {
		if commit.JiraIssueID != "" {
			gen.processJiraIssue(commit)
		}
	}

	// Filter using the Processor
	processed := make([]*git.Commit, 0, len(commits))
	for _, commit := range commits {
		// Convert git.Commit to changelog.Commit for processor
		localCommit := convertToLocalCommit(commit)

		// Process the commit
		localProcessed := processor.ProcessCommit(localCommit)
		if localProcessed == nil {
			continue
		}

		// Convert back to git.Commit
		processed = append(processed, convertToGitCommit(localProcessed))
	}

	return processed
}

// Commit is a local representation of a git commit for the changelog
type Commit struct {
	Hash        *Hash
	Author      *Author
	Committer   *Committer
	Merge       *Merge
	Revert      *Revert
	Refs        []*Ref
	Notes       []*Note
	Mentions    []string
	CoAuthors   []Contact
	Signers     []Contact
	JiraIssue   *JiraIssue
	Header      string
	Type        string
	Scope       string
	Subject     string
	JiraIssueID string
	Body        string
	TrimmedBody string
}

// Hash of commit
type Hash struct {
	Long  string
	Short string
}

// Contact of co-authors and signers
type Contact struct {
	Name  string
	Email string
}

// Author of commit
type Author struct {
	Name  string
	Email string
	Date  time.Time
}

// Committer of commit
type Committer struct {
	Name  string
	Email string
	Date  time.Time
}

// Merge info for commit
type Merge struct {
	Ref    string
	Source string
}

// Revert info for commit
type Revert struct {
	Header string
}

// Ref is abstract data related to commit. (e.g. `Issues`, `Pull Request`)
type Ref struct {
	Action string
	Ref    string
	Source string
}

// Note of commit
type Note struct {
	Title string
	Body  string
}

// JiraIssue is information about a jira ticket
type JiraIssue struct {
	Key         string
	Type        string
	Summary     string
	Description string
	Labels      []string
}

// Helper function to convert git.Commit to changelog.Commit
func convertToLocalCommit(commit *git.Commit) *Commit {
	if commit == nil {
		return nil
	}

	// Create a local Hash
	hash := &Hash{}
	if commit.Hash != nil {
		hash.Long = commit.Hash.Long
		hash.Short = commit.Hash.Short
	}

	// Create a local Author
	author := &Author{}
	if commit.Author != nil {
		author.Name = commit.Author.Name
		author.Email = commit.Author.Email
		author.Date = commit.Author.Date
	}

	// Create a local Committer
	committer := &Committer{}
	if commit.Committer != nil {
		committer.Name = commit.Committer.Name
		committer.Email = commit.Committer.Email
		committer.Date = commit.Committer.Date
	}

	// Create a local Merge
	var merge *Merge
	if commit.Merge != nil {
		merge = &Merge{
			Ref:    commit.Merge.Ref,
			Source: commit.Merge.Source,
		}
	}

	// Create a local Revert
	var revert *Revert
	if commit.Revert != nil {
		revert = &Revert{
			Header: commit.Revert.Header,
		}
	}

	// Convert Refs
	refs := make([]*Ref, 0, len(commit.Refs))
	for _, r := range commit.Refs {
		refs = append(refs, &Ref{
			Action: r.Action,
			Ref:    r.Ref,
			Source: r.Source,
		})
	}

	// Convert Notes
	notes := make([]*Note, 0, len(commit.Notes))
	for _, n := range commit.Notes {
		notes = append(notes, &Note{
			Title: n.Title,
			Body:  n.Body,
		})
	}

	// Convert Contact lists
	coAuthors := make([]Contact, 0, len(commit.CoAuthors))
	for _, c := range commit.CoAuthors {
		coAuthors = append(coAuthors, Contact{
			Name:  c.Name,
			Email: c.Email,
		})
	}

	signers := make([]Contact, 0, len(commit.Signers))
	for _, s := range commit.Signers {
		signers = append(signers, Contact{
			Name:  s.Name,
			Email: s.Email,
		})
	}

	// Create a local JiraIssue
	var jiraIssue *JiraIssue
	if commit.JiraIssue != nil {
		jiraIssue = &JiraIssue{
			Key:         commit.JiraIssue.Key,
			Type:        commit.JiraIssue.Type,
			Summary:     commit.JiraIssue.Summary,
			Description: commit.JiraIssue.Description,
			Labels:      commit.JiraIssue.Labels,
		}
	}

	return &Commit{
		Hash:        hash,
		Author:      author,
		Committer:   committer,
		Merge:       merge,
		Revert:      revert,
		Refs:        refs,
		Notes:       notes,
		Mentions:    commit.Mentions,
		CoAuthors:   coAuthors,
		Signers:     signers,
		JiraIssue:   jiraIssue,
		Header:      commit.Header,
		Type:        commit.Type,
		Scope:       commit.Scope,
		Subject:     commit.Subject,
		JiraIssueID: commit.JiraIssueID,
		Body:        commit.Body,
		TrimmedBody: commit.TrimmedBody,
	}
}

// Helper function to convert changelog.Commit to git.Commit
func convertToGitCommit(commit *Commit) *git.Commit {
	if commit == nil {
		return nil
	}

	// Create a git.Hash
	hash := &git.Hash{}
	if commit.Hash != nil {
		hash.Long = commit.Hash.Long
		hash.Short = commit.Hash.Short
	}

	// Create a git.Author
	author := &git.Author{}
	if commit.Author != nil {
		author.Name = commit.Author.Name
		author.Email = commit.Author.Email
		author.Date = commit.Author.Date
	}

	// Create a git.Committer
	committer := &git.Committer{}
	if commit.Committer != nil {
		committer.Name = commit.Committer.Name
		committer.Email = commit.Committer.Email
		committer.Date = commit.Committer.Date
	}

	// Create a git.Merge
	var merge *git.Merge
	if commit.Merge != nil {
		merge = &git.Merge{
			Ref:    commit.Merge.Ref,
			Source: commit.Merge.Source,
		}
	}

	// Create a git.Revert
	var revert *git.Revert
	if commit.Revert != nil {
		revert = &git.Revert{
			Header: commit.Revert.Header,
		}
	}

	// Convert Refs
	refs := make([]*git.Ref, 0, len(commit.Refs))
	for _, r := range commit.Refs {
		refs = append(refs, &git.Ref{
			Action: r.Action,
			Ref:    r.Ref,
			Source: r.Source,
		})
	}

	// Convert Notes
	notes := make([]*git.Note, 0, len(commit.Notes))
	for _, n := range commit.Notes {
		notes = append(notes, &git.Note{
			Title: n.Title,
			Body:  n.Body,
		})
	}

	// Convert Contact lists
	coAuthors := make([]git.Contact, 0, len(commit.CoAuthors))
	for _, c := range commit.CoAuthors {
		coAuthors = append(coAuthors, git.Contact{
			Name:  c.Name,
			Email: c.Email,
		})
	}

	signers := make([]git.Contact, 0, len(commit.Signers))
	for _, s := range commit.Signers {
		signers = append(signers, git.Contact{
			Name:  s.Name,
			Email: s.Email,
		})
	}

	// Create a git.JiraIssue
	var jiraIssue *git.JiraIssue
	if commit.JiraIssue != nil {
		jiraIssue = &git.JiraIssue{
			Key:         commit.JiraIssue.Key,
			Type:        commit.JiraIssue.Type,
			Summary:     commit.JiraIssue.Summary,
			Description: commit.JiraIssue.Description,
			Labels:      commit.JiraIssue.Labels,
		}
	}

	return &git.Commit{
		Hash:        hash,
		Author:      author,
		Committer:   committer,
		Merge:       merge,
		Revert:      revert,
		Refs:        refs,
		Notes:       notes,
		Mentions:    commit.Mentions,
		CoAuthors:   coAuthors,
		Signers:     signers,
		JiraIssue:   jiraIssue,
		Header:      commit.Header,
		Type:        commit.Type,
		Scope:       commit.Scope,
		Subject:     commit.Subject,
		JiraIssueID: commit.JiraIssueID,
		Body:        commit.Body,
		TrimmedBody: commit.TrimmedBody,
	}
}

func (gen *Generator) processJiraIssue(commit *git.Commit) {
	issue, err := gen.jiraClient.GetJiraIssue(commit.JiraIssueID)
	if err != nil {
		gen.logger.Error(fmt.Sprintf("Failed to parse Jira story %s: %s\n", commit.JiraIssueID, err))
		return
	}

	// Update commit with Jira information
	commit.Type = gen.config.Options.JiraTypeMaps[issue.Fields.Type.Name]
	commit.JiraIssue = &git.JiraIssue{
		Key:         issue.Key,
		Type:        issue.Fields.Type.Name,
		Summary:     issue.Fields.Summary,
		Description: issue.Fields.Description,
		Labels:      issue.Fields.Labels,
	}

	// Apply JiraIssueDescriptionPattern if configured
	if gen.config.Options.JiraIssueDescriptionPattern != "" {
		reJiraIssueDescription := regexp.MustCompile(gen.config.Options.JiraIssueDescriptionPattern)
		res := reJiraIssueDescription.FindStringSubmatch(commit.JiraIssue.Description)
		if len(res) > 1 {
			commit.JiraIssue.Description = res[1]
		}
	}
}

// extractCommits extrahiert Commit-Gruppen, Merge- und Revert-Commits sowie Notizen aus den Commits
// Nutzt die zentrale Extrakt-Funktionalität des Git-Services
func (gen *Generator) extractCommits(commits []*git.Commit) ([]*git.CommitGroup, []*git.Commit, []*git.Commit, []*git.NoteGroup) {
	// Git-Service-Optionen aus Changelog-Optionen erstellen
	gitOpts := &git.Options{
		CommitFilters:         gen.config.Options.CommitFilters,
		CommitSortBy:          gen.config.Options.CommitSortBy,
		CommitGroupBy:         gen.config.Options.CommitGroupBy,
		CommitGroupSortBy:     gen.config.Options.CommitGroupSortBy,
		CommitGroupTitleOrder: gen.config.Options.CommitGroupTitleOrder,
		CommitGroupTitleMaps:  gen.config.Options.CommitGroupTitleMaps,
		NoCaseSensitive:       gen.config.Options.NoCaseSensitive,
	}

	return gen.gitService.ExtractCommits(commits, gitOpts)
}
