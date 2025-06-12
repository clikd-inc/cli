package bubble

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewInputModel tests input model creation
func TestNewInputModel(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		placeholder string
	}{
		{
			name:        "Basic input model",
			title:       "Enter Name",
			description: "Please enter your name",
			placeholder: "John Doe",
		},
		{
			name:        "Empty fields",
			title:       "",
			description: "",
			placeholder: "",
		},
		{
			name:        "Long content",
			title:       "Very Long Title That Should Be Handled Properly",
			description: "This is a very long description that should be handled properly by the input component",
			placeholder: "Very long placeholder text that should fit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewInputModel(tt.title, tt.description, tt.placeholder)

			assert.Equal(t, tt.title, model.Title)
			assert.Equal(t, tt.description, model.Description)
			assert.Equal(t, tt.placeholder, model.Placeholder)
			assert.Equal(t, 80, model.Width)
			assert.Equal(t, 156, model.CharLimit)
			assert.False(t, model.IsPassword)
			assert.Equal(t, tt.placeholder, model.TextInput.Placeholder)
			assert.True(t, model.TextInput.Focused())
		})
	}
}

// TestNewPasswordInputModel tests password input model creation
func TestNewPasswordInputModel(t *testing.T) {
	model := NewPasswordInputModel("Enter Password", "Please enter your password", "password123")

	assert.Equal(t, "Enter Password", model.Title)
	assert.Equal(t, "Please enter your password", model.Description)
	assert.Equal(t, "password123", model.Placeholder)
	assert.True(t, model.IsPassword)
	assert.Equal(t, textinput.EchoPassword, model.TextInput.EchoMode)
}

// TestInputModel_Init tests model initialization
func TestInputModel_Init(t *testing.T) {
	model := NewInputModel("Test", "Test description", "test")
	cmd := model.Init()

	// Should return textinput.Blink command
	assert.NotNil(t, cmd)
}

// TestInputModel_Update tests model updates
func TestInputModel_Update(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		msg           tea.Msg
		expectedValue string
		expectQuit    bool
		expectResult  bool
	}{
		{
			name:          "Enter key with text",
			initialValue:  "test input",
			msg:           tea.KeyMsg{Type: tea.KeyEnter},
			expectedValue: "test input",
			expectResult:  true,
		},
		{
			name:          "Enter key with empty text uses placeholder",
			initialValue:  "",
			msg:           tea.KeyMsg{Type: tea.KeyEnter},
			expectedValue: "default",
			expectResult:  true,
		},
		{
			name:       "Escape key quits",
			msg:        tea.KeyMsg{Type: tea.KeyEsc},
			expectQuit: true,
		},
		{
			name:       "Ctrl+C quits",
			msg:        tea.KeyMsg{Type: tea.KeyCtrlC},
			expectQuit: true,
		},
		{
			name: "Regular character input",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewInputModel("Test", "Test description", "default")

			// Set initial value if provided
			if tt.initialValue != "" {
				model.TextInput.SetValue(tt.initialValue)
			}

			updatedModel, cmd := model.Update(tt.msg)
			inputModel, ok := updatedModel.(InputModel)
			require.True(t, ok)

			if tt.expectQuit {
				// Should return quit command
				assert.NotNil(t, cmd)
				if cmd != nil {
					msg := cmd()
					_, isQuit := msg.(tea.QuitMsg)
					assert.True(t, isQuit)
				}
			} else if tt.expectResult {
				// Should return InputResultMsg
				assert.NotNil(t, cmd)
				if cmd != nil {
					msg := cmd()
					resultMsg, ok := msg.(InputResultMsg)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedValue, resultMsg.Value)
					assert.Equal(t, tt.expectedValue, inputModel.Value)
				}
			} else {
				// Regular input should update the text input
				// The exact behavior depends on the textinput component
			}
		})
	}
}

