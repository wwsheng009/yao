//go:build demo
// +build demo

package main

import (
	"fmt"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/display"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/form"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/interactive"
	"github.com/yaoapp/yao/tui/framework/layout"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/framework/validation"
)

// =============================================================================
// Demo Application - TUI Framework V3 Showcase
// =============================================================================

func main() {
	fmt.Println("==============================================")
	fmt.Println("   TUI Framework V3 - 组件演示程序")
	fmt.Println("==============================================")
	fmt.Println()

	// 演示基础组件
	demoBasicComponents()

	// 演示布局组件
	demoLayouts()

	// 演示表单组件
	demoForms()

	// 演示列表和表格
	demoListsAndTables()

	// 演示 AI Controller
	demoAIController()

	// 演示事件系统
	demoEventSystem()

	fmt.Println("\n==============================================")
	fmt.Println("   所有演示完成！")
	fmt.Println("==============================================")
}

// =============================================================================
// 基础组件演示
// =============================================================================

func demoBasicComponents() {
	fmt.Println(">>> 基础组件演示")

	// 1. Text 组件
	fmt.Println("\n--- Text 组件 ---")

	title := display.NewText("TUI Framework V3")
	title.SetStyle(style.Style{}.Foreground(style.Blue).Bold(true))
	fmt.Printf("标题: %s\n", title.GetContent())

	subtitle := display.NewText("新一代终端 UI 框架")
	fmt.Printf("副标题: %s\n", subtitle.GetContent())

	multiline := display.NewText("支持多行文本\n自动换行\n样式应用")
	fmt.Printf("多行文本:\n%s\n", multiline.GetContent())

	// 2. TextInput 组件
	fmt.Println("\n--- TextInput 组件 ---")

	username := input.NewTextInput()
	username.SetPlaceholder("请输入用户名")
	username.SetValue("demo_user")
	fmt.Printf("用户名: %s\n", username.GetValue())

	password := input.NewTextInput()
	password.SetPlaceholder("请输入密码")
	password.SetPassword(true)
	password.SetValue("secret123")
	fmt.Printf("密码: %s (长度=%d)\n", maskString(password.GetValue()), len(username.GetValue()))

	// 3. Button 组件
	fmt.Println("\n--- Button 组件 ---")

	submitBtn := interactive.NewButton("提交")
	fmt.Printf("按钮标签: %s\n", submitBtn.GetLabel())

	cancelBtn := interactive.NewButton("取消")
	cancelBtn.SetNormalStyle(style.Style{}.Foreground(style.Red))
	fmt.Printf("取消按钮: %s\n", cancelBtn.GetLabel())
}

// =============================================================================
// 布局组件演示
// =============================================================================

func demoLayouts() {
	fmt.Println("\n>>> 布局组件演示")

	// Box 容器
	fmt.Println("\n--- Box 容器 ---")

	box := layout.NewBox().WithBorder(true).WithBorderColor(style.Cyan).WithPadding(1)

	fmt.Printf("Box 边框: %v\n", box.GetBorder() != nil)
	fmt.Printf("Box 内边距: %+v\n", box.GetPadding())

	// Flex 布局
	fmt.Println("\n--- Flex 布局 ---")

	_ = layout.NewFlex(layout.Row)
	fmt.Printf("Flex 方向: Row\n")

	_ = layout.NewFlex(layout.Column)
	fmt.Printf("Flex 方向: Column\n")

	flexWithGap := layout.NewFlex(layout.Row).WithGap(2)
	fmt.Printf("Flex 间距: %d (已设置)\n", flexWithGap)
}

// =============================================================================
// 表单组件演示
// =============================================================================

