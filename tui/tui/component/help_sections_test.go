package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHelpPropsWithSections(t *testing.T) {
	props := map[string]interface{}{
		"style": "sections",
		"sections": []interface{}{
			map[string]interface{}{
				"title": "Navigation",
				"items": []interface{}{
					map[string]interface{}{
						"key":         "Tab",
						"description": "Focus next input",
					},
					map[string]interface{}{
						"key":         "Shift+Tab",
						"description": "Focus previous input",
					},
				},
			},
			map[string]interface{}{
				"title": "Actions",
				"items": []interface{}{
					map[string]interface{}{
						"key":         "Enter",
						"description": "Submit/Confirm",
					},
					map[string]interface{}{
						"key":         "Esc",
						"description": "Cancel/Back",
					},
				},
			},
		},
		"width":  80,
		"height": 15,
		"border": true,
	}

	result := ParseHelpProps(props)

	// Check basic properties
	assert.Equal(t, "sections", result.Style)
	assert.Equal(t, 80, result.Width)
	assert.Equal(t, 15, result.Height)
	assert.True(t, result.Border)

	// Check sections
	assert.Len(t, result.Sections, 2)

	// Check first section
	assert.Equal(t, "Navigation", result.Sections[0].Title)
	assert.Len(t, result.Sections[0].Items, 2)
	assert.Equal(t, "Tab", result.Sections[0].Items[0].Key)
	assert.Equal(t, "Focus next input", result.Sections[0].Items[0].Description)
	assert.Equal(t, "Shift+Tab", result.Sections[0].Items[1].Key)
	assert.Equal(t, "Focus previous input", result.Sections[0].Items[1].Description)

	// Check second section
	assert.Equal(t, "Actions", result.Sections[1].Title)
	assert.Len(t, result.Sections[1].Items, 2)
	assert.Equal(t, "Enter", result.Sections[1].Items[0].Key)
	assert.Equal(t, "Submit/Confirm", result.Sections[1].Items[1].Description)
	assert.Equal(t, "Esc", result.Sections[1].Items[1].Key)
	assert.Equal(t, "Cancel/Back", result.Sections[1].Items[1].Description)
}

func TestHelpModelRenderSections(t *testing.T) {
	props := HelpProps{
		Style: "sections",
		Sections: []HelpSection{
			{
				Title: "Navigation",
				Items: []HelpItem{
					{Key: "Tab", Description: "Focus next input"},
					{Key: "Esc", Description: "Cancel"},
				},
			},
			{
				Title: "Actions",
				Items: []HelpItem{
					{Key: "Enter", Description: "Submit"},
					{Key: "q", Description: "Quit"},
				},
			},
		},
		SectionSeparator: "\n",
		ItemSeparator:    "\n",
	}

	model := NewHelpModel(props, "test-help")
	result := model.View()

	assert.Contains(t, result, "Navigation")
	assert.Contains(t, result, "Actions")
	assert.Contains(t, result, "Tab                   Focus next input")
	assert.Contains(t, result, "Esc                   Cancel")
	assert.Contains(t, result, "Enter                 Submit")
	assert.Contains(t, result, "q                     Quit")
}

func TestHelpModelRenderSectionsWithCustomColors(t *testing.T) {
	props := HelpProps{
		Style:            "sections",
		SectionTitleColor: "cyan",
		Color:            "white",
		Sections: []HelpSection{
			{
				Title: "Navigation",
				Items: []HelpItem{
					{Key: "Tab", Description: "Focus next input"},
				},
			},
		},
	}

	model := NewHelpModel(props, "test-help")
	result := model.View()

	assert.Contains(t, result, "Navigation")
	assert.Contains(t, result, "Tab")
	assert.Contains(t, result, "Focus next input")
}

func TestHelpModelRenderSectionsEmpty(t *testing.T) {
	props := HelpProps{
		Style:    "sections",
		Sections: []HelpSection{},
	}

	model := NewHelpModel(props, "test-help")
	result := model.View()

	assert.Contains(t, result, "No help sections available")
}

func TestHelpItemParsing(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{
			"key":         "Ctrl+C",
			"description": "Force quit",
		},
		map[string]interface{}{
			"key":         "Ctrl+R",
			"description": "Refresh UI",
		},
	}

	result := parseHelpItems(items)

	assert.Len(t, result, 2)
	assert.Equal(t, "Ctrl+C", result[0].Key)
	assert.Equal(t, "Force quit", result[0].Description)
	assert.Equal(t, "Ctrl+R", result[1].Key)
	assert.Equal(t, "Refresh UI", result[1].Description)
}

func TestHelpSectionParsing(t *testing.T) {
	sections := []interface{}{
		map[string]interface{}{
			"title": "General",
			"items": []interface{}{
				map[string]interface{}{
					"key":         "h",
					"description": "Show help",
				},
			},
		},
	}

	result := parseHelpSections(sections)

	assert.Len(t, result, 1)
	assert.Equal(t, "General", result[0].Title)
	assert.Len(t, result[0].Items, 1)
	assert.Equal(t, "h", result[0].Items[0].Key)
	assert.Equal(t, "Show help", result[0].Items[0].Description)
}

func TestParseHelpPropsWithKeyMap(t *testing.T) {
	props := map[string]interface{}{
		"keyMap": map[string]interface{}{
			"h": "Toggle help",
			"q": "Quit",
		},
		"style": "compact",
	}

	result := ParseHelpProps(props)

	assert.Equal(t, "compact", result.Style)
	assert.NotNil(t, result.KeyMap)
	assert.Equal(t, "Toggle help", result.KeyMap["h"])
	assert.Equal(t, "Quit", result.KeyMap["q"])
}

func TestHelpModelRenderBackwardCompatibility(t *testing.T) {
	// Test that non-sections styles still work
	props := HelpProps{
		Style:  "compact",
		KeyMap: map[string]interface{}{},
	}

	model := NewHelpModel(props, "test-help")
	result := model.View()

	assert.Contains(t, result, "↑↓: Navigate")
	assert.Contains(t, result, "Enter: Select")
	assert.Contains(t, result, "Esc: Back")
}

func TestHelpModelRenderFullStyle(t *testing.T) {
	props := HelpProps{
		Style:  "full",
		KeyMap: map[string]interface{}{},
	}

	model := NewHelpModel(props, "test-help")
	result := model.View()

	assert.Contains(t, result, "Navigation: ↑↓←→")
	assert.Contains(t, result, "Select: Enter")
	assert.Contains(t, result, "Quit: Ctrl+C or Esc")
}

func TestHelpModelRenderMinimalStyle(t *testing.T) {
	props := HelpProps{
		Style:  "minimal",
		KeyMap: map[string]interface{}{},
	}

	model := NewHelpModel(props, "test-help")
	result := model.View()

	assert.Contains(t, result, "↑↓")
	assert.Contains(t, result, "Enter")
	assert.Contains(t, result, "Esc")
}
