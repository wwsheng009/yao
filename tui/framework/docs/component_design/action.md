
### 43. 动作系统设计 (Action System)

不要在组件内部硬编码业务逻辑。组件应该发出一个 **Action**，由顶层 Runtime 解释执行。

#### Action 定义

我们需要一个通用的 Action 结构来描述“意图”。

```go
// framework/action/types.go

type Type string

const (
    // 内置 UI 行为
    ActPushRoute  Type = "Route.Push"  // 跳转页面
    ActPopRoute   Type = "Route.Pop"   // 返回
    ActSetState   Type = "State.Set"   // 修改本地状态
    
    // Yao 业务行为
    ActRunProcess Type = "Service.Run" // 调用 Yao Process (e.g., models.user.save)
    ActEmitEvent  Type = "Event.Emit"  // 发送事件给父组件
)

type Action struct {
    Type    Type                   `json:"type"`
    Payload map[string]interface{} `json:"payload"`
    Handler func(ctx *Context)     `json:"-"` // 可选：Go 原生回调
}

// 示例：定义一个保存动作
func NewProcessAction(process string, args ...interface{}) Action {
    return Action{
        Type: ActRunProcess,
        Payload: map[string]interface{}{
            "process": process,
            "args":    args,
        },
    }
}

```

---

### 44. 组件的事件绑定 (Event Binding)

在组件中，我们通常有 `OnClick`, `OnSubmit`, `OnChange` 等钩子。这些钩子不应该直接执行代码，而是返回一个 `Action` 列表。

#### 修改 BaseComponent 或具体组件

让组件能够存储和触发 Action。

```go
// framework/component/component.go 扩展

type Component interface {
    // ... 原有接口
    GetAction(trigger string) *action.Action
}

// Button 组件示例
type Button struct {
    *component.BaseComponent
    text    string
    onClick *action.Action // 点击时触发的动作
}

func (b *Button) SetOnClick(act action.Action) {
    b.onClick = &act
}

// 处理按键或鼠标点击
func (b *Button) HandleAction(a action.Action) bool {
    // 如果是回车键或鼠标左键
    if a.Type == action.ActionSubmit || a.Type == action.ActionClick {
        if b.onClick != nil {
            // 关键：这里我们不直接执行，而是通过 dispatch 发送出去
            // 但 HandleAction 的返回值通常是 bool。
            // 我们需要一个机制向上冒泡这个“业务动作”
            b.GetContext().Dispatch(*b.onClick)
            return true
        }
    }
    return false
}

```

---

### 45. 运行时桥接 (Runtime Bridge)

这是连接 TUI 和 Yao 内核（Gou 引擎）的桥梁。你需要一个 `ActionExecutor`。

```go
// runtime/executor.go

type Executor struct {
    engine *gou.Engine // 假设这是 Yao 的处理器引擎引用
    store  *state.Store
}

func (e *Executor) Execute(ctx context.Context, act action.Action) error {
    switch act.Type {
    
    case action.ActRunProcess:
        // 1. 解析参数
        name := act.Payload["process"].(string)
        args := act.Payload["args"].([]interface{})
        
        // 2. 处理参数绑定 (例如把 {{form.data}} 解析为实际值)
        resolvedArgs := e.resolveArgs(args)
        
        // 3. 调用 Yao Process (需要在 Goroutine 中运行以防阻塞 UI)
        go func() {
            res, err := gou.Process(name, resolvedArgs...)
            if err != nil {
                e.ShowToast("Error: " + err.Error(), "error")
                return
            }
            
            // 4. 处理回调 (Callback)
            if callback, ok := act.Payload["onSuccess"]; ok {
                e.Execute(ctx, callback.(action.Action))
            }
        }()

    case action.ActPushRoute:
        path := act.Payload["path"].(string)
        e.NavigateTo(path)
        
    case action.ActSetState:
        key := act.Payload["key"].(string)
        val := act.Payload["value"]
        e.store.Set(key, val) // 这会自动触发 UI 绑定的重绘
    }
    return nil
}

```

---

### 46. 表单提交完整链路 (The Full Form Submission Flow)

让我们把之前的所有设计串起来，看看一个“保存表单”的流程是如何工作的。

**场景**：用户在一个输入框输入姓名，点击“保存”按钮，调用 `models.user.Save`。

1. **DSL 定义 (界面描述)**:
```json
{
  "type": "Form",
  "state": { "data": {} }, // 本地状态
  "children": [
    {
      "type": "Input",
      "props": { "value": "{{data.name}}", "onChange": { "type": "State.Set", "key": "data.name" } }
    },
    {
      "type": "Button",
      "props": {
        "text": "Save",
        "onClick": {
          "type": "Service.Run",
          "process": "models.user.Save",
          "args": [ "{{data}}" ], // 将整个 data 对象作为参数
          "onSuccess": { "type": "Route.Pop" } // 保存成功后返回
        }
      }
    }
  ]
}

```


2. **解析与绑定 (Init)**:
* `Form` 初始化 `Store`，key 为 `data`。
* `Input` 的 `valueProp` 绑定到 `Store["data.name"]`。
* `Button` 解析 `onClick` 配置，生成 `Action` 结构体。


3. **用户输入 (Interaction)**:
* 用户按键 -> `Input` 处理按键 -> 触发 `onChange` Action -> `Store` 更新 `data.name` -> `Input` (以及其他绑定了该值的组件) 被标记为 Dirty -> `Paint` 重绘显示新字符。


4. **提交 (Submission)**:
* 用户点击 Button -> `Button` 发出 `ActRunProcess`。
* **Runtime 拦截**:
* 发现参数包含 `{{data}}`。
* 从 `Store` 中提取当前 `data` 的完整 JSON 对象：`{"name": "Alice"}`。
* 调用 Go 函数: `gou.Process("models.user.Save", map[string]interface{}{"name": "Alice"})`。


* **回调**: Process 返回成功 -> Runtime 执行 `onSuccess` Action (`Route.Pop`) -> 界面切换。



---

### 47. 键盘导航系统 (Keymap System)

为了让应用像 Vim 或 Emacs 一样高效，我们需要一个集中的键盘映射系统，而不是让每个组件硬编码按键。

#### 设计 Keymap

```go
// framework/input/keymap.go

type KeyBinding struct {
    Keys   []string // e.g. ["enter", "ctrl+s"]
    Action string   // e.g. "submit", "save"
    Desc   string   // 用于显示帮助栏
}

type ComponentKeyMap struct {
    Shortcuts []KeyBinding
}

// 默认映射
var ListKeyMap = ComponentKeyMap{
    Shortcuts: []KeyBinding{
        {Keys: []string{"j", "down"}, Action: "next_item", Desc: "Select next"},
        {Keys: []string{"k", "up"}, Action: "prev_item", Desc: "Select prev"},
        {Keys: []string{"/", "ctrl+f"}, Action: "search", Desc: "Search"},
    },
}

```