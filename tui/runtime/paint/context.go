package paint

import (
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// ==============================================================================
// Paint Context (V3)
// ==============================================================================
// PaintContext 提供组件绘制时需要的上下文信息

// PaintContext 绘制上下文
type PaintContext struct {
	// Buffer 绘制目标缓冲区
	Buffer *Buffer

	// Bounds 组件的边界（相对于父容器）
	Bounds Rect

	// ============================================================================
	// Compatibility Fields - Framework components access these directly
	// These are kept in sync with Bounds for backward compatibility.
	// ============================================================================
	// Deprecated: Use Bounds.X instead
	X int
	// Deprecated: Use Bounds.Y instead
	Y int
	// Deprecated: Use Bounds.Width instead
	AvailableWidth int
	// Deprecated: Use Bounds.Height instead
	AvailableHeight int

	// FocusPath 当前焦点路径
	FocusPath state.FocusPath

	// Focused 当前组件是否有焦点
	Focused bool

	// Disabled 当前组件是否禁用
	Disabled bool

	// ZIndex 当前组件的Z层级
	ZIndex int

	// DirtyTracker 脏区域跟踪器
	DirtyTracker *DirtyTracker

	// viewport 视口偏移（用于滚动）
	viewportX int
	viewportY int
}

// NewPaintContext 创建绘制上下文
func NewPaintContext(buf *Buffer, bounds Rect) *PaintContext {
	return &PaintContext{
		Buffer:         buf,
		Bounds:         bounds,
		X:              bounds.X,
		Y:              bounds.Y,
		AvailableWidth: bounds.Width,
		AvailableHeight: bounds.Height,
		FocusPath:      make(state.FocusPath, 0),
		Focused:        false,
		Disabled:       false,
		ZIndex:         0,
		DirtyTracker:   NewDirtyTracker(),
		viewportX:      0,
		viewportY:      0,
	}
}

// WithFocus 设置焦点状态
func (c *PaintContext) WithFocus(focused bool) *PaintContext {
	ctx := c.clone()
	ctx.Focused = focused
	return ctx
}

// WithDisabled 设置禁用状态
func (c *PaintContext) WithDisabled(disabled bool) *PaintContext {
	ctx := c.clone()
	ctx.Disabled = disabled
	return ctx
}

// WithZIndex 设置Z层级
func (c *PaintContext) WithZIndex(zindex int) *PaintContext {
	ctx := c.clone()
	ctx.ZIndex = zindex
	return ctx
}

// WithBounds 设置边界
func (c *PaintContext) WithBounds(bounds Rect) *PaintContext {
	ctx := c.clone()
	ctx.Bounds = bounds
	ctx.X = bounds.X
	ctx.Y = bounds.Y
	ctx.AvailableWidth = bounds.Width
	ctx.AvailableHeight = bounds.Height
	return ctx
}

// WithViewport 设置视口偏移
func (c *PaintContext) WithViewport(x, y int) *PaintContext {
	ctx := c.clone()
	ctx.viewportX = x
	ctx.viewportY = y
	return ctx
}

// WithFocusPath 设置焦点路径
func (c *PaintContext) WithFocusPath(path state.FocusPath) *PaintContext {
	ctx := c.clone()
	ctx.FocusPath = path
	return ctx
}

// Child 创建子上下文
func (c *PaintContext) Child(id string, bounds Rect) *PaintContext {
	childPath := c.FocusPath.Clone()
	childPath = append(childPath, id)

	return &PaintContext{
		Buffer:          c.Buffer,
		Bounds:          bounds,
		X:               bounds.X,
		Y:               bounds.Y,
		AvailableWidth:  bounds.Width,
		AvailableHeight: bounds.Height,
		FocusPath:       childPath,
		Focused:         c.Focused && c.FocusPath.Current() == id,
		Disabled:        c.Disabled,
		ZIndex:          c.ZIndex,
		DirtyTracker:    c.DirtyTracker,
		viewportX:       c.viewportX,
		viewportY:       c.viewportY,
	}
}

// SetCell 在当前位置设置单元格
func (c *PaintContext) SetCell(x, y int, char rune, s style.Style) {
	// 防护：检查 Buffer 是否存在
	if c.Buffer == nil {
		return
	}

	// 应用视口偏移
	actualX := c.Bounds.X + x - c.viewportX
	actualY := c.Bounds.Y + y - c.viewportY

	// 边界检查
	if actualX < 0 || actualX >= c.Buffer.Width ||
		actualY < 0 || actualY >= c.Buffer.Height {
		return
	}

	c.Buffer.SetCell(actualX, actualY, char, s)

	// 标记脏区域
	if c.DirtyTracker != nil {
		c.DirtyTracker.MarkCell(actualX, actualY)
	}
}

// SetString 在当前位置写入字符串
func (c *PaintContext) SetString(x, y int, text string, s style.Style) {
	// 防护：检查 Buffer 是否存在
	if c.Buffer == nil {
		return
	}

	// 应用视口偏移
	actualX := c.Bounds.X + x - c.viewportX
	actualY := c.Bounds.Y + y - c.viewportY

	// 边界检查
	if actualY < 0 || actualY >= c.Buffer.Height {
		return
	}

	col := actualX
	for _, char := range text {
		if col < 0 || col >= c.Buffer.Width {
			break
		}
		c.Buffer.SetCell(col, actualY, char, s)

		// 标记脏区域
		if c.DirtyTracker != nil {
			c.DirtyTracker.MarkCell(col, actualY)
		}
		col++
	}
}

// Fill 填充矩形区域
func (c *PaintContext) Fill(rect Rect, char rune, s style.Style) {
	// 转换为绝对坐标
	absRect := Rect{
		X:      c.Bounds.X + rect.X - c.viewportX,
		Y:      c.Bounds.Y + rect.Y - c.viewportY,
		Width:  rect.Width,
		Height: rect.Height,
	}

	for y := absRect.Y; y < absRect.Y+absRect.Height; y++ {
		for x := absRect.X; x < absRect.X+absRect.Width; x++ {
			if x >= 0 && x < c.Buffer.Width && y >= 0 && y < c.Buffer.Height {
				c.Buffer.SetCell(x, y, char, s)
				if c.DirtyTracker != nil {
					c.DirtyTracker.MarkCell(x, y)
				}
			}
		}
	}
}

// DrawBox 绘制边框
func (c *PaintContext) DrawBox(rect Rect, boxStyle BoxStyle) {
	// 绘制水平边框
	if rect.Width > 2 {
		for x := 1; x < rect.Width-1; x++ {
			c.SetCell(x, 0, boxStyle.Horizontal, boxStyle.Style)
			c.SetCell(x, rect.Height-1, boxStyle.Horizontal, boxStyle.Style)
		}
	}

	// 绘制垂直边框
	if rect.Height > 2 {
		for y := 1; y < rect.Height-1; y++ {
			c.SetCell(0, y, boxStyle.Vertical, boxStyle.Style)
			c.SetCell(rect.Width-1, y, boxStyle.Vertical, boxStyle.Style)
		}
	}

	// 绘制角落
	if rect.Width > 0 && rect.Height > 0 {
		c.SetCell(0, 0, boxStyle.TopLeft, boxStyle.Style)
		c.SetCell(rect.Width-1, 0, boxStyle.TopRight, boxStyle.Style)
		c.SetCell(0, rect.Height-1, boxStyle.BottomLeft, boxStyle.Style)
		c.SetCell(rect.Width-1, rect.Height-1, boxStyle.BottomRight, boxStyle.Style)
	}
}

// DrawText 绘制文本，支持对齐
func (c *PaintContext) DrawText(x, y int, text string, align TextAlign, s style.Style) {
	// 计算可用宽度
	availableWidth := c.Bounds.Width - x
	if availableWidth <= 0 {
		return
	}

	// 截断文本
	text = c.truncateText(text, availableWidth)

	// 根据对齐方式计算起始位置
	var startX int
	switch align {
	case AlignLeft:
		startX = x
	case AlignCenter:
		startX = x + (availableWidth-len(text))/2
	case AlignRight:
		startX = x + availableWidth - len(text)
	}

	c.SetString(startX, y, text, s)
}

// truncateText 截断文本到指定宽度
func (c *PaintContext) truncateText(text string, maxWidth int) string {
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}

// Width 返回上下文宽度
func (c *PaintContext) Width() int {
	return c.Bounds.Width
}

// Height 返回上下文高度
func (c *PaintContext) Height() int {
	return c.Bounds.Height
}

// Contains 检查点是否在边界内
func (c *PaintContext) Contains(x, y int) bool {
	return x >= 0 && x < c.Bounds.Width &&
		y >= 0 && y < c.Bounds.Height
}

// Clamp 将坐标限制在边界内
func (c *PaintContext) Clamp(x, y int) (int, int) {
	if x < 0 {
		x = 0
	}
	if x >= c.Bounds.Width {
		x = c.Bounds.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= c.Bounds.Height {
		y = c.Bounds.Height - 1
	}
	return x, y
}

// clone 克隆上下文
func (c *PaintContext) clone() *PaintContext {
	return &PaintContext{
		Buffer:          c.Buffer,
		Bounds:          c.Bounds,
		X:               c.Bounds.X,
		Y:               c.Bounds.Y,
		AvailableWidth:  c.Bounds.Width,
		AvailableHeight: c.Bounds.Height,
		FocusPath:       c.FocusPath.Clone(),
		Focused:         c.Focused,
		Disabled:        c.Disabled,
		ZIndex:          c.ZIndex,
		DirtyTracker:    c.DirtyTracker,
		viewportX:       c.viewportX,
		viewportY:       c.viewportY,
	}
}

// BoxStyle 边框样式
type BoxStyle struct {
	// TopLeft 左上角字符
	TopLeft rune

	// TopRight 右上角字符
	TopRight rune

	// BottomLeft 左下角字符
	BottomLeft rune

	// BottomRight 右下角字符
	BottomRight rune

	// Horizontal 水平线字符
	Horizontal rune

	// Vertical 垂直线字符
	Vertical rune

	// Style 边框样式
	Style style.Style
}

// WithStyle returns a new BoxStyle with the given style applied.
func (b BoxStyle) WithStyle(s style.Style) BoxStyle {
	b.Style = s
	return b
}

// DefaultBoxStyle 默认边框样式
var DefaultBoxStyle = BoxStyle{
	TopLeft:     '┌',
	TopRight:    '┐',
	BottomLeft:  '└',
	BottomRight: '┘',
	Horizontal:  '─',
	Vertical:    '│',
	Style:       style.Style{},
}

// DoubleBoxStyle 双线边框样式
var DoubleBoxStyle = BoxStyle{
	TopLeft:     '╔',
	TopRight:    '╗',
	BottomLeft:  '╚',
	BottomRight: '╝',
	Horizontal:  '═',
	Vertical:    '║',
	Style:       style.Style{},
}

// RoundedBoxStyle 圆角边框样式
var RoundedBoxStyle = BoxStyle{
	TopLeft:     '╭',
	TopRight:    '╮',
	BottomLeft:  '╰',
	BottomRight: '╯',
	Horizontal:  '─',
	Vertical:    '│',
	Style:       style.Style{},
}

// TextAlign 文本对齐
type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)
