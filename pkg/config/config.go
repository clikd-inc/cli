package config

import (
	"fmt"
	"os"
	"sync"
)

var (
	// globalManager ist die Singleton-Instanz des Konfigurationsmanagers
	globalManager *Manager

	// mutex für thread-sichere Initialisierung
	managerMutex sync.Mutex
)

// Initialize initialisiert die globale Konfiguration
// Wenn configFile leer ist, wird der Standardpfad verwendet
func Initialize(configFile string) error {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	// Neue Manager-Instanz erstellen
	manager := NewManager()

	// Konfiguration laden
	if err := manager.InitConfig(configFile); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Umgebungsvariablen explizit überprüfen und setzen (für Tests)
	// Diese Werte überschreiben alles, was über die loadSensitiveEnvVars-Methode geladen wurde
	if logLevel := os.Getenv("CLIKD_GENERAL_LOG_LEVEL"); logLevel != "" {
		manager.config.General.LogLevel = logLevel
	}

	if model := os.Getenv("CLIKD_AI_MODEL"); model != "" {
		manager.config.AI.Model = model
	}

	if provider := os.Getenv("CLIKD_AI_PROVIDER"); provider != "" {
		manager.config.AI.Provider = provider
	}

	// API-Schlüssel explizit für Tests prüfen
	if apiKey := os.Getenv("CLIKD_API_KEY"); apiKey != "" {
		manager.config.AI.APIKey = apiKey
	} else if openaiKey := os.Getenv("CLIKD_OPENAI_API_KEY"); openaiKey != "" && manager.config.AI.Provider == "openai" {
		manager.config.AI.APIKey = openaiKey
	} else if mistralKey := os.Getenv("CLIKD_MISTRAL_API_KEY"); mistralKey != "" && manager.config.AI.Provider == "mistral" {
		manager.config.AI.APIKey = mistralKey
	} else if anthropicKey := os.Getenv("CLIKD_ANTHROPIC_API_KEY"); anthropicKey != "" && manager.config.AI.Provider == "anthropic" {
		manager.config.AI.APIKey = anthropicKey
	}

	// Globale Instanz setzen
	globalManager = manager

	return nil
}

// Get gibt die aktuelle Konfiguration zurück
// Wenn die Konfiguration noch nicht initialisiert wurde, wird ein Fehler zurückgegeben
func Get() (*ConfigData, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	// Konvertiere die neue Config-Struktur in die alte ConfigData-Struktur
	config := globalManager.GetConfig()
	configData := convertToConfigData(config)
	return configData, nil
}

// GetManager gibt den globalen Konfigurationsmanager zurück
// Wenn der Manager noch nicht initialisiert wurde, wird ein Fehler zurückgegeben
func GetManager() (*Manager, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	return globalManager, nil
}

