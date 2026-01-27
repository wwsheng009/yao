是的，你的直觉非常敏锐。如果在每一个组件的 `Paint` 方法中都直接使用 `buf.SetCell(x, y, rune, style)`，在开发复杂组件时确实会**非常痛苦**且**极易出错**。

主要问题包括：

1. **坐标计算繁琐**：你需要手动处理 `x + offset`，一旦涉及相对定位，代码难以维护。
2. **边界溢出 (Clipping)**：如果一个子组件画到了父组件给定的区域之外，会覆盖掉其他 UI 元素。
3. **重复代码**：绘制字符串、绘制边框、绘制背景色等逻辑在每个组件里都要写一遍循环。

为了解决这个问题，在 TUI 架构设计中，标准做法是引入一个 **"Painter" (绘制器) 抽象层**。

以下是具体的优化设计方案：

### 1. 引入 Painter 结构体

不要直接把 `*paint.Buffer` 传给组件，而是封装一个 `Painter` 对象。这个对象负责处理**坐标转换**和**剪裁**。

```go
// framework/paint/painter.go

type Painter struct {
    Buffer   *Buffer
    OriginX  int      // 相对原点 X
    OriginY  int      // 相对原点 Y
    ClipRect Rect     // 剪裁区域 (绝对坐标)
    Style    Style    // 当前默认样式
}

// NewPainter 创建一个新的绘制器
func NewPainter(buf *Buffer, ctx component.PaintContext) *Painter {
    return &Painter{
        Buffer:   buf,
        OriginX:  ctx.X,
        OriginY:  ctx.Y,
        ClipRect: ctx.ClipRect,
        Style:    style.Default,
    }
}

// Translate 创建一个基于当前坐标系偏移的子绘制器
// 用于绘制子组件
func (p *Painter) Translate(dx, dy int, w, h int) *Painter {
    newOriginX := p.OriginX + dx
    newOriginY := p.OriginY + dy
    
    // 计算新的剪裁区域（取交集）
    newRect := Rect{X: newOriginX, Y: newOriginY, Width: w, Height: h}
    intersectRect := p.ClipRect.Intersect(newRect)
    
    return &Painter{
        Buffer:   p.Buffer,
        OriginX:  newOriginX,
        OriginY:  newOriginY,
        ClipRect: *intersectRect,
        Style:    p.Style,
    }
}

```

### 2. 封装原子绘图方法

在 `Painter` 上实现常用的绘图原语，组件只需要调用这些方法，而不需要关心底层的 `SetCell`。

```go
// SetCell 绘制单个字符（自动处理相对坐标和剪裁）
func (p *Painter) SetCell(x, y int, char rune, s Style) {
    // 1. 转换为绝对坐标
    absX := p.OriginX + x
    absY := p.OriginY + y

    // 2. 剪裁检查 (Clipping Check)
    // 如果超出了给定的绘制区域，直接忽略，防止破坏其他 UI
    if !p.ClipRect.Contains(absX, absY) {
        return
    }

    // 3. 写入 Buffer
    p.Buffer.SetCell(absX, absY, char, s)
}

// Print 绘制字符串
func (p *Painter) Print(x, y int, text string, s Style) {
    // 使用 runewidth 库处理宽字符（如中文）
    cx := x
    for _, char := range text {
        w := runewidth.RuneWidth(char)
        if w == 0 { continue }
        
        // 绘制字符
        p.SetCell(cx, y, char, s)
        
        // 如果是全角字符，占位符处理（视具体 Buffer 实现而定）
        if w > 1 {
            // p.SetCell(cx+1, y, 0, s) // 或设置为空占位
        }
        cx += w
    }
}

// DrawRect 绘制填充矩形
func (p *Painter) FillRect(x, y, w, h int, char rune, s Style) {
    for ly := 0; ly < h; ly++ {
        for lx := 0; lx < w; lx++ {
            p.SetCell(x+lx, y+ly, char, s)
        }
    }
}

// DrawBorder 绘制边框
func (p *Painter) DrawBorder(x, y, w, h int, s Style) {
    // 绘制四个角
    p.SetCell(x, y, '┌', s)
    p.SetCell(x+w-1, y, '┐', s)
    p.SetCell(x, y+h-1, '└', s)
    p.SetCell(x+w-1, y+h-1, '┘', s)

    // 绘制横线
    for i := 1; i < w-1; i++ {
        p.SetCell(x+i, y, '─', s)
        p.SetCell(x+i, y+h-1, '─', s)
    }
    
    // 绘制竖线
    for i := 1; i < h-1; i++ {
        p.SetCell(x, y+i, '│', s)
        p.SetCell(x+w-1, y+i, '│', s)
    }
}

```

