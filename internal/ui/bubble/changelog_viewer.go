package bubble

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"clikd/internal/services/changelog"
	"clikd/internal/ui/styles"
	"clikd/internal/utils"
)

// ChangelogViewerModel is a model for displaying changelog with copy functionality
type ChangelogViewerModel struct {
	Title           string
	Content         string
	RenderedContent string
	Width           int
	Height          int
	Viewport        int
	Scroll          int
	MaxScroll       int
	ShowHelp        bool
	CopyMessage     string
	CopyMessageTime int

	// Generator-spezifische Felder
	IsGenerating    bool
	GeneratorConfig *changelog.CommandConfig
	Query           string
	Progress        progress.Model
	ProgressPercent float64

	// Save-to-File Felder
	ShowingSaveInput bool
	SaveInput        textinput.Model
	SaveMessage      string
	SaveMessageTime  int
}

// ProgressTickMsg wird für die Progress-Animation verwendet
type ProgressTickMsg time.Time

// GenerateCompleteMsg wird gesendet, wenn die Generierung abgeschlossen ist
type GenerateCompleteMsg struct {
	Content string
	Error   error
}

// SaveCompleteMsg wird gesendet, wenn das Speichern abgeschlossen ist
type SaveCompleteMsg struct {
	Filename string
	Error    error
}

// NewChangelogViewerModel creates a new changelog viewer model
func NewChangelogViewerModel(title, content string) ChangelogViewerModel {
	// Use plain text to preserve original markdown structure
	// This prevents glamour from moving reference links to the top
	rendered := content

	lines := strings.Split(rendered, "\n")
	viewport := 30
	maxScroll := len(lines) - viewport
	if maxScroll < 0 {
		maxScroll = 0
	}

	return ChangelogViewerModel{
		Title:           title,
		Content:         content,
		RenderedContent: rendered,
		Viewport:        viewport,
		MaxScroll:       maxScroll,
		ShowHelp:        false,
	}
}

// NewChangelogViewerModelWithGenerator creates a new changelog viewer model that generates content
func NewChangelogViewerModelWithGenerator(title string, config *changelog.CommandConfig, query string) ChangelogViewerModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(60),
	)

	return ChangelogViewerModel{
		Title:           title,
		IsGenerating:    true,
		GeneratorConfig: config,
		Query:           query,
		Progress:        p,
		ProgressPercent: 0.0,
		Viewport:        30,
		ShowHelp:        false,
	}
}

// initSaveInput initialisiert das Save-Input-Feld
func (m *ChangelogViewerModel) initSaveInput() {
	ti := textinput.New()
	ti.Placeholder = "CHANGELOG.md"
	ti.CharLimit = 200
	ti.Width = 60
	ti.Prompt = "> "
	ti.TextStyle = styles.Normal
	ti.PromptStyle = styles.InputPrompt
	ti.Focus()
	m.SaveInput = ti
}

// Init initializes the model
func (m ChangelogViewerModel) Init() tea.Cmd {
	if m.IsGenerating {
		return tea.Batch(
			m.generateChangelog,
			m.tickProgress(),
		)
	}
	return nil
}

// generateChangelog führt die Changelog-Generierung als Command aus
func (m ChangelogViewerModel) generateChangelog() tea.Msg {
	// Schritt 1: Konfiguration laden
	configPath := utils.ResolveConfigPath(m.GeneratorConfig.ConfigPath)
	configFileInfo, err := os.Stat(configPath)
	if err != nil {
		return GenerateCompleteMsg{Error: fmt.Errorf("configuration file not found at %s: %v", configPath, err)}
	}
	if configFileInfo.IsDir() {
		configPath = filepath.Join(configPath, "config.yml")
	}

	// Config-Pfad korrigieren
	m.GeneratorConfig.ConfigPath = configPath

	logger := utils.NewLogger("error", !m.GeneratorConfig.NoColor)
	config, err := changelog.LoadConfigFromCommand(m.GeneratorConfig)
	if err != nil {
		return GenerateCompleteMsg{Error: err}
	}

	generator := changelog.NewGenerator(logger, config)
	buffer := &bytes.Buffer{}

	err = generator.Generate(buffer, m.Query)
	if err != nil {
		return GenerateCompleteMsg{Error: err}
	}

	content := buffer.String()
	return GenerateCompleteMsg{Content: content}
}

