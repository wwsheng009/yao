# Event System Design

## 概述

事件系统是 TUI 框架的核心，负责处理所有用户输入和系统事件。本文档详细描述了事件系统的架构和实现。

## 事件流架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Event Flow                                     │
└─────────────────────────────────────────────────────────────────────────┘

User Input
    │
    ▼
┌─────────────────┐
│  Platform Layer │ 原始输入 (stdin, 信号)
│  (OS/Driver)    │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│  Event Pump     │ 持续读取输入，生成原始事件
│  (Goroutine)    │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Event Classifier │ 解析 ANSI 序列，分类事件
└─────────────────┘
    │
    ▼
┌─────────────────┐
│  Event Router   │ 路由到目标组件
└─────────────────┘
    │
    ├──► Capture Phase ──► 全局处理器
    │
    ├──► Target Phase ────► 目标组件
    │
    └──► Bubble Phase ────► 冒泡到父组件
```

## 核心类型定义

### 1. 事件接口

```go
// 位于: tui/framework/event/event.go

package event

// Event 事件接口
type Event interface {
    // Type 返回事件类型
    Type() EventType

    // Timestamp 返回事件时间戳
    Timestamp() time.Time

    // Source 返回事件源组件
    Source() Component

    // Target 返回目标组件
    Target() Component

    // PreventDefault 阻止默认行为
    PreventDefault()

    // IsDefaultPrevented 是否已阻止默认行为
    IsDefaultPrevented() bool

    // StopPropagation 停止事件传播
    StopPropagation()

    // IsPropagationStopped 是否已停止传播
    IsPropagationStopped() bool
}

// BaseEvent 基础事件实现
type BaseEvent struct {
    eventType    EventType
    timestamp    time.Time
    source       Component
    target       Component
    prevented    bool
    stopped      bool
}

func (e *BaseEvent) Type() EventType { return e.eventType }
func (e *BaseEvent) Timestamp() time.Time { return e.timestamp }
func (e *BaseEvent) Source() Component { return e.source }
func (e *BaseEvent) Target() Component { return e.target }
func (e *BaseEvent) PreventDefault() { e.prevented = true }
func (e *BaseEvent) IsDefaultPrevented() bool { return e.prevented }
func (e *BaseEvent) StopPropagation() { e.stopped = true }
func (e *BaseEvent) IsPropagationStopped() bool { return e.stopped }
```

### 2. 事件类型

```go
// 位于: tui/framework/event/types.go

package event

// EventType 事件类型
type EventType int

const (
    // 键盘事件
    EventKeyPress EventType = iota + 1000
    EventKeyRelease
    EventKeyRepeat

    // 鼠标事件
    EventMousePress
    EventMouseRelease
    EventMouseMove
    EventMouseWheel
    EventMouseEnter
    EventMouseLeave

    // 窗口事件
    EventResize
    EventFocus
    EventBlur
    EventClose

    // 组件事件
    EventClick
    EventDoubleClick
    EventContextMenu
    EventChange
    EventSubmit
    EventCancel
    EventSelect
    EventExpand
    EventCollapse

    // 拖放事件
    EventDragStart
    EventDrag
    EventDragEnd
    EventDrop

    // 焦点事件
    EventFocusIn
    EventFocusOut
    EventFocusGained
    EventFocusLost

    // 自定义事件
    EventCustom = 10000
)

// String 返回事件类型名称
func (t EventType) String() string {
    switch t {
    case EventKeyPress: return "KeyPress"
    case EventKeyRelease: return "KeyRelease"
    case EventMousePress: return "MousePress"
    case EventMouseRelease: return "MouseRelease"
    // ...
    default:
        if t >= EventCustom {
            return fmt.Sprintf("Custom(%d)", t)
        }
        return fmt.Sprintf("Unknown(%d)", t)
    }
}
```

### 3. 键盘事件

```go
// 位于: tui/framework/event/keyboard.go

package event

