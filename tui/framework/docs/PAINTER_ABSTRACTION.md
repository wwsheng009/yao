# Painter Abstraction System Design (V3)

> **优先级**: P1 (显示层抽象)
> **目标**: 隔离平台差异，支持多显示后端
> **关键特性**: Painter 接口、CellBuffer、双缓冲、HTML 后端

## 概述

Painter 抽象系统是 TUI 框架与显示层之间的桥梁。它定义了统一的绘图接口，使得组件代码可以在不同显示后端（终端、HTML、原生窗口）上运行而无需修改。

### 为什么需要 Painter 抽象？

**没有抽象的问题**：
```go
// ❌ 组件直接依赖终端
func (t *Text) Paint(x, y int) {
    // 直接操作终端
    terminal.MoveCursor(x, y)
    terminal.WriteString(t.content)
    terminal.SetColor(t.color)
}

// 问题：
// - 无法切换到 HTML 显示
// - 无法实现双缓冲
// - 无法在不同终端间移植
// - 测试困难（需要真实终端）
```

**有 Painter 抽象的优势**：
```go
// ✅ 组件使用 Painter 接口
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer) {
    buf.SetCell(x, y, char, style)  // 写虚拟缓冲区
}

// 优势：
// - 支持 HTML/终端双后端
// - 天然双缓冲
// - 平台差异隔离
// - 易于测试（使用 MockPainter）
```

## 设计目标

1. **平台无关**: 组件代码不依赖具体平台
2. **多后端支持**: 终端、HTML、原生窗口
3. **双缓冲**: 支持离屏渲染和 Diff
4. **高性能**: 最小化不必要更新
5. **可测试**: 易于 mock 和测试

## 核心架构

```
┌─────────────────────────────────────────────────────────┐
│                    Component Layer                       │
│  (只调用 Painter 接口，不知道具体后端)                    │
└────────────────────┬────────────────────────────────────┘
                     │ Paint(ctx, buf)
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    Painter Interface                     │
│                                                         │
│  - SetCell(x, y, char, style)                           │
│  - FillRect(x, y, w, h, char, style)                    │
│  - DrawText(x, y, text, style)                          │
│  - DrawLine(x1, y1, x2, y2, style)                     │
└────────────────────┬────────────────────────────────────┘
                     │
         ┌───────────┼───────────┐
         ▼           ▼           ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │Terminal │ │  HTML   │ │ Native  │
    │Painter  │ │ Painter │ │ Painter │
    └─────────┘ └─────────┘ └─────────┘
         │           │           │
         ▼           ▼           ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │ stdout/ │ │Browser  │ │ Win32/  │
    │ ANSI    │ │ DOM     │ │ Cocoa   │
    └─────────┘ └─────────┘ └─────────┘
```

## 核心类型定义

### 1. Painter 接口

