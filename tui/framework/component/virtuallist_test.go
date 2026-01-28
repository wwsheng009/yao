package component

import (
	"testing"

	"github.com/yaoapp/yao/tui/runtime/style"
)

// =============================================================================
// VirtualList Tests
// =============================================================================

func TestVirtualList_NewVirtualList(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	ds := NewStringDataSource(items)

	config := VirtualListConfig{
		DataSource: ds,
		ItemHeight: 1,
		ItemRenderer: func(index int, item interface{}, selected bool) string {
			return item.(string)
		},
		ShowCursor: true,
	}

	vlist := NewVirtualList(config)

	if vlist == nil {
		t.Fatal("NewVirtualList returned nil")
	}

	if vlist.Type() != "virtuallist" {
		t.Errorf("Expected type 'virtuallist', got '%s'", vlist.Type())
	}

	if vlist.itemCount != 3 {
		t.Errorf("Expected itemCount 3, got %d", vlist.itemCount)
	}
}

func TestVirtualList_NewVirtualListFromStrings(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	selectedStyle := style.Style{}.Reverse(true)

	vlist := NewVirtualListFromStrings(items, selectedStyle)

	if vlist == nil {
		t.Fatal("NewVirtualListFromStrings returned nil")
	}

	if vlist.itemCount != 3 {
		t.Errorf("Expected itemCount 3, got %d", vlist.itemCount)
	}
}

func TestVirtualList_SelectItem(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	vlist := NewVirtualListFromStrings(items, style.Style{})

	// Select first item
	vlist.SelectItem(0)
	if idx := vlist.GetSelectedIndex(); idx != 0 {
		t.Errorf("Expected selected index 0, got %d", idx)
	}

	// Select second item
	vlist.SelectItem(1)
	if idx := vlist.GetSelectedIndex(); idx != 1 {
		t.Errorf("Expected selected index 1, got %d", idx)
	}

	// Try to select out of bounds (should clamp)
	vlist.SelectItem(10)
	if idx := vlist.GetSelectedIndex(); idx != 2 {
		t.Errorf("Expected selected index 2 (clamped), got %d", idx)
	}
}

func TestVirtualList_GetSelectedItem(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	vlist := NewVirtualListFromStrings(items, style.Style{})

	vlist.SelectItem(1)
	item := vlist.GetSelectedItem()

	if item == nil {
		t.Fatal("Expected non-nil item")
	}

	if str, ok := item.(string); ok {
		if str != "Banana" {
			t.Errorf("Expected 'Banana', got '%s'", str)
		}
	} else {
		t.Errorf("Expected string item, got %T", item)
	}
}

func TestVirtualList_ScrollBy(t *testing.T) {
	items := make([]string, 100)
	for i := 0; i < 100; i++ {
		items[i] = "Item"
	}
	vlist := NewVirtualListFromStrings(items, style.Style{})

	// Initial scroll position should be 0,0
	x, y := vlist.GetScrollPosition()
	if x != 0 || y != 0 {
		t.Errorf("Expected initial position (0, 0), got (%d, %d)", x, y)
	}

	// Scroll down
	vlist.ScrollBy(0, 10)
	x, y = vlist.GetScrollPosition()
	if y != 10 {
		t.Errorf("Expected y position 10, got %d", y)
	}

	// Scroll up
	vlist.ScrollBy(0, -5)
	x, y = vlist.GetScrollPosition()
	if y != 5 {
		t.Errorf("Expected y position 5, got %d", y)
	}
}

func TestVirtualList_ScrollTo(t *testing.T) {
	items := make([]string, 100)
	for i := 0; i < 100; i++ {
		items[i] = "Item"
	}
	vlist := NewVirtualListFromStrings(items, style.Style{})

	vlist.ScrollTo(0, 50)
	_, y := vlist.GetScrollPosition()
	if y != 50 {
		t.Errorf("Expected y position 50, got %d", y)
	}
}

