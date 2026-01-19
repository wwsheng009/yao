package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/yaoapp/yao/tui/core"
)

func TestTextareaESCAndQuitKeys(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Initially focused
	assert.True(t, wrapper.HasFocus(), "Textarea should be initially focused")

	// Test ESC key - should return command and Ignored and blur
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedWrapper, cmd, response := wrapper.UpdateMsg(escMsg)

	assert.NotNil(t, cmd)
	assert.Equal(t, core.Ignored, response)
	updatedWrapperTyped := updatedWrapper.(*TextareaComponentWrapper)
	assert.False(t, updatedWrapperTyped.HasFocus(), "Textarea should lose focus on ESC")

	// Test that when blurred, 'q' key returns Ignored (allows global binding)
	wrapper.SetFocus(true) // Refocus
	assert.True(t, wrapper.HasFocus(), "Textarea should be focused")

	qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, _, response2 := wrapper.UpdateMsg(qMsg)

	// 'q' should be handled by textarea (it will add 'q' to the textarea value)
	assert.Equal(t, core.Handled, response2, "'q' should be handled by textarea when focused")
}

func TestTextAreaHandlesRegularKeys(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Test regular input keys
	oldValue := wrapper.GetValue()

	testKeys := []string{"h", "e", "l", "l", "o"}
	for _, key := range testKeys {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(key[0])}}
		_, cmd, response := wrapper.UpdateMsg(keyMsg)
		assert.NotNil(t, cmd, "Should return a command for key input")
		assert.Equal(t, core.Handled, response, "Should handle regular key input")
	}

	newValue := wrapper.GetValue()
	assert.Contains(t, newValue, "hello", "Textarea should contain 'hello'")
	assert.NotEqual(t, oldValue, newValue, "Value should have changed")
}

func TestTextAreaIgnoresKeysWhenBlurred(t *testing.T) {
	props := TextareaProps{
		Placeholder: "Enter text...",
		Disabled:    false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Blur the textarea
	wrapper.SetFocus(false)
	assert.False(t, wrapper.HasFocus(), "Textarea should not be focused")

	// Test that all keys are ignored when blurred
	oldValue := wrapper.GetValue()

	testKeys := []struct {
		msg  tea.KeyMsg
		desc string
	}{
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, "'a' key"},
		{tea.KeyMsg{Type: tea.KeyEsc}, "ESC key"},
		{tea.KeyMsg{Type: tea.KeyEnter}, "Enter key"},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, "'q' key"},
	}

	for _, test := range testKeys {
		_, _, response := wrapper.UpdateMsg(test.msg)
		assert.Equal(t, core.Ignored, response, "Should ignore all keys when blurred: "+test.desc)
	}

	newValue := wrapper.GetValue()
	assert.Equal(t, oldValue, newValue, "Value should not change when blurred")
}
