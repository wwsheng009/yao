# Style

样式系统。

## 职责

- 样式定义和应用
- 颜色管理
- 主题桥接层
- 样式继承和合并

## 架构

`style` 包是渲染层的样式抽象，与主题系统通过桥接函数连接：

```
组件层 → style.GetStyle() → 主题系统
```

### 关键文件

- `style.go` - 样式定义、GetStyle 桥接函数
- `color.go` - 颜色常量和映射

### 全局样式获取

```go
// 组件获取样式的入口点
// 如果主题系统未初始化，返回空样式
func GetStyle(componentID, state string) Style
```

### 样式类型

```go
type Style struct {
    FG            Color  // 前景色
    BG            Color  // 背景色
    isBold        bool   // 粗体
    isItalic      bool   // 斜体
    isUnderline   bool   // 下划线
    isStrikethrough bool  // 删除线
    isReverse     bool   // 反白
    isBlink       bool   // 闪烁
}
```

### 颜色类型

```go
type Color string

const (
    Black   Color = "black"
    Red     Color = "red"
    Green   Color = "green"
    Yellow  Color = "yellow"
    Blue    Color = "blue"
    Magenta Color = "magenta"
    Cyan    Color = "cyan"
    White   Color = "white"

    BrightBlack   Color = "bright-black"
    BrightRed     Color = "bright-red"
    BrightGreen   Color = "bright-green"
    BrightYellow  Color = "bright-yellow"
    BrightBlue    Color = "bright-blue"
    BrightMagenta Color = "bright-magenta"
    BrightCyan    Color = "bright-cyan"
    BrightWhite   Color = "bright-white"
)
```

## 使用方式

### 组件开发

```go
package mycomponent

import (
    "github.com/yaoapp/yao/tui/runtime/style"
)

func (c *MyComponent) Paint(ctx PaintContext, buf *Buffer) {
    // 获取焦点状态样式
    focusStyle := style.GetStyle("mycomponent", "focus")

    // 获取默认样式
    defaultStyle := style.GetStyle("mycomponent", "")

    // 使用样式绘制
    buf.SetCell(x, y, ch, focusStyle)
}
```

### 构建器模式

```go
import "github.com/yaoapp/yao/tui/runtime/style"

s := style.NewBuilder().
    Foreground("cyan").
    Bold().
    Underline().
    Build()
```

### 直接创建

```go
s := style.Style{}.
    Foreground(style.Red).
    Bold(true)

// 应用到文本
colored := s.Apply("Hello World")
```

### 样式合并

```go
base := style.Style{}.Foreground(style.White)
override := style.Style{}.Bold(true)

// 合并后保留所有属性
merged := base.Merge(override)
// 结果: Foreground=White, Bold=true
```

## 主题集成

`style` 包通过包变量函数连接到主题系统，避免循环导入：

1. 主题初始化时调用 `style.RegisterStyleGetter()`
2. 组件调用 `style.GetStyle()` 获取样式
3. 内部桥接将请求转发到主题系统

```go
// 由主题系统在初始化时调用
func RegisterStyleGetter(getter func() func(string, string) Style)
```

## ANSI 转换

```go
s := style.Style{}.Foreground(style.Cyan).Bold(true)
ansi := s.ToANSI()
// 输出: "\x1b[36;1m"

// 应用到文本
result := s.Apply("Hello")
// 输出: "\x1b[36;1mHello\x1b[0m"
```

## 相关文档

- [TUI_THEME_DESIGN.md](../docs/TUI_THEME_DESIGN.md) - 主题管理系统设计
- [../styling/styling.go](../styling/styling.go) - 样式提供者抽象层
- [../theme/](../theme/) - 主题系统
