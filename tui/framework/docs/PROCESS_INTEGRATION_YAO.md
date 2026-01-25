# Yao Process 集成方案

> **版本**: V3.0
> **优先级**: P0 (核心集成)
> **复用度**: 95% (直接复用现有 Yao Process 基础设施)

## 概述

本文档说明 TUI V3 组件如何与 Yao Process 系统集成，实现 UI 与业务逻辑的完全解耦。

### 设计目标

1. **单向数据流**: Component → Action → Process → State Update → Render
2. **完全解耦**: TUI 组件不直接依赖业务逻辑
3. **可测试**: Process 可独立测试，不依赖 TUI
4. **多入口复用**: 同一 Process 可被 CLI、HTTP、TUI 调用

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         TUI V3 Framework                                │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                        Component Layer                            │  │
│  │  ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐  │  │
│  │  │  Button  │────│ TextInput │────│   Table  │────│  Modal   │  │  │
│  │  └──────────┘     └──────────┘     └──────────┘     └──────────┘  │  │
│  │        │               │               │               │           │  │
│  │        └───────────────┴───────────────┴───────────────┘           │  │
│  │                            │                                       │  │
│  │                            ▼                                       │  │
│  │                   HandleAction(action)                             │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                  │                                       │
│                                  ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                    Action Dispatcher                              │  │
│  │                                                                   │  │
│  │   ┌─────────────────────────────────────────────────────────┐     │  │
│  │   │  Action → Process Mapping                               │     │  │
│  │   │                                                         │     │  │
│  │   │  "submit" → "scripts.user.FormSubmit"                   │     │  │
│  │   │  "load"   → "models.user.Find"                          │     │  │
│  │   │  "save"   → "models.user.Save"                          │     │  │
│  │   │  "delete" → "models.user.Delete"                        │     │  │
│  │   └─────────────────────────────────────────────────────────┘     │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                  │                                       │
│                                  ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                   Process Bridge Layer                             │  │
│  │                                                                   │  │
│  │   process.New(name, args...)                                      │  │
│  │       .WithGlobal(global)                                         │  │
│  │       .WithSID(sid)                                                │  │
│  │       .WithV8Context(ctx)     // V8 上下文亲和                     │  │
│  │       .Execute()               // 异步执行                          │  │
│  │       .ExecuteSync()           // 同步执行（V8）                   │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Yao Process System                              │
│                            (gou/process)                                │
│                                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │   Models     │  │   Scripts    │  │  Services    │  │   Flows    │ │
│  │   Handler    │  │   Handler    │  │   Handler    │  │   Handler  │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘ │
│                                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │    Agents    │  │    Plugins   │  │   Systems    │  │   Custom   │ │
│  │   Handler    │  │   Handler    │  │   Handler    │  │   Handler  │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Business Logic Layer                              │
│                    (Services, Flows, DSL Scripts)                       │
└─────────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Data Layer                                     │
│                    (Database, Cache, External APIs)                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### 双向通信机制

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        TUI ↔ Process 双向通信                            │
└─────────────────────────────────────────────────────────────────────────┘

    Component ──► Action ──► Process ──► Business Logic ──► Database
        │                                                          │
        │                                                          │
        └──────────────────────────────────────────────────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Result      │
                    │ Callback    │
                    │ Message     │
                    └─────────────┘
                           │
                           ▼
                    State Update
                           │
                           ▼
                       Re-render
```

## 核心实现

### 1. Process Bridge

```go
package framework

import (
    "github.com/yaoapp/gou/process"
)

// ProcessBridge TUI 到 Yao Process 的桥接
type ProcessBridge struct {
    model     Model
    global    bool
    sid       string
    v8Context *v8.Context
}

// NewProcessBridge 创建 Process 桥接
func NewProcessBridge(model Model) *ProcessBridge {
    return &ProcessBridge{
        model:  model,
        global: true, // 默认使用全局进程
    }
}

// WithGlobal 设置全局标志
func (b *ProcessBridge) WithGlobal(global bool) *ProcessBridge {
    b.global = global
    return b
}

