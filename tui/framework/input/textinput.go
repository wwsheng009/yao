package input

import (
	"strings"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// TextInput Component V3
// ==============================================================================
// V3 文本输入组件，使用 CellBuffer 绘制，处理语义化 Action

// TextInput V3 文本输入组件
type TextInput struct {
	*component.BaseComponent
	*component.StateHolder

	mu    sync.RWMutex
	value       string
	cursor      int
	placeholder string
	password    bool
	echo        rune
	maxLength   int
	normalStyle style.Style
	focusStyle  style.Style
	placeholderStyle style.Style

	// 光标闪烁
	cursorVisible bool
	lastBlinkTime time.Time
}

// NewTextInput 创建 V3 文本输入组件
func NewTextInput() *TextInput {
	now := time.Now()
	return &TextInput{
		BaseComponent:  component.NewBaseComponent("input"),
		StateHolder:      component.NewStateHolder(),
		value:            "",
		cursor:           0,
		placeholder:      "",
		password:         false,
		echo:             '*',
		maxLength:        0,
		normalStyle:      style.Style{},
		focusStyle:       style.Style{}.Foreground(style.Cyan),
		placeholderStyle: style.Style{}.Foreground(style.BrightBlack),
		cursorVisible:    true,
		lastBlinkTime:    now,
	}
}

// NewTextInputPlaceholder 创建带占位符的 V3 输入框
func NewTextInputPlaceholder(placeholder string) *TextInput {
	now := time.Now()
	return &TextInput{
		BaseComponent:  component.NewBaseComponent("input"),
		StateHolder:      component.NewStateHolder(),
		value:            "",
		cursor:           0,
		placeholder:      placeholder,
		password:         false,
		echo:             '*',
		maxLength:        0,
		normalStyle:      style.Style{},
		focusStyle:       style.Style{}.Foreground(style.Cyan),
		placeholderStyle: style.Style{}.Foreground(style.BrightBlack),
		cursorVisible:    true,
		lastBlinkTime:    now,
	}
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetValue 设置值
func (t *TextInput) SetValue(value string) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.value = value
	if t.cursor > len([]rune(value)) {
		t.cursor = len([]rune(value))
	}
	return t
}

// GetValue 获取值
func (t *TextInput) GetValue() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.value
}

// SetCursor 设置光标位置
func (t *TextInput) SetCursor(pos int) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	if pos >= 0 && pos <= len([]rune(t.value)) {
		t.cursor = pos
	}
	return t
}

// GetCursor 获取光标位置
func (t *TextInput) GetCursor() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cursor
}

// SetPlaceholder 设置占位符
func (t *TextInput) SetPlaceholder(text string) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.placeholder = text
	return t
}

// SetPassword 设置密码模式
func (t *TextInput) SetPassword(enabled bool) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.password = enabled
	return t
}

// SetEcho 设置密码遮罩字符
func (t *TextInput) SetEcho(echo rune) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.echo = echo
	return t
}

// SetMaxLength 设置最大长度
func (t *TextInput) SetMaxLength(max int) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxLength = max
	return t
}

// Clear 清空输入
func (t *TextInput) Clear() *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.value = ""
	t.cursor = 0
	return t
}

// SetNormalStyle 设置普通状态样式
func (t *TextInput) SetNormalStyle(s style.Style) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.normalStyle = s
	return t
}

// SetFocusStyle 设置焦点状态样式
func (t *TextInput) SetFocusStyle(s style.Style) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.focusStyle = s
	return t
}

// SetPlaceholderStyle 设置占位符样式
func (t *TextInput) SetPlaceholderStyle(s style.Style) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.placeholderStyle = s
	return t
}

