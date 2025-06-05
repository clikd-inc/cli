package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		Enable          bool               `toml:"enable"`
		DefaultModel    string             `toml:"default_model"`
		DefaultProvider string             `toml:"default_provider"`
		Verbose         bool               `toml:"verbose"`
		Models          map[string]AIModel `toml:"models"`
	} `toml:"ai"`
}

// AIModel repräsentiert die Konfiguration eines AI-Modells
type AIModel struct {
	Provider       string  `toml:"provider"`
	ModelID        string  `toml:"model_id"`
	APIKey         string  `toml:"api_key,omitempty"`
	MaxTokens      int     `toml:"max_tokens,omitempty"`
	Temperature    float64 `toml:"temperature,omitempty"`
	TopP           float64 `toml:"top_p,omitempty"`
	ContextWindow  int     `toml:"context_window,omitempty"`
	StreamResponse bool    `toml:"stream_response,omitempty"`
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
	c.AI.DefaultModel = "gpt-4"
	c.AI.DefaultProvider = "openai"
	c.AI.Verbose = false
	c.AI.Models = make(map[string]AIModel)

	// Vorkonfigurierte Modelle
	c.AI.Models["gpt-4"] = AIModel{
		Provider:       "openai",
		ModelID:        "gpt-4",
		MaxTokens:      4000,
		Temperature:    0.7,
		TopP:           1.0,
		ContextWindow:  8000,
		StreamResponse: true,
	}
	c.AI.Models["gpt-3.5-turbo"] = AIModel{
		Provider:       "openai",
		ModelID:        "gpt-3.5-turbo",
		MaxTokens:      4000,
		Temperature:    0.7,
		TopP:           1.0,
		ContextWindow:  4000,
		StreamResponse: true,
	}
	c.AI.Models["claude-3-opus"] = AIModel{
		Provider:       "anthropic",
		ModelID:        "claude-3-opus-20240229",
		MaxTokens:      4000,
		Temperature:    0.7,
		TopP:           1.0,
		ContextWindow:  8000,
		StreamResponse: true,
	}
	c.AI.Models["mistral-large"] = AIModel{
		Provider:       "mistral",
		ModelID:        "mistral-large",
		MaxTokens:      4000,
		Temperature:    0.7,
		TopP:           1.0,
		ContextWindow:  4000,
		StreamResponse: true,
	}

	// Mistral Medium Modell für Tests
	c.AI.Models["mistral-medium"] = AIModel{
		Provider:       "mistral",
		ModelID:        "mistral-medium",
		MaxTokens:      4000,
		Temperature:    0.7,
		TopP:           1.0,
		ContextWindow:  4000,
		StreamResponse: true,
	}

	return c
}

