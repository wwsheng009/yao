# Action System

Action 处理和执行系统。

## 职责

- 定义 Action 接口和类型
- 处理组件 Action
- Action 路由和分发

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `action.go` - Action 接口定义
- `handler.go` - Action 处理器
