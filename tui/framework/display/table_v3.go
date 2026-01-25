package display

import (
	"fmt"
	"sync"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Table Component V3 (with Sub-Focus & Virtual Scrolling)
// ==============================================================================
// V3 表格组件，支持子焦点、单元格导航、选择、虚拟滚动

// TableColumnV3 表格列定义
type TableColumnV3 struct {
	Title    string
	Width    int
	MinWidth int
	MaxWidth int
	Align    component.TextAlign
	Sortable bool
	Visible  bool
}

// TableCellPosition 单元格位置
type TableCellPosition struct {
	Row int
	Col int
}

// TableSelectionV3 表格选择区域
type TableSelectionV3 struct {
	fromRow int
	fromCol int
	toRow   int
	toCol   int
	active  bool
	mu      sync.RWMutex
}

// NewTableSelectionV3 创建选择区域
func NewTableSelectionV3() *TableSelectionV3 {
	return &TableSelectionV3{}
}

// Set 设置选择区域
func (s *TableSelectionV3) Set(fromRow, fromCol, toRow, toCol int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 规范化坐标
	if fromRow > toRow {
		fromRow, toRow = toRow, fromRow
	}
	if fromCol > toCol {
		fromCol, toCol = toCol, fromCol
	}

	s.fromRow, s.toRow = fromRow, toRow
	s.fromCol, s.toCol = fromCol, toCol
	s.active = true
}

// Clear 清除选择
func (s *TableSelectionV3) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
}

// IsEmpty 是否为空
func (s *TableSelectionV3) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.active
}

// Contains 检查是否包含指定单元格
func (s *TableSelectionV3) Contains(row, col int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.active {
		return false
	}
	return row >= s.fromRow && row <= s.toRow &&
		col >= s.fromCol && col <= s.toCol
}

// Range 返回选择范围
func (s *TableSelectionV3) Range() (fromRow, fromCol, toRow, toCol int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fromRow, s.fromCol, s.toRow, s.toCol
}

// TableSubFocusV3 表格子焦点
type TableSubFocusV3 struct {
	mu sync.RWMutex

	// 当前焦点位置
	row int
	col int

	// 选择状态
	selection *TableSelectionV3

	// 约束
	maxRow int
	maxCol int

	// 回调
	onCellChange func(row, col int)
}

// NewTableSubFocusV3 创建子焦点
func NewTableSubFocusV3(rows, cols int) *TableSubFocusV3 {
	return &TableSubFocusV3{
		row:       0,
		col:       0,
		selection: NewTableSelectionV3(),
		maxRow:    maxInt(0, rows-1),
		maxCol:    maxInt(0, cols-1),
	}
}

// SetPosition 设置焦点位置
func (f *TableSubFocusV3) SetPosition(row, col int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isValidPositionUnsafe(row, col) {
		return false
	}

	oldRow, oldCol := f.row, f.col
	f.row = row
	f.col = col

	// 触发回调
	if f.onCellChange != nil && (oldRow != row || oldCol != col) {
		f.onCellChange(row, col)
	}

	return true
}

// Position 获取焦点位置
func (f *TableSubFocusV3) Position() (row, col int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.row, f.col
}

// Move 移动焦点
func (f *TableSubFocusV3) Move(deltaRow, deltaCol int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	newRow := f.row + deltaRow
	newCol := f.col + deltaCol

	// 限制范围
	if newRow < 0 {
		newRow = 0
	}
	if newRow > f.maxRow {
		newRow = f.maxRow
	}
	if newCol < 0 {
		newCol = 0
	}
	if newCol > f.maxCol {
		newCol = f.maxCol
	}

	if f.row != newRow || f.col != newCol {
		oldRow, oldCol := f.row, f.col
		f.row = newRow
		f.col = newCol

		if f.onCellChange != nil && (oldRow != newRow || oldCol != newCol) {
			f.onCellChange(newRow, newCol)
		}
		return true
	}

	return false
}

// SetSelection 设置选择区域
func (f *TableSubFocusV3) SetSelection(fromRow, fromCol, toRow, toCol int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selection.Set(fromRow, fromCol, toRow, toCol)
}

