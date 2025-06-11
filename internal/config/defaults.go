package config

import "clikd/internal/cli/version"

// DefaultConfig gibt eine Konfiguration mit Standardwerten zurück
func DefaultConfig() *ConfigData {
	return &ConfigData{
		Version: version.GetVersion(),
		General: GeneralConfig{
			LogLevel: "info",
		},
		AI: AIConfig{
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

// DefaultGeneralConfig returns the default general configuration
func DefaultGeneralConfig() GeneralConfig {
	return GeneralConfig{
		LogLevel: "info",
		// Color default removed - each service manages its own color settings
	}
}
