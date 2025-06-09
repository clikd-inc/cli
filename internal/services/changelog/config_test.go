package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigBasic(t *testing.T) {
	assert := assert.New(t)

	// Einfacher Test der Config-Struktur
	config := &Config{
		Bin:        "git",
		WorkingDir: "/test",
		Template:   "/test/CHANGELOG.tpl.md",
		Info: &Info{
			Title:         "CHANGELOG",
			RepositoryURL: "https://example.com/foo/bar",
		},
		Options: &Options{
			NextTag:          "",
			TagFilterPattern: "v[0-9]*",
		},
	}

	assert.Equal("git", config.Bin)
	assert.Equal("/test", config.WorkingDir)
	assert.Equal("/test/CHANGELOG.tpl.md", config.Template)
	assert.Equal("CHANGELOG", config.Info.Title)
	assert.Equal("https://example.com/foo/bar", config.Info.RepositoryURL)
	assert.Equal("v[0-9]*", config.Options.TagFilterPattern)
}
