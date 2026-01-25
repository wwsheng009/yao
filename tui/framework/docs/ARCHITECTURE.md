# TUI Framework Architecture Design V3

> **版本说明**: 本文档基于全面的架构审查结果重新设计，目标是构建一个"可长期演进、AI 友好、工程化"的 TUI 框架。

## 概述

本框架采用四层架构，严格分离关注点，确保每层可独立测试、复用和演进。

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Layer                               │
│  用户应用代码 - 使用 Framework API 构建具体应用                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Layer                                  │
│  组件系统 + Action 路由 + 适配器                                          │
│  - Component (Capability Interfaces)                                    │
│  - Action Router & Event Handling                                       │
│  - Runtime ↔ Framework Adapters                                        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Layer                                    │
│  纯内核 - 无外部依赖，可独立测试                                          │
│  - Layout Engine (Flexbox)                                              │
│  - Paint System (CellBuffer + RenderTree)                              │
│  - Focus Manager (Path + Scope)                                         │
│  - Dirty Region Tracker                                                 │
│  - Animation Manager (按需 tick)                                        │
│  - Input Processor (KeyMap → Action)                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Platform Layer                                   │
│  平台抽象 - 最小化接口                                                    │
│  - Screen (输出抽象)                                                     │
│  - Cursor (光标抽象)                                                     │
│  - InputReader (输入抽象)                                                │
│  - SignalHandler (信号抽象)                                              │
└─────────────────────────────────────────────────────────────────────────┘
```

## 设计目标（按优先级）

### 1. AI 友好（核心目标）

- UI 状态完全可描述（`StateSnapshot`）
- 所有操作可回放（`ActionLog`）
- AI 通过语义 `Action` 操作，而非模拟按键
- 组件树可查询（类似 DOM Selector）

### 2. 长期演进

- 架构不变量明确且强制执行
- Runtime/Foundation 物理隔离
- 接口为替换而设计，不为实现方便

### 3. 高性能

- Dirty Region 跟踪
- 增量渲染（只渲染变化区域）
- 动画按需触发（无动画时零 CPU）

### 4. 工程化

- 所有状态可枚举、可快照
- 事件流阶段明确（Capture/Target/Bubble）
- 完整的可测试性

## 核心数据流（单向）

```
User Input (stdin/signals)
    │
    ▼
┌─────────────────────────────────────────┐
│ Platform: InputReader.Read()           │ → RawInput
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Runtime: InputProcessor.Parse()        │ → KeyEvent/MouseEvent
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Runtime: KeyMap.Map(contexts)          │ → Action (语义化)
└─────────────────────────────────────────┘
    │
    ├──► Capture Phase ──► 全局拦截器
    │
    ├──► Target Phase ───► FocusTarget.HandleAction()
    │
    └──► Bubble Phase ───► 父组件处理
    │
    ▼
