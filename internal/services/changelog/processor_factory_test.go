package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessorFactory(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	processor, err := factory.Create(&Config{
		Info: &Info{
			RepositoryURL: "https://example.com/owner/repo",
		},
	})

	assert.Nil(err)
	assert.Nil(processor)
}

func TestProcessorFactoryForGitHub(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	// github.com
	processor, err := factory.Create(&Config{
		Info: &Info{
			RepositoryURL: "https://github.com/owner/repo",
		},
	})

	assert.Nil(err)
	assert.IsType(&GitHubProcessorAdapter{}, processor)
	githubProcessor := processor.(*GitHubProcessorAdapter)
	assert.Equal("https://github.com", githubProcessor.Host)

	// Selbst-gehostetes GitHub wurde entfernt, da Style-Eigenschaft nicht mehr unterstützt wird
}

func TestProcessorFactoryForGitLab(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	// gitlab.com
	processor, err := factory.Create(&Config{
		Info: &Info{
			RepositoryURL: "https://gitlab.com/owner/repo",
		},
	})

	assert.Nil(err)
	assert.IsType(&GitLabProcessorAdapter{}, processor)
	gitlabProcessor := processor.(*GitLabProcessorAdapter)
	assert.Equal("https://gitlab.com", gitlabProcessor.Host)

	// Selbst-gehostetes GitLab wurde entfernt, da Style-Eigenschaft nicht mehr unterstützt wird
}

func TestProcessorFactoryForBitbucket(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	// bitbucket.org
	processor, err := factory.Create(&Config{
		Info: &Info{
			RepositoryURL: "https://bitbucket.org/owner/repo",
		},
	})

	assert.Nil(err)
	assert.IsType(&BitbucketProcessorAdapter{}, processor)
	bitbucketProcessor := processor.(*BitbucketProcessorAdapter)
	assert.Equal("https://bitbucket.org", bitbucketProcessor.Host)

	// Selbst-gehostetes Bitbucket wurde entfernt, da Style-Eigenschaft nicht mehr unterstützt wird
}