func TestVirtualList_Refresh(t *testing.T) {
	items := []string{"Item 1", "Item 2"}
	ds := NewStringDataSource(items)
	vlist := NewVirtualList(VirtualListConfig{
		DataSource: ds,
		ItemHeight: 1,
		ItemRenderer: func(index int, item interface{}, selected bool) string {
			return item.(string)
		},
	})

	// Add item to data source
	ds.Add("Item 3")
	vlist.Refresh()

	if vlist.itemCount != 3 {
		t.Errorf("Expected itemCount 3 after refresh, got %d", vlist.itemCount)
	}
}

func TestVirtualList_SetDataSource(t *testing.T) {
	items1 := []string{"A", "B"}
	items2 := []string{"X", "Y", "Z"}

	vlist := NewVirtualListFromStrings(items1, style.Style{})
	if vlist.itemCount != 2 {
		t.Errorf("Expected initial itemCount 2, got %d", vlist.itemCount)
	}

	newDS := NewStringDataSource(items2)
	vlist.SetDataSource(newDS)

	if vlist.itemCount != 3 {
		t.Errorf("Expected itemCount 3 after SetDataSource, got %d", vlist.itemCount)
	}
}

// =============================================================================
// SearchableVirtualList Tests
// =============================================================================

func TestSearchableVirtualList_New(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	ds := NewStringDataSource(items)

	slist := NewSearchableVirtualList(VirtualListConfig{
		DataSource: ds,
		ItemHeight: 1,
		ItemRenderer: func(index int, item interface{}, selected bool) string {
			return item.(string)
		},
	})

	if slist == nil {
		t.Fatal("NewSearchableVirtualList returned nil")
	}
}

func TestSearchableVirtualList_Filter(t *testing.T) {
	items := []string{"Apple", "Banana", "Apricot", "Cherry"}
	ds := NewStringDataSource(items)

	slist := NewSearchableVirtualList(VirtualListConfig{
		DataSource: ds,
		ItemHeight: 1,
		ItemRenderer: func(index int, item interface{}, selected bool) string {
			return item.(string)
		},
	})

	// Filter for items starting with "Ap"
	slist.Filter("Ap")

	// Should have 2 items: Apple, Apricot
	if slist.itemCount != 2 {
		t.Errorf("Expected 2 items after filter, got %d", slist.itemCount)
	}

	// Clear filter
	slist.Filter("")
	if len(slist.filteredIndices) != 0 {
		t.Errorf("Expected filteredIndices to be cleared, got %d items", len(slist.filteredIndices))
	}
}

func TestSearchableVirtualList_GetOriginalIndex(t *testing.T) {
	items := []string{"Apple", "Banana", "Apricot", "Cherry"}
	ds := NewStringDataSource(items)

	slist := NewSearchableVirtualList(VirtualListConfig{
		DataSource: ds,
		ItemHeight: 1,
		ItemRenderer: func(index int, item interface{}, selected bool) string {
			return item.(string)
		},
	})

	slist.Filter("Ap")

	// First filtered item should be "Apple" at index 0
	origIdx := slist.GetOriginalIndex(0)
	if origIdx != 0 {
		t.Errorf("Expected original index 0, got %d", origIdx)
	}

	// Second filtered item should be "Apricot" at index 2
	origIdx = slist.GetOriginalIndex(1)
	if origIdx != 2 {
		t.Errorf("Expected original index 2, got %d", origIdx)
	}
}

// =============================================================================
// LazyDataSource Tests
// =============================================================================

func TestLazyDataSource_New(t *testing.T) {
	loadFunc := func(offset, limit int) []interface{} {
		result := make([]interface{}, limit)
		for i := 0; i < limit; i++ {
			result[i] = offset + i
		}
		return result
	}

	lds := NewLazyDataSource(100, 10, loadFunc)

	if lds.Count() != 100 {
		t.Errorf("Expected count 100, got %d", lds.Count())
	}
}

func TestLazyDataSource_Get(t *testing.T) {
	loadFunc := func(offset, limit int) []interface{} {
		result := make([]interface{}, limit)
		for i := 0; i < limit; i++ {
			result[i] = offset + i
		}
		return result
	}

	lds := NewLazyDataSource(100, 10, loadFunc)

	// Get item at index 5 (should be in page 0)
	item := lds.Get(5)
	if item == nil {
		t.Fatal("Expected non-nil item")
	}
	if val, ok := item.(int); ok {
		if val != 5 {
			t.Errorf("Expected value 5, got %d", val)
		}
	} else {
		t.Errorf("Expected int item, got %T", item)
	}

	// Get item at index 15 (should be in page 1)
	item = lds.Get(15)
	if item == nil {
		t.Fatal("Expected non-nil item")
	}
	if val, ok := item.(int); ok {
		if val != 15 {
			t.Errorf("Expected value 15, got %d", val)
		}
	}
}