// InitConfig initialisiert die Konfiguration aus einer Datei
func (m *Manager) InitConfig(configPath string) error {
	if configPath == "" {
		// Wenn kein Pfad angegeben wurde, verwende die Standardkonfiguration
		m.config = createDefaultConfig()
		return nil
	}

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

// SaveConfig speichert die Konfiguration in einer Datei
func (m *Manager) SaveConfig(configPath string) error {
	// Speichere API-Schlüssel (temporär)
	apiKeys := make(map[string]map[string]string)

	// Sichere vorhandene API-Schlüssel
	for modelName, model := range m.config.AI.Models {
		if model.APIKey != "" {
			if apiKeys[model.Provider] == nil {
				apiKeys[model.Provider] = make(map[string]string)
			}
			apiKeys[model.Provider][modelName] = model.APIKey
		}
	}

	// API-Schlüssel für TOML-Serialisierung entfernen (sollten nicht in Datei gespeichert werden)
	// Nur in globaler Konfiguration speichern wir API-Schlüssel
	isGlobalConfig := false
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfigDir := filepath.Join(homeDir, ".clikd")
		isGlobalConfig = strings.HasPrefix(configPath, globalConfigDir)
	}

	// API-Schlüssel nur in globaler Konfiguration speichern
	if !isGlobalConfig {
		for modelName, model := range m.config.AI.Models {
			if model.APIKey != "" {
				model.APIKey = "" // Entferne API-Schlüssel aus lokaler Konfiguration
				m.config.AI.Models[modelName] = model
			}
		}

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

	// API-Schlüssel wiederherstellen (nach dem Speichern)
	for provider, models := range apiKeys {
		for modelName, apiKey := range models {
			if model, ok := m.config.AI.Models[modelName]; ok && model.Provider == provider {
				model.APIKey = apiKey
				m.config.AI.Models[modelName] = model
			}
		}
	}

	return nil
}

// SetConfigValue setzt einen Konfigurationswert anhand eines Pfades (z.B. "general.log_level")
func (m *Manager) SetConfigValue(path, value string) error {
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
// Diese Methode sollte NICHT Umgebungsvariablen aus der Shell laden
func (m *Manager) loadSensitiveEnvVars() {
	// Diese Methode lädt keine API-Schlüssel mehr aus Umgebungsvariablen
	// Stattdessen erfolgt das API-Key-Management ausschließlich über:
	// 1. .env-Datei für lokale Projekte (über loadEnvFile)
	// 2. Globale Konfiguration für globale Einstellungen

	// Die Überprüfung auf Umgebungsvariablen wurde entfernt, da API-Schlüssel
	// nicht aus Shell-Umgebungsvariablen geladen werden sollen.

	// Generelle Konfigurationswerte (keine API-Schlüssel) dürfen weiterhin
	// aus Umgebungsvariablen kommen.

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

	if model := os.Getenv("CLIKD_AI_DEFAULT_MODEL"); model != "" {
		m.config.AI.DefaultModel = model
	}

	if provider := os.Getenv("CLIKD_AI_DEFAULT_PROVIDER"); provider != "" {
		m.config.AI.DefaultProvider = provider
	}

	if verbose := os.Getenv("CLIKD_AI_VERBOSE"); verbose == "true" || verbose == "1" {
		m.config.AI.Verbose = true
	} else if verbose == "false" || verbose == "0" {
		m.config.AI.Verbose = false
	}

	// Andere, nicht-API-Key-Konfigurationswerte aus Umgebungsvariablen
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
	// Versuche, .env-Datei zu öffnen
	envFile, err := os.Open(".env")
	if err != nil {
		// .env-Datei existiert nicht oder kann nicht geöffnet werden
		return
	}
	defer envFile.Close()

	// Definiere eine Liste von API-Schlüsseln, die geladen werden sollen
	apiKeyPatterns := []struct {
		Provider string
		EnvVar   string
	}{
		{"openai", "CLIKD_OPENAI_API_KEY"},
		{"openai", "OPENAI_API_KEY"}, // Unterstützung für Legacy-Format
		{"mistral", "CLIKD_MISTRAL_API_KEY"},
		{"mistral", "MISTRAL_API_KEY"}, // Unterstützung für Legacy-Format
		{"anthropic", "CLIKD_ANTHROPIC_API_KEY"},
		{"anthropic", "ANTHROPIC_API_KEY"}, // Unterstützung für Legacy-Format
		{"azure-openai", "CLIKD_AZURE_OPENAI_API_KEY"},
		{"azure-openai", "AZURE_OPENAI_API_KEY"}, // Unterstützung für Legacy-Format
	}

	// Jira-API-Schlüssel
	jiraKeyPatterns := []string{
		"CLIKD_JIRA_API_KEY",
		"JIRA_API_KEY", // Unterstützung für Legacy-Format
	}

	// .env-Datei zeilenweise lesen
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Kommentare und leere Zeilen überspringen
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Nach KEY=VALUE-Format suchen
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Anführungszeichen entfernen, falls vorhanden
		value = strings.Trim(value, `"'`)

		// API-Schlüssel für AI-Modelle setzen
		for _, pattern := range apiKeyPatterns {
			if key == pattern.EnvVar {
				// Für jedes Modell mit dem passenden Provider den API-Schlüssel setzen
				for modelName, modelConfig := range m.config.AI.Models {
					if modelConfig.Provider == pattern.Provider {
						modelConfig.APIKey = value
						m.config.AI.Models[modelName] = modelConfig
					}
				}
				break
			}
		}

		// Jira-API-Schlüssel setzen
		for _, pattern := range jiraKeyPatterns {
			if key == pattern {
				m.config.Changelog.Jira.APIKey = value
				break
			}
		}
	}
}

// MaskAPIKey maskiert einen API-Schlüssel aus Sicherheitsgründen
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
