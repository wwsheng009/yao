package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestMenuDelegateRender verifies that itemDelegate.Render is called in interactive mode
func TestMenuDelegateRender(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "Test Item 1",
			Description: "First test item",
			Value:       "item1",
		},
		{
			Title:       "Test Item 2", 
			Description: "Second test item",
			Value:       "item2",
		},
	}

	props := MenuProps{
		Title: "Debug Test Menu",
		Items: menuItems,
	}

	// Create interactive menu model
	model := NewMenuInteractiveModel(props)
	
	// Verify model was created properly
	assert.Equal(t, 2, len(model.Items()), "Should have 2 menu items")
	assert.Equal(t, "Debug Test Menu", model.Title, "Should have correct title")
	
	// Test that we can get the view (this internally calls itemDelegate.Render)
	view := model.View()
	assert.NotEmpty(t, view, "View should not be empty")
	
	// Test navigation (this will trigger itemDelegate.Render for highlighting)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := HandleMenuUpdate(downMsg, &model)
	
	// Verify navigation worked
	assert.Equal(t, 1, updatedModel.Index(), "Should navigate to second item")
	
	// Get view after navigation
	viewAfterNav := updatedModel.View()
	assert.NotEmpty(t, viewAfterNav, "View after navigation should not be empty")
	
	// Test getting selected item
	selectedItem, ok := updatedModel.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item")
	assert.Equal(t, "Test Item 2", selectedItem.Title, "Should have correct selected item after navigation")
}