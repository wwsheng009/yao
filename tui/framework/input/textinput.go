package input

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/cursor"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

var (
	debugInput   = os.Getenv("TUI_INPUT_DEBUG") == "1"
	debugFile    *os.File
	debugFileMu  sync.Mutex
	debugStarted bool
)

// initDebugFile 初始化调试日志文件
func initDebugFile() {
	debugFileMu.Lock()
	defer debugFileMu.Unlock()

	if debugStarted {
		return
	}
	debugStarted = true

	if !debugInput {
		return
	}

	filename := os.Getenv("TUI_INPUT_DEBUG_FILE")
	if filename == "" {
		filename = fmt.Sprintf("tui_input_debug_%s.log", time.Now().Format("20060102_150405"))
	}

	var err error
	debugFile, err = os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create debug file: %v\n", err)
		debugInput = false
		return
	}

	fmt.Fprintf(os.Stderr, "[TextInput] Debug logging enabled, writing to: %s\n", filename)
}

// debugLog 调试日志输出到文件
func debugLog(format string, args ...interface{}) {
	
	if !debugInput {
		return
	}

	if debugFile == nil {
		initDebugFile()
		if debugFile == nil {
			return
		}
	}

	timestamp := time.Now().Format("15:04:05.000")
	fullFormat := fmt.Sprintf("[%s] [TextInput] %s\n", timestamp, format)
	msg := fmt.Sprintf(fullFormat, args...)
	
	debugFileMu.Lock()
	defer debugFileMu.Unlock()
	debugFile.WriteString(msg)
}

// ==============================================================================
// TextInput Component V4
// ==============================================================================
// V4 文本输入组件，使用独立的 Cursor 组件处理光标

// TextInput V4 文本输入组件
type TextInput struct {
	*component.BaseComponent
	*component.StateHolder

	mu sync.RWMutex

	// 文本内容
	value    string
	cursor   int // 文本光标位置（字符索引）

	// 配置
	placeholder string
	password    bool
	echo        rune
	maxLength   int

	// 样式
	normalStyle      style.Style
	focusStyle       style.Style
	placeholderStyle style.Style

	// 光标 - 使用独立的 Cursor 组件
	cursorComp *cursor.Cursor
}

// NewTextInput 创建 V4 文本输入组件
func NewTextInput() *TextInput {
	return &TextInput{
		BaseComponent:    component.NewBaseComponent("input"),
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
		cursorComp:       cursor.NewCursor(),
	}
}