// GetConfigFilePath gibt den Pfad zur verwendeten Konfigurationsdatei zurück
func GetConfigFilePath() (string, error) {
	if globalManager == nil {
		return "", fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	// In der neuen Implementierung ist dies einfach configPath
	return globalManager.configPath, nil
}

// Set setzt einen Konfigurationswert im globalen Manager
func Set(key string, value interface{}) error {
	if globalManager == nil {
		return fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	// Konvertiere den Wert in einen String für die neue SetConfigValue-Methode
	valueStr := fmt.Sprintf("%v", value)

	// Für einige häufig verwendete Schlüssel direkt setzen
	if key == "general.log_level" {
		globalManager.config.General.LogLevel = valueStr
		return nil
	} else if key == "ai.model" {
		globalManager.config.AI.Model = valueStr
		return nil
	} else if key == "ai.provider" {
		globalManager.config.AI.Provider = valueStr
		return nil
	} else if key == "ai.enable" {
		globalManager.config.AI.Enable = (valueStr == "true")
		return nil
	}

	// Ansonsten die allgemeine SetConfigValue-Methode verwenden
	return globalManager.SetConfigValue(key, valueStr)
}

// Save speichert die aktuelle Konfiguration in die Datei
func Save(filePath string) error {
	if globalManager == nil {
		return fmt.Errorf("configuration not initialized, call Initialize() first")
	}

	return globalManager.SaveConfig(filePath)
}

// GetAIModelConfig gibt die Konfiguration für ein bestimmtes KI-Modell zurück
func GetAIModelConfig(modelName string) (ModelConfig, error) {
	config, err := Get()
	if err != nil {
		return ModelConfig{}, err
	}

	// Validiere die Provider-Modell-Kombination
	provider := config.AI.Provider
	model := config.AI.Model

	// Falls ein spezifisches Modell angefragt wurde, verwende dieses für die Validierung
	if modelName != "" && modelName != model {
		// Wenn ein anderes Modell als das konfigurierte angefragt wurde,
		// prüfen, ob es mit dem Provider kompatibel ist
		if err := ValidateProviderModel(provider, modelName); err != nil {
			return ModelConfig{}, fmt.Errorf("ungültige Konfiguration: %w", err)
		}
		// Falls kompatibel, verwende das angeforderte Modell für die Rückgabe
		model = modelName
	} else {
		// Ansonsten validiere die konfigurierte Kombination
		if err := ValidateProviderModel(provider, model); err != nil {
			return ModelConfig{}, fmt.Errorf("ungültige Konfiguration: %w", err)
		}
	}

	// ModelConfig erstellen
	modelConfig := ModelConfig{
		Provider: provider,
		ModelID:  model,
		APIKey:   config.AI.APIKey,
		Endpoint: config.AI.APIURL,
	}

	return modelConfig, nil
}

// EnsureInitialized stellt sicher, dass die Konfiguration initialisiert ist
// Wenn nicht, wird sie mit Standardwerten initialisiert
func EnsureInitialized() (*ConfigData, error) {
	if globalManager == nil {
		if err := Initialize(""); err != nil {
			return nil, err
		}
	}

	// Konvertiere die neue Config-Struktur in die alte ConfigData-Struktur
	config := globalManager.GetConfig()
	configData := convertToConfigData(config)
	return configData, nil
}

// Reset setzt die globale Konfiguration zurück (hauptsächlich für Tests)
func Reset() {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	globalManager = nil
}

// convertToConfigData konvertiert die neue Config-Struktur in die alte ConfigData-Struktur
func convertToConfigData(config Config) *ConfigData {
	configData := &ConfigData{
		Version: config.Version,
		General: GeneralConfig{
			LogLevel: config.General.LogLevel,
			Color:    config.General.Color,
		},
		AI: AIConfig{
			Enable:           config.AI.Enable,
			Provider:         config.AI.Provider,
			Model:            config.AI.Model,
			APIKey:           config.AI.APIKey,
			APIURL:           config.AI.APIURL,
			APICustomHeaders: config.AI.APICustomHeaders,
			TokensMaxInput:   config.AI.TokensMaxInput,
			TokensMaxOutput:  config.AI.TokensMaxOutput,
		},
		Changelog: ChangelogConfig{
			Style:            config.Changelog.Style,
			Template:         config.Changelog.Template,
			JiraIntegration:  config.Changelog.JiraIntegration,
			Sort:             boolToSortString(config.Changelog.Sort),
			TagFilterPattern: config.Changelog.TagFilterPattern,
			Path:             config.Changelog.Path,
			NoCase:           config.Changelog.NoCase,
			Jira: JiraConfig{
				BaseURL:      config.Changelog.Jira.BaseURL,
				Username:     config.Changelog.Jira.Username,
				APIKey:       config.Changelog.Jira.APIKey,
				ProjectKey:   config.Changelog.Jira.ProjectKey,
				IssuePattern: config.Changelog.Jira.IssuePattern,
			},
			Info: ChangelogInfoConfig{
				Title:         config.Changelog.Info.Title,
				RepositoryURL: config.Changelog.Info.RepositoryURL,
			},
			Options: ChangelogOptionsConfig{
				Commits: ChangelogCommitsConfig{
					SortBy:  config.Changelog.Options.Commits.SortBy,
					Filters: make(map[string][]string),
				},
				CommitGroups: ChangelogCommitGroupsConfig{
					GroupBy:   config.Changelog.Options.CommitGroups.GroupBy,
					SortBy:    config.Changelog.Options.CommitGroups.SortBy,
					TitleMaps: config.Changelog.Options.CommitGroups.TitleMaps,
				},
				Header: ChangelogHeaderConfig{
					Pattern:     config.Changelog.Options.Header.Pattern,
					PatternMaps: []string{},
				},
				Notes: ChangelogNotesConfig{
					Keywords: config.Changelog.Options.Notes.Keywords,
				},
			},
		},
	}

	// Konvertiere PatternMaps
	for _, patternMap := range config.Changelog.Options.Header.PatternMaps {
		if pattern, ok := patternMap["pattern"]; ok {
			configData.Changelog.Options.Header.PatternMaps = append(
				configData.Changelog.Options.Header.PatternMaps, pattern)
		}
	}

	// Konvertiere Commit-Filters
	// In der neuen Struktur ist es map[string]string, in der alten map[string][]string
	for key, value := range config.Changelog.Options.Commits.Filters {
		configData.Changelog.Options.Commits.Filters[key] = []string{value}
	}

	return configData
}

// boolToSortString konvertiert einen bool-Wert in einen Sort-String ("asc" oder "desc")
func boolToSortString(sort bool) string {
	if sort {
		return "asc"
	}
	return "desc"
}
