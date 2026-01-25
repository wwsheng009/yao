# Input Processing

输入处理核心层。

## 职责

- 输入事件抽象
- 输入转换和标准化
- 输入验证

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 相关文件

- `processor.go` - 输入处理器
- `keyboard.go` - 键盘输入
- `mouse.go` - 鼠标输入
