package runtime

import (
	"strings"
)

// SelectionManager manages text selection state within the Runtime.
// It handles mouse-based selection and provides selected text for clipboard operations.
type SelectionManager struct {
	enabled   bool
	active    bool
	startX    int
	startY    int
	currentX  int
	currentY  int
	anchorX   int // Selection anchor point (for extending selections)
	anchorY   int
	mode      SelectionMode
	buffer    *CellBuffer
	highlight CellStyle
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

// NewSelectionManager creates a new selection manager.
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{
		enabled:   true,
		mode:      SelectionModeChar,
		highlight: DefaultSelectionHighlight(),
	}
}

// SetEnabled enables or disables text selection.
func (m *SelectionManager) SetEnabled(enabled bool) {
	m.enabled = enabled
	if !enabled {
		m.Clear()
	}
}

// IsEnabled returns whether text selection is enabled.
func (m *SelectionManager) IsEnabled() bool {
	return m.enabled
}

// SetBuffer sets the cell buffer for selection operations.
func (m *SelectionManager) SetBuffer(buffer *CellBuffer) {
	m.buffer = buffer
}

// GetBuffer returns the current cell buffer.
func (m *SelectionManager) GetBuffer() *CellBuffer {
	return m.buffer
}

// Start begins a new selection at the given position.
// The selection coordinates are tracked even without a buffer.
// The buffer will be set later during Render for text extraction and highlighting.
func (m *SelectionManager) Start(x, y int) {
	if !m.enabled {
		return
	}

	m.active = true
	// Clamp coordinates to reasonable defaults even without buffer
	// These will be re-clamped when buffer is available
	m.startX = x
	m.startY = y
	m.currentX = x
	m.currentY = y
	m.anchorX = x
	m.anchorY = y
	m.mode = SelectionModeChar
}

// Update updates the current selection end position.
// The selection coordinates are tracked even without a buffer.
func (m *SelectionManager) Update(x, y int) {
	if !m.active {
		return
	}

	m.currentX = x
	m.currentY = y
}

// Extend extends the selection from the anchor to the new position.
// This is used for Shift+Click selections.
func (m *SelectionManager) Extend(x, y int) {
	if m.buffer == nil {
		m.Start(x, y)
		return
	}

	if !m.active {
		m.Start(x, y)
		return
	}

	m.currentX = clampInt(x, 0, m.buffer.width-1)
	m.currentY = clampInt(y, 0, m.buffer.height-1)
}

// Clear clears the current selection.
func (m *SelectionManager) Clear() {
	m.active = false
	m.startX = 0
	m.startY = 0
	m.currentX = 0
	m.currentY = 0
	m.anchorX = 0
	m.anchorY = 0
	m.mode = SelectionModeChar
}

// IsActive returns whether a selection is active.
func (m *SelectionManager) IsActive() bool {
	return m.enabled && m.active
}

// IsSelected returns whether the cell at (x, y) is selected.
func (m *SelectionManager) IsSelected(x, y int) bool {
	if !m.IsActive() {
		return false
	}

	startX, endX, startY, endY := m.normalize()
	return x >= startX && x <= endX && y >= startY && y <= endY
}

