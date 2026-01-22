package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestLayoutPaddingRendering(t *testing.T) {
	// Setup: A fixed 40x20 container with padding 2, containing a text "X"
	cfg := &Config{
		Name:     "Test Padding",
		LogLevel: "trace",
		Layout: Layout{
			Direction: "vertical",
			Padding:   []int{2, 2, 2, 2}, // Top, Right, Bottom, Left
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "X",
					},
					// Explicitly fill available space to make size predictable
					Width:  func(i int) interface{} { return i }(36), // 40 - 4
					Height: func(i int) interface{} { return i }(16), // 20 - 4
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Force window size larger than container to ensure it's not constrained by window
	model.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	m := model

	// Render
	result := m.RenderLayout()

	// Analysis
	lines := strings.Split(result, "\n")
	height := len(lines)

	// Check max width
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	t.Logf("Rendered Result:\n%s", result)
	t.Logf("Dimensions: %dx%d", maxWidth, height)

	// Expectation: The output should be 40x20 (Container size)
	// Current Prediction: It might be 36x16 (Child size) because Renderer ignores Container Padding/Size wrapping

	// If renderer works correctly, it should include the padding
	assert.Equal(t, 20, height, "Height should include padding (total 20)")
	assert.Equal(t, 40, maxWidth, "Width should include padding (total 40)")
}
