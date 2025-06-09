package changelog

import (
	"clikd/internal/services/git"
	"time"
)

// ChangelogCommit ist ein Wrapper um git.Commit für Changelog-spezifische Funktionalität
type ChangelogCommit struct {
	*git.Commit
}

// Note für Commits
type Note struct {
	Title string
	Body  string
}

// NoteGroup ist eine Sammlung von Notes, gruppiert nach Titeln
type NoteGroup struct {
	Title string
	Notes []*Note
}

// Revert-Informationen für einen Commit
type Revert struct {
	Header string
}

// Tag enthält Daten eines Git-Tags
type Tag struct {
	Name     string
	Subject  string
	Date     time.Time
	Next     *RelateTag
	Previous *RelateTag
}

// RelateTag enthält Beziehungsinformationen zu einem Tag
type RelateTag struct {
	Name    string
	Subject string
	Date    time.Time
}

// ChangelogVersion ist ein Tag-separierter Datensatz für das CHANGELOG
type ChangelogVersion struct {
	Tag           *git.Tag
	CommitGroups  []*ChangelogCommitGroup
	Commits       []*ChangelogCommit
	MergeCommits  []*ChangelogCommit
	RevertCommits []*ChangelogCommit
	NoteGroups    []*git.NoteGroup
}

// ChangelogCommitGroup ist eine Sammlung von Commits, gruppiert nach dem CommitGroupBy-Parameter
type ChangelogCommitGroup struct {
	RawTitle string // Ursprünglicher Titel vor der Konvertierung (z.B. `build`)
	Title    string // Titel nach der Konvertierung (z.B. `Build`)
	Commits  []*ChangelogCommit
}

// ChangelogUnreleased enthält nicht freigegebene Commit-Daten
type ChangelogUnreleased struct {
	CommitGroups  []*ChangelogCommitGroup
	Commits       []*ChangelogCommit
	MergeCommits  []*ChangelogCommit
	RevertCommits []*ChangelogCommit
	NoteGroups    []*git.NoteGroup
}

// Processor interface für Commit-Verarbeitung
type Processor interface {
	Bootstrap(config *Config) error
	ProcessCommit(commit *ChangelogCommit) *ChangelogCommit
}

// ChangelogAnswer enthält die Antworten des Benutzers für die Initialisierung
type ChangelogAnswer struct {
	RepositoryURL       string `survey:"repository_url"`
	Style               string `survey:"style"`
	CommitMessageFormat string `survey:"commit_message_format"`
	Template            string `survey:"template"`
	IncludeMerges       bool   `survey:"include_merges"`
	IncludeReverts      bool   `survey:"include_reverts"`
	ConfigDir           string `survey:"config_dir"`
}

// Answer ist ein Alias für ChangelogAnswer für Kompatibilität mit Builder-Interfaces
type Answer = ChangelogAnswer

// Config for generating CHANGELOG
type Config struct {
	Bin        string // Git execution command
	WorkingDir string // Working directory
	Template   string // Path for template file. If a relative path is specified, it depends on the value of `WorkingDir`.
	Info       *Info
	Options    *Options
}

// Info is metadata related to CHANGELOG
type Info struct {
	Title         string // Title of CHANGELOG
	RepositoryURL string // URL of git repository
}

// RenderData is the data passed to the template
type RenderData struct {
	Info       *Info
	Unreleased *ChangelogUnreleased
	Versions   []*ChangelogVersion
}

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
