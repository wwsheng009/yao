package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/runtime"
)

// TestNewViewport tests viewport creation
func TestNewViewport(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	assert.Equal(t, 100, vp.contentWidth)
	assert.Equal(t, 200, vp.contentHeight)
	assert.Equal(t, 20, vp.viewportWidth)
	assert.Equal(t, 10, vp.viewportHeight)
	assert.Equal(t, 0, vp.scrollX)
	assert.Equal(t, 0, vp.scrollY)
}

// TestViewportGetVisibleRange tests visible range calculation
func TestViewportGetVisibleRange(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	startX, endX, startY, endY := vp.GetVisibleRange()

	assert.Equal(t, 0, startX)
	assert.Equal(t, 20, endX)
	assert.Equal(t, 0, startY)
	assert.Equal(t, 10, endY)
}

// TestViewportScrollTo tests scrolling to a position
func TestViewportScrollTo(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollTo(50, 100)

	x, y := vp.GetScrollPosition()
	assert.Equal(t, 50, x)
	assert.Equal(t, 100, y)
}

// TestViewportScrollToClamping tests that scroll position is clamped
func TestViewportScrollToClamping(t *testing.T) {
	vp := NewViewport(100, 100, 20, 10)

	// Scroll past the end
	vp.ScrollTo(200, 200)

	x, y := vp.GetScrollPosition()
	assert.Equal(t, 80, x) // max scroll X = 100 - 20 = 80
	assert.Equal(t, 90, y) // max scroll Y = 100 - 10 = 90
}

// TestViewportScrollBy tests relative scrolling
func TestViewportScrollBy(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollBy(10, 20)

	x, y := vp.GetScrollPosition()
	assert.Equal(t, 10, x)
	assert.Equal(t, 20, y)
}

// TestViewportScrollToTop tests scrolling to top
func TestViewportScrollToTop(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollTo(50, 100)
	vp.ScrollToTop()

	_, y := vp.GetScrollPosition()
	assert.Equal(t, 0, y)
}

// TestViewportScrollToBottom tests scrolling to bottom
func TestViewportScrollToBottom(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollToBottom()

	_, y := vp.GetScrollPosition()
	assert.Equal(t, 190, y) // 200 - 10 = 190
}

// TestViewportScrollPageDown tests page down scrolling
func TestViewportScrollPageDown(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollPageDown()

	_, y := vp.GetScrollPosition()
	assert.Equal(t, 10, y) // Viewport height is 10

	vp.ScrollPageDown()
	_, y = vp.GetScrollPosition()
	assert.Equal(t, 20, y)
}

// TestViewportScrollPageUp tests page up scrolling
func TestViewportScrollPageUp(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollTo(0, 50)
	vp.ScrollPageUp()

	_, y := vp.GetScrollPosition()
	assert.Equal(t, 40, y) // 50 - 10 = 40

	vp.ScrollPageUp()
	_, y = vp.GetScrollPosition()
	assert.Equal(t, 30, y)
}

// TestViewportGetScrollPercent tests scroll percentage
func TestViewportGetScrollPercent(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	percent := vp.GetScrollPercent()
	assert.Equal(t, 0.0, percent)

	vp.ScrollTo(0, 95)
	percent = vp.GetScrollPercent()
	assert.InDelta(t, 0.5, percent, 0.01) // ~50%

	vp.ScrollToBottom()
	percent = vp.GetScrollPercent()
	assert.Equal(t, 1.0, percent)
}

// TestViewportSetContentSize tests updating content size
func TestViewportSetContentSize(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.SetContentSize(150, 250)

	assert.Equal(t, 150, vp.contentWidth)
	assert.Equal(t, 250, vp.contentHeight)
}

// TestViewportSetContentSizeWithScroll tests that scroll is adjusted when content shrinks
func TestViewportSetContentSizeWithScroll(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.ScrollTo(50, 100)
	vp.SetContentSize(60, 80) // Shrink content

	x, y := vp.GetScrollPosition()
	assert.Equal(t, 40, x) // Clamped to new max
	assert.Equal(t, 70, y) // Clamped to new max
}

// TestViewportSetViewportSize tests updating viewport size
func TestViewportSetViewportSize(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	vp.SetViewportSize(30, 15)

	assert.Equal(t, 30, vp.viewportWidth)
	assert.Equal(t, 15, vp.viewportHeight)
}

// TestViewportIsVisible tests visibility checking
func TestViewportIsVisible(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	// Visible item
	assert.True(t, vp.IsVisible(5, 5, 5, 5))

	// Item above viewport
	assert.False(t, vp.IsVisible(5, -10, 5, 5))

	// Item below viewport
	assert.False(t, vp.IsVisible(5, 15, 5, 5))

	// Partially visible item
	assert.True(t, vp.IsVisible(5, 8, 5, 5)) // Starts at 8, ends at 13 (viewport ends at 10)
}

// TestViewportIsRowVisible tests row visibility
func TestViewportIsRowVisible(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	assert.True(t, vp.IsRowVisible(0))
	assert.True(t, vp.IsRowVisible(5))
	assert.True(t, vp.IsRowVisible(9))
	assert.False(t, vp.IsRowVisible(10))
	assert.False(t, vp.IsRowVisible(20))
}

// TestViewportIsColumnVisible tests column visibility
func TestViewportIsColumnVisible(t *testing.T) {
	vp := NewViewport(100, 200, 20, 10)

	assert.True(t, vp.IsColumnVisible(0))
	assert.True(t, vp.IsColumnVisible(10))
	assert.True(t, vp.IsColumnVisible(19))
	assert.False(t, vp.IsColumnVisible(20))
	assert.False(t, vp.IsColumnVisible(50))
}

