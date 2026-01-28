package action

import (
	"testing"

	"github.com/yaoapp/yao/tui/runtime/state"
)

// TestActionSystemIntegration 测试 Action 系统的完整流程
// RawInput → KeyMap → Action → Dispatcher → Component → State Update
func TestActionSystemIntegration(t *testing.T) {
	// 1. 创建 Dispatcher
	disp := NewDispatcher()

	// 2. 创建 State Tracker
	tracker := state.NewTracker()

	// 3. 创建测试组件
	component := NewMockComponent("test-input")

	// 4. 注册组件
	disp.Register(component)

	// 5. 记录初始状态
	before := tracker.BeforeAction()

	// 6. 创建并分发 Action
	act := NewAction(ActionInputText).WithTarget("test-input").WithPayload("hello")
	handled := disp.Dispatch(act)

	// 7. 验证结果
	if !handled {
		t.Errorf("Action should be handled")
	}

	if component.GetLastValue() != "hello" {
		t.Errorf("Component should receive 'hello', got '%v'", component.GetLastValue())
	}

	// 8. 记录执行后状态
	after := tracker.AfterAction(before)

	if after == nil {
		t.Errorf("AfterAction should return a snapshot")
	}

	// 9. 验证状态变化
	if before == after {
		t.Errorf("State should have changed")
	}
}

// TestActionChain 测试 Action 链式处理
func TestActionChain(t *testing.T) {
	disp := NewDispatcher()

	// 创建测试组件
	component := NewMockComponent("test")
	disp.Register(component)

	// 订阅 Action
	globalHandled := false
	unsub := disp.Subscribe(ActionInputText, func(a *Action) bool {
		globalHandled = true
		// 不处理，让组件处理
		return false
	})
	defer unsub()

	// 分发 Action
	act := NewAction(ActionInputText).WithTarget("test").WithPayload("test")
	handled := disp.Dispatch(act)

	if !handled {
		t.Errorf("Action should be handled")
	}

	// 全局处理器被调用但不返回 true
	if !globalHandled {
		t.Errorf("Global handler should be called")
	}

	// 组件应该接收到
	if component.GetLastValue() != "test" {
		t.Errorf("Component should receive 'test'")
	}
}

// TestActionPriority 测试 Action 处理优先级
// 全局处理器 > 指定目标 > 默认处理器
func TestActionPriority(t *testing.T) {
	disp := NewDispatcher()

	component := NewMockComponent("test")
	disp.Register(component)

	// 设置全局处理器，返回 true
	globalCalled := false
	disp.Subscribe(ActionQuit, func(a *Action) bool {
		globalCalled = true
		return true // 阻止传递
	})

	// 设置默认处理器
	defaultCalled := false
	disp.SetDefaultHandler(func(a *Action) bool {
		defaultCalled = true
		return true
	})

	// 分发 Action（有目标）
	act := NewAction(ActionQuit).WithTarget("test")
	disp.Dispatch(act)

	// 全局处理器应该被调用
	if !globalCalled {
		t.Errorf("Global handler should be called")
	}

	// 组件不应该被调用（全局处理器返回 true）
	if component.WasHandled() {
		t.Errorf("Component should not be called when global handler returns true")
	}

	// 默认处理器不应该被调用
	if defaultCalled {
		t.Errorf("Default handler should not be called when global handler returns true")
	}
}

// TestActionWithoutTarget 测试没有目标时的 Action 处理
func TestActionWithoutTarget(t *testing.T) {
	disp := NewDispatcher()

	component := NewMockComponent("test")
	disp.Register(component)

	// 设置默认处理器
	defaultCalled := false
	disp.SetDefaultHandler(func(a *Action) bool {
		defaultCalled = true
		return true
	})

	// 分发 Action（没有目标）
	act := NewAction(ActionQuit)
	disp.Dispatch(act)

	// 默认处理器应该被调用
	if !defaultCalled {
		t.Errorf("Default handler should be called for action without target")
	}

	// 组件不应该被调用
	if component.WasHandled() {
		t.Errorf("Component should not be called for action without target")
	}
}

// TestMultipleSubscriptions 测试多个订阅者
func TestMultipleSubscriptions(t *testing.T) {
	disp := NewDispatcher()

	order := make([]int, 0)

	disp.Subscribe(ActionInputText, func(a *Action) bool {
		order = append(order, 1)
		return false // 继续传递
	})

	disp.Subscribe(ActionInputText, func(a *Action) bool {
		order = append(order, 2)
		return false // 继续传递
	})

	component := NewMockComponent("test")
	disp.Register(component)

	act := NewAction(ActionInputText).WithTarget("test").WithPayload("x")
	disp.Dispatch(act)

	// 两个订阅者都应该被调用
	if len(order) != 2 {
		t.Errorf("Expected 2 subscriptions called, got %d", len(order))
	}

	if order[0] != 1 || order[1] != 2 {
		t.Errorf("Subscriptions should be called in order: %v", order)
	}

	// 组件应该被处理
	if !component.WasHandled() {
		t.Errorf("Component should be called")
	}
}

// TestDispatcherLifecycle 测试 Dispatcher 的生命周期
func TestDispatcherLifecycle(t *testing.T) {
	disp := NewDispatcher()

	// 初始状态
	stats := disp.Stats()
	if targets, _ := stats["targets"].(int); targets != 0 {
		t.Errorf("Initial targets should be 0")
	}

	// 注册组件
	comp1 := NewMockComponent("comp1")
	comp2 := NewMockComponent("comp2")

	disp.Register(comp1)
	disp.Register(comp2)

	stats = disp.Stats()
	if targets, _ := stats["targets"].(int); targets != 2 {
		t.Errorf("Should have 2 targets after registration")
	}

	// 注销组件
	disp.Unregister("comp1")

	stats = disp.Stats()
	if targets, _ := stats["targets"].(int); targets != 1 {
		t.Errorf("Should have 1 target after unregister")
	}

	// 订阅/取消订阅
	callCount := 0
	unsub := disp.Subscribe(ActionQuit, func(a *Action) bool {
		callCount++
		return true
	})

	disp.Dispatch(NewAction(ActionQuit))
	if callCount != 1 {
		t.Errorf("Handler should be called once")
	}

	unsub()

	disp.Dispatch(NewAction(ActionQuit))
	if callCount != 1 {
		t.Errorf("Handler should not be called after unsubscribe")
	}
}
