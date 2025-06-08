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

// SelectItem represents an item in a selection list
type SelectItem struct {
	Title       string
	Description string
	Value       interface{}
}

// SelectModel is a model for a selection list
type SelectModel struct {
	Title       string
	Items       []SelectItem
	Cursor      int
	Selected    *SelectItem
	Width       int
	Description bool
}

// NewSelectModel creates a new selection model
func NewSelectModel(title string, items []SelectItem) SelectModel {
	return SelectModel{
		Title:       title,
		Items:       items,
		Cursor:      0,
		Description: true,
		Width:       80,
	}
}

// Init initializes the model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		s += "\n"
	}

	s += "\n" + styles.Subtle.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")

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
