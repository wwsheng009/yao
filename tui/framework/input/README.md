# Input

输入组件集合。

## 职责

- 文本输入组件
- 数值输入组件
- 输入验证和格式化

## 组件列表

- `TextInput` - 文本输入
- `NumberInput` - 数值输入
- `PasswordInput` - 密码输入
- `SearchInput` - 搜索输入

## 相关文件

- `textinput.go` - 文本输入组件
- `numberinput.go` - 数值输入组件
- `passwordinput.go` - 密码输入组件
- `cursor.go` - 光标闪烁管理器

---

## 光标闪烁实现

### 问题描述

在实现 TextInput 组件的光标闪烁功能时，遇到了以下几个问题：

1. **光标完全不可见**
   - 原因：渲染函数 `outputBuffer()` 只输出字符，忽略了 Cell 中的 Style 信息
   - 表现：即使设置了 `drawStyle.Reverse(true)`，终端也没有显示反白效果

2. **整个屏幕闪烁**
   - 原因：每次渲染都使用 `\x1b[2J` 清屏，且逐字符打印导致终端渲染不连贯
   - 表现：光标切换时整个界面都在闪烁

3. **焦点未正确传播**
   - 原因：Form 组件的 `OnFocus()` 没有调用子组件的 `OnFocus()`
   - 表现：光标状态一直为不可见，因为 `IsFocused()` 返回 false

### 解决方案

#### 1. 样式渲染修复

**修改文件**: `tui/framework/app.go`

在 `outputBuffer()` 函数中添加样式支持：

```go
// 跟踪当前样式，避免重复输出
var currentStyle style.Style

for y := 0; y < buf.Height; y++ {
    for x := 0; x < buf.Width; x++ {
        cell := buf.Cells[y][x]
        char := cell.Char
        if char == 0 {
            char = ' '
        }

        // 检查样式是否改变
        if cell.Style != currentStyle {
            if currentStyle != (style.Style{}) {
                output.WriteString("\x1b[0m")  // 重置样式
            }
            if cell.Style != (style.Style{}) {
                output.WriteString(cell.Style.ToANSI())  // 应用新样式
            }
            currentStyle = cell.Style
        }

        output.WriteRune(char)
    }
    // 每行结束重置样式
    if currentStyle != (style.Style{}) {
        output.WriteString("\x1b[0m")
        currentStyle = style.Style{}
    }
}
```

#### 2. 减少屏幕闪烁

**修改文件**: `tui/framework/app.go`

添加 `firstRender` 标记，只在首次渲染时清屏：

```go
type App struct {
    // ...
    firstRender bool
}

func (a *App) outputBuffer(buf *paint.Buffer) {
    var output bytes.Buffer

    // 首次渲染时清屏，后续只移动光标到左上角
    if a.firstRender {
        output.WriteString("\x1b[2J")
        a.firstRender = false
    }
    output.WriteString("\x1b[H\x1b[?25l")  // 移动光标、隐藏终端光标

    // ... 构建输出
    fmt.Print(output.String())  // 一次性输出
}
```

#### 3. 全局光标闪烁管理器

**新建文件**: `tui/framework/input/cursor.go`

```go
var (
    cursorMutex     sync.RWMutex
    focusedInputs   []*TextInput
    lastBlinkCheck  time.Time
    dirtyCallback   func()
)

// CursorBlinkTick 更新所有光标的闪烁状态
func CursorBlinkTick() bool {
    cursorMutex.RLock()
    inputs := make([]*TextInput, len(focusedInputs))
    copy(inputs, focusedInputs)
    callback := dirtyCallback
    cursorMutex.RUnlock()

    now := time.Now()
    needsRedraw := false

    if now.Sub(lastBlinkCheck) >= 500*time.Millisecond {
        for _, txt := range inputs {
            if txt.UpdateCursorBlink() {
                needsRedraw = true
            }
        }
        lastBlinkCheck = now

        if needsRedraw && callback != nil {
            callback()
        }
    }

    return needsRedraw
}
```