// UpdateCursorBlink 更新光标闪烁状态，返回是否需要重新渲染
func (t *TextInput) UpdateCursorBlink() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.IsFocused() {
		return false
	}

	now := time.Now()
	if now.Sub(t.lastBlinkTime) >= 500*time.Millisecond {
		t.cursorVisible = !t.cursorVisible  // 切换可见性
		t.lastBlinkTime = now
		return true  // 状态改变，需要重绘
	}
	return false
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (t *TextInput) Measure(maxWidth, maxHeight int) (width, height int) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	placeholderWidth := len([]rune(t.placeholder))
	valueWidth := len([]rune(t.value))

	contentWidth := placeholderWidth
	if valueWidth > contentWidth {
		contentWidth = valueWidth
	}

	if t.maxLength > 0 && contentWidth > t.maxLength {
		contentWidth = t.maxLength
	}

	// 加上边框 (左右各1个字符)
	width = contentWidth + 2
	height = 1

	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}

	return width, height
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件到 CellBuffer
func (t *TextInput) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !t.IsVisible() {
		return
	}

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	// 确定显示内容
	displayValue := t.value
	drawStyle := t.normalStyle

	if t.IsFocused() {
		drawStyle = t.focusStyle
	}

	if t.value == "" && t.placeholder != "" {
		displayValue = t.placeholder
		drawStyle = t.placeholderStyle
	}

	// 密码模式
	if t.password && t.value != "" {
		displayValue = strings.Repeat(string(t.echo), len([]rune(t.value)))
	}

	// 限制长度
	contentWidth := width - 2 // 减去左右边框
	if contentWidth < 1 {
		contentWidth = 1 // 至少显示1个字符
	}
	if t.maxLength > 0 && contentWidth > t.maxLength {
		contentWidth = t.maxLength
	}

	runes := []rune(displayValue)
	if len(runes) > contentWidth {
		runes = runes[:contentWidth]
		displayValue = string(runes)
	}

	// 计算垂直位置
	y := ctx.Y
	if height > 1 {
		y += (height - 1) / 2
	}

	x := ctx.X

	// 绘制左边框
	buf.SetCell(x, y, '[', drawStyle)
	x++

	// 绘制内容
	for i := 0; i < contentWidth; i++ {
		if i < len(runes) {
			// 检查是否在光标位置
			if t.IsFocused() && i == t.cursor && t.cursorVisible {
				// 绘制光标（闪烁时高亮显示）
				buf.SetCell(x+i, y, runes[i], drawStyle.Reverse(true))
			} else {
				buf.SetCell(x+i, y, runes[i], drawStyle)
			}
		} else if t.IsFocused() && i == t.cursor && t.cursorVisible {
			// 光标在内容末尾
			buf.SetCell(x+i, y, ' ', drawStyle.Reverse(true))
		} else {
			buf.SetCell(x+i, y, ' ', drawStyle)
		}
	}

	x += contentWidth

	// 绘制右边框
	buf.SetCell(x, y, ']', drawStyle)
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction 处理语义化 Action
func (t *TextInput) HandleAction(a action.Action) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch a.Type {
	case action.ActionInputChar:
		// 输入字符
		char, ok := a.Payload.(rune)
		if !ok {
			return false
		}
		return t.handleInputChar(char)

	case action.ActionInputText:
		// 输入文本
		text, ok := a.Payload.(string)
		if !ok {
			return false
		}
		return t.handleInputText(text)

	case action.ActionBackspace:
		return t.handleBackspace()

	case action.ActionDeleteChar:
		return t.handleDelete()

	case action.ActionDeleteWord:
		return t.handleDeleteWord()

	case action.ActionDeleteLine:
		return t.handleDeleteLine()

	case action.ActionCursorLeft:
		return t.handleCursorLeft()

	case action.ActionCursorRight:
		return t.handleCursorRight()

	case action.ActionCursorHome:
		return t.handleCursorHome()

	case action.ActionCursorEnd:
		return t.handleCursorEnd()

	case action.ActionCursorWordLeft:
		return t.handleCursorWordLeft()

	case action.ActionCursorWordRight:
		return t.handleCursorWordRight()

	case action.ActionClear:
		t.value = ""
		t.cursor = 0
		return true
	}

	return false
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (t *TextInput) FocusID() string {
	return t.ID()
}

// OnFocus 获得焦点时调用
func (t *TextInput) OnFocus() {
	t.mu.Lock()
	t.cursorVisible = true
	t.lastBlinkTime = time.Now()
	t.mu.Unlock()

	// 调用 BaseComponent 的 OnFocus (设置 focused = true)
	t.BaseComponent.OnFocus()

	// 注册到全局光标管理器
	RegisterCursor(t)
}

// OnBlur 失去焦点时调用
func (t *TextInput) OnBlur() {
	t.mu.Lock()
	t.cursorVisible = false
	t.mu.Unlock()

	// 调用 BaseComponent 的 OnBlur (设置 focused = false)
	t.BaseComponent.OnBlur()

	// 从全局光标管理器注销
	UnregisterCursor(t)
}

// ============================================================================
// V2 Component 接口兼容
// ============================================================================

// HandleEvent 处理事件 (Component接口)
func (t *TextInput) HandleEvent(ev component.Event) bool {
	if keyEv, ok := ev.(*event.KeyEvent); ok {
		// 处理退格键
		if keyEv.Special == event.KeyBackspace {
			return t.HandleAction(*action.NewAction(action.ActionBackspace))
		}

		// 处理删除键
		if keyEv.Special == event.KeyDelete {
			return t.HandleAction(*action.NewAction(action.ActionDeleteChar))
		}

		// 处理光标键
		if keyEv.Special == event.KeyLeft {
			return t.HandleAction(*action.NewAction(action.ActionCursorLeft))
		}
		if keyEv.Special == event.KeyRight {
			return t.HandleAction(*action.NewAction(action.ActionCursorRight))
		}
		if keyEv.Special == event.KeyHome {
			return t.HandleAction(*action.NewAction(action.ActionCursorHome))
		}
		if keyEv.Special == event.KeyEnd {
			return t.HandleAction(*action.NewAction(action.ActionCursorEnd))
		}

		// 处理普通字符输入
		if keyEv.Key != 0 && keyEv.Special == event.KeyUnknown {
			a := action.NewAction(action.ActionInputChar).WithPayload(keyEv.Key)
			return t.HandleAction(*a)
		}
	}
	return false
}

