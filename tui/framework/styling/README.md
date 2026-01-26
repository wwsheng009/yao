# Styling

样式提供者抽象层，实现组件与主题的隔离。

## 职责

- 提供样式获取接口（StyleProvider）
- 全局样式提供者管理（使用 atomic.Value）
- 主题切换支持（ThemeProvider）
- 组件与主题解耦

## 架构位置

```
组件层 → style.GetStyle()
         ↓
styling 包 → 全局原子存储
         ↓
主题层 → ThemeStyleProvider
```

## 设计原则

1. **避免循环导入** - `styling` 导入 `style`，`style` 不导入 `styling`
2. **接口隔离** - 组件依赖抽象接口，不依赖具体实现
3. **运行时切换** - 使用 `atomic.Value` 支持原子性主题切换

## 核心 API

### StyleGetter 函数类型

```go
// 样式获取函数类型
// 组件可以持有此函数类型，而不是 StyleProvider 接口
type StyleGetter func(componentID, state string) style.Style
```

### 全局样式管理

```go
// 获取全局样式获取函数
func GetStyleGetter() StyleGetter

// 设置全局样式获取函数（由主题系统调用）
func SetStyleGetter(getter StyleGetter)
```

### StyleProvider 接口（向后兼容）

```go
// 样式提供者接口
type StyleProvider interface {
    GetStyle(componentID, state string) style.Style
}

// 全局提供者管理
func GetProvider() StyleProvider
func SetProvider(provider StyleProvider)
```

### ThemeProvider 接口

```go
// 主题样式提供者接口
type ThemeProvider interface {
    StyleProvider

    Name() string
    SetTheme(name string) error
    CurrentTheme() string
}

// 便捷主题切换
func SetTheme(name string) error
func CurrentTheme() string
```

## 使用方式

### 主题系统初始化

```go
import "github.com/yaoapp/yao/tui/framework/theme"

// InitThemes 会自动注册到 styling 包
themeMgr, err := theme.InitThemes("dark")
```

### 运行时切换主题

```go
import "github.com/yaoapp/yao/tui/framework/styling"

// 所有组件立即生效
err := styling.SetTheme("light")
```

### 组件使用

组件通常不需要直接导入 `styling` 包，而是通过 `style.GetStyle()` 间接使用：

```go
import "github.com/yaoapp/yao/tui/framework/style"

// 在组件内部
s := style.GetStyle("input", "focus")
```

## 相关文档

- [TUI_THEME_DESIGN.md](../docs/TUI_THEME_DESIGN.md) - 完整的主题系统设计文档
- [../style/README.md](../style/README.md) - 渲染层样式
- [../theme/adapter.go](../theme/adapter.go) - 主题适配器实现
