# Focus System Design (V3)

> **版本**: V3
> **核心概念**: Focus Path（焦点路径）
> **关键特性**: Scope 支持、可恢复性、AI 友好

## 概述

Focus Path 系统是 TUI 框架管理焦点状态的核心机制。与简单的 `bool` 焦点不同，Focus Path 提供了一个层级化的、可追溯的焦点管理方式，支持 Modal、Dialog 等复杂场景。

### 为什么需要 Focus Path？

**简单布尔焦点的问题**：
```go
// ❌ 简单布尔焦点
type Component struct {
    focused bool  // 只知道是否聚焦，不知道如何到达
}

// 问题：
// 1. 关闭 Modal 后，焦点无法恢复
// 2. 无法追踪焦点历史
// 3. AI 无法知道焦点层级
```

**Focus Path 的优势**：
```go
// ✅ Focus Path
type FocusPath []string  // ["root", "main", "form", "input-email"]

// 优势：
// 1. 完整记录焦点层级
// 2. 关闭 Modal 后可恢复
// 3. AI 可以精确控制和查询
```

## 设计目标

1. **层级化**: 焦点路径反映组件树的层级结构
2. **可追溯**: 知道焦点如何到达当前位置
3. **Scope 支持**: 支持 Modal/Dialog 的焦点锁定
4. **可恢复**: 关闭 Modal 后焦点能回到之前位置
5. **AI 友好**: AI 可以精确查询和控制焦点

## 核心概念

### Focus Path

```
FocusPath = ["root", "main", "form", "input-email"]
                           └───── 当前焦点 ─────┘
```

**特点**：
- 完整记录从根到当前焦点的路径
- 可以追踪焦点历史
- 支持"跳转到父级"等操作

### Focus Scope

```
正常状态:
  FocusPath: ["root", "main", "form", "input-email"]
  Scopes:    []

Modal 打开:
  FocusPath: ["root", "main", "form", "input-email", "modal-confirm", "button-ok"]
  Scopes:    ["modal-confirm"]  ← 焦点被锁定在这个 scope

Modal 关闭:
  FocusPath: ["root", "main", "form", "input-email"]  ← 恢复之前的焦点
  Scopes:    []
```

## 核心类型定义

### 1. Focus Path

```go
// 位于: tui/runtime/focus/path.go

package focus

// FocusPath 焦点路径
type FocusPath []string

// NewFocusPath 创建焦点路径
func NewFocusPath(ids ...string) FocusPath {
    return FocusPath(ids)
}

// Current 获取当前焦点 ID
func (p FocusPath) Current() string {
    if len(p) == 0 {
        return ""
    }
    return p[len(p)-1]
}

// Parent 获取父级焦点 ID
func (p FocusPath) Parent() string {
    if len(p) <= 1 {
        return ""
    }
    return p[len(p)-2]
}

// Root 获取根节点 ID
func (p FocusPath) Root() string {
    if len(p) == 0 {
        return ""
    }
    return p[0]
}

// Append 追加节点
func (p FocusPath) Append(id string) FocusPath {
    newPath := make(FocusPath, len(p)+1)
    copy(newPath, p)
    newPath[len(p)] = id
    return newPath
}

// Remove 移除最后节点
func (p FocusPath) Remove() FocusPath {
    if len(p) == 0 {
        return p
    }
    return p[:len(p)-1]
}

// Contains 检查是否包含某个节点
func (p FocusPath) Contains(id string) bool {
    for _, node := range p {
        if node == id {
            return true
        }
    }
    return false
}

// Depth 获取深度
func (p FocusPath) Depth() int {
    return len(p)
}

// String 返回字符串表示
func (p FocusPath) String() string {
    return strings.Join(p, " → ")
}

// CommonPrefix 获取公共前缀
func (p FocusPath) CommonPrefix(other FocusPath) FocusPath {
    i := 0
    for i < len(p) && i < len(other) && p[i] == other[i] {
        i++
    }
    return p[:i]
}

// IsAncestor 检查是否是祖先
func (p FocusPath) IsAncestor(other FocusPath) bool {
    if len(p) >= len(other) {
        return false
    }
    for i, id := range p {
        if other[i] != id {
            return false
        }
    }
    return true
}

// Equals 检查是否相等
func (p FocusPath) Equals(other FocusPath) bool {
    if len(p) != len(other) {
        return false
    }
    for i, id := range p {
        if other[i] != id {
            return false
        }
    }
    return true
}

// Clone 克隆路径
func (p FocusPath) Clone() FocusPath {
    newPath := make(FocusPath, len(p))
    copy(newPath, p)
    return newPath
}
```

### 2. Focus Scope

