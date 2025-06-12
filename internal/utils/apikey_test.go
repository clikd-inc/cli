package utils

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAPIKey(t *testing.T) {
	// Save original environment and viper state
	originalAPIKey := os.Getenv("CLIKD_API_KEY")
	originalAIEnabled := os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED")
	originalViperAPIKey := viper.GetString("ai.api_key")

	// Clean up after test
	defer func() {
		os.Setenv("CLIKD_API_KEY", originalAPIKey)
		os.Setenv("CLIKD_AI_EXPLICITLY_ENABLED", originalAIEnabled)
		viper.Set("ai.api_key", originalViperAPIKey)
	}()

	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	t.Run("API key from .env file (local config present)", func(t *testing.T) {
		// Change to temp directory
		os.Chdir(tempDir)

		// Create clikd directory to simulate local config
		err := os.Mkdir("clikd", 0755)
		require.NoError(t, err)

		// Create .env file with API key
		envContent := `# Test .env file
CLIKD_API_KEY=test_api_key_from_env
OTHER_VAR=other_value
`
		err = os.WriteFile(".env", []byte(envContent), 0644)
		require.NoError(t, err)

		// Clear environment and viper
		os.Unsetenv("CLIKD_API_KEY")
		viper.Set("ai.api_key", "")

		key, err := GetAPIKey()
		assert.NoError(t, err)
		assert.Equal(t, "test_api_key_from_env", key)

		// Clean up
		os.RemoveAll("clikd")
		os.Remove(".env")
	})

	t.Run("API key from environment variable", func(t *testing.T) {
		// Change to temp directory (no local config)
		os.Chdir(tempDir)

		// Set environment variable
		os.Setenv("CLIKD_API_KEY", "test_api_key_from_env_var")
		viper.Set("ai.api_key", "")

		key, err := GetAPIKey()
		assert.NoError(t, err)
		assert.Equal(t, "test_api_key_from_env_var", key)
	})

	t.Run("API key from viper config", func(t *testing.T) {
		// Change to temp directory (no local config)
		os.Chdir(tempDir)

		// Clear environment
		os.Unsetenv("CLIKD_API_KEY")
		viper.Set("ai.api_key", "test_api_key_from_viper")

		key, err := GetAPIKey()
		assert.NoError(t, err)
		assert.Equal(t, "test_api_key_from_viper", key)
	})

	t.Run("No API key found - basic error", func(t *testing.T) {
		// Change to temp directory (no local config)
		os.Chdir(tempDir)

		// Clear all sources
		os.Unsetenv("CLIKD_API_KEY")
		os.Unsetenv("CLIKD_AI_EXPLICITLY_ENABLED")
		viper.Set("ai.api_key", "")

		key, err := GetAPIKey()
		assert.Error(t, err)
		assert.Empty(t, key)
		assert.Contains(t, err.Error(), "CLIKD_API_KEY not found")
		assert.Contains(t, err.Error(), "clikd config set ai.api_key")
	})

	t.Run("No API key found - with AI explicitly enabled", func(t *testing.T) {
		// Change to temp directory (no local config)
		os.Chdir(tempDir)

		// Clear all sources but set AI explicitly enabled
		os.Unsetenv("CLIKD_API_KEY")
		os.Setenv("CLIKD_AI_EXPLICITLY_ENABLED", "true")
		viper.Set("ai.api_key", "")

		key, err := GetAPIKey()
		assert.Error(t, err)
		assert.Empty(t, key)
		assert.Contains(t, err.Error(), "AI features were explicitly enabled")
		assert.Contains(t, err.Error(), "but no API key found")
	})

	t.Run("No API key found - with local config present", func(t *testing.T) {
		// Change to temp directory
		os.Chdir(tempDir)

		// Create clikd directory to simulate local config
		err := os.Mkdir("clikd", 0755)
		require.NoError(t, err)

		// Clear all sources
		os.Unsetenv("CLIKD_API_KEY")
		os.Unsetenv("CLIKD_AI_EXPLICITLY_ENABLED")
		viper.Set("ai.api_key", "")

		key, err := GetAPIKey()
		assert.Error(t, err)
		assert.Empty(t, key)
		assert.Contains(t, err.Error(), "Create a .env file in the project directory")

		// Clean up
		os.RemoveAll("clikd")
	})
}

