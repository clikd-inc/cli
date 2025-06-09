package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessorFactory(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	processor, err := factory.CreateProcessorFromString("")
	assert.Nil(err)
	assert.Nil(processor)
}

func TestProcessorFactoryForGitHub(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	// github.com
	processor, err := factory.CreateProcessorFromString("github")

	assert.Nil(err)
	assert.IsType(&GitHubProcessor{}, processor)

	// Bootstrap aufrufen, um Host zu setzen
	config := &Config{
		Info: &Info{
			RepositoryURL: "https://github.com/owner/repo",
		},
	}
	processor.Bootstrap(config)

	githubProcessor := processor.(*GitHubProcessor)
	assert.Equal("https://github.com", githubProcessor.Host)

	// Mit angegebenem Host
	processor, err = factory.CreateProcessorFromString("github:https://enterprise.github.com")
	assert.Nil(err)
	assert.IsType(&GitHubProcessor{}, processor)
	processor.Bootstrap(config)
	githubProcessor = processor.(*GitHubProcessor)
	assert.Equal("https://enterprise.github.com", githubProcessor.Host)
}

func TestProcessorFactoryForGitLab(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	// gitlab.com
	processor, err := factory.CreateProcessorFromString("gitlab")

	assert.Nil(err)
	assert.IsType(&GitLabProcessor{}, processor)

	// Bootstrap aufrufen, um Host zu setzen
	config := &Config{
		Info: &Info{
			RepositoryURL: "https://gitlab.com/owner/repo",
		},
	}
	processor.Bootstrap(config)

	gitlabProcessor := processor.(*GitLabProcessor)
	assert.Equal("https://gitlab.com", gitlabProcessor.Host)

	// Mit angegebenem Host
	processor, err = factory.CreateProcessorFromString("gitlab:https://gitlab.example.com")
	assert.Nil(err)
	assert.IsType(&GitLabProcessor{}, processor)
	processor.Bootstrap(config)
	gitlabProcessor = processor.(*GitLabProcessor)
	assert.Equal("https://gitlab.example.com", gitlabProcessor.Host)
}

func TestProcessorFactoryForBitbucket(t *testing.T) {
	assert := assert.New(t)
	factory := NewProcessorFactory()

	// bitbucket.org
	processor, err := factory.CreateProcessorFromString("bitbucket")

	assert.Nil(err)
	assert.IsType(&BitbucketProcessor{}, processor)

	// Bootstrap aufrufen, um Host zu setzen
	config := &Config{
		Info: &Info{
			RepositoryURL: "https://bitbucket.org/owner/repo",
		},
	}
	processor.Bootstrap(config)

	bitbucketProcessor := processor.(*BitbucketProcessor)
	assert.Equal("https://bitbucket.org", bitbucketProcessor.Host)

	// Mit angegebenem Host
	processor, err = factory.CreateProcessorFromString("bitbucket:https://bitbucket.example.com")
	assert.Nil(err)
	assert.IsType(&BitbucketProcessor{}, processor)
	processor.Bootstrap(config)
	bitbucketProcessor = processor.(*BitbucketProcessor)
	assert.Equal("https://bitbucket.example.com", bitbucketProcessor.Host)
}
