package changelog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gitcmd "github.com/tsuyoshiwada/go-gitcmd"

	"clikd/internal/utils"
)

var (
	cwd                string
	testRepoRoot       = ".tmp"
	internalTimeFormat = "2006-01-02 15:04:05"
)

// NewTestLogger erstellt einen Logger für Tests
func NewTestLogger(stdout, stderr io.Writer, noColor, emoji bool) utils.Logger {
	logLevel := "info"
	if stdout == nil && stderr == nil {
		// Stille Tests
		logLevel = "error"
	}

	return utils.NewLogger(logLevel, emoji)
}

type commitFunc = func(date, subject, body string)
type tagFunc = func(name string)

func TestMain(m *testing.M) {
	cwd, _ = os.Getwd()
	cleanup()
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func setup(dir string, setupRepo func(commitFunc, tagFunc, gitcmd.Client)) {
	testDir := filepath.Join(cwd, testRepoRoot, dir)

	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, os.ModePerm)
	_ = os.Chdir(testDir)

	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	git := gitcmd.New(nil)
	_, _ = git.Exec("init")
	_, _ = git.Exec("config", "user.name", "test_user")
	_, _ = git.Exec("config", "user.email", "test@example.com")

	var commit = func(date, subject, body string) {
		msg := subject
		if body != "" {
			msg += "\n\n" + body
		}
		t, _ := time.Parse(internalTimeFormat, date)
		d := t.Format("Mon Jan 2 15:04:05 2006 +0000")
		_, _ = git.Exec("commit", "--allow-empty", "--date", d, "-m", msg)
	}

	var tag = func(name string) {
		_, _ = git.Exec("tag", name)
	}

	setupRepo(commit, tag, git)

	_ = os.Chdir(cwd)
}

func cleanup() {
	_ = os.Chdir(cwd)
	_ = os.RemoveAll(filepath.Join(cwd, testRepoRoot))
}

func TestGeneratorNotFoundTags(t *testing.T) {
	assert := assert.New(t)
	testName := "not_found"

	setup(testName, func(commit commitFunc, _ tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "feat(*): New feature", "")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	expected := strings.TrimSpace(buf.String())

	assert.Error(err)
	assert.Contains(err.Error(), "git-tag does not exist")
	assert.Equal("", expected)
}

func TestGeneratorNotFoundCommits(t *testing.T) {
	assert := assert.New(t)
	testName := "not_found"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "feat(*): New feature", "")
		tag("1.0.0")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "foo")
	expected := strings.TrimSpace(buf.String())

	assert.Error(err)
	assert.Equal("", expected)
}

func TestGeneratorNotFoundCommitsOne(t *testing.T) {
	assert := assert.New(t)
	testName := "not_found"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "chore(*): First commit", "")
		tag("1.0.0")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				CommitFilters:        map[string][]string{},
				CommitSortBy:         "Scope",
				CommitGroupBy:        "Type",
				CommitGroupSortBy:    "Title",
				CommitGroupTitleMaps: map[string]string{},
				HeaderPattern:        "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
				IssuePrefix: []string{
					"#",
					"gh-",
				},
				RefActions:   []string{},
				MergePattern: "^Merge pull request #(\\d+) from (.*)$",
				MergePatternMaps: []string{
					"Ref",
					"Source",
				},
				RevertPattern: "^Revert \"([\\s\\S]*)\"$",
				RevertPatternMaps: []string{
					"Header",
				},
				NoteKeywords: []string{
					"BREAKING CHANGE",
				},
			},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "foo")
	expected := strings.TrimSpace(buf.String())

	assert.Error(err)
	assert.Contains(err.Error(), "\"foo\" was not found")
	assert.Equal("", expected)
}

