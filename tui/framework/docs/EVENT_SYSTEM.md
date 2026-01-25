# Event System Design (V3)

> **版本说明**: 本文档定义了 TUI 框架的事件系统。V3 明确定义了事件流的三个阶段：Capture、Target、Bubble，确保事件传播顺序可预测、可测试。

## 概述

事件系统负责处理用户输入和系统事件。V3 核心设计原则：

1. **明确的阶段**: Capture → Target → Bubble
2. **可预测的顺序**: 事件传播路径完全确定
3. **可中断**: 任何处理器都可以停止传播
4. **AI 友好**: 事件可记录、可回放
5. **与 Action 解耦**: 底层 Event → 语义 Action

## 与 Action 系统的关系

```
Platform Input (RawInput)
    │
    ▼
┌─────────────────────────────────────────┐
│  Event System (Runtime Layer)          │
│  - 解析 ANSI 序列                        │
│  - 分类事件类型                         │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  InputProcessor + KeyMap               │
│  - RawInput → Action (语义化)            │
│  - 支持上下文感知                        │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  Action Dispatcher (Framework Layer)    │
│  - Capture Phase (全局拦截)               │
│  - Target Phase (焦点目标)                │
│  - Bubble Phase (父组件冒泡)               │
└─────────────────────────────────────────┘
```

> **重要**: Component 应该实现 `ActionTarget` 来响应操作，而不是直接监听 `KeyEvent`。这确保 UI 可回放、可测试。

## 事件流阶段（V3 核心）

### 三阶段模型

```
User Input (e.g., Enter key)
    │
    ▼
┌─────────────────────────────────────────┐
│  Capture Phase (捕获阶段)               │
│  Root → Parent → Target                  │
│  ↓                                       │
│  全局拦截器优先执行                      │
│  用途: 快捷键、全局状态                  │
└─────────────────────────────────────────┘
    │ (如果未停止)
    ▼
┌─────────────────────────────────────────┐
│  Target Phase (目标阶段)                │
│  Target Component                       │
│  ↓                                       │
│  焦点目标处理 Action                   │
└─────────────────────────────────────────┘
    │ (如果未停止)
    ▼
┌─────────────────────────────────────────┐
│  Bubble Phase (冒泡阶段)                │
│  Target → Parent → Root                  │
│  ↓                                       │
│  父组件可以处理子组件未处理的事件       │
└─────────────────────────────────────────┘
```

### 阶段定义

```go
// 位于: tui/runtime/event/phase.go

package event

// EventPhase 事件阶段
type EventPhase int

const (
    PhaseCapture EventPhase = iota // 捕获阶段：从 Root 到 Target
    PhaseTarget                  // 目标阶段：Target 本身
    PhaseBubble                  // 冒泡阶段：从 Target 到 Root
)

// String 返回阶段名称
func (p EventPhase) String() string {
    switch p {
    case PhaseCapture:
        return "Capture"
    case PhaseTarget:
        return "Target"
    case PhaseBubble:
        return "Bubble"
    default:
        return "Unknown"
    }
}
```

### 传播规则

| 阶段 | 顺序 | 处理器 | 停止条件 |
|------|------|--------|---------|
| Capture | Root → Target | CaptureHandler | `StopPropagation()` |
| Target | Target | ActionTarget | 返回 `true` |
| Bubble | Target → Root | BubbleHandler | `StopPropagation()` |

## 核心类型定义

### 1. Event 接口

