package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// CellStyleExtended extends CellStyle with lipgloss support
type CellStyleExtended struct {
	Foreground lipgloss.TerminalColor
	Background lipgloss.TerminalColor
	Bold       bool
	Italic     bool
	Underline  bool
}

// BufferWriter defines the interface for writing to a cell buffer
// This avoids circular dependency with runtime package
type BufferWriter interface {
	Width() int
	SetContent(x, y, z int, char rune, style CellStyle, nodeID string)
}

// RuntimeCellBuffer is a minimal interface for the runtime CellBuffer
// This allows us to wrap it without importing the runtime package
type RuntimeCellBuffer interface {
	Width() int
	SetContentRuntime(x, y, z int, char rune, bold, underline, italic bool, nodeID string)
}

// CellBufferAdapter wraps RuntimeCellBuffer to implement BufferWriter
type CellBufferAdapter struct {
	buf RuntimeCellBuffer
}

// NewCellBufferAdapter creates a new adapter for a RuntimeCellBuffer
func NewCellBufferAdapter(buf RuntimeCellBuffer) BufferWriter {
	return &CellBufferAdapter{buf: buf}
}

func (a *CellBufferAdapter) Width() int {
	return a.buf.Width()
}

func (a *CellBufferAdapter) SetContent(x, y, z int, char rune, style CellStyle, nodeID string) {
	a.buf.SetContentRuntime(x, y, z, char, style.Bold, style.Underline, style.Italic, nodeID)
}

// LipglossToCell converts lipgloss.Style to CellStyleExtended
func LipglossToCell(style lipgloss.Style) CellStyleExtended {
	return CellStyleExtended{
		Foreground: style.GetForeground(),
		Background: style.GetBackground(),
		Bold:       style.GetBold(),
		Italic:     style.GetItalic(),
		Underline:  style.GetUnderline(),
	}
}

// RenderLipglossToBuffer renders a lipgloss-styled string to CellBuffer
// This parses the lipgloss output and writes styled cells to the buffer
func RenderLipglossToBuffer(buf BufferWriter, text string, style lipgloss.Style, x, y, zIndex int) {
	if buf == nil || text == "" {
		return
	}

	// Render text with lipgloss to get the final styled string
	styledText := style.Render(text)

	// Parse the styled text and write to buffer
	lines := strings.Split(styledText, "\n")
	currentY := y

	for _, line := range lines {
		// Extract ANSI sequences and content
		runes, styles := parseStyledLine(line)

		currentX := x
		for i, r := range runes {
			if currentX >= x+buf.Width() {
				break // Exceeds buffer width
			}

			cellStyle := CellStyle{}
			if i < len(styles) {
				cellStyle = convertStyleExtended(styles[i])
			}

			buf.SetContent(currentX, currentY, zIndex, r, cellStyle, "")
			currentX++
		}
		currentY++
	}
}

// parseStyledLine parses a line with ANSI escape codes
// Returns the runes and their corresponding styles
func parseStyledLine(line string) ([]rune, []CellStyleExtended) {
	runes := []rune{}
	styles := []CellStyleExtended{}

	// Simple implementation: strip ANSI codes for now
	// A full implementation would parse ANSI SGR sequences
	currentStyle := CellStyleExtended{}

	i := 0
	for i < len(line) {
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			// ANSI escape sequence - parse SGR codes
			j := i + 2
			for j < len(line) && line[j] != 'm' {
				j++
			}
			if j < len(line) {
				// Parse the SGR sequence (simplified)
				sgr := line[i+2 : j]
				currentStyle = parseSGR(sgr, currentStyle)
				i = j + 1
			} else {
				break
			}
		} else {
			r := rune(line[i])
			runes = append(runes, r)
			styles = append(styles, currentStyle)
			i++
		}
	}

	return runes, styles
}

// parseSGR parses ANSI SGR (Select Graphic Rendition) sequences
func parseSGR(sgr string, current CellStyleExtended) CellStyleExtended {
	// Simple SGR parser for common codes
	// Full implementation would handle all SGR codes
	if sgr == "" || sgr == "0" {
		return CellStyleExtended{} // Reset
	}
	if sgr == "1" {
		current.Bold = true
	}
	if sgr == "3" {
		current.Italic = true
	}
	if sgr == "4" {
		current.Underline = true
	}
	if sgr == "22" {
		current.Bold = false
	}
	if sgr == "23" {
		current.Italic = false
	}
	if sgr == "24" {
		current.Underline = false
	}
	return current
}

// convertStyleExtended converts CellStyleExtended to CellStyle
func convertStyleExtended(ext CellStyleExtended) CellStyle {
	return CellStyle{
		Bold:      ext.Bold,
		Underline: ext.Underline,
		Italic:    ext.Italic,
		// Note: v1 simplified, colors will be handled in String() output
	}
}

// ApplyLipglossStyle applies lipgloss style to text and returns the result
func ApplyLipglossStyle(text string, style lipgloss.Style) string {
	return style.Render(text)
}

// MeasureTextWithStyle measures text with lipgloss styling applied
func MeasureTextWithStyle(text string, style lipgloss.Style) (width, height int) {
	styled := style.Render(text)
	lines := strings.Split(styled, "\n")

	maxWidth := 0
	for _, line := range lines {
		// Strip ANSI codes for accurate width measurement
		cleaned := stripANSI(line)
		lineWidth := lipgloss.Width(cleaned)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return maxWidth, len(lines)
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	result := strings.Builder{}
	i := 0

	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Find the end of the escape sequence
			end := strings.Index(s[i:], "m")
			if end == -1 {
				break
			}
			i += end + 1
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// ComponentRenderer is an interface for components that can render themselves.
type ComponentRenderer interface {
	// View returns the visual representation of the component
	View() string
}

// RenderNode renders a layout node to a cell buffer using the configured style.
// This function handles the rendering logic without exposing lipgloss to the runtime package.
//
// Parameters:
//   - buf: The buffer to render to (via BufferWriter interface)
//   - text: The text content to render (from Component.View())
//   - x, y: The position to render at
//   - zIndex: The Z-index for rendering order
//   - width: The maximum width for wrapping (0 for no limit)
func RenderNode(buf BufferWriter, text string, x, y, zIndex, width int) {
	if buf == nil || text == "" {
		return
	}

	// Create a basic lipgloss style for rendering
	// v2: Components can expose their own style preferences
	style := lipgloss.NewStyle()

	// Apply width constraint if specified
	if width > 0 {
		style = style.Width(width)
	}

	// Render text to buffer using lipgloss adapter
	RenderLipglossToBuffer(buf, text, style, x, y, zIndex)
}

// RenderNodeWithStyle renders a layout node with custom styling.
// This allows components to provide their own lipgloss styles.
func RenderNodeWithStyle(buf BufferWriter, text string, style lipgloss.Style, x, y, zIndex, width int) {
	if buf == nil || text == "" {
		return
	}

	// Apply width constraint if specified
	if width > 0 {
		style = style.Width(width)
	}

	RenderLipglossToBuffer(buf, text, style, x, y, zIndex)
}
