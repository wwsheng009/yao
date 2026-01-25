# Form Validation System Design (V3)

> **优先级**: P2 (表单验证)
> **目标**: 支持用户表单输入验证
> **关键特性**: 字段验证、实时反馈、错误显示、提交检查

## 概述

表单验证是 TUI 应用中常见的功能，用户需要填写各种信息（如登录、注册、配置等），系统需要验证输入的有效性。

### 为什么需要表单验证系统？

**传统方式的问题**：
```go
// ❌ 分散的验证逻辑
func HandleLogin(username, password string) error {
    if username == "" {
        return errors.New("username required")
    }
    if len(username) < 3 {
        return errors.New("username too short")
    }
    if password == "" {
        return errors.New("password required")
    }
    // ...
}

// 问题：
// - 验证逻辑分散
// - 错误处理不统一
// - 无实时反馈
// - 难以复用
```

**表单验证系统的优势**：
```go
// ✅ 声明式验证
form := NewForm()
form.AddField("username", NewTextInput().
    WithValidator(
        Required(),
        MinLength(3),
        MaxLength(20),
    ))

form.AddField("password", NewTextInput().
    WithValidator(
        Required(),
        MinLength(8),
    ))

// 优势：
// - 验证逻辑集中
// - 实时错误反馈
// - 统一错误处理
// - 可复用验证器
```

## 设计目标

1. **声明式验证**: 通过配置定义验证规则
2. **实时反馈**: 输入时即时显示验证结果
3. **错误显示**: 清晰的错误提示
4. **可组合**: 支持复杂验证规则组合
5. **可扩展**: 支持自定义验证器
6. **国际化**: 支持多语言错误消息

## 核心类型定义

### 1. Validator 接口

```go
// 位于: tui/framework/validation/validator.go

package validation

// Validator 验证器接口
type Validator interface {
    // Validate 验证值
    Validate(value interface{}) error

    // Message 获取错误消息
    Message() string

    // WithMessage 设置错误消息
    WithMessage(msg string) Validator
}

// ValidatorFunc 验证函数类型
type ValidatorFunc func(value interface{}) error

// FuncValidator 函数验证器
type FuncValidator struct {
    fn      ValidatorFunc
    message string
}

// NewFuncValidator 创建函数验证器
func NewFuncValidator(fn ValidatorFunc, message string) *FuncValidator {
    return &FuncValidator{
        fn:      fn,
        message: message,
    }
}

// Validate 验证
func (v *FuncValidator) Validate(value interface{}) error {
    if err := v.fn(value); err != nil {
        if v.message != "" {
            return fmt.Errorf(v.message)
        }
        return err
    }
    return nil
}

// Message 获取消息
func (v *FuncValidator) Message() string {
    return v.message
}

// WithMessage 设置消息
func (v *FuncValidator) WithMessage(msg string) Validator {
    v.message = msg
    return v
}

// CompositeValidator 组合验证器
type CompositeValidator struct {
    validators []Validator
    message    string
    mode       CompositeMode
}

type CompositeMode int

const (
    // ModeAll 所有验证器都必须通过（AND）
    ModeAll CompositeMode = iota

    // ModeAny 至少一个验证器通过（OR）
    ModeAny
)

// NewAllValidator 创建 AND 验证器
func NewAllValidator(validators ...Validator) *CompositeValidator {
    return &CompositeValidator{
        validators: validators,
        mode:       ModeAll,
    }
}

// NewAnyValidator 创建 OR 验证器
func NewAnyValidator(validators ...Validator) *CompositeValidator {
    return &CompositeValidator{
        validators: validators,
        mode:       ModeAny,
    }
}

// Validate 验证
func (v *CompositeValidator) Validate(value interface{}) error {
    var errors []error

    for _, validator := range v.validators {
        err := validator.Validate(value)
        if v.mode == ModeAll {
            if err != nil {
                return err
            }
        } else { // ModeAny
            if err == nil {
                return nil
            }
            errors = append(errors, err)
        }
    }

    if v.mode == ModeAny && len(errors) > 0 {
        return fmt.Errorf("none of the validators passed")
    }

    return nil
}

// Message 获取消息
func (v *CompositeValidator) Message() string {
    return v.message
}

// WithMessage 设置消息
func (v *CompositeValidator) WithMessage(msg string) Validator {
    v.message = msg
    return v
}
```

