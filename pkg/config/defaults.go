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
			Enable:          true,
			DefaultModel:    "mistral-medium",
			DefaultProvider: "mistral",
			Verbose:         false,
			Models: map[string]ModelConfig{
				"mistral-medium": {
					Provider:       "mistral",
					ModelID:        "mistral-medium",
					MaxTokens:      1024,
					Temperature:    0.7,
					TopP:           0.9,
					ContextWindow:  8192,
					StreamResponse: false,
				},
				"mistral-small": {
					Provider:       "mistral",
					ModelID:        "mistral-small",
					MaxTokens:      1024,
					Temperature:    0.7,
					TopP:           0.9,
					ContextWindow:  8192,
					StreamResponse: false,
				},
				"gpt-3.5-turbo": {
					Provider:       "openai",
					ModelID:        "gpt-3.5-turbo",
					MaxTokens:      1024,
					Temperature:    0.7,
					TopP:           0.9,
					ContextWindow:  4096,
					StreamResponse: false,
				},
				"gpt-4": {
					Provider:       "openai",
					ModelID:        "gpt-4",
					MaxTokens:      1024,
					Temperature:    0.7,
					TopP:           0.9,
					ContextWindow:  8192,
					StreamResponse: false,
				},
			},
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
