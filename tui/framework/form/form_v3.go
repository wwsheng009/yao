package form

import (
	"fmt"
	"sync"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Form Component V3
// ==============================================================================
// V3 表单组件，支持验证、字段导航、提交/取消

// FormFieldV3 表单字段
type FormFieldV3 struct {
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

// NewFormFieldV3 创建表单字段
func NewFormFieldV3(name string) *FormFieldV3 {
	return &FormFieldV3{
		Name:        name,
		Visible:     true,
		Disabled:    false,
		Validators:  make([]validation.Validator, 0),
	}
}

// SetValue 设置值
func (f *FormFieldV3) SetValue(value interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 根据输入组件类型设置值
	switch input := f.Input.(type) {
	case *input.TextInputV3:
		if str, ok := value.(string); ok {
			input.SetValue(str)
		}
	}

	f.Touched = true
	return nil
}

// GetValue 获取值
func (f *FormFieldV3) GetValue() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	switch input := f.Input.(type) {
	case *input.TextInputV3:
		return input.GetValue()
	default:
		return nil
	}
}

// Validate 验证字段
func (f *FormFieldV3) Validate() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	value := f.GetValue()
	for _, validator := range f.Validators {
		if err := validator.Validate(value); err != nil {
			f.Error = err
			return err
		}
	}

	f.Error = nil
	return nil
}

// FormV3 V3 表单组件
type FormV3 struct {
	*component.BaseComponentV3
	*component.StateHolder

	mu sync.RWMutex

	// 字段
	fields      map[string]*FormFieldV3
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
}

// NewFormV3 创建 V3 表单组件
func NewFormV3() *FormV3 {
	return &FormV3{
		BaseComponentV3: component.NewBaseComponentV3("form"),
		StateHolder:     component.NewStateHolder(),
		fields:          make(map[string]*FormFieldV3),
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
func (f *FormV3) AddField(field *FormFieldV3) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.fields[field.Name] = field
	f.fieldOrder = append(f.fieldOrder, field.Name)
	return f
}

// RemoveField 移除字段
func (f *FormV3) RemoveField(name string) *FormV3 {
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
func (f *FormV3) GetField(name string) (*FormFieldV3, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	field, ok := f.fields[name]
	return field, ok
}

// GetFields 获取所有字段
func (f *FormV3) GetFields() map[string]*FormFieldV3 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fields
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetOnSubmit 设置提交回调
func (f *FormV3) SetOnSubmit(fn func(data map[string]interface{}) error) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onSubmit = fn
	return f
}

// SetOnCancel 设置取消回调
func (f *FormV3) SetOnCancel(fn func()) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onCancel = fn
	return f
}

// SetLabelStyle 设置标签样式
func (f *FormV3) SetLabelStyle(s style.Style) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.labelStyle = s
	return f
}

// SetErrorStyle 设置错误样式
func (f *FormV3) SetErrorStyle(s style.Style) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errorStyle = s
	return f
}

// SetHelpStyle 设置帮助文本样式
func (f *FormV3) SetHelpStyle(s style.Style) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.helpStyle = s
	return f
}

// SetFocusLabelStyle 设置焦点标签样式
func (f *FormV3) SetFocusLabelStyle(s style.Style) *FormV3 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.focusLabelStyle = s
	return f
}

// ============================================================================
// 表单操作
// ============================================================================

// Validate 验证表单
func (f *FormV3) Validate() error {
	f.mu.Lock()
	defer f.mu.Unlock()

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
func (f *FormV3) ValidateField(name string) error {
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
func (f *FormV3) IsValid() bool {
	return f.Validate() == nil
}

// GetValues 获取所有字段的值
func (f *FormV3) GetValues() map[string]interface{} {
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
func (f *FormV3) SetValue(name string, value interface{}) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	field, ok := f.fields[name]
	if !ok {
		return fmt.Errorf("field %s not found", name)
	}

	return field.SetValue(value)
}

// Reset 重置表单
func (f *FormV3) Reset() {
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
func (f *FormV3) Submit() error {
	// 验证所有字段
	if err := f.Validate(); err != nil {
		return err
	}

	// 标记所有字段为已触摸
	for _, field := range f.fields {
		field.Touched = true
	}

	// 调用提交回调
	if f.onSubmit != nil {
		values := f.GetValues()
		if err := f.onSubmit(values); err != nil {
			return err
		}
	}

	f.submitted = true
	return nil
}

// Cancel 取消表单
func (f *FormV3) Cancel() {
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
func (f *FormV3) Measure(maxWidth, maxHeight int) (width, height int) {
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
func (f *FormV3) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !f.IsVisible() {
		return
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	x := ctx.X
	y := ctx.Y

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
		f.drawText(buf, x, y, label, labelStyle)
		y++

		// 绘制输入组件
		if paintable, ok := field.Input.(interface {
			Paint(component.PaintContext, *paint.Buffer)
		}); ok {
			inputCtx := component.PaintContext{
				AvailableWidth:  width - 2,
				AvailableHeight: 1,
				X:                x + 2,
				Y:                y,
			}
			paintable.Paint(inputCtx, buf)
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
func (f *FormV3) drawText(buf *paint.Buffer, x, y int, text string, s style.Style) {
	runes := []rune(text)
	for i, r := range runes {
		buf.SetCell(x+i, y, r, s)
	}
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction 处理语义化 Action
func (f *FormV3) HandleAction(a action.Action) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch a.Type {
	case action.ActionNavigateDown:
		f.navigateField(1)
		return true

	case action.ActionNavigateUp:
		f.navigateField(-1)
		return true

	case action.ActionSubmit:
		// 提交表单
		if err := f.Submit(); err != nil {
			// 保留错误显示
			return true
		}
		return f.onSubmit != nil

	case action.ActionCancel:
		f.Cancel()
		return true

	default:
		// 转发给当前焦点字段
		currentField := f.getCurrentField()
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

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (f *FormV3) FocusID() string {
	return f.ID()
}

// OnFocus 获得焦点时调用
func (f *FormV3) OnFocus() {
	// 可以在这里添加获得焦点时的逻辑
}

// OnBlur 失去焦点时调用
func (f *FormV3) OnBlur() {
	// 可以在这里添加失去焦点时的逻辑
}

// ============================================================================
// 内部方法
// ============================================================================

// navigateField 导航到下一个/上一个字段
func (f *FormV3) navigateField(delta int) {
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

	f.currentField = visibleFields[newIdx]
}

// getCurrentField 获取当前焦点字段
func (f *FormV3) getCurrentField() *FormFieldV3 {
	if f.currentField < 0 || f.currentField >= len(f.fieldOrder) {
		return nil
	}
	name := f.fieldOrder[f.currentField]
	return f.fields[name]
}

// getVisibleFieldIndices 获取可见字段索引
func (f *FormV3) getVisibleFieldIndices() []int {
	indices := make([]int, 0)
	for i, name := range f.fieldOrder {
		field := f.fields[name]
		if field.Visible && !field.Disabled {
			indices = append(indices, i)
		}
	}
	return indices
}