┌─────────────────────────────────────────┐
│ State Update (显式，可追踪)              │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Mark Dirty Region (主动标记)             │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Runtime: Layout.Compute()               │ → RenderNode Tree
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Runtime: Paint.ToBuffer()               │ → CellBuffer
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Runtime: Dirty.Diff()                  │ → 变更集
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│ Platform: Screen.Write()               │ → ANSI 输出
└─────────────────────────────────────────┘
```

## V2 → V3 主要变更

| 方面 | V2 | V3 |
|------|----|----|
| Component | 胖接口 | Capability Interfaces (8+ 能力接口) |
| Render | 返回 `string` | `Paint(ctx, buf)` + RenderTree 中间态 |
| Event | 阶段不明确 | 明确 Capture/Target/Bubble |
| Focus | `bool` | FocusPath + Scope 栈 |
| Animation | 全局 tick | 按需 tick，有时间轴 |
| State | 隐式 | StateSnapshot 显式 |
| Platform | Terminal 大接口 | Screen/Cursor/Input/Signal 分离 |
| Input | Component 处理 KeyEvent | 只处理 Action |
| Dirty | Cell diff | 主动标记 Dirty Region |

## 目录结构

```
tui/framework/
├── docs/                        # 架构文档
│   ├── ARCHITECTURE.md          # 架构总览 (本文档)
│   ├── ARCHITECTURE_INVARIANTS.md  # 架构不变量 ⚠️
│   ├── BOUNDARIES.md            # 层级边界
│   ├── ACTION_SYSTEM.md         # Action 系统
│   ├── FOCUS_SYSTEM.md          # Focus 系统
│   ├── COMPONENTS.md            # Component V3
│   ├── EVENT_SYSTEM.md          # 事件系统
│   ├── RENDERING.md             # 渲染系统
│   ├── STATE_MANAGEMENT.md      # 状态管理 ⭐ NEW
│   ├── ANIMATION_SYSTEM.md      # 动画系统 ⭐ NEW
│   ├── AI_INTEGRATION.md        # AI 集成
│   ├── STYLING.md               # 样式系统
│   ├── THEME_SYSTEM.md          # 主题系统 ⭐ NEW
│   ├── VIRTUAL_SCROLL.md        # 虚拟滚动 ⭐ NEW
│   ├── ERROR_HANDLING.md        # 错误处理 ⭐ NEW
│   ├── PAINTER_ABSTRACTION.md   # Painter 抽象 ⭐ NEW
│   ├── ASYNC_TASK.md            # 异步任务 ⭐ NEW
│   ├── TABLE_SUBFOCUS.md        # Table 子焦点 ⭐ NEW
│   ├── STREAM_DATA.md           # 流式数据 ⭐ NEW
│   ├── V8_INTEGRATION.md        # V8 集成 ⭐ NEW
│   └── FORM_VALIDATION.md       # Form 验证 ⭐ NEW
│
├── runtime/                     # Runtime 内核 (纯 Go，无依赖)
│   ├── layout/                  # 布局引擎
│   │   ├── flex.go
│   │   ├── constraint.go
│   │   └── measure.go
│   ├── paint/                   # 绘制系统
│   │   ├── cell.go              # Cell 定义
│   │   ├── buffer.go            # CellBuffer
│   │   ├── tree.go              # RenderTree
│   │   ├── dirty.go             # DirtyRegion
│   │   └── diff.go              # Diff 引擎
│   ├── focus/                   # 焦点系统
│   │   ├── path.go              # FocusPath
│   │   ├── scope.go             # FocusScope
│   │   ├── manager.go           # FocusManager
│   │   └── modal.go             # ModalManager
│   ├── input/                   # 输入处理
│   │   ├── raw.go               # RawInput
│   │   ├── keymap.go            # KeyMap (支持 Context)
│   │   └── processor.go         # InputProcessor
│   ├── action/                  # Action 系统
│   │   ├── action.go
│   │   ├── dispatcher.go
│   │   ├── router.go            # Capture/Target/Bubble
│   │   └── log.go               # ActionLog (回放)
│   ├── animation/               # 动画系统 ⭐ NEW
│   │   ├── animation.go         # Animation 定义
│   │   ├── manager.go           # AnimationManager (按需 tick)
│   │   └── easing.go            # Easing 函数
│   ├── state/                   # 状态管理 ⭐ NEW
│   │   ├── snapshot.go          # StateSnapshot
│   │   └── tracker.go           # StateTracker
│   └── runtime.go               # Runtime 核心
│
├── platform/                   # 平台抽象
│   ├── screen.go                # Screen 接口
│   ├── cursor.go                # Cursor 接口
│   ├── input.go                 # InputReader 接口
│   ├── signal.go                # SignalHandler 接口
│   └── impl/                    # 平台实现
│       ├── default/             # 默认实现
│       └── windows/             # Windows 实现
│
├── component/                   # Framework 组件
│   ├── node.go                  # Node 基础接口
│   ├── mountable.go             # Mountable 能力
│   ├── measurable.go            # Measurable 能力
│   ├── paintable.go             # Paintable 能力
│   ├── actionable.go            # ActionTarget 能力
│   ├── focusable.go             # Focusable 能力
│   ├── scrollable.go            # Scrollable 能力
│   ├── validatable.go           # Validatable 能力
│   ├── base.go                  # BaseComponent 组合
│   ├── container.go             # Container 接口
│   └── factory.go               # ComponentFactory (DSL 入口)
│
├── screen/                      # Framework 屏幕管理
│   ├── manager.go               # ScreenManager
│   ├── painter.go               # Painter (Framework → Runtime)
│   └── viewport.go              # Viewport 管理
│
├── input/                       # 输入组件
│   ├── textinput.go
│   ├── textarea.go
│   └── password.go
│
├── display/                     # 显示组件
│   ├── text.go
│   ├── list.go
│   ├── table.go
│   └── tree.go
│
├── layout/                      # 布局组件
│   ├── flex.go
│   ├── box.go
│   └── grid.go
│
├── interactive/                 # 交互组件
│   ├── button.go
│   ├── checkbox.go
│   ├── radio.go
│   └── switch.go
│
├── overlay/                     # 覆盖层组件 ⭐ NEW
│   ├── modal.go
│   ├── dialog.go
│   ├── menu.go
│   └── tooltip.go
│
├── form/                        # 表单组件 ⭐ NEW
│   ├── form.go
│   ├── field.go
│   └── validator.go
│
├── style/                       # 样式系统
│   ├── style.go
│   ├── color.go
│   ├── theme.go
│   ├── palette.go
│   └── manager.go               # ThemeManager ⭐ NEW
│
├── validation/                  # 验证系统 ⭐ NEW
│   ├── validator.go
│   ├── builtin.go
│   └── composite.go
│
├── async/                       # 异步任务系统 ⭐ NEW
│   ├── task.go
│   ├── manager.go
│   └── progress.go
│
├── stream/                      # 流式数据系统 ⭐ NEW
│   ├── stream.go
│   ├── buffer.go
│   └── sources.go
│
├── result/                      # Result 类型 ⭐ NEW
│   └── result.go
│
├── paint/                       # Painter 抽象 ⭐ NEW
│   ├── painter.go
│   ├── cellbuffer.go
│   ├── terminal_painter.go
│   └── html_painter.go
│
├── v8/                          # V8 集成 ⭐ NEW
│   ├── registry.go
│   └── script_component.go
│
├── event/                       # Framework 事件
│   ├── event.go
│   ├── handler.go
│   └── bus.go                   # EventBus (仅用于 System Event)
│
└── app.go                       # 应用入口
```

## 核心接口定义

### 1. Component 能力接口（V3）

```go
// 基础节点
type Node interface {
    ID() string
    Type() string
}

