// Package runtime provides the TUI runtime engine.
//
// This file implements debug utilities for inspecting rendered content.
package runtime

import (
	"fmt"
	"strings"
)

// RenderDebug provides detailed debugging information about a rendered frame.
type RenderDebug struct {
	Frame       *Frame
	LayoutResult *LayoutResult
	BufferInfo  *BufferDebugInfo
	Boxes       []BoxDebugInfo
	Summary     string
}

// BufferDebugInfo contains debug information about the CellBuffer.
type BufferDebugInfo struct {
	Width       int
	Height      int
	TotalCells  int
	NonEmpty    int
	LineLengths []int
}

// BoxDebugInfo contains debug information about a single layout box.
type BoxDebugInfo struct {
	ID      string
	X, Y    int
	Width   int
	Height  int
	Content string
}

// DebugFrame creates a detailed debug representation of a frame.
func DebugFrame(frame *Frame, layoutResult *LayoutResult) *RenderDebug {
	if frame == nil {
		return nil
	}

	debug := &RenderDebug{
		Frame:        frame,
		LayoutResult: layoutResult,
		Boxes:        make([]BoxDebugInfo, 0),
		Summary:      fmt.Sprintf("Frame: %dx%d", frame.Buffer.Width(), frame.Buffer.Height()),
	}

	if layoutResult != nil {
		debug.Summary += fmt.Sprintf(", %d boxes", len(layoutResult.Boxes))
	}

	// Analyze buffer
	debug.BufferInfo = analyzeBuffer(frame.Buffer)

	// Analyze boxes from layout result
	if layoutResult != nil {
		for _, box := range layoutResult.Boxes {
			debug.Boxes = append(debug.Boxes, debugBox(&box, frame.Buffer))
		}
	}

	return debug
}

// analyzeBuffer analyzes the CellBuffer and returns debug info.
func analyzeBuffer(buf *CellBuffer) *BufferDebugInfo {
	if buf == nil {
		return nil
	}

	width := buf.Width()
	height := buf.Height()
	info := &BufferDebugInfo{
		Width:       width,
		Height:      height,
		TotalCells:  width * height,
		LineLengths: make([]int, height),
	}

	for y := 0; y < height; y++ {
		lineLen := 0
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char != ' ' && cell.Char != 0 {
				lineLen = x + 1
				info.NonEmpty++
			}
		}
		info.LineLengths[y] = lineLen
	}

	return info
}

// debugBox creates debug info for a single box.
func debugBox(box *LayoutBox, buf *CellBuffer) BoxDebugInfo {
	info := BoxDebugInfo{
		ID:     box.NodeID,
		X:      box.X,
		Y:      box.Y,
		Width:  box.W,
		Height: box.H,
	}

	// Extract content from buffer
	if buf != nil && box.W > 0 && box.H > 0 {
		info.Content = extractBoxContent(box, buf)
	}

	return info
}

// extractBoxContent extracts the text content of a box from the buffer.
func extractBoxContent(box *LayoutBox, buf *CellBuffer) string {
	if buf == nil {
		return ""
	}

	var content strings.Builder
	maxY := minInt(box.Y+box.H, buf.Height())
	maxX := minInt(box.X+box.W, buf.Width())

	for y := box.Y; y < maxY; y++ {
		if y > box.Y {
			content.WriteByte('\n')
		}
		lastX := box.X
		for x := box.X; x < maxX; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char != ' ' && cell.Char != 0 {
				lastX = x + 1
			}
		}
		for x := box.X; x < lastX; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char != 0 {
				content.WriteRune(cell.Char)
			} else {
				content.WriteByte(' ')
			}
		}
	}

	return strings.Trim(content.String(), " \n")
}

