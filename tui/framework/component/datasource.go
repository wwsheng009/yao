package component

import "sync"

// ==============================================================================
// Data Source (V3)
// ==============================================================================
// 数据源接口，用于虚拟滚动组件

// DataSource 数据源接口
// 为 List、Table 等组件提供数据抽象
type DataSource interface {
	// Count 返回数据总数
	Count() int

	// Get 获取指定索引的数据
	Get(index int) interface{}
}

// SimpleDataSource 简单数据源
// 基于 slice 的内存数据源
type SimpleDataSource struct {
	mu    sync.RWMutex
	items []interface{}
}

// NewSimpleDataSource 创建简单数据源
func NewSimpleDataSource(items []interface{}) *SimpleDataSource {
	return &SimpleDataSource{
		items: items,
	}
}

// Count 返回数据总数
func (s *SimpleDataSource) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// Get 获取指定索引的数据
func (s *SimpleDataSource) Get(index int) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index >= 0 && index < len(s.items) {
		return s.items[index]
	}
	return nil
}

// Set 设置指定索引的数据
func (s *SimpleDataSource) Set(index int, item interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index >= 0 && index < len(s.items) {
		s.items[index] = item
	}
}

// Add 添加数据
func (s *SimpleDataSource) Add(item interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// Remove 移除指定索引的数据
func (s *SimpleDataSource) Remove(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index >= 0 && index < len(s.items) {
		s.items = append(s.items[:index], s.items[index+1:]...)
	}
}

// Clear 清空数据
func (s *SimpleDataSource) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make([]interface{}, 0)
}

// Items 返回所有数据
func (s *SimpleDataSource) Items() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]interface{}, len(s.items))
	copy(result, s.items)
	return result
}

// ==============================================================================
// String Data Source
// ==============================================================================

// StringDataSource 字符串数据源
type StringDataSource struct {
	*SimpleDataSource
}

// NewStringDataSource 创建字符串数据源
func NewStringDataSource(items []string) *StringDataSource {
	wrapped := make([]interface{}, len(items))
	for i, item := range items {
		wrapped[i] = item
	}
	return &StringDataSource{
		SimpleDataSource: NewSimpleDataSource(wrapped),
	}
}

// GetString 获取字符串数据
func (s *StringDataSource) GetString(index int) string {
	item := s.Get(index)
	if str, ok := item.(string); ok {
		return str
	}
	return ""
}

// AddString 添加字符串
func (s *StringDataSource) AddString(item string) {
	s.Add(item)
}

// ==============================================================================
// Lazy Data Source (On-demand loading)
// ==============================================================================

// LazyLoadFunc is a function that loads a page of data
type LazyLoadFunc func(offset int, limit int) []interface{}

// LazyDataSource is a data source that loads data on demand
// Useful for large datasets where you don't want to load everything at once
type LazyDataSource struct {
	mu            sync.RWMutex
	loadFunc      LazyLoadFunc
	pageSize      int
	cachedPages   map[int][]interface{} // page index -> items
	totalCount    int
	loading       map[int]bool         // Currently loading pages
	errorAtIndex  map[int]error        // Load errors
}

// NewLazyDataSource creates a new lazy-loading data source
func NewLazyDataSource(totalCount int, pageSize int, loadFunc LazyLoadFunc) *LazyDataSource {
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	return &LazyDataSource{
		loadFunc:     loadFunc,
		pageSize:     pageSize,
		cachedPages:  make(map[int][]interface{}),
		totalCount:   totalCount,
		loading:      make(map[int]bool),
		errorAtIndex: make(map[int]error),
	}
}

// Count returns the total number of items
func (l *LazyDataSource) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.totalCount
}

// Get retrieves an item, loading its page if necessary
func (l *LazyDataSource) Get(index int) interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index < 0 || index >= l.totalCount {
		return nil
	}

	pageIndex := index / l.pageSize

	// Check if page is already cached
	if items, ok := l.cachedPages[pageIndex]; ok {
		offsetInPage := index % l.pageSize
		if offsetInPage < len(items) {
			return items[offsetInPage]
		}
		return nil
	}

	// Check if we're already loading this page
	if l.loading[pageIndex] {
		// Return nil while loading (could return a loading indicator)
		return nil
	}

	// Trigger async load (in real implementation, use goroutines)
	// For now, load synchronously
	l.loadPage(pageIndex)

	// Try to return the loaded item
	if items, ok := l.cachedPages[pageIndex]; ok {
		offsetInPage := index % l.pageSize
		if offsetInPage < len(items) {
			return items[offsetInPage]
		}
	}

	return nil
}

