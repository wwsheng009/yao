package components

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

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

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// TableModel wraps the table.Model to handle TUI integration
type TableModel struct {
	table.Model
	props               TableProps
	data                [][]interface{} // Store the original data
	id                  string          // Unique identifier for this instance
	previousSelectedRow int             // Track previous selection for change detection
	ID                  string          // For component interface
}

// RenderTable renders a table component
func RenderTable(props TableProps, width int) string {
	// Validate input: ensure we have columns
	if len(props.Columns) == 0 {
		return ""
	}

	// Build table configuration using helper functions
	columns := buildTableColumns(props.Columns)
	rows := buildTableRows(props.Data, props.Columns)

	// Create table model
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(props.Focused),
	)

	// Apply styles
	applyTableStyles(&t, props)

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

// ============================================================================
// Helper Functions - Extract common logic for code reuse
// ============================================================================

// buildTableColumns builds table columns from column definitions
func buildTableColumns(columns []Column) []table.Column {
	result := make([]table.Column, len(columns))
	for i, col := range columns {
		colWidth := col.Width
		if colWidth <= 0 {
			colWidth = 10 // Default width
		}
		result[i] = table.Column{
			Title: col.Title,
			Width: colWidth,
		}
	}
	return result
}

// buildTableRows builds table rows from data and column definitions
func buildTableRows(data [][]interface{}, columns []Column) []table.Row {
	rows := make([]table.Row, 0, len(data))
	for _, rowData := range data {
		// Skip rows that don't match column count
		if len(rowData) != len(columns) {
			continue
		}
		row := make([]string, len(rowData))
		for j, cell := range rowData {
			// Apply column-specific style if defined, otherwise use default formatting
			if j < len(columns) && columns[j].Style.GetStyle().String() != lipgloss.NewStyle().String() {
				row[j] = columns[j].Style.GetStyle().Render(formatCell(cell))
			} else {
				row[j] = formatCell(cell)
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// buildTableStyles builds table styles from props
func buildTableStyles(props TableProps) (headerStyle, cellStyle, selectedStyle lipgloss.Style) {
	headerStyle = props.HeaderStyle.GetStyle()
	cellStyle = props.CellStyle.GetStyle()
	selectedStyle = props.SelectedStyle.GetStyle()

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

	return
}

// applyTableStyles applies styles to a table model
func applyTableStyles(t *table.Model, props TableProps) {
	headerStyle, cellStyle, selectedStyle := buildTableStyles(props)

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
}

// applyTableDimensions applies width and height to a table model
func applyTableDimensions(t *table.Model, width, height int) {
	if width > 0 {
		t.SetWidth(width)
	}
	if height > 0 {
		t.SetHeight(height)
	}
}

// ============================================================================
// Update Functions - For updating existing table models without recreation
// ============================================================================

// updateTableModelData updates table data without recreating the model
func updateTableModelData(model *table.Model, data [][]interface{}, columns []Column) {
	if data == nil {
		return
	}

	// Save current cursor position to restore after update
	oldCursor := model.Cursor()

	// Build new rows
	rows := buildTableRows(data, columns)

	// Update model with new rows
	model.SetRows(rows)

	// Restore cursor position if still valid
	if oldCursor >= 0 && oldCursor < len(rows) {
		// model.SetCursor(oldCursor) - Note: bubbles/table doesn't have SetCursor
		// The cursor position is automatically adjusted by the library
	}
}

// updateTableModelStyles updates table styles without recreating the model
func updateTableModelStyles(model *table.Model, props TableProps) {
	applyTableStyles(model, props)
}

// updateTableModelDimensions updates table dimensions without recreating the model
func updateTableModelDimensions(model *table.Model, props TableProps) {
	if props.Width > 0 {
		model.SetWidth(props.Width)
	}
	if props.Height > 0 {
		model.SetHeight(props.Height)
	}
}

// updateTableModelFocus updates table focus state without recreating the model
func updateTableModelFocus(model *table.Model, focused bool) {
	if focused {
		model.Focus()
	} else {
		model.Blur()
	}
}

// ============================================================================
// Core Table Creation Functions
// ============================================================================

// createNativeTableModel creates a new native table.Model from TableProps
// This function should ONLY be called during initialization, not during updates
func createNativeTableModel(props TableProps) table.Model {
	// Validate input: ensure we have columns
	if len(props.Columns) == 0 {
		return table.New()
	}

	// Build table configuration using helper functions
	columns := buildTableColumns(props.Columns)
	rows := buildTableRows(props.Data, props.Columns)

	// Create table model
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(props.Focused),
	)

	// Apply styles
	applyTableStyles(&t, props)

	// Set size if specified
	applyTableDimensions(&t, props.Width, props.Height)

	return t
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
// This function now uses createNativeTableModel to avoid code duplication
func NewTableModel(props TableProps, id string) TableModel {
	// Validate input: ensure we have columns
	if len(props.Columns) == 0 {
		return TableModel{props: props, id: id}
	}

	// Create native table model
	t := createNativeTableModel(props)

	return TableModel{
		Model: t,
		props: props,
		data:  props.Data,
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

// TableStateHelper 表格组件状态捕获助手
type TableStateHelper struct {
	Indexer     interface{ Index() int }
	Selector    interface{ SelectedItem() interface{} }
	Focuser     interface{ Focused() bool }
	ComponentID string
}

func (h *TableStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"index":    h.Indexer.Index(),
		"selected": h.Selector.SelectedItem(),
		"focused":  h.Focuser.Focused(),
	}
}

func (h *TableStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	// 检测索引变化
	if old["index"] != new["index"] {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventRowSelected, map[string]interface{}{
			"oldIndex": old["index"],
			"newIndex": new["index"],
		}))
	}

	// 检测选中项变化
	if !compareRows(old["selected"], new["selected"]) {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, "TABLE_ITEM_SELECTED", map[string]interface{}{
			"oldSelected": old["selected"],
			"newSelected": new["selected"],
		}))
	}

	// 检测焦点变化
	if old["focused"] != new["focused"] {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventFocusChanged, map[string]interface{}{
			"focused": new["focused"],
		}))
	}

	return cmds
}

