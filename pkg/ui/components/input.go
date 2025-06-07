package components

import (
	"strings"

	"clikd/pkg/ui"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputKeyMap definiert die Tastaturbefehle für die Eingabekomponente
type InputKeyMap struct {
	Submit key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// ShortHelp gibt die wichtigsten Tastaturbefehle zurück
func (k InputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Quit}
}

// FullHelp gibt alle Tastaturbefehle zurück
func (k InputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit},
		{k.Help, k.Quit},
	}
}

// DefaultInputKeyMap gibt die Standardtastaturbefehle zurück
func DefaultInputKeyMap() InputKeyMap {
	return InputKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "bestätigen"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "abbrechen"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "hilfe"),
		),
	}
}

// InputModel ist das Model für die Eingabekomponente
type InputModel struct {
	Title       string
	Description string
	Placeholder string
	TextInput   textinput.Model
	Keys        InputKeyMap
	Help        help.Model
	ShowHelp    bool
	Value       string
	Width       int
	Height      int
	Submitted   bool
	Quitting    bool
}

// NewInputModel erstellt ein neues InputModel mit Standardwerten
func NewInputModel(title, description, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.Width = 40
	ti.Prompt = "> "
	ti.PromptStyle = ui.Selected

	return InputModel{
		Title:       title,
		Description: description,
		Placeholder: placeholder,
		TextInput:   ti,
		Keys:        DefaultInputKeyMap(),
		Help:        help.New(),
		ShowHelp:    true,
		Width:       80,
		Height:      15,
	}
}

// Init initialisiert das Model
func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update aktualisiert das Model basierend auf Nachrichten
func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			m.Quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Submit):
			m.Value = m.TextInput.Value()
			m.Submitted = true
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Help):
			m.ShowHelp = !m.ShowHelp
		}
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

// View rendert die Komponente
func (m InputModel) View() string {
	var s strings.Builder

	// Titel
	s.WriteString(ui.H1.Render(m.Title) + "\n\n")

	// Beschreibung
	if m.Description != "" {
		descStyle := ui.NormalText
		s.WriteString(descStyle.Render(m.Description) + "\n\n")
	}

	// TextInput
	s.WriteString(m.TextInput.View() + "\n\n")

	// Hilfe
	helpView := ""
	if m.ShowHelp {
		helpView = m.Help.View(m.Keys)
	}

	// Padding hinzufügen
	padding := strings.Repeat("\n", max(0, m.Height-strings.Count(s.String(), "\n")-4))

	if helpView != "" {
		return s.String() + padding + "\n" + helpView
	}

	return s.String() + padding
}

// GetValue gibt den eingegebenen Wert zurück
func (m InputModel) GetValue() string {
	return m.Value
}

// IsSubmitted gibt zurück, ob der Benutzer die Eingabe bestätigt hat
func (m InputModel) IsSubmitted() bool {
	return m.Submitted
}

// IsQuitting gibt zurück, ob der Benutzer abbrechen möchte
func (m InputModel) IsQuitting() bool {
	return m.Quitting
}

// RunInput führt eine Eingabeaufforderung aus und gibt den eingegebenen Wert zurück
// Gibt einen leeren String zurück, wenn der Benutzer abbricht
func RunInput(title, description, placeholder string) string {
	m := NewInputModel(title, description, placeholder)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return ""
	}

	finalInputModel, ok := finalModel.(InputModel)
	if !ok || !finalInputModel.IsSubmitted() {
		return ""
	}

	return finalInputModel.GetValue()
}

// RunInputWithDefault führt eine Eingabeaufforderung aus und gibt den eingegebenen Wert zurück
// Wenn der Benutzer abbricht, wird der Standardwert zurückgegeben
func RunInputWithDefault(title, description, placeholder, defaultValue string) string {
	value := RunInput(title, description, placeholder)
	if value == "" {
		return defaultValue
	}
	return value
}

// SetInput initialisiert das Textfeld mit einem Wert
func (m *InputModel) SetInput(value string) {
	m.TextInput.SetValue(value)
}

// NewInputModelWithValue erstellt ein neues InputModel mit einem vordefinierten Wert
func NewInputModelWithValue(title, description, placeholder, value string) InputModel {
	m := NewInputModel(title, description, placeholder)
	m.TextInput.SetValue(value)
	return m
}