```go
// 位于: tui/runtime/focus/scope.go

package focus

// FocusScope 焦点作用域
type FocusScope struct {
    // 唯一标识
    ID string

    // 类型
    Type ScopeType

    // 焦点策略
    Strategy FocusStrategy

    // 可聚焦的组件（有序列表）
    Focusables []string

    // 循环导航
    Cyclic bool

    // 元数据
    Metadata map[string]interface{}
}

// ScopeType 作用域类型
type ScopeType int

const (
    ScopeNormal ScopeType = iota // 普通作用域
    ScopeModal                   // 模态框（独占焦点）
    ScopeDialog                  // 对话框（独占焦点）
    ScopePopover                 // 弹出菜单（独占焦点）
    ScopeForm                    // 表单内部作用域
)

// FocusStrategy 焦点策略
type FocusStrategy int

const (
    FocusSequential FocusStrategy = iota // 顺序导航
    FocusDirectional                      // 方向导航（上下左右）
    FocusPositional                       // 位置导航（基于几何位置）
    FocusCustom                           // 自定义导航
)

// NewScope 创建作用域
func NewScope(id string, scopeType ScopeType) *FocusScope {
    return &FocusScope{
        ID:         id,
        Type:       scopeType,
        Strategy:   FocusSequential,
        Focusables: make([]string, 0),
        Cyclic:     true,
        Metadata:   make(map[string]interface{}),
    }
}

// IsModal 是否是模态作用域
func (s *FocusScope) IsModal() bool {
    return s.Type == ScopeModal || s.Type == ScopeDialog
}

// AddFocusable 添加可聚焦组件
func (s *FocusScope) AddFocusable(id string) {
    // 避免重复
    for _, fid := range s.Focusables {
        if fid == id {
            return
        }
    }
    s.Focusables = append(s.Focusables, id)
}

// RemoveFocusable 移除可聚焦组件
func (s *FocusScope) RemoveFocusable(id string) {
    for i, fid := range s.Focusables {
        if fid == id {
            s.Focusables = append(s.Focusables[:i], s.Focusables[i+1:]...)
            break
        }
    }
}

// Next 获取下一个焦点
func (s *FocusScope) Next(current string) string {
    if len(s.Focusables) == 0 {
        return ""
    }

    if current == "" {
        return s.Focusables[0]
    }

    for i, fid := range s.Focusables {
        if fid == current {
            if i < len(s.Focusables)-1 {
                return s.Focusables[i+1]
            }
            if s.Cyclic {
                return s.Focusables[0]
            }
            return current
        }
    }

    return s.Focusables[0]
}

// Prev 获取上一个焦点
func (s *FocusScope) Prev(current string) string {
    if len(s.Focusables) == 0 {
        return ""
    }

    if current == "" {
        return s.Focusables[len(s.Focusables)-1]
    }

    for i, fid := range s.Focusables {
        if fid == current {
            if i > 0 {
                return s.Focusables[i-1]
            }
            if s.Cyclic {
                return s.Focusables[len(s.Focusables)-1]
            }
            return current
        }
    }

    return s.Focusables[len(s.Focusables)-1]
}

// Contains 检查是否包含组件
func (s *FocusScope) Contains(id string) bool {
    for _, fid := range s.Focusables {
        if fid == id {
            return true
        }
    }
    return false
}

// SetFocusables 设置可聚焦组件列表
func (s *FocusScope) SetFocusables(ids []string) {
    s.Focusables = make([]string, len(ids))
    copy(s.Focusables, ids)
}

// GetIndex 获取组件在焦点列表中的索引
func (s *FocusScope) GetIndex(id string) int {
    for i, fid := range s.Focusables {
        if fid == id {
            return i
        }
    }
    return -1
}

// FocusAt 获取指定索引的焦点
func (s *FocusScope) FocusAt(index int) string {
    if index < 0 || index >= len(s.Focusables) {
        return ""
    }
    return s.Focusables[index]
}
```

### 3. Focus Manager

