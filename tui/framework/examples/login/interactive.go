//go:build interactive
// +build interactive

package main

import (
	"fmt"
	"os"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/form"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// main 交互式登录表单示例 - 等待真实用户输入
func main() {
	fmt.Println("========================================")
	fmt.Println("  TUI Framework - 交互式登录表单")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("使用说明:")
	fmt.Println("  Tab/方向键  - 在字段间导航")
	fmt.Println("  Enter       - 提交表单")
	fmt.Println("  Esc         - 退出")
	fmt.Println()
	fmt.Println("调试模式:")
	fmt.Println("  设置环境变量 TUI_DEBUG=true 启用调试日志")
	fmt.Println("  设置环境变量 TUI_INPUT_DEBUG=1 启用TextInput调试")
	fmt.Println()
	fmt.Print("按 Enter 开始...")
	fmt.Scanln()

	// 创建应用
	app := framework.NewApp()

	// 启用调试模式（如果设置了环境变量）
	if os.Getenv("TUI_DEBUG") == "true" {
		app.SetDebugMode(true)
		fmt.Println("调试模式已启用")
	}

	// 创建登录表单
	loginForm := createLoginForm()

	// 设置为根组件
	app.SetRoot(loginForm)

	// 订阅键盘事件用于退出
	app.OnEvent(event.EventKeyPress, event.EventHandlerFunc(func(ev event.Event) bool {
		if keyEv, ok := ev.(*event.KeyEvent); ok {
			if keyEv.Special == event.KeyEscape {
				// 显示光标
				fmt.Print("\x1b[?25h")
				fmt.Println("\n用户取消登录")
				app.Quit()
				return true
			}
		}
		return false
	}))

	// 运行应用
	fmt.Println("启动交互式登录界面...")
	if err := app.Run(); err != nil {
		fmt.Printf("运行出错: %v\n", err)
	}
	// 确保换行
	fmt.Println()
}

// createLoginForm 创建登录表单
func createLoginForm() *form.Form {
	f := form.NewForm()
	f.SetID("login-form")

	// 标题样式
	f.SetLabelStyle(style.Style{}.Foreground(style.Cyan))

	// 用户名字段
	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")
	usernameInput.SetPlaceholder("请输入用户名")

	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名: *"
	usernameField.Input = usernameInput
	usernameField.HelpText = "至少3个字符"
	usernameField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("用户名不能为空")
			}
			return nil
		}, "必填"),
		validation.NewFuncValidator(func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("值类型错误")
			}
			if len([]rune(str)) < 3 {
				return fmt.Errorf("至少3个字符")
			}
			return nil
		}, "长度"),
	}
	f.AddField(usernameField)

	// 密码字段
	passwordInput := input.NewTextInput()
	passwordInput.SetID("password-input")
	passwordInput.SetPassword(true)
	passwordInput.SetPlaceholder("请输入密码")

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码: *"
	passwordField.Input = passwordInput
	passwordField.HelpText = "至少6个字符"
	passwordField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("密码不能为空")
			}
			return nil
		}, "必填"),
		validation.NewFuncValidator(func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("值类型错误")
			}
			if len([]rune(str)) < 6 {
				return fmt.Errorf("至少6个字符")
			}
			return nil
		}, "长度"),
	}
	f.AddField(passwordField)

	// 设置提交回调
	f.SetOnSubmit(func(data map[string]interface{}) error {
		fmt.Println("\n========================================")
		fmt.Println("           登录成功!")
		fmt.Println("========================================")
		fmt.Printf("用户名: %v\n", data["username"])
		fmt.Printf("密码:   ******\n")
		fmt.Println("========================================")

		// 退出应用
		return fmt.Errorf("exit") // 使用错误来退出循环
	})

	return f
}

// renderForm 渲染表单到控制台
func renderForm(f *form.Form) {
	// 获取表单尺寸
	width, height := f.Measure(50, 20)

	// 创建渲染缓冲区
	buf := paint.NewBuffer(width+4, height+6)

	// 绘制外边框
	drawDoubleBox(buf, 0, 0, width+4, height+6)

	// 绘制标题
	title := " 用户登录 "
	titleX := (width+4 - len([]rune(title))) / 2
	drawText(buf, titleX, 0, title, style.Style{}.Foreground(style.Yellow).Bold(true))

	// 绘制表单内容
	ctx := component.PaintContext{
		AvailableWidth:  width,
		AvailableHeight: height,
		X:               2,
		Y:               2,
	}

	f.Paint(ctx, buf)

	// 绘制底部提示
	helpText := " ↑↓导航 Enter提交 Esc取消 "
	helpX := (width+4 - len([]rune(helpText))) / 2
	helpY := height + 5
	drawText(buf, helpX, helpY, helpText, style.Style{}.Foreground(style.BrightBlack))

	// 输出渲染结果
	for y := 0; y < buf.Height; y++ {
		line := ""
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Char == 0 {
				line += " "
			} else {
				line += string(cell.Char)
			}
		}
		fmt.Println(line)
	}
}

// drawDoubleBox 绘制双线边框
func drawDoubleBox(buf *paint.Buffer, x, y, width, height int) {
	s := style.Style{} // 默认样式

	// 顶部边框
	buf.SetCell(x, y, '╔', s)
	for i := 1; i < width-1; i++ {
		buf.SetCell(x+i, y, '═', s)
	}
	buf.SetCell(x+width-1, y, '╗', s)

	// 两侧边框
	for i := 1; i < height-1; i++ {
		buf.SetCell(x, y+i, '║', s)
		buf.SetCell(x+width-1, y+i, '║', s)
	}

	// 底部边框
	buf.SetCell(x, y+height-1, '╚', s)
	for i := 1; i < width-1; i++ {
		buf.SetCell(x+i, y+height-1, '═', s)
	}
	buf.SetCell(x+width-1, y+height-1, '╝', s)
}

// drawText 绘制文本
func drawText(buf *paint.Buffer, x, y int, text string, s style.Style) {
	runes := []rune(text)
	for i, r := range runes {
		buf.SetCell(x+i, y, r, s)
	}
}
