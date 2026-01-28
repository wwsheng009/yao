package framework

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yaoapp/yao/tui/runtime/paint"
	"github.com/yaoapp/yao/tui/runtime/style"
)

// =============================================================================
// 终端输出优化 - Inline Diff
// =============================================================================
// 该模块负责比较两个 paint.Buffer 并生成优化的 ANSI 输出序列
// 与 runtime/paint/dirty.go 不同，这里关注终端输出层面的优化（光标移动）

// CellChange 表示一个单元格的变化
type CellChange struct {
	X, Y  int
	Force bool // 是否强制输出（用于光标位置刷新）
}

// BufferDiffResult 表示两个 buffer 比较的结果
type BufferDiffResult struct {
	Changes        []CellChange
	CursorX        int // 当前光标位置（反转样式的单元格）
	CursorY        int
	HasChanges     bool
}

// CompareBuffers 比较新旧 buffer，返回变化列表
// 这是输出层优化的核心，与渲染层的 DirtyTracker 不同
func CompareBuffers(newBuf *paint.Buffer, prevBuf [][]paint.Cell, lastCursorX, lastCursorY int) BufferDiffResult {
	result := BufferDiffResult{
		Changes:    make([]CellChange, 0),
		CursorX:    -1,
		CursorY:    -1,
		HasChanges: false,
	}

	// 扫描缓冲区，查找光标位置（有反转样式的单元格）
	// 同时收集所有变化的单元格
	for y := 0; y < newBuf.Height; y++ {
		for x := 0; x < newBuf.Width; x++ {
			newCell := newBuf.Cells[y][x]
			oldCell := prevBuf[y][x]

			// 检测光标位置（有反转样式的单元格）
			if newCell.Style.IsReverse() {
				result.CursorX = x
				result.CursorY = y
			}

			// 检查单元格是否改变
			cellChanged := newCell.Char != oldCell.Char || newCell.Style != oldCell.Style

			if cellChanged {
				result.Changes = append(result.Changes, CellChange{X: x, Y: y, Force: false})
			}
		}
	}

	// 如果光标位置改变，强制刷新新旧光标位置
	if result.CursorX != lastCursorX || result.CursorY != lastCursorY {
		// 旧光标位置需要刷新（清除反转样式）
		if lastCursorX >= 0 && lastCursorY >= 0 {
			result.Changes = append(result.Changes, CellChange{X: lastCursorX, Y: lastCursorY, Force: true})
		}
		// 新光标位置需要刷新（确保反转样式生效）
		if result.CursorX >= 0 && result.CursorY >= 0 {
			result.Changes = append(result.Changes, CellChange{X: result.CursorX, Y: result.CursorY, Force: true})
		}
	}

	result.HasChanges = len(result.Changes) > 0
	return result
}

// SortChanges 按位置排序变化（从上到下，从左到右）以优化输出
func SortChanges(changes []CellChange) {
	// 使用简单的冒泡排序（因为变化数量通常不大）
	for i := 0; i < len(changes)-1; i++ {
		for j := i + 1; j < len(changes); j++ {
			if changes[j].Y < changes[i].Y || (changes[j].Y == changes[i].Y && changes[j].X < changes[i].X) {
				changes[i], changes[j] = changes[j], changes[i]
			}
		}
	}
}

// FormatChangesAsANSI 将变化列表格式化为 ANSI 输出序列
// 返回输出字符串和新的光标位置
func FormatChangesAsANSI(buf *paint.Buffer, diffResult BufferDiffResult, firstRender bool) string {
	var output bytes.Buffer

	// 首次渲染时清屏
	if firstRender {
		output.WriteString("\x1b[2J") // 清屏
	}
	// 隐藏终端光标
	output.WriteString("\x1b[?25l")

	// 如果没有变化，返回空
	if !diffResult.HasChanges {
		// 仍然需要重置样式并移动光标到末尾
		output.WriteString("\x1b[0m")
		output.WriteString(fmt.Sprintf("\x1b[%d;%dH", buf.Height, 1))
		return output.String()
	}

	// 跟踪当前样式和位置
	var currentStyle style.Style
	var lastX, lastY int = 0, 0
	currentY := -1
	cursorX := 0 // 追踪终端光标的实际列位置

	for _, change := range diffResult.Changes {
		x, y := change.X, change.Y
		newCell := buf.Cells[y][x]

		// 跳过空字符和宽字符的填充单元格
		if newCell.Char == 0 || newCell.Width == 0 {
			continue
		}

		// 如果当前单元格在光标位置之前（被宽字符占据），跳过
		if y == currentY && x < cursorX {
			continue
		}

		// 如果换行了，移动到新行的开头
		if y != currentY {
			output.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
			currentY = y
			lastX, lastY = x, y
			cursorX = x
		} else if x != lastX || y != lastY {
			// 同一行内，使用相对移动
			if x > lastX {
				output.WriteString(strings.Repeat("\x1b[C", x-lastX))
			} else if x < lastX {
				// 需要向左移动，使用绝对定位
				output.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
			}
			lastX, lastY = x, y
			cursorX = x
		}

		// 设置字符
		char := newCell.Char
		if char == 0 {
			char = ' '
		}

		// 应用样式（如果改变）
		if newCell.Style != currentStyle {
			if currentStyle != (style.Style{}) {
				output.WriteString("\x1b[0m")
			}
			if newCell.Style != (style.Style{}) {
				output.WriteString(newCell.Style.ToANSI())
			}
			currentStyle = newCell.Style
		}

		output.WriteRune(char)
		// 更新光标位置（宽字符占据 2 列）
		cursorX += newCell.Width
		if cursorX == 0 {
			cursorX = 1
		}
		lastX = cursorX // 用于下次相对移动
	}

	// 重置样式
	if currentStyle != (style.Style{}) {
		output.WriteString("\x1b[0m")
	}

	// 移动光标到末尾（避免残留）
	output.WriteString(fmt.Sprintf("\x1b[%d;%dH", buf.Height, 1))

	return output.String()
}

// EnsurePrevBufferSize 确保 prevBuffer 的大小与新 buffer 匹配
func EnsurePrevBufferSize(prevBuf [][]paint.Cell, width, height int) [][]paint.Cell {
	if prevBuf == nil || len(prevBuf) != height || (len(prevBuf) > 0 && len(prevBuf[0]) != width) {
		prevBuf = make([][]paint.Cell, height)
		for y := 0; y < height; y++ {
			prevBuf[y] = make([]paint.Cell, width)
		}
	}
	return prevBuf
}

// UpdatePrevBuffer 用新 buffer 的内容更新 prevBuffer
func UpdatePrevBuffer(prevBuf [][]paint.Cell, newBuf *paint.Buffer) {
	for y := 0; y < newBuf.Height; y++ {
		copy(prevBuf[y], newBuf.Cells[y])
	}
}
