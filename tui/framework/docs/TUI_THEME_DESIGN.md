# TUI 主题管理系统设计文档

## 概述

本文档描述 Yao TUI Framework 的颜色主题管理方案，实现**组件与主题的完全隔离**，支持全局主题切换和本地样式覆盖。

## 设计目标

1. **全局一致性** - 所有组件使用同一套主题，一键切换
2. **组件隔离** - 组件不依赖主题实现，只依赖样式抽象
3. **易于扩展** - 新组件只需声明需要的样式状态
4. **向后兼容** - 渐进式升级，不影响现有代码

## 架构设计

### 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│  应用层 (App)                                               │
│  - theme.InitThemes("dark")  // 初始化                     │
│  - styling.SetTheme("light")    // 运行时切换               │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  组件层 (Components)                                        │
│  - 使用 style.GetStyle("input", "focus") 获取样式            │
│  - 本地覆盖: SetFocusStyle() etc.                           │
│  - 不依赖 theme 包 ✓                                        │
│  - 不依赖 styling 包 ✓                                      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  桥接层 (Style Bridge)                                       │
│  - style.GetStyle(componentID, state) style.Style            │
│  - 通过包变量函数避免循环导入                                │
│  - 连接到 styling 包的全局获取器                             │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  抽象层 (Styling)                                            │
│  - StyleGetter 函数类型                                      │
│  - globalGetter (atomic.Value)                               │
│  - 支持运行时切换                                           │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  主题层 (Theme)                                              │
│  - ThemeStyleProvider 适配器                                 │
│  - StyleConfig → style.Style 转换                           │
│  - 主题定义: Light, Dark, Dracula...                         │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  渲染层 (Paint)                                              │
│  - Buffer.SetCell(x, y, char, style.Style)                  │
│  - 直接使用 style.Style，无需转换                             │
└─────────────────────────────────────────────────────────────┘
```

### 依赖关系

```
组件 → style.Style (渲染层抽象)
      → style.GetStyle() (桥接函数)
      ↓
styling 包 → 全局原子存储 (atomic.Value)
      ↓
主题层 → 实现 StyleGetter
```

**关键设计**: 组件层和主题层通过 `styling` 包隔离，避免循环导入。

## 核心接口

### StyleGetter 函数类型

```go
// styling/styling.go

// StyleGetter 样式获取函数类型
// 组件可以持有此函数类型，而不是 StyleProvider 接口
// 这样组件只需要导入 style 包，而不需要导入 styling 包
type StyleGetter func(componentID, state string) style.Style

// 获取全局样式获取函数
func GetStyleGetter() StyleGetter

// 设置全局样式获取函数（由主题系统调用）
func SetStyleGetter(getter StyleGetter)
```

**参数说明：**
- `componentID`: 组件标识，如 `"input"`, `"button"`
- `state`: 状态标识，如 `""`(默认), `"focus"`, `"placeholder"`, `"disabled"`

**返回值：**
- `style.Style`: 可直接用于渲染的样式

### StyleProvider 接口（向后兼容）

```go
// styling/styling.go

// StyleProvider 样式提供者接口
type StyleProvider interface {
    GetStyle(componentID, state string) style.Style
}

// ThemeProvider 主题样式提供者接口
type ThemeProvider interface {
    StyleProvider

    // 返回主题名称
    Name() string

    // 切换主题
    SetTheme(name string) error

    // 返回当前主题
    CurrentTheme() string
}

// 全局提供者管理
func GetProvider() StyleProvider
func SetProvider(provider StyleProvider)

// 便捷主题切换
func SetTheme(name string) error
func CurrentTheme() string
```

### Style 桥接层

```go
// style/style.go

// RegisterStyleGetter 注册全局样式获取函数
// 由主题系统在初始化时调用
func RegisterStyleGetter(getter func() func(string, string) Style)

// GetStyle 获取组件样式
// 这是组件获取样式的入口点
// 如果主题系统未初始化，返回空样式
func GetStyle(componentID, state string) Style
```

### 主题适配器

```go
// theme/adapter.go
type ThemeStyleProvider struct {
    manager *Manager
}

// 实现 StyleProvider 接口
func (p *ThemeStyleProvider) GetStyle(componentID, state string) style.Style

