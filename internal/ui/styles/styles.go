package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors - Definiert die in der Anwendung verwendeten Farben
var (
	// Primary palette
	Primary   = lipgloss.Color("#9D5CFF") // Main brand color
	Secondary = lipgloss.Color("#C17BFF") // Lighter purple, complementary to Primary
	Accent    = lipgloss.Color("#FFCC00") // Accent color

	// Status colors
	Success = lipgloss.Color("#2ECC71") // Green
	Error   = lipgloss.Color("#E74C3C") // Red
	Warning = lipgloss.Color("#F39C12") // Orange
	Info    = lipgloss.Color("#5E9CF9") // Lighter blue, more compatible with purple

	// Neutral colors
	Text         = lipgloss.Color("#FFFFFF") // Main text
	SubtleText   = lipgloss.Color("#AAAAAA") // Secondary text
	Background   = lipgloss.Color("#111111") // Background
	Border       = lipgloss.Color("#333333") // Border
	Highlight    = lipgloss.Color("#444444") // Highlight
	Selected     = lipgloss.Color("#555555") // Selected item
	Inactive     = lipgloss.Color("#666666") // Inactive elements
	DisabledText = lipgloss.Color("#777777") // Disabled text

	// Link colors
	LinkNormal = lipgloss.Color("#E668FF") // Lighter purple for links
	LinkHover  = lipgloss.Color("#FF99FF") // Even lighter for hover state
)

// TextStyles - Basic text styles
var (
	// General text styles
	Normal = lipgloss.NewStyle().
		Foreground(Text)

	// Same as Normal but with a different name to maintain compatibility
	NormalText = lipgloss.NewStyle().
			Foreground(Text)

	Subtle = lipgloss.NewStyle().
		Foreground(SubtleText)

	Bold = lipgloss.NewStyle().
		Foreground(Text).
		Bold(true)

	// Same as Bold but with a different name to maintain compatibility
	BoldText = lipgloss.NewStyle().
			Foreground(Text).
			Bold(true)

	// Header styles
	H1 = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Padding(0, 0, 1, 0).
		MarginBottom(1)

	H2 = lipgloss.NewStyle().
		Foreground(Secondary).
		Bold(true).
		MarginBottom(1)

	// Status message styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	// Selection styles
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(SubtleText)

	// Container styles
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)

	// Highlight style
	HighlightStyle = lipgloss.NewStyle().
			Background(Highlight).
			Padding(0, 1)

	// Logo style
	Logo = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	// Link styles
	Link = lipgloss.NewStyle().
		Foreground(LinkNormal).
		Underline(true)

	LinkActive = lipgloss.NewStyle().
			Foreground(LinkHover).
			Underline(true).
			Bold(true)

	// Input styles
	InputPrompt = lipgloss.NewStyle().
			Foreground(Primary)
)

// UI Icons for command line interface
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconArrow   = "→"
	IconInfo    = "ℹ"
	IconBullet  = "•"

	// Icon constants with compatibility names
	SuccessIcon = IconSuccess
	ErrorIcon   = IconError
	WarningIcon = IconWarning
	ArrowIcon   = IconArrow
	InfoIcon    = IconInfo
)

// Logos and ASCII Art
const CLIKDLogo = `
   █████████  █████       █████ █████   ████ ██████████  
  ███░░░░░███░░███       ░░███ ░░███   ███░ ░░███░░░░███ 
 ███     ░░░  ░███        ░███  ░███  ███    ░███   ░░███
░███          ░███        ░███  ░███████     ░███    ░███
░███          ░███        ░███  ░███░░███    ░███    ░███
░░███     ███ ░███      █ ░███  ░███ ░░███   ░███    ███ 
 ░░█████████  ███████████ █████ █████ ░░████ ██████████  
  ░░░░░░░░░  ░░░░░░░░░░░ ░░░░░ ░░░░░   ░░░░ ░░░░░░░░░░   
`

// RenderLogo returns the rendered CLIKD ASCII logo
func RenderLogo() string {
	return Logo.Render(CLIKDLogo)
}

// Helper rendering functions

// InfoText renders text in the info style
func InfoText(text string) string {
	return InfoStyle.Render(text)
}

// ErrorText renders text in the error style
func ErrorText(text string) string {
	return ErrorStyle.Render(text)
}

// WarningText renders text in the warning style
func WarningText(text string) string {
	return WarningStyle.Render(text)
}

// SuccessText renders text in the success style
func SuccessText(text string) string {
	return SuccessStyle.Render(text)
}

// HighlightText renders text in the highlight style
func HighlightText(text string) string {
	return HighlightStyle.Render(text)
}

// LinkText renders text as a link
func LinkText(text string) string {
	return Link.Render(text)
}

// ActiveLinkText renders text as an active/hover link
func ActiveLinkText(text string) string {
	return LinkActive.Render(text)
}

// Layout helper functions
func CenterText(text string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, text)
}

// BoxedContent renders content in a box
func BoxedContent(content string) string {
	return Box.Render(content)
}

// SectionTitle renders a section title with decorative elements
func SectionTitle(title string) string {
	return H2.Render("=== " + title + " ===")
}

// RenderNormal renders text in normal style
func RenderNormal(text string) string {
	return Normal.Render(text)
}

// RenderSubtle renders text in subtle style
func RenderSubtle(text string) string {
	return Subtle.Render(text)
}

// RenderBold renders text in bold style
func RenderBold(text string) string {
	return Bold.Render(text)
}
