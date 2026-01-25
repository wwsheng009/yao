# Paint System

绘制系统核心实现。

## 职责

- CellBuffer 管理
- Z-index 渲染顺序
- 绘制命令抽象

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss（仅在 Render 模块可以使用）

## 相关文件

- `buffer.go` - CellBuffer 实现
- `painter.go` - 绘制器接口
- `zindex.go` - Z-index 管理
