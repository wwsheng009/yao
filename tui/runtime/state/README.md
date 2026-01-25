# State Management

状态管理系统核心实现。

## 职责

- 状态定义和类型
- 状态变更通知
- 状态历史和撤销

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `state.go` - 状态接口定义
- `manager.go` - 状态管理器
- `history.go` - 历史记录
