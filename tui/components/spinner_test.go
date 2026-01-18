package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSpinner(t *testing.T) {
	props := SpinnerProps{
		Style: "dots",
		Label: "Loading",
		Color: "blue",
	}

	result := RenderSpinner(props, 80)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Loading")
}

func TestParseSpinnerProps(t *testing.T) {
	props := map[string]interface{}{
		"style": "line",
		"label": "Processing",
		"color": "red",
		"speed": 200,
	}

	result := ParseSpinnerProps(props)
	assert.Equal(t, "line", result.Style)
	assert.Equal(t, "Processing", result.Label)
	assert.Equal(t, "red", result.Color)
	assert.Equal(t, 200, result.Speed)
}