// NewTextInputPlaceholder 创建带占位符的 V4 输入框
func NewTextInputPlaceholder(placeholder string) *TextInput {
	return &TextInput{
		BaseComponent:    component.NewBaseComponent("input"),
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
		cursorComp:       cursor.NewCursor(),
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

	// 只在光标超出范围时调整（保持在原来的位置，除非超出边界）
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

// GetCursor 获取光标组件
// 实现 cursor.Host 接口
func (t *TextInput) GetCursorComp() *cursor.Cursor {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cursorComp
}

// SetCursorStyle 设置光标样式
func (t *TextInput) SetCursorStyle(s style.Style) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorComp.SetStyle(s)
	return t
}

// SetCursorShape 设置光标形状
func (t *TextInput) SetCursorShape(shape cursor.Shape) *TextInput {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorComp.SetShape(shape)
	return t
}

// MountWithContext 使用组件上下文挂载
// 实现 component.MountableWithContext 接口，避免 App 直接依赖 TextInput 类型
func (t *TextInput) MountWithContext(parent component.Container, ctx *component.ComponentContext) {
	// 调用基础挂载
	t.Mount(parent)

	// 从上下文获取 dirty callback（如果有）
	if fn := ctx.GetDirtyCallback(); fn != nil {
		t.SetDirtyCallback(fn)
	}
}

// SetDirtyCallback 设置脏标记回调
func (t *TextInput) SetDirtyCallback(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// 通过 StateHolder 的 MarkDirty 来调用此回调
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

	// 读取状态
	t.mu.RLock()
	isFocused := t.IsFocused()
	value := t.value
	cursorPos := t.cursor
	placeholder := t.placeholder
	password := t.password
	echo := t.echo
	maxLength := t.maxLength
	normalStyle := t.normalStyle
	focusStyle := t.focusStyle
	placeholderStyle := t.placeholderStyle
	t.mu.RUnlock()

	// 记录每次Paint调用（不论是否有焦点）
	debugLog("[%s] PAINT: ctx=(%d,%d), value='%s', cursor=%d, focused=%v",
		t.ID(), ctx.X, ctx.Y, value, cursorPos, isFocused)

	// 确定显示内容和样式
	displayValue := value
	drawStyle := normalStyle

	if isFocused {
		drawStyle = focusStyle
	}

	if value == "" && placeholder != "" {
		displayValue = placeholder
		drawStyle = placeholderStyle
	}

	// 密码模式
	if password && value != "" {
		displayValue = strings.Repeat(string(echo), len([]rune(value)))
	}

	// 限制长度
	contentWidth := width - 2 // 减去左右边框
	if contentWidth < 1 {
		contentWidth = 1 // 至少显示1个字符
	}
	if maxLength > 0 && contentWidth > maxLength {
		contentWidth = maxLength
	}

	runes := []rune(displayValue)
	if len(runes) > contentWidth {
		runes = runes[:contentWidth]
		displayValue = string(runes)
	}

	// 计算垂直位置
	// TextInput 是单行组件，直接使用 ctx.Y，不做垂直居中
	y := ctx.Y

	x := ctx.X

	// 绘制左边框
	buf.SetCell(x, y, '[', drawStyle)
	x++

	// 绘制内容（只绘制实际文字，不填充额外空格）
	for i, r := range runes {
		buf.SetCell(x+i, y, r, drawStyle)
	}
	x += len(runes)

	// 绘制右边框
	buf.SetCell(x, y, ']', drawStyle)

	// 绘制光标（使用独立的 Cursor 组件）
	if isFocused {
		// 计算光标位置
		// cursorPos 表示光标在文本中的索引（0-based）
		// 对于块状光标，我们高亮当前光标位置的字符
		// 如果 cursorPos >= len(value): 光标在末尾，高亮右括号或最后一个字符

		var cursorX int
		if len(runes) == 0 {
			// 空输入，光标在左边框后的位置
			cursorX = 1
		} else if cursorPos >= len(runes) {
			// 光标在末尾或超出，高亮最后一个字符
			cursorX = 1 + len(runes) - 1
		} else {
			// 光标在某个字符上，高亮该字符
			cursorX = 1 + cursorPos
		}

		cursorY := y - ctx.Y // 转换为相对 Y 坐标

		// 确保光标在边界内
		if cursorX < 1 {
			cursorX = 1
		}
		maxCursorX := 1 + len(runes) // 右括号的位置
		if cursorX > maxCursorX {
			cursorX = maxCursorX
		}

		// 计算绝对光标位置（用于调试）
		absCursorX := ctx.X + cursorX
		absCursorY := ctx.Y + cursorY

		debugLog("[%s] FOCUS CURSOR: logical=%d, relative=(%d,%d), absolute=(%d,%d), lenRunes=%d",
			t.ID(), cursorPos, cursorX, cursorY, absCursorX, absCursorY, len(runes))
		
		// 使用 Cursor 组件绘制光标
		// Cursor.Paint 会加上 ctx.X 和 ctx.Y 来得到绝对位置
		t.cursorComp.SetPosition(cursorX, cursorY)
		//print the x and the y
		
		t.cursorComp.Paint(ctx, buf)

		// 验证光标是否正确绘制到缓冲区
		if absCursorX < buf.Width && absCursorY < buf.Height {
			cell := buf.Cells[absCursorY][absCursorX]
			debugLog("[%s] FOCUS RESULT: (%d,%d)='%c' reverse=%v",
				t.ID(), absCursorX, absCursorY, cell.Char, cell.Style.IsReverse())
		}
	}

	// 调试：可视化整行的反转状态
	if debugInput {
		t.visualizeReverseState(ctx, buf)
	}
}

// visualizeReverseState 可视化整行的反转状态，用于调试
// 输出格式如: [___XXX____] 表示位置3-5有反转样式
func (t *TextInput) visualizeReverseState(ctx component.PaintContext, buf *paint.Buffer) {
	y := ctx.Y
	startX := ctx.X

	// 获取输入框的宽度（左边框 + 内容 + 右边框）
	t.mu.RLock()
	valueLen := len([]rune(t.value))
	t.mu.RUnlock()

	width := valueLen + 2 // 左边框 + 内容 + 右边框
	if width > 40 {
		width = 40 // 限制显示宽度
	}

	var reverseState strings.Builder
	reverseState.WriteString("[ ")

	for i := 0; i < width; i++ {
		absX := startX + i
		if absX >= buf.Width || y >= buf.Height {
			break
		}
		cell := buf.Cells[y][absX]
		if cell.Style.IsReverse() {
			reverseState.WriteString("X") // 反转样式
		} else if cell.Char == 0 {
			reverseState.WriteString("_") // 空字符
		} else {
			reverseState.WriteString(".") // 正常字符
		}
	}

	reverseState.WriteString(" ]")

	debugLog("[%s] REVERSE_STATE: %s (显示范围: x=%d, width=%d)",
		t.ID(), reverseState.String(), startX, width)

	// 同时显示对应的字符内容
	var contentState strings.Builder
	contentState.WriteString("[ ")

	for i := 0; i < width; i++ {
		absX := startX + i
		if absX >= buf.Width || y >= buf.Height {
			break
		}
		cell := buf.Cells[y][absX]
		if cell.Char == 0 {
			contentState.WriteString("_")
		} else {
			contentState.WriteRune(cell.Char)
		}
	}

	contentState.WriteString(" ]")
	debugLog("[%s] CONTENT_STATE: %s", t.ID(), contentState.String())
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
	debugLog("[%s] OnFocus: current cursor=%d, value='%s'",
		t.ID(), t.cursor, t.value)

	// 调用 BaseComponent 的 OnFocus (设置 focused = true)
	t.BaseComponent.OnFocus()

	// 启用光标闪烁
	t.cursorComp.SetBlinkEnabled(true)
	t.cursorComp.ResetBlink()
}

// OnBlur 失去焦点时调用
func (t *TextInput) OnBlur() {
	debugLog("[%s] OnBlur: current cursor=%d, value='%s'",
		t.ID(), t.cursor, t.value)

	// 禁用光标闪烁
	t.cursorComp.SetBlinkEnabled(false)

	// 调用 BaseComponent 的 OnBlur (设置 focused = false)
	t.BaseComponent.OnBlur()
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
			return &ValidationError{Field: t.ID(), Message: "最少需要 " + string(rune('0'+minLen)) + " 个字符"}
		}
	}

	// 检查最大长度
	if t.maxLength > 0 {
		if len([]rune(t.value)) > t.maxLength {
			return &ValidationError{Field: t.ID(), Message: "最多允许 " + string(rune('0'+t.maxLength)) + " 个字符"}
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

	// 重置光标闪烁状态，确保输入后光标立即可见
	// 这解决了用户输入时光标可能刚好处于"不可见"状态的问题
	t.cursorComp.ResetBlink()

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

	// 重置光标闪烁状态，确保输入后光标立即可见
	t.cursorComp.ResetBlink()

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
