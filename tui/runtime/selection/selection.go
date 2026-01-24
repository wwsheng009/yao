package selection

import (
	"strings"
	"unicode/utf8"
)

// Manager manages text selection state and operations.
// It handles mouse-based selection and copy operations.
type Manager struct {
	// Selection state
	active       bool
	startX       int
	startY       int
	currentX     int
	currentY     int
	anchorX      int // Selection anchor point (for extending selections)
	anchorY      int

	// Selection mode
	mode         SelectionMode

	// Buffer reference for extracting text
	buffer       TextBuffer
}

// SelectionMode defines how text selection works.
type SelectionMode int

const (
	// SelectionModeChar selects character by character
	SelectionModeChar SelectionMode = iota
	// SelectionModeWord selects whole words
	SelectionModeWord
	// SelectionModeLine selects whole lines
	SelectionModeLine
)

// TextBuffer defines the interface for accessing text content.
type TextBuffer interface {
	GetCell(x, y int) Cell
	Width() int
	Height() int
}

// Cell represents a single cell in the text buffer.
type Cell struct {
	Char  rune
	Empty bool
}

// NewManager creates a new selection manager.
func NewManager(buffer TextBuffer) *Manager {
	return &Manager{
		buffer:  buffer,
		mode:    SelectionModeChar,
		active:  false,
	}
}

// SetBuffer sets the text buffer for selection operations.
func (m *Manager) SetBuffer(buffer TextBuffer) {
	m.buffer = buffer
}

// Start begins a new selection at the given position.
func (m *Manager) Start(x, y int) {
	m.active = true
	m.startX = x
	m.startY = y
	m.currentX = x
	m.currentY = y
	m.anchorX = x
	m.anchorY = y
}

// Update updates the current selection end position.
func (m *Manager) Update(x, y int) {
	if !m.active {
		return
	}

	// Clamp to buffer bounds
	width := m.buffer.Width()
	height := m.buffer.Height()

	if x < 0 {
		x = 0
	} else if x >= width {
		x = width - 1
	}

	if y < 0 {
		y = 0
	} else if y >= height {
		y = height - 1
	}

	m.currentX = x
	m.currentY = y
}

// Extend extends the selection from the anchor to the new position.
// This is used for Shift+Click selections.
func (m *Manager) Extend(x, y int) {
	if !m.active {
		m.Start(x, y)
		return
	}

	// Clamp to buffer bounds
	width := m.buffer.Width()
	height := m.buffer.Height()

	if x < 0 {
		x = 0
	} else if x >= width {
		x = width - 1
	}

	if y < 0 {
		y = 0
	} else if y >= height {
		y = height - 1
	}

	m.currentX = x
	m.currentY = y
}

// Clear clears the current selection.
func (m *Manager) Clear() {
	m.active = false
	m.startX = 0
	m.startY = 0
	m.currentX = 0
	m.currentY = 0
	m.anchorX = 0
	m.anchorY = 0
}

// IsActive returns whether a selection is active.
func (m *Manager) IsActive() bool {
	return m.active
}

// GetSelectedCells returns all cells in the current selection.
// Returns a slice of (x, y) coordinates.
func (m *Manager) GetSelectedCells() []struct{ X, Y int } {
	if !m.active {
		return nil
	}

	var cells []struct{ X, Y int }

	// Normalize selection coordinates
	startX, endX, startY, endY := m.normalize()

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			cells = append(cells, struct{ X, Y int }{X: x, Y: y})
		}
	}

	return cells
}

// IsSelected returns whether the cell at (x, y) is selected.
func (m *Manager) IsSelected(x, y int) bool {
	if !m.active {
		return false
	}

	startX, endX, startY, endY := m.normalize()

	return x >= startX && x <= endX && y >= startY && y <= endY
}

