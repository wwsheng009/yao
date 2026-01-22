package components

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// TextProps defines the properties for the Text component.
type TextProps struct {
	// Content is the text content
	Content string `json:"content"`

	// Align specifies the text alignment: "left", "center", "right"
	Align string `json:"align"`

	// Color specifies the foreground color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Bold makes the text bold
	Bold bool `json:"bold"`

	// Italic makes the text italic
	Italic bool `json:"italic"`

	// Underline underlines the text
	Underline bool `json:"underline"`

	// Width specifies the text width (0 for auto)
	Width int `json:"width"`

	// Padding specifies the padding [vertical, horizontal]
	Padding []int `json:"padding"`

	// WordWrap enables word wrapping
	WordWrap bool `json:"wordWrap"`

	// VerticalAlign 指定垂直对齐方式
	VerticalAlign string `json:"verticalAlign"` // "top", "center", "bottom"

	// HorizontalAlign 指定水平对齐方式
	HorizontalAlign string `json:"horizontalAlign"` // "left", "center", "right"
}

// RenderText renders a text component.
// This is a flexible text component with various styling options.
func RenderText(props TextProps, width, height int) string {
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
		// Legacy support
		switch props.Align {
		case "center":
			style = style.Align(lipgloss.Center)
		case "right":
			style = style.Align(lipgloss.Right)
		default:
			style = style.Align(lipgloss.Left)
		}
	} else if props.HorizontalAlign != "" {
		switch props.HorizontalAlign {
		case "center":
			style = style.Align(lipgloss.Center)
		case "right":
			style = style.Align(lipgloss.Right)
		default:
			style = style.Align(lipgloss.Left)
		}
	}

	// Apply vertical alignment if height is provided
	if height > 0 {
		style = style.Height(height)
		switch props.VerticalAlign {
		case "center":
			style = style.AlignVertical(lipgloss.Center)
		case "bottom":
			style = style.AlignVertical(lipgloss.Bottom)
		default:
			style = style.AlignVertical(lipgloss.Top)
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

// ParseTextProps converts a generic props map to TextProps using JSON unmarshaling.
func ParseTextProps(props map[string]interface{}) TextProps {
	// Set defaults
	tp := TextProps{}

	// Handle Content and Padding separately
	if content, ok := props["content"].(string); ok {
		tp.Content = content
	} else if bindData, ok := props["__bind_data"]; ok {
		// Handle bound data
		tp.Content = fmt.Sprintf("%v", bindData)
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

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &tp)
	}

	return tp
}

// TextModel wraps the Text component to implement ComponentInterface
type TextModel struct {
	Props  TextProps
	Width  int
	Height int
	id     string // Unique identifier for this instance
}

// NewTextComponent creates a new Text component wrapper
func NewTextComponent(config core.RenderConfig, id string) *TextComponentWrapper {
	model := &TextModel{
		id: id,
	}

	// Update with provided configuration if available
	if config.Data != nil {
		if err := model.UpdateRenderConfig(config); err != nil {
			log.Error("Failed to update Text component config: %v", err)
		}
	}

	return &TextComponentWrapper{
		model: *model,
	}
}

// Init implements ComponentInterface
func (m *TextModel) Init() tea.Cmd {
	return nil
}

// UpdateMsg implements ComponentInterface
func (m *TextModel) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == m.id {
			return m, nil, core.Handled
		}
		return m, nil, core.Ignored
	}

	// Text is a static component, no need to handle other messages
	return m, nil, core.Ignored
}

// View implements ComponentInterface
func (m *TextModel) View() string {
	return RenderText(m.Props, m.Width, m.Height)
}

// GetID implements ComponentInterface
func (m *TextModel) GetID() string {
	return m.id
}

// SetFocus implements ComponentInterface
func (m *TextModel) SetFocus(focus bool) {
	// Text component doesn't support focus
}

func (m *TextModel) GetFocus() bool {
	return false // Text component never has focus
}

func (m *TextModel) GetComponentType() string {
	return "text"
}