// saveToFile speichert den Changelog in eine Datei
func (m ChangelogViewerModel) saveToFile(filename string) tea.Cmd {
	return func() tea.Msg {
		// Leeren Dateinamen abfangen
		if strings.TrimSpace(filename) == "" {
			return SaveCompleteMsg{Filename: filename, Error: fmt.Errorf("filename cannot be empty")}
		}

		// Gefährliche Pfade abfangen
		if filename == "/" || filename == "/CHANGELOG.md" {
			return SaveCompleteMsg{Filename: filename, Error: fmt.Errorf("cannot write to root directory. Use a relative path like 'CHANGELOG.md' or 'docs/CHANGELOG.md'")}
		}

		// Repository-Root finden
		repoRoot, err := findRepositoryRoot()
		if err != nil {
			// Fallback zum aktuellen Arbeitsverzeichnis
			wd, wdErr := os.Getwd()
			if wdErr != nil {
				return SaveCompleteMsg{Filename: filename, Error: fmt.Errorf("failed to get working directory: %v", wdErr)}
			}
			repoRoot = wd
		}

		// Absoluten Pfad bestimmen
		var absPath string
		if filepath.IsAbs(filename) {
			// Warnung für absolute Pfade außerhalb des Home-Verzeichnisses
			homeDir, _ := os.UserHomeDir()
			if homeDir != "" && !strings.HasPrefix(filename, homeDir) {
				return SaveCompleteMsg{Filename: filename, Error: fmt.Errorf("absolute paths outside home directory not allowed. Use relative paths like 'CHANGELOG.md' or 'docs/CHANGELOG.md'")}
			}
			absPath = filename
		} else {
			// Relativer Pfad - vom Repository-Root aus
			absPath = filepath.Join(repoRoot, filename)
		}

		// Verzeichnis erstellen, falls es nicht existiert
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return SaveCompleteMsg{Filename: absPath, Error: fmt.Errorf("failed to create directory %s: %v", dir, err)}
		}

		// Datei schreiben
		err = os.WriteFile(absPath, []byte(m.Content), 0644)
		if err != nil {
			return SaveCompleteMsg{Filename: absPath, Error: fmt.Errorf("failed to write file: %v", err)}
		}

		return SaveCompleteMsg{Filename: absPath, Error: nil}
	}
}

// findRepositoryRoot findet das Repository-Root-Verzeichnis
func findRepositoryRoot() (string, error) {
	// Starte vom aktuellen Arbeitsverzeichnis
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Gehe nach oben bis wir .git finden
	dir := wd
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Wir sind am Root angekommen ohne .git zu finden
			break
		}
		dir = parent
	}

	// Kein Git-Repository gefunden, verwende aktuelles Verzeichnis
	return wd, fmt.Errorf("no git repository found")
}

// tickProgress erstellt einen Tick für die Progress-Animation
func (m ChangelogViewerModel) tickProgress() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return ProgressTickMsg(t)
	})
}

