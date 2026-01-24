package runtime

import (
	"testing"
)

// TestSelectionManager_NewManager tests creating a new selection manager.
func TestSelectionManager_NewManager(t *testing.T) {
	manager := NewSelectionManager()

	if manager == nil {
		t.Fatal("NewSelectionManager returned nil")
	}

	if !manager.IsEnabled() {
		t.Error("New manager should be enabled by default")
	}

	if manager.IsActive() {
		t.Error("New manager should not be active")
	}
}

// TestSelectionManager_Start tests starting a selection.
func TestSelectionManager_Start(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(5, 3)

	if !manager.IsActive() {
		t.Error("Manager should be active after Start")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 5 || endX != 5 || startY != 3 || endY != 3 {
		t.Errorf("Expected range (5,5,3,3), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_Update tests updating selection.
func TestSelectionManager_Update(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(2, 2)
	manager.Update(8, 5)

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 2 || endX != 8 || startY != 2 || endY != 5 {
		t.Errorf("Expected range (2,8,2,5), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_IsSelected tests checking if a cell is selected.
func TestSelectionManager_IsSelected(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(2, 2)
	manager.Update(6, 5)

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"inside selection", 4, 3, true},
		{"at start", 2, 2, true},
		{"at end", 6, 5, true},
		{"outside left", 1, 3, false},
		{"outside right", 7, 3, false},
		{"outside top", 3, 1, false},
		{"outside bottom", 3, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.IsSelected(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("IsSelected(%d,%d) = %v, want %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

// TestSelectionManager_Clear tests clearing selection.
func TestSelectionManager_Clear(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(2, 2)
	manager.Update(6, 5)

	if !manager.IsActive() {
		t.Error("Manager should be active")
	}

	manager.Clear()

	if manager.IsActive() {
		t.Error("Manager should not be active after Clear")
	}
}

// TestSelectionManager_SelectAll tests selecting all text.
func TestSelectionManager_SelectAll(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.SelectAll()

	if !manager.IsActive() {
		t.Error("Manager should be active after SelectAll")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 0 || endX != 19 || startY != 0 || endY != 9 {
		t.Errorf("Expected full range (0,19,0,9), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_SelectWord tests word selection.
func TestSelectionManager_SelectWord(t *testing.T) {
	buffer := NewCellBuffer(30, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add some text: "hello world test"
	text := "hello world test"
	for x, ch := range text {
		if x < buffer.width {
			buffer.SetContent(x, 2, 0, ch, CellStyle{}, "")
		}
	}

	// Select "world" (starts at position 6)
	manager.SelectWord(7, 2)

	if !manager.IsActive() {
		t.Fatal("Manager should be active")
	}

	// Check selection range (should be "world" = positions 6-10)
	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 6 || endX != 10 || startY != 2 || endY != 2 {
		t.Errorf("Expected word range (6,10,2,2), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}

	selectedText := manager.GetSelectedText()
	if selectedText != "world" {
		t.Errorf("Expected selected text 'world', got '%s'", selectedText)
	}
}

// TestSelectionManager_SelectLine tests line selection.
func TestSelectionManager_SelectLine(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.SelectLine(2)

	if !manager.IsActive() {
		t.Fatal("Manager should be active")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 0 || endX != 19 || startY != 2 || endY != 2 {
		t.Errorf("Expected line range (0,19,2,2), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_EnableDisable tests enabling/disabling.
func TestSelectionManager_EnableDisable(t *testing.T) {
	manager := NewSelectionManager()

	if !manager.IsEnabled() {
		t.Error("Manager should be enabled by default")
	}

	manager.SetEnabled(false)
	if manager.IsEnabled() {
		t.Error("Manager should be disabled")
	}

	// Try to start selection - should fail when disabled
	buffer := NewCellBuffer(20, 10)
	manager.SetBuffer(buffer)
	manager.Start(5, 3)
	if manager.IsActive() {
		t.Error("Selection should not start when disabled")
	}

	// Re-enable and try again
	manager.SetEnabled(true)
	if !manager.IsEnabled() {
		t.Error("Manager should be enabled")
	}

	manager.Start(5, 3)
	if !manager.IsActive() {
		t.Error("Selection should start when enabled")
	}
}

// TestSelectionManager_GetSelectedText tests getting selected text.
func TestSelectionManager_GetSelectedText(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add text to buffer
	text := "test line"
	for x, ch := range text {
		buffer.SetContent(x, 1, 0, ch, CellStyle{}, "")
	}

	manager.Start(0, 1)
	manager.Update(len(text)-1, 1)

	selectedText := manager.GetSelectedText()
	if selectedText != "test line" {
		t.Errorf("Expected 'test line', got '%s'", selectedText)
	}
}

// TestSelectionManager_MultiLineSelection tests multi-line selection.
func TestSelectionManager_MultiLineSelection(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add multi-line text
	lines := []string{
		"line 1",
		"line 2",
		"line 3",
	}
	for y, line := range lines {
		for x, ch := range line {
			buffer.SetContent(x, y, 0, ch, CellStyle{}, "")
		}
	}

	// Select from (0, 0) to (5, 2) - should span 3 lines
	manager.Start(0, 0)
	manager.Update(5, 2)

	selectedText := manager.GetSelectedText()
	expected := "line 1\nline 2\nline 3"
	if selectedText != expected {
		t.Errorf("Expected multi-line text, got: '%s'", selectedText)
	}
}

// TestSelectionManager_ApplyHighlight tests applying highlight to buffer.
func TestSelectionManager_ApplyHighlight(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add some text
	for x := 0; x < 10; x++ {
		buffer.SetContent(x, 2, 0, 'A', CellStyle{}, "")
	}

	// Select a range
	manager.Start(2, 2)
	manager.Update(7, 2)

	// Get original cell
	beforeCell := buffer.GetContent(5, 2)
	if beforeCell.Char != 'A' {
		t.Errorf("Expected 'A', got '%c'", beforeCell.Char)
	}

	// Apply highlight
	manager.ApplyHighlight()

	// Check that selected cells have Selected flag set
	for x := 2; x <= 7; x++ {
		cell := buffer.GetContent(x, 2)
		if !cell.Selected {
			t.Errorf("Cell at (%d, 2) should have Selected flag set", x)
		}
	}

	// Non-selected cells should not be selected
	unselectedCell := buffer.GetContent(0, 2)
	if unselectedCell.Selected {
		t.Error("Unselected cell should not have Selected flag set")
	}
}

// TestSelectionManager_GetSelectedCells tests getting all selected cells.
func TestSelectionManager_GetSelectedCells(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(2, 1)
	manager.Update(5, 3)

	cells := manager.GetSelectedCells()

	// Should have 4 columns x 3 rows = 12 cells
	expectedCount := 4 * 3
	if len(cells) != expectedCount {
		t.Errorf("Expected %d cells, got %d", expectedCount, len(cells))
	}

	// Verify all cells are within selection
	for _, cell := range cells {
		if cell.X < 2 || cell.X > 5 || cell.Y < 1 || cell.Y > 3 {
			t.Errorf("Cell (%d,%d) is outside selection range", cell.X, cell.Y)
		}
	}
}

// TestSelectionManager_GetRegion tests getting selection region.
func TestSelectionManager_GetRegion(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(2, 1)
	manager.Update(8, 4)

	region := manager.GetRegion()

	if region.StartX != 2 || region.EndX != 8 {
		t.Errorf("X range wrong: got (%d, %d)", region.StartX, region.EndX)
	}
	if region.StartY != 1 || region.EndY != 4 {
		t.Errorf("Y range wrong: got (%d, %d)", region.StartY, region.EndY)
	}
	if region.Width() != 7 {
		t.Errorf("Width wrong: got %d", region.Width())
	}
	if region.Height() != 4 {
		t.Errorf("Height wrong: got %d", region.Height())
	}
}

// TestDefaultSelectionHighlight tests the default highlight style.
func TestDefaultSelectionHighlight(t *testing.T) {
	style := DefaultSelectionHighlight()

	if !style.Reverse {
		t.Error("Default highlight should use reverse video")
	}
}

// TestLightSelectionHighlight tests the light theme highlight.
func TestLightSelectionHighlight(t *testing.T) {
	style := LightSelectionHighlight()

	if style.Background == "" {
		t.Error("Light highlight should have background color")
	}
}

// TestDarkSelectionHighlight tests the dark theme highlight.
func TestDarkSelectionHighlight(t *testing.T) {
	style := DarkSelectionHighlight()

	if style.Background == "" {
		t.Error("Dark highlight should have background color")
	}
}
