package component

import (
	"sort"
	"sync"

	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
	"github.com/yaoapp/yao/tui/runtime/render"
	"github.com/yaoapp/yao/tui/runtime/style"
)

// =============================================================================
// VirtualList Component
// =============================================================================
// VirtualList is a virtual scrolling list component that only renders visible items.
// This enables efficient display of large datasets (millions of items).

// VirtualListConfig configures a VirtualList component
type VirtualListConfig struct {
	// DataSource provides the data for the list
	DataSource DataSource

	// ItemHeight is the height of each item in rows (default: 1)
	ItemHeight int

	// ItemRenderer renders a single item
	// Returns a Frame (or just the content as a string)
	ItemRenderer func(index int, item interface{}, selected bool) string

	// Style for the list
	Style style.Style

	// SelectedStyle for the selected item
	SelectedStyle style.Style

	// ShowCursor indicates if the selection cursor should be shown
	ShowCursor bool
}

// VirtualList is a virtual scrolling list component
type VirtualList struct {
	*BaseComponent

	mu     sync.RWMutex
	config VirtualListConfig

	// Internal state
	listState *render.VirtualListState
	viewportH int
	viewportW int

	// Cached data for rendering
	itemCount int
}

// NewVirtualList creates a new VirtualList component
func NewVirtualList(config VirtualListConfig) *VirtualList {
	v := &VirtualList{
		BaseComponent: NewBaseComponent("virtuallist"),
		config:        config,
		listState:     nil,
		viewportH:     10,
		viewportW:     80,
		itemCount:     0,
	}

	// Initialize item count from data source
	if config.DataSource != nil {
		v.itemCount = config.DataSource.Count()
	}

	// Set default item height
	if config.ItemHeight <= 0 {
		config.ItemHeight = 1
	}

	// Initialize list state with default viewport size
	v.listState = render.NewVirtualListState(render.VirtualListConfig{
		ItemCount:     v.itemCount,
		ItemHeight:    v.config.ItemHeight,
		ViewportWidth:  v.viewportW,
		ViewportHeight: v.viewportH,
	})

	return v
}

// =============================================================================
// Component Interface Implementation
// =============================================================================

// Type returns the component type
func (v *VirtualList) Type() string {
	return "virtuallist"
}

// Measure calculates the ideal size for the list
func (v *VirtualList) Measure(maxWidth, maxHeight int) (width, height int) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Default size if not specified
	if maxWidth <= 0 {
		maxWidth = 80
	}
	if maxHeight <= 0 {
		maxHeight = 24
	}

	return maxWidth, maxHeight
}

// Paint renders the virtual list to the buffer
func (v *VirtualList) Paint(ctx PaintContext, buf *paint.Buffer) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Update viewport size from context
	v.viewportW = ctx.Bounds.Width
	v.viewportH = ctx.Bounds.Height

	// Initialize or update list state
	v.updateListState()

	// Get visible items
	visibleIndices := v.listState.GetVisibleItems(v.config.ItemHeight)
	_, scrollY := v.listState.Viewport.GetScrollPosition()

	// Calculate actual Y position to start rendering
	baseY := ctx.Bounds.Y

	// Render each visible item
	for _, idx := range visibleIndices {
		item := v.config.DataSource.Get(idx)
		selected := idx == v.listState.SelectedIndex

		// Render the item
		content := v.config.ItemRenderer(idx, item, selected)

		// Calculate Y position
		itemY := idx*v.config.ItemHeight - scrollY
		renderY := baseY + itemY

		// Apply style
		s := v.config.Style
		if selected && v.config.ShowCursor {
			s = v.applySelectedStyle(s)
		}

		// Draw the item
		v.drawItem(ctx.Bounds.X, renderY, content, s, buf)

		// Mark area as dirty
		if ctx.DirtyTracker != nil {
			for dy := 0; dy < v.config.ItemHeight && renderY+dy < ctx.Bounds.Y+ctx.Bounds.Height; dy++ {
				for dx := 0; dx < v.viewportW && ctx.Bounds.X+dx < buf.Width; dx++ {
					ctx.DirtyTracker.MarkCell(ctx.Bounds.X+dx, renderY+dy)
				}
			}
		}
	}

	// Draw scroll indicator if content is larger than viewport
	v.drawScrollIndicator(ctx, buf)
}

// =============================================================================
// ActionTarget Interface Implementation
// =============================================================================