```go
// 位于: tui/runtime/focus/manager.go

package focus

// Manager 焦点管理器
type Manager struct {
    mu sync.RWMutex

    // 当前焦点路径
    path FocusPath

    // 作用域栈
    scopes []*FocusScope

    // 历史记录（用于恢复）
    history []FocusPath

    // 最大历史记录数
    maxHistory int

    // 焦点变化监听器
    listeners []FocusChangeListener
}

// FocusChangeListener 焦点变化监听器
type FocusChangeListener func(oldPath, newPath FocusPath)

// FocusChangeResult 焦点变化结果
type FocusChangeResult struct {
    Success bool
    OldPath FocusPath
    NewPath FocusPath
    Error   error
}

// NewManager 创建焦点管理器
func NewManager() *Manager {
    return &Manager{
        path:       make(FocusPath, 0),
        scopes:     make([]*FocusScope, 0),
        history:    make([]FocusPath, 0),
        maxHistory: 100,
        listeners:  make([]FocusChangeListener, 0),
    }
}

// Current 获取当前焦点路径
func (m *Manager) Current() FocusPath {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.path.Clone()
}

// CurrentID 获取当前焦点 ID
func (m *Manager) CurrentID() string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.path.Current()
}

// CurrentScope 获取当前作用域
func (m *Manager) CurrentScope() *FocusScope {
    m.mu.RLock()
    defer m.mu.RUnlock()
    if len(m.scopes) == 0 {
        return nil
    }
    return m.scopes[len(m.scopes)-1]
}

// SetFocus 设置焦点
func (m *Manager) SetFocus(id string) FocusChangeResult {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 检查是否在当前作用域内
    if !m.isInCurrentScope(id) {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("component %s not in current scope", id),
        }
    }

    oldPath := m.path
    m.path = m.path.Append(id)
    m.notify(oldPath, m.path)

    return FocusChangeResult{
        Success: true,
        OldPath: oldPath,
        NewPath: m.path.Clone(),
    }
}

// SetPath 设置焦点路径
func (m *Manager) SetPath(path FocusPath) FocusChangeResult {
    m.mu.Lock()
    defer m.mu.Unlock()

    if !m.isValidPath(path) {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: path,
            Error:   fmt.Errorf("invalid focus path"),
        }
    }

    oldPath := m.path
    m.saveHistory()
    m.path = path.Clone()
    m.notify(oldPath, m.path)

    return FocusChangeResult{
        Success: true,
        OldPath: oldPath,
        NewPath: m.path.Clone(),
    }
}

// PushScope 推入作用域
func (m *Manager) PushScope(scope *FocusScope) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.saveHistory()
    m.scopes = append(m.scopes, scope)
}

// PopScope 弹出作用域
func (m *Manager) PopScope() {
    m.mu.Lock()
    defer m.mu.Unlock()

    if len(m.scopes) == 0 {
        return
    }

    oldPath := m.path
    m.scopes = m.scopes[:len(m.scopes)-1]

    // 恢复到作用域前的焦点
    if len(m.history) > 0 {
        m.path = m.history[len(m.history)-1].Clone()
        m.history = m.history[:len(m.history)-1]
    } else {
        m.path = m.path[:0]
    }

    m.notify(oldPath, m.path)
}

// Next 焦点移到下一个
func (m *Manager) Next() FocusChangeResult {
    m.mu.Lock()
    defer m.mu.Unlock()

    currentID := m.path.Current()
    scope := m.CurrentScope()

    if scope == nil {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("no focus scope"),
        }
    }

    nextID := scope.Next(currentID)
    if nextID == "" || nextID == currentID {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("no next focusable"),
        }
    }

    oldPath := m.path
    m.path = m.path.Remove().Append(nextID)
    m.notify(oldPath, m.path)

    return FocusChangeResult{
        Success: true,
        OldPath: oldPath,
        NewPath: m.path.Clone(),
    }
}

// Prev 焦点移到上一个
func (m *Manager) Prev() FocusChangeResult {
    m.mu.Lock()
    defer m.mu.Unlock()

    currentID := m.path.Current()
    scope := m.CurrentScope()

    if scope == nil {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("no focus scope"),
        }
    }

    prevID := scope.Prev(currentID)
    if prevID == "" || prevID == currentID {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("no previous focusable"),
        }
    }

    oldPath := m.path
    m.path = m.path.Remove().Append(prevID)
    m.notify(oldPath, m.path)

    return FocusChangeResult{
        Success: true,
        OldPath: oldPath,
        NewPath: m.path.Clone(),
    }
}

// Navigate 导航到指定方向
func (m *Manager) Navigate(direction Direction) FocusChangeResult {
    m.mu.Lock()
    defer m.mu.Unlock()

    scope := m.CurrentScope()
    if scope == nil {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("no focus scope"),
        }
    }

    currentID := m.path.Current()
    var targetID string

    switch direction {
    case DirectionNext:
        targetID = scope.Next(currentID)
    case DirectionPrev:
        targetID = scope.Prev(currentID)
    case DirectionUp, DirectionDown, DirectionLeft, DirectionRight:
        if scope.Strategy == FocusDirectional || scope.Strategy == FocusPositional {
            targetID = m.findNeighbor(currentID, direction, scope)
        } else {
            // 顺序策略下的方向映射
            switch direction {
            case DirectionUp:
                targetID = scope.Prev(currentID)
            case DirectionDown:
                targetID = scope.Next(currentID)
            }
        }
    }

    if targetID == "" || targetID == currentID {
        return FocusChangeResult{
            Success: false,
            OldPath: m.path.Clone(),
            NewPath: m.path.Clone(),
            Error:   fmt.Errorf("no target in direction %v", direction),
        }
    }

    oldPath := m.path
    m.path = m.path.Remove().Append(targetID)
    m.notify(oldPath, m.path)

    return FocusChangeResult{
        Success: true,
        OldPath: oldPath,
        NewPath: m.path.Clone(),
    }
}

// Direction 方向
type Direction int

const (
    DirectionNext Direction = iota
    DirectionPrev
    DirectionUp
    DirectionDown
    DirectionLeft
    DirectionRight
)

// Register 注册可聚焦组件
func (m *Manager) Register(scopeID, componentID string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    scope := m.findScope(scopeID)
    if scope == nil {
        scope = NewScope(scopeID, ScopeNormal)
        m.scopes = append(m.scopes, scope)
    }
    scope.AddFocusable(componentID)
}

// Unregister 注销可聚焦组件
func (m *Manager) Unregister(componentID string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, scope := range m.scopes {
        scope.RemoveFocusable(componentID)
    }
}

// Subscribe 订阅焦点变化
func (m *Manager) Subscribe(listener FocusChangeListener) func() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.listeners = append(m.listeners, listener)

    return func() {
        m.Unsubscribe(listener)
    }
}

// Unsubscribe 取消订阅
func (m *Manager) Unsubscribe(listener FocusChangeListener) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i, l := range m.listeners {
        // 使用函数指针比较
        if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
            m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
            break
        }
    }
}

// GetHistory 获取历史记录
func (m *Manager) GetHistory() []FocusPath {
    m.mu.RLock()
    defer m.mu.RUnlock()

    history := make([]FocusPath, len(m.history))
    for i, path := range m.history {
        history[i] = path.Clone()
    }
    return history
}

// notify 通知监听器
func (m *Manager) notify(oldPath, newPath FocusPath) {
    // 复制监听器列表以避免在回调中修改
    listeners := make([]FocusChangeListener, len(m.listeners))
    copy(listeners, m.listeners)

    for _, listener := range listeners {
        listener(oldPath.Clone(), newPath.Clone())
    }
}

// saveHistory 保存历史记录
func (m *Manager) saveHistory() {
    m.history = append(m.history, m.path.Clone())
    if len(m.history) > m.maxHistory {
        m.history = m.history[1:]
    }
}

// isInCurrentScope 检查是否在当前作用域内
func (m *Manager) isInCurrentScope(id string) bool {
    scope := m.CurrentScope()
    if scope == nil {
        return true // 没有作用域限制
    }
    return scope.Contains(id)
}

// isValidPath 检查路径是否有效
func (m *Manager) isValidPath(path FocusPath) bool {
    if len(path) == 0 {
        return true
    }

    currentID := path.Current()
    return m.isInCurrentScope(currentID)
}

// findScope 查找作用域
func (m *Manager) findScope(id string) *FocusScope {
    for _, scope := range m.scopes {
        if scope.ID == id {
            return scope
        }
    }
    return nil
}

// findNeighbor 查找相邻组件（方向导航）
func (m *Manager) findNeighbor(currentID string, direction Direction, scope *FocusScope) string {
    // 需要结合组件位置信息
    // 这里简化处理，实际实现需要组件的位置坐标
    focusables := scope.Focusables
    currentIndex := scope.GetIndex(currentID)
    if currentIndex < 0 {
        return ""
    }

    switch direction {
    case DirectionUp:
        if currentIndex > 0 {
            return focusables[currentIndex-1]
        }
    case DirectionDown:
        if currentIndex < len(focusables)-1 {
            return focusables[currentIndex+1]
        }
    }

    return ""
}

// IsFocused 检查组件是否聚焦
func (m *Manager) IsFocused(id string) bool {
    return m.CurrentID() == id
}

// GetScopes 获取所有作用域
func (m *Manager) GetScopes() []*FocusScope {
    m.mu.RLock()
    defer m.mu.RUnlock()

    scopes := make([]*FocusScope, len(m.scopes))
    copy(scopes, m.scopes)
    return scopes
}
```

