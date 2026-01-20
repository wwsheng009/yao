package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/teatest"
)

func TestInputBlurBehavior(t *testing.T) {
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