// TestViewportGetVisibleRows tests getting visible row indices
func TestViewportGetVisibleRows(t *testing.T) {
	vp := NewViewport(100, 100, 20, 10)

	rows := vp.GetVisibleRows(2) // Each item is 2 rows high

	// Viewport height is 10, item height is 2
	// Should return indices 0, 1, 2, 3, 4 (5 items)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, rows)
}

// TestViewportGetVisibleRowsWithScroll tests visible rows with scrolling
func TestViewportGetVisibleRowsWithScroll(t *testing.T) {
	vp := NewViewport(100, 100, 20, 10)

	vp.ScrollTo(0, 5) // Scroll down 5 rows
	rows := vp.GetVisibleRows(2) // Each item is 2 rows high

	// Viewport shows rows 5-14, which covers items 2-7
	// Item 2: rows 4-5 (row 5 visible)
	// Item 3: rows 6-7 (fully visible)
	// Item 4: rows 8-9 (fully visible)
	// Item 5: rows 10-11 (fully visible)
	// Item 6: rows 12-13 (fully visible)
	// Item 7: rows 14-15 (row 14 visible)
	assert.Equal(t, []int{2, 3, 4, 5, 6, 7}, rows)
}

// TestNewVirtualListState tests virtual list creation
func TestNewVirtualListState(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    3,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)

	assert.NotNil(t, state.Viewport)
	assert.Equal(t, -1, state.SelectedIndex)
	assert.Equal(t, 300, state.Viewport.contentHeight) // 100 * 3
}

// TestVirtualListGetVisibleItems tests getting visible items
func TestVirtualListGetVisibleItems(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)
	items := state.GetVisibleItems(config.ItemHeight)

	// Viewport shows 10 rows, each item is 2 rows
	// Should show 5 items
	assert.Equal(t, []int{0, 1, 2, 3, 4}, items)
}

// TestVirtualListSelectItem tests item selection
func TestVirtualListSelectItem(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)
	state.SelectItem(5, config.ItemCount, config.ItemHeight, config.ViewportHeight)

	assert.Equal(t, 5, state.SelectedIndex)
}

// TestVirtualListSelectItemScrollDown tests selecting an item below the viewport
func TestVirtualListSelectItemScrollDown(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)
	state.SelectItem(10, config.ItemCount, config.ItemHeight, config.ViewportHeight)

	_, scrollY := state.Viewport.GetScrollPosition()
	// Item 10 starts at row 20, viewport shows rows 0-9
	// Should scroll to make item 10 visible
	assert.GreaterOrEqual(t, scrollY, 11) // At least row 11 should be visible
}

// TestVirtualListPageDown tests page down
func TestVirtualListPageDown(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)
	state.SelectItem(0, config.ItemCount, config.ItemHeight, config.ViewportHeight)
	state.PageDown(config.ItemCount, config.ItemHeight, config.ViewportHeight)

	// Viewport shows 5 items (10/2), page down should move 5 items
	assert.Equal(t, 5, state.SelectedIndex)
}

// TestVirtualListPageUp tests page up
func TestVirtualListPageUp(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)
	state.SelectItem(10, config.ItemCount, config.ItemHeight, config.ViewportHeight)
	state.PageUp(config.ItemCount, config.ItemHeight, config.ViewportHeight)

	// Viewport shows 5 items, page up should move back 5 items
	assert.Equal(t, 5, state.SelectedIndex)
}

// TestVirtualListPageUpClamping tests page up doesn't go below 0
func TestVirtualListPageUpClamping(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     100,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 10,
	}

	state := NewVirtualListState(config)
	state.SelectItem(2, config.ItemCount, config.ItemHeight, config.ViewportHeight)
	state.PageUp(config.ItemCount, config.ItemHeight, config.ViewportHeight)

	assert.Equal(t, 0, state.SelectedIndex)
}

// TestRenderVirtualList tests rendering a virtual list
func TestRenderVirtualList(t *testing.T) {
	config := VirtualListConfig{
		ItemCount:     10,
		ItemHeight:    2,
		ViewportWidth:  20,
		ViewportHeight: 6,
	}

	state := NewVirtualListState(config)
	buf := runtime.NewCellBuffer(20, 6)

	renderer := func(index int, selected bool) runtime.Frame {
		text := "Item "
		if selected {
			text += "> "
		}
		text += string(rune('0' + index))

		itemBuf := runtime.NewCellBuffer(20, 2)
		style := runtime.CellStyle{}
		if selected {
			style.Bold = true
		}

		for i, r := range text {
			itemBuf.SetContent(i, 0, 0, r, style, "")
		}

		return runtime.Frame{Buffer: itemBuf, Width: 20, Height: 2}
	}

	RenderVirtualList(buf, state, config, renderer)

	// Check that some items were rendered
	cell := buf.GetCell(0, 0)
	assert.Equal(t, 'I', cell.Char, "Should have rendered first item")

	cell = buf.GetCell(0, 2)
	assert.Equal(t, 'I', cell.Char, "Should have rendered second item")
}

// BenchmarkViewportGetVisibleRange benchmarks visible range calculation
func BenchmarkViewportGetVisibleRange(b *testing.B) {
	vp := NewViewport(1000, 10000, 80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vp.GetVisibleRange()
	}
}

// BenchmarkVirtualListGetVisibleItems benchmarks getting visible items
func BenchmarkVirtualListGetVisibleItems(b *testing.B) {
	config := VirtualListConfig{
		ItemCount:     10000,
		ItemHeight:    1,
		ViewportWidth:  80,
		ViewportHeight: 24,
	}
	state := NewVirtualListState(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.GetVisibleItems(config.ItemHeight)
	}
}
