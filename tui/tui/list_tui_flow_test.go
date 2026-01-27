package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestListThroughTUIFlow tests the list through the full TUI rendering flow
func TestListThroughTUIFlow(t *testing.T) {
	// Create config
	config := &Config{
		Name: "List Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "testList",
					Bind: "fruits",
					Props: map[string]interface{}{
						"height":       10,
						"width":        50,
						"itemTemplate": "{{id}}. {{name}} - {{price}}",
					},
				},
			},
		},
		Data: map[string]interface{}{
			"fruits": []interface{}{
				map[string]interface{}{"id": 1, "name": "Apple", "price": "$1.99"},
				map[string]interface{}{"id": 2, "name": "Banana", "price": "$0.99"},
			},
		},
	}

	model := NewModel(config, nil)
	model.InitializeComponents()

	// Check what's in the state
	model.StateMu.RLock()
	fmt.Printf("State fruits: %v\n", model.State["fruits"])
	model.StateMu.RUnlock()

	// Set window size
	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	// Get the component instance
	compInstance, exists := model.ComponentInstanceRegistry.Get("testList")
	assert.True(t, exists, "Component should exist")
	assert.NotNil(t, compInstance)

	// Access the internal list component to debug
	if listComp, ok := compInstance.Instance.(interface{ GetModel() interface{} }); ok {
		if listModel := listComp.GetModel(); listModel != nil {
			// Try to peek into the list model
			fmt.Printf("List model type: %T\n", listModel)
		}
	}

	model.Ready = true
	rendered := model.View()
	fmt.Printf("\nRendered TUI output:\n%s\n", rendered)
}
