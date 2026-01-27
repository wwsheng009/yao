package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/component"
)

// TestTUIListCommandFlow simulates the data flow from cmd/tui/list.go
// This tests how external data from prepareTUIListData is passed to the list component.
func TestTUIListCommandFlow(t *testing.T) {
	// Simulate fixed prepareTUIListData() output from cmd/tui/list.go
	// Now using []interface{} with map items (after fix)
	tuiItems := []interface{}{
		map[string]interface{}{
			"id":    "demo",
			"name":  "Demo TUI",
			"title": "demo - Demo TUI",
		},
		map[string]interface{}{
			"id":    "list-simple",
			"name":  "Simple List",
			"title": "list-simple - Simple List",
		},
		map[string]interface{}{
			"id":    "table-demo",
			"name":  "Table Demo",
			"title": "table-demo - Table Demo",
		},
	}

	// This is what the fixed prepareTUIListData produces
	externalData := map[string]interface{}{
		"items":      tuiItems,
		"tuiItems":   tuiItems,
		"totalCount": len(tuiItems),
	}

	// Simulate tui-list.tui.yao config
	cfg := &Config{
		ID:   "__yao.tui-list",
		Name: "TUI List",
		Data: map[string]interface{}{
			"title":       "Available TUI Configurations",
			"description": "Select a TUI to run",
			"items":       []interface{}{}, // Will be overwritten by external data
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "tuiList",
					Bind: "items",
					Props: map[string]interface{}{
						"height":           20,
						"width":            80,
						"showTitle":        true,
						"showStatusBar":    true,
						"showFilter":       true,
						"filteringEnabled": true,
					},
				},
			},
		},
	}

	// Step 1: PrepareInitialState - this merges external data into cfg.Data
	PrepareInitialState(cfg, externalData)

	// Check what's in cfg.Data after merging
	fmt.Printf("=== After PrepareInitialState ===\n")
	fmt.Printf("cfg.Data[items] type: %T\n", cfg.Data["items"])
	fmt.Printf("cfg.Data[items] value: %v\n", cfg.Data["items"])

	// Step 2: NewModel - copies cfg.Data to model.State
	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Check what's in model.State
	fmt.Printf("\n=== After NewModel ===\n")
	model.StateMu.RLock()
	fmt.Printf("model.State[items] type: %T\n", model.State["items"])
	fmt.Printf("model.State[items] value: %v\n", model.State["items"])
	model.StateMu.RUnlock()

	// Step 3: Simulate props resolution (as happens during render)
	child := &model.Config.Layout.Children[0]
	props := model.resolveProps(child)

	fmt.Printf("\n=== Props Resolution ===\n")
	fmt.Printf("props[__bind_data] type: %T\n", props["__bind_data"])
	fmt.Printf("props[__bind_data] value: %v\n", props["__bind_data"])

	// Step 4: Parse list props - THIS IS WHERE THE BUG IS
	listProps := component.ParseListPropsWithBinding(props)

	fmt.Printf("\n=== ParseListPropsWithBinding Result ===\n")
	fmt.Printf("Items count: %d\n", len(listProps.Items))
	for i, item := range listProps.Items {
		fmt.Printf("  Item %d: Title='%s'\n", i, item.Title())
	}

	// Assert that items are correctly parsed
	assert.Equal(t, 3, len(listProps.Items), "Should have 3 items from tuiIDs")
}

// TestTUIListCommandFlowWithInterfaceSlice tests with []interface{} instead of []string
// This demonstrates the correct data format
func TestTUIListCommandFlowWithInterfaceSlice(t *testing.T) {
	// Use []interface{} instead of []string
	tuiIDs := []interface{}{"demo", "list-simple", "table-demo"}

	externalData := map[string]interface{}{
		"items":      tuiIDs,
		"totalCount": len(tuiIDs),
	}

	cfg := &Config{
		ID:   "__yao.tui-list",
		Name: "TUI List",
		Data: map[string]interface{}{
			"title": "Available TUI Configurations",
			"items": []interface{}{},
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "tuiList",
					Bind: "items",
					Props: map[string]interface{}{
						"height": 20,
						"width":  80,
					},
				},
			},
		},
	}

	PrepareInitialState(cfg, externalData)
	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Simulate window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	child := &model.Config.Layout.Children[0]
	props := model.resolveProps(child)

	fmt.Printf("\n=== TestTUIListCommandFlowWithInterfaceSlice ===\n")
	fmt.Printf("props[__bind_data] type: %T\n", props["__bind_data"])

	listProps := component.ParseListPropsWithBinding(props)
	fmt.Printf("Items count: %d\n", len(listProps.Items))
	for i, item := range listProps.Items {
		fmt.Printf("  Item %d: Title='%s'\n", i, item.Title())
	}

	// This should work
	assert.Equal(t, 3, len(listProps.Items), "Should have 3 items with []interface{}")
}

// TestTUIListCommandFlowWithMapSlice tests with []map[string]interface{} (detailed items)
func TestTUIListCommandFlowWithMapSlice(t *testing.T) {
	// Use detailed item format like tuiItems
	tuiItems := []interface{}{
		map[string]interface{}{
			"id":      "demo",
			"name":    "Demo TUI",
			"command": "yao tui demo",
		},
		map[string]interface{}{
			"id":      "list-simple",
			"name":    "Simple List",
			"command": "yao tui list-simple",
		},
	}

	externalData := map[string]interface{}{
		"items":      tuiItems,
		"totalCount": len(tuiItems),
	}

	cfg := &Config{
		ID:   "__yao.tui-list",
		Name: "TUI List",
		Data: map[string]interface{}{
			"title": "Available TUI Configurations",
			"items": []interface{}{},
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "tuiList",
					Bind: "items",
					Props: map[string]interface{}{
						"height":       20,
						"width":        80,
						"itemTemplate": "{{id}} - {{name}}",
					},
				},
			},
		},
	}

	PrepareInitialState(cfg, externalData)
	model := NewModel(cfg, nil)
	model.InitializeComponents()

	child := &model.Config.Layout.Children[0]
	props := model.resolveProps(child)

	fmt.Printf("\n=== TestTUIListCommandFlowWithMapSlice ===\n")
	fmt.Printf("props[__bind_data] type: %T\n", props["__bind_data"])
	fmt.Printf("props[itemTemplate]: %v\n", props["itemTemplate"])

	listProps := component.ParseListPropsWithBinding(props)
	fmt.Printf("Items count: %d\n", len(listProps.Items))
	for i, item := range listProps.Items {
		fmt.Printf("  Item %d: Title='%s'\n", i, item.Title())
	}

	// Verify items are parsed correctly with template
	assert.Equal(t, 2, len(listProps.Items), "Should have 2 items")
	assert.Equal(t, "demo - Demo TUI", listProps.Items[0].Title(), "Item 0 should use template")
	assert.Equal(t, "list-simple - Simple List", listProps.Items[1].Title(), "Item 1 should use template")
}
