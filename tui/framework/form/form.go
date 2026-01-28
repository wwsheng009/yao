package form

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

var (
	debugForm = os.Getenv("TUI_FORM_DEBUG") == "1"
)

func formDebugLog(format string, args ...interface{}) {
	if debugForm {
		timestamp := time.Now().Format("15:04:05.000")
		fullFormat := fmt.Sprintf("[%s] [Form] %s\n", timestamp, format)
		fmt.Fprintf(os.Stderr, fullFormat, args...)
	}
}

// ==============================================================================
// Form Component V3
// ==============================================================================
// V3 表单组件，支持验证、字段导航、提交/取消

// FormField 表单字段
type FormField struct {
	// 基本信息
	Name        string
	Label       string
	Placeholder string
	HelpText    string

	// 输入组件
	Input component.Node

	// 验证器
	Validators []validation.Validator

	// 状态
	Error   error
	Touched  bool
	Visible  bool
	Disabled bool

	mu sync.RWMutex
}

// NewFormField 创建表单字段
func NewFormField(name string) *FormField {
	return &FormField{
		Name:        name,
		Visible:     true,
		Disabled:    false,
		Validators:  make([]validation.Validator, 0),
	}
}

// SetValue 设置值
func (f *FormField) SetValue(value interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 根据输入组件类型设置值
	switch input := f.Input.(type) {
	case *input.TextInput:
		if str, ok := value.(string); ok {
			input.SetValue(str)
		}
	}

	f.Touched = true
	return nil
}

// GetValue 获取值
func (f *FormField) GetValue() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	switch input := f.Input.(type) {
	case *input.TextInput:
		return input.GetValue()
	default:
		return nil
	}
}

// Validate 验证字段
func (f *FormField) Validate() error {
	// 先获取值（使用读锁）
	f.mu.RLock()
	value := f.GetValue()
	validators := f.Validators
	f.mu.RUnlock()

	// 然后验证
	for _, validator := range validators {
		if err := validator.Validate(value); err != nil {
			f.mu.Lock()
			f.Error = err
			f.mu.Unlock()
			return err
		}
	}

	// 清除错误
	f.mu.Lock()
	f.Error = nil
	f.mu.Unlock()
	return nil
}

// Form V3 表单组件
type Form struct {
	*component.BaseComponent
	*component.StateHolder

	mu sync.RWMutex

	// 字段
	fields      map[string]*FormField
	fieldOrder  []string

	// 当前焦点字段索引
	currentField int

	// 提交/取消回调
	onSubmit func(data map[string]interface{}) error
	onCancel func()

	// 状态
	validating bool
	submitted  bool

	// 样式
	labelStyle      style.Style
	errorStyle      style.Style
	helpStyle        style.Style
	focusLabelStyle  style.Style

	// 脏标记回调
	dirtyCallback func()
}

// NewForm 创建 V3 表单组件
func NewForm() *Form {
	return &Form{
		BaseComponent: component.NewBaseComponent("form"),
		StateHolder:     component.NewStateHolder(),
		fields:          make(map[string]*FormField),
		fieldOrder:      make([]string, 0),
		currentField:    0,
		labelStyle:      style.Style{}.Foreground(style.Cyan),
		errorStyle:      style.Style{}.Foreground(style.Red),
		helpStyle:       style.Style{}.Foreground(style.BrightBlack),
		focusLabelStyle: style.Style{}.Foreground(style.White).Background(style.Blue),
	}
}

// ============================================================================
// 字段管理
// ============================================================================

// AddField 添加字段
func (f *Form) AddField(field *FormField) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fields[field.Name] = field
	f.fieldOrder = append(f.fieldOrder, field.Name)
	return f
}

// RemoveField 移除字段
func (f *Form) RemoveField(name string) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.fields, name)
	for i, n := range f.fieldOrder {
		if n == name {
			f.fieldOrder = append(f.fieldOrder[:i], f.fieldOrder[i+1:]...)
			break
		}
	}
	return f
}

// GetField 获取字段
func (f *Form) GetField(name string) (*FormField, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	field, ok := f.fields[name]
	return field, ok
}

// GetFields 获取所有字段
func (f *Form) GetFields() map[string]*FormField {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fields
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetOnSubmit 设置提交回调
func (f *Form) SetOnSubmit(fn func(data map[string]interface{}) error) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onSubmit = fn
	return f
}

// SetOnCancel 设置取消回调
func (f *Form) SetOnCancel(fn func()) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onCancel = fn
	return f
}

// SetLabelStyle 设置标签样式
func (f *Form) SetLabelStyle(s style.Style) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.labelStyle = s
	return f
}

// SetErrorStyle 设置错误样式
func (f *Form) SetErrorStyle(s style.Style) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errorStyle = s
	return f
}

