package dsl

import (
	"os"
	"testing"
)

// TestParseNestedLayoutFile tests parsing the actual nested.tui.yao file.
func TestParseNestedLayoutFile(t *testing.T) {
	data, err := os.ReadFile("../demo/tuis/layouts/nested.tui.yao")
	if err != nil {
		t.Skip("Nested layout file not found:", err)
		return
	}

	cfg, root, err := ValidateAndConvert(data, "nested.tui.yao")
	if err != nil {
		t.Fatalf("ValidateAndConvert failed: %v", err)
	}

	if cfg.Name != "Nested Layouts Test" {
		t.Errorf("Expected name 'Nested Layouts Test', got '%s'", cfg.Name)
	}

	stats := cfg.GetLayoutStats()
	t.Logf("Layout Stats: TotalNodes=%d, ContainerNodes=%d, LeafNodes=%d, MaxDepth=%d",
		stats.TotalNodes, stats.ContainerNodes, stats.LeafNodes, stats.MaxDepth)

	if root == nil {
		t.Fatal("Root LayoutNode is nil")
	}

	if len(root.Children) == 0 {
		t.Error("Root has no children")
	}

	t.Logf("Successfully parsed nested layout with %d nodes", stats.TotalNodes)
}