### 3. 组件实现的对比

#### **修改前 (痛苦模式)**：

```go
func (c *Box) Paint(ctx component.PaintContext, buf *Buffer) {
    // 必须手动加 ctx.X, ctx.Y
    // 必须自己写循环画线
    // 很容易忘记边界检查
    
    x, y := ctx.X, ctx.Y
    w, h := ctx.Width, ctx.Height
    
    buf.SetCell(x, y, '┌', style) // 左上角
    for i := 1; i < w-1; i++ {
        buf.SetCell(x+i, y, '─', style) // 上边
    }
    // ... 还有十几行代码 ...
}

```

#### **修改后 (舒适模式)**：

```go
func (c *Box) Paint(ctx component.PaintContext, buf *Buffer) {
    // 创建 Painter
    p := paint.NewPainter(buf, ctx)
    
    // 使用相对坐标 (0,0) 就是组件左上角
    // 自动处理了 ctx.X/Y 的偏移
    p.DrawBorder(0, 0, c.width, c.height, c.style)
    
    // 绘制内容
    p.Print(1, 1, "Hello World", c.textStyle)
}

```

### 4. 进阶建议

1. **Style 继承**: `Painter` 可以保存当前上下文的 `Style`。当你调用 `p.Print(x, y, text, nil)` 时，如果没有传入具体样式，可以使用 `Painter` 中保存的父级样式，这样更容易做主题统一。
2. **Layer (图层)**: 在 `Painter` 内部，你可以不直接写 `Buffer`，而是写到一个临时的 layer 中，最后再合并。但这对于简单 TUI 可能过度设计了，**Clipping (剪裁)** 是最重要的。
3. **Rich Text**: 可以实现一个 `DrawStyledString`，支持解析简单的颜色标记（如 `[red]Error[-]`），这样在做日志显示或富文本时非常方便。

**总结**：绝对不要在业务组件里直接操作 `Buffer` 的 `SetCell`。**请务必封装一个 `Painter` 工具类**，处理好相对坐标和剪裁，这会节省你大量的调试时间。

好的，我们继续。上一节我们讨论了引入 `Painter` 抽象层来简化绘制逻辑的重要性。现在，我将基于你提供的 `style.go` 和 `flex.go`，给出 **Painter 的具体实现**，并展示如何用它来重构 `Flex` 布局和实现基础的 `Box` 与 `Text` 组件。

这是构建 "Phase 1"（静态展示层）的核心步骤。

### 56. 核心工具：Painter 实现

我们需要在 `framework/paint` 包下创建 `painter.go`。它封装了 `Buffer`，自动处理 **绝对坐标转换** 和 **剪裁 (Clipping)**，并集成 `Style`。

