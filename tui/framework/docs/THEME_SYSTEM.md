# Theme System

> **本文档已迁移**

最新的主题管理系统设计文档已移至：[TUI_THEME_DESIGN.md](./TUI_THEME_DESIGN.md)

## 快速概览

### 架构特点

- **组件隔离** - 组件不依赖 `theme` 或 `styling` 包，只使用 `style.Style`
- **全局切换** - 运行时切换主题，所有组件立即生效
- **无循环导入** - 通过桥接函数连接，依赖方向清晰

### 快速开始

```go
import "github.com/yaoapp/yao/tui/framework/theme"

// 初始化主题（应用启动时调用一次）
themeMgr, err := theme.InitThemes("dark")

// 运行时切换主题
themeMgr.Set("light")
```

### 组件使用

```go
import "github.com/yaoapp/yao/tui/framework/style"

// 组件获取样式（无需持有 theme 或 styling 引用）
s := style.GetStyle("input", "focus")
```

### 内置主题

- `light` - 亮色主题
- `dark` - 暗色主题
- `dracula` - Dracula 配色
- `nord` - Nord 冷色调
- `monokai` - Monokai 暗色
- `tokyo-night` - Tokyo Night 暗色

## 详细文档

请查看 [TUI_THEME_DESIGN.md](./TUI_THEME_DESIGN.md) 获取完整的设计文档，包括：

- 分层架构设计
- 核心接口说明
- 组件开发指导
- 主题定义方式
- 优缺点分析
- 内置主题列表
- 未来优化方向

## 相关文件

| 文件 | 说明 |
|------|------|
| `../styling/styling.go` | StyleGetter 函数类型、全局原子存储 |
| `../theme/adapter.go` | 主题适配器、全局注册辅助函数 |
| `../theme/builtin.go` | 内置主题定义 |
| `../theme/manager.go` | 主题管理器 |
| `../style/style.go` | 渲染样式、桥接层 |
| `../input/textinput.go` | 示例组件实现 |

---

**最后更新**: 2026-01-26