// TestInputModel_View tests view rendering
func TestInputModel_View(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		placeholder string
		expectTitle bool
		expectDesc  bool
		expectHelp  bool
	}{
		{
			name:        "Full input with all fields",
			title:       "Enter Name",
			description: "Please enter your name",
			placeholder: "John Doe",
			expectTitle: true,
			expectDesc:  true,
			expectHelp:  true,
		},
		{
			name:        "Input without title",
			title:       "",
			description: "Please enter your name",
			placeholder: "John Doe",
			expectTitle: false,
			expectDesc:  true,
			expectHelp:  true,
		},
		{
			name:        "Input without description",
			title:       "Enter Name",
			description: "",
			placeholder: "John Doe",
			expectTitle: true,
			expectDesc:  false,
			expectHelp:  true,
		},
		{
			name:        "Minimal input",
			title:       "",
			description: "",
			placeholder: "test",
			expectTitle: false,
			expectDesc:  false,
			expectHelp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewInputModel(tt.title, tt.description, tt.placeholder)
			view := model.View()

			if tt.expectTitle {
				assert.Contains(t, view, tt.title)
			} else if tt.title != "" {
				assert.NotContains(t, view, tt.title)
			}

			if tt.expectDesc {
				assert.Contains(t, view, tt.description)
			} else if tt.description != "" {
				assert.NotContains(t, view, tt.description)
			}

			if tt.expectHelp {
				assert.Contains(t, view, "Enter: Confirm")
				assert.Contains(t, view, "Esc: Cancel")
			}

			// Should always contain the text input view
			assert.NotEmpty(t, view)
		})
	}
}

// TestInputModel_PasswordView tests password input view
func TestInputModel_PasswordView(t *testing.T) {
	model := NewPasswordInputModel("Enter Password", "Please enter your password", "password")
	view := model.View()

	assert.Contains(t, view, "Enter Password")
	assert.Contains(t, view, "Please enter your password")
	assert.Contains(t, view, "Enter: Confirm")
	assert.Contains(t, view, "Esc: Cancel")
}

// TestInputModel_EdgeCases tests edge cases
func TestInputModel_EdgeCases(t *testing.T) {
	t.Run("Very long input", func(t *testing.T) {
		model := NewInputModel("Test", "Test", "test")
		longInput := strings.Repeat("a", 200) // Longer than char limit
		model.TextInput.SetValue(longInput)

		// Should handle long input gracefully
		view := model.View()
		assert.NotEmpty(t, view)
	})

	t.Run("Special characters in input", func(t *testing.T) {
		model := NewInputModel("Test", "Test", "test")
		specialInput := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		model.TextInput.SetValue(specialInput)

		// Should handle special characters
		view := model.View()
		assert.NotEmpty(t, view)
	})

	t.Run("Unicode characters", func(t *testing.T) {
		model := NewInputModel("Test", "Test", "test")
		unicodeInput := "Hello 世界 🌍 café"
		model.TextInput.SetValue(unicodeInput)

		// Should handle unicode characters
		view := model.View()
		assert.NotEmpty(t, view)
	})

	t.Run("Empty placeholder behavior", func(t *testing.T) {
		model := NewInputModel("Test", "Test", "")

		// Enter with empty input and empty placeholder
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		inputModel, ok := updatedModel.(InputModel)
		require.True(t, ok)

		if cmd != nil {
			msg := cmd()
			resultMsg, ok := msg.(InputResultMsg)
			if ok {
				assert.Equal(t, "", resultMsg.Value)
				assert.Equal(t, "", inputModel.Value)
			}
		}
	})
}

// TestInputModel_StateConsistency tests state consistency
func TestInputModel_StateConsistency(t *testing.T) {
	model := NewInputModel("Test", "Description", "placeholder")

	// Test multiple updates maintain consistency
	testInput := "test value"
	model.TextInput.SetValue(testInput)

	// Update with non-enter key
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	inputModel, ok := updatedModel.(InputModel)
	require.True(t, ok)

	// State should be preserved
	assert.Equal(t, "Test", inputModel.Title)
	assert.Equal(t, "Description", inputModel.Description)
	assert.Equal(t, "placeholder", inputModel.Placeholder)
	assert.Equal(t, 80, inputModel.Width)
	assert.Equal(t, 156, inputModel.CharLimit)
}

// TestInputModel_FocusState tests focus state
func TestInputModel_FocusState(t *testing.T) {
	model := NewInputModel("Test", "Test", "test")

	// Should start focused
	assert.True(t, model.TextInput.Focused())

	// View should render properly when focused
	view := model.View()
	assert.NotEmpty(t, view)
}
