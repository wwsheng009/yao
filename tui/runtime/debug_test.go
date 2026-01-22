package runtime

import (
	"strings"
	"testing"
)

// TestDebugFrame tests the DebugFrame function.
func TestDebugFrame(t *testing.T) {
	// Create a simple buffer
	buf := NewCellBuffer(40, 10)

	// Add some content
	buf.SetContent(0, 0, 0, 'H', CellStyle{Bold: true}, "test")
	buf.SetContent(1, 0, 0, 'e', CellStyle{}, "test")
	buf.SetContent(2, 0, 0, 'l', CellStyle{}, "test")
	buf.SetContent(3, 0, 0, 'l', CellStyle{}, "test")
	buf.SetContent(4, 0, 0, 'o', CellStyle{}, "test")
	buf.SetContent(0, 1, 0, 'W', CellStyle{}, "test")
	buf.SetContent(1, 1, 0, 'o', CellStyle{}, "test")
	buf.SetContent(2, 1, 0, 'r', CellStyle{}, "test")
	buf.SetContent(3, 1, 0, 'l', CellStyle{}, "test")
	buf.SetContent(4, 1, 0, 'd', CellStyle{}, "test")

	// Create frame
	frame := &Frame{
		Buffer: buf,
		Width:  40,
		Height: 10,
	}

	// Create layout result with boxes
	layoutResult := &LayoutResult{
		Boxes: []LayoutBox{
			{NodeID: "box1", X: 0, Y: 0, W: 5, H: 1},
			{NodeID: "box2", X: 0, Y: 1, W: 5, H: 1},
		},
		RootWidth:  40,
		RootHeight: 10,
	}

	// Create debug info
	debug := DebugFrame(frame, layoutResult)
	if debug == nil {
		t.Fatal("DebugFrame returned nil")
	}

	// Check summary
	if !strings.Contains(debug.Summary, "40x10") {
		t.Errorf("Summary should contain '40x10', got: %s", debug.Summary)
	}

	if !strings.Contains(debug.Summary, "2 boxes") {
		t.Errorf("Summary should contain '2 boxes', got: %s", debug.Summary)
	}

	// Check buffer info
	if debug.BufferInfo == nil {
		t.Fatal("BufferInfo is nil")
	}

	if debug.BufferInfo.Width != 40 {
		t.Errorf("Expected BufferInfo.Width=40, got %d", debug.BufferInfo.Width)
	}

	if debug.BufferInfo.Height != 10 {
		t.Errorf("Expected BufferInfo.Height=10, got %d", debug.BufferInfo.Height)
	}

	if debug.BufferInfo.NonEmpty != 10 { // "Hello" + "World" = 10 chars
		t.Errorf("Expected NonEmpty=10, got %d", debug.BufferInfo.NonEmpty)
	}

	// Check boxes
	if len(debug.Boxes) != 2 {
		t.Errorf("Expected 2 boxes, got %d", len(debug.Boxes))
	}

	if debug.Boxes[0].ID != "box1" {
		t.Errorf("Expected box1 ID='box1', got '%s'", debug.Boxes[0].ID)
	}

	if debug.Boxes[0].Content != "Hello" {
		t.Errorf("Expected box1 content='Hello', got '%s'", debug.Boxes[0].Content)
	}

	if debug.Boxes[1].Content != "World" {
		t.Errorf("Expected box2 content='World', got '%s'", debug.Boxes[1].Content)
	}
}

// TestDebugFrameNil tests DebugFrame with nil inputs.
func TestDebugFrameNil(t *testing.T) {
	debug := DebugFrame(nil, nil)
	if debug != nil {
		t.Error("DebugFrame(nil) should return nil")
	}
}

// TestDebugFrameString tests the String method.
func TestDebugFrameString(t *testing.T) {
	buf := NewCellBuffer(20, 5)
	buf.SetContent(0, 0, 0, 'T', CellStyle{}, "test")
	buf.SetContent(1, 0, 0, 'e', CellStyle{}, "test")
	buf.SetContent(2, 0, 0, 's', CellStyle{}, "test")
	buf.SetContent(3, 0, 0, 't', CellStyle{}, "test")

	frame := &Frame{
		Buffer: buf,
		Width:  20,
		Height: 5,
	}

	layoutResult := &LayoutResult{
		Boxes: []LayoutBox{
			{NodeID: "test_box", X: 0, Y: 0, W: 4, H: 1},
		},
	}

	debug := DebugFrame(frame, layoutResult)
	output := debug.String()

	if !strings.Contains(output, "=== Render Debug ===") {
		t.Error("String output should contain '=== Render Debug ==='")
	}

	if !strings.Contains(output, "20x5") {
		t.Error("String output should contain '20x5'")
	}

	if !strings.Contains(output, "test_box") {
		t.Error("String output should contain 'test_box'")
	}
}

// TestPlainOutput tests the PlainOutput method.
func TestPlainOutput(t *testing.T) {
	buf := NewCellBuffer(10, 3)
	buf.SetContent(0, 0, 0, 'A', CellStyle{}, "test")
	buf.SetContent(1, 0, 0, 'B', CellStyle{}, "test")
	buf.SetContent(0, 1, 0, 'C', CellStyle{}, "test")
	buf.SetContent(1, 1, 0, 'D', CellStyle{}, "test")

	frame := &Frame{
		Buffer: buf,
		Width:  10,
		Height: 3,
	}

	debug := DebugFrame(frame, nil)
	output := debug.PlainOutput()

	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "AB" {
		t.Errorf("Expected line 0 'AB', got '%s'", lines[0])
	}

	if lines[1] != "CD" {
		t.Errorf("Expected line 1 'CD', got '%s'", lines[1])
	}
}