### 4. Modal Manager

```go
// 位于: tui/runtime/focus/modal.go

package focus

// ModalManager 模态框管理器
type ModalManager struct {
    focus *Manager
}

// NewModalManager 创建模态框管理器
func NewModalManager(focus *Manager) *ModalManager {
    return &ModalManager{focus: focus}
}

// ShowModal 显示模态框
func (m *ModalManager) ShowModal(modalID string, options ...ModalOption) FocusChangeResult {
    // 保存当前焦点路径
    beforePath := m.focus.Current()

    // 创建 Modal Scope
    scope := NewScope(modalID, ScopeModal)
    scope.Strategy = FocusSequential
    for _, opt := range options {
        opt(scope)
    }

    // 推入 Scope
    m.focus.PushScope(scope)

    // 设置焦点到 Modal 的第一个可聚焦组件
    if len(scope.Focusables) > 0 {
        result := m.focus.SetFocus(scope.Focusables[0])
        if result.Success {
            return result
        }
    }

    // 如果没有可聚焦组件，至少聚焦到 modal 本身
    return m.focus.SetFocus(modalID)
}

// HideModal 隐藏模态框
func (m *ModalManager) HideModal(modalID string) FocusChangeResult {
    beforePath := m.focus.Current()

    // 检查栈顶是否是这个 modal
    currentScope := m.focus.CurrentScope()
    if currentScope == nil || currentScope.ID != modalID {
        return FocusChangeResult{
            Success: false,
            OldPath: beforePath,
            NewPath: beforePath,
            Error:   fmt.Errorf("modal %s is not on top of stack", modalID),
        }
    }

    // 弹出 Scope（会自动恢复之前的焦点）
    m.focus.PopScope()

    return FocusChangeResult{
        Success: true,
        OldPath: beforePath,
        NewPath: m.focus.Current(),
    }
}

// IsModalActive 检查模态框是否激活
func (m *ModalManager) IsModalActive(modalID string) bool {
    for _, scope := range m.focus.GetScopes() {
        if scope.ID == modalID && scope.IsModal() {
            return true
        }
    }
    return false
}

// ModalOption 模态框选项
type ModalOption func(*FocusScope)

// WithFocusables 设置可聚焦组件
func WithFocusables(ids ...string) ModalOption {
    return func(s *FocusScope) {
        s.Focusables = make([]string, len(ids))
        copy(s.Focusables, ids)
    }
}

// WithInitialFocus 设置初始焦点
func WithInitialFocus(id string) ModalOption {
    return func(s *FocusScope) {
        if len(s.Focusables) == 0 {
            s.Focusables = []string{id}
        } else {
            // 将 id 移到前面
            for i, fid := range s.Focusables {
                if fid == id {
                    s.Focusables = append([]string{id}, append(s.Focusables[:i], s.Focusables[i+1:]...)...)
                    break
                }
            }
        }
    }
}

// WithCyclic 设置循环导航
func WithCyclic(cyclic bool) ModalOption {
    return func(s *FocusScope) {
        s.Cyclic = cyclic
    }
}

// WithStrategy 设置焦点策略
func WithStrategy(strategy FocusStrategy) ModalOption {
    return func(s *FocusScope) {
        s.Strategy = strategy
    }
}
```

