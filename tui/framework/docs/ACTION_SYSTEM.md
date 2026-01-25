# Action System Design (V3)

> **版本**: V3
> **不变量**: Input ≠ Action
> **核心原则**: 所有 UI 状态变化必须能追溯到一次 `Dispatch(Action)`

## 概述

Action 系统是 TUI 框架的"神经系统"，负责将原始输入转换为语义化的操作指令。这是框架最核心的不变量之一。

### 为什么需要 Action 系统？

**没有 Action 系统的问题**：
```go
// ❌ Component 直接处理按键
func (i *Input) HandleKey(ev KeyEvent) {
    if ev.Key == 'a' {
        i.value += "a"  // 无法 replay
    }
}
```

**有了 Action 系统**：
```go
// ✅ Component 处理语义
func (i *Input) HandleAction(a *Action) bool {
    switch a.Type {
    case ActionInputText:
        if text, ok := a.Payload.(string); ok {
            i.value += text  // 可 replay，可测试
            return true
        }
    }
    return false
}
```

## 设计目标

1. **语义化**: Component 处理操作意图，而非物理按键
2. **可回放**: 所有 Action 可以记录和重放
3. **AI 友好**: AI 可以通过抽象操作控制 UI
4. **国际化**: 支持不同键盘布局和输入法
5. **可测试**: Action 是纯数据，易于测试

## 架构分层

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Platform Layer                                   │
│  ┌─────────────┐                                                          │
│  │   stdin     │  原始输入: '\x1b[A', 'a', '\x1b[3~'                   │
│  └─────────────┘                                                          │
│                        ⬇ 只产生 RawInput                                  │
└─────────────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Runtime Layer                                     │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                 │
│  │ InputParser │ → │   KeyMap    │ → │ Dispatcher   │                 │
│  │             │    │ (上下文感知) │    │             │                 │
│  └─────────────┘    └─────────────┘    └─────────────┘                 │
│       │                   │                    │                          │
│       ▼                   ▼                    ▼                          │
│   RawInput          KeyEvent           Action (语义)                      │
│                                              │                           │
│                                        ┌─────┴─────┐                    │
│                                        ▼           ▼                    │
│                                   Component   State Snapshot             │
│                                   Handler        更新                     │
└─────────────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Framework Layer                                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                 │
│  │ Component   │ → │ State       │ → │ Dirty Mark   │                 │
│  │ Handlers    │    │ Updater     │    │             │                 │
│  └─────────────┘    └─────────────┘    └─────────────┘                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## 核心不变量

### Input ≠ Action

```go
// ✅ 正确：分层清晰
Platform (stdin)
    ↓ RawInput
Runtime (KeyMap)
    ↓ Action
Component (HandleAction)
    ↓ State Update

// ❌ 错误：Component 处理原始按键
Component (HandleKey)
```

### 所有状态变化通过 Action

```go
// ✅ 正确：状态变化通过 Action
runtime.Dispatch(Action{
    Type: ActionInputText,
    Payload: "hello",
})
// → State Update → Dirty → Render

// ❌ 错误：直接修改状态
input.value = "world"  // 绕过 Action，无法 replay
```

## 核心类型定义

### 1. Action 结构