// HandleAction handles semantic actions for navigation
func (v *VirtualList) HandleAction(a *action.Action) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch a.Type {
	case action.ActionNavigateDown:
		v.navigateDown()
		return true

	case action.ActionNavigateUp:
		v.navigateUp()
		return true

	case action.ActionNavigatePageDown:
		v.listState.PageDown(v.itemCount, v.config.ItemHeight, v.viewportH)
		v.markDirtyAndNotify()
		return true

	case action.ActionNavigatePageUp:
		v.listState.PageUp(v.itemCount, v.config.ItemHeight, v.viewportH)
		v.markDirtyAndNotify()
		return true

	case action.ActionNavigateFirst:
		v.navigateToIndex(0)
		return true

	case action.ActionNavigateLast:
		v.navigateToIndex(v.itemCount - 1)
		return true

	case action.ActionSelectItem:
		// Emit selection event or callback
		v.markDirtyAndNotify()
		return true
	}

	return false
}

// =============================================================================
// Public Methods
// =============================================================================

// SetDataSource updates the data source
func (v *VirtualList) SetDataSource(ds DataSource) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.config.DataSource = ds
	v.itemCount = ds.Count()
	v.updateListState()
	v.markDirtyAndNotify()
}

// SelectItem selects an item by index
func (v *VirtualList) SelectItem(index int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.navigateToIndex(index)
}

// GetSelectedIndex returns the currently selected index
func (v *VirtualList) GetSelectedIndex() int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.listState.SelectedIndex
}

// GetSelectedItem returns the currently selected item
func (v *VirtualList) GetSelectedItem() interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.listState.SelectedIndex < 0 || v.listState.SelectedIndex >= v.itemCount {
		return nil
	}

	return v.config.DataSource.Get(v.listState.SelectedIndex)
}

// ScrollTo scrolls to a specific position
func (v *VirtualList) ScrollTo(x, y int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.listState != nil && v.listState.Viewport != nil {
		v.listState.Viewport.ScrollTo(x, y)
		v.markDirtyAndNotify()
	}
}

// ScrollBy scrolls by a relative amount
func (v *VirtualList) ScrollBy(dx, dy int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.listState != nil && v.listState.Viewport != nil {
		v.listState.Viewport.ScrollBy(dx, dy)
		v.markDirtyAndNotify()
	}
}

// GetScrollPosition returns the current scroll position
func (v *VirtualList) GetScrollPosition() (x, y int) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.listState != nil && v.listState.Viewport != nil {
		return v.listState.Viewport.GetScrollPosition()
	}
	return 0, 0
}

// Refresh refreshes the data source
func (v *VirtualList) Refresh() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.itemCount = v.config.DataSource.Count()
	v.updateListState()
	v.markDirtyAndNotify()
}

// =============================================================================
// Private Methods
// =============================================================================

// updateListState initializes or updates the internal list state
func (v *VirtualList) updateListState() {
	contentHeight := v.itemCount * v.config.ItemHeight

	if v.listState == nil {
		v.listState = render.NewVirtualListState(render.VirtualListConfig{
			ItemCount:     v.itemCount,
			ItemHeight:    v.config.ItemHeight,
			ViewportWidth:  v.viewportW,
			ViewportHeight: v.viewportH,
		})
	} else {
		// Update viewport and content size
		v.listState.Viewport.SetViewportSize(v.viewportW, v.viewportH)
		v.listState.Viewport.SetContentSize(v.viewportW, contentHeight)
	}
}

// navigateDown moves selection to the next item
func (v *VirtualList) navigateDown() {
	if v.itemCount == 0 {
		return
	}

	newIndex := v.listState.SelectedIndex + 1
	if newIndex >= v.itemCount {
		newIndex = v.itemCount - 1
	}

	v.listState.SelectItem(newIndex, v.itemCount, v.config.ItemHeight, v.viewportH)
	v.markDirtyAndNotify()
}

// navigateUp moves selection to the previous item
func (v *VirtualList) navigateUp() {
	if v.itemCount == 0 {
		return
	}

	newIndex := v.listState.SelectedIndex - 1
	if newIndex < 0 {
		newIndex = 0
	}

	v.listState.SelectItem(newIndex, v.itemCount, v.config.ItemHeight, v.viewportH)
	v.markDirtyAndNotify()
}

