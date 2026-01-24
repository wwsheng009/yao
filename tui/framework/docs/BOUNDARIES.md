# Framework Boundaries and Interfaces

## 概述

本文档定义了新 TUI 框架各层之间的边界和接口契约，确保模块间的清晰分离和可替换性。

## 层次边界

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Boundary                             │
│  应用程序代码 - 用户创建的具体应用                                        │
│  依赖: Framework API                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Boundary                               │
│  tui/framework/* - 新框架代码                                            │
│  依赖: Runtime API, Platform API                                        │
│  提供: Component, Event, Style, Screen 接口                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Boundary                                 │
│  tui/runtime/* - 现有布局引擎 (复用)                                     │
│  依赖: 无 (纯内核)                                                       │
│  提供: Layout, CellBuffer, Focus, Animation, Event 接口                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Platform Boundary                                │
│  终端 I/O - 系统级操作                                                   │
│  依赖: OS, Terminal Driver                                              │
│  提供: 原始输入/输出, 窗口大小, 信号处理                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## 模块接口定义

### 1. Framework → Runtime 接口

#### 1.1 布局接口

```go
// Framework 使用 Runtime 的布局接口
// 位于: tui/framework/internal/runtime_adapter.go

package runtime_adapter

import (
    "github.com/yaoapp/yao/tui/runtime"
)

// RuntimeAdapter 适配 Runtime 接口到 Framework
type RuntimeAdapter struct {
    runtime *runtime.RuntimeImpl
}

// LayoutNode 转换
func (a *RuntimeAdapter) ToRuntimeNode(component Component) *runtime.LayoutNode {
    node := &runtime.LayoutNode{
        ID:       component.ID(),
        Type:     a.mapComponentType(component),
        Children: []*runtime.LayoutNode{},
    }

    // 设置尺寸约束
    node.Constraints = runtime.BoxConstraints{
        Min: runtime.Size{Width: component.MinWidth(), Height: component.MinHeight()},
        Max: runtime.Size{Width: component.MaxWidth(), Height: component.MaxHeight()},
    }

    return node
}

// Frame 转换
func (a *RuntimeAdapter) FromRuntimeFrame(frame runtime.Frame) *Frame {
    return &Frame{
        Buffer: frame.Buffer,
        Width:  frame.Width,
        Height: frame.Height,
    }
}

// 禁止: Runtime 不应依赖 Framework 的任何类型
// 禁止: Runtime 不应导入 framework 包
```

#### 1.2 可复用的 Runtime 能力

| 模块 | 接口 | Framework 使用方式 |
|------|------|-------------------|
| Layout | `runtime.Layout()` | 调用获取布局结果 |
| CellBuffer | `runtime.CellBuffer` | 直接使用虚拟画布 |
| Focus | `runtime.FocusManager` | 委托焦点管理 |
| Event | `runtime.EventDispatcher` | 委托事件分发 |
| Animation | `runtime.AnimationManager` | 委托动画控制 |
| Selection | `runtime.SelectionManager` | 委托选择管理 |
| Clipboard | `runtime.Clipboard` | 委托剪贴板操作 |

### 2. Framework → Platform 接口

#### 2.1 终端接口

```go
// 位于: tui/framework/platform/terminal.go

package platform

// Terminal 终端抽象接口
type Terminal interface {
    // 初始化
    Init() error
    Close() error

    // 屏幕操作
    EnterAlternateScreen() error
    ExitAlternateScreen() error
    EnableRawMode() error
    DisableRawMode() error

    // 光标操作
    ShowCursor() error
    HideCursor() error
    MoveCursor(x, y int) error

    // 输出
    Write(data []byte) (int, error)
    WriteString(s string) (int, error)
    Flush() error

    // 输入
    Read() ([]byte, error)

    // 窗口
    GetSize() (width, height int, err error)
    MonitorSize(callback func(width, height int))

    // 信号
    HandleSignals(signals []os.Signal, handler func(sig os.Signal))
}

// DefaultTerminal 默认实现
type DefaultTerminal struct {
    // 实现 Terminal 接口
}

// WindowsTerminal Windows 特定实现
type WindowsTerminal struct {
    // 使用 Windows Console API
}
```

#### 2.2 输入设备接口

```go
// 位于: tui/framework/platform/input.go

package platform

// InputReader 输入读取器接口
type InputReader interface {
    // 读取输入事件
    ReadEvent() (Event, error)

    // 启动读取循环
    Start(events chan<- Event) error

    // 停止读取
    Stop() error
}

// KeyboardEvent 键盘事件
type KeyboardEvent struct {
    Key       rune
    Modifiers KeyModifier
    Special   SpecialKey
}

// MouseEvent 鼠标事件
type MouseEvent struct {
    X      int
    Y      int
    Button MouseButton
    Action MouseAction
    Modifiers KeyModifier
}

// KeyModifier 键盘修饰键
type KeyModifier uint8

const (
    ModShift KeyModifier = 1 << iota
    ModAlt
    ModCtrl
    ModMeta
)

// SpecialKey 特殊键
type SpecialKey int

const (
    KeyEscape SpecialKey = iota
    KeyEnter
    KeyTab
    KeyBackspace
    KeyDelete
    KeyInsert
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyF1
    // ... F12
)

// MouseButton 鼠标按钮
type MouseButton int

const (
    MouseLeft MouseButton = iota
    MouseMiddle
    MouseRight
)

// MouseAction 鼠标动作
type MouseAction int

const (
    MousePress MouseAction = iota
    MouseRelease
    MouseMove
    MouseWheel
)
```

### 3. Framework 组件接口

#### 3.1 核心组件接口

```go
// 位于: tui/framework/component/component.go

package component

// Component 组件基础接口
type Component interface {
    // 标识
    ID() string

    // 生命周期
    Mount(parent Component)
    Unmount()

    // 尺寸
    SetSize(width, height int)
    GetSize() (width, height int)
    GetPreferredSize() (width, height int)
    GetMinSize() (width, height int)
    GetMaxSize() (width, height int)

    // 渲染
    Render(ctx RenderContext) string

    // 事件
    HandleEvent(ev Event) bool

    // 状态
    SetVisible(bool)
    IsVisible() bool
    SetEnabled(bool)
    IsEnabled() bool
}

// RenderContext 渲染上下文
type RenderContext struct {
    // 可用尺寸
    AvailableWidth  int
    AvailableHeight int

    // 偏移量 (用于滚动)
    OffsetX int
    OffsetY int

    // 样式继承
    InheritStyle Style

    // Z-index
    ZIndex int
}
```

#### 3.2 容器组件接口

```go
// 位于: tui/framework/component/container.go

package component

// Container 容器接口
type Container interface {
    Component

    // 子组件管理
    Add(child Component)
    Remove(child Component)
    RemoveAt(index int)
    GetChildren() []Component
    GetChild(index int) Component
    ChildCount() int

    // 布局
    SetLayout(layout Layout)
    GetLayout() Layout
}

// Layout 布局接口
type Layout interface {
    // 测量
    Measure(container Container, availableWidth, availableHeight int) (width, height int)

    // 布局
    Layout(container Container, x, y, width, height int)

    // 通知变更
    Invalidate()
}
```

#### 3.3 交互组件接口

```go
// 位于: tui/framework/component/interactive.go

package component

// Focusable 可聚焦组件
type Focusable interface {
    Component

    // 焦点
    SetFocus(bool)
    HasFocus() bool
    CanFocus() bool

    // 焦点导航
    FocusNext() Component
    FocusPrev() Component
}

// Validatable 可验证组件 (用于表单)
type Validatable interface {
    Component

    // 验证
    Validate() error
    SetValidator(validator Validator)
    IsValid() bool
}

// Validator 验证器
type Validator interface {
    Validate(value interface{}) error
}

// Updatable 可更新组件
type Updatable interface {
    Component

    // 更新
    Update(value interface{}) error
    GetValue() interface{}
}

// Scrollable 可滚动组件
type Scrollable interface {
    Component

    // 滚动
    ScrollTo(x, y int)
    ScrollBy(dx, dy int)
    GetScrollPosition() (x, y int)
    GetScrollRange() (minX, minY, maxX, maxY int)
    SetScrollSize(width, height int)
}
```

### 4. Framework 事件接口

#### 4.1 事件系统接口

```go
// 位于: tui/framework/event/event.go

package event

// Event 事件接口
type Event interface {
    Type() EventType
    Timestamp() time.Time
    Source() Component
    PreventDefault()
    IsDefaultPrevented() bool
    StopPropagation()
    IsPropagationStopped() bool
}

// EventHandler 事件处理器
type EventHandler interface {
    HandleEvent(Event) bool
}

// EventListener 事件监听器
type EventListener func(Event) bool

// EventBus 事件总线
type EventBus interface {
    // 订阅
    Subscribe(eventType EventType, handler EventHandler) Subscription
    SubscribeTo(component Component, eventType EventType, handler EventHandler) Subscription

    // 取消订阅
    Unsubscribe(sub Subscription)

    // 发布
    Publish(ev Event)

    // 路由
    RouteTo(component Component, ev Event) bool
}

// Subscription 订阅句柄
type Subscription interface {
    Unsubscribe()
}
```

### 5. 样式系统接口

#### 5.1 样式接口

```go
// 位于: tui/framework/style/style.go

package style

// Style 样式接口
type Style interface {
    // 应用到文本
    Apply(text string) string

    // 合并
    Merge(other Style) Style

    // 克隆
    Clone() Style
}

// Styled 可样式化接口
type Styled interface {
    SetStyle(style Style)
    GetStyle() Style
}

// Themable 可主题化接口
type Themable interface {
    SetTheme(theme Theme)
    GetTheme() Theme
}

// Theme 主题接口
type Theme interface {
    GetName() string
    GetColor(name string) Color
    GetStyle(name string) Style
}
```

### 6. 边界约束

#### 6.1 禁止的依赖

```go
// ❌ 禁止: Runtime 依赖 Framework
// tui/runtime/ 不能导入 tui/framework/

// ❌ 禁止: Platform 依赖 Framework 业务逻辑
// platform 包只能提供抽象接口

// ❌ 禁止: Component 直接访问 Terminal
// 必须通过 ScreenManager 间接访问

// ❌ 禁止: 跨层直接访问
// Application 不能直接使用 Runtime 内部类型
```

#### 6.2 允许的依赖

```go
// ✅ 允许: Framework 依赖 Runtime
// import "github.com/yaoapp/yao/tui/runtime"

// ✅ 允许: Framework 依赖 Platform
// import "github.com/yaoapp/yao/tui/framework/platform"

// ✅ 允许: Application 依赖 Framework
// import "github.com/yaoapp/yao/tui/framework"

// ✅ 允许: 通过接口解耦
// 使用接口而非具体实现
```

### 7. 数据边界

#### 7.1 类型转换规则

```go
// 位于: tui/framework/internal/converter.go

package internal

// EventTypeConverter 事件类型转换器
type EventTypeConverter struct{}

// ToFrameworkEvent 转换平台事件到框架事件
func (c *EventTypeConverter) ToFrameworkEvent(platformEv platform.Event) event.Event {
    switch pe := platformEv.(type) {
    case *platform.KeyboardEvent:
        return &event.KeyEvent{
            Key:       pe.Key,
            Modifiers: event.KeyModifier(pe.Modifiers),
        }
    case *platform.MouseEvent:
        return &event.MouseEvent{
            X:      pe.X,
            Y:      pe.Y,
            Button: event.MouseButton(pe.Button),
        }
    }
    return nil
}

// ToRuntimeStyle 转换框架样式到运行时样式
func (c *EventTypeConverter) ToRuntimeStyle(style style.Style) runtime.CellStyle {
    return runtime.CellStyle{
        FG:         c.convertColor(style.FG()),
        BG:         c.convertColor(style.BG()),
        Bold:       style.Bold(),
        Underline:  style.Underline(),
        // ...
    }
}
```

### 8. 生命周期边界

#### 8.1 应用生命周期

```go
// 位于: tui/framework/app.go

type LifecycleState int

const (
    StateCreated LifecycleState = iota
    StateInitializing
    StateRunning
    StatePaused
    StateStopping
    StateStopped
    StateError
)

// Lifecycle 生命周期接口
type Lifecycle interface {
    OnCreate()
    OnInit()
    OnStart()
    OnPause()
    OnResume()
    OnStop()
    OnDestroy()
    OnError(err error)
}
```

#### 8.2 组件生命周期

```
     Created
        │
        ▼
    Mounted ──────┐
        │         │
        ▼         │
    Updated ◄─────┘
        │
        ▼
   Unmounted
        │
        ▼
    Destroyed
```

## 依赖关系图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Dependencies                                  │
└─────────────────────────────────────────────────────────────────────────┘

Application (用户代码)
    ↓ depends on
Framework (tui/framework/*)
    ↓ depends on
    ├─→ Runtime (tui/runtime/*)
    └─→ Platform (tui/framework/platform/*)
           ↓ depends on
           OS / Terminal Driver

禁止的依赖:
    Runtime ←←← Framework (逆向依赖)
    Platform ←←← Application (跨层依赖)
    Runtime ←←← Platform (边界违反)
```

## 接口稳定性

| 层级 | 稳定性 | 说明 |
|------|--------|------|
| Application API | Stable | 对外公开 API |
| Framework Core | Stable | 核心接口 |
| Framework Internal | Volatile | 内部实现可能变化 |
| Runtime API | Stable | 复用的 Runtime 接口 |
| Platform API | Stable | 平台抽象接口 |
| Platform Impl | Volatile | 平台特定实现 |