```go
// 位于: tui/runtime/action/action.go

package action

// ActionType 动作类型
type ActionType string

const (
    // === 导航动作 ===
    ActionNavigateNext     ActionType = "navigate_next"
    ActionNavigatePrev     ActionType = "navigate_prev"
    ActionNavigateUp       ActionType = "navigate_up"
    ActionNavigateDown     ActionType = "navigate_down"
    ActionNavigateLeft     ActionType = "navigate_left"
    ActionNavigateRight    ActionType = "navigate_right"
    ActionNavigateFirst    ActionType = "navigate_first"
    ActionNavigateLast     ActionType = "navigate_last"
    ActionNavigatePageUp   ActionType = "navigate_page_up"
    ActionNavigatePageDown ActionType = "navigate_page_down"

    // === 编辑动作 ===
    ActionInputText    ActionType = "input_text"
    ActionDeleteChar   ActionType = "delete_char"
    ActionDeleteWord   ActionType = "delete_word"
    ActionDeleteLine   ActionType = "delete_line"
    ActionInsertText   ActionType = "insert_text"
    ActionReplaceText  ActionType = "replace_text"

    // === 选择动作 ===
    ActionSelectAll      ActionType = "select_all"
    ActionSelectWord     ActionType = "select_word"
    ActionSelectLine     ActionType = "select_line"
    ActionClearSelection ActionType = "clear_selection"

    // === 光标动作 ===
    ActionCursorMove     ActionType = "cursor_move"
    ActionCursorHome     ActionType = "cursor_home"
    ActionCursorEnd      ActionType = "cursor_end"
    ActionCursorWordNext ActionType = "cursor_word_next"
    ActionCursorWordPrev ActionType = "cursor_word_prev"

    // === 表单动作 ===
    ActionSubmit   ActionType = "submit"
    ActionCancel   ActionType = "cancel"
    ActionReset    ActionType = "reset"
    ActionValidate ActionType = "validate"

    // === 鼠标动作 ===
    ActionMouseClick   ActionType = "mouse_click"
    ActionMouseRelease ActionType = "mouse_release"
    ActionMouseDrag    ActionType = "mouse_drag"
    ActionMouseHover   ActionType = "mouse_hover"

    // === 窗口动作 ===
    ActionClose      ActionType = "close"
    ActionMaximize   ActionType = "maximize"
    ActionMinimize   ActionType = "minimize"
    ActionFullscreen ActionType = "fullscreen"

    // === 视图动作 ===
    ActionScrollUp    ActionType = "scroll_up"
    ActionScrollDown  ActionType = "scroll_down"
    ActionScrollLeft  ActionType = "scroll_left"
    ActionScrollRight ActionType = "scroll_right"
    ActionScrollTo    ActionType = "scroll_to"
    ActionZoomIn      ActionType = "zoom_in"
    ActionZoomOut     ActionType = "zoom_out"

    // === 系统动作 ===
    ActionQuit    ActionType = "quit"
    ActionHelp    ActionType = "help"
    ActionRefresh ActionType = "refresh"
    ActionSearch  ActionType = "search"
    ActionCopy    ActionType = "copy"
    ActionPaste   ActionType = "paste"
    ActionUndo    ActionType = "undo"
    ActionRedo    ActionType = "redo"

    // === AI 专用动作 ===
    ActionAIQuery ActionType = "ai.query"
    ActionAIDump  ActionType = "ai.dump"
    ActionAIReset ActionType = "ai.reset"
)

// Action 动作
type Action struct {
    // 类型
    Type ActionType

    // 负载（可选）
    Payload any

    // 目标组件 ID（可选，为空时使用当前焦点）
    Target string

    // 来源组件 ID
    Source string

    // 时间戳
    Timestamp time.Time

    // 元数据
    Metadata map[string]string
}

// NewAction 创建动作
func NewAction(actionType ActionType) *Action {
    return &Action{
        Type:      actionType,
        Timestamp: time.Now(),
        Metadata:  make(map[string]string),
    }
}

// WithPayload 设置负载
func (a *Action) WithPayload(payload any) *Action {
    a.Payload = payload
    return a
}

// WithTarget 设置目标
func (a *Action) WithTarget(target string) *Action {
    a.Target = target
    return a
}

// WithSource 设置来源
func (a *Action) WithSource(source string) *Action {
    a.Source = source
    return a
}

// String 返回动作的字符串表示
func (a *Action) String() string {
    if a.Payload != nil {
        return fmt.Sprintf("%s(%v)", a.Type, a.Payload)
    }
    if a.Target != "" {
        return fmt.Sprintf("%s→%s", a.Type, a.Target)
    }
    return string(a.Type)
}

// Clone 克隆动作（用于 replay）
func (a *Action) Clone() *Action {
    metadata := make(map[string]string)
    for k, v := range a.Metadata {
        metadata[k] = v
    }
    return &Action{
        Type:      a.Type,
        Payload:   a.Payload,
        Target:    a.Target,
        Source:    a.Source,
        Timestamp: a.Timestamp,
        Metadata:  metadata,
    }
}
```

### 2. RawInput 定义

```go
// 位于: tui/runtime/input/raw.go

package input

// RawInputType 原始输入类型
type RawInputType int

const (
    InputKeyPress RawInputType = iota
    InputKeyRelease
    InputMouse
    InputResize
    InputPaste
    InputSignal
)

// SpecialKey 特殊键
type SpecialKey int

const (
    KeyUnknown SpecialKey = iota
    KeyEscape
    KeyEnter
    KeyTab
    KeyBackspace
    KeyDelete
    KeyInsert
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown
    KeyF1
    KeyF2
    KeyF3
    KeyF4
    KeyF5
    KeyF6
    KeyF7
    KeyF8
    KeyF9
    KeyF10
    KeyF11
    KeyF12
)

// KeyModifier 键盘修饰键
type KeyModifier int

const (
    ModNone  KeyModifier = 0
    ModCtrl  KeyModifier = 1 << iota
    ModAlt
    ModShift
    ModMotion // 鼠标拖拽标志
)

// MouseButton 鼠标按钮
type MouseButton int

const (
    MouseLeft MouseButton = iota
    MouseMiddle
    MouseRight
    WheelUp
    WheelDown
)

// MouseAction 鼠标动作
type MouseAction int

const (
    MousePress MouseAction = iota
    MouseRelease
    MouseMove
    MouseWheel
)

// RawInput 原始输入
type RawInput struct {
    Type RawInputType

    // 按键数据
    Key      rune
    Special  SpecialKey
    Modifiers KeyModifier

    // 鼠标数据
    MouseX     int
    MouseY     int
    MouseButton MouseButton
    MouseAction MouseAction

    // 原始数据
    Data []byte

    // 时间戳
    Timestamp time.Time
}

// String 返回输入的字符串表示
func (r *RawInput) String() string {
    switch r.Type {
    case InputKeyPress:
        if r.Special != KeyUnknown {
            return fmt.Sprintf("KeyPress(%s)", r.Special)
        }
        return fmt.Sprintf("KeyPress(%c)", r.Key)
    case InputMouse:
        return fmt.Sprintf("Mouse(%d,%d)", r.MouseX, r.MouseY)
    default:
        return fmt.Sprintf("Input(%d)", r.Type)
    }
}
```

### 3. KeyMap 定义

