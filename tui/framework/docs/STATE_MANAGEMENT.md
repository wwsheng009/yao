# State Management Design (V3)

> **版本**: V3
> **核心原则**: 状态可枚举、可快照、可追溯
> **关键特性**: 时间旅行、Replay 支持、AI 友好

## 概述

状态管理是 TUI 框架的核心功能之一。V3 架构强调所有状态必须是：
1. **可枚举** - 可以完整列出所有状态
2. **可快照** - 可以保存和恢复状态
3. **可追溯** - 每个状态变化都能追溯到 Action

### 为什么需要显式状态管理？

**隐式状态的问题**：
```go
// ❌ 隐式状态 - 散落在各处
var currentFocus Component    // 谁知道当前焦点在哪？
var globalStyle Style         // 样式从哪来？
var isDirty bool              // 怎么知道谁 dirty？
func makeHandler() func() {
    count := 0                 // 闭包里的隐藏状态
    return func() { count++ }  // 外部无法访问
}
```

**显式状态的优势**：
```go
// ✅ 显式状态 - 集中管理
type AppState struct {
    Components map[string]ComponentState
    Focus      FocusPath
    Modals     []ModalState
    Dirty      DirtyRegion
}

func (s *AppState) Snapshot() StateSnapshot {
    // 完整的状态快照
}
```

## 设计目标

1. **无隐式状态**: 所有状态必须可枚举
2. **可快照**: 任何时刻可以保存完整状态
3. **可回放**: 可以从任意状态恢复
4. **可追溯**: 每个状态变化都关联 Action
5. **AI 友好**: AI 可以读取和操作状态

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        State Flow                                       │
└─────────────────────────────────────────────────────────────────────────┘

    Action
      │
      ▼
┌─────────────────┐
│  State Tracker  │ ◄───── Before Action (Snapshot)
└─────────────────┘
      │
      ▼
┌─────────────────┐
│ Action Dispatch │
└─────────────────┘
      │
      ▼
┌─────────────────┐
│ Component Update│
└─────────────────┘
      │
      ▼
┌─────────────────┐
│  Mark Dirty     │
└─────────────────┘
      │
      ▼
┌─────────────────┐
│ After Action    │ ───► History (for Undo)
│   (Snapshot)    │
└─────────────────┘
```

## 核心类型定义

### 1. State Snapshot

```go
// 位于: tui/runtime/state/snapshot.go

package state

// Snapshot 状态快照
type Snapshot struct {
    // 时间戳
    Timestamp time.Time

    // 焦点路径
    FocusPath FocusPath

    // 组件状态
    Components map[string]ComponentState

    // Modal 栈
    Modals []ModalState

    // 脏区域
    Dirty DirtyRegion

    // 元数据
    Metadata map[string]interface{}
}

// ComponentState 组件状态
type ComponentState struct {
    // 组件标识
    ID string

    // 组件类型
    Type string

    // 静态属性（配置）
    Props map[string]interface{}

    // 动态状态（运行时）
    State map[string]interface{}

    // 布局位置
    Rect Rect

    // 可见性
    Visible bool

    // 可交互性
    Disabled bool
}

// Rect 矩形区域
type Rect struct {
    X      int
    Y      int
    Width  int
    Height int
}

// ModalState Modal 状态
type ModalState struct {
    ID       string
    Type     ModalType
    Focus    string
    Open     bool
    Closable bool
}

// ModalType Modal 类型
type ModalType int

const (
    ModalDialog ModalType = iota
    ModalAlert
    ModalConfirm
    ModalMenu
)

// DirtyRegion 脏区域
type DirtyRegion struct {
    // 脏单元格列表
    Cells []CellRef

    // 脏矩形列表（优化渲染）
    Rects []Rect
}

// CellRef 单元格引用
type CellRef struct {
    X int
    Y int
}

// NewSnapshot 创建状态快照
func NewSnapshot() *Snapshot {
    return &Snapshot{
        Timestamp:  time.Now(),
        FocusPath:  make(FocusPath, 0),
        Components: make(map[string]ComponentState),
        Modals:     make([]ModalState, 0),
        Dirty:      DirtyRegion{},
        Metadata:   make(map[string]interface{}),
    }
}

