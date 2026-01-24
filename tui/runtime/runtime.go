package runtime

import (
	"fmt"
	"strings"
)

// Component is the base interface for all UI components.
//
// This is the minimal interface that all components must implement.
// The View() method returns the component's visual representation.
type Component interface {
	// View returns the component's visual representation as a string.
	// This is called during the Render phase.
	View() string
}

// Event is a placeholder for future event system
// v1: simplified, will be expanded in Phase 3
type Event struct {
	X, Y int
	Type string
	Data interface{}
}

// FocusableComponent is an interface for components that can receive focus.
// This is the minimal interface required for focus management.
type FocusableComponent interface {
	SetFocus(focus bool)
	IsFocusable() bool
}

// Rect represents a rectangle region.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
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
	// StyledText stores the original ANSI-styled text from lipgloss
	// When non-empty, this takes precedence over Char+Style for rendering
	StyledText string
	// Selected indicates this cell is part of a selection
	Selected bool
}

// CellStyle represents rendering style for a cell.
// v1: supports basic text styling + colors for lipgloss integration
type CellStyle struct {
	Bold       bool
	Underline  bool
	Italic     bool
	Strikethrough bool
	Blink      bool
	Reverse    bool
	Foreground string // Hex color or terminal color name
	Background string // Hex color or terminal color name
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

// SetContentRuntime sets a cell at the given position using individual style parameters.
// This method is used by the render package adapter to avoid circular imports.
func (b *CellBuffer) SetContentRuntime(x, y, z int, char rune, bold, underline, italic bool, nodeID string) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}

	// Check Z-Index
	if z < b.cells[y][x].ZIndex {
		return
	}

	b.cells[y][x] = Cell{
		Char:   char,
		Style:  CellStyle{Bold: bold, Underline: underline, Italic: italic},
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

// GetCell returns the cell at the given position (alias for GetContent).
// This is the preferred name for consistency with SetCell.
func (b *CellBuffer) GetCell(x, y int) Cell {
	return b.GetContent(x, y)
}

// SetCell sets a cell at the given position without Z-Index checking.
// This directly overwrites the cell regardless of Z-Index.
// Use SetContent for Z-Index-aware cell setting.
func (b *CellBuffer) SetCell(x, y int, char rune, style CellStyle, zIndex int) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}

	b.cells[y][x] = Cell{
		Char:   char,
		Style:  style,
		ZIndex: zIndex,
	}
}

// SetStyledText sets styled text at a given line.
// The text should contain lipgloss ANSI escape codes for styling.
// This stores the styled text in the first cell of the line and marks
// subsequent cells as part of the styled text region.
func (b *CellBuffer) SetStyledText(x, y, z int, text string, nodeID string) {
	if y < 0 || y >= b.height || x < 0 || x >= b.width {
		return
	}

	// Store the styled text in the first cell
	// Mark the cell as containing styled text
	if z >= b.cells[y][x].ZIndex {
		b.cells[y][x] = Cell{
			Char:       0, // 0 indicates styled text follows
			StyledText: text,
			ZIndex:     z,
			NodeID:     nodeID,
		}

		// Mark the span of this styled text
		// Count visible characters (excluding ANSI codes)
		runes := []rune(text)
		visibleLen := 0
		i := 0
		for i < len(runes) {
			if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
				// Skip ANSI escape sequence
				i += 2
				for i < len(runes) && runes[i] != 'm' {
					i++
				}
				if i < len(runes) {
					i++ // skip 'm'
				}
			} else {
				visibleLen++
				i++
			}
		}

		// Mark subsequent cells as part of this styled region
		for offsetX := 1; offsetX < visibleLen && x+offsetX < b.width; offsetX++ {
			if z >= b.cells[y][x+offsetX].ZIndex {
				b.cells[y][x+offsetX] = Cell{
					Char:   ' ', // Placeholder, styled text will be output
					ZIndex: z,
					NodeID: nodeID,
					// StyledText empty - this is a continuation cell
				}
			}
		}
	}
}