// SetHelpStyle 设置帮助文本样式
func (f *Form) SetHelpStyle(s style.Style) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.helpStyle = s
	return f
}

// SetFocusLabelStyle 设置焦点标签样式
func (f *Form) SetFocusLabelStyle(s style.Style) *Form {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.focusLabelStyle = s
	return f
}

// SetDirtyCallback 设置脏标记回调
func (f *Form) SetDirtyCallback(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dirtyCallback = fn
}

// MountWithContext 使用组件上下文挂载
// 实现 component.MountableWithContext 接口，避免 App 直接依赖 Form 类型
func (f *Form) MountWithContext(parent component.Container, ctx *component.ComponentContext) {
	// 调用基础挂载
	f.Mount(parent)

	// 从上下文设置 dirty callback
	if fn := ctx.GetDirtyCallback(); fn != nil {
		f.SetDirtyCallback(fn)
	}

	// 递归为所有子组件注入上下文
	f.injectContextToFields(ctx)
}

// injectContextToFields 递归为所有字段中的组件注入上下文
func (f *Form) injectContextToFields(ctx *component.ComponentContext) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, field := range f.fields {
		if field.Input == nil {
			continue
		}

		// 如果输入组件实现了 MountableWithContext，调用它
		if mountable, ok := field.Input.(interface {
			MountWithContext(component.Container, *component.ComponentContext)
		}); ok {
			mountable.MountWithContext(nil, ctx)
		}
	}
}

// ============================================================================
// 表单操作
// ============================================================================

// Validate 验证表单
func (f *Form) Validate() error {
	// 不加锁，因为可能从已持有锁的方法调用
	// 调用者需要确保线程安全

	for _, field := range f.fields {
		if field.Visible && !field.Disabled {
			if err := field.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateField 验证单个字段
func (f *Form) ValidateField(name string) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	field, ok := f.fields[name]
	if !ok {
		return fmt.Errorf("field %s not found", name)
	}

	if field.Visible && !field.Disabled {
		return field.Validate()
	}
	return nil
}

// IsValid 检查表单是否有效
func (f *Form) IsValid() bool {
	return f.Validate() == nil
}

// GetValues 获取所有字段的值
func (f *Form) GetValues() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	values := make(map[string]interface{})
	for name, field := range f.fields {
		if field.Visible && !field.Disabled {
			values[name] = field.GetValue()
		}
	}
	return values
}

// SetValue 设置字段值
func (f *Form) SetValue(name string, value interface{}) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	field, ok := f.fields[name]
	if !ok {
		return fmt.Errorf("field %s not found", name)
	}

	return field.SetValue(value)
}

// Reset 重置表单
func (f *Form) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, field := range f.fields {
		field.Error = nil
		field.Touched = false
	}
	f.submitted = false
	f.currentField = 0
}

// Submit 提交表单
func (f *Form) Submit() error {
	// 验证所有字段
	if err := f.Validate(); err != nil {
		return err
	}

	// 标记所有字段为已触摸
	f.mu.Lock()
	for _, field := range f.fields {
		field.Touched = true
	}
	onSubmit := f.onSubmit
	f.mu.Unlock()

	// 调用提交回调
	if onSubmit != nil {
		values := f.GetValues()
		if err := onSubmit(values); err != nil {
			return err
		}
	}

	f.mu.Lock()
	f.submitted = true
	f.mu.Unlock()

	return nil
}

