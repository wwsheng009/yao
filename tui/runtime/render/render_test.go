package render

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

// mockBuffer is a mock implementation of BufferWriter for testing
type mockBuffer struct {
	width  int
	height int
	cells  map[string]struct {
		char  rune
		style CellStyle
		z     int
	}
}

func newMockBuffer(width, height int) *mockBuffer {
	return &mockBuffer{
		width:  width,
		height: height,
		cells:  make(map[string]struct{ char rune; style CellStyle; z int }),
	}
}

func (m *mockBuffer) Width() int {
	return m.width
}

func (m *mockBuffer) SetContent(x, y, z int, char rune, style CellStyle, nodeID string) {
	key := cellKey(x, y)
	m.cells[key] = struct {
		char  rune
		style CellStyle
		z     int
	}{char: char, style: style, z: z}
}

func (m *mockBuffer) GetContent(x, y int) (rune, CellStyle) {
	key := cellKey(x, y)
	if cell, ok := m.cells[key]; ok {
		return cell.char, cell.style
	}
	return ' ', CellStyle{}
}

// mockFrameBuffer is a mock implementation of FrameBuffer for testing
type mockFrameBuffer struct {
	width  int
	height int
	cells  map[string]Cell
}

func newMockFrameBuffer(width, height int) *mockFrameBuffer {
	return &mockFrameBuffer{
		width:  width,
		height: height,
		cells:  make(map[string]Cell),
	}
}

func (m *mockFrameBuffer) GetContent(x, y int) Cell {
	key := cellKey(x, y)
	if cell, ok := m.cells[key]; ok {
		return cell
	}
	return Cell{Char: ' ', Style: CellStyle{}}
}

func (m *mockFrameBuffer) Width() int {
	return m.width
}

func (m *mockFrameBuffer) Height() int {
	return m.height
}

func (m *mockFrameBuffer) SetContent(x, y int, char rune, style CellStyle) {
	key := cellKey(x, y)
	m.cells[key] = Cell{Char: char, Style: style}
}

// TestLipglossToCell tests the lipgloss to cell style conversion
func TestLipglossToCell(t *testing.T) {
	style := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Italic(true)

	cellStyle := LipglossToCell(style)

	assert.True(t, cellStyle.Bold, "Bold should be true")
	assert.True(t, cellStyle.Underline, "Underline should be true")
	assert.True(t, cellStyle.Italic, "Italic should be true")
}

// TestLipglossToCellEmpty tests empty lipgloss style conversion
func TestLipglossToCellEmpty(t *testing.T) {
	style := lipgloss.NewStyle()
	cellStyle := LipglossToCell(style)

	assert.False(t, cellStyle.Bold, "Bold should be false")
	assert.False(t, cellStyle.Underline, "Underline should be false")
	assert.False(t, cellStyle.Italic, "Italic should be false")
}

// TestRenderLipglossToBuffer tests rendering styled text to buffer
func TestRenderLipglossToBuffer(t *testing.T) {
	buf := newMockBuffer(20, 5)
	text := "Hello"

	style := lipgloss.NewStyle().Bold(true)
	RenderLipglossToBuffer(buf, text, style, 0, 0, 0)

	// Verify text was rendered
	char, _ := buf.GetContent(0, 0)
	assert.Equal(t, 'H', char, "First character should be 'H'")

	char, _ = buf.GetContent(1, 0)
	assert.Equal(t, 'e', char, "Second character should be 'e'")
}

// TestRenderLipglossToBufferMultiline tests rendering multi-line text
func TestRenderLipglossToBufferMultiline(t *testing.T) {
	buf := newMockBuffer(20, 5)
	text := "Line 1\nLine 2"

	style := lipgloss.NewStyle()
	RenderLipglossToBuffer(buf, text, style, 0, 0, 0)

	// Verify first line
	char, _ := buf.GetContent(0, 0)
	assert.Equal(t, 'L', char, "First char of line 1 should be 'L'")

	// Verify second line
	char, _ = buf.GetContent(0, 1)
	assert.Equal(t, 'L', char, "First char of line 2 should be 'L'")
}

