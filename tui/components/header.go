package components

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
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
		Align: "left", // Default alignment
		Bold:  false,  // Default to not bold
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &hp)
	}

	return hp
}

// HeaderModel wraps the header properties for TUI integration
type HeaderModel struct {
	props HeaderProps
	id    string // Unique identifier for this instance
}

// NewHeaderModel creates a new HeaderModel
func NewHeaderModel(props HeaderProps, id string) HeaderModel {
	return HeaderModel{
		props: props,
		id: id,
	}
}

// HandleHeaderUpdate handles updates for header components
func HandleHeaderUpdate(msg tea.Msg, headerModel *HeaderModel) (HeaderModel, tea.Cmd) {
	if headerModel == nil {
		return HeaderModel{}, nil
	}

	return *headerModel, nil
}

// Init initializes the header model
func (m *HeaderModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the header
func (m *HeaderModel) View() string {
	return RenderHeader(m.props, m.props.Width)
}

// GetID returns the unique identifier for this component instance
func (m *HeaderModel) GetID() string {
	return m.id
}

// SetFocus sets or removes focus from header component
func (m *HeaderModel) SetFocus(focus bool) {
	// Header doesn't have focus concept
}

// HeaderComponentWrapper wraps HeaderModel to implement ComponentInterface properly
type HeaderComponentWrapper struct {
	model *HeaderModel
}

// NewHeaderComponent creates a new Header component wrapper
func NewHeaderComponent(config core.RenderConfig, id string) *HeaderComponentWrapper {
	var props HeaderProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseHeaderProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Title == "" {
		props = HeaderProps{
			Title: "",
			Align: "left",
			Color: "",
			Background: "",
		}
	}
	
	model := NewHeaderModel(props, id)
	return &HeaderComponentWrapper{
		model: &model,
	}
}

func (w *HeaderComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *HeaderComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// Header is a static component, no need to handle other messages
	return w, nil, core.Ignored
}

func (w *HeaderComponentWrapper) View() string {
	return w.model.View()
}

func (w *HeaderComponentWrapper) GetID() string {
	return w.model.id
}

func (w *HeaderComponentWrapper) SetFocus(focus bool) {
	// Header doesn't have focus concept
}

func (w *HeaderComponentWrapper) GetComponentType() string {
	return "header"
}

func (m *HeaderModel) Render(config core.RenderConfig) (string, error) {
	// 解析配置数据
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("HeaderModel: invalid data type, expected map[string]interface{}, got %T", config.Data)
	}

	// 解析属性
	props := ParseHeaderProps(propsMap)

	// 更新组件属性
	m.props = props

	// 验证必要的属性
	if m.props.Title == "" && propsMap["__bind_data"] == nil {
		return "", fmt.Errorf("HeaderModel: missing required 'title' property")
	}

	// 返回渲染结果
	return m.View(), nil
}

func (w *HeaderComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *HeaderComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("HeaderComponentWrapper: invalid data type")
	}

	// Parse header properties
	props := ParseHeaderProps(propsMap)

	// Update component properties
	w.model.props = props

	return nil
}

// Cleanup cleans up resources used by the header component
func (w *HeaderComponentWrapper) Cleanup() {
	// Header components typically don't need cleanup
	// This is a no-op for header components
}

// GetStateChanges returns the state changes from this component
func (w *HeaderComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Static components don't have state changes
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *HeaderComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}
