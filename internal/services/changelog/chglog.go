package changelog

import (
	"clikd/internal/services/git"
	"clikd/internal/utils"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/tsuyoshiwada/go-gitcmd"
)

// JiraClientInterface Interface für die Interaktion mit Jira
type JiraClientInterface interface {
	FetchIssue(issueID string) (*git.JiraIssue, error)
}

// JiraClientMock ist eine Dummy-Implementierung des JiraClient-Interfaces
type JiraClientMock struct{}

// FetchIssue implementiert die JiraClientInterface
func (j *JiraClientMock) FetchIssue(issueID string) (*git.JiraIssue, error) {
	return &git.JiraIssue{
		Key:         issueID,
		Summary:     "Mock issue summary",
		Description: "Mock issue description",
		Type:        "Mock issue type",
	}, nil
}

// NewJiraClientForChglog erstellt einen neuen Jira-Client speziell für chglog
func NewJiraClientForChglog(config *Config) JiraClientInterface {
	// Implementiere einen Jira-Client oder einen Mock
	return &JiraClientMock{}
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
	client     gitcmd.Client
	config     *Config
	gitService git.Service
}

// NewGenerator receives `Config` and create an new `Generator`
func NewGenerator(logger utils.Logger, config *Config) *Generator {
	client := gitcmd.New(&gitcmd.Config{
		Bin: config.Bin,
	})

	if config.Options.Processor != nil {
		if err := config.Options.Processor.Bootstrap(config); err != nil {
			// Log error but continue
		}
	}

	normalizeConfig(config)

	// Git-Service mit Tag-Filter-Konfiguration erstellen
	gitClientWrapper, _ := git.NewClientWithRepoDir(config.WorkingDir)

	// Tag-Filter-Pattern und Sort-Option aus der Konfiguration verwenden
	tagFilterPattern := config.Options.TagFilterPattern
	tagSortBy := config.Options.Sort
	if tagSortBy == "" {
		tagSortBy = "date" // Standard-Sortierung
	}

	gitService := git.NewServiceWithOptions(gitClientWrapper, tagFilterPattern, tagSortBy)

	return &Generator{
		client:     client,
		config:     config,
		gitService: gitService,
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
	back, err := gen.workdir()
	if err != nil {
		return err
	}
	defer func() {
		if err = back(); err != nil {
			log.Fatal(err)
		}
	}()

	tags, first, err := gen.getTags(query)
	if err != nil {
		return err
	}

	unreleased, err := gen.readUnreleased(tags)
	if err != nil {
		return err
	}

	versions, err := gen.readVersions(tags, first)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		return fmt.Errorf("commits corresponding to \"%s\" was not found", query)
	}

	return gen.render(w, unreleased, versions)
}

func (gen *Generator) readVersions(tags []*git.Tag, first string) ([]*ChangelogVersion, error) {
	next := gen.config.Options.NextTag
	versions := []*ChangelogVersion{}

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
					// For the oldest/only tag, get all commits up to that tag
					// This ensures that when there's only one tag, all commits
					// are included in the version, not in unreleased
					rev = tag.Name
				}
			}
		}

		commits, err := gen.gitService.GetCommits(rev, gen.config.Options.Paths)
		if err != nil {
			return nil, err
		}

		commitGroups, mergeCommits, revertCommits, noteGroups := gen.gitService.ExtractCommits(commits, &git.Options{
			CommitGroupBy:        gen.config.Options.CommitGroupBy,
			CommitGroupSortBy:    gen.config.Options.CommitGroupSortBy,
			CommitGroupTitleMaps: gen.config.Options.CommitGroupTitleMaps,
			CommitFilters:        gen.config.Options.CommitFilters,
			CommitSortBy:         gen.config.Options.CommitSortBy,
		})

		// Konvertiere Git-Commits zu Changelog-Commits
		clCommits := make([]*ChangelogCommit, len(commits))
		for j, commit := range commits {
			clCommits[j] = &ChangelogCommit{Commit: commit}
		}

		// Konvertiere Git-MergeCommits zu Changelog-Commits
		clMergeCommits := make([]*ChangelogCommit, len(mergeCommits))
		for j, commit := range mergeCommits {
			clMergeCommits[j] = &ChangelogCommit{Commit: commit}
		}

		// Konvertiere Git-RevertCommits zu Changelog-Commits
		clRevertCommits := make([]*ChangelogCommit, len(revertCommits))
		for j, commit := range revertCommits {
			clRevertCommits[j] = &ChangelogCommit{Commit: commit}
		}

		// Konvertiere Git-CommitGroups zu Changelog-CommitGroups
		clCommitGroups := make([]*ChangelogCommitGroup, len(commitGroups))
		for j, group := range commitGroups {
			// Konvertiere Git-Commits der Gruppe zu Changelog-Commits
			groupCommits := make([]*ChangelogCommit, len(group.Commits))
			for k, commit := range group.Commits {
				groupCommits[k] = &ChangelogCommit{Commit: commit}
			}

			clCommitGroups[j] = &ChangelogCommitGroup{
				RawTitle: group.RawTitle,
				Title:    group.Title,
				Commits:  groupCommits,
			}
		}

		versions = append(versions, &ChangelogVersion{
			Tag:           tag,
			CommitGroups:  clCommitGroups,
			Commits:       clCommits,
			MergeCommits:  clMergeCommits,
			RevertCommits: clRevertCommits,
			NoteGroups:    noteGroups,
		})

		// Instead of `getTags()`, assign the date to the tag
		if isNext && len(commits) != 0 {
			tag.Date = commits[0].Author.Date
		}
	}

	return versions, nil
}