#### 4. 光标状态管理

**修改文件**: `tui/framework/input/textinput.go`

```go
type TextInput struct {
    // ...
    cursorVisible bool
    lastBlinkTime time.Time
}

func (t *TextInput) UpdateCursorBlink() bool {
    t.mu.Lock()
    defer t.mu.Unlock()

    if !t.IsFocused() {
        return false
    }

    now := time.Now()
    if now.Sub(t.lastBlinkTime) >= 500*time.Millisecond {
        t.cursorVisible = !t.cursorVisible  // 切换可见性
        t.lastBlinkTime = now
        return true  // 状态改变，需要重绘
    }
    return false
}
```

#### 5. 焦点传播

**修改文件**: `tui/framework/form/form.go`

确保 Form 正确传播焦点到子组件：

```go
func (f *Form) OnFocus() {
    visibleFields := f.getVisibleFieldIndices()
    if len(visibleFields) > 0 {
        fieldName := f.fieldOrder[f.currentField]
        if field := f.fields[fieldName]; field != nil {
            if input, ok := field.Input.(*input.TextInput); ok {
                input.OnFocus()  // 传播焦点到输入组件
            }
        }
    }
}
```

### 光标渲染逻辑

在 `Paint()` 方法中，光标通过反白样式显示：

```go
if t.IsFocused() && i == t.cursor && t.cursorVisible {
    // 绘制光标（闪烁时高亮显示）
    buf.SetCell(x+i, y, runes[i], drawStyle.Reverse(true))
} else {
    buf.SetCell(x+i, y, runes[i], drawStyle)
}
```

### ANSI 转义码说明

| 转义码 | 功能 |
|--------|------|
| `\x1b[2J` | 清屏 |
| `\x1b[H` | 移动光标到左上角 |
| `\x1b[?25l` | 隐藏终端光标 |
| `\x1b[?25h` | 显示终端光标 |
| `\x1b[7m` | 反白显示（Reverse） |
| `\x1b[0m` | 重置所有样式 |

### 时序图

```
App.Init()
    │
    ├─> input.SetDirtyCallback(func() { a.dirty = true })
    │
    └─> root.OnFocus()
            │
            └─> Form.OnFocus()
                    │
                    └─> TextInput.OnFocus()
                            │
                            ├─> cursorVisible = true
                            ├─> RegisterCursor(t)
                            └─> BaseComponent.OnFocus()

Main Loop (每 16ms):
    │
    ├─> handleTick()
    │       │
    │       └─> input.CursorBlinkTick()
    │               │
    │               └─> 每 500ms: cursorVisible = !cursorVisible
    │                       │
    │                       └─> dirtyCallback() → a.dirty = true
    │
    └─> if dirty → render()
            │
            └─> outputBuffer() [带样式渲染]
```

### 注意事项

1. **终端光标 vs 假光标**
   - 终端光标：由操作系统/终端控制，使用 `\x1b[?25l` 隐藏
   - 假光标：通过反色渲染实现，可以精确控制位置和样式

2. **样式比较**
   - `Style` 结构体包含未导出字段，跨包比较时需要使用 `ToANSI()` 字符串比较
   - 当前实现依赖同包内的结构体比较

3. **性能优化**
   - 使用 `bytes.Buffer` 构建输出，一次性写入
   - 跟踪当前样式，避免重复 ANSI 码
   - 只在光标状态改变时触发重绘

### 相关文件清单

| 文件 | 修改内容 |
|------|----------|
| `tui/framework/app.go` | 添加样式渲染、firstRender 标记 |
| `tui/framework/input/textinput.go` | 光标闪烁状态、OnFocus/OnBlur |
| `tui/framework/input/cursor.go` | 全局光标管理器（新建） |
| `tui/framework/form/form.go` | 焦点传播修复 |
| `tui/framework/style/style.go` | Reverse() 方法已存在 |
