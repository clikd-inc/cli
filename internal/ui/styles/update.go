package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Update notification styles
var (
	// Container style for update notification
	UpdateNotification = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(1).
				MarginTop(1).
				MarginBottom(1)

	// Update version style
	UpdateVersion = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	// Update indicator style
	UpdateIndicator = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// Update command style
	UpdateCommand = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Update URL style
	UpdateURL = lipgloss.NewStyle().
			Foreground(LinkNormal).
			Underline(true)
)
