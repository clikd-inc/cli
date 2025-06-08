# Implementierungsplan: Verbesserung der Changelog-Funktionalität in der clikd CLI

Dieser Plan beschreibt die notwendigen Änderungen, um die Changelog-Funktionalität der clikd CLI zu verbessern und alle Funktionen des git-chglog-Tools zu unterstützen.

## 1. Problembeschreibung

Die aktuelle Implementierung des Changelog-Befehls in der clikd CLI ist unvollständig und unterstützt nicht alle Funktionen, die im Original-git-chglog-Tool verfügbar sind:

1. **Unvollständige Konfigurationsabfrage:** Die CLI fragt nur einen Teil der in `defaults.go` und `config.go` definierten Konfigurationsoptionen ab.
2. **Fehlende Template-Generierung:** Das ursprüngliche Tool generiert verschiedene Template-Dateien basierend auf Benutzerauswahlen.
3. **Eingeschränkte Formatierungsoptionen:** Verschiedene Commit-Nachrichtenformate werden nicht unterstützt.
4. **Unvollständige Filteroptionen:** Detaillierte Filteroptionen für Commits fehlen.

## 2. Ziel

Vollständige Integration der git-chglog-Funktionalität in die clikd CLI, einschließlich:

1. Erweiterter Konfigurationsabfragen
2. Unterstützung verschiedener Template-Stile
3. Flexibler Commit-Nachrichtenformate
4. Umfassender Filteroptionen
5. JIRA-Integration

## 3. Implementierungsschritte

### 3.1 Überarbeitung der Konfigurationsstruktur

#### 3.1.1 Aktualisierung von config.go/defaults.go

```go
// In internal/config/defaults.go
func DefaultConfig() *ConfigData {
    return &ConfigData{
        // Bestehende Konfiguration...
        
        Changelog: ChangelogConfig{
            // Bestehende Felder...
            CommitMessageFormat: "type-scope-subject",
            TemplateStyle: "standard",
            CommitFilters: map[string][]string{
                "Type": {"feat", "fix", "perf", "refactor", "docs", "style", "test", "build", "ci", "chore", "revert"},
            },
            HeaderPattern: `^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$`,
            HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
            IncludeMerges: true,
            IncludeReverts: true,
            // Weitere Optionen...
        },
    }
}

// In internal/config/config.go
type ChangelogConfig struct {
    // Bestehende Felder...
    CommitMessageFormat string
    TemplateStyle       string
    CommitFilters       map[string][]string
    HeaderPattern       string
    HeaderPatternMaps   []string
    IncludeMerges       bool
    IncludeReverts      bool
    // Weitere Felder...
}
```

### 3.2 Verbesserung der Initialisierungs-UI

#### 3.2.1 Erweiterung der Changelog-Konfiguration in internal/ui/cmd/initialize/changelog.go

