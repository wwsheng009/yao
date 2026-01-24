package component

import (
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Component 组件接口
type Component interface {
	// 标识
	ID() string
	SetID(id string)
	GetType() string

	// 生命周期
	Mount(parent Component)
	Unmount()
	IsMounted() bool

	// 尺寸
	SetSize(width, height int)
	GetSize() (width, height int)
	GetPreferredSize() (width, height int)
	GetMinSize() (width, height int)
	GetMaxSize() (width, height int)

	// 渲染
	Render(ctx *RenderContext) string

	// 事件
	HandleEvent(ev event.Event) bool
	SetEventHandler(handler event.EventHandler)
	GetEventHandler() event.EventHandler

	// 状态
	SetVisible(visible bool)
	IsVisible() bool
	SetEnabled(enabled bool)
	IsEnabled() bool

	// 样式
	SetStyle(s style.Style)
	GetStyle() style.Style
}

// RenderContext 渲染上下文
type RenderContext struct {
	// 可用尺寸
	AvailableWidth  int
	AvailableHeight int

	// 位置 (相对于父组件)
	X int
	Y int

	// 滚动偏移
	OffsetX int
	OffsetY int

	// 继承的样式
	InheritStyle style.Style

	// Z-index
	ZIndex int

	// 裁剪区域
	ClipRect *Rect
}

// Rect 矩形区域
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewRenderContext 创建渲染上下文
func NewRenderContext(width, height int) *RenderContext {
	return &RenderContext{
		AvailableWidth:  width,
		AvailableHeight: height,
	}
}

// WithOffset 创建带偏移的上下文
func (c *RenderContext) WithOffset(dx, dy int) *RenderContext {
	return &RenderContext{
		AvailableWidth:  c.AvailableWidth,
		AvailableHeight: c.AvailableHeight,
		X:               c.X + dx,
		Y:               c.Y + dy,
		OffsetX:         c.OffsetX,
		OffsetY:         c.OffsetY,
		InheritStyle:    c.InheritStyle,
		ZIndex:          c.ZIndex,
		ClipRect:        c.ClipRect,
	}
}

// WithClip 创建带裁剪的上下文
func (c *RenderContext) WithClip(rect *Rect) *RenderContext {
	clip := rect
	if c.ClipRect != nil {
		clip = c.ClipRect.Intersect(rect)
	}
	return &RenderContext{
		AvailableWidth:  c.AvailableWidth,
		AvailableHeight: c.AvailableHeight,
		X:               c.X,
		Y:               c.Y,
		OffsetX:         c.OffsetX,
		OffsetY:         c.OffsetY,
		InheritStyle:    c.InheritStyle,
		ZIndex:          c.ZIndex,
		ClipRect:        clip,
	}
}

// Contains 检查点是否在矩形内
func (r *Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

// Intersect 计算两个矩形的交集
func (r *Rect) Intersect(other *Rect) *Rect {
	if r == nil {
		return other
	}
	if other == nil {
		return r
	}

	x1 := maxInt(r.X, other.X)
	y1 := maxInt(r.Y, other.Y)
	x2 := minInt(r.X+r.Width, other.X+other.Width)
	y2 := minInt(r.Y+r.Height, other.Y+other.Height)

	if x1 >= x2 || y1 >= y2 {
		return nil
	}

	return &Rect{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