## 与 Action 系统集成

### 导航 Action 处理

```go
// 位于: tui/framework/component/focus_handler.go

package component

// FocusHandler 焦点处理器
type FocusHandler struct {
    focusMgr *focus.Manager
}

// NewFocusHandler 创建焦点处理器
func NewFocusHandler(focusMgr *focus.Manager) *FocusHandler {
    return &FocusHandler{focusMgr: focusMgr}
}

// HandleAction 处理导航动作
func (h *FocusHandler) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateNext:
        result := h.focusMgr.Next()
        return result.Success

    case action.ActionNavigatePrev:
        result := h.focusMgr.Prev()
        return result.Success

    case action.ActionNavigateUp:
        result := h.focusMgr.Navigate(focus.DirectionUp)
        return result.Success

    case action.ActionNavigateDown:
        result := h.focusMgr.Navigate(focus.DirectionDown)
        return result.Success

    case action.ActionNavigateLeft:
        result := h.focusMgr.Navigate(focus.DirectionLeft)
        return result.Success

    case action.ActionNavigateRight:
        result := h.focusMgr.Navigate(focus.DirectionRight)
        return result.Success
    }

    return false
}

// IsFocused 检查是否聚焦
func (h *FocusHandler) IsFocused(id string) bool {
    return h.focusMgr.IsFocused(id)
}

// CurrentFocus 获取当前焦点
func (h *FocusHandler) CurrentFocus() string {
    return h.focusMgr.CurrentID()
}
```

### Container 组件集成

