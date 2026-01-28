package action

import (
	"sync"
	"testing"
)

// MockComponent 测试用 Mock 组件
type MockComponent struct {
	id        string
	lastValue interface{}
	handled   bool
	mu        sync.Mutex
}

func NewMockComponent(id string) *MockComponent {
	return &MockComponent{id: id}
}

func (m *MockComponent) ID() string {
	return m.id
}

func (m *MockComponent) HandleAction(a *Action) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastValue = a.Payload
	m.handled = true
	return true
}

func (m *MockComponent) GetLastValue() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastValue
}

func (m *MockComponent) WasHandled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.handled
}

func (m *MockComponent) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastValue = nil
	m.handled = false
}

// MockComponentReject 总是返回 false 的 Mock 组件
type MockComponentReject struct {
	id string
}

func NewMockComponentReject(id string) *MockComponentReject {
	return &MockComponentReject{id: id}
}

func (m *MockComponentReject) ID() string {
	return m.id
}

func (m *MockComponentReject) HandleAction(a *Action) bool {
	return false
}

func TestDispatcherRegister(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")

	d.Register(mock)

	target, exists := d.GetTarget("test-1")
	if !exists {
		t.Errorf("target was not registered")
	}

	if target.ID() != "test-1" {
		t.Errorf("registered target ID mismatch")
	}
}

func TestDispatcherUnregister(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")

	d.Register(mock)
	d.Unregister("test-1")

	_, exists := d.GetTarget("test-1")
	if exists {
		t.Errorf("target was not unregistered")
	}
}

func TestDispatcherDispatch(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	a := NewAction(ActionInputText).WithTarget("test-1").WithPayload("hello")
	handled := d.Dispatch(a)

	if !handled {
		t.Errorf("expected action to be handled")
	}

	if mock.GetLastValue() != "hello" {
		t.Errorf("expected payload 'hello', got '%v'", mock.GetLastValue())
	}
}

func TestDispatcherDispatchToTarget(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	a := NewAction(ActionInputText).WithPayload("hello")
	handled := d.DispatchToTarget("test-1", a)

	if !handled {
		t.Errorf("expected action to be handled")
	}

	if mock.GetLastValue() != "hello" {
		t.Errorf("expected payload 'hello', got '%v'", mock.GetLastValue())
	}

	// 注意：由于 DispatchToTarget 使用 defer 恢复了原始 Target，
	// 所以这里 a.Target 应该是空的（或者是原来的值）
}

func TestDispatcherGlobalHandler(t *testing.T) {
	d := NewDispatcher()

	called := false
	d.Subscribe(ActionQuit, func(a *Action) bool {
		called = true
		return true
	})

	a := NewAction(ActionQuit)
	d.Dispatch(a)

	if !called {
		t.Errorf("global handler was not called")
	}
}

func TestDispatcherGlobalHandlerPriority(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	globalCalled := false
	d.Subscribe(ActionInputText, func(a *Action) bool {
		globalCalled = true
		return true // 全局处理器返回 true，阻止传递
	})

	a := NewAction(ActionInputText).WithTarget("test-1").WithPayload("hello")
	d.Dispatch(a)

	if !globalCalled {
		t.Errorf("global handler was not called")
	}

	if mock.WasHandled() {
		t.Errorf("component should not have been called when global handler returns true")
	}
}

func TestDispatcherUnsubscribe(t *testing.T) {
	d := NewDispatcher()

	callCount := 0
	handler := func(a *Action) bool {
		callCount++
		return true
	}

	unsub := d.Subscribe(ActionQuit, handler)

	// 第一次调用
	d.Dispatch(NewAction(ActionQuit))

	// 取消订阅
	unsub()

	// 第二次调用（不应该触发处理器）
	d.Dispatch(NewAction(ActionQuit))

	if callCount != 1 {
		t.Errorf("expected handler to be called once, got %d times", callCount)
	}
}

func TestDispatcherDefaultHandler(t *testing.T) {
	d := NewDispatcher()

	defaultCalled := false
	d.SetDefaultHandler(func(a *Action) bool {
		defaultCalled = true
		return true
	})

	// 分发到未注册的目标
	a := NewAction(ActionInputText).WithTarget("non-existent")
	d.Dispatch(a)

	if !defaultCalled {
		t.Errorf("default handler was not called")
	}
}

func TestDispatcherStats(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	stats := d.Stats()

	if targets, ok := stats["targets"].(int); !ok || targets != 1 {
		t.Errorf("expected 1 target, got %v", stats["targets"])
	}

	if handlers, ok := stats["global_handlers"].(int); !ok || handlers != 0 {
		t.Errorf("expected 0 global handlers, got %v", stats["global_handlers"])
	}
}

func TestDispatcherLog(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	d.EnableLog(true)

	a := NewAction(ActionInputText).WithTarget("test-1").WithPayload("hello")
	d.Dispatch(a)

	logs := d.GetLog()
	if len(logs) == 0 {
		t.Errorf("expected log entries")
	}

	if !logs[0].Handled {
		t.Errorf("expected log entry to be marked as handled")
	}
}

func TestDispatcherClearLog(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	d.EnableLog(true)
	d.Dispatch(NewAction(ActionInputText).WithTarget("test-1"))

	d.ClearLog()

	logs := d.GetLog()
	if len(logs) != 0 {
		t.Errorf("expected log to be cleared")
	}
}

func TestDispatcherMultipleTargets(t *testing.T) {
	d := NewDispatcher()
	mock1 := NewMockComponent("test-1")
	mock2 := NewMockComponent("test-2")

	d.Register(mock1)
	d.Register(mock2)

	// 分发到第一个目标
	d.DispatchToTarget("test-1", NewAction(ActionInputText).WithPayload("first"))

	// 分发到第二个目标
	d.DispatchToTarget("test-2", NewAction(ActionInputText).WithPayload("second"))

	if mock1.GetLastValue() != "first" {
		t.Errorf("expected 'first' for mock1, got '%v'", mock1.GetLastValue())
	}

	if mock2.GetLastValue() != "second" {
		t.Errorf("expected 'second' for mock2, got '%v'", mock2.GetLastValue())
	}
}

func TestDispatcherTargetNotHandled(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponentReject("test-1")
	d.Register(mock)

	a := NewAction(ActionInputText).WithTarget("test-1")
	handled := d.Dispatch(a)

	if handled {
		t.Errorf("expected action to not be handled")
	}
}

func TestDispatcherString(t *testing.T) {
	d := NewDispatcher()
	mock := NewMockComponent("test-1")
	d.Register(mock)

	d.Subscribe(ActionQuit, func(a *Action) bool {
		return true
	})

	s := d.String()
	if s == "" {
		t.Errorf("String() returned empty string")
	}
}
