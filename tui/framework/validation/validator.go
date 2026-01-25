package validation

import (
	"fmt"
)

// ==============================================================================
// Validation System (V3)
// ==============================================================================
// 表单验证系统

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

// ==============================================================================
// Composite Validator
// ==============================================================================

// CompositeMode 组合模式
type CompositeMode int

const (
	// ModeAll 所有验证器都必须通过（AND）
	ModeAll CompositeMode = iota

	// ModeAny 至少一个验证器通过（OR）
	ModeAny
)

// CompositeValidator 组合验证器
type CompositeValidator struct {
	validators []Validator
	message    string
	mode       CompositeMode
}

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
