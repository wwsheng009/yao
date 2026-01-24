# TUI 文本选择功能集成指南

本文档说明如何将文本选择功能集成到现有的 TUI Runtime 中。

## 架构概述

```
┌─────────────────────────────────────────────────────────────┐
│                        Bubble Tea                           │
│                    (tea.KeyMsg, tea.MouseMsg)               │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Event Adapter                             │
│              (tui/tea/adapter/event_adapter.go)              │
│                  ConvertBubbleTeaMsg()                      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    TUI Model                                │
│                   (tui/model.go)                            │
│            handleGeometryEvent() / handleComponentEvent()   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Runtime Engine                            │
│              (tui/runtime/runtime_impl.go)                  │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Layout     │  │   Render     │  │   Events     │      │
│  │              │  │              │  │              │      │
│  │ Measure      │  │ Component →  │  │ Dispatch     │      │
│  │ Layout       │  │ CellBuffer   │  │ Focus        │      │
│  │              │  │              │  │              │      │
│  └──────────────┘  │ Selection    │  │ Selection    │      │
│                    │ Highlight    │  │ Events       │      │
│                    └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              Text Selection Module                         │
│             (tui/runtime/selection/)                       │
│                                                               │
│  • Manager      - 选择状态管理                              │
│  • Renderer     - 选择高亮渲染                              │
│  • MouseHandler - 鼠标拖动选择                              │
│  • KeyboardHandler - 键盘快捷键                             │
│  • Clipboard    - 剪贴板集成                                │
└─────────────────────────────────────────────────────────────┘
```

## 集成步骤

### 步骤 1: 在 RuntimeImpl 中添加文本选择支持

修改 `tui/runtime/runtime_impl.go`：

```go
package runtime

import (
    "time"
    "github.com/yaoapp/yao/tui/runtime/animation"
    sel "github.com/yaoapp/yao/tui/runtime/selection"  // 新增
)

type RuntimeImpl struct {
    width       int
    height      int
    lastFrame   *Frame
    lastResult  LayoutResult
    dirtyRegions []Rect
    forceFullRender bool
    isDirty     bool
    focusMgr    *FocusManager
    lastRoot    *LayoutNode
    animMgr     *animation.Manager
    animationsRunning bool

    // 新增：文本选择支持
    textSelection *sel.TextSelection
    selectionEnabled bool
}

// NewRuntime 创建 Runtime 实例
func NewRuntime(width, height int) *RuntimeImpl {
    r := &RuntimeImpl{
        width:    width,
        height:   height,
        focusMgr: NewFocusManager(),
        animMgr:  animation.NewManager(),
        animationsRunning: false,
        selectionEnabled: true,  // 默认启用
    }

    // 创建选择系统（稍后在首次渲染时初始化 buffer）
    return r
}
```

### 步骤 2: 修改 Render 方法以应用选择高亮

在 `Render()` 方法中，渲染完所有组件后应用选择高亮：

```go
func (r *RuntimeImpl) Render(result LayoutResult) Frame {
    // ... 现有渲染代码 ...

    frame := Frame{
        Buffer: buf,
        Width:  r.width,
        Height: r.height,
        Dirty:  len(r.dirtyRegions) > 0 || r.isDirty,
    }

    // 新增：应用选择高亮（在所有组件渲染后）
    if r.selectionEnabled && r.textSelection != nil {
        // 确保 textSelection 使用最新的 buffer
        r.textSelection.UpdateBuffer(buf)
        // 应用选择高亮
        r.textSelection.ApplySelection()
    }

    // ... 现有代码 ...

    return frame
}
```

### 步骤 3: 添加选择相关方法到 RuntimeImpl

```go
// ============================================================================
// 文本选择支持方法
// ============================================================================

// EnableSelection 启用或禁用文本选择
func (r *RuntimeImpl) EnableSelection(enabled bool) {
    r.selectionEnabled = enabled
    if r.textSelection != nil {
        r.textSelection.SetEnabled(enabled)
    }
}

// IsSelectionEnabled 返回文本选择是否启用
func (r *RuntimeImpl) IsSelectionEnabled() bool {
    return r.selectionEnabled
}

// GetTextSelection 返回文本选择系统实例
func (r *RuntimeImpl) GetTextSelection() *sel.TextSelection {
    return r.textSelection
}

// HandleSelectionEvent 处理选择相关的事件
// 返回 true 表示事件被处理
func (r *RuntimeImpl) HandleSelectionEvent(ev interface{}) bool {
    if !r.selectionEnabled || r.textSelection == nil {
        return false
    }

    switch e := ev.(type) {
    case MouseEvent:
        return r.textSelection.HandleMouseEvent(&e)
    case KeyEvent:
        return r.textSelection.HandleKeyEvent(&e)
    }

    return false
}

// CopySelection 复制选中的文本到剪贴板
func (r *RuntimeImpl) CopySelection() (string, error) {
    if r.textSelection == nil {
        return "", nil
    }
    return r.textSelection.Copy()
}

// GetSelectedText 返回选中的文本
func (r *RuntimeImpl) GetSelectedText() string {
    if r.textSelection == nil {
        return ""
    }
    return r.textSelection.GetSelectedText()
}

// ClearSelection 清除当前选择
func (r *RuntimeImpl) ClearSelection() {
    if r.textSelection != nil {
        r.textSelection.Clear()
    }
}

// IsSelectionActive 返回是否有活动选择
func (r *RuntimeImpl) IsSelectionActive() bool {
    if r.textSelection == nil {
        return false
    }
    return r.textSelection.IsActive()
}

// SelectAll 选择全部文本
func (r *RuntimeImpl) SelectAll() {
    if r.textSelection != nil {
        r.textSelection.SelectAll()
    }
}
```

