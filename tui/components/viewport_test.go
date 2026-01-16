package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseViewportProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected ViewportProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: ViewportProps{
				EnableGlamour: false,
				GlamourStyle:  "dark",
				AutoScroll:    false,
				ShowScrollbar: true,
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"content":       "Hello world",
				"width":         80,
				"height":        20,
				"showScrollbar": false,
				"enableGlamour": true,
				"glamourStyle":  "light",
				"autoScroll":    true,
			},
			expected: ViewportProps{
				Content:       "Hello world",
				Width:         80,
				Height:        20,
				ShowScrollbar: false,
				EnableGlamour: true,
				GlamourStyle:  "light",
				AutoScroll:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseViewportProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderViewport(t *testing.T) {
	props := ViewportProps{
		Content:       "# Hello\n\nThis is a test.",
		Width:         80,
		Height:        10,
		EnableGlamour: true,
	}

	result := RenderViewport(props, 80)
	assert.NotEmpty(t, result)
}