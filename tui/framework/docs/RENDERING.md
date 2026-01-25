# Rendering System Design (V3)

> **版本说明**: 本文档定义了 TUI 框架的渲染系统。V3 采用审查建议，引入 RenderTree 中间态、Dirty Region 主动标记、按需重绘机制。

## 概述

渲染系统负责将组件绘制到终端屏幕。V3 设计核心原则：

1. **Paint 不返回 string**: 直接写入 CellBuffer
2. **RenderTree 中间态**: Component → RenderNode → Paint
3. **Dirty Region 主动标记**: 不依赖 Cell diff
4. **局部重绘**: 只渲染变化的区域
5. **幂等性**: 相同输入必定产生相同输出

## 渲染管线（V3）

```
Component.Paint(ctx, buf)
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 1: Paint (Component Side)       │
│  - 组件生成 RenderNode 描述             │
│  - 不直接操作 Buffer                    │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 2: Build RenderTree            │
│  - Runtime 组合 RenderNode 树          │
│  - 应用 Z-order                        │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 3: Layout (Runtime Side)       │
│  - 计算每个节点的 Bounds               │
│  - 应用 Flexbox 约束                    │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 4: Paint to Buffer              │
│  - RenderTree → CellBuffer              │
│  - 应用 Z-order 和裁剪                   │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 5: Dirty Region Diff           │
│  - 只比较 Dirty Region                 │
│  - 生成最小变更集                       │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 6: Terminal Output              │
│  - 输出 ANSI 转义码                      │
└─────────────────────────────────────────┘
```

## 核心类型定义

### 1. Cell - 基础单元格

```go
// 位于: tui/runtime/paint/cell.go

package paint

// Cell 单元格
type Cell struct {
    // 内容
    Char rune

    // 样式
    Style CellStyle

    // 元数据
    ZIndex    int
    Modified  bool  // 是否被修改（用于 diff）
}

// CellStyle 单元格样式
type CellStyle struct {
    FG       Color
    BG       Color
    Bold     bool
    Dim      bool
    Italic   bool
    Underline bool
    Blink    bool
    Reverse  bool
}

// Color 颜色
type Color struct {
    Type  ColorType
    Value uint8
}

type ColorType int

const (
    ColorDefault ColorType = iota
    ColorBasic
    Color256
    ColorRGB
)

// Clear 清空单元格
func (c *Cell) Clear() {
    c.Char = ' '
    c.Style = CellStyle{}
    c.Modified = false
}

// Clone 克隆单元格
func (c *Cell) Clone() Cell {
    return Cell{
        Char:     c.Char,
        Style:    c.Style,
        ZIndex:   c.ZIndex,
        Modified: c.Modified,
    }
}
```

### 2. CellBuffer - 虚拟画布