// WithSID 设置会话 ID
func (b *ProcessBridge) WithSID(sid string) *ProcessBridge {
    b.sid = sid
    return b
}

// WithV8Context 设置 V8 上下文
func (b *ProcessBridge) WithV8Context(ctx *v8.Context) *ProcessBridge {
    b.v8Context = ctx
    return b
}

// Execute 异步执行 Process
func (b *ProcessBridge) Execute(name string, args ...interface{}) error {
    p := process.New(name, args...)

    if b.global {
        p = p.WithGlobal(b.global)
    }

    if b.sid != "" {
        p = p.WithSID(b.sid)
    }

    if b.v8Context != nil {
        p = p.WithV8Context(b.v8Context)
    }

    // 发送消息到 Bubble Tea Program
    b.model.sendProcessMessage(p)

    return nil
}

// ExecuteSync 同步执行 Process（V8 上下文共享）
func (b *ProcessBridge) ExecuteSync(name string, args ...interface{}) (interface{}, error) {
    p := process.New(name, args...)

    if b.global {
        p = p.WithGlobal(b.global)
    }

    if b.sid != "" {
        p = p.WithSID(b.sid)
    }

    if b.v8Context != nil {
        p = p.WithV8Context(b.v8Context)
    }

    return p.ExecuteSync()
}
```

### 2. Action → Process 映射

```go
package framework

// ActionToProcessConfig Action 到 Process 的映射配置
type ActionToProcessConfig struct {
    Action    string
    Process   string
    Transform func(action *Action) []interface{}
    Async     bool
}

// ProcessRegistry Action → Process 注册表
type ProcessRegistry struct {
    mappings map[string]*ActionToProcessConfig
}

// NewProcessRegistry 创建注册表
func NewProcessRegistry() *ProcessRegistry {
    return &ProcessRegistry{
        mappings: make(map[string]*ActionToProcessConfig),
    }
}

// Register 注册映射
func (r *ProcessRegistry) Register(config *ActionToProcessConfig) {
    r.mappings[config.Action] = config
}

// Get 获取映射配置
func (r *ProcessRegistry) Get(action string) (*ActionToProcessConfig, bool) {
    config, ok := r.mappings[action]
    return config, ok
}

// DefaultMappings 默认映射配置
func DefaultMappings() *ProcessRegistry {
    r := NewProcessRegistry()

    // 表单相关
    r.Register(&ActionToProcessConfig{
        Action:  "form.submit",
        Process: "scripts.form.Submit",
        Transform: func(a *Action) []interface{} {
            return []interface{}{a.Payload}
        },
        Async: true,
    })

    r.Register(&ActionToProcessConfig{
        Action:  "form.validate",
        Process: "scripts.form.Validate",
        Transform: func(a *Action) []interface{} {
            return []interface{}{a.Payload}
        },
        Async: false,
    })

    // 数据操作
    r.Register(&ActionToProcessConfig{
        Action:  "data.load",
        Process: "models.user.Find",
        Transform: func(a *Action) []interface{} {
            id := a.Payload["id"]
            return []interface{}{id}
        },
        Async: true,
    })

    r.Register(&ActionToProcessConfig{
        Action:  "data.save",
        Process: "models.user.Save",
        Transform: func(a *Action) []interface{} {
            return []interface{}{a.Payload}
        },
        Async: true,
    })

    r.Register(&ActionToProcessConfig{
        Action:  "data.delete",
        Process: "models.user.Delete",
        Transform: func(a *Action) []interface{} {
            id := a.Payload["id"]
            return []interface{}{id}
        },
        Async: true,
    })

    // 列表操作
    r.Register(&ActionToProcessConfig{
        Action:  "list.load",
        Process: "models.user.List",
        Transform: func(a *Action) []interface{} {
            page := a.Payload["page"]
            size := a.Payload["size"]
            return []interface{}{page, size}
        },
        Async: true,
    })

    // 搜索
    r.Register(&ActionToProcessConfig{
        Action:  "search",
        Process: "services.search.Query",
        Transform: func(a *Action) []interface{} {
            query := a.Payload["query"]
            return []interface{}{query}
        },
        Async: true,
    })

    return r
}
```

### 3. Component 集成

```go
package component

