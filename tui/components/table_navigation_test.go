package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

func TestTableComponentWrapper_UpdateMsg_KeyDown(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
			{Key: "email", Title: "Email", Width: 30},
		},
		Data: [][]interface{}{
			{1, "Alice", "alice@example.com"},
			{2, "Bob", "bob@example.com"},
			{3, "Charlie", "charlie@example.com"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      60,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Test initial selection
	initialCursor := wrapper.model.Cursor()
	if initialCursor != 0 {
		t.Errorf("Expected initial cursor at 0, got %d", initialCursor)
	}

	// Test Down key
	msg := tea.KeyMsg{Type: tea.KeyDown}
	comp, cmd, response := wrapper.UpdateMsg(msg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
	if cmd == nil {
		t.Error("Expected command to be returned (event command)")
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}

	// Check cursor moved down
	newCursor := wrapper.model.Cursor()
	if newCursor != 1 {
		t.Errorf("Expected cursor at 1 after Down key, got %d", newCursor)
	}
}

func TestTableComponentWrapper_UpdateMsg_NavigationKeys(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
			{2, "Bob"},
			{3, "Charlie"},
			{4, "David"},
			{5, "Eve"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	testCases := []struct {
		name         string
		msg          tea.KeyMsg
		expectedRow  int
	}{
		{"Down key", tea.KeyMsg{Type: tea.KeyDown}, 1},
		{"Down key again", tea.KeyMsg{Type: tea.KeyDown}, 2},
		{"Up key", tea.KeyMsg{Type: tea.KeyUp}, 1},
		{"Down key", tea.KeyMsg{Type: tea.KeyDown}, 2},
		{"Page Down", tea.KeyMsg{Type: tea.KeyPgDown}, 5},
		{"Page Up", tea.KeyMsg{Type: tea.KeyPgUp}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, response := wrapper.UpdateMsg(tc.msg)
			if response != core.Handled {
				t.Errorf("Expected Handled response, got %v", response)
			}

			cursor := wrapper.model.Cursor()
			if cursor != tc.expectedRow {
				t.Errorf("Expected cursor at %d, got %d", tc.expectedRow, cursor)
			}
		})
	}
}

func TestTableComponentWrapper_UpdateMsg_EnterKey(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
			{2, "Bob"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Move to second row
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter on second row
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	comp, cmd, response := wrapper.UpdateMsg(msg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}

	// Verify cursor is still at row 1
	cursor := wrapper.model.Cursor()
	if cursor != 1 {
		t.Errorf("Expected cursor to remain at 1, got %d", cursor)
	}
}

func TestTableComponentWrapper_UpdateMsg_TargetedMsg(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
		},
		ShowBorder: true,
		Focused:    true,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Test targeted message to this component
	innerMsg := tea.KeyMsg{Type: tea.KeyDown}
	targetedMsg := core.TargetedMsg{
		TargetID: "test-table",
		InnerMsg: innerMsg,
	}
	comp, _, response := wrapper.UpdateMsg(targetedMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}

	// Test targeted message to different component
	targetedMsg = core.TargetedMsg{
		TargetID: "other-table",
		InnerMsg: innerMsg,
	}
	comp, _, response = wrapper.UpdateMsg(targetedMsg)
	if response != core.Ignored {
		t.Errorf("Expected Ignored response, got %v", response)
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}
}

func TestTableComponentWrapper_UpdateMsg_RowSelectionEvent(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
			{2, "Bob"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Move down to trigger selection event
	_, cmd, _ := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})

	if cmd == nil {
		t.Error("Expected command to be returned for selection event")
	}
}

func TestNewTableModel_DefaultStyles(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	tableModel := NewTableModel(props, "test-table")

	// Verify the table was created successfully
	if tableModel.id != "test-table" {
		t.Errorf("Expected id 'test-table', got '%s'", tableModel.id)
	}

	// Verify the view is not empty
	view := tableModel.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestNewTableModel_SelectedRowVisibility(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
			{2, "Bob"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	tableModel := NewTableModel(props, "test-table")

	// Get initial view
	initialView := tableModel.View()

	// Move selection to second row
	tableModel.Model.SetCursor(1)
	selectedView := tableModel.View()

	// Views should be different due to selected row highlighting
	if initialView == selectedView {
		t.Error("Expected views to be different when selection changes")
	}

	// Both should be non-empty
	if initialView == "" {
		t.Error("Expected initial view to be non-empty")
	}
	if selectedView == "" {
		t.Error("Expected selected view to be non-empty")
	}
}

func TestNewTableModel_CustomSelectedStyle(t *testing.T) {
	// Create a custom selected style
	customStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("196"))

	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
		},
		ShowBorder:   true,
		Focused:      true,
		Height:       10,
		Width:        30,
		SelectedStyle: lipglossStyleWrapper{Style: &customStyle},
	}
	tableModel := NewTableModel(props, "test-table")

	// Verify the table was created successfully
	view := tableModel.View()
	if view == "" {
		t.Error("Expected non-empty view with custom style")
	}
}

func TestTableComponentWrapper_SetFocus(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
		},
		Data: [][]interface{}{
			{1, "Alice"},
		},
		ShowBorder: true,
		Focused:    false,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Set focus
	wrapper.SetFocus(true)
	if !wrapper.model.Focused() {
		t.Error("Expected table to be focused")
	}

	// Remove focus
	wrapper.SetFocus(false)
	if wrapper.model.Focused() {
		t.Error("Expected table to be blurred")
	}
}

func TestTableComponentWrapper_GetID(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
		},
		Data: [][]interface{}{
			{1, "Alice"},
		},
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-id-789")

	if wrapper.GetID() != "test-id-789" {
		t.Errorf("Expected id 'test-id-789', got '%s'", wrapper.GetID())
	}
}

func TestTableComponentWrapper_View(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{1, "Alice"},
			{2, "Bob"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	view := wrapper.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestTableComponentWrapper_Init(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
		},
		Data: [][]interface{}{
			{1, "Alice"},
		},
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	cmd := wrapper.Init()
	if cmd != nil {
		t.Error("Expected nil command from Init")
	}
}

func TestTableComponentWrapper_BoundaryNavigation(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "id", Title: "ID", Width: 5},
		},
		Data: [][]interface{}{
			{1, "Alice"},
			{2, "Bob"},
		},
		ShowBorder: true,
		Focused:    true,
		Height:     10,
		Width:      30,
	}
	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Move to last row
	wrapper.model.SetCursor(1)

	// Try to move down (should stay at last row)
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	cursor := wrapper.model.Cursor()
	if cursor != 1 {
		t.Errorf("Expected cursor to stay at 1, got %d", cursor)
	}

	// Try to move up (should stay at first row)
	wrapper.model.SetCursor(0)
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	cursor = wrapper.model.Cursor()
	if cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", cursor)
	}
}