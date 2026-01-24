# Rendering System Design

## 概述

渲染系统负责将组件绘制到终端屏幕。本文档详细描述了渲染管线、缓冲区管理和差分渲染。

## 渲染管线

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Rendering Pipeline                               │
└─────────────────────────────────────────────────────────────────────────┘

Component.Render()
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 1: Component Render              │
│  - 每个组件生成自己的内容                 │
│  - 返回样式化的文本或 CellBuffer         │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 2: Layout Composition            │
│  - Runtime 布局引擎计算位置              │
│  - 生成 LayoutBox 层级结构               │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 3: Buffer Composition            │
│  - 将组件内容合并到 CellBuffer           │
│  - 应用 Z-order                         │
│  - 处理重叠                             │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 4: Diff Computation              │
│  - 比较前后缓冲区                        │
│  - 生成变更列表                         │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 5: Terminal Output               │
│  - 只输出变化的区域                      │
│  - 优化 ANSI 序列                       │
└─────────────────────────────────────────┘
```

## 核心类型

### 1. 缓冲区

```go
// 位于: tui/framework/screen/buffer.go

package screen

// Buffer 渲染缓冲区
type Buffer struct {
    // 尺寸
    width  int
    height int

    // 单元格数据
    cells  [][]Cell

    // 元数据
    dirty   bool
    version int
}

// Cell 单元格
type Cell struct {
    // 内容
    Char       rune
    StyledText string  // ANSI 样式文本

    // 样式
    Style      Style

    // 元数据
    ZIndex     int
    Selected   bool
    Modified   bool  // 标记是否被修改
}

// NewBuffer 创建缓冲区
func NewBuffer(width, height int) *Buffer {
    b := &Buffer{
        width:  width,
        height: height,
        cells:  make([][]Cell, height),
    }

    for y := 0; y < height; y++ {
        b.cells[y] = make([]Cell, width)
        for x := 0; x < width; x++ {
            b.cells[y][x] = Cell{
                Char:  ' ',
                Style: DefaultStyle(),
            }
        }
    }

    return b
}

// SetCell 设置单元格
func (b *Buffer) SetCell(x, y int, char rune, style Style) {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return
    }

    cell := &b.cells[y][x]
    cell.Char = char
    cell.Style = style
    cell.Modified = true
    b.dirty = true
}

// SetStyledText 设置样式化文本
func (b *Buffer) SetStyledText(x, y int, text string, style Style) {
    if y < 0 || y >= b.height || x < 0 {
        return
    }

    // 解析 ANSI 序列，分割文本和样式
    runes := []rune(text)
    pos := x

    for i := 0; i < len(runes) && pos < b.width; {
        // 处理 ANSI 转义序列
        if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
            // 跳过 ANSI 序列
            end := i + 2
            for end < len(runes) && runes[end] != 'm' {
                end++
            }
            if end < len(runes) {
                end++  // 跳过 'm'
            }

            // 解析样式
            ansiStyle := string(runes[i:end])
            style = style.Merge(ParseANSIStyle(ansiStyle))
            i = end
        } else {
            // 普通字符
            b.SetCell(pos, y, runes[i], style)
            pos++
            i++
        }
    }
}

// SetLine 设置一行文本
func (b *Buffer) SetLine(y int, text string, style Style) {
    runes := []rune(text)
    for x, r := range runes {
        if x >= b.width {
            break
        }
        b.SetCell(x, y, r, style)
    }
}

// Fill 填充区域
func (b *Buffer) Fill(x, y, width, height int, char rune, style Style) {
    for py := y; py < y+height && py < b.height; py++ {
        for px := x; px < x+width && px < b.width; px++ {
            b.SetCell(px, py, char, style)
        }
    }
}

// Clear 清空缓冲区
func (b *Buffer) Clear() {
    for y := 0; y < b.height; y++ {
        for x := 0; x < b.width; x++ {
            b.cells[y][x] = Cell{
                Char:  ' ',
                Style: DefaultStyle(),
            }
        }
    }
    b.dirty = false
}

// GetSize 获取缓冲区尺寸
func (b *Buffer) GetSize() (width, height int) {
    return b.width, b.height
}

// GetCell 获取单元格
func (b *Buffer) GetCell(x, y int) Cell {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return Cell{}
    }
    return b.cells[y][x]
}

