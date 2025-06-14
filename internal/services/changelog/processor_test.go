package changelog

import (
	"testing"

	"clikd/internal/services/git"

	"github.com/stretchr/testify/assert"
)

func TestGitHubProcessor(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Info: &Info{
			RepositoryURL: "https://example.com",
		},
	}

	processor := &GitHubProcessor{}

	processor.Bootstrap(config)

	assert.Equal(
		&ChangelogCommit{
			Commit: &git.Commit{
				Header:  "message [@foo](https://github.com/foo) [#123](https://example.com/issues/123)",
				Subject: "message [@foo](https://github.com/foo) [#123](https://example.com/issues/123)",
				Body: `issue [#456](https://example.com/issues/456)
multiline [#789](https://example.com/issues/789)
[@foo](https://github.com/foo), [@bar](https://github.com/bar)`,
				Notes: []*git.Note{
					{
						Body: `issue1 [#11](https://example.com/issues/11)
issue2 [#22](https://example.com/issues/22)
[gh-56](https://example.com/issues/56) hoge fuga`,
					},
				},
			},
		},
		processor.ProcessCommit(
			&ChangelogCommit{
				Commit: &git.Commit{
					Header:  "message @foo #123",
					Subject: "message @foo #123",
					Body: `issue #456
multiline #789
@foo, @bar`,
					Notes: []*git.Note{
						{
							Body: `issue1 #11
issue2 #22
gh-56 hoge fuga`,
						},
					},
				},
			},
		),
	)

	assert.Equal(
		&ChangelogCommit{
			Commit: &git.Commit{
				Revert: &git.Revert{
					Header: "revert header [@mention](https://github.com/mention) [#123](https://example.com/issues/123)",
				},
			},
		},
		processor.ProcessCommit(
			&ChangelogCommit{
				Commit: &git.Commit{
					Revert: &git.Revert{
						Header: "revert header @mention #123",
					},
				},
			},
		),
	)
}

func TestGitLabProcessor(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Info: &Info{
			RepositoryURL: "https://example.com",
		},
	}

	processor := &GitLabProcessor{}

	processor.Bootstrap(config)

	assert.Equal(
		&ChangelogCommit{
			Commit: &git.Commit{
				Header:  "message [@foo](https://gitlab.com/foo) [#123](https://example.com/issues/123) [!345](https://example.com/merge_requests/345)",
				Subject: "message [@foo](https://gitlab.com/foo) [#123](https://example.com/issues/123) [!345](https://example.com/merge_requests/345)",
				Body: `issue [#456](https://example.com/issues/456)
multiline [#789](https://example.com/issues/789)
merge request [!345](https://example.com/merge_requests/345)
[@foo](https://gitlab.com/foo), [@bar](https://gitlab.com/bar)`,
				Notes: []*git.Note{
					{
						Body: `issue1 [#11](https://example.com/issues/11) [!33](https://example.com/merge_requests/33)
issue2 [#22](https://example.com/issues/22)
merge request [!33](https://example.com/merge_requests/33)
gh-56 hoge fuga`,
					},
				},
			},
		},
		processor.ProcessCommit(
			&ChangelogCommit{
				Commit: &git.Commit{
					Header:  "message @foo #123 !345",
					Subject: "message @foo #123 !345",
					Body: `issue #456
multiline #789
merge request !345
@foo, @bar`,
					Notes: []*git.Note{
						{
							Body: `issue1 #11 !33
issue2 #22
merge request !33
gh-56 hoge fuga`,
						},
					},
				},
			},
		),
	)

	assert.Equal(
		&ChangelogCommit{
			Commit: &git.Commit{
				Revert: &git.Revert{
					Header: "revert header [@mention](https://gitlab.com/mention) [#123](https://example.com/issues/123) [!345](https://example.com/merge_requests/345)",
				},
			},
		},
		processor.ProcessCommit(
			&ChangelogCommit{
				Commit: &git.Commit{
					Revert: &git.Revert{
						Header: "revert header @mention #123 !345",
					},
				},
			},
		),
	)
}

func TestBitbucketProcessor(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Info: &Info{
			RepositoryURL: "https://example.com",
		},
	}

	processor := &BitbucketProcessor{}

	processor.Bootstrap(config)

	assert.Equal(
		&ChangelogCommit{
			Commit: &git.Commit{
				Header:  "message [@foo](https://bitbucket.org/foo/) [#123](https://example.com/issues/123/)",
				Subject: "message [@foo](https://bitbucket.org/foo/) [#123](https://example.com/issues/123/)",
				Body: `issue [#456](https://example.com/issues/456/)
multiline [#789](https://example.com/issues/789/)
[@foo](https://bitbucket.org/foo/), [@bar](https://bitbucket.org/bar/)`,
				Notes: []*git.Note{
					{
						Body: `issue1 [#11](https://example.com/issues/11/)
issue2 [#22](https://example.com/issues/22/)
gh-56 hoge fuga`,
					},
				},
			},
		},
		processor.ProcessCommit(
			&ChangelogCommit{
				Commit: &git.Commit{
					Header:  "message @foo #123",
					Subject: "message @foo #123",
					Body: `issue #456
multiline #789
@foo, @bar`,
					Notes: []*git.Note{
						{
							Body: `issue1 #11
issue2 #22
gh-56 hoge fuga`,
						},
					},
				},
			},
		),
	)

	assert.Equal(
		&ChangelogCommit{
			Commit: &git.Commit{
				Revert: &git.Revert{
					Header: "revert header [@mention](https://bitbucket.org/mention/) [#123](https://example.com/issues/123/)",
				},
			},
		},
		processor.ProcessCommit(
			&ChangelogCommit{
				Commit: &git.Commit{
					Revert: &git.Revert{
						Header: "revert header @mention #123",
					},
				},
			},
		),
	)
}