// KeyEvent 键盘事件
type KeyEvent struct {
    BaseEvent

    // 按键
    Key      rune   // 字符键 (如 'a', 'A', '1')
    Special  SpecialKey  // 特殊键 (如 Enter, Escape)

    // 修饰键
    Modifiers KeyModifier

    // 重复
    Repeat    bool
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
    KeyF4
    KeyF5
    KeyF6
    KeyF7
    KeyF8
    KeyF9
    KeyF10
    KeyF11
    KeyF12

    // 组合键
    KeySpace
    KeyCtrlA
    KeyCtrlB
    KeyCtrlC
    KeyCtrlD
    KeyCtrlE
    KeyCtrlF
    KeyCtrlG
    KeyCtrlH
    KeyCtrlI
    KeyCtrlJ
    KeyCtrlK
    KeyCtrlL
    KeyCtrlM
    KeyCtrlN
    KeyCtrlO
    KeyCtrlP
    KeyCtrlQ
    KeyCtrlR
    KeyCtrlS
    KeyCtrlT
    KeyCtrlU
    KeyCtrlV
    KeyCtrlW
    KeyCtrlX
    KeyCtrlY
    KeyCtrlZ
)

// KeyModifier 修饰键
type KeyModifier uint8

const (
    ModShift KeyModifier = 1 << iota
    ModAlt
    ModCtrl
    ModMeta  // Windows/Cmd 键
)

// Has 检查是否有修饰键
func (m KeyModifier) Has(mod KeyModifier) bool {
    return m&mod != 0
}

// String 返回修饰键字符串
func (m KeyModifier) String() string {
    var parts []string
    if m.Has(ModCtrl) {
        parts = append(parts, "Ctrl")
    }
    if m.Has(ModAlt) {
        parts = append(parts, "Alt")
    }
    if m.Has(ModShift) {
        parts = append(parts, "Shift")
    }
    if m.Has(ModMeta) {
        parts = append(parts, "Meta")
    }
    return strings.Join(parts, "+")
}

// KeyCombo 快捷键组合 (如 "Ctrl+C", "Alt+Shift+Delete")
type KeyCombo struct {
    Key       rune
    Special   SpecialKey
    Modifiers KeyModifier
}

// ParseKeyCombo 解析快捷键字符串
func ParseKeyCombo(s string) (*KeyCombo, error) {
    // 解析 "Ctrl+C", "Alt+Enter", "Shift+Tab" 等
    combo := &KeyCombo{}

    parts := strings.Split(s, "+")
    for _, part := range parts {
        switch strings.ToUpper(part) {
        case "CTRL":
            combo.Modifiers |= ModCtrl
        case "ALT":
            combo.Modifiers |= ModAlt
        case "SHIFT":
            combo.Modifiers |= ModShift
        case "META":
            combo.Modifiers |= ModMeta
        default:
            // 解析具体按键
            if len(part) == 1 {
                combo.Key = rune(part[0])
            } else {
                combo.Special = parseSpecialKey(part)
            }
        }
    }

    return combo, nil
}
```

### 4. 鼠标事件

```go
// 位于: tui/framework/event/mouse.go

package event

// MouseEvent 鼠标事件
type MouseEvent struct {
    BaseEvent

    // 位置
    X        int
    Y        int

    // 按钮
    Button   MouseButton

    // 动作
    Action   MouseAction

    // 修饰键
    Modifiers KeyModifier

    // 滚动
    DeltaX   int  // 水平滚动
    DeltaY   int  // 垂直滚动
}

// MouseButton 鼠标按钮
type MouseButton int

const (
    MouseNone MouseButton = iota
    MouseLeft
    MouseMiddle
    MouseRight
    MouseWheelUp
    MouseWheelDown
    MouseWheelLeft
    MouseWheelRight
)

// MouseAction 鼠标动作
type MouseAction int

const (
    MousePress MouseAction = iota
    MouseRelease
    MouseMove
    MouseDrag
    MouseWheel
)