### 2. 内置验证器

```go
// 位于: tui/framework/validation/builtin.go

package validation

import (
    "regexp"
    "strings"
)

// Required 必填验证器
func Required() Validator {
    return &FuncValidator{
        fn: func(value interface{}) error {
            if value == nil {
                return ErrRequired
            }
            switch v := value.(type) {
            case string:
                if strings.TrimSpace(v) == "" {
                    return ErrRequired
                }
            case int, int64, float64:
                return nil
            case []interface{}:
                if len(v) == 0 {
                    return ErrRequired
                }
            }
            return nil
        },
        message: "This field is required",
    }
}

// MinLength 最小长度验证器
func MinLength(min int) Validator {
    return &FuncValidator{
        fn: func(value interface{}) error {
            str, ok := value.(string)
            if !ok {
                return ErrTypeMismatch
            }
            if len(str) < min {
                return fmt.Errorf("must be at least %d characters", min)
            }
            return nil
        },
        message: fmt.Sprintf("Must be at least %d characters", min),
    }
}

// MaxLength 最大长度验证器
func MaxLength(max int) Validator {
    return &FuncValidator{
        fn: func(value interface{}) error {
            str, ok := value.(string)
            if !ok {
                return ErrTypeMismatch
            }
            if len(str) > max {
                return fmt.Errorf("must be at most %d characters", max)
            }
            return nil
        },
        message: fmt.Sprintf("Must be at most %d characters", max),
    }
}

// Min 最小值验证器
func Min(min float64) Validator {
    return &FuncValidator{
        fn: func(value interface{}) error {
            var num float64
            switch v := value.(type) {
            case int:
                num = float64(v)
            case int64:
                num = float64(v)
            case float64:
                num = v
            case string:
                f, err := strconv.ParseFloat(v, 64)
                if err != nil {
                    return ErrTypeMismatch
                }
                num = f
            default:
                return ErrTypeMismatch
            }
            if num < min {
                return fmt.Errorf("must be at least %v", min)
            }
            return nil
        },
        message: fmt.Sprintf("Must be at least %v", min),
    }
}

// Max 最大值验证器
func Max(max float64) Validator {
    return &FuncValidator{
        fn: func(value interface{}) error {
            var num float64
            switch v := value.(type) {
            case int:
                num = float64(v)
            case int64:
                num = float64(v)
            case float64:
                num = v
            case string:
                f, err := strconv.ParseFloat(v, 64)
                if err != nil {
                    return ErrTypeMismatch
                }
                num = f
            default:
                return ErrTypeMismatch
            }
            if num > max {
                return fmt.Errorf("must be at most %v", max)
            }
            return nil
        },
        message: fmt.Sprintf("Must be at most %v", max),
    }
}

// Pattern 正则表达式验证器
func Pattern(pattern string) Validator {
    re := regexp.MustCompile(pattern)
    return &FuncValidator{
        fn: func(value interface{}) error {
            str, ok := value.(string)
            if !ok {
                return ErrTypeMismatch
            }
            if !re.MatchString(str) {
                return fmt.Errorf("must match pattern: %s", pattern)
            }
            return nil
        },
        message: fmt.Sprintf("Must match pattern: %s", pattern),
    }
}

// Email 邮箱验证器
func Email() Validator {
    emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    return Pattern(emailPattern).WithMessage("Must be a valid email address")
}

// URL URL 验证器
func URL() Validator {
    urlPattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
    return Pattern(urlPattern).WithMessage("Must be a valid URL")
}

// Range 范围验证器
func Range(min, max float64) Validator {
    return NewAllValidator(Min(min), Max(max)).
        WithMessage(fmt.Sprintf("Must be between %v and %v", min, max))
}

// Length 长度验证器
func Length(min, max int) Validator {
    return NewAllValidator(MinLength(min), MaxLength(max)).
        WithMessage(fmt.Sprintf("Must be between %d and %d characters", min, max))
}

// OneOf 枚举验证器
func OneOf(values ...interface{}) Validator {
    return &FuncValidator{
        fn: func(value interface{}) error {
            for _, v := range values {
                if value == v {
                    return nil
                }
            }
            return fmt.Errorf("must be one of: %v", values)
        },
        message: fmt.Sprintf("Must be one of: %v", values),
    }
}

// Custom 自定义验证器
func Custom(fn ValidatorFunc, message string) Validator {
    return NewFuncValidator(fn, message)
}

// 标准错误
var (
    ErrRequired     = errors.New("required")
    ErrTypeMismatch = errors.New("type mismatch")
)
```

