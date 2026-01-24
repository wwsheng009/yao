# Component Design - Phase 2

## 概述

本文档详细描述第二阶段（Phase 2）的组件设计，包括所有基础和高级组件的完整规格说明。

## 组件层次结构

```
Component (接口)
    │
    ├── BaseComponent (基础实现)
    │
    ├── Display Components (显示组件)
    │   ├── Text
    │   ├── Paragraph
    │   ├── Code
    │   └── Separator
    │
    ├── Interactive Components (交互组件)
    │   ├── Button
    │   ├── Checkbox
    │   ├── Radio
    │   ├── Toggle
    │   └── Slider
    │
    ├── Input Components (输入组件)
    │   ├── TextInput
    │   ├── TextArea
    │   ├── PasswordInput
    │   ├── NumberInput
    │   └── Select
    │
    ├── Container Components (容器组件)
    │   ├── Box
    │   ├── Flex
    │   ├── Grid
    │   ├── Stack
    │   └── Tabs
    │
    ├── Display Collections (集合显示)
    │   ├── List
    │   ├── Table
    │   ├── Tree
    │   └── Calendar
    │
    ├── Widget Components (小部件)
    │   ├── ProgressBar
    │   ├── Spinner
    │   ├── Meter
    │   ├── Gauge
    │   └── Chart
    │
    └── Form Components (表单组件)
        ├── Form
        ├── Field
        ├── Label
        ├── Validation
        └── Schema
```

## 基础组件

### 1. Text (文本显示)

#### 接口定义

```go
// 位于: tui/framework/component/display/text.go

package display

// Text 文本显示组件
type Text struct {
    BaseComponent

    // 内容
    content string
    lines   []string

    // 样式
    style   style.Style

    // 布局
    align   TextAlign  // Left, Center, Right, Justify
    wrap    bool       // 自动换行
    maxLines int       // 最大行数 (0 = 无限制)

    // 截断
    truncate string    // "head", "tail", "middle", ""
    ellipsis string    // 省略符号 (默认 "...")
}

// TextAlign 文本对齐
type TextAlign int

const (
    AlignLeft TextAlign = iota
    AlignCenter
    AlignRight
    AlignJustify
)
```

#### API

```go
// 创建
func NewText(content string) *Text
func NewStyledText(content string, style style.Style) *Text

// 配置 (链式)
func (t *Text) WithAlign(align TextAlign) *Text
func (t *Text) WithWrap(wrap bool) *Text
func (t *Text) WithMaxLines(max int) *Text
func (t *Text) WithTruncate(mode string) *Text
func (t *Text) WithEllipsis(ellipsis string) *Text

// 操作
func (t *Text) SetContent(content string)
func (t *Text) GetContent() string
func (t *Text) Append(text string)
func (t *Text) Clear()

// 测量
func (t *Text) Measure(constraints Constraints) Size
```

#### 渲染示例

```
┌─────────────────────────────────┐
│ Left aligned text               │
│         Center text             │
│                    Right text   │
│                                 │
│ Long text that wraps to the    │
│ next line when the container is │
│ not wide enough.                │
└─────────────────────────────────┘
```

### 2. Box (盒子容器)

#### 接口定义

```go
// 位于: tui/framework/component/layout/box.go

package layout

// Box 盒子容器
type Box struct {
    BaseContainer

    // 边框
    border  Border
    borderStyle style.Style

    // 间距
    padding BoxSpacing  // 内边距
    margin  BoxSpacing  // 外边距

    // 尺寸
    width  Dimension
    height Dimension

    // 背景
    bgColor Color
}

// Border 边框定义
type Border struct {
    Top    bool
    Bottom bool
    Left   bool
    Right  bool

    // 样式
    Style BorderStyle  // Normal, Rounded, Double, Thick, Hidden
}

// BorderStyle 边框样式
type BorderStyle int

const (
    BorderNormal BorderStyle = iota
    BorderRounded
    BorderDouble
    BorderThick
    BorderHidden
    BorderCustom  // 使用自定义字符
)

// BoxSpacing 间距
type BoxSpacing struct {
    Top    int
    Right  int
    Bottom int
    Left   int
}

// Dimension 尺寸
type Dimension struct {
    Value   int
    Unit    DimensionUnit
    Percent float64  // 当 Unit = Percent
}

type DimensionUnit int

const (
    UnitPixel DimensionUnit = iota
    UnitPercent
    UnitFlex    // flex-grow
    UnitAuto    // 自动计算
)
```