// 可挂载
type Mountable interface {
    Node
    Mount(parent Container) error
    Unmount() error
    IsMounted() bool
}

// 可测量
type Measurable interface {
    Node
    Measure(constraints Constraints) Size
    GetSize() Size
}

// 可绘制 (V3: 不返回 string)
type Paintable interface {
    Node
    Paint(ctx PaintContext, buf *runtime.CellBuffer)
}

// 可处理 Action (V3: 不处理 KeyEvent)
type ActionTarget interface {
    Node
    HandleAction(a *runtime.Action) bool
}

// 可聚焦 (V3: 返回 FocusID)
type Focusable interface {
    Node
    FocusID() string
    OnFocus()
    OnBlur()
}

// 可滚动
type Scrollable interface {
    Node
    ScrollTo(x, y int)
    ScrollBy(dx, dy int)
    GetScrollPosition() (x, y int)
}

// 可验证
type Validatable interface {
    Node
    Validate() error
    IsValid() bool
}

// 组合接口
type BaseComponent interface {
    Node
    Mountable
    Measurable
    Paintable
}

type InteractiveComponent interface {
    BaseComponent
    ActionTarget
    Focusable
}
```

### 2. Platform 接口（V3: 拆分）

```go
// 屏幕输出
type Screen interface {
    Init() error
    Close() error
    Size() (width, height int)
    Write(data []byte) (int, error)
    Flush() error
    Clear() error
}

// 光标控制
type Cursor interface {
    Show() error
    Hide() error
    Move(x, y int) error
    Position() (x, y int, err error)
}

// 输入读取
type InputReader interface {
    ReadEvent() (RawInput, error)
    Start(events chan<- RawInput) error
    Stop() error
}

// 信号处理
type SignalHandler interface {
    Handle(signals []os.Signal, handler func(os.Signal))
}
```

### 3. Runtime 核心

```go
// Runtime 内核
type Runtime struct {
    layout     *layout.Engine
    paint      *paint.Engine
    focus      *focus.Manager
    dirty      *dirty.Tracker
    animation  *animation.Manager  // ⭐ 按需 tick
    input      *input.Processor
    action     *action.Dispatcher
    state      *state.Tracker      // ⭐ 状态快照
}