```go
// 位于: tui/runtime/paint/buffer.go

package paint

// CellBuffer 虚拟画布
type CellBuffer struct {
    cells  [][]Cell
    width  int
    height int
}

// NewBuffer 创建缓冲区
func NewBuffer(width, height int) *CellBuffer {
    b := &CellBuffer{
        width:  width,
        height: height,
        cells:  make([][]Cell, height),
    }

    for y := 0; y < height; y++ {
        b.cells[y] = make([]Cell, width)
        for x := 0; x < width; x++ {
            b.cells[y][x] = Cell{Char: ' '}
        }
    }

    return b
}

// SetCell 设置单元格
func (b *CellBuffer) SetCell(x, y int, char rune, style CellStyle) {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return
    }

    cell := &b.cells[y][x]
    cell.Char = char
    cell.Style = style
    cell.Modified = true
}

// SetCellV 设置带变体的单元格
func (b *CellBuffer) SetCellV(x, y int, char rune, variant CellVariant) {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return
    }

    cell := &b.cells[y][x]
    cell.Char = char
    cell.Modified = true

    switch variant {
    case VariantBold:
        cell.Style.Bold = true
    case VariantDim:
        cell.Style.Dim = true
    case VariantUnderline:
        cell.Style.Underline = true
    }
}

// GetCell 获取单元格
func (b *CellBuffer) GetCell(x, y int) Cell {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return Cell{}
    }
    return b.cells[y][x]
}

// Size 获取缓冲区尺寸
func (b *CellBuffer) Size() (width, height int) {
    return b.width, b.height
}

// Resize 调整尺寸
func (b *CellBuffer) Resize(width, height int) {
    newCells := make([][]Cell, height)
    for y := 0; y < height; y++ {
        newCells[y] = make([]Cell, width)
        if y < b.height {
            copyX := min(width, b.width)
            for x := 0; x < copyX; x++ {
                newCells[y][x] = b.cells[y][x]
            }
        }
    }
    b.cells = newCells
    b.width = width
    b.height = height
}

// Clear 清空缓冲区
func (b *CellBuffer) Clear() {
    for y := 0; y < b.height; y++ {
        for x := 0; x < b.width; x++ {
            b.cells[y][x] = Cell{Char: ' '}
        }
    }
}

// Fill 填充区域
func (b *CellBuffer) Fill(rect Rect, char rune, style CellStyle) {
    for y := rect.Y; y < rect.Y+rect.Height && y < b.height; y++ {
        for x := rect.X; x < rect.X+rect.Width && x < b.width; x++ {
            b.SetCell(x, y, char, style)
        }
    }
}

// Clone 克隆缓冲区
func (b *CellBuffer) Clone() *CellBuffer {
    clone := NewBuffer(b.width, b.height)
    for y := 0; y < b.height; y++ {
        copy(clone.cells[y], b.cells[y])
    }
    return clone
}

// CellVariant 单元格变体
type CellVariant int

const (
    VariantNormal CellVariant = iota
    VariantBold
    VariantDim
    VariantItalic
    VariantUnderline
    VariantBlink
    VariantReverse
)
```

### 3. RenderNode - 渲染节点

```go
// 位于: tui/runtime/paint/node.go

package paint

// RenderNode 渲染节点
type RenderNode struct {
    // 标识
    ID string

    // 位置和尺寸（由 Layout 计算）
    Bounds Rect

    // Z-index
    Z int

    // 绘制函数（由 Component 提供）
    PaintFunc func(ctx PaintContext, buf *CellBuffer)

    // 子节点
    Children []*RenderNode

    // 样式
    Style Style

    // 可见性
    Visible bool

    // 裁剪区域
    ClipRect *Rect
}

// Rect 矩形区域
type Rect struct {
    X, Y, Width, Height int
}

// Empty 检查是否为空
func (r Rect) Empty() bool {
    return r.Width <= 0 || r.Height <= 0
}

// Contains 检查点是否在矩形内
func (r Rect) Contains(x, y int) bool {
    return x >= r.X && x < r.X+r.Width &&
           y >= r.Y && y < r.Y+r.Height
}

// Intersect 计算交集
func (r Rect) Intersect(other Rect) Rect {
    x1 := max(r.X, other.X)
    y1 := max(r.Y, other.Y)
    x2 := min(r.X+r.Width, other.X+other.Width)
    y2 := min(r.Y+r.Height, other.Y+other.Height)

    if x1 >= x2 || y1 >= y2 {
        return Rect{} // 无交集
    }

    return Rect{
        X:      x1,
        Y:      y1,
        Width:  x2 - x1,
        Height: y2 - y1,
    }
}

// Union 计算并集
func (r Rect) Union(other Rect) Rect {
    if r.Empty() {
        return other
    }
    if other.Empty() {
        return r
    }

    x1 := min(r.X, other.X)
    y1 := min(r.Y, other.Y)
    x2 := max(r.X+r.Width, other.X+other.Width)
    y2 := max(r.Y+r.Height, other.Y+other.Height)

    return Rect{
        X:      x1,
        Y:      y1,
        Width:  x2 - x1,
        Height: y2 - y1,
    }
}

// NewRenderNode 创建渲染节点
func NewRenderNode(id string) *RenderNode {
    return &RenderNode{
        ID:       id,
        Children: make([]*RenderNode, 0),
        Visible:  true,
    }
}

// AddChild 添加子节点
func (n *RenderNode) AddChild(child *RenderNode) {
    n.Children = append(n.Children, child)
}

// Find 查找节点
func (n *RenderNode) Find(id string) *RenderNode {
    if n.ID == id {
        return n
    }
    for _, child := range n.Children {
        if found := child.Find(id); found != nil {
            return found
        }
    }
    return nil
}

// Render 渲染到缓冲区
func (n *RenderNode) Render(buf *CellBuffer, offsetX, offsetY int) {
    if !n.Visible {
        return
    }

    ctx := PaintContext{
        Bounds:    n.Bounds,
        AbsoluteX: offsetX + n.Bounds.X,
        AbsoluteY: offsetY + n.Bounds.Y,
        ZIndex:    n.Z,
        ClipRect:  n.ClipRect,
    }

    // 绘制自身
    if n.PaintFunc != nil {
        n.PaintFunc(ctx, buf)
    }

    // 绘制子节点
    for _, child := range n.Children {
        child.Render(buf, offsetX+n.Bounds.X, offsetY+n.Bounds.Y)
    }
}

// CalculateBounds 计算所有节点的边界
func (n *RenderNode) CalculateBounds() {
    // 递归计算子节点边界
    for _, child := range n.Children {
        child.CalculateBounds()
    }

    // 根据子节点调整自身边界
    // （由 Layout Engine 完成）
}
```