```go
package paint

import (
	"github.com/mattn/go-runewidth"
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Painter 绘制辅助器
// 负责处理坐标偏移、剪裁和样式应用
type Painter struct {
	Buffer   *Buffer
	OriginX  int            // 相对原点 X (绝对坐标)
	OriginY  int            // 相对原点 Y (绝对坐标)
	ClipRect component.Rect // 剪裁区域 (绝对坐标)
	Style    style.Style    // 当前默认样式
}

// NewPainter 创建一个新的绘制器
func NewPainter(buf *Buffer, ctx component.PaintContext) *Painter {
	return &Painter{
		Buffer:   buf,
		OriginX:  ctx.X,
		OriginY:  ctx.Y,
		ClipRect: ctx.ClipRect,
		Style:    style.NewStyle(), // 默认无样式
	}
}

// WithStyle 返回带有新样式的 Painter 副本
func (p *Painter) WithStyle(s style.Style) *Painter {
	newP := *p
	newP.Style = s
	return &newP
}

// Translate 创建一个基于当前坐标系偏移的子绘制器
// 通常用于绘制子组件
func (p *Painter) Translate(dx, dy int, w, h int) *Painter {
	newOriginX := p.OriginX + dx
	newOriginY := p.OriginY + dy

	// 计算新的剪裁区域（取交集）
	// 子组件不能超出父组件定义的区域
	childRect := component.Rect{X: newOriginX, Y: newOriginY, Width: w, Height: h}
	intersectRect := p.ClipRect.Intersect(&childRect)
	
	// 如果完全没有交集，ClipRect 会是 nil 或者空，这里简化处理
	if intersectRect == nil {
		intersectRect = &component.Rect{}
	}

	return &Painter{
		Buffer:   p.Buffer,
		OriginX:  newOriginX,
		OriginY:  newOriginY,
		ClipRect: *intersectRect,
		Style:    p.Style,
	}
}

// SetCell 在相对坐标 (x, y) 绘制一个字符
func (p *Painter) SetCell(x, y int, char rune, s style.Style) {
	// 1. 转换为绝对坐标
	absX := p.OriginX + x
	absY := p.OriginY + y

	// 2. 剪裁检查 (Clipping Check)
	if !p.ClipRect.Contains(absX, absY) {
		return
	}

	// 3. 样式合并 (传入的样式优先级高于 Painter 默认样式)
	finalStyle := p.Style.Merge(s)

	// 4. 写入 Buffer
	p.Buffer.SetCell(absX, absY, char, finalStyle)
}

// Print 绘制字符串 (单行)
func (p *Painter) Print(x, y int, text string, s style.Style) {
	cx := x
	for _, char := range text {
		w := runewidth.RuneWidth(char)
		if w == 0 {
			continue
		}
		
		p.SetCell(cx, y, char, s)
		
		// 宽字符占位处理 (视 Buffer 实现而定，通常需要跳过下一个格)
		if w > 1 {
			// p.SetCell(cx+1, y, 0, s) // 可选：设置占位符
		}
		cx += w
	}
}

// FillRect 填充矩形区域
func (p *Painter) FillRect(x, y, w, h int, char rune, s style.Style) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			p.SetCell(x+lx, y+ly, char, s)
		}
	}
}

// DrawBorder 绘制边框
func (p *Painter) DrawBorder(x, y, w, h int, s style.Style) {
	if w < 2 || h < 2 {
		return
	}
	
	// 绘制四个角
	p.SetCell(x, y, '┌', s)
	p.SetCell(x+w-1, y, '┐', s)
	p.SetCell(x, y+h-1, '└', s)
	p.SetCell(x+w-1, y+h-1, '┘', s)

	// 绘制横线
	for i := 1; i < w-1; i++ {
		p.SetCell(x+i, y, '─', s)
		p.SetCell(x+i, y+h-1, '─', s)
	}

	// 绘制竖线
	for i := 1; i < h-1; i++ {
		p.SetCell(x, y+i, '│', s)
		p.SetCell(x+w-1, y+i, '│', s)
	}
}

```

---

### 57. 重构 Flex 组件

使用新的 `Painter`，我们可以简化 `Flex` 的 `Paint` 方法。注意不需要再手动计算 `ctx.X + currentPos`，而是创建一个子 `Painter`。

```go
// tui/framework/layout/flex.go (更新 Paint 方法)

func (f *Flex) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !f.IsVisible() || f.ChildCount() == 0 {
		return
	}

	// 创建 Painter
	p := paint.NewPainter(buf, ctx)

	children := f.GetChildren()
	availableWidth := ctx.AvailableWidth
	availableHeight := ctx.AvailableHeight

	// 计算布局 (复用之前的逻辑)
	sizes := f.calculateSizes(availableWidth, availableHeight)

	currentPos := 0
	for i, child := range children {
		childW, childH := sizes[i].width, sizes[i].height
		
		// 确定子组件的相对坐标
		childX, childY := 0, 0
		if f.direction == Row {
			childX = currentPos
		} else {
			childY = currentPos
		}

		// 创建子组件的上下文和剪裁区域
		// 注意：这里我们使用 Painter 的 Translate 来生成逻辑上的子区域
		// 但由于 component.Paintable 接口接收的是 ctx 和 buf，
		// 我们需要手动构建新的 childCtx
		
		// 计算绝对坐标用于 Context
		absX := ctx.X + childX
		absY := ctx.Y + childY
		
		childCtx := component.PaintContext{
			AvailableWidth:  childW,
			AvailableHeight: childH,
			X:               absX,
			Y:               absY,
			// 关键：计算剪裁区域交集，防止子组件溢出 Flex 容器
			ClipRect: *ctx.ClipRect.Intersect(&component.Rect{
				X: absX, Y: absY, Width: childW, Height: childH,
			}),
			Data: ctx.Data, // 传递数据上下文
		}

		if paintable, ok := child.(component.Paintable); ok {
			paintable.Paint(childCtx, buf)
		}

		// 更新位置
		if f.direction == Row {
			currentPos += childW + f.gap
		} else {
			currentPos += childH + f.gap
		}
	}
}

```