func demoForms() {
	fmt.Println("\n>>> 表单组件演示")

	// 创建表单
	loginForm := form.NewForm()
	loginForm.SetLabelStyle(style.Style{}.Foreground(style.Cyan))
	loginForm.SetErrorStyle(style.Style{}.Foreground(style.Red))

	// 添加字段
	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名"
	usernameField.Placeholder = "请输入用户名"
	usernameField.Input = input.NewTextInput()
	usernameField.Validators = []validation.Validator{
		validation.Required(),
		validation.MinLength(3),
	}

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码"
	passwordField.Placeholder = "请输入密码"
	passwordField.Input = input.NewTextInput()
	passwordField.Input.(*input.TextInput).SetPassword(true)
	passwordField.Validators = []validation.Validator{
		validation.Required(),
		validation.MinLength(6),
	}

	emailField := form.NewFormField("email")
	emailField.Label = "邮箱"
	emailField.Placeholder = "请输入邮箱"
	emailField.Input = input.NewTextInput()
	emailField.Validators = []validation.Validator{
		validation.Required(),
		validation.Email(),
	}

	loginForm.AddField(usernameField)
	loginForm.AddField(passwordField)
	loginForm.AddField(emailField)

	// 设置提交回调
	loginForm.SetOnSubmit(func(data map[string]interface{}) error {
		fmt.Printf("\n✓ 表单提交成功!\n")
		fmt.Printf("  用户名: %v\n", data["username"])
		fmt.Printf("  密码: %v\n", data["password"])
		fmt.Printf("  邮箱: %v\n", data["email"])
		return nil
	})

	loginForm.SetOnCancel(func() {
		fmt.Println("\n✗ 表单已取消")
	})

	// 测试表单
	fmt.Println("\n表单字段:")
	for _, field := range loginForm.GetFields() {
		fmt.Printf("  - %s: %s\n", field.Label, field.Placeholder)
		if len(field.Validators) > 0 {
			fmt.Printf("    验证器: %d 个\n", len(field.Validators))
		}
	}

	// 验证测试
	fmt.Println("\n验证测试:")
	valid := loginForm.IsValid()
	fmt.Printf("  表单有效: %v (初始状态)\n", valid)

	// 设置值并测试验证
	_ = loginForm.SetValue("username", "ab") // 太短
	_ = loginForm.SetValue("password", "123") // 太短
	_ = loginForm.SetValue("email", "invalid") // 无效邮箱

	field, _ := loginForm.GetField("username")
	err := field.Validate()
	fmt.Printf("  用户名验证: %v\n", err)

	field, _ = loginForm.GetField("email")
	err = field.Validate()
	fmt.Printf("  邮箱验证: %v\n", err)

	// 有效数据
	_ = loginForm.SetValue("username", "demo_user")
	_ = loginForm.SetValue("password", "password123")
	_ = loginForm.SetValue("email", "user@example.com")

	valid = loginForm.IsValid()
	fmt.Printf("  表单有效: %v (有效数据)\n", valid)
}

// =============================================================================
// 列表和表格演示
// =============================================================================

func demoListsAndTables() {
	fmt.Println("\n>>> 列表和表格演示")

	// List 组件
	fmt.Println("\n--- List 组件 ---")

	items := []string{
		"项目 1: 学习 Go 语言",
		"项目 2: 开发 TUI 框架",
		"项目 3: 编写文档",
		"项目 4: 单元测试",
		"项目 5: 发布版本",
	}
	_ = display.NewListStrings(items)

	fmt.Printf("列表项数量: %d\n", len(items))

	// Table 组件
	fmt.Println("\n--- Table 组件 ---")

	table := display.NewTable([]display.TableColumn{
		{Title: "ID", Width: 10},
		{Title: "名称", Width: 30},
		{Title: "状态", Width: 15},
		{Title: "创建时间", Width: 20},
	})

	table.SetRows([][]string{
		{"1", "用户管理", "完成", "2024-01-20"},
		{"2", "订单系统", "进行中", "2024-01-21"},
		{"3", "数据同步", "待开始", "2024-01-22"},
	})

	fmt.Printf("表格列数: %d\n", table.GetColumnCount())
	fmt.Printf("表格行数: %d\n", table.GetRowCount())
}

// =============================================================================
// AI Controller 演示
// =============================================================================

func demoAIController() {
	fmt.Println("\n>>> AI Controller 演示")

	// 模拟组件状态
	componentStates := map[string]map[string]interface{}{
		"username": {
			"value":    "testuser",
			"focused":  true,
			"disabled": false,
		},
		"password": {
			"value":    "",
			"focused":  false,
			"disabled": false,
		},
		"submit": {
			"disabled": true,
		},
	}

	fmt.Println("组件状态查询:")
	for id, state := range componentStates {
		fmt.Printf("  %s:\n", id)
		for key, val := range state {
			fmt.Printf("    %s: %v\n", key, val)
		}
	}

	// 模拟 AI 操作序列
	fmt.Println("\nAI 操作序列演示:")
	fmt.Println("  1. 检查 UI 状态")
	fmt.Println("  2. 查找可交互组件")
	fmt.Println("  3. 输入用户名")
	fmt.Println("  4. 输入密码")
	fmt.Println("   5. 点击提交按钮")
	fmt.Println("   6. 等待提交完成")

	fmt.Println("\n选择器语法:")
	fmt.Println("  - ID 选择器:   #username")
	fmt.Println("  - 类型选择器: .TextInput")
	fmt.Println("  - 属性选择器: [focused=\"true\"]")
}