### 3. Form 表单

```go
// 位于: tui/framework/component/form.go

package component

import (
    "github.com/yaoapp/yao/tui/framework/validation"
)

// Form 表单组件
type Form struct {
    BaseComponent
    *Measurable
    *ThemeHolder

    // 字段
    fields map[string]*FormField

    // 布局
    layout FormLayout

    // 提交回调
    onSubmit func(data map[string]interface{}) error

    // 取消回调
    onCancel func()

    // 状态
    validating bool
    submitted  bool
}

// FormField 表单字段
type FormField struct {
    // 基本信息
    Name        string
    Label       string
    Placeholder string
    HelpText    string

    // 输入组件
    Input Component

    // 验证器
    Validators []validation.Validator

    // 状态
    Error      error
    Touched    bool
    Dirty      bool

    // 条件
    Visible    bool
    Disabled   bool
    Condition  func(values map[string]interface{}) bool

    // 依赖
    DependsOn  []string
}

// FormLayout 表单布局
type FormLayout int

const (
    // LayoutVertical 垂直布局
    LayoutVertical FormLayout = iota

    // LayoutHorizontal 水平布局
    LayoutHorizontal

    // LayoutGrid 网格布局
    LayoutGrid
)

// NewForm 创建表单
func NewForm() *Form {
    form := &Form{
        fields: make(map[string]*FormField),
        layout: LayoutVertical,
    }

    form.Measurable = NewMeasurable()
    form.ThemeHolder = NewThemeHolder(nil)

    return form
}

// AddField 添加字段
func (f *Form) AddField(field *FormField) *Form {
    f.fields[field.Name] = field
    return f
}

// RemoveField 移除字段
func (f *Form) RemoveField(name string) *Form {
    delete(f.fields, name)
    return f
}

// GetField 获取字段
func (f *Form) GetField(name string) (*FormField, bool) {
    field, ok := f.fields[name]
    return field, ok
}

// SetLayout 设置布局
func (f *Form) SetLayout(layout FormLayout) *Form {
    f.layout = layout
    return f
}

// SetOnSubmit 设置提交回调
func (f *Form) SetOnSubmit(fn func(data map[string]interface{}) error) *Form {
    f.onSubmit = fn
    return f
}

// SetOnCancel 设置取消回调
func (f *Form) SetOnCancel(fn func()) *Form {
    f.onCancel = fn
    return f
}

// Validate 验证表单
func (f *Form) Validate() error {
    var errors []error

    for _, field := range f.fields {
        if !field.Visible || field.Disabled {
            continue
        }

        // 获取字段值
        value := f.getFieldValue(field)

        // 验证
        for _, validator := range field.Validators {
            if err := validator.Validate(value); err != nil {
                field.Error = err
                errors = append(errors, fmt.Errorf("%s: %w", field.Name, err))
                break
            } else {
                field.Error = nil
            }
        }
    }

    if len(errors) > 0 {
        return errors[0]
    }

    return nil
}

// ValidateField 验证单个字段
func (f *Form) ValidateField(name string) error {
    field, ok := f.fields[name]
    if !ok {
        return fmt.Errorf("field %s not found", name)
    }

    if !field.Visible || field.Disabled {
        return nil
    }

    value := f.getFieldValue(field)

    for _, validator := range field.Validators {
        if err := validator.Validate(value); err != nil {
            field.Error = err
            return err
        }
    }

    field.Error = nil
    return nil
}

// IsValid 检查表单是否有效
func (f *Form) IsValid() bool {
    return f.Validate() == nil
}

// GetValues 获取所有字段的值
func (f *Form) GetValues() map[string]interface{} {
    values := make(map[string]interface{})

    for name, field := range f.fields {
        if field.Visible && !field.Disabled {
            values[name] = f.getFieldValue(field)
        }
    }

    return values
}

// SetValue 设置字段值
func (f *Form) SetValue(name string, value interface{}) error {
    field, ok := f.fields[name]
    if !ok {
        return fmt.Errorf("field %s not found", name)
    }

    f.setFieldValue(field, value)
    field.Dirty = true

    // 验证字段
    if field.Touched {
        f.ValidateField(name)
    }

    f.MarkDirty()
    return nil
}

// getFieldValue 获取字段值
func (f *Form) getFieldValue(field *FormField) interface{} {
    switch input := field.Input.(type) {
    case *TextInput:
        return input.Value()
    case *NumberInput:
        return input.Value()
    case *CheckBox:
        return input.IsChecked()
    case *Select:
        return input.Selected()
    case *MultiSelect:
        return input.SelectedItems()
    default:
        return nil
    }
}

// setFieldValue 设置字段值
func (f *Form) setFieldValue(field *FormField, value interface{}) {
    switch input := field.Input.(type) {
    case *TextInput:
        if str, ok := value.(string); ok {
            input.SetValue(str)
        }
    case *NumberInput:
        switch v := value.(type) {
        case int:
            input.SetValue(float64(v))
        case float64:
            input.SetValue(v)
        case string:
            if num, err := strconv.ParseFloat(v, 64); err == nil {
                input.SetValue(num)
            }
        }
    case *CheckBox:
        if b, ok := value.(bool); ok {
            input.SetChecked(b)
        }
    case *Select:
        if str, ok := value.(string); ok {
            input.SelectOption(str)
        }
    }
}

// Reset 重置表单
func (f *Form) Reset() {
    for _, field := range f.fields {
        field.Error = nil
        field.Touched = false
        field.Dirty = false

        // 重置输入组件
        if resettable, ok := field.Input.(Resettable); ok {
            resettable.Reset()
        }
    }

    f.submitted = false
    f.MarkDirty()
}

// Submit 提交表单
func (f *Form) Submit() error {
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
func (f *Form) Cancel() {
    if f.onCancel != nil {
        f.onCancel()
    }
}

// HandleAction 处理 Action
func (f *Form) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateDown, action.ActionNavigateUp:
        // 焦点导航
        return f.handleNavigation(a)

    case action.ActionSubmit:
        // 提交表单
        if err := f.Submit(); err != nil {
            // 显示错误
            f.MarkDirty()
            return true
        }
        return f.onSubmit != nil

    case action.ActionCancel:
        f.Cancel()
        return true

    default:
        // 转发给当前焦点字段
        if focusField := f.getFocusedField(); focusField != nil {
            if actionTarget, ok := focusField.Input.(ActionTarget); ok {
                return actionTarget.HandleAction(a)
            }
        }
        return false
    }
}

// handleNavigation 处理导航
func (f *Form) handleNavigation(a *action.Action) bool {
    fields := f.getVisibleFields()
    if len(fields) == 0 {
        return false
    }

    currentIdx := f.getFocusedFieldIndex()
    var newIdx int

    if a.Type == action.ActionNavigateDown {
        newIdx = currentIdx + 1
        if newIdx >= len(fields) {
            newIdx = 0
        }
    } else { // NavigateUp
        newIdx = currentIdx - 1
        if newIdx < 0 {
            newIdx = len(fields) - 1
        }
    }

    // 设置焦点
    f.focusField(fields[newIdx])
    return true
}

// getVisibleFields 获取可见字段
func (f *Form) getVisibleFields() []*FormField {
    fields := make([]*FormField, 0)
    for _, field := range f.fields {
        if field.Visible && !field.Disabled {
            fields = append(fields, field)
        }
    }
    return fields
}

// getFocusedField 获取当前焦点字段
func (f *Form) getFocusedField() *FormField {
    for _, field := range f.fields {
        if focusable, ok := field.Input.(Focusable); ok {
            if focusable.IsFocused() {
                return field
            }
        }
    }
    return nil
}

// getFocusedFieldIndex 获取焦点字段索引
func (f *Form) getFocusedFieldIndex() int {
    fields := f.getVisibleFields()
    for i, field := range fields {
        if focusable, ok := field.Input.(Focusable); ok {
            if focusable.IsFocused() {
                return i
            }
        }
    }
    return 0
}

// focusField 聚焦字段
func (f *Form) focusField(field *FormField) {
    // 取消其他字段焦点
    for _, f := range f.fields {
        if focusable, ok := f.Input.(Focusable); ok {
            focusable.SetFocused(false)
        }
    }

    // 设置焦点
    if focusable, ok := field.Input.(Focusable); ok {
        focusable.SetFocused(true)
    }
}

// Paint 绘制表单
func (f *Form) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    theme := f.GetTheme()
    bounds := f.Bounds()

    y := bounds.Y

    // 绘制每个字段
    for _, field := range f.getVisibleFields() {
        // 绘制标签
        label := field.Label
        if field.Error != nil {
            label += " *"
        }

        labelStyle := theme.GetStyle("form.label")
        if field.Error != nil {
            labelStyle = theme.GetStyle("form.label.error")
        }

        buf.DrawText(bounds.X, y, label, labelStyle)
        y++

        // 绘制输入组件
        if paintable, ok := field.Input.(Paintable); ok {
            inputBounds := runtime.Rect{
                X:      bounds.X + 2,
                Y:      y,
                Width:  bounds.Width - 4,
                Height: 1,
            }
            field.Input.SetBounds(inputBounds)
            paintable.Paint(ctx, buf)
        }
        y++

        // 绘制错误提示
        if field.Error != nil {
            errorStyle := theme.GetStyle("form.error")
            buf.DrawText(bounds.X + 2, y, field.Error.Error(), errorStyle)
            y++
        }

        // 绘制帮助文本
        if field.HelpText != "" && field.Error == nil {
            helpStyle := theme.GetStyle("form.help")
            buf.DrawText(bounds.X + 2, y, field.HelpText, helpStyle)
            y++
        }

        y++ // 字段间距
    }
}

// Measure 测量尺寸
func (f *Form) Measure(maxWidth, maxHeight int) (width, height int) {
    fieldCount := len(f.getVisibleFields())

    // 每个字段：标签(1) + 输入(1) + 错误/帮助(0-1) + 间距(1)
    // 估算高度
    return maxWidth, fieldCount * 4
}
```