#### API

```go
// 创建
func NewBox() *Box

// 边框
func (b *Box) Border(enabled bool) *Box
func (b *Box) BorderStyle(style BorderStyle) *Box
func (b *Box) BorderColor(color Color) *Box

// 间距
func (b *Box) Padding(all int) *Box
func (b *Box) PaddingV(vertical int) *Box
func (b *Box) PaddingH(horizontal int) *Box
func (b *Box) PaddingTop(v int) *Box
func (b *Box) PaddingRight(v int) *Box
func (b *Box) PaddingBottom(v int) *Box
func (b *Box) PaddingLeft(v int) *Box

// 尺寸
func (b *Box) Width(w int) *Box
func (b *Box) Height(h int) *Box
func (b *Box) WidthPercent(p float64) *Box
func (b *Box) HeightPercent(p float64) *Box
func (b *Box) Flex() *Box

// 背景
func (b *Box) Background(color Color) *Box
```

#### 渲染示例

```
┌─────────────────────────────────┐
│ ┌───────┐                       │
│ │       │  Margin (外边距)        │
│ │ ┌───┐ │                       │
│ │ │   │ │  Padding (内边距)      │
│ │ │ X │ │  Content              │
│ │ │   │ │                       │
│ │ └───┘ │                       │
│ │       │                       │
│ └───────┘                       │
│                                 │
└─────────────────────────────────┘
```

### 3. ProgressBar (进度条)

#### 接口定义

```go
// 位于: tui/framework/component/widget/progress.go

package widget

// ProgressBar 进度条
type ProgressBar struct {
    InteractiveComponent

    // 值
    value    float64  // 当前值
    max      float64  // 最大值

    // 显示
    showPercent bool
    showValue   bool

    // 样式
    fillStyle   style.Style
    emptyStyle  style.Style
    textStyle   style.Style

    // 不确定状态
    indeterminate bool
    indeterminatePos int

    // 方向
    horizontal bool
}

// ProgressBarStyle 进度条样式
type ProgressBarStyle struct {
    Fill    style.Style
    Empty   style.Style
    Text    style.Style

    // 字符
    FillChar    rune  // 默认 '█'
    EmptyChar   rune  // 默认 '░'

    // 边框
    ShowBorder  bool
    BorderStyle BorderStyle
}
```

#### API

```go
// 创建
func NewProgressBar() *ProgressBar
func NewProgressBarMax(max float64) *ProgressBar

// 值操作
func (p *ProgressBar) SetValue(value float64)
func (p *ProgressBar) GetValue() float64
func (p *ProgressBar) SetMax(max float64)
func (p *ProgressBar) Increment(delta float64)
func (p *ProgressBar) GetPercent() float64

// 显示
func (p *ProgressBar) ShowPercent(show bool) *ProgressBar
func (p *ProgressBar) ShowValue(show bool) *ProgressBar

// 样式
func (p *ProgressBar) SetFillStyle(style style.Style)
func (p *ProgressBar) SetEmptyStyle(style style.Style)

// 不确定状态
func (p *ProgressBar) SetIndeterminate(indeterminate bool)
func (p *ProgressBar) IsIndeterminate() bool
```

#### 渲染示例

```
确定状态:
┌─────────────────────────────────┐
│ Progress: [████████████░░░░] 75%│
│ Progress: [████████████████] 100%│
│         [████████░░░░░░░░]  40% │
└─────────────────────────────────┘

不确定状态:
┌─────────────────────────────────┐
│ Loading:  [░░░████░░░░░░░░░░░░] │
│           [░░░░░░░██████░░░░░░] │
└─────────────────────────────────┘
```

## 输入组件

### 4. TextInput (文本输入)

#### 接口定义

