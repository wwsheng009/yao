# Table Sub-Focus System Design (V3)

> **优先级**: P1 (Table 编辑功能)
> **目标**: 支持 Table 内部单元格级别的焦点管理
> **关键特性**: 子焦点、单元格导航、内联编辑、选择模式

## 概述

Table 组件是 TUI 应用中最复杂的组件之一。除了常规的组件焦点外，Table 还需要管理内部的"子焦点"——即单个单元格的焦点状态。这使得用户可以在 Table 内部导航和编辑单元格。

### 为什么需要 Table 子焦点系统？

**传统焦点系统的问题**：
```go
// ❌ 只能聚焦到整个 Table
type Table struct {
    rows [][]string
    selectedRow int
}

func (t *Table) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateDown:
        t.selectedRow++  // 只能选择整行
        return true
    }
}

// 问题：
// - 无法编辑单个单元格
// - 无法选择单元格区域
// - 无法复制单个单元格
// - 焦点导航不直观
```

**有子焦点系统的优势**：
```go
// ✅ 支持单元格级别的焦点
type Table struct {
    subFocus *TableSubFocus  // 子焦点管理器
}

// 子焦点可以：
// - 聚焦到单个单元格
// - 在单元格间导航
// - 编辑单元格内容
// - 选择多个单元格
```

## 设计目标

1. **单元格导航**: 支持在单元格间移动焦点
2. **内联编辑**: 支持单元格内编辑
3. **选择模式**: 支持选择单个或多个单元格
4. **键盘快捷键**: 支持常用的快捷键（复制、粘贴、删除等）
5. **与主焦点协调**: 子焦点与组件焦点协调工作

## 核心架构

```
┌─────────────────────────────────────────────────────┐
│                   Focus Manager                      │
│                  (组件级焦点)                         │
└──────────────────┬──────────────────────────────────┘
                   │ Table 获得焦点
                   ▼
┌─────────────────────────────────────────────────────┐
│                    Table 组件                         │
│  ┌───────────────────────────────────────────────┐  │
│  │              Sub-Focus Manager                 │  │
│  │  (单元格级焦点)                                │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───┬───┬───┬───┐                                 │
│  │ A │ B │ C │ D │  ← 当前焦点在 (0,0)             │
│  ├───┼───┼───┼───┤                                 │
│  │ E │ F │ G │ H │                                 │
│  ├───┼───┼───┼───┤                                 │
│  │ I │ J │ K │ L │                                 │
│  └───┴───┴───┴───┘                                 │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │              Selection Manager                 │  │
│  │  (选择区域: (0,0) - (1,2))                     │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## 核心类型定义

### 1. TableSubFocus 子焦点

```go
// 位于: tui/framework/component/table_subfocus.go

package component

import (
    "github.com/yaoapp/yao/tui/framework/action"
    "github.com/yaoapp/yao/tui/runtime/focus"
)

// TableSubFocus Table 子焦点
type TableSubFocus struct {
    // 当前焦点位置
    row int
    col int

    // 选择状态
    selection *TableSelection

    // 编辑状态
    editing      bool
    editCell     *TableCell
    editor       Component // 当前编辑器

    // 导航模式
    mode NavigationMode

    // 约束
    maxRow int
    maxCol int

    // 回调
    onCellChange func(row, col int)
    onEnterEdit  func(row, col int) bool
    onExitEdit   func(row, col int, cancelled bool)
}

// NavigationMode 导航模式
type NavigationMode int

const (
    // ModeCell 单元格模式（焦点在单元格）
    ModeCell NavigationMode = iota

    // ModeRow 行模式（焦点在整行）
    ModeRow

    // ModeBlock 块模式（选择区域）
    ModeBlock
)

// NewTableSubFocus 创建子焦点
func NewTableSubFocus(rows, cols int) *TableSubFocus {
    return &TableSubFocus{
        row:       0,
        col:       0,
        selection: NewTableSelection(),
        mode:      ModeCell,
        maxRow:    rows - 1,
        maxCol:    cols - 1,
    }
}

