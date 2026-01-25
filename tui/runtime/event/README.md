# Event System

事件系统核心实现。

## 职责

- 事件类型定义
- 事件分发机制
- 事件传播和冒泡

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `event.go` - 事件类型定义
- `dispatcher.go` - 事件分发器