```go
// 位于: tui/runtime/paint/painter.go

package paint

// Painter 绘图接口
type Painter interface {
    // === 基础绘图 ===

    // SetCell 设置单元格
    SetCell(x, y int, char rune, style Style)

    // GetCell 获取单元格
    GetCell(x, y int) Cell

    // FillRect 填充矩形
    FillRect(x, y, width, height int, char rune, style Style)

    // Clear 清空缓冲区
    Clear(style Style)

    // === 文本绘制 ===

    // DrawText 绘制文本（自动换行）
    DrawText(x, y int, text string, style Style)

    // DrawTextAligned 绘制对齐文本
    DrawTextAligned(x, y, width int, text string, align Alignment, style Style)

    // === 图形绘制 ===

    // DrawLine 绘制线条
    DrawLine(x1, y1, x2, y2 int, style Style)

    // DrawRect 绘制矩形边框
    DrawRect(x, y, width, height int, style Style)

    // DrawBox 绘制带标题的框
    DrawBox(x, y, width, height int, title string, style Style)

    // === 滚动支持 ===

    // ScrollUp 向上滚动
    ScrollUp(lines int)

    // ScrollDown 向下滚动
    ScrollDown(lines int)

    // === 光标控制 ===

    // ShowCursor 显示光标
    ShowCursor(x, y int)

    // HideCursor 隐藏光标

    // === 缓冲区操作 ===

    // Flush 刷新到显示设备
    Flush() error

    // Size 获取缓冲区大小
    Size() (width, height int)

    // Snapshot 获取快照（用于 Diff）
    Snapshot() *CellBuffer
}

// Alignment 对齐方式
type Alignment int

const (
    AlignLeft Alignment = iota
    AlignCenter
    AlignRight
)

// Style 单元格样式
type Style struct {
    Foreground Color
    Background Color
    Bold       bool
    Dim        bool
    Italic     bool
    Underline  bool
    Blink      bool
    Reverse    bool
}

// Cell 单元格
type Cell struct {
    Char     rune
    Style    Style
    Width    int  // 字符宽度（1 或 2，用于中文等宽字符）
    Modified bool // 是否被修改
}

// Color 颜色
type Color struct {
    Type  ColorType
    Value uint8
}

type ColorType int

const (
    ColorDefault ColorType = iota
    ColorANSI                // 16 色
    Color256                 // 256 色
    ColorRGB                 // 真彩色
)

// 预定义颜色
const (
    ColorBlack   Color = {Type: ColorANSI, Value: 0}
    ColorRed     Color = {Type: ColorANSI, Value: 1}
    ColorGreen   Color = {Type: ColorANSI, Value: 2}
    ColorYellow  Color = {Type: ColorANSI, Value: 3}
    ColorBlue    Color = {Type: ColorANSI, Value: 4}
    ColorMagenta Color = {Type: ColorANSI, Value: 5}
    ColorCyan    Color = {Type: ColorANSI, Value: 6}
    ColorWhite   Color = {Type: ColorANSI, Value: 7}
)

// RGBColor 创建 RGB 颜色
func RGBColor(r, g, b uint8) Color {
    return Color{
        Type:  ColorRGB,
        Value: (r << 16) | (g << 8) | b,
    }
}

// ANSIColor 创建 ANSI 颜色
func ANSIColor(value uint8) Color {
    return Color{
        Type:  ColorANSI,
        Value: value,
    }
}
```

### 2. CellBuffer 实现

