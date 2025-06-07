package components

import (
	"strings"

	"clikd/pkg/ui"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MultiselectKeyMap definiert die Tastaturbefehle für die Multiselect-Komponente
type MultiselectKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	Submit    key.Binding
	Quit      key.Binding
	Help      key.Binding
	ToggleAll key.Binding
}

// ShortHelp gibt die wichtigsten Tastaturbefehle zurück
func (k MultiselectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Submit, k.Quit}
}

// FullHelp gibt alle Tastaturbefehle zurück
func (k MultiselectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.ToggleAll, k.Submit},
		{k.Help, k.Quit},
	}
}

// DefaultMultiselectKeyMap gibt die Standardtastaturbefehle zurück
func DefaultMultiselectKeyMap() MultiselectKeyMap {
	return MultiselectKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "auf"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "ab"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "auswählen/abwählen"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "bestätigen"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q/esc", "abbrechen"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "hilfe"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "alle auswählen/abwählen"),
		),
	}
}

// MultiselectItem repräsentiert ein auswählbares Item
type MultiselectItem struct {
	Title       string
	Description string
	Value       interface{}
	Selected    bool
}

// MultiselectModel ist das Model für die Multiselect-Komponente
type MultiselectModel struct {
	Title       string
	Description string
	Items       []MultiselectItem
	Cursor      int
	Width       int
	Height      int
	Keys        MultiselectKeyMap
	Help        help.Model
	ShowHelp    bool
	Quitting    bool
	Submitted   bool
	ConfirmText string
	CancelText  string
	MaxSelected int // 0 = unbegrenzt
}

// NewMultiselectModel erstellt ein neues MultiselectModel mit Standardwerten
func NewMultiselectModel(title, description string, items []MultiselectItem) MultiselectModel {
	return MultiselectModel{
		Title:       title,
		Description: description,
		Items:       items,
		Cursor:      0,
		Width:       80,
		Height:      24,
		Keys:        DefaultMultiselectKeyMap(),
		Help:        help.New(),
		ShowHelp:    true,
		ConfirmText: "Bestätigen",
		CancelText:  "Abbrechen",
		MaxSelected: 0, // 0 = unbegrenzt
	}
}

// Init initialisiert das Model
func (m MultiselectModel) Init() tea.Cmd {
	return nil
}

// Update aktualisiert das Model basierend auf Nachrichten
func (m MultiselectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		case key.Matches(msg, m.Keys.Toggle):
			if len(m.Items) > 0 {
				// Wenn ein MaxSelected gesetzt ist und wir es erreicht haben, können wir nur abwählen
				if m.MaxSelected > 0 && !m.Items[m.Cursor].Selected {
					// Prüfen, ob wir das Maximum erreicht haben
					selectedCount := 0
					for _, item := range m.Items {
						if item.Selected {
							selectedCount++
						}
					}

					if selectedCount < m.MaxSelected {
						// Wir können noch ein weiteres Item auswählen
						m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
					}
				} else {
					// Kein Maximum oder wir wählen ab
					m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
				}
			}

		case key.Matches(msg, m.Keys.ToggleAll):
			// Bestimme, ob wir alle auswählen oder abwählen sollen
			// Wenn alle ausgewählt sind, wählen wir alle ab, sonst wählen wir alle aus
			allSelected := true
			for _, item := range m.Items {
				if !item.Selected {
					allSelected = false
					break
				}
			}

			// Wenn MaxSelected gesetzt ist, können wir nur bis zu diesem Limit auswählen
			if m.MaxSelected > 0 && !allSelected {
				// Wähle die ersten MaxSelected Items aus
				for i := 0; i < len(m.Items); i++ {
					m.Items[i].Selected = i < m.MaxSelected
				}
			} else {
				// Alle auswählen oder abwählen
				for i := range m.Items {
					m.Items[i].Selected = !allSelected
				}
			}

		case key.Matches(msg, m.Keys.Submit):
			m.Submitted = true
			return m, tea.Quit

		case key.Matches(msg, m.Keys.Help):
			m.ShowHelp = !m.ShowHelp
		}
	}

	return m, nil
}

