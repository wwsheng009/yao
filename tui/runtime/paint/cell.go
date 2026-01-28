package paint

import "github.com/yaoapp/yao/tui/runtime/style"

// Cell represents a single terminal cell with content and style.
type Cell struct {
	// Char is the rune character to display.
	Char rune

	// Style is the visual style (color, attributes) for this cell.
	// We use the framework's style definition for consistency.
	Style style.Style

	// Width is the visual width of the character (usually 1 or 2).
	Width int
}
