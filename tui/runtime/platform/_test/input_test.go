//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"

	"github.com/yaoapp/yao/tui/runtime/platform"
)

func main() {
	fmt.Println("=== 键盘输入测试 ===")
	fmt.Println("按任意键测试输入，按 Ctrl+C 退出")
	fmt.Println()

	input, err := platform.NewInputReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建输入读取器失败: %v\n", err)
		os.Exit(1)
	}

	events := make(chan platform.RawInput, 100)
	if err := input.Start(events); err != nil {
		fmt.Fprintf(os.Stderr, "启动输入读取器失败: %v\n", err)
		os.Exit(1)
	}
	defer input.Stop()

	fmt.Println("[就绪] 等待按键...")

	count := 0
	for raw := range events {
		count++
		fmt.Printf("[事件 %d] Type=%d", count, raw.Type)

		if raw.Type == platform.InputKeyPress {
			if raw.Special != platform.KeyUnknown {
				fmt.Printf(" Special=%d", raw.Special)
			}
			if raw.Key != 0 {
				fmt.Printf(" Key='%c' (0x%X)", raw.Key, raw.Key)
			}
			if raw.Modifiers != 0 {
				fmt.Printf(" Modifiers=0x%02X", raw.Modifiers)
			}
		}

		fmt.Println()

		if count >= 10 {
			fmt.Println("\n测试完成！")
			break
		}
	}
}