// Clone 克隆快照
func (s *Snapshot) Clone() *Snapshot {
    components := make(map[string]ComponentState)
    for id, comp := range s.Components {
        components[id] = comp.Clone()
    }

    modals := make([]ModalState, len(s.Modals))
    copy(modals, s.Modals)

    focusPath := make(FocusPath, len(s.FocusPath))
    copy(focusPath, s.FocusPath)

    return &Snapshot{
        Timestamp:  s.Timestamp,
        FocusPath:  focusPath,
        Components: components,
        Modals:     modals,
        Dirty:      s.Dirty,
        Metadata:   copyMap(s.Metadata),
    }
}

// GetComponent 获取组件状态
func (s *Snapshot) GetComponent(id string) (ComponentState, bool) {
    comp, ok := s.Components[id]
    return comp, ok
}

// SetComponent 设置组件状态
func (s *Snapshot) SetComponent(comp ComponentState) {
    s.Components[comp.ID] = comp
}

// GetFocus 获取焦点
func (s *Snapshot) GetFocus() FocusPath {
    return s.FocusPath
}

// SetFocus 设置焦点
func (s *Snapshot) SetFocus(path FocusPath) {
    s.FocusPath = path
}

// Clone 克隆组件状态
func (c *ComponentState) Clone() ComponentState {
    return ComponentState{
        ID:       c.ID,
        Type:     c.Type,
        Props:    copyMap(c.Properties),
        State:    copyMap(c.State),
        Rect:     c.Rect,
        Visible:  c.Visible,
        Disabled: c.Disabled,
    }
}

// copyMap 复制 map
func copyMap(m map[string]interface{}) map[string]interface{} {
    if m == nil {
        return nil
    }
    result := make(map[string]interface{})
    for k, v := range m {
        result[k] = v
    }
    return result
}

// Equal 比较两个快照是否相等
func (s *Snapshot) Equal(other *Snapshot) bool {
    if s == nil && other == nil {
        return true
    }
    if s == nil || other == nil {
        return false
    }

    // 比较焦点路径
    if !s.FocusPath.Equals(other.FocusPath) {
        return false
    }

    // 比较组件数量
    if len(s.Components) != len(other.Components) {
        return false
    }

    // 比较每个组件状态
    for id, comp := range s.Components {
        otherComp, ok := other.Components[id]
        if !ok || !comp.Equal(otherComp) {
            return false
        }
    }

    // 比较 Modal
    if len(s.Modals) != len(other.Modals) {
        return false
    }
    for i, modal := range s.Modals {
        if modal != other.Modals[i] {
            return false
        }
    }

    return true
}

// Equal 比较组件状态
func (c *ComponentState) Equal(other ComponentState) bool {
    if c.ID != other.ID || c.Type != other.Type {
        return false
    }
    if c.Visible != other.Visible || c.Disabled != other.Disabled {
        return false
    }
    // 简化比较，实际可能需要更深入的比较
    return true
}
```

### 2. State Tracker

```go
// 位于: tui/runtime/state/tracker.go

package state

// Tracker 状态追踪器
type Tracker struct {
    mu sync.RWMutex

    // 当前状态快照
    current *Snapshot

    // 历史状态（用于 Undo/Redo）
    past []*Snapshot

    // 未来状态（用于 Redo）
    future []*Snapshot

    // 最大历史记录数
    maxHistory int

    // 变化监听器
    listeners []ChangeListener
}

// ChangeListener 状态变化监听器
type ChangeListener func(old, new *Snapshot)

// NewTracker 创建状态追踪器
func NewTracker() *Tracker {
    return &Tracker{
        current:    NewSnapshot(),
        past:       make([]*Snapshot, 0),
        future:     make([]*Snapshot, 0),
        maxHistory: 100,
        listeners:  make([]ChangeListener, 0),
    }
}

// Current 获取当前状态
func (t *Tracker) Current() *Snapshot {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.current.Clone()
}

// BeforeAction 在执行 Action 前记录状态
func (t *Tracker) BeforeAction() *Snapshot {
    t.mu.Lock()
    defer t.mu.Unlock()
    return t.current.Clone()
}

