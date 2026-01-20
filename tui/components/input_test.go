package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
)

func TestParseInputProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected InputProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: InputProps{
				Prompt: "> ", // default prompt
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"placeholder": "Enter text",
				"value":       "test value",
				"prompt":      "$ ",
				"color":       "red",
				"background":  "blue",
				"width":       20,
				"height":      1,
				"disabled":    true,
			},
			expected: InputProps{
				Placeholder: "Enter text",
				Value:       "test value",
				Prompt:      "$ ",
				Color:       "red",
				Background:  "blue",
				Width:       20,
				Height:      1,
				Disabled:    true,
			},
		},
		{
			name: "float width",
			props: map[string]interface{}{
				"width": 25.0,
			},
			expected: InputProps{
				Prompt: "> ", // default prompt
				Width:  25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInputProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderInput(t *testing.T) {
	props := InputProps{
		Placeholder: "Enter text",
		Value:       "test",
		Prompt:      "> ",
		Width:       20,
	}

	result := RenderInput(props, 80)
	assert.NotEmpty(t, result)
}

func TestNewInputComponentWrapper(t *testing.T) {
	props := InputProps{
		Placeholder: "Enter text",
		Value:       "test value",
		Prompt:      "$ ",
		Color:       "red",
		Background:  "blue",
		Width:       20,
		Height:      1,
		Disabled:    true,
	}

	wrapper := NewInputComponentWrapper(props, "test-id")

	// Test that configuration was applied correctly
	assert.Equal(t, "test-id", wrapper.GetID())
	assert.Equal(t, "test value", wrapper.GetValue())
	assert.Equal(t, "input", wrapper.GetComponentType())
	// Note: We can't directly access internal fields, but we can test behavior
	view := wrapper.View()
	assert.NotEmpty(t, view)
}

func TestInputComponentWrapperMethods(t *testing.T) {
	props := InputProps{
		Value: "initial value",
	}

	wrapper := NewInputComponentWrapper(props, "test-id")

	// Test GetID
	assert.Equal(t, "test-id", wrapper.GetID())

	// Test GetComponentType
	assert.Equal(t, "input", wrapper.GetComponentType())

	// Test Init
	cmd := wrapper.Init()
	assert.Nil(t, cmd)

	// Test View
	view := wrapper.View()
	assert.NotEmpty(t, view)

	// Test Render
	config := core.RenderConfig{}
	rendered, err := wrapper.Render(config)
	assert.NoError(t, err)
	assert.Equal(t, wrapper.View(), rendered)

	// Test SetFocus
	wrapper.SetFocus(true)
	assert.True(t, wrapper.GetFocus())

	wrapper.SetFocus(false)
	assert.False(t, wrapper.GetFocus())
}

func TestInputComponentWrapperUpdateBehavior(t *testing.T) {
	props := InputProps{
		Value: "test",
	}

	wrapper := NewInputComponentWrapper(props, "test-id")
	wrapper.SetFocus(true)

	// Test ESC key behavior through UpdateMsg
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedWrapper, _, _ := wrapper.UpdateMsg(escMsg)
	assert.NotNil(t, updatedWrapper)
	// ESC handling varies by implementation

	// Test character input
	charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedWrapper, _, _ = updatedWrapper.(*InputComponentWrapper).UpdateMsg(charMsg)
	assert.NotNil(t, updatedWrapper)
	// Character input handling varies by implementation
}

func TestInputComponentWrapper(t *testing.T) {
	props := InputProps{
		Value: "initial value",
	}

	wrapper := NewInputComponentWrapper(props, "test-id")

	// Test Init
	cmd := wrapper.Init()
	assert.Nil(t, cmd)

	// Test GetID
	assert.Equal(t, "test-id", wrapper.GetID())

	// Test GetComponentType
	assert.Equal(t, "input", wrapper.GetComponentType())

	// Test View
	view := wrapper.View()
	assert.NotEmpty(t, view)

	// Test GetValue and SetValue
	assert.Equal(t, "initial value", wrapper.GetValue())
	wrapper.SetValue("new value")
	assert.Equal(t, "new value", wrapper.GetValue())

	// Test SetFocus
	wrapper.SetFocus(true)
	// Note: We can't easily test focus state without accessing internal model directly

	// Test Render
	config := core.RenderConfig{}
	rendered, err := wrapper.Render(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, rendered)

	// Test Cleanup
	wrapper.Cleanup() // Just ensure it doesn't panic

	// Test GetStateChanges
	changes, hasChanges := wrapper.GetStateChanges()
	assert.True(t, hasChanges)
	assert.Contains(t, changes, "test-id")
	assert.Equal(t, "new value", changes["test-id"])
}

func TestInputComponentWrapperUpdateMsg(t *testing.T) {
	props := InputProps{
		Value: "initial value",
	}

	wrapper := NewInputComponentWrapper(props, "test-id")

	// Ensure the component has focus for testing
	wrapper.SetFocus(true)

	// Test ESC key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedWrapper, cmd, response := wrapper.UpdateMsg(escMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Ignored, response)

	// After ESC, component loses focus, need to re-focus for next tests
	updatedWrapperTyped := updatedWrapper.(*InputComponentWrapper)
	updatedWrapperTyped.SetFocus(true)

	// Test Enter key
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedWrapper, cmd, response = updatedWrapperTyped.UpdateMsg(enterMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response)

	// Update reference for next test
	updatedWrapperTyped = updatedWrapper.(*InputComponentWrapper)

	// Test Tab key
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedWrapper, cmd, response = updatedWrapperTyped.UpdateMsg(tabMsg)
	assert.NotNil(t, updatedWrapper)
	assert.Nil(t, cmd) // Tab shouldn't produce a command
	assert.Equal(t, core.Ignored, response)

	// Update reference for next test
	updatedWrapperTyped = updatedWrapper.(*InputComponentWrapper)

	// Test other key (should trigger value change)
	runesMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updatedWrapper, cmd, response = updatedWrapperTyped.UpdateMsg(runesMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response)

	// Update reference for next test
	updatedWrapperTyped = updatedWrapper.(*InputComponentWrapper)

	// Test targeted message
	targetedMsg := core.TargetedMsg{
		TargetID: "test-id",
		InnerMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
	}
	updatedWrapper, cmd, response = updatedWrapperTyped.UpdateMsg(targetedMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response)

	// Update reference for next test
	updatedWrapperTyped = updatedWrapper.(*InputComponentWrapper)

	// Test targeted message for different component (should be ignored)
	differentTargetMsg := core.TargetedMsg{
		TargetID: "different-id",
		InnerMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
	}
	updatedWrapper, cmd, response = updatedWrapperTyped.UpdateMsg(differentTargetMsg)
	assert.NotNil(t, updatedWrapper)
	assert.Nil(t, cmd)
	assert.Equal(t, core.Ignored, response)
}

func TestInputComponentWrapperUpdateMsgEdgeCases(t *testing.T) {
	props := InputProps{
		Value: "initial value",
	}

	// Test 1: Component without focus should ignore key messages
	wrapper := NewInputComponentWrapper(props, "test-no-focus")
	wrapper.SetFocus(false)

	// Test key press when not focused
	runesMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedWrapper, cmd, response := wrapper.UpdateMsg(runesMsg)
	assert.NotNil(t, updatedWrapper)
	assert.Nil(t, cmd) // Should not produce command when not focused
	assert.Equal(t, core.Ignored, response)

	// Test 2: Backspace key
	wrapper2 := NewInputComponentWrapper(props, "test-backspace")
	wrapper2.SetFocus(true)
	wrapper2.SetValue("test")

	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedWrapper2, cmd2, response2 := wrapper2.UpdateMsg(backspaceMsg)
	assert.NotNil(t, updatedWrapper2)
	assert.NotNil(t, cmd2)
	assert.Equal(t, core.Handled, response2)

	// Test 3: Ctrl+C should be ignored (special key not handled)
	ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedWrapper2, cmd2, response2 = wrapper2.UpdateMsg(ctrlCMsg)
	assert.NotNil(t, updatedWrapper2)
	// Ctrl+C is not a special key in HandleSpecialKey, so it should be handled
	// by delegateToBubbles which always returns a command
	assert.NotNil(t, cmd2)
	assert.Equal(t, core.Handled, response2)

	// Test 4: Non-key messages should be handled
	wrapper3 := NewInputComponentWrapper(props, "test-non-key")
	wrapper3.SetFocus(true)

	// Test with a non-key message (e.g., mouse message)
	mouseMsg := tea.MouseMsg{} // Empty mouse message
	updatedWrapper3, cmd3, response3 := wrapper3.UpdateMsg(mouseMsg)
	assert.NotNil(t, updatedWrapper3)
	assert.NotNil(t, cmd3) // delegateToBubbles should return a command
	assert.Equal(t, core.Handled, response3)

	// Test 5: State changes detection
	wrapper4 := NewInputComponentWrapper(props, "test-state")
	wrapper4.SetFocus(true)

	// Get initial state changes
	changes1, hasChanges1 := wrapper4.GetStateChanges()
	assert.True(t, hasChanges1)
	assert.Equal(t, "initial value", changes1["test-state"])

	// Update value
	wrapper4.SetValue("new value")
	changes2, hasChanges2 := wrapper4.GetStateChanges()
	assert.True(t, hasChanges2)
	assert.Equal(t, "new value", changes2["test-state"])
}

func TestInputComponentWrapperUpdateRenderConfig(t *testing.T) {
	props := InputProps{
		Value:       "initial value",
		Placeholder: "Enter text",
	}

	wrapper := NewInputComponentWrapper(props, "test-id")

	// Test valid config update
	newProps := map[string]interface{}{
		"value":       "updated value",
		"placeholder": "New placeholder",
		"prompt":      ">>> ",
	}
	config := core.RenderConfig{
		Data: newProps,
	}

	err := wrapper.UpdateRenderConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, "updated value", wrapper.GetValue())

	// Test invalid data type
	invalidConfig := core.RenderConfig{
		Data: "invalid data type",
	}
	err = wrapper.UpdateRenderConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data type")

	// Test empty props
	emptyProps := map[string]interface{}{}
	emptyConfig := core.RenderConfig{
		Data: emptyProps,
	}
	err = wrapper.UpdateRenderConfig(emptyConfig)
	assert.NoError(t, err)
}
