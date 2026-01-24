package component

import (
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
)

// BaseComponent 组件基础实现
type BaseComponent struct {
	id       string
	comptype string

	// 尺寸
	width    int
	height   int
	minW     int
	minH     int
	maxW     int
	maxH     int

	// 状态
	mounted  bool
	visible  bool
	enabled  bool

	// 样式
	style    style.Style

	// 事件
	handler  event.EventHandler

	// 父组件
	parent   Component
}

// NewBaseComponent 创建基础组件
func NewBaseComponent(typ string) *BaseComponent {
	return &BaseComponent{
		comptype: typ,
		visible:  true,
		enabled:  true,
		minW:     0,
		minH:     0,
		maxW:     -1, // 无限制
		maxH:     -1,
	}
}

// ID 获取组件 ID
func (c *BaseComponent) ID() string {
	return c.id
}

// SetID 设置组件 ID
func (c *BaseComponent) SetID(id string) {
	c.id = id
}

// GetType 获取组件类型
func (c *BaseComponent) GetType() string {
	return c.comptype
}

// Mount 挂载组件
func (c *BaseComponent) Mount(parent Component) {
	c.parent = parent
	c.mounted = true
}

// Unmount 卸载组件
func (c *BaseComponent) Unmount() {
	c.mounted = false
}

// IsMounted 检查是否已挂载
func (c *BaseComponent) IsMounted() bool {
	return c.mounted
}

// SetSize 设置尺寸
func (c *BaseComponent) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// GetSize 获取尺寸
func (c *BaseComponent) GetSize() (width, height int) {
	return c.width, c.height
}

// GetPreferredSize 获取首选尺寸
func (c *BaseComponent) GetPreferredSize() (width, height int) {
	return c.width, c.height
}

// GetMinSize 获取最小尺寸
func (c *BaseComponent) GetMinSize() (width, height int) {
	return c.minW, c.minH
}

// SetMinSize 设置最小尺寸
func (c *BaseComponent) SetMinSize(width, height int) {
	c.minW = width
	c.minH = height
}

// GetMaxSize 获取最大尺寸
func (c *BaseComponent) GetMaxSize() (width, height int) {
	if c.maxW < 0 {
		width = -1
	} else {
		width = c.maxW
	}
	if c.maxH < 0 {
		height = -1
	} else {
		height = c.maxH
	}
	return
}

// SetMaxSize 设置最大尺寸
func (c *BaseComponent) SetMaxSize(width, height int) {
	c.maxW = width
	c.maxH = height
}

// Render 渲染组件 (子类实现)
func (c *BaseComponent) Render(ctx *RenderContext) string {
	return ""
}

// HandleEvent 处理事件
func (c *BaseComponent) HandleEvent(ev event.Event) bool {
	if c.handler != nil {
		return c.handler.HandleEvent(ev)
	}
	return false
}

// SetEventHandler 设置事件处理器
func (c *BaseComponent) SetEventHandler(handler event.EventHandler) {
	c.handler = handler
}

// GetEventHandler 获取事件处理器
func (c *BaseComponent) GetEventHandler() event.EventHandler {
	return c.handler
}

// SetVisible 设置可见性
func (c *BaseComponent) SetVisible(visible bool) {
	c.visible = visible
}

// IsVisible 检查是否可见
func (c *BaseComponent) IsVisible() bool {
	return c.visible
}

// SetEnabled 设置启用状态
func (c *BaseComponent) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// IsEnabled 检查是否启用
func (c *BaseComponent) IsEnabled() bool {
	return c.enabled
}

// SetStyle 设置样式
func (c *BaseComponent) SetStyle(s style.Style) {
	c.style = s
}

// GetStyle 获取样式
func (c *BaseComponent) GetStyle() style.Style {
	return c.style
}

// GetParent 获取父组件
func (c *BaseComponent) GetParent() Component {
	return c.parent
}
