package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
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
	Style lipglossStyleWrapper `json:"style"`
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
	props               TableProps
	data                [][]interface{} // Store the original data
	id                  string        // Unique identifier for this instance
	previousSelectedRow int           // Track previous selection for change detection
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

	// Prepare rows with validation (no column-specific styling to avoid ANSI conflicts)
	rows := make([]table.Row, 0, len(props.Data))
	for _, rowData := range props.Data {
		// Skip rows that don't match column count
		if len(rowData) != len(props.Columns) {
			continue
		}
		row := make([]string, len(rowData))
		for j, cell := range rowData {
			// Apply column-specific style if defined, otherwise use default formatting
			if j < len(props.Columns) && props.Columns[j].Style.GetStyle().String() != lipgloss.NewStyle().String() {
				row[j] = props.Columns[j].Style.GetStyle().Render(formatCell(cell))
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

	// Set default styles if not provided for better visibility
	if headerStyle.String() == lipgloss.NewStyle().String() {
		headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")) // Light orange
	}
	if cellStyle.String() == lipgloss.NewStyle().String() {
		cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Dark gray
	}
	if selectedStyle.String() == lipgloss.NewStyle().String() {
		// High-contrast selected style for better visibility
		selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("231")). // Black
			Background(lipgloss.Color("39")).  // Light blue background
			Underline(true)
	}

	if props.ShowBorder {
		t.SetStyles(table.Styles{
			Header: headerStyle,
			// Cell:     cellStyle,
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

// NewTableModel creates a new TableModel from TableProps
func NewTableModel(props TableProps, id string) TableModel {
	// Validate input: ensure we have columns
	if len(props.Columns) == 0 {
		return TableModel{props: props, id: id}
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
			if j < len(props.Columns) && props.Columns[j].Style.GetStyle().String() != lipgloss.NewStyle().String() {
				row[j] = props.Columns[j].Style.GetStyle().Render(formatCell(cell))
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

	// Set default styles if not provided for better visibility
	if headerStyle.String() == lipgloss.NewStyle().String() {
		headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")) // Light orange
	}
	if cellStyle.String() == lipgloss.NewStyle().String() {
		cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Dark gray
	}
	if selectedStyle.String() == lipgloss.NewStyle().String() {
		// High-contrast selected style for better visibility
		selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("231")). // Black
			Background(lipgloss.Color("39")).  // Light blue background
			Underline(true)
	}

	if props.ShowBorder {
		t.SetStyles(table.Styles{
			Header: headerStyle,
			// Cell:     cellStyle,
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
	}
	if props.Height > 0 {
		t.SetHeight(props.Height)
	}

	return TableModel{
		Model: t,
		props: props,
		data:  props.Data, // Initialize with provided data
		id:    id,
	}
}

// Init initializes the table model
func (m *TableModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the table
func (m *TableModel) View() string {
	return m.Model.View()
}

// GetID returns the unique identifier for this component instance
func (m *TableModel) GetID() string {
	return m.id
}

// GetComponentType returns the component type
func (m *TableModel) GetComponentType() string {
	return "table"
}

// UpdateMsg implements ComponentInterface for table component
func (m *TableModel) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// If table is not focused, ignore keyboard events but allow pass-through
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If not focused, ignore keyboard navigation events
		if !m.Model.Focused() {
			return m, nil, core.Ignored
		}

		// Track current selection before update
		prevSelectedRow := m.Model.Cursor()

		// Update using the underlying model
		var cmd tea.Cmd
		m.Model, cmd = m.Model.Update(msg)

		// Check if selection changed after navigation
		currentSelectedRow := m.Model.Cursor()

		// Publish event for any selection change (including Down key)
		if currentSelectedRow != prevSelectedRow && currentSelectedRow >= 0 {
			// Get row data if available
			var rowData interface{}
			rows := m.Model.Rows()
			if currentSelectedRow < len(rows) {
				rowData = rows[currentSelectedRow]
			}

			// Publish row selected event
			eventCmd := core.PublishEvent(
				m.id,
				core.EventRowSelected,
				map[string]interface{}{
					"rowIndex":      currentSelectedRow,
					"prevRowIndex":  prevSelectedRow,
					"rowData":       rowData,
					"tableID":       m.id,
					"navigationKey": msg.String(),
				},
			)

			// Combine commands if we have an existing cmd
			if cmd != nil {
				cmd = tea.Batch(cmd, eventCmd)
			} else {
				cmd = eventCmd
			}
		}

		return m, cmd, core.Handled
	}

	// For non-key messages, just update the model
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd, core.Handled
}

// TableComponentWrapper wraps TableModel to implement ComponentInterface properly
type TableComponentWrapper struct {
	model *TableModel
}

// NewTableComponentWrapper creates a wrapper that implements ComponentInterface
func NewTableComponentWrapper(tableModel *TableModel) *TableComponentWrapper {
	return &TableComponentWrapper{
		model: tableModel,
	}
}

func (w *TableComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *TableComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case tea.KeyMsg:
		// If not focused, ignore keyboard navigation events
		if !w.model.Model.Focused() {
			return w, nil, core.Ignored
		}

		// Track current selection before update
		prevSelectedRow := w.model.Model.Cursor()

		// Handle navigation keys explicitly
		switch msg.Type {
		case tea.KeyDown, tea.KeyUp, tea.KeyPgUp, tea.KeyPgDown:
			// Explicitly handle navigation keys to ensure proper event publishing
			var cmd tea.Cmd
			w.model.Model, cmd = w.model.Model.Update(msg)

			// Check if selection changed after navigation
			currentSelectedRow := w.model.Model.Cursor()
			if currentSelectedRow != prevSelectedRow && currentSelectedRow >= 0 {
				// Selection changed, publish event
				// Get row data if available
				var rowData interface{}
				rows := w.model.Model.Rows()
				if currentSelectedRow < len(rows) {
					rowData = rows[currentSelectedRow]
				}

				// Publish event with navigation key info
				eventCmd := core.PublishEvent(
					w.model.id,
					core.EventRowSelected,
					map[string]interface{}{
						"rowIndex":      currentSelectedRow,
						"prevRowIndex":  prevSelectedRow,
						"rowData":       rowData,
						"tableID":       w.model.id,
						"navigationKey": msg.String(),
						"isNavigation":  true,
					},
				)

				// Combine commands if we have an existing cmd
				if cmd != nil {
					cmd = tea.Batch(cmd, eventCmd)
				} else {
					cmd = eventCmd
				}
			}

			return w, cmd, core.Handled

		case tea.KeyEnter:
			// Handle Enter key for row selection confirmation
			currentSelectedRow := w.model.Model.Cursor()
			if currentSelectedRow >= 0 {
				// Get row data if available
				var rowData interface{}
				rows := w.model.Model.Rows()
				if currentSelectedRow < len(rows) {
					rowData = rows[currentSelectedRow]
				}

				// Publish row double-click / enter pressed event
				eventCmd := core.PublishEvent(
					w.model.id,
					core.EventRowDoubleClicked,
					map[string]interface{}{
						"rowIndex": currentSelectedRow,
						"rowData":  rowData,
						"tableID":  w.model.id,
						"trigger":  "enter_key",
					},
				)

				return w, eventCmd, core.Handled
			}
			return w, nil, core.Handled

		default:
			// For other key messages, update normally and check for selection changes
			var cmd tea.Cmd
			w.model.Model, cmd = w.model.Model.Update(msg)

			// Check if selection changed
			currentSelectedRow := w.model.Model.Cursor()
			if currentSelectedRow != prevSelectedRow && currentSelectedRow >= 0 {
				// Selection changed, publish event
				// Get row data if available
				var rowData interface{}
				rows := w.model.Model.Rows()
				if currentSelectedRow < len(rows) {
					rowData = rows[currentSelectedRow]
				}

				eventCmd := core.PublishEvent(
					w.model.id,
					core.EventRowSelected,
					map[string]interface{}{
						"rowIndex": currentSelectedRow,
						"rowData":  rowData,
						"tableID":  w.model.id,
					},
				)

				// Combine commands if we have an existing cmd
				if cmd != nil {
					cmd = tea.Batch(cmd, eventCmd)
				} else {
					cmd = eventCmd
				}
			}

			return w, cmd, core.Handled
		}
	}

	// For non-key messages, update using the underlying model
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)
	return w, cmd, core.Handled
}

func (w *TableComponentWrapper) View() string {
	return w.model.View()
}

func (w *TableComponentWrapper) GetID() string {
	return w.model.id
}

// SetFocus sets or removes focus from table component
func (m *TableModel) SetFocus(focus bool) {
	if focus {
		m.Model.Focus()
	} else {
		m.Model.Blur()
	}
}

func (w *TableComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
}

func (w *TableComponentWrapper) GetComponentType() string {
	return "table"
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (m *TableModel) UpdateRenderConfig(config core.RenderConfig) error {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TableModel: invalid data type")
	}

	// Parse table properties
	props := ParseTableProps(propsMap)

	// Update component properties
	m.props = props

	// Update table data if provided
	if props.Data != nil {
		m.data = props.Data
	}

	return nil
}

// Cleanup 清理资源
func (m *TableModel) Cleanup() {
	// TableModel 通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (m *TableModel) GetStateChanges() (map[string]interface{}, bool) {
	// Table state is managed by the wrapper
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (m *TableModel) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
	}
}

func (m *TableModel) Render(config core.RenderConfig) (string, error) {
	// This method is kept for backward compatibility
	// It now delegates to UpdateRenderConfig
	_ = m.UpdateRenderConfig(config)
	return m.View(), nil
}


func (w *TableComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *TableComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	return w.model.UpdateRenderConfig(config)
}

// Cleanup cleans up resources used by the table component
func (w *TableComponentWrapper) Cleanup() {
	// Table components typically don't need cleanup
	// This is a no-op for table components
}

// GetStateChanges returns the state changes from this component
func (w *TableComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	selectedRow := w.model.Model.Cursor()
	rows := w.model.Model.Rows()

	rowData := interface{}(nil)
	if selectedRow >= 0 && selectedRow < len(rows) {
		rowData = rows[selectedRow]
	}

	return map[string]interface{}{
		w.GetID() + "_selected_row": selectedRow,
		w.GetID() + "_selected_data": rowData,
	}, len(rows) > 0 && selectedRow >= 0
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *TableComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}

