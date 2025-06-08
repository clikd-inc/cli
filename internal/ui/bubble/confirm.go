package bubble

import (
	tea "github.com/charmbracelet/bubbletea"

	"clikd/internal/ui/styles"
)

// ConfirmResultMsg is sent when a confirmation dialog is completed
type ConfirmResultMsg struct {
	Result bool
}

// ConfirmModel is a model for a yes/no confirmation dialog
type ConfirmModel struct {
	Title       string
	Message     string
	YesText     string
	NoText      string
	Cursor      int
	Result      bool
	HasSelected bool
}

// NewConfirmModel creates a new confirmation model
func NewConfirmModel(title, message string) ConfirmModel {
	return ConfirmModel{
		Title:   title,
		Message: message,
		YesText: "Yes",
		NoText:  "No",
		Cursor:  0, // Default to "Yes"
	}
}

// Init initializes the model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.Cursor = 0 // Yes
		case "right", "l":
			m.Cursor = 1 // No
		case "enter", " ":
			m.Result = m.Cursor == 0
			m.HasSelected = true
			return m, func() tea.Msg {
				return ConfirmResultMsg{Result: m.Result}
			}
		case "y", "Y":
			m.Result = true
			m.HasSelected = true
			return m, func() tea.Msg {
				return ConfirmResultMsg{Result: true}
			}
		case "n", "N":
			m.Result = false
			m.HasSelected = true
			return m, func() tea.Msg {
				return ConfirmResultMsg{Result: false}
			}
		case "q", "ctrl+c", "esc":
			m.Result = false
			m.HasSelected = true
			return m, func() tea.Msg {
				return ConfirmResultMsg{Result: false}
			}
		}
	}

	return m, nil
}

// View renders the model
func (m ConfirmModel) View() string {
	s := ""
	if m.Title != "" {
		s += styles.H2.Render(m.Title) + "\n\n"
	}

	if m.Message != "" {
		s += styles.Normal.Render(m.Message) + "\n\n"
	}

	// Render Yes/No options
	yesStyle := styles.UnselectedStyle
	noStyle := styles.UnselectedStyle

	if m.Cursor == 0 {
		yesStyle = styles.SelectedStyle
	} else {
		noStyle = styles.SelectedStyle
	}

	s += yesStyle.Render("  " + m.YesText + "  ")
	s += "    "
	s += noStyle.Render("  " + m.NoText + "  ")
	s += "\n\n"

	s += styles.Subtle.Render("←/→: Navigate • Enter: Select • Y/N: Quick select • Esc: Cancel")

	return s
}

// RunConfirm displays a confirmation dialog and returns the user's choice
func RunConfirm(title, message string) bool {
	m := NewConfirmModel(title, message)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false
	}

	result, ok := finalModel.(ConfirmModel)
	if !ok || !result.HasSelected {
		return false
	}

	return result.Result
}