func (gen *Generator) readUnreleased(tags []*git.Tag) (*ChangelogUnreleased, error) {
	if gen.config.Options.NextTag != "" {
		return &ChangelogUnreleased{}, nil
	}

	rev := "HEAD"

	if len(tags) > 0 {
		rev = tags[0].Name + "..HEAD"
	}

	commits, err := gen.gitService.GetCommits(rev, gen.config.Options.Paths)
	if err != nil {
		return nil, err
	}

	commitGroups, mergeCommits, revertCommits, noteGroups := gen.gitService.ExtractCommits(commits, &git.Options{
		CommitGroupBy:        gen.config.Options.CommitGroupBy,
		CommitGroupSortBy:    gen.config.Options.CommitGroupSortBy,
		CommitGroupTitleMaps: gen.config.Options.CommitGroupTitleMaps,
		CommitFilters:        gen.config.Options.CommitFilters,
		CommitSortBy:         gen.config.Options.CommitSortBy,
	})

	// Konvertiere Git-Commits zu Changelog-Commits
	clCommits := make([]*ChangelogCommit, len(commits))
	for i, commit := range commits {
		clCommits[i] = &ChangelogCommit{Commit: commit}
	}

	// Konvertiere Git-MergeCommits zu Changelog-Commits
	clMergeCommits := make([]*ChangelogCommit, len(mergeCommits))
	for i, commit := range mergeCommits {
		clMergeCommits[i] = &ChangelogCommit{Commit: commit}
	}

	// Konvertiere Git-RevertCommits zu Changelog-Commits
	clRevertCommits := make([]*ChangelogCommit, len(revertCommits))
	for i, commit := range revertCommits {
		clRevertCommits[i] = &ChangelogCommit{Commit: commit}
	}

	// Konvertiere Git-CommitGroups zu Changelog-CommitGroups
	clCommitGroups := make([]*ChangelogCommitGroup, len(commitGroups))
	for i, group := range commitGroups {
		// Konvertiere Git-Commits der Gruppe zu Changelog-Commits
		groupCommits := make([]*ChangelogCommit, len(group.Commits))
		for j, commit := range group.Commits {
			groupCommits[j] = &ChangelogCommit{Commit: commit}
		}

		clCommitGroups[i] = &ChangelogCommitGroup{
			RawTitle: group.RawTitle,
			Title:    group.Title,
			Commits:  groupCommits,
		}
	}

	unreleased := &ChangelogUnreleased{
		CommitGroups:  clCommitGroups,
		Commits:       clCommits,
		MergeCommits:  clMergeCommits,
		RevertCommits: clRevertCommits,
		NoteGroups:    noteGroups,
	}

	return unreleased, nil
}

func (gen *Generator) getTags(query string) ([]*git.Tag, string, error) {
	tags, err := gen.gitService.GetAllTagsWithDetails()
	if err != nil {
		return nil, "", err
	}

	next := gen.config.Options.NextTag
	if next != "" {
		for _, tag := range tags {
			if next == tag.Name {
				return nil, "", fmt.Errorf("\"%s\" tag already exists", next)
			}
		}

		var previous *git.RelateTag
		if len(tags) > 0 {
			previous = &git.RelateTag{
				Name:    tags[0].Name,
				Subject: tags[0].Subject,
				Date:    tags[0].Date,
			}
		}

		// Assign the date with `readVersions()`
		tags = append([]*git.Tag{
			{
				Name:     next,
				Subject:  next,
				Previous: previous,
			},
		}, tags...)
	}

	if len(tags) == 0 {
		return nil, "", errors.New("git-tag does not exist")
	}

	first := ""
	if query != "" {
		tags, first, err = gen.gitService.SelectTagsWithQuery(tags, query)
		if err != nil {
			if errors.Is(err, git.ErrNotFoundTag) {
				return nil, "", fmt.Errorf("commits corresponding to \"%s\" was not found", query)
			}
			return nil, "", err
		}
	}

	return tags, first, nil
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

func (gen *Generator) render(w io.Writer, unreleased *ChangelogUnreleased, versions []*ChangelogVersion) error {
	if _, err := os.Stat(gen.config.Template); err != nil {
		return err
	}

	fmap := template.FuncMap{
		// format the input time according to layout
		"datetime": func(layout string, input time.Time) string {
			return input.Format(layout)
		},
		// upper case the first character of a string
		"upperFirst": func(s string) string {
			if len(s) > 0 {
				return strings.ToUpper(string(s[0])) + s[1:]
			}
			return ""
		},
		// indent all lines of s n spaces
		"indent": func(s string, n int) string {
			if len(s) == 0 {
				return ""
			}
			pad := strings.Repeat(" ", n)
			return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
		},
		// While Sprig provides these functions, they change the standard input
		// order which leads to a regression. For an example see:
		// https://github.com/Masterminds/sprig/blob/master/functions.go#L149
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"replace":   strings.Replace,
	}

	fname := filepath.Base(gen.config.Template)

	t := template.Must(template.New(fname).Funcs(sprig.TxtFuncMap()).Funcs(fmap).ParseFiles(gen.config.Template))

	return t.Execute(w, &RenderData{
		Info:       gen.config.Info,
		Unreleased: unreleased,
		Versions:   versions,
	})
}