### 步骤 4: 修改 Model 以处理选择事件

修改 `tui/model.go` 的 `handleGeometryEvent` 方法：

```go
func (m *Model) handleGeometryEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.Runtime == nil {
        return m, nil
    }

    // 转换事件
    runtimeEvent := m.EventAdapter.Convert(msg)

    // 首先让选择系统处理鼠标/键盘事件
    handled := false
    if m.TextSelection != nil {  // 需要添加 TextSelection 字段
        handled = m.TextSelection.HandleEvent(runtimeEvent)
    }

    // 如果选择系统没有处理，继续正常的事件流程
    if !handled {
        switch runtimeEvent.Type {
        case event.EventTypeMouse:
            if runtimeEvent.Mouse != nil {
                result := event.DispatchEvent(runtimeEvent, m.Runtime.GetBoxes())
                // 处理焦点变化等...
            }
        case event.EventTypeKey:
            if runtimeEvent.Key != nil {
                result := event.DispatchEvent(runtimeEvent, m.Runtime.GetBoxes())
                // 处理焦点变化等...
            }
        }
    }

    return m, nil
}
```

### 步骤 5: 添加配置选项

在 `tui/config.go` 中添加选择相关配置：

```go
type Config struct {
    // ... 现有字段 ...

    // 文本选择配置
    EnableTextSelection *bool  // 是否启用文本选择（默认 true）
    SelectionHighlight  string // 选择高亮样式：'reverse', 'light', 'dark'
}
```

### 步骤 6: 简化集成 - 使用独立的适配器

为了最小化对现有代码的修改，可以创建一个独立的适配器：

```go
// tui/runtime/selection/adapter.go
package selection

import (
    "github.com/yaoapp/yao/tui/runtime"
)

// RuntimeAdapter 将文本选择集成到 Runtime
type RuntimeAdapter struct {
    runtime       *runtime.RuntimeImpl
    textSelection *TextSelection
}

// NewRuntimeAdapter 创建一个新的 Runtime 适配器
func NewRuntimeAdapter(rt *runtime.RuntimeImpl) *RuntimeAdapter {
    adapter := &RuntimeAdapter{
        runtime: rt,
    }

    // 延迟初始化：在首次渲染时创建
    return adapter
}

// OnRender 在渲染后调用，应用选择高亮
func (a *RuntimeAdapter) OnRender(frame *runtime.Frame) {
    if a.textSelection == nil {
        a.textSelection = NewTextSelection(frame.Buffer)
    } else {
        a.textSelection.UpdateBuffer(frame.Buffer)
    }

    a.textSelection.ApplySelection()
}

// OnEvent 在事件处理前调用
func (a *RuntimeAdapter) OnEvent(ev interface{}) bool {
    if a.textSelection == nil {
        return false
    }
    return a.textSelection.HandleEvent(ev)
}

// GetSelection 返回选择系统
func (a *RuntimeAdapter) GetSelection() *TextSelection {
    return a.textSelection
}
```

## 使用示例

### 方式 1: 通过 RuntimeImpl 直接使用

```go
// 在初始化时
runtime := runtime.NewRuntime(80, 24)
runtime.EnableSelection(true)

// 在事件处理中
runtime.HandleSelectionEvent(mouseEvent)

// 复制选中文字
text, err := runtime.CopySelection()
```

### 方式 2: 通过 Model 使用（推荐）

```go
// 在 Model 中添加字段
type Model struct {
    // ... 现有字段
    TextSelection *selection.TextSelection
}

// 在 Init() 中初始化
func (m *Model) Init() tea.Cmd {
    // ... 现有代码

    // 初始化文本选择
    m.TextSelection = selection.NewTextSelection(
        m.Runtime.GetLastFrame().Buffer,
    )
    m.TextSelection.SetEnabled(true)

    // ...
}

// 在 Update() 中处理事件
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 先让选择系统处理
    if m.TextSelection != nil {
        if m.TextSelection.HandleEvent(msg) {
            return m, nil  // 事件被处理，不需要进一步传播
        }
    }

    // 继续正常的事件处理...
}
```

### 方式 3: 使用全局选择管理器（最简单）

```go
import "github.com/yaoapp/yao/tui/runtime/selection"

// 在首次渲染后初始化
selection.InitGlobalSelection(frame.Buffer)

// 在事件处理中
selection.GetGlobalSelection().HandleEvent(mouseEvent)

// 复制
text, err := selection.CopyToClipboardGlobal()
```

## 测试清单

- [ ] 鼠标单击拖动选择文字
- [ ] 双击选择单词
- [ ] 三击选择整行
- [ ] Ctrl+C 复制到剪贴板
- [ ] Ctrl+X 剪切
- [ ] Ctrl+A 全选
- [ ] Escape 清除选择
- [ ] 选择高亮正确显示
- [ ] 多行选择正确工作
- [ ] Windows/macOS/Linux 剪贴板都正常

## 注意事项

1. **模块边界**：`runtime` 包不能直接依赖 `selection` 包
   - 解决方案：使用接口或适配器模式

2. **渲染顺序**：选择高亮必须在所有组件渲染后应用
   - 在 `Render()` 方法最后调用 `ApplySelection()`

3. **事件优先级**：选择事件应该在组件事件之前处理
   - 避免干扰其他组件的交互

4. **性能考虑**：选择高亮需要遍历所有选中的单元格
   - 对于大缓冲区可能有性能影响

5. **终端支持**：需要终端支持鼠标事件
   - 使用 `tea.WithMouseCellMotion()` 启用