// compareRows compares two potentially nil row values
func compareRows(a, b interface{}) bool {
	// If both are nil, they are equal
	if a == nil && b == nil {
		return true
	}
	// If one is nil and the other isn't, they are not equal
	if a == nil || b == nil {
		return false
	}
	// Try to cast both to []string for comparison
	rowA, okA := a.([]string)
	rowB, okB := b.([]string)
	if okA && okB {
		// Compare lengths first
		if len(rowA) != len(rowB) {
			return false
		}
		// Compare each element
		for i := range rowA {
			if rowA[i] != rowB[i] {
				return false
			}
		}
		return true
	}

	// For other types, use reflection to check if they are comparable
	// and handle comparison safely
	return safeCompare(a, b)
}

// safeCompare compares two values safely, avoiding panic on uncomparable types
func safeCompare(a, b interface{}) bool {
	// If types are different, they cannot be equal
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	// Use reflection to check if the types are comparable
	// If not, we consider them unequal rather than panicking
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != valB.Kind() {
		return false
	}

	// If either value is not valid, they are not equal
	if !valA.IsValid() || !valB.IsValid() {
		return false
	}

	// Check if the type is comparable using reflection
	if valA.Type().Comparable() {
		return a == b
	}

	// For uncomparable types, return false instead of panicking
	return false
}

// TableComponentWrapper wraps the native table.Model to implement ComponentInterface properly
type TableComponentWrapper struct {
	model       table.Model
	props       TableProps
	id          string
	bindings    []core.ComponentBinding
	stateHelper *TableStateHelper
}