// HitTest 命中测试结果
type HitTest struct {
    Component Component
    X        int  // 相对组件的 X 坐标
    Y        int  // 相对组件的 Y 坐标
    ZIndex   int
}

// HitTestInfo 命中测试信息
type HitTestInfo struct {
    hits     []HitTest
    topMost  *HitTest
}

// Position 位置信息
type Position struct {
    X int
    Y int
}

// Delta 位置变化
func (p Position) Delta(other Position) (dx, dy int) {
    return other.X - p.X, other.Y - p.Y
}

// Distance 距离
func (p Position) Distance(other Position) float64 {
    dx := float64(other.X - p.X)
    dy := float64(other.Y - p.Y)
    return math.Sqrt(dx*dx + dy*dy)
}
```

### 5. 窗口事件

```go
// 位于: tui/framework/event/window.go

package event

// ResizeEvent 窗口大小改变事件
type ResizeEvent struct {
    BaseEvent
    OldWidth  int
    OldHeight int
    NewWidth  int
    NewHeight int
}

// FocusEvent 焦点事件
type FocusEvent struct {
    BaseEvent
    Gained bool  // true = 获得焦点, false = 失去焦点
    Reason FocusReason
}

// FocusReason 焦点变化原因
type FocusReason int

const (
    FocusReasonTab FocusReason = iota  // Tab 导航
    FocusReasonClick                   // 鼠标点击
    FocusReasonProgrammatic            // 程序设置
    FocusReasonModal                   // 模态框
)

// CloseEvent 关闭事件
type CloseEvent struct {
    BaseEvent
    Reason CloseReason
}

// CloseReason 关闭原因
type CloseReason int

const (
    CloseReasonUser CloseReason = iota  // 用户请求 (Ctrl+C, Ctrl+D)
    CloseReasonSignal                   // 系统信号
    CloseReasonError                    // 错误导致
)
```

### 6. 组件事件

```go
// 位于: tui/framework/event/component.go

package event

// ClickEvent 点击事件
type ClickEvent struct {
    BaseEvent
    X        int
    Y        int
    Button   MouseButton
    Count    int  // 点击次数 (1=单击, 2=双击, 3=三击)
}

// ChangeEvent 值改变事件
type ChangeEvent struct {
    BaseEvent
    OldValue interface{}
    NewValue interface{}
}

// SubmitEvent 提交事件
type SubmitEvent struct {
    BaseEvent
    Data map[string]interface{}
}

// SelectEvent 选择事件
type SelectEvent struct {
    BaseEvent
    Index     int
    Selected  bool
    Value     interface{}
}

// ExpandEvent 展开/折叠事件
type ExpandEvent struct {
    BaseEvent
    Expanded  bool
    Index     int
}
```

## 事件处理

### 事件处理器接口

```go
// 位于: tui/framework/event/handler.go

package event

// EventHandler 事件处理器接口
type EventHandler interface {
    HandleEvent(Event) bool
}

// EventHandlerFunc 事件处理器函数
type EventHandlerFunc func(Event) bool

func (f EventHandlerFunc) HandleEvent(ev Event) bool {
    return f(ev)
}

// KeyHandler 键盘事件处理器
type KeyHandler interface {
    HandleKey(*KeyEvent) bool
}

// MouseHandler 鼠标事件处理器
type MouseHandler interface {
    HandleMouse(*MouseEvent) bool
}

// FocusHandler 焦点事件处理器
type FocusHandler interface {
    HandleFocus(*FocusEvent) bool
}
```

### 事件路由器

```go
// 位于: tui/framework/event/router.go

package event

// Router 事件路由器
type Router struct {
    // 全局处理器
    globalHandlers map[EventType][]EventHandler

    // 捕获阶段处理器
    captureHandlers []EventHandler

    // 目标查找
    hitTest func(x, y int) []Component

    // 队列
    eventQueue chan Event
    quit       chan struct{}
}