// AfterAction 在执行 Action 后记录状态
func (t *Tracker) AfterAction(before *Snapshot) *Snapshot {
    t.mu.Lock()
    defer t.mu.Unlock()

    after := t.capture()

    // 只在状态真正变化时才记录历史
    if !t.current.Equal(before) {
        // 清空 future（因为有了新的操作分支）
        t.future = t.future[:0]

        // 保存到 past
        t.past = append(t.past, before)
        if len(t.past) > t.maxHistory {
            t.past = t.past[1:]
        }
    }

    t.notify(before, after)
    return after
}

// capture 捕获当前状态
func (t *Tracker) capture() *Snapshot {
    // 这个方法由 Runtime 实现具体逻辑
    // 这里只返回当前快照的克隆
    return t.current.Clone()
}

// Update 更新当前状态
func (t *Tracker) Update(newSnapshot *Snapshot) {
    t.mu.Lock()
    defer t.mu.Unlock()

    old := t.current
    t.current = newSnapshot
    t.notify(old, newSnapshot)
}

// Undo 撤销
func (t *Tracker) Undo() bool {
    t.mu.Lock()
    defer t.mu.Unlock()

    if len(t.past) == 0 {
        return false
    }

    // 保存当前状态到 future
    t.future = append(t.future, t.current)

    // 恢复到上一个状态
    last := len(t.past) - 1
    t.current = t.past[last]
    t.past = t.past[:last]

    return true
}

// Redo 重做
func (t *Tracker) Redo() bool {
    t.mu.Lock()
    defer t.mu.Unlock()

    if len(t.future) == 0 {
        return false
    }

    // 保存当前状态到 past
    t.past = append(t.past, t.current)

    // 恢复到下一个状态
    last := len(t.future) - 1
    t.current = t.future[last]
    t.future = t.future[:last]

    return true
}

// CanUndo 是否可以撤销
func (t *Tracker) CanUndo() bool {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return len(t.past) > 0
}

// CanRedo 是否可以重做
func (t *Tracker) CanRedo() bool {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return len(t.future) > 0
}

// GetHistory 获取历史记录
func (t *Tracker) GetHistory() []*Snapshot {
    t.mu.RLock()
    defer t.mu.RUnlock()

    history := make([]*Snapshot, len(t.past))
    for i, snap := range t.past {
        history[i] = snap.Clone()
    }
    return history
}

// ClearHistory 清空历史记录
func (t *Tracker) ClearHistory() {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.past = t.past[:0]
    t.future = t.future[:0]
}

// Subscribe 订阅状态变化
func (t *Tracker) Subscribe(listener ChangeListener) func() {
    t.mu.Lock()
    defer t.mu.Unlock()

    t.listeners = append(t.listeners, listener)

    return func() {
        t.Unsubscribe(listener)
    }
}

// Unsubscribe 取消订阅
func (t *Tracker) Unsubscribe(listener ChangeListener) {
    t.mu.Lock()
    defer t.mu.Unlock()

    for i, l := range t.listeners {
        if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
            t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
            break
        }
    }
}

// notify 通知监听器
func (t *Tracker) notify(old, new *Snapshot) {
    listeners := make([]ChangeListener, len(t.listeners))
    copy(listeners, t.listeners)

    for _, listener := range listeners {
        listener(old, new)
    }
}

// SetComponentState 设置组件状态
func (t *Tracker) SetComponentState(id string, state map[string]interface{}) {
    t.mu.Lock()
    defer t.mu.Unlock()

    comp, ok := t.current.Components[id]
    if !ok {
        comp = ComponentState{
            ID:    id,
            State: make(map[string]interface{}),
        }
    }

    comp.State = state
    t.current.Components[id] = comp
}

// GetComponentState 获取组件状态
func (t *Tracker) GetComponentState(id string) (map[string]interface{}, bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()

    comp, ok := t.current.Components[id]
    if !ok {
        return nil, false
    }
    return comp.State, true
}

// SetFocusPath 设置焦点路径
func (t *Tracker) SetFocusPath(path FocusPath) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.current.FocusPath = path.Clone()
}

// GetFocusPath 获取焦点路径
func (t *Tracker) GetFocusPath() FocusPath {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.current.FocusPath.Clone()
}
```

### 3. Component State 接口

```go
// 位于: tui/framework/component/state.go

package component

// Stateful 组件状态接口
type Stateful interface {
    // GetState 获取组件状态
    GetState() map[string]interface{}

    // SetState 设置组件状态
    SetState(state map[string]interface{})

    // GetProps 获取组件属性
    GetProps() map[string]interface{}

    // SetProps 设置组件属性
    SetProps(props map[string]interface{})
}

