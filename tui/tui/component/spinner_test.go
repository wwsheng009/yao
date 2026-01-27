package component

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

// TestRenderSpinnerDotsStyle tests dots spinner style
func TestRenderSpinnerDotsStyle(t *testing.T) {
	props := SpinnerProps{
		Style: "dots",
		Label: "Loading...",
		Color: "blue",
	}

	result := RenderSpinner(props, 80)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Loading")
}

// TestRenderSpinnerLineStyle tests line spinner style
func TestRenderSpinnerLineStyle(t *testing.T) {
	props := SpinnerProps{
		Style: "line",
		Label: "Processing...",
		Color: "green",
	}

	result := RenderSpinner(props, 80)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Processing")
}

// TestSpinnerWithDifferentSpeeds tests spinner with different animation speeds
func TestSpinnerWithDifferentSpeeds(t *testing.T) {
	testCases := []struct {
		name  string
		speed int
		props SpinnerProps
	}{
		{
			name:  "Very slow spinner",
			speed: 1000,
			props: SpinnerProps{Style: "dots", Speed: 1000},
		},
		{
			name:  "Fast spinner",
			speed: 100,
			props: SpinnerProps{Style: "dots", Speed: 100},
		},
		{
			name:  "Default speed",
			speed: 300,
			props: SpinnerProps{Style: "dots"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result SpinnerProps
			if tc.speed == 300 {
				// Default case, Speed not set in props
				result = ParseSpinnerProps(map[string]interface{}{
					"style": "dots",
				})
			} else {
				result = ParseSpinnerProps(map[string]interface{}{
					"style": tc.props.Style,
					"speed": tc.props.Speed,
				})
			}
			assert.GreaterOrEqual(t, result.Speed, 0)
		})
	}
}

// TestSpinnerWithEmptyLabel tests spinner without label
func TestSpinnerWithEmptyLabel(t *testing.T) {
	props := SpinnerProps{
		Style: "dots",
		Label: "",
		Color: "white",
	}

	result := RenderSpinner(props, 80)
	assert.NotEmpty(t, result)
}

// TestSpinnerWrapBehavior tests spinner behavior at different widths
func TestSpinnerWrapBehavior(t *testing.T) {
	props := SpinnerProps{
		Style: "dots",
		Label: "Loading data from server...",
		Color: "blue",
	}

	// Test with different widths
	widths := []int{20, 40, 60, 80, 100}
	for _, width := range widths {
		result := RenderSpinner(props, width)
		assert.NotEmpty(t, result)
	}
}

// Helper function to convert SpinnerProps to map[string]interface{}
func mapToInterface(props SpinnerProps) map[string]interface{} {
	return map[string]interface{}{
		"style": props.Style,
		"label": props.Label,
		"color": props.Color,
		"speed": props.Speed,
	}
}