// String returns a string representation of the debug info.
func (d *RenderDebug) String() string {
	if d == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Render Debug ===\n"))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", d.Summary))

	if d.BufferInfo != nil {
		sb.WriteString(fmt.Sprintf("Buffer: %dx%d, %d/%d cells non-empty\n\n",
			d.BufferInfo.Width, d.BufferInfo.Height,
			d.BufferInfo.NonEmpty, d.BufferInfo.TotalCells))
	}

	sb.WriteString(fmt.Sprintf("Boxes (%d):\n", len(d.Boxes)))
	for i, box := range d.Boxes {
		sb.WriteString(fmt.Sprintf("  [%d] %s at (%d,%d) size %dx%d",
			i, box.ID, box.X, box.Y, box.Width, box.Height))
		if box.Content != "" {
			content := strings.ReplaceAll(box.Content, "\n", "\\n")
			if len(content) > 40 {
				content = content[:37] + "..."
			}
			sb.WriteString(fmt.Sprintf(" content: %q", content))
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

// DiffOutput returns a formatted string showing the rendered output
// with line numbers and visible markers for whitespace/empty cells.
func (d *RenderDebug) DiffOutput() string {
	if d == nil || d.Frame == nil || d.Frame.Buffer == nil {
		return ""
	}

	buf := d.Frame.Buffer
	width := buf.Width()
	height := buf.Height()
	var result strings.Builder

	// Calculate line number width
	lineWidth := len(fmt.Sprintf("%d", height))

	for y := 0; y < height; y++ {
		// Line number
		result.WriteString(fmt.Sprintf("%*d│", lineWidth, y+1))

		// Content with visible markers
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char == 0 || cell.Char == ' ' {
				result.WriteString("·") // Mark empty/space cells visibly
			} else {
				result.WriteRune(cell.Char)
			}
		}
		result.WriteString("│\n")
	}

	return result.String()
}

// PlainOutput returns a plain text version of the rendered output.
// Useful for automated testing and comparison.
func (d *RenderDebug) PlainOutput() string {
	if d == nil || d.Frame == nil || d.Frame.Buffer == nil {
		return ""
	}

	buf := d.Frame.Buffer
	width := buf.Width()
	height := buf.Height()
	var result strings.Builder

	for y := 0; y < height; y++ {
		if y > 0 {
			result.WriteByte('\n')
		}
		lineEnd := 0
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char != 0 && cell.Char != ' ' {
				lineEnd = x + 1
			}
		}
		for x := 0; x < lineEnd; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char == 0 {
				result.WriteByte(' ')
			} else {
				result.WriteRune(cell.Char)
			}
		}
	}

	return result.String()
}

// JSONOutput returns a JSON-serializable representation of the frame.
// Useful for automated testing and comparison.
type JSONOutput struct {
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Lines  []string  `json:"lines"`
	Boxes  []BoxJSON `json:"boxes"`
	Stats  StatsJSON `json:"stats"`
}

// BoxJSON represents a box in JSON format.
type BoxJSON struct {
	ID      string `json:"id"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Content string `json:"content,omitempty"`
}

// StatsJSON represents statistics in JSON format.
type StatsJSON struct {
	TotalCells int     `json:"total_cells"`
	NonEmpty   int     `json:"non_empty_cells"`
	Density    float64 `json:"density"`
}

// ToJSON converts the debug info to a JSONOutput structure.
func (d *RenderDebug) ToJSON() *JSONOutput {
	if d == nil || d.Frame == nil {
		return nil
	}

	buf := d.Frame.Buffer
	width := buf.Width()
	height := buf.Height()
	output := &JSONOutput{
		Width:  width,
		Height: height,
		Lines:  make([]string, height),
		Boxes:  make([]BoxJSON, len(d.Boxes)),
	}

	// Extract lines
	for y := 0; y < height; y++ {
		line := make([]rune, width)
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Char == 0 {
				line[x] = ' '
			} else {
				line[x] = cell.Char
			}
		}
		output.Lines[y] = strings.TrimRight(string(line), " ")
	}

	// Extract boxes
	for i, box := range d.Boxes {
		output.Boxes[i] = BoxJSON{
			ID:      box.ID,
			X:       box.X,
			Y:       box.Y,
			Width:   box.Width,
			Height:  box.Height,
			Content: box.Content,
		}
	}

	// Calculate stats
	if d.BufferInfo != nil {
		output.Stats = StatsJSON{
			TotalCells: d.BufferInfo.TotalCells,
			NonEmpty:   d.BufferInfo.NonEmpty,
			Density:    float64(d.BufferInfo.NonEmpty) / float64(d.BufferInfo.TotalCells),
		}
	}

	return output
}

// String returns a compact string representation of JSONOutput.
func (j *JSONOutput) String() string {
	if j == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Render Output: %dx%d\n", j.Width, j.Height))
	for i, line := range j.Lines {
		if line != "" {
			sb.WriteString(fmt.Sprintf("%3d: %s\n", i+1, line))
		}
	}
	return sb.String()
}

// Compare compares two JSONOutputs and returns the differences.
func (j *JSONOutput) Compare(other *JSONOutput) []string {
	var diffs []string

	if j == nil && other == nil {
		return diffs
	}
	if j == nil {
		diffs = append(diffs, "first output is nil")
		return diffs
	}
	if other == nil {
		diffs = append(diffs, "second output is nil")
		return diffs
	}

	if j.Width != other.Width {
		diffs = append(diffs, fmt.Sprintf("width differs: %d vs %d", j.Width, other.Width))
	}
	if j.Height != other.Height {
		diffs = append(diffs, fmt.Sprintf("height differs: %d vs %d", j.Height, other.Height))
	}

	maxLines := max(len(j.Lines), len(other.Lines))
	for i := 0; i < maxLines; i++ {
		var line1, line2 string
		if i < len(j.Lines) {
			line1 = j.Lines[i]
		}
		if i < len(other.Lines) {
			line2 = other.Lines[i]
		}
		if line1 != line2 {
			diffs = append(diffs, fmt.Sprintf("line %d differs:\n  got:  %q\n  want: %q", i+1, line1, line2))
		}
	}

	return diffs
}

// Helper function
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
