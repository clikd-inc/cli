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
	// EnvPrefix ist das Präfix für Umgebungsvariablen
	EnvPrefix = "CLIKD"

	// DefaultConfigFileName ist der Standardname für die Konfigurationsdatei
	DefaultConfigFileName = "config"

	// DefaultConfigFileExt ist die Standarderweiterung für die Konfigurationsdatei
	DefaultConfigFileExt = "toml"
)

// Config repräsentiert die Konfiguration der Anwendung
type Config struct {
	Version string `toml:"version"`
	General struct {
		LogLevel string `toml:"log_level"`
	} `toml:"general"`
	AI struct {
		// Alle weiteren KI-Einstellungen werden über Umgebungsvariablen gesteuert.
		// CLIKD_API_KEY         - API-Schlüssel für den gewählten Provider
		// CLIKD_AI_PROVIDER     - KI-Provider ('mistral', 'openai', 'anthropic', etc.)
		// CLIKD_MODEL           - Modell (z.B. 'mistral-medium', 'gpt-4o')
		// CLIKD_API_URL         - URL für Proxy oder alternative API-Endpunkte
		// CLIKD_API_CUSTOM_HEADERS - Benutzerdefinierte HTTP-Header für API-Anfragen
		// CLIKD_TOKENS_MAX_INPUT  - Maximales Token-Limit für Eingaben (Standard: 4096)
		// CLIKD_TOKENS_MAX_OUTPUT - Maximales Token-Limit für Ausgaben (Standard: 500)
		Provider         string `toml:"provider"`
		Model            string `toml:"model"`
		APIKey           string `toml:"api_key"`
		APIURL           string `toml:"api_url"`
		APICustomHeaders string `toml:"api_custom_headers"`
		TokensMaxInput   int    `toml:"tokens_max_input"`
		TokensMaxOutput  int    `toml:"tokens_max_output"`
	} `toml:"ai"`
}

// Manager verwaltet die Konfiguration
type Manager struct {
	config     Config
	configPath string
}

// NewManager erstellt einen neuen Manager
func NewManager() *Manager {
	return &Manager{
		config: createDefaultConfig(),
	}
}

// createDefaultConfig erstellt eine Standardkonfiguration
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

// InitConfig initialisiert die Konfiguration aus einer Datei
func (m *Manager) InitConfig(configPath string) error {
	// Wenn ein expliziter Konfigurationspfad angegeben wurde, verwenden wir nur diese Datei
	if configPath != "" {
		m.configPath = configPath

		// Datei lesen
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				// Wenn die Datei nicht existiert, verwende die Standardkonfiguration
				m.config = createDefaultConfig()
				return nil
			}
			return fmt.Errorf("error reading config file: %w", err)
		}

		// TOML parsen
		if err := toml.Unmarshal(data, &m.config); err != nil {
			return fmt.Errorf("error parsing config file: %w", err)
		}

		// Sensible Daten aus Umgebungsvariablen laden
		m.loadSensitiveEnvVars()

		// Repository-spezifische .env-Datei laden, falls vorhanden
		m.loadEnvFile()

		return nil
	}

	// Wenn kein expliziter Pfad angegeben wurde, verwenden wir die Prioritätsreihenfolge
	// 1. Standardkonfiguration laden
	m.config = createDefaultConfig()

	// 2. Globale Konfiguration laden, falls vorhanden
	homedir, err := os.UserHomeDir()
	if err == nil {
		globalConfigPath := filepath.Join(homedir, ".clikd", "config.toml")
		if _, err := os.Stat(globalConfigPath); err == nil {
			// Globale Konfiguration laden
			data, err := os.ReadFile(globalConfigPath)
			if err == nil {
				// TOML parsen
				if err := toml.Unmarshal(data, &m.config); err == nil {
					// Globale Konfiguration geladen
					m.configPath = globalConfigPath
				}
			}
		}
	}

	// 3. Projektspezifische Konfiguration laden, falls vorhanden
	// Die aktuelle Projektspezifische Konfiguration überschreibt die globale
	wd, err := os.Getwd()
	if err == nil {
		localConfigPath := filepath.Join(wd, "clikd", "config.toml")
		if _, err := os.Stat(localConfigPath); err == nil {
			// Lokale Konfiguration laden
			data, err := os.ReadFile(localConfigPath)
			if err == nil {
				// Temporäre Konfiguration erstellen, um keine nicht-spezifizierten Werte zu überschreiben
				tempConfig := Config{}
				if err := toml.Unmarshal(data, &tempConfig); err == nil {
					// Nur die in der lokalen Konfiguration spezifizierten Werte übernehmen
					// Allgemeine Einstellungen
					if tempConfig.General.LogLevel != "" {
						m.config.General.LogLevel = tempConfig.General.LogLevel
					}

					// KI-Einstellungen explizit übernehmen

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

					// Pfad aktualisieren, da die lokale Konfiguration Vorrang hat
					m.configPath = localConfigPath
				}
			}
		}
	}

	// 4. Umgebungsvariablen laden (haben höchste Priorität)
	m.loadSensitiveEnvVars()

	// 5. Repository-spezifische .env-Datei laden, falls vorhanden
	m.loadEnvFile()

	return nil
}