```go
// 位于: tui/runtime/input/keymap.go

package input

// KeyMap 按键映射
type KeyMap struct {
    // 默认映射
    defaultMap map[string]ActionType

    // 用户自定义映射
    customMap map[string]ActionType

    // 上下文映射（如 modal、form 等不同上下文）
    contextMaps map[string]map[string]ActionType

    // 当前上下文栈
    contextStack []string
}

// NewKeyMap 创建默认按键映射
func NewKeyMap() *KeyMap {
    return &KeyMap{
        defaultMap: map[string]ActionType{
            // === 导航 ===
            "Tab":        action.ActionNavigateNext,
            "Shift+Tab":  action.ActionNavigatePrev,
            "Up":         action.ActionNavigateUp,
            "Down":       action.ActionNavigateDown,
            "Left":       action.ActionNavigateLeft,
            "Right":      action.ActionNavigateRight,
            "Home":       action.ActionNavigateFirst,
            "End":        action.ActionNavigateLast,
            "PageUp":     action.ActionNavigatePageUp,
            "PageDown":   action.ActionNavigatePageDown,

            // === 编辑 ===
            "Backspace":  action.ActionDeleteChar,
            "Delete":     action.ActionDeleteChar,
            "Ctrl+W":     action.ActionDeleteWord,
            "Ctrl+K":     action.ActionDeleteLine,

            // === 选择 ===
            "Ctrl+A":     action.ActionSelectAll,

            // === 光标 ===
            "Ctrl+E":     action.ActionCursorEnd,
            // Ctrl+A 已用于 SelectAll

            // === 表单 ===
            "Enter":      action.ActionSubmit,
            "Escape":     action.ActionCancel,

            // === 系统 ===
            "Ctrl+C":     action.ActionQuit,
            "Ctrl+Q":     action.ActionQuit,
            "F1":         action.ActionHelp,
            "Ctrl+R":     action.ActionRefresh,
            "Ctrl+F":     action.ActionSearch,

            // === 撤销/重做 ===
            "Ctrl+Z":     action.ActionUndo,
            "Ctrl+Y":     action.ActionRedo,
        },
        customMap:    make(map[string]ActionType),
        contextMaps: make(map[string]map[string]ActionType),
        contextStack: make([]string, 0),
    }
}

// Map 将原始输入转换为 Action
func (km *KeyMap) Map(raw RawInput) *action.Action {
    // 1. 处理鼠标事件
    if raw.Type == InputMouse {
        return km.mapMouse(raw)
    }

    // 2. 只处理按键事件
    if raw.Type != InputKeyPress {
        return nil
    }

    // 生成按键组合键
    key := km.makeKey(raw.Special, raw.Key, raw.Modifiers)

    // 3. 优先查找上下文映射（栈顶优先）
    for i := len(km.contextStack) - 1; i >= 0; i-- {
        ctx := km.contextStack[i]
        if ctxMap, ok := km.contextMaps[ctx]; ok {
            if actionType, found := ctxMap[key]; found {
                return action.NewAction(actionType).
                    WithPayload(km.buildPayload(raw)).
                    WithMetadata("context", ctx)
            }
        }
    }

    // 4. 查找自定义全局映射
    if actionType, found := km.customMap[key]; found {
        return action.NewAction(actionType).
            WithPayload(km.buildPayload(raw))
    }

    // 5. 查找默认全局映射
    if actionType, found := km.defaultMap[key]; found {
        return action.NewAction(actionType).
            WithPayload(km.buildPayload(raw))
    }

    // 6. 普通字符输入（仅当没有修饰键或仅有 Shift 时）
    if raw.Key > 0 && (raw.Modifiers == 0 || raw.Modifiers == ModShift) {
        return action.NewAction(action.ActionInputText).
            WithPayload(string(raw.Key))
    }

    return nil
}

// mapMouse 处理鼠标映射
func (km *KeyMap) mapMouse(raw RawInput) *action.Action {
    payload := map[string]any{
        "x":      raw.MouseX,
        "y":      raw.MouseY,
        "button": int(raw.MouseButton),
    }

    switch raw.MouseAction {
    case MousePress:
        return action.NewAction(action.ActionMouseClick).WithPayload(payload)
    case MouseRelease:
        return action.NewAction(action.ActionMouseRelease).WithPayload(payload)
    case MouseMove:
        if raw.Modifiers&ModMotion != 0 {
            return action.NewAction(action.ActionMouseDrag).WithPayload(payload)
        }
        return nil // 普通移动不生成 Action
    case MouseWheel:
        if raw.MouseButton == WheelUp {
            return action.NewAction(action.ActionScrollUp).WithPayload(payload)
        }
        if raw.MouseButton == WheelDown {
            return action.NewAction(action.ActionScrollDown).WithPayload(payload)
        }
    }

    return nil
}

// makeKey 生成映射键
func (km *KeyMap) makeKey(special SpecialKey, key rune, modifiers KeyModifier) string {
    var parts []string

    if modifiers&ModCtrl != 0 {
        parts = append(parts, "Ctrl")
    }
    if modifiers&ModAlt != 0 {
        parts = append(parts, "Alt")
    }
    if modifiers&ModShift != 0 {
        parts = append(parts, "Shift")
    }

    if special != KeyUnknown {
        parts = append(parts, special.String())
    } else if key > 0 {
        parts = append(parts, string(key))
    }

    return strings.Join(parts, "+")
}

// buildPayload 构建负载
func (km *KeyMap) buildPayload(raw RawInput) any {
    switch raw.Special {
    case KeyUp, KeyDown, KeyLeft, KeyRight:
        return map[string]int{"step": 1}
    case KeyPageUp, KeyPageDown:
        return map[string]int{"step": 10}
    default:
        return nil
    }
}

// Bind 绑定自定义按键
func (km *KeyMap) Bind(combo string, actionType action.ActionType) {
    km.customMap[combo] = actionType
}

// Unbind 解除绑定
func (km *KeyMap) Unbind(combo string) {
    delete(km.customMap, combo)
}

// BindContext 绑定上下文相关按键
func (km *KeyMap) BindContext(context, combo string, actionType action.ActionType) {
    if km.contextMaps[context] == nil {
        km.contextMaps[context] = make(map[string]action.ActionType)
    }
    km.contextMaps[context][combo] = actionType
}

// PushContext 推入上下文
func (km *KeyMap) PushContext(context string) {
    km.contextStack = append(km.contextStack, context)
}

// PopContext 弹出上下文
func (km *KeyMap) PopContext() {
    if len(km.contextStack) > 0 {
        km.contextStack = km.contextStack[:len(km.contextStack)-1]
    }
}

// String 返回 SpecialKey 的字符串表示
func (k SpecialKey) String() string {
    switch k {
    case KeyEscape:
        return "Escape"
    case KeyEnter:
        return "Enter"
    case KeyTab:
        return "Tab"
    case KeyBackspace:
        return "Backspace"
    case KeyDelete:
        return "Delete"
    case KeyInsert:
        return "Insert"
    case KeyUp:
        return "Up"
    case KeyDown:
        return "Down"
    case KeyLeft:
        return "Left"
    case KeyRight:
        return "Right"
    case KeyHome:
        return "Home"
    case KeyEnd:
        return "End"
    case KeyPageUp:
        return "PageUp"
    case KeyPageDown:
        return "PageDown"
    case KeyF1, KeyF2, KeyF3, KeyF4, KeyF5, KeyF6, KeyF7, KeyF8, KeyF9, KeyF10, KeyF11, KeyF12:
        return fmt.Sprintf("F%d", int(k-KeyF1+1))
    default:
        return "Unknown"
    }
}

// Has 检查是否有修饰键
func (m KeyModifier) Has(mod KeyModifier) bool {
    return m&mod != 0
}
```

