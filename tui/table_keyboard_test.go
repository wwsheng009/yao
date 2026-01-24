package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/ui/components"
)

// TestTableKeyboardInput tests that tables respond to keyboard input
func TestTableKeyboardInput(t *testing.T) {
	config := &Config{
		Name: "Table Keyboard Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "table",
					ID:   "test-table",
					Bind: "users", // Bind to the users data
					Props: map[string]interface{}{
						"columns": []interface{}{
							map[string]interface{}{"key": "id", "title": "ID", "width": 10},
							map[string]interface{}{"key": "name", "title": "Name", "width": 20},
						},
						"height": 5,
					},
				},
			},
		},
		Data: map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "John Doe"},
				map[string]interface{}{"id": 2, "name": "Jane Smith"},
				map[string]interface{}{"id": 3, "name": "Bob Johnson"},
			},
		},
	}

	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	model.Init()
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Trigger rendering to populate focus list
	_ = model.View()

	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	t.Logf("Focus list: %v", model.runtimeFocusList)
	t.Logf("m.Components has %d items", len(model.Components))

	// Check if the table is in m.Components
	for id, comp := range model.Components {
		t.Logf("m.Components[%s]: Type=%s, Instance=%T", id, comp.Type, comp.Instance)
	}

	// Press Tab to set focus on first table
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*Model)

	// Execute the focus command
	if cmd != nil {
		msg := cmd()
		newModel, _ := model.Update(msg)
		model = newModel.(*Model)
	}

	t.Logf("After Tab, CurrentFocus: %s", model.CurrentFocus)

	// Now press down key - the table should handle it
	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(*Model)

	// Execute any commands from the table
	if cmd != nil {
		msg := cmd()
		newModel, newCmd := model.Update(msg)
		model = newModel.(*Model)
		if newCmd != nil {
			newMsg := newCmd()
			newModel, _ := model.Update(newMsg)
			model = newModel.(*Model)
		}
	}

	t.Logf("After Down, CurrentFocus: %s", model.CurrentFocus)

	// Test multiple down keys
	for i := 0; i < 3; i++ {
		newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = newModel.(*Model)
		t.Logf("After down %d, CurrentFocus: %s", i+1, model.CurrentFocus)
	}

	// Check the table component state and cursor position
	var initialCursor, finalCursor int
	if model.RuntimeRoot != nil {
		for _, child := range model.RuntimeRoot.Children {
			if child.ID == "test-table" && child.Component != nil && child.Component.Instance != nil {
				// In runtime mode, the component is wrapped
				if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
					if table, ok := wrapper.Component.(*components.TableComponent); ok {
						t.Logf("Table state: focused=%v, cursor=%d, rows=%d", table.IsFocused(), table.GetSelectedRow(), len(table.GetRows()))
						initialCursor = table.GetSelectedRow()
					}
				}
			}
		}
	}

	// Now press down key again to check cursor movement
	t.Logf("About to press Down key, CurrentFocus=%s", model.CurrentFocus)

	// Check if we can access the table directly before key press
	var tableBefore *components.TableComponent
	if model.RuntimeRoot != nil {
		for _, child := range model.RuntimeRoot.Children {
			if child.ID == "test-table" && child.Component != nil && child.Component.Instance != nil {
				if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
					if table, ok := wrapper.Component.(*components.TableComponent); ok {
						tableBefore = table
						t.Logf("Before Down: table=%p, cursor=%d", table, table.GetSelectedRow())
					}
				}
			}
		}
	}

	// Find the runtime node before the key press
	if model.RuntimeRoot != nil {
		for _, child := range model.RuntimeRoot.Children {
			if child.ID == "test-table" {
				t.Logf("Found runtime node, Component=%v, Instance type=%T",
					child.Component != nil,
					child.Component.Instance)
			}
		}
	}

	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(*Model)
	t.Logf("After Down key, cmd=%v", cmd != nil)

	if cmd != nil {
		msg := cmd()
		t.Logf("Command returned msg type: %T", msg)
		newModel, newCmd := model.Update(msg)
		model = newModel.(*Model)
		if newCmd != nil {
			newMsg := newCmd()
			t.Logf("Nested command returned msg type: %T", newMsg)
			newModel2, _ := model.Update(newMsg)
			model = newModel2.(*Model)
		}
	}

	// Check if the component instance is the same after the update
	if model.RuntimeRoot != nil {
		for _, child := range model.RuntimeRoot.Children {
			if child.ID == "test-table" {
				t.Logf("After update, runtime node Instance type=%T",
					child.Component.Instance)
				if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
					if table, ok := wrapper.Component.(*components.TableComponent); ok {
						t.Logf("After Down: table=%p (same? %v), cursor=%d", table, table == tableBefore, table.GetSelectedRow())
					}
				}
			}
		}
	}

	// Check final cursor position
	if model.RuntimeRoot != nil {
		for _, child := range model.RuntimeRoot.Children {
			if child.ID == "test-table" && child.Component != nil && child.Component.Instance != nil {
				if wrapper, ok := child.Component.Instance.(*NativeComponentWrapper); ok {
					if table, ok := wrapper.Component.(*components.TableComponent); ok {
						finalCursor = table.GetSelectedRow()
						t.Logf("After additional Down, cursor=%d (was %d)", finalCursor, initialCursor)
						// Cursor should have moved (if there are rows to move to)
						if len(table.GetRows()) > 1 {
							if finalCursor <= initialCursor && initialCursor < len(table.GetRows())-1 {
								t.Errorf("Cursor should have moved down, was %d, now %d", initialCursor, finalCursor)
							}
						}
					}
				}
			}
		}
	}

	t.Log("Table keyboard input test passed!")
}
