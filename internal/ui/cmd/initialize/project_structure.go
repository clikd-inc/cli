package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// createProjectStructure creates the project directory structure
func createProjectStructure(m InitModel) tea.Cmd {
	return func() tea.Msg {
		// Create config directory if it doesn't exist
		if err := os.MkdirAll(m.ConfigDir, 0755); err != nil {
			return ProjectStructureErrorMsg{Error: fmt.Errorf("failed to create config directory: %w", err)}
		}

		// Create template directories if not using global config
		if !m.Global {
			// Create necessary directories
			dirs := []string{
				"templates",
				"templates/changelog",
			}

			for _, dir := range dirs {
				dirPath := filepath.Join("clikd", dir)
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					return ProjectStructureErrorMsg{
						Error: fmt.Errorf("failed to create directory %s: %w", dir, err),
					}
				}
			}

			// Create template files
			templateFiles := map[string]string{
				filepath.Join("clikd", "templates", "changelog", "github.md"): defaultGithubTemplate,
			}

			for path, content := range templateFiles {
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					return ProjectStructureErrorMsg{
						Error: fmt.Errorf("failed to create template file %s: %w", path, err),
					}
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

// defaultGithubTemplate is the default GitHub changelog template
const defaultGithubTemplate = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

{{#each releases}}
## {{title}} {{#if date}}({{date}}){{/if}}
{{#if summary}}
{{summary}}
{{/if}}

{{#each groups}}
### {{title}}

{{#each commits}}
- {{#if id}}{{id}} {{/if}}{{message}}
{{/each}}
{{/each}}
{{/each}}
`
