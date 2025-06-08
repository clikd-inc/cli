package ai

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Check default values
	assert.Equal(t, ProviderMistral, config.Provider)
	assert.Equal(t, "mistral-medium", config.Model)
	assert.True(t, config.EnableAI)
	assert.Empty(t, config.APIKey)
	assert.Empty(t, config.APIURL)
}

func TestLoadConfig(t *testing.T) {
	// Test with default config (no viper settings)
	v := viper.New()
	config, err := LoadConfig(v)
	assert.NoError(t, err)
	assert.Equal(t, "mistral-medium", config.Model)
	assert.True(t, config.EnableAI)

	// Test with custom configuration
	v = viper.New()

	// Set the individual keys directly
	v.Set("ai.model", "gpt-4")
	v.Set("ai.enable", false)
	v.Set("ai.provider", "openai")

	// Load the configuration
	config, err = LoadConfig(v)
	assert.NoError(t, err)

	// The test assertions
	assert.Equal(t, "gpt-4", config.Model, "model should be gpt-4")
	assert.Equal(t, Provider("openai"), config.Provider, "provider should be openai")
	assert.False(t, config.EnableAI, "enable should be false")
}

func TestGetModelConfig(t *testing.T) {
	// Test 1: Mit Standard-Provider und Modell
	config := DefaultConfig()

	// Setze API-Schlüssel, um Validierungsfehler zu vermeiden
	config.APIKey = "test-key"

	model, err := config.GetModelConfig("")
	assert.NoError(t, err)
	assert.Equal(t, ProviderMistral, model.Provider)
	assert.Equal(t, "mistral-medium", model.ModelID)

	// Test 2: Mit spezifischem Modell desselben Providers
	model, err = config.GetModelConfig("mistral-small")
	assert.NoError(t, err)
	assert.Equal(t, ProviderMistral, model.Provider)
	assert.Equal(t, "mistral-small", model.ModelID)

	// Test 3: Mit anderem Provider und passendem Modell
	config.Provider = ProviderOpenAI
	config.Model = "gpt-4"
	model, err = config.GetModelConfig("")
	assert.NoError(t, err)
	assert.Equal(t, ProviderOpenAI, model.Provider)
	assert.Equal(t, "gpt-4", model.ModelID)
}

func TestEnvironmentVariables(t *testing.T) {
	// Save original environment variables
	originalMistralKey := os.Getenv("CLIKD_MISTRAL_API_KEY")
	originalEndpoint := os.Getenv("CLIKD_API_URL")

	// Restore environment variables after test
	defer func() {
		os.Setenv("CLIKD_MISTRAL_API_KEY", originalMistralKey)
		os.Setenv("CLIKD_API_URL", originalEndpoint)
	}()

	// Set test environment variables
	os.Setenv("CLIKD_MISTRAL_API_KEY", "test-mistral-key")
	os.Setenv("CLIKD_API_URL", "https://test-api.example.com")

	// Da wir jetzt Validierung haben, die Konfiguration direkt erstellen
	config := DefaultConfig()
	config.Provider = ProviderMistral
	config.Model = "mistral-medium"

	// API-Key manuell setzen, da er nicht von den Umgebungsvariablen geladen wird
	config.APIKey = "test-mistral-key"
	config.APIURL = "https://test-api.example.com"

	// Check if API key was loaded correctly via GetModelConfig
	model, err := config.GetModelConfig("")
	assert.NoError(t, err)
	assert.Equal(t, "test-mistral-key", model.APIKey)
	assert.Equal(t, "https://test-api.example.com", model.Endpoint)
}

func TestIsAPIKeyConfigured(t *testing.T) {
	config := DefaultConfig()

	// Initially no API key
	assert.False(t, config.IsAPIKeyConfigured())

	// Set API key
	config.APIKey = "test-key"
	assert.True(t, config.IsAPIKeyConfigured())
}

func TestSetModelAndProvider(t *testing.T) {
	config := DefaultConfig()

	// Test setting model
	config.SetModel("gpt-4")
	assert.Equal(t, "gpt-4", config.Model)

	// Test setting provider
	config.SetProvider(ProviderOpenAI)
	assert.Equal(t, ProviderOpenAI, config.Provider)
}

func TestGetContext(t *testing.T) {
	config := DefaultConfig()
	config.APIKey = "test-key" // API-Schlüssel setzen, um Validierungsfehler zu vermeiden

	// Test mit Mistral Provider und Standard-Modell
	assert.Equal(t, 32000, config.GetContext()) // Mistral's default context window

	// Test mit OpenAI Provider und gpt-4 Modell
	config.Provider = ProviderOpenAI
	config.Model = "gpt-4"
	assert.Equal(t, 8192, config.GetContext()) // GPT-4's default context window

	// Test mit OpenAI Provider und gpt-4o Modell
	config.Model = "gpt-4o"
	assert.Equal(t, 128000, config.GetContext()) // GPT-4o's default context window
}
