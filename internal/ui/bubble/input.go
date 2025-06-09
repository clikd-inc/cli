package bubble

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"clikd/internal/ui/styles"
)

// InputResultMsg is sent when input is submitted
type InputResultMsg struct {
	Value string
}

// InputModel is a model for a text input
type InputModel struct {
	Title       string
	Description string
	Placeholder string
	TextInput   textinput.Model
	Value       string
	Width       int
	CharLimit   int
	IsPassword  bool
}

// NewInputModel creates a new input model
func NewInputModel(title, description, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 156
	ti.Width = 60
	ti.Prompt = "> "
	ti.TextStyle = styles.Normal
	ti.PromptStyle = styles.InputPrompt
	ti.Cursor.Blink = true
	ti.Focus()

	return InputModel{
		Title:       title,
		Description: description,
		Placeholder: placeholder,
		TextInput:   ti,
		Width:       80,
		CharLimit:   156,
	}
}

// NewPasswordInputModel creates a new password input model
func NewPasswordInputModel(title, description, placeholder string) InputModel {
	m := NewInputModel(title, description, placeholder)
	m.IsPassword = true
	m.TextInput.EchoMode = textinput.EchoPassword
	return m
}

// Init initializes the model
func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update updates the model
func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Value = m.TextInput.Value()
			// If the input is empty, use the placeholder value
			if m.Value == "" && m.Placeholder != "" {
				m.Value = m.Placeholder
			}
			return m, func() tea.Msg {
				return InputResultMsg{Value: m.Value}
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

// View renders the model
func (m InputModel) View() string {
	s := ""
	if m.Title != "" {
		s += styles.H2.Render(m.Title) + "\n\n"
	}

	if m.Description != "" {
		s += styles.Normal.Render(m.Description) + "\n\n"
	}

	s += m.TextInput.View() + "\n\n"

	s += styles.Subtle.Render("Enter: Confirm • Esc: Cancel")

	return s
}

// RunInput displays an input dialog and returns the entered text
func RunInput(title, description, placeholder string) string {
	m := NewInputModel(title, description, placeholder)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return ""
	}

	result, ok := finalModel.(InputModel)
	if !ok {
		return ""
	}

	return result.Value
}

// RunPasswordInput displays a password input dialog and returns the entered password
func RunPasswordInput(title, description, placeholder string) string {
	m := NewPasswordInputModel(title, description, placeholder)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return ""
	}

	result, ok := finalModel.(InputModel)
	if !ok {
		return ""
	}

	return result.Value
}
