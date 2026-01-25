package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ==============================================================================
// Built-in Validators
// ==============================================================================

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
		message: "此字段为必填项",
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
			if len([]rune(str)) < min {
				return fmt.Errorf("最少需要 %d 个字符", min)
			}
			return nil
		},
		message: fmt.Sprintf("最少需要 %d 个字符", min),
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
			if len([]rune(str)) > max {
				return fmt.Errorf("最多允许 %d 个字符", max)
			}
			return nil
		},
		message: fmt.Sprintf("最多允许 %d 个字符", max),
	}
}

// Length 长度范围验证器
func Length(min, max int) Validator {
	return NewAllValidator(MinLength(min), MaxLength(max)).
		WithMessage(fmt.Sprintf("长度必须在 %d 到 %d 之间", min, max))
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
				return fmt.Errorf("必须大于等于 %v", min)
			}
			return nil
		},
		message: fmt.Sprintf("必须大于等于 %v", min),
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
				return fmt.Errorf("必须小于等于 %v", max)
			}
			return nil
		},
		message: fmt.Sprintf("必须小于等于 %v", max),
	}
}

// Range 范围验证器
func Range(min, max float64) Validator {
	return NewAllValidator(Min(min), Max(max)).
		WithMessage(fmt.Sprintf("必须在 %v 到 %v 之间", min, max))
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
				return fmt.Errorf("格式不正确")
			}
			return nil
		},
		message: "格式不正确",
	}
}

// Email 邮箱验证器
func Email() Validator {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return Pattern(emailPattern).WithMessage("请输入有效的邮箱地址")
}

// URL URL 验证器
func URL() Validator {
	urlPattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
	return Pattern(urlPattern).WithMessage("请输入有效的 URL")
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
			return fmt.Errorf("必须是以下值之一: %v", values)
		},
		message: fmt.Sprintf("必须是以下值之一: %v", values),
	}
}

// Custom 自定义验证器
func Custom(fn ValidatorFunc, message string) Validator {
	return NewFuncValidator(fn, message)
}

// ==============================================================================
// 标准错误
// ==============================================================================

var (
	ErrRequired     = fmt.Errorf("required")
	ErrTypeMismatch = fmt.Errorf("type mismatch")
)
