package components

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// ViewportProps defines the properties for the Viewport component
type ViewportProps struct {
	// Content is the content to display in viewport
	Content string `json:"content"`

	// Width specifies the viewport width (0 for auto)
	Width int `json:"width"`

	// Height specifies the viewport height (0 for auto)
	Height int `json:"height"`

	// ShowScrollbar determines if scrollbar is shown
	ShowScrollbar bool `json:"showScrollbar"`

	// ScrollbarStyle is the style for scrollbar
	ScrollbarStyle lipglossStyleWrapper `json:"scrollbarStyle"`

	// BorderStyle is the style for viewport border
	BorderStyle lipglossStyleWrapper `json:"borderStyle"`

	// Style is the general viewport style
	Style lipglossStyleWrapper `json:"style"`

	// EnableGlamour enables Markdown rendering with Glamour
	EnableGlamour bool `json:"enableGlamour"`

	// GlamourStyle sets the Glamour style for Markdown rendering
	GlamourStyle string `json:"glamourStyle"`

	// AutoScroll determines if viewport automatically scrolls to bottom
	AutoScroll bool `json:"autoScroll"`
}

// ViewportModel wraps the viewport.Model to handle TUI integration
type ViewportModel struct {
	viewport.Model
	props ViewportProps
}

// RenderViewport renders a viewport component
func RenderViewport(props ViewportProps, width int) string {
	vp := viewport.New(0, 0) // width and height will be set later

	// Set content
	content := props.Content

	// Apply Markdown rendering if enabled
	if props.EnableGlamour {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0), // Use viewport width for wrapping
		)
		if err == nil {
			rendered, err := renderer.Render(content)
			if err == nil {
				content = rendered
			}
		}
	}

	vp.SetContent(content)

	// Set dimensions
	viewWidth := props.Width
	if viewWidth <= 0 && width > 0 {
		viewWidth = width
	}

	viewHeight := props.Height
	if viewHeight <= 0 {
		// Estimate height based on content if not specified
		lineCount := strings.Count(content, "\n") + 1
		if lineCount < 10 {
			viewHeight = lineCount + 2 // Add some padding
		} else {
			viewHeight = 15 // Default height
		}
	}

	if viewWidth > 0 {
		vp.Width = viewWidth
	}
	if viewHeight > 0 {
		vp.Height = viewHeight
	}

	// Apply styles
	style := props.Style.GetStyle()
	if style.String() != lipgloss.NewStyle().String() {
		vp.Style = style
	}

	return vp.View()
}

// ParseViewportProps converts a generic props map to ViewportProps using JSON unmarshaling
func ParseViewportProps(props map[string]interface{}) ViewportProps {
	// Set defaults
	vp := ViewportProps{
		EnableGlamour: false,
		GlamourStyle:  "dark",
		AutoScroll:    false,
		ShowScrollbar: true,
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &vp)
	}

	return vp
}

// HandleViewportUpdate handles updates for viewport components
// This is used when the viewport is interactive (scrolling, etc.)
func HandleViewportUpdate(msg tea.Msg, viewportModel *ViewportModel) (ViewportModel, tea.Cmd) {
	if viewportModel == nil {
		return ViewportModel{}, nil
	}

	var cmd tea.Cmd
	viewportModel.Model, cmd = viewportModel.Model.Update(msg)
	return *viewportModel, cmd
}
