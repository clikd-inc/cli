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
	assert.Equal(t, ProviderMistral, config.DefaultProvider)
	assert.Equal(t, "mistral-medium", config.DefaultModel)
	assert.True(t, config.EnableAI)
	assert.False(t, config.Verbose)

	// Check default models
	assert.Contains(t, config.Models, "mistral-medium")
	assert.Contains(t, config.Models, "mistral-small")
	assert.Contains(t, config.Models, "gpt-3.5-turbo")
	assert.Contains(t, config.Models, "gpt-4")

	// Check model configuration
	mistralModel := config.Models["mistral-medium"]
	assert.Equal(t, ProviderMistral, mistralModel.Provider)
	assert.Equal(t, "mistral-medium", mistralModel.ModelID)
	assert.Equal(t, 1024, mistralModel.MaxTokens)
	assert.Equal(t, 0.7, mistralModel.Temperature)
	assert.Equal(t, 0.9, mistralModel.TopP)
}

func TestLoadConfig(t *testing.T) {
	// Test with default config (no viper settings)
	v := viper.New()
	config, err := LoadConfig(v)
	assert.NoError(t, err)
	assert.Equal(t, "mistral-medium", config.DefaultModel)
	assert.True(t, config.EnableAI)

	// Test with custom configuration
	// Using direct key setting instead of nested map
	v = viper.New()

	// Set the individual keys directly instead of using a map
	v.Set("ai.default_model", "gpt-4")
	v.Set("ai.enable_ai", false)

	// Load the configuration
	config, err = LoadConfig(v)
	assert.NoError(t, err)

	// The test assertions
	assert.Equal(t, "gpt-4", config.DefaultModel, "default_model should be gpt-4")
	assert.False(t, config.EnableAI, "enable_ai should be false")
}

func TestGetModelConfig(t *testing.T) {
	config := DefaultConfig()

	// Test getting an existing model
	model, err := config.GetModelConfig("mistral-medium")
	assert.NoError(t, err)
	assert.Equal(t, ProviderMistral, model.Provider)
	assert.Equal(t, "mistral-medium", model.ModelID)

	// Test getting a non-existent model (should fall back to default)
	model, err = config.GetModelConfig("non-existent")
	assert.NoError(t, err)
	assert.Equal(t, "mistral-medium", model.ModelID)

	// Test with custom default model
	config.DefaultModel = "gpt-4"
	model, err = config.GetModelConfig("non-existent")
	assert.NoError(t, err)
	assert.Equal(t, "gpt-4", model.ModelID)

	// Test with non-existent default model
	config.DefaultModel = "non-existent-default"
	_, err = config.GetModelConfig("non-existent")
	assert.Error(t, err)
}

func TestEnvironmentVariables(t *testing.T) {
	// Save original environment variables
	originalMistralKey := os.Getenv("MISTRAL_API_KEY")
	originalOpenAIKey := os.Getenv("OPENAI_API_KEY")

	// Restore environment variables after test
	defer func() {
		os.Setenv("MISTRAL_API_KEY", originalMistralKey)
		os.Setenv("OPENAI_API_KEY", originalOpenAIKey)
	}()

	// Set test environment variables
	os.Setenv("MISTRAL_API_KEY", "test-mistral-key")
	os.Setenv("OPENAI_API_KEY", "test-openai-key")

	// Test loading from environment
	v := viper.New()
	config, err := LoadConfig(v)
	assert.NoError(t, err)

	// Check if API keys were loaded correctly
	mistralModel := config.Models["mistral-medium"]
	assert.Equal(t, "test-mistral-key", mistralModel.APIKey)

	openaiModel := config.Models["gpt-4"]
	assert.Equal(t, "test-openai-key", openaiModel.APIKey)
}

func TestAddModel(t *testing.T) {
	config := DefaultConfig()

	// Add a new model
	config.AddModel("custom-model", ModelConfig{
		Provider:    ProviderLocal,
		ModelID:     "custom",
		Endpoint:    "http://localhost:8000",
		MaxTokens:   512,
		Temperature: 0.5,
	})

	// Check if model was added
	assert.Contains(t, config.Models, "custom-model")

	// Check model properties
	model := config.Models["custom-model"]
	assert.Equal(t, ProviderLocal, model.Provider)
	assert.Equal(t, "custom", model.ModelID)
	assert.Equal(t, "http://localhost:8000", model.Endpoint)
	assert.Equal(t, 512, model.MaxTokens)
	assert.Equal(t, 0.5, model.Temperature)
}

func TestSetDefaultModel(t *testing.T) {
	config := DefaultConfig()

	// Test setting an existing model as default
	err := config.SetDefaultModel("gpt-4")
	assert.NoError(t, err)
	assert.Equal(t, "gpt-4", config.DefaultModel)
	assert.Equal(t, ProviderOpenAI, config.DefaultProvider)

	// Test setting a non-existent model
	err = config.SetDefaultModel("non-existent")
	assert.Error(t, err)
}

func TestGetAvailableModels(t *testing.T) {
	config := DefaultConfig()

	// Add a custom model
	config.AddModel("custom-model", ModelConfig{
		Provider: ProviderLocal,
		ModelID:  "custom",
	})

	// Get available models
	models := config.GetAvailableModels()

	// Check if all models are included
	assert.Contains(t, models, "mistral-medium")
	assert.Contains(t, models, "gpt-4")
	assert.Contains(t, models, "custom-model")
	assert.Len(t, models, len(config.Models))
}

func TestIsAPIKeyConfigured(t *testing.T) {
	config := DefaultConfig()

	// Set API key for a model
	mistralModel := config.Models["mistral-medium"]
	mistralModel.APIKey = "test-key"
	config.Models["mistral-medium"] = mistralModel

	// Test with API key configured
	assert.True(t, config.IsAPIKeyConfigured("mistral-medium"))

	// Test with no API key
	assert.False(t, config.IsAPIKeyConfigured("gpt-4"))

	// Test with non-existent model (falls back to default)
	assert.True(t, config.IsAPIKeyConfigured("non-existent"))

	// Test with different default (no API key)
	config.DefaultModel = "gpt-4"
	assert.False(t, config.IsAPIKeyConfigured("non-existent"))
}