```go
// 位于: tui/framework/component/input/textinput.go

package input

// TextInput 文本输入框
type TextInput struct {
    InteractiveComponent

    // 值
    value     string
    cursor    int
    selection Selection

    // 配置
    placeholder string
    password    bool
    echo        rune  // 密码遮罩字符

    // 限制
    maxLength   int
    validator   Validator

    // 样式
    placeholderStyle style.Style
    cursorStyle     style.Style
    selectionStyle  style.Style

    // 历史
    history    []string
    historyPos int
}

// Selection 文本选择
type Selection struct {
    Start int
    End   int
    Active bool
}

// Validator 验证器
type Validator interface {
    Validate(value string) error
}

// InputValidator 输入验证函数
type InputValidator func(string) error

func (v InputValidator) Validate(value string) error {
    return v(value)
}
```

#### API

```go
// 创建
func NewTextInput() *TextInput
func NewTextInputPlaceholder(placeholder string) *TextInput

// 值操作
func (t *TextInput) SetValue(value string)
func (t *TextInput) GetValue() string
func (t *TextInput) Clear()

// 光标
func (t *TextInput) SetCursor(pos int)
func (t *TextInput) GetCursor() int
func (t *TextInput) CursorStart()
func (t *TextInput) CursorEnd()

// 选择
func (t *TextInput) SelectAll()
func (t *TextInput) ClearSelection()
func (t *TextInput) GetSelection() string
func (t *TextInput) DeleteSelection()

// 配置
func (t *TextInput) SetPlaceholder(text string)
func (t *TextInput) SetPassword(enabled bool)
func (t *TextInput) SetEcho(char rune)
func (t *TextInput) SetMaxLength(max int)

// 验证
func (t *TextInput) SetValidator(validator Validator)
func (t *TextInput) Validate() error
func (t *TextInput) IsValid() bool

// 历史
func (t *TextInput) HistoryAdd(value string)
func (t *TextInput) HistoryPrev()
func (t *TextInput) HistoryNext()

// 样式
func (t *TextInput) SetPlaceholderStyle(style style.Style)
func (t *TextInput) SetCursorStyle(style style.Style)
func (t *TextInput) SetSelectionStyle(style style.Style)
```

#### 键盘快捷键

| 按键 | 动作 |
|------|------|
| `←` `→` | 移动光标 |
| `Ctrl+A` | 光标到开头 |
| `Ctrl+E` | 光标到末尾 |
| `Backspace` | 删除前一个字符 |
| `Delete` | 删除后一个字符 |
| `Ctrl+W` | 删除前一个单词 |
| `Ctrl+K` | 删除到行尾 |
| `Ctrl+U` | 删除到行首 |
| `Ctrl+C` | 复制选择 |
| `Ctrl+X` | 剪切选择 |
| `Ctrl+V` | 粘贴 |
| `Ctrl+A` | 全选 |
| `↑` `↓` | 历史记录 |

#### 渲染示例

```
普通输入:
┌─────────────────────────────────┐
│ Username: [john_doe           ] │
│          └─────光标─────┘         │
└─────────────────────────────────┘

密码输入:
┌─────────────────────────────────┐
│ Password: [••••••••            ] │
└─────────────────────────────────┘

占位符:
┌─────────────────────────────────┐
│ Search: [Type to search...    ] │
│         └─────灰色─────┘         │
└─────────────────────────────────┘

选择文本:
┌─────────────────────────────────┐
│ Input: [Hello[ world]         ] │
│        └─────反白─────┘          │
└─────────────────────────────────┘
```

### 5. TextArea (多行输入)

#### 接口定义

```go
// 位于: tui/framework/component/input/textarea.go

package input

// TextArea 多行文本输入
type TextArea struct {
    InteractiveComponent

    // 内容
    lines    []string
    cursor   CursorPos

    // 滚动
    scrollX  int
    scrollY  int

    // 配置
    lineNumbers bool
    wordWrap    bool
    syntax      SyntaxHighlighter

    // 标尺
    showRuler    bool
    rulerColumns []int
}

// CursorPos 光标位置
type CursorPos struct {
    Line int
    Col  int
}

// SyntaxHighlighter 语法高亮
type SyntaxHighlighter interface {
    Highlight(line string, col int) []style.Style
}

// TextAreaStyle 多行输入样式
type TextAreaStyle struct {
    LineNumberStyle style.Style
    CurrentLineStyle style.Style
    RulerStyle      style.Style
}
```

