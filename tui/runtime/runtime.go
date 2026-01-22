package runtime

// Event is a placeholder for future event system
// v1: simplified, will be expanded in Phase 3
type Event struct {
	X, Y int
	Type string
	Data interface{}
}

// Runtime is the main interface for the Yao TUI Runtime.
//
// It provides a clean API for:
//   - Layout: Calculate geometry (measure + layout phases)
//   - Render: Generate frames from layout results
//   - Dispatch: Handle events (Phase 3)
//   - Focus: Manage keyboard navigation (Phase 3)
type Runtime interface {
	// Layout performs a complete layout pass on the root node.
	//
	// This includes:
	//   1. Measure phase: Calculate intrinsic sizes bottom-up
	//   2. Layout phase: Assign positions top-down
	//
	// The constraints (c) are the available space from the screen/window.
	//
	// Returns a LayoutResult containing all positioned nodes.
	Layout(root *LayoutNode, c BoxConstraints) LayoutResult

	// Render generates a Frame from a LayoutResult.
	//
	// This is the Render phase, which:
	//   - Creates a CellBuffer (virtual canvas)
	//   - Renders all nodes in Z-Index order
	//   - Returns a Frame that can be output to the terminal
	//
	// The resulting Frame.String() can be used to update the terminal.
	Render(result LayoutResult) Frame

	// Dispatch handles an input event (keyboard, mouse, etc.).
	//
	// v1: placeholder, will be implemented in Phase 3
	// For now, events are handled by existing Bubble Tea system.
	Dispatch(ev Event)

	// FocusNext moves focus to the next focusable component.
	//
	// v1: placeholder, will be implemented in Phase 3
	// For now, focus is handled by existing focus manager.
	FocusNext()
}

// Frame represents a rendered frame (virtual canvas).
//
// It contains the complete rendered output that can be sent to the terminal.
type Frame struct {
	Buffer *CellBuffer
	Width  int
	Height int
	Dirty  bool // True if this frame has changes from previous
}

// String returns the frame as a string for terminal output.
// This is the primary way to send a frame to Bubble Tea's View() method.
func (f Frame) String() string {
	if f.Buffer == nil {
		return ""
	}
	return f.Buffer.String()
}

// CellBuffer is a virtual canvas for rendering.
//
// It represents the terminal screen as a 2D array of cells.
// Each cell contains a character and its style.
//
// Z-Index support is built-in: later writes to a position will
// overwrite earlier writes, but cells compare Z-Index to decide.
type CellBuffer struct {
	cells  [][]Cell
	width  int
	height int
}

// Cell represents a single cell in the CellBuffer.
type Cell struct {
	Char   rune
	Style  CellStyle
	ZIndex int
	NodeID string // For hit testing
}

// CellStyle represents rendering style for a cell.
// v1: simplified version, will be expanded to support lipgloss.Style in render module
type CellStyle struct {
	Bold      bool
	Underline bool
	// v1: foreground/background will be handled by text rendering
	// Full lipgloss.Style support will be in render module
}

// NewCellBuffer creates a new CellBuffer with the given dimensions.
func NewCellBuffer(width, height int) *CellBuffer {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = Cell{
				Char:   ' ',
				Style:  CellStyle{},
				ZIndex: 0,
			}
		}
	}

	return &CellBuffer{
		cells:  cells,
		width:  width,
		height: height,
	}
}

// SetContent sets a cell at the given position.
// If the new Z-Index is greater than or equal to existing, overwrites the cell.
func (b *CellBuffer) SetContent(x, y, z int, char rune, style CellStyle, nodeID string) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}

	// Check Z-Index
	if z < b.cells[y][x].ZIndex {
		return
	}

	b.cells[y][x] = Cell{
		Char:   char,
		Style:  style,
		ZIndex: z,
		NodeID: nodeID,
	}
}

// GetContent returns the cell at the given position.
func (b *CellBuffer) GetContent(x, y int) Cell {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return Cell{}
	}
	return b.cells[y][x]
}

// Clear clears the entire buffer
func (b *CellBuffer) Clear() {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.cells[y][x] = Cell{
				Char:   ' ',
				Style:  CellStyle{},
				ZIndex: 0,
			}
		}
	}
}

// Width returns the buffer width.
func (b *CellBuffer) Width() int {
	return b.width
}

// Height returns the buffer height.
func (b *CellBuffer) Height() int {
	return b.height
}

// String returns the buffer as a string.
// This is a simple implementation that will be enhanced in the render module
// to properly handle styles and lipgloss integration.
func (b *CellBuffer) String() string {
	lines := make([]string, b.height)
	for y := 0; y < b.height; y++ {
		runes := make([]rune, b.width)
		for x := 0; x < b.width; x++ {
			runes[x] = b.cells[y][x].Char
		}
		lines[y] = string(runes)
	}
	return joinLines(lines)
}

// joinLines joins lines with newline characters
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	if len(lines) == 1 {
		return lines[0]
	}
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
}
