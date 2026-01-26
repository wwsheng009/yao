# 光标闪烁原理与逻辑分析

## 概述

本文档详细说明TUI框架中光标的闪烁机制、位置计算和用户输入处理流程。

---

## 一、光标闪烁原理

### 1.1 闪烁状态机

```
时间轴: ────────────────────────────────────────────────────────────▶
       │     │     │     │     │     │     │     │     │     │
       0ms  500ms  1s   1.5s  2s   2.5s  3s   3.5s  4s ...

可见性: ┌─────┐──────┌─────┐──────┌─────┐──────┌─────┐
       │ ON  │ OFF  │ ON  │ OFF  │ ON  │ OFF  │ ON  │
       └─────┘──────└─────┘──────└─────┘──────└─────┘

绘制:  [███]     [███]     [███]     [███]     [███]
       绘制     跳过      绘制     跳过      绘制
       反转     (提前     反转     (提前     反转
       样式     返回)     样式     返回)     样式
```

### 1.2 IsVisible() 方法

```go
func (c *Cursor) IsVisible() bool {
    c.mu.Lock()
    defer c.mu.Unlock()

    if !c.blinkEnabled {
        return true  // 禁用闪烁时始终可见
    }

    now := time.Now()
    // 每次调用都检查是否需要切换状态
    if now.Sub(c.lastBlinkTime) >= c.blinkInterval {
        c.visible = !c.visible  // 切换状态
        c.lastBlinkTime = now
    }

    return c.visible
}
```

**关键点**：
- `IsVisible()` 在每次 `Paint()` 时被调用
- 它根据时间自动切换 `visible` 状态
- 返回 `false` 时，`Paint()` 提前返回，不绘制任何东西

---

## 二、光标位置计算

### 2.1 坐标系统

```
TextInput 的渲染布局:

相对坐标 (相对于TextInput的左上角):
  ┌───┬───┬───┬───┬───┬───┬───┐
  │ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ ...
  ├───┼───┼───┼───┼───┼───┼───┤
  │ [ │ a │ b │ c │ ] │   │   │
  └───┴───┴───┴───┴───┴───┴───┘
    ↑   ↑   ↑   ↑   ↑
  左边 第 第 第 右边
  框  一 二 三 框
      个 个 个 字
      字 字 符 字
      符 符

绝对坐标 (相对于终端):
  absX = ctx.X + 相对X
  absY = ctx.Y + 相对Y
```

### 2.2 光标位置计算公式

```go
// TextInput.Paint() 中的计算
cursorX = 1 + cursorPos  // 左边框(1) + 光标索引
absCursorX = ctx.X + cursorX
```

**示例**：

| cursorPos | cursorX | ctx.X | absCursorX | 位置说明 |
|-----------|---------|-------|------------|----------|
| 0 | 1 | 2 | 3 | 第1个字符位置 |
| 1 | 2 | 2 | 4 | 第2个字符位置 |
| 2 | 3 | 2 | 5 | 第3个字符位置 |
| 3 | 4 | 2 | 6 | 第4个字符位置 |

**问题**：当 value="a", cursorPos=1 时：
- 字符'a'在位置3
- 右括号']'在位置4
- 光标计算位置是4（右括号），而不是3（字符'a'）

---

## 三、用户输入后的处理流程

### 3.1 完整调用链