#### API

```go
// 创建
func NewTextArea() *TextArea
func NewTextAreaLines(lines []string) *TextArea

// 内容
func (t *TextArea) SetLines(lines []string)
func (t *TextArea) GetLines() []string
func (t *TextArea) GetText() string
func (t *TextArea) SetText(text string)
func (t *TextArea) AppendLine(line string)
func (t *TextArea) InsertLine(pos int, line string)
func (t *TextArea) DeleteLine(pos int)

// 光标
func (t *TextArea) SetCursor(line, col int)
func (t *TextArea) GetCursor() CursorPos

// 滚动
func (t *TextArea) ScrollTo(line, col int)
func (t *TextArea) ScrollBy(deltaLine, deltaCol int)

// 配置
func (t *TextArea) SetLineNumbers(show bool)
func (t *TextArea) SetWordWrap(enabled bool)
func (t *TextArea) SetSyntax(highlighter SyntaxHighlighter)
func (t *TextArea) SetRuler(columns []int)
```

#### 渲染示例

```
带行号:
┌─────────────────────────────────┐
│ 1 │ func hello() {              │
│ 2 │     fmt.Println("Hello")    │
│ 3 │ }                           │
│   └─光标                         │
└─────────────────────────────────┘

带标尺:
┌─────────────────────────────────┐
│   │    10    20    30    40     │
│ 1 │ func hello() {              │
│ 2 │     fmt.Println("Hello")    │
│ 3 │ }                           │
└─────────────────────────────────┘
```

## 集合组件

### 6. List (列表)

#### 接口定义

```go
// 位于: tui/framework/component/display/list.go

package display

// List 列表组件
type List struct {
    InteractiveComponent

    // 数据
    items    []ListItem
    cursor   int
    selected map[int]bool  // 多选

    // 滚动
    offset   int

    // 配置
    showCursor bool
    multiSelect bool
    showFilter bool  // 过滤器

    // 样式
    cursorStyle style.Style
    selectedStyle style.Style
    dimStyle    style.Style  // 非选中项样式
}

// ListItem 列表项
type ListItem struct {
    Text      string
    Value     interface{}
    Disabled  bool
    Icon      rune  // 前缀图标
    Secondary string  // 次要文本
}

// ListStyle 列表样式
type ListStyle struct {
    Cursor      style.Style
    Selected    style.Style
    Dim         style.Style

    // 前缀
    ShowPrefix  bool
    Prefix      map[string]rune  // "selected": "•", "unselected": " "

    // 分隔符
    ShowDivider bool
    Divider     rune
}
```

#### API

```go
// 创建
func NewList() *List
func NewListItems(items []string) *List

// 数据
func (l *List) SetItems(items []ListItem)
func (l *List) AddItem(item ListItem)
func (l *List) RemoveItem(index int)
func (l *List) ClearItems()
func (l *List) GetItem(index int) ListItem
func (l *List) GetItemCount() int

// 选择
func (l *List) SetCursor(index int)
func (l *List) GetCursor() int
func (l *List) CursorDown()
func (l *List) CursorUp()
func (l *List) PageDown()
func (l *List) PageUp()
func (l *List) Top()
func (l *List) Bottom()

// 多选
func (l *List) SetMultiSelect(enabled bool)
func (l *List) Select(index int)
func (l *List) Deselect(index int)
func (l *List) ToggleSelect(index int)
func (l *List) SelectAll()
func (l *List) DeselectAll()
func (l *List) GetSelected() []int
func (l *List) GetSelectedItems() []ListItem
```

#### 渲染示例

```
单选:
┌─────────────────────────────────┐
│ ▶ Item 1                        │
│   Item 2                        │
│   Item 3                        │
│   Item 4                        │
└─────────────────────────────────┘

多选:
┌─────────────────────────────────┐
│ ▶ ✓ Item 1                      │
│   ✓ Item 2                      │
│     Item 3                      │
│     Item 4                      │
└─────────────────────────────────┘

带次要文本:
┌─────────────────────────────────┐
│ ▶ ✓ John Doe                    │
│     john@example.com            │
│   ✓ Jane Smith                  │
│     jane@example.com            │
│     Bob Johnson                 │
│     bob@example.com            │
└─────────────────────────────────┘
```