// NewTableComponentWrapper creates a wrapper that implements ComponentInterface
func NewTableComponentWrapper(props TableProps, id string) *TableComponentWrapper {
	// 创建原生 table.Model
	nativeModel := createNativeTableModel(props)

	// 初始化 wrapper，直接使用原生模型
	wrapper := &TableComponentWrapper{
		model:    nativeModel,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	// 初始化 stateHelper，使用 wrapper 自身
	wrapper.stateHelper = &TableStateHelper{
		Indexer:     wrapper, // wrapper 自己实现 Index() 方法
		Selector:    wrapper, // wrapper 自己实现 SelectedItem() 方法
		Focuser:     wrapper, // wrapper 自己实现 Focused() 方法
		ComponentID: id,
	}

	return wrapper
}

func (w *TableComponentWrapper) Init() tea.Cmd {
	return nil
}

// Index returns the current cursor position
func (w *TableComponentWrapper) Index() int {
	return w.model.Cursor()
}

// SelectedItem returns the currently selected item
func (w *TableComponentWrapper) SelectedItem() interface{} {
	cursor := w.model.Cursor()
	rows := w.model.Rows()
	if cursor >= 0 && cursor < len(rows) {
		return rows[cursor]
	}
	return nil
}

// Focused returns whether the table is focused
func (w *TableComponentWrapper) Focused() bool {
	return w.model.Focused()
}

// SetFocus sets or removes focus from table component
func (w *TableComponentWrapper) SetFocus(focus bool) {
	currentFocus := w.model.Focused()
	if focus != currentFocus {
		if focus {
			w.model.Focus()
		} else {
			w.model.Blur()
		}
	}
}

func (w *TableComponentWrapper) GetFocus() bool {
	return w.model.Focused()
}

// GetValue returns the current value (for Valuer interface)
func (w *TableComponentWrapper) GetValue() string {
	item := w.SelectedItem()
	if item != nil {
		// Return a string representation of the selected item
		return fmt.Sprintf("%v", item)
	}
	return ""
}

// GetModel returns the underlying model
func (w *TableComponentWrapper) GetModel() interface{} {
	return w.model
}

// GetID returns the component ID
func (w *TableComponentWrapper) GetID() string {
	return w.id
}

// View returns the view of the component
func (w *TableComponentWrapper) View() string {
	return w.model.View()
}

// PublishEvent creates and returns a command to publish an event
func (w *TableComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *TableComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For table component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}

func (w *TableComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,                   // 实现了 InteractiveBehavior 接口的组件
		msg,                 // 接收的消息
		w.getBindings,       // 获取按键绑定的函数
		w.handleBinding,     // 处理按键绑定的函数
		w.delegateToBubbles, // 委托给原 bubbles 组件的函数
	)

	return w, cmd, response
}

// 实现 InteractiveBehavior 接口的方法

func (w *TableComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *TableComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// 直接使用wrapper本身实现 ComponentWrapper 接口
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *TableComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *TableComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *TableComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HasFocus 方法
func (w *TableComponentWrapper) HasFocus() bool {
	return w.model.Focused()
}

// 实现 HandleSpecialKey 方法
func (w *TableComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyTab:
		// 让Tab键冒泡以处理组件导航
		return nil, core.Ignored, true
	case tea.KeyEnter:
		// Handle Enter key for row selection confirmation
		currentSelectedRow := w.model.Cursor()
		if currentSelectedRow >= 0 {
			// Get row data if available
			var rowData interface{}
			rows := w.model.Rows()
			if currentSelectedRow < len(rows) {
				rowData = rows[currentSelectedRow]
			}

			// Publish row double-click / enter pressed event
			eventCmd := core.PublishEvent(
				w.id,
				core.EventRowDoubleClicked,
				map[string]interface{}{
					"rowIndex": currentSelectedRow,
					"rowData":  rowData,
					"tableID":  w.id,
					"trigger":  "enter_key",
				},
			)

			return eventCmd, core.Handled, true
		}
		return nil, core.Handled, true
	}

	// 其他按键不由这个函数处理
	return nil, core.Ignored, false
}

