package cursor

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

var (
	debugCursor = os.Getenv("TUI_CURSOR_DEBUG") == "1"
)

// cursorDebugLog 调试日志输出
func cursorDebugLog(format string, args ...interface{}) {
	if debugCursor {
		timestamp := time.Now().Format("15:04:05.000")
		fullFormat := fmt.Sprintf("[%s] [Cursor] %s\n", timestamp, format)
		fmt.Fprintf(os.Stderr, fullFormat, args...)
		// fmt.Printf(fullFormat, args...)
	}
}

// ==============================================================================
// Cursor - 原子光标组件
// ==============================================================================
// Cursor 是一个独立、自包含的光标组件，具备以下特性：
// 1. 原子性：完全自我管理，不依赖外部状态
// 2. 自我绘制：实现 Paintable 接口，自己负责渲染
// 3. 自我闪烁：内部管理闪烁计时，不需要外部 Tick
// 4. 松耦合：通过接口与其他组件交互

// Cursor 光标组件
type Cursor struct {
	*component.BaseComponent

	mu sync.RWMutex

	// 位置状态
	x      int  // 光标在父组件中的相对 X 坐标
	y      int  // 光标在父组件中的相对 Y 坐标
	visible bool // 光标当前是否可见（用于闪烁）

	// 闪烁状态
	blinkEnabled bool          // 是否启用闪烁
	blinkInterval time.Duration // 闪烁间隔
	lastBlinkTime time.Time    // 上次闪烁时间

	// 样式
	style style.Style // 光标样式（前景色、背景色等）

	// 形状
	shape Shape // 光标形状
}

// Shape 光标形状
type Shape int

const (
	// ShapeBlock 块状光标（覆盖整个字符）
	ShapeBlock Shape = iota
	// ShapeUnderline 下划线光标
	ShapeUnderline
	// ShapeBar 竖线光标
	ShapeBar
)

// NewCursor 创建新的光标组件
func NewCursor() *Cursor {
	return &Cursor{
		BaseComponent:  component.NewBaseComponent("cursor"),
		x:              0,
		y:              0,
		visible:        true,
		blinkEnabled:   true,
		blinkInterval:  500 * time.Millisecond,
		lastBlinkTime:  time.Now(),
		style:          style.Style{}.Reverse(true),
		shape:          ShapeBlock,
	}
}

// ==============================================================================
// 位置管理
// ==============================================================================

// SetPosition 设置光标位置（相对于父组件）
func (c *Cursor) SetPosition(x, y int) *Cursor {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.x = x
	c.y = y
	return c
}

// GetPosition 获取光标位置
func (c *Cursor) GetPosition() (x, y int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.x, c.y
}

// ==============================================================================
// 闪烁管理
// ==============================================================================

// SetBlinkEnabled 设置是否启用闪烁
func (c *Cursor) SetBlinkEnabled(enabled bool) *Cursor {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blinkEnabled = enabled
	if enabled {
		c.lastBlinkTime = time.Now()
		c.visible = true
	}
	return c
}

// IsBlinkEnabled 检查是否启用闪烁
func (c *Cursor) IsBlinkEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blinkEnabled
}

// SetBlinkInterval 设置闪烁间隔
func (c *Cursor) SetBlinkInterval(interval time.Duration) *Cursor {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blinkInterval = interval
	return c
}

// IsVisible 检查光标当前是否可见
// 此方法会在每次绘制时被调用，以更新闪烁状态
func (c *Cursor) IsVisible() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果禁用闪烁，始终可见
	if !c.blinkEnabled {
		return true
	}

	// 检查是否需要切换闪烁状态
	now := time.Now()
	if now.Sub(c.lastBlinkTime) >= c.blinkInterval {
		c.visible = !c.visible
		c.lastBlinkTime = now
	}
	
	return c.visible
}

// ResetBlink 重置闪烁状态（使光标立即可见）
func (c *Cursor) ResetBlink() *Cursor {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastBlinkTime = time.Now()
	c.visible = true
	return c
}

// ==============================================================================
// 样式管理
// ==============================================================================

// SetStyle 设置光标样式
func (c *Cursor) SetStyle(s style.Style) *Cursor {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.style = s
	return c
}

// GetStyle 获取光标样式
func (c *Cursor) GetStyle() style.Style {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.style
}

// SetShape 设置光标形状
func (c *Cursor) SetShape(shape Shape) *Cursor {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shape = shape
	return c
}