### 4. DirtyRegion - 脏区域追踪

```go
// 位于: tui/runtime/paint/dirty.go

package paint

// DirtyRegion 脏区域
type DirtyRegion struct {
    X, Y, Width, Height int
}

// DirtyTracker 脏区域追踪器
type DirtyTracker struct {
    regions []DirtyRegion
    enabled bool
}

// NewDirtyTracker 创建脏区域追踪器
func NewDirtyTracker() *DirtyTracker {
    return &DirtyTracker{
        regions: make([]DirtyRegion, 0),
        enabled: true,
    }
}

// Enable 启用脏区域追踪
func (t *DirtyTracker) Enable() {
    t.enabled = true
}

// Disable 禁用脏区域追踪
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

// MarkComponent 标记组件为脏
func (t *DirtyTracker) MarkComponent(node *RenderNode) {
    if node == nil || !node.Visible {
        return
    }
    t.MarkDirty(
        node.Bounds.X,
        node.Bounds.Y,
        node.Bounds.Width,
        node.Bounds.Height,
    )
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

// HasAny 是否有任何脏区域
func (t *DirtyTracker) HasAny() bool {
    return len(t.regions) > 0
}

// MarkAll 标记全屏为脏
func (t *DirtyTracker) MarkAll(width, height int) {
    t.regions = []DirtyRegion{
        {X: 0, Y: 0, Width: width, Height: height},
    }
}
```

### 5. DiffEngine - 差分引擎

```go
// 位于: tui/runtime/paint/diff.go

package paint

// DiffChange 变更记录
type DiffChange struct {
    X, Y int
    Old  Cell
    New  Cell
}

// DiffEngine 差分引擎
type DiffEngine struct {
    // 只比较 Dirty Region
    dirtyOnly bool
}

// NewDiffEngine 创建差分引擎
func NewDiffEngine() *DiffEngine {
    return &DiffEngine{
        dirtyOnly: true,
    }
}

// Diff 计算两个缓冲区的差异
func (e *DiffEngine) Diff(old, new *CellBuffer) []DiffChange {
    if old.width != new.width || old.height != new.height {
        // 尺寸不同，全量更新
        return e.fullDiff(new)
    }

    var changes []DiffChange

    // 如果启用 dirtyOnly，检查是否有脏区域
    if e.dirtyOnly {
        // TODO: 从 DirtyTracker 获取脏区域
        // 这里简化处理，全量比较
    }

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

    return changes
}

// cellsEqual 比较单元格是否相等
func (e *DiffEngine) cellsEqual(a, b Cell) bool {
    return a.Char == b.Char && a.Style == b.Style
}

// fullDiff 全量差异
func (e *DiffEngine) fullDiff(buf *CellBuffer) []DiffChange {
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

// Optimize 优化变更列表
func (e *DiffEngine) Optimize(changes []DiffChange) []DiffChange {
    if len(changes) == 0 {
        return changes
    }

    // 按行分组
    byLine := make(map[int][]DiffChange)
    for _, c := range changes {
        byLine[c.Y] = append(byLine[c.Y], c)
    }

    var optimized []DiffChange

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
            if i-start > 3 {
                // 使用范围
                optimized = append(optimized, DiffChange{
                    Y:   y,
                    X:   lineChanges[start].X,
                    New: lineChanges[start].New,
                    Count: i - start, // 新增字段
                })
            } else {
                // 单独输出
                for j := start; j < i; j++ {
                    optimized = append(optimized, lineChanges[j])
                }
            }
        }
    }

    return optimized
}
```