// Clone 克隆缓冲区
func (b *Buffer) Clone() *Buffer {
    clone := NewBuffer(b.width, b.height)
    for y := 0; y < b.height; y++ {
        copy(clone.cells[y], b.cells[y])
    }
    clone.dirty = b.dirty
    clone.version = b.version + 1
    return clone
}
```

### 2. 差分引擎

```go
// 位于: tui/framework/screen/diff.go

package screen

// DiffChange 变更记录
type DiffChange struct {
    X     int
    Y     int
    Old   Cell
    New   Cell
}

// DiffEngine 差分引擎
type DiffEngine struct {
    // 优化配置
    mergeChanges bool  // 合并相邻变更
    maxBatchSize int   // 最大批量大小
}

// NewDiffEngine 创建差分引擎
func NewDiffEngine() *DiffEngine {
    return &DiffEngine{
        mergeChanges: true,
        maxBatchSize: 1000,
    }
}

// Diff 计算两个缓冲区的差异
func (e *DiffEngine) Diff(old, new *Buffer) []DiffChange {
    if old.width != new.width || old.height != new.height {
        // 尺寸不同，全量更新
        return e.fullDiff(new)
    }

    var changes []DiffChange

    for y := 0; y < new.height; y++ {
        for x := 0; x < new.width; x++ {
            oldCell := old.cells[y][x]
            newCell := new.cells[y][x]

            if !e.cellsEqual(oldCell, newCell) {
                changes = append(changes, DiffChange{
                    X:   x,
                    Y:   y,
                    Old: oldCell,
                    New: newCell,
                })
            }
        }
    }

    if e.mergeChanges {
        changes = e.mergeAdjacentChanges(changes)
    }

    return changes
}

// cellsEqual 比较单元格是否相等
func (e *DiffEngine) cellsEqual(a, b Cell) bool {
    return a.Char == b.Char &&
           a.Style == b.Style &&
           a.Selected == b.Selected
}

// fullDiff 全量差异
func (e *DiffEngine) fullDiff(buf *Buffer) []DiffChange {
    changes := make([]DiffChange, 0, buf.width*buf.height)

    for y := 0; y < buf.height; y++ {
        for x := 0; x < buf.width; x++ {
            cell := buf.cells[y][x]
            changes = append(changes, DiffChange{
                X:   x,
                Y:   y,
                New: cell,
            })
        }
    }

    return changes
}

// mergeAdjacentChanges 合并相邻的变更
func (e *DiffEngine) mergeAdjacentChanges(changes []DiffChange) []DiffChange {
    if len(changes) == 0 {
        return changes
    }

    // 按行分组
    byLine := make(map[int][]DiffChange)
    for _, c := range changes {
        byLine[c.Y] = append(byLine[c.Y], c)
    }

    var merged []DiffChange

    for y, lineChanges := range byLine {
        // 按X排序
        sort.Slice(lineChanges, func(i, j int) bool {
            return lineChanges[i].X < lineChanges[j].X
        })

        // 合并相邻的相同样式变更
        i := 0
        for i < len(lineChanges) {
            start := i
            startStyle := lineChanges[i].New.Style

            for i < len(lineChanges) &&
                lineChanges[i].New.Style == startStyle &&
                (i == start || lineChanges[i].X == lineChanges[i-1].X+1) {
                i++
            }

            // 检查是否可以合并为范围输出
            if i-start > 5 {  // 超过5个连续字符，使用范围
                merged = append(merged, DiffChange{
                    Y:   y,
                    X:   lineChanges[start].X,
                    New: lineChanges[start].New,
                })
            } else {
                // 单独输出
                for j := start; j < i; j++ {
                    merged = append(merged, lineChanges[j])
                }
            }
        }
    }

    return merged
}
```

### 3. 绘制器

```go
// 位于: tui/framework/screen/painter.go

package screen

// Painter 绘制器
type Painter struct {
    terminal platform.Terminal
    buffer   *Buffer
}

// NewPainter 创建绘制器
func NewPainter(terminal platform.Terminal) *Painter {
    return &Painter{
        terminal: terminal,
    }
}

// SetBuffer 设置当前缓冲区
func (p *Painter) SetBuffer(buf *Buffer) {
    p.buffer = buf
}

// Draw 绘制缓冲区到终端
func (p *Painter) Draw(changes []DiffChange) error {
    // 保存光标位置
    p.saveCursor()

    for _, change := range changes {
        // 移动光标
        p.moveCursor(change.X, change.Y)

        // 设置样式
        p.applyStyle(change.New.Style)

        // 绘制字符
        p.drawRune(change.New.Char)
    }

    // 重置样式
    p.resetStyle()

    // 恢复光标位置
    p.restoreCursor()

    // 刷新输出
    return p.terminal.Flush()
}

