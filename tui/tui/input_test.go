package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/component"
	"github.com/yaoapp/yao/tui/teatest"
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
	inputWrapper, ok := comp.Instance.(*component.InputComponentWrapper)
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

	// Manually set focus to first input (auto-focus only happens via Init)
	model.setFocus("first-input")
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

func TestInputModelBlurBehavior(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test Input Blur",
		AutoFocus: &autofocus,
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "test-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Test input",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Initialize components - use teatest utility for proper batch processing
	cmd := model.Init()
	model = teatest.ProcessSequentialCmd(model, cmd).(*Model)

	// Set window size
	model = teatest.ProcessSequentialCmd(model, func() tea.Msg {
		return tea.WindowSizeMsg{Width: 80, Height: 24}
	}).(*Model)

	// Set focus to input (returns cmd)
	cmd = model.setFocus("test-input")
	model = teatest.ProcessSequentialCmd(model, cmd).(*Model)
	t.Logf("After setFocus and process, CurrentFocus: %s", model.CurrentFocus)

	// Check if component has focus
	comp, exists := model.Components["test-input"]
	if !exists {
		t.Fatal("Component not found")
	}
	t.Logf("Component GetFocus() before ESC: %v", comp.Instance.GetFocus())
	if !comp.Instance.GetFocus() {
		t.Error("Expected component to have focus after setFocus")
	}

	// Send ESC key - component should handle it and lose focus
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(escMsg)
	m := updatedModel.(*Model)

	// Process any returned command (should be FocusMsg if component implements)
	if cmd != nil {
		m = teatest.ProcessSequentialCmd(m, cmd).(*Model)
	}

	t.Logf("After ESC and process command, CurrentFocus: %s", m.CurrentFocus)
	t.Logf("Component GetFocus() after ESC: %v", comp.Instance.GetFocus())

	if m.CurrentFocus != "" {
		t.Errorf("Expected CurrentFocus to be empty, got %s", m.CurrentFocus)
	}

	if comp.Instance.GetFocus() {
		t.Error("Expected component to not have focus after ESC")
	}
}