```go
// 位于: tui/framework/component/container.go

package component

// Container 容器组件
type Container struct {
    BaseComponent

    // 子组件
    children []Component

    // 焦点管理
    focusHandler *FocusHandler
}

// HandleAction 处理动作
func (c *Container) HandleAction(a *action.Action) bool {
    // 优先处理导航动作
    if c.focusHandler != nil {
        if c.focusHandler.HandleAction(a) {
            return true
        }
    }

    // 分发给当前焦点子组件
    currentID := c.focusHandler.CurrentFocus()
    if currentID != "" {
        for _, child := range c.children {
            if child.ID() == currentID {
                return child.HandleAction(a)
            }
        }
    }

    return false
}

// Mount 挂载组件
func (c *Container) Mount(rt Runtime) error {
    // 注册子组件到焦点管理器
    for _, child := range c.children {
        if focusable, ok := child.(Focusable); ok && focusable.CanFocus() {
            c.focusMgr.Register(c.ID(), child.ID())
        }
    }
    return nil
}
```

## 使用示例

### 基本使用

```go
// 创建焦点管理器
focusMgr := focus.NewManager()

// 创建作用域并添加可聚焦组件
formScope := focus.NewScope("login-form", focus.ScopeForm)
formScope.AddFocusable("username")
formScope.AddFocusable("password")
formScope.AddFocusable("remember")
formScope.AddFocusable("submit")

focusMgr.PushScope(formScope)

// 设置初始焦点
result := focusMgr.SetFocus("username")
if !result.Success {
    log.Printf("Failed to set focus: %v", result.Error)
}

// 导航
result = focusMgr.Next()
if result.Success {
    fmt.Printf("Focus moved: %s → %s\n", result.OldPath, result.NewPath)
}

result = focusMgr.Prev()
fmt.Printf("Focus moved: %s → %s\n", result.OldPath, result.NewPath)
```

### Modal 使用

```go
focusMgr := focus.NewManager()
modalMgr := focus.NewModalManager(focusMgr)

// 设置初始焦点
focusMgr.SetFocus("input-1")
fmt.Println(focusMgr.CurrentID()) // "input-1"

// 显示 Modal
result := modalMgr.ShowModal("confirm-dialog",
    focus.WithFocusables("btn-yes", "btn-no"),
    focus.WithInitialFocus("btn-no"),
)

if result.Success {
    fmt.Println(focusMgr.CurrentID()) // "btn-no"
    fmt.Println(len(focusMgr.GetScopes())) // 2 (form + modal)
}

// 用户导航...
focusMgr.Navigate(focus.DirectionNext)
fmt.Println(focusMgr.CurrentID()) // "btn-yes"

// 关闭 Modal
result = modalMgr.HideModal("confirm-dialog")
if result.Success {
    fmt.Println(focusMgr.CurrentID()) // "input-1" - 自动恢复
    fmt.Println(len(focusMgr.GetScopes())) // 1 (只有 form)
}
```

### 表单使用

```go
// 创建表单 Scope
formScope := focus.NewScope("login-form", focus.ScopeForm)
formScope.AddFocusable("username")
formScope.AddFocusable("password")
formScope.AddFocusable("remember")
formScope.AddFocusable("submit")

focusMgr.PushScope(formScope)
focusMgr.SetFocus("username")

// Tab 导航
focusMgr.Next()  // username → password
focusMgr.Next()  // password → remember
focusMgr.Next()  // remember → submit
focusMgr.Next()  // submit → username (循环)
```

### 监听焦点变化

```go
focusMgr := focus.NewManager()

// 订阅焦点变化
unsubscribe := focusMgr.Subscribe(func(oldPath, newPath focus.FocusPath) {
    fmt.Printf("Focus changed: %s → %s\n", oldPath, newPath)
})

// ... 执行一些焦点操作

// 取消订阅
unsubscribe()
```

## AI 集成

### AI 查询焦点

```go
// AI 可以精确查询当前焦点状态
type FocusInfo struct {
    Current   string
    Path      []string
    Scope     string
    ScopeType string
    Available []string
}

func (ai *AIController) GetFocusInfo() FocusInfo {
    mgr := ai.runtime.focusMgr

    currentScope := mgr.CurrentScope()
    scopeInfo := ""
    scopeType := ""
    if currentScope != nil {
        scopeInfo = currentScope.ID
        scopeType = fmt.Sprintf("%v", currentScope.Type)
    }

    return FocusInfo{
        Current:   mgr.CurrentID(),
        Path:      mgr.Current(),
        Scope:     scopeInfo,
        ScopeType: scopeType,
        Available: currentScope.Focusables,
    }
}
```

### AI 控制焦点

```go
// AI 可以精确控制焦点
func (ai *AIController) SetFocus(componentID string) error {
    result := ai.runtime.focusMgr.SetFocus(componentID)
    if !result.Success {
        return result.Error
    }
    return nil
}

func (ai *AIController) NavigateNext() error {
    result := ai.runtime.focusMgr.Next()
    if !result.Success {
        return result.Error
    }
    return nil
}

func (ai *AIController) NavigatePrev() error {
    result := ai.runtime.focusMgr.Prev()
    if !result.Success {
        return result.Error
    }
    return nil
}
```

