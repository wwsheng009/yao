package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestListKeyEventHandling tests if the list component handles keyboard events correctly
func TestListKeyEventHandling(t *testing.T) {
	config := &Config{
		Name: "List Key Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "itemList",
					Bind: "items",
					Props: map[string]interface{}{
						"height":       12,
						"width":        50,
						"itemTemplate": "{{id}}. {{name}} - {{price}}",
						"showTitle":    false,
						"showFilter":   false,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"id": 1, "name": "Apple", "price": "$1.99"},
				map[string]interface{}{"id": 2, "name": "Banana", "price": "$0.99"},
				map[string]interface{}{"id": 3, "name": "Cherry", "price": "$2.99"},
			},
		},
	}

	model := NewModel(config, nil)
	model.InitializeComponents()

	// Set window size
	msg := tea.WindowSizeMsg{Width: 80, Height: 30}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	// Verify list component exists
	compInstance, exists := model.ComponentInstanceRegistry.Get("itemList")
	assert.True(t, exists, "List component should exist")

	// Set focus to the list component to enable keyboard navigation
	model.CurrentFocus = "itemList"

	// Check if component has focus after setting it
	comp, exists := model.Components["itemList"]
	assert.True(t, exists)
	assert.NotNil(t, comp)
	_ = compInstance

	// Manually set focus on the component
	if focuser, ok := comp.Instance.(interface{ SetFocus(bool) }); ok {
		focuser.SetFocus(true)
		t.Logf("Component SetFocus(true) called")
	}

	// Test component focus by calling GetFocus
	if focuser, ok := comp.Instance.(interface{ GetFocus() bool }); ok {
		t.Logf("Component GetFocus() before: %v", focuser.GetFocus())
	}

	t.Logf("Model.CurrentFocus before: %s", model.CurrentFocus)

	// Now send a down key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	t.Logf("Sending down key to component 'itemList'")

	updatedModel, _ := model.Update(downMsg)
	model = updatedModel.(*Model)
	t.Logf("After down key - CurrentFocus: %s", model.CurrentFocus)

	// Check if the component received the update
	if focuser, ok := comp.Instance.(interface{ GetFocus() bool }); ok {
		t.Logf("Component GetFocus() after: %v", focuser.GetFocus())
	}
}