// NewRouter 创建事件路由器
func NewRouter() *Router {
    return &Router{
        globalHandlers: make(map[EventType][]EventHandler),
        eventQueue:     make(chan Event, 100),
        quit:          make(chan struct{}),
    }
}

// Subscribe 订阅事件
func (r *Router) Subscribe(eventType EventType, handler EventHandler) func() {
    r.globalHandlers[eventType] = append(r.globalHandlers[eventType], handler)

    // 返回取消订阅函数
    return func() {
        r.Unsubscribe(eventType, handler)
    }
}

// Unsubscribe 取消订阅
func (r *Router) Unsubscribe(eventType EventType, handler EventHandler) {
    handlers := r.globalHandlers[eventType]
    for i, h := range handlers {
        if h == handler {
            r.globalHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
            break
        }
    }
}

// AddCaptureHandler 添加捕获阶段处理器
func (r *Router) AddCaptureHandler(handler EventHandler) {
    r.captureHandlers = append(r.captureHandlers, handler)
}

// Route 路由事件到目标
func (r *Router) Route(ev Event) {
    // 1. 捕获阶段
    for _, handler := range r.captureHandlers {
        if handler.HandleEvent(ev) {
            if ev.IsPropagationStopped() {
                return
            }
        }
    }

    // 2. 全局处理器
    if handlers, ok := r.globalHandlers[ev.Type()]; ok {
        for _, handler := range handlers {
            if handler.HandleEvent(ev) {
                if ev.IsPropagationStopped() {
                    return
                }
            }
        }
    }

    // 3. 目标阶段
    if target := ev.Target(); target != nil {
        if handler, ok := target.(EventHandler); ok {
            handler.HandleEvent(ev)
        }
    }

    // 4. 冒泡阶段
    if !ev.IsPropagationStopped() {
        r.bubble(ev)
    }
}

// bubble 事件冒泡
func (r *Router) bubble(ev Event) {
    target := ev.Target()
    if target == nil {
        return
    }

    // 获取父组件链
    parents := r.getParentChain(target)

    // 向上冒泡
    for _, parent := range parents {
        if handler, ok := parent.(EventHandler); ok {
            if handler.HandleEvent(ev) {
                break
            }
        }
        if ev.IsPropagationStopped() {
            break
        }
    }
}

// SetHitTest 设置命中测试函数
func (r *Router) SetHitTest(fn func(x, y int) []Component) {
    r.hitTest = fn
}
```

## 事件泵

```go
// 位于: tui/framework/event/pump.go

package event

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
            data, err := p.reader.Read()
            if err != nil {
                // 处理错误
                continue
            }

            events := p.parser.Parse(data)
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

## 事件解析器