// SetSelected sets the Selected flag for a cell at the given position.
func (b *CellBuffer) SetSelected(x, y int, selected bool) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	b.cells[y][x].Selected = selected
}

// ClearSelection clears the selection flag for all cells.
func (b *CellBuffer) ClearSelection() {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.cells[y][x].Selected = false
		}
	}
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

// String returns the buffer as a string with ANSI escape codes for styling.
// This outputs the buffer with proper terminal styling support.
// Handles both CellStyle and lipgloss StyledText (with embedded ANSI codes).
// Also applies selection highlighting (reverse video) to selected cells.
func (b *CellBuffer) String() string {
	if b.height == 0 {
		return ""
	}

	lines := make([]string, b.height)
	for y := 0; y < b.height; y++ {
		var lineBuilder strings.Builder
		x := 0

		for x < b.width {
			cell := b.cells[y][x]

			// Check if this cell has styled text (from lipgloss)
			if cell.StyledText != "" {
				// Check if any part of this styled text region is selected
				visibleLen := countVisibleChars(cell.StyledText)
				hasSelection := false
				for offsetX := 0; offsetX < visibleLen && x+offsetX < b.width; offsetX++ {
					if b.cells[y][x+offsetX].Selected {
						hasSelection = true
						break
					}
				}

				if hasSelection {
					// For selected styled text, process to ensure reverse works correctly
					// The issue is that lipgloss \x1b[0m resets our reverse video
					processed := applySelectionToStyledText(cell.StyledText)
					lineBuilder.WriteString(processed)
				} else {
					// Output the styled text directly (it already contains ANSI codes)
					lineBuilder.WriteString(cell.StyledText)
				}

				// Skip continuation cells
				x += visibleLen
			} else {
				// Regular cell: output character with style
				// Apply reverse video if selected
				if cell.Selected {
					lineBuilder.WriteString("\x1b[7m")
				}

				if cell.Char == 0 {
					lineBuilder.WriteRune(' ')
				} else {
					lineBuilder.WriteRune(cell.Char)
				}

				if cell.Selected {
					lineBuilder.WriteString("\x1b[27m")
				}

				x++
			}
		}

		// Reset style at end of line
		lineBuilder.WriteString("\x1b[0m")
		lines[y] = lineBuilder.String()
	}

	return joinLines(lines)
}

// countVisibleChars counts the number of visible characters in a string,
// excluding ANSI escape sequences.
func countVisibleChars(s string) int {
	runes := []rune(s)
	count := 0
	i := 0
	for i < len(runes) {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// Skip ANSI escape sequence
			i += 2
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
			if i < len(runes) {
				i++ // skip 'm'
			}
		} else {
			count++
			i++
		}
	}
	return count
}

// applySelectionToStyledText processes styled text to ensure selection highlighting works correctly.
// The problem is that lipgloss uses \x1b[0m which resets all attributes including our reverse video.
// This function processes the styled text to ensure reverse video is maintained.
func applySelectionToStyledText(styledText string) string {
	runes := []rune(styledText)
	var result strings.Builder
	i := 0

	for i < len(runes) {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// Copy the ANSI escape sequence
			start := i
			i += 2
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
			if i < len(runes) {
				i++ // skip 'm'
			}

			// Check if this is a reset sequence
			seq := string(runes[start:i])
			if seq == "\x1b[0m" {
				// Replace with reset+reverse to maintain highlighting
				result.WriteString("\x1b[0m\x1b[7m") // Reset then reverse again
			} else {
				result.WriteString(seq)
			}
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}

	// Ensure we end with reverse still on (in case there was no reset at the end)
	// and add final reset at the end
	return result.String() + "\x1b[27m\x1b[0m"
}