// DrawFull 完整绘制
func (p *Painter) DrawFull(buf *Buffer) error {
    p.buffer = buf

    // 清屏
    p.clearScreen()

    // 逐行绘制
    for y := 0; y < buf.height; y++ {
        p.moveCursor(0, y)
        p.drawRow(y)
    }

    return p.terminal.Flush()
}

// drawRow 绘制一行
func (p *Painter) drawRow(y int) {
    if y < 0 || y >= p.buffer.height {
        return
    }

    var currentStyle Style

    for x := 0; x < p.buffer.width; x++ {
        cell := p.buffer.cells[y][x]

        // 样式变化时输出 ANSI 序列
        if cell.Style != currentStyle {
            p.applyStyle(cell.Style)
            currentStyle = cell.Style
        }

        p.drawRune(cell.Char)
    }
}

// saveCursor 保存光标位置
func (p *Painter) saveCursor() {
    p.terminal.WriteString("\x1b[s")
}

// restoreCursor 恢复光标位置
func (p *Painter) restoreCursor() {
    p.terminal.WriteString("\x1b[u")
}

// moveCursor 移动光标
func (p *Painter) moveCursor(x, y int) {
    // 使用 CSI H: ESC [ <row> ; <col> H
    // 注意: 终端坐标从 1 开始
    p.terminal.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
}

// applyStyle 应用样式
func (p *Painter) applyStyle(style Style) {
    var codes []string

    // 前景色
    if style.FG.Color != "" {
        codes = append(codes, style.FG.toANSI(false))
    }

    // 背景色
    if style.BG.Color != "" {
        codes = append(codes, style.BG.toANSI(true))
    }

    // 粗体
    if style.Bold {
        codes = append(codes, "1")
    }

    // 斜体
    if style.Italic {
        codes = append(codes, "3")
    }

    // 下划线
    if style.Underline {
        codes = append(codes, "4")
    }

    // 反白
    if style.Reverse {
        codes = append(codes, "7")
    }

    if len(codes) > 0 {
        p.terminal.WriteString("\x1b[" + strings.Join(codes, ";") + "m")
    } else {
        p.resetStyle()
    }
}

// resetStyle 重置样式
func (p *Painter) resetStyle() {
    p.terminal.WriteString("\x1b[0m")
}

// drawRune 绘制字符
func (p *Painter) drawRune(r rune) {
    p.terminal.WriteString(string(r))
}

// clearScreen 清屏
func (p *Painter) clearScreen() {
    p.terminal.WriteString("\x1b[2J")
    p.moveCursor(0, 0)
}

// clearLine 清除行
func (p *Painter) clearLine(y int) {
    p.moveCursor(0, y)
    p.terminal.WriteString("\x1b[2K")
}
```

### 4. 屏幕管理器

```go
// 位于: tui/framework/screen/manager.go

package screen

// Manager 屏幕管理器
type Manager struct {
    terminal platform.Terminal

    // 双缓冲
    front    *Buffer  // 当前显示的缓冲区
    back     *Buffer  // 后台缓冲区

    // 差分引擎
    diff     *DiffEngine

    // 绘制器
    painter  *Painter

    // 光标
    cursor   Cursor
    cursorVisible bool
}

// NewManager 创建屏幕管理器
func NewManager(terminal platform.Terminal) *Manager {
    return &Manager{
        terminal: terminal,
        diff:     NewDiffEngine(),
        painter:  NewPainter(terminal),
    }
}

// Init 初始化屏幕
func (m *Manager) Init() error {
    // 获取终端尺寸
    width, height, err := m.terminal.GetSize()
    if err != nil {
        return err
    }

    // 创建缓冲区
    m.front = NewBuffer(width, height)
    m.back = NewBuffer(width, height)

    // 进入备用屏幕
    m.terminal.WriteString("\x1b[?1049h")

    // 启用原始模式
    m.terminal.EnableRawMode()

    // 隐藏光标
    m.hideCursor()

    // 清屏
    m.clearScreen()

    return nil
}

// Close 关闭屏幕
func (m *Manager) Close() error {
    // 显示光标
    m.showCursor()

    // 退出原始模式
    m.terminal.DisableRawMode()

    // 退出备用屏幕
    m.terminal.WriteString("\x1b[?1049l")

    return nil
}

