# Yao TUI 系统开发手册

## 版本信息

- **版本**: v3.0
- **更新日期**: 2025-01
- **适用范围**: `tui/framework` 和 `tui/runtime`

---

## 目录

1. [系统架构](#系统架构)
2. [功能列表](#功能列表)
3. [开发规则](#开发规则)
4. [接口参考](#接口参考)
5. [限制条件](#限制条件)
6. [正确示例](#正确示例)
7. [错误示例](#错误示例)
8. [附录](#附录)

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Application Layer                          │
│                    (用户代码、DSL、业务逻辑)                              │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Framework Layer                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                    │
│  │ Component │  │   Event  │  │   Form   │  │  Theme   │                    │
│  │          │  │          │  │          │  │          │                    │
│  │  Binding  │  │  System  │  │          │  │          │                    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘                    │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Runtime Layer                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                    │
│  │  Action  │  │  Focus   │  │  Layout  │  │  State   │                    │
│  │          │  │          │  │  Engine  │  │  Tracker │                    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                    │
│  │  Paint   │  │  Input   │  │ Platform│  │  Paint   │                    │
│  │  Engine  │  │          │  │          │  │  Buffer  │                    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘                    │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Platform Layer                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                                    │
│  │ Terminal │  │  Console │  │  Signal  │                                    │
│  │  Driver  │  │          │  │          │                                    │
│  └──────────┘  └──────────┘  └──────────┘                                    │
└─────────────────────────────────────────────────────────────────────┘
```

### 架构边界与依赖规则

| 层级 | 依赖 | 禁止 |
|------|------|------|
| **Framework** | Runtime, Platform | 终端 I/O, Bubble Tea |
| **Runtime** | Platform | Framework, 组件类型, Bubble Tea |
| **Platform** | Go 标准库, OS 驱动 | Runtime, Framework, 组件 |

### 核心设计原则

1. **关注点分离**: Framework 处理组件抽象，Runtime 处理布局引擎
2. **依赖倒置**: 高层依赖低层接口，而非反向
3. **单一职责**: 每个模块只负责一个明确的功能
4. **开闭原则**: 对扩展开放，对修改封闭

---

## 功能列表

### Framework Layer 功能

| 模块 | 功能 | 说明 |
|------|------|------|
| **component/** | 组件基础架构 | Node 接口、能力接口、BaseComponent |
| **component/binding/** | 数据绑定 | Prop[T]、Scope、响应式存储 |
| **event/** | 事件系统 | Event 定义、事件处理器、事件分发 |
| **form/** | 表单系统 | Form 容器、字段验证、数据收集 |
| **input/** | 输入组件 | TextInput、光标管理 |
| **layout/** | 布局组件 | Box、Flex 容器 |
| **paint/** | 绘制抽象 | Painter 工具、绘制上下文 |
| **theme/** | 主题系统 | 样式定义、主题管理、颜色系统 |
| **display/** | 显示组件 | List、Table、Text 显示 |
| **validation/** | 验证系统 | Validator 接口、内置验证器 |

### Runtime Layer 功能

| 模块 | 功能 | 说明 |
|------|------|------|
| **action/** | Action 系统 | Action 定义、分发器、组合 Action |
| **layout/** | 布局引擎 | Flexbox、约束系统、布局缓存 |
| **paint/** | 绘制引擎 | CellBuffer、样式应用、脏区域跟踪 |
| **focus/** | 焦点管理 | 焦点域、焦点导航、焦点陷阱 |
| **input/** | 输入处理 | 键盘映射、鼠标处理、输入转换 |
| **platform/** | 平台抽象 | 终端驱动、信号处理、平台检测 |
| **state/** | 状态管理 | Snapshot、Tracker、Diff、序列化 |
| **animation/** | 动画系统 | Easing、动画管理、缓动函数 |
| **event/** | 事件处理 | 命中测试、三阶段分发 |
| **render/** | 渲染优化 | 节流、缓存、批量渲染 |

---

## 开发规则

### 规则 1: 组件必须实现最小接口

所有组件必须实现 `Node` 接口：

```go
type Node interface {
    ID() string           // 组件唯一标识
    Type() string          // 组件类型名称
}
```

**❌ 错误示例**:
```go
type MyComponent struct {
    name string
}

// 缺少 ID() 和 Type() 方法
```

**✅ 正确示例**:
```go
type MyComponent struct {
    base *component.BaseComponent
}

func (c *MyComponent) ID() string {
    return c.base.ID()
}

func (c *MyComponent) Type() string {
    return "mycomponent"
}
```

### 规则 2: 使用组合而非继承

组件应该组合多个能力接口，而不是深层继承。

**❌ 错误示例**:
```go
// 过深的继承链
type BaseClickable struct {}
type BaseButton struct {
    BaseClickable
}
type MyButton struct {
    BaseButton
}
```

**✅ 正确示例**:
```go
// 组合多个能力接口
type MyComponent struct {
    *component.BaseComponent
}

// 通过接口检查能力
func (c *MyComponent) HandleAction(a action.Action) bool {
    // 实现交互能力
}

func (c *MyComponent) Paint(ctx PaintContext, buf *paint.Buffer) {
    // 实现绘制能力
}
```

### 规则 3: Runtime 层禁止导入 Framework

Runtime 层（`tui/runtime/`）绝对不能导入 Framework 层包。

**❌ 禁止**:
```go
// runtime/layout/engine.go
import "github.com/yaoapp/yao/tui/framework/component"  // ❌
```

**✅ 允许**:
```go
// runtime/layout/engine.go
import (
    "github.com/yaoapp/yao/tui/runtime/node"     // ✅ runtime 包
    "github.com/yaoapp/yao/tui/runtime/constraint" // ✅ runtime 包
)
```

### 规则 4: 组件状态必须可枚举

组件的所有状态必须能被序列化和枚举。

**❌ 错误示例**:
```go
type MyComponent struct {
    mu    sync.Mutex
    cache map[string]interface{}  // 隐藏状态，无法枚举
}

func (c *MyComponent) GetHiddenState() interface{} {
    return c.cache  // 无法通过枚举发现
}
```

**✅ 正确示例**:
```go
type MyComponent struct {
    *component.StateHolder  // 所有状态通过 StateHolder 管理
}

// 状态可枚举
state := comp.GetState()
for key, value := range state {
    // 所有状态都可访问
}
```

### 规则 5: 渲染必须是幂等的

相同输入必须产生相同输出，Render 不能读取外部状态。

**❌ 错误示例**:
```go
func (c *MyComponent) Paint(ctx PaintContext, buf *paint.Buffer) {
    now := time.Now()  // ❌ 读取外部状态
    ctx.SetCell(0, 0, now.Format("15:04:05"), style)
}
```

**✅ 正确示例**:
```go
func (c *MyComponent) Paint(ctx PaintContext, buf *paint.Buffer) {
    // 只从 ctx 或组件状态获取数据
    text := c.GetText()
    ctx.SetCell(0, 0, text, style)
}
```

### 规则 6: 禁止在 Runtime 层判断组件类型

Runtime 层不知道组件的具体类型，只通过接口操作。

**❌ 错误示例**:
```go
// runtime/event/dispatch.go
func DispatchToComponent(comp interface{}) {
    if btn, ok := comp.(*input.Button); ok {  // ❌ 类型断言
        // 处理按钮逻辑
    }
}
```

**✅ 正确示例**:
```go
// runtime/event/dispatch.go
func DispatchToComponent(comp InteractiveComponent) {
    comp.HandleAction(action.ActionNavigate{})  // 通过接口调用
}
```

### 规则 7: 状态变化必须通过 Action

禁止直接修改状态，必须通过 Action 分发。

**❌ 错误示例**:
```go
func (c *MyComponent) UpdateValue() {
    c.value = newValue  // ❌ 直接修改
}
```

**✅ 正确示例**:
```go
func (c *MyComponent) HandleAction(a action.Action) bool {
    if a.Type == action.ActionUpdate {
        c.SetValue(a.Payload)  // 通过 Action 更新
        c.MarkDirty()
        return true
    }
    return false
}
```

### 规则 8: 组件必须线程安全

由于可能并发访问（如动画、输入），组件状态操作必须加锁。

```go
type MyComponent struct {
    mu sync.RWMutex
    value string
}

func (c *MyComponent) GetValue() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.value
}
```

---

## 接口参考

### 核心接口

#### Node 接口

所有组件必须实现的最小接口：

```go
type Node interface {
    ID() string           // 组件唯一标识符
    Type() string          // 组件类型名称
}
```

#### ComponentNode 组合接口

静态组件应该实现的组合接口：

```go
type ComponentNode interface {
    Node
    Positionable          // GetPosition(), SetPosition()
    Sizable              // GetSize(), SetSize()
    Measurable            // Measure()
}
```

#### InteractiveComponent 交互组件接口

交互组件（按钮、输入框等）应该实现：

```go
type InteractiveComponent interface {
    ComponentNode
    Focusable            // FocusID(), OnFocus(), OnBlur()
    ActionTarget         // HandleAction()
}
```

#### ContainerComponent 容器组件接口

容器组件（Box、Flex）应该实现：

```go
type ContainerComponent interface {
    Parent                // Children(), Add(), Remove()
    Measurable            // Measure()
    Paintable             // Paint()
}
```

### Framework 组件接口

#### Paintable 可绘制接口

```go
type Paintable interface {
    Node
    Paint(ctx PaintContext, buf *paint.Buffer)
}
```

#### Mountable 可挂载接口

```go
type Mountable interface {
    Node
    Mount(parent Container)
    Unmount()
    IsMounted() bool
}
```

#### Measurable 可测量接口

```go
type Measurable interface {
    Node
    Measure(maxWidth, maxHeight int) (width, height int)
}
```

#### Focusable 可聚焦接口

```go
type Focusable interface {
    Node
    FocusID() string
    OnFocus()
    OnBlur()
    IsFocused() bool
}
```

#### ActionTarget 动作目标接口

```go
type ActionTarget interface {
    Node
    HandleAction(a action.Action) bool
}
```

#### Validatable 可验证接口

```go
type Validatable interface {
    Node
    Validate() error
    IsValid() bool
}
```

### Runtime 层接口

#### Container 容器接口

```go
type Container interface {
    Node
    Children() []Node
    ChildCount() int
}
```

#### Parent 父节点接口

```go
type Parent interface {
    Node
    Children() []Node
}
```

#### Child 子节点接口

```go
type Child interface {
    Node
    // 可被添加/移除
}
```

---

## 限制条件

### 1. 架构边界限制

| 层级 | 可以使用 | 禁止使用 |
|------|----------|----------|
| **Framework** | Runtime 接口, Platform 接口 | Platform 驱动, 终端 I/O |
| **Runtime** | Platform 接口, Go 标准库 | Framework 包, 组件类型 |
| **Platform** | Go 标准库, OS 驱动 | Runtime, Framework, 组件 |

### 2. 接口使用限制

| 接口 | 可实现位置 | 说明 |
|------|------------|------|
| `Node` | Runtime | Framework 通过别名使用 |
| `Paintable` | Framework | 自定义绘制行为 |
| `Measurable` | Runtime | 组件尺寸计算 |
| `Focusable` | Runtime | 焦点管理 |
| `ActionTarget` | Framework | 处理语义动作 |
| `Container` | Framework | 容器管理 |
| `Mountable` | Framework | 组件挂载 |

### 3. 方法调用限制

| 操作 | 允许的上下文 | 禁止的上下文 |
|------|-------------|-------------|
| 绘制 (`Paint()`) | 可调用 Buffer 方法, 上下文方法 | 禁止读取外部状态 |
| 测量 (`Measure()`) | 只读取约束参数 | 禁止读取组件状态 |
| 事件处理 (`HandleAction()`) | 可访问组件状态 | 禁止修改其他组件状态 |
| 状态修改 | 只能在 HandleAction 中 | 禁止在 Paint 中修改 |

### 4. 状态管理限制

```go
// ✅ 允许：通过 StateHolder 管理状态
comp.SetStateValue("key", "value")
value := comp.GetStateValue("key")

// ✅ 允许：通过 ReactiveStore 管理全局状态
store.Set("global.key", "value")

// ❌ 禁止：闭包中的隐藏状态
func (c *Component) GetHiddenValue() string {
    return "hardcoded"  // 无法被枚举或序列化
}
```

### 5. 渲染限制

```go
// ❌ 禁止：读取外部状态
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    now := time.Now()  // ❌ 非幂等
}

// ❌ 禁止：副作用操作
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    fmt.Println("painting...")  // ❌ 副作用
}

// ❌ 禁止：修改全局状态
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    globalCounter++  // ❌ 副作用
}

// ✅ 允许：使用上下文和组件状态
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    text := c.GetText()
    width := ctx.AvailableWidth
    // 渲染...
}
```

### 6. 并发限制

```go
// 组件状态访问必须加锁
type MyComponent struct {
    mu sync.RWMutex
    value string
}

// 读操作使用读锁
func (c *MyComponent) GetValue() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.value
}

// 写操作使用写锁
func (c *MyComponent) SetValue(value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value = value
}
```

---

## 正确示例

### 示例 1: 创建静态组件

```go
package display

import (
    "github.com/yaoapp/yao/tui/framework/component"
    "github.com/yaoapp/yao/tui/runtime/paint"
)

type Text struct {
    *component.BaseComponent
    *component.StateHolder
    content string
}

func NewText(content string) *Text {
    return &Text{
        BaseComponent: component.NewBaseComponent("text"),
        StateHolder:   component.NewStateHolder(),
        content:      content,
    }
}

// 实现 Node 接口
func (t *Text) ID() string {
    return t.BaseComponent.ID()
}

func (t *Text) Type() string {
    return t.BaseComponent.Type()
}

// 实现 Paintable 接口
func (t *Text) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if !t.IsVisible() {
        return
    }

    // 绘制文本
    for i, r := range t.content {
        ctx.SetCell(ctx.X+i, ctx.Y, r)
    }
}

// 实现 Measurable 接口
func (t *Text) Measure(maxWidth, maxHeight int) (width, height int) {
    textLen := len([]rune(t.content))
    width = textLen
    height = 1

    if maxWidth > 0 && width > maxWidth {
        width = maxWidth
    }
    if maxHeight > 0 && height > maxHeight {
        height = maxHeight
    }
    return width, height
}
```

### 示例 2: 创建交互组件

```go
package interactive

import (
    "github.com/yaoapp/yao/tui/framework/component"
    "github.com/yaoapp/yao/tui/framework/event"
    "github.com/yaoapp/yao/tui/runtime/action"
    "github.com/yaoapp/yao/tui/runtime/paint"
)

type Button struct {
    *component.BaseComponent
    *component.StateHolder

    label     string
    disabled  bool
    onClick  func()
}

func NewButton(label string, onClick func()) *Button {
    return &Button{
        BaseComponent: component.NewBaseComponent("button"),
        StateHolder:   component.NewStateHolder(),
        label:        label,
        disabled:     false,
        onClick:      onClick,
    }
}

// 实现 Paintable
func (b *Button) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if !b.IsVisible() {
        return
    }

    disabled := b.IsDisabled()
    focused := b.IsFocused()

    // 绘制按钮边框和文本
    style := b.getButtonStyle(disabled, focused)
    text := b.label

    // 绘制边框
    ctx.DrawBox(0, 0, ctx.AvailableWidth, 1, style)

    // 绘制文本
    ctx.DrawText(0, 0, text, "center", style)
}

