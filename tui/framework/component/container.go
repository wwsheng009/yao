package component

// Container 容器接口
type Container interface {
	Component

	// 子组件管理
	Add(child Component)
	Remove(child Component)
	RemoveAt(index int)
	GetChildren() []Component
	GetChild(index int) Component
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

// BaseContainer 基础容器实现
type BaseContainer struct {
	*BaseComponent
	children []Component
	layout   Layout
}

// NewBaseContainer 创建基础容器
func NewBaseContainer(typ string) *BaseContainer {
	return &BaseContainer{
		BaseComponent: NewBaseComponent(typ),
		children:     make([]Component, 0),
	}
}

// Add 添加子组件
func (c *BaseContainer) Add(child Component) {
	c.children = append(c.children, child)
	child.Mount(c)
}

// Remove 移除子组件
func (c *BaseContainer) Remove(child Component) {
	for i, ch := range c.children {
		if ch == child {
			c.children = append(c.children[:i], c.children[i+1:]...)
			child.Unmount()
			break
		}
	}
}

// RemoveAt 移除指定位置的子组件
func (c *BaseContainer) RemoveAt(index int) {
	if index >= 0 && index < len(c.children) {
		child := c.children[index]
		c.children = append(c.children[:index], c.children[index+1:]...)
		child.Unmount()
	}
}

// GetChildren 获取子组件列表
func (c *BaseContainer) GetChildren() []Component {
	return c.children
}

// GetChild 获取指定位置的子组件
func (c *BaseContainer) GetChild(index int) Component {
	if index >= 0 && index < len(c.children) {
		return c.children[index]
	}
	return nil
}

// ChildCount 获取子组件数量
func (c *BaseContainer) ChildCount() int {
	return len(c.children)
}

// SetLayout 设置布局
func (c *BaseContainer) SetLayout(layout Layout) {
	c.layout = layout
}

// GetLayout 获取布局
func (c *BaseContainer) GetLayout() Layout {
	return c.layout
}

// Render 渲染容器
func (c *BaseContainer) Render(ctx *RenderContext) string {
	// 由具体布局实现
	return ""
}
