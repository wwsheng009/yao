package layout

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ==============================================================================
// Layout Cache (V3)
// ==============================================================================
// 布局结果缓存，避免重复计算

// cacheKey 缓存键
type cacheKey struct {
	nodesHash string
	constraintsKey string
}

// Cache 布局缓存
type Cache struct {
	entries map[string]*CachedLayout
	maxSize int
}

// CachedLayout 缓存的布局结果
type CachedLayout struct {
	Result     *LayoutResult
	Timestamp  time.Time
	HitCount   int
}

// Get 获取缓存
func (c *Cache) Get(nodes []Node, constraints Constraints) *LayoutResult {
	key := c.makeKey(nodes, constraints)
	if entry, ok := c.entries[key]; ok {
		entry.HitCount++
		// 返回克隆避免修改缓存
		return c.cloneResult(entry.Result)
	}
	return nil
}

// Put 存入缓存
func (c *Cache) Put(nodes []Node, constraints Constraints, result *LayoutResult) {
	key := c.makeKey(nodes, constraints)

	// 如果缓存已满，删除最旧的条目
	if len(c.entries) >= c.maxSize {
		c.evict()
	}

	c.entries[key] = &CachedLayout{
		Result:     result,
		Timestamp:  time.Now(),
		HitCount:   0,
	}
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.entries = make(map[string]*CachedLayout)
}

// RemoveByNode 删除特定节点的缓存
func (c *Cache) RemoveByNode(id string) {
	// 需要检查每个缓存条目是否包含该节点
	// 简化实现：清空所有缓存
	c.Clear()
}

// evict 驱逐最旧的条目
func (c *Cache) evict() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// makeKey 生成缓存键
func (c *Cache) makeKey(nodes []Node, constraints Constraints) string {
	// 简化实现：基于约束生成键
	// 实际应该基于节点树结构
	constraintKey := c.constraintsKey(constraints)
	nodesHash := c.nodesHash(nodes)

	return constraintKey + ":" + nodesHash
}

// constraintsKey 约束键
func (c *Cache) constraintsKey(constraints Constraints) string {
	return string(rune(constraints.MinWidth)) + "," +
		string(rune(constraints.MaxWidth)) + "," +
		string(rune(constraints.MinHeight)) + "," +
		string(rune(constraints.MaxHeight))
}

// nodesHash 节点哈希
func (c *Cache) nodesHash(nodes []Node) string {
	h := sha256.New()
	for _, node := range nodes {
		h.Write([]byte(node.ID()))
		h.Write([]byte(node.Type()))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// cloneResult 克隆布局结果
func (c *Cache) cloneResult(result *LayoutResult) *LayoutResult {
	if result == nil {
		return nil
	}
	clone := &LayoutResult{
		ContentSize: result.ContentSize,
		Dirty:       result.Dirty,
	}

	if result.Root != nil {
		clone.Root = c.cloneBox(result.Root)
	}

	clone.Boxes = make([]LayoutBox, len(result.Boxes))
	copy(clone.Boxes, result.Boxes)

	return clone
}

// cloneBox 克隆布局盒子
func (c *Cache) cloneBox(box *LayoutBox) *LayoutBox {
	if box == nil {
		return nil
	}
	clone := &LayoutBox{
		ID:       box.ID,
		X:        box.X,
		Y:        box.Y,
		Width:    box.Width,
		Height:   box.Height,
		Baseline: box.Baseline,
	}

	if len(box.Children) > 0 {
		clone.Children = make([]*LayoutBox, len(box.Children))
		for i, child := range box.Children {
			clone.Children[i] = c.cloneBox(child)
		}
	}

	return clone
}
