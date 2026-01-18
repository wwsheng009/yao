package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// TextareaProps defines the properties for the Textarea component
type TextareaProps struct {
	// Placeholder is the placeholder text when the textarea is empty
	Placeholder string `json:"placeholder"`

	// Value is the initial value of the textarea
	Value string `json:"value"`

	// Prompt is the prompt character/string before the textarea
	Prompt string `json:"prompt"`

	// Color specifies the text color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Width specifies the textarea width (0 for auto)
	Width int `json:"width"`

	// Height specifies the textarea height (0 for auto)
	Height int `json:"height"`

	// MaxHeight specifies the maximum height
	MaxHeight int `json:"maxHeight"`

	// Disabled determines if the textarea is disabled
	Disabled bool `json:"disabled"`

	// ShowLineNumbers shows/hides line numbers
	ShowLineNumbers bool `json:"showLineNumbers"`

	// CharLimit specifies the maximum character limit (0 for unlimited)
	CharLimit int `json:"charLimit"`
}

// TextareaModel wraps the textarea.Model to handle TUI integration
type TextareaModel struct {
	textarea.Model
	props TextareaProps
	id    string // Unique identifier for this instance
}

// RenderTextarea renders a textarea component
func RenderTextarea(props TextareaProps, width int) string {
	ta := textarea.New()

	// Set placeholder
	if props.Placeholder != "" {
		ta.Placeholder = props.Placeholder
	}

	// Set initial value
	if props.Value != "" {
		ta.SetValue(props.Value)
	}

	// Set prompt
	if props.Prompt != "" {
		ta.Prompt = props.Prompt
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to textarea
	ta.FocusedStyle.Text = ta.FocusedStyle.Text.Inherit(style)
	ta.BlurredStyle.Text = ta.BlurredStyle.Text.Inherit(style)

	// Set dimensions
	if props.Width > 0 {
		ta.SetWidth(props.Width)
	} else if width > 0 {
		ta.SetWidth(width)
	}

	if props.Height > 0 {
		ta.SetHeight(props.Height)
	}

	if props.MaxHeight > 0 {
		ta.MaxHeight = props.MaxHeight
	}

	// Set other properties
	ta.ShowLineNumbers = props.ShowLineNumbers
	if props.CharLimit > 0 {
		ta.CharLimit = props.CharLimit
	}

	// Disable if needed
	if props.Disabled {
		ta.Blur()
	} else {
		ta.Focus()
	}

	return ta.View()
}

// ParseTextareaProps converts a generic props map to TextareaProps using JSON unmarshaling
func ParseTextareaProps(props map[string]interface{}) TextareaProps {
	// Set defaults
	tp := TextareaProps{
		Prompt:          "> ", // Default prompt
		ShowLineNumbers: false,
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &tp)
	}

	return tp
}

// NewTextareaModel creates a new TextareaModel from TextareaProps
func NewTextareaModel(props TextareaProps, id string) TextareaModel {
	ta := textarea.New()

	// Set placeholder
	if props.Placeholder != "" {
		ta.Placeholder = props.Placeholder
	}

	// Set initial value
	if props.Value != "" {
		ta.SetValue(props.Value)
	}

	// Set prompt
	if props.Prompt != "" {
		ta.Prompt = props.Prompt
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to textarea
	ta.FocusedStyle.Text = ta.FocusedStyle.Text.Inherit(style)
	ta.BlurredStyle.Text = ta.BlurredStyle.Text.Inherit(style)

	// Set dimensions
	if props.Width > 0 {
		ta.SetWidth(props.Width)
	}

	if props.Height > 0 {
		ta.SetHeight(props.Height)
	}

	if props.MaxHeight > 0 {
		ta.MaxHeight = props.MaxHeight
	}

	// Set other properties
	ta.ShowLineNumbers = props.ShowLineNumbers
	if props.CharLimit > 0 {
		ta.CharLimit = props.CharLimit
	}

	// Disable if needed
	if props.Disabled {
		ta.Blur()
	} else {
		ta.Focus()
	}

	return TextareaModel{
		Model: ta,
		props: props,
		id:    id,
	}
}

// HandleTextareaUpdate handles updates for textarea components
func HandleTextareaUpdate(msg tea.Msg, textareaModel *TextareaModel) (TextareaModel, tea.Cmd) {
	if textareaModel == nil {
		return TextareaModel{}, nil
	}

	// Check for key press events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Blur the textarea when ESC is pressed
			textareaModel.Blur()
			// Return the updated model and a command to refresh the view
			return *textareaModel, nil
		}
	}

	var cmd tea.Cmd
	textareaModel.Model, cmd = textareaModel.Model.Update(msg)
	return *textareaModel, cmd
}