### 7. Table (表格)

#### 接口定义

```go
// 位于: tui/framework/component/display/table.go

package display

// Table 表格组件
type Table struct {
    InteractiveComponent

    // 列定义
    columns []Column

    // 数据
    rows    []TableRow
    cursor  int
    selected map[int]bool

    // 滚动
    offsetX  int
    offsetY  int

    // 排序
    sortBy   int
    sortDesc bool

    // 固定
    fixedRows int
    fixedCols int

    // 配置
    showHeader bool
    showCursor bool
    multiSelect bool
    striped bool

    // 样式
    headerStyle  style.Style
    cursorStyle  style.Style
    selectedStyle style.Style
    stripedStyle style.Style
}

// Column 列定义
type Column struct {
    Title    string
    Width    int  // 0 = 自动
    Align    TextAlign
    Sortable bool
    Visible  bool
}

// TableRow 表格行
type TableRow struct {
    Cells   []TableCell
    Data    interface{}  // 原始数据
    Expander *Expander  // 展开子行
}

// TableCell 表格单元格
type TableCell struct {
    Text    string
    Style   style.Style
    ColSpan int
    RowSpan int
}

// Expander 展开器
type Expander struct {
    Expanded bool
    Content []TableRow
}
```

#### API

```go
// 创建
func NewTable() *Table
func NewTableColumns(columns []Column) *Table

// 列
func (t *Table) SetColumns(columns []Column)
func (t *Table) AddColumn(column Column)
func (t *Table) RemoveColumn(index int)
func (t *Table) GetColumn(index int) Column

// 行
func (t *Table) SetRows(rows []TableRow)
func (t *Table) AddRow(row TableRow)
func (t *Table) RemoveRow(index int)
func (t *Table) ClearRows()
func (t *Table) GetRow(index int) TableRow
func (t *Table) GetRowCount() int

// 单元格
func (t *Table) SetCell(row, col int, cell TableCell)
func (t *Table) GetCell(row, col int) TableCell

// 选择
func (t *Table) SetCursor(row, col int)
func (t *Table) GetCursor() (row, col int)
func (t *Table) SelectRow(index int)
func (t *Table) GetSelectedRows() []int

// 排序
func (t *Table) SortBy(column int, desc bool)
func (t *Table) GetSort() (column int, desc bool)

// 滚动
func (t *Table) ScrollTo(row, col int)
func (t *Table) ScrollBy(deltaRow, deltaCol int)

// 固定
func (t *Table) SetFixedRows(count int)
func (t *Table) SetFixedCols(count int)
```

#### 渲染示例

```
基本表格:
┌─────────────────────────────────┐
│ ┌───┬────────────┬──────────┐ │
│ │ # │ Name       │ Status   │ │
│ ├───┼────────────┼──────────┤ │
│ │ ▶ │ Item 1     │ Active   │ │
│ │   │ Item 2     │ Inactive │ │
│ │   │ Item 3     │ Active   │ │
│ └───┴────────────┴──────────┘ │
└─────────────────────────────────┘

固定表头和首列:
┌─────────────────────────────────┐
│ ┌───┬────────┬────────┬──────┐ │
│ │ ■ │ Col 1  │ Col 2  │ Col 3│ │ ← 固定
│ ├───┼────────┼────────┼──────┤ │
│ │ ▶ │ A      │ B      │ C    │ │ ← 固定
│ ├───┼────────┼────────┼──────┤ │
│ │   │ D      │ E      │ F    │ │
│ │   │ G      │ H      │ I    │ │
│ └───┴────────┴────────┴──────┘ │
└─────────────────────────────────┘

可展开:
┌─────────────────────────────────┐
│ ┌──────┬────────────┐          │
│ │ ▼    │ Category 1  │          │
│ │      ├───┬────────┤          │
│ │      │ • │ Item 1.1│          │
│ │      │ • │ Item 1.2│          │
│ ├──────┼────────────┤          │
│ │ ▶    │ Category 2  │          │
│ └──────┴────────────┘          │
└─────────────────────────────────┘
```

