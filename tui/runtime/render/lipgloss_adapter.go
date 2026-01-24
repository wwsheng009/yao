package render

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/runtime"
)

// LipglossToCellStyle converts lipgloss.Style to runtime.CellStyle
// This extracts color, bold, underline, italic properties from lipgloss style
func LipglossToCellStyle(style lipgloss.Style) runtime.CellStyle {
	cellStyle := runtime.CellStyle{}

	// Extract foreground color using type switch for TerminalColor interface
	fg := style.GetForeground()
	if fg != nil {
		if color, ok := fg.(lipgloss.Color); ok && color != "" {
			colorStr := string(color)
			if colorStr != "NoColor" {
				cellStyle.Foreground = colorToHex(color)
			}
		}
		// Note: NoColor and other TerminalColor types are handled by returning ""
		// AdaptiveColor is not supported for now
	}

	// Extract background color using type switch for TerminalColor interface
	bg := style.GetBackground()
	if bg != nil {
		if color, ok := bg.(lipgloss.Color); ok && color != "" {
			colorStr := string(color)
			if colorStr != "NoColor" {
				cellStyle.Background = colorToHex(color)
			}
		}
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

	// Convert ANSI color codes to hex
	// Standard terminal colors
	ansiToHex := map[string]string{
		"0":   "#000000", // black
		"1":   "#800000", // red
		"2":   "#008000", // green
		"3":   "#808000", // yellow
		"4":   "#000080", // blue
		"5":   "#800080", // magenta
		"6":   "#008080", // cyan
		"7":   "#c0c0c0", // white
		"8":   "#808080", // bright black (gray)
		"9":   "#ff0000", // bright red
		"10":  "#00ff00", // bright green
		"11":  "#ffff00", // bright yellow
		"12":  "#0000ff", // bright blue
		"13":  "#ff00ff", // bright magenta
		"14":  "#00ffff", // bright cyan
		"15":  "#ffffff", // bright white
	}

	// Named color mapping (for lipgloss named colors)
	namedColors := map[string]string{
		"black":         "#000000",
		"red":           "#800000",
		"green":         "#008000",
		"yellow":        "#808000",
		"blue":          "#000080",
		"magenta":       "#800080",
		"cyan":          "#008080",
		"white":         "#c0c0c0",
		"bright-black":  "#808080",
		"bright-red":    "#ff0000",
		"bright-green":  "#00ff00",
		"bright-yellow": "#ffff00",
		"bright-blue":   "#0000ff",
		"bright-magenta": "#ff00ff",
		"bright-cyan":   "#00ffff",
		"bright-white":  "#ffffff",
	}

	// Check if it's a numeric ANSI color code
	if hex, ok := ansiToHex[colorStr]; ok {
		return hex
	}

	// Check if it's a named color
	if hex, ok := namedColors[strings.ToLower(colorStr)]; ok {
		return hex
	}

	// 256-color mode (colors 16-255)
	if num, err := strconv.Atoi(colorStr); err == nil {
		if num >= 16 && num <= 231 {
			// 216-color cube (6x6x6)
			return rgb256ToHex(num)
		} else if num >= 232 && num <= 255 {
			// Grayscale
			return gray256ToHex(num)
		}
	}

	// Return as-is (unknown color)
	return colorStr
}

// rgb256ToHex converts 216-color cube color to hex
func rgb256ToHex(color int) string {
	color = color - 16
	r := (color / 36) % 6
	g := (color / 6) % 6
	b := color % 6

	// Convert 0-5 range to 0-255
	if r > 0 {
		r = 55 + r*40
	}
	if g > 0 {
		g = 55 + g*40
	}
	if b > 0 {
		b = 55 + b*40
	}

	return "#" + toHex(r) + toHex(g) + toHex(b)
}

// gray256ToHex converts grayscale color to hex
func gray256ToHex(color int) string {
	gray := 8 + (color-232)*10
	hex := strconv.FormatInt(int64(gray), 16)
	if len(hex) == 1 {
		hex = "0" + hex
	}
	return "#" + hex + hex + hex
}

// toHex converts a number to 2-digit hex string
func toHex(n int) string {
	hex := strconv.FormatInt(int64(n), 16)
	if len(hex) == 1 {
		hex = "0" + hex
	}
	return hex
}

// RenderLipglossToBuffer renders lipgloss-styled text to CellBuffer
// This parses the lipgloss rendered output and writes styled cells to the buffer
func RenderLipglossToBuffer(buf *runtime.CellBuffer, text string, style lipgloss.Style, x, y, zIndex int) {
	if buf == nil || text == "" {
		return
	}

	// Render the text with lipgloss
	rendered := style.Render(text)

	// Parse ANSI codes and render each character with proper styling
	lines := splitLines(rendered)
	for lineIdx, line := range lines {
		// Parse line into styled segments
		segments := parseANSILine(line)

		currentX := x
		for _, segment := range segments {
			// Write each character in the segment
			runes := []rune(segment.Text)
			for i, r := range runes {
				if currentX+i < buf.Width() && y+lineIdx < buf.Height() {
					cellStyle := segment.Style
					if cellStyle.Foreground == "" {
						// Use lipgloss style as fallback
						cellStyle = LipglossToCellStyle(style)
					}
					buf.SetContent(currentX+i, y+lineIdx, zIndex, r, cellStyle, "")
				}
			}
			currentX += len(runes)
		}
	}
}

// StyledSegment represents a text segment with consistent styling
type StyledSegment struct {
	Text  string
	Style runtime.CellStyle
}

// parseANSILine parses a line of text with ANSI codes into styled segments
func parseANSILine(line string) []StyledSegment {
	var segments []StyledSegment
	currentSegment := StyledSegment{
		Style: runtime.CellStyle{},
	}

	var inAnsi bool
	var ansiCode strings.Builder

	for i := 0; i < len(line); i++ {
		c := line[i]

		if c == '\x1b' {
			// Start of ANSI escape sequence
			inAnsi = true
			ansiCode.Reset()
			continue
		}

		if inAnsi {
			// Inside ANSI sequence
			if c == 'm' {
				// End of ANSI sequence
				inAnsi = false
				// Apply the style
				newStyle := currentSegment.Style
				parseANSICode(ansiCode.String(), &newStyle)

				// Start new segment with updated style
				if currentSegment.Text != "" {
					segments = append(segments, currentSegment)
				}
				currentSegment = StyledSegment{
					Style: newStyle,
				}
			} else {
				ansiCode.WriteByte(c)
			}
			continue
		}

		// Regular character
		currentSegment.Text += string(c)
	}

	// Add final segment
	if currentSegment.Text != "" {
		segments = append(segments, currentSegment)
	}

	return segments
}

// parseANSICode parses ANSI escape sequence and updates CellStyle
func parseANSICode(code string, style *runtime.CellStyle) {
	// Remove ESC[ prefix and trailing m
	code = strings.TrimPrefix(code, "\x1b[")
	code = strings.TrimSuffix(code, "m")

	if code == "" || code == "0" {
		// Reset
		*style = runtime.CellStyle{}
		return
	}

	// Parse codes separated by ;
	codes := strings.Split(code, ";")
	for _, c := range codes {
		num, err := strconv.Atoi(c)
		if err != nil {
			continue
		}

		switch {
		case num == 1:
			style.Bold = true
		case num == 3:
			style.Italic = true
		case num == 4:
			style.Underline = true
		case num == 5:
			style.Blink = true
		case num == 7:
			style.Reverse = true
		case num == 9:
			style.Strikethrough = true
		case num == 22:
			style.Bold = false
		case num == 23:
			style.Italic = false
		case num == 24:
			style.Underline = false
		case num == 25:
			style.Blink = false
		case num == 27:
			style.Reverse = false
		case num == 29:
			style.Strikethrough = false
		case num >= 30 && num <= 37:
			// Foreground color (standard)
			style.Foreground = colorToHex(lipgloss.Color(strconv.Itoa(num - 30)))
		case num == 38:
			// Foreground color (extended 256/truecolor)
			// Skip for now, needs more parsing
		case num == 39:
			style.Foreground = ""
		case num >= 40 && num <= 47:
			// Background color (standard)
			style.Background = colorToHex(lipgloss.Color(strconv.Itoa(num - 40)))
		case num == 48:
			// Background color (extended 256/truecolor)
			// Skip for now, needs more parsing
		case num == 49:
			style.Background = ""
		}
	}
}

// splitLines splits text into lines handling both \n and \r\n
func splitLines(text string) []string {
	return strings.Split(text, "\n")
}

// stripANSI removes ANSI escape codes from a string
// DEPRECATED: Use parseANSILine for proper ANSI parsing
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

// MeasureLipglossTextHeight measures the height (number of lines) of text when rendered
func MeasureLipglossTextHeight(text string, style lipgloss.Style) int {
	rendered := style.Render(text)
	lines := splitLines(rendered)
	return len(lines)
}
