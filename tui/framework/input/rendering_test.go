package input

import (
	"fmt"
	"testing"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// TestTextInputRendering 详细测试 TextInput 渲染逻辑
func TestTextInputRendering(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		cursorPos int
	}{
		{"空输入", "", 0},
		{"单字符", "a", 1},
		{"两个字符", "ab", 2},
		{"四个字符", "test", 4},
		{"光标在中间", "test", 2},
		{"光标在开头", "test", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txt := NewTextInput()
			txt.SetValue(tt.value)
			txt.SetCursor(tt.cursorPos)
			txt.OnFocus()

			buf := paint.NewBuffer(80, 24)
			ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

			txt.Paint(ctx, buf)

			// 分析渲染结果
			fmt.Printf("\n=== %s ===\n", tt.name)
			fmt.Printf("值: '%s', 光标: %d\n", tt.value, tt.cursorPos)

			// 打印前 20 个字符
			fmt.Print("渲染: ")
			for x := 0; x < 20; x++ {
				cell := buf.Cells[0][x]
				if cell.Char == 0 {
					fmt.Print(".")
				} else {
					fmt.Printf("%c", cell.Char)
				}
			}
			fmt.Println()

			// 详细分析每个位置
			fmt.Println("位置分析:")
			for x := 0; x < len(tt.value)+3; x++ {
				cell := buf.Cells[0][x]
				char := " "
				if cell.Char != 0 {
					char = fmt.Sprintf("%c", cell.Char)
				}
				fmt.Printf("  x=%d: '%s' reverse=%v\n", x, char, cell.Style.IsReverse())
			}

			// 验证左边框
			if buf.Cells[0][0].Char != '[' {
				t.Errorf("位置 0 应该是 '['")
			}

			// 验证右边框位置（左边框 + 文字长度）
			expectedRightBracket := 1 + len(tt.value) // 左边框后跟文字，然后是右边框
			if buf.Cells[0][expectedRightBracket].Char != ']' {
				t.Errorf("位置 %d 应该是 ']', 找到 '%c'", expectedRightBracket, buf.Cells[0][expectedRightBracket].Char)
			}

			// 验证文字内容
			for i, ch := range tt.value {
				pos := 1 + i // 左边框后第 i 个位置
				if buf.Cells[0][pos].Char != ch {
					t.Errorf("位置 %d 应该是 '%c'", pos, ch)
				}
			}

			// 验证光标位置（应该在 cursorPos + 1 位置有反色）
			cursorScreenPos := 1 + tt.cursorPos
			if buf.Cells[0][cursorScreenPos].Style.IsReverse() {
				fmt.Printf("✓ 光标在正确位置: %d\n", cursorScreenPos)
			} else {
				t.Errorf("光标应该在位置 %d 有反色样式", cursorScreenPos)
			}

			// 验证只有一个反色单元格
			reverseCount := 0
			for x := 0; x < 30; x++ {
				if buf.Cells[0][x].Style.IsReverse() {
					reverseCount++
				}
			}
			if reverseCount != 1 {
				t.Errorf("应该只有 1 个反色单元格，但有 %d 个", reverseCount)
			}
		})
	}
}

// TestTextInputWithOffset 测试有偏移时的渲染
func TestTextInputWithOffset(t *testing.T) {
	txt := NewTextInput()
	txt.SetValue("hello")
	txt.SetCursor(5) // 光标在末尾
	txt.OnFocus()

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 10, 5, 80, 24) // 偏移 10, 5

	txt.Paint(ctx, buf)

	fmt.Printf("\n=== 有偏移渲染 ===\n")
	fmt.Printf("ctx.X=%d, ctx.Y=%d\n", ctx.X, ctx.Y)
	fmt.Printf("值: 'hello', 光标: 5\n")

	// 打印第 5 行的内容
	fmt.Print("第 5 行 (x=0-25): ")
	for x := 0; x < 25; x++ {
		cell := buf.Cells[5][x]
		if cell.Char == 0 {
			fmt.Print(".")
		} else {
			fmt.Printf("%c", cell.Char)
		}
	}
	fmt.Println()

	// 验证：左边框应该在 10
	if buf.Cells[5][10].Char != '[' {
		t.Errorf("左边框应该在 x=10, 找到 '%c'", buf.Cells[5][10].Char)
	} else {
		fmt.Printf("✓ 左边框在 x=10\n")
	}

	// 验证：'h' 应该在 11
	if buf.Cells[5][11].Char != 'h' {
		t.Errorf("'h' 应该在 x=11, 找到 '%c'", buf.Cells[5][11].Char)
	} else {
		fmt.Printf("✓ 'h' 在 x=11\n")
	}

	// 验证：光标应该在 10+1+5=16
	cursorPos := 16
	if buf.Cells[5][cursorPos].Style.IsReverse() {
		fmt.Printf("✓ 光标在 x=%d\n", cursorPos)
	} else {
		t.Errorf("光标应该在 x=%d", cursorPos)
	}
}