// Prefetch loads a page in advance
func (l *LazyDataSource) Prefetch(pageIndex int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.cachedPages[pageIndex]; ok {
		return // Already cached
	}

	if pageIndex < 0 || pageIndex >= (l.totalCount+l.pageSize-1)/l.pageSize {
		return // Invalid page
	}

	go l.loadPageAsync(pageIndex)
}

// Invalidate clears cached data
func (l *LazyDataSource) Invalidate() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cachedPages = make(map[int][]interface{})
	l.errorAtIndex = make(map[int]error)
}

// InvalidatePage clears a specific page from cache
func (l *LazyDataSource) InvalidatePage(pageIndex int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.cachedPages, pageIndex)
	delete(l.errorAtIndex, pageIndex)
}

// IsLoading checks if a page is currently being loaded
func (l *LazyDataSource) IsLoading(index int) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	pageIndex := index / l.pageSize
	return l.loading[pageIndex]
}

// GetError returns any error for an item
func (l *LazyDataSource) GetError(index int) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	pageIndex := index / l.pageSize
	return l.errorAtIndex[pageIndex]
}

// loadPage loads a page synchronously
func (l *LazyDataSource) loadPage(pageIndex int) {
	if l.loading[pageIndex] {
		return
	}

	l.loading[pageIndex] = true
	defer func() {
		delete(l.loading, pageIndex)
	}()

	offset := pageIndex * l.pageSize
	items := l.loadFunc(offset, l.pageSize)
	l.cachedPages[pageIndex] = items
}

// loadPageAsync loads a page asynchronously
func (l *LazyDataSource) loadPageAsync(pageIndex int) {
	l.mu.Lock()
	if l.loading[pageIndex] {
		l.mu.Unlock()
		return
	}
	l.loading[pageIndex] = true
	l.mu.Unlock()

	defer func() {
		l.mu.Lock()
		delete(l.loading, pageIndex)
		l.mu.Unlock()
	}()

	offset := pageIndex * l.pageSize
	items := l.loadFunc(offset, l.pageSize)

	l.mu.Lock()
	l.cachedPages[pageIndex] = items
	l.mu.Unlock()
}

// ==============================================================================
// Paged Data Source
// ==============================================================================

// PageFetcher is a function that fetches a specific page
type PageFetcher func(page int, pageSize int) ([]interface{}, error)

// PagedDataSource is a data source for paginated data (e.g., from an API)
// Unlike LazyDataSource, it doesn't know the total count upfront
type PagedDataSource struct {
	mu            sync.RWMutex
	fetcher       PageFetcher
	pageSize      int
	cachedPages   map[int][]interface{}
	currentPage   int
	hasMore       bool
	pagesLoaded   int
	errorAtIndex  map[int]error
}

// NewPagedDataSource creates a new paged data source
func NewPagedDataSource(pageSize int, fetcher PageFetcher) *PagedDataSource {
	if pageSize <= 0 {
		pageSize = 50
	}

	ds := &PagedDataSource{
		fetcher:      fetcher,
		pageSize:     pageSize,
		cachedPages:  make(map[int][]interface{}),
		currentPage:  -1,
		hasMore:      true,
		pagesLoaded:  0,
		errorAtIndex: make(map[int]error),
	}

	return ds
}

// Count returns the currently known count
// For paged data sources, this may grow as more pages are loaded
func (p *PagedDataSource) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return count of loaded items
	count := 0
	for i := 0; i <= p.currentPage; i++ {
		if items, ok := p.cachedPages[i]; ok {
			count += len(items)
		}
	}
	return count
}

// Get retrieves an item at the given index
// If the item's page hasn't been loaded, it will trigger a load
func (p *PagedDataSource) Get(index int) interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate which page this index is on
	pageIndex := index / p.pageSize

	// If page isn't loaded, load it
	if _, ok := p.cachedPages[pageIndex]; !ok {
		if !p.hasMore && pageIndex > p.currentPage {
			return nil
		}

		// Load the page
		p.loadPage(pageIndex)
	}

	// Get the item from the cached page
	if items, ok := p.cachedPages[pageIndex]; ok {
		offsetInPage := index % p.pageSize
		if offsetInPage < len(items) {
			return items[offsetInPage]
		}
	}

	return nil
}

// LoadNext loads the next page of data
func (p *PagedDataSource) LoadNext() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	nextPage := p.currentPage + 1
	if !p.hasMore {
		return false
	}

	return p.loadPage(nextPage)
}

// LoadPage loads a specific page
func (p *PagedDataSource) LoadPage(pageIndex int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.loadPage(pageIndex)
}

