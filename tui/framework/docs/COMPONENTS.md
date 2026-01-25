# Component Interface Design V3

> **版本说明**: 本文档定义了重构后的 Component 接口系统。V3 采用 Capability Interfaces 模式，将 Component 拆分为多个小型、可组合的能力接口。

## 设计原则

1. **最小接口**: 每个接口只定义一种核心能力
2. **可组合**: 通过接口组合实现复杂组件
3. **无强制**: 组件只实现需要的能力
4. **可替换**: 接口为替换而设计

## 为什么拆分 Component？

### ❌ V1: 胖接口

```go
type Component interface {
    ID() string
    Mount(parent Container) error
    Unmount() error
    Measure(constraints Constraints) Size
    Render(ctx RenderContext) string  // 返回 string!
    HandleEvent(ev Event) bool
    IsFocused() bool
    IsEnabled() bool
    IsVisible() bool
}
```

**问题**：
- 所有组件必须实现所有方法
- 大量 no-op 实现
- Render 返回 string 无法支持 diff
- 直接处理 KeyEvent 导致无法回放

### ✅ V3: 能力接口

```go
// 基础
type Node interface {
    ID() string
    Type() string
}

// 能力（按需实现）
type Paintable interface {
    Node
    Paint(ctx PaintContext, buf *runtime.CellBuffer)
}

type ActionTarget interface {
    Node
    HandleAction(a *runtime.Action) bool
}

type Focusable interface {
    Node
    FocusID() string
    OnFocus()
    OnBlur()
}

// 组合接口
type BaseComponent interface {
    Node
    Paintable
}
```

**优势**：
- 无 no-op 实现
- 组件只实现需要的能力
- AI 可以查询组件能力
- 易于测试和 Mock

## 核心能力接口

### 1. Node - 基础节点

```go
// 位于: tui/framework/component/node.go

package component

// Node 所有组件的基础接口
type Node interface {
    // ID 返回组件唯一标识
    ID() string

    // Type 返回组件类型
    Type() string
}

// NewNode 创建基础节点
func NewNode(id, componentType string) Node {
    return &baseNode{
        id:   id,
        typ:  componentType,
    }
}

type baseNode struct {
    id  string
    typ string
}

func (n *baseNode) ID() string { return n.id }
func (n *baseNode) Type() string { return n.typ }
```

### 2. Mountable - 可挂载

```go
// 位于: tui/framework/component/mountable.go

package component

// Mountable 可挂载组件
type Mountable interface {
    Node

    // Mount 挂载到父组件
    Mount(parent Container) error

    // Unmount 卸载
    Unmount() error

    // IsMounted 是否已挂载
    IsMounted() bool
}

// Container 容器接口
type Container interface {
    Node
    Add(child Node) error
    Remove(id string) error
    Children() []Node
}

// MountableState 可挂载状态
type MountableState struct {
    mounted bool
    parent  Container
}

func (s *MountableState) IsMounted() bool {
    return s.mounted
}
```

### 3. Measurable - 可测量

```go
// 位于: tui/framework/component/measurable.go

package component

// Constraints 尺寸约束
type Constraints struct {
    MinWidth  int
    MaxWidth  int
    MinHeight int
    MaxHeight int
}

// Size 尺寸
type Size struct {
    Width  int
    Height int
}

// Measurable 可测量组件
type Measurable interface {
    Node

    // Measure 测量首选尺寸
    Measure(constraints Constraints) Size

    // GetSize 获取当前尺寸
    GetSize() Size
}

// MeasurableState 可测量状态
type MeasurableState struct {
    width  int
    height int
}

func (s *MeasurableState) Width() int  { return s.width }
func (s *MeasurableState) Height() int { return s.height }

func (s *MeasurableState) SetSize(w, h int) {
    s.width = w
    s.height = h
}
```

### 4. Paintable - 可绘制（V3 关键变更）

```go
// 位于: tui/framework/component/paintable.go

package component

import (
    "github.com/yaoapp/yao/tui/runtime/paint"
)

// PaintContext 绘制上下文
type PaintContext struct {
    // 可用区域
    Bounds Rect

    // 绝对位置
    AbsoluteX int
    AbsoluteY int

    // 继承样式
    InheritStyle Style

    // Z-index
    ZIndex int

    // 裁剪区域
    ClipRect *Rect
}

// Paintable 可绘制组件
type Paintable interface {
    Node

    // Paint 绘制到缓冲区
    // V3: 不返回 string，而是直接写入 buffer
    Paint(ctx PaintContext, buf *paint.CellBuffer)
}

// PaintableState 可绘制状态
type PaintableState struct {
    visible bool
    opacity float32
}

func (s *PaintableState) IsVisible() bool {
    return s.visible
}
```

