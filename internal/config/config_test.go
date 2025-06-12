package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Get a default config instance
	cfg, err := EnsureInitialized()
	require.NoError(t, err)

	// Test some default values
	if cfg.General.LogLevel != "info" {
		t.Errorf("expected default log level to be 'info', got %s", cfg.General.LogLevel)
	}

	if cfg.AI.Model != "mistral-medium" {
		t.Errorf("expected default AI model to be 'mistral-medium', got %s", cfg.AI.Model)
	}

	// Test provider exists
	if cfg.AI.Provider != "mistral" {
		t.Error("expected default provider to be 'mistral'")
	}
}

func TestInitializeWithDefaults(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Initialize with no config file
	err := Initialize("")
	if err != nil {
		t.Fatalf("Initialize() with empty path failed: %v", err)
	}

	// Get the config
	cfg, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Check that we have default values
	if cfg.General.LogLevel != "info" {
		t.Errorf("expected default log level to be 'info', got %s", cfg.General.LogLevel)
	}
}

func TestInitializeWithCustomFile(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "clikd-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.toml")
	configContent := `
version = "dev"

[general]
log_level = "debug"
color = false

[ai]
enable = true
model = "gpt-4"
provider = "openai"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Initialize with the test config file
	err = Initialize(configPath)
	if err != nil {
		t.Fatalf("Initialize() with custom path failed: %v", err)
	}

	// Get the config
	cfg, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Check that we have the custom values
	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected log level to be 'debug', got %s", cfg.General.LogLevel)
	}

	if cfg.AI.Model != "gpt-4" {
		t.Errorf("expected AI model to be 'gpt-4', got %s", cfg.AI.Model)
	}
}

func TestSetAndGet(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Initialize with defaults
	err := Initialize("")
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Set a value
	err = Set("general.log_level", "debug")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get the config
	cfg, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Check that the value was updated
	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected log level to be 'debug', got %s", cfg.General.LogLevel)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Erstelle ein temporäres Verzeichnis
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Erstelle einen Konfigurationsmanager
	manager := NewManager()

	// Konfigurationsdatei-Pfad
	configPath := filepath.Join(tempDir, "config.toml")

	// Setze einige Werte
	manager.config.General.LogLevel = "debug"
	manager.config.AI.Model = "gpt-4"

	// Speichere die Konfiguration
	err = manager.SaveConfig(configPath)
	require.NoError(t, err)

	// Erstelle einen neuen Manager und lade die Konfiguration
	newManager := NewManager()
	err = newManager.InitConfig(configPath)
	require.NoError(t, err)

	// Überprüfe, dass die Werte korrekt geladen wurden
	require.Equal(t, "debug", newManager.config.General.LogLevel)
	require.Equal(t, "gpt-4", newManager.config.AI.Model)
}

func TestConfig_LoadWithOverride(t *testing.T) {
	// Erstelle ein temporäres Verzeichnis
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Erstelle eine Konfigurationsdatei
	configPath := filepath.Join(tempDir, "config.toml")
	configContent := `
version = "dev"

[general]
log_level = "debug"
color = false

[ai]
enable = true
model = "gpt-4"
provider = "openai"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Setze Umgebungsvariablen, die die Konfiguration überschreiben sollen
	os.Setenv("CLIKD_GENERAL_LOG_LEVEL", "trace")
	defer os.Unsetenv("CLIKD_GENERAL_LOG_LEVEL")

	// Erstelle einen Manager und lade die Konfiguration
	manager := NewManager()
	err = manager.InitConfig(configPath)
	require.NoError(t, err)

	// Überprüfe, dass die Umgebungsvariable die Konfigurationsdatei überschreibt
	require.Equal(t, "trace", manager.config.General.LogLevel)
	require.Equal(t, "gpt-4", manager.config.AI.Model)
}