// GetShape 获取光标形状
func (c *Cursor) GetShape() Shape {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shape
}

// ==============================================================================
// Paintable 接口实现
// ==============================================================================

// Paint 绘制光标到缓冲区
// ctx.X, ctx.Y 是父组件的起始位置
// 光标绘制在 ctx.X + c.x, ctx.Y + c.y
func (c *Cursor) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	// 加锁保护所有字段的读取
	c.mu.RLock()
	x := ctx.X + c.x
	y := ctx.Y + c.y
	cursorStyle := c.style
	shape := c.shape

	// 检查是否需要更新闪烁状态（不修改状态，只读取）
	// 如果启用闪烁，检查当前是否在"显示"周期内
	visible := c.visible
	if c.blinkEnabled {
		now := time.Now()
		// 计算当前周期
		elapsed := now.Sub(c.lastBlinkTime)
		// 在闪烁周期内，前半段显示，后半段隐藏
		period := elapsed / c.blinkInterval
		visible = (period % 2) == 0
	}
	c.mu.RUnlock()

	// 如果光标不可见，直接返回（不修改缓冲区）
	if !visible {
		cursorDebugLog("CURSOR HIDDEN: absolute=(%d,%d)", x, y)
		return
	}

	// 边界检查
	if x < 0 || x >= buf.Width || y < 0 || y >= buf.Height {
		cursorDebugLog("FOCUS ERROR: cursor out of bounds at (%d,%d), buffer is %dx%d",
			x, y, buf.Width, buf.Height)
		return
	}

	// 记录焦点状态下的光标位置
	cursorDebugLog("FOCUS: absolute=(%d,%d), ctx=(%d,%d), visible=%v, shape=%d",
		x, y, ctx.X, ctx.Y, visible, shape)

	// 根据形状绘制光标
	switch shape {
	case ShapeBlock:
		// 块状光标：反转当前单元格的样式
		// 读取当前单元格的原始样式
		cell := buf.Cells[y][x]

		// 在原始样式的基础上应用反转
		// 保留原始样式的所有属性，只添加反转效果
		baseStyle := cell.Style

		// 应用反转效果到基础样式
		// 如果样式未设置，使用默认反转样式
		reverseStyle := baseStyle
		if baseStyle == (style.Style{}) {
			// 没有样式，使用光标默认样式
			reverseStyle = cursorStyle.Reverse(true)
		} else {
			// 在现有样式基础上反转
			reverseStyle = baseStyle.Reverse(true)
		}

		cursorDebugLog("FOCUS RENDER: drew cursor at (%d,%d) on char '%c', baseStyle=%v, reverseStyle.IsReverse=%v",
			x, y, cell.Char, baseStyle.IsReverse(), reverseStyle.IsReverse())

		buf.SetCell(x, y, cell.Char, reverseStyle)
	case ShapeUnderline:
		buf.SetCell(x, y, '_', cursorStyle)
	case ShapeBar:
		buf.SetCell(x, y, '|', cursorStyle)
	}
}

// ==============================================================================
// 焦点管理
// ==============================================================================

// FocusID 返回焦点标识符
func (c *Cursor) FocusID() string {
	return c.ID()
}

// OnFocus 获得焦点时启用闪烁
func (c *Cursor) OnFocus() {
	c.SetBlinkEnabled(true)
	c.BaseComponent.OnFocus()
}

// OnBlur 失去焦点时禁用闪烁（隐藏光标）
func (c *Cursor) OnBlur() {
	c.SetBlinkEnabled(false)
	c.BaseComponent.OnBlur()
}

// ==============================================================================
// 辅助接口 - 用于其他组件嵌入光标
// ==============================================================================

// Host 需要嵌入光标的组件接口
// 任何组件可以实现此接口来提供光标绘制能力
type Host interface {
	// GetCursor 获取光标组件
	GetCursor() *Cursor
}

// PaintCursor 在指定位置绘制光标
// 这是一个便捷函数，用于在组件的 Paint 方法中嵌入光标绘制
func PaintCursor(host Host, ctx component.PaintContext, buf *paint.Buffer, x, y int) {
	cursor := host.GetCursor()
	if cursor == nil {
		return
	}

	// 更新光标位置并绘制
	cursor.SetPosition(x, y)
	cursor.Paint(ctx, buf)
}
