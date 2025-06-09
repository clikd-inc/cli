package bubble

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"clikd/internal/ui/styles"
)

// PreviewModel is a model for displaying markdown previews
type PreviewModel struct {
	Title           string
	Content         string
	RenderedContent string
	Width           int
	Height          int
	Viewport        int
	Scroll          int
	MaxScroll       int
	ShowHelp        bool
}

// NewPreviewModel creates a new preview model
func NewPreviewModel(title, content string) PreviewModel {
	// Render the markdown content with glamour using Tokyo Night style
	rendered, err := glamour.Render(content, "tokyo-night")
	if err != nil {
		rendered = content // Fallback to raw content if rendering fails
	}

	// Calculate viewport dimensions for better content display
	lines := strings.Split(rendered, "\n")
	height := 30 // Much larger viewport to show more content
	maxScroll := len(lines) - height + 2
	if maxScroll < 0 {
		maxScroll = 0
	}

	return PreviewModel{
		Title:           title,
		Content:         content,
		RenderedContent: rendered,
		Width:           100,
		Height:          height,
		Viewport:        height,
		Scroll:          0,
		MaxScroll:       maxScroll,
		ShowHelp:        false, // Start with help hidden
	}
}

// Init initializes the model
func (m PreviewModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Scroll > 0 {
				m.Scroll--
			}
		case "down", "j":
			if m.Scroll < m.MaxScroll {
				m.Scroll++
			}
		case "pgup", "b":
			m.Scroll -= m.Viewport / 2
			if m.Scroll < 0 {
				m.Scroll = 0
			}
		case "pgdown", "f":
			m.Scroll += m.Viewport / 2
			if m.Scroll > m.MaxScroll {
				m.Scroll = m.MaxScroll
			}
		case "home", "g":
			m.Scroll = 0
		case "end", "G":
			m.Scroll = m.MaxScroll
		case "h", "?":
			m.ShowHelp = !m.ShowHelp
		case "q", "ctrl+c", "esc", "enter", " ":
			// These will be handled by the parent SelectModel
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// Adjust viewport size based on terminal size
		m.Height = msg.Height - 6 // Leave minimal space for title and status
		m.Viewport = m.Height
		lines := strings.Split(m.RenderedContent, "\n")
		m.MaxScroll = len(lines) - m.Viewport + 1
		if m.MaxScroll < 0 {
			m.MaxScroll = 0
		}
		if m.Scroll > m.MaxScroll {
			m.Scroll = m.MaxScroll
		}
	}
	return m, nil
}

// View renders the model
func (m PreviewModel) View() string {
	var s strings.Builder

	// Simple title line
	if m.Title != "" {
		s.WriteString(styles.H2.Render(m.Title) + "\n")
		s.WriteString(strings.Repeat("─", 80) + "\n\n")
	}

	// Content viewport - clean display without borders
	lines := strings.Split(m.RenderedContent, "\n")
	start := m.Scroll
	end := start + m.Viewport
	if end > len(lines) {
		end = len(lines)
	}

	// Display the visible portion cleanly
	for i := start; i < end; i++ {
		if i < len(lines) {
			s.WriteString(lines[i] + "\n")
		}
	}

	// Compact help if shown
	if m.ShowHelp {
		s.WriteString("\n" + styles.Subtle.Render("↑/↓: Scroll • PgUp/PgDn: Page • Home/End: Jump • h: Hide help • Enter/Esc: Back") + "\n")
	}

	// Compact status bar
	statusInfo := ""
	if m.MaxScroll > 0 {
		percentage := int(float64(m.Scroll) / float64(m.MaxScroll) * 100)
		if m.Scroll == 0 {
			statusInfo = "Top"
		} else if m.Scroll >= m.MaxScroll {
			statusInfo = "Bottom"
		} else {
			statusInfo = strings.Repeat("█", percentage/10) + strings.Repeat("░", 10-percentage/10) + " " +
				string(rune(48+percentage/10)) + string(rune(48+percentage%10)) + "%"
		}
	} else {
		statusInfo = "Complete"
	}

	s.WriteString("\n")
	if !m.ShowHelp {
		s.WriteString(styles.Subtle.Render("Position: " + statusInfo + " • h: Help • Enter/Esc: Back"))
	} else {
		s.WriteString(styles.Subtle.Render("Position: " + statusInfo))
	}

	return s.String()
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RunPreview displays a markdown preview and waits for user input
// This function is kept for backward compatibility but is no longer used
// in the new integrated preview system
func RunPreview(title, content string) {
	m := NewPreviewModel(title, content)
	p := tea.NewProgram(m)
	_, _ = p.Run()
}
