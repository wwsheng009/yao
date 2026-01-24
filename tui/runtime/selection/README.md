# Text Selection for TUI Runtime

本文档介绍 TUI Runtime 中的文本选择功能。该功能允许用户使用鼠标和键盘选择 TUI 界面中的文字，并将其复制到剪贴板。

## 功能概述

文本选择系统提供以下功能：

- **鼠标选择**：点击拖动选择文字
- **双击**：选择整个单词
- **三击**：选择整行
- **键盘快捷键**：
  - `Ctrl+C` - 复制选中的文字
  - `Ctrl+X` - 剪切（复制并清除选择）
  - `Ctrl+A` - 全选
  - `Escape` - 清除选择
- **选择高亮**：选中的文字会以反色显示
- **剪贴板集成**：支持 Windows、macOS 和 Linux

## 快速开始

### 基本使用

```go
import (
    "github.com/yaoapp/yao/tui/runtime"
    "github.com/yaoapp/yao/tui/runtime/selection"
)

// 创建文本选择系统
buffer := runtime.NewCellBuffer(80, 24)
textSelection := selection.NewTextSelection(buffer)

// 启用文本选择
textSelection.SetEnabled(true)
```

### 处理事件

```go
// 在事件处理函数中处理鼠标和键盘事件
func handleEvent(ev interface{}) {
    switch e := ev.(type) {
    case *event.MouseEvent:
        textSelection.HandleMouseEvent(e)
    case *event.KeyEvent:
        textSelection.HandleKeyEvent(e)
    }
}
```

### 应用选择高亮

```go
// 渲染完成后，应用选择高亮
func render(frame *runtime.Frame) {
    // 先渲染组件到 buffer
    renderComponents(frame.Buffer)

    // 然后应用选择高亮
    textSelection.ApplySelection()
}
```

### 复制选中的文字

```go
// 复制到剪贴板
text, err := textSelection.Copy()
if err != nil {
    // 处理错误
}

// 或者获取选中的文字而不复制
text := textSelection.GetSelectedText()
```

## API 参考

### TextSelection

主要的选择系统接口。

#### 方法

| 方法 | 描述 |
|------|------|
| `HandleEvent(ev)` | 处理任意事件 |
| `HandleMouseEvent(ev)` | 处理鼠标事件 |
| `HandleKeyEvent(ev)` | 处理键盘事件 |
| `ApplySelection()` | 应用选择高亮 |
| `Copy()` | 复制选中的文字到剪贴板 |
| `GetSelectedText()` | 获取选中的文字 |
| `IsActive()` | 是否有活动选择 |
| `IsSelected(x, y)` | 指定位置是否被选中 |
| `Clear()` | 清除选择 |
| `SelectAll()` | 全选 |
| `SelectWord(x, y)` | 选择指定位置的单词 |
| `SelectLine(y)` | 选择指定行 |
| `SetEnabled(bool)` | 启用/禁用选择 |
| `SetHighlightStyle(style)` | 设置高亮样式 |

### SelectionController

提供完整的选择控制接口。

```go
controller := selection.NewSelectionController(buffer)

// 处理事件
controller.HandleEvent(mouseEvent)

// 复制
text, err := controller.Copy()

// 检查状态
if controller.IsActive() {
    // ...
}
```

### Manager

低级别的选择状态管理器。

```go
adapter := selection.NewTextBufferAdapter(buffer)
manager := selection.NewManager(adapter)

// 开始选择
manager.Start(x, y)

// 更新选择
manager.Update(x, y)

// 检查是否选中
if manager.IsSelected(x, y) {
    // ...
}
```

## 配置

### 选择高亮样式

```go
// 使用默认的反色高亮
textSelection.SetHighlightStyle(selection.DefaultHighlightStyle())

// 亮色主题
textSelection.SetHighlightStyle(selection.LightHighlightStyle())

// 暗色主题
textSelection.SetHighlightStyle(selection.DarkHighlightStyle())

// 自定义样式
customStyle := selection.CellStyle{
    Background: "#4A90E2",
    Foreground: "white",
    Bold:       true,
}
textSelection.SetHighlightStyle(customStyle)
```

### 选择模式

```go
// 字符模式（默认）
textSelection.SetSelectionMode(selection.SelectionModeChar)

// 单词模式
textSelection.SetSelectionMode(selection.SelectionModeWord)

// 行模式
textSelection.SetSelectionMode(selection.SelectionModeLine)
```

### 使用配置创建

```go
config := selection.SelectionConfig{
    Enabled:         true,
    HighlightStyle:  selection.DefaultHighlightStyle(),
    SelectionMode:   selection.SelectionModeChar,
    EnableMouse:     true,
    EnableKeyboard:  true,
    EnableClipboard: true,
}

textSelection := selection.NewTextSelectionWithConfig(buffer, config)
```

## 集成示例

### 完整示例

