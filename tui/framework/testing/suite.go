package testing

import (
	"fmt"
	stdtesting "testing"
	"time"

	"github.com/yaoapp/yao/tui/runtime/ai"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/focus"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// =============================================================================
// TestSuite (V3)
// =============================================================================
// TestSuite 测试套件
// 用于集成测试和端到端测试

// TestSuite 测试套件
type TestSuite struct {
	t             *stdtesting.T
	controller    *ai.RuntimeController
	dispatcher    *action.Dispatcher
	tracker       *state.Tracker
	focusMgr      *focus.Manager
	teardownFuncs []func() // 清理函数列表
}

// NewTestSuite 创建测试套件
func NewTestSuite(t *stdtesting.T) *TestSuite {
	dispatcher := action.NewDispatcher()
	tracker := state.NewTracker()
	focusMgr := focus.NewManager(nil) // root node can be nil for basic tests

	controller := ai.NewRuntimeController(dispatcher, tracker, focusMgr)

	return &TestSuite{
		t:          t,
		controller: controller,
		dispatcher: dispatcher,
		tracker:    tracker,
		focusMgr:   focusMgr,
	}
}

// T 获取 stdtesting.T
func (s *TestSuite) T() *stdtesting.T {
	return s.t
}

// Controller 获取 AI 控制器
func (s *TestSuite) Controller() *ai.RuntimeController {
	return s.controller
}

// Dispatcher 获取 Action 分发器
func (s *TestSuite) Dispatcher() *action.Dispatcher {
	return s.dispatcher
}

// Tracker 获取状态追踪器
func (s *TestSuite) Tracker() *state.Tracker {
	return s.tracker
}

// FocusManager 获取焦点管理器
func (s *TestSuite) FocusManager() *focus.Manager {
	return s.focusMgr
}

// =============================================================================
// 测试辅助方法
// =============================================================================

// Must 不为 nil，否则测试失败
func (s *TestSuite) Must(err error) {
	if err != nil {
		s.t.Fatalf("unexpected error: %v", err)
	}
}

// MustEq 值必须相等，否则测试失败
func (s *TestSuite) MustEq(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		s.t.Helper()
		msg := fmt.Sprintf("expected %v, got %v", expected, actual)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustTrue 条件必须为真，否则测试失败
func (s *TestSuite) MustTrue(condition bool, msgAndArgs ...interface{}) {
	if !condition {
		s.t.Helper()
		msg := "condition is not true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustFalse 条件必须为假，否则测试失败
func (s *TestSuite) MustFalse(condition bool, msgAndArgs ...interface{}) {
	s.MustTrue(!condition, msgAndArgs...)
}

// MustNotNil 值不能为 nil，否则测试失败
func (s *TestSuite) MustNotNil(v interface{}, msgAndArgs ...interface{}) {
	if v == nil {
		s.t.Helper()
		msg := "value is nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}

	// 检查接口是否为 nil
	if isNil(v) {
		s.t.Helper()
		msg := "value is nil (interface)"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustNil 值必须为 nil，否则测试失败
func (s *TestSuite) MustNil(v interface{}, msgAndArgs ...interface{}) {
	if !isNil(v) {
		s.t.Helper()
		msg := fmt.Sprintf("expected nil, got %v", v)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustNoPanic 代码不能 panic，否则测试失败
func (s *TestSuite) MustNoPanic(f func()) {
	defer func() {
		if r := recover(); r != nil {
			s.t.Helper()
			s.t.Fatalf("unexpected panic: %v", r)
		}
	}()
	f()
}

// MustErr 期望有错误，如果没有错误则测试失败
func (s *TestSuite) MustErr(err error, msgAndArgs ...interface{}) {
	if err == nil {
		s.t.Helper()
		msg := "expected error, got nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustNoErr 期望没有错误，如果有错误则测试失败
func (s *TestSuite) MustNoErr(err error, msgAndArgs ...interface{}) {
	if err != nil {
		s.t.Helper()
		msg := fmt.Sprintf("unexpected error: %v", err)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustContain 切片必须包含元素
func (s *TestSuite) MustContain(slice interface{}, elem interface{}, msgAndArgs ...interface{}) {
	s.t.Helper()
	contains := false
	switch sl := slice.(type) {
	case []string:
		for _, v := range sl {
			if v == elem {
				contains = true
				break
			}
		}
	case []int:
		for _, v := range sl {
			if v == elem {
				contains = true
				break
			}
		}
	case []interface{}:
		for _, v := range sl {
			if v == elem {
				contains = true
				break
			}
		}
	default:
		s.t.Fatalf("unsupported slice type: %T", slice)
	}

	if !contains {
		msg := fmt.Sprintf("expected %v to contain %v", slice, elem)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// MustLen 切片/映射长度必须等于预期
func (s *TestSuite) MustLen(v interface{}, expected int, msgAndArgs ...interface{}) {
	s.t.Helper()
	length := 0

	switch val := v.(type) {
	case []string:
		length = len(val)
	case []int:
		length = len(val)
	case []interface{}:
		length = len(val)
	case map[string]interface{}:
		length = len(val)
	case map[string]string:
		length = len(val)
	default:
		s.t.Fatalf("unsupported type for MustLen: %T", v)
	}

	if length != expected {
		msg := fmt.Sprintf("expected length %d, got %d", expected, length)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...) + ": " + msg
		}
		s.t.Fatal(msg)
	}
}

// =============================================================================
// 清理和生命周期
// =============================================================================

// Defer 注册清理函数，在测试结束时执行
func (s *TestSuite) Defer(f func()) {
	s.teardownFuncs = append(s.teardownFuncs, f)
}

// TearDown 执行所有清理函数
func (s *TestSuite) TearDown() {
	// 倒序执行清理函数
	for i := len(s.teardownFuncs) - 1; i >= 0; i-- {
		s.teardownFuncs[i]()
	}
	s.teardownFuncs = nil
}

// =============================================================================
// 辅助函数
// =============================================================================

// isNil 检查值是否为 nil（包括接口的 nil 情况）
func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := [1]interface{}{v}
	switch rv[0].(type) {
	case interface{}:
		return false
	default:
		return rv[0] == nil
	}
}

// =============================================================================
// 快捷操作
// =============================================================================

// Sleep 暂停测试执行
func (s *TestSuite) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Eventually 最终断言（重试直到条件满足）
func (s *TestSuite) Eventually(condition func() bool, timeout time.Duration, interval time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}

		if time.Now().After(deadline) {
			s.t.Fatalf("eventuality condition not met after %v", timeout)
			return
		}

		<-ticker.C
	}
}

// Consistently 一致性断言（条件必须持续满足）
func (s *TestSuite) Consistently(condition func() bool, duration time.Duration, interval time.Duration) {
	deadline := time.Now().Add(duration)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if !condition() {
			s.t.Fatalf("consistency condition failed after %v", duration)
			return
		}

		if time.Now().After(deadline) {
			return
		}

		<-ticker.C
	}
}
