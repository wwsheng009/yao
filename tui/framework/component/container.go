package component

import (
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Container 容器接口 (V3: 使用 Node 而非 Component)
// V2 兼容：也实现了 Component 接口
type Container interface {
	Component // V2 兼容

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

// BaseContainer 基础容器实现 (V3: 使用 Node 类型)
type BaseContainer struct {
	*BaseComponent
	children []Node
	layout   Layout

	// V2 兼容字段
	eventHandler event.EventHandler
	v2Style      style.Style
}

// NewBaseContainer 创建基础容器
func NewBaseContainer(typ string) *BaseContainer {
	return &BaseContainer{
		BaseComponent: NewBaseComponent(typ),
		children:     make([]Node, 0),
	}
}

// Add 添加子组件 (V3: 使用 Node 类型)
func (c *BaseContainer) Add(child Node) {
	c.children = append(c.children, child)
	if mountable, ok := child.(Mountable); ok {
		mountable.Mount(c)
	}
}

// Remove 移除子组件 (V3: 使用 Node 类型)
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

// GetChildren 获取子组件列表 (V3: 返回 Node 类型)
func (c *BaseContainer) GetChildren() []Node {
	return c.children
}

// GetChild 获取指定位置的子组件 (V3: 返回 Node 类型)
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

// SetLayout 设置布局
func (c *BaseContainer) SetLayout(layout Layout) {
	c.layout = layout
}

// GetLayout 获取布局
func (c *BaseContainer) GetLayout() Layout {
	return c.layout
}

// Render 渲染容器 (V2 兼容)
func (c *BaseContainer) Render(ctx *RenderContext) string {
	// 由具体布局实现
	return ""
}

// ============================================================================
// V2 兼容方法
// ============================================================================

// Mount 挂载到父容器 (V2 兼容：接受 Component 参数)
// V3 的 Mount(Container) 由 BaseComponent 提供
func (c *BaseContainer) Mount(parent Component) {
	// 将 Component 转换为 Container（如果可能）
	if container, ok := parent.(Container); ok {
		c.BaseComponent.Mount(container)
	}
}

// GetType 返回组件类型 (V2 兼容)
func (c *BaseContainer) GetType() string {
	return c.Type()
}

// GetPreferredSize 获取首选尺寸 (V2 兼容)
func (c *BaseContainer) GetPreferredSize() (width, height int) {
	return c.Measure(1000, 1000)
}

// GetMinSize 获取最小尺寸 (V2 兼容)
func (c *BaseContainer) GetMinSize() (width, height int) {
	return 0, 0
}

// GetMaxSize 获取最大尺寸 (V2 兼容)
func (c *BaseContainer) GetMaxSize() (width, height int) {
	return 1000, 1000
}

// HandleEvent 处理事件 (V2 兼容)
func (c *BaseContainer) HandleEvent(ev event.Event) bool {
	if c.eventHandler != nil {
		return c.eventHandler.HandleEvent(ev)
	}
	return false
}

// SetEventHandler 设置事件处理器 (V2 兼容)
func (c *BaseContainer) SetEventHandler(handler event.EventHandler) {
	c.eventHandler = handler
}

// GetEventHandler 获取事件处理器 (V2 兼容)
func (c *BaseContainer) GetEventHandler() event.EventHandler {
	return c.eventHandler
}

// SetEnabled 设置启用状态 (V2 兼容)
func (c *BaseContainer) SetEnabled(enabled bool) {
	c.SetDisabled(!enabled)
}

// IsEnabled 检查是否启用 (V2 兼容)
func (c *BaseContainer) IsEnabled() bool {
	return !c.IsDisabled()
}

// SetStyle 设置样式 (V2 兼容)
func (c *BaseContainer) SetStyle(s style.Style) {
	c.v2Style = s
}

// GetStyle 获取样式 (V2 兼容)
func (c *BaseContainer) GetStyle() style.Style {
	return c.v2Style
}
