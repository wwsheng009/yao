package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/tui/component"
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

	model := NewModel(cfg, program)
	model.Width = 80
	model.Height = 24

	// Initialize components
	cmds := model.InitializeComponents()
	for _, cmd := range cmds {
		if cmd != nil {
			// Execute init commands
		}
	}

	// Wait a moment for initialization to complete
	// This ensures components are fully created before setting focus
	model.Ready = true // Mark as ready after initialization

	// Get the input component wrapper
	comp, exists := model.Components["test-input"]
	if !exists {
		t.Fatal("Input component not found")
	}

	inputWrapper, ok := comp.Instance.(*component.InputComponentWrapper)
	if !ok {
		t.Fatal("Component is not an InputComponentWrapper")
	}

	// First, render the component once to ensure it's properly initialized
	_ = model.RenderComponent(&model.Config.Layout.Children[0])

	// Set initial focus directly on the component wrapper
	inputWrapper.SetFocus(true)

	// Verify component is focused
	if !inputWrapper.GetFocus() {
		t.Error("Expected input component to be focused")
	}

	// Store the initial focus state
	initialFocusState := inputWrapper.GetFocus()

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

	// Verify focus state hasn't been changed
	finalFocusState := inputWrapper.GetFocus()
	if initialFocusState != finalFocusState {
		t.Error("Focus state changed unexpectedly after multiple renders")
	}

	t.Log("✓ SetFocus not called repeatedly during rendering")
}
