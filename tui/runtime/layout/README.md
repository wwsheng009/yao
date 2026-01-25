# Layout Engine

布局引擎核心实现（Flexbox 算法）。

## 职责

- Flexbox 布局算法
- 约束计算
- 尺寸分配

## 纯 Go 约束

此目录是纯布局内核，必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `flex.go` - Flexbox 布局算法
- `constraint.go` - 约束系统
- `calculator.go` - 尺寸计算器
