package components

import (
	"fmt"
	"strings"

	"clikd/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ListItem repräsentiert ein Listenelement
type ListItem struct {
	Title       string
	Description string
	Status      string
	Tags        []string
	Metadata    map[string]string
}

// ListModel ist das Model für die Listenkomponente
type ListModel struct {
	Title       string
	Description string
	Items       []ListItem
	Width       int
	Height      int
	PageSize    int
	Page        int
	ShowTags    bool
	ShowStatus  bool
	AutoClose   bool
}

// NewListModel erstellt ein neues ListModel mit Standardwerten
func NewListModel(title, description string, items []ListItem) ListModel {
	return ListModel{
		Title:       title,
		Description: description,
		Items:       items,
		Width:       80,
		Height:      24,
		PageSize:    10,
		Page:        0,
		ShowTags:    true,
		ShowStatus:  true,
		AutoClose:   true,
	}
}

// Init initialisiert das Model
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update aktualisiert das Model basierend auf Nachrichten
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "n", "j", "down", "pgdown":
			totalPages := (len(m.Items) + m.PageSize - 1) / m.PageSize
			if m.Page < totalPages-1 {
				m.Page++
			}
		case "p", "k", "up", "pgup":
			if m.Page > 0 {
				m.Page--
			}
		case "home":
			m.Page = 0
		case "end":
			totalPages := (len(m.Items) + m.PageSize - 1) / m.PageSize
			m.Page = totalPages - 1
		case "enter", " ":
			if m.AutoClose {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View rendert die Komponente
func (m ListModel) View() string {
	var s strings.Builder

	// Titel
	s.WriteString(ui.H1.Render(m.Title) + "\n\n")

	// Beschreibung
	if m.Description != "" {
		s.WriteString(ui.NormalText.Render(m.Description) + "\n\n")
	}

	// Keine Items
	if len(m.Items) == 0 {
		s.WriteString(ui.SubtleText.Render("Keine Einträge vorhanden.") + "\n")
		return s.String()
	}

	// Paginierung berechnen
	start := m.Page * m.PageSize
	end := start + m.PageSize
	if end > len(m.Items) {
		end = len(m.Items)
	}

	// Items anzeigen
	for i := start; i < end; i++ {
		item := m.Items[i]

		// Titelstyling
		titleStyle := ui.BoldText
		title := titleStyle.Render(item.Title)

		// Status, falls vorhanden
		statusText := ""
		if m.ShowStatus && item.Status != "" {
			// Statusfarben basierend auf Status-Text
			switch strings.ToLower(item.Status) {
			case "done", "completed", "fertig", "abgeschlossen":
				statusText = ui.SuccessText(" [" + item.Status + "]")
			case "pending", "offen", "todo", "ausstehend":
				statusText = ui.WarningText(" [" + item.Status + "]")
			case "error", "failed", "fehler", "fehlgeschlagen":
				statusText = ui.ErrorText(" [" + item.Status + "]")
			case "in progress", "in bearbeitung", "läuft":
				statusText = ui.InfoText(" [" + item.Status + "]")
			default:
				statusText = ui.SubtleText.Render(" [" + item.Status + "]")
			}
		}

		// Zeilenzusammensetzung: Titel + Status
		line := fmt.Sprintf("%d. %s%s", i+1, title, statusText)
		s.WriteString(line + "\n")

		// Beschreibung, falls vorhanden
		if item.Description != "" {
			description := ui.SubtleText.Render("   " + item.Description)
			s.WriteString(description + "\n")
		}

		// Tags, falls vorhanden
		if m.ShowTags && len(item.Tags) > 0 {
			var tagList []string
			for _, tag := range item.Tags {
				tagStyle := ui.SubtleText.Copy().Background(ui.HighlightColor)
				tagList = append(tagList, tagStyle.Render(" "+tag+" "))
			}
			tags := "   " + strings.Join(tagList, " ")
			s.WriteString(tags + "\n")
		}

		// Metadata, falls vorhanden
		if len(item.Metadata) > 0 {
			for key, value := range item.Metadata {
				metaLine := fmt.Sprintf("   %s: %s",
					ui.SubtleText.Render(key),
					ui.NormalText.Render(value))
				s.WriteString(metaLine + "\n")
			}
		}

		s.WriteString("\n")
	}

	// Paginierungsinfo
	totalPages := (len(m.Items) + m.PageSize - 1) / m.PageSize
	if totalPages > 1 {
		pageInfo := fmt.Sprintf("Seite %d von %d", m.Page+1, totalPages)
		navigation := "Verwende ↑/↓ zum Navigieren, Enter zum Schließen"

		paginationText := ui.SubtleText.Render(pageInfo + " • " + navigation)
		s.WriteString("\n" + paginationText)
	}

	return s.String()
}

// ShowList zeigt eine einfache Liste an
func ShowList(title, description string, items []ListItem) {
	m := NewListModel(title, description, items)
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Println("Fehler beim Anzeigen der Liste:", err)
	}
}

// ShowFormattedList zeigt eine formatierte Liste mit benutzerdefinierten Optionen an
func ShowFormattedList(title, description string, items []ListItem, showTags, showStatus bool, pageSize int) {
	m := NewListModel(title, description, items)
	m.ShowTags = showTags
	m.ShowStatus = showStatus
	m.PageSize = pageSize
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Println("Fehler beim Anzeigen der Liste:", err)
	}
}

// FormatList formatiert eine Liste von Elementen als String
func FormatList(title string, items []ListItem, showTags, showStatus bool) string {
	m := NewListModel(title, "", items)
	m.ShowTags = showTags
	m.ShowStatus = showStatus
	m.AutoClose = false
	return m.View()
}

// CreateListItemWithStatus erstellt ein neues ListItem mit Status
func CreateListItemWithStatus(title, description, status string) ListItem {
	return ListItem{
		Title:       title,
		Description: description,
		Status:      status,
	}
}

// CreateListItemWithTags erstellt ein neues ListItem mit Tags
func CreateListItemWithTags(title, description string, tags []string) ListItem {
	return ListItem{
		Title:       title,
		Description: description,
		Tags:        tags,
	}
}

// CreateListItemWithMetadata erstellt ein neues ListItem mit Metadata
func CreateListItemWithMetadata(title, description string, metadata map[string]string) ListItem {
	return ListItem{
		Title:       title,
		Description: description,
		Metadata:    metadata,
	}
}