```go
// 位于: tui/runtime/event/event.go

package event

// Event 事件接口
type Event interface {
    // Type 返回事件类型
    Type() EventType

    // Phase 返回当前阶段
    Phase() EventPhase

    // Timestamp 返回时间戳
    Timestamp() time.Time

    // Target 返回目标组件
    Target() Node

    // PreventDefault 阻止默认行为
    PreventDefault()

    // IsDefaultPrevented 是否已阻止默认行为
    IsDefaultPrevented() bool

    // StopPropagation 停止传播
    StopPropagation()

    // IsPropagationStopped 是否已停止传播
    IsPropagationStopped() bool
}

// BaseEvent 基础事件实现
type BaseEvent struct {
    eventType EventType
    phase     EventPhase
    timestamp time.Time
    target    Node
    prevented bool
    stopped   bool
}

func (e *BaseEvent) Type() EventType             { return e.eventType }
func (e *BaseEvent) Phase() EventPhase           { return e.phase }
func (e *BaseEvent) Timestamp() time.Time       { return e.timestamp }
func (e *BaseEvent) Target() Node               { return e.target }
func (e *BaseEvent) PreventDefault()          { e.prevented = true }
func (e *BaseEvent) IsDefaultPrevented() bool  { return e.prevented }
func (e *BaseEvent) StopPropagation()         { e.stopped = true }
func (e *BaseEvent) IsPropagationStopped() bool { return e.stopped }

// SetPhase 设置阶段（内部使用）
func (e *BaseEvent) SetPhase(phase EventPhase) {
    e.phase = phase
}
```

### 2. 事件类型

```go
// 位于: tui/runtime/event/types.go

package event

// EventType 事件类型
type EventType int

const (
    // 系统事件
    EventInit EventType = iota + 1000
    EventTick
    EventResize
    EventSignal
    EventQuit

    // 键盘事件（原始，Platform 层）
    EventKeyPress
    EventKeyRelease

    // 鼠标事件（原始，Platform 层）
    EventMousePress
    EventMouseRelease
    EventMouseMove
    EventMouseWheel

    // Action 事件（语义化，Framework 层）
    EventAction

    // 组件事件
    EventClick
    EventChange
    EventFocus
    EventBlur
    EventSubmit
    EventCancel

    // 自定义事件
    EventCustom = 20000
)

// String 返回事件类型名称
func (t EventType) String() string {
    switch t {
    case EventInit:
        return "Init"
    case EventTick:
        return "Tick"
    case EventResize:
        return "Resize"
    case EventSignal:
        return "Signal"
    case EventQuit:
        return "Quit"
    case EventKeyPress:
        return "KeyPress"
    case EventAction:
        return "Action"
    default:
        if t >= EventCustom {
            return fmt.Sprintf("Custom(%d)", t)
        }
        return fmt.Sprintf("Unknown(%d)", t)
    }
}
```

### 3. 键盘事件（Platform 层）

```go
// 位于: tui/runtime/event/keyboard.go

package event

// KeyEvent 键盘事件（原始输入）
type KeyEvent struct {
    BaseEvent

    // 按键
    Key      rune      // 字符键 (如 'a', 'A', '1')
    Special  SpecialKey // 特殊键 (如 Enter, Escape)

    // 修饰键
    Modifiers KeyModifier

    // 重复
    Repeat bool
}

// SpecialKey 特殊键定义
type SpecialKey int

const (
    KeyUnknown SpecialKey = iota

    // 控制键
    KeyEscape
    KeyEnter
    KeyTab
    KeyBackspace
    KeyDelete
    KeyInsert

    // 光标键
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown

    // 功能键
    KeyF1
    KeyF2
    KeyF3
    // ... F12
)

// KeyModifier 修饰键
type KeyModifier uint8

const (
    ModShift KeyModifier = 1 << iota
    ModAlt
    ModCtrl
    ModMeta
)

// Has 检查是否有修饰键
func (m KeyModifier) Has(mod KeyModifier) bool {
    return m&mod != 0
}
```

### 4. Action 事件（Framework 层）

