package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHelpAutoDetectStyle tests that sections style is auto-detected when style is not explicitly set
func TestHelpAutoDetectStyle(t *testing.T) {
	// Case 1: Sections provided, style not set (defaults to "compact")
	propsMap1 := map[string]interface{}{
		"sections": []interface{}{
			map[string]interface{}{
				"title": "Navigation",
				"items": []interface{}{
					map[string]interface{}{
						"key":         "Tab",
						"description": "Focus next input",
					},
				},
			},
		},
	}
	props1 := ParseHelpProps(propsMap1)
	model1 := NewHelpModel(props1, "test-help")
	view1 := model1.View()

	// Should auto-detect and use sections style
	assert.Contains(t, view1, "Navigation", "Should render sections even when style is default")
	assert.Contains(t, view1, "Tab", "Should render section items")
	assert.Contains(t, view1, "Focus next input", "Should render item descriptions")

	// Case 2: Sections provided with explicit "sections" style
	propsMap2 := map[string]interface{}{
		"style": "sections",
		"sections": []interface{}{
			map[string]interface{}{
				"title": "Actions",
				"items": []interface{}{
					map[string]interface{}{
						"key":         "Enter",
						"description": "Submit",
					},
				},
			},
		},
	}
	props2 := ParseHelpProps(propsMap2)
	model2 := NewHelpModel(props2, "test-help")
	view2 := model2.View()

	// Should use sections style
	assert.Contains(t, view2, "Actions", "Should render sections with explicit style")
	assert.Contains(t, view2, "Enter", "Should render section items")
	assert.Contains(t, view2, "Submit", "Should render item descriptions")

	// Case 3: No sections, style defaults to "compact"
	propsMap3 := map[string]interface{}{}
	props3 := ParseHelpProps(propsMap3)
	model3 := NewHelpModel(props3, "test-help")
	view3 := model3.View()

	// Should use compact style (default)
	assert.Contains(t, view3, "↑↓: Navigate", "Should use compact style when no sections")
	assert.NotContains(t, view3, "Navigation", "Should not render sections")
}

// TestHelpExplicitStyleTakesPrecedence tests that explicit style is respected
func TestHelpExplicitStyleTakesPrecedence(t *testing.T) {
	// Even if sections are provided, explicit "compact" style should use compact
	propsMap := map[string]interface{}{
		"style": "compact",
		"sections": []interface{}{
			map[string]interface{}{
				"title": "Navigation",
				"items": []interface{}{
					map[string]interface{}{
						"key":         "Tab",
						"description": "Focus next input",
					},
				},
			},
		},
	}
	props := ParseHelpProps(propsMap)
	model := NewHelpModel(props, "test-help")
	view := model.View()

	// Should use compact style (explicitly set)
	assert.Contains(t, view, "↑↓: Navigate", "Should respect explicit compact style")
	assert.NotContains(t, view, "Navigation", "Should not render sections with explicit compact style")
}

// TestHelpUserScenario tests the exact user scenario from the conversation
func TestHelpUserScenario(t *testing.T) {
	// This is the exact configuration from the user's test case
	propsMap := map[string]interface{}{
		"sections": []interface{}{
			map[string]interface{}{
				"title": "Navigation",
				"items": []interface{}{
					map[string]interface{}{"key": "Tab", "description": "Focus next input"},
					map[string]interface{}{"key": "Shift+Tab", "description": "Focus previous input"},
					map[string]interface{}{"key": "↑ / ↓", "description": "Navigate up/down"},
					map[string]interface{}{"key": "← / →", "description": "Navigate left/right"},
				},
			},
			map[string]interface{}{
				"title": "Actions",
				"items": []interface{}{
					map[string]interface{}{"key": "Enter", "description": "Submit/Confirm"},
					map[string]interface{}{"key": "Esc", "description": "Cancel/Back"},
					map[string]interface{}{"key": "q", "description": "Quit application"},
				},
			},
			map[string]interface{}{
				"title": "General",
				"items": []interface{}{
					map[string]interface{}{"key": "Ctrl+C", "description": "Force quit"},
					map[string]interface{}{"key": "Ctrl+R", "description": "Refresh UI"},
					map[string]interface{}{"key": "Ctrl+Z", "description": "Suspend application"},
				},
			},
		},
		"height": 15,
		"width":  80,
		"border": true,
	}

	props := ParseHelpProps(propsMap)
	model := NewHelpModel(props, "test-help")
	view := model.View()

	// Should render all sections
	assert.Contains(t, view, "Navigation", "Should render Navigation section")
	assert.Contains(t, view, "Actions", "Should render Actions section")
	assert.Contains(t, view, "General", "Should render General section")

	// Should render items
	assert.Contains(t, view, "Tab", "Should render Tab key")
	assert.Contains(t, view, "Focus next input", "Should render Tab description")
	assert.Contains(t, view, "Enter", "Should render Enter key")
	assert.Contains(t, view, "Submit/Confirm", "Should render Enter description")
	assert.Contains(t, view, "Ctrl+C", "Should render Ctrl+C key")
	assert.Contains(t, view, "Force quit", "Should render Ctrl+C description")
}
