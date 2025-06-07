package components

import (
	"strings"

	"clikd/pkg/ui"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectKeyMap definiert die Tastaturbefehle für die Auswahlkomponente
type SelectKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// ShortHelp gibt die wichtigsten Tastaturbefehle zurück
func (k SelectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Quit}
}

// FullHelp gibt alle Tastaturbefehle zurück
func (k SelectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		{k.Help, k.Quit},
	}
}

// DefaultSelectKeyMap gibt die Standardtastaturbefehle zurück
func DefaultSelectKeyMap() SelectKeyMap {
	return SelectKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "auf"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "ab"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "auswählen"),
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

// SelectItem repräsentiert ein Auswahlitem
type SelectItem struct {
	Title       string
	Description string
	Value       interface{}
}

// SelectModel ist das Model für die Auswahlkomponente
type SelectModel struct {
	Title       string
	Items       []SelectItem
	Cursor      int
	Selected    int
	Width       int
	Height      int
	Keys        SelectKeyMap
	Help        help.Model
	ShowHelp    bool
	Quitting    bool
	ConfirmText string
	CancelText  string
}

// NewSelectModel erstellt ein neues SelectModel mit Standardwerten
func NewSelectModel(title string, items []SelectItem) SelectModel {
	return SelectModel{
		Title:       title,
		Items:       items,
		Cursor:      0,
		Selected:    -1,
		Width:       80,
		Height:      24,
		Keys:        DefaultSelectKeyMap(),
		Help:        help.New(),
		ShowHelp:    true,
		ConfirmText: "Bestätigen",
		CancelText:  "Abbrechen",
	}
}

// Init initialisiert das Model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update aktualisiert das Model basierend auf Nachrichten
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			m.Quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Up):
			if m.Cursor > 0 {
				m.Cursor--
			} else {
				// Wrap around
				m.Cursor = len(m.Items) - 1
			}

		case key.Matches(msg, m.Keys.Down):
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			} else {
				// Wrap around
				m.Cursor = 0
			}

		case key.Matches(msg, m.Keys.Select):
			m.Selected = m.Cursor
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Help):
			m.ShowHelp = !m.ShowHelp
		}
	}

	return m, nil
}

// View rendert die Komponente
func (m SelectModel) View() string {
	var s strings.Builder

	// Titel
	s.WriteString(ui.H1.Render(m.Title) + "\n\n")

	// Optionen
	for i, item := range m.Items {
		cursor := " "
		style := ui.NormalText

		// Cursor oder ausgewählt?
		if m.Cursor == i {
			cursor = ">"
			style = ui.Selected
		}

		// Haupteintrag mit Beschreibung
		line := style.Render(cursor + " " + item.Title)

		// Wenn eine Beschreibung vorhanden ist, diese hinzufügen
		if item.Description != "" {
			line += "\n  " + ui.SubtleText.Render(item.Description)
		}

		s.WriteString(line + "\n\n")
	}

	// Hilfe
	helpView := ""
	if m.ShowHelp {
		helpView = "\n" + m.Help.View(m.Keys)
	}

	// Actionbar am unteren Rand
	buttons := []string{}

	// Bestätigungsbutton
	okButton := ui.SuccessText(" " + m.ConfirmText + " ")
	if m.Cursor >= 0 && m.Cursor < len(m.Items) {
		buttons = append(buttons, ui.Selected.Render("↵")+" "+okButton)
	}

	// Abbruchbutton
	cancelButton := ui.ErrorText(" " + m.CancelText + " ")
	buttons = append(buttons, ui.Selected.Render("q")+" "+cancelButton)

	actionbar := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)

	padding := strings.Repeat("\n", max(0, m.Height-strings.Count(s.String(), "\n")-4))

	return s.String() + padding + helpView + "\n" + actionbar
}

// GetSelected gibt das ausgewählte Item zurück oder nil, wenn abgebrochen wurde
func (m SelectModel) GetSelected() *SelectItem {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		return &m.Items[m.Selected]
	}
	return nil
}

// GetSelectedValue gibt den Wert des ausgewählten Items zurück oder nil, wenn abgebrochen wurde
func (m SelectModel) GetSelectedValue() interface{} {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		return m.Items[m.Selected].Value
	}
	return nil
}

// IsSelected gibt zurück, ob ein Item ausgewählt wurde
func (m SelectModel) IsSelected() bool {
	return m.Selected >= 0 && m.Selected < len(m.Items)
}

// RunSelect führt eine Auswahl aus und gibt das ausgewählte Item zurück
func RunSelect(title string, items []SelectItem) *SelectItem {
	m := NewSelectModel(title, items)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil
	}

	finalSelectModel, ok := finalModel.(SelectModel)
	if !ok {
		return nil
	}

	return finalSelectModel.GetSelected()
}

// RunSelectWithValues führt eine Auswahl aus und gibt den Wert des ausgewählten Items zurück
func RunSelectWithValues(title string, items []SelectItem) interface{} {
	m := NewSelectModel(title, items)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil
	}

	finalSelectModel, ok := finalModel.(SelectModel)
	if !ok {
		return nil
	}

	return finalSelectModel.GetSelectedValue()
}

// Helper für max-Funktion
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
