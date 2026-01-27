# Table 组件完整样式配置示例

## 对应 bubbles/table 原生 API 的配置

你想要的配置：
```go
s := table.DefaultStyles()
s.Header = s.Header.
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("240")).
    BorderBottom(true).
    Bold(false)
s.Selected = s.Selected.
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("57")).
    Bold(false)
t.SetStyles(s)
```

## 等价的 DSL 配置

### 方式 1: 完整 JSON 配置

```json
{
  "type": "table",
  "id": "myTable",
  "props": {
    "columns": [
      {"key": "name", "title": "Name", "width": 30},
      {"key": "status", "title": "Status", "width": 15}
    ],
    "showBorder": true,
    "focused": true,
    
    // Header 样式配置
    "headerColor": "",           // 不设置前景色（使用默认）
    "headerBackground": "",      // 不设置背景色
    "headerBold": false,         // Bold(false)
    
    // 边框配置
    "borderColor": "240",        // BorderForeground(lipgloss.Color("240"))
    "borderStyle": "normal",     // BorderStyle(lipgloss.NormalBorder())
    "borderBottom": true,        // BorderBottom(true) - 通常是默认的
    
    // 选中行样式
    "selectedColor": "229",      // Foreground(lipgloss.Color("229"))
    "selectedBackground": "57",  // Background(lipgloss.Color("57"))
    "selectedBold": false        // Bold(false)
  }
}
```

### 方式 2: Go 代码链式配置

```go
package main

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/yaoapp/yao/tui/component"
)

func main() {
    columns := []components.RuntimeColumn{
        {Key: "name", Title: "Name", Width: 30},
        {Key: "status", Title: "Status", Width: 15},
    }
    
    rows := [][]interface{}{
        {"Server 1", "Running"},
        {"Server 2", "Stopped"},
        {"Server 3", "Running"},
    }
    
    // 完全等价于你提供的 bubbles/table 配置
    table := components.NewTable().
        WithColumns(columns).
        WithData(rows).
        WithFocused(true).
        WithHeight(7).
        
    // Header 样式：对应 s.Header 配置
    table.WithHeaderBold(false).  // Bold(false)
        
    // 边框样式：对应 BorderStyle + BorderForeground + BorderBottom
    table.WithBorderType(lipgloss.NormalBorder()).      // BorderStyle
    table.WithBorderColor("240").                       // BorderForeground
    table.WithShowBorder(true).                         // 启用边框（BorderBottom 自动为 true）
    
    // Selected 样式：对应 s.Selected 配置
    table.WithSelectedColor("229").                     // Foreground
    table.WithSelectedBackground("57").                 // Background
    table.WithSelectedBold(false)                       // Bold(false)
}
```

## 配置项完整对照表

| bubbles/table API | DSL 属性 | Go 链式方法 | 说明 |
|-------------------|---------|------------|------|
| `BorderStyle(lipgloss.NormalBorder())` | `borderStyle: "normal"` | `WithBorderType(lipgloss.NormalBorder())` | 设置边框类型 |
| `BorderForeground(lipgloss.Color("240"))` | `borderColor: "240"` | `WithBorderColor("240")` | 边框颜色 |
| `BorderBottom(true)` | `borderBottom: true` | `WithShowBorder(true)` | 显示底部边框（通常是默认的） |
| `Bold(false)` | `headerBold: false` | `WithHeaderBold(false)` | 表头不加粗 |
| `Foreground(lipgloss.Color("229"))` | `selectedColor: "229"` | `WithSelectedColor("229")` | 选中行前景色 |
| `Background(lipgloss.Color("57"))` | `selectedBackground: "57"` | `WithSelectedBackground("57")` | 选中行背景色 |
| `Bold(false)` | `selectedBold: false` | `WithSelectedBold(false)` | 选中行不加粗 |

## 完整的 TUI DSL YAML 示例