---

### 58. 新增组件：Box (容器与边框)

`Box` 是最常用的布局组件，用于添加边框、内边距 (Padding) 和背景色。

```go
package widget

import (
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/paint"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Box 基础容器组件
type Box struct {
	*component.BaseContainer
	style   style.Style
	padding [4]int // Top, Right, Bottom, Left
	border  bool
	title   string
}

func NewBox() *Box {
	return &Box{
		BaseContainer: component.NewBaseContainer("box"),
		style:         style.NewStyle(),
		border:        true,
	}
}

func (b *Box) SetStyle(s style.Style) *Box {
	b.style = s
	return b
}

func (b *Box) SetTitle(title string) *Box {
	b.title = title
	return b
}

func (b *Box) SetPadding(all int) *Box {
	b.padding = [4]int{all, all, all, all}
	return b
}

// Measure 测量：包含 Padding 和 Border 的尺寸
func (b *Box) Measure(maxWidth, maxHeight int) (int, int) {
	// 计算内容可用空间
	hPad := b.padding[1] + b.padding[3]
	vPad := b.padding[0] + b.padding[2]
	if b.border {
		hPad += 2
		vPad += 2
	}

	contentMaxW := maxWidth - hPad
	contentMaxH := maxHeight - vPad
	if contentMaxW < 0 { contentMaxW = 0 }
	if contentMaxH < 0 { contentMaxH = 0 }

	// 测量子组件 (假设 Box 只有一个子组件，或者自己管理布局)
	// 如果 Box 作为 Flex 的 item，通常由 Flex 决定大小
	// 这里简单返回 0,0 或者子组件的大小
	
	// 如果 Box 只有一个子组件
	if len(b.GetChildren()) > 0 {
		child := b.GetChildren()[0]
		if m, ok := child.(component.Measurable); ok {
			cw, ch := m.Measure(contentMaxW, contentMaxH)
			return cw + hPad, ch + vPad
		}
	}
	
	// 默认填满或者最小尺寸
	return hPad, vPad
}

func (b *Box) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	p := paint.NewPainter(buf, ctx)
	w, h := ctx.AvailableWidth, ctx.AvailableHeight

	// 1. 绘制背景
	p.FillRect(0, 0, w, h, ' ', b.style)

	// 2. 绘制边框
	if b.border {
		p.DrawBorder(0, 0, w, h, b.style)
		// 绘制标题
		if b.title != "" {
			p.Print(2, 0, " "+b.title+" ", b.style)
		}
	}

	// 3. 绘制子组件
	// 计算内容区域的偏移量
	offsetX := b.padding[3] // Left
	offsetY := b.padding[0] // Top
	if b.border {
		offsetX++
		offsetY++
	}

	contentW := w - offsetX - b.padding[1] // Right
	if b.border { contentW-- }
	contentH := h - offsetY - b.padding[2] // Bottom
	if b.border { contentH-- }

	if contentW <= 0 || contentH <= 0 {
		return
	}

	// 递归绘制子组件
	// Box 通常只有一个孩子，作为 wrapper
	for _, child := range b.GetChildren() {
		// 手动构建子 Context，复用我们在 Flex 中讨论的逻辑
		absX := ctx.X + offsetX
		absY := ctx.Y + offsetY
		
		childCtx := component.PaintContext{
			AvailableWidth:  contentW,
			AvailableHeight: contentH,
			X:               absX,
			Y:               absY,
			ClipRect:        *ctx.ClipRect.Intersect(&component.Rect{X: absX, Y: absY, Width: contentW, Height: contentH}),
			Data:            ctx.Data,
		}
		
		if paintable, ok := child.(component.Paintable); ok {
			paintable.Paint(childCtx, buf)
		}
	}
}

```

---

### 59. 新增组件：Text (文本显示)

这是最基本的叶子组件，用于显示静态文本或绑定数据。