## 几何导航

### 位置感知的焦点导航

```go
// 位于: tui/runtime/focus/geometry.go

package focus

// GeometryNavigator 几何导航器
type GeometryNavigator struct {
    // 组件位置缓存
    positions map[string]Rect
}

// Rect 矩形区域
type Rect struct {
    X      int
    Y      int
    Width  int
    Height int
}

// NewGeometryNavigator 创建几何导航器
func NewGeometryNavigator() *GeometryNavigator {
    return &GeometryNavigator{
        positions: make(map[string]Rect),
    }
}

// UpdatePosition 更新组件位置
func (n *GeometryNavigator) UpdatePosition(id string, rect Rect) {
    n.positions[id] = rect
}

// FindNeighbor 查找相邻组件
func (n *GeometryNavigator) FindNeighbor(current string, direction Direction, focusables []string) string {
    currentRect, ok := n.positions[current]
    if !ok {
        return ""
    }

    var best string
    var bestDist int = math.MaxInt32

    for _, id := range focusables {
        if id == current {
            continue
        }

        rect, ok := n.positions[id]
        if !ok {
            continue
        }

        dist := n.distance(currentRect, rect, direction)
        if dist > 0 && dist < bestDist {
            best = id
            bestDist = dist
        }
    }

    return best
}

// distance 计算距离（考虑方向）
func (n *GeometryNavigator) distance(from, to Rect, direction Direction) int {
    fromCenter := n.center(from)
    toCenter := n.center(to)

    switch direction {
    case DirectionUp:
        // 必须在上方
        if toCenter.Y >= fromCenter.Y {
            return -1
        }
        // 计算垂直距离 + 水平偏移
        dy := fromCenter.Y - toCenter.Y
        dx := abs(fromCenter.X - toCenter.X)
        return dy*dy + dx*dx

    case DirectionDown:
        // 必须在下方
        if toCenter.Y <= fromCenter.Y {
            return -1
        }
        dy := toCenter.Y - fromCenter.Y
        dx := abs(fromCenter.X - toCenter.X)
        return dy*dy + dx*dx

    case DirectionLeft:
        // 必须在左方
        if toCenter.X >= fromCenter.X {
            return -1
        }
        dx := fromCenter.X - toCenter.X
        dy := abs(fromCenter.Y - toCenter.Y)
        return dx*dx + dy*dy

    case DirectionRight:
        // 必须在右方
        if toCenter.X <= fromCenter.X {
            return -1
        }
        dx := toCenter.X - fromCenter.X
        dy := abs(fromCenter.Y - toCenter.Y)
        return dx*dx + dy*dy
    }

    return -1
}

// center 计算中心点
func (n *GeometryNavigator) center(r Rect) Point {
    return Point{
        X: r.X + r.Width/2,
        Y: r.Y + r.Height/2,
    }
}

// Point 点
type Point struct {
    X int
    Y int
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

## 测试

### 单元测试

```go
// 位于: tui/runtime/focus/path_test.go

package focus

func TestFocusPath(t *testing.T) {
    path := NewFocusPath("root", "main", "form", "input")

    assert.Equal(t, "input", path.Current())
    assert.Equal(t, "form", path.Parent())
    assert.Equal(t, "root", path.Root())
    assert.Equal(t, 4, path.Depth())
    assert.True(t, path.Contains("form"))
    assert.True(t, path.IsAncestor(NewFocusPath("root", "main", "form", "input", "cursor")))

    // 测试追加
    newPath := path.Append("cursor")
    assert.Equal(t, 5, newPath.Depth())
    assert.Equal(t, "cursor", newPath.Current())

    // 测试移除
    removed := newPath.Remove()
    assert.Equal(t, 4, removed.Depth())
    assert.Equal(t, path, removed)
}

func TestFocusPathEquals(t *testing.T) {
    path1 := NewFocusPath("root", "main", "form")
    path2 := NewFocusPath("root", "main", "form")
    path3 := NewFocusPath("root", "main", "input")

    assert.True(t, path1.Equals(path2))
    assert.False(t, path1.Equals(path3))
}

func TestFocusPathCommonPrefix(t *testing.T) {
    path1 := NewFocusPath("root", "main", "form", "input")
    path2 := NewFocusPath("root", "main", "dialog", "button")

    prefix := path1.CommonPrefix(path2)
    assert.Equal(t, 2, len(prefix))
    assert.Equal(t, []string{"root", "main"}, []string(prefix))
}
```

```go
// 位于: tui/runtime/focus/manager_test.go