// SetPosition 设置焦点位置
func (f *TableSubFocus) SetPosition(row, col int) bool {
    if !f.isValidPosition(row, col) {
        return false
    }

    // 如果正在编辑，先退出编辑
    if f.editing {
        f.ExitEdit(false)
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
func (f *TableSubFocus) Position() (row, col int) {
    return f.row, f.col
}

// Move 移动焦点
func (f *TableSubFocus) Move(direction focus.Direction) bool {
    newRow, newCol := f.row, f.col

    switch direction {
    case focus.DirectionUp:
        newRow = f.row - 1
        if newRow < 0 {
            newRow = 0
        }
    case focus.DirectionDown:
        newRow = f.row + 1
        if newRow > f.maxRow {
            newRow = f.maxRow
        }
    case focus.DirectionLeft:
        newCol = f.col - 1
        if newCol < 0 {
            newCol = 0
        }
    case focus.DirectionRight:
        newCol = f.col + 1
        if newCol > f.maxCol {
            newCol = f.maxCol
        }
    case focus.DirectionPageUp:
        newRow = f.row - 10
        if newRow < 0 {
            newRow = 0
        }
    case focus.DirectionPageDown:
        newRow = f.row + 10
        if newRow > f.maxRow {
            newRow = f.maxRow
        }
    case focus.DirectionHome:
        newCol = 0
    case focus.DirectionEnd:
        newCol = f.maxCol
    case focus.DirectionFirst:
        newRow, newCol = 0, 0
    case focus.DirectionLast:
        newRow, newCol = f.maxRow, f.maxCol
    }

    return f.SetPosition(newRow, newCol)
}

// EnterEdit 进入编辑模式
func (f *TableSubFocus) EnterEdit(editor Component) bool {
    if f.editing {
        return false
    }

    // 检查是否可编辑
    if f.onEnterEdit != nil {
        if !f.onEnterEdit(f.row, f.col) {
            return false
        }
    }

    f.editing = true
    f.editor = editor

    return true
}

// ExitEdit 退出编辑模式
func (f *TableSubFocus) ExitEdit(cancelled bool) bool {
    if !f.editing {
        return false
    }

    // 触发回调
    if f.onExitEdit != nil {
        f.onExitEdit(f.row, f.col, cancelled)
    }

    f.editing = false
    f.editor = nil

    return true
}

// IsEditing 是否在编辑模式
func (f *TableSubFocus) IsEditing() bool {
    return f.editing
}

// Editor 获取当前编辑器
func (f *TableSubFocus) Editor() Component {
    return f.editor
}

// SetSelection 设置选择区域
func (f *TableSubFocus) SetSelection(fromRow, fromCol, toRow, toCol int) {
    f.selection.Set(fromRow, fromCol, toRow, toCol)
}

// GetSelection 获取选择区域
func (f *TableSubFocus) GetSelection() *TableSelection {
    return f.selection
}

// SelectAll 全选
func (f *TableSubFocus) SelectAll() {
    f.selection.Set(0, 0, f.maxRow, f.maxCol)
}

// SelectRow 选择当前行
func (f *TableSubFocus) SelectRow() {
    f.selection.Set(f.row, 0, f.row, f.maxCol)
}

// SelectCol 选择当前列
func (f *TableSubFocus) SelectCol() {
    f.selection.Set(0, f.col, f.maxRow, f.col)
}

// ClearSelection 清除选择
func (f *TableSubFocus) ClearSelection() {
    f.selection.Clear()
}

// HasSelection 是否有选择
func (f *TableSubFocus) HasSelection() bool {
    return !f.selection.IsEmpty()
}

// IsSelected 检查单元格是否被选中
func (f *TableSubFocus) IsSelected(row, col int) bool {
    return f.selection.Contains(row, col)
}

// Resize 调整大小
func (f *TableSubFocus) Resize(rows, cols int) {
    f.maxRow = rows - 1
    f.maxCol = cols - 1

    // 确保焦点在范围内
    if f.row > f.maxRow {
        f.row = f.maxRow
    }
    if f.col > f.maxCol {
        f.col = f.maxCol
    }
}

// SetOnCellChange 设置单元格变化回调
func (f *TableSubFocus) SetOnCellChange(fn func(row, col int)) {
    f.onCellChange = fn
}

// SetOnEnterEdit 设置进入编辑回调
func (f *TableSubFocus) SetOnEnterEdit(fn func(row, col int) bool) {
    f.onEnterEdit = fn
}

// SetOnExitEdit 设置退出编辑回调
func (f *TableSubFocus) SetOnExitEdit(fn func(row, col int, cancelled bool)) {
    f.onExitEdit = fn
}

// isValidPosition 检查位置是否有效
func (f *TableSubFocus) isValidPosition(row, col int) bool {
    return row >= 0 && row <= f.maxRow && col >= 0 && col <= f.maxCol
}
```

### 2. TableSelection 选择区域

```go
// 位于: tui/framework/component/table_selection.go

package component

// TableSelection Table 选择区域
type TableSelection struct {
    fromRow int
    fromCol int
    toRow   int
    toCol   int
    active  bool
}

// NewTableSelection 创建选择区域
func NewTableSelection() *TableSelection {
    return &TableSelection{}
}

// Set 设置选择区域
func (s *TableSelection) Set(fromRow, fromCol, toRow, toCol int) {
    // 规范化坐标
    s.fromRow, s.toRow = minMax(fromRow, toRow)
    s.fromCol, s.toCol = minMax(fromCol, toCol)
    s.active = true
}

// Clear 清除选择
func (s *TableSelection) Clear() {
    s.active = false
}

// IsEmpty 是否为空
func (s *TableSelection) IsEmpty() bool {
    return !s.active
}

// Contains 检查是否包含指定单元格
func (s *TableSelection) Contains(row, col int) bool {
    if !s.active {
        return false
    }
    return row >= s.fromRow && row <= s.toRow &&
        col >= s.fromCol && col <= s.toCol
}

// Range 返回选择范围
func (s *TableSelection) Range() (fromRow, fromCol, toRow, toCol int) {
    return s.fromRow, s.fromCol, s.toRow, s.toCol
}

// Cells 返回所有选中的单元格
func (s *TableSelection) Cells() []CellPosition {
    if !s.active {
        return []CellPosition{}
    }

    cells := make([]CellPosition, 0)
    for row := s.fromRow; row <= s.toRow; row++ {
        for col := s.fromCol; col <= s.toCol; col++ {
            cells = append(cells, CellPosition{Row: row, Col: col})
        }
    }
    return cells
}

// CellCount 单元格数量
func (s *TableSelection) CellCount() int {
    if !s.active {
        return 0
    }
    return (s.toRow - s.fromRow + 1) * (s.toCol - s.fromCol + 1)
}

// IsSingleRow 是否单行选择
func (s *TableSelection) IsSingleRow() bool {
    return s.active && s.fromRow == s.toRow
}

// IsSingleCol 是否单列选择
func (s *TableSelection) IsSingleCol() bool {
    return s.active && s.fromCol == s.toCol
}

// IsSingleCell 是否单单元格选择
func (s *TableSelection) IsSingleCell() bool {
    return s.IsSingleRow() && s.IsSingleCol()
}

// CellPosition 单元格位置
type CellPosition struct {
    Row int
    Col int
}

func minMax(a, b int) (int, int) {
    if a <= b {
        return a, b
    }
    return b, a
}
```

### 3. TableCell 单元格

```go
// 位于: tui/framework/component/table_cell.go

package component

// TableCell 表格单元格
type TableCell struct {
    // 内容
    value interface{}

    // 样式
    style CellStyle

    // 状态
    readOnly bool
    visible  bool

    // 编辑器
    editor  Component // nil 表示使用默认编辑器
    editorType EditorType

    // 验证器
    validators []CellValidator

    // 格式化
    formatter func(interface{}) string
    parser    func(string) interface{}
}

// CellStyle 单元格样式
type CellStyle struct {
    Align     TextAlignment
    Foreground Color
    Background Color
    Bold      bool
    Italic    bool
}

// TextAlignment 文本对齐
type TextAlignment int

const (
    AlignLeft TextAlignment = iota
    AlignCenter
    AlignRight
)

// EditorType 编辑器类型
type EditorType int

const (
    // EditorText 文本编辑器
    EditorText EditorType = iota

    // EditorNumber 数字编辑器
    EditorNumber

    // EditorDate 日期编辑器
    EditorDate

    // EditorSelect 选择编辑器
    EditorSelect

    // EditorBoolean 布尔编辑器
    EditorBoolean

    // EditorCustom 自定义编辑器
    EditorCustom
)

// CellValidator 单元格验证器
type CellValidator interface {
    Validate(value interface{}) error
    Message() string
}

// NewTableCell 创建单元格
func NewTableCell(value interface{}) *TableCell {
    return &TableCell{
        value:     value,
        visible:   true,
        formatter: defaultFormatter,
    }
}

// Value 获取值
func (c *TableCell) Value() interface{} {
    return c.value
}

// SetValue 设置值
func (c *TableCell) SetValue(value interface{}) error {
    // 验证
    for _, validator := range c.validators {
        if err := validator.Validate(value); err != nil {
            return err
        }
    }

    c.value = value
    return nil
}

// String 获取字符串表示
func (c *TableCell) String() string {
    if c.formatter != nil {
        return c.formatter(c.value)
    }
    return defaultFormatter(c.value)
}

// SetString 设置字符串值
func (c *TableCell) SetString(s string) error {
    var value interface{}

    if c.parser != nil {
        value = c.parser(s)
    } else {
        value = s
    }

    return c.SetValue(value)
}

// SetReadOnly 设置只读
func (c *TableCell) SetReadOnly(readOnly bool) {
    c.readOnly = readOnly
}

// IsReadOnly 是否只读
func (c *TableCell) IsReadOnly() bool {
    return c.readOnly
}

// SetVisible 设置可见
func (c *TableCell) SetVisible(visible bool) {
    c.visible = visible
}

// IsVisible 是否可见
func (c *TableCell) IsVisible() bool {
    return c.visible
}

// SetStyle 设置样式
func (c *TableCell) SetStyle(style CellStyle) {
    c.style = style
}

// Style 获取样式
func (c *TableCell) Style() CellStyle {
    return c.style
}

// SetEditor 设置编辑器
func (c *TableCell) SetEditor(editor Component) {
    c.editor = editor
    c.editorType = EditorCustom
}

// SetEditorType 设置编辑器类型
func (c *TableCell) SetEditorType(editorType EditorType) {
    c.editorType = editorType
}

// Editor 获取编辑器
func (c *TableCell) Editor() Component {
    return c.editor
}

// EditorType 获取编辑器类型
func (c *TableCell) EditorType() EditorType {
    return c.editorType
}

// AddValidator 添加验证器
func (c *TableCell) AddValidator(validator CellValidator) {
    c.validators = append(c.validators, validator)
}

// SetFormatter 设置格式化函数
func (c *TableCell) SetFormatter(fn func(interface{}) string) {
    c.formatter = fn
}

// SetParser 设置解析函数
func (c *TableCell) SetParser(fn func(string) interface{}) {
    c.parser = fn
}

func defaultFormatter(value interface{}) string {
    if value == nil {
        return ""
    }
    return fmt.Sprintf("%v", value)
}
```

### 4. Table 组件集成

```go
// 位于: tui/framework/component/table.go

package component

// Table 表格组件
type Table struct {
    BaseComponent
    *Measurable
    *ThemeHolder

    // 数据
    columns []TableColumn
    rows    [][]*TableCell

    // 子焦点
    subFocus *TableSubFocus

    // 视口（用于虚拟滚动）
    viewport *Viewport

    // 表头
    showHeader bool
    headerHeight int

    // 样式
    gridLines    bool
    rowNumbers   bool
    alternateRow bool
}

// TableColumn 表格列
type TableColumn struct {
    Title      string
    Width      int
    MinWidth   int
    MaxWidth   int
    Resizable  bool
    Sortable   bool
    Visible    bool
}

// NewTable 创建表格
func NewTable(columns []TableColumn, rows int) *Table {
    table := &Table{
        columns:     columns,
        rows:        make([][]*TableCell, rows),
        subFocus:    NewTableSubFocus(rows, len(columns)),
        showHeader:  true,
        headerHeight: 1,
        gridLines:   true,
        rowNumbers:  false,
        alternateRow: true,
    }

    table.Measurable = NewMeasurable()
    table.ThemeHolder = NewThemeHolder(nil)

    // 设置子焦点回调
    table.setupSubFocusCallbacks()

    return table
}

// SetRowCount 设置行数
func (t *Table) SetRowCount(count int) {
    t.rows = make([][]*TableCell, count)
    t.subFocus.Resize(count, len(t.columns))
}

// SetCell 设置单元格
func (t *Table) SetCell(row, col int, cell *TableCell) {
    if row >= len(t.rows) {
        t.SetRowCount(row + 1)
    }

    if len(t.rows[row]) <= col {
        t.rows[row] = make([]*TableCell, len(t.columns))
    }

    t.rows[row][col] = cell
}

// GetCell 获取单元格
func (t *Table) GetCell(row, col int) *TableCell {
    if row < 0 || row >= len(t.rows) {
        return nil
    }
    if col < 0 || col >= len(t.rows[row]) {
        return nil
    }
    return t.rows[row][col]
}

// GetSubFocus 获取子焦点
func (t *Table) GetSubFocus() *TableSubFocus {
    return t.subFocus
}

// setupSubFocusCallbacks 设置子焦点回调
func (t *Table) setupSubFocusCallbacks() {
    t.subFocus.SetOnEnterEdit(func(row, col int) bool {
        cell := t.GetCell(row, col)
        if cell == nil || cell.IsReadOnly() {
            return false
        }

        // 创建编辑器
        editor := t.createEditor(cell)
        if editor == nil {
            return false
        }

        return t.subFocus.EnterEdit(editor)
    })

    t.subFocus.SetOnExitEdit(func(row, col int, cancelled bool) {
        if !cancelled {
            // 保存编辑内容
            editor := t.subFocus.Editor()
            if textInput, ok := editor.(*TextInput); ok {
                cell := t.GetCell(row, col)
                if cell != nil {
                    cell.SetString(textInput.Value())
                }
            }
        }
    })

    t.subFocus.SetOnCellChange(func(row, col int) {
        t.MarkDirty()
    })
}

// createEditor 创建编辑器
func (t *Table) createEditor(cell *TableCell) Component {
    // 如果有自定义编辑器，使用它
    if cell.Editor() != nil {
        return cell.Editor()
    }

    // 根据类型创建编辑器
    switch cell.EditorType() {
    case EditorText:
        return t.createTextEditor(cell)
    case EditorNumber:
        return t.createNumberEditor(cell)
    case EditorBoolean:
        return t.createBooleanEditor(cell)
    case EditorSelect:
        return t.createSelectEditor(cell)
    default:
        return t.createTextEditor(cell)
    }
}

// createTextEditor 创建文本编辑器
func (t *Table) createTextEditor(cell *TableCell) Component {
    editor := NewTextInput()
    editor.SetValue(cell.String())

    // 设置样式
    theme := t.GetTheme()
    editor.SetTheme(theme)

    return editor
}

// createNumberEditor 创建数字编辑器
func (t *Table) createNumberEditor(cell *TableCell) Component {
    editor := NewNumberInput()
    if num, ok := cell.Value().(float64); ok {
        editor.SetValue(num)
    }

    theme := t.GetTheme()
    editor.SetTheme(theme)

    return editor
}

// createBooleanEditor 创建布尔编辑器
func (t *Table) createBooleanEditor(cell *TableCell) Component {
    editor := NewCheckBox()
    if b, ok := cell.Value().(bool); ok {
        editor.SetChecked(b)
    }

    theme := t.GetTheme()
    editor.SetTheme(theme)

    return editor
}

// createSelectEditor 创建选择编辑器
func (t *Table) createSelectEditor(cell *TableCell) Component {
    // 假设验证器中有选项列表
    options := []string{}
    for _, validator := range cell.validators {
        if selectValidator, ok := validator.(*SelectValidator); ok {
            options = selectValidator.Options
            break
        }
    }

    editor := NewSelect(options)
    if selected, ok := cell.Value().(string); ok {
        editor.SelectOption(selected)
    }

    theme := t.GetTheme()
    editor.SetTheme(theme)

    return editor
}

// HandleAction 处理 Action
func (t *Table) HandleAction(a *action.Action) bool {
    // 如果正在编辑，转发给编辑器
    if t.subFocus.IsEditing() {
        editor := t.subFocus.Editor()
        if actionTarget, ok := editor.(ActionTarget); ok {
            handled := actionTarget.HandleAction(a)

            // 检查是否应该退出编辑
            if t.shouldExitEdit(a) {
                t.subFocus.ExitEdit(false)
            }

            return handled
        }
    }

    // 处理 Table 相关的 Action
    switch a.Type {
    case action.ActionNavigateUp, action.ActionNavigateDown,
         action.ActionNavigateLeft, action.ActionNavigateRight,
         action.ActionNavigatePageUp, action.ActionNavigatePageDown,
         action.ActionNavigateHome, action.ActionNavigateEnd,
         action.ActionNavigateFirst, action.ActionNavigateLast:
        direction := actionToDirection(a.Type)
        t.subFocus.Move(direction)
        return true

    case action.ActionEdit:
        t.subFocus.EnterEdit(nil)
        return true

    case action.ActionSubmit:
        // 进入编辑模式
        return t.subFocus.EnterEdit(nil)

    case action.ActionCancel:
        // 退出编辑模式
        if t.subFocus.IsEditing() {
            t.subFocus.ExitEdit(true)
            return true
        }
        return false

    case action.ActionSelectAll:
        t.subFocus.SelectAll()
        return true

    case action.ActionCopy:
        return t.copySelection()

    case action.ActionPaste:
        return t.pasteSelection()

    case action.ActionDelete:
        return t.deleteSelection()

    default:
        return false
    }
}

// shouldExitEdit 检查是否应该退出编辑
func (t *Table) shouldExitEdit(a *action.Action) bool {
    switch a.Type {
    case action.ActionSubmit:
        return true
    case action.ActionNavigateUp, action.ActionNavigateDown:
        return true
    case action.ActionCancel:
        return true
    default:
        return false
    }
}

// Paint 绘制表格
func (t *Table) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    theme := t.GetTheme()
    bounds := t.Bounds()

    y := bounds.Y

    // 绘制表头
    if t.showHeader {
        t.paintHeader(ctx, buf, y)
        y += t.headerHeight + 1
    }

    // 绘制行
    for row := 0; row < len(t.rows); row++ {
        // 检查是否可见（虚拟滚动）
        if !t.isRowVisible(row) {
            continue
        }

        t.paintRow(ctx, buf, row, y)
        y++
    }

    // 绘制编辑器
    if t.subFocus.IsEditing() {
        editor := t.subFocus.Editor()
        row, col := t.subFocus.Position()
        cellRect := t.getCellRect(row, col)

        editor.SetBounds(cellRect)
        if paintable, ok := editor.(Paintable); ok {
            paintable.Paint(ctx, buf)
        }
    }
}

// paintHeader 绘制表头
func (t *Table) paintHeader(ctx PaintContext, buf *runtime.CellBuffer, y int) {
    theme := t.GetTheme()
    headerStyle := theme.GetStyle("table.header")

    x := t.Bounds().X
    for _, col := range t.columns {
        if !col.Visible {
            continue
        }

        // 绘制列标题
        buf.DrawTextAligned(x, y, col.Width, col.Title,
            AlignCenter, headerStyle)

        x += col.Width
        if t.gridLines {
            buf.DrawText(x, y, "│", headerStyle)
            x++
        }
    }
}

// paintRow 绘制行
func (t *Table) paintRow(ctx PaintContext, buf *runtime.CellBuffer, row int, y int) {
    theme := t.GetTheme()
    bounds := t.Bounds()

    // 行样式
    rowStyle := theme.GetStyle("table.row")
    if t.alternateRow && row%2 == 1 {
        rowStyle = theme.GetStyle("table.alternate")
    }

    // 选中样式
    row, col := t.subFocus.Position()
    focusStyle := theme.GetStyle("table.focus")

    x := bounds.X
    for c := 0; c < len(t.columns); c++ {
        if !t.columns[c].Visible {
            continue
        }

        cell := t.GetCell(row, c)
        if cell == nil {
            continue
        }

        // 确定样式
        style := rowStyle
        if row == t.subFocus.row && c == t.subFocus.col {
            style = focusStyle
        } else if t.subFocus.IsSelected(row, c) {
            style = theme.GetStyle("table.selected")
        }

        // 绘制单元格
        text := cell.String()
        buf.DrawTextAligned(x, y, t.columns[c].Width, text,
            cell.style.Align, style)

        x += t.columns[c].Width
        if t.gridLines {
            buf.DrawText(x, y, "│", rowStyle)
            x++
        }
    }
}

// getCellRect 获取单元格矩形
func (t *Table) getCellRect(row, col int) runtime.Rect {
    bounds := t.Bounds()

    x := bounds.X
    for c := 0; c < col; c++ {
        if t.columns[c].Visible {
            x += t.columns[c].Width
            if t.gridLines {
                x++
            }
        }
    }

    y := bounds.Y
    if t.showHeader {
        y += t.headerHeight + 1
    }
    y += row

    width := t.columns[col].Width
    height := 1

    return runtime.Rect{
        X:      x,
        Y:      y,
        Width:  width,
        Height: height,
    }
}

// isRowVisible 检查行是否可见
func (t *Table) isRowVisible(row int) bool {
    if t.viewport == nil {
        return true
    }

    start, end := t.viewport.GetVisibleRange()
    return row >= start && row < end
}
```

## 使用示例

### 示例 1：基础表格

```go
// ✅ 创建基础表格
table := component.NewTable([]component.TableColumn{
    {Title: "Name", Width: 20},
    {Title: "Age", Width: 10},
    {Title: "Email", Width: 30},
}, 3)

// 设置数据
table.SetCell(0, 0, component.NewTableCell("John Doe"))
table.SetCell(0, 1, component.NewTableCell(30))
table.SetCell(0, 2, component.NewTableCell("john@example.com"))

table.SetCell(1, 0, component.NewTableCell("Jane Smith"))
table.SetCell(1, 1, component.NewTableCell(25))
table.SetCell(1, 2, component.NewTableCell("jane@example.com"))
```

### 示例 2：可编辑表格

```go
// ✅ 可编辑表格
table := component.NewTable(columns, 10)

// 设置单元格为可编辑
for row := 0; row < 10; row++ {
    for col := 0; col < 3; col++ {
        cell := component.NewTableCell("")
        cell.SetEditorType(component.EditorText)
        table.SetCell(row, col, cell)
    }
}

// 用户可以：
// 1. 导航到单元格（方向键）
// 2. 进入编辑模式（Enter）
// 3. 编辑内容
// 4. 退出并保存（Enter 或方向键）
```

### 示例 3：带验证的表格

```go
// ✅ 添加验证
emailCell := component.NewTableCell("")

// 添加邮箱验证
emailCell.AddValidator(&EmailValidator{})

// 只读单元格
readOnlyCell := component.NewTableCell("Fixed Value")
readOnlyCell.SetReadOnly(true)
```

### 示例 4：选择和复制

```go
// ✅ 选择区域
// 用户可以：
// 1. Shift + 方向键 - 扩展选择
// 2. Ctrl+A - 全选
// 3. Ctrl+C - 复制选择

// 获取选择内容
selection := table.GetSubFocus().GetSelection()
cells := selection.Cells()

for _, pos := range cells {
    cell := table.GetCell(pos.Row, pos.Col)
    fmt.Println(cell.String())
}
```

### 示例 5：虚拟滚动

```go
// ✅ 大数据量表格
table := component.NewTable(columns, 100000)

// 设置视口
table.SetViewport(component.NewViewport(20, 1))

// Table 会自动只渲染可见行
```

## 测试

```go
// 位于: tui/framework/component/table_test.go

func TestTableSubFocus(t *testing.T) {
    table := NewTestTable(5, 3)
    subFocus := table.GetSubFocus()

    // 测试移动
    assert.True(t, subFocus.Move(focus.DirectionRight))
    row, col := subFocus.Position()
    assert.Equal(t, 0, row)
    assert.Equal(t, 1, col)

    // 测试边界
    subFocus.Move(focus.DirectionUp)
    row, col = subFocus.Position()
    assert.Equal(t, 0, row) // 不会超出上边界
}

func TestTableSelection(t *testing.T) {
    table := NewTestTable(5, 3)
    subFocus := table.GetSubFocus()

    // 选择区域
    subFocus.SetSelection(0, 0, 2, 2)

    // 验证选择
    assert.True(t, subFocus.IsSelected(1, 1))
    assert.False(t, subFocus.IsSelected(3, 3))
}

func TestTableEdit(t *testing.T) {
    table := NewTestEditableTable(5, 3)
    subFocus := table.GetSubFocus()

    // 移动到 (1, 1)
    subFocus.SetPosition(1, 1)

    // 进入编辑
    assert.True(t, subFocus.EnterEdit(nil))
    assert.True(t, subFocus.IsEditing())

    // 退出编辑
    subFocus.ExitEdit(false)
    assert.False(t, subFocus.IsEditing())
}
```

## 总结

Table 子焦点系统提供：

1. **单元格导航**: 支持在单元格间移动焦点
2. **内联编辑**: 支持单元格内编辑
3. **选择模式**: 支持选择单个或多个单元格
4. **键盘快捷键**: 支持常用的快捷键
5. **虚拟滚动**: 支持大数据量表格
6. **验证**: 支持单元格级别验证

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [FOCUS_SYSTEM.md](./FOCUS_SYSTEM.md) - 焦点系统
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [VIRTUAL_SCROLL.md](./VIRTUAL_SCROLL.md) - 虚拟滚动