```go
package widget

import (
	"github.com/mattn/go-runewidth"
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/paint"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Text 文本组件
type Text struct {
	*component.BaseComponent
	content string
	align   Align // Left, Center, Right
	style   style.Style
}

type Align int
const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

func NewText(content string) *Text {
	return &Text{
		BaseComponent: component.NewBaseComponent("text"),
		content:       content,
		style:         style.NewStyle(),
	}
}

func (t *Text) SetStyle(s style.Style) *Text {
	t.style = s
	return t
}

func (t *Text) SetAlign(a Align) *Text {
	t.align = a
	return t
}

func (t *Text) Measure(maxWidth, maxHeight int) (int, int) {
	w := runewidth.StringWidth(t.content)
	return w, 1 // 简单实现：单行文本高度为 1
}

func (t *Text) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	p := paint.NewPainter(buf, ctx)
	
	// 计算对齐
	textWidth := runewidth.StringWidth(t.content)
	x := 0
	
	switch t.align {
	case AlignCenter:
		x = (ctx.AvailableWidth - textWidth) / 2
	case AlignRight:
		x = ctx.AvailableWidth - textWidth
	}
	
	if x < 0 { x = 0 }

	p.Print(x, 0, t.content, t.style)
}

```

### 总结

现在我们拥有了：

1. **Painter**: 强大的绘图工具，屏蔽了底层 Buffer 操作。
2. **Flex**: 布局容器，已适配 PaintContext。
3. **Box**: 装饰容器，支持边框和 Padding。
4. **Text**: 内容展示组件。

这四个文件构成了 TUI 的 **"DOM 树"** 基础。你可以使用它们组合出像下面这样的界面：

```go
// 示例：创建一个带边框的居中欢迎语
root := layout.NewFlex(layout.Column).
    WithChildren(
        widget.NewBox().
            SetTitle("Welcome").
            SetStyle(style.NewStyle().Foreground(style.Green)).
            Add(
                widget.NewText("Hello, Yao TUI!").SetAlign(widget.AlignCenter),
            ),
    )

```

接下来你可以继续实现 `framework/event/pump.go` (事件泵) 来让这些静态组件动起来。

好的，我们已经有了静态的“DOM 树”结构（Box, Text, Flex）和强大的绘制工具（Painter）。现在的界面能看，但还不能动。

接下来的关键步骤是实现 **Phase 2: 交互 (The Interaction)**。我们需要让应用能够接收键盘输入，并让组件对这些输入做出反应。

这将涉及三个核心模块：

1. **Event System**: 定义标准化的事件（键盘、鼠标、Resize）。
2. **Event Pump**: 事件泵，负责从底层终端（如 `tcell`）抽取事件并标准化。
3. **Focus Manager**: 焦点管理器，决定哪个组件当前接收键盘输入。

---

### 60. 定义事件系统 (Event System)

虽然 `tcell` 有自己的事件类型，但为了解耦，我们需要定义框架自己的事件接口。

在 `framework/event/event.go` 中：

```go
package event

import (
	"fmt"
	"time"
)

// Type 事件类型
type Type int

const (
	TypeKey    Type = iota // 键盘事件
	TypeMouse              // 鼠标事件
	TypeResize             // 窗口调整大小
	TypeError              // 错误事件
)

// Event 基础事件接口
type Event interface {
	Type() Type
	When() time.Time
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	Timestamp time.Time
}

func (e BaseEvent) When() time.Time { return e.Timestamp }

// =============================================================================
// 键盘事件
// =============================================================================

type Key int16

const (
	KeyRune Key = iota + 256 // 普通字符
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyEnter
	KeyEsc
	KeyBackspace
	KeyTab
	KeySpace
	// ... 其他功能键 F1-F12 等
)

// KeyEvent 键盘事件
type KeyEvent struct {
	BaseEvent
	Key  Key
	Rune rune   // 如果是普通字符，这里存储字符值
	Alt  bool
	Ctrl bool
	Meta bool
}

func (e KeyEvent) Type() Type { return TypeKey }
func (e KeyEvent) String() string {
	if e.Key == KeyRune {
		return fmt.Sprintf("Key('%c')", e.Rune)
	}
	return fmt.Sprintf("Key(%d)", e.Key)
}

// NewKey 创建按键事件
func NewKey(k Key, r rune) KeyEvent {
	return KeyEvent{
		BaseEvent: BaseEvent{Timestamp: time.Now()},
		Key:       k,
		Rune:      r,
	}
}

// =============================================================================
// 窗口调整事件
// =============================================================================

// ResizeEvent 调整大小事件
type ResizeEvent struct {
	BaseEvent
	Width  int
	Height int
}

func (e ResizeEvent) Type() Type { return TypeResize }

```

