package tui

import (
	"encoding/json"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/ui/components"
)

// TestInputInRuntimeMode tests input component with runtime mode
func TestInputInRuntimeMode(t *testing.T) {
	jsonConfig := `{
		"name": "Input Test",
		"useRuntime": true,
		"data": {
			"message": "Enter text:"
		},
		"layout": {
			"direction": "vertical",
			"children": [
				{
					"type": "text",
					"props": {
						"content": "{{message}}"
					}
				},
				{
					"id": "my-input",
					"type": "input",
					"props": {
						"placeholder": "Type here...",
						"width": 30
					}
				}
			]
		}
	}`

	var config Config
	if err := json.Unmarshal([]byte(jsonConfig), &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	model := NewModel(&config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Trigger rendering
	_ = model.View()

	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	t.Logf("Focus list: %v", model.runtimeFocusList)

	// Check if input is in focus list
	hasInput := false
	for _, id := range model.runtimeFocusList {
		if id == "my-input" {
			hasInput = true
			break
		}
	}

	if !hasInput {
		t.Error("Input component should be in focus list")
	}

	// Test Tab to set focus
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*Model)

	if cmd != nil {
		msg := cmd()
		newModel, _ = model.Update(msg)
		model = newModel.(*Model)
	}

	t.Logf("After Tab, CurrentFocus: %s", model.CurrentFocus)

	// Test typing
	if model.CurrentFocus == "my-input" {
		// Find the input component
		if model.RuntimeRoot != nil {
			for _, child := range model.RuntimeRoot.Children {
				if child.ID == "my-input" && child.Component != nil && child.Component.Instance != nil {
					if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
						if input, ok := wrapper.Component.(*components.InputComponent); ok {
							initialValue := input.GetValue()
							t.Logf("Input value: '%s', focused: %v", initialValue, input.IsFocused())

							// Type some characters
							model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
							model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

							t.Logf("After typing, value: '%s'", input.GetValue())
						}
					}
				}
			}
		}
	}

	t.Log("Input runtime test passed!")
}

// TestListInRuntimeMode tests list component with runtime mode
func TestListInRuntimeMode(t *testing.T) {
	jsonConfig := `{
		"name": "List Test",
		"useRuntime": true,
		"data": {
			"items": [
				{"id": 1, "name": "Apple"},
				{"id": 2, "name": "Banana"},
				{"id": 3, "name": "Cherry"}
			]
		},
		"layout": {
			"direction": "vertical",
			"children": [
				{
					"type": "text",
					"props": {
						"content": "Select an item:"
					}
				},
				{
					"id": "my-list",
					"type": "list",
					"bind": "items",
					"props": {
						"height": 5,
						"width": 30
					}
				}
			]
		}
	}`

	var config Config
	if err := json.Unmarshal([]byte(jsonConfig), &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	model := NewModel(&config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Trigger rendering
	_ = model.View()

	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	t.Logf("Focus list: %v", model.runtimeFocusList)

	// Check if list is in focus list
	hasList := false
	for _, id := range model.runtimeFocusList {
		if id == "my-list" {
			hasList = true
			break
		}
	}

	if !hasList {
		t.Error("List component should be in focus list")
	}

	// Test Tab to set focus
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*Model)

	if cmd != nil {
		msg := cmd()
		newModel, _ = model.Update(msg)
		model = newModel.(*Model)
	}

	t.Logf("After Tab, CurrentFocus: %s", model.CurrentFocus)

	t.Log("List runtime test passed!")
}

// TestInputFromFile tests with actual input.tui.yao file
func TestInputFromFile(t *testing.T) {
	if _, err := os.Stat("demo/tuis/input.tui.yao"); os.IsNotExist(err) {
		t.Skip("input.tui.yao file not found")
		return
	}

	config := loadConfigFromFile("demo/tuis/input.tui.yao", t)

	model := NewModel(config, nil)
	model.Width = 100
	model.Height = 30

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Trigger rendering
	_ = model.View()

	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	t.Logf("UseRuntime: %v", model.UseRuntime)

	// Count input components in focus list
	inputCount := 0
	for _, id := range model.runtimeFocusList {
		if id == "username-input" || id == "email-input" || id == "password-input" {
			inputCount++
		}
	}

	t.Logf("Found %d input components in focus list", inputCount)
	if inputCount < 3 {
		t.Errorf("Expected at least 3 input components in focus list, got %d", inputCount)
	}

	t.Log("Input file test passed!")
}

// TestListFromFile tests with actual list-simple.tui.yao file
func TestListFromFile(t *testing.T) {
	if _, err := os.Stat("demo/tuis/list-simple.tui.yao"); os.IsNotExist(err) {
		t.Skip("list-simple.tui.yao file not found")
		return
	}

	config := loadConfigFromFile("demo/tuis/list-simple.tui.yao", t)

	model := NewModel(config, nil)
	model.Width = 100
	model.Height = 30

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Trigger rendering
	_ = model.View()

	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	t.Logf("Focus list: %v", model.runtimeFocusList)

	// Check if list is in focus list
	hasList := false
	for _, id := range model.runtimeFocusList {
		if id == "itemList" {
			hasList = true
			break
		}
	}

	if !hasList {
		t.Error("List component should be in focus list")
	}

	t.Log("List file test passed!")
}
