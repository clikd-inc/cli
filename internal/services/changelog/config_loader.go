package changelog

import (
	"clikd/internal/utils"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigLoader ...
type ConfigLoader interface {
	Load(string) (*Config, error)
}

type configLoaderImpl struct {
}

// NewConfigLoader ...
func NewConfigLoader() ConfigLoader {
	return &configLoaderImpl{}
}

func (loader *configLoaderImpl) Load(path string) (*Config, error) {
	fp := filepath.Clean(path)
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// CommandConfig enthält die Konfiguration für den Changelog-Befehl
type CommandConfig struct {
	// Basis-Konfiguration
	WorkingDir    string
	ConfigPath    string
	Template      string
	RepositoryURL string
	OutputPath    string

	// Filter und Optionen
	Silent           bool
	NoColor          bool
	NoEmoji          bool
	NoCaseSensitive  bool
	Query            string
	NextTag          string
	TagFilterPattern string
	Sort             string

	// Jira-Integration
	JiraUsername string
	JiraToken    string
	JiraURL      string

	// Commit-Filter
	Paths []string
}

// LoadConfigFromCommand lädt die Konfiguration aus dem CommandConfig-Objekt
func LoadConfigFromCommand(cmdConfig *CommandConfig) (*Config, error) {
	logger := utils.NewLogger("debug", !cmdConfig.NoColor)
	logger.Debug("Loading configuration from command")

	// Überprüfen, ob die Konfigurationsdatei existiert
	if _, err := os.Stat(cmdConfig.ConfigPath); err != nil {
		return nil, fmt.Errorf("config file not found: %s", cmdConfig.ConfigPath)
	}

	// Template-Existenz prüfen
	if cmdConfig.Template != "" {
		if _, err := os.Stat(cmdConfig.Template); err != nil {
			return nil, fmt.Errorf("template file not found: %s", cmdConfig.Template)
		}
	}

	// CLI-Kontext erstellen
	ctx := &CLIContext{
		WorkingDir:       cmdConfig.WorkingDir,
		ConfigPath:       cmdConfig.ConfigPath,
		Template:         cmdConfig.Template,
		RepositoryURL:    cmdConfig.RepositoryURL,
		OutputPath:       cmdConfig.OutputPath,
		Silent:           cmdConfig.Silent,
		NoColor:          cmdConfig.NoColor,
		NoEmoji:          cmdConfig.NoEmoji,
		NoCaseSensitive:  cmdConfig.NoCaseSensitive,
		Query:            cmdConfig.Query,
		NextTag:          cmdConfig.NextTag,
		TagFilterPattern: cmdConfig.TagFilterPattern,
		JiraUsername:     cmdConfig.JiraUsername,
		JiraToken:        cmdConfig.JiraToken,
		JiraURL:          cmdConfig.JiraURL,
		Paths:            cmdConfig.Paths,
		Sort:             cmdConfig.Sort,
	}

	// Konfiguration aus YAML laden
	config, err := LoadConfig(cmdConfig.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Konfiguration normalisieren
	if err := config.Normalize(ctx); err != nil {
		return nil, err
	}

	// Fertige Konfiguration zurückgeben
	return config, nil
}

// LoadConfig lädt die Konfiguration aus einer YAML-Datei
func LoadConfig(path string) (*Config, error) {
	// Implementation needed - this is a placeholder
	// The actual implementation would read and parse the YAML file
	return &Config{}, nil
}
