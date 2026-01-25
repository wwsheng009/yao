package layout

// ==============================================================================
// Layout Engine (V3)
// ==============================================================================
// Runtime 的布局引擎是纯内核，无外部依赖
// 负责：计算布局，不处理渲染、事件、焦点

// Engine 布局引擎
type Engine struct {
	// 缓存
	cache *Cache
	// 统计
	stats *Stats
}

// Stats 布局统计
type Stats struct {
	TotalLayouts    int64
	CacheHits       int64
	CacheMisses     int64
	DirtyOptimizations int64
}

// NewEngine 创建布局引擎
func NewEngine() *Engine {
	return &Engine{
		cache: &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize: 1000,
		},
		stats: &Stats{},
	}
}

// Layout 计算布局
// 输入：节点树和约束
// 输出：布局结果（每个节点的位置和尺寸）
func (e *Engine) Layout(nodes []Node, constraints Constraints) *LayoutResult {
	// 1. 检查缓存
	if result := e.cache.Get(nodes, constraints); result != nil {
		e.stats.CacheHits++
		e.stats.TotalLayouts++
		return result
	}

	e.stats.CacheMisses++
	e.stats.TotalLayouts++

	// 2. 执行布局
	result := e.computeLayout(nodes, constraints)

	// 3. 缓存结果
	e.cache.Put(nodes, constraints, result)

	return result
}

// computeLayout 执行实际布局计算
func (e *Engine) computeLayout(nodes []Node, constraints Constraints) *LayoutResult {
	result := &LayoutResult{
		Boxes: make([]LayoutBox, 0, len(nodes)),
	}

	// Phase 1: Measure - 测量每个节点的理想尺寸
	e.measureNodes(nodes, constraints)

	// Phase 2: Layout - 计算每个节点的实际位置和尺寸
	e.layoutNodes(nodes, constraints)

	// Phase 3: 收集结果
	for _, node := range nodes {
		box := e.collectBox(node, 0, 0)
		if box != nil {
			result.Boxes = append(result.Boxes, *box)
		}
	}

	return result
}

// measureNodes 测量节点
func (e *Engine) measureNodes(nodes []Node, constraints Constraints) {
	for _, node := range nodes {
		if measurable, ok := node.(Measurable); ok {
			_ = measurable.Measure(constraints)
			// 测量结果会缓存到节点内部
		}
	}
}

// layoutNodes 布局节点
func (e *Engine) layoutNodes(nodes []Node, constraints Constraints) {
	// 根据 Flexbox 或其他布局算法布局
	// 这里简化为垂直排列
	y := 0
	for _, node := range nodes {
		node.SetPosition(0, y)
		height := node.GetHeight()
		y += height
	}
}

// collectBox 收集布局结果
func (e *Engine) collectBox(node Node, x, y int) *LayoutBox {
	return &LayoutBox{
		ID:     node.ID(),
		X:      x,
		Y:      y,
		Width:  node.GetWidth(),
		Height: node.GetHeight(),
	}
}

// Measure 测量节点
func (e *Engine) Measure(node Node, constraints Constraints) Size {
	if measurable, ok := node.(Measurable); ok {
		return measurable.Measure(constraints)
	}
	return Size{Width: constraints.MinWidth, Height: constraints.MinHeight}
}

// Invalidate 作废缓存
func (e *Engine) Invalidate() {
	e.cache.Clear()
}

// InvalidateNode 作废特定节点的缓存
func (e *Engine) InvalidateNode(id string) {
	e.cache.RemoveByNode(id)
}

// GetStats 获取统计信息
func (e *Engine) GetStats() *Stats {
	return e.stats
}

// ClearStats 清空统计信息
func (e *Engine) ClearStats() {
	e.stats = &Stats{}
}
