package main

import (
	"fmt"

	"github.com/yaoapp/yao/tui/framework"
	"github.com/yaoapp/yao/tui/framework/display"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/interactive"
	"github.com/yaoapp/yao/tui/framework/layout"
)

// main 示例主函数
func main() {
	app := framework.NewApp()

	// 创建标题
	title := display.NewText("TUI Framework Demo").
		WithStyle(style.NewBuilder().
			Bold().
			Foreground(style.Blue).
			Build())

	// 创建说明文本
	description := display.NewText(
		"这是一个独立于 Bubble Tea 的 TUI 框架\n" +
			"按 'q' 或 ESC 退出\n" +
			"使用方向键操作",
	)

	// 创建输入框
	nameInput := input.NewTextInputPlaceholder("请输入姓名...")

	// 创建按钮
	quitButton := interactive.NewButtonWithAction("退出", app.Quit)

	// 使用 Flex 垂直布局
	root := layout.NewColumn(layout.Column).
		WithGap(1).
		WithChildren(
			title,
			description,
			nameInput,
			layout.NewRow().WithChildren(
				interactive.NewButton("确定"),
				quitButton,
			),
		)

	// 设置根组件
	app.SetRoot(root)

	// 注册退出快捷键
	app.OnKey('q', app.Quit)
	app.OnKey(event.KeyEscape, app.Quit)

	// 显示欢迎信息
	fmt.Println("TUI Framework 启动...")
	fmt.Println("按回车键开始")

	// 运行应用
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// 事件类型
type event struct{}

// KeyEscape 常量
const event KeyEscape = 27

// style 包引用
type style struct {
	// 导出样式相关
}

// NewStyle 创建样式
func NewStyle() *style {
	return &style{}
}

// style 类型
type style struct{}

// NewBuilder 创建样式构建器
func (s *style) NewBuilder() *style {
	return s
}

// Bold 设置粗体
func (s *style) Bold() *style {
	return s
}

// Foreground 设置前景色
func (s *style) Foreground(c Color) *style {
	return s
}

// Color 颜色类型
type Color string

const (
	Blue    Color = "blue"
	Red     Color = "red"
	Green   Color = "green"
	Yellow  Color = "yellow"
	Black   Color = "black"
	White   Color = "white"
	BrightBlack Color = "bright-black"
)
