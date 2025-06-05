package config

import (
	"fmt"
)

// ConfigData repräsentiert die Hauptkonfigurationsstruktur
type ConfigData struct {
	Version   string          `mapstructure:"version"`
	General   GeneralConfig   `mapstructure:"general"`
	AI        AIConfig        `mapstructure:"ai"`
	Changelog ChangelogConfig `mapstructure:"changelog"`
	// Weitere Funktionen hier hinzufügen
}

// GeneralConfig enthält allgemeine Einstellungen
type GeneralConfig struct {
	LogLevel string `mapstructure:"log_level"`
	Color    bool   `mapstructure:"color"`
}

// AIConfig enthält KI-bezogene Einstellungen
type AIConfig struct {
	Enable          bool                   `mapstructure:"enable"`
	DefaultModel    string                 `mapstructure:"default_model"`
	DefaultProvider string                 `mapstructure:"default_provider"`
	Verbose         bool                   `mapstructure:"verbose"`
	Models          map[string]ModelConfig `mapstructure:"models"`
}

// ModelConfig enthält Einstellungen für ein spezifisches KI-Modell
type ModelConfig struct {
	Provider       string  `mapstructure:"provider"`
	ModelID        string  `mapstructure:"model_id"`
	APIKey         string  `mapstructure:"api_key,omitempty"`
	Endpoint       string  `mapstructure:"endpoint,omitempty"`
	MaxTokens      int     `mapstructure:"max_tokens"`
	Temperature    float64 `mapstructure:"temperature"`
	TopP           float64 `mapstructure:"top_p"`
	ContextWindow  int     `mapstructure:"context_window"`
	StreamResponse bool    `mapstructure:"stream_response"`
}

// ChangelogConfig enthält Changelog-bezogene Einstellungen
type ChangelogConfig struct {
	Style            string     `mapstructure:"style"`
	Template         string     `mapstructure:"template"`
	JiraIntegration  bool       `mapstructure:"jira_integration"`
	Sort             string     `mapstructure:"sort"`
	TagFilterPattern string     `mapstructure:"tag_filter_pattern"`
	Path             string     `mapstructure:"path"`
	NoCase           bool       `mapstructure:"no_case"`
	Jira             JiraConfig `mapstructure:"jira"`

	// Erweiterte Optionen, die in der alten YAML-Konfiguration vorhanden waren
	Info    ChangelogInfoConfig    `mapstructure:"info"`
	Options ChangelogOptionsConfig `mapstructure:"options"`
}

// ChangelogInfoConfig enthält Metadaten für den Changelog
type ChangelogInfoConfig struct {
	Title         string `mapstructure:"title"`
	RepositoryURL string `mapstructure:"repository_url"`
}

// ChangelogOptionsConfig enthält erweiterte Optionen für den Changelog
type ChangelogOptionsConfig struct {
	Commits      ChangelogCommitsConfig      `mapstructure:"commits"`
	CommitGroups ChangelogCommitGroupsConfig `mapstructure:"commit_groups"`
	Header       ChangelogHeaderConfig       `mapstructure:"header"`
	Notes        ChangelogNotesConfig        `mapstructure:"notes"`
}

// ChangelogCommitsConfig enthält Commit-bezogene Optionen
type ChangelogCommitsConfig struct {
	Filters map[string][]string `mapstructure:"filters"`
	SortBy  string              `mapstructure:"sort_by"`
}

// ChangelogCommitGroupsConfig enthält Commit-Gruppierungsoptionen
type ChangelogCommitGroupsConfig struct {
	GroupBy   string            `mapstructure:"group_by"`
	SortBy    string            `mapstructure:"sort_by"`
	TitleMaps map[string]string `mapstructure:"title_maps"`
}

// ChangelogHeaderConfig enthält Header-bezogene Optionen
type ChangelogHeaderConfig struct {
	Pattern     string   `mapstructure:"pattern"`
	PatternMaps []string `mapstructure:"pattern_maps"`
}

// ChangelogNotesConfig enthält Optionen für Notizen
type ChangelogNotesConfig struct {
	Keywords []string `mapstructure:"keywords"`
}

// JiraConfig enthält Jira-spezifische Einstellungen
type JiraConfig struct {
	BaseURL      string `mapstructure:"base_url"`
	Username     string `mapstructure:"username"`
	APIKey       string `mapstructure:"api_key,omitempty"`
	ProjectKey   string `mapstructure:"project_key"`
	IssuePattern string `mapstructure:"issue_pattern"`
}

// GetModelConfig gibt die Konfiguration für ein bestimmtes Modell zurück
func (c *AIConfig) GetModelConfig(modelName string) (ModelConfig, error) {
	model, exists := c.Models[modelName]
	if !exists {
		return ModelConfig{}, fmt.Errorf("model configuration not found for %s", modelName)
	}
	return model, nil
}
