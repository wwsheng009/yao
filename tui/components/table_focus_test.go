package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

// TestTableModel_FocusAndNavigation 测试表格焦点和导航
func TestTableModel_FocusAndNavigation(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
			{Key: "age",  Title: "Age",  Width: 10},
		},
		Data: [][]interface{}{
			{"Alice", 25},
			{"Bob",   30},
			{"Charlie", 35},
		},
		Focused:    true, // 启用焦点
		ShowBorder: true,
	}

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 初始状态：应该有焦点
	if !tableModel.Model.Focused() {
		t.Error("Table should be focused initially")
	}

	// 初始光标应该在第一行 (index 0)
	if cursor := tableModel.Model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0, got %d", cursor)
	}

	// 测试向下导航
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Handled {
		t.Error("Down key should be handled when focused")
	}

	if cursor := tableModel.Model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor at row 1 after Down key, got %d", cursor)
	}

	// 测试向上导航
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	if response != core.Handled {
		t.Error("Up key should be handled when focused")
	}

	if cursor := tableModel.Model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0 after Up key, got %d", cursor)
	}
}

// TestTableModel_FocusLost_IgnoresKeys 测试失去焦点时忽略键盘事件
func TestTableModel_FocusLost_IgnoresKeys(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
			{Key: "age",  Title: "Age",  Width: 10},
		},
		Data: [][]interface{}{
			{"Alice", 25},
			{"Bob",   30},
			{"Charlie", 35},
		},
		Focused:    false, // 不启用焦点
		ShowBorder: true,
	}

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 初始状态：不应该有焦点
	if tableModel.Model.Focused() {
		t.Error("Table should not be focused initially")
	}

	// 测试向下导航（应该被忽略）
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Ignored {
		t.Errorf("Down key should be ignored when not focused, got %v", response)
	}

	// 光标应该保持不变
	if cursor := tableModel.Model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor to remain at row 0 when unfocused, got %d", cursor)
	}
}

// TestTableModel_SetFocus_Dynamic 动态切换焦点
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

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 初始状态：无焦点
	if tableModel.Model.Focused() {
		t.Error("Table should not be focused initially")
	}

	// 键盘事件应该被忽略
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Ignored {
		t.Error("Down key should be ignored when not focused")
	}

	// 设置焦点
	wrapper.SetFocus(true)

	if !tableModel.Model.Focused() {
		t.Error("Table should be focused after SetFocus(true)")
	}

	// 现在键盘事件应该被处理
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Handled {
		t.Error("Down key should be handled when focused")
	}

	// 光标应该移动
	if cursor := tableModel.Model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor to move to row 1 when focused, got %d", cursor)
	}

	// 释放焦点
	wrapper.SetFocus(false)

	if tableModel.Model.Focused() {
		t.Error("Table should not be focused after SetFocus(false)")
	}

	// 键盘事件应该再次被忽略
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Ignored {
		t.Error("Down key should be ignored when not focused again")
	}
}

// TestTableModel_SelectionEvents 测试选择事件发布
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

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 订阅事件
	var receivedEvents []core.ActionMsg
	eventCh := make(chan core.ActionMsg, 10)

	// 模拟事件订阅
	go func() {
		for {
			select {
			case e := <-eventCh:
				receivedEvents = append(receivedEvents, e)
			}
		}
	}()

	// 注意：由于测试环境中事件系统可能不可用，这里主要验证逻辑
	// 实际应用中，事件会被发布到 tea.Cmd 并由主程序处理

	// 按下 Down 键应该触发 EventRowSelected
	_, cmd, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	if response != core.Handled {
		t.Error("Down key should be handled")
	}

	// 应该有一个命令
	if cmd == nil {
		t.Error("Update should return a command for event publishing")
	}

	// 光标应该移动
	if cursor := tableModel.Model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor at row 1, got %d", cursor)
	}
}

// TestTableModel_EnterKey 测试 Enter 键处理
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

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 移动到第二行
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})

	// 按下 Enter 键应该触发 EventRowDoubleClicked
	_, cmd, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyEnter})
	if response != core.Handled {
		t.Error("Enter key should be handled")
	}

	// 应该有一个命令
	if cmd == nil {
		t.Error("Enter key should return a command for event publishing")
	}

	// 光标应该保持在第二行
	if cursor := tableModel.Model.Cursor(); cursor != 1 {
		t.Errorf("Expected cursor to remain at row 1, got %d", cursor)
	}
}

// TestTableModel_Pagination 测试翻页导航
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

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 初始光标在第一行
	if cursor := tableModel.Model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0, got %d", cursor)
	}

	// 按 PgDown 应该翻页
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyPgDown})
	if response != core.Handled {
		t.Error("PgDown key should be handled")
	}

	// 光标应该向下移动（具体行数取决于表格实现）
	newCursor := tableModel.Model.Cursor()
	if newCursor <= 0 {
		t.Errorf("Expected cursor to move after PgDown, got %d", newCursor)
	}

	// 按 PgUp 应该向上翻页
	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyPgUp})
	if response != core.Handled {
		t.Error("PgUp key should be handled")
	}

	// 光标应该向上移动
	finalCursor := tableModel.Model.Cursor()
	if finalCursor >= newCursor {
		t.Errorf("Expected cursor to move up after PgUp, got %d (from %d)", finalCursor, newCursor)
	}
}

// TestTableModel_EmptyTable 测试空表格的处理
func TestTableModel_EmptyTable(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Key: "name", Title: "Name", Width: 20},
		},
		Data:       [][]interface{}{}, // 空数据
		Focused:    true,
		ShowBorder: true,
	}

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 光标应该在 0 或 -1（取决于表格实现）
	cursor := tableModel.Model.Cursor()
	if cursor < -1 {
		t.Errorf("Invalid cursor position: %d", cursor)
	}

	// 按下键应该不会导致崩溃
	_, _, response := wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	_ = response

	_, _, response = wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	_ = response
}

// TestTableModel_SingleRowTable 测试单行表格
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

	tableModel := NewTableModel(props, "test-table")
	wrapper := NewTableComponentWrapper(&tableModel)

	// 初始光标在第一行
	if cursor := tableModel.Model.Cursor(); cursor != 0 {
		t.Errorf("Expected cursor at row 0, got %d", cursor)
	}

	// 按 Down 键应该保持或循环（取决于表格实现）
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyDown})
	cursor := tableModel.Model.Cursor()
	if cursor < 0 {
		t.Errorf("Cursor should not be negative: %d", cursor)
	}

	// 按 Up 键应该保持或循环
	wrapper.UpdateMsg(tea.KeyMsg{Type: tea.KeyUp})
	cursor = tableModel.Model.Cursor()
	if cursor < 0 {
		t.Errorf("Cursor should not be negative: %d", cursor)
	}
}
