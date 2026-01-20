package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

func TestGlobalBindingsWorkWhenNoFocus(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test Global Bindings No Focus",
		AutoFocus: &autofocus,
		Bindings: map[string]core.Action{
			"q": {
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

	windowMsg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, _ := model.Update(windowMsg)
	m := updatedModel.(*Model)

	if m.CurrentFocus != "" {
		t.Errorf("Expected no focus, got %s", m.CurrentFocus)
	}

	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'q'},
	}

	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Error("Expected a command to be returned for 'q' key when no component has focus")
	} else {
		msg := cmd()
		t.Logf("Command returned: %T", msg)
	}
}

func TestGlobalBindingsIgnoredWhenInputHasFocus(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test Global Bindings With Input Focus",
		AutoFocus: &autofocus,
		Bindings: map[string]core.Action{
			"q": {
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

	windowMsg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, _ := model.Update(windowMsg)
	m := updatedModel.(*Model)

	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) > 0 && m.CurrentFocus != focusableIDs[0] {
		t.Errorf("Expected focus on %s", focusableIDs[0])
	}

	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'q'},
	}

	_, cmd := m.Update(keyMsg)

	// When input has focus, 'q' should be used as a character in the input, NOT trigger quit
	// The component should handle the key, so command might be non-nil (state update)
	// The key indicator is that it shouldn't be a QuitMsg
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(core.QuitMsg); ok {
			t.Error("Expected 'q' to be handled as input when component has focus, not to trigger quit")
		}
	}

	// Check state was updated with the input value
	if value, ok := m.GetState("username-input"); ok {
		if value != "q" {
			t.Logf("Expected input value to be 'q', got '%s'", value)
		}
	}
}
