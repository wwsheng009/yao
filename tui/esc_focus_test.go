package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestESCClearsInputFocus(t *testing.T) {
	cfg := &Config{
		Name:      "Test ESC Clears Focus",
		AutoFocus: true,
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

	// Initialize model
	windowMsg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, _ := model.Update(windowMsg)
	m := updatedModel.(*Model)

	initialFocus := m.CurrentFocus
	if initialFocus == "" {
		t.Error("Expected component to have focus after initialization")
	}

	escMsg := tea.KeyMsg{
		Type: tea.KeyEsc,
	}

	updatedModel, _ = m.Update(escMsg)
	m = updatedModel.(*Model)

	if m.CurrentFocus != "" {
		t.Errorf("Expected focus to be cleared after ESC, got %s", m.CurrentFocus)
	}
}
