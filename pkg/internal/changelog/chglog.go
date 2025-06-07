// Package chglog implements main logic for the CHANGELOG generate.
package changelog

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/tsuyoshiwada/go-gitcmd"
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
	Unreleased *Unreleased
	Versions   []*Version
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
	client          gitcmd.Client
	config          *Config
	tagReader       *tagReader
	tagSelector     *tagSelector
	commitParser    *commitParser
	commitExtractor *commitExtractor
}

// NewGenerator receives `Config` and create an new `Generator`
func NewGenerator(logger *Logger, config *Config) *Generator {
	// Zielrepository-Verzeichnis
	repoDir := "/Users/nyxb/Projects/nyxb/cli/clikd/test_repo"

	// Erstelle einen benutzerdefinierten Git-Client, der im Zielrepository arbeitet
	client := newCustomGitClient(&gitcmd.Config{
		Bin: config.Bin,
	}, repoDir)

	jiraClient := NewJiraClient(config)

	if config.Options.Processor != nil {
		config.Options.Processor.Bootstrap(config)
	}

	normalizeConfig(config)

	return &Generator{
		client:          client,
		config:          config,
		tagReader:       newTagReader(client, config.Options.TagFilterPattern, config.Options.Sort),
		tagSelector:     newTagSelector(),
		commitParser:    newCommitParser(logger, client, jiraClient, config),
		commitExtractor: newCommitExtractor(config.Options),
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

	// Überprüfen des aktuellen Arbeitsverzeichnisses
	if currWd, err := os.Getwd(); err == nil {
		fmt.Printf("DEBUG: Current working directory in Generate: %s\n", currWd)
	}

	// Direkt mit der Tag-Verarbeitung fortfahren, da unser customGitClient
	// bereits das richtige Verzeichnis verwendet
	tags, first, err := gen.getTags(query)
	if err != nil {
		fmt.Printf("DEBUG: Error in getTags: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: getTags returned %d tags, first=%s\n", len(tags), first)

	// Debug: Zeige detaillierte Informationen zu jedem Tag
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

	// Debug: Zeige Details zu unreleased Commits
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

		// Debug: Zeige Details zu den Commits in dieser Version
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

func (gen *Generator) readVersions(tags []*Tag, first string) ([]*Version, error) {
	fmt.Printf("DEBUG: readVersions called with %d tags, first=%s\n", len(tags), first)

	next := gen.config.Options.NextTag
	versions := []*Version{}

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

		// Debug: Führe einen Git-Befehl aus, um zu überprüfen, welche Commits tatsächlich vorhanden sind
		gitArgs := []string{"-C", "/Users/nyxb/Projects/nyxb/cli/clikd/test_repo", "log", "--pretty=format:%h - %s", rev}
		fmt.Printf("DEBUG: Executing git %s\n", strings.Join(gitArgs, " "))
		if out, err := exec.Command("git", gitArgs...).CombinedOutput(); err == nil {
			fmt.Printf("DEBUG: Git log for rev %s:\n%s\n", rev, string(out))
		} else {
			fmt.Printf("DEBUG: Error executing git log for rev %s: %v\n%s\n", rev, err, string(out))
		}

		commits, err := gen.commitParser.Parse(rev)
		if err != nil {
			fmt.Printf("DEBUG: Error parsing commits for rev=%s: %v\n", rev, err)
			return nil, err
		}
		fmt.Printf("DEBUG: Found %d commits for rev=%s\n", len(commits), rev)

		// Debug: Detaillierte Informationen zu jedem gefundenen Commit
		for j, commit := range commits {
			fmt.Printf("DEBUG: commit[%d]: hash=%s, subject=%s, type=%s, scope=%s\n",
				j, commit.Hash.Short, commit.Subject, commit.Type, commit.Scope)
		}

		commitGroups, mergeCommits, revertCommits, noteGroups := gen.commitExtractor.Extract(commits)
		fmt.Printf("DEBUG: Extracted %d commitGroups, %d mergeCommits, %d revertCommits, %d noteGroups\n",
			len(commitGroups), len(mergeCommits), len(revertCommits), len(noteGroups))

		// Debug: Zeige Informationen zu den extrahierten Commit-Gruppen
		for j, group := range commitGroups {
			fmt.Printf("DEBUG: commitGroup[%d]: title=%s, commits=%d\n",
				j, group.Title, len(group.Commits))
		}

		versions = append(versions, &Version{
			Tag:           tag,
			CommitGroups:  commitGroups,
			Commits:       commits,
			MergeCommits:  mergeCommits,
			RevertCommits: revertCommits,
			NoteGroups:    noteGroups,
		})

		// Instead of `getTags()`, assign the date to the tag
		if isNext && len(commits) != 0 {
			tag.Date = commits[0].Author.Date
		}
	}

	return versions, nil
}

func (gen *Generator) readUnreleased(tags []*Tag) (*Unreleased, error) {
	fmt.Printf("DEBUG: readUnreleased called with %d tags\n", len(tags))

	if gen.config.Options.NextTag != "" {
		fmt.Printf("DEBUG: NextTag is set to %q, returning empty Unreleased\n", gen.config.Options.NextTag)
		return &Unreleased{}, nil
	}

	rev := "HEAD"

	if len(tags) > 0 {
		rev = tags[0].Name + "..HEAD"
		fmt.Printf("DEBUG: Setting rev to %q using the latest tag\n", rev)
	} else {
		fmt.Printf("DEBUG: No tags found, using rev=%q\n", rev)
	}

	// Debug: Führe einen Git-Befehl aus, um zu überprüfen, welche Commits tatsächlich vorhanden sind
	gitArgs := []string{"-C", "/Users/nyxb/Projects/nyxb/cli/clikd/test_repo", "log", "--pretty=format:%h - %s", rev}
	fmt.Printf("DEBUG: Executing git %s\n", strings.Join(gitArgs, " "))
	if out, err := exec.Command("git", gitArgs...).CombinedOutput(); err == nil {
		if len(out) > 0 {
			fmt.Printf("DEBUG: Git log for unreleased commits (rev=%s):\n%s\n", rev, string(out))
		} else {
			fmt.Printf("DEBUG: No unreleased commits found (rev=%s)\n", rev)
		}
	} else {
		fmt.Printf("DEBUG: Error executing git log for unreleased commits: %v\n%s\n", err, string(out))
	}

	commits, err := gen.commitParser.Parse(rev)
	if err != nil {
		fmt.Printf("DEBUG: Error parsing unreleased commits for rev=%s: %v\n", rev, err)
		return nil, err
	}
	fmt.Printf("DEBUG: Found %d unreleased commits for rev=%s\n", len(commits), rev)

	// Debug: Detaillierte Informationen zu jedem gefundenen Commit
	for i, commit := range commits {
		fmt.Printf("DEBUG: unreleased commit[%d]: hash=%s, subject=%s, type=%s, scope=%s\n",
			i, commit.Hash.Short, commit.Subject, commit.Type, commit.Scope)
	}

	commitGroups, mergeCommits, revertCommits, noteGroups := gen.commitExtractor.Extract(commits)
	fmt.Printf("DEBUG: Extracted %d commitGroups, %d mergeCommits, %d revertCommits, %d noteGroups\n",
		len(commitGroups), len(mergeCommits), len(revertCommits), len(noteGroups))

	// Debug: Zeige Informationen zu den extrahierten Commit-Gruppen
	for i, group := range commitGroups {
		fmt.Printf("DEBUG: unreleased commitGroup[%d]: title=%s, commits=%d\n",
			i, group.Title, len(group.Commits))
	}

	unreleased := &Unreleased{
		CommitGroups:  commitGroups,
		Commits:       commits,
		MergeCommits:  mergeCommits,
		RevertCommits: revertCommits,
		NoteGroups:    noteGroups,
	}

	return unreleased, nil
}

func (gen *Generator) getTags(query string) ([]*Tag, string, error) {
	fmt.Printf("DEBUG: getTags called with query: %q\n", query)

	// Debug: Überprüfen, ob das Git-Arbeitsverzeichnis korrekt gesetzt ist
	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("DEBUG: Current working directory in getTags: %s\n", wd)
	}

	// Debug: Direkter Git-Befehl, um die verfügbaren Tags zu überprüfen
	gitArgs := []string{"-C", "/Users/nyxb/Projects/nyxb/cli/clikd/test_repo", "tag", "-l", "--sort=-creatordate"}
	fmt.Printf("DEBUG: Executing git %s\n", strings.Join(gitArgs, " "))
	if out, err := exec.Command("git", gitArgs...).CombinedOutput(); err == nil {
		fmt.Printf("DEBUG: Available git tags sorted by date:\n%s\n", string(out))
	} else {
		fmt.Printf("DEBUG: Error executing git tag command: %v\n%s\n", err, string(out))
	}

	tags, err := gen.tagReader.ReadAll()
	if err != nil {
		fmt.Printf("DEBUG: Error in tagReader.ReadAll: %v\n", err)
		return nil, "", err
	}
	fmt.Printf("DEBUG: tagReader.ReadAll returned %d tags\n", len(tags))

	// Debug: Detaillierte Informationen zu jedem Tag
	for i, tag := range tags {
		fmt.Printf("DEBUG: ReadAll tag[%d]: name=%s, date=%s\n",
			i, tag.Name, tag.Date.Format("2006-01-02 15:04:05"))
	}

	next := gen.config.Options.NextTag
	if next != "" {
		fmt.Printf("DEBUG: NextTag option is set to %q\n", next)
		for _, tag := range tags {
			if next == tag.Name {
				fmt.Printf("DEBUG: NextTag %q already exists as a tag\n", next)
				return nil, "", fmt.Errorf("\"%s\" tag already exists", next)
			}
		}

		var previous *RelateTag
		if len(tags) > 0 {
			previous = &RelateTag{
				Name:    tags[0].Name,
				Subject: tags[0].Subject,
				Date:    tags[0].Date,
			}
			fmt.Printf("DEBUG: Setting previous tag for NextTag: %s\n", previous.Name)
		} else {
			fmt.Printf("DEBUG: No previous tag found for NextTag\n")
		}

		// Assign the date with `readVersions()`
		tags = append([]*Tag{
			{
				Name:     next,
				Subject:  next,
				Previous: previous,
			},
		}, tags...)
		fmt.Printf("DEBUG: Added NextTag %q to tags list\n", next)
	}

	// IMPORTANT - parse and convert query to tag and revision
	fmt.Printf("DEBUG: Parsing query %q\n", query)
	first := ""
	tgs := tags

	if query != "" {
		tags, first, err = gen.tagSelector.Select(tags, query)
		if err != nil {
			fmt.Printf("DEBUG: Error in tagSelector.Select: %v\n", err)
			return nil, "", err
		}
		fmt.Printf("DEBUG: tagSelector.Select returned %d tags, first=%s\n", len(tags), first)

		// Debug: Zeige die ausgewählten Tags
		for i, tag := range tags {
			fmt.Printf("DEBUG: Selected tag[%d]: %s\n", i, tag.Name)
		}
	} else {
		fmt.Printf("DEBUG: No query provided, using all %d tags\n", len(tags))
	}

	if len(tags) == 0 {
		// Use the oldest tag if first is empty
		if first == "" && len(tgs) != 0 {
			first = tgs[len(tgs)-1].Name
			fmt.Printf("DEBUG: No tags selected, using oldest tag as first: %s\n", first)
		}
	}

	fmt.Printf("DEBUG: getTags returning %d tags, first=%s\n", len(tags), first)
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

func (gen *Generator) render(w io.Writer, unreleased *Unreleased, versions []*Version) error {
	// Überprüfen, ob die Template-Datei existiert und kein Verzeichnis ist
	fileInfo, err := os.Stat(gen.config.Template)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("read %s: is a directory", gen.config.Template)
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

// customGitClient ist ein Wrapper um gitcmd.Client, der alle Befehle im korrekten Repository-Verzeichnis ausführt
type customGitClient struct {
	wrapped gitcmd.Client
	repoDir string
}

// Exec führt den Git-Befehl im Repository-Verzeichnis aus
func (c *customGitClient) Exec(subcmd string, args ...string) (string, error) {
	fmt.Printf("DEBUG: customGitClient.Exec called with subcmd=%q, args=%v\n", subcmd, args)

	// Ursprüngliches Arbeitsverzeichnis speichern
	origWd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	fmt.Printf("DEBUG: customGitClient original working directory: %s\n", origWd)

	// In das Repository-Verzeichnis wechseln
	if err := os.Chdir(c.repoDir); err != nil {
		return "", fmt.Errorf("failed to change to repository directory: %w", err)
	}
	fmt.Printf("DEBUG: customGitClient changed to repository directory: %s\n", c.repoDir)

	// Git-Befehl ausführen
	out, err := c.wrapped.Exec(subcmd, args...)

	// Zum ursprünglichen Verzeichnis zurückkehren
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		fmt.Printf("DEBUG: WARNING: Failed to restore original directory: %v\n", chdirErr)
		// Falls der Git-Befehl erfolgreich war, geben wir den Fehler beim Verzeichniswechsel nicht zurück
		if err == nil {
			err = chdirErr
		}
	} else {
		fmt.Printf("DEBUG: customGitClient restored original directory: %s\n", origWd)
	}

	// Debug-Ausgabe für das Ergebnis
	if err != nil {
		fmt.Printf("DEBUG: customGitClient git command failed: %v\n", err)
	} else {
		preview := out
		if len(out) > 100 {
			preview = out[:100] + "..."
		}
		fmt.Printf("DEBUG: customGitClient git command output (first 100 chars): %s\n", preview)
	}

	return out, err
}

// CanExec prüft, ob der Git-Befehl ausgeführt werden kann
func (c *customGitClient) CanExec() error {
	return c.wrapped.CanExec()
}

// InsideWorkTree prüft, ob wir uns innerhalb eines Git-Arbeitsbaumes befinden
func (c *customGitClient) InsideWorkTree() error {
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
	result := c.wrapped.InsideWorkTree()

	// Zum ursprünglichen Verzeichnis zurückkehren
	if chdirErr := os.Chdir(origWd); chdirErr != nil {
		// Falls die InsideWorkTree-Prüfung erfolgreich war, geben wir den Fehler beim Verzeichniswechsel zurück
		if result == nil {
			return chdirErr
		}
	}

	return result
}

// newCustomGitClient erstellt einen neuen Git-Client-Wrapper
func newCustomGitClient(config *gitcmd.Config, repoDir string) gitcmd.Client {
	return &customGitClient{
		wrapped: gitcmd.New(config),
		repoDir: repoDir,
	}
}
