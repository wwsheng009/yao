package components

import (
	"encoding/json"

	"github.com/charmbracelet/lipgloss"
)

// FooterProps defines the properties for the Footer component
type FooterProps struct {
	// Text is the footer text content
	Text string `json:"text"`

	// Height specifies the footer height (0 for auto)
	Height int `json:"height"`

	// Width specifies the footer width (0 for auto)
	Width int `json:"width"`

	// Color specifies the text color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Align specifies the text alignment (left, center, right)
	Align string `json:"align"`

	// Bold makes the text bold
	Bold bool `json:"bold"`

	// Italic makes the text italic
	Italic bool `json:"italic"`

	// Underline adds underline to the text
	Underline bool `json:"underline"`

	// MarginTop sets the top margin
	MarginTop int `json:"marginTop"`

	// MarginBottom sets the bottom margin
	MarginBottom int `json:"marginBottom"`

	// PaddingLeft sets the left padding
	PaddingLeft int `json:"paddingLeft"`

	// PaddingRight sets the right padding
	PaddingRight int `json:"paddingRight"`
}

// RenderFooter renders a footer component
func RenderFooter(props FooterProps, width int) string {
	style := lipgloss.NewStyle()

	// Apply colors
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply text decorations
	if props.Bold {
		style = style.Bold(props.Bold)
	}
	if props.Italic {
		style = style.Italic(props.Italic)
	}
	if props.Underline {
		style = style.Underline(props.Underline)
	}

	// Apply margins
	if props.MarginTop > 0 {
		style = style.MarginTop(props.MarginTop)
	}
	if props.MarginBottom > 0 {
		style = style.MarginBottom(props.MarginBottom)
	}

	// Apply padding
	if props.PaddingLeft > 0 {
		style = style.PaddingLeft(props.PaddingLeft)
	}
	if props.PaddingRight > 0 {
		style = style.PaddingRight(props.PaddingRight)
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

	// Apply width if specified
	finalWidth := props.Width
	if finalWidth <= 0 && width > 0 {
		finalWidth = width
	}
	if finalWidth > 0 {
		style = style.Width(finalWidth)
	}

	// Apply height if specified
	if props.Height > 0 {
		style = style.Height(props.Height)
	}

	// Render the footer text with the applied style
	return style.Render(props.Text)
}

// ParseFooterProps converts a generic props map to FooterProps using JSON unmarshaling
func ParseFooterProps(props map[string]interface{}) FooterProps {
	// Set defaults
	fp := FooterProps{
		Align: "left", // Default alignment
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &fp)
	}

	return fp
}
