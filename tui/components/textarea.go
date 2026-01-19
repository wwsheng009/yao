package components

import (
	"encoding/json"
	"fmt"
	"time"

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
	
	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
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
		// Do not call Focus() here, it should be handled by Init() method
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
		// Do not call Focus() here, it should be handled by Init() method
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
	// Only focus if not disabled
	if !m.props.Disabled {
		m.Model.Focus()
	}
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

func (m *TextareaModel) GetFocus() bool {
	return m.Model.Focused()
}

// HasFocus returns whether the textarea model currently has focus
func (m *TextareaModel) HasFocus() bool {
	return m.Model.Focused()
}



// GetModel returns the underlying model
func (w *TextareaComponentWrapper) GetModel() interface{} {
	return w.model
}

// PublishEvent creates and returns a command to publish an event
func (w *TextareaComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *TextareaComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For textarea component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}



// TextareaStateHelper implements StateCapturable interface for textarea components
type TextareaStateHelper struct {
	component *TextareaComponentWrapper
}

// CaptureState implements StateCapturable interface
func (h *TextareaStateHelper) CaptureState() map[string]interface{} {
	if h.component == nil {
		return map[string]interface{}{}
	}
	
	return map[string]interface{}{
		"value":     h.component.model.Value(),
		"focused":   h.component.model.Focused(),
		"disabled":  h.component.props.Disabled,
		"component": "textarea",
	}
}

// DetectStateChanges implements StateCapturable interface
func (h *TextareaStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd
	
	if old == nil || new == nil {
		return cmds
	}
	
	// 检查值变化
	if oldValue, oldOk := old["value"].(string); oldOk {
		if newValue, newOk := new["value"].(string); newOk && oldValue != newValue {
			cmds = append(cmds, core.PublishEvent(h.component.GetID(), core.EventInputValueChanged, map[string]interface{}{
				"oldValue": oldValue,
				"newValue": newValue,
			}))
		}
	}
	
	// 检查焦点变化
	if oldFocused, oldOk := old["focused"].(bool); oldOk {
		if newFocused, newOk := new["focused"].(bool); newOk && oldFocused != newFocused {
			cmds = append(cmds, core.PublishEvent(h.component.GetID(), core.EventInputFocusChanged, map[string]interface{}{
				"focused": newFocused,
			}))
		}
	}
	
	return cmds
}

// TextareaComponentWrapper wraps the native textarea.Model to implement ComponentInterface properly
type TextareaComponentWrapper struct {
	model       textarea.Model  // Directly use the native model
	props       TextareaProps // Component properties
	id          string        // Component ID
	bindings    []core.ComponentBinding
	stateHelper *TextareaStateHelper
}

// NewTextareaComponentWrapper creates a wrapper that implements ComponentInterface
// This is the unified entry point that accepts props and id, creating the model internally
func NewTextareaComponentWrapper(props TextareaProps, id string) *TextareaComponentWrapper {
	// Directly create textarea.Model
	ta := textarea.New()

	// Apply configuration directly to the native component
	applyTextareaConfigDirect(&ta, props)

	// Create wrapper that directly implements all interfaces
	wrapper := &TextareaComponentWrapper{
		model:    ta,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	// stateHelper uses wrapper itself as the implementation
	wrapper.stateHelper = &TextareaStateHelper{
		component: wrapper,
	}

	return wrapper
}

func (w *TextareaComponentWrapper) Init() tea.Cmd {
	// 如果组件未被禁用，则返回Focus命令以启动光标闪烁
	if !w.props.Disabled {
		return w.SetFocusWithCmd()
	}
	return nil
}

// SetFocusWithCmd sets focus and returns the command for cursor blinking
func (w *TextareaComponentWrapper) SetFocusWithCmd() tea.Cmd {
	w.model.Focus()
	// Note: textarea.Focus() does not return a BlinkCmd like textinput does
	// This method exists for interface consistency with InputComponentWrapper
	return nil
}

func (w *TextareaComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,                           // 实现了 InteractiveBehavior 接口的组件
		msg,                         // 接收的消息
		w.getBindings,              // 获取按键绑定的函数
		w.handleBinding,            // 处理按键绑定的函数
		w.delegateToBubbles,        // 委托给原 bubbles 组件的函数
	)

	return w, cmd, response
}

// 实现 InteractiveBehavior 接口的方法

func (w *TextareaComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *TextareaComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// TextareaComponentWrapper 已经实现了 core.ComponentWrapper 接口，可以直接传递
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *TextareaComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// 如果是Enter键且EnterSubmits为true，则发布Enter按下事件
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEnter && w.props.EnterSubmits {
		w.model, cmd = w.model.Update(msg)
		
		// 发布Enter按下事件
		enterCmd := core.PublishEvent(w.id, core.EventInputEnterPressed, map[string]interface{}{
			"value": w.model.Value(),
		})
		
		// 如果原始命令存在，批处理两个命令
		if cmd != nil {
			return tea.Batch(enterCmd, cmd)
		}
		return enterCmd
	}
	
	w.model, cmd = w.model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *TextareaComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *TextareaComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *TextareaComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyTab:
		// 让Tab键冒泡以处理组件导航
		return nil, core.Ignored, true
	case tea.KeyEscape:
		// 失焦处理
		w.model.Blur()
		cmd := core.PublishEvent(w.id, core.EventEscapePressed, nil)
		return cmd, core.Ignored, true
	}

	// 其他按键不由这个函数处理
	return nil, core.Ignored, false
}

