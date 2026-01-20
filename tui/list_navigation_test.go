package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestListNavigation tests list navigation with arrow keys
func TestListNavigation(t *testing.T) {
	config := &Config{
		Name: "List Navigation Test",
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
				map[string]interface{}{"id": 4, "name": "Date", "price": "$3.99"},
			},
		},
	}

	model := NewModel(config, nil)
	model.InitializeComponents()

	// Set window size
	msg := tea.WindowSizeMsg{Width: 80, Height: 30}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	// Get the list component and check initial index
	comp := model.Components["itemList"]
	listComp, ok := comp.Instance.(interface{ GetModel() interface{} })
	if ok {
		if listModel, ok := listComp.GetModel().(list.Model); ok {
			t.Logf("Initial list index: %d", listModel.Index())
			assert.Equal(t, 0, listModel.Index(), "Initial index should be 0")
		}
	}

	// Set focus and send down keys
	model.CurrentFocus = "itemList"
	if focuser, ok := comp.Instance.(interface{ SetFocus(bool) }); ok {
		focuser.SetFocus(true)
	}

	// Press down twice
	for i := 0; i < 2; i++ {
		t.Logf("Sending down key #%d", i+1)
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = newModel.(*Model)

		// Check if index changed
		if listComp, ok := comp.Instance.(interface{ GetModel() interface{} }); ok {
			if listModel, ok := listComp.GetModel().(list.Model); ok {
				t.Logf("After down key #%d, list index: %d", i+1, listModel.Index())
			}
		}
	}

	// Final check - should be at index 2
	if listComp, ok := comp.Instance.(interface{ GetModel() interface{} }); ok {
		if listModel, ok := listComp.GetModel().(list.Model); ok {
			finalIndex := listModel.Index()
			t.Logf("Final list index: %d", finalIndex)
			assert.Equal(t, 2, finalIndex, "After 2 down keys, index should be 2")
		}
	}
}
