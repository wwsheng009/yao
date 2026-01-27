package component

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/tui/core"
)

// TestTableModel_FocusAndNavigation tests table focus and navigation
func TestTableModel_FocusAndNavigation(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
			{Key: "age", Title: "Age", Width: 10},
		},
		Data: [][]interface{}{
			{"Alice", 25},
			{"Bob", 30},
			{"Charlie", 35},
		},
		Focused:    true, // Enable focus
		ShowBorder: true,
	}

	// Use the new API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Initial state: should have focus
	if !wrapper.model.Focused() {
		t.Error("Table should be focused initially")
	}

	// Initial cursor should be on the first row (index 0)
	if cursor := wrapper.model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0, got %d", cursor)
	}

	// Test down navigation
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Handled {
		t.Error("Down key should be handled when focused")
	}

	if cursor := wrapper.model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor at row 1 after Down key, got %d", cursor)
	}

	// Test up navigation
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	if response != core.Handled {
		t.Error("Up key should be handled when focused")
	}

	if cursor := wrapper.model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0 after Up key, got %d", cursor)
	}
}

// TestTableModel_FocusLost_IgnoresKeys tests that keyboard events are ignored when table loses focus
func TestTableModel_FocusLost_IgnoresKeys(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
			{Key: "age", Title: "Age", Width: 10},
		},
		Data: [][]interface{}{
			{"Alice", 25},
			{"Bob", 30},
			{"Charlie", 35},
		},
		Focused:    false, // Disable focus
		ShowBorder: true,
	}

	// Use the new API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Initial state: should not have focus
	if wrapper.model.Focused() {
		t.Error("Table should not be focused initially")
	}

	// Test down navigation (should be ignored)
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Ignored {
		t.Errorf("Down key should be ignored when not focused, got %v", response)
	}

	// Cursor should remain unchanged
	if cursor := wrapper.model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor to remain at row 0 when unfocused, got %d", cursor)
	}
}

// TestTableModel_SetFocus_Dynamic tests dynamic focus switching
func TestTableModel_SetFocus_Dynamic(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{"Alice"},
			{"Bob"},
		},
		Focused:    false,
		ShowBorder: true,
	}

	// Use the new API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Initial state: no focus
	if wrapper.model.Focused() {
		t.Error("Table should not be focused initially")
	}

	// Keyboard events should be ignored
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Ignored {
		t.Error("Down key should be ignored when not focused")
	}

	// Set focus
	wrapper.SetFocus(true)

	if !wrapper.model.Focused() {
		t.Error("Table should be focused after SetFocus(true)")
	}

	// Now keyboard events should be handled
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Handled {
		t.Error("Down key should be handled when focused")
	}

	// Cursor should move
	if cursor := wrapper.model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor to move to row 1 when focused, got %d", cursor)
	}

	// Release focus
	wrapper.SetFocus(false)

	if wrapper.model.Focused() {
		t.Error("Table should not be focused after SetFocus(false)")
	}

	// Keyboard events should be ignored again
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Ignored {
		t.Error("Down key should be ignored when not focused again")
	}
}

// TestTableModel_SelectionEvents tests selection event publishing
func TestTableModel_SelectionEvents(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{"Alice"},
			{"Bob"},
		},
		Focused:    true,
		ShowBorder: true,
	}

	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Subscribe to events
	var receivedEvents []core.ActionMsg
	eventCh := make(chan core.ActionMsg, 10)

	// Simulate event subscription
	go func() {
		for {
			select {
			case e := <-eventCh:
				receivedEvents = append(receivedEvents, e)
			}
		}
	}()

	// Note: Event systems may not be available in test environment, mainly verifying logic here
	// In actual application, events will be published to tea.Cmd and handled by the main program

	// Pressing Down key should trigger EventRowSelected
	_, cmd, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Handled {
		t.Error("Down key should be handled")
	}

	// Should have a command
	if cmd == nil {
		t.Error("Update should return a command for event publishing")
	}

	// Cursor should move
	if cursor := wrapper.model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor at row 1, got %d", cursor)
	}
}

// TestTableModel_EnterKey tests Enter key handling
func TestTableModel_EnterKey(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{"Alice"},
			{"Bob"},
		},
		Focused:    true,
		ShowBorder: true,
	}

	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")
	// Move to second row
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})

	// Pressing Enter key should trigger EventRowDoubleClicked
	_, cmd, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyEnter})
	if response != core.Handled {
		t.Error("Enter key should be handled")
	}

	// Should have a command
	if cmd == nil {
		t.Error("Enter key should return a command for event publishing")
	}

	// Cursor should remain on second row
	if cursor := wrapper.model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor to remain at row 1, got %d", cursor)
	}
}

// TestTableModel_Pagination tests pagination navigation
func TestTableModel_Pagination(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{"Row1"}, {"Row2"}, {"Row3"}, {"Row4"},
			{"Row5"}, {"Row6"}, {"Row7"}, {"Row8"},
			{"Row9"}, {"Row10"}, {"Row11"},
		},
		Focused:    true,
		Height:     5, // 限制高度
		ShowBorder: true,
	}

	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Initial cursor on first row
	if cursor := wrapper.model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0, got %d", cursor)
	}

	// Pressing PgDown should page down
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyPgDown})
	if response != core.Handled {
		t.Error("PgDown key should be handled")
	}

	// Cursor should move down (exact rows depend on table implementation)
	newCursor := wrapper.model.Cursor()
	if newCursor <= 0 {
		t.Errorf("Expected cursor to move after PgDown, got %d", newCursor)
	}

	// Pressing PgUp should page up
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyPgUp})
	if response != core.Handled {
		t.Error("PgUp key should be handled")
	}

	// Cursor should move up
	finalCursor := wrapper.model.Cursor()
	if finalCursor >= newCursor {
		t.Errorf("Expected cursor to move up after PgUp, got %d (from %d)", finalCursor, newCursor)
	}
}

// TestTableModel_EmptyTable tests empty table handling
func TestTableModel_EmptyTable(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data:       [][]interface{}{}, // 空数据
		Focused:    true,
		ShowBorder: true,
	}

	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Cursor should be at 0 or -1 (depends on table implementation)
	cursor := wrapper.model.Cursor()
	if cursor < -1 {
		t.Errorf("Invalid cursor position: %d", cursor)
	}

	// Pressing keys should not cause panic
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	_ = response

	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	_ = response
}

// TestTableModel_SingleRowTable tests single-row table
func TestTableModel_SingleRowTable(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data: [][]interface{}{
			{"Only row"},
		},
		Focused:    true,
		ShowBorder: true,
	}

	// 使用新的API
	wrapper := NewTableComponentWrapper(props, "test-table")

	// Initial cursor on first row
	if cursor := wrapper.model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0, got %d", cursor)
	}

	// Pressing Down key should stay or cycle (depends on table implementation)
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	cursor := wrapper.model.Cursor()
	if cursor < 0 {
		t.Errorf("Cursor should not be negative: %d", cursor)
	}

	// 按 Up 键应该保持或循环
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	cursor = wrapper.model.Cursor()
	if cursor < 0 {
		t.Errorf("Cursor should not be negative: %d", cursor)
	}
}