func TestLazyDataSource_Prefetch(t *testing.T) {
	loadFunc := func(offset, limit int) []interface{} {
		result := make([]interface{}, limit)
		for i := 0; i < limit; i++ {
			result[i] = offset + i
		}
		return result
	}

	lds := NewLazyDataSource(100, 10, loadFunc)

	// Prefetch page 2
	lds.Prefetch(2)

	// Give some time for async load (in real tests, use sync mechanisms)
	// For now, just verify no panic
}

func TestLazyDataSource_Invalidate(t *testing.T) {
	loadFunc := func(offset, limit int) []interface{} {
		result := make([]interface{}, limit)
		for i := 0; i < limit; i++ {
			result[i] = offset + i
		}
		return result
	}

	lds := NewLazyDataSource(100, 10, loadFunc)

	// Load some items
	lds.Get(5)

	// Invalidate cache
	lds.Invalidate()

	if len(lds.cachedPages) != 0 {
		t.Errorf("Expected empty cache after invalidate, got %d pages", len(lds.cachedPages))
	}
}

// =============================================================================
// PagedDataSource Tests
// =============================================================================

func TestPagedDataSource_New(t *testing.T) {
	fetcher := func(page, pageSize int) ([]interface{}, error) {
		result := make([]interface{}, pageSize)
		for i := 0; i < pageSize; i++ {
			result[i] = page*pageSize + i
		}
		return result, nil
	}

	pds := NewPagedDataSource(10, fetcher)

	if pds.Count() != 0 {
		t.Errorf("Expected initial count 0, got %d", pds.Count())
	}

	if !pds.HasMore() {
		t.Error("Expected HasMore to be true initially")
	}
}

func TestPagedDataSource_LoadNext(t *testing.T) {
	fetcher := func(page, pageSize int) ([]interface{}, error) {
		result := make([]interface{}, pageSize)
		for i := 0; i < pageSize; i++ {
			result[i] = page*pageSize + i
		}
		return result, nil
	}

	pds := NewPagedDataSource(10, fetcher)

	// Load first page
	if !pds.LoadNext() {
		t.Error("LoadNext failed")
	}

	if pds.Count() != 10 {
		t.Errorf("Expected count 10 after loading first page, got %d", pds.Count())
	}

	if pds.GetCurrentPage() != 0 {
		t.Errorf("Expected current page 0, got %d", pds.GetCurrentPage())
	}

	// Load second page
	if !pds.LoadNext() {
		t.Error("LoadNext failed for second page")
	}

	if pds.Count() != 20 {
		t.Errorf("Expected count 20 after loading second page, got %d", pds.Count())
	}

	if pds.GetCurrentPage() != 1 {
		t.Errorf("Expected current page 1, got %d", pds.GetCurrentPage())
	}
}

func TestPagedDataSource_Get(t *testing.T) {
	fetcher := func(page, pageSize int) ([]interface{}, error) {
		result := make([]interface{}, pageSize)
		for i := 0; i < pageSize; i++ {
			result[i] = page*pageSize + i
		}
		return result, nil
	}

	pds := NewPagedDataSource(10, fetcher)

	// Get item should trigger page load
	item := pds.Get(5)
	if item == nil {
		t.Fatal("Expected non-nil item")
	}
	if val, ok := item.(int); ok {
		if val != 5 {
			t.Errorf("Expected value 5, got %d", val)
		}
	} else {
		t.Errorf("Expected int item, got %T", item)
	}
}

