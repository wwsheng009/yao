package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFooterProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected FooterProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: FooterProps{
				Align: "left", // default alignment
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"text":         "Footer text",
				"height":       2,
				"width":        80,
				"color":        "white",
				"background":   "blue",
				"align":        "center",
				"bold":         true,
				"italic":       false,
				"underline":    true,
				"marginTop":    1,
				"marginBottom": 1,
				"paddingLeft":  2,
				"paddingRight": 2,
			},
			expected: FooterProps{
				Text:         "Footer text",
				Height:       2,
				Width:        80,
				Color:        "white",
				Background:   "blue",
				Align:        "center",
				Bold:         true,
				Italic:       false,
				Underline:    true,
				MarginTop:    1,
				MarginBottom: 1,
				PaddingLeft:  2,
				PaddingRight: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFooterProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderFooter(t *testing.T) {
	props := FooterProps{
		Text:       "This is a footer",
		Width:      80,
		Align:      "center",
		Background: "blue",
		Color:      "white",
	}

	result := RenderFooter(props, 80)
	assert.NotEmpty(t, result)
}