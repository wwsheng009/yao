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
	progress1 *components.ProgressComponent
	progress2 *components.ProgressComponent
	progress3 *components.ProgressComponent
	table     *components.TableComponent
	list      *components.ListComponent
	statusBar *components.FooterComponent
	header    *components.HeaderComponent

	// State
	focused int
	err     error
	width   int
	height  int
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
		WithColumns([]components.RuntimeColumn{
			{Key: "name", Title: "Name", Width: 30},
			{Key: "status", Title: "Status", Width: 15},
			{Key: "cpu", Title: "CPU", Width: 10},
			{Key: "memory", Title: "Memory", Width: 10},
		}).
		WithData([][]interface{}{
			{"Server 1", "Running", "45%", "60%"},
			{"Server 2", "Stopped", "0%", "0%"},
			{"Server 3", "Running", "72%", "81%"},
			{"Database", "Running", "30%", "45%"},
		}).
		WithWidth(80).
		WithHeight(10).
		WithShowBorder(true)

	// Create list with log entries
	list := components.NewList().
		WithID("log-list").
		WithTitle("System Logs").
		WithItems([]components.RuntimeListItem{
			components.NewRuntimeListItem("[INFO] Server started on port 8080", ""),
			components.NewRuntimeListItem("[WARN] High memory usage detected", ""),
			components.NewRuntimeListItem("[INFO] Backup completed successfully", ""),
			components.NewRuntimeListItem("[ERROR] Failed to connect to database", ""),
			components.NewRuntimeListItem("[INFO] User authentication successful", ""),
		}).
		WithWidth(80).
		WithHeight(8)

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
		table:     table,
		list:      list,
		statusBar: statusBar,
		header:    header,
		focused:   0,
		width:     80,
		height:    24,
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

	// Simulate progress updates
	case '1':
		newPercent := int(m.progress1.GetPercent()+10) % 101
		m.progress1.WithPercent(float64(newPercent))
	case '2':
		newPercent := int(m.progress2.GetPercent()+10) % 101
		m.progress2.WithPercent(float64(newPercent))
	case '3':
		newPercent := int(m.progress3.GetPercent()+10) % 101
		m.progress3.WithPercent(float64(newPercent))

	// Enter to select from list
	case tea.KeyEnter:
		if m.focused == 5 { // List
			selectedItem := m.list.GetSelectedItem()
			if selectedItem != nil {
				// Handle selection
				fmt.Printf("Selected: %s\n", selectedItem.Title())
			}
		}
		return m, nil
	}

	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	var builder strings.Builder

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