```go
// 位于: tui/runtime/paint/cellbuffer.go

package paint

// CellBuffer 单元格缓冲区
type CellBuffer struct {
    cells  [][]Cell
    width  int
    height int
    dirty  bool
}

// NewCellBuffer 创建缓冲区
func NewCellBuffer(width, height int) *CellBuffer {
    cells := make([][]Cell, height)
    for y := range cells {
        cells[y] = make([]Cell, width)
        for x := range cells[y] {
            cells[y][x] = Cell{Char: ' ', Style: DefaultStyle()}
        }
    }

    return &CellBuffer{
        cells:  cells,
        width:  width,
        height: height,
    }
}

// Resize 调整大小
func (b *CellBuffer) Resize(width, height int) {
    if width == b.width && height == b.height {
        return
    }

    newCells := make([][]Cell, height)
    for y := range newCells {
        newCells[y] = make([]Cell, width)
        for x := range newCells[y] {
            if y < b.height && x < b.width {
                newCells[y][x] = b.cells[y][x]
            } else {
                newCells[y][x] = Cell{Char: ' ', Style: DefaultStyle()}
            }
        }
    }

    b.cells = newCells
    b.width = width
    b.height = height
    b.dirty = true
}

// SetCell 设置单元格
func (b *CellBuffer) SetCell(x, y int, char rune, style Style) {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return
    }

    cell := &b.cells[y][x]
    cell.Char = char
    cell.Style = style
    cell.Width = runeWidth(char)
    cell.Modified = true
    b.dirty = true
}

// GetCell 获取单元格
func (b *CellBuffer) GetCell(x, y int) Cell {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return Cell{Char: ' ', Style: DefaultStyle()}
    }
    return b.cells[y][x]
}

// FillRect 填充矩形
func (b *CellBuffer) FillRect(x, y, width, height int, char rune, style Style) {
    for dy := 0; dy < height; dy++ {
        for dx := 0; dx < width; dx++ {
            b.SetCell(x+dx, y+dy, char, style)
        }
    }
}

// Clear 清空缓冲区
func (b *CellBuffer) Clear(style Style) {
    b.FillRect(0, 0, b.width, b.height, ' ', style)
    b.dirty = false
}

// DrawText 绘制文本
func (b *CellBuffer) DrawText(x, y int, text string, style Style) {
    dx := 0
    dy := 0
    for _, r := range text {
        if r == '\n' {
            dy++
            dx = 0
            continue
        }

        if x+dx >= b.width {
            break
        }

        b.SetCell(x+dx, y+dy, r, style)
        dx += runeWidth(r)
    }
}

// DrawTextAligned 绘制对齐文本
func (b *CellBuffer) DrawTextAligned(x, y, width int, text string, align Alignment, style Style) {
    textWidth := stringWidth(text)

    var startX int
    switch align {
    case AlignLeft:
        startX = x
    case AlignCenter:
        startX = x + (width-textWidth)/2
    case AlignRight:
        startX = x + width - textWidth
    }

    b.DrawText(startX, y, text, style)
}

// DrawLine 绘制线条
func (b *CellBuffer) DrawLine(x1, y1, x2, y2 int, style Style) {
    // Bresenham 算法
    dx := abs(x2 - x1)
    dy := abs(y2 - y1)
    sx := sign(x2 - x1)
    sy := sign(y2 - y1)
    err := dx - dy

    x, y := x1, y1
    for {
        b.SetCell(x, y, lineChar(x, y, x1, y1, x2, y2), style)

        if x == x2 && y == y2 {
            break
        }

        e2 := 2 * err
        if e2 > -dy {
            err -= dy
            x += sx
        }
        if e2 < dx {
            err += dx
            y += sy
        }
    }
}

// DrawRect 绘制矩形边框
func (b *CellBuffer) DrawRect(x, y, width, height int, style Style) {
    // 顶边
    for dx := 0; dx < width; dx++ {
        ch := '─'
        if dx == 0 {
            ch = '┌'
        } else if dx == width-1 {
            ch = '┐'
        }
        b.SetCell(x+dx, y, ch, style)
    }

    // 底边
    for dx := 0; dx < width; dx++ {
        ch := '─'
        if dx == 0 {
            ch = '└'
        } else if dx == width-1 {
            ch = '┘'
        }
        b.SetCell(x+dx, y+height-1, ch, style)
    }

    // 左右边
    for dy := 1; dy < height-1; dy++ {
        b.SetCell(x, y+dy, '│', style)
        b.SetCell(x+width-1, y+dy, '│', style)
    }
}

// DrawBox 绘制带标题的框
func (b *CellBuffer) DrawBox(x, y, width, height int, title string, style Style) {
    b.DrawRect(x, y, width, height, style)

    if title != "" && width > 4 {
        // 绘制标题
        titleX := x + 2
        titleText := " " + title + " "
        if len(titleText) > width-4 {
            titleText = titleText[:width-4]
        }
        b.DrawText(titleX, y, titleText, style)
    }
}

// ShowCursor 显示光标
func (b *CellBuffer) ShowCursor(x, y int) {
    // 在缓冲区中标记光标位置
    // 实际显示由后端处理
}

// HideCursor 隐藏光标
func (b *CellBuffer) HideCursor() {
    // 标记隐藏光标
}

// Flush 刷新到显示设备
func (b *CellBuffer) Flush() error {
    b.dirty = false
    return nil
}

// Size 获取缓冲区大小
func (b *CellBuffer) Size() (int, int) {
    return b.width, b.height
}

// Snapshot 获取快照
func (b *CellBuffer) Snapshot() *CellBuffer {
    snapshot := NewCellBuffer(b.width, b.height)
    for y := range b.cells {
        copy(snapshot.cells[y], b.cells[y])
    }
    return snapshot
}

// Diff 计算与另一个缓冲区的差异
func (b *CellBuffer) Diff(other *CellBuffer) []CellChange {
    changes := make([]CellChange, 0)

    width, height := b.Size()
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            thisCell := b.cells[y][x]
            otherCell := other.cells[y][x]

            if thisCell.Char != otherCell.Char ||
                !styleEqual(thisCell.Style, otherCell.Style) {
                changes = append(changes, CellChange{
                    X:     x,
                    Y:     y,
                    Cell:  thisCell,
                })
            }
        }
    }

    return changes
}

// CellChange 单元格变化
type CellChange struct {
    X    int
    Y    int
    Cell Cell
}

// 辅助函数
func runeWidth(r rune) int {
    if r >= 0x1100 {
        // 中日韩字符等宽为 2
        return 2
    }
    return 1
}

func stringWidth(s string) int {
    width := 0
    for _, r := range s {
        width += runeWidth(r)
    }
    return width
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

func sign(x int) int {
    if x < 0 {
        return -1
    }
    if x > 0 {
        return 1
    }
    return 0
}

func lineChar(x, y, x1, y1, x2, y2 int) rune {
    // 返回合适的线条字符
    if x == x1 && x == x2 {
        return '│'
    }
    if y == y1 && y == y2 {
        return '─'
    }
    return '┼'
}

func styleEqual(a, b Style) bool {
    return a.Foreground == b.Foreground &&
        a.Background == b.Background &&
        a.Bold == b.Bold &&
        a.Dim == b.Dim &&
        a.Italic == b.Italic &&
        a.Underline == b.Underline &&
        a.Blink == b.Blink &&
        a.Reverse == b.Reverse
}

func DefaultStyle() Style {
    return Style{
        Foreground: Color{Type: ColorDefault},
        Background: Color{Type: ColorDefault},
    }
}
```