### 4. FormField 构建器

```go
// 位于: tui/framework/component/form_builder.go

package component

// NewFormField 创建表单字段
func NewFormField(name string) *FormFieldBuilder {
    return &FormFieldBuilder{
        field: &FormField{
            Name:     name,
            Visible:  true,
            Disabled: false,
        },
    }
}

// FormFieldBuilder 表单字段构建器
type FormFieldBuilder struct {
    field *FormField
}

// WithLabel 设置标签
func (b *FormFieldBuilder) WithLabel(label string) *FormFieldBuilder {
    b.field.Label = label
    return b
}

// WithPlaceholder 设置占位符
func (b *FormFieldBuilder) WithPlaceholder(placeholder string) *FormFieldBuilder {
    b.field.Placeholder = placeholder
    return b
}

// WithHelpText 设置帮助文本
func (b *FormFieldBuilder) WithHelpText(text string) *FormFieldBuilder {
    b.field.HelpText = text
    return b
}

// WithInput 设置输入组件
func (b *FormFieldBuilder) WithInput(input Component) *FormFieldBuilder {
    b.field.Input = input
    return b
}

// WithValidators 添加验证器
func (b *FormFieldBuilder) WithValidators(validators ...validation.Validator) *FormFieldBuilder {
    b.field.Validators = append(b.field.Validators, validators...)
    return b
}

// WithRequired 设置必填
func (b *FormFieldBuilder) WithRequired() *FormFieldBuilder {
    b.field.Validators = append(b.field.Validators, validation.Required())
    return b
}

// WithCondition 设置显示条件
func (b *FormFieldBuilder) WithCondition(fn func(values map[string]interface{}) bool) *FormFieldBuilder {
    b.field.Condition = fn
    return b
}

// WithDisabled 设置禁用
func (b *FormFieldBuilder) WithDisabled(disabled bool) *FormFieldBuilder {
    b.field.Disabled = disabled
    return b
}

// WithDependsOn 设置依赖
func (b *FormFieldBuilder) WithDependsOn(fields ...string) *FormFieldBuilder {
    b.field.DependsOn = fields
    return b
}

// Build 构建字段
func (b *FormFieldBuilder) Build() *FormField {
    return b.field
}

// 便捷构建函数

// TextField 创建文本字段
func TextField(name, label string) *FormFieldBuilder {
    input := NewTextInput()
    return NewFormField(name).
        WithLabel(label).
        WithInput(input)
}

// NumberField 创建数字字段
func NumberField(name, label string) *FormFieldBuilder {
    input := NewNumberInput()
    return NewFormField(name).
        WithLabel(label).
        WithInput(input)
}

// EmailField 创建邮箱字段
func EmailField(name, label string) *FormFieldBuilder {
    input := NewTextInput()
    return NewFormField(name).
        WithLabel(label).
        WithInput(input).
        WithValidators(
            validation.Required(),
            validation.Email(),
        )
}

// PasswordField 创建密码字段
func PasswordField(name, label string) *FormFieldBuilder {
    input := NewTextInput()
    input.SetPassword(true)
    return NewFormField(name).
        WithLabel(label).
        WithInput(input).
        WithValidators(
            validation.Required(),
            validation.MinLength(8),
        )
}

// SelectField 创建选择字段
func SelectField(name, label string, options []string) *FormFieldBuilder {
    input := NewSelect(options)
    return NewFormField(name).
        WithLabel(label).
        WithInput(input)
}

// CheckBoxField 创建复选框字段
func CheckBoxField(name, label string) *FormFieldBuilder {
    input := NewCheckBox()
    return NewFormField(name).
        WithLabel(label).
        WithInput(input)
}
```