func (b *Button) getButtonStyle(disabled, focused bool) style.Style {
    // 根据状态返回样式
    // 从主题系统获取
    return style.GetStyle("button", "default")
}

// 实现 Measurable
func (b *Button) Measure(maxWidth, maxHeight int) (width, height int) {
    width = len(b.label) + 4  // 文本 + 边框
    height = 1

    if maxWidth > 0 && width > maxWidth {
        width = maxWidth
    }
    return
}

// 实现 Focusable
func (b *Button) FocusID() string {
    return b.ID()
}

func (b *Button) OnFocus() {
    b.BaseComponent.OnFocus()
}

func (b *Button) OnBlur() {
    b.BaseComponent.OnBlur()
}

// 实现 ActionTarget
func (b *Button) HandleAction(a action.Action) bool {
    if a.Type == action.ActionSelect && !b.IsDisabled() {
        if b.onClick != nil {
            b.onClick()
            return true
        }
    }
    return false
}

// 实现 Validatable
func (b *Button) Validate() error {
    if b.label == "" {
        return &ValidationError{"button", "label cannot be empty"}
    }
    return nil
}
```

### 示例 3: 创建容器组件

```go
package layout

import (
    "github.com/yaoapp/yao/tui/framework/component"
    "github.com/yaoapp/yao/tui/runtime/paint"
)