package focus

func TestFocusManager(t *testing.T) {
    mgr := NewManager()
    scope := NewScope("test", ScopeNormal)
    scope.AddFocusable("a")
    scope.AddFocusable("b")
    scope.AddFocusable("c")

    mgr.PushScope(scope)

    // 测试设置焦点
    result := mgr.SetFocus("a")
    assert.True(t, result.Success)
    assert.Equal(t, "a", mgr.CurrentID())

    // 测试导航
    result = mgr.Next()
    assert.True(t, result.Success)
    assert.Equal(t, "b", mgr.CurrentID())

    result = mgr.Next()
    assert.True(t, result.Success)
    assert.Equal(t, "c", mgr.CurrentID())

    // 测试循环
    result = mgr.Next()
    assert.True(t, result.Success)
    assert.Equal(t, "a", mgr.CurrentID())
}

func TestFocusManagerHistory(t *testing.T) {
    mgr := NewManager()
    scope := NewScope("test", ScopeNormal)
    scope.AddFocusable("a", "b", "c")
    mgr.PushScope(scope)

    mgr.SetFocus("a")
    mgr.SetFocus("b")
    mgr.SetFocus("c")

    history := mgr.GetHistory()
    assert.Equal(t, 2, len(history)) // a, b
    assert.Equal(t, []string{"root", "a"}, []string(history[0]))
}

func TestFocusManagerListener(t *testing.T) {
    mgr := NewManager()
    scope := NewScope("test", ScopeNormal)
    scope.AddFocusable("a", "b")
    mgr.PushScope(scope)

    var oldPath, newPath FocusPath
    called := false
    mgr.Subscribe(func(old, new FocusPath) {
        called = true
        oldPath = old
        newPath = new
    })

    mgr.SetFocus("a")

    assert.True(t, called)
    assert.Equal(t, 0, len(oldPath))
    assert.Equal(t, 1, len(newPath))
}
```

```go
// 位于: tui/runtime/focus/modal_test.go

package focus

func TestModalScope(t *testing.T) {
    focusMgr := NewManager()
    modalMgr := NewModalManager(focusMgr)

    // 设置初始焦点
    formScope := NewScope("form", ScopeForm)
    formScope.AddFocusable("input-1", "input-2")
    focusMgr.PushScope(formScope)
    focusMgr.SetFocus("input-1")

    assert.Equal(t, "input-1", focusMgr.CurrentID())

    // 显示 Modal
    result := modalMgr.ShowModal("modal",
        WithFocusables("modal-btn-1", "modal-btn-2"),
        WithInitialFocus("modal-btn-1"),
    )

    assert.True(t, result.Success)
    assert.Equal(t, "modal-btn-1", focusMgr.CurrentID())
    assert.True(t, modalMgr.IsModalActive("modal"))

    // 导航
    focusMgr.Next()
    assert.Equal(t, "modal-btn-2", focusMgr.CurrentID())

    // 关闭 Modal
    result = modalMgr.HideModal("modal")
    assert.True(t, result.Success)
    assert.Equal(t, "input-1", focusMgr.CurrentID()) // 焦点恢复
    assert.False(t, modalMgr.IsModalActive("modal"))
}

func TestNestedModal(t *testing.T) {
    focusMgr := NewManager()
    modalMgr := NewModalManager(focusMgr)

    // 第一个 Modal
    modalMgr.ShowModal("modal1",
        WithFocusables("btn1", "btn2"),
        WithInitialFocus("btn1"),
    )
    assert.Equal(t, "btn1", focusMgr.CurrentID())

    // 第二个 Modal（嵌套）
    modalMgr.ShowModal("modal2",
        WithFocusables("btn3", "btn4"),
        WithInitialFocus("btn3"),
    )
    assert.Equal(t, "btn3", focusMgr.CurrentID())

    // 关闭第二个 Modal
    modalMgr.HideModal("modal2")
    assert.Equal(t, "btn1", focusMgr.CurrentID()) // 恢复到 modal1 的焦点

    // 关闭第一个 Modal
    modalMgr.HideModal("modal1")
    assert.Equal(t, "", focusMgr.CurrentID())
}
```

## 总结

Focus Path 系统提供：

1. **层级化焦点**: FocusPath 反映组件树结构
2. **Scope 支持**: 支持 Modal/Dialog 焦点锁定
3. **可恢复性**: 关闭 Modal 后自动恢复焦点
4. **灵活导航**: 支持顺序导航和方向导航
5. **可测试**: 纯数据结构，易于测试
6. **AI 友好**: AI 可以准确知道和控制焦点

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [EVENT_SYSTEM.md](./EVENT_SYSTEM.md) - 事件系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - AI 集成
