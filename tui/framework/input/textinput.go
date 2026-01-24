package input

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
)

// TextInput 文本输入组件
type TextInput struct {
	*component.BaseComponent

	value      string
	cursor     int
	placeholder string
	password   bool
	echo       rune
	maxLength  int
}

// NewTextInput 创建文本输入框
func NewTextInput() *TextInput {
	return &TextInput{
		BaseComponent: component.NewBaseComponent("input"),
		cursor:        0,
		placeholder:   "",
		password:      false,
		echo:         '*',
		maxLength:     0,
	}
}

// NewTextInputPlaceholder 创建带占位符的输入框
func NewTextInputPlaceholder(placeholder string) *TextInput {
	return &TextInput{
		BaseComponent: component.NewBaseComponent("input"),
		cursor:        0,
		placeholder:   placeholder,
		password:      false,
		echo:         '*',
		maxLength:     0,
	}
}

// SetValue 设置值
func (t *TextInput) SetValue(value string) {
	t.value = value
	if t.cursor > len(value) {
		t.cursor = len(value)
	}
}

// GetValue 获取值
func (t *TextInput) GetValue() string {
	return t.value
}

// SetCursor 设置光标位置
func (t *TextInput) SetCursor(pos int) {
	if pos >= 0 && pos <= len(t.value) {
		t.cursor = pos
	}
}

// GetCursor 获取光标位置
func (t *TextInput) GetCursor() int {
	return t.cursor
}

// SetPlaceholder 设置占位符
func (t *TextInput) SetPlaceholder(text string) {
	t.placeholder = text
}

// SetPassword 设置密码模式
func (t *TextInput) SetPassword(enabled bool) {
	t.password = enabled
}

// SetEcho 设置密码遮罩字符
func (t *TextInput) SetEcho(echo rune) {
	t.echo = echo
}

// SetMaxLength 设置最大长度
func (t *TextInput) SetMaxLength(max int) {
	t.maxLength = max
}

// Clear 清空输入
func (t *TextInput) Clear() {
	t.value = ""
	t.cursor = 0
}

// Render 渲染输入框
func (t *TextInput) Render(ctx *component.RenderContext) string {
	if !t.IsVisible() {
		return ""
	}

	s := t.GetStyle()

	// 确定显示内容
	displayValue := t.value
	if t.value == "" && t.placeholder != "" {
		displayValue = t.placeholder
		s = s.Foreground(style.BrightBlack)
	}

	// 密码模式
	if t.password && t.value != "" {
		displayValue = strings.Repeat(string(t.echo), len(t.value))
	}

	// 限制长度
	if t.maxLength > 0 && len(displayValue) > t.maxLength {
		displayValue = displayValue[:t.maxLength]
	}

	width := ctx.AvailableWidth
	if width <= 0 {
		return ""
	}

	// 渲染
	result := "["
	if width > 2 {
		contentWidth := width - 2
		if len(displayValue) > contentWidth {
			displayValue = displayValue[:contentWidth]
		}

		// 显示内容
		cursorPos := t.cursor
		if cursorPos > len(displayValue) {
			cursorPos = len(displayValue)
		}

		// 插入光标
		if t.HasFocus() && t.value != "" {
			before := displayValue[:cursorPos]
			after := displayValue[cursorPos:]
			result += s.Apply(before) + "▌" + after + strings.Repeat(" ", contentWidth-len(displayValue)-1)
		} else {
			result += s.Apply(displayValue) + strings.Repeat(" ", contentWidth-len(displayValue))
		}
	}

	result += "]"

	return result
}

// HasFocus 检查是否有焦点
func (t *TextInput) HasFocus() bool {
	return false
}

// HandleEvent 处理事件
func (t *TextInput) HandleEvent(ev component.Event) bool {
	switch e := ev.(type) {
	case *event.KeyEvent:
		return t.handleKey(e)
	}
	return false
}

// handleKey 处理键盘事件
func (t *TextInput) handleKey(ev *event.KeyEvent) bool {
	switch {
	case ev.Key >= 32 && ev.Key <= 126:
		// 可打印字符
		if t.maxLength > 0 && len(t.value) >= t.maxLength {
			return true
		}
		t.value = t.value[:t.cursor] + string(ev.Key) + t.value[t.cursor:]
		t.cursor++
		return true

	case ev.Special == event.KeyBackspace:
		if t.cursor > 0 {
			t.value = t.value[:t.cursor-1] + t.value[t.cursor:]
			t.cursor--
		}
		return true

	case ev.Special == event.KeyDelete:
		if t.cursor < len(t.value) {
			t.value = t.value[:t.cursor] + t.value[t.cursor+1:]
		}
		return true

	case ev.Special == event.KeyLeft:
		if t.cursor > 0 {
			t.cursor--
		}
		return true

	case ev.Special == event.KeyRight:
		if t.cursor < len(t.value) {
			t.cursor++
		}
		return true

	case ev.Special == event.KeyHome:
		t.cursor = 0
		return true

	case ev.Special == event.KeyEnd:
		t.cursor = len(t.value)
		return true
	}

	return false
}

// GetPreferredSize 获取首选尺寸
func (t *TextInput) GetPreferredSize() (width, height int) {
	placeholderLen := len(t.placeholder)
	if len(t.value) > placeholderLen {
		placeholderLen = len(t.value)
	}
	if t.maxLength > 0 && placeholderLen > t.maxLength {
		placeholderLen = t.maxLength
	}

	return placeholderLen + 2, 1 // +2 for brackets
}
