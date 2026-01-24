# TUI Framework Architecture Design

## 概述

本文档描述了完全独立于 Bubble Tea 的新 TUI 框架架构设计。该框架复用现有的 `tui/runtime/` 布局引擎，提供完整的事件处理、渲染和组件系统。

## 设计目标

1. **完全独立**: 不依赖 Bubble Tea，拥有自主的消息循环和生命周期管理
2. **复用 Runtime**: 最大化复用现有的 `tui/runtime/` 布局引擎
3. **高性能**: 差分渲染、脏区域检测、增量更新
4. **可扩展**: 清晰的组件接口和扩展点
5. **易于使用**: 简洁的 API，支持 DSL 配置

## 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Application Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Table     │  │    List     │  │   Form      │  │   Input     │   │
│  │ Component   │  │  Component  │  │  Component  │  │  Component  │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Layer (NEW)                            │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       Component System                           │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │  Component  │  │  Container  │  │  Composite  │              │  │
│  │  │   Base      │  │   Wrapper   │  │   Pattern   │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       Event System                               │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │ Event Pump  │  │Event Router │  │Event Handler│              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     Style System                                 │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │Style Builder│  │Theme Manager│  │Color Palette│              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                   Screen & Renderer                              │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │   Screen    │  │   Painter   │  │  Buffer     │              │  │
│  │  │  Manager    │  │             │  │  Manager    │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Layer (REUSED)                           │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     Layout Engine                                │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │   Flexbox   │  │   Measure   │  │ Constraints │              │  │
│  │  │   Layout    │  │   System    │  │   Solver    │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     CellBuffer                                   │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │ Virtual     │  │ Z-Index     │  │ Styled      │              │  │
│  │  │ Canvas      │  │ Support     │  │ Text        │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     Focus Manager                                │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │ Tab Navigation│ Focus Store  │ Modal Trap    │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     Event & Selection                            │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │  │
│  │  │Event Dispatcher│ Selection   │ Clipboard     │              │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Platform Layer                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Terminal  │  │   Input     │  │   Output    │  │   Signals   │   │
│  │   I/O       │  │   Reader    │  │   Writer    │  │  Handler    │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 核心模块设计

### 1. 主循环 (Main Loop)

```go
// 位于: tui/framework/app.go
type App struct {
    // Runtime 复用现有布局引擎
    runtime   *runtime.RuntimeImpl

    // 屏幕管理
    screen    *ScreenManager

    // 事件系统
    events    chan Event
    quit      chan struct{}

    // 组件树
    root      Component

    // 状态
    running   bool
    dirty     bool
}

// Run 启动应用程序
func (a *App) Run() error {
    a.Init()
    defer a.Close()

    // 启动事件监听
    go a.pumpEvents()

    // 主循环
    for a.running {
        select {
        case ev := <-a.events:
            a.handleEvent(ev)
        case <-a.time.Cick(Tick):
            a.handleTick()
        case <-a.quit:
            return nil
        }

        // 渲染
        if a.dirty {
            a.render()
        }
    }

    return nil
}
```

### 2. 组件系统 (Component System)

```go
// 位于: tui/framework/component/component.go

// Component 是所有组件的基础接口
type Component interface {
    // 渲染组件
    Render() string

    // 处理事件
    HandleEvent(ev Event) bool

    // 设置尺寸
    SetSize(width, height int)

    // 获取首选尺寸
    GetPreferredSize() Size
}

// BaseComponent 提供组件的默认实现
type BaseComponent struct {
    id       string
    width    int
    height   int
    style    Style
    visible  bool
    focused  bool
}

// InteractiveComponent 支持交互的组件
type InteractiveComponent struct {
    BaseComponent
    focusable bool
    enabled   bool
}

// ContainerComponent 容器组件
type ContainerComponent struct {
    InteractiveComponent
    children []Component
    layout   Layout
}
```

### 3. 事件系统 (Event System)

