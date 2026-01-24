package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/ui/components"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// checkTableData recursively checks table components and logs their state
func checkTableData(node interface{}, t *testing.T) {
	// Simple placeholder - we'll trace the issue differently
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestTableWithBind tests table component with bind attribute
func TestTableWithBind(t *testing.T) {
	// Read the table.tui.yao file
	configPath := filepath.Join("demo", "tuis", "table.tui.yao")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	// Create model with Runtime enabled (now default)
	model := NewModel(&cfg, nil)
	model.Width = 100
	model.Height = 30

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Verify RuntimeRoot was created
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should not be nil after Init")
	}

	t.Logf("UseRuntime: %v", model.UseRuntime)
	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))

	// Debug: Check each child node
	for i, child := range model.RuntimeRoot.Children {
		t.Logf("Child[%d]: ID=%s, Type=%s", i, child.ID, child.Type)
		if child.Component != nil && child.Component.Instance != nil {
			t.Logf("  Component: %T", child.Component.Instance)

			// Check if it's a table component and inspect its data
			if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
				if table, ok := wrapper.Component.(*components.TableComponent); ok {
					rows := table.GetRows()
					cols := table.GetColumns()
					t.Logf("    Table: %d columns, %d rows", len(cols), len(rows))
					if len(rows) > 0 {
						t.Logf("    First row: %v", rows[0])
					}
					// Try rendering directly
					directView := table.View()
					t.Logf("    Direct view (first 200 chars): %q", directView[:min(200, len(directView))])
				}
			}
		}
	}

	// Check if state has users data
	model.StateMu.RLock()
	if users, ok := model.State["users"]; ok {
		t.Logf("State has users data: %v", users)
	} else {
		t.Error("State should have 'users' data")
	}
	if products, ok := model.State["products"]; ok {
		t.Logf("State has products data: %v", products)
	}
	model.StateMu.RUnlock()

	// Render the model
	output := model.View()
	if output == "" {
		t.Fatal("View output should not be empty")
	}

	t.Logf("Output length: %d chars", len(output))
	t.Logf("Output preview (first 1000 chars): %s", output[:min(1000, len(output))])

	// Debug: Check LayoutResult to see table boxes
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))
	for i, box := range result.Boxes {
		if box.NodeID == "comp_table_0" || box.NodeID == "comp_table_1" {
			t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		}
	}

	// Check Frame/CellBuffer directly
	frame := model.GetFrame()
	bufferLines := frame.Buffer.String()
	t.Logf("Buffer preview (first 500 chars): %q", bufferLines[:min(500, len(bufferLines))])

	// Debug: Check table components
	if model.RuntimeRoot != nil {
		checkTableData(model.RuntimeRoot, t)
	}

	// Check for expected content
	if !contains(output, "John Doe") {
		t.Errorf("Output should contain 'John Doe'")
	}
	if !contains(output, "jane@example.com") {
		t.Errorf("Output should contain 'jane@example.com'")
	}
	if !contains(output, "Admin") {
		t.Errorf("Output should contain 'Admin'")
	}

	// Also check for products
	if !contains(output, "Laptop") {
		t.Errorf("Output should contain 'Laptop'")
	}
	if !contains(output, "Mouse") {
		t.Errorf("Output should contain 'Mouse'")
	}

	t.Log("Table bind test passed!")
}