### 3. Terminal Painter

```go
// 位于: tui/runtime/paint/terminal_painter.go

package paint

import (
    "io"
    "github.com/yaoapp/yao/tui/framework/platform/terminal"
)

// TerminalPainter 终端绘图器
type TerminalPainter struct {
    output    io.Writer
    buffer    *CellBuffer
    prev      *CellBuffer  // 上一帧，用于 Diff
    cursorVisible bool
}

// NewTerminalPainter 创建终端绘图器
func NewTerminalPainter(output io.Writer, width, height int) *TerminalPainter {
    return &TerminalPainter{
        output: output,
        buffer: NewCellBuffer(width, height),
        prev:   NewCellBuffer(width, height),
    }
}

// SetCell 设置单元格
func (p *TerminalPainter) SetCell(x, y int, char rune, style Style) {
    p.buffer.SetCell(x, y, char, style)
}

// GetCell 获取单元格
func (p *TerminalPainter) GetCell(x, y int) Cell {
    return p.buffer.GetCell(x, y)
}

// FillRect 填充矩形
func (p *TerminalPainter) FillRect(x, y, width, height int, char rune, style Style) {
    p.buffer.FillRect(x, y, width, height, char, style)
}

// Clear 清空缓冲区
func (p *TerminalPainter) Clear(style Style) {
    p.buffer.Clear(style)
}

// DrawText 绘制文本
func (p *TerminalPainter) DrawText(x, y int, text string, style Style) {
    p.buffer.DrawText(x, y, text, style)
}

// DrawTextAligned 绘制对齐文本
func (p *TerminalPainter) DrawTextAligned(x, y, width int, text string, align Alignment, style Style) {
    p.buffer.DrawTextAligned(x, y, width, text, align, style)
}

// DrawLine 绘制线条
func (p *TerminalPainter) DrawLine(x1, y1, x2, y2 int, style Style) {
    p.buffer.DrawLine(x1, y1, x2, y2, style)
}

// DrawRect 绘制矩形边框
func (p *TerminalPainter) DrawRect(x, y, width, height int, style Style) {
    p.buffer.DrawRect(x, y, width, height, style)
}

// DrawBox 绘制带标题的框
func (p *TerminalPainter) DrawBox(x, y, width, height int, title string, style Style) {
    p.buffer.DrawBox(x, y, width, height, title, style)
}

// ShowCursor 显示光标
func (p *TerminalPainter) ShowCursor(x, y int) {
    p.cursorVisible = true
    p.buffer.ShowCursor(x, y)
}

// HideCursor 隐藏光标
func (p *TerminalPainter) HideCursor() {
    p.cursorVisible = false
    p.buffer.HideCursor()
}

// Flush 刷新到终端
func (p *TerminalPainter) Flush() error {
    // 计算 Diff
    changes := p.buffer.Diff(p.prev)

    // 只更新变化的单元格
    for _, change := range changes {
        p.drawCell(change.X, change.Y, change.Cell)
    }

    // 处理光标
    if p.cursorVisible {
        // terminal.MoveCursor(...)
    } else {
        terminal.HideCursor(p.output)
    }

    // 更新上一帧
    p.prev = p.buffer.Snapshot()

    return nil
}

// drawCell 绘制单元格到终端
func (p *TerminalPainter) drawCell(x, y int, cell Cell) {
    // 移动光标
    terminal.MoveCursor(p.output, x, y)

    // 设置样式
    p.applyStyle(cell.Style)

    // 输出字符
    if cell.Width == 2 {
        // 宽字符，需要特殊处理
        p.output.Write([]byte{0xEF, 0xBC, 0x80}) // placeholder
    } else {
        p.output.Write([]byte{byte(cell.Char)})
    }

    // 重置样式
    terminal.ResetStyle(p.output)
}

// applyStyle 应用样式
func (p *TerminalPainter) applyStyle(style Style) {
    // 前景色
    terminal.SetForeground(p.output, style.Foreground)

    // 背景色
    terminal.SetBackground(p.output, style.Background)

    // 样式
    if style.Bold {
        terminal.SetBold(p.output)
    }
    if style.Dim {
        terminal.SetDim(p.output)
    }
    if style.Italic {
        terminal.SetItalic(p.output)
    }
    if style.Underline {
        terminal.SetUnderline(p.output)
    }
    if style.Blink {
        terminal.SetBlink(p.output)
    }
    if style.Reverse {
        terminal.SetReverse(p.output)
    }
}

// Size 获取缓冲区大小
func (p *TerminalPainter) Size() (int, int) {
    return p.buffer.Size()
}

// Snapshot 获取快照
func (p *TerminalPainter) Snapshot() *CellBuffer {
    return p.buffer.Snapshot()
}
```

