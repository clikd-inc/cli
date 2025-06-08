package initialize

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// RunInitializationUI is the main function to start the initialization UI process
func RunInitializationUI(global, force, yes bool) error {
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