### 4. Action Dispatcher

```go
// 位于: tui/runtime/action/dispatcher.go

package action

// Dispatcher 动作分发器
type Dispatcher struct {
    // 目标注册（Component ID → Component）
    targets map[string]ActionTarget

    // 焦点管理器
    focus *focus.Manager

    // 全局处理器（ActionType → Handlers）
    globalHandlers map[ActionType][]ActionHandler

    // 日志记录器
    logger *Logger

    // 错误处理器
    errorHandler ErrorHandler

    // 统计
    stats DispatchStats
}

// ActionTarget 动作目标（即 Component）
type ActionTarget interface {
    ID() string
    HandleAction(a *Action) bool
}

// ActionHandler 动作处理器
type ActionHandler func(*Action) bool

// ErrorHandler 错误处理器
type ErrorHandler func(*Action, error)

// Logger 动作日志记录器
type Logger struct {
    enabled bool
    entries []LogEntry
    maxSize int
}

// LogEntry 日志条目
type LogEntry struct {
    Timestamp time.Time
    Action    *Action
    Target    string
    Handled   bool
    Error     error
}

// DispatchStats 分发统计
type DispatchStats struct {
    TotalDispatched int64
    TotalHandled    int64
    TotalFailed     int64
}

// NewDispatcher 创建分发器
func NewDispatcher() *Dispatcher {
    return &Dispatcher{
        targets:        make(map[string]ActionTarget),
        globalHandlers: make(map[ActionType][]ActionHandler),
        logger:         NewLogger(1000),
        errorHandler:   defaultErrorHandler,
    }
}

// NewLogger 创建日志记录器
func NewLogger(maxSize int) *Logger {
    return &Logger{
        enabled: true,
        entries: make([]LogEntry, 0, maxSize),
        maxSize: maxSize,
    }
}

// Dispatch 分发动作
func (d *Dispatcher) Dispatch(a *Action) bool {
    d.stats.TotalDispatched++

    // 1. 全局处理器（优先级最高）
    for _, handler := range d.globalHandlers[a.Type] {
        if handler(a) {
            d.log(a, "", true, nil)
            d.stats.TotalHandled++
            return true
        }
    }

    // 2. 确定目标
    targetID := a.Target
    if targetID == "" && d.focus != nil {
        targetID = d.focus.CurrentID()
    }

    if targetID == "" {
        d.log(a, "", false, fmt.Errorf("no target"))
        d.stats.TotalFailed++
        return false
    }

    // 3. 分发到目标
    target := d.targets[targetID]
    if target == nil {
        err := fmt.Errorf("target not found: %s", targetID)
        d.handleError(a, err)
        d.log(a, targetID, false, err)
        d.stats.TotalFailed++
        return false
    }

    handled := target.HandleAction(a)
    d.log(a, targetID, handled, nil)

    if handled {
        d.stats.TotalHandled++
    } else {
        d.stats.TotalFailed++
    }

    return handled
}

// DispatchTo 分发到指定目标
func (d *Dispatcher) DispatchTo(targetID string, a *Action) bool {
    a.Target = targetID
    return d.Dispatch(a)
}

// Register 注册目标
func (d *Dispatcher) Register(target ActionTarget) {
    d.targets[target.ID()] = target
}

// Unregister 注销目标
func (d *Dispatcher) Unregister(id string) {
    delete(d.targets, id)
}

// Subscribe 订阅动作类型
func (d *Dispatcher) Subscribe(actionType ActionType, handler ActionHandler) func() {
    d.globalHandlers[actionType] = append(d.globalHandlers[actionType], handler)

    return func() {
        d.Unsubscribe(actionType, handler)
    }
}

// Unsubscribe 取消订阅
func (d *Dispatcher) Unsubscribe(actionType ActionType, handler ActionHandler) {
    handlers := d.globalHandlers[actionType]
    for i, h := range handlers {
        // 使用函数指针比较
        if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
            d.globalHandlers[actionType] = append(handlers[:i], handlers[i+1:]...)
            break
        }
    }
}

// SetFocusManager 设置焦点管理器
func (d *Dispatcher) SetFocusManager(focus *focus.Manager) {
    d.focus = focus
}

// SetErrorHandler 设置错误处理器
func (d *Dispatcher) SetErrorHandler(handler ErrorHandler) {
    d.errorHandler = handler
}

// EnableLog 启用日志
func (d *Dispatcher) EnableLog(enabled bool) {
    d.logger.enabled = enabled
}

// GetLogs 获取日志
func (d *Dispatcher) GetLogs() []LogEntry {
    return d.logger.entries
}

// GetStats 获取统计
func (d *Dispatcher) GetStats() DispatchStats {
    return d.stats
}

// log 记录日志
func (d *Dispatcher) log(a *Action, target string, handled bool, err error) {
    if !d.logger.enabled {
        return
    }

    entry := LogEntry{
        Timestamp: time.Now(),
        Action:    a.Clone(),
        Target:    target,
        Handled:   handled,
        Error:     err,
    }

    d.logger.entries = append(d.logger.entries, entry)

    if len(d.logger.entries) >= d.logger.maxSize {
        // 滚动日志
        d.logger.entries = d.logger.entries[1:]
    }
}

// handleError 处理错误
func (d *Dispatcher) handleError(a *Action, err error) {
    if d.errorHandler != nil {
        d.errorHandler(a, err)
    }
}

// defaultErrorHandler 默认错误处理器
func defaultErrorHandler(a *Action, err error) {
    // 默认只记录，不中断程序
    log.Printf("[Action] Error dispatching %s: %v", a.Type, err)
}
```

