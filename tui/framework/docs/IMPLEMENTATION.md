# Implementation Plan

## 概述

本文档提供了分阶段实施新 TUI 框架的详细步骤计划。

## 实施总览

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Implementation Phases                            │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 0: Foundation            Week 1-2                                │
│  - 项目结构搭建                                                          │
│  - 核心接口定义                                                          │
│  - Platform 抽象层                                                       │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 1: Core Framework      Week 3-5                                  │
│  - 主循环和生命周期                                                        │
│  - 事件系统                                                              │
│  - 屏幕管理                                                              │
│  - 样式系统                                                              │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 2: Runtime Integration Week 6-7                                 │
│  - Runtime 适配器                                                        │
│  - 布局集成                                                              │
│  - CellBuffer 集成                                                       │
│  - 焦点管理集成                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 3: Basic Components    Week 8-11                                │
│  - Text (文本显示)                                                        │
│  - Box (容器)                                                            │
│  - ProgressBar (进度条)                                                  │
│  - TextInput (文本输入)                                                  │
│  - Button (按钮)                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 4: Advanced Components Week 12-16                               │
│  - List (列表)                                                           │
│  - Table (表格)                                                          │
│  - Tree (树形)                                                           │
│  - Form (表单)                                                           │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 5: Polish & Testing    Week 17-20                               │
│  - 动画效果                                                              │
│  - 主题系统                                                              │
│  - 性能优化                                                              │
│  - 文档和示例                                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

## Phase 0: Foundation (Week 1-2)

### Step 0.1: 创建目录结构

```bash
tui/framework/
├── docs/
├── app.go
├── component/
│   ├── component.go
│   ├── container.go
│   └── registry.go
├── event/
│   ├── event.go
│   └── types.go
├── style/
│   ├── style.go
│   └── color.go
├── screen/
│   ├── screen.go
│   └── buffer.go
├── platform/
│   ├── terminal.go
│   └── input.go
└── util/
    └── utils.go
```

**任务清单:**
- [ ] 创建所有目录
- [ ] 创建 README.md
- [ ] 设置 go.mod
- [ ] 配置 gitignore

**验收标准:**
- 目录结构完整
- 编译通过
- 文档就绪

### Step 0.2: 定义核心接口

**文件: `tui/framework/component/component.go`**

```go
package component

type Component interface {
    ID() string
    Render(ctx RenderContext) string
    HandleEvent(ev event.Event) bool
    SetSize(width, height int)
    GetSize() (width, height int)
}

type RenderContext struct {
    AvailableWidth  int
    AvailableHeight int
    OffsetX         int
    OffsetY         int
}
```

**任务清单:**
- [ ] 定义 Component 接口
- [ ] 定义 Container 接口
- [ ] 定义 Interactive 接口
- [ ] 添加单元测试

**验收标准:**
- 接口定义清晰
- 测试覆盖核心方法

### Step 0.3: 实现 Platform 抽象

**文件: `tui/framework/platform/terminal.go`**

```go
package platform

type Terminal interface {
    Init() error
    Close() error
    WriteString(s string) (int, error)
    Read() ([]byte, error)
    GetSize() (width, height int, err error)
}

type DefaultTerminal struct {
    // 实现
}
```

**任务清单:**
- [ ] 实现 Terminal 接口
- [ ] 实现 UnixTerminal
- [ ] 实现 WindowsTerminal
- [ ] 添加输入解析

**验收标准:**
- 支持 Linux/macOS
- 支持 Windows
- 能读取键盘输入

## Phase 1: Core Framework (Week 3-5)

### Step 1.1: 主循环实现

**文件: `tui/framework/app.go`**

```go
package framework

type App struct {
    screen    *ScreenManager
    events    chan event.Event
    quit      chan struct{}
    root      component.Component
    running   bool
}

func (a *App) Run() error {
    if err := a.Init(); err != nil {
        return err
    }
    defer a.Close()

    go a.pumpEvents()

    for a.running {
        select {
        case ev := <-a.events:
            a.handleEvent(ev)
        case <-a.quit:
            return nil
        }

        if a.dirty {
            a.render()
        }
    }

    return nil
}
```

**任务清单:**
- [ ] 实现 App 结构
- [ ] 实现 Init 方法
- [ ] 实现 Run 方法
- [ ] 实现 Close 方法
- [ ] 添加生命周期钩子

**验收标准:**
- 能启动和退出
- 能处理事件
- 能渲染内容

### Step 1.2: 事件系统

**文件: `tui/framework/event/pump.go`**