// StateHolder 状态持有者（嵌入到组件中）
type StateHolder struct {
    // 静态属性（配置）
    props map[string]interface{}

    // 动态状态（运行时）
    state map[string]interface{}
}

// NewStateHolder 创建状态持有者
func NewStateHolder() *StateHolder {
    return &StateHolder{
        props: make(map[string]interface{}),
        state: make(map[string]interface{}),
    }
}

// GetState 获取状态
func (h *StateHolder) GetState() map[string]interface{} {
    return copyMap(h.state)
}

// SetState 设置状态
func (h *StateHolder) SetState(state map[string]interface{}) {
    h.state = copyMap(state)
}

// GetStateValue 获取单个状态值
func (h *StateHolder) GetStateValue(key string) (interface{}, bool) {
    val, ok := h.state[key]
    return val, ok
}

// SetStateValue 设置单个状态值
func (h *StateHolder) SetStateValue(key string, value interface{}) {
    h.state[key] = value
}

// GetProps 获取属性
func (h *StateHolder) GetProps() map[string]interface{} {
    return copyMap(h.props)
}

// SetProps 设置属性
func (h *StateHolder) SetProps(props map[string]interface{}) {
    h.props = copyMap(props)
}

// GetProp 获取单个属性值
func (h *StateHolder) GetProp(key string) (interface{}, bool) {
    val, ok := h.props[key]
    return val, ok
}

// SetProp 设置单个属性值
func (h *StateHolder) SetProp(key string, value interface{}) {
    h.props[key] = value
}

// ExportState 导出状态（用于快照）
func (h *StateHolder) ExportState() state.ComponentState {
    return state.ComponentState{
        Props: copyMap(h.props),
        State: copyMap(h.state),
    }
}

// ImportState 导入状态（用于恢复）
func (h *StateHolder) ImportState(compState state.ComponentState) {
    h.props = copyMap(compState.Props)
    h.state = copyMap(compState.State)
}
```

## 与 Runtime 集成

### Runtime 收集状态

```go
// 位于: tui/runtime/runtime.go

package runtime

func (r *Runtime) collectState() *state.Snapshot {
    snapshot := state.NewSnapshot()

    // 1. 收集焦点路径
    snapshot.FocusPath = r.focus.Current()

    // 2. 收集所有组件状态
    for _, comp := range r.components {
        if stateful, ok := comp.(component.Stateful); ok {
            compState := state.ComponentState{
                ID:       comp.ID(),
                Type:     comp.Type(),
                Props:    stateful.GetProps(),
                State:    stateful.GetState(),
                Rect:     r.layout.GetBounds(comp.ID()),
                Visible:  true,  // 从布局获取
                Disabled: false, // 从组件获取
            }
            snapshot.Components[comp.ID()] = compState
        }
    }

    // 3. 收集 Modal 状态
    snapshot.Modals = r.collectModalStates()

    // 4. 收集脏区域
    snapshot.Dirty = r.dirty.GetRegion()

    return snapshot
}

// ProcessInput 处理输入（完整流程）
func (r *Runtime) ProcessInput(raw RawInput) error {
    // 1. 记录执行前状态
    before := r.stateTracker.BeforeAction()

    // 2. RawInput → Action
    act := r.keyMap.Map(raw)
    if act == nil {
        return nil
    }

    // 3. 分发 Action
    handled := r.dispatcher.Dispatch(act)

    // 4. 收集执行后状态
    after := r.collectState()

    // 5. 记录到历史
    r.stateTracker.AfterAction(before)
    r.stateTracker.Update(after)

    // 6. 如果处理成功，标记 Dirty
    if handled {
        r.dirty.MarkAll()
    }

    return nil
}
```

## 组件状态管理

### TextInput 状态管理

```go
// 位于: tui/framework/component/textinput.go

package component

// TextInput 文本输入组件
type TextInput struct {
    BaseComponent
    *StateHolder

    // 状态字段（方便访问，实际存在 StateHolder.state 中）
    value     string
    cursor    int
    selection Selection
}

