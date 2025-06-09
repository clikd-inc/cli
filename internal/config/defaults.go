package config

// DefaultConfig gibt eine Konfiguration mit Standardwerten zurück
func DefaultConfig() *ConfigData {
	return &ConfigData{
		Version: "1.0.0",
		General: GeneralConfig{
			LogLevel: "info",
			Color:    true,
		},
		AI: AIConfig{
			Enable:           true,
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
