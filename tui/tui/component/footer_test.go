package component

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/core"
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

func TestFooterModel_UpdateMsg(t *testing.T) {
	// Create wrapper
	wrapper := NewFooterComponentWrapper(FooterProps{Text: "Test Footer"}, "test-footer")

	// Test targeted message
	targetedMsg := core.TargetedMsg{
		TargetID: "test-footer",
		InnerMsg: nil,
	}

	updatedWrapper, cmd, response := wrapper.UpdateMsg(targetedMsg)
	assert.Equal(t, core.Ignored, response)
	assert.Nil(t, cmd)
	assert.Equal(t, wrapper, updatedWrapper)

	// Test non-targeted message
	updatedWrapper, cmd, response = wrapper.UpdateMsg(tea.KeyMsg{})
	assert.Equal(t, core.Ignored, response)
	assert.Nil(t, cmd)
	assert.Equal(t, wrapper, updatedWrapper)
}

func TestFooterModel_View(t *testing.T) {
	wrapper := NewFooterComponentWrapper(FooterProps{Text: "Test Footer"}, "test-footer")
	result := wrapper.View()
	assert.NotEmpty(t, result)
}

func TestFooterModel_GetID(t *testing.T) {
	wrapper := NewFooterComponentWrapper(FooterProps{Text: "Test Footer"}, "test-footer")
	assert.Equal(t, "test-footer", wrapper.GetID())
}

func TestFooterModel_SetFocus(t *testing.T) {
	wrapper := NewFooterComponentWrapper(FooterProps{Text: "Test Footer"}, "test-footer")
	// Should not panic
	wrapper.SetFocus(true)
	wrapper.SetFocus(false)
}