## 与状态管理集成

### Action → State 流程

```go
// 位于: tui/runtime/state/tracker.go

package state

// Tracker 状态追踪器
type Tracker struct {
    // 当前状态快照
    current *StateSnapshot

    // 历史状态（用于 undo）
    history []*StateSnapshot

    // 最大历史记录数
    maxHistory int
}

// StateSnapshot 状态快照
type StateSnapshot struct {
    Timestamp time.Time
    FocusPath []string
    Components map[string]ComponentState
    Modals     []string
}

// ComponentState 组件状态
type ComponentState struct {
    ID         string
    Type       string
    Props      map[string]interface{} // 静态属性
    State      map[string]interface{} // 动态状态
    Rect       Rect                   // 布局位置
}

// BeforeAction 在执行 Action 前记录状态
func (t *Tracker) BeforeAction() *StateSnapshot {
    return t.capture()
}

// AfterAction 在执行 Action 后记录状态
func (t *Tracker) AfterAction(before *StateSnapshot) *StateSnapshot {
    after := t.capture()

    // 只在状态真正变化时才记录历史
    if !t.equal(before, after) {
        t.history = append(t.history, before)
        if len(t.history) > t.maxHistory {
            t.history = t.history[1:]
        }
    }

    t.current = after
    return after
}

// capture 捕获当前状态
func (t *Tracker) capture() *StateSnapshot {
    // 从 Runtime 收集所有组件状态
    return &StateSnapshot{
        Timestamp: time.Now(),
        FocusPath: t.getFocusPath(),
        Components: t.collectComponentStates(),
        Modals:     t.getActiveModals(),
    }
}
```

### 完整的 Action 执行流程

```go
// 位于: tui/runtime/runtime.go

func (r *Runtime) ProcessInput(raw RawInput) error {
    // 1. RawInput → Action
    act := r.keyMap.Map(raw)
    if act == nil {
        return nil
    }

    // 2. 记录执行前状态
    before := r.stateTracker.BeforeAction()

    // 3. 分发 Action
    handled := r.dispatcher.Dispatch(act)

    // 4. 记录执行后状态
    after := r.stateTracker.AfterAction(before)

    // 5. 如果处理成功，标记 Dirty
    if handled {
        r.dirty.MarkAll() // 简化处理，实际可以更精确
    }

    return nil
}
```

## Component Action 处理

### 示例：TextInput

