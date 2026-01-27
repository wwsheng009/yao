package tui

import (
	"testing"

	"github.com/yaoapp/yao/tui/teatest"
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

	// Process the command (which might be a batch) using teatest helper
	msgs := teatest.ExecuteBatchCommand(cmd)
	for _, msg := range msgs {
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*Model)
	}

	t.Logf("After processing Init command, CurrentFocus: %s", model.CurrentFocus)
	t.Logf("After processing Init command, AutoFocusApplied: %v", model.AutoFocusApplied)

	// Verify focus was set to first component
	if model.CurrentFocus != "input1" {
		t.Errorf("Expected CurrentFocus to be 'input1', got '%s'", model.CurrentFocus)
	}

	// Verify AutoFocusApplied flag is set
	if !model.AutoFocusApplied {
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
		// Process command (which might be a batch) using teatest helper
		msgs := teatest.ExecuteBatchCommand(cmd)
		for _, msg := range msgs {
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(*Model)
		}
	}

	t.Logf("After Init, CurrentFocus: %s", model.CurrentFocus)
	t.Logf("After Init, AutoFocusApplied: %v", model.AutoFocusApplied)

	// Verify no focus was set when AutoFocus is disabled
	if model.CurrentFocus != "" {
		t.Errorf("Expected CurrentFocus to be empty when AutoFocus=false, got '%s'", model.CurrentFocus)
	}

	// AutoFocusApplied should be true even when disabled (flag gets set regardless)
	// Wait, checking the implementation in message_handlers.go:
	// handlers["FocusFirstComponentMsg"] only sets AutoFocusApplied=true IF it actually applies focus.
	// So if AutoFocus is false, AutoFocusApplied might remain false or depend on logic.
	// Let's check if the test expects true or false. Original test implied expectation.
	// Actually, if AutoFocus is false, FocusFirstComponentMsg is NOT sent in Init().
	// So Update handler is never called for it.
	// So AutoFocusApplied should be false.
	if model.AutoFocusApplied {
		t.Errorf("Expected AutoFocusApplied to be false when AutoFocus=false (commands not sent), got %v", model.AutoFocusApplied)
	}
}