// navigateToIndex moves selection to a specific index
func (v *VirtualList) navigateToIndex(index int) {
	if index < 0 {
		index = 0
	}
	if index >= v.itemCount {
		index = v.itemCount - 1
	}

	v.listState.SelectItem(index, v.itemCount, v.config.ItemHeight, v.viewportH)
	v.markDirtyAndNotify()
}

// drawItem draws a single item to the buffer
func (v *VirtualList) drawItem(x, y int, content string, s style.Style, buf *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buf.Height {
		return
	}

	// Draw the content
	buf.SetString(x, y, content, s)
}

// applySelectedStyle applies the selected style (typically reverse video)
func (v *VirtualList) applySelectedStyle(base style.Style) style.Style {
	selected := v.config.SelectedStyle
	if selected.FG == "" && selected.BG == "" && !selected.IsBold() && !selected.IsUnderline() {
		// Default selection style: reverse video
		return base.Reverse(true)
	}
	return selected
}

// drawScrollIndicator draws a scroll position indicator
func (v *VirtualList) drawScrollIndicator(ctx PaintContext, buf *paint.Buffer) {
	if v.viewportH <= 2 {
		return
	}

	contentHeight := v.itemCount * v.config.ItemHeight
	if contentHeight <= v.viewportH {
		return // No scrolling needed
	}

	scrollPercent := v.listState.Viewport.GetScrollPercent()
	indicatorSize := maxInt(1, v.viewportH*v.viewportH/contentHeight)

	// Calculate indicator position
	indicatorY := int(float64(v.viewportH-indicatorSize) * scrollPercent)

	// Draw indicator on the right edge
	indicatorX := ctx.Bounds.X + ctx.Bounds.Width - 1
	baseY := ctx.Bounds.Y

	for i := 0; i < indicatorSize && baseY+indicatorY+i < buf.Height; i++ {
		y := baseY + indicatorY + i
		if y >= 0 && y < buf.Height {
			buf.SetCell(indicatorX, y, '▐', style.Style{})
		}
	}
}

// markDirtyAndNotify marks the component as dirty
func (v *VirtualList) markDirtyAndNotify() {
	v.MarkDirty()
}

// =============================================================================
// Helper Types
// =============================================================================

// VirtualListItem is a simple item with display text and optional data
type VirtualListItem struct {
	// Text is the display text
	Text string

	// Data is optional associated data
	Data interface{}
}

// NewVirtualListFromStrings creates a VirtualList from a slice of strings
func NewVirtualListFromStrings(items []string, selectedStyle style.Style) *VirtualList {
	ds := NewStringDataSource(items)

	return NewVirtualList(VirtualListConfig{
		DataSource: ds,
		ItemHeight: 1,
		ItemRenderer: func(index int, item interface{}, selected bool) string {
			str := item.(string)
			// Pad to viewport width for proper rendering
			return str
		},
		SelectedStyle: selectedStyle,
		ShowCursor:    true,
	})
}

// =============================================================================
// Searchable Virtual List Extension
// =============================================================================

// SearchableVirtualList adds search/filter capabilities to VirtualList
type SearchableVirtualList struct {
	*VirtualList

	filteredIndices []int
	filterText      string
	caseSensitive   bool
}

// NewSearchableVirtualList creates a new searchable virtual list
func NewSearchableVirtualList(config VirtualListConfig) *SearchableVirtualList {
	base := NewVirtualList(config)

	return &SearchableVirtualList{
		VirtualList:    base,
		filteredIndices: nil,
		filterText:      "",
		caseSensitive:   false,
	}
}

// Filter filters the list by text
func (s *SearchableVirtualList) Filter(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.filterText = text

	if text == "" {
		s.filteredIndices = nil
		s.ResetDataSource()
		return
	}

	// Build filtered indices
	indices := make([]int, 0)
	count := s.config.DataSource.Count()

	for i := 0; i < count; i++ {
		item := s.config.DataSource.Get(i)
		if s.itemMatches(item, text) {
			indices = append(indices, i)
		}
	}

	s.filteredIndices = indices

	// Create filtered data source
	filteredDS := &filteredDataSource{
		parent:    s.config.DataSource,
		indices:   indices,
		itemCache: make(map[int]interface{}),
	}

	s.config.DataSource = filteredDS
	s.itemCount = filteredDS.Count()
	s.updateListState()
	s.listState.SelectedIndex = -1
	s.MarkDirty()
}