type VBox struct {
    *component.BaseComponent
    children []component.Node
    spacing int
}

func NewVBox() *VBox {
    return &VBox{
        BaseComponent: component.NewBaseComponent("vbox"),
        children:   make([]component.Node, 0),
        spacing:    0,
    }
}

// 实现 Container 接口
func (v *VBox) Add(child component.Node) {
    v.children = append(v.children, child)
    child.Mount(v)
}

func (v *VBox) Remove(child component.Node) {
    for i, c := range v.children {
        if c == child {
            v.children = append(v.children[:i], v.children[i+1:]...)
            child.Unmount()
            break
        }
    }
}

func (v *VBox) GetChildren() []component.Node {
    return v.children
}

func (v *VBox) ChildCount() int {
    return len(v.children)
}

// 实现 Measurable
func (v *VBox) Measure(maxWidth, maxHeight int) (width, height int) {
    if len(v.children) == 0 {
        return 0, 0
    }

    // 测量每个子组件
    childWidths := make([]int, len(v.children))
    childHeights := make([]int, len(v.children))

    maxWidth := 0
    totalHeight := 0

    for i, child := range v.children {
        if measurable, ok := child.(component.Measurable); ok {
            w, h := measurable.Measure(maxWidth, maxHeight)
            childWidths[i] = w
            childHeights[i] = h

            if w > maxWidth {
                maxWidth = w
            }
        } else {
            childWidths[i] = maxWidth
            childHeights[i] = 1
        }

        totalHeight += childHeights[i]
        if i > 0 {
            totalHeight += v.spacing
        }
    }

    return maxWidth, totalHeight
}