## 使用示例

### 示例 1：基础表单

```go
// ✅ 创建登录表单
form := component.NewForm()

form.AddField(
    component.TextField("username", "Username").
        WithPlaceholder("Enter your username").
        WithRequired().
        WithValidators(
            validation.MinLength(3),
            validation.MaxLength(20),
        ).
        Build(),
)

form.AddField(
    component.PasswordField("password", "Password").
        WithPlaceholder("Enter your password").
        Build(),
)

form.AddField(
    component.CheckBoxField("remember", "Remember me").
        Build(),
)

form.SetOnSubmit(func(data map[string]interface{}) error {
    username := data["username"].(string)
    password := data["password"].(string)
    remember := data["remember"].(bool)

    // 执行登录
    return doLogin(username, password, remember)
})

form.SetOnCancel(func() {
    // 取消登录
    app.Exit()
})

app.Mount(form)
```

### 示例 2：条件显示字段

```go
// ✅ 条件字段
form := component.NewForm()

// 是否有订阅
form.AddField(
    component.CheckBoxField("has_subscription", "Has Subscription").
        Build(),
)

// 订阅类型（仅当有订阅时显示）
form.AddField(
    component.SelectField("subscription_type", "Subscription Type",
        []string{"Basic", "Premium", "Enterprise"}).
        WithCondition(func(values map[string]interface{}) bool {
            hasSub, ok := values["has_subscription"].(bool)
            return ok && hasSub
        }).
        Build(),
)

// 每次值变化时检查条件
form.SetOnValueChange(func(name string, value interface{}) {
    form.UpdateVisibility()
})
```