### 5. ActionTarget - Action 处理（V3 关键变更）

```go
// 位于: tui/framework/component/actionable.go

package component

import (
    "github.com/yaoapp/yao/tui/runtime/action"
)

// ActionTarget 可处理 Action 的组件
type ActionTarget interface {
    Node

    // HandleAction 处理语义化动作
    // V3: 不处理 KeyEvent，只处理 Action
    HandleAction(a *action.Action) bool
}

// 示例: TextInput 实现
func (t *TextInput) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionInputText:
        if text, ok := a.Payload.(string); ok {
            t.Insert(text)
            return true
        }
    case action.ActionDeleteChar:
        t.DeleteChar()
        return true
    case action.ActionNavigateLeft:
        t.MoveCursor(-1)
        return true
    }
    return false
}
```

### 6. Focusable - 可聚焦（V3 关键变更）

```go
// 位于: tui/framework/component/focusable.go

package component

// Focusable 可聚焦组件
type Focusable interface {
    Node

    // FocusID 返回焦点标识
    FocusID() string

    // OnFocus 获得焦点
    OnFocus()

    // OnBlur 失去焦点
    OnBlur()
}

// FocusableState 焦点状态
type FocusableState struct {
    focused bool
}

func (s *FocusableState) IsFocused() bool {
    return s.focused
}

func (s *FocusableState) SetFocus(focused bool) {
    s.focused = focused
}
```

### 7. Scrollable - 可滚动

```go
// 位于: tui/framework/component/scrollable.go

package component

// Scrollable 可滚动组件
type Scrollable interface {
    Node

    // ScrollTo 滚动到指定位置
    ScrollTo(x, y int)

    // ScrollBy 相对滚动
    ScrollBy(dx, dy int)

    // GetScrollPosition 获取滚动位置
    GetScrollPosition() (x, y int)

    // GetScrollRange 获取滚动范围
    GetScrollRange() (minX, minY, maxX, maxY int)
}

// ScrollableState 滚动状态
type ScrollableState struct {
    offsetX       int
    offsetY       int
    contentWidth  int
    contentHeight int
}
```

### 8. Validatable - 可验证

```go
// 位于: tui/framework/component/validatable.go

package component

// Validatable 可验证组件
type Validatable interface {
    Node

    // Validate 验证
    Validate() error

    // IsValid 是否有效
    IsValid() bool

    // ValidationMessage 验证消息
    ValidationMessage() string
}
```

### 9. Updatable - 可更新

```go
// 位于: tui/framework/component/updatable.go

package component

// Updatable 可更新组件
type Updatable interface {
    Node

    // GetValue 获取值
    GetValue() any

    // SetValue 设置值
    SetValue(value any) error

    // OnChange 值变化回调
    OnChange(fn func(any))
}
```

## 组合接口

### BaseComponent - 基础组件

```go
// 位于: tui/framework/component/base.go

package component

// BaseComponent 基础组件（组合常用能力）
type BaseComponent interface {
    Node
    Paintable
}

// 大多数组件应该实现 BaseComponent
```

### InteractiveComponent - 交互组件

```go
// 位于: tui/framework/component/interactive.go

package component

// InteractiveComponent 交互组件
type InteractiveComponent interface {
    BaseComponent
    ActionTarget
    Focusable
}

// Button, Input 等交互组件应该实现 InteractiveComponent
```

### ContainerComponent - 容器组件

```go
// 位于: tui/framework/component/container.go

package component

// ContainerComponent 容器组件
type ContainerComponent interface {
    BaseComponent

    // Add 添加子组件
    Add(child Node) error

    // Remove 移除子组件
    Remove(id string) error

    // Children 获取子组件
    Children() []Node

    // ChildCount 子组件数量
    ChildCount() int

    // Child 获取子组件
    Child(id string) (Node, bool)
}
```

## 接口断言和类型检查

