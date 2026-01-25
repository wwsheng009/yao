# Framework Boundaries and Interfaces (V3)

> **版本说明**: 本文档定义了 TUI 框架各层之间的边界和接口契约，确保模块间的清晰分离和可替换性。

## 概述

本文档定义了四层架构之间的边界和接口契约。这些边界是架构的"物理隔离线"，**违反这些边界的代码不应被接受**。

## 四层边界

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Boundary                             │
│  应用程序代码 - 用户创建的具体应用                                        │
│  依赖: Framework API                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    ↓ depends on
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Boundary                               │
│  tui/framework/* - 新框架代码                                            │
│  依赖: Runtime API, Platform API                                        │
│  提供: Component, Factory, Style 接口                                  │
│  禁止: 直接操作 Terminal, 处理 RawInput                                  │
└─────────────────────────────────────────────────────────────────────────┘
                                    ↓ depends on
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Boundary                                 │
│  tui/runtime/* - 布局引擎内核 (纯 Go，无外部依赖)                         │
│  依赖: 无 (纯内核)                                                       │
│  提供: Layout, Paint, Focus, Action 接口                                │
│  禁止: import framework, 知道 Component 类型                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    ↓ depends on
┌─────────────────────────────────────────────────────────────────────────┐
│                        Platform Boundary                                │
│  tui/platform/* - 平台抽象                                              │
│  依赖: OS, Terminal Driver                                              │
│  提供: Screen, Cursor, Input, Signal 接口                              │
│  禁止: 知道 Framework, Runtime, Component                               │
└─────────────────────────────────────────────────────────────────────────┘
```

## 依赖规则（强制）

### ✅ 允许的依赖

```go
// ✅ Application 可以依赖 Framework
import "github.com/yaoapp/yao/tui/framework"

// ✅ Framework 可以依赖 Runtime
import "github.com/yaoapp/yao/tui/runtime"

// ✅ Framework 可以依赖 Platform
import "github.com/yaoapp/yao/tui/platform"

// ✅ Platform 实现可以依赖 OS
import "os"
import "syscall"
```

### ❌ 禁止的依赖

```go
// ❌ Runtime 绝不依赖 Framework
// tui/runtime/ 不能导入 tui/framework/
import "github.com/yaoapp/yao/tui/framework"  // 禁止！

// ❌ Platform 绝不依赖 Framework 或 Runtime
// platform 包只提供抽象接口
import "github.com/yaoapp/yao/tui/framework"  // 禁止！

// ❌ Component 绝不直接操作 Terminal
// 必须通过 Runtime CellBuffer
import "github.com/yaoapp/yao/tui/platform"   // 禁止！

// ❌ 跨层直接访问
// Application 不能直接使用 Runtime 内部类型
import "github.com/yaoapp/yao/tui/runtime"    // 禁止！
```

## Platform 层接口

### 设计原则

Platform 层只提供"能力抽象"，不包含"语义"。

- ✅ 提供原始输入/输出能力
- ❌ 不理解 Focus、Event、Component、Layout

### 1. Screen 接口（V3: 从 Terminal 拆分）

```go
// 位于: tui/platform/screen.go

package platform

// Screen 屏幕输出抽象
type Screen interface {
    // 初始化
    Init() error
    Close() error

    // 尺寸
    Size() (width, height int)

    // 输出
    Write(data []byte) (int, error)
    Flush() error

    // 清屏
    Clear() error

    // 备用屏幕
    EnterAlternateScreen() error
    ExitAlternateScreen() error
}

// DefaultScreen 默认实现 (Unix)
type DefaultScreen struct {
    file    *os.File
    oldState *term.State
}

// WindowsScreen Windows 实现
type WindowsScreen struct {
    handle uintptr
    // Windows Console API
}
```

### 2. Cursor 接口（V3: 从 Terminal 拆分）

```go
// 位于: tui/platform/cursor.go

package platform

// Cursor 光标控制抽象
type Cursor interface {
    // 显示控制
    Show() error
    Hide() error

    // 移动
    Move(x, y int) error

    // 位置查询
    Position() (x, y int, err error)

    // 样式
    SetStyle(style CursorStyle) error
}

// CursorStyle 光标样式
type CursorStyle int

const (
    CursorBlock   CursorStyle = iota
    CursorUnderline
    CursorBar
)
```

### 3. InputReader 接口（V3: 从 Terminal 拆分）

```go
// 位于: tui/platform/input.go

package platform

// InputReader 输入读取抽象
type InputReader interface {
    // 读取单个输入
    ReadEvent() (RawInput, error)

    // 启动读取循环
    Start(events chan<- RawInput) error

    // 停止读取
    Stop() error
}

// RawInput 原始输入
type RawInput struct {
    Type RawInputType

    // 键盘
    Key      rune
    Special  SpecialKey
    Modifiers KeyModifier

    // 鼠标
    MouseX   int
    MouseY   int
    MouseButton MouseButton
    MouseAction MouseAction

    // 其他
    Data     []byte
    Timestamp time.Time
}

// RawInputType 输入类型
type RawInputType int

const (
    InputKeyPress RawInputType = iota
    InputKeyRelease
    InputMouse
    InputResize
    InputPaste
    InputSignal
)

// SpecialKey 特殊键
type SpecialKey int

const (
    KeyUnknown SpecialKey = iota
    KeyEscape
    KeyEnter
    KeyTab
    KeyBackspace
    KeyDelete
    // ...
)

// KeyModifier 修饰键
type KeyModifier uint8

const (
    ModShift KeyModifier = 1 << iota
    ModAlt
    ModCtrl
    ModMeta
)
```

### 4. SignalHandler 接口（V3: 从 Terminal 拆分）

```go
// 位于: tui/platform/signal.go

package platform

// SignalHandler 信号处理抽象
type SignalHandler interface {
    // 注册信号处理
    Handle(signals []os.Signal, handler func(os.Signal))

    // 启动监听
    Start() error

    // 停止监听
    Stop() error
}

// DefaultSignalHandler 默认实现
type DefaultSignalHandler struct {
    signals []os.Signal
    handler func(os.Signal)
    stop    chan struct{}
}
```

## Runtime 层接口

### 设计原则

Runtime 是"纯内核"，可独立测试、复用。

- ✅ 提供布局、绘制、焦点、Action 处理
- ❌ 不依赖 Framework、Platform、Component

### 1. Layout Engine

```go
// 位于: tui/runtime/layout/engine.go

package layout

// Engine 布局引擎
type Engine struct {
    // 纯布局逻辑，无外部依赖
}

// Layout 计算布局
func (e *Engine) Layout(nodes []Node, constraints Constraints) []LayoutBox {
    // Flexbox 算法
}

// Measure 测量尺寸
func (e *Engine) Measure(node Node, constraints Constraints) Size {
    // 测量逻辑
}

// Node 布局节点
type Node interface {
    ID() string
    Children() []Node
    Constraints() Constraints
}

// LayoutBox 布局结果
type LayoutBox struct {
    ID     string
    X, Y   int
    Width  int
    Height int
}
```

### 2. Paint Engine

```go
// 位于: tui/runtime/paint/engine.go

package paint

// Engine 绘制引擎
type Engine struct {
    buffer *CellBuffer
    dirty  *DirtyTracker
}

// Paint 绘制组件
func (e *Engine) Paint(node Node, box LayoutBox) {
    // 绘制逻辑
}

// Diff 计算差异
func (e *Engine) Diff(old, new *CellBuffer) []DiffChange {
    // Diff 逻辑
}

// CellBuffer 虚拟画布
type CellBuffer struct {
    cells  [][]Cell
    width  int
    height int
}

// Cell 单元格
type Cell struct {
    Char   rune
    Style  CellStyle
}

// CellStyle 单元格样式
type CellStyle struct {
    FG       Color
    BG       Color
    Bold     bool
    Underline bool
}
```

### 3. Focus Manager

```go
// 位于: tui/runtime/focus/manager.go

package focus

// Manager 焦点管理器
type Manager struct {
    path   FocusPath
    scopes []*FocusScope
}

// FocusPath 焦点路径
type FocusPath []string

// FocusScope 焦点作用域
type FocusScope struct {
    ID         string
    Type       ScopeType
    Focusables []string
}

// SetFocus 设置焦点
func (m *Manager) SetFocus(id string) bool

// PushScope 推入作用域
func (m *Manager) PushScope(scope *FocusScope)

// PopScope 弹出作用域
func (m *Manager) PopScope()
```

### 4. Action Dispatcher

```go
// 位于: tui/runtime/action/dispatcher.go

package action

// Dispatcher Action 分发器
type Dispatcher struct {
    targets        map[string]Target
    globalHandlers map[ActionType][]Handler
    focus          *focus.Manager
}

// Dispatch 分发 Action
func (d *Dispatcher) Dispatch(a *Action) bool {
    // 1. 全局处理器
    // 2. 焦点目标
    // 3. 指定目标
}

// Action Action 定义
type Action struct {
    Type      ActionType
    Payload   any
    Source    string
    Timestamp time.Time
}

// Target Action 目标
type Target interface {
    ID() string
    HandleAction(a *Action) bool
}
```

## Framework 层接口

### 设计原则

Framework 是"应用层桥接"，组装 Runtime 能力。

- ✅ 提供 Component、Factory、Style
- ✅ 适配 Runtime 到应用需求
- ❌ 不直接操作 Platform，不处理 RawInput

### 1. Component 接口

```go
// 位于: tui/framework/component/node.go

package component

// Node 基础节点接口
type Node interface {
    ID() string
    Type() string
}

// Paintable 可绘制组件
type Paintable interface {
    Node
    Paint(ctx PaintContext, buf *runtime.CellBuffer)
}

// ActionTarget 可处理 Action 的组件
type ActionTarget interface {
    Node
    HandleAction(a *runtime.Action) bool
}

// Focusable 可聚焦组件
type Focusable interface {
    Node
    FocusID() string
    OnFocus()
    OnBlur()
}

// BaseComponent 基础组件（组合）
type BaseComponent interface {
    Node
    Paintable
}

// InteractiveComponent 交互组件（组合）
type InteractiveComponent interface {
    BaseComponent
    ActionTarget
    Focusable
}
```

### 2. Factory 接口

```go
// 位于: tui/framework/component/factory.go

package component

// Factory 组件工厂（DSL 入口）
type Factory struct {
    runtime *runtime.Runtime
}

// CreateFromSpec 从 Spec 创建组件
func (f *Factory) CreateFromSpec(spec ComponentSpec) (Node, error) {
    switch spec.Type {
    case "text":
        return f.createText(spec)
    case "button":
        return f.createButton(spec)
    // ...
    }
}

// ComponentSpec 组件规格（DSL）
type ComponentSpec struct {
    Type   string                 // "text", "button", ...
    Props  map[string]interface{} // 组件属性
    Style  StyleSpec              // 样式规格
    Events map[string]string      // 事件绑定
    Children []ComponentSpec      // 子组件
}
```

### 3. Screen Manager（Framework 层）

```go
// 位于: tui/framework/screen/manager.go

package screen

// Manager 屏幕管理器
type Manager struct {
    screen   platform.Screen
    runtime  *runtime.Runtime
    frontBuf *runtime.CellBuffer
    backBuf  *runtime.CellBuffer
}

// Render 渲染缓冲区到屏幕
func (m *Manager) Render(buf *runtime.CellBuffer) error {
    // 1. Diff
    changes := m.runtime.Diff(m.frontBuf, buf)

    // 2. 输出
    for _, change := range changes {
        m.screen.Write(change.ToANSI())
    }

    // 3. Flush
    return m.screen.Flush()
}
```

## 边界检查清单

在提交代码前，请确认：

- [ ] Runtime 没有导入 Framework
- [ ] Platform 没有导入 Framework 或 Runtime
- [ ] Component 没有导入 Platform
- [ ] Component 只实现需要的能力接口
- [ ] Component 只处理 Action，不处理 RawInput
- [ ] 所有状态变化可追溯到 Action
- [ ] Render 函数只使用传入的 Context 和 Buffer

## 边界违反后果

| 严重程度 | 后果 |
|---------|------|
| 轻微 | 警告，要求修复 |
| 中等 | 拒绝合并，要求重构 |
| 严重 | 阻止发布，要求架构重新评审 |

> **记住：这些边界不是限制创造力，而是保护架构长期健康的护栏。**