```go
// 位于: tui/framework/event/

// EventType 事件类型
type EventType int

const (
    // 键盘事件
    EventKeyPress EventType = iota
    EventKeyRelease

    // 鼠标事件
    EventMousePress
    EventMouseRelease
    EventMouseMove
    EventMouseWheel

    // 窗口事件
    EventResize
    EventFocus
    EventBlur

    // 组件事件
    EventClick
    EventChange
    EventSubmit
    EventCancel

    // 自定义事件
    EventCustom
)

// Event 统一事件结构
type Event struct {
    Type      EventType
    Timestamp time.Time
    Source    Component

    // 事件数据
    Key       *KeyEvent
    Mouse     *MouseEvent
    Resize    *ResizeEvent
    Custom    interface{}
}

// EventHandler 事件处理器
type EventHandler interface {
    HandleEvent(ev Event) bool
}

// EventRouter 事件路由器
type EventRouter struct {
    handlers map[EventType][]EventHandler
    capture  []EventHandler
    bubble   []EventHandler
}
```

### 4. 样式系统 (Style System)

```go
// 位于: tui/framework/style/

// Color 颜色表示
type Color struct {
    Type  ColorType
    Value interface{} // string for named, int for 256, [3]int for RGB
}

type ColorType int

const (
    ColorNamed ColorType = iota
    Color256
    ColorRGB
)

// Style 样式定义
type Style struct {
    FG         Color
    BG         Color
    Bold       bool
    Italic     bool
    Underline  bool
    Strikethrough bool
    Reverse    bool
    Blink      bool
    Border     BorderStyle
    Margin     Box
    Padding    Box
}

// StyleBuilder 样式构建器
type StyleBuilder struct {
    style Style
}

func (s *StyleBuilder) Foreground(c Color) *StyleBuilder {
    s.style.FG = c
    return s
}

func (s *StyleBuilder) Build() Style {
    return s.style
}
```

### 5. 屏幕管理 (Screen Manager)

```go
// 位于: tui/framework/screen/

// ScreenManager 屏幕管理器
type ScreenManager struct {
    // 终端
    terminal *Terminal

    // 前后缓冲区
    front    *Buffer
    back     *Buffer

    // 光标
    cursor   Cursor

    // 原始模式
    rawMode  bool
}

// Init 初始化屏幕
func (s *ScreenManager) Init() error {
    // 进入备用屏幕
    s.enterAlternateScreen()

    // 启用原始模式
    s.enableRawMode()

    // 隐藏光标
    s.hideCursor()

    // 初始化缓冲区
    s.initBuffers()

    return nil
}

// Close 关闭屏幕
func (s *ScreenManager) Close() error {
    // 显示光标
    s.showCursor()

    // 退出原始模式
    s.disableRawMode()

    // 退出备用屏幕
    s.exitAlternateScreen()

    return nil
}

// Draw 绘制缓冲区内容到屏幕
func (s *ScreenManager) Draw(buf *Buffer) error {
    // 计算差异
    diff := s.front.Diff(buf)

    // 只绘制变化的区域
    for _, change := range diff {
        s.moveCursor(change.X, change.Y)
        s.writeString(change.Text)
    }

    // 更新前缓冲
    s.front = buf

    return nil
}
```

## 目录结构