// 实现 Paintable
func (v *VBox) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if !v.IsVisible() {
        return
    }

    y := ctx.Y
    for _, child := range v.children {
        childHeight := 1
        if measurable, ok := child.(component.Measurable); ok {
            _, h := measurable.Measure(ctx.AvailableWidth, ctx.AvailableHeight)
            childHeight = h
        }

        // 创建子组件绘制上下文
        childCtx := component.PaintContext{
            PaintContext: ctx.PaintContext,
            X:             ctx.X,
            Y:             y,
            AvailableWidth: ctx.AvailableWidth,
            AvailableHeight: childHeight,
        }

        // 绘制子组件
        if paintable, ok := child.(component.Paintable); ok {
            paintable.Paint(childCtx, buf)
        }

        y += childHeight + v.spacing
    }
}
```

### 示例 4: 使用数据绑定

```go
package display

import (
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
    "github.com/yaoapp/yao/tui/framework/binding"
    "github.com/yaoapp/yao/tui/framework/component"
)

type BoundLabel struct {
    *cb.BaseBindable
    textProp binding.Prop[string]
}

func NewBoundLabel(textPath string) *BoundLabel {
    return &BoundLabel{
        BaseBindable: cb.NewBaseBindable("label"),
        textProp:     binding.NewBinding[string](textPath),
    }
}

