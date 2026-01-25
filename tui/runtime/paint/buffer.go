package paint

import "github.com/yaoapp/yao/tui/framework/style"

// Buffer represents a grid of cells that components paint into.
// It acts as the "canvas" for the TUI rendering engine.
type Buffer struct {
	// Width and Height of the buffer
	Width  int
	Height int

	// Cells stores the grid content.
	// Access via GetCell/SetCell.
	Cells [][]Cell
}

// NewBuffer creates a new buffer with the specified dimensions.
func NewBuffer(width, height int) *Buffer {
	b := &Buffer{
		Width:  width,
		Height: height,
		Cells:  make([][]Cell, height),
	}

	for y := 0; y < height; y++ {
		b.Cells[y] = make([]Cell, width)
		// Initialize with empty cells if needed, or rely on zero value
	}

	return b
}

// SetCell sets the character and style at the given coordinates.
// It handles boundary checks safely.
func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	b.Cells[y][x] = Cell{
		Char:  char,
		Style: s,
		Width: 1, // Simplified width assumption for now
	}
}

// SetString writes a string starting at (x, y) with the given style.
func (b *Buffer) SetString(x, y int, text string, s style.Style) {
	if y < 0 || y >= b.Height {
		return
	}

	col := x
	for _, char := range text {
		if col >= b.Width {
			break
		}
		b.SetCell(col, y, char, s)
		col++
	}
}

// Fill fills a rectangular area with a character and style.
func (b *Buffer) Fill(rect Rect, char rune, s style.Style) {
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			b.SetCell(x, y, char, s)
		}
	}
}

// Rect represents a rectangular area.
// We duplicate this simple struct here or share it.
// For now, let's define it here to make paint package self-contained.
type Rect struct {
	X, Y, Width, Height int
}
