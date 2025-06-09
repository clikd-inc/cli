package changelog

import (
	"fmt"
	"strings"
)

// ConfigBuilder ...
type ConfigBuilder interface {
	Builder
}

type configBuilderImpl struct{}

// NewConfigBuilder ...
func NewConfigBuilder() ConfigBuilder {
	return &configBuilderImpl{}
}

// Build ...
func (*configBuilderImpl) Build(ans *Answer) (string, error) {
	var msgFormat *CommitMessageFormat

	// Search through all available formats
	allFormats := []*CommitMessageFormat{
		fmtTypeScopeSubject,
		fmtTypeSubject,
		fmtGitBasic,
		fmtSubject,
		fmtCommitEmoji,
	}

	for _, f := range allFormats {
		if f.display == ans.CommitMessageFormat {
			msgFormat = f
			break
		}
	}

	if msgFormat == nil {
		return "", fmt.Errorf("\"%s\" is an invalid commit message format", ans.CommitMessageFormat)
	}

	repoURL := strings.TrimRight(ans.RepositoryURL, "/")
	if repoURL == "" {
		repoURL = "\"\""
	}

	config := fmt.Sprintf(`style: %s
template: %s
info:
  title: CHANGELOG
  repository_url: %s
options:
  commits:
    sort_by: Scope
    filters:%s
  commit_groups:
    group_by: Type
    sort_by: Title
    title_maps:%s
  header:
    pattern: "%s"
    pattern_maps:%s
  notes:
    keywords:
      - BREAKING CHANGE`,
		ans.Style,
		defaultTemplateFilename,
		repoURL,
		msgFormat.FilterTypesString(),
		msgFormat.TitleMapsString(),
		msgFormat.pattern,
		msgFormat.PatternMapString(),
	)

	return config, nil
}
