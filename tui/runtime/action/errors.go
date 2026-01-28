package action

import (
	"fmt"
)

// ==============================================================================
// Action Errors (V3)
// ==============================================================================
// 结构化 Action 处理错误，提供清晰的错误信息用于调试和日志

// ErrorType 错误类型
type ErrorType string

const (
	// 目标相关错误
	ErrTargetNotFound      ErrorType = "target_not_found"       // 目标组件未找到
	ErrTargetDisabled      ErrorType = "target_disabled"         // 目标组件已禁用
	ErrTargetNotInteractable ErrorType = "target_not_interactable" // 目标组件不可交互

	// Payload 相关错误
	ErrInvalidPayload      ErrorType = "invalid_payload"        // 无效的 Payload
	ErrMissingPayload      ErrorType = "missing_payload"        // 缺少必需的 Payload
	ErrPayloadTypeMismatch ErrorType = "payload_type_mismatch"  // Payload 类型不匹配

	// Action 相关错误
	ErrActionNotSupported  ErrorType = "action_not_supported"   // 组件不支持此 Action
	ErrActionNotAllowed    ErrorType = "action_not_allowed"     // 当前状态不允许此操作
	ErrActionFailed        ErrorType = "action_failed"          // Action 执行失败

	// 系统错误
	ErrDispatchFailed      ErrorType = "dispatch_failed"        // 分发失败
	ErrTimeout             ErrorType = "timeout"                // 操作超时
)

// Error Action 处理错误
type Error struct {
	// Type 错误类型
	Type ErrorType

	// Message 错误消息
	Message string

	// Action 触发错误的 Action
	Action *Action

	// Target 目标组件 ID
	Target string

	// ComponentType 目标组件类型
	ComponentType string

	// Details 详细信息
	Details map[string]interface{}
}

// NewError 创建错误
func NewError(errorType ErrorType, message string, a *Action) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Action:  a,
		Details: make(map[string]interface{}),
	}
}

// WithTarget 设置目标
func (e *Error) WithTarget(target string) *Error {
	e.Target = target
	return e
}

// WithComponentType 设置组件类型
func (e *Error) WithComponentType(typ string) *Error {
	e.ComponentType = typ
	return e
}

// WithDetail 添加详情
func (e *Error) WithDetail(key string, value interface{}) *Error {
	e.Details[key] = value
	return e
}

// WithDetails 添加多个详情
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Target != "" {
		return fmt.Sprintf("[%s] %s (action: %s, target: %s)",
			e.Type, e.Message, e.Action.Type, e.Target)
	}
	return fmt.Sprintf("[%s] %s (action: %s)", e.Type, e.Message, e.Action.Type)
}

// String 返回详细的错误字符串
func (e *Error) String() string {
	s := e.Error()
	if len(e.Details) > 0 {
		s += ", details: {"
		first := true
		for k, v := range e.Details {
			if !first {
				s += ", "
			}
			s += fmt.Sprintf("%s: %v", k, v)
			first = false
		}
		s += "}"
	}
	if e.ComponentType != "" {
		s += fmt.Sprintf(", component_type: %s", e.ComponentType)
	}
	return s
}

// ==============================================================================
// 预定义错误构造器
// ==============================================================================

// NewErrTargetNotFound 创建目标未找到错误
func NewErrTargetNotFound(targetID string, a *Action) *Error {
	return NewError(ErrTargetNotFound,
		fmt.Sprintf("target component not found: %s", targetID), a).
		WithTarget(targetID)
}

// NewErrTargetDisabled 创建目标已禁用错误
func NewErrTargetDisabled(targetID string, a *Action) *Error {
	return NewError(ErrTargetDisabled,
		fmt.Sprintf("target component is disabled: %s", targetID), a).
		WithTarget(targetID)
}

// NewErrInvalidPayload 创建无效 Payload 错误
func NewErrInvalidPayload(expectedType string, a *Action) *Error {
	return NewError(ErrInvalidPayload,
		fmt.Sprintf("invalid payload type: expected %s", expectedType), a).
		WithDetail("expected_type", expectedType).
		WithDetail("actual_type", fmt.Sprintf("%T", a.Payload))
}

// NewErrActionNotSupported 创建不支持 Action 错误
func NewErrActionNotSupported(componentType, actionType string, a *Action) *Error {
	return NewError(ErrActionNotSupported,
		fmt.Sprintf("component %s does not support action: %s", componentType, actionType), a).
		WithComponentType(componentType).
		WithDetail("action_type", actionType)
}

// NewErrActionNotAllowed 创建不允许 Action 错误
func NewErrActionNotAllowed(reason string, a *Action) *Error {
	return NewError(ErrActionNotAllowed,
		fmt.Sprintf("action not allowed: %s", reason), a).
		WithDetail("reason", reason)
}

// ==============================================================================
// Payload 验证辅助函数
// ==============================================================================

// ValidateStringPayload 验证字符串 Payload
func ValidateStringPayload(a *Action) (string, bool) {
	s, ok := a.Payload.(string)
	if !ok {
		return "", false
	}
	return s, true
}

// ValidateRunePayload 验证 rune Payload
func ValidateRunePayload(a *Action) (rune, bool) {
	r, ok := a.Payload.(rune)
	if !ok {
		return 0, false
	}
	return r, true
}

// ValidateIntPayload 验证 int Payload
func ValidateIntPayload(a *Action) (int, bool) {
	i, ok := a.Payload.(int)
	if !ok {
		return 0, false
	}
	return i, true
}

// ValidateMapPayload 验证 map Payload
func ValidateMapPayload(a *Action) (map[string]interface{}, bool) {
	m, ok := a.Payload.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return m, true
}