## 表单组件

### 8. Form (表单)

#### 接口定义

```go
// 位于: tui/framework/component/form/form.go

package form

// Form 表单容器
type Form struct {
    ContainerComponent

    // 字段
    fields   []Field
    fieldMap map[string]Field

    // 导航
    current  int

    // 按钮
    buttons []Button

    // 提交
    onSubmit func(data map[string]interface{}) error
    onCancel func()

    // 布局
    layout   FormLayout  // Vertical, Horizontal, Grid
}

// Field 表单字段接口
type Field interface {
    Component

    // 字段信息
    Name() string
    Label() string
    Value() interface{}
    SetValue(value interface{}) error

    // 验证
    Validate() error
    IsValid() bool
    SetValidator(validator Validator)

    // 状态
    SetRequired(required bool)
    IsRequired() bool
    SetEnabled(enabled bool)
    IsEnabled() bool
}

// Validator 验证器
type Validator interface {
    Validate(value interface{}) error
    Message() string
}

// FormLayout 表单布局
type FormLayout int

const (
    FormLayoutVertical FormLayout = iota  // 标签在上，输入在下
    FormLayoutHorizontal                   // 标签在左，输入在右
    FormLayoutGrid                        // 网格布局
)

// FormBuilder 表单构建器
type FormBuilder struct {
    form *Form
}
```

#### API

```go
// 创建
func NewForm() *Form
func NewFormBuilder() *FormBuilder

// 字段
func (f *Form) AddField(field Field)
func (f *Form) RemoveField(name string)
func (f *Form) GetField(name string) (Field, bool)
func (f *Form) GetFields() []Field

// 导航
func (f *Form) NextField()
func (f *Form) PrevField()
func (f *Form) SetCurrent(index int)
func (f *Form) GetCurrent() Field

// 数据
func (f *Form) GetValue(name string) (interface{}, error)
func (f *Form) SetValue(name string, value interface{}) error
func (f *Form) GetValues() map[string]interface{}
func (f *Form) SetValues(data map[string]interface{})

// 验证
func (f *Form) Validate() error
func (f *Form) IsValid() bool

// 提交
func (f *Form) SetOnSubmit(handler func(map[string]interface{}) error)
func (f *Form) Submit() error
func (f *Form) Cancel()
```

#### 预定义字段

```go
// 位于: tui/framework/component/form/fields.go

// TextField 文本字段
type TextField struct {
    BaseField
    input *input.TextInput
}

// NumberField 数字字段
type NumberField struct {
    BaseField
    input   *input.TextInput
    min     float64
    max     float64
    step    float64
}

// SelectField 选择字段
type SelectField struct {
    BaseField
    list    *display.List
    options []SelectOption
}

// CheckboxField 复选框字段
type CheckboxField struct {
    BaseField
    checkbox *interactive.Checkbox
}

// DateField 日期字段
type DateField struct {
    BaseField
    input   *input.TextInput
    format  string
}

// BaseField 字段基类
type BaseField struct {
    name     string
    label    string
    required bool
    enabled  bool
    validator Validator
    help     string
}
```

#### 预定义验证器

```go
// 位于: tui/framework/component/form/validation.go

// Required 必填验证
func Required() Validator

// MinLength 最小长度
func MinLength(min int) Validator

// MaxLength 最大长度
func MaxLength(max int) Validator

// Email 邮箱验证
func Email() Validator

// Pattern 正则验证
func Pattern(pattern string) Validator

// Range 范围验证
func Range(min, max float64) Validator

// Custom 自定义验证
func Custom(fn func(interface{}) error) Validator

// AllOf 所有验证都通过
func AllOf(validators ...Validator) Validator

// AnyOf 任一验证通过
func AnyOf(validators ...Validator) Validator
```

#### 渲染示例

```
垂直布局:
┌─────────────────────────────────┐
│          User Form              │
├─────────────────────────────────┤
│ Name:                          │
│ [____________________________] │
│                                │
│ Email:                         │
│ [____________________________] │
│                                │
│ Age:                           │
│ [____]                         │
│                                │
│ [ ] Subscribe to newsletter    │
│                                │
│ [Submit]  [Cancel]             │
└─────────────────────────────────┘

水平布局:
┌─────────────────────────────────┐
│ Name:     [_______________]     │
│ Email:    [_______________]     │
│ Age:      [____]                │
│                [Submit] [Cancel]│
└─────────────────────────────────┘
```