// styleToANSI converts a CellStyle to ANSI escape codes
// Supports colors, bold, underline, italic, strikethrough, blink, reverse
func styleToANSI(style CellStyle) string {
	// Check if no styling
	if !style.Bold && !style.Underline && !style.Italic &&
	   !style.Strikethrough && !style.Blink && !style.Reverse &&
	   style.Foreground == "" && style.Background == "" {
		return "\x1b[0m" // Reset
	}

	codes := []string{}

	// Add text styling codes
	if style.Bold {
		codes = append(codes, "1")
	}
	if style.Italic {
		codes = append(codes, "3")
	}
	if style.Underline {
		codes = append(codes, "4")
	}
	if style.Strikethrough {
		codes = append(codes, "9")
	}
	if style.Blink {
		codes = append(codes, "5")
	}
	if style.Reverse {
		codes = append(codes, "7")
	}

	// Add foreground color
	if style.Foreground != "" {
		fgCode := colorToANSICode(style.Foreground, false)
		if fgCode != "" {
			// Foreground colors use codes 30-37 (basic), 38 (256 color), 38;5 (RGB)
			// We prefix with 3 for foreground, 4 for background
			codes = append(codes, fgCode)
		}
	}

	// Add background color
	if style.Background != "" {
		bgCode := colorToANSICode(style.Background, true)
		if bgCode != "" {
			codes = append(codes, bgCode)
		}
	}

	if len(codes) == 0 {
		return "\x1b[0m"
	}

	result := "\x1b["
	for i, code := range codes {
		if i > 0 {
			result += ";"
		}
		result += code
	}
	result += "m"
	return result
}

// colorToANSICode converts a color string to ANSI color code
// isBackground: true for background color (40-47), false for foreground (30-37)
// Returns the ANSI code string or empty if no valid color
//
// Supported formats:
// - Named colors: "red", "blue", "green", etc.
// - 256-color mode: "0-255" (e.g., "214" for orange, "33" for blue)
// - Hex colors: "#RRGGBB" (not yet supported, returns empty)
func colorToANSICode(color string, isBackground bool) string {
	if color == "" {
		return ""
	}

	// Hex color: "#RRGGBB"
	if strings.HasPrefix(color, "#") {
		// For now, we don't support truecolor in basic ANSI
		// Would need to use 38;2;R;G;B format for truecolor
		// Just return empty for now
		return ""
	}

	// Try to parse as 256-color code (0-255)
	// This handles numeric strings like "214", "33", "240", etc.
	if len(color) > 0 && color[0] >= '0' && color[0] <= '9' {
		codeNum := 0
		if _, err := fmt.Sscanf(color, "%d", &codeNum); err == nil && codeNum >= 0 && codeNum <= 255 {
			// 256-color mode uses format: 38;5;N (foreground) or 48;5;N (background)
			if isBackground {
				return fmt.Sprintf("48;5;%d", codeNum)
			}
			return fmt.Sprintf("38;5;%d", codeNum)
		}
	}

	// Basic ANSI color names
	baseCode := "30" // Default to foreground
	if isBackground {
		baseCode = "40"
	}

	switch color {
	case "black":
		return baseCode + "0"
	case "red":
		return baseCode + "1"
	case "green":
		return baseCode + "2"
	case "yellow":
		return baseCode + "3"
	case "blue":
		return baseCode + "4"
	case "magenta":
		return baseCode + "5"
	case "cyan":
		return baseCode + "6"
	case "white":
		return baseCode + "7"
	// Bright variants (use 90-97 for foreground, 100-107 for background)
	case "bright-black", "gray":
		if isBackground {
			return "100"
		}
		return "90"
	case "bright-red":
		if isBackground {
			return "101"
		}
		return "91"
	case "bright-green":
		if isBackground {
			return "102"
		}
		return "92"
	case "bright-yellow":
		if isBackground {
			return "103"
		}
		return "93"
	case "bright-blue":
		if isBackground {
			return "104"
		}
		return "94"
	case "bright-magenta":
		if isBackground {
			return "105"
		}
		return "95"
	case "bright-cyan":
		if isBackground {
			return "106"
		}
		return "96"
	case "bright-white":
		if isBackground {
			return "107"
		}
		return "97"
	default:
		// Unknown color, return empty
		return ""
	}
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