```go
package event

type EventPump struct {
    input     platform.InputReader
    events    chan Event
    quit      chan struct{}
}

func (p *EventPump) Start() {
    go func() {
        for {
            select {
            case <-p.quit:
                return
            default:
                data, err := p.input.Read()
                if err != nil {
                    continue
                }
                ev := p.parseEvent(data)
                p.events <- ev
            }
        }
    }()
}

func (p *EventPump) parseEvent(data []byte) Event {
    // 解析 ANSI 转义序列
    // 返回对应的事件类型
}
```

**任务清单:**
- [ ] 实现 EventPump
- [ ] 实现 EventRouter
- [ ] 实现事件分类
- [ ] 添加事件测试

**验收标准:**
- 正确解析键盘输入
- 正确解析鼠标输入
- 支持组合键

### Step 1.3: 屏幕管理

**文件: `tui/framework/screen/manager.go`**

```go
package screen

type ScreenManager struct {
    terminal platform.Terminal
    buffer   *Buffer
    cursor   Cursor
}

func (s *ScreenManager) Init() error {
    s.terminal.EnterAlternateScreen()
    s.terminal.EnableRawMode()
    s.terminal.HideCursor()
    return nil
}

func (s *ScreenManager) Draw(buf *Buffer) error {
    diff := s.buffer.Diff(buf)
    for _, change := range diff {
        s.terminal.MoveCursor(change.X, change.Y)
        s.terminal.WriteString(change.Text)
    }
    s.buffer = buf
    return nil
}
```

**任务清单:**
- [ ] 实现 ScreenManager
- [ ] 实现 Buffer
- [ ] 实现差分渲染
- [ ] 添加 ANSI 支持

**验收标准:**
- 正确初始化屏幕
- 正确清理屏幕
- 高效增量渲染

### Step 1.4: 样式系统

**文件: `tui/framework/style/builder.go`**

```go
package style

type Style struct {
    FG         Color
    BG         Color
    Bold       bool
    Italic     bool
    Underline  bool
    Reverse    bool
}

type StyleBuilder struct {
    style Style
}

func NewStyle() *StyleBuilder {
    return &StyleBuilder{}
}

func (b *StyleBuilder) Foreground(c Color) *StyleBuilder {
    b.style.FG = c
    return b
}

func (b *StyleBuilder) Apply(text string) string {
    // 生成 ANSI 转义码
}
```

**任务清单:**
- [ ] 实现 Style 类型
- [ ] 实现 Color 类型
- [ ] 实现 StyleBuilder
- [ ] 实现 ANSI 生成

**验收标准:**
- 支持命名颜色
- 支持 256 色
- 支持 RGB 颜色
- 正确生成 ANSI

## Phase 2: Runtime Integration (Week 6-7)

### Step 2.1: Runtime 适配器

**文件: `tui/framework/internal/runtime_adapter.go`**

```go
package internal

import (
    "github.com/yaoapp/yao/tui/runtime"
)

type RuntimeAdapter struct {
    runtime *runtime.RuntimeImpl
}

func NewRuntimeAdapter(width, height int) *RuntimeAdapter {
    return &RuntimeAdapter{
        runtime: runtime.NewRuntime(width, height),
    }
}

func (a *RuntimeAdapter) Layout(root *runtime.LayoutNode) runtime.LayoutResult {
    return a.runtime.Layout(root, runtime.BoxConstraints{})
}

func (a *RuntimeAdapter) Render(result runtime.LayoutResult) runtime.Frame {
    return a.runtime.Render(result)
}
```

**任务清单:**
- [ ] 创建 RuntimeAdapter
- [ ] 实现组件到 LayoutNode 的转换
- [ ] 实现 Frame 的处理
- [ ] 添加集成测试

**验收标准:**
- 能调用 Runtime 布局
- 能获取渲染结果
- 测试通过

### Step 2.2: 组件注册表

**文件: `tui/framework/component/registry.go`**

```go
package component

type Registry struct {
    factories map[string]FactoryFunc
}

type FactoryFunc func() Component

func NewRegistry() *Registry {
    return &Registry{
        factories: make(map[string]FactoryFunc),
    }
}

func (r *Registry) Register(name string, factory FactoryFunc) {
    r.factories[name] = factory
}

func (r *Registry) Create(name string) (Component, error) {
    factory, ok := r.factories[name]
    if !ok {
        return nil, fmt.Errorf("unknown component: %s", name)
    }
    return factory(), nil
}
```

**任务清单:**
- [ ] 实现 Registry
- [ ] 注册基础组件
- [ ] 添加 DSL 支持

**验收标准:**
- 能注册组件
- 能创建组件
- 支持类型检查

## Phase 3: Basic Components (Week 8-11)

### Step 3.1: Text 组件

**文件: `tui/framework/component/display/text.go`**

