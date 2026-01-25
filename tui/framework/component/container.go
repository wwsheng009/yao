package component

// ==============================================================================
// Container (V3)
// ==============================================================================
// V3 容器接口，使用 Node 类型而非 Component

// Container 容器接口 (V3)
type Container interface {
	Node
	Mountable

	// 子组件管理 (V3: 使用 Node 类型)
	Add(child Node)
	Remove(child Node)
	RemoveAt(index int)
	GetChildren() []Node
	GetChild(index int) Node
	ChildCount() int

	// 布局
	SetLayout(layout Layout)
	GetLayout() Layout
}

// Layout 布局接口
type Layout interface {
	// 测量
	Measure(container Container, availableWidth, availableHeight int) (width, height int)

	// 布局
	Layout(container Container, x, y, width, height int)

	// 通知变更
	Invalidate()
}

// BaseContainer 基础容器实现 (V3)
type BaseContainer struct {
	*BaseComponent
	children []Node
	layout   Layout
}

// NewBaseContainer 创建基础容器
func NewBaseContainer(typ string) *BaseContainer {
	return &BaseContainer{
		BaseComponent: NewBaseComponent(typ),
		children:      make([]Node, 0),
	}
}

// ============================================================================
// 子组件管理 (V3: 使用 Node 类型)
// ============================================================================

// Add 添加子组件
func (c *BaseContainer) Add(child Node) {
	c.children = append(c.children, child)
	if mountable, ok := child.(Mountable); ok {
		mountable.Mount(c)
	}
}

// Remove 移除子组件
func (c *BaseContainer) Remove(child Node) {
	for i, ch := range c.children {
		if ch == child {
			c.children = append(c.children[:i], c.children[i+1:]...)
			if mountable, ok := child.(Mountable); ok {
				mountable.Unmount()
			}
			break
		}
	}
}

// RemoveAt 移除指定位置的子组件
func (c *BaseContainer) RemoveAt(index int) {
	if index >= 0 && index < len(c.children) {
		child := c.children[index]
		c.children = append(c.children[:index], c.children[index+1:]...)
		if mountable, ok := child.(Mountable); ok {
			mountable.Unmount()
		}
	}
}

// GetChildren 获取子组件列表
func (c *BaseContainer) GetChildren() []Node {
	return c.children
}

// GetChild 获取指定位置的子组件
func (c *BaseContainer) GetChild(index int) Node {
	if index >= 0 && index < len(c.children) {
		return c.children[index]
	}
	return nil
}

// ChildCount 获取子组件数量
func (c *BaseContainer) ChildCount() int {
	return len(c.children)
}

// ============================================================================
// 布局管理
// ============================================================================

// SetLayout 设置布局
func (c *BaseContainer) SetLayout(layout Layout) {
	c.layout = layout
}

// GetLayout 获取布局
func (c *BaseContainer) GetLayout() Layout {
	return c.layout
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (c *BaseContainer) Measure(maxWidth, maxHeight int) (width, height int) {
	if c.layout != nil {
		return c.layout.Measure(c, maxWidth, maxHeight)
	}
	// 默认：计算所有子组件的总尺寸
	for _, child := range c.children {
		if measurable, ok := child.(Measurable); ok {
			w, h := measurable.Measure(maxWidth, maxHeight)
			if w > width {
				width = w
			}
			height += h
		}
	}
	return width, height
}
