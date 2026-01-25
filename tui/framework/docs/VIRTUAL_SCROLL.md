# Virtual Scroll System Design (V3)

> **优先级**: P0 (核心功能)
> **目标**: 支持大数据量列表的高性能渲染
> **关键特性**: 惰性加载、动态缓冲、平滑滚动

## 概述

虚拟滚动是处理大数据量列表的核心技术。通过只渲染可见区域的项，而不是渲染全部数据，实现高性能的列表展示。

### 为什么需要虚拟滚动？

**没有虚拟滚动的问题**：
```go
// ❌ 传统列表：渲染所有项
type List struct {
    items []interface{}  // 可能有 100,000 项
}

func (l *List) Paint(ctx PaintContext, buf *CellBuffer) {
    for i, item := range l.items {  // 全部渲染！
        y := i + l.offset
        renderItem(buf, l.x, y, item)
    }
}

// 问题：
// - 100,000 项 = 100,000 次渲染调用
// - 内存消耗巨大
// - 滚动卡顿
```

**有虚拟滚动的优势**：
```go
// ✅ 虚拟列表：只渲染可见项
type VirtualList struct {
    dataSource DataSource      // 数据源
    viewport   Viewport         // 可见区域
    bufferSize int              // 缓冲区大小
}

func (l *VirtualList) Paint(ctx PaintContext, buf *CellBuffer) {
    start := l.viewport.Offset
    end := start + l.viewport.Count

    for i := start; i < end; i++ {
        item := l.dataSource.Get(i)  // 只获取可见项
        renderItem(buf, l.x, i-start, item)
    }
}

// 优势：
// - 只渲染 20-50 项
// - 内存占用恒定
// - 滚动流畅
```

## 设计目标

1. **高性能**: 只渲染可见项，O(1) 渲染复杂度
2. **可扩展**: 支持自定义数据源
3. **平滑滚动**: 支持像素级平滑滚动
4. **动态加载**: 支持懒加载和分页加载
5. **兼容性**: 与现有组件系统无缝集成

## 核心类型定义

### 1. DataSource 接口

```go
// 位于: tui/framework/component/datasource.go

package component

// DataSource 数据源接口
type DataSource interface {
    // Count 返回数据总数
    Count() int

    // Get 获取指定索引的数据
    Get(index int) interface{}

    // Height 返回指定项的高度（支持不等高）
    Height(index int) int
}

// SimpleDataSource 简单数据源
type SimpleDataSource struct {
    items []interface{}
}

func (s *SimpleDataSource) Count() int {
    return len(s.items)
}

func (s *SimpleDataSource) Get(index int) interface{} {
    if index >= 0 && index < len(s.items) {
        return s.items[index]
    }
    return nil
}

func (s *SimpleDataSource) Height(index int) int {
    return 1  // 默认高度为 1
}

// LazyDataSource 懒加载数据源
type LazyDataSource struct {
    loader func(page int) ([]interface{}, error)
    pages  map[int][]interface{}
    count  int
}

func (l *LazyDataSource) Count() int {
    return l.count
}

func (l *LazyDataSource) Get(index int) interface{} {
    page := index / 100
    if items, ok := l.pages[page]; ok {
        return items[index%100]
    }

    // 异步加载
    items, _ := l.loader(page)
    l.pages[page] = items
    return items[index%100]
}

// PagedDataSource 分页数据源
type PagedDataSource struct {
    fetcher func(page int) ([]interface{}, error)
    cache   map[int][]interface{}
    pages   int
    loading map[int]bool
}

func (p *PagedDataSource) Count() int {
    return p.pages * 100 // 假设每页 100 项
}

func (p *PagedDataSource) Get(index int) interface{} {
    page := index / 100
    items, ok := p.cache[page]
    if !ok {
        go p.fetchPage(page)
        return loadingPlaceholder
    }
    return items[index%100]
}
```

### 2. Viewport 视口

