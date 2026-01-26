//go:build interactive_input
// +build interactive_input

package main

import (
	"fmt"

	"github.com/yaoapp/yao/tui/framework"
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/runtime/paint"
	"github.com/yaoapp/yao/tui/framework/style"
)

// SimpleInputBox 简单的双输入框容器
type SimpleInputBox struct {
	*component.BaseComponent
	*component.StateHolder

	nameInput  *input.TextInput
	emailInput *input.TextInput
}

// NewSimpleInputBox 创建简单输入框容器
func NewSimpleInputBox() *SimpleInputBox {
	nameInput := input.NewTextInput()
	nameInput.SetID("name-input")
	nameInput.SetPlaceholder("请输入名字")

	emailInput := input.NewTextInput()
	emailInput.SetID("email-input")
	emailInput.SetPlaceholder("请输入邮箱")

	return &SimpleInputBox{
		BaseComponent: component.NewBaseComponent("simple-input-box"),
		StateHolder:   component.NewStateHolder(),
		nameInput:     nameInput,
		emailInput:    emailInput,
	}
}

// GetNameValues 获取输入值
func (b *SimpleInputBox) GetNameValues() (string, string) {
	return b.nameInput.GetValue(), b.emailInput.GetValue()
}

// Measure 测量理想尺寸
func (b *SimpleInputBox) Measure(maxWidth, maxHeight int) (width, height int) {
	return 40, 6
}

// Paint 绘制组件
func (b *SimpleInputBox) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !b.IsVisible() {
		return
	}

	x := ctx.X
	y := ctx.Y
	width := ctx.AvailableWidth

	// 绘制标签和输入框
	labelStyle := style.Style{}.Foreground(style.Cyan)

	// 名字标签
	buf.SetCell(x, y, []rune("名字: ")[0], labelStyle)
	buf.SetCell(x+1, y, []rune("名字: ")[1], labelStyle)
	buf.SetCell(x+2, y, []rune("名字: ")[2], labelStyle)
	buf.SetCell(x+3, y, []rune("名字: ")[3], labelStyle)

	// 名字输入框
	inputCtx := component.PaintContext{
		AvailableWidth:  width - 6,
		AvailableHeight: 1,
		X:                x + 6,
		Y:                y,
	}
	b.nameInput.Paint(inputCtx, buf)

	// 邮箱标签
	buf.SetCell(x, y+2, []rune("邮箱: ")[0], labelStyle)
	buf.SetCell(x+1, y+2, []rune("邮箱: ")[1], labelStyle)
	buf.SetCell(x+2, y+2, []rune("邮箱: ")[2], labelStyle)
	buf.SetCell(x+3, y+2, []rune("邮箱: ")[3], labelStyle)

	// 邮箱输入框
	emailInputCtx := component.PaintContext{
		AvailableWidth:  width - 6,
		AvailableHeight: 1,
		X:                x + 6,
		Y:                y + 2,
	}
	b.emailInput.Paint(emailInputCtx, buf)

	// 底部提示
	helpStyle := style.Style{}.Foreground(style.BrightBlack)
	helpText := " Tab导航 Enter提交 Esc取消 "
	helpRunes := []rune(helpText)
	helpX := x + (width - len(helpRunes)) / 2
	for i, r := range helpRunes {
		buf.SetCell(helpX+i, y+5, r, helpStyle)
	}
}

// HandleEvent 处理事件
func (b *SimpleInputBox) HandleEvent(ev event.Event) bool {
	// 转发事件到当前焦点的输入框
	if b.nameInput.IsFocused() {
		if b.nameInput.HandleEvent(ev) {
			return true
		}
	} else if b.emailInput.IsFocused() {
		if b.emailInput.HandleEvent(ev) {
			return true
		}
	}

	// 处理 Tab 键切换焦点
	if keyEv, ok := ev.(*event.KeyEvent); ok {
		if keyEv.Special == event.KeyTab {
			if b.nameInput.IsFocused() {
				b.nameInput.OnBlur()
				b.emailInput.OnFocus()
				return true
			} else if b.emailInput.IsFocused() {
				b.emailInput.OnBlur()
				b.nameInput.OnFocus()
				return true
			}
		}
	}

	return false
}

// main 简单的双输入框示例
func main() {
	fmt.Println("========================================")
	fmt.Println("  TUI Framework - 简单输入框示例")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("使用说明:")
	fmt.Println("  Tab/方向键  - 在字段间导航")
	fmt.Println("  Enter       - 提交表单")
	fmt.Println("  Esc         - 退出")
	fmt.Println()
	fmt.Print("按 Enter 开始...")
	fmt.Scanln()

	// 创建应用
	app := framework.NewApp()

	// 创建简单输入框容器
	inputBox := NewSimpleInputBox()
	inputBox.SetID("input-box")
	inputBox.nameInput.OnFocus()

	// 设置为根组件
	app.SetRoot(inputBox)

	// 订阅键盘事件用于退出和提交
	app.OnEvent(event.EventKeyPress, event.EventHandlerFunc(func(ev event.Event) bool {
		if keyEv, ok := ev.(*event.KeyEvent); ok {
			if keyEv.Special == event.KeyEscape {
				fmt.Print("\x1b[?25h")
				fmt.Println("\n用户取消")
				app.Quit()
				return true
			}
			if keyEv.Special == event.KeyEnter {
				fmt.Println("\n========================================")
				fmt.Println("           提交成功!")
				fmt.Println("========================================")
				name, email := inputBox.GetNameValues()
				fmt.Printf("名字: %s\n", name)
				fmt.Printf("邮箱: %s\n", email)
				fmt.Println("========================================")
				app.Quit()
				return true
			}
		}
		return false
	}))

	// 运行应用
	fmt.Println("启动界面...")
	if err := app.Run(); err != nil {
		fmt.Printf("运行出错: %v\n", err)
	}
	fmt.Println()
}