import (
    "github.com/yaoapp/yao/tui/framework"
    "github.com/yaoapp/yao/tui/runtime"
)

// ActionableComponent 可执行 Action 的组件
type ActionableComponent struct {
    *BaseComponent
    processBridge *framework.ProcessBridge
    processMap    *framework.ProcessRegistry
}

// NewActionableComponent 创建可执行 Action 的组件
func NewActionableComponent(id string) *ActionableComponent {
    return &ActionableComponent{
        BaseComponent: NewBaseComponent(id),
        processMap:    framework.DefaultMappings(),
    }
}

// SetProcessBridge 设置 Process 桥接
func (c *ActionableComponent) SetProcessBridge(bridge *framework.ProcessBridge) {
    c.processBridge = bridge
}

// HandleAction 处理 Action
func (c *ActionableComponent) HandleAction(a *runtime.Action) bool {
    // 检查是否有映射的 Process
    config, ok := c.processMap.Get(a.Type)
    if !ok {
        return false // 没有映射，交给父组件处理
    }

    // 转换 Action 为 Process 参数
    args := config.Transform(a)

    // 执行 Process
    if config.Async {
        err := c.processBridge.Execute(config.Process, args...)
        if err != nil {
            // 错误处理
            c.handleError(err)
            return true
        }
        // 异步执行，返回 true 表示已处理
        return true
    }

    // 同步执行
    result, err := c.processBridge.ExecuteSync(config.Process, args...)
    if err != nil {
        c.handleError(err)
        return true
    }

    // 处理结果
    c.handleResult(result)

    return true
}

// handleError 错误处理
func (c *ActionableComponent) handleError(err error) {
    // 标记组件为错误状态
    c.SetState(map[string]interface{}{
        "error": err.Error(),
    })

    // 触发错误 Action
    c.EmitAction("error", map[string]interface{}{
        "component": c.ID(),
        "error":     err.Error(),
    })
}

// handleResult 处理 Process 结果
func (c *ActionableComponent) handleResult(result interface{}) {
    // 根据结果更新组件状态
    c.SetState(map[string]interface{}{
        "data": result,
    })
}
```

### 4. Button 组件示例

```go
package component

import (
    "github.com/yaoapp/yao/tui/runtime"
)

// Button 按钮组件
type Button struct {
    *ActionableComponent
    label   string
    onClick string // 点击时执行的 Process 名称
}

// NewButton 创建按钮
func NewButton(id, label, onClick string) *Button {
    btn := &Button{
        label:   label,
        onClick: onClick,
    }
    btn.ActionableComponent = NewActionableComponent(id)
    return btn
}

// HandleAction 处理 Action
func (b *Button) HandleAction(a *runtime.Action) bool {
    switch a.Type {
    case "click":
        // 执行配置的 Process
        if b.onClick != "" {
            err := b.processBridge.Execute(b.onClick)
            if err != nil {
                b.handleError(err)
            }
            return true
        }
        // 触发 onClick Action
        b.EmitAction("click", nil)
        return true
    }

    // 交给父类处理
    return b.ActionableComponent.HandleAction(a)
}
```

### 5. Table 组件示例（带编辑）

```go
package component

import (
    "github.com/yaoapp/yao/tui/runtime"
)

// Table 表格组件
type Table struct {
    *ActionableComponent
    columns    []Column
    data       []map[string]interface{}
    editable   bool
    onEdit     string // 编辑时执行的 Process
    onDelete   string // 删除时执行的 Process
}

// Column 列定义
type Column struct {
    Key     string
    Title   string
    Width   int
    Editable bool
}

