package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column defines a table column
type Column struct {
	// Key is the data key for this column
	Key string `json:"key"`
	
	// Title is the display title for this column
	Title string `json:"title"`
	
	// Width is the column width
	Width int `json:"width"`
	
	// Style is the column style
	Style lipgloss.Style `json:"style"`
}



// TableProps defines the properties for the Table component
type TableProps struct {
	// Columns defines the table columns
	Columns []Column `json:"columns"`
	
	// Data contains the table rows
	Data [][]interface{} `json:"data"`
	
	// Focused determines if the table is focused (for selection)
	Focused bool `json:"focused"`
	
	// Height specifies the table height (0 for auto)
	Height int `json:"height"`
	
	// Width specifies the table width (0 for auto)
	Width int `json:"width"`
	
	// ShowBorder determines if borders are shown
	ShowBorder bool `json:"showBorder"`
	
	// BorderStyle is the style for table borders
	BorderStyle lipglossStyleWrapper `json:"borderStyle"`
	
	// HeaderStyle is the style for header cells
	HeaderStyle lipglossStyleWrapper `json:"headerStyle"`
	
	// CellStyle is the style for regular cells
	CellStyle lipglossStyleWrapper `json:"cellStyle"`
	
	// SelectedStyle is the style for selected cells
	SelectedStyle lipglossStyleWrapper `json:"selectedStyle"`
}

// TableModel wraps the table.Model to handle TUI integration
type TableModel struct {
	table.Model
	props TableProps
}

// RenderTable renders a table component
func RenderTable(props TableProps, width int) string {
	// Validate input: ensure we have columns
	if len(props.Columns) == 0 {
		return ""
	}

	// Prepare columns
	columns := make([]table.Column, len(props.Columns))
	for i, col := range props.Columns {
		colWidth := col.Width
		if colWidth <= 0 {
			// Calculate reasonable default width
			colWidth = 10
		}
		columns[i] = table.Column{
			Title: col.Title,
			Width: colWidth,
		}
	}

	// Prepare rows with validation and column-specific styling
	rows := make([]table.Row, 0, len(props.Data))
	for _, rowData := range props.Data {
		// Skip rows that don't match column count
		if len(rowData) != len(props.Columns) {
			continue
		}
		row := make([]string, len(rowData))
		for j, cell := range rowData {
			// Apply column-specific style if defined, otherwise use default formatting
			// Check if style is non-empty by comparing string representation
			if j < len(props.Columns) && props.Columns[j].Style.String() != lipgloss.NewStyle().String() {
				row[j] = props.Columns[j].Style.Render(formatCell(cell))
			} else {
				row[j] = formatCell(cell)
			}
		}
		rows = append(rows, row)
	}

	// Create table model
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(props.Focused),
	)

	// Apply styles
	headerStyle := props.HeaderStyle.GetStyle()
	cellStyle := props.CellStyle.GetStyle()
	selectedStyle := props.SelectedStyle.GetStyle()
	
	if props.ShowBorder {
		t.SetStyles(table.Styles{
			Header:   headerStyle,
			Cell:     cellStyle,
			Selected: selectedStyle,
		})
	} else {
		s := table.DefaultStyles()
		s.Header = headerStyle
		s.Cell = cellStyle
		s.Selected = selectedStyle
		t.SetStyles(s)
	}

	// Set size if specified
	if props.Width > 0 {
		t.SetWidth(props.Width)
	} else if width > 0 {
		t.SetWidth(width)
	}

	if props.Height > 0 {
		t.SetHeight(props.Height)
	}

	return t.View()
}

// ParseTableProps converts a generic props map to TableProps using JSON unmarshaling
func ParseTableProps(props map[string]interface{}) TableProps {
	// Set defaults
	tp := TableProps{
		ShowBorder: true,
		Focused:    false,
	}

	// Unmarshal properties first to get Columns
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &tp)
	}

	// Handle Data separately as it needs special processing
	dataValue := props["data"]

	// Helper function to process data array
	processDataArray := func(dataArray []interface{}) {
		tp.Data = make([][]interface{}, 0, len(dataArray))
		for _, rowIntf := range dataArray {
			// Check if data is already a slice ([][]interface{})
			if rowSlice, ok := rowIntf.([]interface{}); ok {
				tp.Data = append(tp.Data, rowSlice)
				continue
			}

			// Check if data is a map ([]map[string]interface{})
			// Convert object array to array based on column keys
			if rowMap, ok := rowIntf.(map[string]interface{}); ok && len(tp.Columns) > 0 {
				row := make([]interface{}, len(tp.Columns))
				for i, col := range tp.Columns {
					if col.Key != "" {
						row[i] = rowMap[col.Key]
					}
				}
				tp.Data = append(tp.Data, row)
			}
		}
	}

	// Check if data is empty or nil
	if dataValue == nil {
		return tp
	}

	// Case 1: data is already an array ([]interface{})
	if dataArray, ok := dataValue.([]interface{}); ok {
		processDataArray(dataArray)
		return tp
	}

	// Case 2: data is a map ({"users": [...]} type)
	// Extract the first array value from the map
	if dataMap, ok := dataValue.(map[string]interface{}); ok {
		for _, v := range dataMap {
			if dataArray, ok := v.([]interface{}); ok {
				processDataArray(dataArray)
				return tp
			}
		}
	}

	// Case 3: data is a string (template variable like "{{users}}" that was converted to string)
	// This happens when the expr engine converts non-simple types to string
	if dataStr, ok := dataValue.(string); ok {
		// Try to unmarshal as JSON array first
		var dataArray []interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataArray); err == nil {
			processDataArray(dataArray)
			return tp
		}

		// If JSON unmarshal fails, the data might be the string representation of a map
		// Check if we have __bind_data which contains the original data
		if bindData, ok := props["__bind_data"]; ok {
			if bindDataArray, ok := bindData.([]interface{}); ok {
				processDataArray(bindDataArray)
				return tp
			}
			// If __bind_data is a map, extract the first array value
			if bindDataMap, ok := bindData.(map[string]interface{}); ok {
				for _, v := range bindDataMap {
					if dataArray, ok := v.([]interface{}); ok {
						processDataArray(dataArray)
						return tp
					}
				}
			}
		}
	}

	return tp
}

// formatCell formats a cell value for display
func formatCell(cell interface{}) string {
	return fmt.Sprintf("%v", cell)
}

// HandleTableUpdate handles updates for table components
// This is used when the table is interactive (selection, scrolling, etc.)
func HandleTableUpdate(msg tea.Msg, tableModel *TableModel) (TableModel, tea.Cmd) {
	if tableModel == nil {
		return TableModel{}, nil
	}

	var cmd tea.Cmd
	tableModel.Model, cmd = tableModel.Model.Update(msg)
	return *tableModel, cmd
}