package bubble

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"clikd/internal/services"
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
	Timer           timer.Model

	// Real timing tracking
	StartTime time.Time

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
	// Create a countdown timer with more realistic duration based on AI usage
	countdownTime := time.Second * 10 // Default countdown time (10s)
	if !config.NoAI {
		countdownTime = time.Second * 20 // Longer for AI processing (20s)
	}

	t := timer.NewWithInterval(countdownTime, time.Millisecond*100)

	return ChangelogViewerModel{
		Title:           title,
		IsGenerating:    true,
		GeneratorConfig: config,
		Query:           query,
		Timer:           t,
		StartTime:       time.Now(),
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
		// Start timer and generation immediately
		return tea.Batch(
			m.Timer.Start(), // Start the timer immediately
			m.generateChangelogAsync(),
		)
	}
	return nil
}

// generateChangelogAsync starts the changelog generation asynchronously
func (m ChangelogViewerModel) generateChangelogAsync() tea.Cmd {
	return func() tea.Msg {
		// Create a channel for the result
		resultChan := make(chan GenerateCompleteMsg, 1)

		// Start generation in background
		go func() {
			defer close(resultChan)

			// Schritt 1: Konfiguration laden
			configPath := utils.ResolveConfigPath(m.GeneratorConfig.ConfigPath)
			configFileInfo, err := os.Stat(configPath)
			if err != nil {
				resultChan <- GenerateCompleteMsg{Error: fmt.Errorf("configuration file not found at %s: %v", configPath, err)}
				return
			}
			if configFileInfo.IsDir() {
				configPath = filepath.Join(configPath, "config.yml")
			}

			// Config-Pfad korrigieren
			m.GeneratorConfig.ConfigPath = configPath

			// Use the same log level as the main application
			logLevel := "info" // default
			if envLevel := os.Getenv("CLIKD_LOG_LEVEL"); envLevel != "" {
				logLevel = envLevel
			}
			logger := utils.NewLogger(logLevel, !m.GeneratorConfig.NoColor)

			config, err := changelog.LoadConfigFromCommand(m.GeneratorConfig)
			if err != nil {
				resultChan <- GenerateCompleteMsg{Error: err}
				return
			}

			generator := changelog.NewGenerator(logger, config)

			// Try to create and inject AI service for enhancement using ServiceFactory (unless NoAI is set)
			if !m.GeneratorConfig.NoAI {
				ctx := context.Background()

				// Use ServiceFactory for proper AI service creation
				factory, err := services.NewServiceFactory(ctx)
				if err != nil {
					logger.Debug("Could not create service factory, proceeding without AI enhancement: %v", err)
				} else {
					// Try to create AI service via factory
					aiService, err := factory.CreateAIService()
					if err != nil {
						logger.Debug("Could not create AI service, proceeding without AI enhancement: %v", err)
					} else {
						logger.Debug("AI service created successfully, changelog will be enhanced")
						generator.SetAIService(aiService)
					}
				}
			} else {
				logger.Debug("AI enhancement disabled via --no-ai flag")
			}

			buffer := &bytes.Buffer{}

			err = generator.Generate(buffer, m.Query)
			if err != nil {
				resultChan <- GenerateCompleteMsg{Error: err}
				return
			}

			resultChan <- GenerateCompleteMsg{Content: buffer.String()}
		}()

		// Wait for the final result
		return <-resultChan
	}
}

// Helper functions for the changelog viewer
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
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

// tickProgress returns a command that ticks the progress bar
func (m ChangelogViewerModel) tickProgress() tea.Cmd {
	return tea.Tick(time.Millisecond*30, func(t time.Time) tea.Msg {
		return ProgressTickMsg(t)
	})
}

