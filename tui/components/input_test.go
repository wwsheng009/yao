package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInputProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected InputProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: InputProps{
				Prompt: "> ", // default prompt
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"placeholder": "Enter text",
				"value":       "test value",
				"prompt":      "$ ",
				"color":       "red",
				"background":  "blue",
				"width":       20,
				"height":      1,
				"disabled":    true,
			},
			expected: InputProps{
				Placeholder:  "Enter text",
				Value:        "test value",
				Prompt:       "$ ",
				Color:        "red",
				Background:   "blue",
				Width:        20,
				Height:       1,
				Disabled:     true,
			},
		},
		{
			name: "float width",
			props: map[string]interface{}{
				"width": 25.0,
			},
			expected: InputProps{
				Prompt: "> ", // default prompt
				Width:  25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInputProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderInput(t *testing.T) {
	props := InputProps{
		Placeholder: "Enter text",
		Value:       "test",
		Prompt:      "> ",
		Width:       20,
	}

	result := RenderInput(props, 80)
	assert.NotEmpty(t, result)
}