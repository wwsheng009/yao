# Selection System

选择系统核心实现。

## 职责

- 选择状态管理
- 选择区域计算
- 选择操作（全选、反选等）

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `manager.go` - 选择管理器
- `range.go` - 选择范围
