package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Farben, die in der gesamten Anwendung verwendet werden
var (
	// Primäre Farben
	PrimaryColor   = lipgloss.Color("#9D5CFF") // Helles Magenta
	SecondaryColor = lipgloss.Color("#00B3E6") // Cyan
	AccentColor    = lipgloss.Color("#FFCC00") // Gelb/Gold

	// Statusfarben
	SuccessColor = lipgloss.Color("#2ECC71") // Grün
	ErrorColor   = lipgloss.Color("#E74C3C") // Rot
	WarningColor = lipgloss.Color("#F39C12") // Orange
	InfoColor    = lipgloss.Color("#3498DB") // Blau

	// Neutrale Farben
	TextColor         = lipgloss.Color("#FFFFFF")
	SubtleTextColor   = lipgloss.Color("#AAAAAA")
	BackgroundColor   = lipgloss.Color("#111111")
	BorderColor       = lipgloss.Color("#333333")
	HighlightColor    = lipgloss.Color("#444444")
	SelectedColor     = lipgloss.Color("#555555")
	InactiveColor     = lipgloss.Color("#666666")
	DisabledTextColor = lipgloss.Color("#777777")
)

// Styles für verschiedene UI-Elemente
var (
	// Allgemeine Stile
	NormalText = lipgloss.NewStyle().
			Foreground(TextColor)

	SubtleText = lipgloss.NewStyle().
			Foreground(SubtleTextColor)

	BoldText = lipgloss.NewStyle().
			Foreground(TextColor).
			Bold(true)

	// Header-Stile
	H1 = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Padding(0, 0, 1, 0).
		MarginBottom(1)

	H2 = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true).
		MarginBottom(1)

	// Hervorgehobene Elemente
	Highlight = lipgloss.NewStyle().
			Background(HighlightColor).
			Padding(0, 1)

	// Statusmeldungen
	Success = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	Error = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	Warning = lipgloss.NewStyle().
		Foreground(WarningColor)

	Info = lipgloss.NewStyle().
		Foreground(InfoColor)

	// Spezielle Hervorhebungen
	Selected = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	Unselected = lipgloss.NewStyle().
			Foreground(SubtleTextColor)

	// Boxen und Container
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1).
			MarginTop(1).
			MarginBottom(1)

	// Logo & Brand
	LogoStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)
)

// Logos und ASCII-Art
const CLIKDLogo = `
   ______   __     __   __            __  
  / ____/  / /    /  | / /           / /  
 / /      / /    / /||/ / ____ ___  / /__ 
/ /____  / /___ / / |  / / __// _ \/  __/ 
\____/ /_____//_/  |_/ /_/   \___/\__/   
`

// RenderLogo gibt das ASCII-Art-Logo für clikd zurück
func RenderLogo() string {
	return LogoStyle.Render(`
   ______   __     __   __            __  
  / ____/  / /    /  | / /           / /  
 / /      / /    / /||/ / ____ ___  / /__ 
/ /____  / /___ / / |  / / __// _ \/  __/ 
\____/ /_____//_/  |_/ /_/   \___/\__/   
                                          
`)
}

// Helperfunktionen
func ErrorText(text string) string {
	return Error.Render(text)
}

func SuccessText(text string) string {
	return Success.Render(text)
}

func WarningText(text string) string {
	return Warning.Render(text)
}

func InfoText(text string) string {
	return Info.Render(text)
}

func HighlightText(text string) string {
	return Highlight.Render(text)
}

// Einige Layout-Hilfsfunktionen
func CenterText(text string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, text)
}

func Box(content string) string {
	return BoxStyle.Render(content)
}

// SectionTitle rendert einen Abschnittstitel
func SectionTitle(title string) string {
	return H2.Render("=== " + title + " ===")
}
