package components

import (
	"strings"

	"clikd/pkg/ui"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmKeyMap definiert die Tastaturbefehle für die Bestätigungskomponente
type ConfirmKeyMap struct {
	Yes   key.Binding
	No    key.Binding
	Quit  key.Binding
	Help  key.Binding
	Left  key.Binding
	Right key.Binding
}

// ShortHelp gibt die wichtigsten Tastaturbefehle zurück
func (k ConfirmKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Yes, k.No, k.Quit}
}

// FullHelp gibt alle Tastaturbefehle zurück
func (k ConfirmKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Yes, k.No},
		{k.Left, k.Right},
		{k.Help, k.Quit},
	}
}

// DefaultConfirmKeyMap gibt die Standardtastaturbefehle zurück
func DefaultConfirmKeyMap() ConfirmKeyMap {
	return ConfirmKeyMap{
		Yes: key.NewBinding(
			key.WithKeys("y", "j"),
			key.WithHelp("y/j", "ja"),
		),
		No: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "nein"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "links"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "rechts"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q/esc", "abbrechen"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "hilfe"),
		),
	}
}

// ConfirmModel ist das Model für die Bestätigungskomponente
type ConfirmModel struct {
	Title    string
	Message  string
	YesText  string
	NoText   string
	Cursor   int // 0 = Ja, 1 = Nein
	Width    int
	Height   int
	Keys     ConfirmKeyMap
	Help     help.Model
	ShowHelp bool
	Result   *bool // nil = keine Auswahl, true = Ja, false = Nein
}

// NewConfirmModel erstellt ein neues ConfirmModel mit Standardwerten
func NewConfirmModel(title, message string) ConfirmModel {
	return ConfirmModel{
		Title:    title,
		Message:  message,
		YesText:  "Ja",
		NoText:   "Nein",
		Cursor:   0,
		Width:    80,
		Height:   15,
		Keys:     DefaultConfirmKeyMap(),
		Help:     help.New(),
		ShowHelp: true,
		Result:   nil,
	}
}

// Init initialisiert das Model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update aktualisiert das Model basierend auf Nachrichten
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Yes):
			result := true
			m.Result = &result
			m.Cursor = 0
			return m, tea.Quit

		case key.Matches(msg, m.Keys.No):
			result := false
			m.Result = &result
			m.Cursor = 1
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Left):
			m.Cursor = 0 // Ja

		case key.Matches(msg, m.Keys.Right):
			m.Cursor = 1 // Nein

		case key.Matches(msg, m.Keys.Help):
			m.ShowHelp = !m.ShowHelp

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.Cursor == 0 {
				result := true
				m.Result = &result
			} else {
				result := false
				m.Result = &result
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

// View rendert die Komponente
func (m ConfirmModel) View() string {
	var s strings.Builder

	// Titel
	s.WriteString(ui.H1.Render(m.Title) + "\n\n")

	// Nachricht
	msgStyle := ui.BoxStyle.Copy().BorderForeground(ui.BorderColor)
	s.WriteString(msgStyle.Render(m.Message) + "\n\n")

	// Buttons
	yesStyle := ui.NormalText
	noStyle := ui.NormalText

	if m.Cursor == 0 {
		yesStyle = ui.Selected.Copy().Background(ui.SuccessColor).Foreground(ui.TextColor)
	} else {
		noStyle = ui.Selected.Copy().Background(ui.ErrorColor).Foreground(ui.TextColor)
	}

	yesButton := yesStyle.Render(" " + m.YesText + " ")
	noButton := noStyle.Render(" " + m.NoText + " ")

	// Labels für die Buttons
	yesLabel := ""
	noLabel := ""

	if m.Cursor == 0 {
		yesLabel = ui.Selected.Render("[Y]") + " "
	} else {
		yesLabel = ui.SubtleText.Render("[Y]") + " "
	}

	if m.Cursor == 1 {
		noLabel = ui.Selected.Render("[N]") + " "
	} else {
		noLabel = ui.SubtleText.Render("[N]") + " "
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		yesLabel+yesButton,
		"     ",
		noLabel+noButton,
	)

	s.WriteString(lipgloss.Place(m.Width, 3, lipgloss.Center, lipgloss.Center, buttons))

	// Hilfe
	helpView := ""
	if m.ShowHelp {
		helpView = "\n" + m.Help.View(m.Keys)
	}

	padding := strings.Repeat("\n", max(0, m.Height-strings.Count(s.String(), "\n")-4))

	return s.String() + padding + helpView
}

// GetResult gibt das Ergebnis zurück
func (m ConfirmModel) GetResult() *bool {
	return m.Result
}

// RunConfirm führt eine Bestätigungsabfrage aus und gibt das Ergebnis zurück
// Gibt true für Ja, false für Nein und nil für Abbruch zurück
func RunConfirm(title, message string) *bool {
	m := NewConfirmModel(title, message)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil
	}

	finalConfirmModel, ok := finalModel.(ConfirmModel)
	if !ok {
		return nil
	}

	return finalConfirmModel.Result
}

// Confirm ist eine Hilfsfunktion, die eine Bestätigungsabfrage ausführt und ein Boolean zurückgibt
// Gibt true für Ja und false für Nein oder Abbruch zurück
func Confirm(title, message string) bool {
	result := RunConfirm(title, message)
	return result != nil && *result
}

// ConfirmWithDefault ist eine Hilfsfunktion, die eine Bestätigungsabfrage ausführt
// und einen Standardwert zurückgibt, wenn abgebrochen wird
func ConfirmWithDefault(title, message string, defaultValue bool) bool {
	result := RunConfirm(title, message)
	if result == nil {
		return defaultValue
	}
	return *result
}
