package tui

import (
	"testing"
)

func TestInitWithAutoFocus(t *testing.T) {
	autofocus := true
	cfg := &Config{
		Name:      "Test Init AutoFocus",
		AutoFocus: &autofocus,
		LogLevel:  "trace",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "First input",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Second input",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Init should return a command that sends FocusFirstComponentMsg
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Expected Init to return command, got nil")
	}

	t.Logf("Before processing Init commands, CurrentFocus: %s", model.CurrentFocus)
	t.Logf("Before processing Init commands, AutoFocusApplied: %v", model.AutoFocusApplied)

	// Process the command which should send FocusFirstComponentMsg
	msg := cmd()
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	t.Logf("After processing Init command, CurrentFocus: %s", m.CurrentFocus)
	t.Logf("After processing Init command, AutoFocusApplied: %v", m.AutoFocusApplied)

	// Verify focus was set to first component
	if m.CurrentFocus != "input1" {
		t.Errorf("Expected CurrentFocus to be 'input1', got '%s'", m.CurrentFocus)
	}

	// Verify AutoFocusApplied flag is set
	if !m.AutoFocusApplied {
		t.Error("Expected AutoFocusApplied to be true after Init")
	}
}

func TestInitWithoutAutoFocus(t *testing.T) {
	autofocus := false
	cfg := &Config{
		Name:      "Test Init No AutoFocus",
		AutoFocus: &autofocus,
		LogLevel:  "trace",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "First input",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Init should still return command (for component initialization)
	cmd := model.Init()

	if cmd != nil {
		// Process command
		msg := cmd()
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*Model)
	}

	t.Logf("After Init, CurrentFocus: %s", model.CurrentFocus)
	t.Logf("After Init, AutoFocusApplied: %v", model.AutoFocusApplied)

	// Verify no focus was set when AutoFocus is disabled
	if model.CurrentFocus != "" {
		t.Errorf("Expected CurrentFocus to be empty when AutoFocus=false, got '%s'", model.CurrentFocus)
	}

	// AutoFocusApplied should be true even when disabled (flag gets set regardless)
	t.Logf("AutoFocusApplied flag: %v", model.AutoFocusApplied)
}
