# Aktualisierter Implementierungsplan für Changelog-Funktionalität in clikd

Basierend auf der Analyse des git-chglog-master Repositories und dem Verständnis, wie die Komponenten dort interagieren, erstellen wir folgenden überarbeiteten Plan für die Integration der Changelog-Funktionalität in clikd.

## 1. Ordnerstruktur und Dateien

Die aktualisierte Struktur orientiert sich stärker am Original-Repository:

```
internal/
└── services/
    └── changelog/
        ├── template_builders/  # Enthält die Template-Builder für die Generierung von Templates
        │   ├── builder.go                 # Definition der Builder-Schnittstelle
        │   ├── template_builder.go        # Definition der Template-Builder-Schnittstelle und Factory
        │   ├── custom.go                  # Builder für Standard/Cool-Templates
        │   ├── kac.go                     # Builder für Keep-a-Changelog-Templates
        │   └── template_builder_mock.go   # Mock für Tests
        ├── variables.go        # Enthält Definitionen für Formate, Stile und Konstanten (ersetzt formats.go)
        ├── types.go            # Enthält grundlegende Typdefinitionen
        ├── config_builder.go   # Enthält die Logik zur Generierung der Konfigurationsdateien
        ├── service.go          # Enthält die Hauptfunktionalität des Changelog-Services
        ├── previews.go         # Enthält Beispiel-Changelogs für die Vorschau mit Glamour
        └── chglog.go           # Enthält die eigentliche Changelog-Generierungslogik
```

**Wichtiger Hinweis:** Im Gegensatz zur ursprünglichen Planung werden die Templates und Konfigurationen **nicht** als eingebettete Dateien verwendet. Stattdessen werden sie, wie im Original, direkt im Code als Strings generiert. Die Verzeichnisse `templates/` und `configs/` werden daher nicht benötigt.

### 2 Variablen und Konstanten in `variables.go`
siehe `internal/services/changelog/variables.go`

## 3. Integration und Verwendung der Template-Builders

Die Template-Builders sind ein zentraler Teil der Changelog-Funktionalität. Sie wandeln die Benutzereingaben (in Form des Answer-Objekts) in ein vollständiges Go-Template um, das später zur Generierung des Changelogs verwendet wird.

### 3.2 Verwendung der Template-Builders im Initialisierungsprozess

Der Initialisierungsprozess nutzt die Template-Builders, um aus den Benutzerantworten ein vollständiges Template zu generieren:

1. Der Benutzer wählt einen Stil (GitHub, GitLab, etc.)
2. Der Benutzer wählt ein Template-Format (Standard, Cool, Keep-a-Changelog)
3. Die Antworten werden in einem Answer-Objekt gespeichert
4. Der passende Template-Builder wird über `GetTemplateBuilder()` abgerufen
5. Die `Build()`-Methode des Builders generiert das vollständige Template
6. Das generierte Template wird in einer Datei gespeichert

## 4. Integration in die Bubble Tea UI-Komponenten

Die bestehenden Bubble Tea-Komponenten im `internal/ui/bubble`-Verzeichnis werden für die interaktive Konfiguration des Changelogs verwendet. Wir nutzen die folgenden Komponenten:

### 4.1 Erweiterung der Initialisierungs-UI