// Render 渲染缓冲区
func (m *Manager) Render(buf *Buffer) error {
    // 确保尺寸匹配
    if buf.width != m.back.width || buf.height != m.back.height {
        m.resize(buf.width, buf.height)
        m.front = NewBuffer(buf.width, buf.height)
    }

    // 计算差异
    changes := m.diff.Diff(m.front, buf)

    // 应用变更
    m.painter.SetBuffer(buf)
    if err := m.painter.Draw(changes); err != nil {
        return err
    }

    // 更新前缓冲
    m.front = buf

    return nil
}

// resize 调整尺寸
func (m *Manager) resize(width, height int) {
    m.back = NewBuffer(width, height)
}

// clearScreen 清屏
func (m *Manager) clearScreen() {
    m.terminal.WriteString("\x1b[2J")
    m.moveCursor(0, 0)
}

// moveCursor 移动光标
func (m *Manager) moveCursor(x, y int) {
    m.terminal.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
}

// hideCursor 隐藏光标
func (m *Manager) hideCursor() {
    m.terminal.WriteString("\x1b[?25l")
    m.cursorVisible = false
}

// showCursor 显示光标
func (m *Manager) showCursor() {
    m.terminal.WriteString("\x1b[?25h")
    m.cursorVisible = true
}

// SetCursor 设置光标位置
func (m *Manager) SetCursor(x, y int) {
    m.cursor.X = x
    m.cursor.Y = y

    if m.cursorVisible {
        m.moveCursor(x, y)
    }
}

// GetCursor 获取光标位置
func (m *Manager) GetCursor() (x, y int) {
    return m.cursor.X, m.cursor.Y
}

// SetCursorVisible 设置光标可见性
func (m *Manager) SetCursorVisible(visible bool) {
    if visible != m.cursorVisible {
        if visible {
            m.showCursor()
        } else {
            m.hideCursor()
        }
        m.cursorVisible = visible
    }
}

// GetSize 获取屏幕尺寸
func (m *Manager) GetSize() (width, height int) {
    return m.terminal.GetSize()
}

// Cursor 光标结构
type Cursor struct {
    X int
    Y int
}
```

## 组件渲染

### 渲染上下文

```go
// 位于: tui/framework/component/render.go

package component