```
tui/framework/
├── README.md                    # 框架概述
├── docs/                        # 文档目录
│   ├── ARCHITECTURE.md          # 架构设计 (本文档)
│   ├── BOUNDARIES.md            # 边界定义
│   ├── IMPLEMENTATION.md        # 实施步骤
│   ├── COMPONENTS.md            # 组件设计
│   ├── EVENT_SYSTEM.md          # 事件系统
│   ├── RENDERING.md             # 渲染系统
│   ├── STYLING.md               # 样式系统
│   └── EXAMPLES.md              # 示例代码
├── app.go                       # 应用程序入口
├── component/                   # 组件系统
│   ├── component.go             # 组件基础接口
│   ├── container.go             # 容器组件
│   ├── composite.go             # 组合模式
│   └── registry.go              # 组件注册表
├── event/                       # 事件系统
│   ├── event.go                 # 事件定义
│   ├── pump.go                  # 事件泵
│   ├── router.go                # 事件路由
│   ├── keyboard.go              # 键盘处理
│   └── mouse.go                 # 鼠标处理
├── style/                       # 样式系统
│   ├── style.go                 # 样式定义
│   ├── color.go                 # 颜色处理
│   ├── border.go                # 边框样式
│   └── theme.go                 # 主题管理
├── screen/                      # 屏幕管理
│   ├── screen.go                # 屏幕接口
│   ├── buffer.go                # 缓冲区管理
│   ├── terminal.go              # 终端操作
│   └── ansi.go                  # ANSI 转义码
├── input/                       # 输入组件
│   ├── textinput.go             # 文本输入
│   ├── textarea.go              # 多行输入
│   └── password.go              # 密码输入
├── display/                     # 显示组件
│   ├── table.go                 # 表格
│   ├── list.go                  # 列表
│   ├── tree.go                  # 树形
│   └── text.go                  # 文本
├── layout/                      # 布局组件
│   ├── flex.go                  # Flex 容器
│   ├── grid.go                  # Grid 容器
│   ├── box.go                   # 盒子
│   └── stack.go                 # 层叠容器
├── widget/                      # 小部件
│   ├── progress.go              # 进度条
│   ├── spinner.go               # 加载动画
│   ├── meter.go                 # 仪表盘
│   └── chart.go                 # 图表
├── form/                        # 表单组件
│   ├── form.go                  # 表单容器
│   ├── field.go                 # 表单字段
│   ├── validation.go            # 验证器
│   └── schema.go                # 表单模式
└── util/                        # 工具函数
    ├── strings.go               # 字符串工具
    ├── ansi.go                  # ANSI 处理
    └── timer.go                 # 定时器
```

## 数据流

```
User Input (Keyboard/Mouse)
         │
         ▼
┌─────────────────┐
│  Event Pump     │ 采集原始输入
└─────────────────┘
         │
         ▼
┌─────────────────┐
│ Event Classifier │ 分类事件类型
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Event Router   │ 路由到组件
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   Component     │ 处理事件
│  Handlers       │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   State Update  │ 更新状态
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   Runtime       │ 布局计算
│   Layout        │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   Component     │ 渲染组件
│   Render        │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   CellBuffer    │ 合成到虚拟画布
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Diff Engine    │ 计算差异
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Screen Output  │ 输出到终端
└─────────────────┘
```

## 渲染管线

```
Component.Render()
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 1: Measure (自底向上)            │
│  - 计算每个组件的首选尺寸                │
│  - 考虑内容、约束、样式                  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 2: Layout (自顶向下)             │
│  - 应用 Flexbox 算法                     │
│  - 分配位置和尺寸                        │
│  - 处理对齐和间距                        │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 3: Render                        │
│  - 将组件渲染到 CellBuffer               │
│  - 应用样式和颜色                        │
│  - 处理 Z-index 层级                     │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 4: Diff                          │
│  - 比较前后缓冲区                        │
│  - 生成增量更新                          │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Phase 5: Output                        │
│  - 输出 ANSI 转义码                      │
│  - 最小化终端操作                        │
└─────────────────────────────────────────┘
```

## 线程模型

```
┌─────────────────────────────────────────────────────────────┐
│                      Main Thread                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                 Main Loop                            │   │
│  │  while running:                                      │   │
│  │    receive event                                     │   │
│  │    handle event                                      │   │
│  │    update state                                      │   │
│  │    render                                            │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         │                                │
         │ subscribe                      │ trigger
         ▼                                ▼
┌─────────────────────┐         ┌─────────────────────┐
│   Event Pump        │         │   Animation Timer   │
│   (Goroutine)       │         │   (Goroutine)       │
│  - Read stdin       │         │  - Tick events      │
│  - Parse input      │         │  - Frame sync       │
│  - Send events      │         └─────────────────────┘
└─────────────────────┘
```

## 扩展点

1. **自定义组件**: 实现 `Component` 接口
2. **自定义布局**: 实现 `Layout` 接口
3. **自定义样式**: 扩展 `Style` 类型
4. **自定义事件**: 使用 `EventCustom` 类型
5. **自定义渲染**: 覆盖 `Render` 方法

## 兼容性

- **Go 版本**: 1.21+
- **平台**: Linux, macOS, Windows (通过 conpty)
- **终端**: 支持 ANSI 转义码的现代终端
- **编码**: UTF-8
