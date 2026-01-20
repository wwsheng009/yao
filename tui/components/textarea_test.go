package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/yaoapp/yao/tui/core"
)

func TestTextareaComponentWrapperHasFocus(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Initially not focused by default (focus is managed by TUI framework)
	assert.False(t, wrapper.GetFocus(), "Textarea should not be initially focused by default")

	// Add focus manually
	wrapper.SetFocus(true)
	assert.True(t, wrapper.GetFocus(), "Textarea should be focused after SetFocus(true)")

	// Remove focus
	wrapper.SetFocus(false)
	assert.False(t, wrapper.GetFocus(), "Textarea should not be focused after SetFocus(false)")

	// Add focus back
	wrapper.SetFocus(true)
	assert.True(t, wrapper.GetFocus(), "Textarea should be focused after SetFocus(true)")
}

func TestTextareaComponentWrapperUpdateMsg(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Test ESC key when focused
	wrapper.SetFocus(true) // 确保组件获得焦点
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedWrapper, cmd, response := wrapper.UpdateMsg(escMsg)

	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response) // ESC key is now handled by the component to manage focus
	updatedWrapperTyped := updatedWrapper.(*TextareaComponentWrapper)
	assert.False(t, updatedWrapperTyped.GetFocus(), "Textarea should lose focus on ESC")

	// Test that regular keys are handled when focused
	wrapper.SetFocus(true)
	runesMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_, cmd2, response2 := wrapper.UpdateMsg(runesMsg)

	assert.NotNil(t, cmd2, "Should return a command for key input")
	assert.Equal(t, core.Handled, response2)
}

func TestTextareaComponentWrapperGetValue(t *testing.T) {
	props := TextareaProps{
		Value:    "Initial value",
		Disabled: false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	assert.Equal(t, "Initial value", wrapper.GetValue(), "Should return initial value")

	// Update value
	wrapper.SetValue("Updated value")
	assert.Equal(t, "Updated value", wrapper.GetValue(), "Should return updated value")
}

func TestTextareaModelHasFocus(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	model := NewTextareaModel(props, "test-textarea")

	// Initially not focused by default (focus is managed by TUI framework)
	assert.False(t, model.GetFocus(), "Model should not be initially focused by default")

	// Add focus manually
	model.SetFocus(true)
	assert.True(t, model.GetFocus(), "Model should be focused after SetFocus(true)")

	// Remove focus
	model.SetFocus(false)
	assert.False(t, model.GetFocus(), "Model should not be focused after SetFocus(false)")

	// Add focus back
	model.SetFocus(true)
	assert.True(t, model.GetFocus(), "Model should be focused after SetFocus(true)")
}

func TestTextareaModelDisabled(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    true,
	}
	model := NewTextareaModel(props, "test-textarea")

	// Should not be focused when disabled
	assert.False(t, model.GetFocus(), "Model should not be focused when disabled")
}

// TestTextareaMultilineEditing tests multiline editing functionality
func TestTextareaMultilineEditing(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Set focus so the textarea can receive input
	wrapper.SetFocus(true)

	// Type first line
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	// Press Enter to create new line
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyEnter})

	// Type second line
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	value := wrapper.GetValue()

	// Should contain both lines
	assert.Contains(t, value, "Hello")
	assert.Contains(t, value, "World")
}

// TestTextareaDeleteOperations tests delete operations
func TestTextareaDeleteOperations(t *testing.T) {
	props := TextareaProps{
		Value:    "Hello World",
		Disabled: false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Set focus so the textarea can receive input
	wrapper.SetFocus(true)

	// Press backspace to delete character
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyBackspace})

	newValue := wrapper.GetValue()
	assert.NotContains(t, newValue, "Hello World")
}

// TestTextareaNavigationWithinText tests navigation within text
func TestTextareaNavigationWithinText(t *testing.T) {
	props := TextareaProps{
		Value:    "Hello World",
		Disabled: false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Set focus so the textarea can receive input
	wrapper.SetFocus(true)

	// Navigation should not panic
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyLeft})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRight})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})

	// Should still contain content
	assert.Contains(t, wrapper.GetValue(), "Hello World")
}

// TestTextareaPasteBehavior tests paste-like behavior (rapid typing)
func TestTextareaPasteBehavior(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Set focus so the textarea can receive input
	wrapper.SetFocus(true)

	// Simulate rapid typing (like paste)
	text := "Simulating pasted text"
	for _, char := range text {
		wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}

	value := wrapper.GetValue()
	assert.Contains(t, value, "Simulating pasted text")
}

// TestTextareaWrapBehavior tests text wrapping (if implemented)
func TestTextareaWrapBehavior(t *testing.T) {
	longText := "This is a very long line of text that might need to wrap at some point depending on the width of the textarea component in the terminal user interface"
	props := TextareaProps{
		Value:    longText,
		Disabled: false,
		Width:    20, // Short width
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Set focus so the textarea can receive input
	wrapper.SetFocus(true)

	// Render should not panic
	view := wrapper.View()
	assert.NotEmpty(t, view)
	// Check that the view contains some content from the long text
	assert.Contains(t, view, "This is a") // Check that some content is visible
}
