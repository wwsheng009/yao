package dsl

import (
	"testing"
)

// TestParseJSON tests parsing JSON configuration.
func TestParseJSON(t *testing.T) {
	jsonData := []byte(`{
		"name": "Test TUI",
		"data": {
			"title": "Hello World"
		},
		"layout": {
			"type": "row",
			"direction": "horizontal",
			"children": [
				{
					"type": "text",
					"props": {
						"content": "Left"
					},
					"width": "50%"
				},
				{
					"type": "text",
					"props": {
						"content": "Right"
					},
					"width": "50%"
				}
			]
		}
	}`)

	cfg, err := ParseJSON(jsonData)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if cfg.Name != "Test TUI" {
		t.Errorf("Expected name 'Test TUI', got '%s'", cfg.Name)
	}

	if cfg.Layout == nil {
		t.Fatal("Layout is nil")
	}

	if cfg.Layout.Type != "row" {
		t.Errorf("Expected layout type 'row', got '%s'", cfg.Layout.Type)
	}

	if len(cfg.Layout.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(cfg.Layout.Children))
	}
}

// TestParseYAML tests parsing YAML configuration.
func TestParseYAML(t *testing.T) {
	yamlData := []byte(`
name: Test TUI
data:
  title: Hello World
  stats:
    totalUsers: 1250
    activeUsers: 890
layout:
  type: column
  direction: vertical
  children:
    - type: header
      props:
        title: "{{title}}"
        align: center
    - type: text
      props:
        content: "Total: {{stats.totalUsers}}"
bindings:
  q:
    process: "tui.quit"
`)

	cfg, err := ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if cfg.Name != "Test TUI" {
		t.Errorf("Expected name 'Test TUI', got '%s'", cfg.Name)
	}

	if cfg.Layout == nil {
		t.Fatal("Layout is nil")
	}

	if cfg.Layout.Type != "column" {
		t.Errorf("Expected layout type 'column', got '%s'", cfg.Layout.Type)
	}

	if len(cfg.Layout.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(cfg.Layout.Children))
	}

	// Check bindings
	if cfg.Bindings == nil {
		t.Fatal("Bindings is nil")
	}

	if len(cfg.Bindings) != 1 {
		t.Fatalf("Expected 1 binding, got %d", len(cfg.Bindings))
	}
}

// TestParseAutoDetect tests automatic format detection.
func TestParseAutoDetect(t *testing.T) {
	jsonData := []byte(`{"name": "JSON Test", "layout": {"type": "row"}}`)
	yamlData := []byte(`
name: YAML Test
layout:
  type: column
`)

	// Test JSON
	cfg1, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Parse (JSON) failed: %v", err)
	}
	if cfg1.Name != "JSON Test" {
		t.Errorf("Expected name 'JSON Test', got '%s'", cfg1.Name)
	}

	// Test YAML
	cfg2, err := Parse(yamlData)
	if err != nil {
		t.Fatalf("Parse (YAML) failed: %v", err)
	}
	if cfg2.Name != "YAML Test" {
		t.Errorf("Expected name 'YAML Test', got '%s'", cfg2.Name)
	}
}

