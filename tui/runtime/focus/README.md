# Focus System

焦点管理系统核心实现。

## 职责

- 焦点栈管理
- 焦点导航逻辑
- 焦点事件处理

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `manager.go` - 焦点管理器
- `stack.go` - 焦点栈
- `navigation.go` - 导航逻辑
