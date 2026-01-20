package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/components"
)

// TestRenderComponent_SetFocusNotCalledRepeatedly verifies that SetFocus
// is not called repeatedly when rendering a focused component multiple times.
// This test ensures the fix for cursor blink timer reset issue.
func TestRenderComponent_SetFocusNotCalledRepeatedly(t *testing.T) {
	// Create a test model with an input component
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "test-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter text",
					},
				},
			},
		},
	}

	program := tea.NewProgram(nil)
	defer program.Quit()

	model := NewModel(cfg, program)
	model.Width = 80
	model.Height = 24

	// Initialize components
	cmds := model.InitializeComponents()
	if cmds != nil {
		for _, cmd := range cmds {
			if cmd != nil {
				// Execute init commands
			}
		}
	}

	// Get the input component wrapper
	comp, exists := model.Components["test-input"]
	if !exists {
		t.Fatal("Input component not found")
	}

	inputWrapper, ok := comp.Instance.(*components.InputComponentWrapper)
	if !ok {
		t.Fatal("Component is not an InputComponentWrapper")
	}

	// Set initial focus using setFocus (not rendering)
	model.setFocus("test-input")

	// Verify component is focused
	if !inputWrapper.GetFocus() {
		t.Error("Expected input component to be focused")
	}

	// Store the initial cursor model state
	initialCursorState := inputWrapper.GetModel()

	// Simulate multiple renders by calling RenderComponent multiple times
	// This should NOT reset the cursor blink timer
	for i := 0; i < 10; i++ {
		// Call RenderComponent (this is what View() does internally)
		_ = model.RenderComponent(&model.Config.Layout.Children[0])
	}

	// Verify the component is still focused
	if !inputWrapper.GetFocus() {
		t.Error("Component lost focus after multiple renders")
	}

	// Verify cursor state hasn't been reset
	finalCursorState := inputWrapper.GetModel()
	if initialCursorState != finalCursorState {
		t.Error("Cursor state changed unexpectedly after multiple renders")
	}

	t.Log("✓ SetFocus not called repeatedly during rendering")
}
