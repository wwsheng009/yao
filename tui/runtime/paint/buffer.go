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
		Width: runeWidth(char),
	}
}

// runeWidth 返回字符的显示宽度 (1 或 2)
func runeWidth(r rune) int {
	// CJK 字符范围 (中文、日文、韩文等)
	if r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x20000 && r <= 0x2fffd) ||
		(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	// Emoji 和其他符号
	if r >= 0x1f300 && r <= 0x1f9f0 {
		return 2
	}
	return 1
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
		width := runeWidth(char)
		// 对于宽字符，需要检查下一个位置是否可用
		if width == 2 && col+1 >= b.Width {
			break
		}
		b.SetCell(col, y, char, s)
		col += width
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

// Intersect calculates the intersection of two rectangles.
// Returns nil if there is no intersection.
func (r Rect) Intersect(other *Rect) *Rect {
	if other == nil {
		return &r
	}

	x1 := maxInt(r.X, other.X)
	y1 := maxInt(r.Y, other.Y)
	x2 := minInt(r.X+r.Width, other.X+other.Width)
	y2 := minInt(r.Y+r.Height, other.Y+other.Height)

	if x1 >= x2 || y1 >= y2 {
		return nil
	}

	return &Rect{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

// Contains checks if a point is inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