```go
// 位于: tui/framework/event/parser.go

package event

// Parser 事件解析器
type Parser struct {
    buffer []byte
    state  parseState
}

// NewParser 创建解析器
func NewParser() *Parser {
    return &Parser{
        buffer: make([]byte, 0, 256),
    }
}

// Parse 解析输入数据
func (p *Parser) Parse(data []byte) []Event {
    p.buffer = append(p.buffer, data...)

    var events []Event

    for len(p.buffer) > 0 {
        // 检查是否是 ANSI 转义序列
        if p.buffer[0] == '\x1b' {
            ev, consumed := p.parseANSI()
            if ev != nil {
                events = append(events, ev)
            }
            p.buffer = p.buffer[consumed:]
        } else {
            // 普通字符
            events = append(events, &KeyEvent{
                Key:   rune(p.buffer[0]),
            })
            p.buffer = p.buffer[1:]
        }
    }

    return events
}

// parseANSI 解析 ANSI 转义序列
func (p *Parser) parseANSI() (Event, int) {
    if len(p.buffer) < 2 {
        return nil, 0
    }

    // CSI 序列: ESC [
    if p.buffer[1] == '[' {
        return p.parseCSI()
    }

    // 其他序列
    return nil, 0
}

// parseCSI 解析 CSI 序列
func (p *Parser) parseCSI() (Event, int) {
    // CSI 格式: ESC [ <params> <intermediate> <final>
    // 例如: ESC [ A (上键), ESC [ 1 ; 5 B (Ctrl+Shift+下键)

    // 查找结束字符
    end := 2
    for end < len(p.buffer) {
        c := p.buffer[end]
        if c >= 0x40 && c <= 0x7E {
            break
        }
        end++
    }

    if end >= len(p.buffer) {
        return nil, 0  // 序列不完整
    }

    final := p.buffer[end]
    params := string(p.buffer[2:end])

    // 解析参数
    numbers := p.parseNumbers(params)

    // 根据最终字符确定事件类型
    switch final {
    case 'A':  // 上键
        return &KeyEvent{Special: KeyUp}, end + 1
    case 'B':  // 下键
        return &KeyEvent{Special: KeyDown}, end + 1
    case 'C':  // 右键
        return &KeyEvent{Special: KeyRight}, end + 1
    case 'D':  // 左键
        return &KeyEvent{Special: KeyLeft}, end + 1
    case 'Z':  // Shift+Tab
        return &KeyEvent{Special: KeyTab, Modifiers: ModShift}, end + 1
    // ... 更多 CSI 序列
    }

    return nil, end + 1
}

// parseNumbers 解析数字参数
func (p *Parser) parseNumbers(s string) []int {
    if s == "" {
        return []int{}
    }

    parts := strings.Split(s, ";")
    numbers := make([]int, len(parts))
    for i, part := range parts {
        if part == "" {
            numbers[i] = 0
        } else {
            n, _ := strconv.Atoi(part)
            numbers[i] = n
        }
    }
    return numbers
}
```

## 常用 ANSI 转义序列

```go
// 位于: tui/framework/event/ansi_codes.go

package event

// ANSICodes ANSI 转义码映射
var ANSICodes = map[string]Event{
    // 光标键
    "\x1b[A": &KeyEvent{Special: KeyUp},
    "\x1b[B": &KeyEvent{Special: KeyDown},
    "\x1b[C": &KeyEvent{Special: KeyRight},
    "\x1b[D": &KeyEvent{Special: KeyLeft},

    // 功能键
    "\x1bOP": &KeyEvent{Special: KeyF1},
    "\x1bOQ": &KeyEvent{Special: KeyF2},
    "\x1bOR": &KeyEvent{Special: KeyF3},
    "\x1bOS": &KeyEvent{Special: KeyF4},

    // 编辑键
    "\x1b[1~": &KeyEvent{Special: KeyHome},
    "\x1b[2~": &KeyEvent{Special: KeyInsert},
    "\x1b[3~": &KeyEvent{Special: KeyDelete},
    "\x1b[4~": &KeyEvent{Special: KeyEnd},
    "\x1b[5~": &KeyEvent{Special: KeyPageUp},
    "\x1b[6~": &KeyEvent{Special: KeyPageDown},

    // 控制键
    "\x1b[H": &KeyEvent{Special: KeyHome},
    "\x1b[F": &KeyEvent{Special: KeyEnd},

    // Tab
    "\x09":   &KeyEvent{Special: KeyTab},
    "\x1b[Z": &KeyEvent{Special: KeyTab, Modifiers: ModShift},
}

// ModifiedKeyCodes 修饰键组合
// 格式: ESC [ <modifier> + <key>
// modifier: 1=Shift, 2=Alt, 4=Ctrl, 8=Meta
var ModifiedKeyCodes = map[string]KeyModifier{
    "\x1b[1;2": ModShift,    // Shift+Key
    "\x1b[1;3": ModAlt,      // Alt+Key
    "\x1b[1;4": ModAlt | ModShift,    // Alt+Shift+Key
    "\x1b[1;5": ModCtrl,     // Ctrl+Key
    "\x1b[1;6": ModCtrl | ModShift,   // Ctrl+Shift+Key
    "\x1b[1;7": ModCtrl | ModAlt,     // Ctrl+Alt+Key
    "\x1b[1;8": ModCtrl | ModAlt | ModShift,  // Ctrl+Alt+Shift+Key
}
```