// TestDiffOutput tests the DiffOutput method.
func TestDiffOutput(t *testing.T) {
	buf := NewCellBuffer(10, 2)
	buf.SetContent(0, 0, 0, 'X', CellStyle{}, "test")
	buf.SetContent(1, 0, 0, 'Y', CellStyle{}, "test")

	frame := &Frame{
		Buffer: buf,
		Width:  10,
		Height: 2,
	}

	debug := DebugFrame(frame, nil)
	output := debug.DiffOutput()

	// Debug: print actual output
	t.Logf("DiffOutput:\n%s", output)

	// Should contain line numbers
	if !strings.Contains(output, "1│") {
		t.Error("DiffOutput should contain line number '1│'")
	}

	// Should contain visible markers for empty cells
	if !strings.Contains(output, "·") {
		t.Error("DiffOutput should contain visible markers '·' for empty cells")
	}

	// Should contain the actual content
	if !strings.Contains(output, "XY") {
		t.Error("DiffOutput should contain 'XY'")
	}
}

// TestJSONOutput tests the ToJSON method.
func TestJSONOutput(t *testing.T) {
	buf := NewCellBuffer(8, 2)
	buf.SetContent(0, 0, 0, 'T', CellStyle{}, "test")
	buf.SetContent(1, 0, 0, 'e', CellStyle{}, "test")
	buf.SetContent(2, 0, 0, 's', CellStyle{}, "test")
	buf.SetContent(3, 0, 0, 't', CellStyle{}, "test")

	frame := &Frame{
		Buffer: buf,
		Width:  8,
		Height: 2,
	}

	layoutResult := &LayoutResult{
		Boxes: []LayoutBox{
			{NodeID: "test_box", X: 0, Y: 0, W: 4, H: 1},
		},
	}

	debug := DebugFrame(frame, layoutResult)
	jsonOutput := debug.ToJSON()

	if jsonOutput == nil {
		t.Fatal("ToJSON returned nil")
	}

	if jsonOutput.Width != 8 {
		t.Errorf("Expected Width=8, got %d", jsonOutput.Width)
	}

	if jsonOutput.Height != 2 {
		t.Errorf("Expected Height=2, got %d", jsonOutput.Height)
	}

	if len(jsonOutput.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(jsonOutput.Lines))
	}

	if jsonOutput.Lines[0] != "Test" {
		t.Errorf("Expected line 0 'Test', got '%s'", jsonOutput.Lines[0])
	}

	if jsonOutput.Stats.TotalCells != 16 {
		t.Errorf("Expected TotalCells=16, got %d", jsonOutput.Stats.TotalCells)
	}

	if jsonOutput.Stats.NonEmpty != 4 {
		t.Errorf("Expected NonEmpty=4, got %d", jsonOutput.Stats.NonEmpty)
	}
}

// TestJSONOutputCompare tests the Compare method.
func TestJSONOutputCompare(t *testing.T) {
	output1 := &JSONOutput{
		Width:  10,
		Height: 5,
		Lines:  []string{"Line 1", "Line 2", "Line 3"},
		Boxes:  []BoxJSON{{ID: "box1", X: 0, Y: 0, Width: 5, Height: 1}},
	}

	output2 := &JSONOutput{
		Width:  10,
		Height: 5,
		Lines:  []string{"Line 1", "Line 2", "Line 3"},
		Boxes:  []BoxJSON{{ID: "box1", X: 0, Y: 0, Width: 5, Height: 1}},
	}

	// Identical outputs should have no differences
	diffs := output1.Compare(output2)
	if len(diffs) > 0 {
		t.Errorf("Expected no differences, got %d: %v", len(diffs), diffs)
	}

	// Different lines
	output3 := &JSONOutput{
		Width:  10,
		Height: 5,
		Lines:  []string{"Line 1", "Different", "Line 3"},
	}

	diffs = output1.Compare(output3)
	if len(diffs) == 0 {
		t.Error("Expected differences, got none")
	}

	// Different width
	output4 := &JSONOutput{
		Width:  20,
		Height: 5,
		Lines:  []string{"Line 1", "Line 2", "Line 3"},
	}

	diffs = output1.Compare(output4)
	if len(diffs) == 0 {
		t.Error("Expected differences for width, got none")
	}
}

// TestJSONOutputString tests the JSONOutput String method.
func TestJSONOutputString(t *testing.T) {
	output := &JSONOutput{
		Width:  10,
		Height: 3,
		Lines:  []string{"ABC", "", "DEF"},
	}

	str := output.String()
	if !strings.Contains(str, "Render Output: 10x3") {
		t.Error("String should contain 'Render Output: 10x3'")
	}

	if !strings.Contains(str, "ABC") {
		t.Error("String should contain 'ABC'")
	}

	if !strings.Contains(str, "DEF") {
		t.Error("String should contain 'DEF'")
	}
}

// TestJSONOutputNil tests JSONOutput with nil inputs.
func TestJSONOutputNil(t *testing.T) {
	var output *JSONOutput

	// String should handle nil
	str := output.String()
	if str != "" {
		t.Errorf("Expected empty string for nil JSONOutput, got '%s'", str)
	}

	// Compare should handle nil
	diffs := output.Compare(nil)
	if len(diffs) != 0 {
		t.Errorf("Expected 0 diffs for nil vs nil, got %d", len(diffs))
	}

	output2 := &JSONOutput{Width: 10, Height: 5}
	diffs = output.Compare(output2)
	if len(diffs) == 0 {
		t.Error("Expected differences for nil vs non-nil")
	}
}