// Dispatch 唯一入口
func (r *Runtime) Dispatch(a *Action) {
    // 1. Focus 处理
    r.focus.Handle(a)

    // 2. 状态更新
    r.state.Apply(a)

    // 3. 标记脏区域
    r.dirty.MarkFrom(a)

    // 4. 动画处理
    r.animation.Handle(a)

    // 5. 布局计算 (仅当 dirty)
    if r.dirty.HasAny() {
        r.layout.Compute()
    }

    // 6. 绘制 (仅 dirty 区域)
    r.paint.Render(r.dirty.Regions())
}
```

## 扩展点

1. **自定义组件**: 实现需要的 Capability Interfaces
2. **自定义 Action**: 扩展 ActionType，通过 KeyMap 绑定
3. **自定义样式**: 扩展 Style 和 Theme
4. **自定义布局**: 实现 Measurable 接口
5. **自定义动画**: 使用 AnimationManager 注册
6. **AI 接入**: 实现 Controller 接口
7. **自定义验证器**: 实现 Validator 接口
8. **自定义数据源**: 实现 DataSource 接口（虚拟滚动）
9. **脚本组件**: 使用 V8 集成编写 JS/TS 组件
10. **自定义 Painter**: 实现 Painter 接口（支持其他后端）

## 新增功能模块（V3）

| 模块 | 优先级 | 功能描述 | 文档 |
|------|--------|----------|------|
| 主题系统 | P0 | 运行时主题切换、样式继承 | [THEME_SYSTEM.md](THEME_SYSTEM.md) |
| 虚拟滚动 | P0 | 大数据量列表高性能渲染 | [VIRTUAL_SCROLL.md](VIRTUAL_SCROLL.md) |
| 错误处理 | P0 | Result 类型、错误边界、异步错误 | [ERROR_HANDLING.md](ERROR_HANDLING.md) |
| Painter 抽象 | P1 | 多后端支持（终端/HTML） | [PAINTER_ABSTRACTION.md](PAINTER_ABSTRACTION.md) |
| 异步任务 | P1 | 长时间运行任务、超时控制 | [ASYNC_TASK.md](ASYNC_TASK.md) |
| Table 子焦点 | P1 | 单元格级焦点、内联编辑 | [TABLE_SUBFOCUS.md](TABLE_SUBFOCUS.md) |
| 流式数据 | P1 | 实时流式显示、增量更新 | [STREAM_DATA.md](STREAM_DATA.md) |
| V8 集成 | P2 | JS/TS 脚本组件、热重载 | [V8_INTEGRATION_YAO.md](V8_INTEGRATION_YAO.md) |
| Form 验证 | P2 | 表单验证、实时反馈 | [FORM_VALIDATION.md](FORM_VALIDATION.md) |

## 兼容性

- **Go 版本**: 1.21+
- **平台**: Linux, macOS, Windows
- **终端**: 支持 ANSI 转义码的现代终端
- **编码**: UTF-8
- **V8**: rogchap.com/v8go (可选，用于脚本组件)

## 相关文档

### 核心架构
- [ARCHITECTURE_INVARIANTS.md](ARCHITECTURE_INVARIANTS.md) - **必须遵守**的架构不变量
- [BOUNDARIES.md](BOUNDARIES.md) - 层级边界和依赖规则
- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构总览 (本文档)

### 系统模块
- [ACTION_SYSTEM.md](ACTION_SYSTEM.md) - Action 语义事件系统
- [FOCUS_SYSTEM.md](FOCUS_SYSTEM.md) - Focus Path 系统
- [COMPONENTS.md](COMPONENTS.md) - Component V3 设计
- [EVENT_SYSTEM.md](EVENT_SYSTEM.md) - 事件流阶段
- [RENDERING.md](RENDERING.md) - 渲染管线
- [STATE_MANAGEMENT.md](STATE_MANAGEMENT.md) - 状态管理
- [ANIMATION_SYSTEM.md](ANIMATION_SYSTEM.md) - 动画系统
- [AI_INTEGRATION.md](AI_INTEGRATION.md) - AI 集成标准

### 扩展功能
- [THEME_SYSTEM.md](THEME_SYSTEM.md) - 主题系统
- [VIRTUAL_SCROLL.md](VIRTUAL_SCROLL.md) - 虚拟滚动
- [ERROR_HANDLING.md](ERROR_HANDLING.md) - 错误处理
- [PAINTER_ABSTRACTION.md](PAINTER_ABSTRACTION.md) - Painter 抽象
- [ASYNC_TASK.md](ASYNC_TASK.md) - 异步任务
- [TABLE_SUBFOCUS.md](TABLE_SUBFOCUS.md) - Table 子焦点
- [STREAM_DATA.md](STREAM_DATA.md) - 流式数据
- [V8_INTEGRATION_YAO.md](V8_INTEGRATION_YAO.md) - V8 集成（基于 Yao 基础设施）
- [V8_EVENT_CALLBACK.md](V8_EVENT_CALLBACK.md) - V8 事件回调桥接
- [PROCESS_INTEGRATION_YAO.md](PROCESS_INTEGRATION_YAO.md) - Yao Process 集成
- [FORM_VALIDATION.md](FORM_VALIDATION.md) - Form 验证
