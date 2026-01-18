package components

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
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

// FooterModel wraps the footer properties for TUI integration
type FooterModel struct {
	props FooterProps
	id    string // Unique identifier for this instance
}

// NewFooterModel creates a new FooterModel from FooterProps
func NewFooterModel(props FooterProps, id string) FooterModel {
	return FooterModel{
		props: props,
		id:    id,
	}
}

// HandleFooterUpdate handles updates for footer components
func HandleFooterUpdate(msg tea.Msg, footerModel *FooterModel) (FooterModel, tea.Cmd) {
	if footerModel == nil {
		return FooterModel{}, nil
	}

	return *footerModel, nil
}

// Init initializes the footer model
func (m *FooterModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the footer
func (m *FooterModel) View() string {
	return RenderFooter(m.props, m.props.Width)
}

// GetID returns the unique identifier for this component instance
func (m *FooterModel) GetID() string {
	return m.id
}

// SetFocus sets or removes focus from footer component
func (m *FooterModel) SetFocus(focus bool) {
	// Footer doesn't have focus concept
}

func (m *FooterModel) GetComponentType() string {
	return "footer"
}

func (m *FooterModel) Render(config core.RenderConfig) (string, error) {
	// 解析配置数据
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("FooterModel: invalid data type, expected map[string]interface{}, got %T", config.Data)
	}

	// 解析属性
	props := ParseFooterProps(propsMap)

	// 更新组件属性
	m.props = props

	// 验证必要的属性
	if m.props.Text == "" && propsMap["__bind_data"] == nil {
		return "", fmt.Errorf("FooterModel: missing required 'text' property")
	}

	// 返回渲染结果
	return m.View(), nil
}

// FooterComponentWrapper wraps FooterModel to implement ComponentInterface properly
type FooterComponentWrapper struct {
	model *FooterModel
}

// NewFooterComponentWrapper creates a wrapper that implements ComponentInterface
func NewFooterComponentWrapper(footerModel *FooterModel) *FooterComponentWrapper {
	return &FooterComponentWrapper{
		model: footerModel,
	}
}

func (w *FooterComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *FooterComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// Footer is a static component, no need to handle other messages
	return w, nil, core.Ignored
}

func (w *FooterComponentWrapper) View() string {
	return w.model.View()
}

func (w *FooterComponentWrapper) GetID() string {
	return w.model.id
}

func (w *FooterComponentWrapper) SetFocus(focus bool) {
	// Footer doesn't have focus concept
}

func (w *FooterComponentWrapper) GetComponentType() string {
	return "footer"
}

func (w *FooterComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig 更新渲染配置
func (w *FooterComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("FooterComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse footer properties
	props := ParseFooterProps(propsMap)

	// Update component properties
	w.model.props = props

	return nil
}

func (w *FooterComponentWrapper) Cleanup() {
	// Footer组件通常不需要清理资源
	// 这是一个空操作
}

// GetStateChanges returns the state changes from this component
func (w *FooterComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Static components don't have state changes
	return nil, false
}
