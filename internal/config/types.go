package config

import (
	"os"
)

// ConfigData repräsentiert die Hauptkonfigurationsstruktur
type ConfigData struct {
	Version string        `mapstructure:"version"`
	General GeneralConfig `mapstructure:"general"`
	AI      AIConfig      `mapstructure:"ai"`
	// Changelog-Konfiguration wird direkt aus YAML geladen, nicht aus TOML
	// Weitere Funktionen hier hinzufügen
}

// GeneralConfig enthält allgemeine Einstellungen
type GeneralConfig struct {
	LogLevel string `mapstructure:"log_level"`
	// Color field removed - each service manages its own color settings
}

// AIConfig enthält die Konfiguration für KI-bezogene Funktionen
type AIConfig struct {
	// Provider ist der KI-Anbieter (z.B. "mistral", "openai", "anthropic")
	Provider string `json:"provider" mapstructure:"provider" toml:"provider"`
	// Model ist das zu verwendende Modell (z.B. "mistral-medium", "gpt-4o")
	Model string `json:"model" mapstructure:"model" toml:"model"`
	// APIKey ist der API-Schlüssel für den Provider (nur in globaler Konfiguration gespeichert)
	APIKey string `json:"api_key,omitempty" mapstructure:"api_key,omitempty" toml:"api_key,omitempty"`
	// APIURL ist ein optionaler, benutzerdefinierter API-Endpunkt oder Proxy
	APIURL string `json:"api_url,omitempty" mapstructure:"api_url,omitempty" toml:"api_url,omitempty"`
	// APICustomHeaders sind benutzerdefinierte HTTP-Header für API-Anfragen (als JSON-String)
	APICustomHeaders string `json:"api_custom_headers,omitempty" mapstructure:"api_custom_headers,omitempty" toml:"api_custom_headers,omitempty"`
	// TokensMaxInput ist das maximale Token-Limit für Eingaben (Standard: 4096)
	TokensMaxInput int `json:"tokens_max_input" mapstructure:"tokens_max_input" toml:"tokens_max_input"`
	// TokensMaxOutput ist das maximale Token-Limit für Ausgaben (Standard: 500)
	TokensMaxOutput int `json:"tokens_max_output" mapstructure:"tokens_max_output" toml:"tokens_max_output"`
}

// ModelConfig wird für die API-Kompatibilität beibehalten
type ModelConfig struct {
	// Provider ist der KI-Anbieter (z.B. "openai", "mistral")
	Provider string `json:"provider" mapstructure:"provider"`
	// ModelID ist die spezifische Modell-ID beim Anbieter (z.B. "gpt-4", "mistral-medium")
	ModelID string `json:"model_id" mapstructure:"model_id"`
	// APIKey ist der API-Schlüssel für diesen Anbieter (nur in globaler Konfiguration)
	APIKey string `json:"api_key,omitempty" mapstructure:"api_key,omitempty"`
	// Endpoint ist ein optionaler, benutzerdefinierter API-Endpunkt
	Endpoint string `json:"endpoint,omitempty" mapstructure:"endpoint,omitempty"`
}

// ChangelogCommandConfig enthält die Konfiguration für den Changelog-Befehl
// Diese Struktur ersetzt den alten CLIContext des Initializers
type ChangelogCommandConfig struct {
	WorkingDir       string
	ConfigPath       string
	Template         string
	RepositoryURL    string
	OutputPath       string
	Silent           bool
	NoColor          bool
	NoEmoji          bool
	NoCaseSensitive  bool
	Query            string
	NextTag          string
	TagFilterPattern string
	JiraUsername     string
	JiraToken        string
	JiraURL          string
	Paths            []string
	Sort             string
}

// ChangelogConfig enthält Changelog-bezogene Einstellungen
type ChangelogConfig struct {
	Template   string `mapstructure:"template"`
	ConfigFile string `mapstructure:"config_file"`
}

// ChangelogInfoConfig enthält Metadaten für den Changelog
type ChangelogInfoConfig struct {
	Title         string `mapstructure:"title"`
	RepositoryURL string `mapstructure:"repository_url"`
}

// GetModelConfig gibt die Konfiguration für ein bestimmtes Modell zurück
func (c *AIConfig) GetModelConfig(modelName string) (ModelConfig, error) {
	// Zuerst den Provider aus der Umgebungsvariable lesen
	provider := os.Getenv("CLIKD_AI_PROVIDER")
	if provider == "" {
		// Standardwert, wenn nicht gesetzt
		provider = "mistral"
	}

	// Das Modell aus der Umgebungsvariable lesen, wenn nicht explizit angegeben
	if modelName == "" {
		modelName = os.Getenv("CLIKD_MODEL")
		if modelName == "" {
			// Standardwert, wenn nicht gesetzt
			modelName = "mistral-medium"
		}
	}

	// Validiere das angeforderte Modell mit dem Provider
	if err := ValidateProviderModel(provider, modelName); err != nil {
		// Bei fehlerhafter Kombination Fehler zurückgeben
		return ModelConfig{}, err
	}

	// API-URL aus Umgebungsvariable lesen
	endpoint := os.Getenv("CLIKD_API_URL")

	// Falls wir hier ankommen, ist die Kombination gültig
	model := ModelConfig{
		Provider: provider,
		ModelID:  modelName,
		APIKey:   "", // API-Schlüssel wird über GetAPIKey abgerufen
		Endpoint: endpoint,
	}

	return model, nil
}