// Update updates the model
func (m ChangelogViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.IsGenerating {
			// Während der Generierung nur Ctrl+C erlauben
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}
			return m, nil
		}

		// Save-Input-Dialog aktiv
		if m.ShowingSaveInput {
			switch msg.Type {
			case tea.KeyEsc:
				// Save-Dialog abbrechen
				m.ShowingSaveInput = false
				return m, nil
			case tea.KeyEnter:
				// Speichern ausführen
				filename := strings.TrimSpace(m.SaveInput.Value())
				if filename == "" {
					// Verwende Placeholder als Default
					filename = m.SaveInput.Placeholder
				}
				m.ShowingSaveInput = false
				return m, m.saveToFile(filename)
			}

			// Input-Update weiterleiten
			var cmd tea.Cmd
			m.SaveInput, cmd = m.SaveInput.Update(msg)
			return m, cmd
		}

		// Normale Navigation
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "s":
			// Save-to-File Dialog öffnen
			m.ShowingSaveInput = true
			m.initSaveInput()
			return m, textinput.Blink
		case "c":
			// Copy to clipboard
			err := clipboard.WriteAll(m.Content)
			if err == nil {
				m.CopyMessage = "✅ Copied to clipboard!"
			} else {
				m.CopyMessage = "❌ Failed to copy"
			}
			m.CopyMessageTime = 3
			return m, nil
		case "h":
			m.ShowHelp = !m.ShowHelp
			return m, nil
		case "up", "k":
			if m.Scroll > 0 {
				m.Scroll--
			}
			return m, nil
		case "down", "j":
			if m.Scroll < m.MaxScroll {
				m.Scroll++
			}
			return m, nil
		case "pgup":
			m.Scroll -= 10
			if m.Scroll < 0 {
				m.Scroll = 0
			}
			return m, nil
		case "pgdown":
			m.Scroll += 10
			if m.Scroll > m.MaxScroll {
				m.Scroll = m.MaxScroll
			}
			return m, nil
		case "home":
			m.Scroll = 0
			return m, nil
		case "end":
			m.Scroll = m.MaxScroll
			return m, nil
		}

	case SaveCompleteMsg:
		// Speichern abgeschlossen
		if msg.Error != nil {
			m.SaveMessage = fmt.Sprintf("❌ Failed to save: %v", msg.Error)
			m.SaveMessageTime = 8 // Fehler länger anzeigen
			return m, nil
		} else {
			// Zeige relativen Pfad vom Repository-Root wenn möglich
			displayPath := msg.Filename
			if repoRoot, err := findRepositoryRoot(); err == nil {
				if relPath, err := filepath.Rel(repoRoot, msg.Filename); err == nil && !strings.HasPrefix(relPath, "..") {
					displayPath = relPath
				}
			}
			m.SaveMessage = fmt.Sprintf("✅ Saved to %s", displayPath)
			m.SaveMessageTime = 3 // Kurz anzeigen vor dem Schließen

			// Auto-Close nach erfolgreichem Speichern
			return m, tea.Batch(
				tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
					return tea.Quit()
				}),
			)
		}

	case ProgressTickMsg:
		if m.IsGenerating {
			// Animiere Progress-Bar während der Generierung
			m.ProgressPercent += 0.02
			if m.ProgressPercent > 0.95 {
				m.ProgressPercent = 0.1 // Reset für kontinuierliche Animation
			}
			cmd := m.Progress.SetPercent(m.ProgressPercent)
			return m, tea.Batch(cmd, m.tickProgress())
		}
		return m, nil

	case GenerateCompleteMsg:
		if msg.Error != nil {
			// Fehler anzeigen
			m.Content = fmt.Sprintf("Error generating changelog: %v", msg.Error)
			m.RenderedContent = m.Content
		} else {
			// Erfolg: Content setzen und mit korrigiertem Glamour rendern
			m.Content = msg.Content

			// Trenne Hauptinhalt von Referenz-Links für korrektes Rendering
			mainContent, referenceLinks := separateReferenceLinks(msg.Content)

			// Rendere nur den Hauptinhalt mit Glamour für schöne Formatierung
			rendered, err := glamour.Render(mainContent, "tokyo-night")
			if err != nil {
				// Fallback: Verwende originalen Content ohne Glamour
				m.RenderedContent = msg.Content
			} else {
				// Füge die Referenz-Links am Ende hinzu (ohne Glamour-Rendering)
				if referenceLinks != "" {
					m.RenderedContent = rendered + "\n" + referenceLinks
				} else {
					m.RenderedContent = rendered
				}
			}

			lines := strings.Split(m.RenderedContent, "\n")
			m.MaxScroll = len(lines) - m.Viewport
			if m.MaxScroll < 0 {
				m.MaxScroll = 0
			}
		}

		m.IsGenerating = false
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	// Message countdown
	if m.CopyMessageTime > 0 {
		m.CopyMessageTime--
	}
	if m.SaveMessageTime > 0 {
		m.SaveMessageTime--
	}

	return m, nil
}

