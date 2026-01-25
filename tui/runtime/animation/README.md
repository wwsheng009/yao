# Animation System

动画系统核心实现。

## 职责

- 动画状态管理
- 缓动函数（Easing functions）
- 动画帧计算

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `animator.go` - 动画控制器
- `easing.go` - 缓动函数