### 示例 3：自定义验证器

```go
// ✅ 自定义验证器
// 年龄必须大于 18
ageValidator := validation.Custom(func(value interface{}) error {
    age, ok := value.(int)
    if !ok {
        return validation.ErrTypeMismatch
    }
    if age < 18 {
        return errors.New("must be 18 or older")
    }
    return nil
}, "You must be 18 or older to register")

form.AddField(
    component.NumberField("age", "Age").
        WithValidators(ageValidator).
        Build(),
)
```

### 示例 4：实时验证

```go
// ✅ 实时验证（输入时验证）
input := component.NewTextInput()
input.SetOnValueChange(func(value string) {
    // 字段值变化时验证
    form.ValidateField("email")
})

form.AddField(
    component.NewFormField("email").
        WithLabel("Email").
        WithInput(input).
        WithValidators(
            validation.Required(),
            validation.Email(),
        ).
        Build(),
)
```

### 示例 5：多步表单

```go
// ✅ 多步表单
type MultiStepForm struct {
    currentStep int
    steps       []*component.Form
}

func NewMultiStepForm() *MultiStepForm {
    return &MultiStepForm{
        currentStep: 0,
        steps: []*component.Form{
            // 第一步：基本信息
            createBasicInfoForm(),
            // 第二步：联系方式
            createContactForm(),
            // 第三步：确认
            createConfirmForm(),
        },
    }
}

func (m *MultiStepForm) Next() error {
    // 验证当前步骤
    if err := m.steps[m.currentStep].Validate(); err != nil {
        return err
    }

    // 进入下一步
    if m.currentStep < len(m.steps)-1 {
        m.currentStep++
    }

    return nil
}

func (m *MultiStepForm) Prev() {
    if m.currentStep > 0 {
        m.currentStep--
    }
}
```