## PaintContext - 绘制上下文

```go
// 位于: tui/runtime/paint/context.go

package paint

// PaintContext 绘制上下文
type PaintContext struct {
    // 绘制区域（相对于父组件）
    Bounds Rect

    // 绝对屏幕坐标
    AbsoluteX int
    AbsoluteY int

    // 继承样式
    InheritStyle Style

    // Z-index
    ZIndex int

    // 裁剪区域
    ClipRect *Rect

    // 元数据
    Metadata map[string]interface{}
}

// WithOffset 创建带偏移的上下文
func (c PaintContext) WithOffset(dx, dy int) PaintContext {
    return PaintContext{
        Bounds:       c.Bounds,
        AbsoluteX:    c.AbsoluteX + dx,
        AbsoluteY:    c.AbsoluteY + dy,
        InheritStyle: c.InheritStyle,
        ZIndex:       c.ZIndex,
        ClipRect:     c.ClipRect,
        Metadata:     c.Metadata,
    }
}

// WithClip 创建带裁剪的上下文
func (c PaintContext) WithClip(rect Rect) PaintContext {
    clip := rect
    if c.ClipRect != nil {
        clip = c.ClipRect.Intersect(rect)
    }
    return PaintContext{
        Bounds:       c.Bounds,
        AbsoluteX:    c.AbsoluteX,
        AbsoluteY:    c.AbsoluteY,
        InheritStyle: c.InheritStyle,
        ZIndex:       c.ZIndex,
        ClipRect:     &clip,
        Metadata:     c.Metadata,
    }
}

// IsClipped 检查点是否被裁剪
func (c PaintContext) IsClipped(x, y int) bool {
    if c.ClipRect == nil {
        return false
    }
    return !c.ClipRect.Contains(x, y)
}
```

## 输出接口

### TerminalOutput - 终端输出

```go
// 位于: tui/runtime/paint/output.go

package paint

// TerminalOutput 终端输出
type TerminalOutput interface {
    // Write 输出变更
    Write(changes []DiffChange) error

    // Flush 刷新输出
    Flush() error

    // Clear 清屏
    Clear() error
}

// ANSIOutput ANSI 转义码输出
type ANSIOutput struct {
    writer io.Writer
    cursor *CursorState
}

// Write 写入变更
func (o *ANSIOutput) Write(changes []DiffChange) error {
    for _, change := range changes {
        // 移动光标
        o.moveCursor(change.X, change.Y)

        // 设置样式
        o.applyStyle(change.New.Style)

        // 绘制字符
        o.writeRune(change.New.Char)
    }

    // 重置样式
    o.resetStyle()

    return nil
}

// Flush 刷新
func (o *ANSIOutput) Flush() error {
    if f, ok := o.writer.(interface{ Flush() error }); ok {
        return f.Flush()
    }
    return nil
}
```

## V2 → V3 主要变更

| 方面 | V2 | V3 |
|------|----|----|
| Render | 返回 `string` | `Paint(ctx, buf)` 直接写入 |
| 中间态 | 无 | 引入 `RenderNode` |
| Dirty | Cell diff | 主动标记 `DirtyRegion` |
| Diff | 全量比较 | 只比较 Dirty 区域 |
| 性能 | 可能有全量 repaint | 局部重绘 |
| 幂等性 | 未明确 | 明确要求幂等性 |

## 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构总览
- [COMPONENTS.md](COMPONENTS.md) - Component 接口
- [BOUNDARIES.md](BOUNDARIES.md) - 层级边界