// NewTable 创建表格
func NewTable(id string, columns []Column) *Table {
    t := &Table{
        columns:  columns,
        editable: true,
    }
    t.ActionableComponent = NewActionableComponent(id)
    return t
}

// SetData 设置数据
func (t *Table) SetData(data []map[string]interface{}) {
    t.data = data
    t.MarkDirty()
}

// HandleAction 处理 Action
func (t *Table) HandleAction(a *runtime.Action) bool {
    switch a.Type {
    case "edit":
        if !t.editible || t.onEdit == "" {
            return false
        }
        row := a.Payload["row"].(int)
        col := a.Payload["col"].(string)
        value := a.Payload["value"]

        // 执行编辑 Process
        err := t.processBridge.Execute(t.onEdit,
            t.data[row]["id"],
            col,
            value,
        )
        if err != nil {
            t.handleError(err)
        }
        return true

    case "delete":
        if t.onDelete == "" {
            return false
        }
        row := a.Payload["row"].(int)

        // 执行删除 Process
        err := t.processBridge.Execute(t.onDelete,
            t.data[row]["id"],
        )
        if err != nil {
            t.handleError(err)
        }
        return true
    }

    return t.ActionableComponent.HandleAction(a)
}
```

### 6. Process → TUI 回调

```go
package tui

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/yaoapp/gou/process"
)

// ProcessResultMsg Process 结果消息
type ProcessResultMsg struct {
    Process string
    Result  interface{}
    Error   error
}

// ProcessExecutor Process 执行器（在 Model 中使用）
type ProcessExecutor struct {
    model Model
}

// NewProcessExecutor 创建执行器
func NewProcessExecutor(model Model) *ProcessExecutor {
    return &ProcessExecutor{model: model}
}

// Execute 执行 Process 并返回结果消息
func (e *ProcessExecutor) Execute(p *process.Process) tea.Cmd {
    return func() tea.Msg {
        result, err := p.ExecuteSync()
        return ProcessResultMsg{
            Process: p.Name,
            Result:  result,
            Error:   err,
        }
    }
}

// Model 集成示例
type Model struct {
    // ... 其他字段
    executor *ProcessExecutor
}

// Update 处理 Process 结果
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ProcessResultMsg:
        if msg.Error != nil {
            // 处理错误
            return m, m.showError(msg.Error)
        }

        // 根据 Process 名称处理结果
        switch msg.Process {
        case "models.user.Find":
            return m.handleUserLoaded(msg.Result)
        case "models.user.Save":
            return m.handleUserSaved(msg.Result)
        case "models.user.Delete":
            return m.handleUserDeleted(msg.Result)
        }
    }

    return m, nil
}
```

## Yao Process 调用 TUI

### 现有 TUI Process API

| Process | 参数 | 功能 |
|---------|------|------|
| `tui.load` | `config` | 加载 TUI 应用 |
| `tui.get` | `componentId` | 获取组件状态 |
| `tui.list` | - | 列出所有组件 |
| `tui.quit` | - | 退出 TUI |
| `tui.refresh` | - | 刷新屏幕 |
| `tui.focus.next` | - | 移动焦点到下一个 |
| `tui.focus.prev` | - | 移动焦点到上一个 |
| `tui.focus.set` | `componentId` | 设置焦点到指定组件 |
| `tui.state.update` | `componentId, state` | 更新组件状态 |
| `tui.state.batch` | `updates[]` | 批量更新状态 |
| `tui.event.publish` | `event, data` | 发布事件 |
| `tui.message.send` | `msg` | 发送消息 |
| `tui.message.targeted` | `componentId, msg` | 发送目标消息 |
| `tui.data.set` | `key, value` | 设置数据 |
| `tui.data.get` | `key` | 获取数据 |
| `tui.component.add` | `component` | 添加组件 |
| `tui.component.remove` | `componentId` | 移除组件 |
| `tui.component.show` | `componentId` | 显示组件 |
| `tui.component.hide` | `componentId` | 隐藏组件 |

### Process 中调用 TUI 示例

```javascript
// 在 Yao Script 中调用 TUI

