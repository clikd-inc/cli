package bubble

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"clikd/internal/ui/styles"
)

// ProgressModel is a model for a progress bar
type ProgressModel struct {
	Title       string
	Description string
	Progress    progress.Model
	Percent     float64
	Width       int
	Done        bool
	Callback    func(setPercent func(float64), setDone func())
}

// PercentMsg is used to update the progress percentage
type PercentMsg float64

// DoneMsg is used to indicate the progress is complete
type DoneMsg struct{}

// NewProgressModel creates a new progress model
func NewProgressModel(title, description string, width int) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(width),
	)

	return ProgressModel{
		Title:       title,
		Description: description,
		Progress:    p,
		Width:       width,
	}
}

// Init initializes the model
func (m ProgressModel) Init() tea.Cmd {
	if m.Callback != nil {
		return m.runCallback
	}
	return nil
}

// runCallback runs the callback function in a goroutine
func (m ProgressModel) runCallback() tea.Msg {
	go func() {
		if m.Callback == nil {
			return
		}

		m.Callback(
			func(p float64) {
				tea.NewProgram(m).Send(PercentMsg(p))
			},
			func() {
				tea.NewProgram(m).Send(DoneMsg{})
			},
		)
	}()

	return nil
}

// Update updates the model
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case PercentMsg:
		m.Percent = float64(msg)
		cmd := m.Progress.SetPercent(m.Percent)
		return m, cmd
	case DoneMsg:
		m.Done = true
		m.Percent = 1.0
		cmd := m.Progress.SetPercent(1.0)
		return m, tea.Sequence(cmd, tea.Quit)
	}

	return m, nil
}

// View renders the model
func (m ProgressModel) View() string {
	s := ""
	if m.Title != "" {
		s += styles.H2.Render(m.Title) + "\n\n"
	}

	if m.Description != "" {
		s += styles.Normal.Render(m.Description) + "\n\n"
	}

	s += m.Progress.View() + "\n\n"

	if m.Done {
		s += styles.SuccessStyle.Render("Complete!") + "\n"
	} else {
		s += styles.Subtle.Render("Ctrl+C: Cancel")
	}

	return s
}

// RunProgress displays a progress bar and executes a callback function
func RunProgress(title, description string, callback func(setPercent func(float64), setDone func())) {
	m := NewProgressModel(title, description, 50)
	m.Callback = callback
	p := tea.NewProgram(m)
	p.Run()
}
