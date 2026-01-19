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
				Placeholder:  "Enter text",
				Value:        "test value",
				Prompt:       "$ ",
				Color:        "red",
				Background:   "blue",
				Width:        20,
				Height:       1,
				Disabled:     true,
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

func TestNewInputModel(t *testing.T) {
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

	model := NewInputModel(props, "test-id")

	assert.Equal(t, props, model.props)
	assert.Equal(t, "test-id", model.id)
	assert.Equal(t, "test value", model.Value())
	assert.Equal(t, "$ ", model.Prompt)
	assert.Equal(t, 20, model.Width)
	assert.False(t, model.Focused()) // Should be disabled
}

func TestInputModelMethods(t *testing.T) {
	props := InputProps{
		Value: "initial value",
	}

	model := NewInputModel(props, "test-id")

	// Test GetID
	assert.Equal(t, "test-id", model.GetID())

	// Test GetComponentType
	assert.Equal(t, "input", model.GetComponentType())

	// Test Init
	cmd := model.Init()
	assert.Nil(t, cmd)

	// Test View
	view := model.View()
	assert.NotEmpty(t, view)

	// Test Render
	config := core.RenderConfig{}
	rendered, err := model.Render(config)
	assert.NoError(t, err)
	assert.Equal(t, model.View(), rendered)

	// Test SetFocus
	model.SetFocus(true)
	assert.True(t, model.Focused())

	model.SetFocus(false)
	assert.False(t, model.Focused())
}

func TestHandleInputUpdate(t *testing.T) {
	props := InputProps{
		Value: "test",
	}

	model := NewInputModel(props, "test-id")
	inputModelPtr := &model

	// Test nil input model
	resultModel, cmd := HandleInputUpdate(tea.KeyMsg{}, nil)
	assert.Equal(t, InputModel{}, resultModel)
	assert.Nil(t, cmd)

	// Test ESC key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	resultModel, cmd = HandleInputUpdate(escMsg, inputModelPtr)
	assert.NotNil(t, resultModel.Model)
	assert.False(t, resultModel.Focused())

	// Test other key (should pass through)
	otherMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	resultModel, cmd = HandleInputUpdate(otherMsg, inputModelPtr)
	assert.NotNil(t, resultModel.Model)
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

	// Test Enter key
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedWrapper, cmd, response = wrapper.UpdateMsg(enterMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response)

	// Test Tab key
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedWrapper, cmd, response = wrapper.UpdateMsg(tabMsg)
	assert.NotNil(t, updatedWrapper)
	assert.Nil(t, cmd) // Tab shouldn't produce a command
	assert.Equal(t, core.Ignored, response)

	// Test other key (should trigger value change)
	runesMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updatedWrapper, cmd, response = wrapper.UpdateMsg(runesMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response)

	// Test targeted message
	targetedMsg := core.TargetedMsg{
		TargetID: "test-id",
		InnerMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
	}
	updatedWrapper, cmd, response = wrapper.UpdateMsg(targetedMsg)
	assert.NotNil(t, updatedWrapper)
	assert.NotNil(t, cmd)
	assert.Equal(t, core.Handled, response)

	// Test targeted message for different component (should be ignored)
	differentTargetMsg := core.TargetedMsg{
		TargetID: "different-id",
		InnerMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
	}
	updatedWrapper, cmd, response = wrapper.UpdateMsg(differentTargetMsg)
	assert.NotNil(t, updatedWrapper)
	assert.Nil(t, cmd)
	assert.Equal(t, core.Ignored, response)
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