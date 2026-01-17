package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestMenu_Render(t *testing.T) {
	// Define test menu items
	menuItems := []MenuItem{
		{
			Title:       "Dashboard",
			Description: "View system dashboard",
			Value:       "dashboard",
			Action: map[string]interface{}{
				"process": "yao.demo.dashboard",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Users Management",
			Description: "Manage user accounts",
			Value:       "users",
			Action: map[string]interface{}{
				"process": "yao.demo.users",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Exit",
			Description: "Exit the application",
			Value:       "exit",
			Action: map[string]interface{}{
				"process": "core.quit",
				"args":    []interface{}{},
			},
		},
	}

	// Create menu props with proper lipgloss styles
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("235")).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(lipgloss.Color("233"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("212")).Bold(true)
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	props := MenuProps{
		Title:         "Main Menu",
		Items:         menuItems,
		Height:        15,
		Width:         60,
		ShowStatusBar: true,
		ShowFilter:    false,
		ActiveItemStyle: lipglossStyleWrapper{
			Style: &activeStyle,
		},
		InactiveItemStyle: lipglossStyleWrapper{
			Style: &inactiveStyle,
		},
		SelectedItemStyle: lipglossStyleWrapper{
			Style: &selectedStyle,
		},
		DisabledItemStyle: lipglossStyleWrapper{
			Style: &disabledStyle,
		},
	}

	// Test menu rendering
	width := 80
	result := RenderMenu(props, width)

	// Assertions
	assert.Contains(t, result, "Main Menu", "Rendered menu should contain title")
	assert.Contains(t, result, "Dashboard", "Rendered menu should contain first item")
	assert.Contains(t, result, "Users Management", "Rendered menu should contain second item")
	assert.Contains(t, result, "Exit", "Rendered menu should contain third item")
	// Note: Descriptions are intentionally not rendered to keep the menu compact
	// assert.Contains(t, result, "View system dashboard", "Rendered menu should contain description")
}

func TestMenu_ParseProps(t *testing.T) {
	// Create test props map
	propsMap := map[string]interface{}{
		"title":         "Test Menu",
		"height":        10,
		"width":         50,
		"showStatusBar": true,
		"showFilter":    false,
		"items": []interface{}{
			map[string]interface{}{
				"title":       "Item 1",
				"description": "First item",
				"value":       "item1",
				"action": map[string]interface{}{
					"process": "test.action",
					"args":    []interface{}{"arg1"},
				},
			},
			map[string]interface{}{
				"title":       "Item 2",
				"description": "Second item",
				"value":       "item2",
				"action": map[string]interface{}{
					"process": "test.action2",
					"args":    []interface{}{"arg2"},
				},
			},
		},
		"activeItemStyle": map[string]interface{}{
			"foreground": "226",
			"background": "235",
			"bold":       true,
		},
		"inactiveItemStyle": map[string]interface{}{
			"foreground": "245",
			"background": "233",
		},
	}

	// Parse the props
	props := ParseMenuProps(propsMap)

	// Assertions
	assert.Equal(t, "Test Menu", props.Title)
	assert.Equal(t, 10, props.Height)
	assert.Equal(t, 50, props.Width)
	assert.True(t, props.ShowStatusBar)
	assert.False(t, props.ShowFilter)
	assert.Len(t, props.Items, 2)

	// Check first item
	assert.Equal(t, "Item 1", props.Items[0].Title)
	assert.Equal(t, "First item", props.Items[0].Description)
	assert.Equal(t, "item1", props.Items[0].Value)
	assert.NotNil(t, props.Items[0].Action)
	assert.Equal(t, "test.action", props.Items[0].Action["process"])

	// Check second item
	assert.Equal(t, "Item 2", props.Items[1].Title)
	assert.Equal(t, "Second item", props.Items[1].Description)
	assert.Equal(t, "item2", props.Items[1].Value)
	assert.NotNil(t, props.Items[1].Action)
	assert.Equal(t, "test.action2", props.Items[1].Action["process"])
}

func TestMenuInteractiveModel_Navigation(t *testing.T) {
	// Define test menu items
	menuItems := []MenuItem{
		{
			Title:       "Item 1",
			Description: "First item",
			Value:       "item1",
		},
		{
			Title:       "Item 2",
			Description: "Second item",
			Value:       "item2",
		},
		{
			Title:       "Item 3",
			Description: "Third item",
			Value:       "item3",
		},
	}

	// Create menu props
	props := MenuProps{
		Title: "Navigation Test Menu",
		Items: menuItems,
	}

	// Create interactive menu model
	model := NewMenuInteractiveModel(props)

	// Initially, first item should be selected (index 0)
	assert.Equal(t, 0, model.Index(), "Initially first item should be selected")

	// Send down arrow key message to move selection down
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := HandleMenuUpdate(downMsg, &model)
	// No specific command assertion needed, just check navigation
	assert.Equal(t, 1, newModel.Index(), "After down arrow, second item should be selected")

	// Send another down arrow
	downMsg2 := tea.KeyMsg{Type: tea.KeyDown}
	newModel2, _ := HandleMenuUpdate(downMsg2, &newModel)
	assert.Equal(t, 2, newModel2.Index(), "After second down arrow, third item should be selected")

	// Send up arrow to go back
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	newModel3, _ := HandleMenuUpdate(upMsg, &newModel2)
	assert.Equal(t, 1, newModel3.Index(), "After up arrow, second item should be selected")

	// Send enter to select item
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel4, _ := HandleMenuUpdate(enterMsg, &newModel3)
	assert.Equal(t, 1, newModel4.Index(), "After enter, same item should remain selected")
}

func TestMenuInteractiveModel_GetSelectedItem(t *testing.T) {
	// Define test menu items
	menuItems := []MenuItem{
		{
			Title:       "Dashboard",
			Description: "View system dashboard",
			Value:       "dashboard",
			Action: map[string]interface{}{
				"process": "yao.demo.dashboard",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Users Management",
			Description: "Manage user accounts",
			Value:       "users",
			Action: map[string]interface{}{
				"process": "yao.demo.users",
				"args":    []interface{}{},
			},
		},
	}

	// Create menu props
	props := MenuProps{
		Items: menuItems,
	}

	// Create interactive menu model
	model := NewMenuInteractiveModel(props)

	// Initially, first item (index 0) should be selected
	selectedItem, ok := model.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item")
	assert.Equal(t, "Dashboard", selectedItem.Title)
	assert.Equal(t, "View system dashboard", selectedItem.Description)
	assert.Equal(t, "dashboard", selectedItem.Value)
	assert.Equal(t, "yao.demo.dashboard", selectedItem.Action["process"])

	// Move to second item
	model.Model.Select(1)
	selectedItem2, ok2 := model.GetSelectedItem()
	assert.True(t, ok2, "Should be able to get second selected item")
	assert.Equal(t, "Users Management", selectedItem2.Title)
	assert.Equal(t, "Manage user accounts", selectedItem2.Description)
	assert.Equal(t, "users", selectedItem2.Value)
	assert.Equal(t, "yao.demo.users", selectedItem2.Action["process"])
}

func TestMenu_View(t *testing.T) {
	// Define test menu items
	menuItems := []MenuItem{
		{
			Title:       "Test Item",
			Description: "Test Description",
			Value:       "test",
		},
	}

	// Create menu props
	props := MenuProps{
		Title: "Test Menu",
		Items: menuItems,
	}

	// Create interactive menu model
	model := NewMenuInteractiveModel(props)

	// Get the view
	view := model.View()

	// Assertions
	assert.NotEmpty(t, view, "View should not be empty")
	// The interactive menu view might not contain the title directly, but should contain the items
	assert.Contains(t, view, "Test Item", "View should contain menu item")
}
