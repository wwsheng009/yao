package components

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// InputProps defines the properties for the Input component
type InputProps struct {
	// Placeholder is the placeholder text when the input is empty
	Placeholder string `json:"placeholder"`

	// Value is the initial value of the input
	Value string `json:"value"`

	// Prompt is the prompt character/string before the input
	Prompt string `json:"prompt"`

	// Color specifies the text color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Width specifies the input width (0 for auto)
	Width int `json:"width"`

	// Height specifies the input height (0 for auto)
	Height int `json:"height"`

	// Disabled determines if the input is disabled
	Disabled bool `json:"disabled"`

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}



// RenderInput renders an input component
func RenderInput(props InputProps, width int) string {
	input := textinput.New()

	// Set placeholder
	if props.Placeholder != "" {
		input.Placeholder = props.Placeholder
	}

	// Set initial value
	if props.Value != "" {
		input.SetValue(props.Value)
	}

	// Set prompt
	if props.Prompt != "" {
		input.Prompt = props.Prompt
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to input
	input.TextStyle = style
	input.PlaceholderStyle = style

	// Set width if specified
	if props.Width > 0 {
		input.Width = props.Width
	} else if width > 0 {
		// Adjust for available width, accounting for prompt length
		promptLen := len(input.Prompt)
		availableWidth := width - promptLen - 2 // Subtract 2 for padding
		if availableWidth > 0 {
			input.Width = availableWidth
		}
	}

	// Disable if needed
	if props.Disabled {
		input.Blur()
	} else {
		input.Focus()
	}

	return input.View()
}

// ParseInputProps converts a generic props map to InputProps using JSON unmarshaling
func ParseInputProps(props map[string]interface{}) InputProps {
	// Set defaults
	ip := InputProps{
		Prompt: "> ", // Default prompt
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &ip)
	}

	return ip
}

// applyTextInputConfig applies InputProps configuration to textinput.Model
func applyTextInputConfig(input *textinput.Model, props InputProps) {
	// Set placeholder
	if props.Placeholder != "" {
		input.Placeholder = props.Placeholder
	}

	// Set initial value
	if props.Value != "" {
		input.SetValue(props.Value)
	}

	// Set prompt
	if props.Prompt != "" {
		input.Prompt = props.Prompt
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to input
	input.TextStyle = style
	input.PlaceholderStyle = style

	// Set width if specified
	if props.Width > 0 {
		input.Width = props.Width
	}

	// Disable if needed
	if props.Disabled {
		input.Blur()
	} else {
		input.Focus()
	}
}

// InputComponentWrapper directly implements ComponentInterface by wrapping textinput.Model
type InputComponentWrapper struct {
	model       textinput.Model // 直接使用原生组件
	props       InputProps      // 组件属性
	id          string          // 组件ID
	bindings    []core.ComponentBinding
	stateHelper *core.InputStateHelper
}

// NewInputComponentWrapper creates a wrapper that implements ComponentInterface
// This is the unified entry point that accepts props and id, creating the model internally
func NewInputComponentWrapper(props InputProps, id string) *InputComponentWrapper {
	// Directly create textinput.Model
	input := textinput.New()

	// Apply configuration directly to the native component
	applyTextInputConfig(&input, props)

	// Create wrapper that directly implements all interfaces
	wrapper := &InputComponentWrapper{
		model:    input,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	// stateHelper uses wrapper itself as the implementation
	wrapper.stateHelper = &core.InputStateHelper{
		Valuer:      wrapper, // wrapper implements Valuer interface
		Focuser:     wrapper, // wrapper implements Focuser interface
		ComponentID: id,
	}

	return wrapper
}

// InputComponentWrapper directly implements core.Valuer interface
func (w *InputComponentWrapper) GetValue() string {
	return w.model.Value()
}

// InputComponentWrapper directly implements core.Focuser interface
func (w *InputComponentWrapper) Focused() bool {
	return w.model.Focused()
}

func (w *InputComponentWrapper) SetFocus(focus bool) {
	if focus {
		w.model.Focus()
	} else {
		w.model.Blur()
	}
}

func (w *InputComponentWrapper) Init() tea.Cmd {
	return nil
}

// GetModel returns the underlying model
func (w *InputComponentWrapper) GetModel() interface{} {
	return w.model
}

// GetID returns the component ID
func (w *InputComponentWrapper) GetID() string {
	return w.id
}

// PublishEvent creates and returns a command to publish an event
func (w *InputComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *InputComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For input component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}

func (w *InputComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,                   // 实现了 InteractiveBehavior 接口的组件
		msg,                 // 接收的消息
		w.getBindings,       // 获取按键绑定的函数
		w.handleBinding,     // 处理按键绑定的函数
		w.delegateToBubbles, // 委托给原 bubbles 组件的函数
	)

	return w, cmd, response
}

// 实现 InteractiveBehavior 接口的方法

func (w *InputComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *InputComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// InputComponentWrapper 已经实现了 ComponentWrapper 接口，可以直接传递
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *InputComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// 处理按键消息
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// 跳过已在 HandleSpecialKey 中处理的键
		switch keyMsg.Type {
		case tea.KeyTab, tea.KeyEscape:
			// 这些键已由 HandleSpecialKey 处理，跳过委托
			return nil
		case tea.KeyEnter:
			// 特殊处理Enter键
			w.model, cmd = w.model.Update(msg)

			// 发布Enter按下事件
			enterCmd := core.PublishEvent(w.id, core.EventInputEnterPressed, map[string]interface{}{
				"value": w.model.Value(),
			})

			// 合并命令（如果有的话）
			if cmd != nil {
				return tea.Batch(enterCmd, cmd)
			}
			return enterCmd
		default:
			// 处理其他按键（包括字符输入）
			w.model, cmd = w.model.Update(msg)
			return cmd // 可能为 nil
		}
	}

	// 处理非按键消息
	w.model, cmd = w.model.Update(msg)
	return cmd // 可能为 nil
}

// 实现 StateCapturable 接口
func (w *InputComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *InputComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *InputComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
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

func (w *InputComponentWrapper) View() string {
	return w.model.View()
}

// SetValue sets the value of the input component
func (w *InputComponentWrapper) SetValue(value string) {
	w.model.SetValue(value)
}

// HasFocus returns whether the input component currently has focus
func (w *InputComponentWrapper) HasFocus() bool {
	return w.model.Focused()
}

func (w *InputComponentWrapper) GetComponentType() string {
	return "input"
}

func (w *InputComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *InputComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("InputComponentWrapper: invalid data type")
	}

	// Parse input properties
	props := ParseInputProps(propsMap)

	// Update component properties
	w.props = props

	// Apply configuration to the model
	applyTextInputConfig(&w.model, props)

	// Update underlying model if value changed
	if props.Value != "" && w.model.Value() != props.Value {
		w.model.SetValue(props.Value)
	}

	return nil
}

// Cleanup cleans up resources used by the input component
func (w *InputComponentWrapper) Cleanup() {
	// Input components typically don't need cleanup
	// This is a no-op for input components
}

// GetStateChanges returns the state changes from this component
func (w *InputComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// For input components, we always sync the current value
	return map[string]interface{}{
		w.GetID(): w.GetValue(),
	}, true
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *InputComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}
