package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTableProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected TableProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: TableProps{
				ShowBorder: true,
				Focused:    false,
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"key":   "name",
						"title": "Name",
						"width": 20,
					},
					map[string]interface{}{
						"key":   "age",
						"title": "Age",
						"width": 10,
					},
				},
				"data": []interface{}{
					[]interface{}{"John", 30},
					[]interface{}{"Jane", 25},
				},
				"focused":    true,
				"width":      80,
				"height":     20,
				"showBorder": false,
			},
			expected: TableProps{
				Columns: []Column{
					{Key: "name", Title: "Name", Width: 20},
					{Key: "age", Title: "Age", Width: 10},
				},
				Data: [][]interface{}{
					{"John", 30},
					{"Jane", 25},
				},
				Focused:    true,
				Width:      80,
				Height:     20,
				ShowBorder: false,
			},
		},
		{
			name: "object array data",
			props: map[string]interface{}{
				"columns": []interface{}{
					map[string]interface{}{
						"key":   "id",
						"title": "ID",
						"width": 10,
					},
					map[string]interface{}{
						"key":   "name",
						"title": "Name",
						"width": 20,
					},
					map[string]interface{}{
						"key":   "role",
						"title": "Role",
						"width": 15,
					},
				},
				"data": []interface{}{
					map[string]interface{}{
						"id":   1,
						"name": "Alice Johnson",
						"role": "Admin",
					},
					map[string]interface{}{
						"id":   2,
						"name": "Bob Smith",
						"role": "User",
					},
				},
				"focused":    false,
				"showBorder": true,
			},
			expected: TableProps{
				Columns: []Column{
					{Key: "id", Title: "ID", Width: 10},
					{Key: "name", Title: "Name", Width: 20},
					{Key: "role", Title: "Role", Width: 15},
				},
				Data: [][]interface{}{
					{1, "Alice Johnson", "Admin"},
					{2, "Bob Smith", "User"},
				},
				Focused:    false,
				ShowBorder: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTableProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTable(t *testing.T) {
	props := TableProps{
		Columns: []Column{
			{Title: "Name", Width: 20},
			{Title: "Age", Width: 10},
		},
		Data: [][]interface{}{
			{"John", 30},
			{"Jane", 25},
		},
		Width: 80,
	}

	result := RenderTable(props, 80)
	assert.NotEmpty(t, result)
}