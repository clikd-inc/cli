package bubble

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSelectModel tests select model creation
func TestNewSelectModel(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		items         []SelectItem
		expectPreview bool
	}{
		{
			name:  "Basic select model",
			title: "Choose Option",
			items: []SelectItem{
				{Title: "Option 1", Description: "First option", Value: "opt1"},
				{Title: "Option 2", Description: "Second option", Value: "opt2"},
			},
			expectPreview: false,
		},
		{
			name:  "Select with preview items",
			title: "Choose with Preview",
			items: []SelectItem{
				{Title: "Option 1", Description: "First option", Value: "opt1", Preview: "# Preview 1\nThis is preview content"},
				{Title: "Option 2", Description: "Second option", Value: "opt2"},
			},
			expectPreview: true,
		},
		{
			name:          "Empty items",
			title:         "Empty Select",
			items:         []SelectItem{},
			expectPreview: false,
		},
		{
			name:  "Single item",
			title: "Single Choice",
			items: []SelectItem{
				{Title: "Only Option", Description: "The only choice", Value: "only"},
			},
			expectPreview: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSelectModel(tt.title, tt.items)

			assert.Equal(t, tt.title, model.Title)
			assert.Equal(t, tt.items, model.Items)
			assert.Equal(t, 0, model.Cursor)
			assert.Nil(t, model.Selected)
			assert.Equal(t, 80, model.Width)
			assert.True(t, model.Description)
			assert.Equal(t, tt.expectPreview, model.ShowPreview)
			assert.False(t, model.InPreview)
			assert.Nil(t, model.PreviewModel)
		})
	}
}

// TestSelectModel_Init tests model initialization
func TestSelectModel_Init(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Value: "opt1"},
	}
	model := NewSelectModel("Test", items)
	cmd := model.Init()

	// Should return nil (no initialization command needed)
	assert.Nil(t, cmd)
}

// TestSelectModel_Navigation tests cursor navigation
func TestSelectModel_Navigation(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Value: "opt1"},
		{Title: "Option 2", Value: "opt2"},
		{Title: "Option 3", Value: "opt3"},
	}
	model := NewSelectModel("Test", items)

	tests := []struct {
		name           string
		key            tea.KeyMsg
		expectedCursor int
	}{
		{
			name:           "Down arrow moves cursor down",
			key:            tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "j key moves cursor down",
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedCursor: 1,
		},
		{
			name:           "Up arrow moves cursor up",
			key:            tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0, // Should stay at 0 (top)
		},
		{
			name:           "k key moves cursor up",
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedCursor: 0, // Should stay at 0 (top)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cursor to 0
			model.Cursor = 0

			updatedModel, cmd := model.Update(tt.key)
			selectModel, ok := updatedModel.(SelectModel)
			require.True(t, ok)

			assert.Equal(t, tt.expectedCursor, selectModel.Cursor)
			assert.Nil(t, cmd) // Navigation shouldn't return commands
		})
	}
}

// TestSelectModel_NavigationBounds tests navigation boundaries
func TestSelectModel_NavigationBounds(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Value: "opt1"},
		{Title: "Option 2", Value: "opt2"},
		{Title: "Option 3", Value: "opt3"},
	}
	model := NewSelectModel("Test", items)

	// Test moving down to the last item
	model.Cursor = 2 // Last item
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectModel, ok := updatedModel.(SelectModel)
	require.True(t, ok)
	assert.Equal(t, 2, selectModel.Cursor) // Should stay at last item

	// Test moving up from first item
	model.Cursor = 0 // First item
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	selectModel, ok = updatedModel.(SelectModel)
	require.True(t, ok)
	assert.Equal(t, 0, selectModel.Cursor) // Should stay at first item
}

// TestSelectModel_Selection tests item selection
func TestSelectModel_Selection(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Description: "First option", Value: "opt1"},
		{Title: "Option 2", Description: "Second option", Value: "opt2"},
	}
	model := NewSelectModel("Test", items)
	model.Cursor = 1 // Select second item

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{
			name: "Enter key selects item",
			key:  tea.KeyMsg{Type: tea.KeyEnter},
		},
		{
			name: "Space key selects item",
			key:  tea.KeyMsg{Type: tea.KeySpace},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset selection
			model.Selected = nil

			updatedModel, cmd := model.Update(tt.key)
			selectModel, ok := updatedModel.(SelectModel)
			require.True(t, ok)

			// Should have selected the item at cursor position
			assert.NotNil(t, selectModel.Selected)
			assert.Equal(t, &items[1], selectModel.Selected)

			// Should return SelectResultMsg command
			assert.NotNil(t, cmd)
			if cmd != nil {
				msg := cmd()
				resultMsg, ok := msg.(SelectResultMsg)
				assert.True(t, ok)
				assert.Equal(t, "opt2", resultMsg.Value)
			}
		})
	}
}

