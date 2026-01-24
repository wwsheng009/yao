# TUI Examples

This directory contains demonstration applications showcasing the Yao TUI framework capabilities.

## Examples Overview

### [Todo App](./todo_app/)
**Level**: Beginner
**Components**: Input, Button, List, Header, Footer

A simple todo list application demonstrating:
- Basic component usage
- Event handling (keyboard input)
- State management
- Focus coordination
- List manipulation

**Features:**
- Add todos with Enter key
- Navigate with Tab/Arrow keys
- Toggle completion with Space
- Delete selected todos
- View statistics in footer

**Run:**
```bash
go run tui/examples/todo_app/main.go
```

### [Dashboard App](./dashboard_app/)
**Level**: Intermediate
**Components**: Progress, Table, List, Header, Footer

A system monitoring dashboard demonstrating:
- Multiple progress bars
- Data table with columns
- Log list display
- Component focus management
- Interactive updates

**Features:**
- Three progress bars (CPU, Memory, Disk)
- Server status table
- System log viewer
- Focus cycling with Tab
- Simulate data updates (1, 2, 3 keys)

**Run:**
```bash
go run tui/examples/dashboard_app/main.go
```

## Component Reference

All examples use production-ready components from `tui/ui/components/`:

| Component | Description | Example Usage |
|-----------|-------------|---------------|
| **Input** | Text input with validation | Add todo items |
| **Button** | Clickable action buttons | Add/Remove actions |
| **List** | Scrollable item list | Display todos/logs |
| **Table** | Tabular data display | Show server status |
| **Progress** | Progress indicator | Show resource usage |
| **Header** | Title bar with styling | App title |
| **Footer** | Status bar with information | Statistics |
| **Tree** | Hierarchical data display | File browser |
| **Tabs** | Tabbed interface | Multiple panels |
| **Modal** | Overlay dialog | Confirm dialogs |
| **ContextMenu** | Popup menu | Context actions |
| **SplitPane** | Resizable panels | Split view |

## Running Examples

### Prerequisites
Ensure you're in the project root directory:
```bash
cd E:/projects/yao/wwsheng009/yao
```

### Run an Example
```bash
# Simple example
go run tui/examples/todo_app/main.go

# Advanced example
go run tui/examples/dashboard_app/main.go
```

### Build and Run
```bash
# Build
go build -o todo-demo tui/examples/todo_app/main.go

# Run
./todo-demo
```

## Keyboard Shortcuts (Common)

| Key | Action |
|-----|--------|
| Tab | Cycle focus forward |
| Shift+Tab | Cycle focus backward |
| ↑/↓ or j/k | Navigate items (list/tree) |
| Enter | Select item / Confirm action |
| Space | Toggle state (check/expand) |
| Esc / q | Quit application |
| Ctrl+C | Force quit |

## Learning Path

### 1. Start with Todo App
Begin with the todo app to understand:
- Basic component structure
- Event handling patterns
- State management
- Focus coordination

### 2. Explore Dashboard App
Progress to the dashboard to learn:
- Multiple component coordination
- Data display patterns
- Progress indicators
- Table rendering

### 3. Study Component Code
Review source files in `tui/ui/components/`:
- `input.go` - Input handling
- `list.go` - List rendering
- `table.go` - Table layout
- `button.go` - Button actions

### 4. Build Custom Applications
Combine components to create:
- Admin interfaces
- Monitoring dashboards
- Data viewers
- Interactive forms

## Component API Patterns

### Initialization
```go
component := components.NewComponent().
    WithID("my-component").
    WithConfig(value)
```

### Event Handling
```go
component.WithOnChange(func(value string) {
    // Handle change
})
```

### Styling
```go
component.
    WithBold(true).
    WithColor("#ff0000").
    WithBackground("#ffffff")
```

### Focus Management
```go
component.SetFocus(true)
if component.IsFocusable() {
    // Component can receive focus
}
```

## Architecture Patterns

### Model Structure
```go
type MyModel struct {
    components []*Component
    focused    int
    state      MyState
}

func (m *MyModel) Init() tea.Cmd {
    return nil
}

func (m *MyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle updates
    return m, nil
}

func (m MyModel) View() string {
    // Render UI
    return ""
}
```

### Component Composition
```go
func (m *MyModel) View() string {
    var b strings.Builder

    // Render header
    b.WriteString(m.header.View())
    b.WriteString("\n")

    // Render content
    b.WriteString(m.content.View())
    b.WriteString("\n")

    // Render footer
    b.WriteString(m.footer.View())

    return b.String()
}
```

### Focus Cycling
```go
func (m *MyModel) cycleFocus() {
    m.focused = (m.focused + 1) % len(m.components)

    for i, comp := range m.components {
        comp.SetFocus(i == m.focused)
    }
}
```

## Advanced Topics

### RenderToBuffer
For better performance, use direct CellBuffer rendering:
```go
func (c *Component) RenderToBuffer(buffer *runtime.CellBuffer, x, y, width, height int) {
    // Direct rendering for better performance
}
```

### Layout System
Use the runtime layout engine:
```go
import "github.com/yaoapp/yao/tui/runtime"

node := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, style)
node.AddComponent(component)
```

### Animation
Add smooth transitions:
```go
import "github.com/yaoapp/yao/tui/runtime/animation"

anim := animation.FadeIn("fade-in", 500*time.Millisecond)
runtime.AddAnimation(anim)
```

## Tips and Best Practices

### DO:
- ✅ Keep Update() method simple
- ✅ Use With* methods for configuration
- ✅ Handle errors gracefully
- ✅ Test components independently
- ✅ Document keyboard shortcuts
- ✅ Provide visual feedback for focus

### DON'T:
- ❌ Block in Update() method
- ❌ Ignore accessibility
- ❌ Hardcode component sizes
- ❌ Skip error handling
- ❌ Overcomplicate simple tasks
- ❌ Forget to handle quit events

## Troubleshooting

### Component Not Showing
1. Check visibility state
2. Verify component has content
3. Ensure proper sizing
4. Check z-index (for overlays)

### Events Not Working
1. Verify focus state
2. Check event is subscribed
3. Ensure key bindings are correct
4. Add debug logging

### Performance Issues
1. Use RenderToBuffer for complex UIs
2. Implement virtual scrolling
3. Debounce rapid updates
4. Clear unused resources

## Contributing

To add new examples:

1. Create directory in `examples/your_app/`
2. Implement `main.go` with tea.Model
3. Add `README.md` with documentation
4. Update this README with your example
5. Test on multiple terminals

## Resources

- [Component Documentation](../../ui/components/)
- [TUI Runtime Guide](../../runtime/README.md)
- [Layout System](../../runtime/DESIGN.md)
- [Animation System](../../runtime/animation/)

## License

Same as parent Yao project.