// SaveConfig speichert die Konfiguration in einer Datei
func (m *Manager) SaveConfig(configPath string) error {
	// API-Schlüssel für TOML-Serialisierung behandeln
	// Nur in globaler Konfiguration speichern wir API-Schlüssel
	isGlobalConfig := false
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfigDir := filepath.Join(homeDir, ".clikd")
		isGlobalConfig = strings.HasPrefix(configPath, globalConfigDir)
	}

	// Sichere vorhandene API-Schlüssel
	apiKey := m.config.AI.APIKey

	// API-Schlüssel nur in globaler Konfiguration speichern
	if !isGlobalConfig {
		m.config.AI.APIKey = ""
	}

	// TOML serialisieren
	data, err := toml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	// Stelle sicher, dass das Verzeichnis existiert
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Datei schreiben
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	// API-Schlüssel wiederherstellen (falls lokal entfernt)
	m.config.AI.APIKey = apiKey

	return nil
}

// SetConfigValue setzt einen Konfigurationswert anhand eines Pfades (z.B. "general.log_level")
func (m *Manager) SetConfigValue(path, value string) error {
	// AI is now always enabled, no special handling needed for ai.enable

	v := viper.New()
	v.SetConfigType("toml")

	// Aktuelle Konfiguration in Viper laden
	data, err := toml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
		return fmt.Errorf("error loading config into viper: %w", err)
	}

	// Wert setzen
	v.Set(path, value)

	// Zurück in unsere Konfigurationsstruktur laden
	if err := v.Unmarshal(&m.config); err != nil {
		return fmt.Errorf("error updating config: %w", err)
	}

	return nil
}

// GetConfig gibt die aktuelle Konfiguration zurück
func (m *Manager) GetConfig() Config {
	return m.config
}

// loadSensitiveEnvVars lädt sensible Daten aus Umgebungsvariablen
func (m *Manager) loadSensitiveEnvVars() {
	// General-Konfigurationswerte aus Umgebungsvariablen
	if logLevel := os.Getenv("CLIKD_GENERAL_LOG_LEVEL"); logLevel != "" {
		m.config.General.LogLevel = logLevel
	}

	// AI-Konfigurationswerte aus Umgebungsvariablen
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

// loadEnvFile lädt API-Schlüssel aus einer .env-Datei im Projektverzeichnis
// Diese Methode ist speziell für API-Schlüssel und andere sensible Daten vorgesehen
func (m *Manager) loadEnvFile() {
	// AI is now always enabled, always try to load API key
	// Schlüssel laden mit der vereinfachten GetAPIKey Funktion
	apiKey, err := utils.GetAPIKey()
	if err == nil && apiKey != "" {
		// Wenn ein Schlüssel gefunden wurde, in der Konfiguration speichern
		m.config.AI.APIKey = apiKey
	}
}

// Validiere Provider-Modell-Kombination
// Diese Methode bleibt für Kompatibilität bestehen, aber ohne die Konfiguration zu überprüfen
func (m *Manager) validateProviderModel() error {
	// Da wir keine Provider/Modell-Felder mehr in der Konfiguration haben,
	// macht diese Validierung keinen Sinn mehr. Die Validierung erfolgt
	// stattdessen zur Laufzeit, wenn die API aufgerufen wird.
	return nil
}

// MaskAPIKey maskiert einen API-Schlüssel aus Sicherheitsgründen
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