```go
// 位于: tui/framework/component/viewport.go

package component

// Viewport 视口
type Viewport struct {
    // 可见区域的起始索引
    Offset int

    // 可见区域的项数
    Count int

    // 总高度
    TotalHeight int

    // 可见区域的高度
    VisibleHeight int

    // 行高
    ItemHeight int
}

// NewViewport 创建视口
func NewViewport(visibleHeight int, itemHeight int) *Viewport {
    return &Viewport{
        Count:          visibleHeight / itemHeight,
        VisibleHeight:  visibleHeight,
        ItemHeight:    itemHeight,
    }
}

// UpdateOffset 更新偏移量
func (v *Viewport) UpdateOffset(totalCount int, newOffset int) {
    v.TotalHeight = totalCount * v.ItemHeight

    // 限制偏移范围
    if newOffset < 0 {
        newOffset = 0
    }
    maxOffset := totalCount - v.Count
    if newOffset > maxOffset {
        newOffset = maxOffset
    }
    v.Offset = newOffset
}

// GetVisibleRange 获取可见范围
func (v *Viewport) GetVisibleRange() (start, end int) {
    start = v.Offset
    end = v.Offset + v.Count
    return
}

// GetVisibleItems 获取可见项索引
func (v *Viewport) GetVisibleItems() []int {
    items := make([]int, v.Count)
    for i := 0; i < v.Count; i++ {
        items[i] = v.Offset + i
    }
    return items
}

// ScrollTo 滚动到指定项
func (v *Viewport) ScrollTo(index int) {
    // 将 index 滚动到视口中心
    targetOffset := index - v.Count/2
    v.UpdateOffset(v.TotalHeight/v.ItemHeight, targetOffset)
}

// ScrollBy 滚动指定行数
func (v *Viewport) ScrollBy(delta int) int {
    newOffset := v.Offset + delta
    v.UpdateOffset(v.TotalHeight/v.ItemHeight, newOffset)
    return v.Offset - newOffset  // 返回实际滚动的距离
}
```

### 3. VirtualList 组件

