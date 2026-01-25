package state

import (
	"sync"
)

// ==============================================================================
// State Tracker (V3)
// ==============================================================================
// Tracker 负责跟踪状态变化，支持 Undo/Redo 和时间旅行

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
// 这个方法由 Runtime 实现具体逻辑
func (t *Tracker) capture() *Snapshot {
	// 返回当前快照的克隆
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

	// 通知监听器
	for _, listener := range t.listeners {
		listener(nil, t.current)
	}

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

	// 通知监听器
	for _, listener := range t.listeners {
		listener(nil, t.current)
	}

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
// 返回取消订阅的函数
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
		// 简单的引用比较
		if &l == &listener {
			t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
			break
		}
	}
}

// notify 通知监听器
func (t *Tracker) notify(old, new *Snapshot) {
	// 复制监听器列表避免在持有锁的情况下调用
	listeners := make([]ChangeListener, len(t.listeners))
	copy(listeners, t.listeners)

	for _, listener := range listeners {
		listener(old, new)
	}
}

// SetMaxHistory 设置最大历史记录数
func (t *Tracker) SetMaxHistory(max int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxHistory = max
}

// GetMaxHistory 获取最大历史记录数
func (t *Tracker) GetMaxHistory() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.maxHistory
}

// GetHistorySize 获取当前历史记录数
func (t *Tracker) GetHistorySize() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.past)
}

// GetFutureSize 获取未来记录数
func (t *Tracker) GetFutureSize() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.future)
}

// ClearFuture 清空未来记录
func (t *Tracker) ClearFuture() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.future = t.future[:0]
}