```go
// 位于: tui/framework/component/traits.go

package component

// IsPaintable 检查是否可绘制
func IsPaintable(n Node) bool {
    _, ok := n.(Paintable)
    return ok
}

// IsFocusable 检查是否可聚焦
func IsFocusable(n Node) bool {
    _, ok := n.(Focusable)
    return ok
}

// IsActionTarget 检查是否可处理 Action
func IsActionTarget(n Node) bool {
    _, ok := n.(ActionTarget)
    return ok
}

// ToPaintable 转换为 Paintable
func ToPaintable(n Node) (Paintable, bool) {
    p, ok := n.(Paintable)
    return p, ok
}

// ToFocusable 转换为 Focusable
func ToFocusable(n Node) (Focusable, bool) {
    f, ok := n.(Focusable)
    return f, ok
}

// GetCapabilities 获取组件所有能力
func GetCapabilities(n Node) []string {
    var caps []string

    if _, ok := n.(Paintable); ok {
        caps = append(caps, "Paintable")
    }
    if _, ok := n.(ActionTarget); ok {
        caps = append(caps, "ActionTarget")
    }
    if _, ok := n.(Focusable); ok {
        caps = append(caps, "Focusable")
    }
    if _, ok := n.(Scrollable); ok {
        caps = append(caps, "Scrollable")
    }
    if _, ok := n.(Validatable); ok {
        caps = append(caps, "Validatable")
    }

    return caps
}
```

## 组件实现示例

### 示例 1: Text 组件（最小化）

```go
// 位于: tui/framework/display/text.go

package display

type Text struct {
    *component.MountableState
    *component.MeasurableState
    content string
    style   style.Style
}

func NewText(id, content string) *Text {
    return &Text{
        MountableState:   &component.MountableState{},
        MeasurableState: &component.MeasurableState{},
        content:         content,
    }
}

func (t *Text) ID() string   { return t.MountableState.ID }
func (t *Text) Type() string { return "text" }

func (t *Text) Measure(constraints component.Constraints) component.Size {
    // 测量文本尺寸
    width := runewidth.StringWidth(t.content)
    return component.Size{
        Width:  min(width, constraints.MaxWidth),
        Height: 1,
    }
}

func (t *Text) Paint(ctx component.PaintContext, buf *runtime.CellBuffer) {
    // 绘制到 buffer
    for i, r := range t.content {
        buf.SetCell(ctx.AbsoluteX+i, ctx.AbsoluteY, r, t.style)
    }
}
```

### 示例 2: Button 组件（交互）

```go
// 位于: tui/framework/interactive/button.go

package interactive

type Button struct {
    *component.MountableState
    *component.MeasurableState
    *component.FocusableState
    label    string
    onClick  func()
}

func NewButton(id, label string) *Button {
    return &Button{
        MountableState:   &component.MountableState{},
        MeasurableState: &component.MeasurableState{},
        FocusableState:  &component.FocusableState{},
        label:          label,
    }
}

func (b *Button) ID() string   { return b.MountableState.ID }
func (b *Button) Type() string { return "button" }

func (b *Button) FocusID() string { return b.ID() }

func (b *Button) OnFocus() {
    b.FocusableState.SetFocus(true)
}

func (b *Button) OnBlur() {
    b.FocusableState.SetFocus(false)
}

func (b *Button) HandleAction(a *runtime.Action) bool {
    switch a.Type {
    case runtime.ActionSubmit:
        if b.onClick != nil {
            b.onClick()
        }
        return true
    }
    return false
}

func (b *Button) Paint(ctx component.PaintContext, buf *runtime.CellBuffer) {
    // 根据 focused 状态绘制不同样式
    style := b.style
    if b.FocusableState.IsFocused() {
        style = style.WithBorder(style.BorderFocused)
    }
    // ...
}
```

## V2 → V3 变更对照

| 方面 | V2 | V3 |
|------|----|----|
| Component | 单一胖接口 | 能力接口拆分 |
| Render | `Render() string` | `Paint(ctx, buf)` |
| Event | `HandleEvent(Event)` | `HandleAction(Action)` |
| Focus | `IsFocused() bool` | `FocusID() string` |
| 可选能力 | 强制实现所有方法 | 只实现需要的能力 |
| AI 查询 | 无法查询组件能力 | `GetCapabilities()` |

## 总结

V3 Component 接口设计：

1. **9 个能力接口**: Node, Mountable, Measurable, Paintable, ActionTarget, Focusable, Scrollable, Validatable, Updatable
2. **3 个组合接口**: BaseComponent, InteractiveComponent, ContainerComponent
3. **Paint 不返回 string**: 直接写入 CellBuffer
4. **只处理 Action**: 不直接处理 KeyEvent
5. **无 no-op 实现**: 组件只实现需要的能力
6. **完全可测试**: 纯接口，易于 mock
7. **AI 友好**: 可以查询组件能力