### 4. HTML Painter

```go
// 位于: tui/runtime/paint/html_painter.go

package paint

import (
    "strings"
    "github.com/yaoapp/yao/tui/framework/platform/html"
)

// HTMLPainter HTML 绘图器
type HTMLPainter struct {
    buffer     *CellBuffer
    elementIDs map[string]string  // 单元格到 DOM 元素的映射
    container  string             // 容器 ID
}

// NewHTMLPainter 创建 HTML 绘图器
func NewHTMLPainter(container string, width, height int) *HTMLPainter {
    return &HTMLPainter{
        buffer:     NewCellBuffer(width, height),
        elementIDs: make(map[string]string),
        container:  container,
    }
}

// 大部分方法与 TerminalPainter 类似，直接操作 buffer

// Flush 刷新到 HTML DOM
func (p *HTMLPainter) Flush() error {
    var sb strings.Builder

    width, height := p.buffer.Size()

    sb.WriteString("<div class=\"tui-container\" style=\"")
    sb.WriteString(p.getContainerStyle())
    sb.WriteString("\">")

    for y := 0; y < height; y++ {
        sb.WriteString("<div class=\"tui-row\">")
        for x := 0; x < width; x++ {
            cell := p.buffer.GetCell(x, y)
            sb.WriteString(p.renderCell(x, y, cell))
        }
        sb.WriteString("</div>")
    }

    sb.WriteString("</div>")

    // 更新 DOM
    html.SetInnerHTML(p.container, sb.String())

    return nil
}

// renderCell 渲染单元格为 HTML
func (p *HTMLPainter) renderCell(x, y int, cell Cell) string {
    var sb strings.Builder

    sb.WriteString("<span class=\"tui-cell\" style=\"")
    sb.WriteString(p.getCellStyle(cell.Style))
    sb.WriteString("\">")

    // 转义 HTML 字符
    sb.WriteString(html.Escape(string(cell.Char)))

    sb.WriteString("</span>")

    return sb
}

// getCellStyle 获取单元格样式
func (p *HTMLPainter) getCellStyle(style Style) string {
    var styles []string

    // 颜色
    if style.Foreground.Type != ColorDefault {
        styles = append(styles, "color: "+p.colorToCSS(style.Foreground))
    }
    if style.Background.Type != ColorDefault {
        styles = append(styles, "background-color: "+p.colorToCSS(style.Background))
    }

    // 样式
    if style.Bold {
        styles = append(styles, "font-weight: bold")
    }
    if style.Italic {
        styles = append(styles, "font-style: italic")
    }
    if style.Underline {
        styles = append(styles, "text-decoration: underline")
    }
    if style.Blink {
        styles = append(styles, "animation: blink 1s infinite")
    }

    return strings.Join(styles, "; ")
}

// colorToCSS 颜色转 CSS
func (p *HTMLPainter) colorToCSS(color Color) string {
    switch color.Type {
    case ColorANSI:
        return p.ansiToCSS[color.Value]
    case Color256:
        return p.xterm256ToCSS(color.Value)
    case ColorRGB:
        r := (color.Value >> 16) & 0xFF
        g := (color.Value >> 8) & 0xFF
        b := color.Value & 0xFF
        return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)
    default:
        return ""
    }
}

// ANSI 到 CSS 颜色映射
var ansiToCSS = map[uint8]string{
    0:   "#000000",
    1:   "#cd0000",
    2:   "#00cd00",
    3:   "#cdcd00",
    4:   "#0000ee",
    5:   "#cd00cd",
    6:   "#00cdcd",
    7:   "#e5e5e5",
    8:   "#7f7f7f",
    9:   "#ff0000",
    10:  "#00ff00",
    11:  "#ffff00",
    12:  "#5c5cff",
    13:  "#ff00ff",
    14:  "#00ffff",
    15:  "#ffffff",
}

// getContainerStyle 获取容器样式
func (p *HTMLPainter) getContainerStyle() string    width, height := p.buffer.Size()

    styles := []string{
        "display: flex",
        "flex-direction: column",
        "font-family: monospace",
        "white-space: pre",
        "line-height: 1.2",
        fmt.Sprintf("width: %dch", width),
    }

    return strings.Join(styles, "; ")
}

// xterm256ToCSS xterm 256 色转 CSS
func (p *HTMLPainter) xterm256ToCSS(value uint8) string {
    // 简化的映射，实际应该使用完整的 xterm 256 色表
    if value < 16 {
        return p.ansiToCSS[value%16]
    }
    return fmt.Sprintf("#%06x", xterm256[value]))
}

// xterm 256 色表（简化版，完整版有 256 项）
var xterm256 = []uint32{
    0x000000, 0x800000, 0x008000, 0x808000, 0x000080, 0x800080, 0x008080, 0xc0c0c0,
    0x808080, 0xff0000, 0x00ff00, 0xffff00, 0x0000ff, 0xff00ff, 0x00ffff, 0xffffff,
    // ... 更多颜色
}
```

