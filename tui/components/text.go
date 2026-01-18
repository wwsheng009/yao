package components

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	Props TextProps
	Width int
	id    string // Unique identifier for this instance
}

// NewTextComponent creates a new Text component wrapper
func NewTextComponent(id string) *TextComponentWrapper {
	model := &TextModel{
		id: id,
	}
	return &TextComponentWrapper{
		model: model,
	}
}

// Init implements ComponentInterface
func (m *TextModel) Init() tea.Cmd {
	return nil
}

// UpdateMsg implements ComponentInterface
func (m *TextModel) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	switch msg := msg.(type) {
	case core.TargetedMsg:
		if msg.TargetID == "text" {
			return m, nil, core.Handled
		}
	}
	return m, nil, core.Ignored
}

// View implements ComponentInterface
func (m *TextModel) View() string {
	return RenderText(m.Props, m.Width)
}

// GetID implements ComponentInterface
func (m *TextModel) GetID() string {
	return m.id
}

// SetFocus implements ComponentInterface
func (m *TextModel) SetFocus(focus bool) {
	// Text component doesn't support focus
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

	// 验证必要的属性
	if m.Props.Content == "" && propsMap["__bind_data"] == nil {
		return "", fmt.Errorf("TextModel: missing required 'content' property")
	}

	// 返回渲染结果
	return m.View(), nil
}

// TextComponentWrapper wraps TextModel to implement ComponentInterface properly
type TextComponentWrapper struct {
	model *TextModel
}

// NewTextComponentWrapper creates a wrapper that implements ComponentInterface
func NewTextComponentWrapper(textModel *TextModel) *TextComponentWrapper {
	return &TextComponentWrapper{
		model: textModel,
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
		if msg.TargetID == w.model.id {
			return w, nil, core.Handled
		}
		return w, nil, core.Ignored
	}

	// Text is a static component, no need to handle other messages
	return w, nil, core.Ignored
}

func (w *TextComponentWrapper) View() string {
	return w.model.View()
}

func (w *TextComponentWrapper) GetID() string {
	return w.model.id
}

func (w *TextComponentWrapper) SetFocus(focus bool) {
	// Text doesn't have focus concept
}

func (w *TextComponentWrapper) GetComponentType() string {
	return "text"
}

func (w *TextComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}
