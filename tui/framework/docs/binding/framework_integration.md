# 在 Framework 中应用 Binding 模块

## 目录

1. [快速开始](#快速开始)
2. [扩展组件支持绑定](#扩展组件支持绑定)
3. [创建绑定组件](#创建绑定组件)
4. [响应式应用](#响应式应用)
5. [DSL 集成](#dsl-集成)
6. [最佳实践](#最佳实践)

---

## 快速开始

### 步骤 1: 导入模块

```go
import (
    "github.com/yaoapp/yao/tui/framework/binding"
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
    "github.com/yaoapp/yao/tui/framework/component"
)
```

### 步骤 2: 创建响应式存储

```go
// 创建全局状态存储
store := binding.NewReactiveStore()

// 设置初始数据
store.Set("app", map[string]interface{}{
    "title": "My Application",
    "version": "1.0.0",
})
store.Set("user", map[string]interface{}{
    "name": "Alice",
    "role": "admin",
})
```

### 步骤 3: 创建绑定组件

```go
// 创建支持绑定的标签组件
type Label struct {
    *cb.BaseBindable
    textProp binding.Prop[string]
}

func NewLabel(text string) *Label {
    return &Label{
        BaseBindable: cb.NewBaseBindable("label"),
        textProp:     binding.NewStatic(text),
    }
}

// 支持绑定语法
func (l *Label) SetText(text string) *Label {
    l.textProp = binding.NewStringProp(text)
    return l
}
```

### 步骤 4: 实现绘制

```go
func (l *Label) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 创建绑定上下文
    bindCtx := cb.CreateBindingContext(l)

    // 解析属性值
    text := l.textProp.Resolve(bindCtx)

    // 绘制文本
    for i, r := range text {
        ctx.SetCell(ctx.X+i, ctx.Y, r)
    }
}
```

---

## 扩展组件支持绑定

### 现有组件改造

以 `TextInput` 为例，展示如何添加绑定支持：

```go
// input/textinput.go

type TextInput struct {
    *component.BaseComponent
    *component.StateHolder

    // 新增：绑定属性
    valueProp binding.Prop[string]
    placeholderProp binding.Prop[string]
}

// 修改 SetValue 支持自动检测绑定
func (t *TextInput) SetValue(value string) *TextInput {
    t.valueProp = binding.NewStringProp(value)
    return t
}

// 新增：从 Store 同步值
func (t *TextInput) SyncFromStore(ctx binding.Context) {
    if t.valueProp.IsBound() {
        value := t.valueProp.Resolve(ctx)
        t.mu.Lock()
        t.value = value
        t.mu.Unlock()
    }
}
```

### 添加监听支持

```go
// 扩展组件，添加 Store 监听
type BindableTextInput struct {
    *TextInput
    store     *binding.ReactiveStore
    bindPath  string
    cancel    func()
}

func (bt *BindableTextInput) WatchStore() {
    bt.cancel = bt.store.Subscribe(bt.bindPath, func(key string, old, new interface{}) {
        // Store 变化时更新组件
        bt.SetValue(fmt.Sprintf("%v", new))
        bt.MarkDirty()
    })
}
```

---

## 创建绑定组件

### Label 组件

```go
package display

import (
    "github.com/yaoapp/yao/tui/framework/binding"
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
    "github.com/yaoapp/yao/tui/framework/component"
)

type Label struct {
    *cb.BaseBindable
    textProp   binding.Prop[string]
    alignProp  binding.Prop[string]
    styleProp  binding.Prop[string]
}

func NewLabel(text string) *Label {
    return &Label{
        BaseBindable: cb.NewBaseBindable("label"),
        textProp:     binding.NewStatic(text),
        alignProp:    binding.NewStatic("left"),
        styleProp:    binding.NewStatic("default"),
    }
}

// 从 DSL 创建
func NewLabelFromDSL(dsl map[string]interface{}) *Label {
    label := NewLabel("")

    // 解析并设置属性
    if text, ok := dsl["text"].(string); ok {
        label.SetText(text)
    }

    if align, ok := dsl["align"].(string); ok {
        label.SetAlign(align)
    }

    if bind, ok := dsl["bind"].(string); ok {
        label.textProp = binding.NewBinding[string](bind)
    }

    return label
}

func (l *Label) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    bindCtx := cb.CreateBindingContext(l)

    text := l.textProp.Resolve(bindCtx)
    align := l.alignProp.Resolve(bindCtx)

    // 根据对齐绘制
    // ...
}
```

### Button 组件

```go
package interactive

import (
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
)

type Button struct {
    *cb.BaseBindable
    textProp    binding.Prop[string]
    disabledProp binding.Prop[bool]
    onClick     func()
}

func NewButton(text string, onClick func()) *Button {
    return &Button{
        BaseBindable: cb.NewBaseBindable("button"),
        textProp:     binding.NewStatic(text),
        disabledProp: binding.NewStatic(false),
        onClick:      onClick,
    }
}

func (b *Button) SetText(text string) *Button {
    b.textProp = binding.NewStringProp(text)
    return b
}

func (b *Button) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    bindCtx := cb.CreateBindingContext(b)

    text := b.textProp.Resolve(bindCtx)
    disabled := b.disabledProp.Resolve(bindCtx)

    // 根据状态绘制
    if disabled {
        // 绘制禁用状态
    } else {
        // 绘制正常状态
    }
}

func (b *Button) HandleAction(a action.Action) bool {
    if a.Type == action.ActionSelect {
        bindCtx := cb.CreateBindingContext(b)
        disabled := b.disabledProp.Resolve(bindCtx)

        if !disabled && b.onClick != nil {
            b.onClick()
            return true
        }
    }
    return false
}
```

---

## 响应式应用

### 应用结构

```go
type App struct {
    store *binding.ReactiveStore

    // 组件
    header   *Label
    content  *ReactiveList
    footer   *Label
}

func NewApp() *App {
    store := binding.NewReactiveStore()

    app := &App{store: store}

    // 初始化数据
    store.Set("app", map[string]interface{}{
        "title": "Task Manager",
        "status": "active",
    })
    store.Set("tasks", []interface{}{
        map[string]interface{}{"id": 1, "text": "Task 1", "done": false},
        map[string]interface{}{"id": 2, "text": "Task 2", "done": true},
    })

    // 创建组件
    app.header = NewLabelBinding("app.title")
    app.content = NewReactiveList("tasks")
    app.footer = NewLabelBinding("app.status")

    // 订阅变化
    app.setupSubscriptions()

    return app
}

func (a *App) setupSubscriptions() {
    // 标题变化时重绘
    a.store.Subscribe("app.title", func(key, old, new interface{}) {
        a.header.MarkDirty()
    })

    // 任务列表变化时重绘
    a.store.Subscribe("tasks", func(key, old, new interface{}) {
        a.content.MarkDirty()
    })
}
```

### 状态管理

```go
// 集中式状态更新
func (a *App) AddTask(text string) {
    tasks, _ := a.store.Get("tasks")
    newTasks := append(tasks.([]interface{}), map[string]interface{}{
        "id":   len(tasks.([]interface{})) + 1,
        "text": text,
        "done": false,
    })

    a.store.Set("tasks", newTasks)
}

// 计算属性
func (a *App) GetCompletedCount() int {
    tasks, _ := a.store.Get("tasks")
    count := 0
    for _, item := range tasks.([]interface{}) {
        if task, ok := item.(map[string]interface{}); ok {
            if done, ok := task["done"].(bool); ok && done {
                count++
            }
        }
    }
    return count
}

// 使用计算属性
completedCount := binding.NewStoreComputed(a.store,
    []string{"tasks"},
    func() interface{} {
        return a.GetCompletedCount()
    },
)
```

---

## DSL 集成

### DSL 定义

```json
{
    "type": "form",
    "bind": "formData",
    "fields": [
        {
            "type": "input",
            "id": "username",
            "label": "用户名",
            "bind": "user.name",
            "props": {
                "placeholder": "请输入用户名"
            }
        },
        {
            "type": "input",
            "id": "password",
            "label": "密码",
            "bind": "user.password",
            "props": {
                "password": true
            }
        }
    ]
}
```

### DSL 解析器

```go
type DSLParser struct {
    store *binding.ReactiveStore
}

func (p *DSLParser) Parse(dsl map[string]interface{}) component.Node {
    typ, _ := dsl["type"].(string)

    switch typ {
    case "label":
        return p.parseLabel(dsl)
    case "input":
        return p.parseInput(dsl)
    case "button":
        return p.parseButton(dsl)
    case "form":
        return p.parseForm(dsl)
    default:
        return nil
    }
}

func (p *DSLParser) parseLabel(dsl map[string]interface{}) component.Node {
    label := NewLabel("")

    // 处理绑定
    if bind, ok := dsl["bind"].(string); ok {
        label.SetText(bind)  // 自动检测 {{ }} 语法
    } else if text, ok := dsl["text"].(string); ok {
        label.SetText(text)
    }

    // 处理其他属性
    if align, ok := dsl["align"].(string); ok {
        label.SetAlign(align)
    }

    return label
}

func (p *DSLParser) parseInput(dsl map[string]interface{}) component.Node {
    input := input.NewTextInput()

    // 绑定 Store
    if bind, ok := dsl["bind"].(string); ok {
        // 创建绑定适配器
        adapter := cb.NewStoreToComponentAdapter(
            p.store,
            input,
            map[string]string{"value": bind},
        )

        // 初始同步
        adapter.Sync()

        // 监听变化
        adapter.Watch(func() {
            input.MarkDirty()
        })
    }

    // 处理属性
    if props, ok := dsl["props"].(map[string]interface{}); ok {
        if placeholder, ok := props["placeholder"].(string); ok {
            input.SetPlaceholder(placeholder)
        }
        if password, ok := props["password"].(bool); ok && password {
            input.SetPassword(true)
        }
    }

    return input
}
```

---

## 最佳实践

### 1. 分离关注点

```go
// 好：分离数据、逻辑、视图
type App struct {
    store *binding.ReactiveStore  // 数据层
    view  *View                   // 视图层
}

// 避免：组件直接持有业务逻辑
```

### 2. 使用类型安全的 Prop

```go
// 好：使用泛型 Prop
type MyComponent struct {
    textProp binding.Prop[string]
    countProp binding.Prop[int]
}

// 避免：使用 interface{}
type MyComponent struct {
    props map[string]interface{}
}
```

### 3. 合理使用作用域

```go
// 为列表创建临时作用域
func (l *List) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    items := l.itemsProp.Resolve(ctx.Data)

    for i, item := range items {
        // 每项独立作用域
        itemCtx := ctx.Data.New(map[string]interface{}{
            "$index": i,
            "$item":  item,
        })

        renderItem(itemCtx)
    }
}
```

### 4. 及时清理订阅

```go
func (c *Component) Mount(parent component.Container) {
    c.BaseComponent.Mount(parent)

    // 保存取消函数
    c.cancel = c.store.Subscribe("key", c.handler)
}

func (c *Component) Unmount() {
    // 清理订阅
    if c.cancel != nil {
        c.cancel()
    }

    c.BaseComponent.Unmount()
}
```

### 5. 批量更新优化

```go
// 好：使用批量更新
func (a *App) UpdateMultiple() {
    a.store.BeginBatch()
    a.store.Set("key1", "value1")
    a.store.Set("key2", "value2")
    a.store.Set("key3", "value3")
    a.store.EndBatch()  // 只触发一次通知
}

// 避免：多次单独更新
func (a *App) UpdateMultiple() {
    a.store.Set("key1", "value1")  // 触发通知
    a.store.Set("key2", "value2")  // 触发通知
    a.store.Set("key3", "value3")  // 触发通知
}
```

---

## 完整示例：待办事项应用

```go
package main

import (
    "github.com/yaoapp/yao/tui/framework/binding"
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
)

// TodoApp 待办事项应用
type TodoApp struct {
    store *binding.ReactiveStore

    // 组件
    title    *Label
    taskList *TaskList
    addButton *Button
}

func NewTodoApp() *TodoApp {
    app := &TodoApp{
        store: binding.NewReactiveStore(),
    }

    // 初始化数据
    app.store.Set("app", map[string]interface{}{
        "title": "我的待办",
        "count": 0,
    })
    app.store.Set("tasks", []interface{}{})

    // 创建组件
    app.title = NewLabelBinding("app.title")
    app.taskList = NewTaskList("tasks")
    app.addButton = NewButton("添加任务", app.addTask)

    return app
}

func (a *TodoApp) addTask() {
    tasks, _ := a.store.Get("tasks")
    newTasks := append(tasks.([]interface{}), map[string]interface{}{
        "id":   len(tasks.([]interface{})) + 1,
        "text": "新任务",
        "done": false,
    })

    // 批量更新
    a.store.BeginBatch()
    a.store.Set("tasks", newTasks)
    a.store.Set("app.count", len(newTasks))
    a.store.EndBatch()
}

// TaskList 任务列表组件
type TaskList struct {
    *cb.BaseBindable
    itemsProp binding.Prop[[]interface{}]
}

func NewTaskList(path string) *TaskList {
    return &TaskList{
        BaseBindable: cb.NewBaseBindable("tasklist"),
        itemsProp:    binding.NewBinding[[]interface{}](path),
    }
}

func (tl *TaskList) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    bindCtx := cb.CreateBindingContext(tl)
    items := tl.itemsProp.Resolve(bindCtx)

    for i, item := range items {
        task := item.(map[string]interface{})
        text, _ := task["text"].(string)
        done, _ := task["done"].(bool)

        // 渲染任务
        prefix := "[ ]"
        if done {
            prefix = "[x]"
        }

        row := fmt.Sprintf("%s %s", prefix, text)
        for j, r := range row {
            ctx.SetCell(ctx.X+j, ctx.Y+i, r)
        }
    }
}
```

---

## API 速查表

### 创建 Prop

| 方法 | 说明 |
|------|------|
| `binding.NewStatic(v)` | 静态值 |
| `binding.NewBinding(p)` | 数据绑定 |
| `binding.NewStringProp(s)` | 自动检测 |
| `binding.NewExpression(e)` | 表达式 |

### Context 操作

| 方法 | 说明 |
|------|------|
| `ctx.Get(path)` | 获取值 |
| `ctx.Set(path, v)` | 设置值 |
| `ctx.New(data)` | 创建子作用域 |
| `store.ToContext()` | Store → Context |

### Store 操作

| 方法 | 说明 |
|------|------|
| `store.Set(k, v)` | 设置值 |
| `store.Get(k)` | 获取值 |
| `store.Subscribe(k, fn)` | 订阅变化 |
| `store.BeginBatch()` | 开始批量 |
| `store.EndBatch()` | 结束批量 |