```go
// internal/ui/cmd/initialize/update.go (erweitert)

// Hinzufügen neuer Schritte für Changelog-Konfiguration
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.CurrentStep {
	// ... bestehende Schritte ...
	
	case StepChangelogConfig:
		// Abfrage, ob Changelog konfiguriert werden soll
		switch msg := msg.(type) {
		case bubble.ConfirmResultMsg:
			if msg.Result {
				m.CurrentStep = StepChangelogStyle
				m.confirmModel = nil
				
				// Template-Stil-Auswahl vorbereiten
				items := []bubble.SelectItem{
					{Title: "GitHub", Description: "Standard GitHub Format", Value: "github"},
					{Title: "GitLab", Description: "Standard GitLab Format", Value: "gitlab"},
					{Title: "Bitbucket", Description: "Standard Bitbucket Format", Value: "bitbucket"},
					{Title: "None", Description: "Einfaches Format ohne spezielle Links", Value: "none"},
				}
				m.selectModel = &bubble.SelectModel{
					Title:       "Wählen Sie einen Stil für den Changelog",
					Items:       items,
					Description: true,
				}
				return m, nil
			}
			m.CurrentStep = StepProjectStructure
			return m, nil
		}
		
		// Initialisiere Confirm-Modell, wenn es noch nicht existiert
		if m.confirmModel == nil {
			confirm := bubble.NewConfirmModel(
				"Changelog-Konfiguration",
				"Möchten Sie einen Changelog für Ihr Projekt konfigurieren?",
			)
			m.confirmModel = &confirm
		}
		
		// Update des Confirm-Modells
		newConfirm, cmd := m.confirmModel.Update(msg)
		confirm, ok := newConfirm.(bubble.ConfirmModel)
		if ok {
			m.confirmModel = &confirm
		}
		return m, cmd
		
	case StepChangelogStyle:
		// Auswahl des Template-Stils
		switch msg := msg.(type) {
		case bubble.SelectResultMsg:
			m.Manager.Config.Changelog.Style = msg.Value.(string)
			m.CurrentStep = StepChangelogJIRA
			m.selectModel = nil
			
			// JIRA-Integration-Abfrage vorbereiten
			confirm := bubble.NewConfirmModel(
				"JIRA-Integration",
				"Möchten Sie JIRA-Integration für den Changelog aktivieren?",
			)
			m.confirmModel = &confirm
			return m, nil
		}
		
		// Update des Select-Modells
		if m.selectModel != nil {
			newSelect, cmd := m.selectModel.Update(msg)
			selectModel, ok := newSelect.(bubble.SelectModel)
			if ok {
				m.selectModel = &selectModel
			}
			return m, cmd
		}
		
	// ... weitere Schritte ...
	}
	
	// ... Rest der Update-Funktion ...
}
```

### 4.2 Vorschau-Rendering mit Glamour

```go
// internal/ui/cmd/initialize/preview.go
package initialize

import (
	"clikd/internal/services/changelog/previews"

	tea "github.com/charmbracelet/bubbletea"
)

// TemplatePreviewMsg enthält die gerenderte Vorschau
type TemplatePreviewMsg struct {
	RenderedPreview string
}

// showTemplatePreview zeigt eine Vorschau des ausgewählten Templates
func showTemplatePreview(style string) tea.Cmd {
	return func() tea.Msg {
		// Hole die passende Vorschau für den Stil
		previewContent := previews.GetPreview(style)
		
		// Rendere die Vorschau mit Glamour
		rendered, err := previews.RenderPreview(previewContent)
		if err != nil {
			rendered = "Fehler beim Rendern der Vorschau: " + err.Error()
		}
		
		return TemplatePreviewMsg{RenderedPreview: rendered}
	}
}
```

## 5. Anpassung der Konfigurationsstruktur

Die Konfigurationsstruktur in `internal/config/config.go` enthält bereits die notwendigen Felder für die Changelog-Konfiguration. nur müssen wir diese anpassen udn viel weniger einbauen da wir für changelog die yml config nutzen.

### 5.1 aktuelle config.toml Struktur

```toml
# Allgemeine Konfiguration
[general]
log_level = "info"
color = true

# Changelog-Konfiguration
[changelog]
style = "github"                                # github, gitlab, bitbucket, none
commit_format = "type-scope-subject"            # Das Format der Commit-Nachrichten
template = "clikd/changelog/templates/standard.tpl.md"
config_file = "clikd/changelog/config/standard.yml"
include_merges = true
include_reverts = true
tag_filter_pattern = ""
sort = "date"                                   # "date" oder "semver"
path = "CHANGELOG.md"                          # Ausgabepfad für den generierten Changelog
no_case = false                                # Case-insensitive Filterung

[changelog.info]
title = "Changelog"
repository_url = ""

[changelog.options.commits]
sort_by = "scope"
filters = {}

[changelog.options.commit_groups]
group_by = "type"
sort_by = "title"
title_maps = { feat = "Features", fix = "Bug Fixes", perf = "Performance Improvements", refactor = "Code Refactoring", docs = "Documentation", test = "Tests" }

[changelog.options.header]
pattern = "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
pattern_maps = ["Type", "Scope", "Subject"]

[changelog.options.notes]
keywords = ["BREAKING CHANGE", "DEPRECATED"]

[changelog.jira]
base_url = ""
username = ""
project_key = ""
issue_pattern = "([A-Z]+-\\d+)"

# KI-Konfiguration
[ai]
enable = true
provider = "mistral"
model = "mistral-medium"
api_key = ""
```