// GetSelection 获取选择区域
func (f *TableSubFocusV3) GetSelection() *TableSelectionV3 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.selection
}

// ClearSelection 清除选择
func (f *TableSubFocusV3) ClearSelection() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selection.Clear()
}

// HasSelection 是否有选择
func (f *TableSubFocusV3) HasSelection() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return !f.selection.IsEmpty()
}

// IsSelected 检查单元格是否被选中
func (f *TableSubFocusV3) IsSelected(row, col int) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.selection.Contains(row, col)
}

// SelectAll 全选
func (f *TableSubFocusV3) SelectAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selection.Set(0, 0, f.maxRow, f.maxCol)
}

// SelectRow 选择当前行
func (f *TableSubFocusV3) SelectRow() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selection.Set(f.row, 0, f.row, f.maxCol)
}

// SelectCol 选择当前列
func (f *TableSubFocusV3) SelectCol() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selection.Set(0, f.col, f.maxRow, f.col)
}

// Resize 调整大小
func (f *TableSubFocusV3) Resize(rows, cols int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.maxRow = maxInt(0, rows-1)
	f.maxCol = maxInt(0, cols-1)

	// 确保焦点在范围内
	if f.row > f.maxRow {
		f.row = f.maxRow
	}
	if f.col > f.maxCol {
		f.col = f.maxCol
	}
}

// SetOnCellChange 设置单元格变化回调
func (f *TableSubFocusV3) SetOnCellChange(fn func(row, col int)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onCellChange = fn
}

// isValidPositionUnsafe 检查位置是否有效（未加锁）
func (f *TableSubFocusV3) isValidPositionUnsafe(row, col int) bool {
	return row >= 0 && row <= f.maxRow && col >= 0 && col <= f.maxCol
}

// TableV3 V3 表格组件
type TableV3 struct {
	*component.BaseComponentV3
	*component.StateHolder

	mu sync.RWMutex

	// 列定义
	columns []TableColumnV3

	// 数据
	rows    [][]string
	rowCount int

	// 子焦点
	subFocus *TableSubFocusV3

	// 视口状态（虚拟滚动）
	offset   int // 滚动偏移（行索引）
	height   int // 可见高度

	// 表头
	showHeader    bool
	headerHeight  int

	// 样式
	gridLines      bool
	showRowNumbers bool
	alternateRows  bool

	// 样式配置
	headerStyle      style.Style
	normalRowStyle   style.Style
	alternateRowStyle style.Style
	focusStyle       style.Style
	selectedStyle    style.Style
	gridStyle        style.Style

	// 排序
	sortColumn   int
	sortAsc      bool

	// 回调
	onSelect func(row, col int)
	onCellChange func(row, col int)
}

// NewTableV3 创建 V3 表格组件
func NewTableV3(columns []TableColumnV3) *TableV3 {
	return &TableV3{
		BaseComponentV3:  component.NewBaseComponentV3("table"),
		StateHolder:      component.NewStateHolder(),
		columns:          columns,
		rows:             make([][]string, 0),
		rowCount:         0,
		subFocus:         NewTableSubFocusV3(0, len(columns)),
		offset:           0,
		height:           10,
		showHeader:       true,
		headerHeight:     1,
		gridLines:        true,
		showRowNumbers:   false,
		alternateRows:    true,
		headerStyle:      style.Style{}.Foreground(style.White).Background(style.Blue),
		normalRowStyle:   style.Style{},
		alternateRowStyle: style.Style{}.Background(style.BrightBlack),
		focusStyle:       style.Style{}.Reverse(true),
		selectedStyle:    style.Style{}.Foreground(style.White).Background(style.Cyan),
		gridStyle:        style.Style{}.Foreground(style.BrightBlack),
		sortColumn:       -1,
		sortAsc:          true,
	}
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetRows 设置行数据
func (t *TableV3) SetRows(rows [][]string) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows = rows
	t.rowCount = len(rows)
	t.subFocus.Resize(t.rowCount, len(t.columns))
	return t
}

// AddRow 添加行
func (t *TableV3) AddRow(row []string) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows = append(t.rows, row)
	t.rowCount++
	t.subFocus.Resize(t.rowCount, len(t.columns))
	return t
}