### 5. Mock Painter（测试用）

```go
// 位于: tui/runtime/paint/mock_painter.go

package paint

import (
    "strings"
)

// MockPainter 模拟绘图器（用于测试）
type MockPainter struct {
    buffer     *CellBuffer
    operations []string  // 记录所有操作
}

// NewMockPainter 创建模拟绘图器
func NewMockPainter(width, height int) *MockPainter {
    return &MockPainter{
        buffer:     NewCellBuffer(width, height),
        operations: make([]string, 0),
    }
}

// SetCell 设置单元格
func (m *MockPainter) SetCell(x, y int, char rune, style Style) {
    m.buffer.SetCell(x, y, char, style)
    m.operations = append(m.operations, fmt.Sprintf("SetCell(%d,%d,%c)", x, y, char))
}

// ... 其他方法类似

// GetOperations 获取操作记录
func (m *MockPainter) GetOperations() []string {
    return m.operations
}

// String 获取缓冲区内容的字符串表示
func (m *MockPainter) String() string {
    var sb strings.Builder
    width, height := m.buffer.Size()

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            cell := m.buffer.GetCell(x, y)
            sb.WriteRune(cell.Char)
        }
        sb.WriteString("\n")
    }

    return sb.String()
}

// AssertCell 断言单元格内容
func (m *MockPainter) AssertCell(t *testing.T, x, y int, char rune) {
    cell := m.buffer.GetCell(x, y)
    if cell.Char != char {
        t.Errorf("Expected cell (%d,%d) to be '%c', got '%c'", x, y, char, cell.Char)
    }
}

// AssertText 断言文本内容
func (m *MockPainter) AssertText(t *testing.T, expected string) {
    actual := m.String()
    if actual != expected {
        t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, actual)
    }
}
```

