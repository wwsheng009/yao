package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// TextProps defines the properties for the Text component.
type TextProps struct {
	// Content is the text content
	Content string

	// Align specifies the text alignment: "left", "center", "right"
	Align string

	// Color specifies the foreground color
	Color string

	// Background specifies the background color
	Background string

	// Bold makes the text bold
	Bold bool

	// Italic makes the text italic
	Italic bool

	// Underline underlines the text
	Underline bool

	// Width specifies the text width (0 for auto)
	Width int

	// Padding specifies padding [vertical, horizontal]
	Padding []int

	// WordWrap enables word wrapping
	WordWrap bool
}

// RenderText renders a text component.
// This is a flexible text component with various styling options.
func RenderText(props TextProps, width int) string {
	// Default content
	content := props.Content
	if content == "" {
		content = ""
	}

	// Use provided width or default
	textWidth := props.Width
	if textWidth == 0 && width > 0 {
		textWidth = width - 2 // Account for padding
	}

	// Build style
	style := lipgloss.NewStyle()

	// Apply colors
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply text styling
	if props.Bold {
		style = style.Bold(true)
	}
	if props.Italic {
		style = style.Italic(true)
	}
	if props.Underline {
		style = style.Underline(true)
	}

	// Apply alignment
	if props.Align != "" {
		switch props.Align {
		case "center":
			style = style.Align(lipgloss.Center)
		case "right":
			style = style.Align(lipgloss.Right)
		default:
			style = style.Align(lipgloss.Left)
		}
	}

	// Apply width
	if textWidth > 0 {
		style = style.Width(textWidth)
	}

	// Apply padding
	if len(props.Padding) > 0 {
		switch len(props.Padding) {
		case 1:
			// All sides
			style = style.Padding(props.Padding[0])
		case 2:
			// Vertical, Horizontal
			style = style.Padding(props.Padding[0], props.Padding[1])
		case 4:
			// Top, Right, Bottom, Left
			style = style.Padding(props.Padding[0], props.Padding[1], props.Padding[2], props.Padding[3])
		}
	} else {
		// Default padding
		style = style.Padding(0, 1)
	}

	return style.Render(content)
}

// ParseTextProps converts a generic props map to TextProps.
func ParseTextProps(props map[string]interface{}) TextProps {
	tp := TextProps{}

	// Content - can be string or formatted from bound data
	if content, ok := props["content"].(string); ok {
		tp.Content = content
	} else if bindData, ok := props["__bind_data"]; ok {
		// Handle bound data
		tp.Content = fmt.Sprintf("%v", bindData)
	}

	if align, ok := props["align"].(string); ok {
		tp.Align = align
	}

	if color, ok := props["color"].(string); ok {
		tp.Color = color
	}

	if bg, ok := props["background"].(string); ok {
		tp.Background = bg
	}

	if bold, ok := props["bold"].(bool); ok {
		tp.Bold = bold
	}

	if italic, ok := props["italic"].(bool); ok {
		tp.Italic = italic
	}

	if underline, ok := props["underline"].(bool); ok {
		tp.Underline = underline
	}

	if width, ok := props["width"].(int); ok {
		tp.Width = width
	} else if widthFloat, ok := props["width"].(float64); ok {
		tp.Width = int(widthFloat)
	}

	if padding, ok := props["padding"].([]interface{}); ok {
		tp.Padding = make([]int, len(padding))
		for i, v := range padding {
			if intVal, ok := v.(int); ok {
				tp.Padding[i] = intVal
			} else if floatVal, ok := v.(float64); ok {
				tp.Padding[i] = int(floatVal)
			}
		}
	}

	if wordWrap, ok := props["wordWrap"].(bool); ok {
		tp.WordWrap = wordWrap
	}

	return tp
}
