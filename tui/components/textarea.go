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

	// EnterSubmits determines if Enter key submits the form (true) or inserts newline (false)
	EnterSubmits bool `json:"enterSubmits"`
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

	// Configure Enter key behavior
	// If EnterSubmits is true, disable InsertNewline so Enter can submit form (Shift+Enter still works for newline)
	// If EnterSubmits is false (default), Enter inserts newline
	ta.KeyMap.InsertNewline.SetEnabled(!props.EnterSubmits)

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

// HasFocus returns whether the textarea model currently has focus
func (m *TextareaModel) HasFocus() bool {
	return m.Model.Focused()
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
	// Layer 1: Handle targeted messages first
	// 定向消息优先处理，确保消息能正确路由到目标组件
	switch msg := msg.(type) {
	case core.TargetedMsg:
		if msg.TargetID == w.model.id {
			// 递归处理内部消息
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// Layer 2: For KeyMsg, implement layered interception strategy
	// 按键消息采用分层拦截策略
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Layer 2.1: Focus check (MUST keep)
		// 焦点检查：没有焦点的组件不处理按键，让全局绑定生效
		if !w.model.Focused() {
			return w, nil, core.Ignored
		}

		// Layer 2.2: Handle intercepted keys (ESC, Tab, Enter)
		// 处理需要拦截的特殊按键
		switch keyMsg.Type {
		case tea.KeyEsc:
			// ESC: 拦截用于失焦（原始 textarea 不处理 ESC）
			// 失焦并发布焦点变化事件
			oldFocus := w.model.Focused()
			w.model.Blur()
			newFocus := w.model.Focused()

			// 如果焦点确实改变了，发布事件
			if oldFocus != newFocus {
				eventCmd := core.PublishEvent(w.model.id, core.EventInputFocusChanged, map[string]interface{}{
					"focused": newFocus,
				})
				return w, eventCmd, core.Handled
			}
			return w, nil, core.Handled

		case tea.KeyTab:
			// Tab: 拦截用于导航（原始 textarea 不处理 Tab）
			// 返回 Ignored 让上层处理 Tab 导航
			return w, nil, core.Ignored

		case tea.KeyEnter:
			// Enter: 条件拦截用于表单提交
			// 只有当 EnterSubmits=true 时才拦截 Enter，否则让原始 textarea 插入换行
			if w.model.props.EnterSubmits {
				// 发布 Enter 按下事件，返回 Ignored 让上层处理表单提交
				eventCmd := core.PublishEvent(w.model.id, core.EventInputEnterPressed, map[string]interface{}{
					"value": w.model.Value(),
				})
				return w, eventCmd, core.Ignored
			}
			// EnterSubmits=false，fallthrough 让原始 textarea 处理（插入换行）
			fallthrough

		default:
			// Layer 2.3: All other keys - delegate to original textarea
			// 其他所有按键：完全委托给原始 textarea 处理
			// 这保留了所有原生功能：光标移动、文本编辑、复制粘贴等
			oldValue := w.model.Value()
			oldFocus := w.model.Focused()

			// 让原始 textarea 处理这个按键
			var cmd tea.Cmd
			w.model.Model, cmd = w.model.Model.Update(keyMsg)

			// 检测状态变化并发布事件
			newValue := w.model.Value()
			newFocus := w.model.Focused()

			var eventCmds []tea.Cmd

			// 值变化事件
			if oldValue != newValue {
				eventCmds = append(eventCmds, core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
					"oldValue": oldValue,
					"newValue": newValue,
				}))
			}

			// 焦点变化事件（虽然原始 textarea 通常不会自动改变焦点，但保持一致性）
			if oldFocus != newFocus {
				eventCmds = append(eventCmds, core.PublishEvent(w.model.id, core.EventInputFocusChanged, map[string]interface{}{
					"focused": newFocus,
				}))
			}

			// 如果有事件需要发布，批量返回
			if len(eventCmds) > 0 {
				if cmd != nil {
					eventCmds = append([]tea.Cmd{cmd}, eventCmds...)
				}
				return w, tea.Batch(eventCmds...), core.Handled
			}

			// 没有事件，只返回原始命令
			return w, cmd, core.Handled
		}
	}

	// Layer 3: Non-key messages - delegate to original textarea
	// 非按键消息：完全委托给原始 textarea 处理
	oldValue := w.model.Value()
	oldFocus := w.model.Focused()

	// 让原始 textarea 处理所有其他消息
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// 检测状态变化并发布事件
	newValue := w.model.Value()
	newFocus := w.model.Focused()

	var eventCmds []tea.Cmd

	// 值变化事件
	if oldValue != newValue {
		eventCmds = append(eventCmds, core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
			"oldValue": oldValue,
			"newValue": newValue,
		}))
	}

	// 焦点变化事件
	if oldFocus != newFocus {
		eventCmds = append(eventCmds, core.PublishEvent(w.model.id, core.EventInputFocusChanged, map[string]interface{}{
			"focused": newFocus,
		}))
	}

	// 批量返回命令
	if len(eventCmds) > 0 {
		if cmd != nil {
			eventCmds = append([]tea.Cmd{cmd}, eventCmds...)
		}
		return w, tea.Batch(eventCmds...), core.Handled
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

// HasFocus returns whether the textarea component currently has focus
func (w *TextareaComponentWrapper) HasFocus() bool {
	return w.model.Model.Focused()
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

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *TextareaComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}
