# Render Module

渲染模块（可以使用 lipgloss 的唯一模块）。

## 职责

- 样式应用
- 文本渲染
- 最终输出生成

## 注意

这是 runtime 中唯一可以使用 lipgloss 的模块。

## 纯 Go 约束

此目录可以使用：
- lipgloss（样式库）

但不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件

## 相关文件

- `renderer.go` - 渲染器
- `styler.go` - 样式应用器
