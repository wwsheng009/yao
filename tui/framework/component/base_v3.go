package component

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Base Component V3
// ==============================================================================
// V3 基础组件实现，遵循 Capability Interfaces 模式

// BaseComponentV3 V3 基础组件
// 提供所有能力的默认实现，组件可以按需组合
type BaseComponentV3 struct {
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
}

// NewBaseComponentV3 创建 V3 基础组件
func NewBaseComponentV3(typ string) *BaseComponentV3 {
	return &BaseComponentV3{
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
func (c *BaseComponentV3) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.id == "" {
		return c.typ
	}
	return c.id
}

// SetID 设置组件 ID
func (c *BaseComponentV3) SetID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.id = id
}

// Type 返回组件类型
func (c *BaseComponentV3) Type() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.typ
}

// Children 返回子节点（基础组件无子节点）
func (c *BaseComponentV3) Children() []Node {
	return nil
}

// GetPosition 获取位置
func (c *BaseComponentV3) GetPosition() (x, y int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.x, c.y
}

// SetPosition 设置位置
func (c *BaseComponentV3) SetPosition(x, y int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.x = x
	c.y = y
}

// GetSize 获取尺寸
func (c *BaseComponentV3) GetSize() (width, height int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.width, c.height
}

// SetSize 设置尺寸
func (c *BaseComponentV3) SetSize(width, height int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.width = width
	c.height = height
}

// GetWidth 获取宽度
func (c *BaseComponentV3) GetWidth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.width
}

// GetHeight 获取高度
func (c *BaseComponentV3) GetHeight() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.height
}

// ============================================================================
// Mountable 接口实现
// ============================================================================

// Mount 挂载到父容器
func (c *BaseComponentV3) Mount(parent Container) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parent = parent
}

// Unmount 从父容器卸载
func (c *BaseComponentV3) Unmount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parent = nil
}

// IsMounted 检查是否已挂载
func (c *BaseComponentV3) IsMounted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parent != nil
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (c *BaseComponentV3) Measure(maxWidth, maxHeight int) (width, height int) {
	// 默认实现：返回最小尺寸
	return 0, 0
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件
// 默认实现：什么都不绘制
func (c *BaseComponentV3) Paint(ctx PaintContext, buf *paint.Buffer) {
	// 默认不绘制任何内容
	// 子类可以覆盖此方法
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (c *BaseComponentV3) FocusID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.focusID == "" {
		return c.id
	}
	return c.focusID
}

// SetFocusID 设置焦点标识符
func (c *BaseComponentV3) SetFocusID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focusID = id
}

// OnFocus 获得焦点时调用
func (c *BaseComponentV3) OnFocus() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focused = true
}

// OnBlur 失去焦点时调用
func (c *BaseComponentV3) OnBlur() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focused = false
}

// IsFocused 检查是否有焦点
func (c *BaseComponentV3) IsFocused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.focused
}

// ============================================================================
// 状态管理
// ============================================================================

// SetVisible 设置可见性
func (c *BaseComponentV3) SetVisible(visible bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.visible = visible
}

// IsVisible 检查是否可见
func (c *BaseComponentV3) IsVisible() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.visible
}

// SetDisabled 设置禁用状态
func (c *BaseComponentV3) SetDisabled(disabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = disabled
}

// IsDisabled 检查是否禁用
func (c *BaseComponentV3) IsDisabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.disabled
}

// GetParent 获取父容器
func (c *BaseComponentV3) GetParent() Container {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parent
}
