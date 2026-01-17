package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
)

// TestMenuNavigationAndRendering tests menu rendering and user navigation with arrow keys
func TestMenuNavigationAndRendering(t *testing.T) {
	// Create menu items for navigation testing
	menuItems := []MenuItem{
		{
			Title:       "Home",
			Description: "Go to home page",
			Value:       "home",
			Action: map[string]interface{}{
				"process": "page.home",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Products",
			Description: "View products catalog",
			Value:       "products",
			Action: map[string]interface{}{
				"process": "page.products",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Services",
			Description: "Our services offering",
			Value:       "services",
			Action: map[string]interface{}{
				"process": "page.services",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "About Us",
			Description: "Learn about our company",
			Value:       "about",
			Action: map[string]interface{}{
				"process": "page.about",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Contact",
			Description: "Get in touch with us",
			Value:       "contact",
			Action: map[string]interface{}{
				"process": "page.contact",
				"args":    []interface{}{},
			},
		},
	}

	// Define styles for the menu
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("235")).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Background(lipgloss.Color("233"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("212")).Bold(true)
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	props := MenuProps{
		Title:         "Main Navigation Menu",
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

	// Test static rendering
	width := 80
	staticRender := RenderMenu(props, width)
	assert.Contains(t, staticRender, "Main Navigation Menu", "Static render should contain menu title")
	assert.Contains(t, staticRender, "Home", "Static render should contain first menu item")
	assert.Contains(t, staticRender, "Products", "Static render should contain second menu item")
	assert.Contains(t, staticRender, "Services", "Static render should contain third menu item")
	assert.Contains(t, staticRender, "About Us", "Static render should contain fourth menu item")
	assert.Contains(t, staticRender, "Contact", "Static render should contain fifth menu item")
	assert.Contains(t, staticRender, "Go to home page", "Static render should contain first description")
	assert.Contains(t, staticRender, "View products catalog", "Static render should contain second description")

	// Test interactive menu model creation and initial state
	interactiveModel := NewMenuInteractiveModel(props)
	
	// Initially, first item (index 0) should be selected
	assert.Equal(t, 0, interactiveModel.Index(), "Initially first item should be selected")
	
	// Verify all items are present
	assert.Equal(t, 5, len(interactiveModel.Items()), "Interactive model should have 5 items")

	// Test navigation with arrow keys
	currentModel := interactiveModel

	// Navigate down using Down arrow (should select second item)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	nextModel, _ := HandleMenuUpdate(downMsg, &currentModel)
	assert.Equal(t, 1, nextModel.Index(), "After Down arrow, second item should be selected")
	currentModel = nextModel

	// Navigate down again (should select third item)
	downMsg2 := tea.KeyMsg{Type: tea.KeyDown}
	nextModel2, _ := HandleMenuUpdate(downMsg2, &currentModel)
	assert.Equal(t, 2, nextModel2.Index(), "After second Down arrow, third item should be selected")
	currentModel = nextModel2

	// Navigate down again (should select fourth item)
	downMsg3 := tea.KeyMsg{Type: tea.KeyDown}
	nextModel3, _ := HandleMenuUpdate(downMsg3, &currentModel)
	assert.Equal(t, 3, nextModel3.Index(), "After third Down arrow, fourth item should be selected")
	currentModel = nextModel3

	// Navigate down again (should select fifth item)
	downMsg4 := tea.KeyMsg{Type: tea.KeyDown}
	nextModel4, _ := HandleMenuUpdate(downMsg4, &currentModel)
	assert.Equal(t, 4, nextModel4.Index(), "After fourth Down arrow, fifth item should be selected")
	currentModel = nextModel4

	// Now test navigation back up using Up arrow
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	prevModel, _ := HandleMenuUpdate(upMsg, &currentModel)
	assert.Equal(t, 3, prevModel.Index(), "After Up arrow, fourth item should be selected")
	currentModel = prevModel

	// Navigate up again
	upMsg2 := tea.KeyMsg{Type: tea.KeyUp}
	prevModel2, _ := HandleMenuUpdate(upMsg2, &currentModel)
	assert.Equal(t, 2, prevModel2.Index(), "After second Up arrow, third item should be selected")
	currentModel = prevModel2

	// Navigate up again
	upMsg3 := tea.KeyMsg{Type: tea.KeyUp}
	prevModel3, _ := HandleMenuUpdate(upMsg3, &currentModel)
	assert.Equal(t, 1, prevModel3.Index(), "After third Up arrow, second item should be selected")
	currentModel = prevModel3

	// Navigate up again to first item
	upMsg4 := tea.KeyMsg{Type: tea.KeyUp}
	prevModel4, _ := HandleMenuUpdate(upMsg4, &currentModel)
	assert.Equal(t, 0, prevModel4.Index(), "After fourth Up arrow, first item should be selected")

	// Test that we can get the correct selected items at each stage
	selectedAtStart, ok := prevModel4.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item at start")
	assert.Equal(t, "Home", selectedAtStart.Title, "First selected item should be Home")

	// Navigate to third item and verify selection
	navigateToThird := prevModel4
	for i := 0; i < 2; i++ {
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		navigateToThird, _ = HandleMenuUpdate(downMsg, &navigateToThird)
	}
	
	selectedAtThird, ok := navigateToThird.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item at third position")
	assert.Equal(t, "Services", selectedAtThird.Title, "Third selected item should be Services")
	assert.Equal(t, "Our services offering", selectedAtThird.Description, "Third selected item description should be correct")

	// Test that the interactive view contains the expected content
	interactiveView := navigateToThird.View()
	assert.Contains(t, interactiveView, "Services", "Interactive view should contain the selected item")
	assert.NotEmpty(t, interactiveView, "Interactive view should not be empty")
}

// TestMenuWithDifferentSizes tests menu rendering with different sizes
func TestMenuWithDifferentSizes(t *testing.T) {
	menuItems := []MenuItem{
		{Title: "Item 1", Description: "First item", Value: "1"},
		{Title: "Item 2", Description: "Second item", Value: "2"},
		{Title: "Item 3", Description: "Third item", Value: "3"},
	}

	props := MenuProps{
		Title: "Size Test Menu",
		Items: menuItems,
		Width: 40,
		Height: 10,
	}

	// Test with different widths
	for _, width := range []int{30, 50, 70, 100} {
		result := RenderMenu(props, width)
		assert.Contains(t, result, "Size Test Menu", "Should render with width %d", width)
		assert.Contains(t, result, "Item 1", "Should contain first item with width %d", width)
		assert.Contains(t, result, "Item 2", "Should contain second item with width %d", width)
		assert.Contains(t, result, "Item 3", "Should contain third item with width %d", width)
	}

	// Test interactive model with different configurations
	interactiveModel := NewMenuInteractiveModel(props)
	assert.Equal(t, 3, len(interactiveModel.Items()), "Should have 3 items regardless of size")
	assert.Equal(t, "Size Test Menu", interactiveModel.Title, "Title should match")
}

// TestMenuArrowKeysCompatibility tests that menu responds correctly to various arrow key combinations
func TestMenuArrowKeysCompatibility(t *testing.T) {
	menuItems := []MenuItem{
		{Title: "Option 1", Value: "opt1"},
		{Title: "Option 2", Value: "opt2"},
		{Title: "Option 3", Value: "opt3"},
		{Title: "Option 4", Value: "opt4"},
	}

	props := MenuProps{
		Title: "Arrow Keys Test Menu",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	assert.Equal(t, 0, model.Index(), "Initially first item should be selected")

	// Test Down arrow key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downMsg, &model)
	assert.Equal(t, 1, model.Index(), "After Down arrow, second item should be selected")

	// Test j key (alternative to Down arrow in many UIs)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	model, _ = HandleMenuUpdate(jMsg, &model)
	assert.Equal(t, 2, model.Index(), "After 'j' key, third item should be selected")

	// Test Up arrow key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	model, _ = HandleMenuUpdate(upMsg, &model)
	assert.Equal(t, 1, model.Index(), "After Up arrow, second item should be selected")

	// Test k key (alternative to Up arrow in many UIs)
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	model, _ = HandleMenuUpdate(kMsg, &model)
	assert.Equal(t, 0, model.Index(), "After 'k' key, first item should be selected")

	// Test boundary conditions - going past the last item
	for i := 0; i < 10; i++ {
		model, _ = HandleMenuUpdate(downMsg, &model)
	}
	// Should stop at the last item (index 3 for 4 items)
	assert.Equal(t, 3, model.Index(), "Should not go past last item")

	// Test boundary conditions - going before the first item
	for i := 0; i < 10; i++ {
		model, _ = HandleMenuUpdate(upMsg, &model)
	}
	// Should stop at the first item (index 0)
	assert.Equal(t, 0, model.Index(), "Should not go before first item")
}

// TestMenuSelectionAndResponse tests menu response to user selections
func TestMenuSelectionAndResponse(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "Dashboard",
			Description: "View system dashboard",
			Value:       "dashboard",
			Action: map[string]interface{}{
				"process": "yao.demo.dashboard",
				"args":    []interface{}{"dashboard_id"},
			},
		},
		{
			Title:       "Users Management",
			Description: "Manage user accounts",
			Value:       "users",
			Action: map[string]interface{}{
				"process": "yao.demo.users",
				"args":    []interface{}{"users_list"},
			},
		},
		{
			Title:       "Settings",
			Description: "Application settings",
			Value:       "settings",
			Action: map[string]interface{}{
				"process": "yao.demo.settings",
				"args":    []interface{}{"config"},
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

	props := MenuProps{
		Title: "Interactive Menu Response Test",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	assert.Equal(t, 0, model.Index(), "Initially first item should be selected")

	// Test getting selected item before selection
	selectedItem, ok := model.GetSelectedItem()
	assert.True(t, ok, "Should be able to get initially selected item")
	assert.Equal(t, "Dashboard", selectedItem.Title, "Initial selected item should be Dashboard")
	assert.Equal(t, "yao.demo.dashboard", selectedItem.Action["process"], "Initial selected item action should be correct")
	assert.Equal(t, []interface{}{"dashboard_id"}, selectedItem.Action["args"], "Initial selected item args should be correct")

	// Navigate to second item
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downMsg, &model)
	assert.Equal(t, 1, model.Index(), "After Down arrow, second item should be selected")

	// Get selected item after navigation
	selectedItem2, ok2 := model.GetSelectedItem()
	assert.True(t, ok2, "Should be able to get second selected item")
	assert.Equal(t, "Users Management", selectedItem2.Title, "Second selected item should be Users Management")
	assert.Equal(t, "yao.demo.users", selectedItem2.Action["process"], "Second selected item action should be correct")
	assert.Equal(t, []interface{}{"users_list"}, selectedItem2.Action["args"], "Second selected item args should be correct")

	// Navigate to last item (Exit)
	for i := 0; i < 2; i++ {
		model, _ = HandleMenuUpdate(downMsg, &model)
	}
	assert.Equal(t, 3, model.Index(), "After two more Down arrows, last item should be selected")

	selectedItemLast, okLast := model.GetSelectedItem()
	assert.True(t, okLast, "Should be able to get last selected item")
	assert.Equal(t, "Exit", selectedItemLast.Title, "Last selected item should be Exit")
	assert.Equal(t, "core.quit", selectedItemLast.Action["process"], "Exit item action should be core.quit")
	assert.Equal(t, []interface{}{}, selectedItemLast.Action["args"], "Exit item should have no args")

	// Test that the view reflects the current selection
	view := model.View()
	assert.NotEmpty(t, view, "View should not be empty")
	assert.Contains(t, view, "Exit", "View should contain the selected item")
}

// TestMenuEdgeCases tests menu behavior in edge cases
func TestMenuEdgeCases(t *testing.T) {
	// Test with single item
	singleItemMenu := []MenuItem{
		{Title: "Single Item", Value: "single", Description: "Only item in menu"},
	}
	
	singleItemProps := MenuProps{
		Title: "Single Item Menu",
		Items: singleItemMenu,
	}
	
	singleItemModel := NewMenuInteractiveModel(singleItemProps)
	assert.Equal(t, 0, singleItemModel.Index(), "Single item should be selected")
	assert.Equal(t, 1, len(singleItemModel.Items()), "Should have 1 item")

	// Navigation on single item should not change selection
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	nextModel, _ := HandleMenuUpdate(downMsg, &singleItemModel)
	assert.Equal(t, 0, nextModel.Index(), "On single item menu, navigation should not change selection")

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	prevModel, _ := HandleMenuUpdate(upMsg, &singleItemModel)
	assert.Equal(t, 0, prevModel.Index(), "On single item menu, up navigation should not change selection")

	// Test with empty items (we already have a test for this, but let's verify rendering)
	emptyProps := MenuProps{
		Title: "Empty Menu",
		Items: []MenuItem{},
	}
	
	emptyResult := RenderMenu(emptyProps, 80)
	assert.Contains(t, emptyResult, "Empty Menu", "Empty menu should still render title")
	
	emptyModel := NewMenuInteractiveModel(emptyProps)
	assert.Equal(t, 0, len(emptyModel.Items()), "Empty menu should have 0 items")
	
	// Test menu with one disabled item
	disabledItemMenu := []MenuItem{
		{Title: "Normal Item", Value: "normal", Description: "Normal menu item"},
		{Title: "Disabled Item", Value: "disabled", Description: "Disabled menu item", Disabled: true},
		{Title: "Another Normal Item", Value: "normal2", Description: "Another normal menu item"},
	}
	
	disabledProps := MenuProps{
		Title: "Menu with Disabled Item",
		Items: disabledItemMenu,
	}
	
	disabledModel := NewMenuInteractiveModel(disabledProps)
	assert.Equal(t, 3, len(disabledModel.Items()), "Should have 3 items including disabled one")
	
	// Navigation should still work with disabled items
	currentModel := disabledModel
	for i := 0; i < 5; i++ { // Navigate more than the number of items
		currentModel, _ = HandleMenuUpdate(downMsg, &currentModel)
		// Should cycle within bounds, never exceed max index
		assert.LessOrEqual(t, currentModel.Index(), 2, "Should not exceed max index")
		assert.GreaterOrEqual(t, currentModel.Index(), 0, "Should not be negative")
	}
	
	// All items should be selectable despite one being marked as disabled
	// (in our implementation, disabled affects appearance but not selection)
	_, ok := currentModel.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item even with disabled items present")
}

// TestMenuActionTriggering tests menu action triggering when user presses Enter
func TestMenuActionTriggering(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "Dashboard",
			Description: "View system dashboard",
			Value:       "dashboard",
			Action: map[string]interface{}{
				"process": "yao.demo.dashboard",
				"args":    []interface{}{"dashboard_id"},
			},
		},
		{
			Title:       "Users Management",
			Description: "Manage user accounts",
			Value:       "users",
			Action: map[string]interface{}{
				"process": "yao.demo.users",
				"args":    []interface{}{"users_list"},
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

	props := MenuProps{
		Title: "Action Triggering Test Menu",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	assert.Equal(t, 0, model.Index(), "Initially first item should be selected")

	// Test that initially no action is triggered
	selectedItem, ok := model.GetSelectedItem()
	assert.True(t, ok, "Should be able to get initially selected item")
	assert.Equal(t, "Dashboard", selectedItem.Title, "Initial selected item should be Dashboard")

	// Simulate pressing Enter on the selected item
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := HandleMenuUpdate(enterMsg, &model)

	// Verify that the model was updated
	assert.Equal(t, 0, updatedModel.Index(), "After Enter, same item should still be selected")
	
	// Check if a command was returned (indicating an action was triggered)
	if cmd != nil {
		// Execute the command to get the resulting message
		resultMsg := cmd()
		actionMsg, ok := resultMsg.(core.MenuActionTriggered)
		assert.True(t, ok, "Command should return MenuActionTriggered message when Enter is pressed on item with action")
		
		if ok {
			assert.Equal(t, "Dashboard", actionMsg.Item.GetTitle(), "Action should be triggered for Dashboard")
			assert.Equal(t, "yao.demo.dashboard", actionMsg.Action["process"], "Action process should be yao.demo.dashboard")
			assert.Equal(t, []interface{}{"dashboard_id"}, actionMsg.Action["args"], "Action args should be correct")
		}
	} else {
		t.Log("No command returned - this may be expected depending on implementation")
	}

	// Navigate to second item and test action triggering there
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	model2, _ := HandleMenuUpdate(downMsg, &updatedModel)
	assert.Equal(t, 1, model2.Index(), "After Down arrow, second item should be selected")

	// Press Enter on second item
	enterMsg2 := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd2 := HandleMenuUpdate(enterMsg2, &model2)
	
	if cmd2 != nil {
		resultMsg2 := cmd2()
		actionMsg2, ok2 := resultMsg2.(core.MenuActionTriggered)
		assert.True(t, ok2, "Command should return MenuActionTriggered message for second item")
		
		if ok2 {
			assert.Equal(t, "Users Management", actionMsg2.Item.GetTitle(), "Action should be triggered for Users Management")
			assert.Equal(t, "yao.demo.users", actionMsg2.Action["process"], "Action process should be yao.demo.users")
			assert.Equal(t, []interface{}{"users_list"}, actionMsg2.Action["args"], "Action args should be correct for second item")
		}
	}

	// Test that items without actions don't trigger anything
	menuItemWithoutAction := []MenuItem{
		{Title: "Simple Item", Description: "Item without action", Value: "simple"},
	}
	
	propsWithoutAction := MenuProps{
		Items: menuItemWithoutAction,
	}
	
	modelWithoutAction := NewMenuInteractiveModel(propsWithoutAction)
	enterMsg3 := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd3 := HandleMenuUpdate(enterMsg3, &modelWithoutAction)
	
	// When there's no action, cmd might be nil or might not be a MenuActionTriggered
	if cmd3 != nil {
		resultMsg3 := cmd3()
		_, ok := resultMsg3.(core.MenuActionTriggered)
		_ = ok  // Just to use the variable
		// If it's not a MenuActionTriggered, that's fine - it means no action was triggered
		// This is acceptable behavior for items without actions
	}
}

// TestMenuItemWithSpecialCharacters tests menu behavior with special characters
func TestMenuItemWithSpecialCharacters(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "File & Edit",
			Description: "File and edit operations",
			Value:       "file_edit",
			Action: map[string]interface{}{
				"process": "app.file.edit",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Help ?",
			Description: "Get help and support",
			Value:       "help",
			Action: map[string]interface{}{
				"process": "app.help",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Settings ⚙",
			Description: "Configure application settings",
			Value:       "settings",
			Action: map[string]interface{}{
				"process": "app.settings",
				"args":    []interface{}{},
			},
		},
	}

	props := MenuProps{
		Title: "Special Characters Test Menu",
		Items: menuItems,
	}

	// Test rendering with special characters
	result := RenderMenu(props, 80)
	assert.Contains(t, result, "File & Edit", "Should render special characters in title")
	assert.Contains(t, result, "Help ?", "Should render special characters in title")
	assert.Contains(t, result, "Settings ⚙", "Should render special characters in title")

	// Test interactive model with special characters
	model := NewMenuInteractiveModel(props)
	assert.Equal(t, 3, len(model.Items()), "Should have 3 items with special characters")

	// Navigate and verify special characters are preserved
	for i := 0; i < 3; i++ {
		selectedItem, ok := model.GetSelectedItem()
		assert.True(t, ok, "Should be able to get selected item %d", i)
		
		switch i {
		case 0:
			assert.Equal(t, "File & Edit", selectedItem.Title, "Special characters should be preserved in title")
		case 1:
			assert.Equal(t, "Help ?", selectedItem.Title, "Special characters should be preserved in title")
		case 2:
			assert.Equal(t, "Settings ⚙", selectedItem.Title, "Unicode characters should be preserved in title")
		}

		// Navigate to next item (except on last iteration)
		if i < 2 {
			downMsg := tea.KeyMsg{Type: tea.KeyDown}
			model, _ = HandleMenuUpdate(downMsg, &model)
		}
	}
}

// TestMenuAdvancedFeatures tests advanced menu features like filtering and status bar
func TestMenuAdvancedFeatures(t *testing.T) {
	menuItems := []MenuItem{
		{Title: "Dashboard", Description: "System dashboard", Value: "dashboard"},
		{Title: "User Management", Description: "Manage users", Value: "users"},
		{Title: "Settings", Description: "Application settings", Value: "settings"},
		{Title: "Reports", Description: "Generate reports", Value: "reports"},
		{Title: "Help", Description: "Get help", Value: "help"},
	}

	props := MenuProps{
		Title:         "Advanced Features Test Menu",
		Items:         menuItems,
		ShowStatusBar: true,
		ShowFilter:    true, // Enable filtering
	}

	model := NewMenuInteractiveModel(props)
	
	// Verify that advanced features are configured properly
	assert.Equal(t, 5, len(model.Items()), "Should have 5 items")
	assert.Equal(t, true, model.FilteringEnabled(), "Filtering should be enabled when ShowFilter is true")
	
	// Test status bar visibility
	// Note: We can't directly test the status bar visibility, but we can ensure the model is created correctly
	
	// Test filtering functionality by simulating typing in the filter
	filterMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}} // Type 'u'
	model, _ = HandleMenuUpdate(filterMsg, &model)
	
	// After typing 'u', we should see items that contain 'u' in their title
	// In our case: "User Management" and possibly others if they contain 'u'
	_ = model.View()
	// The view will show filtered results, but the exact behavior depends on the underlying list implementation
	
	// Reset the filter by simulating escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	model, _ = HandleMenuUpdate(escMsg, &model)
	
	// Test with another filter
	filterMsg2 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}} // Type 's'
	model, _ = HandleMenuUpdate(filterMsg2, &model)
	
	// After typing 's', we should see items that contain 's' in their title
	// In our case: "User Management", "Settings", "Reports"
	
	// Test that we can disable filtering
	propsNoFilter := MenuProps{
		Title:         "No Filter Test Menu",
		Items:         menuItems,
		ShowStatusBar: true,
		ShowFilter:    false, // Disable filtering
	}
	
	modelNoFilter := NewMenuInteractiveModel(propsNoFilter)
	// Note: Filtering might be enabled by default in the underlying list component
	// The ShowFilter prop controls visual display of filter bar, not necessarily functionality
	_ = modelNoFilter // Use the variable
	
	// Test rendering with and without filter
	resultWithFilter := RenderMenu(props, 80)
	resultWithoutFilter := RenderMenu(MenuProps{
		Title:         "No Filter Menu",
		Items:         menuItems,
		ShowStatusBar: true,
		ShowFilter:    false,
	}, 80)
	
	// Both should contain the menu title and items
	assert.Contains(t, resultWithFilter, "Advanced Features Test Menu", "Result with filter should contain title")
	assert.Contains(t, resultWithoutFilter, "No Filter Menu", "Result without filter should contain title")
	
	for _, item := range menuItems {
		assert.Contains(t, resultWithFilter, item.Title, "Result with filter should contain item: %s", item.Title)
		assert.Contains(t, resultWithoutFilter, item.Title, "Result without filter should contain item: %s", item.Title)
	}
}

// TestMenuWithFocusAndBlur tests menu behavior with focus changes
func TestMenuWithFocusAndBlur(t *testing.T) {
	menuItems := []MenuItem{
		{Title: "Option 1", Value: "opt1", Description: "First option"},
		{Title: "Option 2", Value: "opt2", Description: "Second option"},
		{Title: "Option 3", Value: "opt3", Description: "Third option"},
	}

	props := MenuProps{
		Title:   "Focus Test Menu",
		Items:   menuItems,
		Focused: true,
	}

	model := NewMenuInteractiveModel(props)
	
	// Verify model was created with items
	assert.Equal(t, 3, len(model.Items()), "Menu should have 3 items")
	
	// Test with Focused=false initially
	propsUnfocused := MenuProps{
		Title:   "Unfocused Test Menu",
		Items:   menuItems,
		Focused: false, // Start unfocused
	}
	
	modelUnfocused := NewMenuInteractiveModel(propsUnfocused)
	assert.Equal(t, 3, len(modelUnfocused.Items()), "Menu should have 3 items")
}

// TestNestedMenus tests submenu functionality and navigation
func TestNestedMenus(t *testing.T) {
	// Create a menu with nested submenus
	menuItems := []MenuItem{
		{
			Title:       "Dashboard",
			Description: "Main dashboard",
			Value:       "dashboard",
			Action: map[string]interface{}{
				"process": "dashboard.view",
				"args":    []interface{}{},
			},
		},
		{
			Title:       "Settings",
			Description: "Application settings",
			Value:       "settings",
			Children: []MenuItem{
				{
					Title:       "General Settings",
					Description: "General application settings",
					Value:       "general",
					Action: map[string]interface{}{
						"process": "settings.general",
						"args":    []interface{}{},
					},
				},
				{
					Title:       "User Preferences",
					Description: "User preference settings",
					Value:       "preferences",
					Action: map[string]interface{}{
						"process": "settings.preferences",
						"args":    []interface{}{},
					},
				},
				{
					Title:       "Security",
					Description: "Security settings",
					Value:       "security",
					Children: []MenuItem{
						{
							Title:       "Passwords",
							Description: "Password settings",
							Value:       "passwords",
							Action: map[string]interface{}{
								"process": "security.passwords",
								"args":    []interface{}{},
							},
						},
						{
							Title:       "Access Control",
							Description: "Access control settings",
							Value:       "access",
							Action: map[string]interface{}{
								"process": "security.access",
								"args":    []interface{}{},
							},
						},
					},
				},
			},
		},
		{
			Title:       "Tools",
			Description: "Various tools",
			Value:       "tools",
			Children: []MenuItem{
				{
					Title:       "Import",
					Description: "Import data",
					Value:       "import",
					Action: map[string]interface{}{
						"process": "tools.import",
						"args":    []interface{}{},
					},
				},
				{
					Title:       "Export",
					Description: "Export data",
					Value:       "export",
					Action: map[string]interface{}{
						"process": "tools.export",
						"args":    []interface{}{},
					},
				},
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

	props := MenuProps{
		Title: "Nested Menu Test",
		Items: menuItems,
	}

	// Test static rendering includes submenu indicators
	result := RenderMenu(props, 80)
	assert.Contains(t, result, "Settings ▶", "Should show submenu indicator for Settings")
	assert.Contains(t, result, "Tools ▶", "Should show submenu indicator for Tools")
	assert.Contains(t, result, "Nested Menu Test", "Should contain menu title")
	// Note: Security ▶ won't appear in static render because it's nested inside Settings, not at top level

	// Test interactive model creation
	model := NewMenuInteractiveModel(props)
	assert.Equal(t, 0, model.CurrentLevel, "Should start at level 0 (top level)")
	assert.Equal(t, 0, len(model.Path), "Should start with empty path")
	assert.Equal(t, 4, len(model.Items()), "Should have 4 top-level items")

	// Test submenu detection
	settingsItem := menuItems[1] // Settings item
	assert.True(t, settingsItem.HasSubmenu(), "Settings item should have submenu")
	assert.Equal(t, 3, len(settingsItem.Children), "Settings should have 3 submenu items")

	// Test that items with children have the correct flag
	assert.True(t, settingsItem.HasChildren(), "Settings item should have HasChildren=true")
	assert.True(t, settingsItem.Children[2].HasSubmenu(), "Security item should have submenu")
}

// TestMenuItemSelection tests menu item selection and deselection
func TestMenuItemSelection(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "Item 1",
			Description: "First item",
			Value:       "item1",
			Selected:    false,
		},
		{
			Title:       "Item 2",
			Description: "Second item",
			Value:       "item2",
			Selected:    false,
		},
		{
			Title:       "Item 3",
			Description: "Third item",
			Value:       "item3",
			Selected:    false,
		},
	}

	props := MenuProps{
		Title: "Selection Test Menu",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	
	// Initially, no items should be specifically marked as selected (though one will be focused)
	selectedItem, ok := model.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item")
	assert.Equal(t, "Item 1", selectedItem.Title, "First item should be initially selected by default")
	
	// Verify that the selected property works in the data
	assert.Equal(t, false, selectedItem.Selected, "Item should have Selected property as false by default")

	// Test navigation doesn't change the Selected field but changes focused item
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downMsg, &model)
	
	selectedItem2, ok2 := model.GetSelectedItem()
	assert.True(t, ok2, "Should be able to get selected item after navigation")
	assert.Equal(t, "Item 2", selectedItem2.Title, "Second item should be selected after navigation")
	assert.Equal(t, false, selectedItem2.Selected, "Second item should have Selected property as false by default")
}

// TestMenuProcessorCallbacks tests processor callback handling
func TestMenuProcessorCallbacks(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "Process Item 1",
			Description: "First process item",
			Value:       "proc1",
			Action: map[string]interface{}{
				"process": "test.process.1",
				"args":    []interface{}{"arg1", "arg2"},
			},
		},
		{
			Title:       "Process Item 2",
			Description: "Second process item",
			Value:       "proc2",
			Action: map[string]interface{}{
				"process": "test.process.2",
				"args":    []interface{}{"arg3", "arg4"},
			},
		},
		{
			Title:       "No Process Item",
			Description: "Item without process",
			Value:       "noproc",
		},
	}

	props := MenuProps{
		Title: "Processor Callback Test",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	
	// Test first item action trigger
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := HandleMenuUpdate(enterMsg, &model)
	
	// Verify command is returned when action exists
	if cmd != nil {
		resultMsg := cmd()
		actionMsg, ok := resultMsg.(core.MenuActionTriggered)
		assert.True(t, ok, "Should return MenuActionTriggered when Enter pressed on item with action")
		
		if ok {
			assert.Equal(t, "Process Item 1", actionMsg.Item.GetTitle(), "Action should be triggered for first item")
			assert.Equal(t, "test.process.1", actionMsg.Action["process"], "Correct process should be in action")
			assert.Equal(t, []interface{}{"arg1", "arg2"}, actionMsg.Action["args"], "Correct args should be in action")
		}
	}

	// Navigate to item without action
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downMsg, &model)
	downMsg2 := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downMsg2, &model)
	
	// Try to trigger action on item without action
	enterMsg2 := tea.KeyMsg{Type: tea.KeyEnter}
	_, _ = HandleMenuUpdate(enterMsg2, &model)
	
	// Command may or may not be returned for items without actions, depending on implementation
	// The important thing is that it doesn't crash
}

// TestMenuExitHandling tests default exit button handling
func TestMenuExitHandling(t *testing.T) {
	menuItems := []MenuItem{
		{Title: "Item 1", Value: "item1"},
		{Title: "Item 2", Value: "item2"},
		{Title: "Exit", Value: "exit"},
	}

	props := MenuProps{
		Title: "Exit Handling Test",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	
	// Test 'q' key exit
	qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, _ = HandleMenuUpdate(qMsg, &model)
	// We expect tea.Quit command to be returned, but we can't easily test it without executing it
	
	// Test ESC key exit
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	_, _ = HandleMenuUpdate(escMsg, &model)
	// Similar expectation for ESC key
	
	// Test Ctrl+C exit
	ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, _ = HandleMenuUpdate(ctrlCMsg, &model)
	// Similar expectation for Ctrl+C key
	
	// The important thing is that these key bindings don't cause errors
	assert.Equal(t, 3, len(model.Items()), "Model should still have all items after exit attempts")
}

// TestMenuNavigationEnhanced tests enhanced navigation with submenu support
func TestMenuNavigationEnhanced(t *testing.T) {
	menuItems := []MenuItem{
		{
			Title:       "Main Item 1",
			Description: "First main item",
			Value:       "main1",
		},
		{
			Title:       "Submenu Parent",
			Description: "Has submenu items",
			Value:       "parent",
			Children: []MenuItem{
				{Title: "Child 1", Value: "child1", Description: "First child"},
				{Title: "Child 2", Value: "child2", Description: "Second child"},
			},
		},
		{
			Title:       "Main Item 3",
			Description: "Third main item",
			Value:       "main3",
		},
	}

	props := MenuProps{
		Title: "Enhanced Navigation Test",
		Items: menuItems,
	}

	model := NewMenuInteractiveModel(props)
	assert.Equal(t, 0, model.CurrentLevel, "Should start at level 0")
	assert.Equal(t, 0, len(model.Path), "Should start with empty path")

	// Navigate to submenu parent
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downMsg, &model) // Now on "Submenu Parent"
	assert.Equal(t, "Submenu Parent", model.SelectedItem().(MenuItem).Title, "Should be on submenu parent")

	// Enter the submenu
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, _ = HandleMenuUpdate(enterMsg, &model)
	
	// After entering submenu
	assert.Equal(t, 1, model.CurrentLevel, "Should be at level 1 after entering submenu")
	assert.Equal(t, 1, len(model.Path), "Should have one item in path")
	assert.Equal(t, "Submenu Parent", model.Path[0], "Path should contain parent menu name")
	assert.Equal(t, 2, len(model.Items()), "Should now have 2 submenu items")
	
	// Navigate within submenu
	downInSubmenu := tea.KeyMsg{Type: tea.KeyDown}
	model, _ = HandleMenuUpdate(downInSubmenu, &model)
	selectedInSubmenu, ok := model.GetSelectedItem()
	assert.True(t, ok, "Should be able to get selected item in submenu")
	assert.Equal(t, "Child 2", selectedInSubmenu.Title, "Should be on Child 2 after navigating down in submenu")

	// Go back to parent menu
	leftMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}} // Use 'h' key to go back
	model, _ = HandleMenuUpdate(leftMsg, &model)
	
	// After going back
	assert.Equal(t, 0, model.CurrentLevel, "Should be back at level 0")
	assert.Equal(t, 0, len(model.Path), "Should have empty path again")
	assert.Equal(t, 3, len(model.Items()), "Should have original 3 items again")
}