// GetSelectedText returns the selected text as a string.
// For cells with StyledText, it extracts the visible characters (excluding ANSI codes).
func (m *SelectionManager) GetSelectedText() string {
	if !m.IsActive() || m.buffer == nil {
		return ""
	}

	startX, endX, startY, endY := m.normalize()
	var lines []string

	// Debug: log the selection range
	for y := startY; y <= endY; y++ {
		var lineBuilder strings.Builder
		lineStart := startX
		lineEnd := endX

		// For multi-line selections, use full line width for middle lines
		if y > startY && y < endY {
			lineStart = 0
			lineEnd = m.buffer.width - 1
		}

		// Track styled text regions to avoid duplication
		x := lineStart
		for x <= lineEnd {
			cell := m.buffer.GetContent(x, y)

			// Check if this cell has styled text
			if cell.StyledText != "" {
				// This is the start of a styled text region
				// Extract visible characters from styled text (excluding ANSI codes)
				visibleChars := extractVisibleText(cell.StyledText)
				visibleLen := countVisibleCharsInStyledText(cell.StyledText)

				// Calculate the actual visible text we can take
				// We need to account for any offset within the styled text
				offset := 0
				if x > lineStart {
					// If we didn't start at the beginning of this styled text,
					// we need to figure out the offset
					// For now, take from current position to end of styled text or selection
					offset = 0 // We're at the start, so no offset
				}

				// Determine how many characters to take
				remainingWidth := lineEnd - x + 1
				availableChars := visibleLen - offset
				takeCount := availableChars
				if takeCount > remainingWidth {
					takeCount = remainingWidth
				}
				if takeCount > len(visibleChars) {
					takeCount = len(visibleChars)
				}

				if takeCount > 0 && offset < len(visibleChars) {
					endIdx := offset + takeCount
					if endIdx > len(visibleChars) {
						endIdx = len(visibleChars)
					}
					lineBuilder.WriteString(visibleChars[offset:endIdx])
				}

				// Skip to the end of this styled text region
				x += visibleLen
			} else if cell.Char == 0 && cell.StyledText == "" {
				// This might be a continuation cell of a styled text region
				// Scan backwards to find the start of the styled text
				styledText, spatialOffset := findStyledTextStart(m.buffer, x, y)
				if styledText != "" {
					// We found the styled text that this cell continues
					// Extract visible characters from the styled text
					visibleChars := extractVisibleText(styledText)
					visibleLen := countVisibleCharsInStyledText(styledText)

					// spatialOffset is how many cells from the start we are
					// This should approximately equal the character offset
					offset := spatialOffset
					if offset < 0 {
						offset = 0
					}
					if offset > len(visibleChars) {
						offset = len(visibleChars)
					}

					// Calculate how many characters to take
					remainingWidth := lineEnd - x + 1
					endOffset := offset + remainingWidth
					if endOffset > len(visibleChars) {
						endOffset = len(visibleChars)
					}

					if offset < endOffset && offset < len(visibleChars) {
						lineBuilder.WriteString(visibleChars[offset:endOffset])
					}

					// Skip to the end of this styled text region
					remainingLen := visibleLen - offset
					x += remainingLen
				} else {
					// No styled text found, this is just an empty cell
					lineBuilder.WriteRune(' ')
					x++
				}
			} else {
				// Regular cell: use the character
				if cell.Char != 0 {
					lineBuilder.WriteRune(cell.Char)
				} else {
					lineBuilder.WriteRune(' ')
				}
				x++
			}
		}

		// Trim trailing spaces
		line := strings.TrimRight(lineBuilder.String(), " ")
		if line != "" || y < endY {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// extractVisibleText extracts visible characters from a string, excluding ANSI escape codes.
func extractVisibleText(s string) string {
	runes := []rune(s)
	var result strings.Builder
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
			result.WriteRune(runes[i])
			i++
		}
	}
	return result.String()
}

// countVisibleCharsInStyledText counts visible characters in styled text.
func countVisibleCharsInStyledText(s string) int {
	return len([]rune(extractVisibleText(s)))
}

// findStyledTextStart scans backwards from position (x, y) to find the start of a styled text region.
// Returns the styled text and the offset of (x, y) within that styled text.
// If no styled text region is found, returns ("", 0).
//
// Note: Continuation cells have Char=' ' (space) and StyledText="".
// We need to skip these to find the actual styled text start.
func findStyledTextStart(buffer *CellBuffer, x, y int) (string, int) {
	if buffer == nil || x < 0 || y < 0 || y >= buffer.height {
		return "", 0
	}

	// Scan backwards from x to find the cell with StyledText
	for scanX := x; scanX >= 0; scanX-- {
		cell := buffer.GetContent(scanX, y)
		if cell.StyledText != "" {
			// Found the start of the styled text region
			// Calculate the offset from this start to position x
			// The offset is simply the distance from scanX to x
			offset := x - scanX
			return cell.StyledText, offset
		}
		// Also check if we hit a cell with a real character (not part of styled text)
		// Note: Continuation cells have Char=' ', so we exclude space from this check
		if cell.Char != 0 && cell.Char != ' ' {
			// Hit a regular cell with a non-space character
			// This is not part of a styled text region
			return "", 0
		}
		// Continue scanning for continuation cells (Char=' ') and empty cells (Char=0)
	}

	return "", 0
}

// GetSelectedTextCompact returns the selected text with trailing whitespace trimmed.
func (m *SelectionManager) GetSelectedTextCompact() string {
	text := m.GetSelectedText()
	return strings.TrimRight(text, "\n")
}

// GetSelectionRange returns the normalized selection range.
// Returns (startX, endX, startY, endY).
func (m *SelectionManager) GetSelectionRange() (startX, endX, startY, endY int) {
	if !m.IsActive() {
		return 0, 0, 0, 0
	}
	return m.normalize()
}

// normalize returns the selection coordinates in normalized order.
// Also clamps coordinates to buffer bounds if buffer is available.
func (m *SelectionManager) normalize() (startX, endX, startY, endY int) {
	startX = m.startX
	endX = m.currentX
	startY = m.startY
	endY = m.currentY

	// Clamp to buffer bounds if available
	if m.buffer != nil {
		startX = clampInt(startX, 0, m.buffer.width-1)
		endX = clampInt(endX, 0, m.buffer.width-1)
		startY = clampInt(startY, 0, m.buffer.height-1)
		endY = clampInt(endY, 0, m.buffer.height-1)
	}

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
func (m *SelectionManager) SelectWord(x, y int) {
	if m.buffer == nil {
		return
	}

	x = clampInt(x, 0, m.buffer.width-1)
	y = clampInt(y, 0, m.buffer.height-1)

	// Find word boundaries
	startX := x
	endX := x

	// Find start of word (going left)
	for startX > 0 {
		cell := m.buffer.GetContent(startX-1, y)
		if isWhitespace(cell.Char) {
			break
		}
		startX--
	}

	// Find end of word (going right)
	for endX < m.buffer.width-1 {
		cell := m.buffer.GetContent(endX+1, y)
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
func (m *SelectionManager) SelectLine(y int) {
	if m.buffer == nil {
		return
	}

	y = clampInt(y, 0, m.buffer.height-1)

	m.active = true
	m.startX = 0
	m.startY = y
	m.currentX = m.buffer.width - 1
	m.currentY = y
	m.anchorX = 0
	m.anchorY = y
	m.mode = SelectionModeLine
}

// SelectAll selects the entire buffer.
func (m *SelectionManager) SelectAll() {
	if m.buffer == nil {
		return
	}

	m.active = true
	m.startX = 0
	m.startY = 0
	m.currentX = m.buffer.width - 1
	m.currentY = m.buffer.height - 1
	m.anchorX = 0
	m.anchorY = 0
	m.mode = SelectionModeChar
}

// SetMode sets the selection mode.
func (m *SelectionManager) SetMode(mode SelectionMode) {
	m.mode = mode
}

// GetMode returns the current selection mode.
func (m *SelectionManager) GetMode() SelectionMode {
	return m.mode
}

// SetHighlightStyle sets the style used for selection highlighting.
func (m *SelectionManager) SetHighlightStyle(style CellStyle) {
	m.highlight = style
}

// GetHighlightStyle returns the current highlight style.
func (m *SelectionManager) GetHighlightStyle() CellStyle {
	return m.highlight
}

// ApplyHighlight applies the selection highlight to the buffer.
// This sets the Selected flag on cells in the selection range.
// The actual visual highlighting is applied during String() rendering.
func (m *SelectionManager) ApplyHighlight() {
	if !m.IsActive() || m.buffer == nil {
		return
	}

	startX, endX, startY, endY := m.normalize()

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			m.buffer.SetSelected(x, y, true)
		}
	}
}

// combineStyles combines two cell styles, with the highlight taking precedence.
func (m *SelectionManager) combineStyles(original, highlight CellStyle) CellStyle {
	result := original

	// If highlight uses reverse, that's our primary selection indicator
	if highlight.Reverse {
		result.Reverse = true
	}

	// Override colors if specified
	if highlight.Foreground != "" {
		result.Foreground = highlight.Foreground
	}
	if highlight.Background != "" {
		result.Background = highlight.Background
	}

	// Add additional styling from highlight
	if highlight.Bold {
		result.Bold = true
	}
	if highlight.Underline {
		result.Underline = true
	}
	if highlight.Italic {
		result.Italic = true
	}

	return result
}

// GetSelectedCells returns all cells in the current selection.
func (m *SelectionManager) GetSelectedCells() []struct{ X, Y int } {
	if !m.IsActive() {
		return nil
	}

	var cells []struct{ X, Y int }
	startX, endX, startY, endY := m.normalize()

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			cells = append(cells, struct{ X, Y int }{X: x, Y: y})
		}
	}

	return cells
}

// SelectionRegion represents a rectangular selection region.
type SelectionRegion struct {
	StartX int
	EndX   int
	StartY int
	EndY   int
}

// GetRegion returns the selection as a SelectionRegion.
func (m *SelectionManager) GetRegion() SelectionRegion {
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

// DefaultSelectionHighlight returns the default selection highlight style.
func DefaultSelectionHighlight() CellStyle {
	return CellStyle{
		Reverse: true, // Reverse video is most visible
		Bold:    false,
	}
}

// LightSelectionHighlight returns a light theme selection highlight.
func LightSelectionHighlight() CellStyle {
	return CellStyle{
		Background: "#4A90E2",
		Foreground: "white",
		Bold:       true,
	}
}

// DarkSelectionHighlight returns a dark theme selection highlight.
func DarkSelectionHighlight() CellStyle {
	return CellStyle{
		Background: "#607D8B",
		Foreground: "white",
		Bold:       true,
	}
}

// clampInt clamps an integer between min and max.
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// RuneWidth returns the display width of a rune.
func RuneWidth(r rune) int {
	// CJK characters are width 2
	if r >= 0x1100 {
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