// View rendert die Komponente
func (m MultiselectModel) View() string {
	var s strings.Builder

	// Titel
	s.WriteString(ui.H1.Render(m.Title) + "\n\n")

	// Beschreibung
	if m.Description != "" {
		s.WriteString(ui.NormalText.Render(m.Description) + "\n\n")
	}

	// Anzahl der ausgewählten Items
	selectedCount := 0
	for _, item := range m.Items {
		if item.Selected {
			selectedCount++
		}
	}

	// Anzeige der Anzahl ausgewählter Items
	countStr := string('0' + rune(selectedCount))
	totalStr := string('0' + rune(len(m.Items)))
	countText := ui.InfoText("Ausgewählt: " + ui.Selected.Render(countStr) + "/" + ui.BoldText.Render(totalStr))

	// Max-Limit-Anzeige, falls vorhanden
	if m.MaxSelected > 0 {
		countText += ui.SubtleText.Render(" (max: " + string('0'+rune(m.MaxSelected)) + ")")
	}

	s.WriteString(countText + "\n\n")

	// Optionen
	for i, item := range m.Items {
		cursor := " "
		checkbox := "[ ]"
		style := ui.NormalText

		// Cursor oder ausgewählt?
		if m.Cursor == i {
			cursor = ">"
			style = ui.Selected
		}

		// Checkbox-Status
		if item.Selected {
			checkbox = "[x]"
		}

		// Haupteintrag mit Beschreibung
		line := style.Render(cursor + " " + checkbox + " " + item.Title)

		// Wenn eine Beschreibung vorhanden ist, diese hinzufügen
		if item.Description != "" {
			line += "\n    " + ui.SubtleText.Render(item.Description)
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
	buttons = append(buttons, ui.Selected.Render("↵")+" "+okButton)

	// Abbruchbutton
	cancelButton := ui.ErrorText(" " + m.CancelText + " ")
	buttons = append(buttons, ui.Selected.Render("q")+" "+cancelButton)

	actionbar := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)

	padding := strings.Repeat("\n", max(0, m.Height-strings.Count(s.String(), "\n")-4))

	return s.String() + padding + helpView + "\n" + actionbar
}

// GetSelectedItems gibt die ausgewählten Items zurück
func (m MultiselectModel) GetSelectedItems() []MultiselectItem {
	var selected []MultiselectItem
	for _, item := range m.Items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// GetSelectedValues gibt die Werte der ausgewählten Items zurück
func (m MultiselectModel) GetSelectedValues() []interface{} {
	var values []interface{}
	for _, item := range m.Items {
		if item.Selected {
			values = append(values, item.Value)
		}
	}
	return values
}

// RunMultiselect führt eine Mehrfachauswahl aus und gibt die ausgewählten Items zurück
func RunMultiselect(title, description string, items []MultiselectItem) []MultiselectItem {
	m := NewMultiselectModel(title, description, items)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil
	}

	finalMultiselectModel, ok := finalModel.(MultiselectModel)
	if !ok || !finalMultiselectModel.Submitted {
		return nil
	}

	return finalMultiselectModel.GetSelectedItems()
}

// RunMultiselectWithValues führt eine Mehrfachauswahl aus und gibt die Werte der ausgewählten Items zurück
func RunMultiselectWithValues(title, description string, items []MultiselectItem) []interface{} {
	selected := RunMultiselect(title, description, items)
	if selected == nil {
		return nil
	}

	var values []interface{}
	for _, item := range selected {
		values = append(values, item.Value)
	}
	return values
}

// NewMultiselectModelWithMaxSelected erstellt ein neues MultiselectModel mit einer maximalen Anzahl an auswählbaren Items
func NewMultiselectModelWithMaxSelected(title, description string, items []MultiselectItem, maxSelected int) MultiselectModel {
	m := NewMultiselectModel(title, description, items)
	m.MaxSelected = maxSelected
	return m
}

// RunMultiselectWithMaxSelected führt eine Mehrfachauswahl mit einer maximalen Anzahl an auswählbaren Items aus
func RunMultiselectWithMaxSelected(title, description string, items []MultiselectItem, maxSelected int) []MultiselectItem {
	m := NewMultiselectModelWithMaxSelected(title, description, items, maxSelected)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil
	}

	finalMultiselectModel, ok := finalModel.(MultiselectModel)
	if !ok || !finalMultiselectModel.Submitted {
		return nil
	}

	return finalMultiselectModel.GetSelectedItems()
}
