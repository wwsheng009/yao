package selection

import (
	"testing"
)

// mockBuffer implements TextBuffer for testing.
type mockBuffer struct {
	cells  [][]Cell
	width  int
	height int
}

func newMockBuffer(width, height int) *mockBuffer {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = Cell{
				Char:  rune('a' + x%26),
				Empty: false,
			}
		}
	}
	return &mockBuffer{
		cells:  cells,
		width:  width,
		height: height,
	}
}

func (m *mockBuffer) GetCell(x, y int) Cell {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return Cell{Empty: true}
	}
	return m.cells[y][x]
}

func (m *mockBuffer) Width() int {
	return m.width
}

func (m *mockBuffer) Height() int {
	return m.height
}

func TestNewManager(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.IsActive() {
		t.Error("New manager should not be active")
	}
}

func TestManager_Start(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 3)

	if !manager.IsActive() {
		t.Error("Manager should be active after Start")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 2 || endX != 2 || startY != 3 || endY != 3 {
		t.Errorf("Expected range (2,2,3,3), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

func TestManager_Update(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 2)
	manager.Update(5, 4)

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 2 || endX != 5 || startY != 2 || endY != 4 {
		t.Errorf("Expected range (2,5,2,4), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

func TestManager_Update_Reversed(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(5, 4)
	manager.Update(2, 2)

	// Normalized range should have start < end
	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 2 || endX != 5 || startY != 2 || endY != 4 {
		t.Errorf("Expected normalized range (2,5,2,4), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

func TestManager_IsSelected(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 2)
	manager.Update(5, 4)

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"inside selection", 3, 3, true},
		{"at start", 2, 2, true},
		{"at end", 5, 4, true},
		{"outside left", 1, 3, false},
		{"outside right", 6, 3, false},
		{"outside top", 3, 1, false},
		{"outside bottom", 3, 5, false},
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

func TestManager_Clear(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 2)
	manager.Update(5, 4)

	if !manager.IsActive() {
		t.Error("Manager should be active")
	}

	manager.Clear()

	if manager.IsActive() {
		t.Error("Manager should not be active after Clear")
	}

	if manager.IsSelected(3, 3) {
		t.Error("No cells should be selected after Clear")
	}
}

func TestManager_SelectWord(t *testing.T) {
	buffer := &mockBuffer{
		width:  20,
		height: 3,
		cells:  make([][]Cell, 3),
	}

	// Create a buffer with words "hello world test"
	for y := 0; y < 3; y++ {
		buffer.cells[y] = make([]Cell, 20)
		word := "hello world test"
		x := 0
		for _, ch := range word {
			if x < 20 {
				buffer.cells[y][x] = Cell{Char: ch, Empty: false}
				x++
			}
		}
		for ; x < 20; x++ {
			buffer.cells[y][x] = Cell{Char: ' ', Empty: false}
		}
	}

	manager := NewManager(buffer)

	// Select "world" starting at position 7
	manager.SelectWord(7, 0)

	if !manager.IsActive() {
		t.Fatal("Manager should be active after SelectWord")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	// "world" is at positions 6-10
	if startX != 6 || endX != 10 || startY != 0 || endY != 0 {
		t.Errorf("Expected word range (6,10,0,0), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}

	text := manager.GetSelectedText()
	if text != "world" {
		t.Errorf("Expected selected text 'world', got '%s'", text)
	}
}

func TestManager_SelectLine(t *testing.T) {
	buffer := newMockBuffer(20, 5)
	manager := NewManager(buffer)

	manager.SelectLine(2)

	if !manager.IsActive() {
		t.Fatal("Manager should be active after SelectLine")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	// Should select entire line 2 (0 to width-1)
	if startX != 0 || endX != 19 || startY != 2 || endY != 2 {
		t.Errorf("Expected line range (0,19,2,2), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

func TestManager_SelectAll(t *testing.T) {
	buffer := newMockBuffer(20, 5)
	manager := NewManager(buffer)

	manager.SelectAll()

	if !manager.IsActive() {
		t.Fatal("Manager should be active after SelectAll")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 0 || endX != 19 || startY != 0 || endY != 4 {
		t.Errorf("Expected full range (0,19,0,4), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

func TestManager_GetSelectedCells(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 2)
	manager.Update(4, 3)

	cells := manager.GetSelectedCells()

	// Should have (4-2+1) * (3-2+1) = 3 * 2 = 6 cells
	expectedCount := 3 * 2
	if len(cells) != expectedCount {
		t.Errorf("Expected %d cells, got %d", expectedCount, len(cells))
	}

	// Check that all returned cells are within the selection
	for _, cell := range cells {
		if cell.X < 2 || cell.X > 4 || cell.Y < 2 || cell.Y > 3 {
			t.Errorf("Cell (%d,%d) is outside selection range", cell.X, cell.Y)
		}
	}
}

func TestManager_MoveStart(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 2)
	manager.Update(5, 4)
	manager.MoveStart(1, 0)

	startX, _, startY, _ := manager.GetSelectionRange()
	if startX != 3 || startY != 2 {
		t.Errorf("Expected start (3,2), got (%d,%d)", startX, startY)
	}
}

func TestManager_MoveEnd(t *testing.T) {
	buffer := newMockBuffer(10, 5)
	manager := NewManager(buffer)

	manager.Start(2, 2)
	manager.Update(5, 4)
	manager.MoveEnd(1, 0)

	_, endX, _, endY := manager.GetSelectionRange()
	if endX != 6 || endY != 4 {
		t.Errorf("Expected end (6,4), got (%d,%d)", endX, endY)
	}
}

func TestSelectionRegion(t *testing.T) {
	region := SelectionRegion{
		StartX: 2,
		EndX:   5,
		StartY: 3,
		EndY:   7,
	}

	tests := []struct {
		name     string
		x, y     int
		expected bool
	}{
		{"inside", 3, 5, true},
		{"at top-left", 2, 3, true},
		{"at bottom-right", 5, 7, true},
		{"outside left", 1, 5, false},
		{"outside right", 6, 5, false},
		{"outside top", 3, 2, false},
		{"outside bottom", 3, 8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := region.Contains(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("Region.Contains(%d,%d) = %v, want %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestSelectionRegion_Width(t *testing.T) {
	region := SelectionRegion{
		StartX: 2,
		EndX:   5,
		StartY: 3,
		EndY:   7,
	}

	if region.Width() != 4 {
		t.Errorf("Expected width 4, got %d", region.Width())
	}

	if region.Height() != 5 {
		t.Errorf("Expected height 5, got %d", region.Height())
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		r        rune
		expected int
	}{
		{'a', 1},
		{'A', 1},
		{' ', 1},
		{'中', 2},  // Chinese character
		{'日', 2},  // Japanese character
		{'한', 2},  // Korean character
	}

	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			result := RuneWidth(tt.r)
			if result != tt.expected {
				t.Errorf("RuneWidth(%c) = %d, want %d", tt.r, result, tt.expected)
			}
		})
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		s        string
		expected int
	}{
		{"hello", 5},
		{"hello world", 11},
		{"你好", 4},  // 2 Chinese chars = 4 width
		{"hello世界", 9}, // 5 + 4 = 9
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			result := StringWidth(tt.s)
			if result != tt.expected {
				t.Errorf("StringWidth(%s) = %d, want %d", tt.s, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		s        string
		maxWidth int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello wo"},
		{"你好世界", 4, "你好"},
		{"abc", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			result := TruncateString(tt.s, tt.maxWidth)
			if StringWidth(result) > tt.maxWidth {
				t.Errorf("TruncateString(%s, %d) = %s has width %d, want <= %d",
					tt.s, tt.maxWidth, result, StringWidth(result), tt.maxWidth)
			}
		})
	}
}