// TestSelectModel_Quit tests quit functionality
func TestSelectModel_Quit(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Value: "opt1"},
	}
	model := NewSelectModel("Test", items)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{
			name: "q key quits",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
		},
		{
			name: "Escape key quits",
			key:  tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name: "Ctrl+C quits",
			key:  tea.KeyMsg{Type: tea.KeyCtrlC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := model.Update(tt.key)
			_, ok := updatedModel.(SelectModel)
			require.True(t, ok)

			// Should return quit command
			assert.NotNil(t, cmd)
			if cmd != nil {
				msg := cmd()
				_, isQuit := msg.(tea.QuitMsg)
				assert.True(t, isQuit)
			}
		})
	}
}

// TestSelectModel_Preview tests preview functionality
func TestSelectModel_Preview(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Value: "opt1", Preview: "# Preview 1\nThis is preview content"},
		{Title: "Option 2", Value: "opt2"}, // No preview
	}
	model := NewSelectModel("Test", items)

	t.Run("Enter preview mode with p key", func(t *testing.T) {
		model.Cursor = 0 // Item with preview

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		selectModel, ok := updatedModel.(SelectModel)
		require.True(t, ok)

		assert.True(t, selectModel.InPreview)
		assert.NotNil(t, selectModel.PreviewModel)
		assert.Nil(t, cmd) // Entering preview shouldn't return command
	})

	t.Run("p key with no preview does nothing", func(t *testing.T) {
		model.Cursor = 1 // Item without preview
		model.InPreview = false
		model.PreviewModel = nil

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		selectModel, ok := updatedModel.(SelectModel)
		require.True(t, ok)

		assert.False(t, selectModel.InPreview)
		assert.Nil(t, selectModel.PreviewModel)
		assert.Nil(t, cmd)
	})

	t.Run("Exit preview mode", func(t *testing.T) {
		// Set up preview mode
		model.InPreview = true
		previewModel := NewPreviewModel("Test Preview", "# Test\nContent")
		model.PreviewModel = &previewModel

		exitKeys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'q'}},
			{Type: tea.KeyEsc},
			{Type: tea.KeyEnter},
			{Type: tea.KeySpace},
		}

		for _, key := range exitKeys {
			// Reset preview state
			model.InPreview = true
			model.PreviewModel = &previewModel

			updatedModel, cmd := model.Update(key)
			selectModel, ok := updatedModel.(SelectModel)
			require.True(t, ok)

			assert.False(t, selectModel.InPreview)
			assert.Nil(t, selectModel.PreviewModel)
			assert.Nil(t, cmd)
		}
	})
}

// TestSelectModel_View tests view rendering
func TestSelectModel_View(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		items         []SelectItem
		cursor        int
		expectTitle   bool
		expectItems   bool
		expectHelp    bool
		expectPreview bool
	}{
		{
			name:  "Basic view with title and items",
			title: "Choose Option",
			items: []SelectItem{
				{Title: "Option 1", Description: "First option", Value: "opt1"},
				{Title: "Option 2", Description: "Second option", Value: "opt2"},
			},
			cursor:        0,
			expectTitle:   true,
			expectItems:   true,
			expectHelp:    true,
			expectPreview: false,
		},
		{
			name:  "View with preview items",
			title: "Choose with Preview",
			items: []SelectItem{
				{Title: "Option 1", Description: "First option", Value: "opt1", Preview: "# Preview"},
				{Title: "Option 2", Description: "Second option", Value: "opt2"},
			},
			cursor:        0,
			expectTitle:   true,
			expectItems:   true,
			expectHelp:    true,
			expectPreview: true,
		},
		{
			name:  "View without title",
			title: "",
			items: []SelectItem{
				{Title: "Option 1", Value: "opt1"},
			},
			cursor:        0,
			expectTitle:   false,
			expectItems:   true,
			expectHelp:    true,
			expectPreview: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSelectModel(tt.title, tt.items)
			model.Cursor = tt.cursor
			view := model.View()

			if tt.expectTitle && tt.title != "" {
				assert.Contains(t, view, tt.title)
			}

			if tt.expectItems {
				for _, item := range tt.items {
					assert.Contains(t, view, item.Title)
					if item.Description != "" {
						assert.Contains(t, view, item.Description)
					}
				}
			}

			if tt.expectHelp {
				assert.Contains(t, view, "Navigate")
				assert.Contains(t, view, "Select")
			}

			if tt.expectPreview {
				assert.Contains(t, view, "Preview")
			}

			assert.NotEmpty(t, view)
		})
	}
}