### 5.2 neue config.toml Struktur
```toml
# Allgemeine Konfiguration
[general]
log_level = "info"
color = true

# KI-Konfiguration
[ai]
enable = true
provider = 'mistral'
model = 'mistral-medium'
api_key = ''
api_url = ''
api_custom_headers = ''
tokens_max_input = 4096
tokens_max_output = 500

# Changelog-Konfiguration
[changelog]
style = "github"                                # github, gitlab, bitbucket, none
commit_format = "type-scope-subject"            # Das Format der Commit-Nachrichten
template = "clikd/changelog/templates/standard.tpl.md"
config_file = "clikd/changelog/config/standard.yml"
include_merges = true
include_reverts = true

```

## 6. Konfigurationsintegration

Die Integration zwischen der TOML-basierten Konfiguration von clikd und der YAML-basierten Konfiguration des Changelogs erfordert besondere Aufmerksamkeit. Wir implementieren einen durchdachten Ansatz für die Zusammenarbeit beider Konfigurationsformate.

### 6.1 Verzeichnisstruktur nach der Initialisierung

Nach dem Ausführen des `clikd init`-Befehls wird folgende Verzeichnisstruktur erstellt:

```
 clikd/
   ├── config.toml             # Enthält nur Verweise auf die gewählten Dateien
   └── changelog
       ├── standard.tpl.md     # Template-Datei für den Changelog
       └── standard.yml        # YAML-Konfigurationsdatei für den Changelog
```

### 6.2 Konfigurationshierarchie

Der Changelog-Befehl sucht und lädt die Konfigurationen in folgender Reihenfolge:

1. **TOML-Konfiguration**: Zunächst wird die `config.toml` geladen, in der die Pfade zu den Changelog-spezifischen Dateien definiert sind.
2. **YAML-Konfiguration**: Basierend auf dem in der TOML definierten Pfad wird die `config.yml` geladen.
3. **Template-Datei**: Ebenfalls wird der Pfad zum Template aus der TOML-Konfiguration gelesen.

Wenn nicht explizit in der TOML-Konfiguration definiert, werden Standardpfade verwendet:
wird der user darauf hingeweisen das er die config.yml und die template.tpl.md dateien in den changelog ordner legen muss. oder er den init befehl ausführen muss.

### 6.3 Konfigurationstypen

wir müssen die bestehende `ChangelogConfig`-Struktur in `internal/config/types.go`, anpassen und notwendige Felder entfernen:

```go
// ChangelogConfig enthält Changelog-bezogene Einstellungen
type ChangelogConfig struct {
	Style            string     `mapstructure:"style"`
	Template         string     `mapstructure:"template"`
	JiraIntegration  bool       `mapstructure:"jira_integration"`
	Sort             string     `mapstructure:"sort"`
	TagFilterPattern string     `mapstructure:"tag_filter_pattern"`
	Path             string     `mapstructure:"path"`
	NoCase           bool       `mapstructure:"no_case"`
	Jira             JiraConfig `mapstructure:"jira"`

	// Erweiterte Optionen aus der YAML-Konfiguration
	Info    ChangelogInfoConfig    `mapstructure:"info"`
	Options ChangelogOptionsConfig `mapstructure:"options"`
}
```

## Die config für den changelog befehl
Die config für den changelog befehl wird mit dem config_builder in `internal/config/changelog` erstellt.

#### 6.4 Preview-Befehl
der changelog hat einen preview befehl diesen werden wir auch mit charmbracelet/glamour rendern udn anzeigen.

```go
// pkg/commands/changelog/cmd.go
func newPreviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview [tag-query]",
		Short: "Zeige eine Vorschau des Changelogs im Terminal",
		Long: `Zeige eine Vorschau des Changelogs im Terminal an.
Optionales Argument <tag-query> spezifiziert den Bereich der Tags (siehe 'generate').`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Service mit Konfiguration erstellen
			service, err := changelog.NewChangelogService(config.NewLoader())
			if err != nil {
				return err
			}
			
			// Tag-Query extrahieren
			var query string
			if len(args) > 0 {
				query = args[0]
			}
			
			// Changelog generieren und im Terminal anzeigen
			content, err := service.GenerateToString(query)
			if err != nil {
				return err
			}
			
			// Ausgabe
			fmt.Println(content)
			return nil
		},
	}
	
	// Flags registrieren (ähnlich wie bei generate)
	
	return cmd
}
```