```go
package display

type Text struct {
    BaseComponent
    content string
    align   TextAlign
    wrap    bool
}

func NewText(content string) *Text {
    return &Text{
        content: content,
        align:   AlignLeft,
        wrap:    true,
    }
}

func (t *Text) Render(ctx RenderContext) string {
    // 处理文本换行和对齐
    // 应用样式
}
```

**任务清单:**
- [ ] 实现 Text 组件
- [ ] 支持多行文本
- [ ] 支持文本对齐
- [ ] 支持自动换行
- [ ] 添加测试

**验收标准:**
- 正确显示文本
- 支持样式
- 支持换行

### Step 3.2: Box 容器

**文件: `tui/framework/component/layout/box.go`**

```go
package layout

type Box struct {
    BaseContainer
    border   Border
    padding  BoxSpacing
    margin   BoxSpacing
}

func (b *Box) Render(ctx RenderContext) string {
    // 计算内容区域（减去边距和内边距）
    // 渲染边框
    // 渲染子组件
}
```

**任务清单:**
- [ ] 实现 Box 容器
- [ ] 支持边框样式
- [ ] 支持内边距
- [ ] 支持外边距
- [ ] 添加测试

**验收标准:**
- 正确渲染边框
- 正确计算空间
- 支持嵌套

### Step 3.3: ProgressBar 组件

**文件: `tui/framework/component/widget/progress.go`**

```go
package widget

type ProgressBar struct {
    BaseComponent
    value    float64
    max      float64
    indeterminate bool
}

func (p *ProgressBar) Render(ctx RenderContext) string {
    // 计算进度条宽度
    // 渲染进度条
    // 显示百分比
}
```

**任务清单:**
- [ ] 实现 ProgressBar
- [ ] 支持确定/不确定状态
- [ ] 支持自定义样式
- [ ] 添加动画
- [ ] 添加测试

**验收标准:**
- 正确显示进度
- 支持样式自定义

### Step 3.4: TextInput 组件

**文件: `tui/framework/component/input/textinput.go`**

```go
package input

type TextInput struct {
    InteractiveComponent
    value    string
    cursor   int
    placeholder string
    password bool
    echo     rune
}

func (t *TextInput) HandleEvent(ev event.Event) bool {
    switch e := ev.(type) {
    case *event.KeyEvent:
        return t.handleKey(e)
    }
    return false
}

func (t *TextInput) handleKey(ev *event.KeyEvent) bool {
    // 处理字符输入
    // 处理光标移动
    // 处理删除
}
```

**任务清单:**
- [ ] 实现 TextInput
- [ ] 支持字符输入
- [ ] 支持光标移动
- [ ] 支持删除操作
- [ ] 支持密码模式
- [ ] 添加测试

**验收标准:**
- 正确接收输入
- 光标操作正确
- 支持复制粘贴

### Step 3.5: Button 组件

**文件: `tui/framework/component/interactive/button.go`**

```go
package interactive

type Button struct {
    InteractiveComponent
    label    string
    onClick  func()
}

func (b *Button) HandleEvent(ev event.Event) bool {
    switch e := ev.(type) {
    case *event.KeyEvent:
        if e.Key == '\n' || e.Key == ' ' {
            b.onClick()
            return true
        }
    case *event.MouseEvent:
        if e.Button == MouseLeft && e.Action == MousePress {
            b.onClick()
            return true
        }
    }
    return false
}
```

**任务清单:**
- [ ] 实现 Button
- [ ] 支持点击事件
- [ ] 支持快捷键
- [ ] 支持样式变化
- [ ] 添加测试

**验收标准:**
- 正确响应点击
- 显示正确状态

## Phase 4: Advanced Components (Week 12-16)

### Step 4.1: List 组件

**文件: `tui/framework/component/display/list.go`**

```go
package display

type List struct {
    InteractiveComponent
    items    []string
    cursor   int
    offset   int
}

func (l *List) Render(ctx RenderContext) string {
    // 渲染可见项
    // 高亮选中项
    // 处理滚动
}

func (l *List) HandleEvent(ev event.Event) bool {
    // 处理上下键
    // 处理 PageUp/PageDown
}
```

**任务清单:**
- [ ] 实现 List 组件
- [ ] 支持虚拟滚动
- [ ] 支持键盘导航
- [ ] 支持多选
- [ ] 添加测试

**验收标准:**
- 正确显示列表
- 导航流畅
- 性能良好

### Step 4.2: Table 组件

**文件: `tui/framework/component/display/table.go`**

```go
package display

type Table struct {
    InteractiveComponent
    columns []Column
    rows    [][]string
    cursor  int
    offset  int
}

type Column struct {
    Title    string
    Width    int
    Align    TextAlign
}

func (t *Table) Render(ctx RenderContext) string {
    // 渲染表头
    // 渲染数据行
    // 处理排序
}
```