```go
// In internal/ui/cmd/initialize/changelog.go
func configureChangelog() map[string]string {
    options := make(map[string]string)

    // Bestehender Code...
    
    // 1. Commit-Nachrichtenformat-Auswahl
    commitFormatItems := []bubble.SelectItem{
        {
            Title:       "<type>(<scope>): <subject>",
            Description: "RECOMMENDED - feat(core): Add new feature",
            Value:       "type-scope-subject",
        },
        {
            Title:       "<type>: <subject>",
            Description: "feat: Add new feature",
            Value:       "type-subject",
        },
        {
            Title:       "<<type> subject>",
            Description: "Basic Git commit - Add new feature",
            Value:       "git-basic",
        },
        {
            Title:       "<subject>",
            Description: "Simple text without type detection",
            Value:       "subject",
        },
        {
            Title:       ":<emoji>: <subject>",
            Description: ":sparkles: Add new feature",
            Value:       "emoji",
        },
    }
    
    selectedFormat := bubble.RunSelect("Select Commit Message Format", commitFormatItems)
    if selectedFormat != nil {
        options["commit_format"] = selectedFormat.Value.(string)
    } else {
        options["commit_format"] = "type-scope-subject"
    }
    
    // 2. Template-Stil-Auswahl
    templateItems := []bubble.SelectItem{
        {
            Title:       "keep-a-changelog",
            Description: "Format similar to Keep A Changelog",
            Value:       "keep-a-changelog",
        },
        {
            Title:       "standard",
            Description: "Standard format with markdown",
            Value:       "standard",
        },
        {
            Title:       "cool",
            Description: "Cool format with emojis and better styling",
            Value:       "cool",
        },
    }
    
    selectedTemplate := bubble.RunSelect("Select Template Style", templateItems)
    if selectedTemplate != nil {
        options["template_style"] = selectedTemplate.Value.(string)
    } else {
        options["template_style"] = "standard"
    }
    
    // 3. Commit-Filter-Konfiguration
    if bubble.RunConfirm("Configure Commit Filters", "Do you want to configure commit filters?") {
        // Default commit types
        defaultTypes := []string{"feat", "fix", "perf", "refactor", "docs", "style", "test", "build", "ci", "chore", "revert"}
        
        // Multiselect für Commit-Typen
        typeItems := make([]bubble.SelectItem, len(defaultTypes))
        for i, t := range defaultTypes {
            typeItems[i] = bubble.SelectItem{
                Title:       t,
                Description: getTypeDescription(t),
                Value:       t,
            }
        }
        
        selectedTypes := bubble.RunMultiSelect(
            "Select Commit Types to Include",
            "Choose which commit types should be included in the changelog",
            typeItems,
        )
        
        if len(selectedTypes) > 0 {
            typeValues := make([]string, len(selectedTypes))
            for i, item := range selectedTypes {
                typeValues[i] = item.Value.(string)
            }
            options["commit_types"] = strings.Join(typeValues, ",")
        } else {
            options["commit_types"] = "feat,fix,perf,refactor"
        }
    }
    
    // 4. Merge/Revert Commit-Optionen
    options["include_merges"] = BoolToString(bubble.RunConfirm(
        "Include Merge Commits",
        "Do you want to include merge commits in the changelog?",
    ))
    
    options["include_reverts"] = BoolToString(bubble.RunConfirm(
        "Include Revert Commits",
        "Do you want to include revert commits in the changelog?",
    ))
    
    // Bestehender JIRA-Konfigurationscode...
    
    return options
}

// Hilfsfunktion für Beschreibungen der Commit-Typen
func getTypeDescription(typeName string) string {
    descriptions := map[string]string{
        "feat":     "New features",
        "fix":      "Bug fixes",
        "perf":     "Performance improvements",
        "refactor": "Code refactoring",
        "docs":     "Documentation updates",
        "style":    "Code style changes (formatting, etc.)",
        "test":     "Adding or updating tests",
        "build":    "Build system changes",
        "ci":       "CI configuration changes",
        "chore":    "Regular maintenance tasks",
        "revert":   "Reverting previous changes",
    }
    
    if desc, ok := descriptions[typeName]; ok {
        return desc
    }
    return typeName
}
```

### 3.3 Implementierung der Template-Generierung

#### 3.3.1 Neue Datei: internal/ui/cmd/initialize/templates.go

```go
package initialize

import (
    "fmt"
    "os"
    "path/filepath"
)

// Generiert Templates basierend auf den ausgewählten Optionen
func generateTemplates(options map[string]string, configDir string) error {
    templateStyle := options["template_style"]
    if templateStyle == "" {
        templateStyle = "standard"
    }
    
    var templateContent string
    switch templateStyle {
    case "keep-a-changelog":
        templateContent = keepAChangelogTemplate
    case "cool":
        templateContent = coolTemplate
    case "standard":
        fallthrough
    default:
        templateContent = standardTemplate
    }
    
    // Verzeichnis erstellen falls notwendig
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return fmt.Errorf("failed to create config directory: %w", err)
    }
    
    // Template-Datei schreiben
    templatePath := filepath.Join(configDir, "CHANGELOG.tpl.md")
    if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
        return fmt.Errorf("failed to write template file: %w", err)
    }
    
    return nil
}

// Konstantendefinitionen für die verschiedenen Template-Stile
const (
    // Standard Template (basierend auf git-chglog's standard template)
    standardTemplate = `{{ if .Versions -}}
