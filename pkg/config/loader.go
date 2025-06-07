package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"clikd/pkg/utils"

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
		Color    bool   `toml:"color"`
	} `toml:"general"`
	Changelog struct {
		Style            string `toml:"style"`
		Template         string `toml:"template"`
		Sort             bool   `toml:"sort"`
		JiraIntegration  bool   `toml:"jira_integration"`
		TagFilterPattern string `toml:"tag_filter_pattern"`
		Path             string `toml:"path"`
		NoCase           bool   `toml:"no_case"`
		Info             struct {
			Title         string `toml:"title"`
			RepositoryURL string `toml:"repository_url"`
		} `toml:"info"`
		Options struct {
			Commits struct {
				SortBy  string            `toml:"sort_by"`
				Filters map[string]string `toml:"filters"`
			} `toml:"commits"`
			CommitGroups struct {
				GroupBy   string            `toml:"group_by"`
				SortBy    string            `toml:"sort_by"`
				TitleMaps map[string]string `toml:"title_maps"`
			} `toml:"commit_groups"`
			Header struct {
				Pattern     string              `toml:"pattern"`
				PatternMaps []map[string]string `toml:"pattern_maps"`
			} `toml:"header"`
			Notes struct {
				Keywords []string `toml:"keywords"`
			} `toml:"notes"`
		} `toml:"options"`
		Jira struct {
			URL          string `toml:"url"`
			BaseURL      string `toml:"base_url"`
			Username     string `toml:"username"`
			APIKey       string `toml:"api_key"`
			ProjectKey   string `toml:"project_key"`
			IssuePattern string `toml:"issue_pattern"`
		} `toml:"jira"`
	} `toml:"changelog"`
	AI struct {
		// Enable aktiviert oder deaktiviert alle KI-Funktionen global.
		// Alle weiteren KI-Einstellungen werden über Umgebungsvariablen gesteuert.
		// CLIKD_API_KEY         - API-Schlüssel für den gewählten Provider
		// CLIKD_AI_PROVIDER     - KI-Provider ('mistral', 'openai', 'anthropic', etc.)
		// CLIKD_MODEL           - Modell (z.B. 'mistral-medium', 'gpt-4o')
		// CLIKD_API_URL         - URL für Proxy oder alternative API-Endpunkte
		// CLIKD_API_CUSTOM_HEADERS - Benutzerdefinierte HTTP-Header für API-Anfragen
		// CLIKD_TOKENS_MAX_INPUT  - Maximales Token-Limit für Eingaben (Standard: 4096)
		// CLIKD_TOKENS_MAX_OUTPUT - Maximales Token-Limit für Ausgaben (Standard: 500)
		Enable           bool   `toml:"enable"`
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
		Version: "1.0.0",
	}

	// General
	c.General.LogLevel = "info"
	c.General.Color = true

	// Changelog
	c.Changelog.Style = "conventional"
	c.Changelog.Template = ""
	c.Changelog.Sort = true
	c.Changelog.JiraIntegration = false
	c.Changelog.TagFilterPattern = ""
	c.Changelog.Path = "CHANGELOG.md"
	c.Changelog.NoCase = false

	// Changelog Info
	c.Changelog.Info.Title = ""
	c.Changelog.Info.RepositoryURL = ""

	// Changelog Options - Commits
	c.Changelog.Options.Commits.SortBy = "scope"
	c.Changelog.Options.Commits.Filters = make(map[string]string)

	// Changelog Options - CommitGroups
	c.Changelog.Options.CommitGroups.GroupBy = "type"
	c.Changelog.Options.CommitGroups.SortBy = "title"
	c.Changelog.Options.CommitGroups.TitleMaps = map[string]string{
		"feat":     "Features",
		"fix":      "Bug Fixes",
		"perf":     "Performance Improvements",
		"refactor": "Code Refactoring",
		"docs":     "Documentation",
		"test":     "Tests",
		"build":    "Build System",
		"ci":       "Continuous Integration",
	}

	// Changelog Options - Header
	c.Changelog.Options.Header.Pattern = "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
	c.Changelog.Options.Header.PatternMaps = []map[string]string{
		{"pattern": "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"},
	}

	// Changelog Options - Notes
	c.Changelog.Options.Notes.Keywords = []string{"BREAKING CHANGE", "DEPRECATED"}

	// Changelog Jira
	c.Changelog.Jira.BaseURL = ""
	c.Changelog.Jira.Username = ""
	c.Changelog.Jira.ProjectKey = ""
	c.Changelog.Jira.IssuePattern = "([A-Z]+-\\d+)"

	// AI
	c.AI.Enable = true
	c.AI.Provider = "mistral"
	c.AI.Model = "mistral-medium"
	c.AI.APIKey = "" // Wird nur in der globalen Konfiguration gespeichert
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
					// AI.Enable muss immer übernommen werden, unabhängig vom Wert
					m.config.AI.Enable = tempConfig.AI.Enable

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

					// Changelog-Einstellungen übernehmen
					if tempConfig.Changelog.Style != "" {
						m.config.Changelog.Style = tempConfig.Changelog.Style
					}
					if tempConfig.Changelog.Template != "" {
						m.config.Changelog.Template = tempConfig.Changelog.Template
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

		// Jira API-Schlüssel ebenfalls entfernen
		if m.config.Changelog.Jira.APIKey != "" {
			m.config.Changelog.Jira.APIKey = ""
		}
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
	// Sonderbehandlung für AI-Einstellungen
	if path == "ai.enable" {
		if value == "true" || value == "1" {
			m.config.AI.Enable = true
		} else if value == "false" || value == "0" {
			m.config.AI.Enable = false
		}
		return nil
	}

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

// loadSensitiveEnvVars lädt sensible Daten aus der globalen Konfiguration
// Diese Methode sollte NICHT API-Schlüssel aus der Shell laden
func (m *Manager) loadSensitiveEnvVars() {
	// General-Konfigurationswerte aus Umgebungsvariablen
	if logLevel := os.Getenv("CLIKD_GENERAL_LOG_LEVEL"); logLevel != "" {
		m.config.General.LogLevel = logLevel
	}

	// AI-Konfigurationswerte aus Umgebungsvariablen
	if enableAI := os.Getenv("CLIKD_AI_ENABLE"); enableAI == "true" || enableAI == "1" {
		m.config.AI.Enable = true
	} else if enableAI == "false" || enableAI == "0" {
		m.config.AI.Enable = false
	}

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

	// Andere, nicht-AI-Konfigurationswerte aus Umgebungsvariablen
	if style := os.Getenv("CLIKD_CHANGELOG_STYLE"); style != "" {
		m.config.Changelog.Style = style
	}

	if template := os.Getenv("CLIKD_CHANGELOG_TEMPLATE"); template != "" {
		m.config.Changelog.Template = template
	}

	if sort := os.Getenv("CLIKD_CHANGELOG_SORT"); sort == "true" || sort == "1" {
		m.config.Changelog.Sort = true
	} else if sort == "false" || sort == "0" {
		m.config.Changelog.Sort = false
	}
}

// loadEnvFile lädt API-Schlüssel aus einer .env-Datei im Projektverzeichnis
// Diese Methode ist speziell für API-Schlüssel und andere sensible Daten vorgesehen
func (m *Manager) loadEnvFile() {
	// Prüfen, ob eine lokale Konfiguration existiert
	localConfigExists := utils.IsLocalConfigPresent()

	// Provider-spezifische Informationen basierend auf der aktuellen Konfiguration
	var providerInfo utils.ProviderKeyInfo

	// Provider-spezifische Konfiguration auswählen
	switch strings.ToLower(m.config.AI.Provider) {
	case "openai":
		providerInfo = utils.OpenAIProvider
	case "mistral":
		providerInfo = utils.MistralProvider
	case "anthropic":
		providerInfo = utils.AnthropicProvider
	case "groq":
		providerInfo = utils.GroqProvider
	case "openrouter":
		providerInfo = utils.OpenRouterProvider
	default:
		// Fallback zu generischem Provider
		providerInfo = utils.ProviderKeyInfo{
			Name:            m.config.AI.Provider,
			ConfigKey:       "ai.api_key",
			EnvVarName:      "CLIKD_API_KEY",
			EnvVarNameShort: "API_KEY",
			Required:        false,
		}
	}

	// API-Schlüssel für den aktuellen Provider abrufen
	// Nur wenn KI aktiviert ist, markieren wir den Schlüssel als erforderlich
	if m.config.AI.Enable {
		providerInfo.Required = true
	}

	// Schlüssel laden mit Fallback-Logik
	apiKey, err := utils.GetAPIKey(providerInfo, localConfigExists)
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
