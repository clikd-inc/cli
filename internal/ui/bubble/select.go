package bubble

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"clikd/internal/ui/styles"
)

// SelectResultMsg is sent when a selection is made
type SelectResultMsg struct {
	Value interface{}
}

// PreviewModeMsg is sent when entering preview mode
type PreviewModeMsg struct{}

// ExitPreviewModeMsg is sent when exiting preview mode
type ExitPreviewModeMsg struct{}

// SelectItem represents an item in a selection list
type SelectItem struct {
	Title       string
	Description string
	Value       interface{}
	Preview     string // Optional preview content (markdown)
}

// SelectModel is a model for a selection list
type SelectModel struct {
	Title        string
	Items        []SelectItem
	Cursor       int
	Selected     *SelectItem
	Width        int
	Description  bool
	ShowPreview  bool          // Whether to show preview option
	InPreview    bool          // Whether currently in preview mode
	PreviewModel *PreviewModel // The preview model when active
}

// NewSelectModel creates a new selection model
func NewSelectModel(title string, items []SelectItem) SelectModel {
	// Check if any items have preview content
	showPreview := false
	for _, item := range items {
		if item.Preview != "" {
			showPreview = true
			break
		}
	}

	return SelectModel{
		Title:       title,
		Items:       items,
		Cursor:      0,
		Description: true,
		Width:       80,
		ShowPreview: showPreview,
		InPreview:   false,
	}
}

// Init initializes the model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle preview mode
	if m.InPreview && m.PreviewModel != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c", "esc", "enter", " ":
				// Exit preview mode
				m.InPreview = false
				m.PreviewModel = nil
				return m, nil
			default:
				// Forward other keys to preview model
				updatedPreview, cmd := m.PreviewModel.Update(msg)
				if previewModel, ok := updatedPreview.(PreviewModel); ok {
					m.PreviewModel = &previewModel
				}
				return m, cmd
			}
		default:
			// Forward other messages to preview model
			updatedPreview, cmd := m.PreviewModel.Update(msg)
			if previewModel, ok := updatedPreview.(PreviewModel); ok {
				m.PreviewModel = &previewModel
			}
			return m, cmd
		}
	}

	// Handle normal selection mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
		case "p":
			// Show preview if available for current item
			if m.ShowPreview && m.Cursor < len(m.Items) && m.Items[m.Cursor].Preview != "" {
				previewTitle := "Preview: " + m.Items[m.Cursor].Title
				previewModel := NewPreviewModel(previewTitle, m.Items[m.Cursor].Preview)
				m.PreviewModel = &previewModel
				m.InPreview = true
				return m, nil
			}
		case "enter", " ":
			m.Selected = &m.Items[m.Cursor]
			return m, tea.Batch(
				func() tea.Msg {
					return SelectResultMsg{Value: m.Selected.Value}
				},
			)
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the model
func (m SelectModel) View() string {
	// If in preview mode, show only the preview
	if m.InPreview && m.PreviewModel != nil {
		return m.PreviewModel.View()
	}

	// Normal selection view
	s := ""
	if m.Title != "" {
		s += styles.H2.Render(m.Title) + "\n\n"
	}

	for i, item := range m.Items {
		cursor := " "
		if i == m.Cursor {
			cursor = styles.IconArrow + " "
			s += styles.SelectedStyle.Render(cursor + item.Title)
		} else {
			s += styles.UnselectedStyle.Render(cursor + item.Title)
		}

		if m.Description && item.Description != "" {
			s += "\n" + strings.Repeat(" ", 4) + styles.Subtle.Render(item.Description)
		}

		// Show preview indicator if available
		if item.Preview != "" {
			s += "\n" + strings.Repeat(" ", 4) + styles.Subtle.Render("(Press 'p' to preview)")
		}

		s += "\n"
	}

	s += "\n"
	if m.ShowPreview {
		s += styles.Subtle.Render("↑/↓: Navigate • p: Preview • Enter: Select • Esc: Cancel")
	} else {
		s += styles.Subtle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")
	}

	return s
}

// RunSelect displays a selection list and returns the selected item
func RunSelect(title string, items []SelectItem) *SelectItem {
	m := NewSelectModel(title, items)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil
	}

	result, ok := finalModel.(SelectModel)
	if !ok {
		return nil
	}

	return result.Selected
}
