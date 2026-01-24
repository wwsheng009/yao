# 文本选择功能集成到现有 TUI Runtime 的步骤

## 概述

文本选择功能已经完整实现在 `tui/runtime/selection/` 包中。要将其集成到现有的 TUI Runtime，需要进行以下修改。

## 方式一：最小化集成（推荐）

这种方式对现有代码改动最小，通过在 Model 层添加选择功能。

### 1. 修改 `tui/model.go`

在 Model 结构体中添加选择字段：

```go
// Model 是 TUI 应用程序的状态
type Model struct {
    // ... 现有字段 ...

    // 新增：文本选择支持
    textSelection *selection.TextSelection
    selectionEnabled bool
}
```

在 `Init()` 中初始化：

```go
func (m *Model) Init() tea.Cmd {
    // ... 现有代码 ...

    // 初始化文本选择（在 Runtime 初始化后）
    if m.UseRuntime {
        m.selectionEnabled = true
        // textSelection 会在首次渲染时初始化
    }

    // ...
}
```

在 `Update()` 中添加事件处理：

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 处理 Ctrl+C 退出
    if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyCtrlC {
        return m, tea.Quit
    }

    // 新增：处理选择事件（在其他事件处理之前）
    if m.selectionEnabled && m.textSelection != nil {
        // 转换事件
        runtimeEvent := m.EventAdapter.Convert(msg)

        // 让选择系统处理
        if handled := m.textSelection.HandleEvent(runtimeEvent); handled {
            // 如果是选择操作，需要刷新视图
            m.messageReceived = true
            return m, nil
        }
    }

    // ... 现有事件处理代码 ...
}
```

在 `View()` 中应用选择高亮：

```go
func (m Model) View() string {
    if !m.UseRuntime {
        return m.viewLegacy()
    }

    // 布局
    result := m.Runtime.Layout(m.LayoutRoot, m.constraints)

    // 渲染
    frame := m.Runtime.Render(result)

    // 新增：初始化或更新选择系统
    if m.selectionEnabled {
        if m.textSelection == nil {
            m.textSelection = selection.NewTextSelection(frame.Buffer)
            m.textSelection.SetEnabled(true)
        } else {
            m.textSelection.UpdateBuffer(frame.Buffer)
        }

        // 应用选择高亮
        m.textSelection.ApplySelection()
    }

    return frame.String()
}
```

### 2. 修改 `tui/config.go`

添加配置选项（可选）：

```go
type Config struct {
    // ... 现有字段 ...

    // TextSelection 文本选择配置
    EnableTextSelection *bool   // 是否启用文本选择
}
```

## 方式二：深度集成到 Runtime

这种方式将选择功能直接集成到 RuntimeImpl 中。

### 需要修改的文件

**1. `tui/runtime/runtime_impl.go`**

添加导入：
```go
import (
    "time"
    sel "github.com/yaoapp/yao/tui/runtime/selection"
    "github.com/yaoapp/yao/tui/runtime/animation"
)
```

修改结构体：
```go
type RuntimeImpl struct {
    // ... 现有字段 ...
    selectionAdapter *sel.RuntimeAdapter
    selectionEnabled bool
}
```

修改 `NewRuntime`：
```go
func NewRuntime(width, height int) *RuntimeImpl {
    return &RuntimeImpl{
        // ... 现有初始化 ...
        selectionEnabled: true,
        selectionAdapter: sel.NewRuntimeAdapter(),
    }
}
```

修改 `Render()` 方法（在返回 frame 之前）：
```go
frame := Frame{
    Buffer: buf,
    Width:  r.width,
    Height: r.height,
    Dirty:  len(r.dirtyRegions) > 0 || r.isDirty,
}

// 应用选择高亮
if r.selectionEnabled {
    r.selectionAdapter.OnRender(&frame)
}

return frame
```

添加选择相关方法：
```go
// EnableSelection 启用/禁用文本选择
func (r *RuntimeImpl) EnableSelection(enabled bool)

// HandleSelectionEvent 处理选择事件
func (r *RuntimeImpl) HandleSelectionEvent(ev interface{}) bool

// CopySelection 复制选中文字
func (r *RuntimeImpl) CopySelection() (string, error)

// ... 其他方法见 runtime_integration.go
```

**2. `tui/model.go`**

在 `handleGeometryEvent` 中调用：
```go
func (m *Model) handleGeometryEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 转换事件
    runtimeEvent := m.EventAdapter.Convert(msg)

    // 让选择系统处理
    if m.Runtime.HandleSelectionEvent(runtimeEvent) {
        return m, nil
    }

    // ... 现有事件处理 ...
}
```

## 使用示例

### 基本使用

```go
// 启用文本选择
model.selectionEnabled = true

// 复制选中文字
text, err := model.textSelection.Copy()
if err != nil {
    // 处理错误
}

// 清除选择
model.textSelection.Clear()
```

### 检查选择状态

```go
// 检查是否有选择
if model.textSelection.IsActive() {
    text := model.textSelection.GetSelectedText()
    fmt.Printf("选中: %s\n", text)
}

// 检查特定位置是否被选中
if model.textSelection.IsSelected(x, y) {
    // 处理选中单元格
}
```

### 全选

```go
model.textSelection.SelectAll()
```

## 文件清单

集成后涉及的文件：

| 文件 | 操作 | 说明 |
|------|------|------|
| `tui/runtime/selection/` | 已存在 | 选择功能实现 |
| `tui/runtime/selection/adapter.go` | 新增 | Runtime 适配器 |
| `tui/model.go` | 修改 | 添加选择支持 |
| `tui/runtime/runtime_impl.go` | 修改（可选）| 深度集成 |

## 测试

运行测试确保功能正常：

```bash
# 测试选择功能
go test ./tui/runtime/selection -v

# 测试 TUI
go test ./tui -v -run TestModel
```

## 注意事项

1. **终端鼠标支持**：确保启用鼠标事件
   ```go
   tea.WithMouseCellMotion()
   ```

2. **渲染顺序**：选择高亮必须在所有组件渲染后应用

3. **剪贴板支持**：Linux 用户需要安装 `xclip` 或 `wl-copy`

4. **性能**：大缓冲区可能有轻微性能影响
