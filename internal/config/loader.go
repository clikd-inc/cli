package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"clikd/internal/cli/version"
	"clikd/internal/utils"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

const (
	// EnvPrefix is the prefix for environment variables
	EnvPrefix = "CLIKD"

	// DefaultConfigFileName is the standard name for the configuration file
	DefaultConfigFileName = "config"

	// DefaultConfigFileExt is the standard extension for the configuration file
	DefaultConfigFileExt = "toml"
)

// Config represents the application configuration
type Config struct {
	Version string `toml:"version"`
	General struct {
		LogLevel string `toml:"log_level"`
	} `toml:"general"`
	AI struct {
		// All further AI settings are controlled via environment variables.
		// CLIKD_API_KEY         - API key for the chosen provider
		// CLIKD_AI_PROVIDER     - AI provider ('mistral', 'openai', 'anthropic', etc.)
		// CLIKD_MODEL           - Model (e.g. 'mistral-medium', 'gpt-4o')
		// CLIKD_API_URL         - URL for proxy or alternative API endpoints
		// CLIKD_API_CUSTOM_HEADERS - Custom HTTP headers for API requests
		// CLIKD_TOKENS_MAX_INPUT  - Maximum token limit for inputs (default: 4096)
		// CLIKD_TOKENS_MAX_OUTPUT - Maximum token limit for outputs (default: 500)
		Provider         string `toml:"provider"`
		Model            string `toml:"model"`
		APIKey           string `toml:"api_key"`
		APIURL           string `toml:"api_url"`
		APICustomHeaders string `toml:"api_custom_headers"`
		TokensMaxInput   int    `toml:"tokens_max_input"`
		TokensMaxOutput  int    `toml:"tokens_max_output"`
	} `toml:"ai"`
}

// Manager manages the configuration
type Manager struct {
	config     Config
	configPath string
}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{
		config: createDefaultConfig(),
	}
}

// createDefaultConfig creates a default configuration
func createDefaultConfig() Config {
	c := Config{
		Version: version.GetVersion(),
	}

	// General
	c.General.LogLevel = "info"

	// AI
	c.AI.Provider = "mistral"
	c.AI.Model = "mistral-medium"
	c.AI.APIKey = ""
	c.AI.APIURL = ""
	c.AI.APICustomHeaders = ""
	c.AI.TokensMaxInput = 4096
	c.AI.TokensMaxOutput = 500

	return c
}

// InitConfig initializes the configuration from a file
func (m *Manager) InitConfig(configPath string) error {
	// If an explicit configuration path is specified, we use only this file
	if configPath != "" {
		m.configPath = configPath

		// Read file
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				// If the file doesn't exist, use the default configuration
				m.config = createDefaultConfig()
				return nil
			}
			return fmt.Errorf("error reading config file: %w", err)
		}

		// Parse TOML
		if err := toml.Unmarshal(data, &m.config); err != nil {
			return fmt.Errorf("error parsing config file: %w", err)
		}

		// Load sensitive data from environment variables
		m.loadSensitiveEnvVars()

		// Load repository-specific .env file if available
		m.loadEnvFile()

		return nil
	}

	// If no explicit path is specified, we use the priority order
	// 1. Load default configuration
	m.config = createDefaultConfig()

	// 2. Load global configuration if available
	homedir, err := os.UserHomeDir()
	if err == nil {
		globalConfigPath := filepath.Join(homedir, ".clikd", "config.toml")
		if _, err := os.Stat(globalConfigPath); err == nil {
			// Load global configuration
			data, err := os.ReadFile(globalConfigPath)
			if err == nil {
				// Parse TOML
				if err := toml.Unmarshal(data, &m.config); err == nil {
					// Global configuration loaded
					m.configPath = globalConfigPath
				}
			}
		}
	}

	// 3. Load project-specific configuration if available
	// The current project-specific configuration overrides the global one
	wd, err := os.Getwd()
	if err == nil {
		localConfigPath := filepath.Join(wd, "clikd", "config.toml")
		if _, err := os.Stat(localConfigPath); err == nil {
			// Load local configuration
			data, err := os.ReadFile(localConfigPath)
			if err == nil {
				// Create temporary configuration to avoid overwriting unspecified values
				tempConfig := Config{}
				if err := toml.Unmarshal(data, &tempConfig); err == nil {
					// Only adopt values specified in the local configuration
					// General settings
					if tempConfig.General.LogLevel != "" {
						m.config.General.LogLevel = tempConfig.General.LogLevel
					}

					// Explicitly adopt AI settings

					if tempConfig.AI.Provider != "" {
						m.config.AI.Provider = tempConfig.AI.Provider
					}
					if tempConfig.AI.Model != "" {
						m.config.AI.Model = tempConfig.AI.Model
					}
					if tempConfig.AI.APIURL != "" {
						m.config.AI.APIURL = tempConfig.AI.APIURL
					}
					if tempConfig.AI.APICustomHeaders != "" {
						m.config.AI.APICustomHeaders = tempConfig.AI.APICustomHeaders
					}
					if tempConfig.AI.TokensMaxInput > 0 {
						m.config.AI.TokensMaxInput = tempConfig.AI.TokensMaxInput
					}
					if tempConfig.AI.TokensMaxOutput > 0 {
						m.config.AI.TokensMaxOutput = tempConfig.AI.TokensMaxOutput
					}

					// Update path as the local configuration takes precedence
					m.configPath = localConfigPath
				}
			}
		}
	}

	// 4. Load environment variables (highest priority)
	m.loadSensitiveEnvVars()

	// 5. Load repository-specific .env file if available
	m.loadEnvFile()

	return nil
}