// RenderContext 渲染上下文
type RenderContext struct {
    // 可用尺寸
    AvailableWidth  int
    AvailableHeight int

    // 组件位置 (相对于父组件)
    X int
    Y int

    // 滚动偏移
    OffsetX int
    OffsetY int

    // 继承的样式
    InheritStyle Style

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

// Contains 检查点是否在矩形内
func (r *Rect) Contains(x, y int) bool {
    return x >= r.X && x < r.X+r.Width &&
           y >= r.Y && y < r.Y+r.Height
}

// Intersect 计算两个矩形的交集
func (r *Rect) Intersect(other *Rect) *Rect {
    x1 := max(r.X, other.X)
    y1 := max(r.Y, other.Y)
    x2 := min(r.X+r.Width, other.X+other.Width)
    y2 := min(r.Y+r.Height, other.Y+other.Height)

    if x1 >= x2 || y1 >= y2 {
        return nil  // 无交集
    }

    return &Rect{
        X:      x1,
        Y:      y1,
        Width:  x2 - x1,
        Height: y2 - y1,
    }
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
```

### 基础渲染器

```go
// 位于: tui/framework/component/renderer.go

package component

// Renderer 组件渲染器
type Renderer struct {
    buffer *screen.Buffer
}

// NewRenderer 创建渲染器
func NewRenderer(buffer *screen.Buffer) *Renderer {
    return &Renderer{buffer: buffer}
}

// Render 渲染组件
func (r *Renderer) Render(comp Component, ctx *RenderContext) {
    // 设置组件尺寸
    comp.SetSize(ctx.AvailableWidth, ctx.AvailableHeight)

    // 获取渲染内容
    content := comp.Render(ctx)

    // 绘制到缓冲区
    r.drawContent(content, ctx)
}

// drawContent 绘制内容
func (r *Renderer) drawContent(content string, ctx *RenderContext) {
    lines := strings.Split(content, "\n")

    for y, line := range lines {
        if y >= ctx.AvailableHeight {
            break
        }
        r.buffer.SetStyledText(ctx.X, ctx.Y+y, line, ctx.InheritStyle)
    }
}

// RenderContainer 渲染容器组件
func (r *Renderer) RenderContainer(container Container, ctx *RenderContext) {
    children := container.GetChildren()

    // 布局计算
    layout := container.GetLayout()
    if layout != nil {
        layout.Layout(container, ctx.X, ctx.Y, ctx.AvailableWidth, ctx.AvailableHeight)
    }

    // 渲染子组件
    for _, child := range children {
        if !child.IsVisible() {
            continue
        }

        // 获取子组件位置
        childX := child.GetX()
        childY := child.GetY()
        childW, childH := child.GetSize()

        // 创建子组件上下文
        childCtx := ctx.WithOffset(childX, childY)
        childCtx.AvailableWidth = childW
        childCtx.AvailableHeight = childH

        // 递归渲染
        r.Render(child, childCtx)
    }
}
```

## 渲染优化

### 1. 脏区域标记

```go
// 位于: tui/framework/screen/dirty.go

package screen

// DirtyRegion 脏区域
type DirtyRegion struct {
    X      int
    Y      int
    Width  int
    Height int
}

// DirtyTracker 脏区域跟踪器
type DirtyTracker struct {
    regions []DirtyRegion
    enabled bool
}

// NewDirtyTracker 创建脏区域跟踪器
func NewDirtyTracker() *DirtyTracker {
    return &DirtyTracker{
        regions: make([]DirtyRegion, 0),
        enabled: true,
    }
}

// Enable 启用脏区域跟踪
func (t *DirtyTracker) Enable() {
    t.enabled = true
}

// Disable 禁用脏区域跟踪
func (t *DirtyTracker) Disable() {
    t.enabled = false
}

// MarkDirty 标记脏区域
func (t *DirtyTracker) MarkDirty(x, y, width, height int) {
    if !t.enabled {
        return
    }

    region := DirtyRegion{
        X:      x,
        Y:      y,
        Width:  width,
        Height: height,
    }

    // 尝试与现有区域合并
    for i, existing := range t.regions {
        if t.mergeable(&existing, &region) {
            t.regions[i] = t.merge(&existing, &region)
            return
        }
    }

    // 添加新区域
    t.regions = append(t.regions, region)
}

// mergeable 检查是否可合并
func (t *DirtyTracker) mergeable(a, b *DirtyRegion) bool {
    // 检查是否相邻或重叠
    return !(a.X+a.Width < b.X || b.X+b.Width < a.X ||
             a.Y+a.Height < b.Y || b.Y+b.Height < a.Y)
}

// merge 合并区域
func (t *DirtyTracker) merge(a, b *DirtyRegion) DirtyRegion {
    x1 := min(a.X, b.X)
    y1 := min(a.Y, b.Y)
    x2 := max(a.X+a.Width, b.X+b.Width)
    y2 := max(a.Y+a.Height, b.Y+b.Height)

    return DirtyRegion{
        X:      x1,
        Y:      y1,
        Width:  x2 - x1,
        Height: y2 - y1,
    }
}

// GetDirtyRegions 获取脏区域
func (t *DirtyTracker) GetDirtyRegions() []DirtyRegion {
    return t.regions
}

// Clear 清除脏区域标记
func (t *DirtyTracker) Clear() {
    t.regions = t.regions[:0]
}

// IsDirty 检查是否有脏区域
func (t *DirtyTracker) IsDirty() bool {
    return len(t.regions) > 0
}
```

### 2. 部分渲染

```go
// 位于: tui/framework/screen/partial.go

package screen

// PartialRenderer 部分渲染器
type PartialRenderer struct {
    buffer *Buffer
    tracker *DirtyTracker
}

// NewPartialRenderer 创建部分渲染器
func NewPartialRenderer(buffer *Buffer) *PartialRenderer {
    return &PartialRenderer{
        buffer:  buffer,
        tracker: NewDirtyTracker(),
    }
}

// RenderComponent 渲染组件到指定区域
func (r *PartialRenderer) RenderComponent(comp Component, x, y, width, height int) {
    // 标记脏区域
    r.tracker.MarkDirty(x, y, width, height)

    // 渲染组件
    ctx := NewRenderContext(width, height)
    ctx.X = x
    ctx.Y = y

    renderer := NewRenderer(r.buffer)
    renderer.Render(comp, ctx)
}

// GetChanges 获取变更
func (r *PartialRenderer) GetChanges(old *Buffer) []DiffChange {
    diff := NewDiffEngine()

    if len(r.tracker.GetDirtyRegions()) == 0 {
        // 无脏区域，无变更
        return []DiffChange{}
    }

    // 只比较脏区域
    changes := make([]DiffChange, 0)

    for _, region := range r.tracker.GetDirtyRegions() {
        for y := region.Y; y < region.Y+region.Height && y < r.buffer.height; y++ {
            for x := region.X; x < region.X+region.Width && x < r.buffer.width; x++ {
                oldCell := old.GetCell(x, y)
                newCell := r.buffer.GetCell(x, y)

                if !cellsEqual(oldCell, newCell) {
                    changes = append(changes, DiffChange{
                        X:   x,
                        Y:   y,
                        Old: oldCell,
                        New: newCell,
                    })
                }
            }
        }
    }

    return changes
}

// ClearDirty 清除脏标记
func (r *PartialRenderer) ClearDirty() {
    r.tracker.Clear()
}
```

## ANSI 转义码

### 常用转义码

```go
// 位于: tui/framework/screen/ansi.go

package screen

// ANSICodes ANSI 转义码常量
const (
    // 光标控制
    CursorHome          = "\x1b[H"
    CursorUp            = "\x1b[A"
    CursorDown          = "\x1b[B"
    CursorForward       = "\x1b[C"
    CursorBack          = "\x1b[D"
    CursorSave          = "\x1b[s"
    CursorRestore       = "\x1b[u"
    CursorHide          = "\x1b[?25l"
    CursorShow          = "\x1b[?25h"

    // 屏幕控制
    ScreenClear         = "\x1b[2J"
    ScreenErase         = "\x1b[3J"
    LineClear           = "\x1b[2K"
    LineErase           = "\x1b[3K"

    // 滚动
    ScrollUp            = "\x1b[%dS"
    ScrollDown          = "\x1b[%dT"

    // 备用屏幕
    AltScreenEnter      = "\x1b[?1049h"
    AltScreenExit       = "\x1b[?1049l"

    // 样式重置
    ResetAll            = "\x1b[0m"

    // 颜色
    ResetFG             = "\x1b[39m"
    ResetBG             = "\x1b[49m"

    // 粗体
    BoldOn              = "\x1b[1m"
    BoldOff             = "\x1b[22m"

    // 斜体
    ItalicOn            = "\x1b[3m"
    ItalicOff           = "\x1b[23m"

    // 下划线
    UnderlineOn         = "\x1b[4m"
    UnderlineOff        = "\x1b[24m"

    // 反白
    ReverseOn           = "\x1b[7m"
    ReverseOff          = "\x1b[27m"

    // 闪烁
    BlinkOn             = "\x1b[5m"
    BlinkOff            = "\x1b[25m"
)

// ColorCodes 256色代码
var ColorCodes = map[string]int{
    // 标准色
    "black":   0,
    "red":     1,
    "green":   2,
    "yellow":  3,
    "blue":    4,
    "magenta": 5,
    "cyan":    6,
    "white":   7,

    // 高亮色
    "bright-black":   8,
    "bright-red":     9,
    "bright-green":   10,
    "bright-yellow":  11,
    "bright-blue":    12,
    "bright-magenta": 13,
    "bright-cyan":    14,
    "bright-white":   15,
}

// FormatColor 格式化颜色代码
func FormatColor(color string, bg bool) string {
    if code, ok := ColorCodes[color]; ok {
        if bg {
            return fmt.Sprintf("48;5;%d", code)
        }
        return fmt.Sprintf("38;5;%d", code)
    }

    // 尝试解析为数字
    if n, err := strconv.Atoi(color); err == nil && n >= 0 && n <= 255 {
        if bg {
            return fmt.Sprintf("48;5;%d", n)
        }
        return fmt.Sprintf("38;5;%d", n)
    }

    return ""
}

// FormatRGB 格式化 RGB 颜色
func FormatRGB(r, g, b int, bg bool) string {
    if bg {
        return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
    }
    return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
}
```

## 渲染示例

```go
// 完整的渲染流程示例

func ExampleRender() {
    // 1. 创建屏幕管理器
    terminal := platform.NewDefaultTerminal()
    screen := screen.NewManager(terminal)
    screen.Init()
    defer screen.Close()

    // 2. 创建组件树
    root := layout.NewFlex().
        WithDirection(layout.Column).
        WithChildren(
            display.NewText("Hello, World!"),
            input.NewTextInput(),
        )

    // 3. 创建渲染缓冲区
    width, height := screen.GetSize()
    buffer := screen.NewBuffer(width, height)

    // 4. 渲染组件
    renderer := component.NewRenderer(buffer)
    ctx := component.NewRenderContext(width, height)
    renderer.Render(root, ctx)

    // 5. 输出到屏幕
    screen.Render(buffer)
}
```
