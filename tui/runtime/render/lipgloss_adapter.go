package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/runtime"
)

// LipglossToCellStyle converts lipgloss.Style to runtime.CellStyle
// This extracts color, bold, underline, italic properties from lipgloss style
func LipglossToCellStyle(style lipgloss.Style) runtime.CellStyle {
	cellStyle := runtime.CellStyle{}

	// Extract foreground color
	fg := style.GetForeground()
	if fg != "" && string(fg) != "NoColor" {
		cellStyle.Foreground = colorToHex(fg)
	}

	// Extract background color
	bg := style.GetBackground()
	if bg != "" && string(bg) != "NoColor" {
		cellStyle.Background = colorToHex(bg)
	}

	// Extract text styling
	cellStyle.Bold = style.GetBold()
	cellStyle.Italic = style.GetItalic()
	cellStyle.Underline = style.GetUnderline()
	cellStyle.Strikethrough = style.GetStrikethrough()
	cellStyle.Blink = style.GetBlink()
	cellStyle.Reverse = style.GetReverse()

	return cellStyle
}

// colorToHex converts lipgloss.Color to hex string
// This handles both terminal color names and hex colors
func colorToHex(color lipgloss.Color) string {
	if color == "" {
		return ""
	}

	colorStr := string(color)

	// Check for NoColor
	if colorStr == "NoColor" {
		return ""
	}

	// If it's already a hex color, return it
	if strings.HasPrefix(colorStr, "#") {
		return colorStr
	}

	// ANSI color codes (e.g., "5", "21", etc.)
	// For now, just return the string representation
	// TODO: Map ANSI color codes to hex values
	return colorStr
}

// RenderLipglossToBuffer renders lipgloss-styled text to CellBuffer
// This parses the lipgloss rendered output and writes styled cells to the buffer
func RenderLipglossToBuffer(buf *runtime.CellBuffer, text string, style lipgloss.Style, x, y, zIndex int) {
	if buf == nil || text == "" {
		return
	}

	// Render the text with lipgloss
	rendered := style.Render(text)

	// Parse the rendered output to extract ANSI codes
	// For now, we'll do a simple implementation that strips ANSI and stores the plain text
	plainText := stripANSI(rendered)

	// Get cell style from lipgloss style
	cellStyle := LipglossToCellStyle(style)

	// Write each character to the buffer
	runes := []rune(plainText)
	for i, r := range runes {
		if x+i < buf.Width && y < buf.Height {
			buf.SetCell(x+i, y, r, cellStyle, zIndex)
		}
	}
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	var result strings.Builder
	var ansiCode bool

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == '\x1b' {
			// Start of ANSI escape sequence
			ansiCode = true
			continue
		}

		if ansiCode {
			// Inside ANSI sequence, look for end marker
			if c == 'm' {
				ansiCode = false
			}
			continue
		}

		// Regular character
		result.WriteRune(rune(c))
	}

	return result.String()
}

// ApplyStyleToNode applies lipgloss styling to a LayoutNode's Style
// NOTE: This is a no-op for now. Lipgloss styling is applied during the render phase
// when writing to the CellBuffer, not to the LayoutNode's Style struct.
// The LayoutNode.Style is for layout properties (width, height, padding, etc.),
// while CellStyle is for rendering properties (colors, bold, underline, etc.).
func ApplyStyleToNode(node *runtime.LayoutNode, lipglossStyle lipgloss.Style) {
	// No-op: lipgloss styling is applied during rendering, not to layout node
	// This function is kept for API compatibility but does nothing
	_ = lipglossStyle
}

// MeasureLipglossText measures the width of text when rendered with lipgloss style
func MeasureLipglossText(text string, style lipgloss.Style) int {
	rendered := style.Render(text)
	visualText := stripANSI(rendered)
	return len([]rune(visualText))
}