// ClearRows 清空所有行
func (t *TableV3) ClearRows() *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows = make([][]string, 0)
	t.rowCount = 0
	t.offset = 0
	t.subFocus.Resize(0, len(t.columns))
	return t
}

// SetCell 设置单元格
func (t *TableV3) SetCell(row, col int, value string) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 扩展行数
	for row >= len(t.rows) {
		newRow := make([]string, len(t.columns))
		t.rows = append(t.rows, newRow)
		t.rowCount++
	}

	// 扩展列
	if col >= len(t.rows[row]) {
		newRow := make([]string, col+1)
		copy(newRow, t.rows[row])
		t.rows[row] = newRow
	}

	t.rows[row][col] = value
	t.subFocus.Resize(t.rowCount, len(t.columns))
	return t
}

// SetHeight 设置可见高度
func (t *TableV3) SetHeight(h int) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.height = h
	return t
}

// SetShowHeader 设置是否显示表头
func (t *TableV3) SetShowHeader(show bool) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.showHeader = show
	return t
}

// SetGridLines 设置是否显示网格线
func (t *TableV3) SetGridLines(show bool) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.gridLines = show
	return t
}

// SetAlternateRows 设置是否交替行颜色
func (t *TableV3) SetAlternateRows(enable bool) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.alternateRows = enable
	return t
}

// SetOnSelect 设置选中回调
func (t *TableV3) SetOnSelect(fn func(row, col int)) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onSelect = fn
	return t
}

// SetOnCellChange 设置单元格变化回调
func (t *TableV3) SetOnCellChange(fn func(row, col int)) *TableV3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onCellChange = fn
	t.subFocus.SetOnCellChange(fn)
	return t
}

// GetSubFocus 获取子焦点
func (t *TableV3) GetSubFocus() *TableSubFocusV3 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.subFocus
}

// GetRowCount 获取行数
func (t *TableV3) GetRowCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.rowCount
}

// GetColumnCount 获取列数
func (t *TableV3) GetColumnCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.columns)
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (t *TableV3) Measure(maxWidth, maxHeight int) (width, height int) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 计算宽度
	width = 0
	for _, col := range t.columns {
		if col.Visible {
			width += col.Width
		}
	}
	if t.gridLines {
		width += len(t.columns) - 1 // 网格线
	}
	if t.showRowNumbers {
		width += 4 // 行号列
	}

	// 计算高度
	height = t.rowCount
	if t.showHeader {
		height += t.headerHeight + 1 // 表头 + 分隔线
	}

	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}

	return width, height
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件到 CellBuffer
func (t *TableV3) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !t.IsVisible() {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	t.height = height
	y := ctx.Y

	// 绘制表头
	if t.showHeader {
		t.paintHeader(ctx, buf, y)
		y += t.headerHeight + 1
		if t.gridLines {
			t.paintSeparator(ctx, buf, y-1)
		}
	}

	// 绘制数据行
	dataHeight := height - (y - ctx.Y)
	startRow, endRow := t.getVisibleRange(dataHeight)

	for row := startRow; row < endRow && row < t.rowCount && y < ctx.Y+height; row++ {
		t.paintRow(ctx, buf, row, y)
		y++
	}

	// 填充剩余空间
	for y < ctx.Y+height {
		t.paintEmptyRow(ctx, buf, y)
		y++
	}

	// 绘制滚动条
	if t.rowCount > dataHeight {
		t.paintScrollbar(ctx, buf)
	}
}

// paintHeader 绘制表头
func (t *TableV3) paintHeader(ctx component.PaintContext, buf *paint.Buffer, y int) {
	x := ctx.X

	// 行号列
	if t.showRowNumbers {
		t.paintCell(buf, x, y, "#", 4, t.headerStyle, component.AlignCenter)
		x += 4
		if t.gridLines {
			buf.SetCell(x, y, '│', t.headerStyle)
			x++
		}
	}

	// 列标题
	for i, col := range t.columns {
		if !col.Visible {
			continue
		}

		title := col.Title
		if t.sortColumn == i {
			if t.sortAsc {
				title += " ▲"
			} else {
				title += " ▼"
			}
		}

		t.paintCell(buf, x, y, title, col.Width, t.headerStyle, col.Align)
		x += col.Width

		if t.gridLines && i < len(t.columns)-1 {
			buf.SetCell(x, y, '│', t.headerStyle)
			x++
		}
	}
}