```go
// 位于: tui/framework/component/textinput.go

package component

// HandleAction 处理动作
func (t *TextInput) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionInputText:
        if text, ok := a.Payload.(string); ok {
            t.insert(text)
            return true
        }

    case action.ActionDeleteChar:
        t.deleteChar()
        return true

    case action.ActionDeleteWord:
        t.deleteWord()
        return true

    case action.ActionCursorHome:
        t.setCursor(0)
        return true

    case action.ActionCursorEnd:
        t.setCursor(len(t.value))
        return true

    case action.ActionSelectAll:
        t.selection = Selection{Start: 0, End: len(t.value)}
        return true

    case action.ActionSubmit:
        if t.onSubmit != nil {
            t.onSubmit(t.value)
        }
        return true
    }

    return false
}

// insert 插入文本
func (t *TextInput) insert(text string) {
    if t.selection.Active {
        // 删除选中内容
        t.deleteSelection()
    }

    t.value = t.value[:t.cursor] + text + t.value[t.cursor:]
    t.cursor += len(text)
}

// deleteChar 删除字符
func (t *TextInput) deleteChar() {
    if t.selection.Active {
        t.deleteSelection()
        return
    }

    if t.cursor < len(t.value) {
        t.value = t.value[:t.cursor] + t.value[t.cursor+1:]
    }
}
```

### 示例：List

```go
// 位于: tui/framework/component/list.go

package component

// HandleAction 处理动作
func (l *List) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateDown:
        l.moveCursor(1)
        return true

    case action.ActionNavigateUp:
        l.moveCursor(-1)
        return true

    case action.ActionNavigatePageDown:
        l.moveCursor(l.visibleHeight)
        return true

    case action.ActionNavigatePageUp:
        l.moveCursor(-l.visibleHeight)
        return true

    case action.ActionNavigateFirst:
        l.setCursor(0)
        return true

    case action.ActionNavigateLast:
        l.setCursor(len(l.items) - 1)
        return true

    case action.ActionSubmit:
        if l.onSelect != nil && len(l.items) > 0 {
            l.onSelect(l.items[l.cursor])
        }
        return true
    }

    return false
}
```

### 示例：Form

```go
// 位于: tui/framework/component/form.go

package component

// HandleAction 处理动作
func (f *Form) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateNext:
        return f.navigateNext()

    case action.ActionNavigatePrev:
        return f.navigatePrev()

    case action.ActionSubmit:
        return f.submit()

    case action.ActionCancel:
        if f.onCancel != nil {
            f.onCancel()
        }
        return true

    case action.ActionValidate:
        return f.validate()

    case action.ActionReset:
        return f.reset()
    }

    // 尝试分发给当前焦点字段
    if f.currentField != nil {
        return f.currentField.HandleAction(a)
    }

    return false
}
```

## 错误处理

### Action 处理失败

当 Action 处理失败时，框架提供清晰的错误信息：

```go
// 错误示例
{
    "error": "ComponentNotInteractable",
    "message": "Target 'submit-btn' is currently disabled",
    "action": "submit",
    "target": "submit-btn",
    "component_state": {
        "disabled": true
    }
}
```

### 错误处理策略

```go
// 位于: tui/runtime/action/errors.go

package action

// Error Action 处理错误
type Error struct {
    Type    ErrorType
    Message string
    Action  *Action
    Target  string
    Details map[string]interface{}
}

// ErrorType 错误类型
type ErrorType string

const (
    ErrTargetNotFound      ErrorType = "target_not_found"
    ErrTargetDisabled      ErrorType = "target_disabled"
    ErrTargetNotInteractable ErrorType = "target_not_interactable"
    ErrInvalidPayload      ErrorType = "invalid_payload"
    ErrActionNotSupported  ErrorType = "action_not_supported"
)

// NewError 创建错误
func NewError(errorType ErrorType, message string, a *Action) *Error {
    return &Error{
        Type:    errorType,
        Message: message,
        Action:  a,
        Details: make(map[string]interface{}),
    }
}

// WithTarget 设置目标
func (e *Error) WithTarget(target string) *Error {
    e.Target = target
    return e
}

// WithDetail 添加详情
func (e *Error) WithDetail(key string, value interface{}) *Error {
    e.Details[key] = value
    return e
}

// Error 实现 error 接口
func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s (action: %s, target: %s)",
        e.Type, e.Message, e.Action.Type, e.Target)
}
```

## 快捷键配置

### 配置文件格式

```yaml
# keymap.yaml
navigation:
  next: ["Tab", "Down"]
  prev: ["Shift+Tab", "Up"]
  up: "Up"
  down: "Down"
  left: "Left"
  right: "Right"
  first: "Home"
  last: "End"
  page_up: "PageUp"
  page_down: "PageDown"

editing:
  delete_char: ["Backspace", "Delete"]
  delete_word: "Ctrl+W"
  delete_line: "Ctrl+K"
  select_all: "Ctrl+A"

form:
  submit: ["Enter", "Ctrl+Enter"]
  cancel: "Escape"
  validate: "Ctrl+V"
  reset: "Ctrl+R"

system:
  quit: ["Ctrl+C", "Ctrl+Q"]
  help: "F1"
  search: "Ctrl+F"

# 上下文特定映射
contexts:
  modal:
    submit: ["Enter", "Ctrl+Enter"]
    cancel: "Escape"

  text_input:
    left: "Ctrl+B"
    right: "Ctrl+F"
    home: "Ctrl+A"
    end: "Ctrl+E"
```