// SaveConfig saves the configuration to a file
func (m *Manager) SaveConfig(configPath string) error {
	// Handle API key for TOML serialization
	// Only save API keys in global configuration
	isGlobalConfig := false
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfigDir := filepath.Join(homeDir, ".clikd")
		isGlobalConfig = strings.HasPrefix(configPath, globalConfigDir)
	}

	// Secure existing API keys
	apiKey := m.config.AI.APIKey

	// Only save API key in global configuration
	if !isGlobalConfig {
		m.config.AI.APIKey = ""
	}

	// Serialize TOML
	data, err := toml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	// Restore API key (if locally removed)
	m.config.AI.APIKey = apiKey

	return nil
}

// SetConfigValue sets a configuration value based on a path (e.g., "general.log_level")
func (m *Manager) SetConfigValue(path, value string) error {
	// AI is now always enabled, no special handling needed for ai.enable

	v := viper.New()
	v.SetConfigType("toml")

	// Load current configuration into Viper
	data, err := toml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
		return fmt.Errorf("error loading config into viper: %w", err)
	}

	// Set value
	v.Set(path, value)

	// Load back into our configuration structure
	if err := v.Unmarshal(&m.config); err != nil {
		return fmt.Errorf("error updating config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() Config {
	return m.config
}

// loadSensitiveEnvVars loads sensitive data from environment variables
func (m *Manager) loadSensitiveEnvVars() {
	// General configuration values from environment variables
	if logLevel := os.Getenv("CLIKD_GENERAL_LOG_LEVEL"); logLevel != "" {
		m.config.General.LogLevel = logLevel
	}

	// AI configuration values from environment variables
	// AI is now always enabled, no need to check CLIKD_AI_ENABLE

	if provider := os.Getenv("CLIKD_AI_PROVIDER"); provider != "" {
		m.config.AI.Provider = provider
	}

	if model := os.Getenv("CLIKD_MODEL"); model != "" {
		m.config.AI.Model = model
	}

	if apiURL := os.Getenv("CLIKD_API_URL"); apiURL != "" {
		m.config.AI.APIURL = apiURL
	}

	if apiHeaders := os.Getenv("CLIKD_API_CUSTOM_HEADERS"); apiHeaders != "" {
		m.config.AI.APICustomHeaders = apiHeaders
	}

	if maxInput := os.Getenv("CLIKD_TOKENS_MAX_INPUT"); maxInput != "" {
		if val, err := strconv.Atoi(maxInput); err == nil && val > 0 {
			m.config.AI.TokensMaxInput = val
		}
	}

	if maxOutput := os.Getenv("CLIKD_TOKENS_MAX_OUTPUT"); maxOutput != "" {
		if val, err := strconv.Atoi(maxOutput); err == nil && val > 0 {
			m.config.AI.TokensMaxOutput = val
		}
	}
}

// loadEnvFile loads API keys from a .env file in the project directory
// This method is specifically intended for API keys and other sensitive data
func (m *Manager) loadEnvFile() {
	// AI is now always enabled, always try to load API key
	// Load key with the simplified GetAPIKey function
	apiKey, err := utils.GetAPIKey()
	if err == nil && apiKey != "" {
		// If a key was found, save it in the configuration
		m.config.AI.APIKey = apiKey
	}
}

// MaskAPIKey masks an API key for security reasons
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
