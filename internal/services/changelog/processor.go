package changelog

import (
	"regexp"
	"strings"
)

// GitHubProcessor is optimized for CHANGELOG used in GitHub
//
// The following processing is performed
//   - Mentions automatic link (@tsuyoshiwada -> [@tsuyoshiwada](https://github.com/tsuyoshiwada))
//   - Automatic link to references (#123 -> [#123](https://github.com/owner/repo/issues/123))
type GitHubProcessor struct {
	Host      string // Host name used for link destination. Note: You must include the protocol (e.g. "https://github.com")
	config    *Config
	reMention *regexp.Regexp
	reIssue   *regexp.Regexp
}

// Bootstrap ...
func (p *GitHubProcessor) Bootstrap(config *Config) error {
	p.config = config

	if p.Host == "" {
		p.Host = "https://github.com"
	} else {
		p.Host = strings.TrimRight(p.Host, "/")
	}

	p.reMention = regexp.MustCompile(`@(\w+)`)
	p.reIssue = regexp.MustCompile(`(?i)(#|gh-)(\d+)`)

	return nil
}

// ProcessCommit ...
func (p *GitHubProcessor) ProcessCommit(commit *ChangelogCommit) *ChangelogCommit {
	commit.Header = p.addLinks(commit.Header)
	commit.Subject = p.addLinks(commit.Subject)
	commit.Body = p.addLinks(commit.Body)

	for _, note := range commit.Notes {
		note.Body = p.addLinks(note.Body)
	}

	if commit.Revert != nil {
		commit.Revert.Header = p.addLinks(commit.Revert.Header)
	}

	return commit
}

func (p *GitHubProcessor) addLinks(input string) string {
	repoURL := strings.TrimRight(p.config.Info.RepositoryURL, "/")

	// mentions
	input = p.reMention.ReplaceAllString(input, "[@$1]("+p.Host+"/$1)")

	// issues
	input = p.reIssue.ReplaceAllString(input, "[$1$2]("+repoURL+"/issues/$2)")

	return input
}

// GitLabProcessor is optimized for CHANGELOG used in GitLab
//
// The following processing is performed
//   - Mentions automatic link (@tsuyoshiwada -> [@tsuyoshiwada](https://gitlab.com/tsuyoshiwada))
//   - Automatic link to references issues (#123 -> [#123](https://gitlab.com/owner/repo/issues/123))
//   - Automatic link to references merge request (!123 -> [#123](https://gitlab.com/owner/repo/merge_requests/123))
type GitLabProcessor struct {
	Host           string // Host name used for link destination. Note: You must include the protocol (e.g. "https://gitlab.com")
	config         *Config
	reMention      *regexp.Regexp
	reIssue        *regexp.Regexp
	reMergeRequest *regexp.Regexp
}

// Bootstrap ...
func (p *GitLabProcessor) Bootstrap(config *Config) error {
	p.config = config

	if p.Host == "" {
		p.Host = "https://gitlab.com"
	} else {
		p.Host = strings.TrimRight(p.Host, "/")
	}

	p.reMention = regexp.MustCompile(`@(\w+)`)
	p.reIssue = regexp.MustCompile(`(?i)#(\d+)`)
	p.reMergeRequest = regexp.MustCompile(`(?i)!(\d+)`)

	return nil
}

// ProcessCommit ...
func (p *GitLabProcessor) ProcessCommit(commit *ChangelogCommit) *ChangelogCommit {
	commit.Header = p.addLinks(commit.Header)
	commit.Subject = p.addLinks(commit.Subject)
	commit.Body = p.addLinks(commit.Body)

	for _, note := range commit.Notes {
		note.Body = p.addLinks(note.Body)
	}

	if commit.Revert != nil {
		commit.Revert.Header = p.addLinks(commit.Revert.Header)
	}

	return commit
}

func (p *GitLabProcessor) addLinks(input string) string {
	repoURL := strings.TrimRight(p.config.Info.RepositoryURL, "/")

	// mentions
	input = p.reMention.ReplaceAllString(input, "[@$1]("+p.Host+"/$1)")

	// issues
	input = p.reIssue.ReplaceAllString(input, "[#$1]("+repoURL+"/issues/$1)")

	// merge requests
	input = p.reMergeRequest.ReplaceAllString(input, "[!$1]("+repoURL+"/merge_requests/$1)")

	return input
}

// BitbucketProcessor is optimized for CHANGELOG used in Bitbucket
//
// The following processing is performed
//   - Mentions automatic link (@tsuyoshiwada -> [@tsuyoshiwada](https://bitbucket.org/tsuyoshiwada/))
//   - Automatic link to references (#123 -> [#123](https://bitbucket.org/owner/repo/issues/123/))
type BitbucketProcessor struct {
	Host      string // Host name used for link destination. Note: You must include the protocol (e.g. "https://bitbucket.org")
	config    *Config
	reMention *regexp.Regexp
	reIssue   *regexp.Regexp
}

// Bootstrap ...
func (p *BitbucketProcessor) Bootstrap(config *Config) error {
	p.config = config

	if p.Host == "" {
		p.Host = "https://bitbucket.org"
	} else {
		p.Host = strings.TrimRight(p.Host, "/")
	}

	p.reMention = regexp.MustCompile(`@(\w+)`)
	p.reIssue = regexp.MustCompile(`(?i)#(\d+)`)

	return nil
}

// ProcessCommit ...
func (p *BitbucketProcessor) ProcessCommit(commit *ChangelogCommit) *ChangelogCommit {
	commit.Header = p.addLinks(commit.Header)
	commit.Subject = p.addLinks(commit.Subject)
	commit.Body = p.addLinks(commit.Body)

	for _, note := range commit.Notes {
		note.Body = p.addLinks(note.Body)
	}

	if commit.Revert != nil {
		commit.Revert.Header = p.addLinks(commit.Revert.Header)
	}

	return commit
}

func (p *BitbucketProcessor) addLinks(input string) string {
	repoURL := strings.TrimRight(p.config.Info.RepositoryURL, "/")

	// mentions
	input = p.reMention.ReplaceAllString(input, "[@$1]("+p.Host+"/$1/)")

	// issues
	input = p.reIssue.ReplaceAllString(input, "[#$1]("+repoURL+"/issues/$1/)")

	return input
}