### 加载配置

```go
// 位于: tui/runtime/action/keymap_loader.go

package action

// KeyMapConfig 配置结构
type KeyMapConfig struct {
    Navigation map[string][]string `yaml:"navigation"`
    Editing    map[string][]string `yaml:"editing"`
    Form       map[string][]string `yaml:"form"`
    System     map[string][]string `yaml:"system"`
    Contexts   map[string]map[string][]string `yaml:"contexts"`
}

// LoadKeyMap 从配置文件加载
func LoadKeyMap(path string) (*input.KeyMap, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read keymap file: %w", err)
    }

    var config KeyMapConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse keymap file: %w", err)
    }

    km := input.NewKeyMap()

    // 应用导航映射
    for name, keys := range config.Navigation {
        var actionType ActionType
        switch name {
        case "next":
            actionType = ActionNavigateNext
        case "prev":
            actionType = ActionNavigatePrev
        case "up":
            actionType = ActionNavigateUp
        case "down":
            actionType = ActionNavigateDown
        case "left":
            actionType = ActionNavigateLeft
        case "right":
            actionType = ActionNavigateRight
        case "first":
            actionType = ActionNavigateFirst
        case "last":
            actionType = ActionNavigateLast
        case "page_up":
            actionType = ActionNavigatePageUp
        case "page_down":
            actionType = ActionNavigatePageDown
        }

        for _, key := range keys {
            km.Bind(key, actionType)
        }
    }

    // 应用编辑映射
    for name, keys := range config.Editing {
        var actionType ActionType
        switch name {
        case "delete_char":
            actionType = ActionDeleteChar
        case "delete_word":
            actionType = ActionDeleteWord
        case "delete_line":
            actionType = ActionDeleteLine
        case "select_all":
            actionType = ActionSelectAll
        }

        for _, key := range keys {
            km.Bind(key, actionType)
        }
    }

    // 应用表单映射
    for name, keys := range config.Form {
        var actionType ActionType
        switch name {
        case "submit":
            actionType = ActionSubmit
        case "cancel":
            actionType = ActionCancel
        case "validate":
            actionType = ActionValidate
        case "reset":
            actionType = ActionReset
        }

        for _, key := range keys {
            km.Bind(key, actionType)
        }
    }

    // 应用系统映射
    for name, keys := range config.System {
        var actionType ActionType
        switch name {
        case "quit":
            actionType = ActionQuit
        case "help":
            actionType = ActionHelp
        case "search":
            actionType = ActionSearch
        }

        for _, key := range keys {
            km.Bind(key, actionType)
        }
    }

    // 应用上下文映射
    for ctx, mappings := range config.Contexts {
        for name, keys := range mappings {
            var actionType ActionType
            switch name {
            case "submit":
                actionType = ActionSubmit
            case "cancel":
                actionType = ActionCancel
            }

            for _, key := range keys {
                km.BindContext(ctx, key, actionType)
            }
        }
    }

    return km, nil
}
```

## AI 集成

### AI Controller 接口

```go
// 位于: tui/runtime/ai/controller.go

package ai

// Controller AI 控制器接口
type Controller interface {
    // 感知：获取当前完整的 UI 状态快照
    Inspect() (*state.StateSnapshot, error)

    // 操作：向指定组件发送语义指令
    Dispatch(a *action.Action) error

    // 查询：查找组件（类似于 DOM 选择器）
    Find(selector string) ([]ComponentInfo, error)

    // 等待：等待特定状态出现
    WaitUntil(condition func(*state.StateSnapshot) bool, timeout time.Duration) error
}

// ComponentInfo 组件信息
type ComponentInfo struct {
    ID    string
    Type  string
    Props map[string]interface{}
    State map[string]interface{}
}

// RuntimeController Runtime 实现的 AI 控制器
type RuntimeController struct {
    runtime    *Runtime
    dispatcher *action.Dispatcher
    tracker    *state.Tracker
}

func (c *RuntimeController) Inspect() (*state.StateSnapshot, error) {
    return c.tracker.Current(), nil
}

func (c *RuntimeController) Dispatch(a *action.Action) error {
    handled := c.dispatcher.Dispatch(a)
    if !handled {
        return fmt.Errorf("action not handled: %s", a.Type)
    }
    return nil
}

func (c *RuntimeController) Find(selector string) ([]ComponentInfo, error) {
    // 简单的 ID 选择器
    if strings.HasPrefix(selector, "#") {
        id := selector[1:]
        comp := c.runtime.FindComponent(id)
        if comp == nil {
            return nil, fmt.Errorf("component not found: %s", id)
        }
        return []ComponentInfo{{
            ID:    comp.ID(),
            Type:  comp.Type(),
            Props: comp.Props(),
            State: comp.State(),
        }}, nil
    }

    // 类型选择器
    if strings.HasPrefix(selector, ".") {
        typ := selector[1:]
        return c.runtime.FindComponentsByType(typ)
    }

    return nil, fmt.Errorf("invalid selector: %s", selector)
}
```

## 测试

### 单元测试

