package config

// DefaultConfig gibt eine Konfiguration mit Standardwerten zurück
func DefaultConfig() *ConfigData {
	return &ConfigData{
		Version: "1.0.0",
		General: GeneralConfig{
			LogLevel: "info",
			Color:    true,
		},
		AI: AIConfig{
			Enable:           true,
			Provider:         "mistral",
			Model:            "mistral-medium",
			APIKey:           "",
			APIURL:           "",
			APICustomHeaders: "",
			TokensMaxInput:   4096,
			TokensMaxOutput:  500,
		},
		Changelog: ChangelogConfig{
			Style:            "github",
			Template:         "templates/changelog.md",
			JiraIntegration:  false,
			Sort:             "semver",
			TagFilterPattern: "v*",
			Path:             "",
			NoCase:           false,
			Jira: JiraConfig{
				BaseURL:      "",
				Username:     "",
				ProjectKey:   "",
				IssuePattern: "[A-Z]+-[0-9]+",
			},
			Info: ChangelogInfoConfig{
				Title:         "CHANGELOG",
				RepositoryURL: "",
			},
			Options: ChangelogOptionsConfig{
				Commits: ChangelogCommitsConfig{
					SortBy: "Scope",
					Filters: map[string][]string{
						"Type": {"feat", "fix", "perf", "refactor"},
					},
				},
				CommitGroups: ChangelogCommitGroupsConfig{
					GroupBy: "Type",
					SortBy:  "Title",
					TitleMaps: map[string]string{
						"feat":     "Features",
						"fix":      "Bug Fixes",
						"perf":     "Performance Improvements",
						"refactor": "Code Refactoring",
					},
				},
				Header: ChangelogHeaderConfig{
					Pattern:     "^(\\w*)\\:\\s(.*)$",
					PatternMaps: []string{"Type", "Subject"},
				},
				Notes: ChangelogNotesConfig{
					Keywords: []string{"BREAKING CHANGE"},
				},
			},
		},
	}
}
