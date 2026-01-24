package render

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime"
)

// Viewport represents a visible region within a larger content area
// It enables virtual scrolling by tracking which content is currently visible
type Viewport struct {
	mu sync.RWMutex

	// Content dimensions
	contentWidth  int
	contentHeight int

	// Viewport dimensions (visible area)
	viewportWidth  int
	viewportHeight int

	// Scroll position
	scrollX int
	scrollY int

	// Visible range (cached)
	visibleStartY int
	visibleEndY   int
	visibleStartX int
	visibleEndX   int
	dirty         bool
}

// NewViewport creates a new viewport
func NewViewport(contentWidth, contentHeight, viewportWidth, viewportHeight int) *Viewport {
	return &Viewport{
		contentWidth:   contentWidth,
		contentHeight:  contentHeight,
		viewportWidth:  viewportWidth,
		viewportHeight: viewportHeight,
		scrollX:        0,
		scrollY:        0,
		dirty:          true,
	}
}

// GetVisibleRange returns the visible content range
// This is useful for determining which items to render in a virtual list
func (v *Viewport) GetVisibleRange() (startX, endX, startY, endY int) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.dirty {
		v.updateVisibleRange()
	}

	return v.visibleStartX, v.visibleEndX, v.visibleStartY, v.visibleEndY
}

// ScrollTo scrolls the viewport to the specified position
func (v *Viewport) ScrollTo(x, y int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Clamp scroll position
	v.scrollX = clamp(x, 0, v.maxScrollX())
	v.scrollY = clamp(y, 0, v.maxScrollY())
	v.dirty = true
}

// ScrollBy scrolls the viewport by the specified delta
func (v *Viewport) ScrollBy(dx, dy int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.scrollX = clamp(v.scrollX+dx, 0, v.maxScrollX())
	v.scrollY = clamp(v.scrollY+dy, 0, v.maxScrollY())
	v.dirty = true
}

// ScrollToTop scrolls to the top of the content
func (v *Viewport) ScrollToTop() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.scrollY = 0
	v.dirty = true
}

// ScrollToBottom scrolls to the bottom of the content
func (v *Viewport) ScrollToBottom() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.scrollY = v.maxScrollY()
	v.dirty = true
}

// ScrollPageDown scrolls down by one page
func (v *Viewport) ScrollPageDown() {
	v.ScrollBy(0, v.viewportHeight)
}

// ScrollPageUp scrolls up by one page
func (v *Viewport) ScrollPageUp() {
	v.ScrollBy(0, -v.viewportHeight)
}

// GetScrollPosition returns the current scroll position
func (v *Viewport) GetScrollPosition() (x, y int) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.scrollX, v.scrollY
}

// GetScrollPercent returns the scroll position as a percentage (0-1)
func (v *Viewport) GetScrollPercent() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.maxScrollY() == 0 {
		return 0
	}
	return float64(v.scrollY) / float64(v.maxScrollY())
}

// SetContentSize updates the content dimensions
func (v *Viewport) SetContentSize(width, height int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.contentWidth = width
	v.contentHeight = height
	v.dirty = true

	// Adjust scroll position if needed
	if v.scrollX > v.maxScrollX() {
		v.scrollX = v.maxScrollX()
	}
	if v.scrollY > v.maxScrollY() {
		v.scrollY = v.maxScrollY()
	}
}

// SetViewportSize updates the viewport dimensions
func (v *Viewport) SetViewportSize(width, height int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.viewportWidth = width
	v.viewportHeight = height
	v.dirty = true
}

// IsVisible checks if a content region is visible in the viewport
func (v *Viewport) IsVisible(x, y, width, height int) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.dirty {
		v.updateVisibleRange()
	}

	// Check if the region intersects with the visible range
	return x+width > v.visibleStartX && x < v.visibleEndX &&
		y+height > v.visibleStartY && y < v.visibleEndY
}

// IsRowVisible checks if a row is at least partially visible
func (v *Viewport) IsRowVisible(row int) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.dirty {
		v.updateVisibleRange()
	}

	return row >= v.visibleStartY && row < v.visibleEndY
}

// IsColumnVisible checks if a column is at least partially visible
func (v *Viewport) IsColumnVisible(col int) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.dirty {
		v.updateVisibleRange()
	}

	return col >= v.visibleStartX && col < v.visibleEndX
}

// GetVisibleRows returns a slice of row indices that are visible
func (v *Viewport) GetVisibleRows(itemHeight int) []int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.dirty {
		v.updateVisibleRange()
	}

	// Calculate which items are visible
	startItem := v.visibleStartY / itemHeight
	endItem := (v.visibleEndY + itemHeight - 1) / itemHeight

	// Clamp to content bounds
	totalItems := (v.contentHeight + itemHeight - 1) / itemHeight
	if startItem < 0 {
		startItem = 0
	}
	if endItem > totalItems {
		endItem = totalItems
	}

	// Build slice of visible item indices
	items := make([]int, 0, endItem-startItem)
	for i := startItem; i < endItem; i++ {
		items = append(items, i)
	}

	return items
}