```go
// 位于: tui/runtime/event/action_event.go

package event

// ActionEvent Action 事件（语义化）
type ActionEvent struct {
    BaseEvent

    // Action 类型
    Action ActionType

    // Payload
    Payload any
}

// ActionType Action 类型
type ActionType string

const (
    // 导航
    ActionNavigateNext   ActionType = "navigate_next"
    ActionNavigatePrev   ActionType = "navigate_prev"
    ActionNavigateUp     ActionType = "navigate_up"
    ActionNavigateDown   ActionType = "navigate_down"

    // 编辑
    ActionInputText      ActionType = "input_text"
    ActionDeleteChar     ActionType = "delete_char"
    ActionDeleteWord     ActionType = "delete_word"

    // 表单
    ActionSubmit         ActionType = "submit"
    ActionCancel         ActionType = "cancel"

    // 系统
    ActionQuit           ActionType = "quit"
)
```

### 5. 处理器接口

```go
// 位于: tui/runtime/event/handler.go

package event

// CaptureHandler 捕获阶段处理器
type CaptureHandler interface {
    // HandleCapture 处理捕获阶段的事件
    HandleCapture(ev Event) bool
}

// TargetHandler 目标阶段处理器
type TargetHandler interface {
    // HandleTarget 处理目标阶段的事件
    HandleTarget(ev Event) bool
}

// BubbleHandler 冒泡阶段处理器
type BubbleHandler interface {
    // HandleBubble 处理冒泡阶段的事件
    HandleBubble(ev Event) bool
}
```

## 事件路由器

```go
// 位于: tui/runtime/event/router.go

package event

// Router 事件路由器
type Router struct {
    // 捕获阶段处理器（全局）
    captureHandlers []CaptureHandler

    // 焦点管理器
    focus *focus.Manager

    // 组件树根
    root Node
}

// NewRouter 创建事件路由器
func NewRouter(root Node) *Router {
    return &Router{
        captureHandlers: make([]CaptureHandler, 0),
        root:            root,
    }
}

// AddCaptureHandler 添加捕获处理器
func (r *Router) AddCaptureHandler(handler CaptureHandler) {
    r.captureHandlers = append(r.captureHandlers, handler)
}

// Route 路由事件
func (r *Router) Route(ev Event) bool {
    // 1. Capture Phase
    if r.capturePhase(ev) {
        return true
    }

    // 2. Target Phase
    if r.targetPhase(ev) {
        return true
    }

    // 3. Bubble Phase
    return r.bubblePhase(ev)
}

// capturePhase 捕获阶段
func (r *Router) capturePhase(ev Event) bool {
    for _, handler := range r.captureHandlers {
        ev.SetPhase(PhaseCapture)

        if handler.HandleCapture(ev) {
            return true // 停止传播
        }

        if ev.IsPropagationStopped() {
            return true
        }
    }
    return false
}

// targetPhase 目标阶段
func (r *Router) targetPhase(ev Event) bool {
    ev.SetPhase(PhaseTarget)

    target := ev.Target()
    if target == nil {
        return false
    }

    // 尝试作为 TargetHandler 处理
    if handler, ok := target.(TargetHandler); ok {
        if handler.HandleTarget(ev) {
            return true
        }
    }

    // 尝试作为 ActionTarget 处理
    if handler, ok := target.(component.ActionTarget); ok {
        // 将 Event 转换为 Action
        if actionEv, ok := ev.(*ActionEvent); ok {
            return handler.HandleAction(&action.Action{
                Type:    actionEv.Action,
                Payload: actionEv.Payload,
                Source:  actionEv.Source,
            })
        }
    }

    return false
}

// bubblePhase 冒泡阶段
func (r *Router) bubblePhase(ev Event) bool {
    ev.SetPhase(PhaseBubble)

    target := ev.Target()
    if target == nil {
        return false
    }

    // 获取父组件链
    parents := r.getParentChain(target)

    // 向上冒泡
    for i := len(parents) - 1; i >= 0; i-- {
        if handler, ok := parents[i].(BubbleHandler); ok {
            if handler.HandleBubble(ev) {
                return true // 停止传播
            }
        }

        if ev.IsPropagationStopped() {
            return true
        }
    }

    return false
}

// getParentChain 获取父组件链
func (r *Router) getParentChain(target Node) []Node {
    // 遍历组件树，找到从 root 到 target 的路径
    var path []Node

    current := r.root
    path = append(path, current)

    // TODO: 实现树遍历算法
    // 这里简化处理

    return path
}
```

