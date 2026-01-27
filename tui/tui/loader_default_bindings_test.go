package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/core"
)

func TestDefaultBindings(t *testing.T) {
	// Create a config without any bindings
	cfg := &Config{
		Name: "Test Config",
		Data: map[string]interface{}{
			"message": "Hello World",
		},
		Layout: Layout{
			Direction: "vertical",
			Children:  []Component{},
		},
		// Note: No Bindings field set
	}

	// Simulate what happens in loadFile
	if cfg.Bindings == nil {
		cfg.Bindings = make(map[string]core.Action)
	}

	// Apply default bindings logic
	setMissingBinding(cfg.Bindings, "q", core.Action{Process: "tui.quit"})
	setMissingBinding(cfg.Bindings, "ctrl+c", core.Action{Process: "tui.quit"})
	setMissingBinding(cfg.Bindings, "tab", core.Action{Process: "tui.focus.next"})
	setMissingBinding(cfg.Bindings, "shift+tab", core.Action{Process: "tui.focus.prev"})
	setMissingBinding(cfg.Bindings, "enter", core.Action{Process: "tui.form.submit"})
	setMissingBinding(cfg.Bindings, "ctrl+r", core.Action{Process: "tui.refresh"})
	setMissingBinding(cfg.Bindings, "ctrl+l", core.Action{Process: "tui.refresh"})
	setMissingBinding(cfg.Bindings, "ctrl+z", core.Action{Process: "tui.suspend"})

	// Verify default bindings were added
	assert.Contains(t, cfg.Bindings, "q")
	assert.Contains(t, cfg.Bindings, "ctrl+c")
	assert.Contains(t, cfg.Bindings, "tab")
	assert.Contains(t, cfg.Bindings, "shift+tab")
	assert.Contains(t, cfg.Bindings, "enter")
	assert.Contains(t, cfg.Bindings, "ctrl+r")
	assert.Contains(t, cfg.Bindings, "ctrl+l")
	assert.Contains(t, cfg.Bindings, "ctrl+z")

	// Verify the actions are correct
	assert.Equal(t, "tui.quit", cfg.Bindings["q"].Process)
	assert.Equal(t, "tui.quit", cfg.Bindings["ctrl+c"].Process)
	assert.Equal(t, "tui.focus.next", cfg.Bindings["tab"].Process)
	assert.Equal(t, "tui.focus.prev", cfg.Bindings["shift+tab"].Process)
	assert.Equal(t, "tui.form.submit", cfg.Bindings["enter"].Process)
	assert.Equal(t, "tui.refresh", cfg.Bindings["ctrl+r"].Process)
	assert.Equal(t, "tui.refresh", cfg.Bindings["ctrl+l"].Process)
	assert.Equal(t, "tui.suspend", cfg.Bindings["ctrl+z"].Process)

	// Test that existing bindings are preserved
	existingAction := core.Action{Process: "custom.action"}
	cfg.Bindings["q"] = existingAction  // Override the default

	// Try to set default again - should not override
	setMissingBinding(cfg.Bindings, "q", core.Action{Process: "tui.quit"})

	// Should still have the custom action, not the default
	assert.Equal(t, "custom.action", cfg.Bindings["q"].Process)
}