func (m *TextModel) Render(config core.RenderConfig) (string, error) {
	// 解析配置数据
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TextModel: invalid data type, expected map[string]interface{}, got %T", config.Data)
	}

	// 解析属性
	props := ParseTextProps(propsMap)

	// 更新组件属性
	m.Props = props
	m.Width = config.Width
	m.Height = config.Height

	// 返回渲染结果
	return m.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (m *TextModel) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TextModel: invalid data type, expected map[string]interface{}, got %T", config.Data)
	}

	// 解析属性
	props := ParseTextProps(propsMap)

	// 更新组件属性
	m.Props = props
	m.Width = config.Width
	m.Height = config.Height

	return nil
}

// Cleanup 清理资源
func (m *TextModel) Cleanup() {
	// TextModel 通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (m *TextModel) GetStateChanges() (map[string]interface{}, bool) {
	// Text component is static, no state changes
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (m *TextModel) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}

// TextComponentWrapper wraps TextModel to implement ComponentInterface properly
type TextComponentWrapper struct {
	model TextModel
	props TextProps
	id    string
}

// NewTextComponentWrapper creates a wrapper that implements ComponentInterface
func NewTextComponentWrapper(props TextProps, id string) *TextComponentWrapper {
	model := TextModel{
		Props: props,
		Width: props.Width,
		id:    id,
	}
	return &TextComponentWrapper{
		model: model,
		props: props,
		id:    id,
	}
}

func (w *TextComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *TextComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w, nil, core.Handled
		}
		return w, nil, core.Ignored
	}

	// Text is a static component, no need to handle other messages
	return w, nil, core.Ignored
}

func (w *TextComponentWrapper) View() string {
	return RenderText(w.props, w.props.Width, w.model.Height)
}

func (w *TextComponentWrapper) GetID() string {
	return w.id
}

func (w *TextComponentWrapper) SetFocus(focus bool) {
	// Text doesn't have focus concept
}

func (w *TextComponentWrapper) GetFocus() bool {
	return false // Text component never has focus
}

func (w *TextComponentWrapper) GetComponentType() string {
	return "text"
}

func (w *TextComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// 解析配置数据
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TextComponentWrapper: invalid data type")
	}

	// 解析属性
	props := ParseTextProps(propsMap)

	// 更新组件属性
	w.props = props
	w.model.Props = props
	w.model.Width = config.Width
	w.model.Height = config.Height

	// 返回渲染结果
	return w.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *TextComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TextComponentWrapper: invalid data type")
	}

	// Parse text properties
	props := ParseTextProps(propsMap)

	// Update component properties
	w.props = props
	w.model.Props = props
	w.model.Width = config.Width
	w.model.Height = config.Height

	return nil
}

// Cleanup cleans up resources used by the text component
func (w *TextComponentWrapper) Cleanup() {
	// Text components typically don't need cleanup
	// This is a no-op for text components
}

// GetStateChanges returns the state changes from this component
func (w *TextComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Static components don't have state changes
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *TextComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}

// SetSize 更新文本组件的实际显示尺寸
func (w *TextComponentWrapper) SetSize(width, height int) {
	w.model.Width = width
	w.model.Height = height
}

// Measure 返回文本组件的理想尺寸
func (w *TextComponentWrapper) Measure(maxWidth, maxHeight int) (width, height int) {
	content := w.props.Content
	
	// 计算内容宽度
	contentWidth := lipgloss.Width(content)
	
	// 如果指定了宽度，则使用指定宽度，否则使用内容宽度
	if w.props.Width > 0 {
		width = w.props.Width
	} else {
		width = contentWidth
		// 限制在 maxWidth 内
		if width > maxWidth {
			width = maxWidth
		}
	}
	
	// 计算内容高度（考虑换行）
	lines := strings.Split(content, "\n")
	height = len(lines)
	
	// 如果指定了高度，则使用指定高度，否则根据内容计算
	if w.model.Height > 0 {
		height = w.model.Height
	} else {
		// 限制在 maxHeight 内
		if height > maxHeight {
			height = maxHeight
		}
	}
	
	return width, height
}