// HasMore returns true if there may be more data to load
func (p *PagedDataSource) HasMore() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.hasMore
}

// GetCurrentPage returns the current page index
func (p *PagedDataSource) GetCurrentPage() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPage
}

// Invalidate clears all cached data
func (p *PagedDataSource) Invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cachedPages = make(map[int][]interface{})
	p.currentPage = -1
	p.hasMore = true
	p.pagesLoaded = 0
	p.errorAtIndex = make(map[int]error)
}

// loadPage loads a specific page (must be called with lock held)
func (p *PagedDataSource) loadPage(pageIndex int) bool {
	// Check if already loaded
	if _, ok := p.cachedPages[pageIndex]; ok {
		return true
	}

	items, err := p.fetcher(pageIndex, p.pageSize)
	if err != nil {
		p.errorAtIndex[pageIndex] = err
		return false
	}

	p.cachedPages[pageIndex] = items

	// Update state
	if pageIndex > p.currentPage {
		p.currentPage = pageIndex
	}

	// Check if we got a full page (if not, might be no more data)
	p.hasMore = len(items) == p.pageSize
	p.pagesLoaded++

	return true
}

// Prefetch prefetches the next page
func (p *PagedDataSource) Prefetch() {
	go func() {
		if p.HasMore() {
			p.LoadNext()
		}
	}()
}

// ==============================================================================
// Multi Data Source (combines multiple data sources)
// ==============================================================================

// MultiDataSource combines multiple data sources into one
type MultiDataSource struct {
	mu     sync.RWMutex
	sources []DataSource
	offsets []int // Cumulative offsets for each source
}

// NewMultiDataSource creates a new multi data source
func NewMultiDataSource(sources ...DataSource) *MultiDataSource {
	m := &MultiDataSource{
		sources: sources,
		offsets: make([]int, len(sources)),
	}

	// Calculate offsets
	offset := 0
	for i, src := range sources {
		m.offsets[i] = offset
		offset += src.Count()
	}

	return m
}

// Count returns the total count across all sources
func (m *MultiDataSource) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	for _, src := range m.sources {
		total += src.Count()
	}
	return total
}

// Get retrieves an item, finding the correct source
func (m *MultiDataSource) Get(index int) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find which source contains this index
	for _, src := range m.sources {
		count := src.Count()
		if index < count {
			return src.Get(index)
		}
		index -= count
	}

	return nil
}

// AddSource adds a new data source
func (m *MultiDataSource) AddSource(source DataSource) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate offset directly from existing sources to avoid deadlock
	// (Count() would try to acquire a read lock while we hold write lock)
	offset := 0
	for _, src := range m.sources {
		offset += src.Count()
	}

	m.sources = append(m.sources, source)
	m.offsets = append(m.offsets, offset)
}

// ==============================================================================
// Filtered Data Source
// ==============================================================================

// FilterFunc determines if an item should be included
type FilterFunc func(item interface{}) bool

// FilteredDataSource wraps a data source with filtering
type FilteredDataSource struct {
	mu       sync.RWMutex
	source   DataSource
	filter   FilterFunc
	indices  []int // Indices in source that pass filter
	invalidateCount int // Track changes
}

// NewFilteredDataSource creates a new filtered data source
func NewFilteredDataSource(source DataSource, filter FilterFunc) *FilteredDataSource {
	f := &FilteredDataSource{
		source:  source,
		filter:  filter,
		indices: make([]int, 0),
	}

	f.rebuildIndices()
	return f
}

// Count returns the count of filtered items
func (f *FilteredDataSource) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.indices)
}

// Get retrieves a filtered item
func (f *FilteredDataSource) Get(index int) interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if index < 0 || index >= len(f.indices) {
		return nil
	}

	originalIndex := f.indices[index]
	return f.source.Get(originalIndex)
}

// SetFilter updates the filter and rebuilds indices
func (f *FilteredDataSource) SetFilter(filter FilterFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.filter = filter
	f.rebuildIndices()
}

// GetOriginalIndex returns the index in the original source
func (f *FilteredDataSource) GetOriginalIndex(filteredIndex int) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if filteredIndex < 0 || filteredIndex >= len(f.indices) {
		return -1
	}
	return f.indices[filteredIndex]
}

// rebuildIndices rebuilds the filtered indices (must be called with lock held)
func (f *FilteredDataSource) rebuildIndices() {
	f.indices = make([]int, 0)
	count := f.source.Count()

	for i := 0; i < count; i++ {
		item := f.source.Get(i)
		if f.filter(item) {
			f.indices = append(f.indices, i)
		}
	}
	f.invalidateCount++
}
