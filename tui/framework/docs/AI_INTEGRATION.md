# AI Integration Standard (V3)

> **版本**: V3
> **核心原则**: AI 是一级操作者，不是外部模拟器
> **关键特性**: 语义化接口、结构化状态、无需截图/OCR

## 概述

本框架专为 AI Agent 和自动化设计了原生接口。AI 不被视为"外部模拟器"，而是与人类用户平级的"一级操作者"。

本文档定义了 AI Agent 与 TUI Runtime 交互的标准协议。

### 传统 AI 自动化的问题

**传统方式（基于截图/OCR）**：
```
1. 截屏 → 2. OCR 识别 → 3. 分析文本 → 4. 计算坐标 → 5. 模拟输入
```

问题：
- 不稳定（字体、颜色变化影响 OCR）
- 慢（截图和 OCR 耗时）
- 脆弱（UI 布局变化就失效）
- 无法获取内部状态（如 disabled、验证错误）

**V3 原生 AI 接口**：
```
1. Inspect() 获取结构化状态 → 2. Dispatch(语义 Action)
```

优势：
- 稳定（直接访问数据结构）
- 快速（无需图像处理）
- 健壮（布局变化不影响）
- 完整状态（包括 disabled、hidden 等）

## 交互模型

```
Human User      AI Agent
    │              │
    ▼              ▼
[ KeyMap ]    [ Controller ]
    │              │
    └─────┬────────┘
          ▼
   [ Action Dispatcher ]
          │
          ▼
   [ Component Tree ]
          │
          ▼
   [ State Snapshot ]
          │
          ▼
    Feed back to AI
```

## AI Controller 接口

### 核心接口

```go
// 位于: tui/runtime/ai/controller.go

package ai

// Controller AI 控制器接口
type Controller interface {
    // === 感知能力 ===

    // Inspect 获取当前完整的 UI 状态快照
    Inspect() (*state.Snapshot, error)

    // Find 查找组件（类似 DOM 选择器）
    Find(selector string) ([]ComponentInfo, error)

    // Query 查询状态
    Query(query StateQuery) (map[string]interface{}, error)

    // WaitUntil 等待特定状态出现
    WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error

    // === 操作能力 ===

    // Dispatch 向指定组件发送语义指令
    Dispatch(a *action.Action) error

    // Click 点击组件
    Click(componentID string) error

    // Input 输入文本
    Input(componentID, text string) error

    // Navigate 焦点导航
    Navigate(direction focus.Direction) error

    // === 高级能力 ===

    // Execute 执行复杂操作序列
    Execute(ops ...Operation) error

    // Watch 监控状态变化
    Watch(callback func(*state.Snapshot)) func()
}
```

### ComponentInfo

```go
// ComponentInfo 组件信息
type ComponentInfo struct {
    // 基本信息
    ID   string
    Type string

    // 状态
    Props    map[string]interface{} // 静态属性
    State    map[string]interface{} // 动态状态
    Rect     Rect                   // 布局位置
    Visible  bool                   // 可见性
    Disabled bool                   // 可交互性

    // 父子关系
    ParentID string
    Children []string
}
```

### StateQuery

```go
// StateQuery 状态查询
type StateQuery struct {
    // 组件筛选
    ComponentID   string
    ComponentType string

    // 状态键筛选
    StateKey string

    // 值筛选
    Value interface{}
}

// QueryResult 查询结果
type QueryResult struct {
    Components map[string]ComponentInfo
    Values     map[string]interface{}
    Count      int
}
```

### Operation

```go
// Operation 操作
type Operation interface {
    Execute(ctrl Controller) error
}

// ClickOperation 点击操作
type ClickOperation struct {
    ComponentID string
}

func (op *ClickOperation) Execute(ctrl Controller) error {
    return ctrl.Click(op.ComponentID)
}

// InputOperation 输入操作
type InputOperation struct {
    ComponentID string
    Text        string
}

func (op *InputOperation) Execute(ctrl Controller) error {
    return ctrl.Input(op.ComponentID, op.Text)
}

// WaitOperation 等待操作
type WaitOperation struct {
    Condition func(*state.Snapshot) bool
    Timeout   time.Duration
}

func (op *WaitOperation) Execute(ctrl Controller) error {
    return ctrl.WaitUntil(op.Condition, op.Timeout)
}
```

## Runtime 实现

### RuntimeController

