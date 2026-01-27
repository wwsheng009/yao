package dsl

import (
	"testing"

	tuiruntime "github.com/yaoapp/yao/tui/tui/runtime"
)

// TestToLayoutNode tests converting DSL Node to runtime LayoutNode.
func TestToLayoutNode(t *testing.T) {
	dslNode := &Node{
		ID:   "root",
		Type: "row",
		Style: &NodeStyle{
			Width:     "100%",
			Height:    "auto",
			FlexGrow:  1.0,
			Direction: "horizontal",
			Padding:   []int{1, 2, 1, 2},
			Gap:       2,
		},
		Children: []*Node{
			{
				ID:    "child1",
				Type:  "text",
				Width: "50%",
				Props: map[string]interface{}{
					"content": "Hello",
				},
			},
			{
				ID:    "child2",
				Type:  "text",
				Width: "50%",
				Props: map[string]interface{}{
					"content": "World",
				},
			},
		},
	}

	layoutNode := dslNode.ToLayoutNode()

	if layoutNode == nil {
		t.Fatal("ToLayoutNode returned nil")
	}

	if layoutNode.ID != "root" {
		t.Errorf("Expected ID 'root', got '%s'", layoutNode.ID)
	}

	if layoutNode.Type != tuiruntime.NodeTypeRow {
		t.Errorf("Expected type NodeTypeRow, got %v", layoutNode.Type)
	}

	if layoutNode.Style.Direction != tuiruntime.DirectionRow {
		t.Errorf("Expected DirectionRow, got %v", layoutNode.Style.Direction)
	}

	if layoutNode.Style.FlexGrow != 1.0 {
		t.Errorf("Expected FlexGrow=1.0, got %v", layoutNode.Style.FlexGrow)
	}

	// Check padding
	expectedPadding := tuiruntime.Insets{Top: 1, Right: 2, Bottom: 1, Left: 2}
	if layoutNode.Style.Padding != expectedPadding {
		t.Errorf("Expected padding %+v, got %+v", expectedPadding, layoutNode.Style.Padding)
	}

	// Check children
	if len(layoutNode.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layoutNode.Children))
	}

	// Check that children have proper parent references
	for _, child := range layoutNode.Children {
		if child.Parent != layoutNode {
			t.Error("Child does not have proper parent reference")
		}
	}
}

// TestToLayoutNodeWithPercentage tests percentage width/height conversion.
func TestToLayoutNodeWithPercentage(t *testing.T) {
	dslNode := &Node{
		ID:     "test",
		Type:   "column",
		Width:  "50%",
		Height: "100%",
	}

	layoutNode := dslNode.ToLayoutNode()

	// Percentage is encoded as negative value: -50 means 50%
	if layoutNode.Style.Width != -50 {
		t.Errorf("Expected Width=-50 (50%%), got %d", layoutNode.Style.Width)
	}

	if layoutNode.Style.Height != -100 {
		t.Errorf("Expected Height=-100 (100%%), got %d", layoutNode.Style.Height)
	}
}

// TestToLayoutNodeWithFlex tests flex size conversion.
func TestToLayoutNodeWithFlex(t *testing.T) {
	dslNode := &Node{
		ID:     "test",
		Type:   "row",
		Height: "flex",
	}

	layoutNode := dslNode.ToLayoutNode()

	if layoutNode.Style.FlexGrow != 1 {
		t.Errorf("Expected FlexGrow=1 for 'flex', got %v", layoutNode.Style.FlexGrow)
	}
}

// TestToLayoutNodeWithBorder tests border conversion.
func TestToLayoutNodeWithBorder(t *testing.T) {
	tests := []struct {
		name   string
		border interface{}
		want   tuiruntime.Insets
	}{
		{
			name:   "uniform border",
			border: 2,
			want:   tuiruntime.Insets{Top: 2, Right: 2, Bottom: 2, Left: 2},
		},
		{
			name:   "array border",
			border: []interface{}{1, 2, 3, 4},
			want:   tuiruntime.Insets{Top: 1, Right: 2, Bottom: 3, Left: 4},
		},
		{
			name: "spec border",
			border: &BorderSpec{
				Top:    1,
				Right:  2,
				Bottom: 3,
				Left:   4,
			},
			want: tuiruntime.Insets{Top: 1, Right: 2, Bottom: 3, Left: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dslNode := &Node{
				ID:         "test",
				Type:       "row",
				Border:     toBorderSpec(tt.border),
				BorderWidth: tt.border,
			}

			layoutNode := dslNode.ToLayoutNode()

			if layoutNode.Style.Border != tt.want {
				t.Errorf("Expected border %+v, got %+v", tt.want, layoutNode.Style.Border)
			}
		})
	}
}

// toBorderSpec is a helper to convert interface{} to BorderSpec for testing.
func toBorderSpec(v interface{}) *BorderSpec {
	switch val := v.(type) {
	case *BorderSpec:
		return val
	default:
		return nil
	}
}

// TestToLayoutNodeWithAlignment tests alignment properties.
func TestToLayoutNodeWithAlignment(t *testing.T) {
	dslNode := &Node{
		ID:         "test",
		Type:       "row",
		AlignItems: "center",
		Justify:    "space-between",
	}

	layoutNode := dslNode.ToLayoutNode()

	if layoutNode.Style.AlignItems != tuiruntime.AlignCenter {
		t.Errorf("Expected AlignCenter, got %v", layoutNode.Style.AlignItems)
	}

	if layoutNode.Style.Justify != tuiruntime.JustifySpaceBetween {
		t.Errorf("Expected JustifySpaceBetween, got %v", layoutNode.Style.Justify)
	}
}

