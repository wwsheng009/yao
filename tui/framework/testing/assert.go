package testing

import (
	"fmt"
	"reflect"
	stdtesting "testing"

	"github.com/yaoapp/yao/tui/runtime/state"
)

// =============================================================================
// UI Assert (V3)
// =============================================================================
// UIAssert UI 断言工具
// 专门用于测试 TUI 组件状态

// UIAssert UI 断言器
type UIAssert struct {
	t     *stdtesting.T
	snap  *state.Snapshot
	fatal bool
}

// NewAssert 创建断言器
func NewAssert(t *stdtesting.T, snapshot *state.Snapshot) *UIAssert {
	return &UIAssert{
		t:     t,
		snap:  snapshot,
		fatal: true,
	}
}

// =============================================================================
// 组件存在性断言
// =============================================================================

// ComponentExists 断言组件存在
func (a *UIAssert) ComponentExists(id string) *UIAssert {
	a.t.Helper()
	if _, ok := a.snap.GetComponent(id); !ok {
		a.failf("component does not exist: %s", id)
	}
	return a
}

// ComponentNotExists 断言组件不存在
func (a *UIAssert) ComponentNotExists(id string) *UIAssert {
	a.t.Helper()
	if _, ok := a.snap.GetComponent(id); ok {
		a.failf("component exists but should not: %s", id)
	}
	return a
}

// ComponentVisible 断言组件可见
func (a *UIAssert) ComponentVisible(id string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}
	if !comp.Visible {
		a.failf("component is not visible: %s", id)
	}
	return a
}

// ComponentNotVisible 断言组件不可见
func (a *UIAssert) ComponentNotVisible(id string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		// 不存在的组件也算不可见，这可能是期望的行为
		return a
	}
	if comp.Visible {
		a.failf("component is visible but should not be: %s", id)
	}
	return a
}

// ComponentDisabled 断言组件已禁用
func (a *UIAssert) ComponentDisabled(id string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}
	if !comp.Disabled {
		a.failf("component is not disabled: %s", id)
	}
	return a
}

// ComponentNotDisabled 断言组件未禁用
func (a *UIAssert) ComponentNotDisabled(id string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}
	if comp.Disabled {
		a.failf("component is disabled: %s", id)
	}
	return a
}

// ComponentType 断言组件类型
func (a *UIAssert) ComponentType(id, expectedType string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}
	if comp.Type != expectedType {
		a.failf("component %s has type %s, expected %s", id, comp.Type, expectedType)
	}
	return a
}

// =============================================================================
// 状态断言
// =============================================================================

// StateEq 断言状态值相等
func (a *UIAssert) StateEq(id, key string, expected interface{}) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	value, ok := comp.State[key]
	if !ok {
		a.failf("component %s has no state key: %s", id, key)
		return a
	}

	if !reflect.DeepEqual(value, expected) {
		a.failf("component %s state %s = %v, expected %v", id, key, value, expected)
	}
	return a
}

// StateNotEq 断言状态值不相等
func (a *UIAssert) StateNotEq(id, key string, notExpected interface{}) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	value, ok := comp.State[key]
	if !ok {
		// 键不存在也算不相等
		return a
	}

	if reflect.DeepEqual(value, notExpected) {
		a.failf("component %s state %s = %v, should not equal %v", id, key, value, notExpected)
	}
	return a
}

// StateHasKey 断言状态键存在
func (a *UIAssert) StateHasKey(id, key string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	if _, ok := comp.State[key]; !ok {
		a.failf("component %s has no state key: %s", id, key)
	}
	return a
}

// StateNotHasKey 断言状态键不存在
func (a *UIAssert) StateNotHasKey(id, key string) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	if _, ok := comp.State[key]; ok {
		a.failf("component %s has state key: %s", id, key)
	}
	return a
}

// StateLen 断言状态值的长度
func (a *UIAssert) StateLen(id, key string, expectedLen int) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	value, ok := comp.State[key]
	if !ok {
		a.failf("component %s has no state key: %s", id, key)
		return a
	}

	length := 0
	switch v := value.(type) {
	case []string:
		length = len(v)
	case []interface{}:
		length = len(v)
	case []int:
		length = len(v)
	case map[string]interface{}:
		length = len(v)
	case string:
		length = len(v)
	default:
		a.failf("component %s state %s is not a supported type for length check: %T", id, key, value)
		return a
	}

	if length != expectedLen {
		a.failf("component %s state %s has length %d, expected %d", id, key, length, expectedLen)
	}
	return a
}

// StateContains 断言状态包含值
func (a *UIAssert) StateContains(id, key string, expected interface{}) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	value, ok := comp.State[key]
	if !ok {
		a.failf("component %s has no state key: %s", id, key)
		return a
	}

	contains := false
	switch v := value.(type) {
	case []string:
		for _, item := range v {
			if item == fmt.Sprint(expected) {
				contains = true
				break
			}
		}
	case []interface{}:
		for _, item := range v {
			if item == expected {
				contains = true
				break
			}
		}
	case string:
		contains = containsString(v, fmt.Sprint(expected))
	default:
		a.failf("component %s state %s is not a supported type for contains check: %T", id, key, value)
		return a
	}

	if !contains {
		a.failf("component %s state %s = %v does not contain %v", id, key, value, expected)
	}
	return a
}