// 1. 更新组件状态
function afterSave(data) {
    Process('tui.state.update', 'user-form', {
        status: 'saved',
        data: data
    });
}

// 2. 刷新列表
function refreshList() {
    var users = Process('models.user.List', 1, 20);
    Process('tui.state.update', 'user-table', {
        data: users
    });
}

// 3. 显示通知
function showNotification(message, type) {
    Process('tui.component.show', 'notification');
    Process('tui.state.update', 'notification', {
        message: message,
        type: type
    });
}

// 4. 设置焦点
function focusNext() {
    Process('tui.focus.next');
}

// 5. 发布事件
function publishEvent() {
    Process('tui.event.publish', 'data.changed', {
        source: 'user-save',
        timestamp: Date.now()
    });
}
```

## 错误处理

### 1. Process 执行错误

```go
// 在 Component 中处理
func (c *ActionableComponent) HandleAction(a *runtime.Action) bool {
    config, ok := c.processMap.Get(a.Type)
    if !ok {
        return false
    }

    result, err := c.processBridge.ExecuteSync(config.Process, config.Transform(a)...)

    if err != nil {
        // 1. 设置错误状态
        c.SetState(map[string]interface{}{
            "error": err.Error(),
        })

        // 2. 触发错误 Action
        c.EmitAction("error", map[string]interface{}{
            "action": a.Type,
            "error":  err.Error(),
        })

        // 3. 可选：显示错误通知
        c.processBridge.Execute("tui.notification.show", map[string]interface{}{
            "type":    "error",
            "message": err.Error(),
        })

        return true
    }

    c.handleResult(result)
    return true
}
```

### 2. 超时处理

```go
// 带超时的 Process 执行
func (b *ProcessBridge) ExecuteWithTimeout(name string, timeout time.Duration, args ...interface{}) (interface{}, error) {
    p := process.New(name, args...)

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // 使用 context 控制
    done := make(chan error)
    var result interface{}

    go func() {
        var err error
        result, err = p.ExecuteSync()
        done <- err
    }()

    select {
    case err := <-done:
        return result, err
    case <-ctx.Done():
        return nil, fmt.Errorf("process %s timeout after %v", name, timeout)
    }
}
```

### 3. 重试机制

```go
// 带重试的 Process 执行
func (b *ProcessBridge) ExecuteWithRetry(name string, maxRetries int, args ...interface{}) (interface{}, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        p := process.New(name, args...)
        result, err := p.ExecuteSync()
        if err == nil {
            return result, nil
        }

        lastErr = err

        // 指数退避
        time.Sleep(time.Duration(1<<i) * time.Second)
    }

    return nil, fmt.Errorf("process %s failed after %d retries: %w", name, maxRetries, lastErr)
}
```

## 异步状态管理

### 1. 加载状态

```go
// AsyncComponent 支持异步操作的组件
type AsyncComponent struct {
    *ActionableComponent
    loading bool
}

// HandleAction 处理异步 Action
func (c *AsyncComponent) HandleAction(a *runtime.Action) bool {
    config, ok := c.processMap.Get(a.Type)
    if !ok {
        return false
    }

    // 设置加载状态
    c.loading = true
    c.SetState(map[string]interface{}{
        "loading": true,
    })
    c.MarkDirty()

    // 异步执行
    go func() {
        result, err := c.processBridge.ExecuteSync(config.Process, config.Transform(a)...)

        // 通过消息更新 UI
        tea.Send(ProcessResultMsg{
            Process: config.Process,
            Result:  result,
            Error:   err,
        })
    }()

    return true
}
```

### 2. 进度反馈

```go
// ProcessWithProgress 带进度的 Process 执行
type ProcessWithProgress struct {
    progress float64
    status   string
}

// 在 Process 中报告进度
func reportProgress(processID string, progress float64, status string) {
    process.New("tui.state.update", "progress-bar", map[string]interface{}{
        "processID": processID,
        "progress":  progress,
        "status":    status,
    }).Execute()
}