## 测试

```go
// 位于: tui/framework/component/form_test.go

func TestFormValidation(t *testing.T) {
    form := component.NewForm()

    form.AddField(
        component.TextField("email", "Email").
            WithValidators(
                validation.Required(),
                validation.Email(),
            ).
            Build(),
    )

    // 测试有效值
    form.SetValue("email", "test@example.com")
    err := form.ValidateField("email")
    assert.NoError(t, err)

    // 测试无效值
    form.SetValue("email", "invalid")
    err = form.ValidateField("email")
    assert.Error(t, err)
}

func TestFormSubmit(t *testing.T) {
    submitted := false
    submitData := make(map[string]interface{})

    form := component.NewForm()

    form.AddField(
        component.TextField("name", "Name").
            WithRequired().
            Build(),
    )

    form.SetOnSubmit(func(data map[string]interface{}) error {
        submitted = true
        submitData = data
        return nil
    })

    // 提交表单
    form.SetValue("name", "John")
    err := form.Submit()

    assert.NoError(t, err)
    assert.True(t, submitted)
    assert.Equal(t, "John", submitData["name"])
}
```

## 总结

表单验证系统提供：

1. **声明式验证**: 通过配置定义验证规则
2. **实时反馈**: 输入时即时显示验证结果
3. **错误显示**: 清晰的错误提示
4. **可组合**: 支持复杂验证规则组合
5. **可扩展**: 支持自定义验证器
6. **条件字段**: 支持基于其他字段值的条件显示

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [ERROR_HANDLING.md](./ERROR_HANDLING.md) - 错误处理
