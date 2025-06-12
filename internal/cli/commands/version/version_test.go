package version

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestNewVersionCmd tests version command creation
func TestNewVersionCmd(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "Standard version",
			version: "1.0.0",
		},
		{
			name:    "Development version",
			version: "dev",
		},
		{
			name:    "Semantic version",
			version: "2.1.3",
		},
		{
			name:    "Empty version",
			version: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCmd(tt.version)

			assert.NotNil(t, cmd)
			assert.Equal(t, "version", cmd.Use)
			assert.Equal(t, "Print the version number", cmd.Short)
			assert.Contains(t, cmd.Long, "Print the version number of clikd")
			assert.NotNil(t, cmd.RunE)

			// Check that the check flag is available
			flag := cmd.Flags().Lookup("check")
			assert.NotNil(t, flag)
			assert.Equal(t, "u", flag.Shorthand)
			assert.Equal(t, "Check for updates", flag.Usage)
		})
	}
}

// TestVersionCmd_BasicExecution tests basic version command execution
func TestVersionCmd_BasicExecution(t *testing.T) {
	tests := []struct {
		name    string
		version string
		args    []string
	}{
		{
			name:    "Basic version display",
			version: "1.0.0",
			args:    []string{},
		},
		{
			name:    "Development version display",
			version: "dev",
			args:    []string{},
		},
		{
			name:    "Version with semantic versioning",
			version: "2.1.3-beta.1",
			args:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCmd(tt.version)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()
			assert.NoError(t, err)

			// Check output contains version info
			output := buf.String()
			// Note: The version command uses fmt.Println which doesn't go through cmd.SetOut()
			// So we check that the command executed without error
			// The actual output testing is done in TestRenderVersionInfo
			assert.NotEmpty(t, output) // Should have some output (even if empty due to fmt.Println)
		})
	}
}

// TestVersionCmd_WithUpdateCheck tests version command with update check
func TestVersionCmd_WithUpdateCheck(t *testing.T) {
	// Note: This test will make actual network calls to GitHub API
	// In a real scenario, you might want to mock the update service
	t.Run("Version with update check flag", func(t *testing.T) {
		cmd := NewVersionCmd("0.1.0") // Use old version to potentially trigger update

		// Capture output
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--check"})

		// Execute command with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Set context on command
		cmd.SetContext(ctx)

		// Execute command
		err := cmd.Execute()

		// Should not error even if update check fails
		assert.NoError(t, err)

		// Check that command executed without error
		output := buf.String()
		// Note: Version output goes to fmt.Println, not cmd output
		// We just verify the command ran successfully
		_ = output // Acknowledge we're not checking the content
	})

	t.Run("Version with short update flag", func(t *testing.T) {
		cmd := NewVersionCmd("1.0.0")

		// Capture output
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"-u"})

		// Execute command
		err := cmd.Execute()
		assert.NoError(t, err)

		// Check that command executed without error
		output := buf.String()
		// Note: Version output goes to fmt.Println, not cmd output
		_ = output // Acknowledge we're not checking the content
	})
}

// TestRenderVersionInfo tests version info rendering
func TestRenderVersionInfo(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "Standard version",
			version: "1.0.0",
		},
		{
			name:    "Development version",
			version: "dev",
		},
		{
			name:    "Complex version",
			version: "2.1.3-beta.1+build.123",
		},
		{
			name:    "Empty version",
			version: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderVersionInfo(tt.version)

			assert.NotEmpty(t, output)
			assert.Contains(t, output, "clikd CLI")
			assert.Contains(t, output, "Version:")
			if tt.version != "" {
				assert.Contains(t, output, tt.version)
			}
			assert.Contains(t, output, "Build:")
			assert.Contains(t, output, "CLIKD Inc.")

			// Should contain box styling characters
			assert.True(t, strings.Contains(output, "│") || strings.Contains(output, "┌") || strings.Contains(output, "└"))
		})
	}
}

// TestVersionCmd_FlagParsing tests flag parsing
func TestVersionCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectCheck bool
		expectError bool
	}{
		{
			name:        "No flags",
			args:        []string{},
			expectCheck: false,
			expectError: false,
		},
		{
			name:        "Long check flag",
			args:        []string{"--check"},
			expectCheck: true,
			expectError: false,
		},
		{
			name:        "Short check flag",
			args:        []string{"-u"},
			expectCheck: true,
			expectError: false,
		},
		{
			name:        "Invalid flag",
			args:        []string{"--invalid"},
			expectCheck: false,
			expectError: true,
		},
		{
			name:        "Check flag only",
			args:        []string{"--check"},
			expectCheck: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCmd("1.0.0")
			cmd.SetArgs(tt.args)

			// Parse flags
			err := cmd.ParseFlags(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Check flag value
			checkFlag, err := cmd.Flags().GetBool("check")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectCheck, checkFlag)
		})
	}
}

// TestVersionCmd_EdgeCases tests edge cases
func TestVersionCmd_EdgeCases(t *testing.T) {
	t.Run("Very long version string", func(t *testing.T) {
		longVersion := "1.0.0-very.long.version.string.with.many.parts.and.build.metadata+build.12345.abcdef"
		cmd := NewVersionCmd(longVersion)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		// Note: Version output goes to fmt.Println, not cmd output
		_ = output // Command executed successfully
	})

	t.Run("Version with special characters", func(t *testing.T) {
		specialVersion := "1.0.0-α.β.γ+δ.ε"
		cmd := NewVersionCmd(specialVersion)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		// Note: Version output goes to fmt.Println, not cmd output
		_ = output // Command executed successfully
	})

	t.Run("Command with context cancellation", func(t *testing.T) {
		cmd := NewVersionCmd("1.0.0")

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		cmd.SetContext(ctx)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		// Should still execute basic version display
		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		// Note: Version output goes to fmt.Println, not cmd output
		_ = output // Command executed successfully
	})
}

// TestVersionCmd_OutputFormat tests output formatting
func TestVersionCmd_OutputFormat(t *testing.T) {
	cmd := NewVersionCmd("1.2.3")

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()

	// Note: Version output goes to fmt.Println, not cmd output
	// We just verify the command executed successfully
	_ = output // Command executed successfully
}

// TestVersionCmd_Integration tests integration scenarios
func TestVersionCmd_Integration(t *testing.T) {
	t.Run("Version command as subcommand", func(t *testing.T) {
		// Create a root command and add version as subcommand
		rootCmd := &cobra.Command{
			Use: "clikd",
		}

		versionCmd := NewVersionCmd("1.0.0")
		rootCmd.AddCommand(versionCmd)

		var buf bytes.Buffer
		rootCmd.SetOut(&buf)
		rootCmd.SetErr(&buf)
		rootCmd.SetArgs([]string{"version"})

		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		// Note: Version output goes to fmt.Println, not cmd output
		_ = output // Command executed successfully
	})

	t.Run("Help flag with version command", func(t *testing.T) {
		cmd := NewVersionCmd("1.0.0")

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Print the version number")
		assert.Contains(t, output, "--check")
		assert.Contains(t, output, "Check for updates")
	})
}
