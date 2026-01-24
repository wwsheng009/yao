package display

import (
	"fmt"
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Table 表格组件
type Table struct {
	*component.BaseComponent

	columns   []Column
	rows      [][]string
	cursor    int
	offset    int
	offsetX   int
	fixedRows int
	fixedCols int
	showHeader bool
}

// Column 列定义
type Column struct {
	Title   string
	Width   int
	Align   AlignType
	Sortable bool
}

// AlignType 对齐类型
type AlignType int

const (
	AlignLeft AlignType = iota
	AlignCenter
	AlignRight
)

// NewTable 创建表格
func NewTable() *Table {
	return &Table{
		BaseComponent: component.NewBaseComponent("table"),
		columns:       make([]Column, 0),
		rows:          make([][]string, 0),
		cursor:        -1,
		showHeader:    true,
	}
}

// NewTableColumns 创建带列的表格
func NewTableColumns(columns []Column) *Table {
	return &Table{
		BaseComponent: component.NewBaseComponent("table"),
		columns:       columns,
		rows:          make([][]string, 0),
		cursor:        -1,
		showHeader:    true,
	}
}

// SetColumns 设置列
func (t *Table) SetColumns(columns []Column) {
	t.columns = columns
}

// AddColumn 添加列
func (t *Table) AddColumn(column Column) {
	t.columns = append(t.columns, column)
}

// SetRows 设置行
func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
}

// AddRow 添加行
func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// SetCell 设置单元格
func (t *Table) SetCell(row, col int, value string) {
	// 扩展行数
	for len(t.rows) <= row {
		t.rows = append(t.rows, make([]string, len(t.columns)))
	}
	// 扩展列数
	for len(t.rows[row]) <= col {
		t.rows[row] = append(t.rows[row], "")
	}
	t.rows[row][col] = value
}

// GetRowCount 获取行数
func (t *Table) GetRowCount() int {
	return len(t.rows)
}

// GetCell 获取单元格
func (t *Table) GetCell(row, col int) string {
	if row < 0 || row >= len(t.rows) {
		return ""
	}
	if col < 0 || col >= len(t.rows[row]) {
		return ""
	}
	return t.rows[row][col]
}

// SetCursor 设置光标
func (t *Table) SetCursor(row int) {
	if row >= -1 && row < len(t.rows) {
		t.cursor = row
		t.updateOffset()
	}
}

// GetCursor 获取光标位置
func (t *Table) GetCursor() (row, col int) {
	return t.cursor, 0
}

// Render 渲染表格
func (t *Table) Render(ctx *component.RenderContext) string {
	if !t.IsVisible() {
		return ""
	}

	width, height := ctx.AvailableWidth, ctx.AvailableHeight

	// 计算列宽
	colWidths := t.calculateColWidths(width)

	var result []string

	currentY := 0

	// 渲染表头
	if t.showHeader && len(t.columns) > 0 {
		headerLine := "┌"
		for i, col := range t.columns {
			w := colWidths[i]
			headerLine += strings.Repeat("─", w) + "┬"
		}
		headerLine = headerLine[:len(headerLine)-1] + "┐"

		// 分隔符样式
		borderStyle := style.NewStyle().Foreground(style.BrightBlack)
		result = append(result, borderStyle.Apply(headerLine))

		// 列标题行
		titleLine := "│"
		for i, col := range t.columns {
			title := col.Title
			w := colWidths[i]
			title = t.alignText(title, w, col.Align)
			titleLine += " " + title + " │"
		}
		result = append(result, titleLine)

		// 表头下划线
		separatorLine := "├"
		for i, w := range colWidths {
			separatorLine += strings.Repeat("─", w) + "┼"
		}
		separatorLine = separatorLine[:len(separatorLine)-1] + "┤"
		result = append(result, borderStyle.Apply(separatorLine))

		currentY += 3
	}

	// 渲染数据行
	startRow := t.offset
	maxRows := height - currentY
	if startRow+maxRows > len(t.rows) {
		maxRows = len(t.rows) - startRow
	}
	if maxRows < 0 {
		maxRows = 0
	}

	for i := startRow; i < startRow+maxRows; i++ {
		if i >= len(t.rows) {
			break
		}

		dataLine := "│"
		for j, col := range t.columns {
			value := ""
			if j < len(t.rows[i]) {
				value = t.rows[i][j]
			}
			w := colWidths[j]
			value = t.alignText(value, w, col.Align)

			// 高亮选中行
			if i == t.cursor && t.cursor >= 0 {
				dataLine += style.NewStyle().Reverse(true).Apply(" "+value+" ")
			} else {
				dataLine += " " + value + " "
			}
			dataLine += "│"
		}
		result = append(result, dataLine)

		currentY++
	}

	// 底边框
	if len(result) > 0 {
		bottomLine := "└"
		for _, w := range colWidths {
			bottomLine += strings.Repeat("─", w) + "┴"
		}
		bottomLine = bottomLine[:len(bottomLine)-1] + "┘"
		borderStyle := style.NewStyle().Foreground(style.BrightBlack)
		result = append(result, borderStyle.Apply(bottomLine))
	}

	return strings.Join(result, "\n")
}