// TestSelectModel_PreviewView tests preview view rendering
func TestSelectModel_PreviewView(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Value: "opt1", Preview: "# Preview Content\nThis is a test"},
	}
	model := NewSelectModel("Test", items)

	// Enter preview mode
	model.InPreview = true
	previewModel := NewPreviewModel("Preview: Option 1", items[0].Preview)
	model.PreviewModel = &previewModel

	view := model.View()

	// Should show preview content, not selection list
	// Note: The preview title might contain the original item title
	assert.NotEmpty(t, view)                                                   // Should have preview content
	assert.Contains(t, view, "Preview")                                        // Should contain preview text (rendered by Glamour)
	assert.Contains(t, view, "This is a")                                      // Should contain part of the actual content (Glamour may split text)
	assert.Contains(t, view, "test")                                           // Should contain the rest of the content
	assert.Contains(t, view, "Position: Complete • h: Help • Enter/Esc: Back") // Should show preview navigation
}

// TestSelectModel_EdgeCases tests edge cases
func TestSelectModel_EdgeCases(t *testing.T) {
	t.Run("Empty items list", func(t *testing.T) {
		model := NewSelectModel("Test", []SelectItem{})
		view := model.View()

		assert.Contains(t, view, "Test") // Title should still appear
		assert.NotEmpty(t, view)
	})

	t.Run("Very long item titles", func(t *testing.T) {
		longTitle := "This is a very long title that should be handled properly by the select component without breaking the layout"
		items := []SelectItem{
			{Title: longTitle, Description: "Long description", Value: "long"},
		}
		model := NewSelectModel("Test", items)
		view := model.View()

		assert.Contains(t, view, longTitle)
		assert.NotEmpty(t, view)
	})

	t.Run("Special characters in items", func(t *testing.T) {
		items := []SelectItem{
			{Title: "Option with émojis 🎉", Description: "Special chars: !@#$%", Value: "special"},
		}
		model := NewSelectModel("Test", items)
		view := model.View()

		assert.Contains(t, view, "🎉")
		assert.Contains(t, view, "!@#$%")
		assert.NotEmpty(t, view)
	})

	t.Run("Navigation with single item", func(t *testing.T) {
		items := []SelectItem{
			{Title: "Only Option", Value: "only"},
		}
		model := NewSelectModel("Test", items)

		// Try to move down
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		selectModel, ok := updatedModel.(SelectModel)
		require.True(t, ok)
		assert.Equal(t, 0, selectModel.Cursor) // Should stay at 0

		// Try to move up
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		selectModel, ok = updatedModel.(SelectModel)
		require.True(t, ok)
		assert.Equal(t, 0, selectModel.Cursor) // Should stay at 0
	})
}

// TestSelectModel_StateConsistency tests state consistency
func TestSelectModel_StateConsistency(t *testing.T) {
	items := []SelectItem{
		{Title: "Option 1", Description: "First", Value: "opt1"},
		{Title: "Option 2", Description: "Second", Value: "opt2"},
	}
	model := NewSelectModel("Test", items)

	// Multiple updates should maintain state consistency
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectModel, ok := updatedModel.(SelectModel)
	require.True(t, ok)

	// State should be preserved
	assert.Equal(t, "Test", selectModel.Title)
	assert.Equal(t, items, selectModel.Items)
	assert.Equal(t, 80, selectModel.Width)
	assert.True(t, selectModel.Description)
	assert.Equal(t, 1, selectModel.Cursor) // Should have moved down
}