---

### 61. 实现事件泵 (Event Pump)

事件泵的作用是运行在一个独立的 Goroutine 中，阻塞地从终端读取事件，将其转换为我们的 `event.Event`，然后发送到 `App` 的主循环通道。

在 `framework/event/pump.go` 中：

```go
package event

import (
	"github.com/gdamore/tcell/v2" // 假设底层使用 tcell，也可以是 bubbles/tea 或 termbox
)

// Pump 事件泵
type Pump struct {
	screen  tcell.Screen
	outChan chan Event
	stopChan chan struct{}
}

// NewPump 创建事件泵
func NewPump(s tcell.Screen) *Pump {
	return &Pump{
		screen:   s,
		outChan:  make(chan Event, 10), // 带缓冲，防止阻塞
		stopChan: make(chan struct{}),
	}
}

// Events 返回事件通道
func (p *Pump) Events() <-chan Event {
	return p.outChan
}

// Start 启动泵
func (p *Pump) Start() {
	go func() {
		for {
			select {
			case <-p.stopChan:
				close(p.outChan)
				return
			default:
				// 阻塞读取 tcell 事件
				ev := p.screen.PollEvent()
				if ev == nil { // 屏幕关闭
					return
				}

				// 转换并发送
				if converted := p.convert(ev); converted != nil {
					// 非阻塞发送，如果通道满了就丢弃（防止 UI 卡死导致事件积压）
					select {
					case p.outChan <- converted:
					default:
						// Log: event dropped
					}
				}
			}
		}
	}()
}

// Stop 停止泵
func (p *Pump) Stop() {
	close(p.stopChan)
}

// convert 将 tcell 事件转换为框架事件
func (p *Pump) convert(te tcell.Event) Event {
	switch ev := te.(type) {
	case *tcell.EventResize:
		w, h := ev.Size()
		return ResizeEvent{
			BaseEvent: BaseEvent{Timestamp: ev.When()},
			Width:     w,
			Height:    h,
		}
	case *tcell.EventKey:
		return p.convertKey(ev)
	// case *tcell.EventMouse:
	// 	   return p.convertMouse(ev)
	}
	return nil
}

func (p *Pump) convertKey(tek *tcell.EventKey) KeyEvent {
	k := KeyRune
	r := tek.Rune()
	
	// 简单的映射逻辑
	switch tek.Key() {
	case tcell.KeyUp: k = KeyUp
	case tcell.KeyDown: k = KeyDown
	case tcell.KeyLeft: k = KeyLeft
	case tcell.KeyRight: k = KeyRight
	case tcell.KeyEnter: k = KeyEnter
	case tcell.KeyEscape: k = KeyEsc
	case tcell.KeyBackspace, tcell.KeyBackspace2: k = KeyBackspace
	case tcell.KeyTab: k = KeyTab
	}
	
	return KeyEvent{
		BaseEvent: BaseEvent{Timestamp: tek.When()},
		Key:       k,
		Rune:      r,
		Ctrl:      tek.Modifiers()&tcell.ModCtrl != 0,
		Alt:       tek.Modifiers()&tcell.ModAlt != 0,
	}
}

```

---

### 62. 焦点管理器 (Focus Manager)

这是交互逻辑的核心。我们不能让事件广播给所有组件，只有拥有焦点的组件才能处理键盘事件。

设计思路：

* **Focusable**: 只有实现了该接口的组件才能获得焦点。
* **Ordered List**: 焦点管理器维护一个扁平化的可聚焦组件列表（Tab 顺序）。
* **Container Integration**: 容器组件（如 Form）负责将其子组件注册到焦点管理器中。

在 `framework/focus/manager.go` 中：