// TestRenderLipglossToBufferWithNilBuffer tests nil buffer handling
func TestRenderLipglossToBufferWithNilBuffer(t *testing.T) {
	var buf BufferWriter = nil
	text := "Test"
	style := lipgloss.NewStyle()

	// Should not panic
	RenderLipglossToBuffer(buf, text, style, 0, 0, 0)
}

// TestRenderLipglossToBufferWithEmptyText tests empty text handling
func TestRenderLipglossToBufferWithEmptyText(t *testing.T) {
	buf := newMockBuffer(20, 5)
	text := ""
	style := lipgloss.NewStyle()

	// Should not panic
	RenderLipglossToBuffer(buf, text, style, 0, 0, 0)
}

// TestApplyLipglossStyle tests applying lipgloss style to text
func TestApplyLipglossStyle(t *testing.T) {
	text := "Hello World"
	style := lipgloss.NewStyle().Bold(true)

	result := ApplyLipglossStyle(text, style)
	assert.Contains(t, result, "Hello World", "Should contain original text")
}

// TestMeasureTextWithStyle tests measuring text with style
func TestMeasureTextWithStyle(t *testing.T) {
	text := "Hello"
	style := lipgloss.NewStyle()

	width, height := MeasureTextWithStyle(text, style)
	assert.Equal(t, 5, width, "Width should be 5")
	assert.Equal(t, 1, height, "Height should be 1")
}

// TestMeasureTextWithStyleMultiline tests measuring multi-line text
func TestMeasureTextWithStyleMultiline(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	style := lipgloss.NewStyle()

	width, height := MeasureTextWithStyle(text, style)
	assert.Equal(t, 6, width, "Width should be 6 (longest line)")
	assert.Equal(t, 3, height, "Height should be 3 (three lines)")
}

// TestStripANSI tests stripping ANSI codes
func TestStripANSI(t *testing.T) {
	// ANSI escape sequence for bold
	textWithANSI := "\x1b[1mBold Text\x1b[0m"
	stripped := stripANSI(textWithANSI)

	assert.Equal(t, "Bold Text", stripped, "Should strip ANSI codes")
}

// TestStripANSIWithEmptyString tests empty string handling
func TestStripANSIWithEmptyString(t *testing.T) {
	stripped := stripANSI("")
	assert.Equal(t, "", stripped, "Empty string should remain empty")
}

// TestStripANSIWithNoANSICodes tests string without ANSI codes
func TestStripANSIWithNoANSICodes(t *testing.T) {
	text := "Plain text"
	stripped := stripANSI(text)
	assert.Equal(t, "Plain text", stripped, "Plain text should be unchanged")
}

// TestComputeDiff tests frame diffing
func TestComputeDiff(t *testing.T) {
	buf1 := newMockFrameBuffer(10, 5)
	buf2 := newMockFrameBuffer(10, 5)

	// Set different content
	buf1.SetContent(0, 0, 'A', CellStyle{})
	buf2.SetContent(0, 0, 'B', CellStyle{})

	frame1 := Frame{Buffer: buf1, Width: 10, Height: 5}
	frame2 := Frame{Buffer: buf2, Width: 10, Height: 5}

	result := ComputeDiff(frame1, frame2)

	assert.True(t, result.HasChanges, "Should detect changes")
	assert.NotEmpty(t, result.DirtyRegions, "Should have dirty regions")
}

// TestComputeDiffWithNoChanges tests identical frames
func TestComputeDiffWithNoChanges(t *testing.T) {
	buf1 := newMockFrameBuffer(10, 5)
	buf2 := newMockFrameBuffer(10, 5)

	// Set same content
	buf1.SetContent(0, 0, 'A', CellStyle{})
	buf2.SetContent(0, 0, 'A', CellStyle{})

	frame1 := Frame{Buffer: buf1, Width: 10, Height: 5}
	frame2 := Frame{Buffer: buf2, Width: 10, Height: 5}

	result := ComputeDiff(frame1, frame2)

	assert.False(t, result.HasChanges, "Should not detect changes")
	assert.Empty(t, result.DirtyRegions, "Should have no dirty regions")
}

