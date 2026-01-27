package tui

import (
	"encoding/json"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	tuiruntime "github.com/yaoapp/yao/tui/runtime"
)

// TestTableFocusNavigation tests Tab navigation between two tables
func TestTableFocusNavigation(t *testing.T) {
	// Create config with two tables
	config := &Config{
		Name: "Two Tables Focus Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "table",
					ID:   "table1",
					Props: map[string]interface{}{
						"columns": []interface{}{
							map[string]interface{}{"key": "id", "title": "ID", "width": 10},
							map[string]interface{}{"key": "name", "title": "Name", "width": 20},
						},
						"height": 5,
					},
				},
				{
					Type: "table",
					ID:   "table2",
					Props: map[string]interface{}{
						"columns": []interface{}{
							map[string]interface{}{"key": "id", "title": "ID", "width": 10},
							map[string]interface{}{"key": "value", "title": "Value", "width": 20},
						},
						"height": 5,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "Item 1"},
				map[string]interface{}{"id": 2, "name": "Item 2"},
			},
			"items": []interface{}{
				map[string]interface{}{"id": "A", "value": "Value A"},
				map[string]interface{}{"id": "B", "value": "Value B"},
			},
		},
	}

	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Trigger rendering by calling View() - this updates the focus list
	_ = model.View()

	t.Logf("UseRuntime: %v", model.UseRuntime)
	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))

	// Debug: Check each child
	for i, child := range model.RuntimeRoot.Children {
		t.Logf("Child[%d]: ID=%s, Type=%s", i, child.ID, child.Type)
		if child.Component != nil && child.Component.Instance != nil {
			t.Logf("  Component instance: %T", child.Component.Instance)
			// Check if wrapper implements FocusableComponent
			if focusable, ok := child.Component.Instance.(tuiruntime.FocusableComponent); ok {
				t.Logf("  IsFocusable: %v", focusable.IsFocusable())
			}
		}
	}

	// Check the layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult has %d boxes", len(result.Boxes))
	for i, box := range result.Boxes {
		t.Logf("Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
	}

	// Check focus list
	t.Logf("Focus list: %v", model.runtimeFocusList)
	if len(model.runtimeFocusList) == 0 {
		t.Fatal("Focus list should not be empty - tables should be focusable")
	}

	// Verify both tables are in the focus list
	hasTable1 := false
	hasTable2 := false
	for _, id := range model.runtimeFocusList {
		if id == "table1" || id == "comp_table_0" {
			hasTable1 = true
		}
		if id == "table2" || id == "comp_table_1" {
			hasTable2 = true
		}
	}

	if !hasTable1 {
		t.Errorf("Focus list should contain table1, got: %v", model.runtimeFocusList)
	}
	if !hasTable2 {
		t.Errorf("Focus list should contain table2, got: %v", model.runtimeFocusList)
	}

	// Test Tab navigation
	// Press Tab to move to first focusable component
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*Model)

	// Execute any commands from the update
	if cmd != nil {
		msg := cmd()
		newModel, _ = model.Update(msg)
		model = newModel.(*Model)
	}

	t.Logf("After first Tab, CurrentFocus: %s", model.CurrentFocus)
	if model.CurrentFocus == "" {
		t.Error("Focus should be set after pressing Tab")
	}

	// Press Tab again to move to next component
	firstFocus := model.CurrentFocus
	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*Model)

	if cmd != nil {
		msg := cmd()
		newModel, _ = model.Update(msg)
		model = newModel.(*Model)
	}

	t.Logf("After second Tab, CurrentFocus: %s (was %s)", model.CurrentFocus, firstFocus)
	if model.CurrentFocus == firstFocus && len(model.runtimeFocusList) > 1 {
		t.Error("Focus should have moved to next component")
	}

	t.Log("Table focus navigation test passed!")
}

// TestTableFocusWithConfigFile tests focus with the actual table.tui.yao file
func TestTableFocusWithConfigFile(t *testing.T) {
	config := loadConfigFromFile("demo/tuis/table.tui.yao", t)

	model := NewModel(config, nil)
	model.Width = 100
	model.Height = 30

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Trigger rendering by calling View() - this updates the focus list
	_ = model.View()

	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	t.Logf("Focus list: %v", model.runtimeFocusList)

	if len(model.runtimeFocusList) == 0 {
		t.Error("Focus list should not be empty for table.tui.yao")
	}

	// Check that tables are in focus list
	tableCount := 0
	for _, id := range model.runtimeFocusList {
		if id == "comp_table_0" || id == "comp_table_1" {
			tableCount++
		}
	}

	if tableCount < 2 {
		t.Errorf("Both tables should be focusable, found %d", tableCount)
	}

	t.Log("Table focus with config file test passed!")
}

func loadConfigFromFile(path string, t *testing.T) *Config {
	// Simple JSON loader - in real code this would be more robust
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}
	return &cfg
}
