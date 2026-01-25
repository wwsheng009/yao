# AI Integration

AI 功能集成层。

## 职责

- AI Controller 定义
- AI 组件交互接口
- 智能辅助功能

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `controller.go` - AI 控制器接口
- `testing.go` - AI 测试框架
