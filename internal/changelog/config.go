package changelog

import (
	"fmt"

	"clikd/internal/config"
)

// CommandConfig enthält die Konfiguration für den Changelog-Befehl
// Diese Struktur entspricht der config.ChangelogCommandConfig
type CommandConfig struct {
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

// LoadConfigFromGlobal lädt die Changelog-Konfiguration und konvertiert sie in das Format,
// das vom Changelog-Generator benötigt wird. Diese Funktion ersetzt die ehemalige
// LoadChangelogConfig-Funktion aus dem config-Paket.
func LoadConfigFromGlobal(globalConfig *config.ConfigData, cmdConfig *CommandConfig) (*Config, error) {
	if cmdConfig == nil {
		return nil, fmt.Errorf("changelog command config not specified")
	}

	// Wir verwenden die globale Konfiguration
	cfg := globalConfig.Changelog

	// Erstellen einer changelog.Config-Instanz
	chglogCfg := &Config{
		Bin:        "git",
		WorkingDir: cmdConfig.WorkingDir,
		Template:   cmdConfig.Template,
		Info: &Info{
			Title:         cfg.Info.Title,
			RepositoryURL: cmdConfig.RepositoryURL,
		},
		Options: &Options{
			NextTag:                     cmdConfig.NextTag,
			TagFilterPattern:            cmdConfig.TagFilterPattern,
			Sort:                        cmdConfig.Sort,
			NoCaseSensitive:             cmdConfig.NoCaseSensitive,
			Paths:                       cmdConfig.Paths,
			CommitFilters:               cfg.Options.Commits.Filters,
			CommitSortBy:                cfg.Options.Commits.SortBy,
			CommitGroupBy:               cfg.Options.CommitGroups.GroupBy,
			CommitGroupSortBy:           cfg.Options.CommitGroups.SortBy,
			CommitGroupTitleMaps:        cfg.Options.CommitGroups.TitleMaps,
			HeaderPattern:               cfg.Options.Header.Pattern,
			HeaderPatternMaps:           cfg.Options.Header.PatternMaps,
			IssuePrefix:                 []string{"#", "gh-"},
			RefActions:                  []string{"close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved"},
			NoteKeywords:                cfg.Options.Notes.Keywords,
			JiraUsername:                cmdConfig.JiraUsername,
			JiraToken:                   cmdConfig.JiraToken,
			JiraURL:                     cmdConfig.JiraURL,
			JiraTypeMaps:                make(map[string]string),
			JiraIssueDescriptionPattern: "",
		},
	}

	// Wenn kein RepositoryURL gesetzt ist, verwenden wir den aus der Konfiguration
	if chglogCfg.Info.RepositoryURL == "" {
		chglogCfg.Info.RepositoryURL = cfg.Info.RepositoryURL
	}

	// Wenn kein Template gesetzt ist, verwenden wir das aus der Konfiguration
	if chglogCfg.Template == "" {
		chglogCfg.Template = cfg.Template
	}

	// Wenn keine TagFilterPattern gesetzt ist, verwenden wir das aus der Konfiguration
	if chglogCfg.Options.TagFilterPattern == "" {
		chglogCfg.Options.TagFilterPattern = cfg.TagFilterPattern
	}

	// Wenn kein Sort gesetzt ist, verwenden wir das aus der Konfiguration
	if chglogCfg.Options.Sort == "" {
		chglogCfg.Options.Sort = cfg.Sort
	}

	return chglogCfg, nil
}

// Diese Funktion kann von CLI-Befehlen verwendet werden, um die CommandConfig aus dem
// Cobra-Befehlskontext zu erstellen und dann LoadConfigFromGlobal aufzurufen
func LoadConfigFromCommand(cmdConfig *CommandConfig) (*Config, error) {
	// Sicherstellen, dass die globale Konfiguration initialisiert ist
	globalConfig, err := config.EnsureInitialized()
	if err != nil {
		return nil, fmt.Errorf("error initializing global config: %w", err)
	}

	// Konfiguration laden
	return LoadConfigFromGlobal(globalConfig, cmdConfig)
}
