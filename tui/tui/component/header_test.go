package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeaderProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected HeaderProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: HeaderProps{
				Align: "left", // default alignment
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"title":      "Header text",
				"width":      80,
				"color":      "white",
				"background": "blue",
				"align":      "center",
				"bold":       true,
			},
			expected: HeaderProps{
				Title:      "Header text",
				Width:      80,
				Color:      "white",
				Background: "blue",
				Align:      "center",
				Bold:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseHeaderProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderHeader(t *testing.T) {
	props := HeaderProps{
		Title:      "This is a header",
		Width:      80,
		Align:      "center",
		Background: "blue",
		Color:      "white",
	}

	result := RenderHeader(props, 80)
	assert.NotEmpty(t, result)
}

func TestHeaderModel_View(t *testing.T) {
	model := NewHeaderModel(HeaderProps{Title: "Test Header"}, "test-header")
	result := RenderHeader(model.props, 30)
	assert.NotEmpty(t, result)
}

func TestHeaderModel_GetID(t *testing.T) {
	model := NewHeaderModel(HeaderProps{}, "test-header")
	assert.Equal(t, "test-header", model.GetID())
}

func TestHeaderModel_SetFocus(t *testing.T) {
	model := NewHeaderModel(HeaderProps{}, "test-header")
	// Should not panic
	model.SetFocus(true)
	model.SetFocus(false)
}