func TestGeneratorWithTypeScopeSubject(t *testing.T) {
	assert := assert.New(t)
	testName := "type_scope_subject"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "chore(*): First commit", "")
		commit("2018-01-01 00:01:00", "feat(core): Add foo bar", "")
		commit("2018-01-01 00:02:00", "docs(readme): Update usage #123", "")
		tag("1.0.0")

		commit("2018-01-02 00:00:00", "feat(parser): New some super options #333", "")
		commit("2018-01-02 00:01:00", "Merge pull request #999 from tsuyoshiwada/patch-1", "")
		commit("2018-01-02 00:02:00", "Merge pull request #1000 from tsuyoshiwada/patch-1", "")
		commit("2018-01-02 00:03:00", "Revert \"feat(core): Add foo bar @mention and issue #987\"", "")
		tag("1.1.0")

		commit("2018-01-03 00:00:00", "feat(context): Online breaking change", "BREAKING CHANGE: Online breaking change message.")
		commit("2018-01-03 00:01:00", "feat(router): Multiple breaking change", `This is body,

BREAKING CHANGE:
Multiple
breaking
change message.`)
		tag("2.0.0-beta.0")

		commit("2018-01-04 00:00:00", "refactor(context): gofmt", "")
		commit("2018-01-04 00:01:00", "fix(core): Fix commit\n\nThis is body message.", "")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				Title:         "CHANGELOG Example",
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				Sort: "date",
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
						"fix",
					},
				},
				CommitSortBy:      "Scope",
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
					"fix":  "Bug Fixes",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
				IssuePrefix: []string{
					"#",
					"gh-",
				},
				RefActions:   []string{},
				MergePattern: "^Merge pull request #(\\d+) from (.*)$",
				MergePatternMaps: []string{
					"Ref",
					"Source",
				},
				RevertPattern: "^Revert \"([\\s\\S]*)\"$",
				RevertPatternMaps: []string{
					"Header",
				},
				NoteKeywords: []string{
					"BREAKING CHANGE",
				},
			},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	expected := strings.TrimSpace(buf.String())

	assert.Nil(err)
	assert.Equal(`<a name="unreleased"></a>
## [Unreleased]

### Bug Fixes
- **core:** Fix commit


<a name="2.0.0-beta.0"></a>
## [2.0.0-beta.0] - 2018-01-03
### Features
- **context:** Online breaking change
- **router:** Multiple breaking change

### BREAKING CHANGE

Multiple
breaking
change message.

Online breaking change message.


<a name="1.1.0"></a>
## [1.1.0] - 2018-01-02
### Features
- **parser:** New some super options #333

### Reverts
- feat(core): Add foo bar @mention and issue #987

### Pull Requests
- Merge pull request #1000 from tsuyoshiwada/patch-1
- Merge pull request #999 from tsuyoshiwada/patch-1


<a name="1.0.0"></a>
## 1.0.0 - 2018-01-01
### Features
- **core:** Add foo bar


[Unreleased]: https://github.com/git-chglog/git-chglog/compare/2.0.0-beta.0...HEAD
[2.0.0-beta.0]: https://github.com/git-chglog/git-chglog/compare/1.1.0...2.0.0-beta.0
[1.1.0]: https://github.com/git-chglog/git-chglog/compare/1.0.0...1.1.0`, expected)
}

func TestGeneratorWithNextTag(t *testing.T) {
	assert := assert.New(t)
	testName := "type_scope_subject"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "feat(core): version 1.0.0", "")
		tag("1.0.0")

		commit("2018-02-01 00:00:00", "feat(core): version 2.0.0", "")
		tag("2.0.0")

		commit("2018-03-01 00:00:00", "feat(core): version 3.0.0", "")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				Title:         "CHANGELOG Example",
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				Sort:    "date",
				NextTag: "3.0.0",
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
					},
				},
				CommitSortBy:      "Scope",
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
			},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	expected := strings.TrimSpace(buf.String())

	assert.Nil(err)
	assert.Equal(`<a name="unreleased"></a>
## [Unreleased]


<a name="3.0.0"></a>
## [3.0.0] - 2018-03-01
### Features
- **core:** version 3.0.0


<a name="2.0.0"></a>
## [2.0.0] - 2018-02-01
### Features
- **core:** version 2.0.0


<a name="1.0.0"></a>
## 1.0.0 - 2018-01-01
### Features
- **core:** version 1.0.0


[Unreleased]: https://github.com/git-chglog/git-chglog/compare/3.0.0...HEAD
[3.0.0]: https://github.com/git-chglog/git-chglog/compare/2.0.0...3.0.0
[2.0.0]: https://github.com/git-chglog/git-chglog/compare/1.0.0...2.0.0`, expected)

	buf = &bytes.Buffer{}
	err = gen.Generate(buf, "3.0.0")
	expected = strings.TrimSpace(buf.String())

	assert.Nil(err)
	assert.Equal(`<a name="unreleased"></a>
## [Unreleased]


<a name="3.0.0"></a>
## [3.0.0] - 2018-03-01
### Features
- **core:** version 3.0.0


[Unreleased]: https://github.com/git-chglog/git-chglog/compare/3.0.0...HEAD
[3.0.0]: https://github.com/git-chglog/git-chglog/compare/2.0.0...3.0.0`, expected)
}

