# Todo Demo Application

A demonstration TUI application showcasing the Yao TUI framework components.

## Features

This demo application showcases:

### Components Used
- **Input**: Text input field for adding new todos
- **Button**: Add button with click handler
- **List**: Display todo items with selection
- **Header**: App title with centering
- **Footer**: Status bar showing todo statistics

### Functionality
- **Add todos**: Type in the input field and press Enter
- **Navigate**: Use Tab/Shift+Tab to cycle between components
- **List navigation**: Use Arrow keys or j/k to move through todos
- **Delete todos**: Press Delete/Backspace to remove selected todo
- **Toggle complete**: Press Space to toggle completion status
- **Quit**: Press Ctrl+C or Esc to exit

## Running the Demo

### From the project root:

```bash
go run tui/examples/todo_app/main.go
```

### Or build and run:

```bash
go build -o todo-demo tui/examples/todo_app/main.go
./todo-demo
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Enter | Add todo (when input is focused) |
| Tab | Cycle focus forward (Input → Button → List) |
| Shift+Tab | Cycle focus backward |
| Arrow Up/Down or j/k | Navigate list items |
| Space | Toggle completed status |
| Delete/Backspace | Remove selected todo |
| Ctrl+C or Esc | Quit application |

## Architecture

This demo demonstrates:

1. **Component Composition**: How to combine multiple TUI components
2. **Event Handling**: Keyboard navigation and component interaction
3. **State Management**: Managing application state and updating UI
4. **Focus Management**: Coordinating focus across multiple components
5. **Layout**: Simple vertical layout using View() composition

## Component Integration

### Input Component
```go
input := components.NewInput().
    WithPlaceholder("Add a new todo...").
    WithWidth(50)
```

### Button Component
```go
button := components.NewButton("Add").
    WithOnClick(func() {
        // Handle click
    })
```

### List Component
```go
list := components.NewList().
   WithTitle("My Todos").
    WithSize(60, 20)
```

### Header/Footer Components
```go
header := components.NewHeader("Todo App").
    WithAlign("center").
    WithBold(true)

footer := components.NewFooter("Status text").
    WithAlign("center")
```

## Implementation Details

### Focus Management
The application uses a `focusedComponent` enum to track which component has focus:
```go
const (
    focusedInput focusedComponent = iota
    focusedList
    focusedAddButton
)
```

Focus is cycled using Tab/Shift+Tab, and appropriate key events are forwarded to the focused component.

### Event Flow
1. User presses a key
2. TodoModel.Update() receives the tea.Msg
3. Based on focus state, the event is:
   - Handled globally (Enter, Tab, Esc, etc.)
   - Forwarded to the focused component
4. State is updated
5. View() is called to re-render

### State Updates
When state changes (add/remove/toggle todo), the model:
1. Updates internal state (`m.todos`)
2. Refreshes the List display
3. Updates Footer statistics
4. Triggers re-render

## Extending the Demo

### Adding More Features

#### Edit Todos
```go
func (m *TodoModel) editTodo(id string, newText string) {
    for i, todo := range m.todos {
        if todo.ID == id {
            todo.Text = newText
            m.updateList()
            break
        }
    }
}
```

#### Persist Todos
```go
func saveTodos(todos []Todo) error {
    // Save to file or database
    return nil
}

func loadTodos() ([]Todo, error) {
    // Load from file or database
    return []Todo{}, nil
}
```

#### Add Confirm Dialog
```go
func (m *TodoModel) confirmDelete() bool {
    // Show modal dialog
    // Return true if user confirms
    return true
}
```

## Performance Considerations

The demo uses simple View() string rendering for simplicity. For production applications with complex UIs, consider using:
- **RenderToBuffer()**: Direct CellBuffer rendering for better performance
- **Virtual Scrolling**: For large lists
- **Lazy Loading**: Load todos on demand
- **Differential Rendering**: Only re-render changed components

## See Also

- [Component Documentation](../ui/components/)
- [TUI Framework Guide](../../runtime/)
- [Layout System](../../runtime/README.md)
