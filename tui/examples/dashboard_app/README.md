# Dashboard Demo Application

A system monitoring dashboard demonstrating advanced TUI components including Progress bars, Tables, and Lists.

## Features

This demo showcases:

### Components Demonstrated
- **ProgressComponent**: Multiple progress bars with percentages
- **TableComponent**: Data table with columns and rows
- **ListComponent**: Log display with selection
- **HeaderComponent**: Centered application title
- **FooterComponent**: Status information and shortcuts

### Interactive Features
- **Focus Cycling**: Tab/Shift+Tab to cycle between components
- **List Navigation**: Arrow keys or j/k to navigate logs
- **Selection**: Enter to select a log entry
- **Progress Updates**: Press 1, 2, or 3 to update progress bars
- **Quit**: q, Esc, or Ctrl+C to exit

## Running the Demo

```bash
# From project root
go run tui/examples/dashboard_app/main.go

# Or build and run
go build -o dashboard tui/examples/dashboard_app/main.go
./dashboard
```

## Component Details

### Progress Bars
```go
progress := components.NewProgress().
    WithLabel("CPU Usage").
    WithPercent(75).
    WithShowPercentage(true).
    WithWidth(40)
```

Features:
- Real-time percentage display
- Configurable width and label
- Filled and empty character customization
- Color support

### Data Table
```go
table := components.NewTable().
    WithColumns([]*TableColumn{
        {ID: "name", Title: "Name", Width: 30},
        {ID: "status", Title: "Status", Width: 15},
        {ID: "cpu", Title: "CPU", Width: 10},
    }).
    WithData([]map[string]interface{}{
        {"name": "Server 1", "status": "Running", "cpu": "45%"},
        {"name": "Server 2", "status": "Stopped", "cpu": "0%"},
    }).
    WithShowBorder(true)
```

Features:
- Column layout (fixed and flex widths)
- Header row with underlines
- Grid lines
- Selection highlighting
- Border rendering

### Log List
```go
list := components.NewList().
    WithTitle("System Logs").
    WithItems([]*ListItem{
        {ID: "log1", Label: "[INFO] Server started"},
        {ID: "log2", Label: "[WARN] High memory usage"},
    }).
    WithShowBorder(true)
```

Features:
- Title bar
- Item selection
- Keyboard navigation
- Border rendering
- Scroll support

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Tab | Cycle focus forward |
| Shift+Tab | Cycle focus backward |
| ↑/↓ or j/k | Navigate list (when focused) |
| Enter | Select list item (when focused) |
| 1/2/3 | Update progress 1/2/3 |
| q/Esc/Ctrl+C | Quit application |

## Focus Components

The dashboard has 6 focusable components:
1. **CPU Progress Bar** - Display CPU usage
2. **Memory Progress Bar** - Display memory usage
3. **Disk Progress Bar** - Display disk usage
4. **Data Table** - Show server status
5. **Log List** - Display system logs
6. **Status Bar** - Show shortcuts

## Data Display

### Table Data
The table displays mock server data:
- Server name and status
- CPU and memory usage
- Real-time updates possible

### Log Entries
System logs with different levels:
- **INFO**: Informational messages
- **WARN**: Warning messages
- **ERROR**: Error messages

### Progress Bars
Three progress bars showing:
- CPU Usage (percentage)
- Memory Usage (percentage)
- Disk Usage (percentage)

Press 1, 2, or 3 to simulate updating the respective progress bar.

## Component Styling

### Customizing Colors
```go
progress := components.NewProgress().
    WithColor("#00ff00").      // Green
    WithFilledChar('█').      // Block character
    WithEmptyChar('░').       // Shade character
```

### Customizing Table
```go
table.WithColumns([]*TableColumn{
    {ID: "name", Title: "Name", Width: 30},
    {ID: "status", Title: "Status", Width: 15, Align: "left"},
    {ID: "cpu", Title: "CPU", Width: 10, Align: "right"},
}).
WithShowBorder(true).
WithShowGrid(true)
```

## Architecture

### Model Structure
```go
type DashboardModel struct {
    progress1, progress2, progress3 *ProgressComponent
    table                          *TableComponent
    list                           *ListComponent
    statusBar                      *FooterComponent
    header                         *HeaderComponent
    focused                        int
    err                            error
}
```

### Update Flow
1. User input received as tea.Msg
2. Update() processes the message
3. State is modified
4. View() re-renders the UI

### Focus Management
Components are focused sequentially:
- Integer `focused` tracks current component (0-5)
- Tab/Shift+Tab cycles forward/backward
- Keyboard events forwarded to focused component

## Extending the Dashboard

### Add Real Data

```go
// Simulate real-time updates
func (m *DashboardModel) startUpdates() tea.Cmd {
	return tea.TickEvery(time.Second, func(t time.Time) tea.Msg {
		// Fetch real metrics
		cpu := getCPUUsage()
		m.progress1.WithPercent(cpu)
		return nil
	})
}
```

### Add Refresh Button
```go
refreshBtn := components.NewButton("Refresh").
    WithOnClick(func() {
        // Reload data
        m.refreshData()
    })
```

### Add Filters
```go
filterInput := components.NewInput().
    WithPlaceholder("Filter logs...").
    WithOnChange(func(value string) {
        m.filterLogs(value)
    })
```

### Add Charts
```go
// Create a bar chart using Lists
chart := components.NewList().
   WithTitle("Bar Chart").
    WithItems(createChartItems(data))
```

## Performance Considerations

### Optimization Tips

1. **Virtual Scrolling**: For large data sets
   ```go
   list.WithVirtualized(true).
       WithVisibleRange(0, 20)
   ```

2. **Debouncing**: For rapid updates
   ```go
   input.WithOnChange(debounce(func(value string) {
       m.updateFilter(value)
   }, 300*time.Millisecond))
   ```

3. **Lazy Loading**: Load data on demand
   ```go
   if m.table.NeedsData() {
       m.table.LoadData(fetchData())
   }
   ```

### Memory Management

- Clear old data before loading new
- Use pagination for large datasets
- Implement data caching strategies

## Testing

### Component Testing
```go
func TestDashboardModel(t *testing.T) {
    model := NewDashboardModel()

    // Test focus cycling
    model.focused = 0
    model.Update(tea.KeyMsg{Type: tea.KeyTab})

    assert.Equal(t, 1, model.focused)
}
```

### Integration Testing
```go
func TestComponentInteraction(t *testing.T) {
    model := NewDashboardModel()

    // Simulate user interaction
    model.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // Verify state changes
    assert.True(t, model.hasActiveSelection())
}
```

## See Also

- [Todo Demo](../todo_app/) - Simpler component example
- [Component Documentation](../../ui/components/)
- [Progress Component](../../ui/components/progress.go)
- [Table Component](../../ui/components/table.go)
- [TUI Framework Guide](../../runtime/)
