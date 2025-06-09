package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"clikd/internal/services/changelog"

	tea "github.com/charmbracelet/bubbletea"
)

// createProjectStructure creates the project directory structure
func createProjectStructure(m InitModel) tea.Cmd {
	return func() tea.Msg {
		// Create config directory if it doesn't exist
		if err := os.MkdirAll(m.ConfigDir, 0755); err != nil {
			return ProjectStructureErrorMsg{Error: fmt.Errorf("failed to create config directory: %w", err)}
		}

		// Configure changelog if enabled - KEINE Changelog-Konfiguration in config.toml!
		// Die config.toml enthält nur AI und General-Konfiguration

		// Create changelog directories and files if enabled (only for local configuration)
		if !m.Global && m.ChangelogEnabled {
			// Create changelog directory structure
			changelogDir := filepath.Join(m.ConfigDir, "changelog")
			if err := os.MkdirAll(changelogDir, 0755); err != nil {
				return ProjectStructureErrorMsg{
					Error: fmt.Errorf("failed to create changelog directory: %w", err),
				}
			}

			// Create Answer object for the changelog service
			answer := &changelog.ChangelogAnswer{
				RepositoryURL:       m.ChangelogRepositoryURL,
				Style:               m.ChangelogStyle,
				CommitMessageFormat: m.ChangelogFormat,
				Template:            m.ChangelogTemplate,
				ColorEnabled:        m.ChangelogColorEnabled,
				IncludeMerges:       m.ChangelogIncludeMerges,
				IncludeReverts:      m.ChangelogIncludeReverts,
				ConfigDir:           "clikd/changelog",
			}

			// Generate config using the config builder
			configBuilder := changelog.NewConfigBuilder()
			configContent, err := configBuilder.Build(answer)
			if err != nil {
				return ProjectStructureErrorMsg{
					Error: fmt.Errorf("failed to generate changelog config: %w", err),
				}
			}

			// Generate template using the appropriate template builder
			var templateBuilder changelog.TemplateBuilder
			switch m.ChangelogTemplate {
			case "keep-a-changelog":
				templateBuilder = changelog.NewKACTemplateBuilder()
			default:
				templateBuilder = changelog.NewCustomTemplateBuilder()
			}

			templateContent, err := templateBuilder.Build(answer)
			if err != nil {
				return ProjectStructureErrorMsg{
					Error: fmt.Errorf("failed to generate changelog template: %w", err),
				}
			}

			// Write config file: changelog/config.yml
			configPath := filepath.Join(changelogDir, "config.yml")
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return ProjectStructureErrorMsg{
					Error: fmt.Errorf("failed to create changelog config file: %w", err),
				}
			}

			// Write template file: changelog/CHANGELOG.tpl.md
			templatePath := filepath.Join(changelogDir, "CHANGELOG.tpl.md")
			if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
				return ProjectStructureErrorMsg{
					Error: fmt.Errorf("failed to create changelog template file: %w", err),
				}
			}
		}

		// Save configuration
		if err := m.Manager.SaveConfig(m.ConfigPath); err != nil {
			return ProjectStructureErrorMsg{
				Error: fmt.Errorf("failed to save configuration: %w", err),
			}
		}

		return ProjectStructureCompleteMsg{}
	}
}