```go
// 位于: tui/framework/component/virtuallist.go

package component

// VirtualList 虚拟列表
type VirtualList struct {
    BaseComponent
    *Measurable
    *ThemeHolder

    // 数据源
    dataSource DataSource

    // 视口
    viewport *Viewport

    // 缓冲区
    buffer []*ListItem

    // 项目渲染器
    renderItem RenderItemFunc

    // 滚动位置（平滑滚动用）
    scrollPosition float64

    // 滚动目标
    scrollTarget *float64

    // 选中状态
    selected int

    // 焦点状态
    focused bool
}

// ListItem 列表项
type ListItem struct {
    Index     int
    Data      interface{}
    Height    int
    OffsetY   int
    Selected  bool
    Focused   bool
}

// RenderItemFunc 渲染函数
type RenderItemFunc func(buf *runtime.CellBuffer, item *ListItem)

// NewVirtualList 创建虚拟列表
func NewVirtualList(dataSource DataSource, renderItem RenderItemFunc) *VirtualList {
    list := &VirtualList{
        dataSource:   dataSource,
        renderItem:    renderItem,
        scrollPosition: 0,
        selected:      -1,
    }

    // 设置默认主题
    list.Measurable = NewMeasurable()
    list.ThemeHolder = NewThemeHolder(nil)

    return list
}

// Measure 测量尺寸
func (l *VirtualList) Measure(maxWidth, maxHeight int) (width, height int) {
    return l.dataSource.Count() * 20, l.dataSource.Count() * 1
}

// Paint 绘制列表
func (l *VirtualList) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    // 更新视口大小
    visibleHeight := l.bounds.Height
    if l.viewport == nil || l.viewport.VisibleHeight != visibleHeight {
        l.viewport = NewViewport(visibleHeight, 1)
    }

    // 更新总高度
    l.viewport.UpdateOffset(l.dataSource.Count(), int(l.scrollPosition))

    // 获取可见范围
    start, end := l.viewport.GetVisibleRange()

    // 渲染可见项
    y := 0
    for i := start; i < end; i++ {
        item := &ListItem{
            Index:    i,
            Data:     l.dataSource.Get(i),
            Height:   l.dataSource.Height(i),
            OffsetY:  y,
            Selected: i == l.selected,
            Focused:  l.focused && i == l.selected,
        }

        l.renderItem(buf, item)
        y += item.Height
    }
}

// HandleAction 处理动作
func (l *VirtualList) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateDown:
        delta := l.viewport.ScrollBy(1)
        l.scrollPosition += float64(delta)
        l.MarkDirty()
        return true

    case action.ActionNavigateUp:
        delta := l.viewport.ScrollBy(-1)
        l.scrollPosition += float64(delta)
        l.MarkDirty()
        return true

    case action.ActionNavigatePageDown:
        delta := l.viewport.ScrollBy(l.viewport.Count)
        l.scrollPosition += float64(delta)
        l.MarkDirty()
        return true

    case action.ActionNavigatePageUp:
        delta := l.viewport.ScrollBy(-l.viewport.Count)
        l.scrollPosition += float64(delta)
        l.MarkDirty()
        return true

    case action.ActionNavigateFirst:
        l.scrollPosition = 0
        l.viewport.UpdateOffset(l.dataSource.Count(), 0)
        l.MarkDirty()
        return true

    case action.ActionNavigateLast:
        l.scrollPosition = float64(l.dataSource.Count() - l.viewport.Count)
        l.viewport.UpdateOffset(l.dataSource.Count(), int(l.scrollPosition))
        l.MarkDirty()
        return true

    case action.ActionSubmit:
        if l.selected >= 0 {
            l.onSelect(l.dataSource.Get(l.selected))
        }
        return true
    }

    return false
}

// ScrollTo 滚动到指定项
func (l *VirtualList) ScrollTo(index int) {
    l.scrollTarget = new(float64(float64(index))
    l.MarkDirty()
}

// Select 选择指定项
func (l *VirtualList) Select(index int) {
    if index >= 0 && index < l.dataSource.Count() {
        l.selected = index
        l.ScrollTo(index)
    }
}

// GetSelected 获取选中项
func (l *VirtualList) GetSelected() interface{} {
    if l.selected >= 0 && l.selected < l.dataSource.Count() {
        return l.dataSource.Get(l.selected)
    }
    return nil
}

// Update 更新动画
func (l *VirtualList) Update(dt time.Duration) {
    // 平滑滚动逻辑
    if l.scrollTarget != nil {
        diff := *l.scrollTarget - l.scrollPosition
        if abs(diff) < 0.01 {
            l.scrollPosition = *l.scrollTarget
            l.scrollTarget = nil
        } else {
            // 缓动动画
            l.scrollPosition += diff * 0.3
        }
        l.MarkDirty()
    }
}
```

### 4. 不等高列表支持

```go
// 位于: tui/framework/component/virtuallist_variable_height.go

package component

// VariableHeightList 不等高列表
type VariableHeightList struct {
    *VirtualList

    // 位置缓存
    positionCache map[int]int
}

// NewVariableHeightList 创建不等高列表
func NewVariableHeightList(dataSource DataSource, renderItem RenderItemFunc) *VariableHeightList {
    base := NewVirtualList(dataSource, renderItem)

    return &VariableHeightList{
        VirtualList:   base,
        positionCache: make(map[int]int),
    }
}

// BuildPositionCache 构建位置缓存
func (l *VariableHeightList) BuildPositionCache() {
    total := 0
    for i := 0; i < l.dataSource.Count(); i++ {
        l.positionCache[i] = total
        total += l.dataSource.Height(i)
    }
}

// GetPosition 获取指定索引的位置
func (l *VariableHeightList) GetPosition(index int) int {
    if pos, ok := l.positionCache[index]; ok {
        return pos
    }

    // 缓存未命中时构建缓存
    l.BuildPositionCache()
    return l.positionCache[index]
}

// GetIndexAtPosition 根据位置获取索引
func (l *VariableHeightList) GetIndexAtPosition(pos int) int {
    // 二分查找
    low, high := 0, l.dataSource.Count()
    for low < high {
        mid := (low + high) / 2
        if l.GetPosition(mid) <= pos {
            low = mid + 1
        } else {
            high = mid
        }
    }
    return low - 1
}

// ScrollBy 按位置滚动
func (l *VariableHeightList) ScrollBy(deltaPixels int) int {
    currentPos := l.GetPosition(l.viewport.Offset)

    newPos := currentPos + deltaPixels
    if newPos < 0 {
        newPos = 0
    }

    newIndex := l.GetIndexAtPosition(newPos)
    oldIndex := l.viewport.Offset

    l.viewport.Offset = newIndex
    l.scrollPosition = float64(newIndex)

    return newIndex - oldIndex
}
```

