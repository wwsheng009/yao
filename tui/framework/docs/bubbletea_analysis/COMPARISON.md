# Bubble Tea vs Yao TUI v3 详细对比分析

本文档详细对比 Bubble Tea 和 Yao TUI v3 两个框架的架构、功能和设计模式。

## 目录

1. [架构模式对比](#架构模式对比)
2. [核心组件对比](#核心组件对比)
3. [事件系统对比](#事件系统对比)
4. [渲染系统对比](#渲染系统对比)
5. [异步操作对比](#异步操作对比)
6. [布局系统对比](#布局系统对比)
7. [测试支持对比](#测试支持对比)
8. [性能特性对比](#性能特性对比)
9. [可扩展性对比](#可扩展性对比)
10. [总结与建议](#总结与建议)

---

## 架构模式对比

### Bubble Tea: Elm Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Bubble Tea Architecture                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   User Input ──→ Msg ──→ Update() ──→ Model ──→ View()    │
│        ↑                                            ↓       │
│        └────────────────── Cmd ←────────────────────┘       │
│                        (async operations)                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**核心特点:**
- 单向数据流
- Model 是纯数据结构
- Update 是纯函数
- View 返回渲染字符串
- Cmd 处理异步操作

### Yao TUI v3: Three-Phase Rendering

```
┌─────────────────────────────────────────────────────────────────┐
│                    Yao TUI v3 Architecture                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐    │
│   │  Measure    │ ─→  │   Layout    │ ─→  │   Render    │    │
│   │   (自下而上) │     │   (自上而下) │     │  (Z-order)  │    │
│   └─────────────┘     └─────────────┘     └─────────────┘    │
│                                                                 │
│   ┌─────────────────────────────────────────────────────┐     │
│   │            Event System (Three-Phase)               │     │
│   │  Capture → Target → Bubble (几何优先的事件分发)      │     │
│   └─────────────────────────────────────────────────────┘     │
│                                                                 │
│   ┌─────────────────────────────────────────────────────┐     │
│   │            Runtime (Pure Layout Kernel)              │     │
│   │     独立于 Bubble Tea、DSL、具体组件                 │     │
│   └─────────────────────────────────────────────────────┘     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**核心特点:**
- 三阶段渲染流水线
- 几何优先的事件分发
- 分层架构 (Runtime/Framework)
- 虚拟画布 (CellBuffer)
- Flexbox 布局系统

### 架构对比表

| 维度 | Bubble Tea | Yao TUI v3 | 优势方 |
|------|-----------|------------|--------|
| **数据流** | 单向 MVU | 三阶段渲染 | 各有优势 |
| **状态管理** | 集中式 Model | 分布式 State | Yao TUI |
| **渲染方式** | 字符串拼接 | 虚拟画布 | Yao TUI |
| **布局能力** | 手动定位 | Flexbox 自动布局 | Yao TUI |
| **分层设计** | 单层 | Runtime/Framework 分层 | Yao TUI |
| **学习曲线** | 简单 | 较复杂 | Bubble Tea |
| **类型安全** | 接口约束 | Capability 接口 | Yao TUI |

---

## 核心组件对比

### Model 接口

#### Bubble Tea

```go
type Model interface {
    Init() Cmd
    Update(Msg) (Model, Cmd)
    View() string
}

// 典型实现
type model struct {
    choices []string
    cursor  int
    selected map[int]struct{}
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up":
            m.cursor--
        case "enter":
            m.selected[m.cursor] = struct{}{}
        }
    }
    return m, nil
}

func (m model) View() string {
    // 返回渲染字符串
    return "..."
}
```

#### Yao TUI v3

```go
// 组件能力接口
type Component interface {
    ID() string
}

type Measurable interface {
    Measure(maxWidth, maxHeight int) (width, height int)
}

type Paintable interface {
    Paint(ctx *PaintContext) error
}

type EventHandler interface {
    Handle(ctx *EventContext, event Event) bool
}

// 典型实现
type Button struct {
    BaseComponent
    Text    string
    OnClick Action
}

func (b *Button) Paint(ctx *PaintContext) error {
    // 绘制到虚拟画布
    return nil
}

func (b *Button) Handle(ctx *EventContext, event Event) bool {
    if e, ok := event.(*ClickEvent); ok {
        b.OnClick.Execute(ctx)
        return true
    }
    return false
}
```

### 组件对比表

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **组件定义** | 单一 Model 接口 | 多个 Capability 接口 |
| **状态存储** | Model 结构体 | StateHolder |
| **渲染方式** | View() 字符串 | Paint() 画布操作 |
| **事件处理** | Update() 模式匹配 | Handle() 几何分发 |
| **组合方式** | 嵌套 Model | Container 树 |

---

## 事件系统对比

### Bubble Tea 事件流

```
┌────────────────────────────────────────────────────────────┐
│                   Bubble Tea Event Flow                    │
├────────────────────────────────────────────────────────────┤
│                                                            │
│   Input ──→ WithFilter? ──→ Update() ──→ (Model, Cmd)     │
│       (可选拦截)         (模式匹配)                        │
│                                                            │
│   支持的事件类型:                                          │
│   • KeyMsg      - 键盘输入                                 │
│   • MouseMsg    - 鼠标事件                                 │
│   • WindowSizeMsg - 窗口大小变化                           │
│   • Custom Msg - 自定义消息                                │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Yao TUI v3 事件流

```
┌─────────────────────────────────────────────────────────────────┐
│                    Yao TUI v3 Event Flow                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Input ──→ Filter? ──→ Capture ──→ Target ──→ Bubble          │
│             (可选)      (根→叶)       (组件)      (叶→根)       │
│                                                                 │
│   支持的事件类型:                                               │
│   • KeyEvent      - 键盘事件                                   │
│   • MouseEvent    - 鼠标事件                                   │
│   • ResizeEvent   - 窗口大小变化                               │
│   • ActionEvent   - 语义化动作                                 │
│   • Custom Event  - 自定义事件                                 │
│                                                                 │
│   特性:                                                         │
│   • 基于几何的命中测试                                          │
│   • 事件优先级支持                                              │
│   • 停止传播机制                                                │
│   • 委托模式                                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 事件系统对比表

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **传播模式** | 集中式 Update | 三阶段几何传播 |
| **事件过滤** | WithFilter 拦截 | 事件过滤器 |
| **命中测试** | 无 | 基于 LayoutBox |
| **停止传播** | 返回特定值 | PropagationStopped |
| **优先级** | 无 | Priority 支持 |
| **异步事件** | Cmd | Action + Stream |

---

## 渲染系统对比

### Bubble Tea 渲染

```go
// 简单的字符串拼接
func (m model) View() string {
    s := "What should we buy at the market?\n\n"

    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        s += fmt.Sprintf("%s %s\n", cursor, choice)
    }

    s += "\nPress q to quit.\n"
    return s
}
```

**特点:**
- 返回渲染字符串
- 框架负责 diff 和更新
- 支持 lipgloss 样式库
- 帧率限制 (60/120 FPS)

### Yao TUI v3 渲染

```go
// 三阶段渲染
func (c *Container) Measure(maxWidth, maxHeight int) (int, int) {
    // 计算自身和子组件的期望尺寸
    return c.preferredWidth, c.preferredHeight
}

func (c *Container) Layout(box *LayoutBox) {
    // 接收分配的布局空间
    c.box = box
    // 为子组件分配空间
    for child, constraints := range c.children {
        child.Layout(childBox)
    }
}

func (c *Container) Paint(ctx *PaintContext) error {
    // 绘制到虚拟画布
    for _, child := range c.children {
        child.Paint(ctx)
    }
    return nil
}
```

**特点:**
- Measure → Layout → Render 三阶段
- 虚拟画布 CellBuffer
- Z-index 分层支持
- 脏区域优化
- Flexbox 自动布局

### 渲染系统对比表

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **渲染方式** | 字符串 | 虚拟画布 |
| **布局算法** | 手动 | Flexbox |
| **Z-index** | 无 | 支持 |
| **脏区域** | diff | 脏区域跟踪 |
| **帧率限制** | 内置 | 无 (可添加) |
| **样式系统** | lipgloss | 内置 Style |

---

## 异步操作对比

### Bubble Tea Cmd 系统

```go
// Cmd 是一个返回消息的函数
type Cmd func() Msg

// 执行异步操作
func waitForTick(total time.Duration, interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

// 批量执行 (并发)
tea.Batch(cmd1, cmd2, cmd3)

// 顺序执行
tea.Sequence(cmd1, cmd2, cmd3)
```

### Yao TUI v3 Action 系统

```go
// Action 语义化操作
type Action interface {
    Execute(ctx *Context) ActionResult
}

// 异步组件
type AsyncComponent struct {
    // ...
}

func (a *AsyncComponent) LoadData() {
    go func() {
        data := fetchFromAPI()
        a.Dispatch(NewDataLoadedAction(data))
    }()
}
```

### 异步操作对比表

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **抽象方式** | Cmd 函数 | Action 接口 |
| **并发控制** | Batch/Sequence | AsyncComponent |
| **取消机制** | context.Context | 有限支持 |
| **错误处理** | ErrorMsg | ActionResult |
| **流式数据** | 重复 Cmd | Stream 组件 |

---

## 布局系统对比

### Bubble Tea 布局

```go
// 手动布局，需要计算每个组件的位置
func (m model) View() string {
    // 左侧面板
    left := lipgloss.NewStyle().
        Width(20).
        Height(20).
        Render(m.leftView())

    // 右侧面板
    right := lipgloss.NewStyle().
        Width(40).
        Height(20).
        Render(m.rightView())

    // 水平排列
    return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
```

### Yao TUI v3 Flexbox 布局

```yaml
# DSL 声明式布局
layout:
  direction: horizontal
  children:
    - type: sidebar
      width: 20%

    - type: main
      width: flex  # 占据剩余空间
```

```go
// 或者代码方式
container := &Container{
    Direction: DirectionHorizontal,
    Children: []Component{
        &Sidebar{Width: Percent(20)},
        &Main{Flex: 1},
    },
}
```

### 布局系统对比表

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **布局方式** | 手动计算 | Flexbox 自动 |
| **响应式** | 手动处理 | 自动适配 |
| **嵌套支持** | 手动 Join | 天然支持 |
| **约束系统** | 无 | Min/Max/Pref |
| **定位类型** | 相对 | 相对/绝对 |

---

## 测试支持对比

### Bubble Tea 测试

```go
// 基础测试
func TestModelUpdate(t *testing.T) {
    m := model{}
    msg := tea.KeyMsg{Type: tea.KeyEnter}

    newModel, cmd := m.Update(msg)

    if newModel.(model).selected != nil {
        t.Error("expected selection")
    }
}
```

### Yao TUI v3 测试

```go
// 组件测试
func TestButtonClick(t *testing.T) {
    btn := NewButton("Click Me")
    clicked := false
    btn.OnClick = NewAction(func(ctx *Context) ActionResult {
        clicked = true
        return ActionResultOK
    })

    ctx := NewTestContext()
    event := NewClickEvent(btn)
    btn.Handle(ctx, event)

    assert.True(t, clicked)
}

// AI 测试
func TestWithAI(t *testing.T) {
    ai := NewAIController(app)
    ai.Query(".button").Click()
    ai.WaitSelector(".modal")
    assert.TextEqual(t, ai.GetText(".modal"), "Success!")
}
```

### 测试支持对比表

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **单元测试** | Update 纯函数 | 组件独立测试 |
| **集成测试** | Program 选项 | 测试运行器 |
| **AI 测试** | 无 | 内置 AI 控制器 |
| **断言库** | 基础 assert | 丰富的断言 |
| **回放支持** | 无 | Action 录制回放 |

---

## 性能特性对比

| 特性 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **渲染优化** | diff 算法 | 脏区域 + 增量渲染 |
| **帧率限制** | 内置 60/120 FPS | 无 (可添加) |
| **布局缓存** | 无 | 有 |
| **惰性计算** | 无 | Measure 懒加载 |
| **内存占用** | 较低 | 较高 (虚拟画布) |
| **CPU 占用** | diff 消耗 | 布局计算消耗 |

---

## 可扩展性对比

### Bubble Tea 扩展点

```go
// Program 选项模式
type ProgramOption func(*Program)

func WithRenderer(renderer renderer) ProgramOption
func WithInput(reader io.Reader) ProgramOption
func WithOutput(writer io.Writer) ProgramOption
func WithFilter(filter func(Model, Msg) Msg) ProgramOption
func WithContext(ctx context.Context) ProgramOption
```

### Yao TUI v3 扩展点

```go
// 能力接口扩展
type Measurable interface { Measure(...) }
type Paintable interface { Paint(...) }
type Focusable interface { Focus(), Blur() }
type EventHandler interface { Handle(...) }

// 组件扩展
type Component struct {
    BaseComponent
    // 添加任意字段和方法
}
```

### 扩展性对比表

| 维度 | Bubble Tea | Yao TUI v3 |
|------|-----------|------------|
| **组件扩展** | 实现 Model | 实现能力接口 |
| **生命周期** | Init/Update/View | 多个生命周期方法 |
| **插件系统** | Program 选项 | 能力组合 |
| **中间件** | WithFilter | 事件过滤器 |
| **主题系统** | lipgloss | 内置主题 |

---

## 总结与建议

### Yao TUI 应保留的独特优势

1. **三阶段渲染** - 更灵活的渲染控制
2. **Flexbox 布局** - 强大的自动布局
3. **虚拟画布** - Z-index 和分层渲染
4. **AI 集成** - 独特的开发和测试体验
5. **Action 系统** - 语义化的事件处理
6. **几何优先事件** - 更精确的事件分发

### 从 Bubble Tea 借鉴的功能

| 功能 | 优先级 | 理由 |
|------|--------|------|
| **事件过滤** | 高 | 调试、日志、权限控制 |
| **上下文取消** | 高 | 优雅关闭、超时控制 |
| **批量 Action** | 中 | 简化异步操作 |
| **帧率限制** | 中 | 性能优化 |
| **恐慌恢复** | 低 | 稳定性 |
| **输入取消** | 低 | 跨平台一致性 |

### 最终建议

Yao TUI v3 已经具备了比 Bubble Tea 更强大的基础架构，借鉴的重点应该放在**开发体验**和**系统稳定性**上，而不是核心架构。具体来说：

1. 保持三阶段渲染和 Flexbox 布局
2. 添加事件过滤器提升调试能力
3. 完善上下文管理实现优雅关闭
4. 引入批量操作简化异步编程
5. 增强错误处理和恢复机制

这样既能保持 Yao TUI 的技术优势，又能获得 Bubble Tea 的开发体验优势。