// calculateColWidths 计算列宽
func (t *Table) calculateColWidths(availableWidth int) []int {
	colCount := len(t.columns)
	if colCount == 0 {
		return []int{}
	}

	// 边框占用 2 个字符，每个单元格两侧各占 1 个字符
	availableWidth -= 2
	if availableWidth < colCount*2 {
		availableWidth = colCount * 2
	}

	widths := make([]int, colCount)
	totalFixed := 0
	autoCols := 0

	// 计算固定宽度和自动列数
	for i, col := range t.columns {
		if col.Width > 0 {
			widths[i] = col.Width
			totalFixed += col.Width
		} else {
			autoCols++
		}
	}

	// 分配剩余宽度给自动列
	remainingWidth := availableWidth - totalFixed - (colCount-1)*2 // 列间隔
	if autoCols > 0 && remainingWidth > 0 {
		avgWidth := remainingWidth / autoCols
		for i := range widths {
			if widths[i] == 0 {
				widths[i] = avgWidth
			}
		}
	}

	// 确保最小宽度
	minWidth := 3
	for i := range widths {
		if widths[i] < minWidth {
			widths[i] = minWidth
		}
	}

	return widths
}

// alignText 对齐文本
func (t *Table) alignText(text string, width int, align AlignType) string {
	if len(text) > width {
		return text[:width]
	}

	switch align {
	case AlignCenter:
		padding := (width - len(text)) / 2
		return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-padding-len(text))
	case AlignRight:
		return strings.Repeat(" ", width-len(text)) + text
	default:
		return text + strings.Repeat(" ", width-len(text))
	}
}

// updateOffset 更新滚动偏移
func (t *Table) updateOffset() {
	// TODO: 根据光标位置和可见高度计算偏移
}

// HandleEvent 处理事件
func (t *Table) HandleEvent(ev component.Event) bool {
	switch e := ev.(type) {
	case *event.KeyEvent:
		return t.handleKey(e)
	}
	return false
}

// handleKey 处理键盘事件
func (t *Table) handleKey(ev *event.KeyEvent) bool {
	switch ev.Special {
	case event.KeyUp, event.KeyK:
		if t.cursor > 0 {
			t.cursor--
			t.updateOffset()
		}
		return true
	case event.KeyDown, event.KeyJ:
		if t.cursor < len(t.rows)-1 {
			t.cursor++
			t.updateOffset()
		}
		return true
	case event.KeyPageUp:
		t.cursor -= 5
		if t.cursor < -1 {
			t.cursor = -1
		}
		t.updateOffset()
		return true
	case event.KeyPageDown:
		t.cursor += 5
		if t.cursor >= len(t.rows) {
			t.cursor = len(t.rows) - 1
		}
		t.updateOffset()
		return true
	case event.KeyHome:
		t.cursor = -1
		return true
	case event.KeyEnd:
		t.cursor = len(t.rows) - 1
		return true
	}
	return false
}

// GetPreferredSize 获取首选尺寸
func (t *Table) GetPreferredSize() (width, height int) {
	// 计算宽度
	totalWidth := 2 // 边框
	for _, col := range t.columns {
		if col.Width > 0 {
			totalWidth += col.Width
		} else {
			totalWidth += 10 // 默认列宽
		}
		totalWidth += 1 // 列间隔
	}

	// 计算高度
	height = 2 // 边框
	if t.showHeader {
		height += 2 // 表头
	}
	height += len(t.rows) // 数据行
	if t.fixedRows > 0 && height > t.fixedRows {
		height = t.fixedRows
	}

	return width, height
}
