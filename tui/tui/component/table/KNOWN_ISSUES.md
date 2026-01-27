# Table 组件已知问题：选中行高亮不完整

## 问题描述

在 Table 组件中，当设置了 `Cell` 样式时，选中行的 `Selected` 样式只会应用到第一列，导致整行高亮不完整。

## 复现步骤

1. 创建一个 Table 组件，设置 `CellStyle` 和 `SelectedStyle`
2. 选中任意一行
3. 观察到只有第一列应用了选中样式，其他列保持 `CellStyle` 样式

## 根本原因

这是 `github.com/charmbracelet/bubbles/table` 库的实现限制。

### 源码分析

在 `bubbles/table` 的 `renderRow` 函数中（v0.20.0，第 395-413 行）：

```go
func (m *Model) renderRow(r int) string {
	s := make([]string, 0, len(m.cols))
	for i, value := range m.rows[r] {
		if m.cols[i].Width <= 0 {
			continue
		}
		style := lipgloss.NewStyle().Width(m.cols[i].Width).MaxWidth(m.cols[i].Width).Inline(true)
		// ⚠️ 问题：每个单元格先应用 Cell 样式
		renderedCell := m.styles.Cell.Render(style.Render(runewidth.Truncate(value, m.cols[i].Width, "…")))
		s = append(s, renderedCell)
	}

	// 拼接所有单元格
	row := lipgloss.JoinHorizontal(lipgloss.Top, s...)

	// 如果是选中行，应用 Selected 样式
	if r == m.cursor {
		return m.styles.Selected.Render(row)  // ⚠️ 但此时 Cell 样式已经应用
	}

	return row
}
```

### 样式嵌套问题

1. **第一步**（第 406 行）：每个单元格内容被 `m.styles.Cell` 渲染，设置前景色/背景色
2. **第二步**（第 410 行）：已渲染的单元格用 `lipgloss.JoinHorizontal` 拼接
3. **第三步**（第 412 行）：对整行应用 `m.styles.Selected` 样式

由于 lipgloss 的样式是**嵌套应用**的，`Cell` 样式的渲染结果作为 `Selected` 样式的输入。这导致：
- 如果 `Cell` 设置了前景色，`Selected` 的前景色可能无法生效
- 如果 `Cell` 设置了背景色，`Selected` 的背景色会被覆盖或混合
- 只有当 `Cell` 样式为空或只设置 padding 时，`Selected` 样式才能完全生效

### 复现代码

```go
// 问题复现
tbl.SetStyles(table.Styles{
    Header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
    Cell:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),  // ❌ 导致选中行高亮不完整
    Selected: lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("231")).
        Background(lipgloss.Color("39")).
        Underline(true),
})
```

## 解决方案

### 方案 1：不设置 Cell 样式（当前采用）✅

在 `applyTableStyles` 函数中，不设置 `Cell` 样式，让表格使用默认样式：

```go
func applyTableStyles(t *table.Model, props TableProps) {
    headerStyle, _, selectedStyle := buildTableStyles(props)

    // 使用默认样式作为基础
    s := table.DefaultStyles()
    s.Header = headerStyle
    // ⚠️ 不设置 s.Cell，让表格使用默认样式，这样 Selected 样式才能应用到整行
    // s.Cell = cellStyle // 注释掉此行以确保整行高亮正常工作
    s.Selected = selectedStyle
    t.SetStyles(s)
}
```

**优点**：
- ✅ 选中行可以正常高亮整行
- ✅ 实现简单，无需修改 bubbles/table 源码
- ✅ 兼容性好，不依赖特定版本

**缺点**：
- ❌ 失去对普通单元格样式的精细控制
- ❌ 所有单元格使用默认样式（带 padding）
- ❌ 无法自定义单元格的前景色/背景色

### 方案 2：修改 bubbles/table 源码

修改 `renderRow` 函数，让选中行跳过 `Cell` 样式：

