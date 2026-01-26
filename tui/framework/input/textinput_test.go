package input

import (
	"testing"

	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/runtime/action"
)

// TestTextInputBasicInsertion 测试基本字符输入
func TestTextInputBasicInsertion(t *testing.T) {
	input := NewTextInput()
	input.SetID("test-input")

	// 初始状态
	if input.GetValue() != "" {
		t.Errorf("expected empty value, got '%s'", input.GetValue())
	}
	if input.GetCursor() != 0 {
		t.Errorf("expected cursor at 0, got %d", input.GetCursor())
	}

	// 输入第一个字符
	act := action.NewAction(action.ActionInputChar).WithPayload('a')
	handled := input.HandleAction(*act)
	if !handled {
		t.Error("HandleAction should return true for ActionInputChar")
	}

	// 验证状态
	if input.GetValue() != "a" {
		t.Errorf("expected value 'a', got '%s'", input.GetValue())
	}
	if input.GetCursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", input.GetCursor())
	}

	// 输入第二个字符
	act2 := action.NewAction(action.ActionInputChar).WithPayload('b')
	input.HandleAction(*act2)

	// 验证状态
	if input.GetValue() != "ab" {
		t.Errorf("expected value 'ab', got '%s'", input.GetValue())
	}
	if input.GetCursor() != 2 {
		t.Errorf("expected cursor at 2, got %d", input.GetCursor())
	}
}

// TestTextInputHandleEvent 测试 HandleEvent 方法
func TestTextInputHandleEvent(t *testing.T) {
	input := NewTextInput()
	input.SetID("test-input")

	// 使用工厂函数创建 KeyEvent
	keyEv := event.NewKeyEvent('x')
	keyEv.Special = event.KeyUnknown // 确保不是特殊键

	handled := input.HandleEvent(keyEv)
	if !handled {
		t.Error("HandleEvent should return true for character input")
	}

	if input.GetValue() != "x" {
		t.Errorf("expected value 'x', got '%s'", input.GetValue())
	}
	if input.GetCursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", input.GetCursor())
	}
}

// TestTextInputWithFocus 测试焦点状态
func TestTextInputWithFocus(t *testing.T) {
	input := NewTextInput()
	input.SetID("test-input")

	// 初始状态：无焦点
	if input.IsFocused() {
		t.Error("should not be focused initially")
	}

	// 获得焦点
	input.OnFocus()
	if !input.IsFocused() {
		t.Error("should be focused after OnFocus")
	}

	// 输入字符
	act := action.NewAction(action.ActionInputChar).WithPayload('t')
	input.HandleAction(*act)

	if input.GetValue() != "t" {
		t.Errorf("expected value 't', got '%s'", input.GetValue())
	}
	if input.GetCursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", input.GetCursor())
	}

	// 失去焦点
	input.OnBlur()
	if input.IsFocused() {
		t.Error("should not be focused after OnBlur")
	}
}

// TestTextInputBackspace 测试退格键
func TestTextInputBackspace(t *testing.T) {
	input := NewTextInput()
	input.SetValue("hello")

	// 光标在末尾
	input.SetCursor(5)

	// 按退格键
	act := action.NewAction(action.ActionBackspace)
	input.HandleAction(*act)

	if input.GetValue() != "hell" {
		t.Errorf("expected value 'hell', got '%s'", input.GetValue())
	}
	if input.GetCursor() != 4 {
		t.Errorf("expected cursor at 4, got %d", input.GetCursor())
	}
}

// TestTextInputCursorMovement 测试光标移动
func TestTextInputCursorMovement(t *testing.T) {
	input := NewTextInput()
	input.SetValue("abc")

	// 测试右移
	act1 := action.NewAction(action.ActionCursorRight)
	input.HandleAction(*act1)
	if input.GetCursor() != 1 {
		t.Errorf("expected cursor at 1 after right, got %d", input.GetCursor())
	}

	// 测试左移
	act2 := action.NewAction(action.ActionCursorLeft)
	input.HandleAction(*act2)
	if input.GetCursor() != 0 {
		t.Errorf("expected cursor at 0 after left, got %d", input.GetCursor())
	}

	// 测试移动到行尾
	act3 := action.NewAction(action.ActionCursorEnd)
	input.HandleAction(*act3)
	if input.GetCursor() != 3 {
		t.Errorf("expected cursor at 3 after End, got %d", input.GetCursor())
	}
}

// TestTextInputInsertInMiddle 测试在中间插入字符
func TestTextInputInsertInMiddle(t *testing.T) {
	input := NewTextInput()
	input.SetValue("ac")

	// 将光标移到中间
	input.SetCursor(1)

	// 在中间插入 'b'
	act := action.NewAction(action.ActionInputChar).WithPayload('b')
	input.HandleAction(*act)

	if input.GetValue() != "abc" {
		t.Errorf("expected value 'abc', got '%s'", input.GetValue())
	}
	if input.GetCursor() != 2 {
		t.Errorf("expected cursor at 2, got %d", input.GetCursor())
	}
}
