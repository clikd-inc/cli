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

func (loader *configLoaderImpl) Load(configPath string) (*Config, error) {
	// Direkt die YAML-Changelog-Konfiguration laden
	config, err := loader.loadYAMLConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML config from %s: %w", configPath, err)
	}

	// Working Directory setzen (relativ zur Konfigurationsdatei)
	config.WorkingDir = filepath.Dir(configPath)

	return config, nil
}

func (loader *configLoaderImpl) loadYAMLConfig(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
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

// Normalize applies CLI context values to the config
func (config *Config) Normalize(ctx *CLIContext) error {
	// Override config with CLI context values
	if ctx.Template != "" {
		config.Template = ctx.Template
	}

	if ctx.WorkingDir != "" {
		config.WorkingDir = ctx.WorkingDir
	}

	if ctx.RepositoryURL != "" {
		config.Info.RepositoryURL = ctx.RepositoryURL
	}

	if ctx.NoCaseSensitive {
		config.Options.NoCaseSensitive = true
	}

	if ctx.NextTag != "" {
		config.Options.NextTag = ctx.NextTag
	}

	if ctx.TagFilterPattern != "" {
		config.Options.TagFilterPattern = ctx.TagFilterPattern
	}

	if ctx.Sort != "" {
		config.Options.Sort = ctx.Sort
	}

	if len(ctx.Paths) > 0 {
		config.Options.Paths = ctx.Paths
	}

	// Jira integration
	if ctx.JiraUsername != "" {
		config.Options.JiraUsername = ctx.JiraUsername
	}

	if ctx.JiraToken != "" {
		config.Options.JiraToken = ctx.JiraToken
	}

	if ctx.JiraURL != "" {
		config.Options.JiraURL = ctx.JiraURL
	}

	// Template-Pfad auflösen: Wenn der Template-Pfad relativ ist,
	// löse ihn relativ zum Konfigurationsverzeichnis auf
	if config.Template != "" && !filepath.IsAbs(config.Template) {
		config.Template = filepath.Join(filepath.Dir(ctx.ConfigPath), config.Template)
	}

	return nil
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

	// Konfiguration laden (direkt YAML, kein TOML)
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

// LoadConfig lädt die Konfiguration direkt aus der YAML-Datei
func LoadConfig(path string) (*Config, error) {
	loader := NewConfigLoader()
	return loader.Load(path)
}