```go
package main

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/yaoapp/yao/tui/runtime"
    "github.com/yaoapp/yao/tui/runtime/selection"
    "github.com/yaoapp/yao/tui/runtime/event"
)

type Model struct {
    buffer        *runtime.CellBuffer
    textSelection *selection.TextSelection
}

func NewModel() Model {
    buffer := runtime.NewCellBuffer(80, 24)

    // 创建文本选择系统
    textSelection := selection.NewTextSelection(buffer)

    return Model{
        buffer:        buffer,
        textSelection: textSelection,
    }
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // 转换为 runtime 鼠标事件
        mouseEv := convertMouseMsg(msg)
        m.textSelection.HandleMouseEvent(mouseEv)

    case tea.KeyMsg:
        // 转换为 runtime 键盘事件
        keyEv := convertKeyMsg(msg)
        m.textSelection.HandleKeyEvent(keyEv)
    }

    return m, nil
}

func (m Model) View() string {
    // 创建 frame
    frame := &runtime.Frame{
        Buffer: m.buffer,
        Width:  80,
        Height: 24,
    }

    // 渲染组件
    m.renderComponents(frame.Buffer)

    // 应用选择高亮
    m.textSelection.ApplySelection()

    return frame.String()
}

func (m *Model) renderComponents(buffer *runtime.CellBuffer) {
    // 渲染你的组件...
}
```

## 平台支持

### 剪贴板支持

| 平台 | 支持状态 | 使用的命令 |
|------|----------|-----------|
| Windows | ✅ | PowerShell (Set-Clipboard/Get-Clipboard) |
| macOS | ✅ | pbcopy / pbpaste |
| Linux (Wayland) | ✅ | wl-copy / wl-paste |
| Linux (X11) | ✅ | xclip 或 xsel |
| 其他 | ❌ | 无 |

### 检查剪贴板支持

```go
if textSelection.IsClipboardSupported() {
    // 剪贴板可用
} else {
    // 提供替代方案（如保存到文件）
}
```

## 高级用法

### 全局选择管理器

对于简单的应用，可以使用全局选择管理器：

```go
// 初始化全局选择
selection.InitGlobalSelection(buffer)

// 在任何地方使用
text := selection.GetSelectedTextGlobal()
selection.ClearSelectionGlobal()
```

### 自定义键盘快捷键

```go
// 创建可配置的键盘处理器
bindings := selection.KeyBindings{
    Copy:      'c',
    Cut:       'x',
    SelectAll: 'a',
    Escape:    27,
}

keyHandler := selection.NewConfigurableKeyboardHandler(
    manager,
    clipboard,
    bindings,
)
```

### 与现有组件集成

如果组件已经实现了 `MouseEventHandler` 接口，可以委托给文本选择系统：

```go
func (c *MyComponent) HandleMouse(ev *event.MouseEvent, localX, localY int) bool {
    // 先让选择系统处理
    if c.textSelection.HandleMouseEvent(ev) {
        return true
    }

    // 如果不是选择操作，处理其他鼠标事件
    // ...
    return false
}
```

## 注意事项

1. **渲染顺序**：必须在渲染完所有组件后调用 `ApplySelection()`，否则选择高亮会被组件覆盖

2. **事件优先级**：鼠标选择事件应该在组件处理之前检查，以免干扰其他交互

3. **性能**：选择高亮需要遍历所有选中的单元格，对于大缓冲区可能有性能影响

4. **剪贴板限制**：
   - Windows 使用 PowerShell，首次复制可能较慢
   - Linux 需要安装剪贴板工具（xclip、wl-copy 等）
   - 某些终端可能不支持剪贴板操作

5. **终端支持**：
   - 需要终端支持鼠标事件
   - 使用 Bubble Tea 时需要启用鼠标：`tea.WithMouseCellMotion()`

## 故障排除

### 问题：选择高亮不显示

**解决方案**：
- 确保 `ApplySelection()` 在所有组件渲染后调用
- 检查选择是否已启用：`textSelection.IsEnabled()`
- 检查是否有活动选择：`textSelection.IsActive()`

### 问题：复制不工作

**解决方案**：
- 检查剪贴板支持：`textSelection.IsClipboardSupported()`
- Linux 用户：确保安装了 `xclip` 或 `wl-copy`
- Windows 用户：确保 PowerShell 可用

### 问题：鼠标选择不响应

**解决方案**：
- 确保启用了鼠标：`tea.WithMouseCellMotion()`
- 检查事件是否正确传递给 `HandleMouseEvent()`
- 确认选择系统已启用：`textSelection.SetEnabled(true)`

## 扩展

### 添加新的选择模式

```go
const SelectionModeParagraph SelectionMode = 3

func (m *Manager) SelectParagraph(x, y int) {
    // 实现段落选择逻辑
}
```

### 自定义剪贴板后端

```go
type CustomClipboard struct{}

func (c *CustomClipboard) Copy(text string) error {
    // 自定义复制实现
}

func (c *CustomClipboard) Paste() (string, error) {
    // 自定义粘贴实现
}
```