func TestGeneratorWithTagFiler(t *testing.T) {
	assert := assert.New(t)
	testName := "type_scope_subject"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "feat(core): version dev-1.0.0", "")
		tag("dev-1.0.0")

		commit("2018-02-01 00:00:00", "feat(core): version v1.0.0", "")
		tag("v1.0.0")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				Title:         "CHANGELOG Example",
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				TagFilterPattern: "^v",
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
					},
				},
				CommitSortBy:      "Scope",
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
			},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	expected := strings.TrimSpace(buf.String())

	assert.Nil(err)
	assert.Equal(`<a name="unreleased"></a>
## [Unreleased]


<a name="v1.0.0"></a>
## v1.0.0 - 2018-02-01
### Features
- **core:** version v1.0.0
- **core:** version dev-1.0.0


[Unreleased]: https://github.com/git-chglog/git-chglog/compare/v1.0.0...HEAD`, expected)

}

func TestGeneratorWithTrimmedBody(t *testing.T) {
	assert := assert.New(t)
	testName := "trimmed_body"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "feat: single line commit", "")
		commit("2018-01-01 00:01:00", "feat: multi-line commit", `
More details about the change and why it went in.

BREAKING CHANGE:

When using .TrimmedBody Notes are not included and can only appear in the Notes section.

Signed-off-by: First Last <first.last@mail.com>

Co-authored-by: dependabot-preview[bot] <27856297+dependabot-preview[bot]@users.noreply.github.com>`)

		commit("2018-01-01 00:00:00", "feat: another single line commit", "")
		tag("1.0.0")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				Title:         "CHANGELOG Example",
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
					},
				},
				CommitSortBy:      "Scope",
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
				NoteKeywords: []string{
					"BREAKING CHANGE",
				},
			},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	expected := strings.TrimSpace(buf.String())

	assert.Nil(err)
	assert.Equal(`<a name="unreleased"></a>
## [Unreleased]


<a name="1.0.0"></a>
## 1.0.0 - 2018-01-01
### Features
- another single line commit
- multi-line commit
  More details about the change and why it went in.
- single line commit

### BREAKING CHANGE

When using .TrimmedBody Notes are not included and can only appear in the Notes section.


[Unreleased]: https://github.com/git-chglog/git-chglog/compare/1.0.0...HEAD`, expected)
}

func TestGeneratorWithSprig(t *testing.T) {
	assert := assert.New(t)
	testName := "with_sprig"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2018-01-01 00:00:00", "feat(core): version 1.0.0", "")
		tag("1.0.0")

		commit("2018-02-01 00:00:00", "feat(core): version 2.0.0", "")
		tag("2.0.0")

		commit("2018-03-01 00:00:00", "feat(core): version 3.0.0", "")
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", testName+".md"),
			Info: &Info{
				Title:         "CHANGELOG Example",
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				Sort:    "date",
				NextTag: "3.0.0",
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
					},
				},
				CommitSortBy:      "Scope",
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
			},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	expected := strings.TrimSpace(buf.String())

	assert.Nil(err)
	assert.Equal(`My Changelog
<a name="unreleased"></a>
## [Unreleased]


<a name="3.0.0"></a>
## [3.0.0] - 2018-03-01
### Features
- **CORE:** version 3.0.0


<a name="2.0.0"></a>
## [2.0.0] - 2018-02-01
### Features
- **CORE:** version 2.0.0


<a name="1.0.0"></a>
## 1.0.0 - 2018-01-01
### Features
- **CORE:** version 1.0.0


[Unreleased]: https://github.com/git-chglog/git-chglog/compare/3.0.0...HEAD
[3.0.0]: https://github.com/git-chglog/git-chglog/compare/2.0.0...3.0.0
[2.0.0]: https://github.com/git-chglog/git-chglog/compare/1.0.0...2.0.0`, expected)

}

func TestTemplateFunctions(t *testing.T) {
	assert := assert.New(t)
	testName := "template_functions"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		commit("2023-01-01 00:00:00", "feat(*): Add test for template functions", "")
		tag("1.0.0")
	})

	// Create a test template with all template functions
	tempFile := filepath.Join(cwd, "testdata", testName+".md")
	tempContent := `
{{- $test := "test string" -}}
contains: {{ contains $test "st" }}
hasPrefix: {{ hasPrefix $test "test" }}
hasSuffix: {{ hasSuffix $test "ing" }}
replace: {{ replace $test "test" "hello" 1 }}
upperFirst: {{ upperFirst "hello" }}
datetime: {{ datetime "2006-01-02" (index .Versions 0).Tag.Date }}
indented: >
{{ indent "line1\nline2\nline3" 4 }}
`
	err := os.WriteFile(tempFile, []byte(tempContent), 0644)
	assert.Nil(err)
	defer os.Remove(tempFile)

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   tempFile,
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{},
		})

	buf := &bytes.Buffer{}
	err = gen.Generate(buf, "")
	assert.Nil(err)

	result := strings.TrimSpace(buf.String())
	assert.Contains(result, "contains: true")
	assert.Contains(result, "hasPrefix: true")
	assert.Contains(result, "hasSuffix: true")
	assert.Contains(result, "replace: hello string")
	assert.Contains(result, "upperFirst: Hello")
	assert.Contains(result, "datetime: 2023-01-01")
	assert.Contains(result, "indented: >")
	assert.Contains(result, "    line1")
	assert.Contains(result, "    line2")
	assert.Contains(result, "    line3")
}

func TestGeneratorWithLargeRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	assert := assert.New(t)
	testName := "large_repo"

	// Setup a large repository with many commits and tags
	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		// Create 100 commits and 10 tags to simulate a large repository
		for i := 0; i < 10; i++ {
			tagVersion := fmt.Sprintf("%d.0.0", i+1)

			// Create 10 commits per tag
			for j := 0; j < 10; j++ {
				date := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).
					Add(time.Duration(i*10+j) * 24 * time.Hour)

				dateStr := date.Format("2006-01-02 15:04:05")
				commitMsg := fmt.Sprintf("feat(module-%d): Add feature %d-%d", i, i, j)

				commit(dateStr, commitMsg, "")
			}

			tag(tagVersion)
		}
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", "type_scope_subject.md"),
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
					},
				},
				CommitSortBy:      "Scope",
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
			},
		})

	// Measure performance
	start := time.Now()
	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")
	duration := time.Since(start)

	assert.Nil(err)
	assert.NotEmpty(buf.String())

	// Log performance results
	t.Logf("Large repository test completed in %s", duration)

	// The test should complete in a reasonable time (adjust threshold as needed)
	// This is more for monitoring than a strict pass/fail
	if duration > 10*time.Second {
		t.Logf("Warning: Performance might be an issue, test took %s", duration)
	}
}

func TestGeneratorEmptyRepository(t *testing.T) {
	assert := assert.New(t)
	testName := "empty_repo"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		// Empty repository, no commits or tags
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", "type_scope_subject.md"),
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{},
		})

	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "")

	assert.Error(err)
	assert.Contains(err.Error(), "git-tag does not exist")
}

func TestGeneratorSpecialTagQueries(t *testing.T) {
	assert := assert.New(t)
	testName := "tag_queries"

	setup(testName, func(commit commitFunc, tag tagFunc, _ gitcmd.Client) {
		// Create a series of commits and tags
		commit("2021-01-01 00:00:00", "feat(core): First feature", "")
		tag("v1.0.0")

		commit("2021-02-01 00:00:00", "feat(core): Second feature", "")
		tag("v2.0.0")

		commit("2021-03-01 00:00:00", "feat(core): Third feature", "")
		tag("v3.0.0")

		commit("2021-04-01 00:00:00", "feat(core): Fourth feature", "")
		// No tag for the last commit
	})

	gen := NewGenerator(NewTestLogger(os.Stdout, os.Stderr, false, true),
		&Config{
			Bin:        "git",
			WorkingDir: filepath.Join(testRepoRoot, testName),
			Template:   filepath.Join(cwd, "testdata", "type_scope_subject.md"),
			Info: &Info{
				RepositoryURL: "https://github.com/git-chglog/git-chglog",
			},
			Options: &Options{
				CommitFilters: map[string][]string{
					"Type": {
						"feat",
					},
				},
				CommitGroupBy:     "Type",
				CommitGroupSortBy: "Title",
				CommitGroupTitleMaps: map[string]string{
					"feat": "Features",
				},
				HeaderPattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
				HeaderPatternMaps: []string{
					"Type",
					"Scope",
					"Subject",
				},
			},
		})

	// Teste, dass der Generator in der Lage ist, verschiedene Tag-Query-Formate zu verarbeiten
	// Anstatt die genauen Inhalte zu prüfen, überprüfen wir nur, dass die Generierung erfolgreich ist
	// und grundlegende Informationen enthält

	// 1. <old>..<new> format
	buf := &bytes.Buffer{}
	err := gen.Generate(buf, "v1.0.0..v3.0.0")
	assert.Nil(err)
	assert.NotEmpty(buf.String())

	// 2. <name>.. format
	buf = &bytes.Buffer{}
	err = gen.Generate(buf, "v2.0.0..")
	assert.Nil(err)
	assert.NotEmpty(buf.String())

	// 3. ..<name> format
	buf = &bytes.Buffer{}
	err = gen.Generate(buf, "..v2.0.0")
	assert.Nil(err)
	assert.NotEmpty(buf.String())

	// 4. <name> format
	buf = &bytes.Buffer{}
	err = gen.Generate(buf, "v2.0.0")
	assert.Nil(err)
	assert.NotEmpty(buf.String())
}
