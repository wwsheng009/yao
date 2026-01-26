package component

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Base Component V3
// ==============================================================================
// V3 基础组件实现，遵循 Capability Interfaces 模式

// BaseComponent V3 基础组件
// 提供所有能力的默认实现，组件可以按需组合
type BaseComponent struct {
	mu sync.RWMutex

	// 基础属性
	id   string
	typ  string

	// 布局位置和尺寸
	x      int
	y      int
	width  int
	height int

	// 可见性
	visible bool

	// 禁用状态
	disabled bool

	// 焦点
	focusID string
	focused bool

	// 父容器
	parent Container

	// 组件上下文（用于脏标记等运行时功能）
	context *ComponentContext
}

// NewBaseComponent 创建 V3 基础组件
func NewBaseComponent(typ string) *BaseComponent {
	return &BaseComponent{
		typ:      typ,
		visible:  true,
		disabled: false,
		focused:  false,
	}
}

// ============================================================================
// Node 接口实现
// ============================================================================

// ID 返回组件 ID
func (c *BaseComponent) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.id == "" {
		return c.typ
	}
	return c.id
}

// SetID 设置组件 ID
func (c *BaseComponent) SetID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.id = id
}

// Type 返回组件类型
func (c *BaseComponent) Type() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.typ
}

// Children 返回子节点（基础组件无子节点）
func (c *BaseComponent) Children() []Node {
	return nil
}

// GetPosition 获取位置
func (c *BaseComponent) GetPosition() (x, y int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.x, c.y
}

// SetPosition 设置位置
func (c *BaseComponent) SetPosition(x, y int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.x = x
	c.y = y
}

// GetSize 获取尺寸
func (c *BaseComponent) GetSize() (width, height int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.width, c.height
}

// SetSize 设置尺寸
func (c *BaseComponent) SetSize(width, height int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.width = width
	c.height = height
}

// GetWidth 获取宽度
func (c *BaseComponent) GetWidth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.width
}

// GetHeight 获取高度
func (c *BaseComponent) GetHeight() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.height
}

// ============================================================================
// Mountable 接口实现
// ============================================================================

// Mount 挂载到父容器
func (c *BaseComponent) Mount(parent Container) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parent = parent
}

// Unmount 从父容器卸载
func (c *BaseComponent) Unmount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parent = nil
	c.context = nil
}

// IsMounted 检查是否已挂载
func (c *BaseComponent) IsMounted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parent != nil
}

// MountWithContext 挂载到父容器并接收组件上下文
// 实现 MountableWithContext 接口
func (c *BaseComponent) MountWithContext(parent Container, ctx *ComponentContext) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parent = parent
	c.context = ctx
}

// GetComponentContext 获取组件上下文
func (c *BaseComponent) GetComponentContext() *ComponentContext {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.context
}

// MarkDirty 标记组件为脏状态，触发重新渲染
func (c *BaseComponent) MarkDirty() {
	c.mu.RLock()
	ctx := c.context
	c.mu.RUnlock()

	if ctx != nil {
		ctx.MarkDirty()
	}
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (c *BaseComponent) Measure(maxWidth, maxHeight int) (width, height int) {
	// 默认实现：返回最小尺寸
	return 0, 0
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件
// 默认实现：什么都不绘制
func (c *BaseComponent) Paint(ctx PaintContext, buf *paint.Buffer) {
	// 默认不绘制任何内容
	// 子类可以覆盖此方法
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (c *BaseComponent) FocusID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.focusID == "" {
		return c.id
	}
	return c.focusID
}

// SetFocusID 设置焦点标识符
func (c *BaseComponent) SetFocusID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focusID = id
}

// OnFocus 获得焦点时调用
func (c *BaseComponent) OnFocus() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focused = true
}

// OnBlur 失去焦点时调用
func (c *BaseComponent) OnBlur() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focused = false
}

// IsFocused 检查是否有焦点
func (c *BaseComponent) IsFocused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.focused
}

// ============================================================================
// 状态管理
// ============================================================================

// SetVisible 设置可见性
func (c *BaseComponent) SetVisible(visible bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.visible = visible
}

// IsVisible 检查是否可见
func (c *BaseComponent) IsVisible() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.visible
}

// SetDisabled 设置禁用状态
func (c *BaseComponent) SetDisabled(disabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = disabled
}

// IsDisabled 检查是否禁用
func (c *BaseComponent) IsDisabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.disabled
}

// GetParent 获取父容器
func (c *BaseComponent) GetParent() Container {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parent
}