// GetSelectedText returns the selected text as a string.
// Lines are joined with newlines, preserving the visual layout.
func (m *Manager) GetSelectedText() string {
	if !m.active || m.buffer == nil {
		return ""
	}

	startX, endX, startY, endY := m.normalize()
	var lines []string

	for y := startY; y <= endY; y++ {
		var lineBuilder strings.Builder
		lineStart := startX
		lineEnd := endX

		// For multi-line selections, use full line width for middle lines
		if y > startY && y < endY {
			lineStart = 0
			lineEnd = m.buffer.Width() - 1
		}

		// Collect cells for this line
		for x := lineStart; x <= lineEnd; x++ {
			cell := m.buffer.GetCell(x, y)
			if !cell.Empty {
				lineBuilder.WriteRune(cell.Char)
			} else {
				lineBuilder.WriteRune(' ')
			}
		}

		// Trim trailing spaces for each line
		line := strings.TrimRight(lineBuilder.String(), " ")
		if line != "" || y < endY {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// GetSelectedTextCompact returns the selected text with trailing whitespace trimmed.
// This is useful for copy operations where you don't want trailing newlines.
func (m *Manager) GetSelectedTextCompact() string {
	text := m.GetSelectedText()
	return strings.TrimRight(text, "\n")
}

// GetSelectionRange returns the normalized selection range.
// Returns (startX, endX, startY, endY).
func (m *Manager) GetSelectionRange() (startX, endX, startY, endY int) {
	if !m.active {
		return 0, 0, 0, 0
	}
	return m.normalize()
}

// normalize returns the selection coordinates in normalized order.
// Ensures start <= end for both X and Y coordinates.
func (m *Manager) normalize() (startX, endX, startY, endY int) {
	startX = m.startX
	endX = m.currentX
	startY = m.startY
	endY = m.currentY

	// Swap if needed
	if startX > endX {
		startX, endX = endX, startX
	}
	if startY > endY {
		startY, endY = endY, startY
	}

	return startX, endX, startY, endY
}

// SelectWord selects the word at the given position.
// A word is defined as a sequence of non-whitespace characters.
func (m *Manager) SelectWord(x, y int) {
	if m.buffer == nil {
		return
	}

	width := m.buffer.Width()
	if x < 0 || x >= width {
		return
	}

	// Find word boundaries
	startX := x
	endX := x

	// Find start of word (going left)
	for startX > 0 {
		cell := m.buffer.GetCell(startX-1, y)
		if isWhitespace(cell.Char) {
			break
		}
		startX--
	}

	// Find end of word (going right)
	for endX < width-1 {
		cell := m.buffer.GetCell(endX+1, y)
		if isWhitespace(cell.Char) {
			break
		}
		endX++
	}

	m.active = true
	m.startX = startX
	m.startY = y
	m.currentX = endX
	m.currentY = y
	m.anchorX = startX
	m.anchorY = y
	m.mode = SelectionModeWord
}

// SelectLine selects the entire line at the given Y position.
func (m *Manager) SelectLine(y int) {
	if m.buffer == nil {
		return
	}

	height := m.buffer.Height()
	if y < 0 || y >= height {
		return
	}

	width := m.buffer.Width()

	m.active = true
	m.startX = 0
	m.startY = y
	m.currentX = width - 1
	m.currentY = y
	m.anchorX = 0
	m.anchorY = y
	m.mode = SelectionModeLine
}

// SelectAll selects the entire buffer.
func (m *Manager) SelectAll() {
	if m.buffer == nil {
		return
	}

	m.active = true
	m.startX = 0
	m.startY = 0
	m.currentX = m.buffer.Width() - 1
	m.currentY = m.buffer.Height() - 1
	m.anchorX = 0
	m.anchorY = 0
}

// SetMode sets the selection mode.
func (m *Manager) SetMode(mode SelectionMode) {
	m.mode = mode
}

// GetMode returns the current selection mode.
func (m *Manager) GetMode() SelectionMode {
	return m.mode
}

// MoveStart moves the selection start by the given delta.
func (m *Manager) MoveStart(dx, dy int) {
	if !m.active {
		return
	}

	m.startX += dx
	m.startY += dy

	// Clamp to buffer bounds
	if m.buffer != nil {
		if m.startX < 0 {
			m.startX = 0
		} else if m.startX >= m.buffer.Width() {
			m.startX = m.buffer.Width() - 1
		}

		if m.startY < 0 {
			m.startY = 0
		} else if m.startY >= m.buffer.Height() {
			m.startY = m.buffer.Height() - 1
		}
	}
}

// MoveEnd moves the selection end by the given delta.
func (m *Manager) MoveEnd(dx, dy int) {
	if !m.active {
		return
	}

	m.currentX += dx
	m.currentY += dy

	// Clamp to buffer bounds
	if m.buffer != nil {
		if m.currentX < 0 {
			m.currentX = 0
		} else if m.currentX >= m.buffer.Width() {
			m.currentX = m.buffer.Width() - 1
		}

		if m.currentY < 0 {
			m.currentY = 0
		} else if m.currentY >= m.buffer.Height() {
			m.currentY = m.buffer.Height() - 1
		}
	}
}

// SelectionRegion represents a rectangular selection region.
type SelectionRegion struct {
	StartX int
	EndX   int
	StartY int
	EndY   int
}

// GetRegion returns the selection as a SelectionRegion.
func (m *Manager) GetRegion() SelectionRegion {
	startX, endX, startY, endY := m.normalize()
	return SelectionRegion{
		StartX: startX,
		EndX:   endX,
		StartY: startY,
		EndY:   endY,
	}
}

// Contains checks if the region contains the given point.
func (r SelectionRegion) Contains(x, y int) bool {
	return x >= r.StartX && x <= r.EndX && y >= r.StartY && y <= r.EndY
}

// IsEmpty returns whether the selection region is empty.
func (r SelectionRegion) IsEmpty() bool {
	return r.StartX == 0 && r.EndX == 0 && r.StartY == 0 && r.EndY == 0
}

// Width returns the width of the selection region.
func (r SelectionRegion) Width() int {
	return r.EndX - r.StartX + 1
}

// Height returns the height of the selection region.
func (r SelectionRegion) Height() int {
	return r.EndY - r.StartY + 1
}

// isWhitespace returns whether a rune is whitespace.
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// RuneWidth returns the display width of a rune.
// Most runes are width 1, but some CJK characters are width 2.
func RuneWidth(r rune) int {
	// Simple implementation: most characters are width 1
	// CJK characters are width 2
	if r >= 0x1100 {
		// Unicode ranges for CJK characters
		if (r >= 0x1100 && r <= 0x115F) || // Hangul Jamo
			(r >= 0x2E80 && r <= 0x9FFF) || // CJK
			(r >= 0xAC00 && r <= 0xD7A3) || // Hangul Syllables
			(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
			(r >= 0xFE10 && r <= 0xFE19) ||
			(r >= 0xFE30 && r <= 0xFE6F) ||
			(r >= 0xFF00 && r <= 0xFF60) || // Fullwidth Forms
			(r >= 0xFFE0 && r <= 0xFFE6) ||
			(r >= 0x20000 && r <= 0x2FFFD) ||
			(r >= 0x30000 && r <= 0x3FFFD) {
			return 2
		}
	}
	return 1
}

// StringWidth returns the display width of a string.
func StringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}

// TruncateString truncates a string to fit within a given width.
func TruncateString(s string, maxWidth int) string {
	if StringWidth(s) <= maxWidth {
		return s
	}

	width := 0
	var runes []rune
	for _, r := range s {
		rw := RuneWidth(r)
		if width+rw > maxWidth {
			break
		}
		width += rw
		runes = append(runes, r)
	}

	return string(runes)
}

// CellCountToRuneIndex converts a cell count to a rune index in a string.
// This is useful for converting cursor positions (in cells) to string indices.
func CellCountToRuneIndex(s string, cellCount int) int {
	width := 0
	for i, r := range s {
		if width >= cellCount {
			return i
		}
		width += RuneWidth(r)
	}
	return utf8.RuneCountInString(s)
}

// RuneIndexToCellCount converts a rune index to a cell count in a string.
func RuneIndexToCellCount(s string, runeIndex int) int {
	runes := []rune(s)
	if runeIndex > len(runes) {
		runeIndex = len(runes)
	}

	width := 0
	for i := 0; i < runeIndex; i++ {
		width += RuneWidth(runes[i])
	}
	return width
}
