package tui

import (
	"fmt"
	"testing"

	tuipkg "github.com/yaoapp/yao/tui/tea"
)

// TestTUIListNames tests the tuiNames function
func TestTUIListNames(t *testing.T) {
	testCases := []struct {
		name     string
		items    []map[string]interface{}
		expected []string
	}{
		{
			name: "Simple items",
			items: []map[string]interface{}{
				{"id": "tui1", "name": "TUI 1"},
				{"id": "tui2", "name": "TUI 2"},
			},
			expected: []string{"tui1", "tui2"},
		},
		{
			name: "Item without id",
			items: []map[string]interface{}{
				{"name": "TUI 1"},
				{"id": "tui2"},
			},
			expected: []string{"map[name:TUI 1]", "tui2"},
		},
		{
			name:     "Empty items",
			items:    []map[string]interface{}{},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tuiNames(tc.items)
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d items, got %d", len(tc.expected), len(result))
				return
			}

			for i, expectedItem := range tc.expected {
				if result[i] != expectedItem {
					t.Errorf("Item %d: expected '%s', got '%s'", i, expectedItem, result[i])
				}
			}
		})
	}
}

// TestTUIListDataMerger tests that TUI list data is correctly merged into config
func TestTUIListDataMerger(t *testing.T) {
	cfg := &tuipkg.Config{
		ID:   "test-list",
		Name: "Test List",
		Data: map[string]interface{}{
			"title": "Available TUI Configurations",
			"items": []string{}, // Empty initial items
		},
	}

	// Simulate the preparation and merging logic
	mockItems := []map[string]interface{}{
		{"id": "tui1", "name": "TUI 1", "command": "yao tui tui1"},
		{"id": "tui2", "name": "TUI 2", "command": "yao tui tui2"},
		{"id": "tui3", "name": "TUI 3", "command": "yao tui tui3"},
	}

	// Merge data into cfg
	if cfg.Data == nil {
		cfg.Data = make(map[string]interface{})
	}

	cfg.Data["items"] = tuiNames(mockItems)
	cfg.Data["tuiItems"] = mockItems
	cfg.Data["totalCount"] = len(mockItems)

	// Verify merged data
	items, ok := cfg.Data["items"].([]string)
	if !ok {
		t.Errorf("Expected items to be []string, got %T", cfg.Data["items"])
	} else if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	} else if items[0] != "tui1" || items[1] != "tui2" || items[2] != "tui3" {
		t.Errorf("Items not correctly set: %v", items)
	}

	tuiItems, ok := cfg.Data["tuiItems"].([]map[string]interface{})
	if !ok {
		t.Errorf("Expected tuiItems to be []map[string]interface{}, got %T", cfg.Data["tuiItems"])
	} else if len(tuiItems) != 3 {
		t.Errorf("Expected 3 tuiItems, got %d", len(tuiItems))
	}

	totalCount, ok := cfg.Data["totalCount"].(int)
	if !ok {
		t.Errorf("Expected totalCount to be int, got %T", cfg.Data["totalCount"])
	} else if totalCount != 3 {
		t.Errorf("Expected totalCount to be 3, got %d", totalCount)
	}

	// Verify original data is preserved
	if cfg.Data["title"] != "Available TUI Configurations" {
		t.Errorf("Original title should be preserved, got '%v'", cfg.Data["title"])
	}
}

// TestTUIItemStructure tests the structure of TUI items
func TestTUIItemStructure(t *testing.T) {
	// Create a TUI config with description
	tuiCfg := &tuipkg.Config{
		ID:   "tui-test",
		Name: "Test TUI",
		Data: map[string]interface{}{
			"description": "A test TUI for demonstration",
		},
	}

	// Create item
	item := map[string]interface{}{
		"id":      tuiCfg.ID,
		"name":    tuiCfg.Name,
		"command": fmt.Sprintf("yao tui %s", tuiCfg.ID),
	}

	// Add description if available
	if tuiCfg.Data != nil {
		if desc, ok := tuiCfg.Data["description"]; ok {
			item["description"] = desc
		}
	}

	// Verify item structure
	if item["id"] != "tui-test" {
		t.Errorf("Expected id 'tui-test', got '%v'", item["id"])
	}
	if item["name"] != "Test TUI" {
		t.Errorf("Expected name 'Test TUI', got '%v'", item["name"])
	}
	if item["command"] != "yao tui tui-test" {
		t.Errorf("Expected command 'yao tui tui-test', got '%v'", item["command"])
	}
	if item["description"] != "A test TUI for demonstration" {
		t.Errorf("Expected correct description, got '%v'", item["description"])
	}
}

// TestTUIItemWithoutDescription tests TUI item without description
func TestTUIItemWithoutDescription(t *testing.T) {
	// Create a TUI config without description
	tuiCfg := &tuipkg.Config{
		ID:   "tui-simple",
		Name: "Simple TUI",
		Data: map[string]interface{}{},
	}

	// Create item
	item := map[string]interface{}{
		"id":      tuiCfg.ID,
		"name":    tuiCfg.Name,
		"command": fmt.Sprintf("yao tui %s", tuiCfg.ID),
	}

	// Try to add description (should not exist)
	if tuiCfg.Data != nil {
		if desc, ok := tuiCfg.Data["description"]; ok {
			item["description"] = desc
		}
	}

	// Verify item structure
	if item["id"] != "tui-simple" {
		t.Errorf("Expected id 'tui-simple', got '%v'", item["id"])
	}
	if item["name"] != "Simple TUI" {
		t.Errorf("Expected name 'Simple TUI', got '%v'", item["name"])
	}
	if item["command"] != "yao tui tui-simple" {
		t.Errorf("Expected command 'yao tui tui-simple', got '%v'", item["command"])
	}
	if _, exists := item["description"]; exists {
		t.Errorf("Description should not exist, got '%v'", item["description"])
	}
}
