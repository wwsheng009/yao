package paint

import (
	"github.com/mattn/go-runewidth"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Painter provides a high-level drawing interface on top of PaintContext.
// It simplifies common drawing operations while maintaining flexibility.
//
// Designer's Note:
// - Painter wraps PaintContext, providing relative coordinate drawing
// - For simple components, use Painter's convenience methods
// - For complex components, access PaintContext directly via Context()
type Painter struct {
	context *PaintContext
}

// NewPainter creates a new Painter from a PaintContext.
func NewPainter(ctx *PaintContext) *Painter {
	return &Painter{context: ctx}
}

// =============================================================================
// Coordinate Translation
// =============================================================================

// Translate creates a child Painter with an offset origin.
// This is useful for drawing child components within a container.
//
// Example:
//   childPainter := painter.Translate(5, 2, 20, 10)
//   childPainter.Print(0, 0, "Hello") // Draws at absolute (5, 2)
func (p *Painter) Translate(x, y int, w, h int) *Painter {
	// Create a new bounds rectangle for the child
	childRect := Rect{
		X:      p.context.Bounds.X + x,
		Y:      p.context.Bounds.Y + y,
		Width:  w,
		Height: h,
	}

	// Intersect with current bounds to get the clip region
	intersection := p.context.Bounds.Intersect(&childRect)
	if intersection == nil {
		// No intersection, return a painter with empty bounds
		intersection = &Rect{X: 0, Y: 0, Width: 0, Height: 0}
	}

	// Create child context
	childCtx := &PaintContext{
		Buffer:       p.context.Buffer,
		Bounds:       *intersection,
		FocusPath:    p.context.FocusPath,
		Focused:      p.context.Focused,
		Disabled:     p.context.Disabled,
		ZIndex:       p.context.ZIndex,
		DirtyTracker: p.context.DirtyTracker,
		viewportX:    p.context.viewportX,
		viewportY:    p.context.viewportY,
	}

	return NewPainter(childCtx)
}

// =============================================================================
// Basic Drawing Methods (relative coordinates)
// =============================================================================

// SetCell draws a single character at relative (x, y).
// Coordinates are relative to the Painter's origin.
func (p *Painter) SetCell(x, y int, char rune, s style.Style) {
	p.context.SetCell(x, y, char, s)
}

// Print draws a string starting at relative (x, y).
// Handles wide characters (CJK, emoji) automatically.
func (p *Painter) Print(x, y int, text string, s style.Style) {
	p.context.SetString(x, y, text, s)
}

// FillRect fills a rectangular area with a character.
func (p *Painter) FillRect(x, y, w, h int, char rune, s style.Style) {
	rect := Rect{X: x, Y: y, Width: w, Height: h}
	p.context.Fill(rect, char, s)
}

// Clear clears the painter's entire area with spaces.
func (p *Painter) Clear(s style.Style) {
	p.FillRect(0, 0, p.context.Bounds.Width, p.context.Bounds.Height, ' ', s)
}

// =============================================================================
// Border Drawing
// =============================================================================

// DrawBorder draws a border around the specified area.
func (p *Painter) DrawBorder(x, y, w, h int, s style.Style) {
	rect := Rect{X: x, Y: y, Width: w, Height: h}
	p.context.DrawBox(rect, DefaultBoxStyle.WithStyle(s))
}

// DrawBorderWithStyle draws a border with a custom box style.
func (p *Painter) DrawBorderWithStyle(x, y, w, h int, boxStyle BoxStyle) {
	rect := Rect{X: x, Y: y, Width: w, Height: h}
	p.context.DrawBox(rect, boxStyle)
}

// =============================================================================
// Text Drawing with Alignment
// =============================================================================

// DrawText draws text with alignment within a width.
// align: 0=left, 1=center, 2=right (use PaintContext's AlignLeft, AlignCenter, AlignRight)
func (p *Painter) DrawText(x, y int, text string, width int, align TextAlign, s style.Style) {
	p.context.DrawText(x, y, text, align, s)
}

// DrawLeftText draws left-aligned text (full width).
func (p *Painter) DrawLeftText(y int, text string, s style.Style) {
	p.context.DrawText(0, y, text, AlignLeft, s)
}

// DrawCenterText draws centered text (full width).
func (p *Painter) DrawCenterText(y int, text string, s style.Style) {
	p.context.DrawText(0, y, text, AlignCenter, s)
}

// DrawRightText draws right-aligned text (full width).
func (p *Painter) DrawRightText(y int, text string, s style.Style) {
	p.context.DrawText(0, y, text, AlignRight, s)
}

// =============================================================================
// Helper Methods
// =============================================================================

// Width returns the available drawing width.
func (p *Painter) Width() int {
	return p.context.Bounds.Width
}

// Height returns the available drawing height.
func (p *Painter) Height() int {
	return p.context.Bounds.Height
}

// Bounds returns the drawing bounds.
func (p *Painter) Bounds() Rect {
	return p.context.Bounds
}

// WithFocused returns a new Painter with the focused flag set.
func (p *Painter) WithFocused(focused bool) *Painter {
	newCtx := p.context.WithFocus(focused)
	return NewPainter(newCtx)
}

// WithDisabled returns a new Painter with the disabled flag set.
func (p *Painter) WithDisabled(disabled bool) *Painter {
	newCtx := p.context.WithDisabled(disabled)
	return NewPainter(newCtx)
}

// WithStyle sets the default style for subsequent operations.
// Note: This returns the style for chaining; you still pass style to drawing methods.
func (p *Painter) WithStyle(s style.Style) style.Style {
	return s
}

// =============================================================================
// Component-Specific Helpers
// =============================================================================

// DrawInputBox draws a standard input box with borders.
// width is the total width including borders.
// Returns the content width (width - 2) and x position of content start.
func (p *Painter) DrawInputBox(x, y, width int, content string, cursorPos int, isFocused bool, boxStyle, cursorStyle style.Style) (contentX, contentY int) {
	if width < 2 {
		return 0, 0
	}

	// Draw border
	p.SetCell(x, y, '[', boxStyle)
	p.SetCell(x+width-1, y, ']', boxStyle)

	// Draw content
	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	runes := []rune(content)
	if len(runes) > contentWidth {
		runes = runes[:contentWidth]
	}

	contentX = x + 1
	contentY = y

	for i, r := range runes {
		p.SetCell(x+1+i, y, r, boxStyle)
	}

	// Draw cursor if focused
	if isFocused && cursorPos >= 0 && cursorPos <= len(runes) {
		cursorRelX := cursorPos
		if cursorRelX < 0 {
			cursorRelX = 0
		}
		if cursorRelX >= contentWidth {
			cursorRelX = contentWidth - 1
		}
		// Note: Cursor drawing uses reverse style
		if cursorRelX < len(runes) {
			p.SetCell(x+1+cursorRelX, y, runes[cursorRelX], cursorStyle.Reverse(true))
		} else {
			p.SetCell(x+1+cursorRelX, y, ' ', cursorStyle.Reverse(true))
		}
	}

	return contentX, contentY
}

// DrawButton draws a standard button with borders.
// The button text is centered within the available width.
func (p *Painter) DrawButton(x, y, width, height int, label string, isFocused bool, normalStyle, focusStyle style.Style) {
	if width < 2 || height < 1 {
		return
	}

	drawStyle := normalStyle
	if isFocused {
		drawStyle = focusStyle
	}

	buttonText := "[" + label + "]"
	buttonWidth := runewidth.StringWidth(label) + 2

	// Center the button
	paddingLeft := (width - buttonWidth) / 2
	if paddingLeft < 0 {
		paddingLeft = 0
	}

	// Draw empty cells before button
	for i := 0; i < paddingLeft && i < width; i++ {
		p.SetCell(x+i, y, ' ', style.Style{})
	}

	// Draw button
	for i, r := range buttonText {
		pos := x + paddingLeft + i
		if pos < x+width {
			p.SetCell(pos, y, r, drawStyle)
		}
	}

	// Draw empty cells after button
	endPos := x + paddingLeft + buttonWidth
	for i := endPos; i < x+width; i++ {
		p.SetCell(i, y, ' ', style.Style{})
	}
}

// =============================================================================
// Direct Buffer Access (for advanced use)
// =============================================================================

// Buffer returns direct access to the underlying buffer.
// Use this only when Painter methods are insufficient.
func (p *Painter) Buffer() *Buffer {
	return p.context.Buffer
}

// Context returns the underlying PaintContext.
// Use this to access advanced features not exposed by Painter.
func (p *Painter) Context() *PaintContext {
	return p.context
}