func TestPagedDataSource_Invalidate(t *testing.T) {
	fetcher := func(page, pageSize int) ([]interface{}, error) {
		result := make([]interface{}, pageSize)
		for i := 0; i < pageSize; i++ {
			result[i] = page*pageSize + i
		}
		return result, nil
	}

	pds := NewPagedDataSource(10, fetcher)
	pds.LoadNext()
	pds.LoadNext()

	if pds.Count() != 20 {
		t.Errorf("Expected count 20 before invalidate, got %d", pds.Count())
	}

	// Invalidate
	pds.Invalidate()

	if pds.Count() != 0 {
		t.Errorf("Expected count 0 after invalidate, got %d", pds.Count())
	}

	if pds.GetCurrentPage() != -1 {
		t.Errorf("Expected current page -1 after invalidate, got %d", pds.GetCurrentPage())
	}
}

// =============================================================================
// FilteredDataSource Tests
// =============================================================================

func TestFilteredDataSource_New(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry", "Apricot"}
	ds := NewStringDataSource(items)

	filter := func(item interface{}) bool {
		s := item.(string)
		return len(s) > 5
	}

	fds := NewFilteredDataSource(ds, filter)

	// Should have items with length > 5: Banana, Cherry, Apricot
	if fds.Count() != 3 {
		t.Errorf("Expected 3 items after filter, got %d", fds.Count())
	}
}

func TestFilteredDataSource_Get(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	ds := NewStringDataSource(items)

	filter := func(item interface{}) bool {
		s := item.(string)
		return s[0] == 'B' // Only items starting with B
	}

	fds := NewFilteredDataSource(ds, filter)

	if fds.Count() != 1 {
		t.Errorf("Expected 1 item, got %d", fds.Count())
	}

	item := fds.Get(0)
	if item == nil {
		t.Fatal("Expected non-nil item")
	}
	if str, ok := item.(string); ok {
		if str != "Banana" {
			t.Errorf("Expected 'Banana', got '%s'", str)
		}
	}
}

func TestFilteredDataSource_GetOriginalIndex(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry", "Blueberry"}
	ds := NewStringDataSource(items)

	filter := func(item interface{}) bool {
		s := item.(string)
		return s[0] == 'B'
	}

	fds := NewFilteredDataSource(ds, filter)

	// Filtered items: Banana (index 1), Blueberry (index 3)
	origIdx := fds.GetOriginalIndex(0)
	if origIdx != 1 {
		t.Errorf("Expected original index 1, got %d", origIdx)
	}

	origIdx = fds.GetOriginalIndex(1)
	if origIdx != 3 {
		t.Errorf("Expected original index 3, got %d", origIdx)
	}
}

func TestFilteredDataSource_SetFilter(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	ds := NewStringDataSource(items)

	filter := func(item interface{}) bool {
		return true // Accept all
	}

	fds := NewFilteredDataSource(ds, filter)
	if fds.Count() != 3 {
		t.Errorf("Expected 3 items, got %d", fds.Count())
	}

	// Change filter to only "A" items
	newFilter := func(item interface{}) bool {
		s := item.(string)
		return s[0] == 'A'
	}
	fds.SetFilter(newFilter)

	if fds.Count() != 1 {
		t.Errorf("Expected 1 item after SetFilter, got %d", fds.Count())
	}
}

// =============================================================================
// MultiDataSource Tests
// =============================================================================

func TestMultiDataSource_New(t *testing.T) {
	ds1 := NewStringDataSource([]string{"A", "B"})
	ds2 := NewStringDataSource([]string{"C", "D", "E"})

	mds := NewMultiDataSource(ds1, ds2)

	if mds.Count() != 5 {
		t.Errorf("Expected count 5, got %d", mds.Count())
	}
}

func TestMultiDataSource_Get(t *testing.T) {
	ds1 := NewStringDataSource([]string{"A", "B"})
	ds2 := NewStringDataSource([]string{"C", "D", "E"})

	mds := NewMultiDataSource(ds1, ds2)

	// Items from first source
	item := mds.Get(0)
	if str, ok := item.(string); ok {
		if str != "A" {
			t.Errorf("Expected 'A', got '%s'", str)
		}
	}

	item = mds.Get(1)
	if str, ok := item.(string); ok {
		if str != "B" {
			t.Errorf("Expected 'B', got '%s'", str)
		}
	}

	// Items from second source
	item = mds.Get(2)
	if str, ok := item.(string); ok {
		if str != "C" {
			t.Errorf("Expected 'C', got '%s'", str)
		}
	}
}