// itemMatches checks if an item matches the filter text
func (s *SearchableVirtualList) itemMatches(item interface{}, text string) bool {
	var str string

	switch v := item.(type) {
	case string:
		str = v
	case *VirtualListItem:
		str = v.Text
	case VirtualListItem:
		str = v.Text
	default:
		str = ""
	}

	if s.caseSensitive {
		return contains(str, text)
	}
	return contains(lower(str), lower(text))
}

// ResetDataSource resets the data source to original
func (s *SearchableVirtualList) ResetDataSource() {
	// This would need to store the original data source
	// For now, just mark dirty
	s.MarkDirty()
}

// GetOriginalIndex maps a filtered index back to the original data source
func (s *SearchableVirtualList) GetOriginalIndex(filteredIndex int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filteredIndices != nil && filteredIndex >= 0 && filteredIndex < len(s.filteredIndices) {
		return s.filteredIndices[filteredIndex]
	}
	return filteredIndex
}

// =============================================================================
// Filtered Data Source
// =============================================================================

// filteredDataSource wraps a data source with filtered indices
type filteredDataSource struct {
	mu        sync.RWMutex
	parent    DataSource
	indices   []int
	itemCache map[int]interface{}
}

func (f *filteredDataSource) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.indices)
}

func (f *filteredDataSource) Get(index int) interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if index < 0 || index >= len(f.indices) {
		return nil
	}

	// Check cache
	if item, ok := f.itemCache[index]; ok {
		return item
	}

	// Get from parent
	originalIndex := f.indices[index]
	item := f.parent.Get(originalIndex)
	f.itemCache[index] = item
	return item
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func lower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// =============================================================================
// Position Cache for Variable Height Lists
// =============================================================================

// PositionCache caches the Y positions of items in a variable-height list
type PositionCache struct {
	mu       sync.RWMutex
	positions []int // Cumulative Y positions
	totalHeight int
	dirty     bool
}

// NewPositionCache creates a new position cache
func NewPositionCache() *PositionCache {
	return &PositionCache{
		positions:  make([]int, 0, 100),
		totalHeight: 0,
		dirty:      true,
	}
}

// SetHeight sets the height of an item at the given index
func (p *PositionCache) SetHeight(index int, height int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Extend positions slice if needed
	for len(p.positions) <= index {
		if len(p.positions) == 0 {
			p.positions = append(p.positions, 0)
		} else {
			lastPos := p.positions[len(p.positions)-1]
			p.positions = append(p.positions, lastPos+1) // Default height 1
		}
	}

	var oldHeight int
	if index > 0 {
		oldHeight = p.positions[index] - p.positions[index-1]
	} else {
		oldHeight = p.positions[index]
	}

	// Update positions if height changed
	if oldHeight != height {
		delta := height - oldHeight
		for i := index; i < len(p.positions); i++ {
			p.positions[i] += delta
		}
		p.totalHeight += delta
		p.dirty = true
	}
}

// GetY returns the Y position of an item (starting position)
func (p *PositionCache) GetY(index int) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if index < 0 {
		return 0
	}
	if index >= len(p.positions) {
		// Assume height 1 for uncached items
		if len(p.positions) == 0 {
			return index
		}
		return p.positions[len(p.positions)-1] + (index - len(p.positions) + 1)
	}
	// Return starting Y position
	if index == 0 {
		return 0
	}
	return p.positions[index-1]
}

// GetHeight returns the height of an item
func (p *PositionCache) GetHeight(index int) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if index < 0 || index >= len(p.positions) {
		return 1 // Default height
	}

	if index == 0 {
		return p.positions[0]
	}
	return p.positions[index] - p.positions[index-1]
}

// TotalHeight returns the total content height
func (p *PositionCache) TotalHeight() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.positions) == 0 {
		return 0
	}
	return p.positions[len(p.positions)-1] + 1
}

// FindIndexAtY finds the item index at a given Y position
func (p *PositionCache) FindIndexAtY(y int) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.positions) == 0 || y < 0 {
		return 0
	}

	// Binary search for the position
	// positions store ENDING Y positions, so we search for first position > y
	idx := sort.Search(len(p.positions), func(i int) bool {
		return p.positions[i] > y
	})

	if idx >= len(p.positions) {
		return len(p.positions) - 1
	}

	return idx
}

// Invalidate marks the cache as dirty
func (p *PositionCache) Invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dirty = true
}

// Clear clears the cache
func (p *PositionCache) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.positions = make([]int, 0, 100)
	p.totalHeight = 0
	p.dirty = true
}