```go
func (m *Model) renderRow(r int) string {
    isSelected := r == m.cursor

    s := make([]string, 0, len(m.cols))
    for i, value := range m.rows[r] {
        if m.cols[i].Width <= 0 {
            continue
        }
        style := lipgloss.NewStyle().Width(m.cols[i].Width).MaxWidth(m.cols[i].Width).Inline(true)

        // ✅ 修复：选中行不应用 Cell 样式
        var renderedCell string
        if isSelected {
            renderedCell = style.Render(runewidth.Truncate(value, m.cols[i].Width, "…"))
        } else {
            renderedCell = m.styles.Cell.Render(style.Render(runewidth.Truncate(value, m.cols[i].Width, "…")))
        }
        s = append(s, renderedCell)
    }

    row := lipgloss.JoinHorizontal(lipgloss.Top, s...)

    if isSelected {
        return m.styles.Selected.Render(row)
    }

    return row
}
```

**优点**：
- ✅ 完全解决问题，支持 Cell 样式和 Selected 样式
- ✅ 可以自定义普通单元格样式

**缺点**：
- ❌ 需要维护 fork 的版本
- ❌ 无法使用上游更新

### 方案 3：条件性设置 Cell 样式（推荐）

根据 `Selected` 样式是否设置，动态决定是否应用 `Cell` 样式：

```go
func applyTableStyles(t *table.Model, props TableProps) {
    headerStyle, cellStyle, selectedStyle := buildTableStyles(props)

    s := table.DefaultStyles()
    s.Header = headerStyle

    // 检查是否设置了自定义的 Selected 样式
    hasCustomSelectedStyle := selectedStyle.String() != lipgloss.NewStyle().String()

    if hasCustomSelectedStyle {
        // 如果有自定义选中样式，不设置 Cell 样式，确保整行高亮正常工作
        // s.Cell = cellStyle // 注释掉
    } else {
        // 如果没有自定义选中样式，可以安全地设置 Cell 样式
        s.Cell = cellStyle
    }

    s.Selected = selectedStyle
    t.SetStyles(s)
}
```

**优点**：
- ✅ 灵活性高，根据用户配置自动选择
- ✅ 不需要修改外部依赖

**缺点**：
- ⚠️ 用户需要理解权衡：要么有 Cell 样式，要么有 Selected 样式

### 方案 4：使用 lipgloss 样式继承

利用 lipgloss 的样式继承机制，让 `Cell` 样式只设置 padding，不设置颜色：

```go
func buildTableStyles(props TableProps) (headerStyle, cellStyle, selectedStyle lipgloss.Style) {
    headerStyle = props.HeaderStyle.GetStyle()
    cellStyle = props.CellStyle.GetStyle()
    selectedStyle = props.SelectedStyle.GetStyle()

    // 设置默认样式
    if headerStyle.String() == lipgloss.NewStyle().String() {
        headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
    }

    // ⚠️ 关键：Cell 样式只设置 padding，不设置颜色
    if cellStyle.String() == lipgloss.NewStyle().String() {
        cellStyle = lipgloss.NewStyle().Padding(0, 1) // 只设置 padding
    }

    if selectedStyle.String() == lipgloss.NewStyle().String() {
        selectedStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("231")).
            Background(lipgloss.Color("39")).
            Underline(true)
    }

    return
}
```

**优点**：
- ✅ 保持 padding 功能
- ✅ 不干扰 Selected 样式

**缺点**：
- ❌ 无法自定义单元格颜色

## 影响范围

此问题影响以下文件：
- `tui/components/table.go` - DSL 表格组件
- `/tui/tea/component/table.go` - Runtime 表格组件

## 测试影响

### 失败的测试

1. **TestNewTableModel_SelectedRowVisibility**
   - **原因**：移除 `Cell` 样式后，视图内容的文本差异变小（之前 Cell 使用灰色前景色 "240"）
   - **状态**：测试失败，但功能正常
   - **修复建议**：调整测试逻辑，检查 `Selected` 样式的背景色而非整个字符串相等性

