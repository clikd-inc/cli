package config

import "clikd/internal/cli/version"

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: version.GetVersion(),
		General: struct {
			LogLevel string `toml:"log_level"`
		}{
			LogLevel: "info",
		},
		AI: struct {
			Provider         string `toml:"provider"`
			Model            string `toml:"model"`
			APIKey           string `toml:"api_key"`
			APIURL           string `toml:"api_url"`
			APICustomHeaders string `toml:"api_custom_headers"`
			TokensMaxInput   int    `toml:"tokens_max_input"`
			TokensMaxOutput  int    `toml:"tokens_max_output"`
		}{
			Provider:         "mistral",
			Model:            "mistral-medium",
			APIKey:           "",
			APIURL:           "",
			APICustomHeaders: "",
			TokensMaxInput:   4096,
			TokensMaxOutput:  500,
		},
	}
}
