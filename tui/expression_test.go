package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyStateExpressions(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Dynamic Title",
			"count": 5,
			"user": map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
			"items": []interface{}{"apple", "banana", "cherry"},
			"stats": map[string]interface{}{
				"total":     10,
				"active":    7,
				"completed": 3,
			},
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	t.Run("simple key substitution", func(t *testing.T) {
		result := model.applyState("Welcome: {{title}}")
		assert.Equal(t, "Welcome: Dynamic Title", result)
	})

	t.Run("numeric substitution", func(t *testing.T) {
		result := model.applyState("Count: {{count}}")
		assert.Equal(t, "Count: 5", result)
	})

	t.Run("nested key substitution", func(t *testing.T) {
		result := model.applyState("User: {{user.name}}, Age: {{user.age}}")
		assert.Equal(t, "User: Alice, Age: 30", result)
	})

	t.Run("length function", func(t *testing.T) {
		result := model.applyState("Items count: {{len(items)}}")
		assert.Equal(t, "Items count: 3", result)
	})

	t.Run("index function", func(t *testing.T) {
		result := model.applyState("Total: {{index(stats, \"total\")}}")
		assert.Equal(t, "Total: 10", result)
		
		result = model.applyState("Active: {{index(stats, \"active\")}}")
		assert.Equal(t, "Active: 7", result)
		
		result = model.applyState("Completed: {{index(stats, \"completed\")}}")
		assert.Equal(t, "Completed: 3", result)
	})

	t.Run("non-existent key", func(t *testing.T) {
		result := model.applyState("Missing: {{missing}}")
		assert.Equal(t, "Missing: ", result)
	})

	t.Run("complex expression", func(t *testing.T) {
		result := model.applyState("Summary: {{title}} has {{len(items)}} items, total: {{index(stats, \"total\")}}")
		assert.Equal(t, "Summary: Dynamic Title has 3 items, total: 10", result)
	})
}