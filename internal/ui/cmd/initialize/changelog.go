package initialize

import (
	"clikd/internal/config"
	"clikd/internal/ui/bubble"
)

// configureChangelog handles the changelog configuration process
// and returns a map of configuration values
func configureChangelog() map[string]string {
	options := make(map[string]string)

	// Ask if user wants to configure changelog features
	configureChangelog := bubble.RunConfirm(
		"Changelog Configuration",
		"Do you want to configure changelog features?",
	)

	if !configureChangelog {
		return options
	}

	// Select changelog style
	styleItems := []bubble.SelectItem{
		{
			Title:       "github",
			Description: "RECOMMENDED - GitHub-style with Markdown (most widely used)",
			Value:       "github",
		},
		{
			Title:       "gitlab",
			Description: "GitLab-style with Markdown",
			Value:       "gitlab",
		},
		{
			Title:       "bitbucket",
			Description: "Bitbucket-style with Markdown",
			Value:       "bitbucket",
		},
	}
	selectedStyle := bubble.RunSelect("Select Changelog Style", styleItems)
	if selectedStyle != nil {
		options["style"] = selectedStyle.Value.(string)
	} else {
		options["style"] = "github"
	}

	// Configure JIRA integration
	configureJira := bubble.RunConfirm(
		"JIRA Integration",
		"Do you want to enable JIRA integration for your changelog?",
	)
	options["jira"] = BoolToString(configureJira)

	if configureJira {
		jiraPrefix := bubble.RunInput(
			"JIRA Prefix",
			"Enter your JIRA project prefix (e.g., PROJ)",
			"PROJ",
		)
		options["jira_prefix"] = jiraPrefix
	}

	// Configure sort order
	sortItems := []bubble.SelectItem{
		{
			Title:       "Newest first",
			Description: "Show newest entries at the top",
			Value:       "desc",
		},
		{
			Title:       "Oldest first",
			Description: "Show oldest entries at the top",
			Value:       "asc",
		},
	}
	selectedSort := bubble.RunSelect("Changelog Sort Order", sortItems)
	if selectedSort != nil {
		options["sort"] = selectedSort.Value.(string)
	} else {
		options["sort"] = "desc"
	}

	// Configure advanced options
	configureAdvanced := bubble.RunConfirm(
		"Advanced Changelog Options",
		"Do you want to configure advanced changelog options?",
	)

	if !configureAdvanced {
		// Use defaults
		options["tag_filter_pattern"] = "v*"
		options["path"] = "CHANGELOG.md"
		options["no_case"] = "false"
		return options
	}

	// Configure tag filter pattern
	tagFilter := bubble.RunInput(
		"Tag Filter Pattern",
		"Enter pattern to filter tags (default: v*)",
		"v*",
	)
	options["tag_filter_pattern"] = tagFilter

	// Configure changelog path
	changelogPath := bubble.RunInput(
		"Changelog Path",
		"Enter path to the changelog file",
		"CHANGELOG.md",
	)
	options["path"] = changelogPath

	// Configure case sensitivity
	caseInsensitive := bubble.RunConfirm(
		"Case Sensitivity",
		"Make changelog generation case-insensitive?",
	)
	options["no_case"] = BoolToString(caseInsensitive)

	return options
}

// setupDefaultChangelogOptions sets up default changelog options in the config manager
func setupDefaultChangelogOptions(manager *config.Manager) {
	// Set default commit types and patterns
	manager.SetConfigValue("changelog.types", "feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert")
	manager.SetConfigValue("changelog.types.feat", "Features")
	manager.SetConfigValue("changelog.types.fix", "Bug Fixes")
	manager.SetConfigValue("changelog.types.docs", "Documentation")
	manager.SetConfigValue("changelog.types.style", "Styles")
	manager.SetConfigValue("changelog.types.refactor", "Code Refactoring")
	manager.SetConfigValue("changelog.types.perf", "Performance Improvements")
	manager.SetConfigValue("changelog.types.test", "Tests")
	manager.SetConfigValue("changelog.types.build", "Builds")
	manager.SetConfigValue("changelog.types.ci", "Continuous Integration")
	manager.SetConfigValue("changelog.types.chore", "Chores")
	manager.SetConfigValue("changelog.types.revert", "Reverts")
}
