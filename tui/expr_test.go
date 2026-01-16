package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExprEngine(t *testing.T) {
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

	t.Run("simple variable substitution", func(t *testing.T) {
		result := model.applyState("Welcome: {{title}}")
		assert.Equal(t, "Welcome: Dynamic Title", result)
	})

	t.Run("numeric substitution", func(t *testing.T) {
		result := model.applyState("Count: {{count}}")
		assert.Equal(t, "Count: 5", result)
	})

	t.Run("nested property access", func(t *testing.T) {
		result := model.applyState("User: {{user.name}}, Age: {{user.age}}")
		assert.Equal(t, "User: Alice, Age: 30", result)
	})

	t.Run("len function", func(t *testing.T) {
		result := model.applyState("Items count: {{len(items)}}")
		assert.Equal(t, "Items count: 3", result)
		
		result = model.applyState("Title length: {{len(title)}}")
		assert.Equal(t, "Title length: 13", result)
	})

	t.Run("index function", func(t *testing.T) {
		result := model.applyState("Total: {{index(stats, \"total\")}}")
		assert.Equal(t, "Total: 10", result)
	})

	t.Run("ternary operator", func(t *testing.T) {
		result := model.applyState("Status: {{count > 0 ? \"active\" : \"inactive\"}}")
		assert.Equal(t, "Status: active", result)
		
		result = model.applyState("Status: {{count < 0 ? \"active\" : \"inactive\"}}")
		assert.Equal(t, "Status: inactive", result)
	})

	t.Run("complex expressions", func(t *testing.T) {
		result := model.applyState("First item: {{len(items) > 0 ? items[0] : \"none\"}}")
		assert.Equal(t, "First item: apple", result)
	})
}