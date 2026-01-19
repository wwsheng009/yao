package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/components"
)

func TestInputComponent(t *testing.T) {
	cfg := &Config{
		Name: "Input Test",
		Data: map[string]interface{}{
			"username": "",
			"email":    "",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "username-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter username",
						"prompt":      "> ",
					},
				},
				{
					ID:   "email-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter email",
						"prompt":      "> ",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Initialize components first
	model.InitializeComponents()

	// Test that input models are created when components are rendered
	view1 := model.RenderComponent(&cfg.Layout.Children[0])
	view2 := model.RenderComponent(&cfg.Layout.Children[1])
	assert.NotEmpty(t, view1)
	assert.NotEmpty(t, view2)

	// Check that input components were created
	assert.Contains(t, model.Components, "username-input")
	assert.Contains(t, model.Components, "email-input")
	assert.Equal(t, "input", model.Components["username-input"].Type)
	assert.Equal(t, "input", model.Components["email-input"].Type)

	// Test input value update - need to get the wrapper
	comp := model.Components["username-input"]
	inputWrapper, ok := comp.Instance.(*components.InputComponentWrapper)
	assert.True(t, ok, "Expected InputComponentWrapper")
	inputWrapper.SetValue("testuser")

	// Simulate updating state with input value
	model.StateMu.Lock()
	model.State["username-input"] = inputWrapper.GetValue()
	model.StateMu.Unlock()

	// Verify state was updated
	value, exists := model.getStateValue("username-input")
	assert.True(t, exists)
	assert.Equal(t, "testuser", value)
}

func TestInputNavigation(t *testing.T) {
	cfg := &Config{
		Name: "Input Navigation Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "first-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "First input",
						"prompt":      "> ",
					},
				},
				{
					ID:   "second-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Second input",
						"prompt":      "> ",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Initialize components first
	model.InitializeComponents()

	// Manually render the components to initialize input models
	for i := range cfg.Layout.Children {
		model.RenderComponent(&cfg.Layout.Children[i])
	}

	// Initially, first input should be focused
	assert.Equal(t, "first-input", model.CurrentFocus)

	// Simulate tabbing to next input
	model.focusNextInput()
	assert.Equal(t, "second-input", model.CurrentFocus)

	// Tab again should wrap to first
	model.focusNextInput()
	assert.Equal(t, "first-input", model.CurrentFocus)
}

// TestHandleInputUpdate has been removed as InputModel and HandleInputUpdate are deprecated
// Use TestInputComponentWrapperUpdateBehavior instead
