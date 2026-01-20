package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

func TestESCWithoutBindingClearsInputFocus(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test ESC Clears Focus",
		AutoFocus: &autofocus,
		LogLevel:  "trace",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "username-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter username",
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

	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(windowMsg)
	m := updatedModel.(*Model)

	t.Logf("After window init, CurrentFocus: %s", m.CurrentFocus)

	// Manually set focus to input component
	m.setFocus("username-input")

	if m.CurrentFocus != "username-input" {
		t.Error("Failed to set focus on component")
	}

	t.Logf("Before ESC, CurrentFocus: %s", m.CurrentFocus)

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(escMsg)
	m = updatedModel.(*Model)

	t.Logf("After ESC, CurrentFocus: %s", m.CurrentFocus)

	if m.CurrentFocus != "" {
		t.Errorf("Expected focus to be cleared after ESC, got %s", m.CurrentFocus)
	}
}

func TestESCWithQuitBinding(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test ESC With Quit Binding",
		AutoFocus: &autofocus,
		Bindings: map[string]core.Action{
			"esc": {
				Process: "tui.quit",
			},
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "username-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter username",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, cmd := model.Update(windowMsg)
	m := updatedModel.(*Model)

	if cmd != nil {
		msg := cmd()
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(*Model)
	}

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd = m.Update(escMsg)

	if cmd == nil {
		t.Error("Expected quit command when ESC binding is defined")
	}
}
