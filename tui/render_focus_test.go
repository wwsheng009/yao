package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/components"
	"github.com/yaoapp/yao/tui/core"
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

	// Verify LastFocusState is set correctly
	if !comp.LastFocusState {
		t.Error("Expected LastFocusState to be true")
	}

	// Store the initial cursor model state
	initialCursorState := inputWrapper.GetModel()

	// Simulate multiple renders by calling RenderComponent multiple times
	// This should NOT reset the cursor blink timer
	for i := 0; i < 10; i++ {
		// Call RenderComponent (this is what View() does internally)
		_ = model.RenderComponent(&model.Config.Layout.Children[0])
	}

	// Verify the cursor model hasn't been reset (i.e., Focus() wasn't called again)
	// We can't directly check the blink timer, but we can verify that
	// the component is still focused and LastFocusState hasn't changed
	if !inputWrapper.GetFocus() {
		t.Error("Input component lost focus after multiple renders")
	}

	if !comp.LastFocusState {
		t.Error("LastFocusState changed to false after multiple renders")
	}

	// Verify the cursor model reference hasn't changed (indicating Focus wasn't called again)
	if inputWrapper.GetModel() != initialCursorState {
		t.Error("Cursor model was reset, indicating Focus() was called again")
	}
}

// TestSetFocus_ClearFocus_updatesLastFocusState verifies that setFocus and clearFocus
// properly update the LastFocusState field.
func TestSetFocus_ClearFocus_updatesLastFocusState(t *testing.T) {
	// Create a test model with two input components
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
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
	model.InitializeComponents()

	// Initially, no component should be focused
	if model.CurrentFocus != "" {
		t.Error("Expected no focus initially")
	}

	// Set focus to input1
	model.setFocus("input1")

	// Verify focus state
	if model.CurrentFocus != "input1" {
		t.Error("Expected focus on input1")
	}

	comp1, exists := model.Components["input1"]
	if !exists {
		t.Fatal("input1 not found")
	}

	if !comp1.LastFocusState {
		t.Error("Expected LastFocusState of input1 to be true")
	}

	comp2, exists := model.Components["input2"]
	if !exists {
		t.Fatal("input2 not found")
	}

	if comp2.LastFocusState {
		t.Error("Expected LastFocusState of input2 to be false")
	}

	// Switch focus to input2
	model.setFocus("input2")

	// Verify focus state changed
	if model.CurrentFocus != "input2" {
		t.Error("Expected focus on input2")
	}

	if comp1.LastFocusState {
		t.Error("Expected LastFocusState of input1 to be false")
	}

	if !comp2.LastFocusState {
		t.Error("Expected LastFocusState of input2 to be true")
	}

	// Clear all focus
	model.clearFocus()

	// Verify no component is focused
	if model.CurrentFocus != "" {
		t.Error("Expected no focus after clearFocus")
	}

	if comp1.LastFocusState {
		t.Error("Expected LastFocusState of input1 to be false after clearFocus")
	}

	if comp2.LastFocusState {
		t.Error("Expected LastFocusState of input2 to be false after clearFocus")
	}
}

// TestUpdateInputFocusStates_avoidsRedundantCalls verifies that updateInputFocusStates
// only calls SetFocus when focus state actually changes.
func TestUpdateInputFocusStates_avoidsRedundantCalls(t *testing.T) {
	// Create a test model with two input components
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
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
	model.InitializeComponents()

	// Set focus to input1
	model.setFocus("input1")

	comp1, _ := model.Components["input1"]
	input1Wrapper := comp1.Instance.(*components.InputComponentWrapper)
	initialCursorState := input1Wrapper.GetModel()

	// Call updateInputFocusStates multiple times
	// This should NOT cause redundant SetFocus calls
	for i := 0; i < 10; i++ {
		model.updateInputFocusStates()
	}

	// Verify cursor model hasn't been reset
	if input1Wrapper.GetModel() != initialCursorState {
		t.Error("Cursor model was reset, indicating redundant SetFocus(true) was called")
	}

	// Verify LastFocusState is still true
	if !comp1.LastFocusState {
		t.Error("LastFocusState changed unexpectedly")
	}
}

// TestComponentInstance_LastFocusState_initialization verifies that LastFocusState
// is initialized to false for new component instances.
func TestComponentInstance_LastFocusState_initialization(t *testing.T) {
	instance := core.ComponentInstance{
		ID:       "test-component",
		Type:     "input",
		Instance: &components.InputComponentWrapper{},
	}

	// LastFocusState should default to false (zero value)
	if instance.LastFocusState {
		t.Error("Expected LastFocusState to be false (zero value) by default")
	}
}