// NewTextInput 创建文本输入
func NewTextInput() *TextInput {
    ti := &TextInput{
        StateHolder: NewStateHolder(),
        value:       "",
        cursor:      0,
        selection:   Selection{},
    }

    // 初始化状态
    ti.SetStateValue("value", "")
    ti.SetStateValue("cursor", 0)
    ti.SetStateValue("placeholder", "")

    return ti
}

// GetState 实现 Stateful 接口
func (t *TextInput) GetState() map[string]interface{} {
    return map[string]interface{}{
        "value":      t.value,
        "cursor":     t.cursor,
        "selection":  t.selection,
        "placeholder": t.GetStateValue("placeholder"),
    }
}

// SetState 实现 Stateful 接口
func (t *TextInput) SetState(state map[string]interface{}) {
    if v, ok := state["value"].(string); ok {
        t.value = v
    }
    if v, ok := state["cursor"].(int); ok {
        t.cursor = v
    }
    // ...
}

// SetValue 设置值（通过 Action）
func (t *TextInput) SetValue(value string) {
    t.value = value
    t.cursor = len(value)
    t.SetStateValue("value", value)
    t.SetStateValue("cursor", t.cursor)
    t.MarkDirty()
}

// GetValue 获取值
func (t *TextInput) GetValue() string {
    return t.value
}
```

### List 状态管理

```go
// 位于: tui/framework/component/list.go

package component

// List 列表组件
type List struct {
    BaseComponent
    *StateHolder

    items    []interface{}
    cursor   int
    offset   int
    selected map[int]bool
}

// NewList 创建列表
func NewList() *List {
    l := &List{
        StateHolder: NewStateHolder(),
        items:       make([]interface{}, 0),
        cursor:      0,
        offset:      0,
        selected:    make(map[int]bool),
    }

    l.SetStateValue("items", l.items)
    l.SetStateValue("cursor", l.cursor)
    l.SetStateValue("offset", l.offset)

    return l
}

// GetState 实现 Stateful 接口
func (l *List) GetState() map[string]interface{} {
    return map[string]interface{}{
        "items":    l.items,
        "cursor":   l.cursor,
        "offset":   l.offset,
        "selected": l.selected,
    }
}

// SetItems 设置列表项
func (l *List) SetItems(items []interface{}) {
    l.items = items
    l.SetStateValue("items", items)
    l.MarkDirty()
}

// GetItems 获取列表项
func (l *List) GetItems() []interface{} {
    return l.items
}

// SetCursor 设置光标位置
func (l *List) SetCursor(cursor int) {
    if cursor >= 0 && cursor < len(l.items) {
        l.cursor = cursor
        l.SetStateValue("cursor", cursor)
        l.MarkDirty()
    }
}

// GetCursor 获取光标位置
func (l *List) GetCursor() int {
    return l.cursor
}
```

## 状态序列化

### JSON 序列化

```go
// 位于: tui/runtime/state/serialize.go

package state

// Serialize 序列化状态为 JSON
func (s *Snapshot) Serialize() ([]byte, error) {
    return json.MarshalIndent(s, "", "  ")
}

// Deserialize 从 JSON 反序列化状态
func Deserialize(data []byte) (*Snapshot, error) {
    var snapshot Snapshot
    if err := json.Unmarshal(data, &snapshot); err != nil {
        return nil, err
    }
    return &snapshot, nil
}

