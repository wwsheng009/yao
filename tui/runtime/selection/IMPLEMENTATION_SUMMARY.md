# TUI 文本选择功能实现总结

## 概述

为 TUI Runtime 添加了完整的文本选择功能，允许用户使用鼠标和键盘选择界面中的文字，并将其复制到剪贴板。

## 实现的文件

```
tui/runtime/selection/
├── selection.go      # 选择状态管理核心
├── clipboard.go      # 剪贴板集成（Windows/macOS/Linux）
├── render.go         # 选择高亮渲染
├── mouse_handler.go  # 鼠标事件处理
├── keyboard.go       # 键盘快捷键处理
├── integration.go    # 统一 API 接口
├── selection_test.go # 单元测试
└── README.md         # 使用文档
```

## 功能特性

### 1. 鼠标选择
- **单击拖动**：字符级选择
- **双击**：单词级选择
- **三击**：行级选择
- **Shift+点击**：扩展选择

### 2. 键盘快捷键
- `Ctrl+C` - 复制选中文字
- `Ctrl+X` - 剪切
- `Ctrl+A` - 全选
- `Escape` - 清除选择

### 3. 选择高亮
- 默认使用反色显示
- 支持自定义高亮样式
- 亮色/暗色主题预设

### 4. 剪贴板支持
- **Windows**: PowerShell (Set-Clipboard/Get-Clipboard)
- **macOS**: pbcopy/pbpaste
- **Linux**: wl-copy/wl-paste (Wayland), xclip/xsel (X11)

## 核心 API

### TextSelection - 主要接口

```go
// 创建选择系统
textSelection := selection.NewTextSelection(buffer)

// 处理事件
textSelection.HandleMouseEvent(mouseEvent)
textSelection.HandleKeyEvent(keyEvent)

// 应用高亮
textSelection.ApplySelection()

// 复制
text, err := textSelection.Copy()
```

### 快速开始

```go
import "github.com/yaoapp/yao/tui/runtime/selection"

// 1. 创建选择系统
buffer := runtime.NewCellBuffer(80, 24)
textSelection := selection.NewTextSelection(buffer)

// 2. 在事件处理中调用
func handleEvent(ev interface{}) {
    switch e := ev.(type) {
    case *event.MouseEvent:
        textSelection.HandleMouseEvent(e)
    case *event.KeyEvent:
        textSelection.HandleKeyEvent(e)
    }
}

// 3. 渲染后应用高亮
func render(frame *runtime.Frame) {
    renderComponents(frame.Buffer)
    textSelection.ApplySelection()
}
```

## 测试结果

所有测试通过 (20/20)：

```
=== RUN   TestNewManager
--- PASS: TestNewManager (0.00s)
=== RUN   TestManager_Start
--- PASS: TestManager_Start (0.00s)
=== RUN   TestManager_Update
--- PASS: TestManager_Update (0.00s)
=== RUN   TestManager_Update_Reversed
--- PASS: TestManager_Update_Reversed (0.00s)
=== RUN   TestManager_IsSelected
--- PASS: TestManager_IsSelected (0.00s)
=== RUN   TestManager_Clear
--- PASS: TestManager_Clear (0.00s)
=== RUN   TestManager_SelectWord
--- PASS: TestManager_SelectWord (0.00s)
=== RUN   TestManager_SelectLine
--- PASS: TestManager_SelectLine (0.00s)
=== RUN   TestManager_SelectAll
--- PASS: TestManager_SelectAll (0.00s)
=== RUN   TestManager_GetSelectedCells
--- PASS: TestManager_GetSelectedCells (0.00s)
=== RUN   TestManager_MoveStart
--- PASS: TestManager_MoveStart (0.00s)
=== RUN   TestManager_MoveEnd
--- PASS: TestManager_MoveEnd (0.00s)
=== RUN   TestSelectionRegion
--- PASS: TestSelectionRegion (0.00s)
=== RUN   TestSelectionRegion_Width
--- PASS: TestSelectionRegion_Width (0.00s)
=== RUN   TestRuneWidth
--- PASS: TestRuneWidth (0.00s)
=== RUN   TestStringWidth
--- PASS: TestStringWidth (0.00s)
=== RUN   TestTruncateString
--- PASS: TestTruncateString (0.00s)
PASS
```

## 使用示例

### 基本集成

```go
type Model struct {
    buffer        *runtime.CellBuffer
    textSelection *selection.TextSelection
}

func NewModel() Model {
    buffer := runtime.NewCellBuffer(80, 24)
    textSelection := selection.NewTextSelection(buffer)
    return Model{buffer, textSelection}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        ev := convertToMouseEvent(msg)
        m.textSelection.HandleMouseEvent(ev)
    case tea.KeyMsg:
        ev := convertToKeyEvent(msg)
        m.textSelection.HandleKeyEvent(ev)
    }
    return m, nil
}

func (m Model) View() string {
    // 渲染组件
    m.renderComponents(m.buffer)
    // 应用选择高亮
    m.textSelection.ApplySelection()
    return m.buffer.String()
}
```

### 全局选择管理器

```go
// 初始化全局选择
selection.InitGlobalSelection(buffer)

// 在任何地方使用
text := selection.GetSelectedTextGlobal()
selection.ClearSelectionGlobal()
```

## 平台兼容性

| 平台 | 选择功能 | 复制功能 |
|------|---------|---------|
| Windows | ✅ | ✅ |
| macOS | ✅ | ✅ |
| Linux (Wayland) | ✅ | ✅ (需 wl-copy) |
| Linux (X11) | ✅ | ✅ (需 xclip/xsel) |

## 注意事项

1. **渲染顺序**：`ApplySelection()` 必须在所有组件渲染后调用
2. **事件优先级**：选择事件应在组件处理前检查
3. **终端支持**：需要终端支持鼠标事件
4. **剪贴板工具**：Linux 用户需安装 `xclip` 或 `wl-copy`

## 文档

详细使用文档请参阅：`tui/runtime/selection/README.md`

## 下一步

可选的增强功能：

1. **列选择模式**（Alt+拖动）
2. **正则表达式搜索选择**
3. **多区域选择**
4. **选择历史记录**
5. **自定义复制格式**（HTML、Markdown 等）