<a name="unreleased"></a>
## [Unreleased]

{{ if .Unreleased.CommitGroups -}}
{{ range .Unreleased.CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]{{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}
{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .RevertCommits -}}
### Reverts
{{ range .RevertCommits -}}
- {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
### Pull Requests
{{ range .MergeCommits -}}
- {{ .Header }}
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

{{- if .Versions }}
[Unreleased]: {{ .Info.RepositoryURL }}/compare/{{ $latest := index .Versions 0 }}{{ $latest.Tag.Name }}...HEAD
{{ range .Versions -}}
{{ if .Tag.Previous -}}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
{{ end -}}
{{ end -}}
{{ end -}}`

    // Keep-A-Changelog Template
    keepAChangelogTemplate = `# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

{{ if .Versions -}}
{{ if .Unreleased.CommitGroups -}}
## [Unreleased]
{{ range .Unreleased.CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ range .Versions }}
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]{{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}
{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .RevertCommits -}}
### Reverts
{{ range .RevertCommits -}}
- {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
### Merged
{{ range .MergeCommits -}}
- {{ .Header }}
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

{{- if .Versions }}
[Unreleased]: {{ .Info.RepositoryURL }}/compare/{{ $latest := index .Versions 0 }}{{ $latest.Tag.Name }}...HEAD
{{ range .Versions -}}
{{ if .Tag.Previous -}}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
{{ end -}}
{{ end -}}
{{ end -}}`

    // Cool Template mit Emojis
    coolTemplate = `# 📝 Changelog
{{ if .Versions -}}
{{ if .Unreleased.CommitGroups -}}
## 🚀 Upcoming Release
{{ range .Unreleased.CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{ range .Versions }}
## 🔖 {{ if .Tag.Previous }}[{{ .Tag.Name }}]{{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}
{{ range .CommitGroups -}}
### {{ if eq .Title "Features" }}✨ {{ else if eq .Title "Bug Fixes" }}🐛 {{ else if eq .Title "Performance Improvements" }}⚡ {{ else if eq .Title "Code Refactoring" }}♻️ {{ else }}{{ end }}{{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .RevertCommits -}}
### ⏪ Reverts
{{ range .RevertCommits -}}
- {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
### 🔀 Merged
{{ range .MergeCommits -}}
- {{ .Header }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### ❗ {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}

{{- if .Versions }}
[Unreleased]: {{ .Info.RepositoryURL }}/compare/{{ $latest := index .Versions 0 }}{{ $latest.Tag.Name }}...HEAD
{{ range .Versions -}}
{{ if .Tag.Previous -}}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
{{ end -}}
{{ end -}}
{{ end -}}`
)
```

### 3.4 Generierung der Konfigurationsdatei

#### 3.4.1 Erweiterte Implementierung in internal/ui/cmd/initialize/project_structure.go

```go
func createProjectStructure(model *InitModel) error {
    // Bestehender Code...
    
    // Changelog-Konfiguration generieren, wenn entsprechende Optionen vorhanden sind
    if _, ok := model.ConfigOptions["style"]; ok {
        // Konfigurationsverzeichnis erstellen
        configDir := filepath.Join(model.ProjectDir, "clikd")
        if err := os.MkdirAll(configDir, 0755); err != nil {
            return fmt.Errorf("failed to create config directory: %w", err)
        }
        
        // Templates generieren
        if err := generateTemplates(model.ConfigOptions, configDir); err != nil {
            return fmt.Errorf("error generating templates: %w", err)
        }
        
        // Changelog-Konfigurationsdatei generieren
        if err := generateChangelogConfig(model.ConfigOptions, configDir); err != nil {
            return fmt.Errorf("error generating changelog config: %w", err)
        }
    }
    
    return nil
}

// Neue Funktion: Generiert die Changelog-Konfigurationsdatei
func generateChangelogConfig(options map[string]string, configDir string) error {
    // Standardwerte für die Konfiguration
    commitFormat := options["commit_format"]
    if commitFormat == "" {
        commitFormat = "type-scope-subject"
    }
    
    // Header-Pattern und PatternMaps basierend auf dem ausgewählten Format
    var headerPattern string
    var patternMaps []string
    
    switch commitFormat {
    case "type-scope-subject":
        headerPattern = `^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$`
        patternMaps = []string{"Type", "Scope", "Subject"}
    case "type-subject":
        headerPattern = `^(\\w*)\\:\\s(.*)$`
        patternMaps = []string{"Type", "Subject"}
    case "git-basic":
        headerPattern = `^((\\w+)\\s.*)$`
        patternMaps = []string{"Subject", "Type"}
    case "subject":
        headerPattern = `^(.*)$`
        patternMaps = []string{"Subject"}
    case "emoji":
        headerPattern = `^:(\\w*)\\:\\s(.*)$`
        patternMaps = []string{"Type", "Subject"}
    default:
        headerPattern = `^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$`
        patternMaps = []string{"Type", "Scope", "Subject"}
    }
    
    // Commit-Typen verarbeiten
    commitTypes := []string{"feat", "fix", "perf", "refactor"}
    if types, ok := options["commit_types"]; ok && types != "" {
        commitTypes = strings.Split(types, ",")
    }
    
    // TitleMaps für die verschiedenen Typen
    titleMaps := map[string]string{
        "feat":     "Features",
        "fix":      "Bug Fixes",
        "perf":     "Performance Improvements",
        "refactor": "Code Refactoring",
        "docs":     "Documentation",
        "style":    "Styles",
        "test":     "Tests",
        "build":    "Builds",
        "ci":       "Continuous Integration",
        "chore":    "Chores",
        "revert":   "Reverts",
    }
    
    // Weitere Konfigurationsoptionen
    includeMerges := true
    if val, ok := options["include_merges"]; ok {
        includeMerges = val == "true"
    }
    
    includeReverts := true
    if val, ok := options["include_reverts"]; ok {
        includeReverts = val == "true"
    }
    
    // Konfiguration für Jira
    jiraEnabled := false
    if val, ok := options["jira"]; ok {
        jiraEnabled = val == "true"
    }
    
    jiraPrefix := "PROJ"
    if val, ok := options["jira_prefix"]; ok && val != "" {
        jiraPrefix = val
    }
    
    // Konfigurationsdatei-Struktur erstellen
    configContent := fmt.Sprintf(`style: %s
template: CHANGELOG.tpl.md
info:
  title: CHANGELOG
  repository_url: %s
options:
  commits:
    # Filters to select specific commits
    filters:
      Type:
        %s
    sort_by: Scope
  commit_groups:
    # Group commits by field and sort groups
    group_by: Type
    sort_by: Title
    title_maps:
      %s
  header:
    # Commit message header pattern
    pattern: "%s"
    pattern_maps:
      %s
  notes:
    keywords:
      - BREAKING CHANGE
`, 
        options["style"],
        options["repository_url"],
        formatCommitTypes(commitTypes),
        formatTitleMaps(titleMaps, commitTypes),
        headerPattern,
        formatPatternMaps(patternMaps))
    
    // Konfigurationsdatei schreiben
    configPath := filepath.Join(configDir, "config.yml")
    if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }
    
    return nil
}

// Hilfsfunktionen für die Formatierung
func formatCommitTypes(types []string) string {
    result := ""
    for _, t := range types {
        result += fmt.Sprintf("- %s\n        ", t)
    }
    return result
}

func formatTitleMaps(maps map[string]string, types []string) string {
    result := ""
    for _, t := range types {
        if title, ok := maps[t]; ok {
            result += fmt.Sprintf("%s: %s\n      ", t, title)
        } else {
            result += fmt.Sprintf("%s: %s\n      ", t, strings.Title(t))
        }
    }
    return result
}

func formatPatternMaps(maps []string) string {
    result := ""
    for _, m := range maps {
        result += fmt.Sprintf("- %s\n      ", m)
    }
    return result
}
```

### 3.5 Anpassung der Konfigurationsverwaltung

#### 3.5.1 Aktualisierung in internal/ui/cmd/initialize/state.go

```go
func (m InitModel) Init() tea.Cmd {
    // Bestehender Code...
    
    // Konfigurationswerte vorinitialisieren
    m.ConfigOptions = make(map[string]string)
    
    return nil
}

// Update-Funktion aktualisieren, um die Erweiterungen zu unterstützen
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Bestehender Code...
    
    // Nach der Änderung des Zustands
    switch m.State {
    // Bestehende Cases...
    
    case StateChangelogConfig:
        // Changelog-Konfiguration erfassen
        m.ConfigOptions = configureChangelog()
        return m, m.nextState()
    
    // Bestehende Cases...
    }
    
    return m, nil
}
```

### 3.6 Integration mit dem changelog-Befehl

#### 3.6.1 Erweiterung des Changelog-Commands in internal/cli/commands/changelog/changelog.go

```go
// Flags in NewChangelogCmd erweitern
func NewChangelogCmd() *cobra.Command {
    // Bestehender Code...
    
    // Erweiterte Flags hinzufügen
    cmd.Flags().BoolVar(&includeMergesFlag, "include-merges", true, "Include merge commits in changelog")
    cmd.Flags().BoolVar(&includeRevertsFlag, "include-reverts", true, "Include revert commits in changelog")
    cmd.Flags().StringVar(&commitFormatFlag, "commit-format", "", "Commit message format to use")
    cmd.Flags().StringVar(&commitGroupByFlag, "group-by", "", "Field to group commits by")
    cmd.Flags().StringVar(&commitSortByFlag, "sort-by", "", "Field to sort commits by")
    
    return cmd
}

// Aktualisierung in runGenerator
func runGenerator(query string) error {
    // Bestehender Code...
    
    // Erweiterte Konfigurationsparameter übernehmen
    cmdConfig := &changelog.CommandConfig{
        // Bestehende Felder...
        
        IncludeMerges:   includeMergesFlag,
        IncludeReverts:  includeRevertsFlag,
        CommitFormat:    commitFormatFlag,
        CommitGroupBy:   commitGroupByFlag,
        CommitSortBy:    commitSortByFlag,
    }
    
    // Bestehender Code...
}
```

### 3.7 Aktualisierung des Changelog-Generators

#### 3.7.1 Erweiterung in internal/services/changelog/generator.go

```go
// CommandConfig erweitern
type CommandConfig struct {
    // Bestehende Felder...
    
    IncludeMerges   bool
    IncludeReverts  bool
    CommitFormat    string
    CommitGroupBy   string
    CommitSortBy    string
}

// Generator anpassen, um alle Parameter zu berücksichtigen
func (g *Generator) Generate(writer io.Writer, query string) error {
    // Bestehender Code...
    
    // Commit-Format, Gruppierung und Sortierung aus der Konfiguration übernehmen
    // Entsprechende Änderungen an der Generierungslogik...
    
    return nil
}
```

### 3.8 Updates in der Dokumentation

#### 3.8.1 Aktualisierung der Hilfetexte und Dokumentation

```go
// In internal/cli/commands/changelog/changelog.go
func NewChangelogCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "changelog [options] <tag query>",
        Short: "Generate a changelog from git history",
        Long: `Generate a changelog from git history using conventional commits.

This command generates a changelog based on conventional commit messages
in your git history. It supports various customization options and integrations.

Tag Query Formats:
  1. <old>..<new>  - Commits between <old> tag and <new> tag (e.g., v1.0.0..v2.0.0)
  2. <tag>..       - Commits from <tag> to the latest (e.g., v1.0.0..)
  3. ..<tag>       - Commits from the oldest tag to <tag> (e.g., ..v1.0.0)
  4. <tag>         - Commits contained in <tag> only (e.g., v1.0.0)

Special Features:
  - Multiple Template Styles: Choose from "standard", "keep-a-changelog", or "cool" templates
  - Flexible Commit Formats: Support for various commit message formats
    - <type>(<scope>): <subject> (e.g., feat(core): Add feature)
    - <type>: <subject> (e.g., feat: Add feature)
    - And more...
  - Jira Integration: Automatically fetches ticket information when Jira IDs are 
    present in commit messages. Configure via --jira-* flags or environment variables.
  - Path Filtering: Filter commits by specific files or directories with --path.
  - Tag Filtering: Filter tags using regular expressions with --tag-filter-pattern.
  - Semver Sorting: Sort tags by semantic version instead of date with --sort=semver.

Examples:
  # Initialize changelog configuration with the global init command
  clikd init

  # Generate changelog for all tags to stdout
  clikd changelog

  # Generate changelog to file
  clikd changelog -o CHANGELOG.md

  # Generate changelog for specific tag range
  clikd changelog v1.0.0..v2.0.0 -o CHANGELOG.md

  # Generate changelog with "unreleased" commits as next version
  clikd changelog --next-tag v2.0.0 -o CHANGELOG.md

  # Filter commits by path
  clikd changelog --path="internal/,cmd/" -o CHANGELOG.md

  # Include or exclude merge commits
  clikd changelog --include-merges=false -o CHANGELOG.md`,
        // ...
    }
    // ...
}
```

## 4. Testplan

1. **Grundlegende Initialisierung testen:**
   - `clikd init` ausführen und die Changelog-Konfigurationsoption auswählen
   - Verschiedene Kombinationen von Optionen testen

2. **Template-Generierung testen:**
   - Für jeden Template-Stil überprüfen, ob die Dateien korrekt generiert werden
   - Inhalt der Templates validieren

3. **Commit-Format-Parsing testen:**
   - Verschiedene Commit-Formate mit dem Changelog-Generator testen
   - Überprüfen, ob die Ausgabe korrekt formatiert ist

4. **JIRA-Integration testen:**
   - Commits mit JIRA-Referenzen erstellen
   - Testen, ob JIRA-Integration funktioniert

5. **Filteroptionen testen:**
   - Verschiedene Filter anwenden und Ergebnisse überprüfen

## 5. Zeitplan

1. **Phase 1 (Vorbereitung): 2-3 Tage**
   - Analyse des bestehenden Codes
   - Detaillierter Entwurf der Änderungen

2. **Phase 2 (Implementierung): 5-7 Tage**
   - Konfigurationsänderungen
   - UI-Erweiterungen
   - Template-Generierung
   - Changelog-Generator-Anpassungen

3. **Phase 3 (Tests und Bugfixes): 3-4 Tage**
   - Umfassende Tests
   - Bugfixes und Optimierungen

4. **Phase 4 (Dokumentation und Finalisierung): 2 Tage**
   - Dokumentation aktualisieren
   - Code-Review und letzte Anpassungen

## 6. Datei-Mapping und Übernahme-Analyse

### 6.1 Direkt übernehmbare Dateien aus git-chglog-master

#### 6.1.1 Template-Builder (können direkt kopiert und angepasst werden)

**Quelle:** `example/git-chglog-master/cmd/git-chglog/kac_template_builder.go`
**Ziel:** `internal/ui/cmd/initialize/template_builders/kac_template_builder.go`
```bash
cp example/git-chglog-master/cmd/git-chglog/kac_template_builder.go internal/ui/cmd/initialize/template_builders/kac_template_builder.go
```
**Anpassungen erforderlich:**
- Package-Name von `main` zu `template_builders` ändern
- Import-Pfade anpassen
- Konstanten aus `variables.go` übernehmen

**Quelle:** `example/git-chglog-master/cmd/git-chglog/custom_template_builder.go`
**Ziel:** `internal/ui/cmd/initialize/template_builders/custom_template_builder.go`
```bash
cp example/git-chglog-master/cmd/git-chglog/custom_template_builder.go internal/ui/cmd/initialize/template_builders/custom_template_builder.go
```
**Anpassungen erforderlich:**
- Package-Name von `main` zu `template_builders` ändern
- Import-Pfade anpassen
- Template-Konstanten aus `variables.go` übernehmen

#### 6.1.2 Konfigurationsdefinitionen (teilweise übernehmbar)

**Quelle:** `example/git-chglog-master/cmd/git-chglog/config.go` (Strukturen)
**Ziel:** `internal/config/changelog_types.go` (neue Datei)
```bash
# Manuelle Übernahme der Strukturen erforderlich
```
**Übernehmbare Strukturen:**
- `CommitOptions`
- `CommitGroupOptions` 
- `PatternOptions`
- `IssueOptions`
- `RefOptions`
- `NoteOptions`
- `JiraClientInfoOptions`
- `JiraIssueOptions`
- `JiraOptions`

**Anpassungen erforderlich:**
- Package-Name anpassen
- YAML-Tags beibehalten
- In bestehende `ChangelogConfig` integrieren

#### 6.1.3 Commit-Format-Definitionen

**Quelle:** `example/git-chglog-master/cmd/git-chglog/variables.go`
**Ziel:** `internal/ui/cmd/initialize/commit_formats.go` (neue Datei)
```bash
# Manuelle Übernahme erforderlich
```
**Übernehmbare Definitionen:**
- `CommitMessageFormat` Struktur
- Alle Format-Definitionen (`fmtTypeScopeSubject`, `fmtTypeSubject`, etc.)
- Template-Style-Definitionen (`tplKeepAChangelog`, `tplStandard`, `tplCool`)

#### 6.1.4 Konfigurationsnormalisierung

**Quelle:** `example/git-chglog-master/cmd/git-chglog/config.go` (Normalisierungsmethoden)
**Ziel:** `internal/config/changelog_normalization.go` (neue Datei)
**Übernehmbare Methoden:**
- `normalizeStyleOfGitHub()`
- `normalizeStyleOfGitLab()`
- `normalizeStyleOfBitbucket()`
- `normalizeTagSortBy()`

### 6.2 Bereits vorhandene Funktionalitäten in unserer CLI

#### 6.2.1 Grundlegende Konfigurationsstruktur
**Vorhanden in:** `internal/config/defaults.go` und `internal/config/config.go`
**Status:** ✅ Grundstruktur vorhanden, aber unvollständig
**Fehlende Teile:**
- Detaillierte Commit-Filter-Optionen
- Header-Pattern-Maps
- Issue/Ref-Aktionen
- Merge/Revert-Pattern
- JIRA-Konfiguration (nur teilweise vorhanden)

#### 6.2.2 UI-Komponenten
**Vorhanden in:** `internal/ui/bubble/`
**Status:** ✅ Grundkomponenten vorhanden
**Verfügbare Komponenten:**
- `select.go` - Einfache Auswahl
- `input.go` - Texteingabe
- `confirm.go` - Ja/Nein-Bestätigung
- `progress.go` - Fortschrittsanzeige

**Fehlende Komponenten:**
- Multiselect-Funktionalität (für Commit-Typen-Auswahl)

#### 6.2.3 Changelog-Service
**Vorhanden in:** `internal/services/changelog/`
**Status:** ✅ Grundfunktionalität vorhanden
**Verfügbare Dateien:**
- `chglog.go` - Hauptgenerator
- `config.go` - Konfigurationslogik
- `processor.go` - Commit-Verarbeitung
- `jira.go` - JIRA-Integration

**Fehlende Funktionalitäten:**
- Template-basierte Generierung
- Verschiedene Ausgabeformate
- Erweiterte Filteroptionen

### 6.3 Neue Dateien, die erstellt werden müssen

#### 6.3.1 Template-Builder-System
```
internal/ui/cmd/initialize/template_builders/
├── builder.go                    # Interface-Definition
├── kac_template_builder.go      # Keep-A-Changelog Builder (aus git-chglog kopiert)
├── custom_template_builder.go   # Custom Template Builder (aus git-chglog kopiert)
├── standard_template_builder.go # Standard Template Builder (neu)
└── factory.go                   # Factory für Template-Builder
```

#### 6.3.2 Commit-Format-Definitionen
```
internal/ui/cmd/initialize/
├── commit_formats.go            # Commit-Format-Definitionen (aus git-chglog übernommen)
├── template_styles.go           # Template-Stil-Definitionen (aus git-chglog übernommen)
└── multiselect.go              # Multiselect-UI-Komponente (neu)
```

#### 6.3.3 Erweiterte Konfiguration
```
internal/config/
├── changelog_types.go           # Erweiterte Changelog-Typen (aus git-chglog übernommen)
├── changelog_normalization.go  # Normalisierungslogik (aus git-chglog übernommen)
└── changelog_validation.go     # Validierungslogik (neu)
```

### 6.4 Konkrete Kopier-Befehle

#### 6.4.1 Verzeichnisse erstellen
```bash
mkdir -p internal/ui/cmd/initialize/template_builders
mkdir -p internal/config/changelog
```

#### 6.4.2 Dateien kopieren und anpassen
```bash
# Template-Builder kopieren
cp example/git-chglog-master/cmd/git-chglog/kac_template_builder.go internal/ui/cmd/initialize/template_builders/
cp example/git-chglog-master/cmd/git-chglog/custom_template_builder.go internal/ui/cmd/initialize/template_builders/

# Config-Builder als Referenz kopieren
cp example/git-chglog-master/cmd/git-chglog/config_builder.go internal/ui/cmd/initialize/template_builders/

# Variablen-Definitionen als Referenz kopieren
cp example/git-chglog-master/cmd/git-chglog/variables.go internal/ui/cmd/initialize/commit_formats_reference.go
```

#### 6.4.3 Anpassungsschritte nach dem Kopieren

**Für alle kopierten Dateien:**
1. Package-Namen von `main` zu entsprechendem Package ändern
2. Import-Pfade anpassen
3. Externe Abhängigkeiten durch interne ersetzen

**Spezifische Anpassungen:**

**kac_template_builder.go:**
```go
// Ändern von:
package main

// Zu:
package template_builders

// Import hinzufügen:
import (
    "fmt"
    "clikd/internal/ui/cmd/initialize/types"
)

// Answer-Typ durch internen Typ ersetzen:
func (t *kacTemplateBuilderImpl) Build(ans *types.InitializationAnswer) (string, error) {
    // ...
}
```

**custom_template_builder.go:**
```go
// Ähnliche Anpassungen wie bei kac_template_builder.go
```

**variables.go → commit_formats.go:**
```go
// Ändern von:
package main

// Zu:
package initialize

// Strukturen beibehalten, aber in unsere Architektur integrieren
```

### 6.5 Bestehende Dateien, die erweitert werden müssen

#### 6.5.1 internal/config/defaults.go
**Erweitern um:**
- Detaillierte Commit-Filter-Optionen
- Header-Pattern-Definitionen
- Issue/Ref-Aktionen
- Merge/Revert-Pattern

#### 6.5.2 internal/ui/cmd/initialize/changelog.go
**Erweitern um:**
- Commit-Format-Auswahl
- Template-Stil-Auswahl
- Erweiterte Filter-Konfiguration
- Multiselect für Commit-Typen

#### 6.5.3 internal/ui/bubble/select.go
**Erweitern um:**
- Multiselect-Funktionalität
- Checkbox-Unterstützung

### 6.6 Prioritätenliste für die Implementierung

#### Phase 1: Grundlagen (1-2 Tage)
1. **Dateien kopieren und anpassen:**
   - Template-Builder aus git-chglog übernehmen
   - Commit-Format-Definitionen übernehmen
   - Konfigurationsstrukturen erweitern

#### Phase 2: UI-Erweiterungen (2-3 Tage)
2. **Multiselect-Komponente implementieren**
3. **Erweiterte Changelog-Konfiguration in der Init-UI**
4. **Template-Builder-Integration**

#### Phase 3: Service-Integration (2-3 Tage)
5. **Changelog-Service erweitern**
6. **Template-basierte Generierung implementieren**
7. **Erweiterte Filteroptionen**

#### Phase 4: Tests und Finalisierung (2 Tage)
8. **Tests für alle neuen Funktionen**
9. **Dokumentation aktualisieren**
10. **End-to-End-Tests**

## 7. Fazit

Durch die systematische Übernahme und Anpassung der bewährten Komponenten aus git-chglog können wir die Entwicklungszeit erheblich verkürzen. Etwa 60-70% der benötigten Funktionalität kann direkt aus dem Beispiel übernommen und an unsere Architektur angepasst werden. Die verbleibenden 30-40% bestehen hauptsächlich aus UI-Integrationen und Anpassungen an unser bestehendes Konfigurationssystem. 
