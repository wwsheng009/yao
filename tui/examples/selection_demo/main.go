// selection_demo/main.go
//
// 文本选择功能演示程序
//
// 功能说明：
// - 鼠标左键拖动：选择文字
// - 鼠标双击：选择单词
// - 鼠标三击：选择整行
// - Ctrl+A：全选
// - Ctrl+C：复制选中文本（有选择时）
// - Escape：清除选择
//
// 运行方式：
//   go run examples/selection_demo/main.go

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/runtime"
)

type selectionDemoModel struct {
	runtime      *runtime.RuntimeImpl
	width        int
	height       int
	ready        bool
}

func (m selectionDemoModel) Init() tea.Cmd {
	return nil
}

func (m selectionDemoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 窗口大小初始化
	if !m.ready {
		if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = sizeMsg.Width
			m.height = sizeMsg.Height
			m.ready = true
			m.runtime = runtime.NewRuntime(m.width, m.height)
		}
	}

	// 处理键盘快捷键
	if keyMsg, ok := msg.(tea.KeyMsg); ok && m.runtime != nil {
		selection := m.runtime.GetSelection()

		switch {
		case keyMsg.Type == tea.KeyCtrlC || keyMsg.String() == "ctrl+c":
			if selection.IsActive() {
				// 有选择时复制
				text, err := m.runtime.CopySelection()
				if err == nil && text != "" {
					m.showCopySuccess(text)
				}
				selection.Clear()
				return m, nil
			} else {
				return m, tea.Quit
			}

		case keyMsg.Type == tea.KeyCtrlA || keyMsg.String() == "ctrl+a":
			selection.SelectAll()
			return m, nil

		case keyMsg.Type == tea.KeyEscape:
			selection.Clear()
			return m, nil

		case keyMsg.Type == tea.KeyCtrlX || keyMsg.String() == "ctrl+x":
			if selection.IsActive() {
				text, err := m.runtime.CopySelection()
				if err == nil && text != "" {
					m.showCutSuccess(text)
				}
			}
			return m, nil

		case keyMsg.String() == "q":
			return m, tea.Quit
		}
	}

	// 处理鼠标事件
	if mouseMsg, ok := msg.(tea.MouseMsg); ok && m.runtime != nil {
		selection := m.runtime.GetSelection()

		switch mouseMsg.Action {
		case tea.MouseActionPress:
			if mouseMsg.Button == tea.MouseButtonLeft {
				m.runtime.StartSelection(mouseMsg.X, mouseMsg.Y)
			}

		case tea.MouseActionMotion:
			selection.Update(mouseMsg.X, mouseMsg.Y)

		case tea.MouseActionRelease:
			// 鼠标释放
		}
	}

	// 窗口大小变化
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = sizeMsg.Width
		m.height = sizeMsg.Height
		if m.runtime != nil {
			m.runtime.UpdateDimensions(m.width, m.height)
		}
	}

	return m, nil
}

func (m selectionDemoModel) View() string {
	if !m.ready {
		return "Initializing...\n"
	}

	// 渲染内容
	frame := runtime.Frame{
		Buffer: runtime.NewCellBuffer(m.width, m.height),
		Width:  m.width,
		Height: m.height,
	}

	// 填充演示文本
	lines := []string{
		"Text Selection Demo - 文本选择演示",
		"",
		"Features: 单击拖动 | 双击选词 | 三击选行",
		"Ctrl+A 全选 | Ctrl+C 复制 | Escape 清除",
		"",
		"This is a demonstration of the text selection feature.",
		"You can select text by clicking and dragging with the mouse.",
		"",
		"中文文本选择演示 - 支持中文",
		"Supports Chinese characters and wide unicode",
		"",
		"Code example - 代码示例:",
		"func hello() { return 42 }",
		"",
		"Data table - 数据表格:",
		"Name    Age    City           Status",
		"Alice   25     New York       Active",
		"Bob     30     London         Inactive",
		"Charlie 35     Tokyo          Pending",
		"",
	}

	for y, line := range lines {
		if y >= m.height-3 { // 留出空间给底部信息
			break
		}
		for x, ch := range line {
			if x >= m.width {
				break
			}
			style := runtime.CellStyle{}
			if y == 0 || y == 2 || y == 6 {
				style.Bold = true
				style.Foreground = "214" // Orange
			} else if y == 9 {
				style.Foreground = "33" // Blue
			}
			frame.Buffer.SetContent(x, y, 0, ch, style, "")
		}
	}

	// 应用选择高亮
	if m.runtime != nil {
		selection := m.runtime.GetSelection()
		selection.SetBuffer(frame.Buffer)
		selection.ApplyHighlight()
	}

	// 显示帮助信息
	helpText := m.getHelpText()
	selectionInfo := m.getSelectionInfo()

	// 组合输出
	return lipgloss.JoinVertical(lipgloss.Left,
		frame.String(),
		"",
		selectionInfo,
		helpText,
	)
}

func (m selectionDemoModel) getSelectionInfo() string {
	if m.runtime == nil {
		return ""
	}

	selection := m.runtime.GetSelection()
	if !selection.IsActive() {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("No selection | 无选择")
	}

	text := m.runtime.GetSelectedTextCompact()
	if len(text) > 50 {
		text = text[:47] + "..."
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Green
		Bold(true).
		Render(fmt.Sprintf("Selected | 已选择: %s (%d chars)", text, len(text)))
}

func (m selectionDemoModel) getHelpText() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")).
		Render("Help | 帮助:")

	help := []string{
		"Mouse:  Click-drag | Double-click (word) | Triple-click (line)",
		"鼠标:  点击拖动  | 双击(单词)       | 三击(整行)",
		"",
		"Ctrl+A  Select all | 全选",
		"Ctrl+C  Copy (or Quit if no selection) | 复制(无选择时退出)",
		"Ctrl+X  Cut | 剪切",
		"Esc     Clear selection | 清除选择",
		"Q       Quit | 退出",
	}

	lines := make([]string, 0, len(help))
	for _, h := range help {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(h))
	}

	return title + "\n" + lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m selectionDemoModel) showCopySuccess(text string) {
	fmt.Printf("\n[Copied to clipboard] 复制到剪贴板:\n%s\n", text)
}

func (m selectionDemoModel) showCutSuccess(text string) {
	fmt.Printf("\n[Cut to clipboard] 剪切到剪贴板:\n%s\n", text)
}

func main() {
	if !isTerminalSupported() {
		fmt.Println("Error: This program requires a terminal with mouse support.")
		fmt.Println("错误：此程序需要支持鼠标的终端。")
		os.Exit(1)
	}

	model := selectionDemoModel{
		width:  80,
		height: 24,
	}

	p := tea.NewProgram(
		model,
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func isTerminalSupported() bool {
	return os.Getenv("TERM") != "" || os.Getenv("WT_SESSION") != ""
}