```yaml
# example-table.tui.yao
{
  "name": "Styled Table Example",
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "table",
        "id": "myTable",
        "height": 10,
        "width": 60,
        "props": {
          "columns": [
            {"key": "name", "title": "名称", "width": 30},
            {"key": "value", "title": "数值", "width": 15},
            {"key": "status", "title": "状态", "width": 10}
          ],
          
          // 数据
          "data": [
            ["项目 A", "100", "完成"],
            ["项目 B", "200", "进行中"],
            ["项目 C", "150", "待处理"]
          ],
          
          // 显示设置
          "showBorder": true,
          "focused": true,
          
          // === Header 样式（对应 s.Header 配置）===
          "headerColor": "",           // 使用默认前景色
          "headerBackground": "",      // 使用默认背景色
          "headerBold": false,         // Bold(false) - 不加粗
          
          // === 边框样式（对应 BorderStyle + BorderForeground + BorderBottom）===
          "borderColor": "240",        // BorderForeground(lipgloss.Color("240"))
          "borderStyle": "normal",     // BorderStyle(lipgloss.NormalBorder())
          "borderBottom": true,        // BorderBottom(true) - 显示底部边框
          
          // === 选中行样式（对应 s.Selected 配置）===
          "selectedColor": "229",      // Foreground(lipgloss.Color("229"))
          "selectedBackground": "57",  // Background(lipgloss.Color("57"))
          "selectedBold": false        // Bold(false) - 不加粗
        }
      }
    ]
  }
}
```

## 关键点说明

### 1. BorderBottom 通常是默认的

在 `bubbles/table` 中，当你设置了 `BorderStyle` 后，`BorderBottom` 默认就是 `true`。所以：

```go
// 这两个配置等价
s.Header = s.Header.BorderStyle(lipgloss.NormalBorder())
// 和
s.Header = s.Header.
    BorderStyle(lipgloss.NormalBorder()).
    BorderBottom(true)
```

在我们的 DSL 中也是一样：
```json
{
  "borderStyle": "normal"
  // borderBottom 默认就是 true，不需要显式设置
}
```

### 2. 如果要隐藏底部边框

只有在你明确想**隐藏**底部边框时才需要设置：

```json
{
  "borderBottom": false  // 隐藏底部边框
}
```

### 3. Bold(false) 的作用

```go
s.Header = s.Header.Bold(false)
```

这意味着表头**不加粗**。默认情况下 `bubbles/table` 的表头是加粗的，设置为 `false` 后就使用正常字体。

对应 DSL：
```json
{
  "headerBold": false
}
```

### 4. 颜色的 ANSI 代码

- `"240"` - 深灰色（用于边框）
- `"229"` - 浅粉色/桃色（用于选中行文字）
- `"57"` - 深紫色/蓝紫色（用于选中行背景）

这些是标准的 256 色 ANSI 颜色代码。

## 验证配置是否生效

创建测试程序：

```go
package main

import (
    "fmt"
    "github.com/yaoapp/yao/tui/component"
)

func main() {
    // 按照你的要求配置表格
    table := components.NewTable().
        WithBorderType(lipgloss.NormalBorder()).
        WithBorderColor("240").
        WithHeaderBold(false).
        WithSelectedColor("229").
        WithSelectedBackground("57").
        WithSelectedBold(false)
    
    // 验证配置
    fmt.Printf("showBorder: %v\n", table.showBorder)
    fmt.Printf("borderType: %v\n", table.borderType)
    fmt.Printf("borderColor: %v\n", table.borderStyle.GetForeground())
    fmt.Printf("headerBold: %v\n", table.headerStyle.GetBold())
    fmt.Printf("selectedColor: %v\n", table.selectedStyle.GetForeground())
    fmt.Printf("selectedBackground: %v\n", table.selectedStyle.GetBackground())
    fmt.Printf("selectedBold: %v\n", table.selectedStyle.GetBold())
}
```

运行后应该看到：
```
showBorder: false         ← 需要设置为 true
borderType: NormalBorder  ✓
borderColor: 240          ✓
headerBold: false         ✓
selectedColor: 229        ✓
selectedBackground: 57    ✓
selectedBold: false       ✓
```

## 完整的正确配置

要让配置完全等价于你的示例，需要：

```go
table := components.NewTable().
    WithBorderType(lipgloss.NormalBorder()).      // ✓
    WithBorderColor("240").                       // ✓
    WithShowBorder(true).                         // ⚠️ 必须显式设置！
    WithHeaderBold(false).                        // ✓
    WithSelectedColor("229").                     // ✓
    WithSelectedBackground("57").                 // ✓
    WithSelectedBold(false)                       // ✓
```

或者使用 DSL：

```json
{
  "props": {
    "showBorder": true,           // ⚠️ 重要！必须设置为 true
    "borderStyle": "normal",
    "borderColor": "240",
    "headerBold": false,
    "selectedColor": "229",
    "selectedBackground": "57",
    "selectedBold": false
  }
}
```

现在 `showBorder` 配置应该能正常工作了！