// =============================================================================
// 事件系统演示
// =============================================================================

func demoEventSystem() {
	fmt.Println("\n>>> 事件系统演示")

	// 创建各种事件
	fmt.Println("--- 事件类型 ---")

	// 键盘事件
	keyEvent := event.NewKeyEvent('a')
	fmt.Printf("键盘事件: %c\n", keyEvent.Key)

	specialEvent := event.NewSpecialKeyEvent(event.KeyEnter)
	fmt.Printf("特殊键事件: Enter (类型: %d)\n", specialEvent.Type())

	// 事件属性
	fmt.Println("\n--- 事件属性 ---")

	specialEvent = event.NewSpecialKeyEvent(event.KeyTab)
	fmt.Printf("Tab 键 - 类型: 导航键\n")

	specialEvent = event.NewSpecialKeyEvent(event.KeyEscape)
	fmt.Printf("Esc 键 - 类型: 系统键\n")

	specialEvent = event.NewSpecialKeyEvent(event.KeyF1)
	fmt.Printf("F1 键 - 类型: 功能键\n")

	// 事件处理流程
	fmt.Println("\n--- 事件处理流程 ---")
	fmt.Println("  1. Platform → RawInput")
	fmt.Println("  2. Runtime → KeyMap → Action")
	fmt.Println("  3. Component → HandleAction()")
	fmt.Println("  4. State → Update & Render")

	// Action 类型示例
	fmt.Println("\n--- Action 类型 ---")
	fmt.Println("  导航: navigate_up, navigate_down, navigate_next, navigate_prev")
	fmt.Println("  编辑: input_char, delete_char, backspace")
	fmt.Println("  表单: submit, cancel, validate")
	fmt.Println("  视图: scroll_up, scroll_down, zoom_in, zoom_out")
	fmt.Println("  系统: quit, copy, paste, undo, redo")
}

// =============================================================================
// 辅助函数
// =============================================================================

// maskString 掩码字符串
func maskString(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + "***" + s[len(s)-1:]
}

// =============================================================================
// 附加示例：完整的 UI 应用场景
// =============================================================================

// demoLoginForm 演示登录表单完整流程
func demoLoginForm() {
	fmt.Println("\n>>> 登录表单完整流程")

	loginForm := form.NewForm()

	// 添加字段
	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名"
	usernameField.Placeholder = "请输入用户名"
	usernameField.Input = input.NewTextInput()
	usernameField.Validators = []validation.Validator{
		validation.Required(),
		validation.MinLength(3),
		validation.MaxLength(20),
	}

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码"
	passwordField.Placeholder = "请输入密码"
	passwordField.HelpText = "密码至少6个字符"
	passwordField.Input = input.NewTextInput()
	passwordField.Input.(*input.TextInput).SetPassword(true)
	passwordField.Validators = []validation.Validator{
		validation.Required(),
		validation.MinLength(6),
	}

	rememberField := form.NewFormField("remember")
	rememberField.Label = "记住我"
	rememberField.Input = nil // Checkbox

	loginForm.AddField(usernameField)
	loginForm.AddField(passwordField)
	loginForm.AddField(rememberField)

	// 设置回调
	loginForm.SetOnSubmit(func(data map[string]interface{}) error {
		fmt.Println("\n✓ 登录成功!")
		fmt.Printf("  用户名: %s\n", data["username"])
		fmt.Printf("  记住我: %v\n", data["remember"])
		return nil
	})

	// 模拟用户交互
	fmt.Println("\n模拟用户交互:")
	fmt.Println("  [1] 输入用户名: admin")
	_ = loginForm.SetValue("username", "admin")

	fmt.Println("  [2] 输入密码: 123456")
	_ = loginForm.SetValue("password", "123456")

	fmt.Println("  [3] 切换记住我")
	_ = loginForm.SetValue("remember", true)

	fmt.Println("  [4] 验证表单...")
	field, _ := loginForm.GetField("password")
	err := field.Validate()
	if err != nil {
		fmt.Printf("  验证失败: %v\n", err)
		// 显示字段错误
		for _, field := range loginForm.GetFields() {
			if field.Error != nil {
				fmt.Printf("    %s: %v\n", field.Name, field.Error)
			}
		}
	} else {
		fmt.Println("  验证通过!")
	}
}