```go
// 位于: tui/runtime/ai/runtime_controller.go

package ai

// RuntimeController Runtime 实现的 AI 控制器
type RuntimeController struct {
    runtime    *Runtime
    dispatcher *action.Dispatcher
    tracker    *state.Tracker
    focus      *focus.Manager
}

// NewRuntimeController 创建 Runtime AI 控制器
func NewRuntimeController(rt *Runtime) *RuntimeController {
    return &RuntimeController{
        runtime:    rt,
        dispatcher: rt.dispatcher,
        tracker:    rt.stateTracker,
        focus:      rt.focus,
    }
}

// Inspect 获取当前 UI 状态快照
func (c *RuntimeController) Inspect() (*state.Snapshot, error) {
    return c.tracker.Current(), nil
}

// Find 查找组件
func (c *RuntimeController) Find(selector string) ([]ComponentInfo, error) {
    snapshot := c.tracker.Current()

    // ID 选择器: #input-username
    if strings.HasPrefix(selector, "#") {
        id := selector[1:]
        comp, ok := snapshot.Components[id]
        if !ok {
            return nil, fmt.Errorf("component not found: %s", id)
        }
        return []ComponentInfo{c.toComponentInfo(comp)}, nil
    }

    // 类型选择器: .TextInput
    if strings.HasPrefix(selector, ".") {
        typ := selector[1:]
        return c.findByType(snapshot, typ)
    }

    // 属性选择器: [placeholder="Email"]
    if strings.HasPrefix(selector, "[") && strings.HasSuffix(selector, "]") {
        return c.findByAttribute(snapshot, selector[1:len(selector)-1])
    }

    return nil, fmt.Errorf("invalid selector: %s", selector)
}

// findByType 按类型查找
func (c *RuntimeController) findByType(snapshot *state.Snapshot, typ string) ([]ComponentInfo, error) {
    results := make([]ComponentInfo, 0)
    for _, comp := range snapshot.Components {
        if comp.Type == typ {
            results = append(results, c.toComponentInfo(comp))
        }
    }
    return results, nil
}

// findByAttribute 按属性查找
func (c *RuntimeController) findByAttribute(snapshot *state.Snapshot, attr string) ([]ComponentInfo, error) {
    // 解析属性选择器
    // 例如: placeholder="Email" → key=placeholder, value=Email
    parts := strings.SplitN(attr, "=", 2)
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid attribute selector: %s", attr)
    }
    key := strings.TrimSpace(parts[0])
    value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

    results := make([]ComponentInfo, 0)
    for _, comp := range snapshot.Components {
        if compVal, ok := comp.Props[key]; ok && fmt.Sprintf("%v", compVal) == value {
            results = append(results, c.toComponentInfo(comp))
        }
        if compVal, ok := comp.State[key]; ok && fmt.Sprintf("%v", compVal) == value {
            results = append(results, c.toComponentInfo(comp))
        }
    }
    return results, nil
}

// toComponentInfo 转换为 ComponentInfo
func (c *RuntimeController) toComponentInfo(comp state.ComponentState) ComponentInfo {
    return ComponentInfo{
        ID:       comp.ID,
        Type:     comp.Type,
        Props:    comp.Props,
        State:    comp.State,
        Rect:     comp.Rect,
        Visible:  comp.Visible,
        Disabled: comp.Disabled,
    }
}

// Query 查询状态
func (c *RuntimeController) Query(query StateQuery) (map[string]interface{}, error) {
    snapshot := c.tracker.Current()

    if query.ComponentID != "" {
        comp, ok := snapshot.Components[query.ComponentID]
        if !ok {
            return nil, fmt.Errorf("component not found: %s", query.ComponentID)
        }

        if query.StateKey != "" {
            return map[string]interface{}{
                query.StateKey: comp.State[query.StateKey],
            }, nil
        }

        return comp.State, nil
    }

    if query.ComponentType != "" {
        result := make(map[string]interface{})
        for id, comp := range snapshot.Components {
            if comp.Type == query.ComponentType {
                result[id] = comp.State
            }
        }
        return result, nil
    }

    // 返回所有状态
    result := make(map[string]interface{})
    for id, comp := range snapshot.Components {
        result[id] = comp.State
    }
    return result, nil
}

// Dispatch 分发 Action
func (c *RuntimeController) Dispatch(a *action.Action) error {
    handled := c.dispatcher.Dispatch(a)
    if !handled {
        return fmt.Errorf("action not handled: %s", a.Type)
    }
    return nil
}

// Click 点击组件
func (c *RuntimeController) Click(componentID string) error {
    // 检查组件是否存在且可交互
    snapshot := c.tracker.Current()
    comp, ok := snapshot.Components[componentID]
    if !ok {
        return fmt.Errorf("component not found: %s", componentID)
    }
    if comp.Disabled {
        return fmt.Errorf("component is disabled: %s", componentID)
    }

    // 发送点击 Action
    return c.Dispatch(action.NewAction(action.ActionClick).
        WithTarget(componentID))
}

// Input 输入文本
func (c *RuntimeController) Input(componentID, text string) error {
    // 检查组件是否存在
    snapshot := c.tracker.Current()
    _, ok := snapshot.Components[componentID]
    if !ok {
        return fmt.Errorf("component not found: %s", componentID)
    }

    // 发送输入 Action
    return c.Dispatch(action.NewAction(action.ActionInputText).
        WithTarget(componentID).
        WithPayload(text))
}

// Navigate 焦点导航
func (c *RuntimeController) Navigate(direction focus.Direction) error {
    result := c.focus.Navigate(direction)
    if !result.Success {
        return result.Error
    }
    return nil
}

// Execute 执行操作序列
func (c *RuntimeController) Execute(ops ...Operation) error {
    for _, op := range ops {
        if err := op.Execute(c); err != nil {
            return err
        }
    }
    return nil
}

// WaitUntil 等待条件满足
func (c *RuntimeController) WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(50 * time.Millisecond)
    defer ticker.Stop()

    for {
        snapshot := c.tracker.Current()
        if condition(snapshot) {
            return nil
        }

        if time.Now().After(deadline) {
            return fmt.Errorf("timeout waiting for condition")
        }

        <-ticker.C
    }
}

// Watch 监控状态变化
func (c *RuntimeController) Watch(callback func(*state.Snapshot)) func() {
    return c.tracker.Subscribe(func(old, new *state.Snapshot) {
        callback(new)
    })
}
```

