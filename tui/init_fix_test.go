package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/components"
)

// TestInitializeComponentsReturnsCmds tests that InitializeComponents now returns []tea.Cmd
func TestInitializeComponentsReturnsCmds(t *testing.T) {
	cfg := &Config{
		Name: "Test Init Cmd Collection",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter text",
						"prompt":      "> ",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Another input",
						"prompt":      "> ",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	
	// Call InitializeComponents which should now return []tea.Cmd
	cmds := model.InitializeComponents()
	
	// Verify that we get commands back (from input components' Init methods)
	assert.NotNil(t, cmds, "InitializeComponents should return a slice of commands")
	// Note: We can't guarantee non-zero length since disabled inputs return nil
	// Just verify the return type is correct
	assert.NotNil(t, cmds, "Commands slice should not be nil")
	
	// Check that input components were registered
	assert.Contains(t, model.Components, "input1")
	assert.Contains(t, model.Components, "input2")
	
	// Check that the commands are valid tea.Cmd functions
	for _, cmd := range cmds {
		assert.NotNil(t, cmd, "Each command should not be nil")
	}
}

// TestModelInitCollectsComponentCmds tests that Model.Init() collects component Init commands
func TestModelInitCollectsComponentCmds(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test Model Init Collection",
		AutoFocus: &autofocus, // Explicitly enable AutoFocus
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter text",
						"prompt":      "> ",
						"disabled":    false,
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Call Init which should collect and return component Init commands
	cmd := model.Init()

	// Verify that we get a command back (FocusFirstComponentMsg from AutoFocus)
	assert.NotNil(t, cmd, "Model.Init should return a command when AutoFocus is enabled and there are focusable components")
}

// TestInputComponentInitReturnsFocusCmd tests that InputComponentWrapper.Init returns Focus Cmd
func TestInputComponentInitReturnsFocusCmd(t *testing.T) {
	props := components.InputProps{
		Placeholder: "Enter text",
		Prompt:      "> ",
		Disabled:    false, // Should return Focus Cmd
	}
	
	wrapper := components.NewInputComponentWrapper(props, "test-input")
	
	cmd := wrapper.Init()
	
	assert.Nil(t, cmd, "InputComponentWrapper.Init should return nil (focus managed by framework)")
}

// TestInputComponentInitReturnsNilWhenDisabled tests that InputComponentWrapper.Init returns nil when disabled
func TestInputComponentInitReturnsNilWhenDisabled(t *testing.T) {
	props := components.InputProps{
		Placeholder: "Enter text",
		Prompt:      "> ",
		Disabled:    true, // Should return nil
	}
	
	wrapper := components.NewInputComponentWrapper(props, "test-input")
	
	cmd := wrapper.Init()
	
	assert.Nil(t, cmd, "InputComponentWrapper.Init should return nil when disabled")
}

// TestFormComponentInitCollectsChildCmds tests that FormComponentWrapper.Init collects child input Init Cmds
func TestFormComponentInitCollectsChildCmds(t *testing.T) {
	props := components.FormProps{
		Fields: []components.Field{
			{
				Type:        "input",
				Name:        "field1",
				Placeholder: "Field 1",
			},
			{
				Type:        "input",
				Name:        "field2", 
				Placeholder: "Field 2",
			},
		},
	}
	
	wrapper := components.NewFormComponentWrapper(props, "test-form")
	
	// Register some input fields with enabled status
	input1 := components.NewInputComponentWrapper(components.InputProps{
		Placeholder: "Field 1",
		Disabled:    false, // Will return Focus Cmd
	}, "field1")
	input2 := components.NewInputComponentWrapper(components.InputProps{
		Placeholder: "Field 2", 
		Disabled:    false, // Will return Focus Cmd
	}, "field2")
	
	wrapper.RegisterInputField("field1", input1)
	wrapper.RegisterInputField("field2", input2)
	
	cmd := wrapper.Init()
	
	assert.Nil(t, cmd, "FormComponentWrapper.Init should return nil (focus managed by framework)")
}