// 实现 ThemeProvider 接口
func (p *ThemeStyleProvider) SetTheme(name string) error
func (p *ThemeStyleProvider) CurrentTheme() string
```

## 使用方式

### 应用初始化

```go
package main

import (
    "github.com/yaoapp/yao/tui/framework/theme"
)

func main() {
    // 方式1: 使用 InitThemes 便捷函数
    // 会自动注册到 styling 包和 style 包的桥接
    themeMgr, err := theme.InitThemes("dark")
    if err != nil {
        panic(err)
    }

    // 应用启动...
}
```

### 主题切换

```go
// 方式1: 通过 styling 包（所有组件立即生效）
err := styling.SetTheme("light")
if err != nil {
    log.Printf("切换主题失败: %v", err)
}

// 方式2: 通过主题管理器
err = themeMgr.SetTheme("dracula")

// 获取当前主题
name := styling.CurrentTheme()
```

### 组件开发

```go
package input

import (
    "github.com/yaoapp/yao/tui/framework/style"
    // 不需要导入 styling 包！
)

type TextInput struct {
    // 不需要持有 StyleProvider
    // 样式通过 style.GetStyle() 全局获取

    // 本地样式覆盖（可选）
    focusStyle       *style.Style
    placeholderStyle *style.Style
    normalStyle      *style.Style
}

func NewTextInput() *TextInput {
    return &TextInput{
        // 无需初始化 styleProvider
    }
}

func (t *TextInput) Paint(ctx PaintContext, buf *Buffer) {
    // 获取样式
    drawStyle := t.getDrawStyle(isFocused, hasValue)

    // 直接使用绘制
    buf.SetCell(x, y, '[', drawStyle)
}

func (t *TextInput) getDrawStyle(isFocused, hasValue bool) style.Style {
    // 优先级：本地覆盖 > 主题样式
    if !hasValue && t.placeholderStyle != nil {
        return *t.placeholderStyle
    }
    if isFocused && t.focusStyle != nil {
        return *t.focusStyle
    }
    if t.normalStyle != nil {
        return *t.normalStyle
    }

    // 从主题获取（通过 style.GetStyle）
    var state string
    if !hasValue {
        state = "placeholder"
    } else if isFocused {
        state = "focus"
    }

    // 直接调用 style.GetStyle，无需持有 provider
    return style.GetStyle("input", state)
}
```

### 本地样式覆盖

```go
// 方式1: 设置特定状态样式
input.SetFocusStyle(style.Style{}.Foreground(style.Red).Bold(true))
input.SetPlaceholderStyle(style.Style{}.Foreground(style.BrightBlack))

// 方式2: 直接设置基础样式
input.SetNormalStyle(style.Style{}.Foreground(style.White))
```

## 主题定义

### 定义组件样式

```go
// theme/builtin.go

