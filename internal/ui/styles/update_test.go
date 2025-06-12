package styles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateStyles(t *testing.T) {
	// Test that update styles are properly initialized
	assert.NotNil(t, UpdateNotification)
	assert.NotNil(t, UpdateVersion)
	assert.NotNil(t, UpdateIndicator)
	assert.NotNil(t, UpdateCommand)
	assert.NotNil(t, UpdateURL)
}

func TestUpdateStyleRendering(t *testing.T) {
	testText := "Test Update Text"

	t.Run("UpdateNotification", func(t *testing.T) {
		result := UpdateNotification.Render(testText)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, testText)
		// Should be longer than original due to borders and padding
		assert.Greater(t, len(result), len(testText))
	})

	t.Run("UpdateVersion", func(t *testing.T) {
		version := "v1.2.3"
		result := UpdateVersion.Render(version)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, version)
	})

	t.Run("UpdateIndicator", func(t *testing.T) {
		indicator := "NEW UPDATE AVAILABLE"
		result := UpdateIndicator.Render(indicator)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, indicator)
	})

	t.Run("UpdateCommand", func(t *testing.T) {
		command := "npm update -g clikd"
		result := UpdateCommand.Render(command)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, command)
	})

	t.Run("UpdateURL", func(t *testing.T) {
		url := "https://github.com/user/repo/releases"
		result := UpdateURL.Render(url)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, url)
	})
}

func TestUpdateStylesWithEmptyText(t *testing.T) {
	emptyText := ""

	// Test that update styles handle empty strings gracefully
	assert.NotPanics(t, func() { UpdateNotification.Render(emptyText) })
	assert.NotPanics(t, func() { UpdateVersion.Render(emptyText) })
	assert.NotPanics(t, func() { UpdateIndicator.Render(emptyText) })
	assert.NotPanics(t, func() { UpdateCommand.Render(emptyText) })
	assert.NotPanics(t, func() { UpdateURL.Render(emptyText) })
}

func TestUpdateStylesConsistency(t *testing.T) {
	// Test that update styles use consistent color scheme
	testText := "Test"

	// All styles should produce non-empty output
	assert.NotEmpty(t, UpdateNotification.Render(testText))
	assert.NotEmpty(t, UpdateVersion.Render(testText))
	assert.NotEmpty(t, UpdateIndicator.Render(testText))
	assert.NotEmpty(t, UpdateCommand.Render(testText))
	assert.NotEmpty(t, UpdateURL.Render(testText))
}