func TestEnvironmentVariables(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Set environment variables
	os.Setenv("CLIKD_GENERAL_LOG_LEVEL", "trace")
	os.Setenv("CLIKD_AI_MODEL", "gpt-3.5-turbo")
	os.Setenv("CLIKD_AI_PROVIDER", "openai")
	defer func() {
		os.Unsetenv("CLIKD_GENERAL_LOG_LEVEL")
		os.Unsetenv("CLIKD_AI_MODEL")
		os.Unsetenv("CLIKD_AI_PROVIDER")
	}()

	// Initialize with defaults
	err := Initialize("")
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Get the config
	cfg, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Check that the environment variables were applied
	if cfg.General.LogLevel != "trace" {
		t.Errorf("expected log level to be 'trace', got %s", cfg.General.LogLevel)
	}

	if cfg.AI.Model != "gpt-3.5-turbo" {
		t.Errorf("expected AI model to be 'gpt-3.5-turbo', got %s", cfg.AI.Model)
	}
}

func TestSensitiveEnvironmentVariables(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Set environment variable for API key
	os.Setenv("CLIKD_API_KEY", "test-api-key")
	defer func() {
		os.Unsetenv("CLIKD_API_KEY")
	}()

	// Initialize with defaults
	err := Initialize("")
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Get the config
	cfg, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Check that the API key was loaded
	if cfg.AI.APIKey != "test-api-key" {
		t.Errorf("expected API key to be 'test-api-key', got %s", cfg.AI.APIKey)
	}
}

func TestGetAIModelConfig(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Initialize with defaults - standard-Implementierung verwendet jetzt openai/gpt-4o als Standardwerte
	err := Initialize("")
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Konfiguration für den Test anpassen
	err = Set("ai.provider", "mistral")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	err = Set("ai.model", "mistral-medium")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get a specific model config
	modelConfig, err := GetAIModelConfig("mistral-medium")
	if err != nil {
		t.Fatalf("GetAIModelConfig() failed: %v", err)
	}

	// Check the model config
	if modelConfig.Provider != "mistral" {
		t.Errorf("expected provider to be 'mistral', got %s", modelConfig.Provider)
	}

	if modelConfig.ModelID != "mistral-medium" {
		t.Errorf("expected model ID to be 'mistral-medium', got %s", modelConfig.ModelID)
	}

	// Test für inkompatibles Modell
	_, err = GetAIModelConfig("non-existent-model")
	if err == nil {
		t.Error("expected error for non-existent model, got nil")
	}
}

func TestEnsureInitialized(t *testing.T) {
	// Reset the global manager before test
	Reset()

	// Get config using EnsureInitialized
	cfg, err := EnsureInitialized()
	if err != nil {
		t.Fatalf("EnsureInitialized() failed: %v", err)
	}

	// Check that we have default values
	if cfg.General.LogLevel != "info" {
		t.Errorf("expected default log level to be 'info', got %s", cfg.General.LogLevel)
	}

	// Change a value
	err = Set("general.log_level", "debug")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Call EnsureInitialized again - should return the same instance with the updated value
	cfg, err = EnsureInitialized()
	if err != nil {
		t.Fatalf("EnsureInitialized() failed on second call: %v", err)
	}

	// Check that the value is still updated
	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected log level to be 'debug', got %s", cfg.General.LogLevel)
	}
}

func TestGetManager(t *testing.T) {
	// Reset global state
	Reset()

	// Test when not initialized
	manager, err := GetManager()
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "configuration not initialized")

	// Initialize and test again
	err = Initialize("")
	require.NoError(t, err)

	manager, err = GetManager()
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Cleanup
	Reset()
}

func TestGetConfigFilePath(t *testing.T) {
	// Reset global state
	Reset()

	// Test when not initialized
	path, err := GetConfigFilePath()
	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "configuration not initialized")

	// Initialize and test again
	err = Initialize("")
	require.NoError(t, err)

	path, err = GetConfigFilePath()
	assert.NoError(t, err)
	// Path might be empty if no config file was found, which is valid
	// Just test that no error occurred

	// Cleanup
	Reset()
}