// Cancel 取消表单
func (f *Form) Cancel() {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.onCancel != nil {
		f.onCancel()
	}
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (f *Form) Measure(maxWidth, maxHeight int) (width, height int) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 计算高度：每个字段 = 标签(1) + 输入(1) + 错误/帮助(0-1) + 间距(1)
	visibleFields := 0
	maxLabelWidth := 0

	for _, name := range f.fieldOrder {
		field := f.fields[name]
		if field.Visible && !field.Disabled {
			visibleFields++
			labelWidth := len([]rune(field.Label))
			if labelWidth > maxLabelWidth {
				maxLabelWidth = labelWidth
			}
		}
	}

	// 每个字段大约 3 行
	height = visibleFields * 3
	width = maxLabelWidth + 30 // 输入框宽度

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
func (f *Form) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !f.IsVisible() {
		return
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	formDebugLog("PAINT: ctx.X=%d ctx.Y=%d, focused=%v, currentField=%d",
		ctx.X, ctx.Y, f.IsFocused(), f.currentField)

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	x := ctx.X
	y := ctx.Y

	formDebugLog("PAINT: starting at (x=%d, y=%d)", x, y)

	// 绘制每个字段
	for i, name := range f.fieldOrder {
		field := f.fields[name]
		if !field.Visible || field.Disabled {
			continue
		}

		if y >= ctx.Y+height {
			break
		}

		// 确定标签样式
		labelStyle := f.labelStyle
		if i == f.currentField && f.IsFocused() {
			labelStyle = f.focusLabelStyle
		}
		if field.Error != nil {
			labelStyle = f.errorStyle
		}

		// 绘制标签
		label := field.Label
		if len(field.Validators) > 0 {
			label += " *"
		}
		formDebugLog("FIELD[%d]: drawing label '%s' at (x=%d, y=%d)", i, label, x, y)
		f.drawText(buf, x, y, label, labelStyle)
		y++

		// 绘制输入组件
		if txt, ok := field.Input.(*input.TextInput); ok {
			// 创建子组件的 PaintContext，需要包含 Buffer 引用
			inputCtx := component.NewPaintContext(buf, x+2, y, width-2, 1)
			formDebugLog("FIELD[%d]: calling TextInput.Paint with ctx.X=%d ctx.Y=%d",
				i, inputCtx.X, inputCtx.Y)
			txt.Paint(inputCtx, buf)
		}
		y++

		// 绘制错误提示
		if field.Error != nil {
			f.drawText(buf, x+2, y, "  ⚠ "+field.Error.Error(), f.errorStyle)
			y++
		} else if field.HelpText != "" {
			f.drawText(buf, x+2, y, "  ⓘ "+field.HelpText, f.helpStyle)
			y++
		}

		y++ // 字段间距
	}
}

// drawText 绘制文本
func (f *Form) drawText(buf *paint.Buffer, x, y int, text string, s style.Style) {
	runes := []rune(text)
	for i, r := range runes {
		buf.SetCell(x+i, y, r, s)
	}
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction 处理语义化 Action
func (f *Form) HandleAction(a action.Action) bool {
	switch a.Type {
	case action.ActionNavigateDown:
		f.mu.Lock()
		f.navigateField(1)
		f.mu.Unlock()
		return true

	case action.ActionNavigateUp:
		f.mu.Lock()
		f.navigateField(-1)
		f.mu.Unlock()
		return true

	case action.ActionSubmit:
		// 提交表单 - 不持有锁，避免死锁
		return f.handleSubmit()

	case action.ActionCancel:
		// 取消表单
		f.mu.Lock()
	 onCancel := f.onCancel
		f.mu.Unlock()

		if onCancel != nil {
			onCancel()
		}
		return true

	default:
		// 转发给当前焦点字段
		f.mu.Lock()
		currentField := f.getCurrentField()
		f.mu.Unlock()

		if currentField != nil {
			if actionTarget, ok := currentField.Input.(interface {
				HandleAction(action.Action) bool
			}); ok {
				return actionTarget.HandleAction(a)
			}
		}
		return false
	}
}

// handleSubmit 内部提交方法（不加锁）
func (f *Form) handleSubmit() bool {
	// 验证所有字段
	if err := f.Validate(); err != nil {
		// 验证失败，标记需要重绘以显示错误
		return true
	}

	// 标记所有字段为已触摸
	f.mu.Lock()
	for _, field := range f.fields {
		field.Touched = true
	}
	onSubmit := f.onSubmit
	f.mu.Unlock()

	// 调用提交回调
	if onSubmit != nil {
		values := f.GetValues()
		if err := onSubmit(values); err != nil {
			return true
		}
	}

	f.mu.Lock()
	f.submitted = true
	f.mu.Unlock()

	return true
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (f *Form) FocusID() string {
	return f.ID()
}

// OnFocus 获得焦点时调用
func (f *Form) OnFocus() {
	// 首先调用 BaseComponent 的 OnFocus，设置 focused = true
	f.BaseComponent.OnFocus()

	// 焦点第一个可见字段
	visibleFields := f.getVisibleFieldIndices()
	if len(visibleFields) > 0 {
		// 如果当前字段不是可见字段，重置到第一个可见字段
		isCurrentVisible := false
		for _, idx := range visibleFields {
			if idx == f.currentField {
				isCurrentVisible = true
				break
			}
		}
		if !isCurrentVisible {
			f.currentField = visibleFields[0]
		}

		// 让当前字段获得焦点
		fieldName := f.fieldOrder[f.currentField]
		if field := f.fields[fieldName]; field != nil {
			// 直接调用具体类型的方法
			if input, ok := field.Input.(*input.TextInput); ok {
				input.OnFocus()
			}
		}
	}
}

// OnBlur 失去焦点时调用
func (f *Form) OnBlur() {
	// 首先调用 BaseComponent 的 OnBlur，设置 focused = false
	f.BaseComponent.OnBlur()

	// 让当前字段失去焦点
	fieldName := f.fieldOrder[f.currentField]
	if field := f.fields[fieldName]; field != nil {
		// 直接调用具体类型的方法
		if input, ok := field.Input.(*input.TextInput); ok {
			input.OnBlur()
		}
	}
}

// ============================================================================
// 内部方法
// ============================================================================

// navigateField 导航到下一个/上一个字段
func (f *Form) navigateField(delta int) {
	visibleFields := f.getVisibleFieldIndices()
	if len(visibleFields) == 0 {
		return
	}

	// 找到当前字段在可见字段中的索引
	currentIdx := -1
	for i, idx := range visibleFields {
		if idx == f.currentField {
			currentIdx = i
			break
		}
	}

	// 计算新的索引
	newIdx := currentIdx + delta
	if newIdx < 0 {
		newIdx = len(visibleFields) - 1
	} else if newIdx >= len(visibleFields) {
		newIdx = 0
	}

	// 先让当前字段失去焦点
	if currentIdx >= 0 && currentIdx < len(visibleFields) {
		oldFieldIdx := visibleFields[currentIdx]
		oldFieldName := f.fieldOrder[oldFieldIdx]
		if oldField := f.fields[oldFieldName]; oldField != nil {
			if input, ok := oldField.Input.(*input.TextInput); ok {
				input.OnBlur()
			}
		}
	}

	// 更新当前字段索引
	f.currentField = visibleFields[newIdx]

	// 让新字段获得焦点
	newFieldName := f.fieldOrder[f.currentField]
	if newField := f.fields[newFieldName]; newField != nil {
		if input, ok := newField.Input.(*input.TextInput); ok {
			input.OnFocus()
		}
	}
}

// getCurrentField 获取当前焦点字段
func (f *Form) getCurrentField() *FormField {
	if f.currentField < 0 || f.currentField >= len(f.fieldOrder) {
		return nil
	}
	name := f.fieldOrder[f.currentField]
	return f.fields[name]
}

// getVisibleFieldIndices 获取可见字段索引
func (f *Form) getVisibleFieldIndices() []int {
	indices := make([]int, 0)
	for i, name := range f.fieldOrder {
		field := f.fields[name]
		if field.Visible && !field.Disabled {
			indices = append(indices, i)
		}
	}
	return indices
}

// ============================================================================
// Event Handling (V3)
// ============================================================================

// HandleEvent 处理原始事件 (V3: 将事件转换为 Action)
// 这是 bridge 方法，将旧的事件系统连接到 V3 的 Action 系统
func (f *Form) HandleEvent(ev component.Event) bool {
	if keyEv, ok := ev.(*event.KeyEvent); ok {
		// 处理特殊键
		if keyEv.Special == event.KeyEscape {
			return f.HandleAction(*action.NewAction(action.ActionCancel))
		}
		if keyEv.Special == event.KeyEnter {
			return f.HandleAction(*action.NewAction(action.ActionSubmit))
		}
		if keyEv.Special == event.KeyUp || (keyEv.Special == event.KeyTab && keyEv.Modifiers&event.ModShift != 0) {
			return f.HandleAction(*action.NewAction(action.ActionNavigateUp))
		}
		if keyEv.Special == event.KeyDown || keyEv.Special == event.KeyTab {
			return f.HandleAction(*action.NewAction(action.ActionNavigateDown))
		}

		// 处理普通字符输入 - 转发给当前焦点字段
		if keyEv.Key.Rune != 0 && keyEv.Special == event.KeyUnknown {
			currentField := f.getCurrentField()
			if currentField != nil {
				// 检查 Input 是否有 HandleEvent 方法
				if handler, ok := currentField.Input.(event.EventComponent); ok {
					if handler.HandleEvent(ev) {
						f.markDirty()
						return true
					}
				} else {
					// Type assertion failed - try direct dispatch
					// This shouldn't happen if TextInput is properly set up
					if txt, ok := currentField.Input.(*input.TextInput); ok {
						if txt.HandleEvent(ev) {
							f.markDirty()
							return true
						}
					}
				}
			}
		}

		// 处理退格键
		if keyEv.Special == event.KeyBackspace {
			currentField := f.getCurrentField()
			if currentField != nil {
				// 优先使用接口类型断言
				if handler, ok := currentField.Input.(event.EventComponent); ok {
					if handler.HandleEvent(ev) {
						f.markDirty()
						return true
					}
				} else {
					// 回退到具体类型
					if txt, ok := currentField.Input.(*input.TextInput); ok {
						if txt.HandleEvent(ev) {
							f.markDirty()
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// markDirty 标记表单为需要重绘
func (f *Form) markDirty() {
	f.mu.RLock()
	callback := f.dirtyCallback
	f.mu.RUnlock()

	if callback != nil {
		callback()
	}
}