## 集成示例

### 基础用法

```go
// 创建数据源
dataSource := &SimpleDataSource{
    items: generateLargeDataset(100000),
}

// 创建虚拟列表
list := NewVirtualList(dataSource, func(buf *CellBuffer, item *ListItem) {
    data := item.Data.(string)
    style := getListStyle(item.Selected, item.Focused)

    // 渲染文本
    for i, ch := range data {
        buf.SetCell(item.OffsetY*2, i, ch, style)
    }
})

// 设置大小
list.SetBounds(Rect{X: 10, Y: 5, Width: 60, Height: 20})

// 挂载到应用
app.Mount(list)
```

### 动态数据源

```go
// API 数据源
type APIDataSource struct {
    baseURL string
    cache   map[int][]interface{}
}

func (a *APIDataSource) Count() int {
    // 从 API 获取总数
    resp, _ := http.Get(a.baseURL + "/count")
    // ...
    return 100000
}

func (a *APIDataSource) Get(index int) interface{} {
    page := index / 100
    if items, ok := a.cache[page]; ok {
        return items[index%100]
    }

    // 异步加载
    go a.loadPage(page)

    // 返回占位符
    return loadingPlaceholder
}

func (a *APIDataSource) loadPage(page int) {
    resp, _ := http.Get(a.baseURL + fmt.Sprintf("/items?page=%d", page))
    // ...
}
```

### 搜索高亮

```go
// SearchableList 可搜索的虚拟列表
type SearchableList struct {
    *VirtualList

    // 搜索
    searchQuery string
    matches    []int
    currentMatch int
}

func (l *SearchableList) SetSearch(query string) {
    l.searchQuery = query
    l.matches = l.search(query)

    if len(l.matches) > 0 {
        l.currentMatch = 0
        l.ScrollTo(l.matches[0])
    }
}

func (l *SearchableList) search(query string) []int {
    var matches []int
    for i := 0; i < l.dataSource.Count(); i++ {
        data := l.dataSource.Get(i)
        if strings.Contains(fmt.Sprintf("%v", data), query) {
            matches = append(matches, i)
        }
    }
    return matches
}

// NextMatch 下一个匹配项
func (l *SearchableList) NextMatch() {
    if len(l.matches) == 0 {
        return
    }

    l.currentMatch = (l.currentMatch + 1) % len(l.matches)
    l.ScrollTo(l.matches[l.currentMatch])
}
```

### 分页加载

```go
// PagedList 分页虚拟列表
type PagedList struct {
    *VirtualList

    // 分页配置
    pageSize   int
    currentPage int
    totalPages int

    // 加载状态
    loading map[int]bool
}

func NewPagedList(fetcher func(page int) ([]interface{}, error), pageSize int) *PagedList {
    dataSource := &PagedDataSource{
        fetcher: fetcher,
        cache:   make(map[int][]interface{}),
        loading: make(map[int]bool),
    }

    list := NewVirtualList(dataSource, renderItemFunc)

    return &PagedList{
        VirtualList: list,
        pageSize:    pageSize,
        loading:     dataSource.loading,
    }
}

func (p *PagedList) LoadPage(page int) error {
    dataSource := p.dataSource.(*PagedDataSource)

    // 标记加载中
    dataSource.loading[page] = true

    items, err := dataSource.fetcher(page)
    if err != nil {
        dataSource.loading[page] = false
        return err
    }

    dataSource.cache[page] = items
    dataSource.loading[page] = false

    p.currentPage = page
    p.MarkDirty()

    return nil
}
```