// 使用示例
func longRunningTask() {
    reportProgress("task-1", 0.0, "开始处理...")

    // 步骤 1
    time.Sleep(1 * time.Second)
    reportProgress("task-1", 0.25, "加载数据...")

    // 步骤 2
    time.Sleep(1 * time.Second)
    reportProgress("task-1", 0.5, "处理数据...")

    // 步骤 3
    time.Sleep(1 * time.Second)
    reportProgress("task-1", 0.75, "保存结果...")

    // 完成
    time.Sleep(1 * time.Second)
    reportProgress("task-1", 1.0, "完成")
}
```

## V8 上下文亲和性

### 为什么需要 V8 上下文亲和性

V8 Isolate 不是线程安全的，当 TUI 运行在 V8 环境中时，Process 执行需要：

1. **共享 V8 上下文**: 保持变量、函数、模块的访问
2. **内存效率**: 避免创建多个 Isolate
3. **状态一致性**: 确保全局状态的正确性

### 实现方式

```go
// 在 V8 环境中执行 Process
func (b *ProcessBridge) ExecuteInV8Context(name string, v8Ctx *v8.Context, args ...interface{}) (interface{}, error) {
    // 将参数转换为 JS 值
    jsArgs := make([]*v8go.Value, len(args))
    for i, arg := range args {
        val, err := bridge.JsValue(v8Ctx, arg)
        if err != nil {
            return nil, err
        }
        jsArgs[i] = val
    }

    // 创建 Process 并绑定 V8 上下文
    p := process.New(name, args...)
    p = p.WithV8Context(v8Ctx)

    // 同步执行（保持 V8 上下文）
    return p.ExecuteSync()
}
```

## 最佳实践

### 1. 命名规范

```
// Process 名称使用点分命名
models.user.Find      // ✅ 清晰
UserFind              // ❌ 不规范

// Action 使用动词
data.load             // ✅ 清晰
data                  // ❌ 不清楚是什么操作
```

### 2. 错误传播

```go
// Process 中返回结构化错误
function saveUser(data) {
    if (!data.name) {
        return {
            error: true,
            code: 'VALIDATION_ERROR',
            message: '用户名不能为空',
            field: 'name'
        };
    }
    // ...
}

// 在 Component 中处理
func (c *Component) handleResult(result interface{}) {
    if m, ok := result.(map[string]interface{}); ok {
        if m["error"].(bool) {
            // 显示字段级错误
            c.SetFieldError(m["field"].(string), m["message"].(string))
        }
    }
}
```

### 3. 数据验证

```go
// 在 Process 中验证数据
function validateAndSave(data) {
    // 1. 验证
    var errors = Validate(data);
    if (errors.length > 0) {
        return {
            error: true,
            code: 'VALIDATION_ERROR',
            errors: errors
        };
    }

    // 2. 保存
    var result = Process('models.user.Save', data);

    // 3. 返回结果
    return {
        error: false,
        data: result
    };
}
```

### 4. 乐观更新

```go
// 立即更新 UI，后台执行 Process
func (c *Component) HandleAction(a *runtime.Action) bool {
    switch a.Type {
    case "toggle.like":
        // 1. 立即更新 UI
        c.SetState(map[string]interface{}{
            "liked": !c.State["liked"].(bool),
        })

        // 2. 后台执行 Process
        go func() {
            _, err := c.processBridge.ExecuteSync("models.user.ToggleLike", c.ID())
            if err != nil {
                // 失败时回滚
                c.EmitAction("rollback", nil)
            }
        }()

        return true
    }
    return false
}
```

### 5. 批量操作

```go
// 批量更新状态
func batchUpdate(updates []StateUpdate) {
    process.New("tui.state.batch", updates).Execute()
}

// 使用
batchUpdate([]StateUpdate{
    {Component: "table-1", State: map[string]interface{}{"data": data1}},
    {Component: "status-1", State: map[string]interface{}{"text": "已完成"}},
    {Component: "button-1", State: map[string]interface{}{"disabled": false}},
})
```

## 完整示例：用户管理界面

```go
package main

