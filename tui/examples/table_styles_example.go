package main

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/ui/components"
)

// This example demonstrates how to configure Table with styles
// similar to the bubbles/table native API

func Example1_UsingTableMethods() {
	// Create a table using the new fluent API
	columns := []components.RuntimeColumn{
		{Key: "name", Title: "Name", Width: 30},
		{Key: "status", Title: "Status", Width: 15},
		{Key: "cpu", Title: "CPU", Width: 10},
	}

	rows := [][]interface{}{
		{"Server 1", "Running", "45%"},
		{"Server 2", "Stopped", "0%"},
		{"Server 3", "Running", "72%"},
	}

	tbl := components.NewTable().
		WithID("servers").
		WithColumns(columns).
		WithData(rows).
		WithFocused(true).
		WithHeight(7)

	// Apply styles using the new methods
	// Method 1: Using individual style methods
	tbl.WithHeaderColor("240").
		WithHeaderBold(false).
		WithSelectedColor("229").
		WithSelectedBackground("57").
		WithSelectedBold(false).
		WithBorderType(lipgloss.NormalBorder())

	_ = tbl
}

func Example2_UsingStandardBorder() {
	// Using the convenience method for standard borders
	columns := []components.RuntimeColumn{
		{Key: "id", Title: "ID", Width: 10},
		{Key: "name", Title: "Name", Width: 30},
	}

	rows := [][]interface{}{
		{1, "Item 1"},
		{2, "Item 2"},
	}

	tbl := components.NewTable().
		WithColumns(columns).
		WithData(rows).
		WithHeight(7).
		WithStandardBorder("240") // Quick way to set normal border with color

	_ = tbl
}

func Example3_DSLConfiguration() {
	// Example DSL JSON configuration
	// This would be in a .tui.yao file
	dslConfig := map[string]interface{}{
		"columns": []map[string]interface{}{
			{"key": "commit", "title": "提交哈希", "width": 10},
			{"key": "docType", "title": "文档类型", "width": 40},
			{"key": "fileCount", "title": "文件数", "width": 10},
		},
		"showBorder": true,
		"focused":    true,
		// Style configurations - these are automatically applied
		"headerColor":        "240", // Can use ANSI codes directly
		"headerBackground":   "",
		"headerBold":         false,
		"cellColor":          "15",
		"selectedColor":      "229",
		"selectedBackground": "57",
		"selectedBold":       false,
		"borderColor":        "240",
		"borderStyle":        "normal", // New: can specify border type
		"borderBottom":       true,
	}

	// When the DSL parser processes this config, it will:
	// 1. Convert color names to ANSI codes (e.g., "primary" -> "21")
	// 2. Apply styles using WithXXX methods
	// 3. Set border type using WithBorderType

	_ = dslConfig
}

func Example4_ComparisonWithBubblesTable() {
	// Original bubbles/table API:
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 15},
	}
	rows := []table.Row{
		{"Server 1", "Running"},
		{"Server 2", "Stopped"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

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

	_ = t

	// Equivalent using our TableComponent:
	columns2 := []components.RuntimeColumn{
		{Key: "name", Title: "Name", Width: 30},
		{Key: "status", Title: "Status", Width: 15},
	}
	rows2 := [][]interface{}{
		{"Server 1", "Running"},
		{"Server 2", "Stopped"},
	}

	tbl := components.NewTable().
		WithColumns(columns2).
		WithData(rows2).
		WithFocused(true).
		WithHeight(7).
		WithBorderType(lipgloss.NormalBorder()).
		WithBorderColor("240").
		WithHeaderBold(false).
		WithSelectedColor("229").
		WithSelectedBackground("57").
		WithSelectedBold(false)

	_ = tbl
}

func Example5_AdvancedBorderStyles() {
	// Demonstrate different border types
	borderTypes := []struct {
		name   string
		border lipgloss.Border
	}{
		{"Normal", lipgloss.NormalBorder()},
		{"Rounded", lipgloss.RoundedBorder()},
		{"Thick", lipgloss.ThickBorder()},
		{"Double", lipgloss.DoubleBorder()},
	}

	for _, bt := range borderTypes {
		tbl := components.NewTable().
			WithID(bt.name).
			WithBorderType(bt.border). // Apply different border types
			WithBorderColor("240").
			WithHeight(5)

		_ = tbl
	}
}

func Example6_ColorFormats() {
	// The DSL and API support multiple color formats:

	colorFormats := []string{
		"240",              // ANSI code (direct number)
		"primary",          // Semantic color name
		"#FF5733",          // Hex color
		"rgb(255, 87, 51)", // RGB format
		"red",              // Named color
		"brightRed",        // Bright variant
		"muted",            // Semantic (maps to "245")
	}

	for _, color := range colorFormats {
		tbl := components.NewTable().
			WithBorderColor(color).
			WithSelectedColor(color)

		_ = tbl
	}
}

func main() {
	// Run examples
	Example1_UsingTableMethods()
	Example2_UsingStandardBorder()
	Example3_DSLConfiguration()
	Example4_ComparisonWithBubblesTable()
	Example5_AdvancedBorderStyles()
	Example6_ColorFormats()

	// Simple program to demonstrate
	p := tea.NewProgram(nil)
	p.Quit()
}
