# Runtime 中文本选择功能使用指南

## 概述

文本选择功能现已直接集成到 TUI Runtime 中，类似于 FocusManager。这提供了更好的架构一致性和更简洁的 API。

## 新增的文件

```
tui/runtime/
├── selection.go    # SelectionManager - 选择状态管理
└── clipboard.go    # Clipboard - 剪贴板集成
```

## Runtime API

### 选择控制方法

```go
// 获取选择管理器
selection := runtime.GetSelection()

// 启用/禁用选择
runtime.EnableSelection(true)
runtime.IsSelectionEnabled()

// 选择操作
runtime.StartSelection(x, y)      // 开始选择
runtime.UpdateSelection(x, y)     // 更新选择
runtime.ExtendSelection(x, y)     // 扩展选择
runtime.ClearSelection()          // 清除选择

// 快捷选择
runtime.SelectAll()               // 全选
runtime.SelectWord(x, y)          // 选择单词
runtime.SelectLine(y)             // 选择行

// 查询状态
runtime.IsSelectionActive()        // 是否有选择
runtime.IsSelected(x, y)          // 指定位置是否选中
runtime.GetSelectedText()         // 获取选中文本
runtime.GetSelectionRange()       // 获取选择范围

// 样式设置
runtime.SetSelectionHighlightStyle(style)
runtime.SetSelectionMode(mode)
```

### 剪贴板操作

```go
// 复制选中文本
text, err := runtime.CopySelection()

// 从剪贴板粘贴
text, err := runtime.PasteFromClipboard()

// 检查剪贴板支持
runtime.IsClipboardSupported()
```

## 集成到 Model

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yaoapp/yao/tui/runtime"
    "github.com/yaoapp/yao/tui/runtime/event"
)

type Model struct {
    Runtime *runtime.RuntimeImpl
    // ...
}

// 在 Update 中处理鼠标事件
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // 转换为 runtime 事件
        ev := m.EventAdapter.Convert(msg)

        if ev.Mouse != nil {
            // 处理选择
            switch ev.Mouse.Type {
            case event.MousePress:
                if ev.Mouse.Click == event.MouseLeft {
                    m.Runtime.StartSelection(ev.Mouse.X, ev.Mouse.Y)
                }
            case event.MouseMove:
                // 检查鼠标左键是否按下（需要跟踪按键状态）
                m.Runtime.UpdateSelection(ev.Mouse.X, ev.Mouse.Y)
            case event.MouseRelease:
                // 选择完成
            }
        }

    case tea.KeyMsg:
        // 处理键盘快捷键
        ev := m.EventAdapter.Convert(msg)

        if ev.Key != nil {
            switch ev.Key.Key {
            case 'c', 'C':
                if ev.Key.Mod == event.ModCtrl {
                    // Ctrl+C - 复制
                    m.Runtime.CopySelection()
                }
            case 'a', 'A':
                if ev.Key.Mod == event.ModCtrl {
                    // Ctrl+A - 全选
                    m.Runtime.SelectAll()
                }
            case 27: // Escape
                m.Runtime.ClearSelection()
            }
        }
    }

    return m, nil
}

// View 中无需特殊处理
// 选择高亮会在 Render() 中自动应用
func (m Model) View() string {
    result := m.Runtime.Layout(m.LayoutRoot, m.constraints)
    frame := m.Runtime.Render(result)
    return frame.String()
}
```

## 完整示例

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yaoapp/yao/tui/runtime"
)

type Model struct {
    runtime *runtime.RuntimeImpl
    mouseDown bool
}

func NewModel() Model {
    return Model{
        runtime: runtime.NewRuntime(80, 24),
    }
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // 处理鼠标选择
        x, y := msg.X, msg.Y
        if msg.Type == 1 { // Press
            if msg.Button == 1 { // Left button
                m.mouseDown = true
                m.runtime.StartSelection(x, y)
            }
        } else if msg.Type == 3 { // Move
            if m.mouseDown {
                m.runtime.UpdateSelection(x, y)
            }
        } else if msg.Type == 2 { // Release
            m.mouseDown = false
        }

    case tea.KeyMsg:
        // Ctrl+C 复制
        if msg.Type == tea.KeyCtrlC {
            m.runtime.CopySelection()
            return m, nil
        }
        // Ctrl+A 全选
        if msg.String() == "ctrl+a" {
            m.runtime.SelectAll()
            return m, nil
        }
        // Escape 清除选择
        if msg.Type == tea.KeyEscape {
            m.runtime.ClearSelection()
            return m, nil
        }
    }

    return m, nil
}

func (m Model) View() string {
    // 渲染组件...
    // 选择高亮会自动应用
    return m.render()
}

func main() {
    p := tea.NewProgram(
        NewModel(),
        tea.WithMouseCellMotion(), // 启用鼠标支持
    )

    if err := p.Start(); err != nil {
        panic(err)
    }
}
```

## 配置选项

```go
// 使用默认反色高亮
runtime.SetSelectionHighlightStyle(runtime.DefaultSelectionHighlight())

// 亮色主题
runtime.SetSelectionHighlightStyle(runtime.LightSelectionHighlight())

// 暗色主题
runtime.SetSelectionHighlightStyle(runtime.DarkSelectionHighlight())

// 自定义样式
customStyle := runtime.CellStyle{
    Background: "#4A90E2",
    Foreground: "white",
    Bold:       true,
}
runtime.SetSelectionHighlightStyle(customStyle)

// 设置选择模式
runtime.SetSelectionMode(runtime.SelectionModeChar)  // 字符
runtime.SetSelectionMode(runtime.SelectionModeWord)   // 单词
runtime.SetSelectionMode(runtime.SelectionModeLine)   // 行
```

## 事件处理流程

```
┌─────────────────────────────────────────────────────┐
│                    Bubble Tea                        │
│              (tea.MouseMsg, tea.KeyMsg)               │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│                    Model.Update()                    │
│                                                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │ Mouse Press │→ │  Mouse Move │→ │Mouse Release│ │
│  │ StartSelect │→ │ UpdateSelect│→ │    (End)    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
│                                                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │   Ctrl+C    │→ │   Ctrl+A    │→ │   Escape    │ │
│  │   Copy      │→ │ SelectAll   │→ │ ClearSelect │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│                   Runtime.Render()                   │
│                                                       │
│  1. 渲染所有组件 → CellBuffer                       │
│  2. 应用选择高亮 → ApplyHighlight()                  │
│  3. 返回 Frame                                      │
└─────────────────────────────────────────────────────┘
```

## 优势

1. **架构一致**：与 FocusManager 同级，统一管理交互状态
2. **自动化**：选择高亮在 Render() 中自动应用
3. **API 简洁**：通过 Runtime 统一访问
4. **无循环依赖**：所有代码在 runtime 包内
5. **性能更好**：直接访问 CellBuffer，无需接口调用