func TestSave(t *testing.T) {
	// Reset global state
	Reset()

	// Test when not initialized
	err := Save("test.toml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not initialized")

	// Initialize and test again
	err = Initialize("")
	require.NoError(t, err)

	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_config.toml")

	err = Save(testFile)
	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Cleanup
	Reset()
}

func TestDefaultConfigFunction(t *testing.T) {
	config := DefaultConfig()

	// Test that default config is properly initialized
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.Version)
	assert.Equal(t, "info", config.General.LogLevel)
	assert.Equal(t, "mistral", config.AI.Provider)
	assert.Equal(t, "mistral-medium", config.AI.Model)
	assert.Equal(t, "", config.AI.APIKey)
	assert.Equal(t, "", config.AI.APIURL)
	assert.Equal(t, "", config.AI.APICustomHeaders)
	assert.Equal(t, 4096, config.AI.TokensMaxInput)
	assert.Equal(t, 500, config.AI.TokensMaxOutput)
}

func TestSetConfigValue(t *testing.T) {
	// Reset global state
	Reset()

	// Initialize
	err := Initialize("")
	require.NoError(t, err)

	manager, err := GetManager()
	require.NoError(t, err)

	// Test setting a valid config value
	err = manager.SetConfigValue("ai.provider", "openai")
	assert.NoError(t, err)

	// Verify the value was set
	config := manager.GetConfig()
	assert.Equal(t, "openai", config.AI.Provider)

	// Test setting another value using the global Set function
	err = Set("general.log_level", "debug")
	assert.NoError(t, err)

	configPtr, err := Get()
	require.NoError(t, err)
	assert.Equal(t, "debug", configPtr.General.LogLevel)

	// Cleanup
	Reset()
}

func TestMaskAPIKey(t *testing.T) {
	// Test masking empty API key
	masked := MaskAPIKey("")
	assert.Equal(t, "****", masked)

	// Test masking short API key
	masked = MaskAPIKey("abc")
	assert.Equal(t, "****", masked)

	// Test masking normal API key
	masked = MaskAPIKey("sk-1234567890abcdef")
	assert.Equal(t, "sk-1...cdef", masked)

	// Test masking very long API key
	masked = MaskAPIKey("very-long-api-key-that-should-be-masked-properly")
	assert.Equal(t, "very...erly", masked)
}

func TestGetSupportedModelsForProvider(t *testing.T) {
	// Test supported providers
	models, err := GetSupportedModelsForProvider("openai")
	assert.NoError(t, err)
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "gpt-4")

	models, err = GetSupportedModelsForProvider("mistral")
	assert.NoError(t, err)
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "mistral-medium")

	models, err = GetSupportedModelsForProvider("anthropic")
	assert.NoError(t, err)
	assert.NotEmpty(t, models)
	assert.Contains(t, models, "claude-3-sonnet")

	// Test unsupported provider
	models, err = GetSupportedModelsForProvider("unsupported")
	assert.Error(t, err)
	assert.Nil(t, models)
}

func TestGetDefaultModelForProvider(t *testing.T) {
	// Test supported providers
	model, err := GetDefaultModelForProvider("openai")
	assert.NoError(t, err)
	assert.NotEmpty(t, model)
	assert.Equal(t, "gpt-4o", model)

	model, err = GetDefaultModelForProvider("mistral")
	assert.NoError(t, err)
	assert.NotEmpty(t, model)
	assert.Equal(t, "mistral-medium", model)

	model, err = GetDefaultModelForProvider("anthropic")
	assert.NoError(t, err)
	assert.NotEmpty(t, model)
	assert.Equal(t, "claude-3-sonnet", model)

	// Test unsupported provider
	model, err = GetDefaultModelForProvider("unsupported")
	assert.Error(t, err)
	assert.Empty(t, model)
}