// ============================================================================
// Validatable 接口实现
// ============================================================================

// Validate 验证组件状态
func (t *TextInput) Validate() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 检查必填
	if required, ok := t.GetProp("required"); ok && required.(bool) {
		if t.value == "" {
			return &ValidationError{Field: t.ID(), Message: "此字段为必填项"}
		}
	}

	// 检查最小长度
	minLen := t.GetPropInt("minLength", 0)
	if minLen > 0 {
		if len([]rune(t.value)) < minLen {
			return &ValidationError{Field: t.ID(), Message: "最少需要 " + string(rune(minLen+'0')) + " 个字符"}
		}
	}

	// 检查最大长度
	if t.maxLength > 0 {
		if len([]rune(t.value)) > t.maxLength {
			return &ValidationError{Field: t.ID(), Message: "最多允许 " + string(rune(t.maxLength+'0')) + " 个字符"}
		}
	}

	return nil
}

// IsValid 检查是否有效
func (t *TextInput) IsValid() bool {
	return t.Validate() == nil
}

// ============================================================================
// 内部处理方法
// ============================================================================

// handleInputChar 处理字符输入
func (t *TextInput) handleInputChar(char rune) bool {
	if t.maxLength > 0 && len([]rune(t.value)) >= t.maxLength {
		return true
	}

	runes := []rune(t.value)
	t.value = string(append(runes[:t.cursor], append([]rune{char}, runes[t.cursor:]...)...))
	t.cursor++
	return true
}

// handleInputText 处理文本输入
func (t *TextInput) handleInputText(text string) bool {
	runes := []rune(t.value)
	inputRunes := []rune(text)

	for _, r := range inputRunes {
		if t.maxLength > 0 && len(runes) >= t.maxLength {
			break
		}
		runes = append(runes[:t.cursor], append([]rune{r}, runes[t.cursor:]...)...)
		t.cursor++
	}

	t.value = string(runes)
	return true
}

// handleBackspace 处理退格
func (t *TextInput) handleBackspace() bool {
	if t.cursor > 0 {
		runes := []rune(t.value)
		t.value = string(append(runes[:t.cursor-1], runes[t.cursor:]...))
		t.cursor--
		return true
	}
	return false
}

// handleDelete 处理删除
func (t *TextInput) handleDelete() bool {
	runes := []rune(t.value)
	if t.cursor < len(runes) {
		t.value = string(append(runes[:t.cursor], runes[t.cursor+1:]...))
		return true
	}
	return false
}

// handleDeleteWord 删除单词
func (t *TextInput) handleDeleteWord() bool {
	runes := []rune(t.value)
	if t.cursor < len(runes) {
		// 找到下一个单词的起始位置
		end := t.cursor
		for end < len(runes) && runes[end] == ' ' {
			end++
		}
		for end < len(runes) && runes[end] != ' ' {
			end++
		}
		t.value = string(append(runes[:t.cursor], runes[end:]...))
		return true
	}
	return false
}

// handleDeleteLine 删除整行
func (t *TextInput) handleDeleteLine() bool {
	t.value = ""
	t.cursor = 0
	return true
}

// handleCursorLeft 光标左移
func (t *TextInput) handleCursorLeft() bool {
	if t.cursor > 0 {
		t.cursor--
		return true
	}
	return false
}

// handleCursorRight 光标右移
func (t *TextInput) handleCursorRight() bool {
	runes := []rune(t.value)
	if t.cursor < len(runes) {
		t.cursor++
		return true
	}
	return false
}

// handleCursorHome 光标到行首
func (t *TextInput) handleCursorHome() bool {
	t.cursor = 0
	return true
}

// handleCursorEnd 光标到行尾
func (t *TextInput) handleCursorEnd() bool {
	runes := []rune(t.value)
	t.cursor = len(runes)
	return true
}

// handleCursorWordLeft 光标左移一词
func (t *TextInput) handleCursorWordLeft() bool {
	if t.cursor > 0 {
		// 跳过当前词的空格
		for t.cursor > 0 && t.value[t.cursor-1] == ' ' {
			t.cursor--
		}
		// 跳过单词
		for t.cursor > 0 && t.value[t.cursor-1] != ' ' {
			t.cursor--
		}
		return true
	}
	return false
}

// handleCursorWordRight 光标右移一词
func (t *TextInput) handleCursorWordRight() bool {
	runes := []rune(t.value)
	if t.cursor < len(runes) {
		// 跳过当前词
		for t.cursor < len(runes) && runes[t.cursor] != ' ' {
			t.cursor++
		}
		// 跳过空格
		for t.cursor < len(runes) && runes[t.cursor] == ' ' {
			t.cursor++
		}
		return true
	}
	return false
}

// ============================================================================
// 验证错误
// ============================================================================

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
