package runtime

// Scroll and Viewport support for Phase 2.
//
// This file implements scrolling functionality for TUI:
//   - Viewport calculation
//   - Content positioning with scroll offset
//   - Overflow handling (visible, hidden, scroll)
//   - Scroll boundaries and clamping
//
// This implements Phase 2 Scroll/Viewport requirements.

// ScrollPosition represents the current scroll offset.
type ScrollPosition struct {
	X int // Horizontal scroll offset
	Y int // Vertical scroll offset
}

// Scrollable is an interface that components can implement to indicate
// they support scrolling.
// Phase 2: Basic interface, will be enhanced in Phase 3
type Scrollable interface {
	// GetContentSize returns the total content size (may be larger than viewport)
	GetContentSize() (width, height int)

	// SetScrollOffset sets the scroll offset
	SetScrollOffset(x, y int)

	// GetScrollOffset returns the current scroll offset
	GetScrollOffset() (x, y int)
}

// Viewport represents a scrollable viewport with content and position.
type Viewport struct {
	ViewWidth, ViewHeight       int // Viewport dimensions
	ContentWidth, ContentHeight int // Content dimensions
	ScrollX, ScrollY            int // Scroll offset
}

// NewViewport creates a new Viewport.
func NewViewport(viewWidth, viewHeight int) *Viewport {
	return &Viewport{
		ViewWidth:     viewWidth,
		ViewHeight:    viewHeight,
		ContentWidth:  viewWidth,
		ContentHeight: viewHeight,
		ScrollX:       0,
		ScrollY:       0,
	}
}

// SetContentSize sets the content dimensions.
func (v *Viewport) SetContentSize(width, height int) {
	v.ContentWidth = width
	v.ContentHeight = height
}

// SetViewSize sets the viewport dimensions.
func (v *Viewport) SetViewSize(width, height int) {
	v.ViewWidth = width
	v.ViewHeight = height
}

// ScrollBy scrolls by the given delta.
func (v *Viewport) ScrollBy(dx, dy int) {
	v.ScrollTo(v.ScrollX+dx, v.ScrollY+dy)
}

// ScrollTo scrolls to the absolute position.
func (v *Viewport) ScrollTo(x, y int) {
	v.ScrollX = clamp(x, 0, v.maxScrollX())
	v.ScrollY = clamp(y, 0, v.maxScrollY())
}

// GetScrollOffset returns the current scroll offset.
func (v *Viewport) GetScrollOffset() (x, y int) {
	return v.ScrollX, v.ScrollY
}

// maxScrollX returns the maximum horizontal scroll offset.
func (v *Viewport) maxScrollX() int {
	return max(0, v.ContentWidth-v.ViewWidth)
}

// maxScrollY returns the maximum vertical scroll offset.
func (v *Viewport) maxScrollY() int {
	return max(0, v.ContentHeight-v.ViewHeight)
}

// GetVisibleRect returns the visible region (x, y, width, height).
func (v *Viewport) GetVisibleRect() (x, y, width, height int) {
	return v.ScrollX, v.ScrollY, v.ViewWidth, v.ViewHeight
}

// IsContentFullyVisible returns true if all content fits in viewport.
func (v *Viewport) IsContentFullyVisible() bool {
	return v.ContentWidth <= v.ViewWidth && v.ContentHeight <= v.ViewHeight
}

// CanScrollRight returns true if can scroll right.
func (v *Viewport) CanScrollRight() bool {
	return v.ScrollX < v.maxScrollX()
}

// CanScrollLeft returns true if can scroll left.
func (v *Viewport) CanScrollLeft() bool {
	return v.ScrollX > 0
}

// CanScrollDown returns true if can scroll down.
func (v *Viewport) CanScrollDown() bool {
	return v.ScrollY < v.maxScrollY()
}

// CanScrollUp returns true if can scroll up.
func (v *Viewport) CanScrollUp() bool {
	return v.ScrollY > 0
}

// CalculateViewportPosition calculates the position for rendering a node within a viewport.
// This is used during the Render phase to offset children by the scroll position.
func CalculateViewportPosition(nodeX, nodeY, viewportX, viewportY, nodeZIndex, viewportZIndex int) (int, int) {
	// Node's screen position minus scroll offset
	screenX := nodeX - viewportX
	screenY := nodeY - viewportY

	return screenX, screenY
}

// IsNodeInViewport checks if a node is visible within the viewport.
// Returns true if any part of the node is visible.
func IsNodeInViewport(nodeX, nodeY, nodeW, nodeH, viewportX, viewportY, viewportW, viewportH int) bool {
	// Calculate node's visible rect intersection with viewport
	visibleX := max(nodeX, viewportX)
	visibleY := max(nodeY, viewportY)
	visibleX2 := min(nodeX+nodeW, viewportX+viewportW)
	visibleY2 := min(nodeY+nodeH, viewportY+viewportH)

	// Check if intersection exists
	return visibleX < visibleX2 && visibleY < visibleY2
}

// ApplyOverflowStyle applies overflow styling to a node.
// This calculates the effective size and scrollbars based on overflow style.
func ApplyOverflowStyle(node *LayoutNode, parentWidth, parentHeight int) {
	if node == nil {
		return
	}

	switch node.Style.Overflow {
	case OverflowVisible:
		// No clipping, content can extend beyond bounds
		// Node size is unconstrained by parent
		node.MeasuredWidth = max(node.MeasuredWidth, parentWidth)
		node.MeasuredHeight = max(node.MeasuredHeight, parentHeight)

	case OverflowHidden:
		// Clip content to bounds
		// Node size is clamped to parent size
		node.MeasuredWidth = min(node.MeasuredWidth, parentWidth)
		node.MeasuredHeight = min(node.MeasuredHeight, parentHeight)

	case OverflowScroll:
		// Clip content and enable scrolling
		// Node size matches parent, but content may be larger
		// Scrollbars will be rendered if needed
		node.MeasuredWidth = min(node.MeasuredWidth, parentWidth)
		node.MeasuredHeight = min(node.MeasuredHeight, parentHeight)
		// v2: Scrollbar rendering will be added
	}
}