func TestIsLocalConfigPresent(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	t.Run("Local config present", func(t *testing.T) {
		// Change to temp directory
		os.Chdir(tempDir)

		// Create clikd directory
		err := os.Mkdir("clikd", 0755)
		require.NoError(t, err)

		result := IsLocalConfigPresent()
		assert.True(t, result)

		// Clean up
		os.RemoveAll("clikd")
	})

	t.Run("Local config not present", func(t *testing.T) {
		// Change to temp directory (no clikd directory)
		os.Chdir(tempDir)

		result := IsLocalConfigPresent()
		assert.False(t, result)
	})
}

func TestGetEnvKeyFromDotEnv(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	t.Run("Read API key from .env file", func(t *testing.T) {
		// Change to temp directory
		os.Chdir(tempDir)

		// Create .env file with various formats
		envContent := `# This is a comment
CLIKD_API_KEY=test_api_key
OTHER_VAR="quoted_value"
export EXPORT_VAR=exported_value
QUOTED_VAR='single_quoted'

# Another comment
EMPTY_VAR=
INVALID_LINE_NO_EQUALS
`
		err := os.WriteFile(".env", []byte(envContent), 0644)
		require.NoError(t, err)

		// Test reading CLIKD_API_KEY
		result := getEnvKeyFromDotEnv("CLIKD_API_KEY")
		assert.Equal(t, "test_api_key", result)

		// Test reading quoted value
		result = getEnvKeyFromDotEnv("OTHER_VAR")
		assert.Equal(t, "quoted_value", result)

		// Test reading exported value
		result = getEnvKeyFromDotEnv("EXPORT_VAR")
		assert.Equal(t, "exported_value", result)

		// Test reading single quoted value
		result = getEnvKeyFromDotEnv("QUOTED_VAR")
		assert.Equal(t, "single_quoted", result)

		// Test reading empty value
		result = getEnvKeyFromDotEnv("EMPTY_VAR")
		assert.Equal(t, "", result)

		// Test reading non-existent key
		result = getEnvKeyFromDotEnv("NON_EXISTENT")
		assert.Equal(t, "", result)

		// Clean up
		os.Remove(".env")
	})

	t.Run("No .env file", func(t *testing.T) {
		// Change to temp directory (no .env file)
		os.Chdir(tempDir)

		result := getEnvKeyFromDotEnv("CLIKD_API_KEY")
		assert.Equal(t, "", result)
	})

	t.Run("Complex .env file with edge cases", func(t *testing.T) {
		// Change to temp directory
		os.Chdir(tempDir)

		// Create .env file with edge cases
		envContent := `
# Comment line
   # Indented comment

KEY_WITH_SPACES = value_with_spaces   
export EXPORT_KEY_SPACES = "exported with spaces"
KEY_WITH_EQUALS=value=with=equals
MULTILINE_START=value
`
		err := os.WriteFile(".env", []byte(envContent), 0644)
		require.NoError(t, err)

		// Test key with spaces
		result := getEnvKeyFromDotEnv("KEY_WITH_SPACES")
		assert.Equal(t, "value_with_spaces", result)

		// Test exported key with spaces
		result = getEnvKeyFromDotEnv("EXPORT_KEY_SPACES")
		assert.Equal(t, "exported with spaces", result)

		// Test key with equals in value
		result = getEnvKeyFromDotEnv("KEY_WITH_EQUALS")
		assert.Equal(t, "value=with=equals", result)

		// Clean up
		os.Remove(".env")
	})
}