```go
package focus

import (
	"github.com/yaoapp/yao/tui/framework/component"
)

// Manager 焦点管理器
type Manager struct {
	items   []component.Component
	current int // 当前焦点索引
}

// NewManager 创建焦点管理器
func NewManager() *Manager {
	return &Manager{
		items:   make([]component.Component, 0),
		current: -1,
	}
}

// Register 注册可聚焦组件
func (m *Manager) Register(c component.Component) {
	// 只有实现了 Focusable 才能注册
	// 注意：BaseComponent 已经有了 OnFocus/OnBlur，默认都是 Focusable
	// 我们可以加一个 IsFocusable() bool 接口来过滤 Label 等不可聚焦组件
	m.items = append(m.items, c)
}

// Focus 设置焦点到指定索引
func (m *Manager) Focus(index int) {
	if len(m.items) == 0 {
		return
	}
	
	// 范围检查与循环
	if index < 0 {
		index = len(m.items) - 1
	} else if index >= len(m.items) {
		index = 0
	}

	// 1. Blur 当前组件
	if m.current >= 0 && m.current < len(m.items) {
		curr := m.items[m.current]
		curr.OnBlur()
		// 标记脏以便重绘（例如移除高亮边框）
		curr.MarkDirty()
	}

	// 2. 更新索引
	m.current = index

	// 3. Focus 新组件
	next := m.items[m.current]
	next.OnFocus()
	next.MarkDirty()
}

// Next 切换到下一个焦点 (Tab)
func (m *Manager) Next() {
	m.Focus(m.current + 1)
}

// Prev 切换到上一个焦点 (Shift+Tab)
func (m *Manager) Prev() {
	m.Focus(m.current - 1)
}

// Current 获取当前焦点组件
func (m *Manager) Current() component.Component {
	if m.current >= 0 && m.current < len(m.items) {
		return m.items[m.current]
	}
	return nil
}

// HandleKey 处理导航键
// 返回 true 表示已处理
func (m *Manager) HandleKey(key component.Key) bool {
	switch key {
	case component.KeyTab:
		m.Next()
		return true
	// Shift+Tab 处理略复杂，因为通常终端只发 Backtab 或者需要在 KeyMod 中判断
	}
	return false
}

```

---

### 63. 集成到 App 中

最后，我们需要在 `App` 结构体中把这三者串起来。

```go
// framework/app.go (伪代码)

type App struct {
	root         component.Component
	screen       tcell.Screen
	pump         *event.Pump
	focusManager *focus.Manager
	// ...
}

func (app *App) Run() {
	app.screen.Init()
	defer app.screen.Fini()

	// 1. 启动事件泵
	app.pump = event.NewPump(app.screen)
	app.pump.Start()

	// 2. 初始化焦点 (简单的自动收集策略)
	// 实际项目中可能由 DSL 解析器构建焦点链
	app.collectFocusables(app.root)
	app.focusManager.Focus(0)

	// 3. 主循环
	ticker := time.NewTicker(time.Millisecond * 33) // ~30 FPS
	for {
		select {
		case ev := <-app.pump.Events():
			app.handleEvent(ev)
		
		case <-ticker.C:
			// 定时重绘（如果脏）
			if app.isDirty {
				app.Draw()
			}
		}
	}
}

func (app *App) handleEvent(ev event.Event) {
	switch e := ev.(type) {
	case event.ResizeEvent:
		app.width, app.height = e.Width, e.Height
		app.root.Measure(app.width, app.height)
		app.MarkDirty()

	case event.KeyEvent:
		// 1. 全局快捷键 (Ctrl+C)
		if e.Key == event.KeyCtrlC {
			app.Quit()
			return
		}

		// 2. 焦点导航 (Tab)
		// 如果按的是 Tab，直接由 FocusManager 处理，不分发给组件
		if e.Key == event.KeyTab {
			app.focusManager.Next()
			return
		}

		// 3. 分发给当前焦点组件
		target := app.focusManager.Current()
		if target != nil {
			// 将 Event 转换为 Action 并分发
			// 这里假设 Component 实现了 HandleAction
			if handler, ok := target.(component.ActionHandler); ok {
				// 将 Event 包装为 ActionInput
				act := action.New(action.ActionInput, e)
				handler.HandleAction(act)
			}
		}
	}
}

```

### 实施建议

现在你已经有了：

1. 绘制层 (`Painter`)。
2. 布局层 (`Flex`, `Box`)。
3. 交互层 (`Pump`, `FocusManager`)。

**接下来的任务**：
请创建一个简单的 `main.go` 演示程序：

1. 创建一个 `App`。
2. 创建一个 `Column Flex`。
3. 添加两个 `TextInput` (使用之前定义的组件) 和一个 `Button`。
4. 运行 App，测试用 Tab 键在输入框和按钮之间切换焦点（观察边框颜色的变化）。

一旦这个 "Loop" 跑通，你就拥有了一个真正的 TUI 框架核心。