// Update updates the model
func (m ChangelogViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GenerateCompleteMsg:
		// Generation actually completed
		if msg.Error != nil {
			// Handle error case
			m.Content = fmt.Sprintf("Error generating changelog: %v", msg.Error)
			m.RenderedContent = m.Content
			m.IsGenerating = false
			return m, nil
		} else {
			// Process successful content
			m.Content = msg.Content

			// Separate main content from reference links to prevent Glamour from moving them
			mainContent, referenceLinks := separateReferenceLinks(msg.Content)

			// Render main content with Glamour for beautiful styling
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStylePath("tokyo-night"),
				glamour.WithWordWrap(120),
			)

			var rendered string
			if err != nil {
				// Fallback to plain text if glamour fails
				rendered = mainContent
			} else {
				renderedMain, err := renderer.Render(mainContent)
				if err != nil {
					// Fallback to plain text if rendering fails
					rendered = mainContent
				} else {
					rendered = renderedMain
				}
			}

			// Combine rendered main content with unprocessed reference links
			if referenceLinks != "" {
				m.RenderedContent = rendered + "\n" + referenceLinks
			} else {
				m.RenderedContent = rendered
			}

			// Calculate scroll bounds
			lines := strings.Split(m.RenderedContent, "\n")
			m.MaxScroll = len(lines) - m.Viewport
			if m.MaxScroll < 0 {
				m.MaxScroll = 0
			}
		}

		m.IsGenerating = false
		return m, nil

	case SaveCompleteMsg:
		m.ShowingSaveInput = false
		if msg.Error != nil {
			m.SaveMessage = fmt.Sprintf("Error saving file: %v", msg.Error)
		} else {
			m.SaveMessage = fmt.Sprintf("Changelog saved to: %s", msg.Filename)
		}
		m.SaveMessageTime = 150 // Show message for ~3 seconds at 50fps
		return m, nil

	case tea.KeyMsg:
		// Handle save input first
		if m.ShowingSaveInput {
			switch msg.String() {
			case "enter":
				filename := strings.TrimSpace(m.SaveInput.Value())
				if filename == "" {
					filename = "CHANGELOG.md" // Default filename
				}
				return m, m.saveToFile(filename)
			case "esc":
				m.ShowingSaveInput = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.SaveInput, cmd = m.SaveInput.Update(msg)
				return m, cmd
			}
		}

		// Skip input handling during generation
		if m.IsGenerating {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}

		// Normal navigation
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.Scroll > 0 {
				m.Scroll--
			}
		case "down", "j":
			if m.Scroll < m.MaxScroll {
				m.Scroll++
			}
		case "pgup":
			m.Scroll -= 10
			if m.Scroll < 0 {
				m.Scroll = 0
			}
		case "pgdown":
			m.Scroll += 10
			if m.Scroll > m.MaxScroll {
				m.Scroll = m.MaxScroll
			}
		case "home":
			m.Scroll = 0
		case "end":
			m.Scroll = m.MaxScroll
		case "h":
			m.ShowHelp = !m.ShowHelp
		case "c":
			// Copy to clipboard
			err := clipboard.WriteAll(m.Content)
			if err != nil {
				m.CopyMessage = "Failed to copy to clipboard"
			} else {
				m.CopyMessage = "Changelog copied to clipboard!"
			}
			m.CopyMessageTime = 150 // Show message for ~3 seconds at 50fps
		case "s":
			// Show save input
			m.ShowingSaveInput = true
			m.initSaveInput()
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	// Timer messages
	case timer.TickMsg:
		var cmd tea.Cmd
		m.Timer, cmd = m.Timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.Timer, cmd = m.Timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		// Timer finished, but generation might still be running
		// Just let it continue
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
	var s strings.Builder
	padding := "  "

	if m.IsGenerating {
		// Add extra top padding for better spacing
		s.WriteString("\n\n")
		s.WriteString(padding + styles.H2.Render("🔄 Generating Changelog") + "\n\n")

		// Show timer or "still generating" message if timer has timed out
		if !m.Timer.Timedout() {
			// Timer still counting down
			timerText := m.Timer.View()
			s.WriteString(padding + styles.SuccessStyle.Render("⏱️  Generating will complete in: ") +
				styles.InfoStyle.Render(timerText) + "\n\n")
		} else {
			// Timer has reached zero but generation is still running
			elapsed := time.Since(m.StartTime).Round(time.Second)
			s.WriteString(padding + styles.SuccessStyle.Render("⏱️  Still generating... ") +
				styles.InfoStyle.Render(fmt.Sprintf("(%s elapsed)", elapsed)) + "\n\n")
		}

		// Simple status message with more space
		s.WriteString(padding + styles.Subtle.Render("→ Processing commits and generating changelog...") + "\n")

		return s.String()
	}

	// Save-Input-Dialog anzeigen
	if m.ShowingSaveInput {
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

	// Padding zu jeder Zeile hinzufügen (reuse existing padding variable)
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

// truncateCommitMessage truncates a commit message to fit in the UI
func truncateCommitMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}

	// Try to truncate at word boundary
	if maxLength > 3 {
		truncated := message[:maxLength-3]
		if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLength/2 {
			return message[:lastSpace] + "..."
		}
	}

	return message[:maxLength-3] + "..."
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
