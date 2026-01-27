package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/gou/process"
)

func TestProcessFunctions(t *testing.T) {
	// Create a temporary model for testing
	cfg := &Config{Name: "TestProcess"}
	model := NewModel(cfg, nil)
	modelID := "test_model_for_process"
	RegisterModel(modelID, model)
	defer UnregisterModel(modelID) // Clean up after test
	
	// Test tui.exit process
	res := process.New("tui.exit", modelID).Run()
	assert.NotNil(t, res)
	result := res.(map[string]interface{})
	// Note: Since we're in a test environment, exit won't actually occur
	// Just verify that it doesn't panic and returns some result
	assert.Contains(t, result, "action")

	// Test tui.focus.next process
	res = process.New("tui.focus.next", modelID).Run()
	assert.NotNil(t, res)
	result = res.(map[string]interface{})
	assert.Equal(t, "focus_next", result["action"])

	// Test tui.focus.prev process
	res = process.New("tui.focus.prev", modelID).Run()
	assert.NotNil(t, res)
	result = res.(map[string]interface{})
	assert.Equal(t, "focus_prev", result["action"])

	// Test tui.form.submit process
	res = process.New("tui.form.submit", modelID).Run()
	assert.NotNil(t, res)
	result = res.(map[string]interface{})
	assert.Equal(t, "submit_form", result["action"])

	// Test tui.refresh process
	res = process.New("tui.refresh", modelID).Run()
	assert.NotNil(t, res)
	result = res.(map[string]interface{})
	assert.Equal(t, "refresh", result["action"])

	// Test tui.clear process
	res = process.New("tui.clear", modelID).Run()
	assert.NotNil(t, res)
	result = res.(map[string]interface{})
	assert.Equal(t, "clear", result["action"])

	// Test tui.suspend process
	res = process.New("tui.suspend", modelID).Run()
	assert.NotNil(t, res)
	result = res.(map[string]interface{})
	assert.Equal(t, "suspend", result["action"])
}