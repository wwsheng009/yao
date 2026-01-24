package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/ui/components"
)

// DashboardModel demonstrates a TUI dashboard with multiple components
type DashboardModel struct {
	// Components
	progress1  *components.ProgressComponent
	progress2  *components.ProgressComponent
	progress3  *components.ProgressComponent
	table       *components.TableComponent
	list        *components.ListComponent
	statusBar   *components.FooterComponent
	header      *components.HeaderComponent

	// State
	focused     int
	err         error
	width       int
	height      int
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() *DashboardModel {
	// Create progress bars
	progress1 := components.NewProgress().
		WithID("progress-1").
		WithLabel("CPU Usage").
		WithPercent(75).
		WithShowPercentage(true).
		WithWidth(40)

	progress2 := components.NewProgress().
		WithID("progress-2").
		WithLabel("Memory").
		WithPercent(60).
		WithShowPercentage(true).
		WithWidth(40)

	progress3 := components.NewProgress().
		WithID("progress-3").
		WithLabel("Disk").
		WithPercent(45).
		WithShowPercentage(true).
			WithWidth(40)

	// Create table with data
	table := components.NewTable().
		WithID("data-table").
		WithColumns([]*components.TableColumn{
			{ID: "name", Title: "Name", Width: 30},
			{ID: "status", Title: "Status", Width: 15},
			{ID: "cpu", Title: "CPU", Width: 10},
			{ID: "memory", Title: "Memory", Width: 10},
		}).
		WithData([]map[string]interface{}{
			{"name": "Server 1", "status": "Running", "cpu": "45%", "memory": "60%"},
			{"name": "Server 2", "status": "Stopped", "cpu": "0%", "memory": "0%"},
			{"name": "Server 3", "status": "Running", "cpu": "72%", "memory": "81%"},
			{"name": "Database", "status": "Running", "cpu": "30%", "memory": "45%"},
		}).
		WithSize(80, 10).
		WithShowBorder(true)

	// Create list with log entries
	list := components.NewList().
		WithID("log-list").
		WithTitle("System Logs").
		WithItems([]*components.ListItem{
			{ID: "log1", Label: "[INFO] Server started on port 8080"},
			{ID: "log2", Label: "[WARN] High memory usage detected"},
			{ID: "log3", Label: "[INFO] Backup completed successfully"},
			{ID: "log4", Label: "[ERROR] Failed to connect to database"},
			{ID: "log5", Label: "[INFO] User authentication successful"},
		}).
		WithSize(80, 8).
		WithShowBorder(true)

	// Create header
	header := components.NewHeader("📊 System Dashboard").
		WithAlign("center").
		WithBold(true).
		WithColor("#00ff00")

	// Create status bar
	statusBar := components.NewFooter("Press q to quit | Tab to cycle focus").
		WithAlign("center")

	model := &DashboardModel{
		progress1: progress1,
		progress2: progress2,
		progress3: progress3,
		table:      table,
		list:       list,
		statusBar:  statusBar,
		header:     header,
		focused:    0,
		width:      80,
		height:     24,
	}

	return model
}

// Init initializes the model
func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}
	return m, nil
}

// handleKey handles keyboard input
func (m *DashboardModel) handleKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg := msg.(tea.KeyMsg)

	switch keyMsg.Type {
	case tea.KeyCtrlC, tea.KeyEsc, 'q':
		return m, tea.Quit

	case tea.KeyTab, tea.KeyShiftTab:
		// Cycle focus
		if keyMsg.Type == tea.KeyTab {
			m.focused = (m.focused + 1) % 6
		} else {
			m.focused = (m.focused - 1 + 6) % 6
		}
		return m, nil

	// Arrow keys for navigation
	case tea.KeyUp, 'k':
		if m.focused == 5 { // List
			m.list.HandleKey(&components.KeyEvent{Key: 'k'})
		}
		return m, nil

	case tea.KeyDown, 'j':
		if m.focused == 5 { // List
			m.list.HandleKey(&components.KeyEvent{Key: 'j'})
		}
		return m, nil

	// Enter to select
	case tea.KeyEnter:
		if m.focused == 5 { // List
			selectedIdx := m.list.GetSelectedIdx()
			if selectedIdx >= 0 {
				items := m.list.GetItems()
				if selectedIdx < len(items) {
					// Handle selection
					fmt.Printf("Selected: %s\n", items[selectedIdx].Label)
				}
			}
		}
		return m, nil

	// Simulate progress updates
	case '1':
		m.progress1.WithPercent((m.progress1.GetPercent() + 10) % 101)
	case '2':
		m.progress2.WithPercent((m.progress2.GetPercent() + 10) % 101)
	case '3':
		m.progress3.WithPercent((m.progress3.GetPercent() + 10) % 101

	default:
		// Forward key to focused component
		switch m.focused {
		case 0: // Progress 1
			// Progress components don't handle keyboard
		case 1: // Progress 2
		case 2: // Progress 3
		case 3: // Table
			// Table component handles arrow keys
		case 4: // List
			m.list.HandleKey(&components.KeyEvent{Key: int(keyMsg.Type)})
		case 5: // Status bar
			// Static component
		}
		return m, nil
	}

	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	var builder strings.Builder

	// Calculate component sizes
	width := m.width
	height := m.height

	// Header (3 lines)
	builder.WriteString(m.header.View())
	builder.WriteString("\n")

	// Progress bars row
	progressRow := ""
	if m.focused == 0 {
		progressRow += "> "
	} else {
		progressRow += "  "
	}
	progressRow += m.progress1.View() + " "
	if m.focused == 1 {
		progressRow += "> "
	} else {
		progressRow += "  "
	}
	progressRow += m.progress2.View() + " "
	if m.focused == 2 {
		progressRow += "> "
	} else {
		progressRow += "  "
	}
	progressRow += m.progress3.View()
	builder.WriteString(progressRow)
	builder.WriteString("\n\n")

	// Table
	if m.focused == 3 {
		builder.WriteString("> Table View\n")
	} else {
		builder.WriteString("  Table View\n")
	}
	builder.WriteString(m.table.View())
	builder.WriteString("\n\n")

	// List
	if m.focused == 4 {
		builder.WriteString("> System Logs\n")
	} else {
		builder.WriteString("  System Logs\n")
	}
	builder.WriteString(m.list.View())
	builder.WriteString("\n")

	// Status bar
	builder.WriteString(m.statusBar.View())

	return builder.String()
}

// GetFocusInfo returns focus information for the UI
func (m *DashboardModel) GetFocusInfo() string {
	focusedNames := []string{
		"CPU Progress",
		"Memory Progress",
		"Disk Progress",
		"Data Table",
		"Log List",
		"Status Bar",
	}
	current := focusedNames[m.focused]
	return fmt.Sprintf("Focused: %s | Press Tab to cycle", current)
}

func main() {
	p := tea.NewProgram(
		NewDashboardModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
