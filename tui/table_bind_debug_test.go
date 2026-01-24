package tui

import (
	"testing"

	"github.com/yaoapp/yao/tui/runtime/dsl"
	"github.com/yaoapp/yao/tui/ui/components"
)

// TestTableBindDirect tests table component with bind data directly
func TestTableBindDirect(t *testing.T) {
	factory := dsl.NewFactory()

	// Test data as []map[string]interface{}
	testData := []interface{}{
		map[string]interface{}{"id": 1, "name": "John Doe", "email": "john@example.com"},
		map[string]interface{}{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
	}

	// Create table config with columns and __bind_data
	// Columns must be provided in props so they're set before processing bind data
	config := &dsl.ComponentConfig{
		ID:   "test-table",
		Type: "table",
		Props: map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"key": "id", "title": "ID", "width": 10},
				map[string]interface{}{"key": "name", "title": "Name", "width": 20},
				map[string]interface{}{"key": "email", "title": "Email", "width": 30},
			},
			"__bind_data": testData,
		},
	}

	// Create the component
	compIntf, err := factory.Create(config)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	table, ok := compIntf.(*components.TableComponent)
	if !ok {
		t.Fatalf("Expected TableComponent, got %T", compIntf)
	}

	// Check that columns were set
	columns := table.GetColumns()
	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	// Check that data was set
	rows := table.GetRows()
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	// Now test the View method
	view := table.View()
	t.Logf("Table view: %q", view)

	// Check internal state
	t.Logf("Table rows: %v", rows)
	t.Logf("Table columns: %v", columns)

	if !contains(view, "John Doe") {
		t.Error("Expected 'John Doe' in view")
	}
}
