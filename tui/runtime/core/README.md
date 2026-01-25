# Core

Runtime 核心类型和接口定义。

## 职责

- 核心接口定义（Component, Container 等）
- 基础类型（Bounds, Position, Size 等）
- 事件系统基础

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `component.go` - 组件接口
- `container.go` - 容器接口
- `types.go` - 基础类型定义