// Clamp ensures a value is within [min, max]
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// maxScrollX returns the maximum horizontal scroll position
func (v *Viewport) maxScrollX() int {
	max := v.contentWidth - v.viewportWidth
	if max < 0 {
		return 0
	}
	return max
}

// maxScrollY returns the maximum vertical scroll position
func (v *Viewport) maxScrollY() int {
	max := v.contentHeight - v.viewportHeight
	if max < 0 {
		return 0
	}
	return max
}

// updateVisibleRange updates the cached visible range
func (v *Viewport) updateVisibleRange() {
	v.visibleStartY = v.scrollY
	v.visibleEndY = v.scrollY + v.viewportHeight
	if v.visibleEndY > v.contentHeight {
		v.visibleEndY = v.contentHeight
	}

	v.visibleStartX = v.scrollX
	v.visibleEndX = v.scrollX + v.viewportWidth
	if v.visibleEndX > v.contentWidth {
		v.visibleEndX = v.contentWidth
	}

	v.dirty = false
}

// =============================================================================
// Virtual List Support
// =============================================================================

// VirtualListConfig configures a virtual list
type VirtualListConfig struct {
	ItemCount     int // Total number of items
	ItemHeight    int // Height of each item in rows
	ViewportWidth  int // Visible width
	ViewportHeight int // Visible height
}

// VirtualListState represents the state of a virtual list
type VirtualListState struct {
	Viewport     *Viewport
	SelectedIndex int
}

// NewVirtualListState creates a new virtual list state
func NewVirtualListState(config VirtualListConfig) *VirtualListState {
	contentHeight := config.ItemCount * config.ItemHeight

	return &VirtualListState{
		Viewport:      NewViewport(config.ViewportWidth, contentHeight, config.ViewportWidth, config.ViewportHeight),
		SelectedIndex: -1,
	}
}

// GetVisibleItems returns the indices of items that should be rendered
func (l *VirtualListState) GetVisibleItems(itemHeight int) []int {
	return l.Viewport.GetVisibleRows(itemHeight)
}

// SelectItem selects an item and scrolls it into view if needed
func (l *VirtualListState) SelectItem(index int, itemCount, itemHeight, viewportHeight int) {
	l.SelectedIndex = index

	// Scroll to make the selected item visible
	itemY := index * itemHeight
	_, scrollY := l.Viewport.GetScrollPosition()

	if itemY < scrollY {
		// Item is above viewport, scroll up
		l.Viewport.ScrollTo(l.Viewport.scrollX, itemY)
	} else if itemY+itemHeight > scrollY+viewportHeight {
		// Item is below viewport, scroll down
		l.Viewport.ScrollTo(l.Viewport.scrollX, itemY+itemHeight-viewportHeight)
	}
}

// PageDown moves selection down by one page
func (l *VirtualListState) PageDown(itemCount, itemHeight, viewportHeight int) {
	visibleItems := viewportHeight / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	newIndex := l.SelectedIndex + visibleItems
	if newIndex >= itemCount {
		newIndex = itemCount - 1
	}
	l.SelectItem(newIndex, itemCount, itemHeight, viewportHeight)
}

// PageUp moves selection up by one page
func (l *VirtualListState) PageUp(itemCount, itemHeight, viewportHeight int) {
	visibleItems := viewportHeight / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	newIndex := l.SelectedIndex - visibleItems
	if newIndex < 0 {
		newIndex = 0
	}
	l.SelectItem(newIndex, itemCount, itemHeight, viewportHeight)
}

// ItemRenderer is a function that renders an item at a given index
type ItemRenderer func(index int, selected bool) runtime.Frame

// RenderVirtualList renders a virtual list to a buffer
func RenderVirtualList(buf *runtime.CellBuffer, state *VirtualListState, config VirtualListConfig, renderer ItemRenderer) {
	visibleItems := state.GetVisibleItems(config.ItemHeight)
	_, scrollY := state.Viewport.GetScrollPosition()

	for _, index := range visibleItems {
		// Render the item
		selected := index == state.SelectedIndex
		frame := renderer(index, selected)

		// Calculate position
		y := index*config.ItemHeight - scrollY

		// Copy frame to buffer
		if frame.Buffer != nil {
			copyFrameToBuffer(buf, frame, 0, y)
		}
	}
}

// copyFrameToBuffer copies a frame to a buffer at the specified position
func copyFrameToBuffer(dst *runtime.CellBuffer, src runtime.Frame, dstX, dstY int) {
	if dst == nil || src.Buffer == nil {
		return
	}

	srcWidth := src.Buffer.Width()
	srcHeight := src.Buffer.Height()

	for y := 0; y < srcHeight; y++ {
		for x := 0; x < srcWidth; x++ {
			if dstX+x < dst.Width() && dstY+y < dst.Height() {
				cell := src.Buffer.GetCell(x, y)
				dst.SetContent(dstX+x, dstY+y, cell.ZIndex, cell.Char, cell.Style, "")
			}
		}
	}
}