func (w *TableComponentWrapper) GetComponentType() string {
	return "table"
}

// Render renders the table component
// REFACTORED: Now uses update functions instead of recreating the table
func (w *TableComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TableComponentWrapper: invalid data type")
	}
	props := ParseTableProps(propsMap)

	// Update the existing model instead of recreating it
	updateTableModelData(&w.model, props.Data, props.Columns)
	updateTableModelStyles(&w.model, props)
	updateTableModelDimensions(&w.model, props)
	// updateTableModelFocus(&w.model, props.Focused)
	w.props = props

	// Return the view
	return w.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
// REFACTORED: Now uses update functions for cleaner code
func (w *TableComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TableComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse table properties
	props := ParseTableProps(propsMap)

	// Update component properties
	w.props = props

	// Update table data, styles, dimensions, and focus using helper functions
	updateTableModelData(&w.model, props.Data, props.Columns)
	updateTableModelStyles(&w.model, props)
	updateTableModelDimensions(&w.model, props)
	// don't update the focused property
	// updateTableModelFocus(&w.model, props.Focused)

	return nil
}

// Cleanup cleans up resources used by the table component
func (w *TableComponentWrapper) Cleanup() {
	// Table components typically don't need cleanup
	// This is a no-op for table components
}

// GetStateChanges returns the state changes from this component
func (w *TableComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	selectedRow := w.model.Cursor()
	rows := w.model.Rows()

	rowData := interface{}(nil)
	if selectedRow >= 0 && selectedRow < len(rows) {
		rowData = rows[selectedRow]
	}

	return map[string]interface{}{
		w.GetID() + "_selected_row":  selectedRow,
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

// SetFocus sets or removes focus from table component
func (m *TableModel) SetFocus(focus bool) {
	if focus {
		m.Model.Focus()
	} else {
		m.Model.Blur()
	}
}

func (m *TableModel) GetFocus() bool {
	return m.Model.Focused()
}

func (m *TableModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TableModel: invalid data type")
	}

	// Parse table properties
	props := ParseTableProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
// REFACTORED: Now uses update functions for cleaner code
func (m *TableModel) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TableModel: invalid data type for UpdateRenderConfig")
	}

	// Parse table properties
	props := ParseTableProps(propsMap)

	// Update component properties
	m.props = props

	// Update table data, styles, dimensions, and focus using helper functions
	updateTableModelData(&m.Model, props.Data, props.Columns)
	updateTableModelStyles(&m.Model, props)
	updateTableModelDimensions(&m.Model, props)
	// updateTableModelFocus(&m.Model, props.Focused)
	m.data = props.Data

	return nil
}

// Cleanup 清理资源
func (m *TableModel) Cleanup() {
	// TableModel 通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (m *TableModel) GetStateChanges() (map[string]interface{}, bool) {
	selectedRow := m.Model.Cursor()
	rows := m.Model.Rows()

	rowData := interface{}(nil)
	if selectedRow >= 0 && selectedRow < len(rows) {
		rowData = rows[selectedRow]
	}

	return map[string]interface{}{
		m.GetID() + "_selected_row":  selectedRow,
		m.GetID() + "_selected_data": rowData,
	}, len(rows) > 0 && selectedRow >= 0
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (m *TableModel) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}

// Index returns the current cursor position
func (m *TableModel) Index() int {
	return m.Model.Cursor()
}

// SelectedItem returns the currently selected item
func (m *TableModel) SelectedItem() interface{} {
	cursor := m.Model.Cursor()
	rows := m.Model.Rows()
	if cursor >= 0 && cursor < len(rows) {
		return rows[cursor]
	}
	return nil
}

// GetSelected returns the currently selected item and whether anything is selected
func (m *TableModel) GetSelected() (interface{}, bool) {
	item := m.SelectedItem()
	return item, item != nil
}

// Focused returns whether the table is focused
func (m *TableModel) Focused() bool {
	return m.Model.Focused()
}