2. **TestTableComponentWrapper_UpdateMsg_NavigationKeys/Page_Down**
   - **原因**：测试期望值不正确（期望 cursor=5，但只有 5 行数据，最大索引应该是 4）
   - **状态**：预先存在的测试问题，与此修复无关
   - **修复建议**：将期望值从 5 改为 4

### 验证测试

运行以下测试验证修复效果：

```bash
# 测试表格组件基本功能
go test ./tui/components -v -run "TestTableModel"

# 测试表格导航
go test ./tui/components -v -run "TestTableComponentWrapper_UpdateMsg"

# 测试焦点管理
go test ./tui/components -v -run "Focus"
```

## 技术细节

### lipgloss 样式嵌套机制

lipgloss 使用嵌套样式渲染，当多个样式连续应用时：

```go
// 错误示例：样式冲突
cellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("39"))

// 第一步：应用 Cell 样式（设置前景色为灰色）
rendered := cellStyle.Render("Hello")

// 第二步：应用 Selected 样式（设置背景色为蓝色）
final := selectedStyle.Render(rendered)

// 结果：文字仍然是灰色（240），背景是蓝色（39）
// ⚠️ Selected 的前景色无法生效，因为 Cell 已经渲染了文字颜色
```

```go
// 正确示例：样式不冲突
cellStyle := lipgloss.NewStyle().Padding(0, 1)  // 只设置 padding
selectedStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("231")).
    Background(lipgloss.Color("39"))

// 第一步：应用 Cell 样式（只添加 padding）
rendered := cellStyle.Render("Hello")

// 第二步：应用 Selected 样式（设置前景色和背景色）
final := selectedStyle.Render(rendered)

// ✅ 结果：文字是白色（231），背景是蓝色（39），两者都生效
```

### bubbles/table 版本兼容性

| 版本 | 问题状态 | 备注 |
|------|---------|------|
| v0.20.0 | ❌ 存在 | 当前使用版本 |
| v0.19.0 | ❌ 存在 | 相同的 renderRow 实现 |
| v0.18.0 | ❌ 存在 | 相同的 renderRow 实现 |

**注意**：所有 v0.18.0+ 版本都有此问题，因为 `renderRow` 函数的实现逻辑未改变。

## 用户建议

### 对于开发者

1. **使用默认样式**：不要设置 `CellStyle`，让表格使用默认样式
2. **自定义 Selected 样式**：通过 `SelectedStyle` 配置选中行的高亮效果
3. **配置 Header 样式**：`HeaderStyle` 不受此问题影响，可以正常使用

### 配置示例

```yaml
# DSL 配置示例
table:
  columns:
    - key: id
      title: ID
      width: 10
    - key: name
      title: Name
      width: 30
  data: "{{users}}"
  showBorder: true
  # ❌ 不要设置 cellStyle，会导致选中行高亮不完整
  # cellStyle:
  #   foreground: "240"
  # ✅ 使用 selectedStyle 配置选中行高亮
  selectedStyle:
    foreground: "231"
    background: "39"
    bold: true
    underline: true
  # ✅ headerStyle 可以正常使用
  headerStyle:
    foreground: "214"
    bold: true
```

### 对于用户

如果您发现表格选中行只有第一列高亮：
1. 检查是否设置了 `cellStyle` 属性
2. 移除 `cellStyle` 配置
3. 重新运行应用

## 参考链接

- bubbles/table 源码: https://github.com/charmbracelet/bubbles/tree/master/table
- lipgloss 文档: https://github.com/charmbracelet/lipgloss
- 相关 Issue: 暂无（可提交到 bubbles/table 仓库）

## 更新历史

- 2025-01-24: 初始记录，确认问题并实施解决方案 1
- 2025-01-24: 添加源码分析和多种解决方案
- 2025-01-24: 添加技术细节和用户建议
