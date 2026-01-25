# Registry

组件注册表。

## 职责

- 组件类型注册
- 组件工厂管理
- 组件查找和实例化

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `registry.go` - 组件注册表
- `factory.go` - 组件工厂
