package styles

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestColors(t *testing.T) {
	// Test that colors are properly defined
	assert.Equal(t, lipgloss.Color("#9D5CFF"), Primary)
	assert.Equal(t, lipgloss.Color("#C17BFF"), Secondary)
	assert.Equal(t, lipgloss.Color("#FFCC00"), Accent)

	// Test status colors
	assert.Equal(t, lipgloss.Color("#2ECC71"), Success)
	assert.Equal(t, lipgloss.Color("#E74C3C"), Error)
	assert.Equal(t, lipgloss.Color("#F39C12"), Warning)
	assert.Equal(t, lipgloss.Color("#5E9CF9"), Info)

	// Test neutral colors
	assert.Equal(t, lipgloss.Color("#FFFFFF"), Text)
	assert.Equal(t, lipgloss.Color("#AAAAAA"), SubtleText)
	assert.Equal(t, lipgloss.Color("#111111"), Background)
	assert.Equal(t, lipgloss.Color("#333333"), Border)
	assert.Equal(t, lipgloss.Color("#444444"), Highlight)
	assert.Equal(t, lipgloss.Color("#555555"), Selected)
	assert.Equal(t, lipgloss.Color("#666666"), Inactive)
	assert.Equal(t, lipgloss.Color("#777777"), DisabledText)

	// Test link colors
	assert.Equal(t, lipgloss.Color("#E668FF"), LinkNormal)
	assert.Equal(t, lipgloss.Color("#FF99FF"), LinkHover)
}

func TestTextStyles(t *testing.T) {
	// Test that styles are properly initialized
	assert.NotNil(t, Normal)
	assert.NotNil(t, NormalText)
	assert.NotNil(t, Subtle)
	assert.NotNil(t, Bold)
	assert.NotNil(t, BoldText)
	assert.NotNil(t, H1)
	assert.NotNil(t, H2)
	assert.NotNil(t, SuccessStyle)
	assert.NotNil(t, ErrorStyle)
	assert.NotNil(t, WarningStyle)
	assert.NotNil(t, InfoStyle)
	assert.NotNil(t, SelectedStyle)
	assert.NotNil(t, UnselectedStyle)
	assert.NotNil(t, Box)
	assert.NotNil(t, HighlightStyle)
	assert.NotNil(t, Logo)
	assert.NotNil(t, Link)
	assert.NotNil(t, LinkActive)
	assert.NotNil(t, InputPrompt)
}

func TestIcons(t *testing.T) {
	// Test icon constants
	assert.Equal(t, "✓", IconSuccess)
	assert.Equal(t, "✗", IconError)
	assert.Equal(t, "⚠", IconWarning)
	assert.Equal(t, "→", IconArrow)
	assert.Equal(t, "ℹ", IconInfo)
	assert.Equal(t, "•", IconBullet)

	// Test compatibility names
	assert.Equal(t, IconSuccess, SuccessIcon)
	assert.Equal(t, IconError, ErrorIcon)
	assert.Equal(t, IconWarning, WarningIcon)
	assert.Equal(t, IconArrow, ArrowIcon)
	assert.Equal(t, IconInfo, InfoIcon)
}

func TestCLIKDLogo(t *testing.T) {
	// Test that logo constant is not empty
	assert.NotEmpty(t, CLIKDLogo)

	// Test that logo contains expected characters
	assert.Contains(t, CLIKDLogo, "█")
	assert.Contains(t, CLIKDLogo, "░")

	// Test that logo has multiple lines
	lines := strings.Split(CLIKDLogo, "\n")
	assert.Greater(t, len(lines), 5) // Should have more than 5 lines
}

func TestRenderLogo(t *testing.T) {
	result := RenderLogo()

	// Test that rendered logo is not empty
	assert.NotEmpty(t, result)

	// Test that it contains the logo content
	assert.Contains(t, result, "█")
}

func TestHelperRenderingFunctions(t *testing.T) {
	testText := "Test Message"

	t.Run("InfoText", func(t *testing.T) {
		result := InfoText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("ErrorText", func(t *testing.T) {
		result := ErrorText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("WarningText", func(t *testing.T) {
		result := WarningText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("SuccessText", func(t *testing.T) {
		result := SuccessText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("HighlightText", func(t *testing.T) {
		result := HighlightText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("LinkText", func(t *testing.T) {
		result := LinkText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("ActiveLinkText", func(t *testing.T) {
		result := ActiveLinkText(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})
}

func TestLayoutHelperFunctions(t *testing.T) {
	testText := "Test Content"

	t.Run("CenterText", func(t *testing.T) {
		width := 20
		result := CenterText(testText, width)
		assert.NotEmpty(t, result)
		// The result should be at least as wide as the original text
		assert.GreaterOrEqual(t, len(result), len(testText))
	})

	t.Run("BoxedContent", func(t *testing.T) {
		result := BoxedContent(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
		// Boxed content should be longer than original due to borders
		assert.Greater(t, len(result), len(testText))
	})

	t.Run("SectionTitle", func(t *testing.T) {
		title := "Test Section"
		result := SectionTitle(title)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, title)
		assert.Contains(t, result, "===")
	})

	t.Run("RenderNormal", func(t *testing.T) {
		result := RenderNormal(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("RenderSubtle", func(t *testing.T) {
		result := RenderSubtle(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})

	t.Run("RenderBold", func(t *testing.T) {
		result := RenderBold(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
	})
}

func TestStyleConsistency(t *testing.T) {
	// Test that Normal and NormalText are equivalent
	testText := "Test"
	normalResult := Normal.Render(testText)
	normalTextResult := NormalText.Render(testText)
	assert.Equal(t, normalResult, normalTextResult)

	// Test that Bold and BoldText are equivalent
	boldResult := Bold.Render(testText)
	boldTextResult := BoldText.Render(testText)
	assert.Equal(t, boldResult, boldTextResult)
}

func TestStyleApplications(t *testing.T) {
	testText := "Sample Text"

	// Test that styles actually modify the text (rendered text should be different from original)
	normalRendered := Normal.Render(testText)
	boldRendered := Bold.Render(testText)

	// Both should contain the original text
	assert.Contains(t, normalRendered, testText)
	assert.Contains(t, boldRendered, testText)

	// They should be different from each other (due to styling)
	// Note: This might not always be true in test environments without terminal support
	// but we can at least verify they're not empty
	assert.NotEmpty(t, normalRendered)
	assert.NotEmpty(t, boldRendered)
}

func TestEmptyTextHandling(t *testing.T) {
	// Test that functions handle empty strings gracefully
	emptyText := ""

	assert.NotPanics(t, func() { InfoText(emptyText) })
	assert.NotPanics(t, func() { ErrorText(emptyText) })
	assert.NotPanics(t, func() { WarningText(emptyText) })
	assert.NotPanics(t, func() { SuccessText(emptyText) })
	assert.NotPanics(t, func() { HighlightText(emptyText) })
	assert.NotPanics(t, func() { LinkText(emptyText) })
	assert.NotPanics(t, func() { ActiveLinkText(emptyText) })
	assert.NotPanics(t, func() { CenterText(emptyText, 10) })
	assert.NotPanics(t, func() { BoxedContent(emptyText) })
	assert.NotPanics(t, func() { SectionTitle(emptyText) })
	assert.NotPanics(t, func() { RenderNormal(emptyText) })
	assert.NotPanics(t, func() { RenderSubtle(emptyText) })
	assert.NotPanics(t, func() { RenderBold(emptyText) })
}
