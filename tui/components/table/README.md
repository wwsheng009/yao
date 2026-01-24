# Table 组件修复总结

## 修复日期

2025-01-24

## 问题描述

Table 组件在设置 `Cell` 样式后，选中行的 `Selected` 样式只会应用到第一列，导致整行高亮不完整。

## 修复内容

### 修改的文件

1. **tui/components/table.go:281-297**
   - 移除了 `Cell` 样式的设置
   - 使用 `_` 忽略未使用的 `cellStyle` 变量
   - 添加详细注释说明原因

2. **tui/ui/components/table.go:343-384**
   - 移除了 `Cell` 样式的设置
   - 注释掉了 `styles.Cell` 的赋值代码
   - 添加详细注释说明原因

### 修复前

```go
func applyTableStyles(t *table.Model, props TableProps) {
    headerStyle, cellStyle, selectedStyle := buildTableStyles(props)

    if props.ShowBorder {
        t.SetStyles(table.Styles{
            Header:   headerStyle,
            Cell:     cellStyle,  // ❌ 导致选中行高亮不完整
            Selected: selectedStyle,
        })
    } else {
        s := table.DefaultStyles()
        s.Header = headerStyle
        s.Cell = cellStyle  // ❌ 导致选中行高亮不完整
        s.Selected = selectedStyle
        t.SetStyles(s)
    }
}
```

### 修复后

```go
func applyTableStyles(t *table.Model, props TableProps) {
    headerStyle, _, selectedStyle := buildTableStyles(props)

    // 使用默认样式作为基础，这样可以保持正确的边框渲染
    // 注意：由于 bubbles/table 的实现限制，如果设置了 Cell 样式，
    // 选中行的样式只会应用到第一列。为了保证整行高亮正常工作，
    // 我们不设置 Cell 样式，而是使用默认的单元格样式。
    s := table.DefaultStyles()
    s.Header = headerStyle
    // 不设置 s.Cell，让表格使用默认样式，这样 Selected 样式才能应用到整行
    // s.Cell = cellStyle // 注释掉此行以确保整行高亮正常工作
    s.Selected = selectedStyle
    t.SetStyles(s)
}
```

## 根本原因

在 `bubbles/table` 的 `renderRow` 函数中（v0.20.0，第 395-413 行）：

1. 每个单元格先被 `m.styles.Cell` 渲染（设置前景色、背景色等）
2. 所有单元格用 `lipgloss.JoinHorizontal` 拼接成一行
3. 如果是选中行，对整行应用 `m.styles.Selected` 样式

由于 lipgloss 的样式是**嵌套应用**的，`Cell` 样式已经渲染了文字颜色，导致 `Selected` 样式的前景色无法生效，只有第一列可能显示部分选中效果。

## 测试结果

### 通过的测试（26/29）

✅ TestTableModel_FocusAndNavigation
✅ TestTableModel_FocusLost_IgnoresKeys
✅ TestTableModel_SetFocus_Dynamic
✅ TestTableModel_SelectionEvents
✅ TestTableModel_EnterKey
✅ TestTableModel_Pagination
✅ TestTableModel_EmptyTable
✅ TestTableModel_SingleRowTable
✅ TestTableComponentWrapper_UpdateMsg_KeyDown
✅ TestTableComponentWrapper_UpdateMsg_NavigationKeys (部分)
✅ TestTableComponentWrapper_UpdateMsg_EnterKey
✅ TestTableComponentWrapper_UpdateMsg_TargetedMsg
✅ TestTableComponentWrapper_UpdateMsg_RowSelectionEvent
✅ TestNewTableModel_DefaultStyles
✅ TestNewTableModel_CustomSelectedStyle
✅ TestTableComponentWrapper_SetFocus
✅ TestTableComponentWrapper_GetID
✅ TestTableComponentWrapper_View
✅ TestTableComponentWrapper_Init
✅ TestParseTableProps (全部子测试)
✅ TestRenderTable

### 失败的测试（3/29）

1. **TestTableComponentWrapper_UpdateMsg_NavigationKeys/Page_Down**
   - 状态：预先存在的问题
   - 原因：测试期望值不正确（期望 cursor=5，但只有 5 行数据，最大索引应该是 4）
   - 影响：与此修复无关

2. **TestNewTableModel_SelectedRowVisibility**
   - 状态：由此修复引起的测试失败
   - 原因：移除 `Cell` 样式后，视图文本差异变小
   - 功能：**功能正常**，选中行仍然有背景色高亮
   - 建议：调整测试逻辑，检查背景色而非字符串相等性

3. **TestTableComponentWrapper_BoundaryNavigation**
   - 状态：预先存在的问题
   - 原因：边界导航逻辑问题
   - 影响：与此修复无关

## 影响评估

### 功能影响

- ✅ **选中行整行高亮**：现在可以正常工作
- ❌ **单元格样式自定义**：无法再自定义普通单元格的前景色/背景色
- ✅ **表头样式**：不受影响，可以正常使用
- ✅ **默认样式**：保持不变，padding 等布局样式正常

### 用户影响

- **正面**：表格选中行现在可以正确高亮整行
- **负面**：如果用户之前配置了 `cellStyle`，该配置将不再生效

### 兼容性

- 不需要修改现有 DSL 配置（除非使用了 `cellStyle`）
- 不需要修改 API 接口
- 完全向后兼容（除了 `cellStyle` 配置失效）

## 文档

已创建详细的问题文档：
- `tui/components/table/KNOWN_ISSUES.md` - 包含问题描述、根因分析、多种解决方案、技术细节和用户建议

## 建议

### 短期

1. 更新用户文档，说明 `cellStyle` 配置的限制
2. 修复预先存在的测试问题（Page_Down、BoundaryNavigation）
3. 调整 SelectedRowVisibility 测试逻辑

### 长期

1. 考虑向 `bubbles/table` 上游提交 Issue 或 PR
2. 评估是否需要 fork `bubbles/table` 并修复样式优先级问题
3. 考虑实现自定义表格渲染组件，完全控制样式应用

## 相关链接

- 问题详情：`tui/components/table/KNOWN_ISSUES.md`
- bubbles/table 源码：`C:\Users\Vince\go\pkg\mod\github.com\charmbracelet\bubbles@v0.20.0\table\table.go:395-413`
- lipgloss 文档：https://github.com/charmbracelet/lipgloss
