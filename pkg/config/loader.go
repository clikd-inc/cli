package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// Manager verwaltet die Konfiguration
type Manager struct {
	viper      *viper.Viper
	configFile string
	config     *ConfigData
}

// NewManager erstellt einen neuen Konfigurationsmanager
func NewManager() *Manager {
	return &Manager{
		viper:  viper.New(),
		config: DefaultConfig(),
	}
}

// InitConfig initialisiert die Konfiguration
// Wenn configFile leer ist, wird der Standardpfad verwendet
func (m *Manager) InitConfig(configFile string) error {
	m.configFile = configFile
	v := m.viper

	// Standardwerte setzen
	m.setDefaults()

	// Konfigurationspfade hinzufügen
	m.addConfigPaths()

	// Wenn eine spezifische Konfigurationsdatei angegeben wurde, verwende diese
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName(DefaultConfigFileName)
		v.SetConfigType(DefaultConfigFileExt)
	}

	// Umgebungsvariablen einlesen
	m.setupEnvVars()

	// Konfigurationsdatei einlesen
	if err := v.ReadInConfig(); err != nil {
		// Es ist in Ordnung, wenn die Konfigurationsdatei nicht gefunden wird
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Konfiguration in die Struktur übertragen
	if err := v.Unmarshal(m.config); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Sensitive Informationen aus Umgebungsvariablen einlesen
	m.loadSensitiveEnvVars()

	return nil
}

// GetConfig gibt die aktuelle Konfiguration zurück
func (m *Manager) GetConfig() *ConfigData {
	return m.config
}

// GetConfigFilePath gibt den Pfad zur verwendeten Konfigurationsdatei zurück
func (m *Manager) GetConfigFilePath() string {
	return m.viper.ConfigFileUsed()
}

// SetConfigValue setzt einen Konfigurationswert
func (m *Manager) SetConfigValue(key string, value interface{}) error {
	m.viper.Set(key, value)

	// Aktualisiere die interne Konfigurationsstruktur
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("unable to update config struct: %w", err)
	}

	return nil
}

// SaveConfig speichert die aktuelle Konfiguration in die Datei
func (m *Manager) SaveConfig(filePath string) error {
	targetPath := filePath
	if targetPath == "" {
		if m.viper.ConfigFileUsed() != "" {
			targetPath = m.viper.ConfigFileUsed()
		} else {
			// Wenn keine Konfigurationsdatei verwendet wird, speichere in Standardpfad
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("error finding home directory: %w", err)
			}

			configDir := filepath.Join(home, ".clikd")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("error creating config directory: %w", err)
			}

			targetPath = filepath.Join(configDir, "config.toml")
		}
	}

	// Stelle sicher, dass die Konfiguration aktuell ist
	m.updateViperFromConfig()

	// Verwende die WriteConfig-Methode von viper, um die Konfiguration zu speichern
	if err := m.viper.WriteConfigAs(targetPath); err != nil {
		return fmt.Errorf("error writing config to file: %w", err)
	}

	return nil
}

// setDefaults setzt die Standardwerte in viper
func (m *Manager) setDefaults() {
	defaults := DefaultConfig()

	m.viper.SetDefault("version", defaults.Version)

	// General
	m.viper.SetDefault("general.log_level", defaults.General.LogLevel)
	m.viper.SetDefault("general.color", defaults.General.Color)

	// AI
	m.viper.SetDefault("ai.enable", defaults.AI.Enable)
	m.viper.SetDefault("ai.default_model", defaults.AI.DefaultModel)
	m.viper.SetDefault("ai.default_provider", defaults.AI.DefaultProvider)
	m.viper.SetDefault("ai.verbose", defaults.AI.Verbose)

	// Voreingestellte Modelle
	for modelName, modelConfig := range defaults.AI.Models {
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.provider", modelName), modelConfig.Provider)
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.model_id", modelName), modelConfig.ModelID)
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.max_tokens", modelName), modelConfig.MaxTokens)
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.temperature", modelName), modelConfig.Temperature)
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.top_p", modelName), modelConfig.TopP)
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.context_window", modelName), modelConfig.ContextWindow)
		m.viper.SetDefault(fmt.Sprintf("ai.models.%s.stream_response", modelName), modelConfig.StreamResponse)
	}

	// Changelog
	m.viper.SetDefault("changelog.style", defaults.Changelog.Style)
	m.viper.SetDefault("changelog.template", defaults.Changelog.Template)
	m.viper.SetDefault("changelog.jira_integration", defaults.Changelog.JiraIntegration)
	m.viper.SetDefault("changelog.sort", defaults.Changelog.Sort)
	m.viper.SetDefault("changelog.tag_filter_pattern", defaults.Changelog.TagFilterPattern)
	m.viper.SetDefault("changelog.path", defaults.Changelog.Path)
	m.viper.SetDefault("changelog.no_case", defaults.Changelog.NoCase)

	// Jira
	m.viper.SetDefault("changelog.jira.base_url", defaults.Changelog.Jira.BaseURL)
	m.viper.SetDefault("changelog.jira.username", defaults.Changelog.Jira.Username)
	m.viper.SetDefault("changelog.jira.project_key", defaults.Changelog.Jira.ProjectKey)
	m.viper.SetDefault("changelog.jira.issue_pattern", defaults.Changelog.Jira.IssuePattern)
}

