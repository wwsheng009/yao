package components

import (
	"encoding/json"

	"github.com/charmbracelet/lipgloss"
)

// HeaderProps defines the properties for the Header component.
type HeaderProps struct {
	// Title is the header text
	Title string `json:"title"`

	// Align specifies the text alignment: "left", "center", "right"
	Align string `json:"align"`

	// Color specifies the foreground color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Bold makes the text bold
	Bold bool `json:"bold"`

	// Width specifies the header width (0 for auto)
	Width int `json:"width"`
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

// ParseHeaderProps converts a generic props map to HeaderProps using JSON unmarshaling.
func ParseHeaderProps(props map[string]interface{}) HeaderProps {
	// Set defaults
	hp := HeaderProps{
		Bold: true, // Default to bold
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &hp)
	}

	return hp
}