// SaveToFile 保存状态到文件
func (s *Snapshot) SaveToFile(path string) error {
    data, err := s.Serialize()
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

// LoadFromFile 从文件加载状态
func LoadFromFile(path string) (*Snapshot, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return Deserialize(data)
}
```

### 状态差异

```go
// 位于: tui/runtime/state/diff.go

package state

// Diff 状态差异
type Diff struct {
    // 变化的组件
    ChangedComponents []string

    // 变化的字段
    ChangedFields map[string][]string

    // 焦点变化
    FocusChanged bool

    // 新增/删除的组件
    AddedComponents   []string
    RemovedComponents []string
}

// ComputeDiff 计算两个快照的差异
func ComputeDiff(before, after *Snapshot) *Diff {
    diff := &Diff{
        ChangedComponents: make([]string, 0),
        ChangedFields:     make(map[string][]string),
        AddedComponents:   make([]string, 0),
        RemovedComponents: make([]string, 0),
    }

    // 检查焦点变化
    diff.FocusChanged = !before.FocusPath.Equals(after.FocusPath)

    // 检查新增组件
    for id := range after.Components {
        if _, ok := before.Components[id]; !ok {
            diff.AddedComponents = append(diff.AddedComponents, id)
        }
    }

    // 检查删除组件
    for id := range before.Components {
        if _, ok := after.Components[id]; !ok {
            diff.RemovedComponents = append(diff.RemovedComponents, id)
        }
    }

    // 检查组件状态变化
    for id, afterComp := range after.Components {
        beforeComp, ok := before.Components[id]
        if !ok {
            continue
        }

        // 比较状态字段
        changed := compareState(beforeComp.State, afterComp.State)
        if len(changed) > 0 {
            diff.ChangedComponents = append(diff.ChangedComponents, id)
            diff.ChangedFields[id] = changed
        }
    }

    return diff
}

// compareState 比较状态字段
func compareState(before, after map[string]interface{}) []string {
    changed := make([]string, 0)

    for key, afterVal := range after {
        beforeVal, ok := before[key]
        if !ok || !reflect.DeepEqual(beforeVal, afterVal) {
            changed = append(changed, key)
        }
    }

    return changed
}
```

## AI 集成

### AI 读取状态

```go
// 位于: tui/runtime/ai/state.go

package ai

// StateQuery 状态查询
type StateQuery struct {
    // 组件 ID
    ComponentID string

    // 组件类型
    ComponentType string

    // 状态键
    StateKey string
}

// QueryState 查询状态
func (c *Controller) QueryState(query StateQuery) (map[string]interface{}, error) {
    snapshot := c.runtime.stateTracker.Current()

    if query.ComponentID != "" {
        // 查询特定组件
        comp, ok := snapshot.GetComponent(query.ComponentID)
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
        // 查询特定类型的所有组件
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

// SetState 设置状态（通过 Action）
func (c *Controller) SetState(componentID string, state map[string]interface{}) error {
    // 创建设置状态的 Action
    action := action.NewAction(action.ActionSetState).
        WithTarget(componentID).
        WithPayload(state)

    return c.Dispatch(action)
}
```

## 时间旅行调试

### 时间旅行 API

```go
// 位于: tui/runtime/state/timetravel.go

package state

// TimeMachine 时间机器
type TimeMachine struct {
    tracker *Tracker
}

// NewTimeMachine 创建时间机器
func NewTimeMachine(tracker *Tracker) *TimeMachine {
    return &TimeMachine{tracker: tracker}
}

// GoTo 跳转到指定历史状态
func (tm *TimeMachine) GoTo(index int) bool {
    history := tm.tracker.GetHistory()
    if index < 0 || index >= len(history) {
        return false
    }

    target := history[index]
    tm.tracker.Update(target)
    return true
}

// Replay 重放历史
func (tm *TimeMachine) Replay(from, to int) bool {
    history := tm.tracker.GetHistory()
    if from < 0 || to >= len(history) || from > to {
        return false
    }

    for i := from; i <= to; i++ {
        tm.tracker.Update(history[i])
    }

    return true
}

// GetTimeline 获取时间线
func (tm *TimeMachine) GetTimeline() []TimelineEntry {
    history := tm.tracker.GetHistory()
    current := tm.tracker.Current()

    entries := make([]TimelineEntry, len(history)+1)
    for i, snap := range history {
        entries[i] = TimelineEntry{
            Index:     i,
            Timestamp: snap.Timestamp,
            Focus:     snap.FocusPath,
            Summary:   summarizeSnapshot(snap),
        }
    }

    // 添加当前状态
    entries[len(history)] = TimelineEntry{
        Index:     len(history),
        Timestamp: current.Timestamp,
        Focus:     current.FocusPath,
        Summary:   "Current",
        Current:   true,
    }

    return entries
}

// TimelineEntry 时间线条目
type TimelineEntry struct {
    Index     int
    Timestamp time.Time
    Focus     FocusPath
    Summary   string
    Current   bool
}

// summarizeSnapshot 总结快照
func summarizeSnapshot(s *Snapshot) string {
    return fmt.Sprintf("%d components, focus: %s",
        len(s.Components), s.FocusPath)
}
```

## 测试

### 单元测试

```go
// 位于: tui/runtime/state/snapshot_test.go

package state

func TestSnapshotClone(t *testing.T) {
    original := NewSnapshot()
    original.FocusPath = NewFocusPath("root", "main", "form")
    original.Components["input-1"] = ComponentState{
        ID:  "input-1",
        Type: "TextInput",
        State: map[string]interface{}{
            "value": "hello",
        },
    }

    clone := original.Clone()

    // 修改原始不应该影响克隆
    original.FocusPath = NewFocusPath("root")
    original.Components["input-1"].State["value"] = "world"

    assert.Equal(t, NewFocusPath("root", "main", "form"), clone.FocusPath)
    assert.Equal(t, "hello", clone.Components["input-1"].State["value"])
}

func TestSnapshotEqual(t *testing.T) {
    s1 := NewSnapshot()
    s2 := NewSnapshot()

    assert.True(t, s1.Equal(s2))

    s1.FocusPath = NewFocusPath("root")
    assert.False(t, s1.Equal(s2))
}
```

```go
// 位于: tui/runtime/state/tracker_test.go

package state

func TestTrackerUndoRedo(t *testing.T) {
    tracker := NewTracker()

    // 初始状态
    before := tracker.BeforeAction()
    tracker.SetFocusPath(NewFocusPath("input-1"))
    tracker.AfterAction(before)

    // 第一次修改
    before = tracker.BeforeAction()
    tracker.SetFocusPath(NewFocusPath("input-2"))
    tracker.AfterAction(before)

    assert.Equal(t, NewFocusPath("input-2"), tracker.GetFocusPath())
    assert.True(t, tracker.CanUndo())

    // Undo
    tracker.Undo()
    assert.Equal(t, NewFocusPath("input-1"), tracker.GetFocusPath())
    assert.True(t, tracker.CanRedo())

    // Redo
    tracker.Redo()
    assert.Equal(t, NewFocusPath("input-2"), tracker.GetFocusPath())
}

func TestTrackerListener(t *testing.T) {
    tracker := NewTracker()

    var oldState, newState *Snapshot
    called := false

    tracker.Subscribe(func(old, new *Snapshot) {
        called = true
        oldState = old
        newState = new
    })

    before := tracker.BeforeAction()
    tracker.SetFocusPath(NewFocusPath("test"))
    tracker.AfterAction(before)

    assert.True(t, called)
    assert.NotNil(t, oldState)
    assert.NotNil(t, newState)
}
```

## 最佳实践

### 1. 状态设计原则

```go
// ✅ 好的状态设计 - 扁平化
type GoodState struct {
    Value     string
    Cursor    int
    Selected  bool
    Disabled  bool
}

// ❌ 坏的状态设计 - 嵌套过深
type BadState struct {
    UI struct {
        Input struct {
            Value struct {
                Content string
            }
            Cursor struct {
                Position int
            }
        }
    }
}
```

### 2. 状态更新

```go
// ✅ 正确：通过 Action 更新状态
func (c *Component) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionInputText:
        if text, ok := a.Payload.(string); ok {
            c.SetValue(text)  // 内部会调用 SetStateValue
            return true
        }
    }
    return false
}

// ❌ 错误：直接修改内部字段
func (c *Component) HandleAction(a *action.Action) bool {
    c.value = "text"  // 绕过状态管理
    return true
}
```

### 3. 状态派生

```go
// ✅ 派生状态计算，不存储
func (t *TextInput) DisplayValue() string {
    if t.cursor >= len(t.value) {
        return t.value
    }
    // 根据光标位置计算显示
    return t.value[:t.cursor] + "|" + t.value[t.cursor:]
}

// ❌ 存储派生状态
type BadTextInput struct {
    value        string
    cursor       int
    displayValue string  // 这应该动态计算
}
```

## 总结

状态管理系统提供：

1. **可枚举**: 所有状态集中管理，可以完整列出
2. **可快照**: 任何时刻可以保存和恢复状态
3. **可回放**: 支持 Undo/Redo 和时间旅行
4. **可追溯**: 每个状态变化都关联 Action
5. **AI 友好**: AI 可以读取、查询和操作状态
6. **可测试**: 状态是纯数据，易于测试

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [FOCUS_SYSTEM.md](./FOCUS_SYSTEM.md) - 焦点系统
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - AI 集成
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 架构不变量