## 鼠标事件支持

```go
// 位于: tui/framework/event/mouse_support.go

package event

// EnableMouse 启用鼠标支持
func EnableMouse(terminal platform.Terminal) error {
    // 启用鼠标跟踪模式
    // SGR 模式 (最推荐): ESC [ ? 1006 h
    _, err := terminal.WriteString("\x1b[?1006h")
    if err != nil {
        return err
    }

    // 启用鼠标拖拽
    _, err = terminal.WriteString("\x1b[?1002h")
    return err
}

// DisableMouse 禁用鼠标支持
func DisableMouse(terminal platform.Terminal) error {
    _, err := terminal.WriteString("\x1b[?1006l")
    if err != nil {
        return err
    }
    _, err = terminal.WriteString("\x1b[?1002l")
    return err
}

// ParseMouseEvent 解析鼠标事件 (SGR 模式)
// 格式: ESC [ <button> ; <x> ; <y> M
func ParseMouseEvent(data []byte) (*MouseEvent, int) {
    // ESC [ < 0 ; 1 ; 2 M
    //      ^   ^  ^
    //      |   |  └─ Y (从1开始)
    //      |   └──── X (从1开始)
    //      └──────── Button (按下=正数, 释放=负数或+32)

    if len(data) < 6 || data[0] != '\x1b' || data[1] != '[' {
        return nil, 0
    }

    // 查找结束字符 'M' 或 'm'
    end := 2
    for end < len(data) && data[end] != 'M' && data[end] != 'm' {
        end++
    }
    if end >= len(data) {
        return nil, 0
    }

    // 解析参数
    params := strings.Split(string(data[2:end]), ";")
    if len(params) != 3 {
        return nil, 0
    }

    button, _ := strconv.Atoi(params[0])
    x, _ := strconv.Atoi(params[1])
    y, _ := strconv.Atoi(params[2])

    // 判断动作
    action := MousePress
    if data[end] == 'm' || button >= 64 {
        action = MouseRelease
        if button >= 64 {
            button -= 64
        } else {
            button = 0
        }
    }

    // 判断按钮
    mouseBtn := MouseNone
    switch button {
    case 0:
        mouseBtn = MouseLeft
    case 1:
        mouseBtn = MouseMiddle
    case 2:
        mouseBtn = MouseRight
    case 32:
        mouseBtn = MouseLeft
        action = MouseMove  // 拖拽
    case 33:
        mouseBtn = MouseMiddle
        action = MouseMove
    case 34:
        mouseBtn = MouseRight
        action = MouseMove
    case 64, 65:
        mouseBtn = MouseWheelUp
        action = MouseWheel
    case 66, 67:
        mouseBtn = MouseWheelDown
        action = MouseWheel
    }

    return &MouseEvent{
        X:      x - 1,  // 转换为从0开始
        Y:      y - 1,
        Button: mouseBtn,
        Action: action,
    }, end + 1
}
```

## 快捷键绑定

