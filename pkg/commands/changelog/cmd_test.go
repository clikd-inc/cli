package changelog

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewChangelogCmd(t *testing.T) {
	cmd := NewChangelogCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "changelog [options] <tag query>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestChangelogCmdHelp(t *testing.T) {
	cmd := NewChangelogCmd()

	// Test help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	// Help sollte ohne Fehler angezeigt werden
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "changelog")
}

func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}