## AI 专用 Action

除了标准 UI Action，框架还提供 AI 专用元指令：

```go
// 位于: tui/runtime/action/ai_actions.go

package action

const (
    // AI 专用 Action
    ActionAIQuery  ActionType = "ai.query"
    ActionAIDump   ActionType = "ai.dump"
    ActionAIReset  ActionType = "ai.reset"
    ActionAIScript ActionType = "ai.script"
)

// AIQueryPayload AI 查询负载
type AIQueryPayload struct {
    Selector string
    StateKey string
}

// AIDumpPayload AI 导出负载
type AIDumpPayload struct {
    Format string // "json", "yaml"
}
```

## 最佳实践

### ✅ 正确做法

```go
// ✅ 使用语义化接口
func FillLoginForm(ai *ai.RuntimeController) error {
    // 1. 获取状态
    snapshot, _ := ai.Inspect()

    // 2. 查找输入框
    usernameInput, _ := ai.Find("#username")
    passwordInput, _ := ai.Find("#password")

    // 3. 检查状态
    if usernameInput[0].Disabled {
        return errors.New("username input is disabled")
    }

    // 4. 填写表单
    ai.Input("username", "user@example.com")
    ai.Input("password", "secret123")

    // 5. 提交
    ai.Click("#submit")

    // 6. 等待结果
    return ai.WaitUntil(func(s *state.Snapshot) bool {
        msg, _ := s.GetComponent("success-message")
        return msg.Visible
    }, 5*time.Second)
}
```

### ❌ 错误做法

```go
// ❌ 不要使用截图
func FillFormBad(ai *ai.AI) error {
    // 1. 截屏
    screenshot := ai.TakeScreenshot()

    // 2. OCR 识别
    text := ai.OCR(screenshot)

    // 3. 分析文本位置
    coords := ai.FindTextCoords(text, "Username")

    // 4. 模拟鼠标点击
    ai.MoveMouse(coords.X, coords.Y)
    ai.Click()

    // 5. 模拟键盘输入
    for _, ch := range "user@example.com" {
        ai.KeyPress(ch)
    }
}
```

## 使用示例

### 示例 1：表单填写

```go
func ExampleFillForm(ai *ai.RuntimeController) error {
    return ai.Execute(
        &InputOperation{"username", "user@example.com"},
        &InputOperation{"password", "secret123"},
        &ClickOperation{"submit"},
        &WaitOperation{
            Condition: func(s *state.Snapshot) bool {
                if comp, ok := s.GetComponent("success-msg"); ok {
                    return comp.Visible
                }
                return false
            },
            Timeout: 5 * time.Second,
        },
    )
}
```

