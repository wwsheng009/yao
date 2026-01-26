# Examples

示例代码和演示程序。

## 职责

- 提供组件使用示例
- 演示 Framework 功能
- 学习资源

## 示例列表

- `hello.go` - Hello World 示例
- `demo.go` - 功能演示（需要 `-tags demo`）
- `theme_demo.go` - 主题切换演示（需要 `-tags theme_demo`）

## 运行示例

```bash
# 运行 Hello World
go run tui/framework/examples/hello.go

# 运行完整演示
go run -tags demo tui/framework/examples/demo.go

# 运行主题切换演示
go run -tags theme_demo tui/framework/examples/theme_demo.go
```

## 主题演示

主题切换演示展示了主题系统的完整功能：

### 功能特性

- 显示所有可用主题列表
- 演示静态组件在不同主题下的效果
- 演示输入组件的焦点、占位符状态样式
- 主题对比功能

### 可用主题

| 主题 | 说明 | 适用场景 |
|------|------|----------|
| `light` | 亮色主题 | 白天使用 |
| `dark` | 暗色主题 | 夜间使用 |
| `dracula` | Dracula 配色 | 流行暗色主题 |
| `nord` | Nord 冷色调 | 清爽风格 |
| `monokai` | Monokai 暗色 | 代码编辑器风格 |
| `tokyo-night` | Tokyo Night | 现代暗色 |

### 在代码中使用主题

```go
import (
    "github.com/yaoapp/yao/tui/framework"
    "github.com/yaoapp/yao/tui/framework/theme"
)

func main() {
    app := framework.NewApp()

    // 初始化主题
    if err := app.InitTheme("dark"); err != nil {
        panic(err)
    }

    // 运行时切换主题
    app.SetTheme("light")

    // 获取当前主题
    current := app.GetTheme()
    println("Current theme:", current)

    // 运行应用
    app.Run()
}
```

### 组件使用主题样式

```go
import "github.com/yaoapp/yao/tui/framework/style"

// 在组件内部获取主题样式
func (c *MyComponent) getStyle() style.Style {
    // 获取焦点状态样式
    return style.GetStyle("mycomponent", "focus")
}
```

## 相关文件

- `hello.go` - Hello World
- `demo.go` - 完整演示
- `theme_demo.go` - 主题切换演示
- `../docs/TUI_THEME_DESIGN.md` - 主题系统设计文档
