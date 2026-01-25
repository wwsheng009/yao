package action

import (
	"fmt"
	"sync"
)

// ==============================================================================
// Action Dispatcher (V3)
// ==============================================================================
// Dispatcher 负责将 Action 分发到正确的处理器
// 分发顺序：全局处理器 → 焦点目标 → 指定目标

// Target Action 目标接口
// 组件实现此接口以接收 Action
type Target interface {
	// ID 返回组件 ID
	ID() string

	// HandleAction 处理 Action
	// 返回 true 表示已处理，false 表示继续传递
	HandleAction(a *Action) bool
}

// Handler Action 处理器函数类型
type Handler func(a *Action) bool

// Dispatcher Action 分发器
type Dispatcher struct {
	mu             sync.RWMutex
	targets        map[string]Target
	globalHandlers map[ActionType][]Handler
	defaultHandler Handler
	log            bool
	logEntries     []LogEntry
	logMaxSize     int
}

// LogEntry 日志条目
type LogEntry struct {
	Action   *Action
	Target   string
	Handled  bool
	Duration int64 // 纳秒
}

// NewDispatcher 创建 Action 分发器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		targets:        make(map[string]Target),
		globalHandlers: make(map[ActionType][]Handler),
		logMaxSize:     1000,
	}
}

// Register 注册目标
func (d *Dispatcher) Register(target Target) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.targets[target.ID()] = target
}

// Unregister 注销目标
func (d *Dispatcher) Unregister(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.targets, id)
}

// Subscribe 订阅全局 Action 处理
func (d *Dispatcher) Subscribe(actionType ActionType, handler Handler) func() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.globalHandlers[actionType] = append(d.globalHandlers[actionType], handler)

	return func() {
		d.Unsubscribe(actionType, handler)
	}
}

// Unsubscribe 取消订阅
func (d *Dispatcher) Unsubscribe(actionType ActionType, handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	handlers := d.globalHandlers[actionType]
	for i, h := range handlers {
		if &h == &handler { // 简单指针比较
			d.globalHandlers[actionType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// SetDefaultHandler 设置默认处理器
func (d *Dispatcher) SetDefaultHandler(handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.defaultHandler = handler
}

// Dispatch 分发 Action
// 按顺序尝试：全局处理器 → 焦点目标 → 指定目标
// 返回 true 表示 Action 已被处理
func (d *Dispatcher) Dispatch(a *Action) bool {
	d.logAction(a, "dispatch")

	// 1. 全局处理器
	if d.dispatchGlobal(a) {
		d.logAction(a, "handled_by_global")
		return true
	}

	// 2. 指定目标
	if a.Target != "" {
		if d.dispatchToTarget(a, a.Target) {
			d.logAction(a, "handled_by_target")
			return true
		}
	}

	// 3. 默认处理器
	if d.defaultHandler != nil {
		if d.defaultHandler(a) {
			d.logAction(a, "handled_by_default")
			return true
		}
	}

	d.logAction(a, "not_handled")
	return false
}

// dispatchGlobal 分发到全局处理器
func (d *Dispatcher) dispatchGlobal(a *Action) bool {
	d.mu.RLock()
	handlers := d.globalHandlers[a.Type]
	d.mu.RUnlock()

	for _, handler := range handlers {
		if handler(a) {
			return true
		}
	}
	return false
}

// dispatchToTarget 分发到指定目标
func (d *Dispatcher) dispatchToTarget(a *Action, targetID string) bool {
	d.mu.RLock()
	target, exists := d.targets[targetID]
	d.mu.RUnlock()

	if !exists {
		return false
	}

	return target.HandleAction(a)
}

// DispatchToFocus 分发到焦点目标
// 需要焦点管理器配合
func (d *Dispatcher) DispatchToFocus(a *Action, focusID string) bool {
	if focusID == "" {
		return d.Dispatch(a)
	}

	// 临时设置目标
	oldTarget := a.Target
	a.Target = focusID
	defer func() {
		a.Target = oldTarget
	}()

	return d.Dispatch(a)
}

// GetTarget 获取目标
func (d *Dispatcher) GetTarget(id string) (Target, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	target, exists := d.targets[id]
	return target, exists
}

// EnableLog 启用日志
func (d *Dispatcher) EnableLog(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log = enabled
}

// GetLog 获取日志
func (d *Dispatcher) GetLog() []LogEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	log := make([]LogEntry, len(d.logEntries))
	copy(log, d.logEntries)
	return log
}

// ClearLog 清空日志
func (d *Dispatcher) ClearLog() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logEntries = d.logEntries[:0]
}

// logAction 记录 Action
func (d *Dispatcher) logAction(a *Action, status string) {
	if !d.log {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	entry := LogEntry{
		Action:  a,
		Handled: status == "handled",
	}

	d.logEntries = append(d.logEntries, entry)

	// 限制日志大小
	if len(d.logEntries) > d.logMaxSize {
		d.logEntries = d.logEntries[1:]
	}
}

// PrintLog 打印日志
func (d *Dispatcher) PrintLog() {
	entries := d.GetLog()
	for _, entry := range entries {
		if entry.Handled {
			fmt.Printf("[✓] %s -> %s\n", entry.Action, entry.Target)
		} else {
			fmt.Printf("[✗] %s (not handled)\n", entry.Action)
		}
	}
}

// Stats 返回统计信息
func (d *Dispatcher) Stats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"targets":         len(d.targets),
		"global_handlers": len(d.globalHandlers),
		"log_size":        len(d.logEntries),
		"log_enabled":     d.log,
	}
}

// String 返回分���器状态字符串
func (d *Dispatcher) String() string {
	stats := d.Stats()
	return fmt.Sprintf("Dispatcher{targets=%d, handlers=%d}",
		stats["targets"], stats["global_handlers"])
}
