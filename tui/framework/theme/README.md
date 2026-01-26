# Theme

主题系统，提供 UI 外观管理能力。

## 职责

- 主题定义和存储
- 主题管理器（注册、切换、查询）
- 组件样式配置
- 颜色系统
- 主题适配器（连接到 styling 包）

## 架构位置

```
应用层 → theme.InitThemes()
         ↓
主题层 → Manager, Theme
         ↓
适配层 → ThemeStyleProvider
         ↓
styling → 全局原子存储
         ↓
组件层 → style.GetStyle()
```

## 快速开始

```go
import "github.com/yaoapp/yao/tui/framework/theme"

// 初始化主题系统（应用启动时调用一次）
themeMgr, err := theme.InitThemes("dark")
if err != nil {
    panic(err)
}

// 运行时切换主题
err = themeMgr.Set("light")

// 获取当前主题
current := themeMgr.Current()
fmt.Println("当前主题:", current.Name)
```

## 核心 API

### 主题初始化

```go
// InitThemes 初始化主题系统
// 创建管理器，注册内置主题，设置为全局提供者
func InitThemes(initialTheme string) (*Manager, error)
```

### 主题管理器

```go
type Manager struct {
    // 内部字段
}

// Register 注册主题
func (m *Manager) Register(theme *Theme) error

// RegisterMultiple 批量注册主题
func (m *Manager) RegisterMultiple(themes []*Theme)

// Get 获取主题
func (m *Manager) Get(name string) (*Theme, bool)

// Set 设置当前主题
func (m *Manager) Set(name string) error

// Current 获取当前主题
func (m *Manager) Current() *Theme

// List 列出所有主题
func (m *Manager) List() []string

// GetComponentStyle 获取组件样式
func (m *Manager) GetComponentStyle(componentID, state string) StyleConfig
```

### 主题注册为全局提供者

```go
// RegisterAsGlobal 将主题管理器注册为全局样式提供者
// 这是应用初始化时的便捷方法
func (m *Manager) RegisterAsGlobal()
```

## 主题定义

### 主题结构

```go
type Theme struct {
    // 主题名称
    Name string

    // 主题版本
    Version string

    // 父主题（用于继承）
    Parent *Theme

    // 颜色调色板
    Colors ColorPalette

    // 样式映射
    Styles map[string]StyleConfig

    // 组件特定配置
    Components map[string]ComponentStyle

    // 边框样式
    Borders BorderStyles
}
```

### 颜色调色板

```go
type ColorPalette struct {
    Primary    Color
    Secondary  Color
    Accent     Color

    Success    Color
    Warning    Color
    Error      Color
    Info       Color

    Background Color
    Foreground Color
    Muted      Color

    Border     Color
    Focus      Color

    Disabled   Color
    Hover      Color
    Active     Color
}
```

### 组件样式

```go
type ComponentStyle struct {
    Base   StyleConfig
    States map[string]StyleConfig
}

type StyleConfig struct {
    Foreground     *Color
    Background     *Color
    Bold           bool
    Italic         bool
    Underline      bool
    Strikethrough  bool
    Reverse        bool
    Blink          bool
}
```

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

## 自定义主题

```go
import "github.com/yaoapp/yao/tui/framework/theme"

customTheme := &theme.Theme{
    Name:    "mytheme",
    Version: "1.0.0",
    Colors: theme.ColorPalette{
        Primary:    theme.ColorRGB(97, 175, 239),
        Secondary:  theme.ColorRGB(224, 108, 117),
        Background: theme.ColorRGB(17, 24, 39),
        Foreground: theme.ColorRGB(227, 233, 240),
    },
    Components: map[string]theme.ComponentStyle{
        "input": {
            Base: theme.StyleConfig{
                Foreground: theme.ColorRef("Foreground"),
            },
            States: map[string]theme.StyleConfig{
                "focus": {
                    Foreground: theme.ColorRef("Primary"),
                },
            },
        },
    },
}

mgr := theme.NewManager()
mgr.Register(customTheme)
mgr.Set("mytheme")
mgr.RegisterAsGlobal()
```

## 颜色创建

```go
// RGB 颜色
c := theme.ColorRGB(97, 175, 239)

// 十六进制颜色
c := theme.ColorHex("#6193ef")

// 命名颜色
c := theme.ColorNamed("blue")

// 256 色
c := theme.Color256(117)

// 引用调色板
c := theme.ColorRef("Primary")
```

## 边框样式

```go
import "github.com/yaoapp/yao/tui/framework/theme/border"

// 使用预定义边框
theme.WithBorder(theme.Borders{
    "default": border.Rounded,
    "focused": border.Double,
    "error":   border.Bold,
})
```

## 相关文件

| 文件 | 说明 |
|------|------|
| `theme.go` | 主题结构定义 |
| `color.go` | 颜色类型和创建函数 |
| `border.go` | 边框样式 |
| `manager.go` | 主题管理器 |
| `builtin.go` | 内置主题定义 |
| `adapter.go` | 主题适配器 |
| `themed.go` | Themed 接口 |

## 相关文档

- [TUI_THEME_DESIGN.md](../docs/TUI_THEME_DESIGN.md) - 完整的主题系统设计文档
- [../styling/README.md](../styling/README.md) - 样式提供者抽象层
- [../style/README.md](../style/README.md) - 渲染层样式
