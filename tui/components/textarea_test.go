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

	// Initially focused
	assert.True(t, wrapper.HasFocus(), "Textarea should be initially focused")

	// Remove focus
	wrapper.SetFocus(false)
	assert.False(t, wrapper.HasFocus(), "Textarea should not be focused after SetFocus(false)")

	// Add focus back
	wrapper.SetFocus(true)
	assert.True(t, wrapper.HasFocus(), "Textarea should be focused after SetFocus(true)")
}

func TestTextareaComponentWrapperUpdateMsg(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Test ESC key when focused
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedWrapper, cmd, response := wrapper.UpdateMsg(escMsg)

	assert.NotNil(t, cmd)
	assert.Equal(t, core.Ignored, response)
	updatedWrapperTyped := updatedWrapper.(*TextareaComponentWrapper)
	assert.False(t, updatedWrapperTyped.HasFocus(), "Textarea should lose focus on ESC")

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

	// Initially focused (because !Disabled)
	assert.True(t, model.HasFocus(), "Model should be initially focused when not disabled")

	// Remove focus
	model.SetFocus(false)
	assert.False(t, model.HasFocus(), "Model should not be focused after SetFocus(false)")

	// Add focus back
	model.SetFocus(true)
	assert.True(t, model.HasFocus(), "Model should be focused after SetFocus(true)")
}

func TestTextareaModelDisabled(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    true,
	}
	model := NewTextareaModel(props, "test-textarea")

	// Should not be focused when disabled
	assert.False(t, model.HasFocus(), "Model should not be focused when disabled")
}
