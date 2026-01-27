package component

import (
	"encoding/json"

	"github.com/charmbracelet/lipgloss"
)

// lipglossStyleWrapper wraps lipgloss.Style for JSON serialization
type lipglossStyleWrapper struct {
	*lipgloss.Style
}

// UnmarshalJSON implements json.Unmarshaler for lipglossStyleWrapper
func (w *lipglossStyleWrapper) UnmarshalJSON(data []byte) error {
	var styleMap map[string]interface{}
	if err := json.Unmarshal(data, &styleMap); err != nil {
		return err
	}
	style := parseLipglossStyle(styleMap)
	w.Style = &style
	return nil
}

// GetStyle returns the underlying lipgloss.Style from the wrapper
func (w lipglossStyleWrapper) GetStyle() lipgloss.Style {
	if w.Style == nil {
		return lipgloss.NewStyle()
	}
	return *w.Style
}

// parseLipglossStyle parses a style map to lipgloss.Style
func parseLipglossStyle(styleMap map[string]interface{}) lipgloss.Style {
	style := lipgloss.NewStyle()

	if fg, ok := styleMap["foreground"].(string); ok {
		style = style.Foreground(lipgloss.Color(fg))
	}

	if bg, ok := styleMap["background"].(string); ok {
		style = style.Background(lipgloss.Color(bg))
	}

	if bold, ok := styleMap["bold"].(bool); ok && bold {
		style = style.Bold(true)
	}

	if italic, ok := styleMap["italic"].(bool); ok && italic {
		style = style.Italic(true)
	}

	if underline, ok := styleMap["underline"].(bool); ok && underline {
		style = style.Underline(true)
	}

	if align, ok := styleMap["align"].(string); ok {
		switch align {
		case "left":
			style = style.Align(lipgloss.Left)
		case "center":
			style = style.Align(lipgloss.Center)
		case "right":
			style = style.Align(lipgloss.Right)
		}
	}

	return style
}