package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"clikd/pkg/config"
	"clikd/pkg/ui"
	"clikd/pkg/ui/components"
)

// runInitWithBubbleTea führt die Initialisierung mit der Bubble Tea UI durch
func runInitWithBubbleTea(global, force bool) error {
	var configPath string
	var configDir string
	var isInGitRepo bool
	var repoURL string

	// Logo anzeigen
	fmt.Println(ui.RenderLogo())
	fmt.Println(ui.H1.Render("Willkommen beim clikd-Konfigurations-Assistenten!"))
	fmt.Println(ui.NormalText.Render("Dieser Assistent hilft Ihnen bei der Einrichtung von clikd für Ihr Projekt."))
	fmt.Println()

	// Prüfe, ob wir uns in einem Git-Repository befinden
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err == nil {
		isInGitRepo = true

		// Versuche, die Repository-URL zu ermitteln
		remoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
		if remoteOutput, err := remoteCmd.Output(); err == nil {
			repoURL = strings.TrimSpace(string(remoteOutput))
		}
	}

	// Wenn wir in einem Git-Repository sind und nicht im globalen Modus sind,
	// frage den Benutzer, ob er eine lokale oder globale Konfiguration möchte
	if isInGitRepo && !global {
		fmt.Println(ui.InfoText("Git-Repository erkannt: " + repoURL))

		result := components.Confirm(
			"Konfigurationstyp wählen",
			"Möchten Sie eine lokale Konfiguration für dieses Repository erstellen?",
		)

		if !result {
			fmt.Println(ui.InfoText("Erstelle stattdessen eine globale Konfiguration..."))
			global = true
		} else {
			fmt.Println(ui.InfoText("Erstelle lokale Konfiguration für dieses Repository..."))
		}
	}

	// Bestimme Konfigurationspfad
	if global {
		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(ui.ErrorText("Fehler beim Ermitteln des Home-Verzeichnisses: " + err.Error()))
			return fmt.Errorf("error finding home directory: %w", err)
		}

		// Create global config directory
		configDir = filepath.Join(home, ".clikd")
	} else {
		// Lokaler Konfigurationsordner
		configDir = "clikd"
	}

	// Prüfen, ob der Pfad existiert und ob es sich um einen Ordner handelt
	info, err := os.Stat(configDir)
	if err == nil {
		// Pfad existiert, prüfen ob es ein Ordner ist
		if !info.IsDir() {
			fmt.Println(ui.ErrorText("Fehler beim Erstellen des Konfigurationsverzeichnisses: " + configDir + " existiert bereits, ist aber kein Verzeichnis"))
			return fmt.Errorf("error creating config directory: %s already exists but is not a directory", configDir)
		}
	} else if os.IsNotExist(err) {
		// Pfad existiert nicht, erstellen
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println(ui.ErrorText("Fehler beim Erstellen des Konfigurationsverzeichnisses: " + err.Error()))
			return fmt.Errorf("error creating config directory: %w", err)
		}
	} else {
		// Anderer Fehler beim Zugriff
		fmt.Println(ui.ErrorText("Fehler beim Prüfen des Konfigurationsverzeichnisses: " + err.Error()))
		return fmt.Errorf("error checking config directory: %w", err)
	}

	configPath = filepath.Join(configDir, "config.toml")

	// Prüfe, ob Konfiguration bereits existiert
	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		if !force {
			fmt.Println(ui.WarningText("Konfigurationsdatei existiert bereits unter " + configPath))

			result := components.Confirm(
				"Bestehende Konfiguration",
				"Möchten Sie die bestehende Konfiguration überschreiben?",
			)

			if !result {
				fmt.Println(ui.InfoText("Abbruch, bestehende Konfiguration wird nicht überschrieben."))
				return nil
			}
			force = true
		}
	}

	// Create a new config manager
	manager := config.NewManager()

	// Lade bestehende Konfiguration, wenn sie existiert und nicht überschrieben werden soll
	if configExists && !force {
		if err := manager.InitConfig(configPath); err != nil {
			fmt.Println(ui.ErrorText("Fehler beim Laden der bestehenden Konfiguration: " + err.Error()))
			return fmt.Errorf("error loading existing config: %w", err)
		}
		fmt.Println(ui.SuccessText("Bestehende Konfiguration geladen."))
	} else {
		// Initialize with default config
		if err := manager.InitConfig(""); err != nil {
			fmt.Println(ui.ErrorText("Fehler beim Initialisieren der Konfiguration: " + err.Error()))
			return fmt.Errorf("error initializing config: %w", err)
		}
	}

	// AI-Konfiguration
	fmt.Println(ui.SectionTitle("KI-Konfiguration"))

	result := components.Confirm(
		"KI-Funktionen",
		"Möchten Sie KI-Funktionen aktivieren?",
	)

	aiEnabled := result
	manager.SetConfigValue("ai.enable", fmt.Sprintf("%t", aiEnabled))

	if aiEnabled {
		// Provider auswählen
		providerOptions := config.SupportedProviders
		providerItems := make([]components.SelectItem, len(providerOptions))

		for i, provider := range providerOptions {
			defaultModel, _ := config.GetDefaultModelForProvider(provider)
			providerItems[i] = components.SelectItem{
				Title:       provider,
				Description: fmt.Sprintf("Standardmodell: %s", defaultModel),
				Value:       provider,
			}
		}

		selectedProvider := components.RunSelect("Wählen Sie einen Provider", providerItems)
		if selectedProvider == nil {
			return fmt.Errorf("abbruch durch Benutzer")
		}

		provider := selectedProvider.Value.(string)
		manager.SetConfigValue("ai.provider", provider)

		// Modelle für den ausgewählten Provider anzeigen
		supportedModels, _ := config.GetSupportedModelsForProvider(provider)

		modelItems := make([]components.SelectItem, len(supportedModels))

		for i, model := range supportedModels {
			modelItems[i] = components.SelectItem{
				Title:       model,
				Description: fmt.Sprintf("Modell für %s", provider),
				Value:       model,
			}
		}

		selectedModel := components.RunSelect(fmt.Sprintf("Wählen Sie ein Modell für %s", provider), modelItems)
		if selectedModel == nil {
			return fmt.Errorf("abbruch durch Benutzer")
		}

		model := selectedModel.Value.(string)
		manager.SetConfigValue("ai.model", model)

		// API-Schlüssel-Hinweis
		apiKeyInfo := ""
		switch provider {
		case "openai":
			apiKeyInfo = "https://platform.openai.com/api-keys"
		case "mistral":
			apiKeyInfo = "https://console.mistral.ai/api-keys/"
		case "anthropic":
			apiKeyInfo = "https://console.anthropic.com/settings/keys"
		default:
			apiKeyInfo = "der Website des Anbieters"
		}

		fmt.Println(ui.SectionTitle("API-Schlüssel-Konfiguration"))
		if global {
			fmt.Println(ui.InfoText("Für globale KI-Konfiguration können Sie den API-Schlüssel wie folgt setzen:"))
			fmt.Println(ui.HighlightText(fmt.Sprintf("  clikd init config set ai.api_key=IHR_API_SCHLÜSSEL")))
		} else {
			fmt.Println(ui.InfoText("Für lokale Projekte erstellen Sie eine .env-Datei im Projektverzeichnis mit:"))
			fmt.Println(ui.HighlightText(fmt.Sprintf("  CLIKD_%s_API_KEY=IHR_API_SCHLÜSSEL", strings.ToUpper(provider))))
		}
		fmt.Println(ui.InfoText("API-Schlüssel erhalten Sie auf: " + apiKeyInfo))
	}

	// Changelog-Konfiguration
	fmt.Println(ui.SectionTitle("Changelog-Konfiguration"))

	setupChangelog := components.Confirm(
		"Changelog-Funktionen",
		"Möchten Sie Changelog-Funktionen konfigurieren?",
	)

	if setupChangelog {
		// Stil auswählen
		styleOptions := []components.SelectItem{
			{Title: "github", Description: "GitHub-Style mit Markdown", Value: "github"},
			{Title: "gitlab", Description: "GitLab-Style mit Markdown", Value: "gitlab"},
			{Title: "bitbucket", Description: "Bitbucket-Style mit Markdown", Value: "bitbucket"},
		}

		selectedStyle := components.RunSelect("Wählen Sie einen Changelog-Stil", styleOptions)
		if selectedStyle == nil {
			return fmt.Errorf("abbruch durch Benutzer")
		}

		style := selectedStyle.Value.(string)
		manager.SetConfigValue("changelog.style", style)

		// Repository-URL, wenn in einem Git-Repository
		if isInGitRepo && repoURL != "" {
			manager.SetConfigValue("changelog.repository_url", repoURL)
			fmt.Println(ui.SuccessText("Repository-URL auf " + repoURL + " gesetzt."))
		} else if isInGitRepo {
			repoURLInput := components.RunInput(
				"Repository-URL",
				"Geben Sie die URL Ihres Git-Repositories ein:",
				"https://github.com/username/repo",
			)

			if repoURLInput != "" {
				manager.SetConfigValue("changelog.repository_url", repoURLInput)
				fmt.Println(ui.SuccessText("Repository-URL auf " + repoURLInput + " gesetzt."))
			}
		}

		// Erweiterte Changelog-Einstellungen mit Fortschrittsanzeige
		fmt.Println(ui.InfoText("Konfiguriere erweiterte Changelog-Einstellungen..."))

		components.RunProgress(
			"Changelog-Konfiguration",
			"Bitte warten, während die erweiterten Changelog-Einstellungen konfiguriert werden...",
			func(setPercent func(float64), setDone func()) {
				// HeaderPattern und PatternMaps für die Commit-Nachrichtenanalyse
				manager.SetConfigValue("changelog.options.header.pattern", "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$")
				setPercent(0.2)
				time.Sleep(100 * time.Millisecond)

				// PatternMaps als Array setzen
				patternMaps := []string{"Type", "Scope", "Subject"}
				for i, pattern := range patternMaps {
					manager.SetConfigValue(fmt.Sprintf("changelog.options.header.pattern_maps.%d", i), pattern)
				}
				setPercent(0.4)
				time.Sleep(100 * time.Millisecond)

				// Commit-Filter konfigurieren - als Array für jeden Typ
				commitTypes := []string{"feat", "fix", "perf", "refactor", "chore"}
				for i, commitType := range commitTypes {
					manager.SetConfigValue(fmt.Sprintf("changelog.options.commits.filters.Type.%d", i), commitType)
				}
				setPercent(0.6)
				time.Sleep(100 * time.Millisecond)

				// Gruppierung der Commits
				manager.SetConfigValue("changelog.options.commit_groups.group_by", "Type")
				manager.SetConfigValue("changelog.options.commit_groups.sort_by", "Title")

				// Titel-Mappings für Commit-Gruppen
				manager.SetConfigValue("changelog.options.commit_groups.title_maps.feat", "Features")
				manager.SetConfigValue("changelog.options.commit_groups.title_maps.fix", "Bug Fixes")
				manager.SetConfigValue("changelog.options.commit_groups.title_maps.perf", "Performance Improvements")
				manager.SetConfigValue("changelog.options.commit_groups.title_maps.refactor", "Code Refactoring")
				manager.SetConfigValue("changelog.options.commit_groups.title_maps.chore", "Chores")
				setPercent(0.8)
				time.Sleep(100 * time.Millisecond)

				// Note Keywords für Breaking Changes usw.
				noteKeywords := []string{"BREAKING CHANGE", "SECURITY"}
				for i, keyword := range noteKeywords {
					manager.SetConfigValue(fmt.Sprintf("changelog.options.notes.keywords.%d", i), keyword)
				}
				setPercent(1.0)
				time.Sleep(100 * time.Millisecond)
				setDone()
			},
		)
	}

	// Verzeichnisse erstellen mit Fortschrittsanzeige
	components.RunProgressWithValues(
		"Projektstruktur wird erstellt",
		"Erstelle Verzeichnisse und Templates...",
		3, "Schritte",
		func(setValue func(int), setDone func()) {
			// Create templates directory
			var templatesDir string
			if global {
				home, _ := os.UserHomeDir()
				templatesDir = filepath.Join(home, ".clikd", "templates")
			} else {
				templatesDir = filepath.Join("clikd", "templates")
			}

			if err := os.MkdirAll(templatesDir, 0755); err != nil {
				fmt.Println(ui.ErrorText("Fehler beim Erstellen des Templates-Verzeichnisses: " + err.Error()))
				return
			}
			setValue(1)
			time.Sleep(200 * time.Millisecond)

			// Create cache directory
			var cacheDir string
			if global {
				home, _ := os.UserHomeDir()
				cacheDir = filepath.Join(home, ".clikd", "cache")
			} else {
				cacheDir = filepath.Join("clikd", "cache")
			}

			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				fmt.Println(ui.ErrorText("Fehler beim Erstellen des Cache-Verzeichnisses: " + err.Error()))
				return
			}
			setValue(2)
			time.Sleep(200 * time.Millisecond)

			// Create default changelog template
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
				fmt.Println(ui.ErrorText("Fehler beim Erstellen des Changelog-Templates: " + err.Error()))
				return
			}

			// Update template path in config
			if global {
				manager.SetConfigValue("changelog.template", filepath.Join("templates", "changelog.md"))
			} else {
				manager.SetConfigValue("changelog.template", filepath.Join("templates", "changelog.md"))
			}
			setValue(3)
			time.Sleep(200 * time.Millisecond)
			setDone()
		},
	)

	// Save the configuration
	if err := manager.SaveConfig(configPath); err != nil {
		fmt.Println(ui.ErrorText("Fehler beim Speichern der Konfiguration: " + err.Error()))
		return fmt.Errorf("error saving config: %w", err)
	}

	// Prüfen, ob KI aktiviert ist (für "Nächste Schritte")
	cfg := manager.GetConfig()
	aiEnabled = cfg.AI.Enable

	// Zusammenfassung anzeigen
	items := []components.ListItem{
		components.CreateListItemWithStatus("Konfiguration", configPath, "done"),
		components.CreateListItemWithStatus("Verzeichnisse", "templates/, cache/", "done"),
	}

	if aiEnabled {
		items = append(items, components.CreateListItemWithStatus("KI-Funktionen", "Aktiviert", "done"))

		provider := manager.GetConfig().AI.Provider
		model := manager.GetConfig().AI.Model

		modelItem := components.CreateListItemWithStatus("KI-Modell", fmt.Sprintf("%s (%s)", model, provider), "done")
		modelItem.Tags = []string{provider, model}
		items = append(items, modelItem)

		apiKeyItem := components.CreateListItemWithStatus("API-Schlüssel", "Noch nicht konfiguriert", "pending")
		apiKeyItem.Tags = []string{"required"}
		items = append(items, apiKeyItem)
	} else {
		items = append(items, components.CreateListItemWithStatus("KI-Funktionen", "Deaktiviert", "cancelled"))
	}

	// Nächste Schritte
	nextStepsItems := []components.ListItem{}

	if aiEnabled {
		apiKeyItem := components.CreateListItemWithStatus("API-Schlüssel konfigurieren", "", "pending")
		if global {
			apiKeyItem.Description = "clikd init config set ai.api_key=IHR_API_SCHLÜSSEL"
		} else {
			apiKeyItem.Description = fmt.Sprintf("Erstellen Sie eine .env-Datei mit CLIKD_%s_API_KEY=IHR_API_SCHLÜSSEL", strings.ToUpper(manager.GetConfig().AI.Provider))
		}
		nextStepsItems = append(nextStepsItems, apiKeyItem)
	}

	changelogItem := components.CreateListItemWithStatus("Changelog generieren", "clikd changelog -o CHANGELOG.md", "pending")
	nextStepsItems = append(nextStepsItems, changelogItem)

	// Zusammenfassung und Nächste Schritte anzeigen
	fmt.Println(ui.H1.Render("Konfiguration abgeschlossen"))
	components.ShowFormattedList(
		"Erstellte Komponenten",
		"Folgende Komponenten wurden erfolgreich konfiguriert:",
		items,
		true, // Tags anzeigen
		true, // Status anzeigen
		10,   // Einträge pro Seite
	)

	components.ShowFormattedList(
		"Nächste Schritte",
		"Folgende Schritte können Sie als nächstes ausführen:",
		nextStepsItems,
		false, // Tags ausblenden
		true,  // Status anzeigen
		10,    // Einträge pro Seite
	)

	fmt.Println(ui.SuccessText("clikd ist jetzt einsatzbereit! Viel Spaß beim Verwenden!"))

	return nil
}