// paintSeparator 绘制分隔线
func (t *TableV3) paintSeparator(ctx component.PaintContext, buf *paint.Buffer, y int) {
	x := ctx.X
	width := ctx.AvailableWidth

	for i := 0; i < width; i++ {
		buf.SetCell(x+i, y, '─', t.headerStyle)
	}
}

// paintRow 绘制数据行
func (t *TableV3) paintRow(ctx component.PaintContext, buf *paint.Buffer, row int, y int) {
	x := ctx.X

	// 确定行样式
	rowStyle := t.normalRowStyle
	if t.alternateRows && row%2 == 1 {
		rowStyle = t.alternateRowStyle
	}

	// 获取焦点位置
	focusRow, focusCol := t.subFocus.Position()
	isFocused := t.IsFocused() && row == focusRow

	// 行号列
	if t.showRowNumbers {
		rowNum := fmt.Sprintf("%d", row+1)
		style := rowStyle
		if isFocused {
			style = t.focusStyle
		}
		t.paintCell(buf, x, y, rowNum, 4, style, component.AlignRight)
		x += 4
		if t.gridLines {
			buf.SetCell(x, y, '│', t.gridStyle)
			x++
		}
	}

	// 数据列
	for colIndex, col := range t.columns {
		if !col.Visible {
			continue
		}

		// 获取单元格值
		cellValue := ""
		if row < len(t.rows) && colIndex < len(t.rows[row]) {
			cellValue = t.rows[row][colIndex]
		}

		// 确定样式
		cellStyle := rowStyle
		if isFocused && colIndex == focusCol {
			cellStyle = t.focusStyle
		} else if t.subFocus.IsSelected(row, colIndex) {
			cellStyle = t.selectedStyle
		}

		t.paintCell(buf, x, y, cellValue, col.Width, cellStyle, col.Align)
		x += col.Width

		if t.gridLines && colIndex < len(t.columns)-1 {
			buf.SetCell(x, y, '│', t.gridStyle)
			x++
		}
	}
}

// paintEmptyRow 绘制空行
func (t *TableV3) paintEmptyRow(ctx component.PaintContext, buf *paint.Buffer, y int) {
	x := ctx.X
	width := ctx.AvailableWidth

	for i := 0; i < width; i++ {
		buf.SetCell(x+i, y, ' ', style.Style{})
	}
}

// paintCell 绘制单元格
func (t *TableV3) paintCell(buf *paint.Buffer, x, y int, text string, width int, s style.Style, align component.TextAlign) {
	runes := []rune(text)
	textLen := len(runes)

	// 截断过长的文本
	if textLen > width {
		runes = runes[:width]
		textLen = width
	}

	// 计算对齐
	padding := width - textLen
	var leftPad, rightPad int

	switch align {
	case component.AlignCenter:
		leftPad = padding / 2
		rightPad = padding - leftPad
	case component.AlignRight:
		leftPad = padding
		rightPad = 0
	default: // AlignLeft
		leftPad = 0
		rightPad = padding
	}

	// 绘制
	currentX := x
	for i := 0; i < leftPad; i++ {
		buf.SetCell(currentX, y, ' ', s)
		currentX++
	}
	for _, r := range runes {
		buf.SetCell(currentX, y, r, s)
		currentX++
	}
	for i := 0; i < rightPad; i++ {
		buf.SetCell(currentX, y, ' ', s)
		currentX++
	}
}