func TestMultiDataSource_AddSource(t *testing.T) {
	ds1 := NewStringDataSource([]string{"A"})
	mds := NewMultiDataSource(ds1)

	if mds.Count() != 1 {
		t.Errorf("Expected count 1, got %d", mds.Count())
	}

	ds2 := NewStringDataSource([]string{"B", "C"})
	mds.AddSource(ds2)

	if mds.Count() != 3 {
		t.Errorf("Expected count 3 after AddSource, got %d", mds.Count())
	}
}

// =============================================================================
// PositionCache Tests
// =============================================================================

func TestPositionCache_New(t *testing.T) {
	cache := NewPositionCache()

	if cache == nil {
		t.Fatal("NewPositionCache returned nil")
	}

	if len(cache.positions) != 0 {
		t.Errorf("Expected empty positions, got %d", len(cache.positions))
	}
}

func TestPositionCache_SetAndGetHeight(t *testing.T) {
	cache := NewPositionCache()

	// Set first item height
	cache.SetHeight(0, 2)
	if cache.GetHeight(0) != 2 {
		t.Errorf("Expected height 2, got %d", cache.GetHeight(0))
	}

	// Set second item height
	cache.SetHeight(1, 3)
	if cache.GetHeight(1) != 3 {
		t.Errorf("Expected height 3, got %d", cache.GetHeight(1))
	}

	// Get uncached item height (should return default 1)
	if cache.GetHeight(10) != 1 {
		t.Errorf("Expected default height 1, got %d", cache.GetHeight(10))
	}
}

func TestPositionCache_GetY(t *testing.T) {
	cache := NewPositionCache()

	cache.SetHeight(0, 2)
	cache.SetHeight(1, 3)
	cache.SetHeight(2, 1)

	// Y positions should be cumulative
	if cache.GetY(0) != 0 {
		t.Errorf("Expected Y=0, got %d", cache.GetY(0))
	}

	if cache.GetY(1) != 2 {
		t.Errorf("Expected Y=2, got %d", cache.GetY(1))
	}

	if cache.GetY(2) != 5 {
		t.Errorf("Expected Y=5, got %d", cache.GetY(2))
	}
}

func TestPositionCache_TotalHeight(t *testing.T) {
	cache := NewPositionCache()

	if cache.TotalHeight() != 0 {
		// Empty cache returns 0
		t.Errorf("Expected total height 0 for empty cache, got %d", cache.TotalHeight())
	}

	cache.SetHeight(0, 2)
	cache.SetHeight(1, 3)

	// Total should be last position + 1 = 5 + 1 = 6
	if cache.TotalHeight() != 6 {
		t.Errorf("Expected total height 6, got %d", cache.TotalHeight())
	}
}

func TestPositionCache_FindIndexAtY(t *testing.T) {
	cache := NewPositionCache()

	cache.SetHeight(0, 2) // Y: 0-2
	cache.SetHeight(1, 3) // Y: 2-5
	cache.SetHeight(2, 1) // Y: 5-6

	// Find indices at various Y positions
	tests := []struct {
		y        int
		expected int
	}{
		{0, 0},   // In first item
		{1, 0},   // In first item
		{2, 1},   // At start of second item
		{4, 1},   // In middle of second item
		{5, 2},   // At start of third item
		{10, 2},  // Past end (returns last index)
		{-1, 0},  // Before start (returns first index)
	}

	for _, tt := range tests {
		idx := cache.FindIndexAtY(tt.y)
		if idx != tt.expected {
			t.Errorf("FindIndexAtY(%d) = %d, expected %d", tt.y, idx, tt.expected)
		}
	}
}

func TestPositionCache_Clear(t *testing.T) {
	cache := NewPositionCache()

	cache.SetHeight(0, 2)
	cache.SetHeight(1, 3)

	if len(cache.positions) != 2 {
		t.Errorf("Expected 2 positions, got %d", len(cache.positions))
	}

	cache.Clear()

	if len(cache.positions) != 0 {
		t.Errorf("Expected 0 positions after Clear, got %d", len(cache.positions))
	}
}