// Init initializes the textarea model
func (m *TextareaModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the textarea
func (m *TextareaModel) View() string {
	return m.Model.View()
}

// GetID returns the unique identifier for this component instance
func (m *TextareaModel) GetID() string {
	return m.id
}

// SetFocus sets or removes focus from textarea component
func (m *TextareaModel) SetFocus(focus bool) {
	if focus {
		m.Model.Focus()
	} else {
		m.Model.Blur()
	}
}

// TextareaComponentWrapper wraps TextareaModel to implement ComponentInterface properly
type TextareaComponentWrapper struct {
	model *TextareaModel
}

// NewTextareaComponentWrapper creates a wrapper that implements ComponentInterface
func NewTextareaComponentWrapper(textareaModel *TextareaModel) *TextareaComponentWrapper {
	return &TextareaComponentWrapper{
		model: textareaModel,
	}
}

func (w *TextareaComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *TextareaComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case tea.KeyMsg:
		oldValue := w.model.Value()
		var cmds []tea.Cmd

		switch msg.Type {
		case tea.KeyEsc:
			w.model.Blur()
			// Publish focus changed event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputFocusChanged, map[string]interface{}{
				"focused": false,
			}))
			if len(cmds) > 0 {
				return w, tea.Batch(cmds...), core.Ignored
			}
			return w, nil, core.Ignored
		case tea.KeyEnter:
			// Publish enter pressed event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputEnterPressed, map[string]interface{}{
				"value": w.model.Value(),
			}))
			// Let bubble to handleKeyPress for form submission
			if len(cmds) > 0 {
				return w, tea.Batch(cmds...), core.Ignored
			}
			return w, nil, core.Ignored
		case tea.KeyTab:
			// Let bubble to handleKeyPress for navigation
			return w, nil, core.Ignored
		}

		// For other key messages, update the model
		var cmd tea.Cmd
		w.model.Model, cmd = w.model.Model.Update(msg)

		// Check if value changed
		newValue := w.model.Value()
		if oldValue != newValue {
			cmds = append(cmds, cmd)
			// Publish value changed event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
				"oldValue": oldValue,
				"newValue": newValue,
			}))
			return w, tea.Batch(cmds...), core.Handled
		}
		return w, cmd, core.Handled
	}

	// For other messages, update using the underlying model
	oldValue := w.model.Value()
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// Check if value changed
	newValue := w.model.Value()
	if oldValue != newValue {
		// Publish value changed event
		eventCmd := core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
			"oldValue": oldValue,
			"newValue": newValue,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}
	return w, cmd, core.Handled
}

func (w *TextareaComponentWrapper) View() string {
	return w.model.View()
}

func (w *TextareaComponentWrapper) GetID() string {
	return w.model.id
}

// GetValue returns the current value of the textarea component
func (w *TextareaComponentWrapper) GetValue() string {
	return w.model.Value()
}

// SetValue sets the value of the textarea component
func (w *TextareaComponentWrapper) SetValue(value string) {
	w.model.SetValue(value)
}

func (w *TextareaComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
	// Note: We don't publish event here since it would require changing the interface.
	// Events for focus changes are published in the UpdateMsg method for ESC key.
}

func (m *TextareaModel) GetComponentType() string {
	return "textarea"
}

func (w *TextareaComponentWrapper) GetComponentType() string {
	return "textarea"
}

func (m *TextareaModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TextareaModel: invalid data type")
	}

	// Parse textarea properties
	props := ParseTextareaProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (m *TextareaModel) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TextareaModel: invalid data type for UpdateRenderConfig")
	}

	// Parse textarea properties
	props := ParseTextareaProps(propsMap)

	// Update component properties
	m.props = props

	// Update textarea value if provided
	if value, exists := propsMap["value"]; exists {
		if valueStr, ok := value.(string); ok {
			m.SetValue(valueStr)
		}
	}

	return nil
}

// Cleanup 清理资源
func (m *TextareaModel) Cleanup() {
	// TextareaModel 通常不需要特殊清理操作
}

func (w *TextareaComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig 更新渲染配置
func (w *TextareaComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TextareaComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse textarea properties
	props := ParseTextareaProps(propsMap)

	// Update component properties
	w.model.props = props

	// Update textarea value if provided
	if value, exists := propsMap["value"]; exists {
		if valueStr, ok := value.(string); ok {
			w.model.SetValue(valueStr)
		}
	}

	return nil
}

// Cleanup 清理资源
func (w *TextareaComponentWrapper) Cleanup() {
	// 文本区域组件通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (w *TextareaComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Textarea components work like input - sync the current value
	return map[string]interface{}{
		w.GetID(): w.GetValue(),
	}, true
}

