//go:build !demo && !error_test && !visual_test && !interactive
// +build !demo,!error_test,!visual_test,!interactive

package main

import (
	"fmt"
	"strings"

	"github.com/yaoapp/yao/tui/framework/form"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/action"
)

// main 登录表单示例
func main() {
	fmt.Println("========================================")
	fmt.Println("     TUI Framework - 登录表单示例")
	fmt.Println("========================================")
	fmt.Println()

	// 创建登录表单
	loginForm := createLoginForm()

	// 模拟用户输入
	fmt.Println("模拟用户操作:")
	fmt.Println("1. 输入用户名: admin")
	simulateInput(loginForm, action.ActionInputText, "admin")
	fmt.Println("2. 按下 Tab 键切换到密码框")
	simulateInput(loginForm, action.ActionNavigateDown, nil)
	fmt.Println("3. 输入密码: ********")
	simulateInput(loginForm, action.ActionInputText, "password123")
	fmt.Println("4. 按下 Enter 键提交")

	// 模拟提交
	simulateSubmit(loginForm)
}

// createLoginForm 创建登录表单
func createLoginForm() *form.Form {
	f := form.NewForm()
	f.SetID("login-form")

	// 用户名字段
	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")
	usernameInput.SetPlaceholder("请输入用户名")

	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名:"
	usernameField.Input = usernameInput
	usernameField.HelpText = "请输入您的用户名"
	usernameField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("用户名不能为空")
			}
			return nil
		}, "用户名验证"),
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
	passwordInput.SetPlaceholder("请输入密码")

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码:"
	passwordField.Input = passwordInput
	passwordField.HelpText = "请输入您的密码"
	passwordField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("密码不能为空")
			}
			return nil
		}, "密码验证"),
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
		fmt.Println()
		fmt.Println("========================================")
		fmt.Println("           登录信息提交成功!")
		fmt.Println("========================================")
		fmt.Printf("用户名: %s\n", data["username"])
		fmt.Printf("密码:   %s\n", strings.Repeat("*", len([]rune(data["password"].(string)))))
		fmt.Println("========================================")
		return nil
	})

	// 设置取消回调
	f.SetOnCancel(func() {
		fmt.Println()
		fmt.Println("用户取消了登录")
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

// 模拟提交
func simulateSubmit(f *form.Form) {
	// 获取当前值
	values := f.GetValues()
	fmt.Println()
	fmt.Println("--- 当前表单内容 ---")
	fmt.Printf("用户名: %s\n", values["username"])
	fmt.Printf("密码:   %s\n", strings.Repeat("*", len([]rune(values["password"].(string)))))

	// 验证并提交
	if err := f.Submit(); err != nil {
		fmt.Println()
		fmt.Println("验证失败:", err.Error())
		return
	}

	fmt.Println()
	fmt.Println("验证通过，提交成功!")
}