// TestToLayoutTree tests converting entire config to layout tree.
func TestToLayoutTree(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: &Node{
			ID:   "root",
			Type: "column",
			Children: []*Node{
				{ID: "child1", Type: "text"},
				{ID: "child2", Type: "text"},
			},
		},
	}

	root := cfg.ToLayoutTree()

	if root == nil {
		t.Fatal("ToLayoutTree returned nil")
	}

	if len(root.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(root.Children))
	}
}

// TestValidateAndConvert tests the combined validate and convert function.
func TestValidateAndConvert(t *testing.T) {
	jsonData := []byte(`{
		"name": "Test TUI",
		"layout": {
			"type": "row",
			"children": [
				{"type": "text"}
			]
		}
	}`)

	cfg, root, err := ValidateAndConvert(jsonData, "test.json")
	if err != nil {
		t.Fatalf("ValidateAndConvert failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Config is nil")
	}

	if root == nil {
		t.Fatal("LayoutNode is nil")
	}

	if cfg.Name != "Test TUI" {
		t.Errorf("Expected name 'Test TUI', got '%s'", cfg.Name)
	}
}

// TestMapDSLTypeToRuntime tests type mapping.
func TestMapDSLTypeToRuntime(t *testing.T) {
	tests := []struct {
		dslType       string
		expectedType  tuiruntime.NodeType
	}{
		{"row", tuiruntime.NodeTypeRow},
		{"column", tuiruntime.NodeTypeColumn},
		{"vertical", tuiruntime.NodeTypeFlex},
		{"horizontal", tuiruntime.NodeTypeFlex},
		{"layout", tuiruntime.NodeTypeFlex},
		{"text", tuiruntime.NodeTypeCustom},
		{"header", tuiruntime.NodeTypeCustom},
		{"menu", tuiruntime.NodeTypeCustom},
		{"progress", tuiruntime.NodeTypeCustom},
	}

	for _, tt := range tests {
		t.Run(tt.dslType, func(t *testing.T) {
			result := mapDSLTypeToRuntime(tt.dslType)
			if result != tt.expectedType {
				t.Errorf("mapDSLTypeToRuntime(%s) = %v, want %v", tt.dslType, result, tt.expectedType)
			}
		})
	}
}

// TestMapDirection tests direction mapping.
func TestMapDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected tuiruntime.Direction
	}{
		{"row", tuiruntime.DirectionRow},
		{"horizontal", tuiruntime.DirectionRow},
		{"column", tuiruntime.DirectionColumn},
		{"vertical", tuiruntime.DirectionColumn},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapDirection(tt.input)
			if result != tt.expected {
				t.Errorf("mapDirection(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMapAlign tests align mapping.
func TestMapAlign(t *testing.T) {
	tests := []struct {
		input    string
		expected tuiruntime.Align
	}{
		{"start", tuiruntime.AlignStart},
		{"center", tuiruntime.AlignCenter},
		{"end", tuiruntime.AlignEnd},
		{"stretch", tuiruntime.AlignStretch},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapAlign(tt.input)
			if result != tt.expected {
				t.Errorf("mapAlign(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMapJustify tests justify mapping.
func TestMapJustify(t *testing.T) {
	tests := []struct {
		input    string
		expected tuiruntime.Justify
	}{
		{"start", tuiruntime.JustifyStart},
		{"center", tuiruntime.JustifyCenter},
		{"end", tuiruntime.JustifyEnd},
		{"space-between", tuiruntime.JustifySpaceBetween},
		{"space-around", tuiruntime.JustifySpaceAround},
		{"space-evenly", tuiruntime.JustifySpaceEvenly},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapJustify(tt.input)
			if result != tt.expected {
				t.Errorf("mapJustify(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMapOverflow tests overflow mapping.
func TestMapOverflow(t *testing.T) {
	tests := []struct {
		input    string
		expected tuiruntime.Overflow
	}{
		{"visible", tuiruntime.OverflowVisible},
		{"hidden", tuiruntime.OverflowHidden},
		{"scroll", tuiruntime.OverflowScroll},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapOverflow(tt.input)
			if result != tt.expected {
				t.Errorf("mapOverflow(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToPropsConversion tests props are properly copied.
func TestToPropsConversion(t *testing.T) {
	dslNode := &Node{
		ID:    "test",
		Type:  "text",
		Props: map[string]interface{}{
			"content": "Hello",
			"color":   "blue",
			"bold":    true,
		},
	}

	layoutNode := dslNode.ToLayoutNode()

	if len(layoutNode.Props) != 3 {
		t.Fatalf("Expected 3 props, got %d", len(layoutNode.Props))
	}

	if layoutNode.Props["content"] != "Hello" {
		t.Errorf("Expected content 'Hello', got %v", layoutNode.Props["content"])
	}

	if layoutNode.Props["color"] != "blue" {
		t.Errorf("Expected color 'blue', got %v", layoutNode.Props["color"])
	}

	if layoutNode.Props["bold"] != true {
		t.Errorf("Expected bold=true, got %v", layoutNode.Props["bold"])
	}
}
