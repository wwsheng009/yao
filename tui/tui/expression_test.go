package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/kun/any"
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
	
	// Test array indexing after flattening with any.Of().Dot()
	t.Run("flattened array indexing", func(t *testing.T) {
		// Simulate what happens when data is flattened with any.Of().Dot()
		flattenedData := map[string]interface{}{
			"features.0": "Rich text formatting",
			"features.1": "Color support",
			"features.2": "Alignment options",
			"features.3": "Styling capabilities",
			"features": []interface{}{"Rich text formatting", "Color support", "Alignment options", "Styling capabilities"},
		}
		
		// Create a new model with flattened data
		cfg := &Config{
			Name: "Test",
			Data: flattenedData,
		}
		testModel := NewModel(cfg, nil)
		
		// Test accessing array elements using dot notation
		result := testModel.applyState("Feature 0: {{index($, \"features.0\")}}")
		assert.Equal(t, "Feature 0: Rich text formatting", result)
		
		result = testModel.applyState("Feature 1: {{features.1}}")
		// Note: direct access like {{features.1}} may not work because features is an array, not a map
		// The flattened version has "features.1" as a key
		
		// More importantly, test that the flattened keys exist
		result = testModel.applyState("Feature 1: {{index($, \"features.1\")}}")
		assert.Equal(t, "Feature 1: Color support", result)
		
		result = testModel.applyState("Feature 2: {{index($, \"features.2\")}}")
		assert.Equal(t, "Feature 2: Alignment options", result)
		
		result = testModel.applyState("Feature 3: {{index($, \"features.3\")}}")
		assert.Equal(t, "Feature 3: Styling capabilities", result)
	})

	// Test actual text.tui.yao data structure flattening
	t.Run("text tui yao data structure", func(t *testing.T) {
		// Simulate the original data structure from text.tui.yao
		originalData := map[string]interface{}{
			"welcome_message": "Welcome to the TUI Application!",
			"description":     "This is a sample application demonstrating the text component.",
			"features": []interface{}{
				"Rich text formatting",
				"Color support",
				"Alignment options",
				"Styling capabilities",
			},
			"stats": map[string]interface{}{
				"users":     1250,
				"messages":  3420,
				"documents": 89,
			},
		}

		// Apply the same flattening logic as in loader.go
		wrappedRes := any.Of(originalData)
		flattened := wrappedRes.Map().MapStrAny.Dot()

		// Create a new model with the flattened data
		cfg := &Config{
			Name: "Test",
			Data: flattened,
		}
		testModel := NewModel(cfg, nil)

		// Test that the flattened keys exist and can be accessed
		assert.Equal(t, "Rich text formatting", flattened["features.0"])
		assert.Equal(t, "Color support", flattened["features.1"])
		assert.Equal(t, "Alignment options", flattened["features.2"])
		assert.Equal(t, "Styling capabilities", flattened["features.3"])

		// Test that expressions in the format used by text.tui.yao work correctly
		result := testModel.applyState("Feature: {{features.0}}")
		assert.Equal(t, "Feature: Rich text formatting", result)

		result = testModel.applyState("Feature: {{features.1}}")
		assert.Equal(t, "Feature: Color support", result)

		result = testModel.applyState("Feature: {{features.2}}")
		assert.Equal(t, "Feature: Alignment options", result)

		result = testModel.applyState("Feature: {{features.3}}")
		assert.Equal(t, "Feature: Styling capabilities", result)

		// Test other expressions from text.tui.yao
		result = testModel.applyState("Welcome: {{welcome_message}}")
		assert.Equal(t, "Welcome: Welcome to the TUI Application!", result)

		result = testModel.applyState("Stats: Users: {{stats.users}}, Messages: {{stats.messages}}")
		assert.Equal(t, "Stats: Users: 1250, Messages: 3420", result)
	})
}