## 事件泵

```go
// 位于: tui/runtime/event/pump.go

package event

import (
    "github.com/yaoapp/yao/tui/platform"
)

// Pump 事件泵
type Pump struct {
    reader   platform.InputReader
    queue    chan Event
    quit     chan struct{}
    parser   *Parser
}

// NewPump 创建事件泵
func NewPump(reader platform.InputReader) *Pump {
    return &Pump{
        reader: reader,
        queue:  make(chan Event, 100),
        quit:   make(chan struct{}),
        parser: NewParser(),
    }
}

// Start 启动事件泵
func (p *Pump) Start() {
    go p.readLoop()
}

// Stop 停止事件泵
func (p *Pump) Stop() {
    close(p.quit)
}

// Events 返回事件通道
func (p *Pump) Events() <-chan Event {
    return p.queue
}

// readLoop 读取循环
func (p *Pump) readLoop() {
    for {
        select {
        case <-p.quit:
            return

        default:
            input, err := p.reader.ReadEvent()
            if err != nil {
                // 处理错误
                continue
            }

            // 解析事件
            events := p.parser.Parse(input)
            for _, ev := range events {
                select {
                case p.queue <- ev:
                case <-p.quit:
                    return
                }
            }
        }
    }
}
```

## 使用示例

### 示例 1: 注册捕获处理器

```go
// 全局快捷键处理
type GlobalShortcutHandler struct{}

func (h *GlobalShortcutHandler) HandleCapture(ev event.Event) bool {
    if keyEv, ok := ev.(*event.KeyEvent); ok {
        // Ctrl+Q 全局退出
        if keyEv.Key == 'q' && keyEv.Modifiers.Has(event.ModCtrl) {
            app.Quit()
            return true // 停止传播
        }
    }
    return false
}

router.AddCaptureHandler(&GlobalShortcutHandler{})
```

### 示例 2: 组件冒泡处理

```go
type Form struct {
    *component.BaseComponent
    children []Node
}

func (f *Form) HandleBubble(ev event.Event) bool {
    if actionEv, ok := ev.(*event.ActionEvent); ok {
        if actionEv.Action == event.ActionSubmit {
            // 表单级别的提交验证
            if !f.Validate() {
                return true // 阻止冒泡
            }
        }
    }
    return false
}
```

### 示例 3: 事件流测试

```go
func TestEventFlow(t *testing.T) {
    router := event.NewRouter(root)

    // 添加测试处理器
    captureCalled := false
    targetCalled := false
    bubbleCalled := false

    router.AddCaptureHandler(&TestCaptureHandler{
        onHandle: func() { captureCalled = true },
    })

    // 模拟事件
    ev := &TestEvent{target: button}

    router.Route(ev)

    // 验证阶段顺序
    assert.True(t, captureCalled, "Capture phase should be called first")
    assert.True(t, targetCalled, "Target phase should be called second")
    assert.True(t, bubbleCalled, "Bubble phase should be called last")
}
```

## V2 → V3 主要变更

| 方面 | V2 | V3 |
|------|----|----|
| 阶段定义 | 不明确 | 明确 Capture/Target/Bubble |
| 处理器 | 单一 EventHandler | Capture/Target/Bubble 分离 |
| Component | HandleEvent | HandleAction (语义化) |
| 可预测性 | 顺序不确定 | 完全确定 |
| AI 友好 | 难以回放 | 可记录、可测试 |

## 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构总览
- [ACTION_SYSTEM.md](ACTION_SYSTEM.md) - Action 系统
- [FOCUS_SYSTEM.md](FOCUS_SYSTEM.md) - Focus 系统
- [BOUNDARIES.md](BOUNDARIES.md) - 层级边界
