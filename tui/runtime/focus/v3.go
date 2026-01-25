package focus

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime/state"
	"github.com/yaoapp/yao/tui/runtime/layout"
)

// ==============================================================================
// Focus System V3
// ==============================================================================
// V3 焦点系统，使用 FocusPath 而不是索引

// Scope V3 焦点域
type Scope struct {
	mu sync.RWMutex

	// ID 域ID
	ID string

	// Root 根组件ID
	Root string

	// FocusPath 当前焦点路径
	FocusPath state.FocusPath

	// Focusable 可聚焦组件列表
	Focusable []string

	// Active 是否激活
	Active bool

	// Modal 是否为 Modal 域
	Modal bool
}

// NewScope 创建焦点域
func NewScope(id, root string) *Scope {
	return &Scope{
		ID:         id,
		Root:       root,
		FocusPath:  make(state.FocusPath, 0),
		Focusable:  make([]string, 0),
		Active:     false,
		Modal:      false,
	}
}

// SetFocusable 设置可聚焦组件列表
func (s *Scope) SetFocusable(ids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Focusable = ids
}

// GetFocusable 获取可聚焦组件列表
func (s *Scope) GetFocusable() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, len(s.Focusable))
	copy(result, s.Focusable)
	return result
}

// SetFocusPath 设置焦点路径
func (s *Scope) SetFocusPath(path state.FocusPath) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FocusPath = path
}

// GetFocusPath 获取焦点路径
func (s *Scope) GetFocusPath() state.FocusPath {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.FocusPath.Clone()
}

// FocusNext 移动焦点到下一个组件
func (s *Scope) FocusNext() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Focusable) == 0 {
		return "", false
	}

	current := s.FocusPath.Current()
	currentIndex := -1
	for i, id := range s.Focusable {
		if id == current {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(s.Focusable)
	nextID := s.Focusable[nextIndex]

	// 更新焦点路径
	s.FocusPath = s.FocusPath[:0] // 清空当前路径
	s.FocusPath = append(s.FocusPath, nextID)

	return nextID, true
}

// FocusPrev 移动焦点到上一个组件
func (s *Scope) FocusPrev() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Focusable) == 0 {
		return "", false
	}

	current := s.FocusPath.Current()
	currentIndex := -1
	for i, id := range s.Focusable {
		if id == current {
			currentIndex = i
			break
		}
	}

	if currentIndex <= 0 {
		currentIndex = len(s.Focusable)
	}
	prevIndex := currentIndex - 1
	prevID := s.Focusable[prevIndex]

	// 更新焦点路径
	s.FocusPath = s.FocusPath[:0]
	s.FocusPath = append(s.FocusPath, prevID)

	return prevID, true
}

// FocusFirst 移动焦点到第一个组件
func (s *Scope) FocusFirst() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Focusable) == 0 {
		return "", false
	}

	firstID := s.Focusable[0]
	s.FocusPath = state.FocusPath{firstID}
	return firstID, true
}

// FocusLast 移动焦点到最后一个组件
func (s *Scope) FocusLast() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Focusable) == 0 {
		return "", false
	}

	lastID := s.Focusable[len(s.Focusable)-1]
	s.FocusPath = state.FocusPath{lastID}
	return lastID, true
}

// FocusSpecific 设置焦点到指定组件
func (s *Scope) FocusSpecific(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, focusable := range s.Focusable {
		if focusable == id {
			s.FocusPath = state.FocusPath{id}
			return true
		}
	}
	return false
}

// GetFocused 获取当前聚焦的组件ID
func (s *Scope) GetFocused() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.FocusPath) == 0 {
		return "", false
	}
	return s.FocusPath.Current(), true
}

// HasFocus 检查指定组件是否有焦点
func (s *Scope) HasFocus(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.FocusPath) == 0 {
		return false
	}
	return s.FocusPath.Current() == id
}

// IsFocusable 检查组件是否可聚焦
func (s *Scope) IsFocusable(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, focusable := range s.Focusable {
		if focusable == id {
			return true
		}
	}
	return false
}

// Activate 激活焦点域
func (s *Scope) Activate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = true
}

// Deactivate 停用焦点域
func (s *Scope) Deactivate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = false
}

// IsActive 检查域是否激活
func (s *Scope) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Active
}

// IsModal 检查是否为 Modal 域
func (s *Scope) IsModal() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Modal
}

// SetModal 设置是否为 Modal 域
func (s *Scope) SetModal(modal bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Modal = modal
}

// ManagerV3 V3 焦点管理器
type ManagerV3 struct {
	mu sync.RWMutex

	// scopes 焦点域栈
	scopes []*Scope

	// global 全局焦点路径
	globalFocusPath state.FocusPath
}