// demoDataTable 演示数据表格完整功能
func demoDataTable() {
	fmt.Println("\n>>> 数据表格完整功能")

	table := display.NewTable([]display.TableColumn{
		{Title: "ID", Width: 8, Align: component.AlignLeft},
		{Title: "用户名", Width: 20, Align: component.AlignLeft},
		{Title: "邮箱", Width: 30, Align: component.AlignLeft},
		{Title: "角色", Width: 12, Align: component.AlignLeft},
		{Title: "状态", Width: 10, Align: component.AlignCenter, Sortable: true},
	})

	// 添加数据
	table.SetRows([][]string{
		{"1", "admin", "admin@example.com", "管理员", "active"},
		{"2", "user1", "user1@example.com", "用户", "active"},
		{"3", "user2", "user2@example.com", "用户", "inactive"},
		{"4", "guest", "guest@example.com", "访客", "active"},
		{"5", "test", "test@example.com", "测试", "disabled"},
	})

	fmt.Printf("表格列数: %d\n", table.GetColumnCount())
	fmt.Printf("表格行数: %d\n", table.GetRowCount())
}

// demoInteractiveButtons 演示交互按钮
func demoInteractiveButtons() {
	fmt.Println("\n>>> 交互按钮演示")

	// 创建按钮组
	buttons := []*interactive.Button{
		interactive.NewButton("确认").SetNormalStyle(style.Style{}.Foreground(style.Green)),
		interactive.NewButton("取消").SetNormalStyle(style.Style{}.Foreground(style.Red)),
		interactive.NewButton("帮助").SetNormalStyle(style.Style{}.Foreground(style.Blue)),
	}

	fmt.Println("按钮组:")
	for i, btn := range buttons {
		fmt.Printf("  [%d] %s\n", i+1, btn.GetLabel())
	}

	// 快捷键
	fmt.Println("\n快捷键支持:")
	fmt.Println("  Enter - 确认按钮")
	fmt.Println("  Esc - 取消按钮")
	fmt.Println("  F1   - 帮助按钮")
}

// demoStateManagement 演示状态管理
func demoStateManagement() {
	fmt.Println("\n>>> 状态管理演示")

	// 创建状态快照
	fmt.Println("--- 状态快照 ---")
	fmt.Println("  状态快照包含:")
	fmt.Println("    - FocusPath: 当前焦点路径")
	fmt.Println("    - Components: 所有组件状态")
	fmt.Println("    - Modals: 模态窗口栈")
	fmt.Println("    - DirtyRegion: 脏区域标记")

	// Undo/Redo
	fmt.Println("\n--- Undo/Redo ---")
	fmt.Println("  支持操作撤销和重做")
	fmt.Println("    - Undo: 撤销上一步操作")
	fmt.Println("    - Redo: 重做已撤销的操作")
	fmt.Println("    - 历史记录: 最多保存 100 个状态")

	// 状态监听
	fmt.Println("\n--- 状态监听 ---")
	fmt.Println("  可以监听状态变化:")
	fmt.Println("    - OnStateChange: 状态变化时触发")
	fmt.Println("    - OnFocusChange: 焦点变化时触发")
	fmt.Println("    - OnDirty: 脏区域标记时触发")
}

// RunAllDemos 运行所有演示
func RunAllDemos() {
	demos := []struct {
		name string
		fn   func()
	}{
		{"基础组件", demoBasicComponents},
		{"布局组件", demoLayouts},
		{"表单组件", demoForms},
		{"列表和表格", demoListsAndTables},
		{"AI Controller", demoAIController},
		{"事件系统", demoEventSystem},
		{"登录表单", demoLoginForm},
		{"数据表格", demoDataTable},
		{"交互按钮", demoInteractiveButtons},
		{"状态管理", demoStateManagement},
	}

	startTime := time.Now()

	for _, demo := range demos {
		start := time.Now()
		demo.fn()
		duration := time.Since(start)
		fmt.Printf("\n[✓] %s 完成 (耗时: %v)\n", demo.name, duration)
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\n==============================================\n")
	fmt.Printf("   总耗时: %v\n", totalDuration)
	fmt.Println("==============================================")
}