```
用户按键 'a'
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│ KeyEvent('a')                                              │
│    │                                                      │
│    ▼                                                      │
│ TextInput.HandleEvent()                                   │
│    │                                                      │
│    ▼                                                      │
│ TextInput.HandleAction(ActionInputChar)                    │
│    │                                                      │
│    ▼                                                      │
│ TextInput.handleInputChar('a')                             │
│    ├─ t.value = "a" (插入字符)                            │
│    ├─ t.cursor++ (t.cursor = 1)                           │
│    └─ return true                                         │
│         │                                                │
│         ▼                                                │
│ [框架] MarkDirty() → App.render() 被调度                   │
│         │                                                │
│         ▼                                                │
┌─────────────────────────────────────────────────────────────┐
│ App.render()                                               │
│    │                                                      │
│    ├─ buf = paint.NewBuffer(width, height)  ← 新缓冲区      │
│    │                                                      │
│    ├─ paintable.Paint(ctx, buf)                            │
│    │    │                                                 │
│    │    ▼                                                 │
│    │ Form.Paint()                                         │
│    │    │                                                 │
│    │    ├─ f.drawText(label)  // 绘制标签                  │
│    │    │                                                 │
│    │    └─ txt.Paint(inputCtx, buf)                       │
│    │         │                                             │
│    │         ▼                                             │
│    │    ┌─────────────────────────────────────────────────┐ │
│    │    │ TextInput.Paint(ctx, buf)                       │ │
│    │    │    │                                          │ │
│    │    │    ├─ 读取状态: value="a", cursor=1             │ │
│    │    │    │                                          │ │
│    │    │    ├─ 绘制边框: [a]                           │ │
│    │    │    │     buf.SetCell(2, y, '[', style)        │ │
│    │    │    │     buf.SetCell(3, y, 'a', style)        │ │
│    │    │    │     buf.SetCell(4, y, ']', style)        │ │
│    │    │    │                                          │ │
│    │    │    ├─ 计算光标位置:                            │ │
│    │    │    │     cursorX = 1 + 1 = 2                   │ │
│    │    │    │     absCursorX = 2 + 2 = 4                │ │
│    │    │    │                                          │ │
│    │    │    └─ 调用 Cursor.Paint():                     │ │
│    │    │         │                                     │ │
│    │    │         ▼                                     │ │
│    │    │    ┌─────────────────────────────────────────┐ │ │
│    │    │    │ Cursor.Paint(ctx, buf)                  │ │ │
│    │    │    │    │                                  │ │ │
│    │    │    │    ├─ IsVisible()?                      │ │ │
│    │    │    │    │   │                               │ │ │
│    │    │    │    │   ├─ true → 继续绘制               │ │ │
│    │    │    │    │   │   │                          │ │ │
│    │    │    │    │   │   ├─ 计算绝对位置             │ │ │
│    │    │    │    │   │   │  x = ctx.X + c.x         │ │ │
│    │    │    │    │   │   │  y = ctx.Y + c.y         │ │ │
│    │    │    │    │   │   │                          │ │ │
│    │    │    │    │   │   ├─ 获取单元格                │ │ │
│    │    │    │    │   │   │  cell = buf.Cells[y][x]  │ │ │
│    │    │    │    │   │   │                          │ │ │
│    │    │    │    │   │   └─ 绘制反转样式             │ │ │
│    │    │    │    │   │     buf.SetCell(x, y, ...,   │ │ │
│    │    │    │    │   │            reverseStyle)       │ │ │
│    │    │    │    │   │                          │ │ │
│    │    │    │    │    └─ false → 提前返回 (旧光标消失) │ │ │
│    │    │    └─────────────────────────────────────────┘ │ │
│    │    └─────────────────────────────────────────────────┘ │
│    │                                                      │
│    └─ a.outputBuffer(buf)  ← 输出到终端                  │
│                                                          │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 旧光标如何"消失"

**关键理解**：光标不是"移动"的，而是每帧重新计算位置！

```
第N帧 (cursorPos=0):          第N+1帧 (cursorPos=1)       第N+2帧 (cursorPos=1):
┌─────────────────┐           ┌─────────────────┐          ┌─────────────────┐
│ Buffer (全新)   │           │ Buffer (全新)   │          │ Buffer (全新)   │
│                 │           │                 │          │                 │
│ [a]             │           │ [a]             │          │ [a]             │
│  ↑ (位置2)       │           │ [a]             │          │ [a]             │
│ 反转            │           │   ↑ (位置3)     │          │   ↑ (位置3)     │
│                 │           │ 反转            │          │ 反转/隐藏(闪烁) │
└─────────────────┘           └─────────────────┘          └─────────────────┘
                                                           ↑
                                                    旧位置3没有绘制，
                                                    自动"消失"了
```

**不需要"清除"旧光标**：
1. 每次创建新Buffer（全0）
2. 或者使用diff只输出变化的单元格
3. 位置改变后，旧位置不再被绘制，自然消失

---

## 四、当前实现的问题

### 4.1 问题1：光标位置错误

**现象**：
- value="a", cursorPos=1
- 光标应该在字符'a'上（位置3）
- 实际光标在右括号']'上（位置4）

**原因**：
```
cursorX = 1 + cursorPos = 1 + 1 = 2 (相对位置)
absCursorX = ctx.X + cursorX = 2 + 2 = 4 (绝对位置)

位置布局: [  a  ]  ]
          2  3  4  5
              ↑ 光标在这里
```

### 4.2 问题2：光标闪烁时位置跳变

**可能原因**：
1. 某次Paint时cursorPos被重置为0
2. SetPosition没有被正确调用
3. IsVisible()在关键时刻返回false

---

## 五、修复方案

### 5.1 正确的光标位置计算

```go
// 块状光标应该高亮当前光标位置的字符
// cursorPos 表示要高亮的字符索引（0-based）

var cursorX int
if len(runes) == 0 {
    // 空输入，光标在左边框后
    cursorX = 1
} else if cursorPos >= len(runes) {
    // 光标超出范围，高亮最后一个字符
    cursorX = 1 + len(runes) - 1
} else {
    // 正常情况，高亮 cursorPos 位置的字符
    cursorX = 1 + cursorPos
}
```

### 5.2 修复后的行为

| 状态 | value | cursorPos | 光标位置 | 高亮字符 |
|------|-------|-----------|----------|----------|
| 空输入 | "" | 0 | 3 | 无（左边框后） |
| 输入a后 | "a" | 1 | 3 | 'a' |
| 输入ab后 | "ab" | 2 | 4 | 'b' |
| 输入abc后 | "abc" | 3 | 5 | 'c' |

---

## 六、调试方法

### 6.1 启用调试日志

```bash
# TextInput调试（输出到文件）
TUI_INPUT_DEBUG=1 go run ...

# Cursor调试（输出到stderr）
TUI_CURSOR_DEBUG=1 go run ...

# 同时启用
TUI_INPUT_DEBUG=1 TUI_CURSOR_DEBUG=1 go run ...
```

### 6.2 关键日志信息

```
[TextInput] PAINT: ctx=(X,Y), value='...', cursor=N, focused=true/false
[TextInput] FOCUS CURSOR: logical=N, relative=(X,Y), absolute=(X,Y)
[Cursor] FOCUS: absolute=(X,Y), ctx=(X,Y), relative=(X,Y)
[Cursor] FOCUS RENDER: drew cursor at (X,Y) on char 'X', reverse=true/false
[TextInput] FOCUS RESULT: (X,Y)='X' reverse=true/false
```

### 6.3 问题诊断检查清单

- [ ] cursorPos在输入后是否正确增加？
- [ ] SetPosition是否每次Paint都被调用？
- [ ] IsVisible()是否在关键时刻返回false？
- [ ] 绝对位置计算是否正确？
- [ ] 有多次Paint调用导致位置不一致？