// NewManagerV3 创建 V3 焦点管理器
func NewManagerV3() *ManagerV3 {
	return &ManagerV3{
		scopes:           make([]*Scope, 0),
		globalFocusPath:  make(state.FocusPath, 0),
	}
}

// PushScope 推入焦点域
func (m *ManagerV3) PushScope(scope *Scope) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 停用当前域
	if len(m.scopes) > 0 {
		m.scopes[len(m.scopes)-1].Deactivate()
	}

	// 激活新域
	scope.Activate()
	m.scopes = append(m.scopes, scope)
}

// PopScope 弹出焦点域
func (m *ManagerV3) PopScope() *Scope {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.scopes) == 0 {
		return nil
	}

	// 弹出当前域
	scope := m.scopes[len(m.scopes)-1]
	scope.Deactivate()
	m.scopes = m.scopes[:len(m.scopes)-1]

	// 激活上一个域
	if len(m.scopes) > 0 {
		m.scopes[len(m.scopes)-1].Activate()
	}

	return scope
}

// CurrentScope 获取当前焦点域
func (m *ManagerV3) CurrentScope() *Scope {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.scopes) == 0 {
		return nil
	}
	return m.scopes[len(m.scopes)-1]
}

// GetActiveScope 获取激活的焦点域
func (m *ManagerV3) GetActiveScope() *Scope {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.scopes) - 1; i >= 0; i-- {
		if m.scopes[i].IsActive() {
			return m.scopes[i]
		}
	}
	return nil
}

// FocusNext 移动焦点到下一个组件
func (m *ManagerV3) FocusNext() (string, bool) {
	scope := m.GetActiveScope()
	if scope == nil {
		return "", false
	}
	return scope.FocusNext()
}

// FocusPrev 移动焦点到上一个组件
func (m *ManagerV3) FocusPrev() (string, bool) {
	scope := m.GetActiveScope()
	if scope == nil {
		return "", false
	}
	return scope.FocusPrev()
}

// FocusFirst 移动焦点到第一个组件
func (m *ManagerV3) FocusFirst() (string, bool) {
	scope := m.GetActiveScope()
	if scope == nil {
		return "", false
	}
	return scope.FocusFirst()
}

// FocusSpecific 设置焦点到指定组件
func (m *ManagerV3) FocusSpecific(id string) bool {
	scope := m.GetActiveScope()
	if scope == nil {
		return false
	}
	return scope.FocusSpecific(id)
}

// GetFocused 获取当前聚焦的组件ID
func (m *ManagerV3) GetFocused() (string, bool) {
	scope := m.GetActiveScope()
	if scope == nil {
		return "", false
	}
	return scope.GetFocused()
}

// GetFocusPath 获取焦点路径
func (m *ManagerV3) GetFocusPath() state.FocusPath {
	scope := m.GetActiveScope()
	if scope == nil {
		return make(state.FocusPath, 0)
	}
	return scope.GetFocusPath()
}

// HasFocus 检查指定组件是否有焦点
func (m *ManagerV3) HasFocus(id string) bool {
	scope := m.GetActiveScope()
	if scope == nil {
		return false
	}
	return scope.HasFocus(id)
}

// CollectFocusable 收集可聚焦组件
// 从组件树中收集实现了 Focusable 接口的组件
func (m *ManagerV3) CollectFocusable(root layout.Node) []string {
	result := make([]string, 0)

	m.collectFocusableRecursive(root, &result)

	return result
}

// collectFocusableRecursive 递归收集可聚焦组件
func (m *ManagerV3) collectFocusableRecursive(node layout.Node, result *[]string) {
	// 所有节点都是可聚焦的（简化实现）
	// 实际实现中可以根据组件类型或其他属性判断
	*result = append(*result, node.ID())

	// 递归处理子节点
	for _, child := range node.Children() {
		m.collectFocusableRecursive(child, result)
	}
}

// UpdateFocusables 更新当前域的可聚焦组件列表
func (m *ManagerV3) UpdateFocusables(root layout.Node) {
	scope := m.GetActiveScope()
	if scope == nil {
		return
	}

	focusable := m.CollectFocusable(root)
	scope.SetFocusable(focusable)
}

// Clear 清除所有焦点域
func (m *ManagerV3) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, scope := range m.scopes {
		scope.Deactivate()
	}
	m.scopes = make([]*Scope, 0)
	m.globalFocusPath = make(state.FocusPath, 0)
}

// HasActiveScope 检查是否有激活的焦点域
func (m *ManagerV3) HasActiveScope() bool {
	return m.GetActiveScope() != nil
}

// IsModalActive 检查是否有 Modal 域激活
func (m *ManagerV3) IsModalActive() bool {
	scope := m.GetActiveScope()
	if scope == nil {
		return false
	}
	return scope.IsModal()
}