func (l *BoundLabel) SetText(text string) *BoundLabel {
    l.textProp = binding.NewStringProp(text)
    return l
}

func (l *BoundLabel) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 创建绑定上下文
    bindCtx := cb.CreateBindingContext(l)

    // 解析属性值
    text := l.textProp.Resolve(bindCtx)

    // 绘制
    for i, r := range text {
        ctx.SetCell(ctx.X+i, ctx.Y, r)
    }
}
```

### 示例 5: 响应式组件

```go
package display

import (
    "github.com/yaoapp/yao/tui/framework/binding"
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
    "github.com/yaoapp/yao/tui/framework/component"
)

type ReactiveText struct {
    *cb.BaseBindable
    valueProp binding.Prop[string]
    store     *binding.ReactiveStore
    cancel    func()
}

func NewReactiveText(store *binding.ReactiveStore, path string) *ReactiveText {
    rt := &ReactiveText{
        BaseBindable: cb.NewBaseBindable("reactivetxt"),
        valueProp:    binding.NewBinding[string](path),
        store:        store,
    }

    // 订阅数据变化
    rt.cancel = store.Subscribe(path, func(key string, old, new interface{}) {
        rt.MarkDirty()
    })

    return rt
}

func (rt *ReactiveText) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 创建绑定上下文（包含 Store 数据）
    bindCtx := rt.store.ToContext()

    // 解析当前值
    text := rt.valueProp.Resolve(bindCtx)

    // 绘制
    for i, r := range text {
        ctx.SetCell(ctx.X+i, ctx.Y, r)
    }
}

