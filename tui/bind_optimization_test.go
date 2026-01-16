package tui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/kun/any"
)

func TestBindOptimization(t *testing.T) {
	// Test data flattening functionality
	t.Run("test data flattening", func(t *testing.T) {
		// Complex nested data structure
		nestedData := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "张三",
				"age":  30,
				"address": map[string]interface{}{
					"city":    "北京",
					"street":  "长安街",
					"details": map[string]interface{}{"zip": "100000"},
				},
			},
			"orders": []interface{}{
				map[string]interface{}{"id": 1, "amount": 100.50},
				map[string]interface{}{"id": 2, "amount": 200.75},
			},
		}

		// Use kun/any to flatten the data
		wrappedRes := any.Of(nestedData)
		flattened := wrappedRes.Map().MapStrAny.Dot()

		// Verify flattened structure
		assert.Equal(t, "张三", flattened["user.name"])
		assert.Equal(t, 30, flattened["user.age"])
		assert.Equal(t, "北京", flattened["user.address.city"])
		assert.Equal(t, "长安街", flattened["user.address.street"])
		assert.Equal(t, "100000", flattened["user.address.details.zip"])
		assert.Equal(t, 1, flattened["orders[0].id"])
		assert.Equal(t, 100.50, flattened["orders[0].amount"])
		assert.Equal(t, 2, flattened["orders[1].id"])
		assert.Equal(t, 200.75, flattened["orders[1].amount"])
	})

	// Test applyState with helper.Bind and expression combination
	t.Run("test applyState with helper.Bind and expression", func(t *testing.T) {
		cfg := &Config{
			Name: "Test",
			Data: map[string]interface{}{
				"title": "Hello World",
				"count": 42,
				"user": map[string]interface{}{
					"name": "Alice",
					"age":  30,
				},
				"items": []interface{}{"apple", "banana", "cherry"},
			},
			Layout: Layout{
				Direction: "vertical",
			},
		}

		model := NewModel(cfg, nil)

		// Test simple binding
		result := model.applyState("Title: {{title}}")
		assert.Equal(t, "Title: Hello World", result)

		// Test numeric substitution
		result = model.applyState("Count: {{count}}")
		assert.Equal(t, "Count: 42", result)

		// Test nested property access
		result = model.applyState("User: {{user.name}}, Age: {{user.age}}")
		assert.Equal(t, "User: Alice, Age: 30", result)

		// Test len function
		result = model.applyState("Items count: {{len(items)}}")
		assert.Equal(t, "Items count: 3", result)

		// Test complex expression
		result = model.applyState("Summary: {{title}} has {{len(items)}} items, count: {{count}}")
		assert.Equal(t, "Summary: Hello World has 3 items, count: 42", result)
	})

	// Test config loading with flattened data
	t.Run("test config loading with flattened data", func(t *testing.T) {
		// Simulate loading a config with nested data
		rawConfig := `{
			"name": "Test TUI",
			"data": {
				"user": {
					"name": "Bob",
					"profile": {
						"email": "bob@example.com",
						"settings": {
							"theme": "dark",
							"notifications": true
						}
					}
				},
				"stats": {
					"visits": 100,
					"pages": ["/home", "/about", "/contact"]
				}
			},
			"layout": {
				"direction": "vertical"
			}
		}`

		// Simulate unmarshaling the config
		var cfg Config
		err := json.Unmarshal([]byte(rawConfig), &cfg)
		assert.NoError(t, err)

		// Manually apply the flattening logic that happens in loadFile
		if cfg.Data != nil {
			wrappedRes := any.Of(cfg.Data)
			flattened := wrappedRes.Map().MapStrAny.Dot()
			cfg.Data = flattened
		}

		// Verify that the data was flattened correctly
		assert.Equal(t, "Bob", cfg.Data["user.name"])
		assert.Equal(t, "bob@example.com", cfg.Data["user.profile.email"])
		assert.Equal(t, "dark", cfg.Data["user.profile.settings.theme"])
		assert.Equal(t, true, cfg.Data["user.profile.settings.notifications"])
		assert.Equal(t, float64(100), cfg.Data["stats.visits"])
		assert.Equal(t, "/home", cfg.Data["stats.pages[0]"])
		assert.Equal(t, "/about", cfg.Data["stats.pages[1]"])
		assert.Equal(t, "/contact", cfg.Data["stats.pages[2]"])

		// Create model and verify state initialization
		model := NewModel(&cfg, nil)
		
		// Check that flattened data is in state
		value, exists := model.getStateValue("user.name")
		assert.True(t, exists)
		assert.Equal(t, "Bob", value)

		value, exists = model.getStateValue("user.profile.email")
		assert.True(t, exists)
		assert.Equal(t, "bob@example.com", value)

		value, exists = model.getStateValue("user.profile.settings.theme")
		assert.True(t, exists)
		assert.Equal(t, "dark", value)
	})
}