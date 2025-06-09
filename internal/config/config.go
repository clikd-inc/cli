package config

import (
	"fmt"
	"os"
	"sync"
)

var (
	// globalManager is the singleton instance of the configuration manager
	globalManager *Manager

	// mutex for thread-safe initialization
	managerMutex sync.Mutex
)

// Initialize initializes the global configuration
// If configFile is empty, the default path is used
func Initialize(configFile string) error {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	// Create new manager instance
	manager := NewManager()

	// Load configuration
	if err := manager.InitConfig(configFile); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Explicitly check and set environment variables (for tests)
	// These values override anything loaded via the loadSensitiveEnvVars method
	if logLevel := os.Getenv("CLIKD_GENERAL_LOG_LEVEL"); logLevel != "" {
		manager.config.General.LogLevel = logLevel
	}

	if model := os.Getenv("CLIKD_AI_MODEL"); model != "" {
		manager.config.AI.Model = model
	}

	if provider := os.Getenv("CLIKD_AI_PROVIDER"); provider != "" {
		manager.config.AI.Provider = provider
	}

	// Explicitly check API keys for tests
	if apiKey := os.Getenv("CLIKD_API_KEY"); apiKey != "" {
		manager.config.AI.APIKey = apiKey
	} else if openaiKey := os.Getenv("CLIKD_OPENAI_API_KEY"); openaiKey != "" && manager.config.AI.Provider == "openai" {
		manager.config.AI.APIKey = openaiKey
	} else if mistralKey := os.Getenv("CLIKD_MISTRAL_API_KEY"); mistralKey != "" && manager.config.AI.Provider == "mistral" {
		manager.config.AI.APIKey = mistralKey
	} else if anthropicKey := os.Getenv("CLIKD_ANTHROPIC_API_KEY"); anthropicKey != "" && manager.config.AI.Provider == "anthropic" {
		manager.config.AI.APIKey = anthropicKey
	}

	// Set global instance
	globalManager = manager

	return nil
}

// Get returns the current configuration
// If the configuration has not been initialized yet, an error is returned
func Get() (*ConfigData, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	// Convert the new config structure to the old ConfigData structure
	config := globalManager.GetConfig()
	configData := convertToConfigData(config)
	return configData, nil
}

// GetManager returns the global configuration manager
// If the manager has not been initialized yet, an error is returned
func GetManager() (*Manager, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	return globalManager, nil
}

// GetConfigFilePath returns the path to the used configuration file
func GetConfigFilePath() (string, error) {
	if globalManager == nil {
		return "", fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	// In the new implementation, this is simply configPath
	return globalManager.configPath, nil
}

// Set sets a configuration value in the global manager
func Set(key string, value interface{}) error {
	if globalManager == nil {
		return fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	// Convert the value to a string for the new SetConfigValue method
	valueStr := fmt.Sprintf("%v", value)

	// For some commonly used keys, set directly
	if key == "general.log_level" {
		globalManager.config.General.LogLevel = valueStr
		return nil
	} else if key == "ai.model" {
		globalManager.config.AI.Model = valueStr
		return nil
	} else if key == "ai.provider" {
		globalManager.config.AI.Provider = valueStr
		return nil
	} else if key == "ai.enable" {
		globalManager.config.AI.Enable = (valueStr == "true")
		return nil
	}

	// Otherwise use the general SetConfigValue method
	return globalManager.SetConfigValue(key, valueStr)
}

// Save saves the current configuration to the file
func Save(filePath string) error {
	if globalManager == nil {
		return fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	return globalManager.SaveConfig(filePath)
}

// GetAIModelConfig returns the configuration for a specific AI model
func GetAIModelConfig(modelName string) (ModelConfig, error) {
	config, err := Get()
	if err != nil {
		return ModelConfig{}, err
	}

	// Validate the provider-model combination
	provider := config.AI.Provider
	model := config.AI.Model

	// If a specific model was requested, use it for validation
	if modelName != "" && modelName != model {
		// If a model different from the configured one was requested,
		// check if it's compatible with the provider
		if err := ValidateProviderModel(provider, modelName); err != nil {
			return ModelConfig{}, fmt.Errorf("invalid configuration: %w", err)
		}
		// If compatible, use the requested model for the return
		model = modelName
	} else {
		// Otherwise validate the configured combination
		if err := ValidateProviderModel(provider, model); err != nil {
			return ModelConfig{}, fmt.Errorf("invalid configuration: %w", err)
		}
	}

	// Create ModelConfig
	modelConfig := ModelConfig{
		Provider: provider,
		ModelID:  model,
		APIKey:   config.AI.APIKey,
		Endpoint: config.AI.APIURL,
	}

	return modelConfig, nil
}

// EnsureInitialized ensures that the configuration is initialized
// If not, it will be initialized with default values
func EnsureInitialized() (*ConfigData, error) {
	if globalManager == nil {
		if err := Initialize(""); err != nil {
			return nil, err
		}
	}

	// Convert the new Config structure to the old ConfigData structure
	config := globalManager.GetConfig()
	configData := convertToConfigData(config)
	return configData, nil
}

// Reset resets the global configuration (mainly for tests)
func Reset() {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	globalManager = nil
}

// convertToConfigData converts the new Config structure to the old ConfigData structure
func convertToConfigData(config Config) *ConfigData {
	configData := &ConfigData{
		Version: config.Version,
		General: GeneralConfig{
			LogLevel: config.General.LogLevel,
		},
		AI: AIConfig{
			Enable:           config.AI.Enable,
			Provider:         config.AI.Provider,
			Model:            config.AI.Model,
			APIKey:           config.AI.APIKey,
			APIURL:           config.AI.APIURL,
			APICustomHeaders: config.AI.APICustomHeaders,
			TokensMaxInput:   config.AI.TokensMaxInput,
			TokensMaxOutput:  config.AI.TokensMaxOutput,
		},
	}

	return configData
}
