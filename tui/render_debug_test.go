package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/ui/components"
)

// TestTableRenderDebug directly tests table rendering in runtime
func TestTableRenderDebug(t *testing.T) {
	config := &Config{
		Name: "Table Render Debug",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "table",
					ID:   "test-table",
					Bind: "users",
					Props: map[string]interface{}{
						"columns": []interface{}{
							map[string]interface{}{"key": "id", "title": "ID", "width": 10},
							map[string]interface{}{"key": "name", "title": "Name", "width": 20},
						},
						"height": 10,
						"showBorder": true,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "John Doe"},
				map[string]interface{}{"id": 2, "name": "Jane Smith"},
			},
		},
	}

	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get the table component directly
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot is nil")
	}

	// Find table node
	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	for i, child := range model.RuntimeRoot.Children {
		t.Logf("Child[%d]: ID=%q, Type=%s", i, child.ID, child.Type)
		if child.Component != nil && child.Component.Instance != nil {
			t.Logf("  Instance: %T", child.Component.Instance)
		}
	}

	var tableNode *runtime.LayoutNode
	for _, child := range model.RuntimeRoot.Children {
		if child.Type == "custom" && child.Component != nil {
			if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
				if _, ok := wrapper.Component.(*components.TableComponent); ok {
					tableNode = child
					break
				}
			}
		}
	}

	if tableNode == nil {
		t.Fatal("Table node not found")
	}

	t.Logf("Found table node: ID=%q", tableNode.ID)

	// Check table component data
	if tableNode.Component != nil && tableNode.Component.Instance != nil {
		if wrapper, ok := tableNode.Component.Instance.(*NativeComponentWrapper); ok {
			if table, ok := wrapper.Component.(*components.TableComponent); ok {
				rows := table.GetRows()
				cols := table.GetColumns()
				t.Logf("Table has %d columns, %d rows", len(cols), len(rows))
				if len(rows) > 0 {
					t.Logf("First row: %v", rows[0])
				}
				// Check direct view
				directView := table.View()
				t.Logf("Direct table view (first 300 chars): %q", directView[:min(300, len(directView))])
			}
		}
	}

	// Check LayoutResult
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))
	for _, box := range result.Boxes {
		if box.NodeID == "test-table" {
			t.Logf("  Table box: X=%d, Y=%d, W=%d, H=%d", box.X, box.Y, box.W, box.H)
		}
	}

	// Render output
	output := model.View()
	t.Logf("Output (first 500 chars): %q", output[:min(500, len(output))])

	if !contains(output, "John Doe") {
		t.Error("Expected 'John Doe' in output")
	}
}

// TestTableRenderTrace traces the rendering process step by step
func TestTableRenderTrace(t *testing.T) {
	config := &Config{
		Name: "Table Render Trace",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "table",
					ID:   "test-table-trace",
					Bind: "users",
					Props: map[string]interface{}{
						"columns": []interface{}{
							map[string]interface{}{"key": "id", "title": "ID", "width": 10},
							map[string]interface{}{"key": "name", "title": "Name", "width": 20},
						},
						"height": 10,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "John Doe"},
				map[string]interface{}{"id": 2, "name": "Jane Smith"},
			},
		},
	}

	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get the table component
	var tableNode *runtime.LayoutNode
	for _, child := range model.RuntimeRoot.Children {
		if child.Type == "custom" && child.Component != nil {
			if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
				if table, ok := wrapper.Component.(*components.TableComponent); ok {
					tableNode = child
					t.Logf("Found table with %d rows", len(table.GetRows()))
					// Get the full view output
					fullView := table.View()
					t.Logf("Full table view:\n%s", fullView)
					t.Logf("View length: %d chars", len(fullView))
					lines := strings.Split(fullView, "\n")
					t.Logf("Number of lines: %d", len(lines))
					for i, line := range lines {
						t.Logf("Line %d: %q (len=%d)", i, line, len(line))
					}
					break
				}
			}
		}
	}

	if tableNode == nil {
		t.Fatal("Table node not found")
	}

	// Now check the CellBuffer content
	frame := model.GetFrame()
	for y := 0; y < 15 && y < frame.Height; y++ {
		lineContent := ""
		for x := 0; x < frame.Width; x++ {
			cell := frame.Buffer.GetContent(x, y)
			lineContent += string(cell.Char)
		}
		t.Logf("Buffer line %d: %q", y, lineContent)
	}

	// Check for John Doe in the buffer
	bufferContent := frame.Buffer.String()
	if !contains(bufferContent, "John Doe") {
		t.Errorf("Buffer should contain 'John Doe', got: %q", bufferContent[:min(500, len(bufferContent))])
	}
}

// TestTableCellBuffer tests table rendering to CellBuffer
func TestTableCellBuffer(t *testing.T) {
	config := &Config{
		Name: "Table CellBuffer Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "table",
					ID:   "test-table-2",
					Bind: "users",
					Props: map[string]interface{}{
						"columns": []interface{}{
							map[string]interface{}{"key": "id", "title": "ID", "width": 10},
							map[string]interface{}{"key": "name", "title": "Name", "width": 20},
						},
						"height": 10,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "John Doe"},
			},
		},
	}

	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Get the frame directly
	frame := model.GetFrame()
	t.Logf("Frame: Width=%d, Height=%d", frame.Width, frame.Height)

	// Check CellBuffer lines
	lines := frame.Buffer.String()
	t.Logf("CellBuffer output (first 500 chars): %q", lines[:min(500, len(lines))])

	if !contains(lines, "John Doe") {
		t.Errorf("Expected 'John Doe' in CellBuffer output")
	}
}