## 性能优化

### 位置缓存

```go
// 位于: tui/framework/component/virtuallist_cache.go

package component

// PositionCache 位置缓存
type PositionCache struct {
    positions []int
    total     int
    dirty     bool
}

// Build 构建缓存
func (c *PositionCache) Build(dataSource DataSource) {
    c.total = 0
    c.positions = make([]int, dataSource.Count())

    for i := 0; i < dataSource.Count(); i++ {
        c.positions[i] = c.total
        c.total += dataSource.Height(i)
    }

    c.dirty = false
}

// Invalidate 使缓存失效
func (c *PositionCache) Invalidate() {
    c.dirty = true
}

// GetPosition 获取位置
func (c *PositionCache) GetPosition(index int) int {
    if c.dirty {
        return -1 // 需要重建
    }
    if index < 0 || index >= len(c.positions) {
        return -1
    }
    return c.positions[index]
}
```

### 渲染缓冲区

```go
// RenderBuffer 渲染缓冲区
type RenderBuffer struct {
    items []*ListItem
    capacity int
}

// Acquire 获取缓冲区项
func (b *RenderBuffer) Acquire(index int) *ListItem {
    // 查找是否已缓存
    for _, item := range b.items {
        if item.Index == index {
            return item
        }
    }

    // 创建新项
    item := &ListItem{Index: index}
    b.items = append(b.items, item)

    // 限制缓冲区大小
    if len(b.items) > b.capacity {
        b.items = b.items[1:] // 移除最旧的
    }

    return item
}

// Update 更新缓冲区
func (b *RenderBuffer) Update(dataSource DataSource, viewport *Viewport) {
    // 只更新可见范围的项
    start, end := viewport.GetVisibleRange()

    for i := start; i < end; i++ {
        item := b.Acquire(i)
        item.Data = dataSource.Get(i)
        item.Height = dataSource.Height(i)
        item.OffsetY = i - start
    }
}
```

## 测试

```go
// 位于: tui/framework/component/virtuallist_test.go

func TestVirtualListPerformance(t *testing.T) {
    // 创建大数据集
    dataSource := &SimpleDataSource{
        items: make([]interface{}, 100000),
    }

    list := NewVirtualList(dataSource, renderItemFunc)
    list.SetBounds(Rect{Height: 20})

    // 测试渲染时间
    start := time.Now()
    list.Paint(context, buffer)
    duration := time.Since(start)

    // 应该在 10ms 内完成
    assert.Less(t, duration.Milliseconds(), 10)
}

func TestVirtualScroll(t *testing.T) {
    dataSource := &SimpleDataSource{
        items: make([]interface{}, 10000),
    }

    list := NewVirtualList(dataSource, renderItemFunc)
    list.SetBounds(Rect{Height: 20})

    // 滚动到中间
    list.ScrollTo(5000)

    // 验证视口
    start, end := list.viewport.GetVisibleRange()
    assert.InDelta(t, start, 5000-list.viewport.Count/2, 5)
}

func TestLazyDataSource(t *testing.T) {
    callCount := 0
    loader := func(page int) ([]interface{}, error) {
        callCount++
        return make([]interface{}, 100), nil
    }

    dataSource := &LazyDataSource{
        loader: loader,
    }

    // 只请求第一页
    _ = dataSource.Get(50)

    assert.Equal(t, 1, callCount)
}
```

## 总结

虚拟滚动系统提供：

1. **高性能**: O(1) 渲染复杂度
2. **可扩展**: DataSource 接口支持各种数据源
3. **灵活**: 支持等高列表、搜索高亮、分页加载
4. **内存友好**: 固定内存占用，与数据量无关
5. **平滑滚动**: 支持像素级平滑滚动

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [THEME_SYSTEM.md](./THEME_SYSTEM.md) - 主题系统
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
