package form

import (
	"testing"

	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/input"
)

// TestFormCharacterInput 测试表单字符输入
func TestFormCharacterInput(t *testing.T) {
	f := NewForm()
	f.SetID("test-form")

	// 创建输入字段
	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")

	usernameField := NewFormField("username")
	usernameField.Label = "用户名"
	usernameField.Input = usernameInput
	f.AddField(usernameField)

	// 表单获得焦点
	f.OnFocus()

	// 验证初始状态
	if !f.IsFocused() {
		t.Error("form should be focused")
	}
	if !usernameInput.IsFocused() {
		t.Error("username input should be focused after form OnFocus")
	}

	// 模拟键盘事件 - 输入 'a'
	keyEv := event.NewKeyEvent(event.Key{Rune: 'a'})
	keyEv.Special = event.KeyUnknown

	handled := f.HandleEvent(keyEv)
	if !handled {
		t.Error("HandleEvent should return true for character input")
	}

	// 验证输入框的值和光标位置
	if usernameInput.GetValue() != "a" {
		t.Errorf("expected value 'a', got '%s'", usernameInput.GetValue())
	}
	if usernameInput.GetCursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", usernameInput.GetCursor())
	}

	// 输入更多字符
	keyEv2 := event.NewKeyEvent(event.Key{Rune: 'b'})
	keyEv2.Special = event.KeyUnknown
	f.HandleEvent(keyEv2)

	keyEv3 := event.NewKeyEvent(event.Key{Rune: 'c'})
	keyEv3.Special = event.KeyUnknown
	f.HandleEvent(keyEv3)

	// 验证最终结果
	if usernameInput.GetValue() != "abc" {
		t.Errorf("expected value 'abc', got '%s'", usernameInput.GetValue())
	}
	if usernameInput.GetCursor() != 3 {
		t.Errorf("expected cursor at 3, got %d", usernameInput.GetCursor())
	}
}

// TestFormBackspace 测试表单中的退格键
func TestFormBackspace(t *testing.T) {
	f := NewForm()
	f.SetID("test-form")

	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")
	usernameInput.SetValue("hello")
	usernameInput.SetCursor(5) // 将光标移到末尾

	usernameField := NewFormField("username")
	usernameField.Label = "用户名"
	usernameField.Input = usernameInput
	f.AddField(usernameField)

	f.OnFocus()

	// 按退格键
	backspaceEv := event.NewSpecialKeyEvent(event.KeyBackspace)
	f.HandleEvent(backspaceEv)

	// 验证值被删除
	if usernameInput.GetValue() != "hell" {
		t.Errorf("expected value 'hell', got '%s'", usernameInput.GetValue())
	}
}

// TestFormMultipleFields 测试多字段导航
func TestFormMultipleFields(t *testing.T) {
	f := NewForm()
	f.SetID("test-form")

	// 创建两个字段
	input1 := input.NewTextInput()
	input1.SetID("field1")
	field1 := NewFormField("field1")
	field1.Label = "字段1"
	field1.Input = input1
	f.AddField(field1)

	input2 := input.NewTextInput()
	input2.SetID("field2")
	field2 := NewFormField("field2")
	field2.Label = "字段2"
	field2.Input = input2
	f.AddField(field2)

	f.OnFocus()

	// 初始焦点应该在第一个字段
	if !input1.IsFocused() {
		t.Error("input1 should be focused initially")
	}
	if input2.IsFocused() {
		t.Error("input2 should not be focused initially")
	}

	// 在第一个字段输入字符
	keyEv := event.NewKeyEvent(event.Key{Rune: 'x'})
	keyEv.Special = event.KeyUnknown
	f.HandleEvent(keyEv)

	if input1.GetValue() != "x" {
		t.Errorf("expected input1 value 'x', got '%s'", input1.GetValue())
	}

	// 导航到下一个字段 (Tab键)
	tabEv := event.NewSpecialKeyEvent(event.KeyTab)
	f.HandleEvent(tabEv)

	// 焦点应该移动到第二个字段
	if input1.IsFocused() {
		t.Error("input1 should not be focused after tab")
	}
	if !input2.IsFocused() {
		t.Error("input2 should be focused after tab")
	}

	// 在第二个字段输入字符
	keyEv2 := event.NewKeyEvent(event.Key{Rune: 'y'})
	keyEv2.Special = event.KeyUnknown
	f.HandleEvent(keyEv2)

	if input2.GetValue() != "y" {
		t.Errorf("expected input2 value 'y', got '%s'", input2.GetValue())
	}
}