### 示例 2：列表导航

```go
func ExampleNavigateList(ai *ai.RuntimeController) error {
    // 查找列表中的目标项
    items, _ := ai.Find(".ListItem")

    for _, item := range items {
        if item.State["title"] == "Target Item" {
            // 点击目标项
            return ai.Click(item.ID)
        }
    }

    return errors.New("target item not found")
}
```

### 示例 3：状态监控

```go
func ExampleWatchState(ai *ai.RuntimeController) {
    cancel := ai.Watch(func(s *state.Snapshot) {
        // 检查进度条
        if comp, ok := s.GetComponent("progress"); ok {
            progress := comp.State["value"].(float64)
            if progress >= 1.0 {
                fmt.Println("Download complete!")
            }
        }
    })

    defer cancel()
    // ... 继续其他操作
}
```

### 示例 4：错误处理

```go
func ExampleWithErrorHandling(ai *ai.RuntimeController) error {
    // 尝试点击按钮
    err := ai.Click("submit-btn")
    if err != nil {
        // 检查为什么失败
        snapshot, _ := ai.Inspect()

        if comp, ok := snapshot.Components["submit-btn"]; ok {
            if comp.Disabled {
                return fmt.Errorf("button is disabled: %+v", comp.State)
            }
        }

        return err
    }

    return nil
}
```

## JSON API

为了方便非 Go 语言的 AI Agent 使用，Runtime 提供 JSON API：

```json
// GET /ai/inspect - 获取当前状态
{
  "focus_path": ["root", "main", "form", "username"],
  "components": {
    "username": {
      "type": "TextInput",
      "state": {"value": "user@example.com"},
      "props": {"placeholder": "Email"},
      "rect": {"x": 10, "y": 5, "width": 40, "height": 1},
      "visible": true,
      "disabled": false
    }
  }
}

// POST /ai/dispatch - 分发 Action
{
  "type": "input_text",
  "target": "username",
  "payload": "new@email.com"
}

// POST /ai/find - 查找组件
{
  "selector": ".TextInput"
}

// POST /ai/query - 查询状态
{
  "component_id": "username",
  "state_key": "value"
}
```

## 测试 AI 集成

```go
// 位于: tui/runtime/ai/controller_test.go

package ai

func TestAIController(t *testing.T) {
    rt := NewTestRuntime()
    ctrl := NewRuntimeController(rt)

    // 添加测试组件
    input := component.NewTextInput()
    input.SetID("username")
    rt.Mount(input)

    // 测试 Inspect
    snapshot, err := ctrl.Inspect()
    assert.NoError(t, err)
    assert.NotNil(t, snapshot)

    // 测试 Find
    results, err := ctrl.Find("#username")
    assert.NoError(t, err)
    assert.Equal(t, 1, len(results))
    assert.Equal(t, "username", results[0].ID)

    // 测试 Input
    err = ctrl.Input("username", "test")
    assert.NoError(t, err)

    // 验证输入
    snapshot, _ = ctrl.Inspect()
    comp := snapshot.Components["username"]
    assert.Equal(t, "test", comp.State["value"])
}

func TestAIControllerWait(t *testing.T) {
    rt := NewTestRuntime()
    ctrl := NewRuntimeController(rt)

    // 异步修改状态
    go func() {
        time.Sleep(100 * time.Millisecond)
        ctrl.Input("test", "value")
    }()

    // 等待状态变化
    err := ctrl.WaitUntil(func(s *state.Snapshot) bool {
        if comp, ok := s.GetComponent("test"); ok {
            return comp.State["value"] == "value"
        }
        return false
    }, 1*time.Second)

    assert.NoError(t, err)
}
```

## 总结

AI 集成标准提供：

1. **一级操作者**: AI 与人类用户平级，不需要特殊通道
2. **语义化接口**: 使用 Action 而不是模拟输入
3. **结构化状态**: 直接访问数据，不需要 OCR
4. **完整信息**: 包括 disabled、hidden 等内部状态
5. **可组合**: 操作可以序列化和组合
6. **可测试**: 完整的状态快照支持测试

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [FOCUS_SYSTEM.md](./FOCUS_SYSTEM.md) - 焦点系统
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 架构不变量
