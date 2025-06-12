package changelog

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChangelogCmd(t *testing.T) {
	cmd := NewChangelogCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "changelog", cmd.Use[:9])
	assert.Equal(t, "Generate a changelog from git history", cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	flags := cmd.Flags()

	configFlag := flags.Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "clikd/changelog/config.yml", configFlag.DefValue)

	outputFlag := flags.Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "", outputFlag.DefValue)

	noColorFlag := flags.Lookup("no-color")
	assert.NotNil(t, noColorFlag)
	assert.Equal(t, "false", noColorFlag.DefValue)
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

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "env var not set, use default true",
			envKey:       "TEST_BOOL_NOT_SET",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "env var not set, use default false",
			envKey:       "TEST_BOOL_NOT_SET",
			envValue:     "",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "env var set to true",
			envKey:       "TEST_BOOL_TRUE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "env var set to false",
			envKey:       "TEST_BOOL_FALSE",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env var set to 1",
			envKey:       "TEST_BOOL_ONE",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "env var set to 0",
			envKey:       "TEST_BOOL_ZERO",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env var set to invalid value, use default",
			envKey:       "TEST_BOOL_INVALID",
			envValue:     "invalid",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.envKey)

			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvBool(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "env var not set, use default",
			envKey:       "TEST_STRING_NOT_SET",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "env var set to non-empty value",
			envKey:       "TEST_STRING_SET",
			envValue:     "custom_value",
			defaultValue: "default_value",
			expected:     "custom_value",
		},
		{
			name:         "env var set to empty string, use default",
			envKey:       "TEST_STRING_EMPTY",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "env var set with spaces",
			envKey:       "TEST_STRING_SPACES",
			envValue:     "  value with spaces  ",
			defaultValue: "default_value",
			expected:     "  value with spaces  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.envKey)

			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvString(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunGenerator(t *testing.T) {
	// Save original flag values
	originalConfigFlag := configFlag
	originalOutputFlag := outputFlag
	originalSilentFlag := silentFlag
	originalNoColorFlag := noColorFlag

	// Restore original values after test
	defer func() {
		configFlag = originalConfigFlag
		outputFlag = originalOutputFlag
		silentFlag = originalSilentFlag
		noColorFlag = originalNoColorFlag
	}()

	t.Run("runGenerator with invalid config path", func(t *testing.T) {
		// Set flags to invalid config path
		configFlag = "/nonexistent/path/config.yml"
		outputFlag = ""
		silentFlag = false
		noColorFlag = false

		err := runGenerator("")

		// Should return an error due to invalid config path
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("runGenerator with no-color flag", func(t *testing.T) {
		// Set flags for file output (should not use UI)
		configFlag = "testdata/test_config.yml"
		outputFlag = "/tmp/test_changelog.md"
		silentFlag = false
		noColorFlag = true

		// Create a minimal test config
		err := os.MkdirAll("testdata", 0755)
		require.NoError(t, err)
		defer os.RemoveAll("testdata")

		testConfig := `style: "standard"
template: "CHANGELOG.tpl.md"
info:
  title: "Test Changelog"
  repository_url: "https://github.com/test/repo"
options:
  commits:
    group_by: "Type"
    sort_by: "Scope"
  commit_groups:
    title_order: ["feat", "fix", "docs"]
    title_maps:
      feat: "Features"
      fix: "Bug Fixes"
      docs: "Documentation"
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - "Type"
      - "Scope" 
      - "Subject"
  notes:
    keywords: ["BREAKING CHANGE"]
`
		err = os.WriteFile("testdata/test_config.yml", []byte(testConfig), 0644)
		require.NoError(t, err)

		err = runGenerator("v1.0.0..v2.0.0")

		// Should return an error because we don't have a real git repository
		assert.Error(t, err)
		// The error should be related to git operations, not config loading
		assert.NotContains(t, err.Error(), "no such file or directory")
	})
}