```go
// 位于: tui/runtime/action/action_test.go

package action

func TestAction(t *testing.T) {
    a := NewAction(ActionInputText).
        WithPayload("hello").
        WithTarget("input-1").
        WithSource("keyboard")

    assert.Equal(t, ActionInputText, a.Type)
    assert.Equal(t, "hello", a.Payload)
    assert.Equal(t, "input-1", a.Target)
    assert.Equal(t, "keyboard", a.Source)
}

func TestActionClone(t *testing.T) {
    a := NewAction(ActionInputText).WithPayload("hello")
    clone := a.Clone()

    // 修改原始不应该影响克隆
    a.Payload = "world"

    assert.Equal(t, "world", a.Payload)
    assert.Equal(t, "hello", clone.Payload)
}
```

```go
// 位于: tui/runtime/input/keymap_test.go

package input

func TestKeyMap(t *testing.T) {
    km := NewKeyMap()

    tests := []struct {
        name     string
        input    RawInput
        expected ActionType
    }{
        {
            name:     "Tab to next",
            input:    RawInput{Special: KeyTab},
            expected: action.ActionNavigateNext,
        },
        {
            name:     "Shift+Tab to prev",
            input:    RawInput{Special: KeyTab, Modifiers: ModShift},
            expected: action.ActionNavigatePrev,
        },
        {
            name:     "Enter to submit",
            input:    RawInput{Special: KeyEnter},
            expected: action.ActionSubmit,
        },
        {
            name:     "Escape to cancel",
            input:    RawInput{Special: KeyEscape},
            expected: action.ActionCancel,
        },
        {
            name:     "character input",
            input:    RawInput{Key: 'a'},
            expected: action.ActionInputText,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := km.Map(tt.input)
            if result == nil {
                t.Fatalf("expected action %s, got nil", tt.expected)
            }
            if result.Type != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result.Type)
            }
        })
    }
}

func TestKeyMapContext(t *testing.T) {
    km := NewKeyMap()

    // 在 modal 上下文中，Enter 有不同的行为
    km.BindContext("modal", "Enter", action.ActionConfirm)

    km.PushContext("modal")
    result := km.Map(RawInput{Special: KeyEnter})
    km.PopContext()

    assert.Equal(t, action.ActionConfirm, result.Type)
}
```

```go
// 位于: tui/runtime/action/dispatcher_test.go

package action

func TestDispatcher(t *testing.T) {
    d := NewDispatcher()

    // 注册 mock 组件
    mockComp := &MockComponent{id: "test-1"}
    d.Register(mockComp)

    // 测试分发
    a := NewAction(ActionInputText).WithTarget("test-1").WithPayload("hello")
    handled := d.Dispatch(a)

    assert.True(t, handled)
    assert.Equal(t, "hello", mockComp.lastValue)
    assert.Equal(t, int64(1), d.GetStats().TotalHandled)
}

func TestDispatcherWithFocus(t *testing.T) {
    d := NewDispatcher()
    focus := focus.NewManager()
    d.SetFocusManager(focus)

    // 注册组件
    comp1 := &MockComponent{id: "comp-1"}
    comp2 := &MockComponent{id: "comp-2"}
    d.Register(comp1)
    d.Register(comp2)

    // 设置焦点
    focus.SetFocus("comp-2")

    // 分发（无 Target，应该使用焦点）
    a := NewAction(ActionInputText).WithPayload("test")
    handled := d.Dispatch(a)

    assert.True(t, handled)
    assert.Equal(t, "test", comp2.lastValue)
    assert.Equal(t, "", comp1.lastValue)
}

func TestDispatcherGlobalHandler(t *testing.T) {
    d := NewDispatcher()

    // 全局处理器
    called := false
    d.Subscribe(ActionQuit, func(a *Action) bool {
        called = true
        return true
    })

    // 分发
    a := NewAction(ActionQuit)
    d.Dispatch(a)

    assert.True(t, called)
}

// MockComponent 测试用 Mock 组件
type MockComponent struct {
    id        string
    lastValue string
}

func (m *MockComponent) ID() string {
    return m.id
}

func (m *MockComponent) HandleAction(a *Action) bool {
    if a.Type == ActionInputText {
        if text, ok := a.Payload.(string); ok {
            m.lastValue = text
            return true
        }
    }
    return false
}
```

### 集成测试

```go
func TestActionFlow(t *testing.T) {
    // 创建 Runtime
    rt := NewRuntime()

    // 添加组件
    input := component.NewTextInput()
    input.SetID("username")
    rt.Mount(input)

    // 模拟输入
    raw := input.RawInput{
        Type: InputKeyPress,
        Key:  'a',
    }

    err := rt.ProcessInput(raw)
    assert.NoError(t, err)

    // 验证状态
    assert.Equal(t, "a", input.GetValue())
}
```

## 总结

Action 系统是 TUI 框架的核心特性，它：

1. **分离关注点**: Platform 只管输入，Component 只管语义
2. **支持自动化**: 所有操作可以记录和回放
3. **AI 友好**: AI 通过抽象操作控制 UI
4. **国际化**: 支持不同键盘布局和语言
5. **可扩展**: 用户可以自定义按键映射
6. **可测试**: Action 是纯数据，易于测试
7. **可追踪**: 所有状态变化都能追溯到 Action

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [EVENT_SYSTEM.md](./EVENT_SYSTEM.md) - 事件系统
- [FOCUS_SYSTEM.md](./FOCUS_SYSTEM.md) - 焦点系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - AI 集成
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 架构不变量
