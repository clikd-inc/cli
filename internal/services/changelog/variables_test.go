package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitMessageFormatPatternMaps(t *testing.T) {
	assert := assert.New(t)

	f := &CommitMessageFormat{
		patternMaps: []string{
			"Type",
			"Scope",
			"Subject",
		},
	}

	assert.Equal(`
      - Type
      - Scope
      - Subject`, f.PatternMapString())

	f = &CommitMessageFormat{
		patternMaps: []string{},
	}

	assert.Equal(" []", f.PatternMapString())
}

func TestCommitMessageFormatFilterTypes(t *testing.T) {
	assert := assert.New(t)

	f := &CommitMessageFormat{
		typeSamples: []typeSample{
			{"feat", "Features"}, {"fix", "Bug Fixes"},
			{"perf", "Performance Improvements"}, {"refactor", "Code Refactoring"},
		},
	}

	expected := `
      Type:
        - feat
        - fix
        - perf
        - refactor`
	assert.Equal(expected, f.FilterTypesString())

	f = &CommitMessageFormat{
		typeSamples: []typeSample{},
	}

	assert.Equal(" {}", f.FilterTypesString())
}

func TestCommitMessageFormatTitleMaps(t *testing.T) {
	assert := assert.New(t)

	f := &CommitMessageFormat{
		typeSamples: []typeSample{
			{"feat", "Features"}, {"fix", "Bug Fixes"},
			{"perf", "Performance Improvements"}, {"refactor", "Code Refactoring"},
		},
	}

	expected := `
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring`
	assert.Equal(expected, f.TitleMapsString())

	f = &CommitMessageFormat{
		typeSamples: []typeSample{},
	}

	assert.Equal(" {}", f.TitleMapsString())
}