import (
    "github.com/yaoapp/yao/tui/framework"
    "github.com/yaoapp/yao/tui/component"
    "github.com/yaoapp/yao/tui/runtime"
)

// UserListScreen 用户列表界面
type UserListScreen struct {
    *framework.BaseScreen
    table    *component.Table
    toolbar  *component.Toolbar
    bridge   *framework.ProcessBridge
}

// NewUserListScreen 创建用户列表界面
func NewUserListScreen() *UserListScreen {
    screen := &UserListScreen{
        BaseScreen: framework.NewBaseScreen("user-list"),
    }

    // 创建 Process 桥接
    screen.bridge = framework.NewProcessBridge(screen)

    // 创建表格
    screen.table = component.NewTable("user-table", []component.Column{
        {Key: "id", Title: "ID", Width: 8},
        {Key: "name", Title: "姓名", Width: 20, Editable: true},
        {Key: "email", Title: "邮箱", Width: 30, Editable: true},
        {Key: "status", Title: "状态", Width: 10},
    })
    screen.table.SetProcessBridge(screen.bridge)

    // 配置 Action → Process 映射
    screen.table.ProcessMap().Register(&framework.ActionToProcessConfig{
        Action:  "load",
        Process: "models.user.List",
        Async:   true,
    })

    screen.table.ProcessMap().Register(&framework.ActionToProcessConfig{
        Action:  "edit",
        Process: "models.user.Update",
        Async:   true,
    })

    screen.table.ProcessMap().Register(&framework.ActionToProcessConfig{
        Action:  "delete",
        Process: "models.user.Delete",
        Async:   true,
    })

    // 创建工具栏
    screen.toolbar = component.NewToolbar("toolbar", []component.ToolButton{
        {Label: "刷新", Action: "refresh"},
        {Label: "新建", Action: "create"},
        {Label: "删除", Action: "delete"},
    })
    screen.toolbar.SetProcessBridge(screen.bridge)

    screen.toolbar.ProcessMap().Register(&framework.ActionToProcessConfig{
        Action:  "refresh",
        Process: "models.user.List",
        Async:   true,
    })

    screen.toolbar.ProcessMap().Register(&framework.ActionToProcessConfig{
        Action:  "create",
        Process: "tui.navigate.to",
        Transform: func(a *runtime.Action) []interface{} {
            return []interface{}{"user-create"}
        },
        Async: true,
    })

    return screen
}

// 初始化加载数据
func (s *UserListScreen) OnInit() {
    s.table.EmitAction("load", map[string]interface{}{
        "page": 1,
        "size": 20,
    })
}
```

## 复用度评估

| 模块 | 现有实现 | 复用度 | 需要补充 |
|------|----------|--------|----------|
| Process 核心 | gou/process | 100% | 无 |
| Process 注册 | gou/process.Handler | 100% | 无 |
| V8 上下文亲和 | gou/process.WithV8Context | 100% | 无 |
| TUI Process | yao/tui/process | 100% | 无 |
| 类型转换 | gou/runtime/v8/bridge | 100% | 无 |
| Action 映射 | 新增 | 0% | ✅ ProcessRegistry |
| Process Bridge | 新增 | 0% | ✅ ProcessBridge |
| 错误处理 | 新增 | 0% | ✅ 错误处理模式 |
| 超时/重试 | 新增 | 0% | ✅ 超时/重试包装 |

**总体复用度**: 95%

**新增代码量**: 约 500 行

## 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构总览
- [ACTION_SYSTEM.md](ACTION_SYSTEM.md) - Action 系统
- [V8_INTEGRATION_YAO.md](V8_INTEGRATION_YAO.md) - V8 集成
- [COMPONENTS.md](COMPONENTS.md) - 组件系统
- [ERROR_HANDLING.md](ERROR_HANDLING.md) - 错误处理
