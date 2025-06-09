package changelog

import (
	"fmt"
	"os"
	"path/filepath"
)

// Service stellt Funktionen für die Changelog-Verwaltung bereit
type Service struct {
	ConfigPath string
}

// NewService erstellt einen neuen Changelog-Service
func NewService(configPath string) *Service {
	return &Service{
		ConfigPath: configPath,
	}
}

// InitializeTemplates erstellt die Template- und Konfigurationsdateien
func (s *Service) InitializeTemplates(style string, configDir string) error {
	// Erstelle die Verzeichnisse
	templateDir := filepath.Join(configDir, "templates")
	configDir = filepath.Join(configDir, "config")

	dirs := []string{templateDir, configDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("fehler beim Erstellen des Verzeichnisses %s: %w", dir, err)
		}
	}

	// Template- und Konfigurationsdateien schreiben
	templatePath := filepath.Join(templateDir, style+".tpl.md")
	configPath := filepath.Join(configDir, style+".yml")

	// Erstelle ein Answer-Objekt (wird nicht verwendet)
	// aber kann in zukünftigen Implementierungen nützlich sein

	// Generiere das Template und die Konfiguration
	templateContent := getDefaultTemplate(style)
	configContent := getDefaultConfig(style)

	// Schreibe das Template
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("fehler beim Schreiben der Template-Datei: %w", err)
	}

	// Schreibe die Konfiguration
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("fehler beim Schreiben der Konfigurations-Datei: %w", err)
	}

	return nil
}

// EnsureTemplateExists stellt sicher, dass die Template-Datei existiert
// Falls nicht, wird sie aus dem eingebetteten Template wiederhergestellt
func (s *Service) EnsureTemplateExists(templatePath, style string) error {
	if templatePath == "" {
		return nil // Keine Template-Datei konfiguriert
	}

	// Prüfe, ob die Datei existiert
	if _, err := os.Stat(templatePath); err == nil {
		return nil // Datei existiert
	}

	// Stelle sicher, dass das Verzeichnis existiert
	dir := filepath.Dir(templatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Template-Verzeichnisses: %w", err)
	}

	// Schreibe das Template
	templateContent := getDefaultTemplate(style)
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("fehler beim Wiederherstellen des Templates: %w", err)
	}

	fmt.Printf("Template-Datei wurde wiederhergestellt: %s\n", templatePath)
	return nil
}

// EnsureConfigExists stellt sicher, dass die Konfigurations-Datei existiert
// Falls nicht, wird sie aus der eingebetteten Konfiguration wiederhergestellt
func (s *Service) EnsureConfigExists(configPath, style string) error {
	if configPath == "" {
		return nil // Keine Konfigurations-Datei konfiguriert
	}

	// Prüfe, ob die Datei existiert
	if _, err := os.Stat(configPath); err == nil {
		return nil // Datei existiert
	}

	// Stelle sicher, dass das Verzeichnis existiert
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Konfigurations-Verzeichnisses: %w", err)
	}

	// Schreibe die Konfiguration
	configContent := getDefaultConfig(style)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("fehler beim Wiederherstellen der Konfiguration: %w", err)
	}

	fmt.Printf("Konfigurations-Datei wurde wiederhergestellt: %s\n", configPath)
	return nil
}

// getDefaultTemplate liefert das Standard-Template für den angegebenen Stil
func getDefaultTemplate(style string) string {
	switch style {
	case "github":
		return `# {{.Info.Title}}
{{range .Versions}}
<a name="{{.Tag.Name}}"></a>
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	case "gitlab":
		return `# {{.Info.Title}}
{{range .Versions}}
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	case "bitbucket":
		return `# {{.Info.Title}}
{{range .Versions}}
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	default:
		return `# {{.Info.Title}}
{{range .Versions}}
## {{if .Tag.Previous}}[{{.Tag.Name}}]({{$.Info.RepositoryURL}}/compare/{{.Tag.Previous.Name}}...{{.Tag.Name}}){{else}}{{.Tag.Name}}{{end}} ({{datetime "2006-01-02" .Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
* {{.Subject}}{{end}}
{{end}}
{{range .NoteGroups}}
### {{.Title}}
{{range .Notes}}
{{.Body}}
{{end}}
{{end}}
{{end}}`
	}
}

// getDefaultConfig liefert die Standard-Konfiguration für den angegebenen Stil
func getDefaultConfig(style string) string {
	switch style {
	case "github":
		return `style: github
template: templates/github.tpl.md
info:
  title: CHANGELOG
  repository_url: https://github.com/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	case "gitlab":
		return `style: gitlab
template: templates/gitlab.tpl.md
info:
  title: CHANGELOG
  repository_url: https://gitlab.com/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	case "bitbucket":
		return `style: bitbucket
template: templates/bitbucket.tpl.md
info:
  title: CHANGELOG
  repository_url: https://bitbucket.org/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	default:
		return `style: github
template: templates/github.tpl.md
info:
  title: CHANGELOG
  repository_url: https://github.com/clikd-inc/cli
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE`
	}
}