var DarkTheme = &Theme{
    Name: "dark",
    Colors: ColorPalette{
        Primary:    Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},
        // ...
    },
    Components: map[string]ComponentStyle{
        "input": {
            Base: StyleConfig{
                Foreground: &Color{Type: ColorRGB, Value: [3]int{227, 233, 240}},
            },
            States: map[string]StyleConfig{
                "focus": {
                    Foreground: &Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},
                },
                "placeholder": {
                    Foreground: &Color{Type: ColorRGB, Value: [3]int{161, 161, 170}},
                },
            },
        },
    },
}
```

### 样式状态命名约定

| 状态 | 说明 | 适用组件 |
|------|------|----------|
| `""` 或 `"default"` | 默认状态 | 所有 |
| `"focus"` | 获得焦点 | 可聚焦组件 |
| `"hover"` | 鼠标悬停 | 交互组件 |
| `"active"` | 激活/按下 | 按钮 |
| `"disabled"` | 禁用状态 | 可禁用组件 |
| `"placeholder"` | 占位符 | 输入组件 |
| `"error"` | 错误状态 | 表单组件 |
| `"success"` | 成功状态 | 表单组件 |

## 优缺点分析

### 优点

1. **全局一致性**
   - 单一真相源，所有组件共享同一主题
   - 一键切换，无需逐个通知组件
   - `atomic.Value` 保证线程安全

2. **组件隔离**
   - 组件不依赖 `theme` 包
   - 组件不依赖 `styling` 包
   - 只依赖 `style.Style` (渲染层抽象)
   - 易于测试，可注入 Mock 获取器

3. **无循环导入**
   - `styling` 包导入 `style` 包
   - `style` 包通过包变量函数桥接
   - `theme` 包适配 `styling` 接口
   - 依赖方向单向清晰

4. **渐进式**
   - 老组件可以继续直接使用 `style.Style`
   - 新组件可以选择接入主题系统
   - `paint.Buffer` 接口保持不变

5. **灵活性**
   - 支持本地样式覆盖
   - 支持自定义样式提供者
   - 支持运行时主题切换

### 缺点

1. **两套系统共存**
   - `theme.StyleConfig` (主题配置)
   - `style.Style` (渲染样式)
   - 需要转换层（在 ThemeStyleProvider 中）

2. **颜色信息损失**
   - RGB 颜色转换为命名颜色 (Cyan, Red, etc.)
   - 受终端 16 色限制
   - 可通过扩展支持真彩色（见优化方向）

3. **全局状态依赖**
   - 依赖全局原子存储
   - 测试时可能需要重置全局状态

### 与传统方案对比

| 方案 | 优点 | 缺点 |
|------|------|------|
| **当前方案** | 全局切换、组件隔离、无循环依赖 | 需要转换层 |
| 直接嵌入样式 | 简单直接 | 无法全局切换 |
| CSS类风格 | 完全解耦 | 复杂度高、调试困难 |
| 渲染器模式 | 组件最轻量 | 需要重新设计渲染系统 |

## 组件颜色开发指导

### 新组件开发步骤

1. **定义组件 ID**
   ```go
   const ComponentID = "mycomponent"
   ```

2. **在所有内置主题中定义样式**
   ```go
   // theme/builtin.go - 每个主题都添加
   Components: map[string]ComponentStyle{
       ComponentID: {
           Base: StyleConfig{ /* 默认样式 */ },
           States: map[string]StyleConfig{
               "focus":    { /* 焦点样式 */ },
               "disabled": { /* 禁用样式 */ },
           },
       },
   }
   ```

3. **组件实现**
   ```go
   package mycomponent

   import (
       "github.com/yaoapp/yao/tui/framework/style"
       // 不需要导入 styling 或 theme 包
   )

   type MyComponent struct {
       // 本地样式覆盖（可选）
       focusStyle *style.Style
   }

   func (c *MyComponent) getStyle(state string) style.Style {
       // 检查本地覆盖
       if state == "focus" && c.focusStyle != nil {
           return *c.focusStyle
       }

       // 从全局主题获取
       return style.GetStyle(ComponentID, state)
   }
   ```

4. **测试**
   ```go
   // 测试默认样式
   c := NewMyComponent()
   defaultStyle := c.getStyle("")

   // 测试主题切换（需要在测试中初始化主题）
   theme.InitThemes("light")
   // 验证样式更新
   ```

### 样式获取优先级

```
1. 本地覆盖 (SetFocusStyle, SetPlaceholderStyle 等)
   ↓
2. 主题状态样式 (theme.Components["comp"].States["focus"])
   ↓
3. 主题基础样式 (theme.Components["comp"].Base)
   ↓
4. 默认样式 (style.Style{})
```

### 颜色选择建议

| 用途 | 建议颜色 | RGB值 |
|------|---------|-------|
| 主要操作 | Cyan (青色) | (0, 255, 255) |
| 次要操作 | Blue (蓝色) | (0, 0, 255) |
| 成功 | Green (绿色) | (0, 255, 0) |
| 警告 | Yellow (黄色) | (255, 255, 0) |
| 错误 | Red (红色) | (255, 0, 0) |
| 禁用 | BrightBlack (亮黑) | (128, 128, 128) |
| 占位符 | BrightBlack (亮黑) | (128, 128, 128) |
| 默认文本 | White (白色) | (255, 255, 255) |

## 内置主题

### Light Theme

亮色主题，适合白天使用。

```go
theme.InitThemes("light")
```

### Dark Theme

暗色主题，适合夜间使用。

```go
theme.InitThemes("dark")
```

### Dracula Theme

基于 Dracula 配色方案的暗色主题。

```go
theme.InitThemes("dracula")
```

### Nord Theme

基于 Nord 配色方案的冷色调主题。

```go
theme.InitThemes("nord")
```

### Monokai Theme

基于 Monokai 配色方案的暗色主题。

```go
theme.InitThemes("monokai")
```

### Tokyo Night Theme

基于 Tokyo Night 配色方案的暗色主题。

```go
theme.InitThemes("tokyo-night")
```

## 未来优化方向

### 1. 支持真彩色

扩展 `style.Style` 支持 RGB：

```go
type Color struct {
    IsRGB   bool
    RGB     [3]int  // 直接存储 RGB
    Name    string  // 兼容命名颜色
    Is256   bool
    Code256 int     // 256色代码
}