func (rt *ReactiveText) Dispose() {
    if rt.cancel != nil {
        rt.cancel()
    }
}
```

---

## 错误示例

### 错误 1: 直接修改状态

```go
// ❌ 错误：在事件处理中直接修改状态
func (c *Component) OnClick() {
    c.value = "new value"  // ❌ 违反状态变化规则
}

// ✅ 正确：通过 Action 更新状态
func (c *Component) HandleAction(a action.Action) bool {
    if a.Type == action.ActionUpdate {
        c.SetValue(a.Payload)  // ✅ 通过 Action 更新
        c.MarkDirty()
        return true
    }
    return false
}
```

### 错误 2: 在 Render 中读取外部状态

```go
// ❌ 错误：Paint 读取外部状态导致非幂等
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    now := time.Now()  // ❌ 读取外部状态
    ctx.SetCell(0, 0, now.Format("15:04:05"), style)
}

// ✅ 正确：将动态数据作为参数传入
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    // 从组件状态获取动态数据
    if timeVal, ok := c.GetStateValue("currentTime"); ok {
        ctx.SetCell(0, 0, timeVal.(string), style)
    }
}
```

### 错误 3: Runtime 层导入 Framework

```go
// ❌ 错误：Runtime 层导入 Framework 包
package layout

import (
    "github.com/yaoapp/yao/tui/framework/component"  // ❌ 违反架构边界
)

func Layout(comp interface{}) LayoutBox {
    if comp, ok := comp.(*framework.Component) {  // ❌ 类型断言
        // ...
    }
}

// ✅ 正确：只使用 Runtime 接口
import (
    "github.com/yaoapp/yao/tui/runtime/node"
    "github.com/yaoapp/yao/tui/runtime/interfaces"
)

func Layout(node runtime.Node) LayoutBox {
    if children := node.Children(); len(children) > 0 {
        // 处理子节点
    }
}
```

### 错误 4: 组件状态不线程安全

```go
// ❌ 错误：无锁保护的并发访问
type MyComponent struct {
    value string  // ❌ 并发不安全
}

func (c *MyComponent) SetValue(v string) {
    c.value = v  // ❌ 并发写入不安全
}

// ✅ 正确：使用读写锁
type MyComponent struct {
    mu    sync.RWMutex
    value string
}

func (c *MyComponent) SetValue(v string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value = v
}

func (c *MyComponent) GetValue() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.value
}
```

### 错误 5: 忽略可见性检查

```go
// ❌ 错误：绘制不可见组件
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    // 忘略了可见性检查
    for i, r := range c.content {
        ctx.SetCell(ctx.X+i, ctx.Y, r)
    }
}

// ✅ 正确：先检查可见性
func (c *Component) Paint(ctx PaintContext, buf *paint.Buffer) {
    if !c.IsVisible() {
        return  // ✅ 早期返回
    }

    for i, r := range c.content {
        ctx.SetCell(ctx.X+i, ctx.Y, r)
    }
}
```

### 错误 6: 硬编码尺寸

```go
// ❌ 错误：硬编码宽度
func (c *Component) Measure(maxWidth, maxHeight int) (width, height int) {
    width = 80  // ❌ 硬编码
    height = 24
    return
}

// ✅ 正确：基于内容和约束计算
func (c *Component) Measure(maxWidth, maxHeight int) (width, height int) {
    contentLen := len([]rune(c.GetText()))
    width = contentLen + 2  // 内容 + 边框

    if maxWidth > 0 && width > maxWidth {
        width = maxWidth
    }

    height = 1
    if maxHeight > 0 && height > maxHeight {
        height = maxHeight
    }

    return width, height
}
```

### 错误 7: 忽略父容器约束

```go
// ❌ 错误：忽略父容器约束
func (c *Component) Measure(maxWidth, maxHeight int) (width, height int) {
    return 100, 50  // ❌ 忽略约束
}