// View renders the model
func (m ChangelogViewerModel) View() string {
	if m.IsGenerating {
		// Progress-Anzeige während der Generierung mit besserem Layout
		var s strings.Builder

		// Padding von links
		padding := "  "

		if m.Title != "" {
			s.WriteString(padding + styles.H2.Render(m.Title) + "\n\n")
		}

		s.WriteString(padding + styles.Normal.Render("Analyzing commits and generating changelog...") + "\n\n")
		s.WriteString(padding + m.Progress.View() + "\n\n")
		s.WriteString(padding + styles.Subtle.Render("This may take a moment for large repositories"))
		s.WriteString("\n\n" + padding + styles.Subtle.Render("Ctrl+C: Cancel"))

		return s.String()
	}

	// Save-Input-Dialog anzeigen
	if m.ShowingSaveInput {
		padding := "  "
		var s strings.Builder

		s.WriteString(padding + styles.H2.Render("Save Changelog") + "\n\n")

		// Repository-Root anzeigen
		repoRoot, err := findRepositoryRoot()
		if err != nil {
			wd, _ := os.Getwd()
			repoRoot = wd
		}

		s.WriteString(padding + styles.Normal.Render("Enter filename to save the changelog:") + "\n")
		s.WriteString(padding + styles.Subtle.Render(fmt.Sprintf("Repository root: %s", repoRoot)) + "\n")
		s.WriteString(padding + styles.Subtle.Render("Relative paths will be saved from repository root") + "\n\n")

		s.WriteString(padding + styles.Subtle.Render("Examples:") + "\n")
		s.WriteString(padding + styles.Subtle.Render("  CHANGELOG.md           → Save in repository root") + "\n")
		s.WriteString(padding + styles.Subtle.Render("  docs/CHANGELOG.md      → Save in docs/ subdirectory") + "\n")
		s.WriteString(padding + styles.Subtle.Render("  .github/CHANGELOG.md   → Save in .github/ subdirectory") + "\n\n")

		s.WriteString(padding + m.SaveInput.View() + "\n\n")
		s.WriteString(padding + styles.Subtle.Render("Enter: Save • Esc: Cancel"))

		return s.String()
	}

	// Normale Changelog-Anzeige mit besserem Layout
	lines := strings.Split(m.RenderedContent, "\n")

	// Calculate visible lines
	start := m.Scroll
	end := start + m.Viewport
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]

	// Padding zu jeder Zeile hinzufügen
	padding := "  "
	paddedLines := make([]string, len(visibleLines))
	for i, line := range visibleLines {
		paddedLines[i] = padding + line
	}

	content := strings.Join(paddedLines, "\n")

	// Status bar mit Padding
	statusBar := fmt.Sprintf("Lines %d-%d of %d", start+1, end, len(lines))
	if m.MaxScroll > 0 {
		percentage := float64(m.Scroll) / float64(m.MaxScroll) * 100
		statusBar += fmt.Sprintf(" (%.0f%%)", percentage)
	}

	// Help text mit Padding - erweitert um Save-Funktion
	helpText := ""
	if m.ShowHelp {
		helpText = "\n" + padding + styles.Subtle.Render("↑/↓: Scroll • PgUp/PgDn: Page • Home/End: Jump • s: Save • c: Copy • h: Help • q/Esc: Quit")
	} else {
		helpText = "\n" + padding + styles.Subtle.Render("h: Help • s: Save • c: Copy • q: Quit")
	}

	// Messages mit Padding
	messages := ""
	if m.SaveMessageTime > 0 && m.SaveMessage != "" {
		messages += "\n\n" + padding + styles.SuccessStyle.Render(m.SaveMessage)
	}
	if m.CopyMessageTime > 0 && m.CopyMessage != "" {
		messages += "\n" + padding + styles.SuccessStyle.Render(m.CopyMessage)
	}

	return content + "\n\n" + padding + styles.Subtle.Render(statusBar) + helpText + messages
}

// RunChangelogViewer displays a changelog with copy functionality
func RunChangelogViewer(title, content string) {
	m := NewChangelogViewerModel(title, content)
	p := tea.NewProgram(m, tea.WithAltScreen())
	p.Run()
}

// RunChangelogViewerWithGenerator displays a changelog viewer that generates content
func RunChangelogViewerWithGenerator(title string, config *changelog.CommandConfig, query string) {
	m := NewChangelogViewerModelWithGenerator(title, config, query)
	p := tea.NewProgram(m, tea.WithAltScreen())
	p.Run()
}

// separateReferenceLinks separates markdown content into main content and reference links
func separateReferenceLinks(content string) (mainContent, referenceLinks string) {
	lines := strings.Split(content, "\n")
	var mainLines []string
	var refLines []string

	// Find where reference links start (lines that match [text]: URL pattern)
	inReferenceSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line is a reference link definition
		isRefLink := strings.HasPrefix(trimmed, "[") && strings.Contains(trimmed, "]:") && strings.Contains(trimmed, "http")

		// If we find a reference link, check if all remaining non-empty lines are also reference links
		if isRefLink && !inReferenceSection {
			allRemainingAreRefs := true
			for j := i; j < len(lines); j++ {
				remainingTrimmed := strings.TrimSpace(lines[j])
				if remainingTrimmed != "" {
					isRemainingRefLink := strings.HasPrefix(remainingTrimmed, "[") && strings.Contains(remainingTrimmed, "]:") && strings.Contains(remainingTrimmed, "http")
					if !isRemainingRefLink {
						allRemainingAreRefs = false
						break
					}
				}
			}

			if allRemainingAreRefs {
				inReferenceSection = true
			}
		}

		if inReferenceSection {
			refLines = append(refLines, line)
		} else {
			mainLines = append(mainLines, line)
		}
	}

	mainContent = strings.Join(mainLines, "\n")
	referenceLinks = strings.Join(refLines, "\n")

	// Clean up trailing whitespace from main content
	mainContent = strings.TrimRight(mainContent, "\n")
	referenceLinks = strings.TrimLeft(referenceLinks, "\n")

	return mainContent, referenceLinks
}