// addConfigPaths fügt die Standardkonfigurationspfade hinzu
func (m *Manager) addConfigPaths() {
	// Suche nach Konfigurationsdateien in Standardpfaden
	home, err := os.UserHomeDir()
	if err == nil {
		m.viper.AddConfigPath(filepath.Join(home, ".clikd"))
	}

	// Aktuelles Verzeichnis
	m.viper.AddConfigPath(".")

	// clikd Verzeichnis im aktuellen Arbeitsverzeichnis
	m.viper.AddConfigPath("clikd")
}

// setupEnvVars konfiguriert die Unterstützung für Umgebungsvariablen
func (m *Manager) setupEnvVars() {
	v := m.viper
	v.SetEnvPrefix(EnvPrefix)

	// Umgebungsvariablen automatisch erkennen
	v.AutomaticEnv()

	// Unterstützung für verschachtelte Konfiguration in Umgebungsvariablen
	// z.B. CLIKD_AI_DEFAULT_MODEL für ai.default_model
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// loadSensitiveEnvVars lädt sensible Daten aus Umgebungsvariablen
func (m *Manager) loadSensitiveEnvVars() {
	// API-Schlüssel für verschiedene Provider
	// Diese werden direkt aus den Umgebungsvariablen geladen und nicht in der Konfigurationsdatei gespeichert
	providerEnvVars := map[string]string{
		"openai":    "OPENAI_API_KEY",
		"mistral":   "MISTRAL_API_KEY",
		"anthropic": "ANTHROPIC_API_KEY",
		// Weitere Provider hier hinzufügen
	}

	// Für jedes Modell in der Konfiguration
	for modelName, modelConfig := range m.config.AI.Models {
		provider := modelConfig.Provider
		if envVar, ok := providerEnvVars[provider]; ok {
			if apiKey := os.Getenv(envVar); apiKey != "" {
				// API-Schlüssel dem Modell hinzufügen
				modelConfig.APIKey = apiKey
				m.config.AI.Models[modelName] = modelConfig
			}
		}
	}

	// Jira-API-Schlüssel
	if apiKey := os.Getenv("JIRA_API_KEY"); apiKey != "" {
		m.config.Changelog.Jira.APIKey = apiKey
	}
}

// updateViperFromConfig aktualisiert Viper mit den Werten aus der Konfigurationsstruktur
func (m *Manager) updateViperFromConfig() {
	// Stelle sicher, dass die Konfiguration aktuell ist
	// Aktualisiere zuerst den viper mit allen Werten aus der Struktur
	m.viper.Set("version", m.config.Version)

	// General
	m.viper.Set("general.log_level", m.config.General.LogLevel)
	m.viper.Set("general.color", m.config.General.Color)

	// AI
	m.viper.Set("ai.enable", m.config.AI.Enable)
	m.viper.Set("ai.default_model", m.config.AI.DefaultModel)
	m.viper.Set("ai.default_provider", m.config.AI.DefaultProvider)
	m.viper.Set("ai.verbose", m.config.AI.Verbose)

	// Models
	for modelName, modelConfig := range m.config.AI.Models {
		m.viper.Set(fmt.Sprintf("ai.models.%s.provider", modelName), modelConfig.Provider)
		m.viper.Set(fmt.Sprintf("ai.models.%s.model_id", modelName), modelConfig.ModelID)
		m.viper.Set(fmt.Sprintf("ai.models.%s.max_tokens", modelName), modelConfig.MaxTokens)
		m.viper.Set(fmt.Sprintf("ai.models.%s.temperature", modelName), modelConfig.Temperature)
		m.viper.Set(fmt.Sprintf("ai.models.%s.top_p", modelName), modelConfig.TopP)
		m.viper.Set(fmt.Sprintf("ai.models.%s.context_window", modelName), modelConfig.ContextWindow)
		m.viper.Set(fmt.Sprintf("ai.models.%s.stream_response", modelName), modelConfig.StreamResponse)
	}

	// Changelog
	m.viper.Set("changelog.style", m.config.Changelog.Style)
	m.viper.Set("changelog.template", m.config.Changelog.Template)
	m.viper.Set("changelog.jira_integration", m.config.Changelog.JiraIntegration)
	m.viper.Set("changelog.sort", m.config.Changelog.Sort)
	m.viper.Set("changelog.tag_filter_pattern", m.config.Changelog.TagFilterPattern)
	m.viper.Set("changelog.path", m.config.Changelog.Path)
	m.viper.Set("changelog.no_case", m.config.Changelog.NoCase)

	// Jira
	m.viper.Set("changelog.jira.base_url", m.config.Changelog.Jira.BaseURL)
	m.viper.Set("changelog.jira.username", m.config.Changelog.Jira.Username)
	m.viper.Set("changelog.jira.project_key", m.config.Changelog.Jira.ProjectKey)
	m.viper.Set("changelog.jira.issue_pattern", m.config.Changelog.Jira.IssuePattern)
}