// TestComputeDiffWithNilOldFrame tests nil old frame
func TestComputeDiffWithNilOldFrame(t *testing.T) {
	buf := newMockFrameBuffer(10, 5)

	var oldFrame Frame
	newFrame := Frame{Buffer: buf, Width: 10, Height: 5}

	result := ComputeDiff(oldFrame, newFrame)

	assert.True(t, result.HasChanges, "Should detect changes with nil old frame")
	assert.Len(t, result.DirtyRegions, 1, "Should have one dirty region (entire frame)")
}

// TestComputeDiffWithDimensionChange tests different dimensions
func TestComputeDiffWithDimensionChange(t *testing.T) {
	buf1 := newMockFrameBuffer(10, 5)
	buf2 := newMockFrameBuffer(20, 10) // Different size

	frame1 := Frame{Buffer: buf1, Width: 10, Height: 5}
	frame2 := Frame{Buffer: buf2, Width: 20, Height: 10}

	result := ComputeDiff(frame1, frame2)

	assert.True(t, result.HasChanges, "Should detect dimension changes")
	assert.Len(t, result.DirtyRegions, 1, "Should mark entire frame as dirty")
}

// TestIsEmptyDiff tests empty diff detection
func TestIsEmptyDiff(t *testing.T) {
	buf1 := newMockFrameBuffer(10, 5)
	buf2 := newMockFrameBuffer(10, 5)

	frame1 := Frame{Buffer: buf1, Width: 10, Height: 5}
	frame2 := Frame{Buffer: buf2, Width: 10, Height: 5}

	result := ComputeDiff(frame1, frame2)
	assert.True(t, IsEmptyDiff(result), "Identical frames should have empty diff")
}

// TestGetTotalDirtyArea tests dirty area calculation
func TestGetTotalDirtyArea(t *testing.T) {
	buf1 := newMockFrameBuffer(10, 5)
	buf2 := newMockFrameBuffer(10, 5)

	// Change multiple cells
	buf1.SetContent(0, 0, 'A', CellStyle{})
	buf1.SetContent(1, 0, 'B', CellStyle{})
	buf1.SetContent(2, 0, 'C', CellStyle{})

	frame1 := Frame{Buffer: buf1, Width: 10, Height: 5}
	frame2 := Frame{Buffer: buf2, Width: 10, Height: 5}

	result := ComputeDiff(frame1, frame2)
	area := GetTotalDirtyArea(result)

	assert.Greater(t, area, 0, "Should have non-zero dirty area")
}

// TestParseSGR tests SGR parsing
func TestParseSGR(t *testing.T) {
	tests := []struct {
		name     string
		sgr      string
		current  CellStyleExtended
		expected CellStyleExtended
	}{
		{"Reset", "0", CellStyleExtended{Bold: true}, CellStyleExtended{}},
		{"Bold", "1", CellStyleExtended{}, CellStyleExtended{Bold: true}},
		{"Italic", "3", CellStyleExtended{}, CellStyleExtended{Italic: true}},
		{"Underline", "4", CellStyleExtended{}, CellStyleExtended{Underline: true}},
		{"BoldOff", "22", CellStyleExtended{Bold: true}, CellStyleExtended{Bold: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSGR(tt.sgr, tt.current)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCellStyleEquality tests CellStyle struct equality
func TestCellStyleEquality(t *testing.T) {
	style1 := CellStyle{Bold: true, Underline: false, Italic: true}
	style2 := CellStyle{Bold: true, Underline: false, Italic: true}
	style3 := CellStyle{Bold: false, Underline: false, Italic: true}

	assert.Equal(t, style1, style2, "Identical styles should be equal")
	assert.NotEqual(t, style1, style3, "Different styles should not be equal")
}
