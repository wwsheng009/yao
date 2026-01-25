# Exports

导出的公共 API 和接口。

## 职责

- 定义公开的 API
- 提供跨包访问的接口
- Facade 模式实现

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `api.go` - 公共 API 导出