```go
// 位于: tui/framework/event/keymap.go

package event

// KeyMap 快捷键映射
type KeyMap struct {
    bindings map[string]EventHandler
}

// NewKeyMap 创建快捷键映射
func NewKeyMap() *KeyMap {
    return &KeyMap{
        bindings: make(map[string]EventHandler),
    }
}

// Bind 绑定快捷键
func (k *KeyMap) Bind(combo string, handler EventHandler) error {
    parsed, err := ParseKeyCombo(combo)
    if err != nil {
        return err
    }

    key := k.makeKey(parsed.Special, parsed.Key, parsed.Modifiers)
    k.bindings[key] = handler
    return nil
}

// BindFunc 绑定快捷键到函数
func (k *KeyMap) BindFunc(combo string, handler func(*KeyEvent)) error {
    return k.Bind(combo, EventHandlerFunc(func(ev Event) bool {
        if keyEv, ok := ev.(*KeyEvent); ok {
            handler(keyEv)
            return true
        }
        return false
    }))
}

// Lookup 查找快捷键处理器
func (k *KeyMap) Lookup(ev *KeyEvent) (EventHandler, bool) {
    key := k.makeKey(ev.Special, ev.Key, ev.Modifiers)
    handler, ok := k.bindings[key]
    return handler, ok
}

// makeKey 生成映射键
func (k *KeyMap) makeKey(special SpecialKey, key rune, modifiers KeyModifier) string {
    return fmt.Sprintf("%d:%d:%d", special, key, modifiers)
}

// 常用快捷键
const (
    KeyQuit    = "Ctrl+C"
    KeyCancel  = "Escape"
    KeyConfirm = "Enter"
    KeySubmit  = "Ctrl+Enter"
    KeyHelp    = "F1"
    KeySave    = "Ctrl+S"
    KeyOpen    = "Ctrl+O"
    KeyNew     = "Ctrl+N"
    KeyClose   = "Ctrl+W"
    KeyCopy    = "Ctrl+C"  // 有选择时
    KeyPaste   = "Ctrl+V"
    KeyCut     = "Ctrl+X"
    KeySelectAll = "Ctrl+A"
    KeyFind    = "Ctrl+F"
    KeyUndo    = "Ctrl+Z"
    KeyRedo    = "Ctrl+Y"
)
```

## 事件队列

```go
// 位于: tui/framework/event/queue.go

package event

// Queue 事件队列
type Queue struct {
    events   []Event
    capacity int
}

// NewQueue 创建事件队列
func NewQueue(capacity int) *Queue {
    return &Queue{
        events:   make([]Event, 0, capacity),
        capacity: capacity,
    }
}

// Push 添加事件到队列
func (q *Queue) Push(ev Event) bool {
    if len(q.events) >= q.capacity {
        return false
    }
    q.events = append(q.events, ev)
    return true
}

// Pop 从队列弹出事件
func (q *Queue) Pop() (Event, bool) {
    if len(q.events) == 0 {
        return nil, false
    }
    ev := q.events[0]
    q.events = q.events[1:]
    return ev, true
}

// Peek 查看队首事件
func (q *Queue) Peek() (Event, bool) {
    if len(q.events) == 0 {
        return nil, false
    }
    return q.events[0], true
}

// Clear 清空队列
func (q *Queue) Clear() {
    q.events = q.events[:0]
}

// Len 返回队列长度
func (q *Queue) Len() int {
    return len(q.events)
}

// Filter 过滤队列
func (q *Queue) Filter(pred func(Event) bool) []Event {
    var result []Event
    for _, ev := range q.events {
        if pred(ev) {
            result = append(result, ev)
        }
    }
    return result
}
```

## 使用示例

```go
// 创建事件路由器
router := event.NewRouter()

// 订阅全局事件
router.Subscribe(event.EventKeyPress, event.EventHandlerFunc(func(ev event.Event) bool {
    if keyEv, ok := ev.(*event.KeyEvent); ok {
        if keyEv.Key == 'q' && keyEv.Modifiers.Has(event.ModCtrl) {
            // 退出应用
            app.Quit()
            return true
        }
    }
    return false
}))

// 订阅窗口大小变化
router.Subscribe(event.EventResize, event.EventHandlerFunc(func(ev event.Event) bool {
    resizeEv := ev.(*event.ResizeEvent)
    app.Resize(resizeEv.NewWidth, resizeEv.NewHeight)
    return true
}))

// 设置组件事件处理器
button.SetEventHandler(event.EventHandlerFunc(func(ev event.Event) bool {
    if ev.Type() == event.EventClick {
        button.OnClick()
        return true
    }
    return false
}))
```