// ToANSI 支持 True Color 转义序列
func (c Color) ToANSI() string {
    if c.IsRGB {
        return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.RGB[0], c.RGB[1], c.RGB[2])
    }
    // ...
}
```

**优点**: 保留颜色精度，支持现代终端
**缺点**: 需要终端支持 True Color

### 2. CSS 类风格

```go
// 组件只声明 class
input.SetClass("input primary")

// 样式在外部定义
theme.SetClassStyles(".input.primary:focus", style.Style{...})
```

**优点**: 完全解耦
**缺点**: 复杂度高，选择器解析成本

### 3. 动态主题

支持从配置文件或远程加载主题：

```go
theme.LoadFromFile("/path/to/theme.yaml")
theme.LoadFromURL("https://example.com/theme.yaml")
```

## 附录

### 相关文件

| 文件 | 说明 |
|------|------|
| `styling/styling.go` | StyleGetter 函数类型、全局原子存储 |
| `theme/adapter.go` | 主题适配器、全局注册辅助函数 |
| `theme/builtin.go` | 内置主题定义 |
| `theme/manager.go` | 主题管理器 |
| `style/style.go` | 渲染样式、桥接层 |
| `input/textinput.go` | 示例组件实现 |

### API 速查

```go
// 初始化（应用启动时调用一次）
themeMgr, err := theme.InitThemes("dark")

// 切换主题（所有组件立即生效）
styling.SetTheme("light")
// 或
themeMgr.SetTheme("light")

// 获取当前主题名称
name := styling.CurrentTheme()

// 组件获取样式（在组件内部调用）
s := style.GetStyle("input", "focus")

// 组件设置本地覆盖
input.SetFocusStyle(style.Style{}.Foreground(style.Cyan))
input.SetPlaceholderStyle(style.Style{}.Foreground(style.BrightBlack))
```

### 依赖关系图

```
┌─────────────────────────────────────────────────────────────┐
│                        Components                            │
│  - textinput.go                                             │
│  - button.go                                                │
│  - ...                                                      │
│                                                              │
│  只依赖: style.Style, style.GetStyle()                       │
│  不依赖: theme, styling                                      │
└─────────────────────────────────────────────────────────────┘
                    │ imports
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                      style/style.go                          │
│  - Style struct                                              │
│  - GetStyle() → 通过包变量函数桥接                            │
│  - RegisterStyleGetter() 由主题系统调用                      │
│                                                              │
│  依赖: 无（通过包变量避免循环导入）                           │
└─────────────────────────────────────────────────────────────┘
                    ▲ 包变量函数
                    │ 调用
                    │
┌─────────────────────────────────────────────────────────────┐
│                   styling/styling.go                         │
│  - StyleGetter 函数类型                                      │
│  - globalGetter (atomic.Value)                               │
│  - GetStyleGetter() / SetStyleGetter()                       │
│  - StyleProvider 接口                                       │
│                                                              │
│  依赖: style.Style                                            │
└─────────────────────────────────────────────────────────────┘
                    ▲ 实现/注册
                    │
┌─────────────────────────────────────────────────────────────┐
│                   theme/adapter.go                           │
│  - ThemeStyleProvider 实现 StyleProvider                     │
│  - RegisterAsGlobal() 注册到 styling 和 style                │
│  - InitThemes() 便捷初始化函数                               │
│                                                              │
│  依赖: style, styling, theme                                 │
└─────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                   theme/builtin.go                           │
│  - Light, Dark, Dracula 等主题定义                           │
│                                                              │
│  依赖: theme (内部包)                                        │
└─────────────────────────────────────────────────────────────┘
```

---

**文档版本**: 2.0
**最后更新**: 2026-01-26
**维护者**: Yao Framework Team
