package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestListIntegration tests the list component integration with the TUI model
// This simulates loading a TUI config with a list component
func TestListIntegration(t *testing.T) {
	// Create a simple TUI config with a list component
	config := &Config{
		Name: "List Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					Props: map[string]interface{}{
						"title": "Fruit List",
						"align": "center",
					},
				},
				{
					Type: "list",
					ID:   "fruitList",
					Bind: "fruits",
					Props: map[string]interface{}{
						"height":       10,
						"width":        40,
						"itemTemplate": "{{id}}. {{name}} - {{price}}",
					},
				},
			},
		},
		Data: map[string]interface{}{
			"fruits": []interface{}{
				map[string]interface{}{"id": 1, "name": "Apple", "price": "$1.99"},
				map[string]interface{}{"id": 2, "name": "Banana", "price": "$0.99"},
				map[string]interface{}{"id": 3, "name": "Cherry", "price": "$2.99"},
				map[string]interface{}{"id": 4, "name": "Date", "price": "$3.99"},
				map[string]interface{}{"id": 5, "name": "Elderberry", "price": "$4.99"},
			},
		},
	}

	// Create model
	model := NewModel(config, nil)

	// Initialize components
	cmds := model.InitializeComponents()
	assert.Nil(t, cmds)

	// Set window size
	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	// Mark as ready
	model.Ready = true

	// Render the layout
	rendered := model.View()
	assert.NotEmpty(t, rendered, "Rendered output should not be empty")

	// Verify the list is rendered
	// The list should show items with the template format
	// Since itemTemplate support was added, items should show "1. Apple - $1.99", etc.
	fmt.Printf("Rendered output:\n%s\n", rendered)

	// Check if list items are present in rendered output
	assert.Contains(t, rendered, "Apple", "List should contain 'Apple'")
	assert.Contains(t, rendered, "Banana", "List should contain 'Banana'")
	assert.Contains(t, rendered, "$1.99", "List should contain price '$1.99'")
	assert.Contains(t, rendered, "$0.99", "List should contain price '$0.99'")

	// Verify component instance exists
	fruitListComp, exists := model.ComponentInstanceRegistry.Get("fruitList")
	assert.True(t, exists, "List component instance should exist")
	assert.NotNil(t, fruitListComp)
	assert.Equal(t, "list", fruitListComp.Type)
}

// TestListComponentWithoutTemplate tests list component with default behavior
func TestListComponentWithoutTemplate(t *testing.T) {
	config := &Config{
		Name: "List Test No Template",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "itemList",
					Bind: "items",
					Props: map[string]interface{}{
						"height": 8,
						"width":  30,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"id": 1, "name": "Item One"},
				map[string]interface{}{"id": 2, "name": "Item Two"},
			},
		},
	}

	model := NewModel(config, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{
		Width:  50,
		Height: 20,
	}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	model.Ready = true
	rendered := model.View()

	fmt.Printf("Rendered (no template):\n%s\n", rendered)

	// Should use the 'name' field as fallback
	// Note: With height=8, only 1-2 items may be visible
	assert.Contains(t, rendered, "Item One", "List should contain 'Item One'")
}