// paintScrollbar 绘制滚动条
func (t *TableV3) paintScrollbar(ctx component.PaintContext, buf *paint.Buffer) {
	dataHeight := t.height
	if t.showHeader {
		dataHeight -= t.headerHeight + 1
	}
	if dataHeight <= 0 {
		return
	}

	if t.rowCount <= dataHeight {
		return
	}

	barHeight := maxInt(1, dataHeight*dataHeight/t.rowCount)
	barPos := t.offset * (dataHeight - barHeight) / maxInt(1, t.rowCount-dataHeight)

	x := ctx.X + ctx.AvailableWidth - 1
	baseY := ctx.Y
	if t.showHeader {
		baseY += t.headerHeight + 1
	}

	for i := 0; i < dataHeight; i++ {
		y := baseY + i
		if i >= barPos && i < barPos+barHeight {
			buf.SetCell(x, y, '█', style.Style{}.Foreground(style.BrightBlack))
		} else {
			buf.SetCell(x, y, '│', style.Style{}.Foreground(style.BrightBlack))
		}
	}
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction 处理语义化 Action
func (t *TableV3) HandleAction(a action.Action) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch a.Type {
	case action.ActionNavigateUp:
		t.navigateAndFocus(0, -1)
		return true

	case action.ActionNavigateDown:
		t.navigateAndFocus(0, 1)
		return true

	case action.ActionNavigateLeft:
		t.navigateAndFocus(-1, 0)
		return true

	case action.ActionNavigateRight:
		t.navigateAndFocus(1, 0)
		return true

	case action.ActionNavigatePageUp:
		pageStep := maxInt(1, t.height-2)
		t.navigateAndFocus(0, -pageStep)
		return true

	case action.ActionNavigatePageDown:
		pageStep := maxInt(1, t.height-2)
		t.navigateAndFocus(0, pageStep)
		return true

	case action.ActionNavigateFirst:
		t.subFocus.SetPosition(0, 0)
		t.ensureVisible(0)
		return true

	case action.ActionNavigateLast:
		lastRow := maxInt(0, t.rowCount-1)
		lastCol := maxInt(0, len(t.columns)-1)
		t.subFocus.SetPosition(lastRow, lastCol)
		t.ensureVisible(lastRow)
		return true

	case action.ActionSelectAll:
		t.subFocus.SelectAll()
		return true

	case action.ActionSelectItem:
		row, col := t.subFocus.Position()
		if t.onSelect != nil {
			t.onSelect(row, col)
		}
		return true

	case action.ActionSubmit:
		row, col := t.subFocus.Position()
		if t.onSelect != nil {
			t.onSelect(row, col)
		}
		return true
	}

	return false
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (t *TableV3) FocusID() string {
	return t.ID()
}

// OnFocus 获得焦点时调用
func (t *TableV3) OnFocus() {
	// 可以在这里添加获得焦点时的逻辑
}

// OnBlur 失去焦点时调用
func (t *TableV3) OnBlur() {
	// 可以在这里添加失去焦点时的逻辑
}

// ============================================================================
// 内部方法
// ============================================================================

// getVisibleRange 获取可见行范围
func (t *TableV3) getVisibleRange(dataHeight int) (start, end int) {
	start = t.offset
	end = minInt(t.offset+dataHeight, t.rowCount)
	return
}

// navigateAndFocus 导航并设置焦点
func (t *TableV3) navigateAndFocus(deltaCol, deltaRow int) {
	row, col := t.subFocus.Position()
	newRow := row + deltaRow
	newCol := col + deltaCol

	// 限制列范围
	if newCol < 0 {
		newCol = 0
	}
	if newCol >= len(t.columns) {
		newCol = len(t.columns) - 1
	}

	// 限制行范围
	if newRow < 0 {
		newRow = 0
	}
	if newRow >= t.rowCount {
		newRow = t.rowCount - 1
	}

	if t.subFocus.SetPosition(newRow, newCol) {
		t.ensureVisible(newRow)
	}
}

// ensureVisible 确保指定行可见
func (t *TableV3) ensureVisible(row int) {
	dataHeight := t.height
	if t.showHeader {
		dataHeight -= t.headerHeight + 1
	}
	if dataHeight <= 0 {
		return
	}

	if row < t.offset {
		t.offset = row
	} else if row >= t.offset+dataHeight {
		t.offset = row - dataHeight + 1
	}

	// 限制偏移范围
	maxOffset := maxInt(0, t.rowCount-dataHeight)
	if t.offset > maxOffset {
		t.offset = maxOffset
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