// TestTextInputTypingSequence 测试连续输入时的渲染
func TestTextInputTypingSequence(t *testing.T) {
	txt := NewTextInput()
	txt.OnFocus()

	inputs := []rune{'h', 'e', 'l', 'l', 'o'}
	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	fmt.Printf("\n=== 连续输入测试 ===\n")

	for i, ch := range inputs {
		// 模拟输入
		txt.HandleAction(*action.NewAction(action.ActionInputChar).WithPayload(ch))

		// 清空 buffer 重新渲染
		buf = paint.NewBuffer(80, 24)
		txt.Paint(ctx, buf)

		expectedValue := string(inputs[:i+1])
		actualValue := txt.GetValue()

		fmt.Printf("输入 '%c': 值='%s', 光标=%d\n", ch, actualValue, txt.GetCursor())

		// 验证值
		if actualValue != expectedValue {
			t.Errorf("值错误: 预期 '%s', 得到 '%s'", expectedValue, actualValue)
		}

		// 验证光标位置
		expectedCursor := i + 1
		if txt.GetCursor() != expectedCursor {
			t.Errorf("光标位置错误: 预期 %d, 得到 %d", expectedCursor, txt.GetCursor())
		}

		// 验证光标有反色
		cursorScreenPos := 1 + txt.GetCursor()
		if !buf.Cells[0][cursorScreenPos].Style.IsReverse() {
			t.Errorf("输入 '%c' 后光标应该在位置 %d", ch, cursorScreenPos)
		}

		// 打印当前渲染
		fmt.Print("  渲染: ")
		for x := 0; x < 15; x++ {
			cell := buf.Cells[0][x]
			if cell.Char == 0 {
				fmt.Print(".")
			} else {
				fmt.Printf("%c", cell.Char)
			}
		}
		fmt.Println()
	}

	fmt.Printf("✓ 连续输入测试通过\n")
}

// TestTextInputWideChar 测试宽字符处理
func TestTextInputWideChar(t *testing.T) {
	txt := NewTextInput()
	txt.SetValue("你好") // 两个中文字符
	txt.SetCursor(2)
	txt.OnFocus()

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	txt.Paint(ctx, buf)

	fmt.Printf("\n=== 宽字符测试 ===\n")

	// 打印渲染结果
	fmt.Print("渲染: ")
	for x := 0; x < 15; x++ {
		cell := buf.Cells[0][x]
		if cell.Char == 0 {
			fmt.Print(".")
		} else {
			fmt.Printf("%c", cell.Char)
		}
	}
	fmt.Println()

	// 验证内容 - 注意：中文字符在 rune 数组中占 1 个元素
	// 渲染时也只占 1 个字符位置（尽管终端显示时可能占 2 列）
	runes := []rune("你好")
	for i, ch := range runes {
		pos := 1 + i // 左边框后第 i 个位置
		if buf.Cells[0][pos].Char != ch {
			t.Errorf("位置 %d 应该是 '%c'", pos, ch)
		} else {
			fmt.Printf("✓ 位置 %d: '%c'\n", pos, ch)
		}
	}

	// 验证右边框 - 应该紧跟在最后一个字符后面
	rightBracketPos := 1 + len(runes) // 位置 3
	if buf.Cells[0][rightBracketPos].Char != ']' {
		t.Errorf("位置 %d 应该是 ']', 找到 '%c'", rightBracketPos, buf.Cells[0][rightBracketPos].Char)
	} else {
		fmt.Printf("✓ 右边框在位置 %d\n", rightBracketPos)
	}

	// 验证光标 - cursor=2 表示在第 3 个字符位置（0-你, 1-好, 2-末尾）
	cursorPos := 1 + 2
	if buf.Cells[0][cursorPos].Style.IsReverse() {
		fmt.Printf("✓ 光标在位置 %d\n", cursorPos)
	} else {
		t.Errorf("光标应该在位置 %d", cursorPos)
	}
}

// TestBackspaceRendering 测试退格时的渲染
func TestBackspaceRendering(t *testing.T) {
	txt := NewTextInput()
	txt.SetValue("test")
	txt.SetCursor(4)
	txt.OnFocus()

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	// 初始渲染
	txt.Paint(ctx, buf)

	fmt.Printf("\n=== 退格测试 ===\n")
	fmt.Print("初始: ")
	for x := 0; x < 10; x++ {
		cell := buf.Cells[0][x]
		if cell.Char == 0 {
			fmt.Print(".")
		} else {
			fmt.Printf("%c", cell.Char)
		}
	}
	fmt.Println()

	// 执行退格
	txt.HandleAction(*action.NewAction(action.ActionBackspace))

	// 重新渲染
	buf = paint.NewBuffer(80, 24)
	txt.Paint(ctx, buf)

	fmt.Print("退格后: ")
	for x := 0; x < 10; x++ {
		cell := buf.Cells[0][x]
		if cell.Char == 0 {
			fmt.Print(".")
		} else {
			fmt.Printf("%c", cell.Char)
		}
	}
	fmt.Println()

	if txt.GetValue() != "tes" {
		t.Errorf("退格后值应该是 'tes', 得到 '%s'", txt.GetValue())
	}

	if txt.GetCursor() != 3 {
		t.Errorf("退格后光标应该是 3, 得到 %d", txt.GetCursor())
	}

	// 验证光标在正确位置
	cursorScreenPos := 1 + 3
	if !buf.Cells[0][cursorScreenPos].Style.IsReverse() {
		t.Errorf("退格后光标应该在位置 %d", cursorScreenPos)
	}

	fmt.Printf("✓ 退格测试通过\n")
}