## 高级特性

### 9. 键盘导航

所有交互组件支持一致的键盘导航：

| 按键 | 动作 |
|------|------|
| `Tab` | 下一个字段/组件 |
| `Shift+Tab` | 上一个字段/组件 |
| `Enter` | 提交/确认 |
| `Escape` | 取消/关闭 |
| `↑` `↓` | 上/下移动 |
| `←` `→` | 左/右移动 |
| `PageUp` `PageDown` | 翻页 |
| `Home` `End` | 跳到首/尾 |

### 10. 焦点管理

```go
// FocusManager 焦点管理器
type FocusManager struct {
    components []Focusable
    current    int
    cyclic     bool  // 循环导航
}

func (f *FocusManager) Register(component Focusable)
func (f *FocusManager) Next()
func (f *FocusManager) Prev()
func (f *FocusManager) SetCurrent(index int)
func (f *FocusManager) GetCurrent() Focusable
```

### 11. 主题支持

所有组件支持主题：

```go
// Theme 主题
type Theme struct {
    // 颜色
    Primary   Color
    Secondary Color
    Success   Color
    Warning   Color
    Error     Color
    Info      Color

    // 组件样式
    Text      style.Style
    Button    style.Style
    Input     style.Style
    List      ListStyle
    Table     TableStyle
    Form      FormStyle
}

// 预设主题
var LightTheme Theme
var DarkTheme Theme
var DraculaTheme Theme
var NordTheme Theme
```

### 12. 可访问性

```go
// Accessible 可访问性接口
type Accessible interface {
    Component

    // 屏幕阅读器
    SetLabel(label string)
    GetLabel() string
    SetHint(hint string)
    GetHint() string

    // 状态
    SetRole(role AriaRole)
    GetRole() AriaRole
}

type AriaRole int

const (
    RoleButton AriaRole = iota
    RoleInput
    RoleList
    RoleListItem
    RoleTable
    RoleGrid
    RoleDialog
)
```

## 组件生命周期

```
     Created
        │
        ▼
    Mounting ────┐
        │         │ Mount(parent)
        ▼         │
     Mounted ─────┘
        │
        ▼
    ┌────┴────┐
    │         │
 Update    Render
    │         │
    └────┬────┘
         │
         ▼
    Updating ───┐
        │        │ HandleEvent
        ▼        │
    Updated ─────┘
        │
        ▼
   Unmounting
        │
        ▼
    Unmounted
        │
        ▼
    Destroyed
```

## 测试示例

```go
func TestTextInput(t *testing.T) {
    input := NewTextInput()
    input.SetSize(20, 1)

    // 测试输入
    input.HandleEvent(&KeyEvent{Key: 'H'})
    input.HandleEvent(&KeyEvent{Key: 'e'})
    input.HandleEvent(&KeyEvent{Key: 'l'})
    input.HandleEvent(&KeyEvent{Key: 'l'})
    input.HandleEvent(&KeyEvent{Key: 'o'})

    assert.Equal(t, "Hello", input.GetValue())
    assert.Equal(t, 5, input.GetCursor())

    // 测试删除
    input.HandleEvent(&KeyEvent{Special: KeyBackspace})
    assert.Equal(t, "Hell", input.GetValue())
    assert.Equal(t, 4, input.GetCursor())

    // 测试验证
    input.SetValidator(MinLength(5))
    assert.False(t, input.IsValid())
}

func TestTable(t *testing.T) {
    table := NewTable()
    table.SetColumns([]Column{
        {Title: "Name", Width: 20},
        {Title: "Age", Width: 10},
    })
    table.AddRow(TableRow{
        Cells: []TableCell{
            {Text: "John"},
            {Text: "30"},
        },
    })

    assert.Equal(t, 1, table.GetRowCount())

    // 测试选择
    table.SelectRow(0)
    selected := table.GetSelectedRows()
    assert.Equal(t, []int{0}, selected)
}
```
