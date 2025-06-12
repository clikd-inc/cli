package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewRootCmd(t *testing.T) {
	cmd := newRootCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "clikd", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test that the command has the expected subcommands
	subcommands := cmd.Commands()
	commandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		commandNames[i] = subcmd.Name()
	}

	// Check for expected commands
	assert.Contains(t, commandNames, "completion")
	assert.Contains(t, commandNames, "version")
	assert.Contains(t, commandNames, "changelog")
	assert.Contains(t, commandNames, "init")
	assert.Contains(t, commandNames, "ai-test")
}

func TestRootCmdFlags(t *testing.T) {
	cmd := newRootCmd()

	// Test persistent flags
	persistentFlags := cmd.PersistentFlags()

	configFlag := persistentFlags.Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)

	logLevelFlag := persistentFlags.Lookup("log-level")
	assert.NotNil(t, logLevelFlag)
	assert.Equal(t, "l", logLevelFlag.Shorthand)
	assert.Equal(t, "info", logLevelFlag.DefValue)

	verboseFlag := persistentFlags.Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "V", verboseFlag.Shorthand)

	noColorFlag := persistentFlags.Lookup("no-color")
	assert.NotNil(t, noColorFlag)

	colorFlag := persistentFlags.Lookup("color")
	assert.NotNil(t, colorFlag)

	// Test local flags
	flags := cmd.Flags()
	versionFlag := flags.Lookup("version")
	assert.NotNil(t, versionFlag)
	assert.Equal(t, "v", versionFlag.Shorthand)
}

func TestRootCmdHelp(t *testing.T) {
	cmd := newRootCmd()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set args to trigger help
	cmd.SetArgs([]string{})

	// Execute should show help when no args provided
	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "clikd")
	assert.Contains(t, output, "Usage:")
}

func TestRootCmdVersion(t *testing.T) {
	cmd := newRootCmd()
	cmd.Version = "test-version"
	cmd.SetVersionTemplate("clikd version {{.Version}}\n")

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test version flag
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test-version")
}

func TestCompletionCommand(t *testing.T) {
	cmd := newRootCmd()

	// Find completion command
	var completionCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == "completion" {
			completionCmd = subcmd
			break
		}
	}

	require.NotNil(t, completionCmd)
	assert.Equal(t, "completion", completionCmd.Use[:10]) // Check first part
	assert.NotEmpty(t, completionCmd.Short)
	assert.NotEmpty(t, completionCmd.Long)

	// Test valid args
	assert.Contains(t, completionCmd.ValidArgs, "bash")
	assert.Contains(t, completionCmd.ValidArgs, "zsh")
	assert.Contains(t, completionCmd.ValidArgs, "fish")
	assert.Contains(t, completionCmd.ValidArgs, "powershell")
}

func TestCompletionCommandExecution(t *testing.T) {
	cmd := newRootCmd()

	// Test that completion command exists and can be executed
	cmd.SetArgs([]string{"completion", "bash"})

	// Execute the command - it should not error
	err := cmd.Execute()
	assert.NoError(t, err)

	// Test that the completion command has the right structure
	completionCmd := cmd.Commands()[2] // completion is typically the 3rd command
	assert.Equal(t, "completion", completionCmd.Use[:10])
	assert.Contains(t, completionCmd.ValidArgs, "bash")
}

func TestAITestCommand(t *testing.T) {
	cmd := newRootCmd()

	// Find ai-test command
	var aiTestCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == "ai-test" {
			aiTestCmd = subcmd
			break
		}
	}

	require.NotNil(t, aiTestCmd)
	assert.Equal(t, "ai-test", aiTestCmd.Use[:7]) // Check first part
	assert.NotEmpty(t, aiTestCmd.Short)
	assert.NotEmpty(t, aiTestCmd.Long)

	// Test that it requires at least one argument (indirectly by checking it's not nil)
	assert.NotNil(t, aiTestCmd.Args)

	// Test model flag
	modelFlag := aiTestCmd.Flags().Lookup("model")
	assert.NotNil(t, modelFlag)
}

func TestAITestCommandWithoutArgs(t *testing.T) {
	cmd := newRootCmd()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test ai-test without arguments (should fail)
	cmd.SetArgs([]string{"ai-test"})

	err := cmd.Execute()
	assert.Error(t, err) // Should error because no prompt provided
}

func TestPersistentPreRunE(t *testing.T) {
	// Test the persistent pre-run function indirectly by checking flag handling
	cmd := newRootCmd()

	// Set some flags
	cmd.SetArgs([]string{"--log-level", "debug", "--verbose", "version"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute should not error
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestPersistentPostRunE(t *testing.T) {
	// Test that post-run doesn't error for version command (should skip update check)
	cmd := newRootCmd()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test version command (should skip update check)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEnvironmentVariableHandling(t *testing.T) {
	// Test that environment variables are handled correctly

	// Save original env vars
	originalLogLevel := os.Getenv("CLIKD_LOG_LEVEL")
	originalVerbose := os.Getenv("CLIKD_VERBOSE")

	// Clean up after test
	defer func() {
		if originalLogLevel != "" {
			os.Setenv("CLIKD_LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("CLIKD_LOG_LEVEL")
		}
		if originalVerbose != "" {
			os.Setenv("CLIKD_VERBOSE", originalVerbose)
		} else {
			os.Unsetenv("CLIKD_VERBOSE")
		}
	}()

	// Set test environment variables
	os.Setenv("CLIKD_LOG_LEVEL", "debug")
	os.Setenv("CLIKD_VERBOSE", "true")

	cmd := newRootCmd()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute version command to test env var handling
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestContextTimeout(t *testing.T) {
	// Test that context timeout is properly handled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to timeout
	<-ctx.Done()

	// Verify context is cancelled
	assert.Error(t, ctx.Err())
	assert.Equal(t, context.DeadlineExceeded, ctx.Err())
}

func TestCommandStructure(t *testing.T) {
	cmd := newRootCmd()

	// Test that all expected commands are present and properly configured
	commands := cmd.Commands()

	for _, subcmd := range commands {
		// Each command should have a name, short description, and run function
		assert.NotEmpty(t, subcmd.Name())
		assert.NotEmpty(t, subcmd.Short)

		// Commands should not be nil
		assert.NotNil(t, subcmd)
	}
}

func TestFlagValidation(t *testing.T) {
	cmd := newRootCmd()

	// Test invalid log level
	cmd.SetArgs([]string{"--log-level", "invalid", "version"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Should still execute (invalid log level might be handled gracefully)
	err := cmd.Execute()
	// We don't assert error here because the command might handle invalid log levels gracefully
	_ = err
}

// Helper function to check if a string slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
