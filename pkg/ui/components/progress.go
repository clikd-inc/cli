package components

import (
	"fmt"
	"strings"
	"time"

	"clikd/pkg/ui"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressModel ist das Model für die Fortschrittskomponente
type ProgressModel struct {
	Title        string
	Description  string
	Progress     progress.Model
	Percent      float64
	Width        int
	ShowPercent  bool
	ShowValue    bool
	Value        int
	MaxValue     int
	ValueUnit    string
	StartTime    time.Time
	ShowDuration bool
	Done         bool
}

// NewProgressModel erstellt ein neues ProgressModel mit Standardwerten
func NewProgressModel(title, description string) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return ProgressModel{
		Title:        title,
		Description:  description,
		Progress:     p,
		Percent:      0.0,
		Width:        80,
		ShowPercent:  true,
		ShowValue:    false,
		Value:        0,
		MaxValue:     100,
		ValueUnit:    "",
		StartTime:    time.Now(),
		ShowDuration: true,
		Done:         false,
	}
}

// Init initialisiert das Model
func (m ProgressModel) Init() tea.Cmd {
	return nil
}

// Update aktualisiert das Model basierend auf Nachrichten
type TickMsg time.Time

func Tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type FrameMsg int

func Frame() tea.Cmd {
	return tea.Tick(time.Second/30, func(time.Time) tea.Msg {
		return FrameMsg(1)
	})
}

type SetPercentMsg float64

// SetPercent sendet eine Nachricht zum Aktualisieren des Fortschritts
func SetPercent(percent float64) tea.Cmd {
	return func() tea.Msg {
		return SetPercentMsg(percent)
	}
}

type SetValueMsg struct {
	Value    int
	MaxValue int
}

// SetValue sendet eine Nachricht zum Aktualisieren des Werts
func SetValue(value, maxValue int) tea.Cmd {
	return func() tea.Msg {
		return SetValueMsg{Value: value, MaxValue: maxValue}
	}
}

type FinishMsg bool

// Finish sendet eine Nachricht zum Beenden der Fortschrittsanzeige
func Finish() tea.Cmd {
	return func() tea.Msg {
		return FinishMsg(true)
	}
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case FrameMsg:
		// Animationsframe aktualisieren
		progressCmd := m.Progress.SetPercent(m.Percent)
		return m, progressCmd

	case SetPercentMsg:
		m.Percent = float64(msg)
		if m.ShowValue {
			m.Value = int(m.Percent * float64(m.MaxValue))
		}
		if m.Percent >= 1.0 {
			m.Done = true
			return m, tea.Sequence(
				m.Progress.SetPercent(1.0),
				tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
					return FinishMsg(true)
				}),
			)
		}
		return m, m.Progress.SetPercent(m.Percent)

	case SetValueMsg:
		m.Value = msg.Value
		m.MaxValue = msg.MaxValue
		m.Percent = float64(m.Value) / float64(m.MaxValue)
		if m.Percent >= 1.0 {
			m.Done = true
			return m, tea.Sequence(
				m.Progress.SetPercent(1.0),
				tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg {
					return FinishMsg(true)
				}),
			)
		}
		return m, m.Progress.SetPercent(m.Percent)

	case FinishMsg:
		m.Done = bool(msg)
		return m, tea.Quit
	}

	return m, nil
}

// View rendert die Komponente
func (m ProgressModel) View() string {
	var s strings.Builder

	// Titel
	s.WriteString(ui.H2.Render(m.Title) + "\n\n")

	// Beschreibung
	if m.Description != "" {
		s.WriteString(ui.NormalText.Render(m.Description) + "\n\n")
	}

	// Fortschrittsbalken
	s.WriteString(m.Progress.View() + "\n\n")

	// Zusätzliche Informationen
	var infos []string

	// Prozent
	if m.ShowPercent {
		percentStr := fmt.Sprintf("%.0f%%", m.Percent*100)
		infos = append(infos, percentStr)
	}

	// Wert
	if m.ShowValue {
		valueStr := fmt.Sprintf("%d/%d", m.Value, m.MaxValue)
		if m.ValueUnit != "" {
			valueStr += " " + m.ValueUnit
		}
		infos = append(infos, valueStr)
	}

	// Dauer
	if m.ShowDuration {
		elapsed := time.Since(m.StartTime).Round(time.Second)
		infos = append(infos, fmt.Sprintf("Zeit: %s", elapsed))
	}

	// Status
	status := "In Bearbeitung..."
	statusStyle := ui.InfoText
	if m.Done {
		status = "Abgeschlossen!"
		statusStyle = ui.SuccessText
	}

	infoLine := strings.Join(infos, " | ")
	if infoLine != "" {
		s.WriteString(ui.SubtleText.Render(infoLine) + "\n\n")
	}

	s.WriteString(statusStyle(status) + "\n")

	return s.String()
}

// RunProgress führt eine Fortschrittsanzeige aus
// progressFn sollte eine Funktion sein, die den Fortschritt aktualisiert und fertig meldet
func RunProgress(title, description string, progressFn func(setPercent func(float64), setDone func())) {
	m := NewProgressModel(title, description)
	p := tea.NewProgram(m)

	go func() {
		time.Sleep(100 * time.Millisecond) // Warten, bis das Programm gestartet ist

		progressFn(
			func(percent float64) {
				p.Send(SetPercentMsg(percent))
			},
			func() {
				p.Send(FinishMsg(true))
			},
		)
	}()

	_, err := p.Run()
	if err != nil {
		fmt.Println("Fehler:", err)
	}
}

// RunProgressWithValues führt eine Fortschrittsanzeige mit Werten aus
func RunProgressWithValues(title, description string, maxValue int, valueUnit string, progressFn func(setValue func(int), setDone func())) {
	m := NewProgressModel(title, description)
	m.ShowValue = true
	m.MaxValue = maxValue
	m.ValueUnit = valueUnit
	p := tea.NewProgram(m)

	go func() {
		time.Sleep(100 * time.Millisecond) // Warten, bis das Programm gestartet ist

		progressFn(
			func(value int) {
				p.Send(SetValueMsg{Value: value, MaxValue: maxValue})
			},
			func() {
				p.Send(FinishMsg(true))
			},
		)
	}()

	_, err := p.Run()
	if err != nil {
		fmt.Println("Fehler:", err)
	}
}