// ✅ 正确：尊重约束
func (c *Component) Measure(maxWidth, maxHeight int) (width, height int) {
    preferredWidth := c.getPreferredWidth()

    width = minInt(preferredWidth, maxWidth)
    height = c.getPreferredHeight()

    return width, height
}
```

---

## 附录

### A. 完整组件生命周期

```
1. 创建阶段
   ├─ NewComponent() - 创建组件实例
   ├─ SetID() - 设置唯一标识
   └─ SetType() - 设置类型

2. 配置阶段
   ├─ SetProps() - 设置属性
   ├─ SetState() - 设置状态
   └─ SetStyle() - 设置样式

3. 挂载阶段
   ├─ Mount(parent) - 挂载到父容器
   └─ MountWithContext(parent, ctx) - 带上下文挂载

4. 测量阶段
   └─ Measure() - 计算理想尺寸

5. 布局阶段
   └─ Layout() - 计算实际位置和尺寸

6. 渲染阶段
   └─ Paint() - 绘制到虚拟画布

7. 更新阶段
   ├─ HandleAction() - 处理用户交互
   ├─ MarkDirty() - 标记需要重绘
   └─ Unmount() - 卸载组件
```

### B. 接口速查表

| 接口 | 方法 | 说明 | 实现位置 |
|------|------|------|------------|
| `Node` | `ID()`, `Type()` | 基础信息 | Runtime |
| `Positionable` | `GetPosition()`, `SetPosition()` | 位置设置 | Runtime |
| `Sizable` | `GetSize()`, `SetSize()` | 尺寸设置 | Runtime |
| `Located` | Node + Positionable + Sizable | 位置+尺寸 | Runtime |
| `Measurable` | `Measure()` | 尺寸计算 | Runtime |
| `Paintable` | `Paint()` | 绘制 | Framework |
| `Focusable` | `FocusID()`, `OnFocus()`, `OnBlur()` | 焦点 | Runtime |
| `ActionTarget` | `HandleAction()` | 动作处理 | Framework |
| `Mountable` | `Mount()`, `Unmount()` | 挂载/卸载 | Framework |
| `Container` | `Children()`, `Add()`, `Remove()` | 子组件 | Framework |
| `Validatable` | `Validate()`, `IsValid()` | 验证 | Framework |
| `Visible` | `IsVisible()`, `SetVisible()` | 可见性 | Runtime |

### C. Action 类型列表

```go
// 导入 Action 类型
import "github.com/yaoapp/yao/tui/runtime/action"

// 导入 Event 类型
import "github.com/yaoapp/yao/tui/framework/event"

// 常用 Action 类型
const (
    // 导航 Action
    ActionNavigate       action.ActionType = "navigate"
    ActionSubmit        action.ActionType = "submit"
    ActionCancel        action.ActionType = "cancel"

    // 输入 Action
    ActionInputChar     action.ActionType = "input.char"
    ActionInputText     action.ActionType = "input.text"
    ActionBackspace    action.ActionType = "backspace"
    ActionDelete        action.ActionType = "delete"
    ActionEnter        action.ActionType = "enter"

    // 光标 Action
    ActionCursorLeft    action.ActionType = "cursor.left"
    ActionCursorRight   action.ActionType = "cursor.right"
    ActionCursorHome    action.ActionType = "cursor.home"
    ActionCursorEnd      action.ActionType = "cursor.end"

    // 选择 Action
    ActionSelect         action.ActionType = "select"
    ActionNext          action.ActionType = "next"
    ActionPrev          action.ActionType = "prev"

    // 滚动 Action
    ActionScrollUp       action.ActionType = "scroll.up"
    ActionScrollDown     action.ActionType = "scroll.down"
    ActionScrollPageUp   action.ActionType = "scroll.page_up"
    ActionScrollPageDown action.ActionType = "scroll.page_down"

    // 系统 Action
    ActionQuit          action.ActionType = "quit"
    ActionRefresh       action.ActionType = "refresh"
    ActionResize        action.ActionType = "resize"
)
```

### D. 事件类型列表

```go
// 导入 Event 类型
import "github.com/yaoapp/yao/tui/framework/event"

const (
    // 事件类型
    EventTypeKey       event.EventType = "key"
    EventTypeMouse     event.EventType = "mouse"
    EventTypeResize    event.EventType = "resize"
    EventTypeFocus     event.EventType = "focus"
    EventTypeBlur      event.EventType = "blur"
    EventTypeSubmit    event.EventType = "submit"
)
```