// TestValidate tests configuration validation.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name: "valid config",
			data: []byte(`{
				"name": "Valid",
				"layout": {
					"type": "row",
					"children": [{"type": "text"}]
				}
			}`),
			wantErr: false,
		},
		{
			name:    "missing name",
			data:    []byte(`{"layout": {"type": "row"}}`),
			wantErr: true,
		},
		{
			name:    "missing layout",
			data:    []byte(`{"name": "Test"}`),
			wantErr: true,
		},
		{
			name: "valid with auto direction",
			data: []byte(`{
				"name": "Valid",
				"layout": {
					"type": "layout"
				}
			}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse(tt.data)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			err = cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAssignIDs tests automatic ID assignment.
func TestAssignIDs(t *testing.T) {
	data := []byte(`{
		"name": "Test",
		"layout": {
			"type": "row",
			"children": [
				{"type": "text"},
				{"type": "text", "id": "custom-id"},
				{"type": "text"}
			]
		}
	}`)

	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	counter := 0
	cfg.Layout.AssignIDs("node", &counter)

	if cfg.Layout.ID == "" {
		t.Error("Root node should have an ID")
	}

	children := cfg.Layout.Children
	if len(children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(children))
	}

	if children[0].ID == "" {
		t.Error("First child should have auto-generated ID")
	}

	if children[1].ID != "custom-id" {
		t.Errorf("Second child should keep custom ID, got '%s'", children[1].ID)
	}

	if children[2].ID == "" {
		t.Error("Third child should have auto-generated ID")
	}
}

// TestFlattenData tests data flattening.
func TestFlattenData(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
		"stats": map[string]interface{}{
			"total": 100,
			"active": map[string]interface{}{
				"count": 50,
			},
		},
	}

	result := FlattenData(data)

	expectedKeys := []string{
		"user.name",
		"user.age",
		"stats.total",
		"stats.active.count",
	}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Expected key '%s' not found in result", key)
		}
	}
}

// TestParseSize tests size parsing.
func TestParseSize(t *testing.T) {
	tests := []struct {
		name       string
		value      interface{}
		wantSize   int
		wantPercent bool
		wantFlex   bool
		wantAuto   bool
	}{
		{"integer", 100, 100, false, false, false},
		{"float", 50.5, 50, false, false, false},
		{"percentage string", "50%", -50, true, false, false},
		{"percentage string 100", "100%", -100, true, false, false},
		{"flex", "flex", 0, false, true, false},
		{"auto", "auto", 0, false, false, true},
		{"nil", nil, 0, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, isPercent, isFlex, isAuto := ParseSize(tt.value)
			if size != tt.wantSize {
				t.Errorf("ParseSize() size = %d, want %d", size, tt.wantSize)
			}
			if isPercent != tt.wantPercent {
				t.Errorf("ParseSize() isPercent = %v, want %v", isPercent, tt.wantPercent)
			}
			if isFlex != tt.wantFlex {
				t.Errorf("ParseSize() isFlex = %v, want %v", isFlex, tt.wantFlex)
			}
			if isAuto != tt.wantAuto {
				t.Errorf("ParseSize() isAuto = %v, want %v", isAuto, tt.wantAuto)
			}
		})
	}
}

// TestParseBorder tests border parsing.
func TestParseBorder(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  BorderSpec
	}{
		{"single int", 2, BorderSpec{Top: 2, Right: 2, Bottom: 2, Left: 2}},
		{"array 1", []interface{}{1}, BorderSpec{Top: 1, Right: 1, Bottom: 1, Left: 1}},
		{"array 2", []interface{}{1, 2}, BorderSpec{Top: 1, Right: 2, Bottom: 1, Left: 2}},
		{"array 4", []interface{}{1, 2, 3, 4}, BorderSpec{Top: 1, Right: 2, Bottom: 3, Left: 4}},
		{"map", map[string]interface{}{"top": 1, "right": 2}, BorderSpec{Top: 1, Right: 2, Bottom: 0, Left: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseBorder(tt.value)
			if result != tt.want {
				t.Errorf("ParseBorder() = %+v, want %+v", result, tt.want)
			}
		})
	}
}

// TestGetNodeByID tests node lookup by ID.
func TestGetNodeByID(t *testing.T) {
	data := []byte(`{
		"name": "Test",
		"layout": {
			"type": "row",
			"id": "root",
			"children": [
				{
					"type": "column",
					"id": "col1",
					"children": [
						{"type": "text", "id": "text1"}
					]
				},
				{"type": "text", "id": "text2"}
			]
		}
	}`)

	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find root
	if root := cfg.Layout.GetNodeByID("root"); root == nil {
		t.Error("Root node not found")
	}

	// Find nested node
	if text1 := cfg.Layout.GetNodeByID("text1"); text1 == nil {
		t.Error("text1 node not found")
	}

	// Find leaf node
	if text2 := cfg.Layout.GetNodeByID("text2"); text2 == nil {
		t.Error("text2 node not found")
	}

	// Non-existent
	if unknown := cfg.Layout.GetNodeByID("unknown"); unknown != nil {
		t.Error("Should not find unknown node")
	}
}

// TestGetAllNodes tests collecting all nodes.
func TestGetAllNodes(t *testing.T) {
	data := []byte(`{
		"name": "Test",
		"layout": {
			"type": "row",
			"id": "root",
			"children": [
				{"type": "text", "id": "a"},
				{"type": "text", "id": "b"}
			]
		}
	}`)

	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	nodes := cfg.Layout.GetAllNodes()

	if len(nodes) != 3 { // root + 2 children
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}

	expectedIDs := []string{"root", "a", "b"}
	for _, id := range expectedIDs {
		if _, ok := nodes[id]; !ok {
			t.Errorf("Expected node '%s' not found", id)
		}
	}
}

// TestGetLayoutStats tests layout statistics.
func TestGetLayoutStats(t *testing.T) {
	data := []byte(`{
		"name": "Test",
		"layout": {
			"type": "row",
			"children": [
				{
					"type": "column",
					"children": [
						{"type": "text"},
						{"type": "text"}
					]
				},
				{"type": "text"}
			]
		}
	}`)

	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stats := cfg.GetLayoutStats()

	if stats.TotalNodes != 5 { // root + column + 3 text nodes
		t.Errorf("Expected TotalNodes=5, got %d", stats.TotalNodes)
	}

	if stats.ContainerNodes != 2 { // root (row) + column
		t.Errorf("Expected ContainerNodes=2, got %d", stats.ContainerNodes)
	}

	if stats.LeafNodes != 3 { // 3 text nodes
		t.Errorf("Expected LeafNodes=3, got %d", stats.LeafNodes)
	}

	if stats.MaxDepth != 3 {
		t.Errorf("Expected MaxDepth=3, got %d", stats.MaxDepth)
	}
}

// TestToJSON and TestToYAML test serialization.
func TestToJSON(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: &Node{
			Type: "row",
		},
	}

	data, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON returned empty data")
	}
}

func TestToYAML(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: &Node{
			Type: "column",
		},
	}

	data, err := cfg.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ToYAML returned empty data")
	}
}