// HasFocus returns whether the component currently has focus
func (w *TextareaComponentWrapper) HasFocus() bool {
	return w.model.Focused()
}

func (w *TextareaComponentWrapper) View() string {
	return w.model.View()
}

func (w *TextareaComponentWrapper) GetID() string {
	return w.id
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
	if focus {
		w.model.Focus()
	} else {
		w.model.Blur()
	}
	// Note: We don't publish event here since it would require changing the interface.
	// Events for focus changes are published in the UpdateMsg method for ESC key.
}

func (w *TextareaComponentWrapper) GetFocus() bool {
	return w.model.Focused()
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
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TextareaComponentWrapper: invalid data type")
	}

	// Parse textarea properties
	props := ParseTextareaProps(propsMap)

	// Update component properties
	w.props = props

	// Update textarea value if provided
	if value, exists := propsMap["value"]; exists {
		if valueStr, ok := value.(string); ok {
			w.model.SetValue(valueStr)
		}
	}

	// Return the view
	return w.View(), nil
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
	w.props = props

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

// applyTextareaConfigDirect applies TextareaProps configuration to textarea.Model
func applyTextareaConfigDirect(textarea *textarea.Model, props TextareaProps) {
	// Set placeholder
	if props.Placeholder != "" {
		textarea.Placeholder = props.Placeholder
	}

	// Set initial value
	if props.Value != "" {
		textarea.SetValue(props.Value)
	}

	// Set prompt
	if props.Prompt != "" {
		textarea.Prompt = props.Prompt
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
	textarea.FocusedStyle.Text = textarea.FocusedStyle.Text.Inherit(style)
	textarea.BlurredStyle.Text = textarea.BlurredStyle.Text.Inherit(style)

	// Set dimensions
	if props.Width > 0 {
		textarea.SetWidth(props.Width)
	}

	if props.Height > 0 {
		textarea.SetHeight(props.Height)
	}

	if props.MaxHeight > 0 {
		textarea.MaxHeight = props.MaxHeight
	}

	// Set other properties
	textarea.ShowLineNumbers = props.ShowLineNumbers
	if props.CharLimit > 0 {
		textarea.CharLimit = props.CharLimit
	}

	// Configure Enter key behavior
	// If EnterSubmits is true, disable InsertNewline so Enter can submit form (Shift+Enter still works for newline)
	// If EnterSubmits is false (default), Enter inserts newline
	textarea.KeyMap.InsertNewline.SetEnabled(!props.EnterSubmits)

	// Disable if needed
	if props.Disabled {
		textarea.Blur()
	} 
}
