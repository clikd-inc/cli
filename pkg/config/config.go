package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	// General configuration
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`

	// API configuration (example)
	API struct {
		Endpoint string `mapstructure:"endpoint"`
		Token    string `mapstructure:"token"`
	} `mapstructure:"api"`

	// AI configuration
	AI struct {
		Enabled            bool   `mapstructure:"enabled"`
		DefaultModel       string `mapstructure:"default_model"`
		EnhanceMessages    bool   `mapstructure:"enhance_messages"`
		GenerateSummaries  bool   `mapstructure:"generate_summaries"`
		CategorizeCommits  bool   `mapstructure:"categorize_commits"`
		SuggestVersionBump bool   `mapstructure:"suggest_version_bump"`
	} `mapstructure:"ai"`

	// Other configuration sections can be added here
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(configFile string) (*Config, error) {
	config := &Config{}

	// Set defaults
	setDefaults()

	// Find home directory for default config location
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error finding home directory: %w", err)
	}

	// Search for config in standard locations
	viper.AddConfigPath(filepath.Join(home, ".clikd"))
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// If a specific config file is specified, use it
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	// Read environment variables prefixed with CLIKD_
	viper.SetEnvPrefix("CLIKD")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return config, nil
}

// setDefaults sets the default values for configuration
func setDefaults() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "text")
	viper.SetDefault("api.endpoint", "https://api.example.com")

	// AI defaults
	viper.SetDefault("ai.enabled", false)
	viper.SetDefault("ai.default_model", "mistral-medium")
	viper.SetDefault("ai.enhance_messages", true)
	viper.SetDefault("ai.generate_summaries", true)
	viper.SetDefault("ai.categorize_commits", true)
	viper.SetDefault("ai.suggest_version_bump", true)
}

// GetConfigFilePath returns the path to the config file that was used
func GetConfigFilePath() string {
	return viper.ConfigFileUsed()
}
