package initialize

import (
	"fmt"

	"clikd/pkg/config"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InitModel represents the main model for the initialization process
type InitModel struct {
	// Configuration options
	Global bool
	Force  bool
	Yes    bool

	// State variables
	CurrentStep      string
	ConfigPath       string
	ConfigDir        string
	IsInGitRepo      bool
	RepoURL          string
	ApiKeyStatus     string
	ConfigExists     bool
	Manager          *config.Manager
	AIEnabled        bool
	AIProvider       string
	AIModel          string
	AICustomSettings bool

	// UI components
	TextInput   textinput.Model
	Progress    progress.Model
	ActiveInput string
	Message     string
	MessageType string // "success", "error", "info", "warning"

	// Selection components
	SelectCursor  int
	SelectOptions []SelectOption

	// Results and errors
	Error error
	Done  bool
}

// SelectOption represents an option in a selection list
type SelectOption struct {
	Title       string
	Description string
	Value       interface{}
}

// NewInitModel creates a new initialization model
func NewInitModel(global, force, yes bool) InitModel {
	// Texteingabemodell mit klaren, konsistenten Stilen
	input := textinput.New()
	input.Placeholder = "Input..."
	input.CharLimit = 156
	input.Width = 60

	// Einfache, aber sichtbare Stilkonfiguration
	input.Prompt = "> "
	input.Cursor.Blink = true

	// Vereinfachter Stil für bessere Kompatibilität
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#5f87ff"))

	// Stelle sicher, dass der Cursor sichtbar ist und das Feld aktiv ist
	input.Focus()

	return InitModel{
		Global:       global,
		Force:        force,
		Yes:          yes,
		CurrentStep:  StepStart,
		ApiKeyStatus: "pending",
		Progress:     progress.New(progress.WithDefaultGradient(), progress.WithWidth(50)),
		TextInput:    input,
	}
}

// Init initializes the model
func (m InitModel) Init() tea.Cmd {
	return tea.Batch(
		checkGitRepo,
	)
}

// RunInitialization is the main function to start the initialization process
func RunInitialization(global, force, yes bool) error {
	model := NewInitModel(global, force, yes)

	// Create program with MouseCellMotion options for better mouse support
	// Remove AltScreen to avoid full-screen box around the UI
	p := tea.NewProgram(
		model,
		tea.WithMouseCellMotion(), // Enable better mouse support
	)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(InitModel)
	if !ok {
		return fmt.Errorf("error executing the model")
	}

	if m.Error != nil {
		return m.Error
	}

	return nil
}
