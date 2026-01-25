package component

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime/state"
)

// ==============================================================================
// State Holder (V3)
// ==============================================================================
// 组件状态管理器

// StateHolder 状态持有者
type StateHolder struct {
	mu    sync.RWMutex
	state map[string]interface{}
	props map[string]interface{}
}

// NewStateHolder 创建状态持有者
func NewStateHolder() *StateHolder {
	return &StateHolder{
		state: make(map[string]interface{}),
		props: make(map[string]interface{}),
	}
}

// ============================================================================
// 状态管理
// ============================================================================

// GetState 获取完整状态
func (s *StateHolder) GetState() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{}, len(s.state))
	for k, v := range s.state {
		result[k] = v
	}
	return result
}

// SetState 设置完整状态
func (s *StateHolder) SetState(state map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = make(map[string]interface{}, len(state))
	for k, v := range state {
		s.state[k] = v
	}
}

// GetStateValue 获取状态值
func (s *StateHolder) GetStateValue(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.state[key]
	return v, ok
}

// SetStateValue 设置状态值
func (s *StateHolder) SetStateValue(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = value
}

// DeleteStateValue 删除状态值
func (s *StateHolder) DeleteStateValue(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.state, key)
}

// HasStateValue 检查状态值是否存在
func (s *StateHolder) HasStateValue(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.state[key]
	return ok
}

// ============================================================================
// 属性管理
// ============================================================================

// GetProps 获取完整属性
func (s *StateHolder) GetProps() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{}, len(s.props))
	for k, v := range s.props {
		result[k] = v
	}
	return result
}

// SetProps 设置完整属性
func (s *StateHolder) SetProps(props map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.props = make(map[string]interface{}, len(props))
	for k, v := range props {
		s.props[k] = v
	}
}

// GetProp 获取属性值
func (s *StateHolder) GetProp(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.props[key]
	return v, ok
}

// SetProp 设置属性值
func (s *StateHolder) SetProp(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.props[key] = value
}

// DeleteProp 删除属性值
func (s *StateHolder) DeleteProp(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.props, key)
}

// HasProp 检查属性是否存在
func (s *StateHolder) HasProp(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.props[key]
	return ok
}

// GetPropString 获取字符串属性
func (s *StateHolder) GetPropString(key string, defaultValue string) string {
	if v, ok := s.GetProp(key); ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetPropInt 获取整数属性
func (s *StateHolder) GetPropInt(key string, defaultValue int) int {
	if v, ok := s.GetProp(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return defaultValue
}

// GetPropBool 获取布尔属性
func (s *StateHolder) GetPropBool(key string, defaultValue bool) bool {
	if v, ok := s.GetProp(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// ============================================================================
// 导入导出
// ============================================================================

// ExportState 导出到 state.ComponentState
func (s *StateHolder) ExportState(id string) *state.ComponentState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &state.ComponentState{
		ID:    id,
		State: s.GetState(),
		Props: s.GetProps(),
	}
}

// ImportState 从 state.ComponentState 导入
func (s *StateHolder) ImportState(cs *state.ComponentState) {
	if cs == nil {
		return
	}
	s.SetState(cs.State)
	s.SetProps(cs.Props)
}

// MergeState 合并状态
func (s *StateHolder) MergeState(other map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range other {
		s.state[k] = v
	}
}

// MergeProps 合并属性
func (s *StateHolder) MergeProps(other map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range other {
		s.props[k] = v
	}
}

// Clear 清空状态和属性
func (s *StateHolder) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = make(map[string]interface{})
	s.props = make(map[string]interface{})
}

// ClearState 清空状态
func (s *StateHolder) ClearState() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = make(map[string]interface{})
}

// ClearProps 清空属性
func (s *StateHolder) ClearProps() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.props = make(map[string]interface{})
}

// StateCount 返回状态数量
func (s *StateHolder) StateCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.state)
}

// PropsCount 返回属性数量
func (s *StateHolder) PropsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.props)
}
