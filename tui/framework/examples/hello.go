package main

import (
	"fmt"

	"github.com/yaoapp/yao/tui/framework/display"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
)

// main 简单示例
func main() {
	// 创建标题
	title := display.NewTextV3("TUI Framework V3")
	title.SetStyle(style.Style{}.Foreground(style.Blue))

	// 创建说明文本
	description := display.NewTextV3(
		"这是一个 TUI 框架示例\n" +
			"按 'q' 或 ESC 退出",
	)

	// 打印渲染结果
	fmt.Println("Title:", title.GetContent())
	fmt.Println("Description:", description.GetContent())

	// 演示事件处理
	fmt.Println("\n--- 事件处理演示 ---")

	// 创建键盘事件
	upEvent := event.NewSpecialKeyEvent(event.KeyUp)
	fmt.Printf("事件类型: %d\n", upEvent.Type())
	fmt.Printf("特殊键: %d\n", upEvent.Special)

	// Vim 风格键
	kEvent := event.NewSpecialKeyEvent(event.KeyK)
	fmt.Printf("Vim K 键: %d\n", kEvent.Special)

	jEvent := event.NewSpecialKeyEvent(event.KeyJ)
	fmt.Printf("Vim J 键: %d\n", jEvent.Special)
}
