package components

import (
	"encoding/json"
	"fmt"
	"time"

	cursor "github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// CursorProps defines the properties for the Cursor component
type CursorProps struct {
	// Style specifies the cursor style: "blink", "static", "hide"
	Style string `json:"style"`

	// Color specifies the cursor color
	Color string `json:"color"`

	// Background specifies the cursor background color
	Background string `json:"background"`

	// Blink makes the cursor blink
	Blink bool `json:"blink"`

	// BlinkSpeed specifies the blink speed in milliseconds
	BlinkSpeed int `json:"blinkSpeed"`

	// Width specifies the cursor width
	Width int `json:"width"`

	// Height specifies the cursor height
	Height int `json:"height"`

	// Position specifies the cursor position
	Position int `json:"position"`

	// Visible determines if the cursor is visible
	Visible bool `json:"visible"`

	// Char specifies the cursor character
	Char string `json:"char"`
}

// CursorModel wraps the cursor.Model to handle TUI integration
type CursorModel struct {
	cursor.Model
	props CursorProps
	id    string // Unique identifier for this instance
}

// RenderCursor renders a cursor component
func RenderCursor(props CursorProps, width int) string {
	c := cursor.New()

	// Set cursor style
	if props.Style != "" {
		c.SetMode(getCursorMode(props.Style))
	}

	// Set blink speed
	if props.BlinkSpeed > 0 {
		c.BlinkSpeed = time.Duration(props.BlinkSpeed) * time.Millisecond
	}

	// Set cursor character
	if props.Char != "" {
		c.SetChar(props.Char)
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Set width if specified
	if props.Width > 0 {
		style = style.Width(props.Width)
	} else if width > 0 {
		style = style.Width(width)
	}

	// Create cursor display
	cursorChar := getCursorChar(props.Style, props.Char)

	return style.Render(cursorChar)
}

// ParseCursorProps converts a generic props map to CursorProps using JSON unmarshaling
func ParseCursorProps(props map[string]interface{}) CursorProps {
	// Set defaults
	cp := CursorProps{
		Style:      "blink",
		Blink:      true,
		BlinkSpeed: 530,
		Visible:    true,
		Char:       "|",
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &cp)
	}

	return cp
}

// NewCursorModel creates a new CursorModel from CursorProps
func NewCursorModel(props CursorProps, id string) CursorModel {
	c := cursor.New()

	// Set cursor style
	if props.Style != "" {
		c.SetMode(getCursorMode(props.Style))
	}

	// Set blink speed
	if props.BlinkSpeed > 0 {
		c.BlinkSpeed = time.Duration(props.BlinkSpeed) * time.Millisecond
	}

	// Set cursor character
	if props.Char != "" {
		c.SetChar(props.Char)
	}

	return CursorModel{
		Model: c,
		props: props,
		id:    id,
	}
}

// HandleCursorUpdate handles updates for cursor components
func HandleCursorUpdate(msg tea.Msg, cursorModel *CursorModel) (CursorModel, tea.Cmd) {
	if cursorModel == nil {
		return CursorModel{}, nil
	}

	var cmd tea.Cmd
	cursorModel.Model, cmd = cursorModel.Model.Update(msg)
	return *cursorModel, cmd
}

// View returns the string representation of the cursor
func (m *CursorModel) View() string {
	if !m.props.Visible {
		return ""
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if m.props.Color != "" {
		style = style.Foreground(lipgloss.Color(m.props.Color))
	}
	if m.props.Background != "" {
		style = style.Background(lipgloss.Color(m.props.Background))
	}

	// Create cursor display
	cursorChar := getCursorChar(m.props.Style, m.props.Char)

	return style.Render(cursorChar)
}

// GetID returns the unique identifier for this component instance
func (m *CursorModel) GetID() string {
	return m.id
}

// SetPosition sets the cursor position
func (m *CursorModel) SetPosition(pos int) {
	m.props.Position = pos
}

// SetVisible sets the cursor visibility
func (m *CursorModel) SetVisible(visible bool) {
	m.props.Visible = visible
}

// CursorComponentWrapper wraps CursorModel to implement ComponentInterface properly
type CursorComponentWrapper struct {
	model *CursorModel
}

// NewCursorComponentWrapper creates a wrapper that implements ComponentInterface
func NewCursorComponentWrapper(cursorModel *CursorModel) *CursorComponentWrapper {
	return &CursorComponentWrapper{
		model: cursorModel,
	}
}

func (w *CursorComponentWrapper) Init() tea.Cmd {
	return w.model.BlinkCmd()
}

func (w *CursorComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For cursor, just update the model
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// Publish cursor blink event
	if w.model.props.Blink {
		eventCmd := core.PublishEvent(w.model.id, "CURSOR_BLINK", map[string]interface{}{
			"visible": w.model.props.Visible,
			"style":   w.model.props.Style,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}

	return w, cmd, core.Handled
}

func (w *CursorComponentWrapper) View() string {
	return w.model.View()
}

func (w *CursorComponentWrapper) GetID() string {
	return w.model.id
}

func (w *CursorComponentWrapper) SetFocus(focus bool) {
	// Cursor focus is handled internally
}

func (w *CursorComponentWrapper) GetComponentType() string {
	return "cursor"
}

func (m *CursorModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("CursorModel: invalid data type")
	}

	// Parse cursor properties
	props := ParseCursorProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

func (w *CursorComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// getCursorMode returns the cursor mode based on style string
func getCursorMode(style string) cursor.Mode {
	switch style {
	case "static":
		return cursor.CursorStatic
	case "hide":
		return cursor.CursorHide
	default: // "blink"
		return cursor.CursorBlink
	}
}

// getCursorChar returns the cursor character based on style
func getCursorChar(style, customChar string) string {
	if customChar != "" {
		return customChar
	}

	switch style {
	case "static":
		return "█"
	case "hide":
		return ""
	default: // "blink"
		return "|"
	}
}
