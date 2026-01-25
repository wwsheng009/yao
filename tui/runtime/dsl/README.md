# DSL Support

DSL（领域特定语言）支持层。

## 职责

- DSL 解析接口
- 配置到组件的映射
- 声明式 UI 支持

## 注意

虽然此目录涉及 DSL，但应只定义接口，具体实现由 framework 层提供。

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- 具体组件
- lipgloss

## 相关文件

- `parser.go` - 解析器接口
- `config.go` - 配置结构
