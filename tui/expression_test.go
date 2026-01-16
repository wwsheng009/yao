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
			"username-input": "john_doe",
			"email-input": "john@example.com",
			"password-input": "secret123",
			"special.key": "dot_value",
			"special key": "space_value",
			"special@key": "at_value",
			"special#key": "hash_value",
			"special$key": "dollar_value",
			"special%key": "percent_value",
			"special&key": "ampersand_value",
			"special*key": "asterisk_value",
			"special!key": "exclamation_value",
			"special?key": "question_value",
			"special/key": "slash_value",
			"special\\key": "backslash_value",
			"nested": map[string]interface{}{
				"deep-key": "deep_value",
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

	t.Run("hyphenated key with index function", func(t *testing.T) {
		result := model.applyState("Username: {{index($, \"username-input\")}}")
		assert.Equal(t, "Username: john_doe", result)
		
		result = model.applyState("Email: {{index($, \"email-input\")}}")
		assert.Equal(t, "Email: john@example.com", result)
		
		result = model.applyState("Password: {{index($, \"password-input\")}}")
		assert.Equal(t, "Password: secret123", result)
	})

	t.Run("non-existent hyphenated key", func(t *testing.T) {
		result := model.applyState("NonExistent: {{index($, \"non-existent-key\")}}")
		assert.Equal(t, "NonExistent: ", result)
	})

	t.Run("special character keys with index function", func(t *testing.T) {
		result := model.applyState("Special dot: {{index($, \"special.key\")}}")
		assert.Equal(t, "Special dot: dot_value", result)
		
		result = model.applyState("Special space: {{index($, \"special key\")}}")
		assert.Equal(t, "Special space: space_value", result)
		
		result = model.applyState("Special at: {{index($, \"special@key\")}}")
		assert.Equal(t, "Special at: at_value", result)
		
		result = model.applyState("Special hash: {{index($, \"special#key\")}}")
		assert.Equal(t, "Special hash: hash_value", result)
		
		result = model.applyState("Special dollar: {{index($, \"special$key\")}}")
		assert.Equal(t, "Special dollar: dollar_value", result)
		
		result = model.applyState("Special percent: {{index($, \"special%key\")}}")
		assert.Equal(t, "Special percent: percent_value", result)
		
		result = model.applyState("Special ampersand: {{index($, \"special&key\")}}")
		assert.Equal(t, "Special ampersand: ampersand_value", result)
		
		result = model.applyState("Special asterisk: {{index($, \"special*key\")}}")
		assert.Equal(t, "Special asterisk: asterisk_value", result)
		
		result = model.applyState("Special exclamation: {{index($, \"special!key\")}}")
		assert.Equal(t, "Special exclamation: exclamation_value", result)
		
		result = model.applyState("Special question: {{index($, \"special?key\")}}")
		assert.Equal(t, "Special question: question_value", result)
		
		result = model.applyState("Special slash: {{index($, \"special/key\")}}")
		assert.Equal(t, "Special slash: slash_value", result)
		
		result = model.applyState("Special backslash: {{index($, \"special\\\\key\")}}")
		assert.Equal(t, "Special backslash: backslash_value", result)
	})

	t.Run("nested object with special keys", func(t *testing.T) {
		result := model.applyState("Nested deep: {{index($, \"nested\")}}")
		// Just check that it doesn't crash
		assert.NotEqual(t, "Nested deep: ", result)
		
		// Access nested object with special key
		result = model.applyState("Deep key: {{index(index($, \"nested\"), \"deep-key\")}}")
		assert.Equal(t, "Deep key: deep_value", result)
	})

	t.Run("mixed expressions with special keys", func(t *testing.T) {
		result := model.applyState("Mixed: {{title}} - {{index($, \"special.key\")}} - {{count}}")
		assert.Equal(t, "Mixed: Dynamic Title - dot_value - 5", result)
	})

	t.Run("complex nested expressions", func(t *testing.T) {
		result := model.applyState("Complex: {{index($, \"username-input\")}} - {{len(items)}} - {{index($, \"special@key\")}}")
		assert.Equal(t, "Complex: john_doe - 3 - at_value", result)
	})
}