package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidator_ValidConfig(t *testing.T) {
	cfg := &Config{
		Name: "Test TUI",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "header",
					Type: "text",
					Props: map[string]interface{}{
						"content": "Test",
					},
				},
			},
		},
		Data: map[string]interface{}{
			"test": "value",
		},
	}

	registry := GetGlobalRegistry()
	validator := NewConfigValidator(cfg, registry)

	assert.True(t, validator.Validate())
	assert.Equal(t, 0, len(validator.GetErrors()))
}

func TestConfigValidator_MissingName(t *testing.T) {
	cfg := &Config{
		Name: "",
		Layout: Layout{
			Direction: "vertical",
			Children:  []Component{},
		},
	}

	registry := GetGlobalRegistry()
	validator := NewConfigValidator(cfg, registry)

	assert.False(t, validator.Validate())
	assert.Greater(t, len(validator.GetErrors()), 0)

	errors := validator.GetErrors()
	found := false
	for _, err := range errors {
		if err.Path == "name" {
			found = true
			assert.Contains(t, err.Message, "required")
			break
		}
	}
	assert.True(t, found, "Name error not found")
}

func TestConfigValidator_OldFormatNestedLayout(t *testing.T) {
	cfg := &Config{
		Name: "Old Format Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Outer",
					},
				},
				{
					Type:      "layout",
					Direction: "horizontal",
					Children: []Component{
						{
							Type:      "layout",
							Direction: "vertical",
							Children: []Component{
								{
									ID:   "nested-text",
									Type: "text",
									Props: map[string]interface{}{
										"content": "Nested",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	validator := NewConfigValidator(cfg, GetGlobalRegistry())

	assert.True(t, validator.Validate())
	assert.Equal(t, 0, len(validator.GetErrors()))
}

func TestConfigValidator_MixedFormatNestedLayout(t *testing.T) {
	cfg := &Config{
		Name: "Mixed Format Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "layout",
					Props: map[string]interface{}{
						"layout": &Layout{
							Direction: "horizontal",
							Children: []Component{
								{
									ID:   "new-format",
									Type: "text",
									Props: map[string]interface{}{
										"content": "New Format",
									},
								},
							},
						},
					},
				},
				{
					Type:      "layout",
					Direction: "horizontal",
					Children: []Component{
						{
							// Old format nested layout
							Type:      "layout",
							Direction: "vertical",
							Children: []Component{
								{
									ID:   "old-format",
									Type: "text",
									Props: map[string]interface{}{
										"content": "Old Format",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	validator := NewConfigValidator(cfg, GetGlobalRegistry())

	assert.True(t, validator.Validate())
	assert.Equal(t, 0, len(validator.GetErrors()))
}