## 与框架集成

```go
// 位于: tui/runtime/runtime.go

package runtime

type Runtime struct {
    // ...
    painter  paint.Painter
}

// SetPainter 设置绘图器
func (r *Runtime) SetPainter(painter paint.Painter) {
    r.painter = painter
}

// GetPainter 获取绘图器
func (r *Runtime) GetPainter() paint.Painter {
    return r.painter
}

// Render 渲染组件树
func (r *Runtime) Render() error {
    // 清空缓冲区
    r.painter.Clear(paint.DefaultStyle())

    // 渲染组件树
    for _, node := range r.tree.Nodes {
        if paintable, ok := node.(component.Paintable); ok {
            ctx := PaintContext{
                Runtime: r,
                Theme:   r.theme,
            }
            paintable.Paint(ctx, r.painter)
        }
    }

    // 刷新到显示设备
    return r.painter.Flush()
}
```

## 使用示例

### 示例 1：终端渲染

```go
// ✅ 使用终端绘图器
func main() {
    painter := paint.NewTerminalPainter(os.Stdout, 80, 24)

    runtime := runtime.New()
    runtime.SetPainter(painter)

    // ... 添加组件

    runtime.Run()
}
```

### 示例 2：HTML 渲染

```go
// ✅ 使用 HTML 绘图器
func main() {
    painter := paint.NewHTMLPainter("tui-container", 80, 24)

    runtime := runtime.New()
    runtime.SetPainter(painter)

    // ... 添加组件

    runtime.Run()

    // 输出 HTML
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="/tui.css">
</head>
<body>
    <div id="tui-container"></div>
    <script src="/tui.js"></script>
</body>
</html>`)
    })
    http.ListenAndServe(":8080", nil)
}
```

### 示例 3：测试组件

```go
// ✅ 使用 Mock Painter 测试
func TestTextComponent(t *testing.T) {
    painter := paint.NewMockPainter(80, 24)

    text := component.NewText("Hello, World!")
    text.Paint(ctx, painter)

    // 断言输出
    painter.AssertText(t, "Hello, World!")
}
```

## 性能优化

### 1. Diff 渲染

```go
// 只更新变化的单元格
func (p *TerminalPainter) Flush() error {
    changes := p.buffer.Diff(p.prev)

    for _, change := range changes {
        p.drawCell(change.X, change.Y, change.Cell)
    }

    p.prev = p.buffer.Snapshot()
    return nil
}
```

### 2. 批量更新

```go
// 批量收集更新，一次性输出
func (p *TerminalPainter) Flush() error {
    var updates []string

    for _, change := range p.buffer.Diff(p.prev) {
        updates = append(updates, p.formatUpdate(change))
    }

    // 一次性输出
    fmt.Fprint(p.output, strings.Join(updates, ""))
    return nil
}
```

### 3. 虚拟滚动

```go
// 只渲染可见区域
func (l *VirtualList) Paint(ctx PaintContext, buf *CellBuffer) {
    start, end := l.viewport.GetVisibleRange()

    for i := start; i < end; i++ {
        item := l.dataSource.Get(i)
        renderItem(buf, i-start, item)
    }
}
```

## 总结

Painter 抽象系统提供：

1. **平台无关**: 组件代码不依赖具体平台
2. **多后端支持**: 终端、HTML、原生窗口
3. **双缓冲**: 支持离屏渲染和 Diff
4. **易测试**: Mock Painter 支持单元测试
5. **高性能**: Diff 渲染、批量更新

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [RENDERING.md](./RENDERING.md) - 渲染系统
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [THEME_SYSTEM.md](./THEME_SYSTEM.md) - 主题系统
- [VIRTUAL_SCROLL.md](./VIRTUAL_SCROLL.md) - 虚拟滚动
