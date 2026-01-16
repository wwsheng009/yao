package components

import (
	"github.com/charmbracelet/lipgloss"
)

// HeaderProps defines the properties for the Header component.
type HeaderProps struct {
	// Title is the header text
	Title string

	// Align specifies the text alignment: "left", "center", "right"
	Align string

	// Color specifies the foreground color
	Color string

	// Background specifies the background color
	Background string

	// Bold makes the text bold
	Bold bool

	// Width specifies the header width (0 for auto)
	Width int
}

// RenderHeader renders a header component.
// This is a standalone component that can be styled and positioned.
func RenderHeader(props HeaderProps, width int) string {
	// Default values
	if props.Title == "" {
		props.Title = "Header"
	}
	if props.Color == "" {
		props.Color = "205" // Default pink/magenta
	}
	if props.Background == "" {
		props.Background = "235" // Default dark gray
	}
	if props.Align == "" {
		props.Align = "left"
	}

	// Use provided width or default
	headerWidth := props.Width
	if headerWidth == 0 {
		headerWidth = width
	}

	// Build style
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(props.Color)).
		Background(lipgloss.Color(props.Background)).
		Padding(0, 1)

	if props.Bold {
		style = style.Bold(true)
	}

	if headerWidth > 0 {
		style = style.Width(headerWidth)
	}

	// Apply alignment
	switch props.Align {
	case "center":
		style = style.Align(lipgloss.Center)
	case "right":
		style = style.Align(lipgloss.Right)
	default:
		style = style.Align(lipgloss.Left)
	}

	return style.Render(props.Title)
}

// ParseHeaderProps converts a generic props map to HeaderProps.
func ParseHeaderProps(props map[string]interface{}) HeaderProps {
	hp := HeaderProps{
		Bold: true, // Default to bold
	}

	if title, ok := props["title"].(string); ok {
		hp.Title = title
	}

	if align, ok := props["align"].(string); ok {
		hp.Align = align
	}

	if color, ok := props["color"].(string); ok {
		hp.Color = color
	}

	if bg, ok := props["background"].(string); ok {
		hp.Background = bg
	}

	if bold, ok := props["bold"].(bool); ok {
		hp.Bold = bold
	}

	if width, ok := props["width"].(int); ok {
		hp.Width = width
	} else if widthFloat, ok := props["width"].(float64); ok {
		hp.Width = int(widthFloat)
	}

	return hp
}
