package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAutofocusConfiguration(t *testing.T) {
	cfg := &Config{
		Name:      "Test AutoFocus Enabled",
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

	windowMsg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, _ := model.Update(windowMsg)
	m := updatedModel.(*Model)

	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) > 0 {
		if m.CurrentFocus != focusableIDs[0] {
			t.Errorf("Expected focus on %s, got %s", focusableIDs[0], m.CurrentFocus)
		}
	}

	cfgDisabled := &Config{
		Name:      "Test AutoFocus Disabled",
		AutoFocus: false,
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

	modelDisabled := NewModel(cfgDisabled, nil)

	updatedModelDisabled, _ := modelDisabled.Update(windowMsg)
	mDisabled := updatedModelDisabled.(*Model)

	if mDisabled.CurrentFocus != "" {
		t.Errorf("Expected no focus, got %s", mDisabled.CurrentFocus)
	}
}
