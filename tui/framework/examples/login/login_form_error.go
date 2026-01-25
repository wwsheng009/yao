//go:build error_test
// +build error_test

package main

import (
	"fmt"

	"github.com/yaoapp/yao/tui/framework/form"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/action"
)

// main 测试验证失败的登录表单示例
func main() {
	fmt.Println("========================================")
	fmt.Println("  TUI Framework - 验证失败测试")
	fmt.Println("========================================")
	fmt.Println()

	// 创建登录表单
	loginForm := createLoginForm()

	// 测试1: 用户名太短
	fmt.Println("测试1: 用户名太短")
	simulateInput(loginForm, action.ActionInputText, "ab")
	if err := loginForm.Submit(); err != nil {
		fmt.Printf("  验证失败 (预期): %v\n", err)
	}
	fmt.Println()

	// 重置表单
	loginForm.Reset()

	// 测试2: 密码太短
	fmt.Println("测试2: 密码太短")
	simulateInput(loginForm, action.ActionInputText, "admin")
	simulateInput(loginForm, action.ActionNavigateDown, nil)
	simulateInput(loginForm, action.ActionInputText, "12345")
	if err := loginForm.Submit(); err != nil {
		fmt.Printf("  验证失败 (预期): %v\n", err)
	}
	fmt.Println()

	// 重置表单
	loginForm.Reset()

	// 测试3: 空用户名
	fmt.Println("测试3: 空用户名")
	simulateInput(loginForm, action.ActionNavigateDown, nil) // 跳到密码
	simulateInput(loginForm, action.ActionInputText, "password123")
	if err := loginForm.Submit(); err != nil {
		fmt.Printf("  验证失败 (预期): %v\n", err)
	}
	fmt.Println()

	// 重置表单
	loginForm.Reset()

	// 测试4: 所有字段有效
	fmt.Println("测试4: 所有字段有效 (应该成功)")
	simulateInput(loginForm, action.ActionInputText, "admin")
	simulateInput(loginForm, action.ActionNavigateDown, nil)
	simulateInput(loginForm, action.ActionInputText, "password123")
	if err := loginForm.Submit(); err != nil {
		fmt.Printf("  验证失败: %v\n", err)
	} else {
		fmt.Println("  验证通过!")
	}
}

// createLoginForm 创建登录表单
func createLoginForm() *form.Form {
	f := form.NewForm()
	f.SetID("login-form")

	// 用户名字段
	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")

	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名:"
	usernameField.Input = usernameInput
	usernameField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("用户名不能为空")
			}
			return nil
		}, "必填验证"),
		validation.NewFuncValidator(func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("值类型错误")
			}
			if len([]rune(str)) < 3 {
				return fmt.Errorf("用户名至少3个字符")
			}
			return nil
		}, "长度验证"),
	}
	f.AddField(usernameField)

	// 密码字段
	passwordInput := input.NewTextInput()
	passwordInput.SetID("password-input")
	passwordInput.SetPassword(true)

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码:"
	passwordField.Input = passwordInput
	passwordField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("密码不能为空")
			}
			return nil
		}, "必填验证"),
		validation.NewFuncValidator(func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("值类型错误")
			}
			if len([]rune(str)) < 6 {
				return fmt.Errorf("密码至少6个字符")
			}
			return nil
		}, "长度验证"),
	}
	f.AddField(passwordField)

	// 设置提交回调
	f.SetOnSubmit(func(data map[string]interface{}) error {
		return nil
	})

	return f
}

// 模拟输入
func simulateInput(f *form.Form, actionType action.ActionType, payload interface{}) {
	a := action.NewAction(actionType)
	if payload != nil {
		a = a.WithPayload(payload)
	}
	f.HandleAction(*a)
}