// =============================================================================
// 布局断言
// =============================================================================

// PositionEq 断言组件位置相等
func (a *UIAssert) PositionEq(id string, x, y int) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	if comp.Rect.X != x || comp.Rect.Y != y {
		a.failf("component %s position is (%d, %d), expected (%d, %d)", id, comp.Rect.X, comp.Rect.Y, x, y)
	}
	return a
}

// SizeEq 断言组件尺寸相等
func (a *UIAssert) SizeEq(id string, width, height int) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	if comp.Rect.Width != width || comp.Rect.Height != height {
		a.failf("component %s size is %dx%d, expected %dx%d", id, comp.Rect.Width, comp.Rect.Height, width, height)
	}
	return a
}

// RectEq 断言组件矩形相等
func (a *UIAssert) RectEq(id string, x, y, width, height int) *UIAssert {
	a.t.Helper()
	comp, ok := a.snap.GetComponent(id)
	if !ok {
		a.failf("component does not exist: %s", id)
		return a
	}

	if comp.Rect.X != x || comp.Rect.Y != y || comp.Rect.Width != width || comp.Rect.Height != height {
		a.failf("component %s rect is (%d, %d, %d, %d), expected (%d, %d, %d, %d)",
			id, comp.Rect.X, comp.Rect.Y, comp.Rect.Width, comp.Rect.Height, x, y, width, height)
	}
	return a
}

// =============================================================================
// 焦点断言
// =============================================================================

// FocusEq 断言焦点路径相等
func (a *UIAssert) FocusEq(expectedPath []string) *UIAssert {
	a.t.Helper()
	actualPath := a.snap.FocusPath
	if !actualPath.Equals(expectedPath) {
		a.failf("focus path is %v, expected %v", actualPath, expectedPath)
	}
	return a
}

// FocusOn 断言焦点在指定组件上
func (a *UIAssert) FocusOn(componentID string) *UIAssert {
	a.t.Helper()
	if a.snap.FocusPath.Current() != componentID {
		a.failf("focus is on %s, expected %s", a.snap.FocusPath.Current(), componentID)
	}
	return a
}

// =============================================================================
// 通用断言
// =============================================================================

// True 通用真值断言
func (a *UIAssert) True(condition bool, msg ...string) *UIAssert {
	a.t.Helper()
	if !condition {
		if len(msg) > 0 {
			a.fail(msg[0])
		} else {
			a.fail("condition is not true")
		}
	}
	return a
}

// False 通用假值断言
func (a *UIAssert) False(condition bool, msg ...string) *UIAssert {
	a.t.Helper()
	if condition {
		if len(msg) > 0 {
			a.fail(msg[0])
		} else {
			a.fail("condition is not false")
		}
	}
	return a
}

// Nil 通用 nil 断言
func (a *UIAssert) Nil(value interface{}, msg ...string) *UIAssert {
	a.t.Helper()
	if !isNil(value) {
		if len(msg) > 0 {
			a.fail(msg[0])
		} else {
			a.fail("value is not nil")
		}
	}
	return a
}

// NotNil 通用非 nil 断言
func (a *UIAssert) NotNil(value interface{}, msg ...string) *UIAssert {
	a.t.Helper()
	if isNil(value) {
		if len(msg) > 0 {
			a.fail(msg[0])
		} else {
			a.fail("value is nil")
		}
	}
	return a
}

// Eq 通用相等断言
func (a *UIAssert) Eq(expected, actual interface{}, msg ...string) *UIAssert {
	a.t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		if len(msg) > 0 {
			a.fail(fmt.Sprintf("%s: expected %v, got %v", msg[0], expected, actual))
		} else {
			a.failf("expected %v, got %v", expected, actual)
		}
	}
	return a
}

// =============================================================================
// 非致命断言
// =============================================================================

// NonFatal 设置为非致命模式（断言失败不停止测试）
func (a *UIAssert) NonFatal() *UIAssert {
	a.fatal = false
	return a
}

// Fatal 设置为致命模式（断言失败停止测试）
func (a *UIAssert) Fatal() *UIAssert {
	a.fatal = true
	return a
}

// =============================================================================
// 辅助方法
// =============================================================================

// fail 失败处理
func (a *UIAssert) fail(msg string) {
	if a.fatal {
		a.t.Fatalf("%s", msg)
	} else {
		a.t.Log(msg)
	}
}

// failf 失败处理（带格式化）
func (a *UIAssert) failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	a.fail(msg)
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexOfString(s, substr) >= 0)
}

// indexOfString 查找子串位置
func indexOfString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
