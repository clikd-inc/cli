package initialize

import (
	"clikd/pkg/config"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// createProjectStructure erstellt die Projektstruktur (Verzeichnisse, Templates)
func createProjectStructure(m InitModel) tea.Cmd {
	return func() tea.Msg {
		// Schrittweise Fortschrittsanzeige simulieren
		var templatesDir string
		var cacheDir string

		// Fortschritt starten
		time.Sleep(200 * time.Millisecond)

		// Verzeichnisse basierend auf Global-Flag erstellen
		if m.Global {
			home, err := os.UserHomeDir()
			if err != nil {
				return ProjectStructureErrorMsg{
					Error: fmt.Errorf("error finding home directory: %w", err),
				}
			}
			templatesDir = filepath.Join(home, ".clikd", "templates")
			cacheDir = filepath.Join(home, ".clikd", "cache")
		} else {
			// Immer "clikd" als Verzeichnisnamen verwenden, auch wenn bereits eine Datei mit diesem Namen existiert
			templatesDir = filepath.Join("clikd", "templates")
			cacheDir = filepath.Join("clikd", "cache")
		}

		// Templates-Verzeichnis erstellen
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			// Spezifische Fehlerbehandlung für Windows-Konflikte
			if runtime.GOOS == "windows" {
				// Prüfen, ob das Problem darin besteht, dass eine Datei mit dem Namen "clikd" existiert
				fileInfo, statErr := os.Stat("clikd")
				if statErr == nil && !fileInfo.IsDir() {
					return ProjectStructureErrorMsg{
						Error: fmt.Errorf("Eine Datei mit dem Namen 'clikd' existiert bereits. Auf Windows können Dateien und Verzeichnisse nicht denselben Namen haben. Bitte umbenennen oder löschen Sie die Datei 'clikd', um fortzufahren."),
					}
				}
			}
			return ProjectStructureErrorMsg{
				Error: fmt.Errorf("error creating templates directory: %w", err),
			}
		}

		time.Sleep(200 * time.Millisecond)

		// Cache-Verzeichnis erstellen
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return ProjectStructureErrorMsg{
				Error: fmt.Errorf("error creating cache directory: %w", err),
			}
		}

		time.Sleep(200 * time.Millisecond)

		// Standard-Changelog-Template erstellen
		templatePath := filepath.Join(templatesDir, "changelog.md")
		templateContent := `# {{ .Info.Title }}

All notable changes to this project will be documented in this file.

{{ if .Versions -}}
{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }} ({{ datetime "2006-01-02" .Tag.Date }})

{{ range .CommitGroups -}}
### {{ .Title }}

{{ range .Commits -}}
* {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .RevertCommits -}}
### Reverts

{{ range .RevertCommits -}}
* {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
### Merges

{{ range .MergeCommits -}}
* {{ .Header }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}

{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}`

		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			return ProjectStructureErrorMsg{
				Error: fmt.Errorf("error creating changelog template: %w", err),
			}
		}

		// Template-Pfad in der Konfiguration aktualisieren
		m.Manager.SetConfigValue("changelog.template", filepath.Join("templates", "changelog.md"))

		time.Sleep(200 * time.Millisecond)

		// Konfiguration speichern
		if err := m.Manager.SaveConfig(m.ConfigPath); err != nil {
			return ProjectStructureErrorMsg{
				Error: fmt.Errorf("error saving configuration: %w", err),
			}
		}

		// Zum Zusammenfassungsschritt wechseln
		return ProjectStructureCompleteMsg{}
	}
}

// setupChangelogDefaultOptions setzt die Standardoptionen für den Changelog
func setupChangelogDefaultOptions(manager *config.Manager) {
	// HeaderPattern und PatternMaps für die Commit-Nachrichtenanalyse
	manager.SetConfigValue("changelog.options.header.pattern", "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$")

	// PatternMaps als Array setzen
	patternMaps := []string{"Type", "Scope", "Subject"}
	for i, pattern := range patternMaps {
		manager.SetConfigValue(fmt.Sprintf("changelog.options.header.pattern_maps.%d", i), pattern)
	}

	// Commit-Filter konfigurieren - als Array für jeden Typ
	commitTypes := []string{"feat", "fix", "perf", "refactor", "chore"}
	for i, commitType := range commitTypes {
		manager.SetConfigValue(fmt.Sprintf("changelog.options.commits.filters.Type.%d", i), commitType)
	}

	// Gruppierung von Commits
	manager.SetConfigValue("changelog.options.commit_groups.group_by", "Type")
	manager.SetConfigValue("changelog.options.commit_groups.sort_by", "Title")

	// Title-Mappings für Commit-Gruppen
	manager.SetConfigValue("changelog.options.commit_groups.title_maps.feat", "Features")
	manager.SetConfigValue("changelog.options.commit_groups.title_maps.fix", "Bug Fixes")
	manager.SetConfigValue("changelog.options.commit_groups.title_maps.perf", "Performance Improvements")
	manager.SetConfigValue("changelog.options.commit_groups.title_maps.refactor", "Code Refactoring")
	manager.SetConfigValue("changelog.options.commit_groups.title_maps.chore", "Chores")

	// Note Keywords für Breaking Changes, etc.
	noteKeywords := []string{"BREAKING CHANGE", "SECURITY"}
	for i, keyword := range noteKeywords {
		manager.SetConfigValue(fmt.Sprintf("changelog.options.notes.keywords.%d", i), keyword)
	}
}
