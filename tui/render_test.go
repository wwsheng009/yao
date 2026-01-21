package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyState(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Hello World",
			"count": 42,
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple substitution",
			input:    "Title: {{title}}",
			expected: "Title: Hello World",
		},
		{
			name:     "number substitution",
			input:    "Count: {{count}}",
			expected: "Count: 42",
		},
		{
			name:     "multiple substitutions",
			input:    "{{title}} - Count: {{count}}",
			expected: "Hello World - Count: 42",
		},
		{
			name:     "non-existent key",
			input:    "Value: {{nonexistent}}",
			expected: "Value: ",
		},
		{
			name:     "no substitution",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "empty template",
			input:    "{{}}",
			expected: "{{}}", // 空的 {{}} 不会被替换
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.applyState(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStateValue(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Hello",
			"user": map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{
			name:     "simple key",
			key:      "title",
			expected: "Hello",
		},
		{
			name:     "nested key",
			key:      "user.name",
			expected: "Alice",
		},
		{
			name:     "nested number",
			key:      "user.age",
			expected: 30,
		},
		{
			name:     "non-existent key",
			key:      "missing",
			expected: nil,
		},
		{
			name:     "non-existent nested",
			key:      "user.missing",
			expected: nil,
		},
		{
			name:     "empty key",
			key:      "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := model.getStateValue(tt.key)
			var result interface{}
			if exists {
				result = value
			} else {
				result = nil
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyStateToProps(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Dynamic Title",
			"users": []string{"Alice", "Bob"},
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	t.Run("substitute string props", func(t *testing.T) {
		comp := &Component{
			Type: "header",
			Props: map[string]interface{}{
				"title": "Welcome: {{title}}",
				"count": 10,
			},
		}

		props := model.applyStateToProps(comp)
		assert.Equal(t, "Welcome: Dynamic Title", props["title"])
		assert.Equal(t, 10, props["count"])
	})

	t.Run("bind attribute", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Bind: "users",
			Props: map[string]interface{}{
				"style": "list",
			},
		}

		props := model.applyStateToProps(comp)
		assert.Equal(t, "list", props["style"])
		assert.Equal(t, []string{"Alice", "Bob"}, props["__bind_data"])
	})

	t.Run("nil props", func(t *testing.T) {
		comp := &Component{
			Type: "text",
		}

		props := model.applyStateToProps(comp)
		assert.NotNil(t, props)
		assert.Empty(t, props)
	})
}

func TestRenderHeaderComponent(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80

	// Test using the new renderer system
	comp := &Component{
		Type: "header",
		Props: map[string]interface{}{
			"title": "Test Header",
		},
	}

	result := model.RenderComponent(comp)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Test Header")
}

func TestRenderTextComponent(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80

	t.Run("with content", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Props: map[string]interface{}{
				"content": "Hello World",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Hello World")
	})

	t.Run("with bind data", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Props: map[string]interface{}{
				"__bind_data": "Bound content",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Bound content")
	})

	t.Run("with alignment", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Props: map[string]interface{}{
				"content": "Centered",
				"align":   "center",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
	})
}

func TestRenderComponent(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Dynamic",
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80
	model.Height = 24

	t.Run("header component", func(t *testing.T) {
		comp := &Component{
			Type: "header",
			Props: map[string]interface{}{
				"title": "{{title}}",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Dynamic")
	})

	t.Run("text component", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			ID:   "test-text-component",
			Props: map[string]interface{}{
				"content": "Some text",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Some text")
	})

	t.Run("unknown component", func(t *testing.T) {
		comp := &Component{
			Type: "unknown",
			Props: map[string]interface{}{
				"foo": "bar",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Unknown component")
	})

	t.Run("nil component", func(t *testing.T) {
		result := model.RenderComponent(nil)
		assert.Empty(t, result)
	})

	t.Run("empty type", func(t *testing.T) {
		comp := &Component{
			Props: map[string]interface{}{
				"foo": "bar",
			},
		}

		result := model.RenderComponent(comp)
		assert.Empty(t, result)
	})
}

func TestRenderLayout(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Test App",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					ID:   "test-layout-header",
					Props: map[string]interface{}{
						"title": "{{title}}",
					},
				},
				{
					Type: "text",
					ID:   "test-layout-text",
					Props: map[string]interface{}{
						"content": "Welcome to the app",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80
	model.Height = 24
	model.Ready = true

	// Initialize components
	model.InitializeComponents()

	result := model.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Test App")
	assert.Contains(t, result, "Welcome to the app")
}

func TestApplyPadding(t *testing.T) {
	tests := []struct {
		name    string
		padding []int
		content string
	}{
		{
			name:    "no padding",
			padding: []int{},
			content: "test",
		},
		{
			name:    "all sides equal",
			padding: []int{1},
			content: "test",
		},
		{
			name:    "vertical and horizontal",
			padding: []int{1, 2},
			content: "test",
		},
		{
			name:    "top, horizontal, bottom",
			padding: []int{1, 2, 3},
			content: "test",
		},
		{
			name:    "all sides different",
			padding: []int{1, 2, 3, 4},
			content: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyPadding(tt.content, tt.padding)
			assert.NotEmpty(t, result)
		})
	}
}

func TestGetPropHelpers(t *testing.T) {
	props := map[string]interface{}{
		"string": "value",
		"int":    42,
		"float":  3.14,
		"bool":   true,
	}

	t.Run("getStringProp", func(t *testing.T) {
		assert.Equal(t, "value", getStringProp(props, "string", "default"))
		assert.Equal(t, "default", getStringProp(props, "missing", "default"))
		assert.Equal(t, "default", getStringProp(props, "int", "default"))
	})

	t.Run("getIntProp", func(t *testing.T) {
		assert.Equal(t, 42, getIntProp(props, "int", 0))
		assert.Equal(t, 3, getIntProp(props, "float", 0))
		assert.Equal(t, 99, getIntProp(props, "missing", 99))
	})

	t.Run("getBoolProp", func(t *testing.T) {
		assert.Equal(t, true, getBoolProp(props, "bool", false))
		assert.Equal(t, false, getBoolProp(props, "missing", false))
		assert.Equal(t, false, getBoolProp(props, "string", false))
	})
}
