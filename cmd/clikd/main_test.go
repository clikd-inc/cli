package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// executeCommand ist ein Hilfshelfer zum Ausführen von Kommandos in Tests
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}

func TestRootCmd(t *testing.T) {
	rootCmd := newRootCmd()

	// Überprüfe, dass das Kommando korrekt initialisiert wurde
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "clikd", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
}

func TestRootCmdFlags(t *testing.T) {
	rootCmd := newRootCmd()

	// Überprüfe, dass die globalen Flags korrekt hinzugefügt wurden
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("config"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("log-level"))

	// Überprüfe, dass die Version-Flag hinzugefügt wurde
	assert.NotNil(t, rootCmd.Flags().Lookup("version"))
}

func TestRootCmdHelp(t *testing.T) {
	rootCmd := newRootCmd()

	// Test help output
	output, err := executeCommand(rootCmd, "--help")
	assert.NoError(t, err)

	// Überprüfe, dass die Hilfe-Ausgabe die globalen Flags enthält
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "--log-level")

	// Überprüfe, dass die Hilfe-Ausgabe die verfügbaren Kommandos enthält
	assert.Contains(t, output, "changelog")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "init")
}
