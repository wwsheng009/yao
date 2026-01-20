package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

	// Initialize components
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		model.Update(msg)
	}

	// Get window size
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(windowMsg)
	m := updatedModel.(*Model)

	// Set focus to input
	m.setFocus("test-input")
	t.Logf("After setFocus, CurrentFocus: %s", m.CurrentFocus)

	// Check if component has focus
	comp, exists := m.Components["test-input"]
	if !exists {
		t.Fatal("Component not found")
	}
	t.Logf("Component GetFocus() before ESC: %v", comp.Instance.GetFocus())

	// Send ESC key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(escMsg)
	m = updatedModel.(*Model)

	t.Logf("After ESC, CurrentFocus: %s", m.CurrentFocus)
	t.Logf("Component GetFocus() after ESC: %v", comp.Instance.GetFocus())

	if m.CurrentFocus != "" {
		t.Errorf("Expected CurrentFocus to be empty, got %s", m.CurrentFocus)
	}

	if comp.Instance.GetFocus() {
		t.Error("Expected component to not have focus")
	}
}