**任务清单:**
- [ ] 实现 Table 组件
- [ ] 支持列定义
- [ ] 支持排序
- [ ] 支持选择
- [ ] 支持固定列
- [ ] 添加测试

**验收标准:**
- 正确显示表格
- 支持交互
- 性能良好

### Step 4.3: Tree 组件

**文件: `tui/framework/component/display/tree.go`**

```go
package display

type TreeNode struct {
    Label    string
    Expanded bool
    Children []*TreeNode
}

type Tree struct {
    InteractiveComponent
    root     *TreeNode
    cursor   int
    flat     []*TreeNode  // 扁平化的显示列表
}

func (t *Tree) Render(ctx RenderContext) string {
    // 渲染树形结构
    // 显示展开/折叠图标
}
```

**任务清单:**
- [ ] 实现 Tree 组件
- [ ] 支持展开/折叠
- [ ] 支持导航
- [ ] 支持懒加载
- [ ] 添加测试

**验收标准:**
- 正确显示树结构
- 交互流畅

### Step 4.4: Form 组件

**文件: `tui/framework/component/form/form.go`**

```go
package form

type Form struct {
    ContainerComponent
    fields   []Field
    submit   func(data map[string]interface{})
}

type Field interface {
    Component
    Name() string
    Value() interface{}
    Validate() error
}

func (f *Form) HandleEvent(ev event.Event) bool {
    // 处理表单提交
    // 处理字段间导航
}
```

**任务清单:**
- [ ] 实现 Form 组件
- [ ] 实现各种 Field 类型
- [ ] 实现验证器
- [ ] 支持表单布局
- [ ] 添加测试

**验收标准:**
- 正确验证输入
- 导航流畅
- 提交正确

## Phase 5: Polish & Testing (Week 17-20)

### Step 5.1: 动画系统

**任务清单:**
- [ ] 集成 Runtime 动画
- [ ] 实现过渡效果
- [ ] 添加缓动函数

### Step 5.2: 主题系统

**任务清单:**
- [ ] 实现主题管理
- [ ] 预设主题
- [ ] 主题切换

### Step 5.3: 性能优化

**任务清单:**
- [ ] 性能分析
- [ ] 渲染优化
- [ ] 内存优化

### Step 5.4: 文档和示例

**任务清单:**
- [ ] API 文档
- [ ] 使用指南
- [ ] 示例应用
- [ ] 教程

## 每日工作流

```bash
# 1. 拉取最新代码
git pull

# 2. 创建功能分支
git checkout -b feat/component-name

# 3. 编写代码
vim component_name.go

# 4. 编写测试
vim component_name_test.go

# 5. 运行测试
go test ./...

# 6. 运行示例
go run examples/demo.go

# 7. 提交代码
git add .
git commit -m "feat: add component"
git push
```

## 测试策略

### 单元测试

```go
func TestTextInput_HandleEvent(t *testing.T) {
    input := NewTextInput()
    input.SetSize(20, 1)

    // 测试字符输入
    ev := &event.KeyEvent{Key: 'a'}
    input.HandleEvent(ev)
    assert.Equal(t, "a", input.GetValue())

    // 测试删除
    ev = &event.KeyEvent{Special: KeyBackspace}
    input.HandleEvent(ev)
    assert.Equal(t, "", input.GetValue())
}
```

### 集成测试

```go
func TestApp_Run(t *testing.T) {
    app := NewApp()
    app.SetRoot(NewText("Hello"))

    // 测试启动
    err := app.Init()
    assert.NoError(t, err)

    // 测试渲染
    app.render()
    assert.True(t, app.screen.Dirty)

    // 测试清理
    err = app.Close()
    assert.NoError(t, err)
}
```

### 端到端测试

```go
func TestE2E_Form(t *testing.T) {
    // 模拟用户输入
    app := NewApp()
    form := NewForm()
    form.AddField("name", NewTextInput())
    app.SetRoot(form)

    // 发送按键事件
    app.SendEvent(Key('J'))
    app.SendEvent(Key('o'))
    app.SendEvent(Key('h'))
    app.SendEvent(Key('n'))

    // 验证结果
    assert.Equal(t, "John", form.GetValue("name"))
}
```

## 里程碑

| 里程碑 | 日期 | 交付物 |
|--------|------|--------|
| M0: Foundation | Week 2 | 项目结构，核心接口 |
| M1: Core Framework | Week 5 | 主循环，事件系统，屏幕管理 |
| M2: Runtime Integration | Week 7 | Runtime 集成完成 |
| M3: Basic Components | Week 11 | 基础组件完成 |
| M4: Advanced Components | Week 16 | 高级组件完成 |
| M5: Production Ready | Week 20 | 完整框架，文档，示例 |
