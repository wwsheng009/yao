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

// InputModel wraps the textinput.Model to handle TUI integration
type InputModel struct {
	textinput.Model
	props InputProps
	id    string // Unique identifier for this instance
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

// NewInputModel creates a new InputModel from InputProps
func NewInputModel(props InputProps, id string) InputModel {
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
	}

	// Disable if needed
	if props.Disabled {
		input.Blur()
	} else {
		input.Focus()
	}

	return InputModel{
		Model: input,
		props: props,
		id:    id,
	}
}

// HandleInputUpdate handles updates for input components
// This is used when the input is interactive
func HandleInputUpdate(msg tea.Msg, inputModel *InputModel) (InputModel, tea.Cmd) {
	if inputModel == nil {
		return InputModel{}, nil
	}

	// Check for key press events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Blur the input when ESC is pressed
			inputModel.Blur()
			// Return the updated model and a command to refresh the view
			return *inputModel, nil
		}
	}

	var cmd tea.Cmd
	inputModel.Model, cmd = inputModel.Model.Update(msg)
	return *inputModel, cmd
}

// Init initializes the input model
func (m *InputModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the input
func (m *InputModel) View() string {
	return m.Model.View()
}

// GetID returns the unique identifier for this component instance
func (m *InputModel) GetID() string {
	return m.id
}

// GetComponentType returns the component type
func (m *InputModel) GetComponentType() string {
	return "input"
}

func (m *InputModel) Render(config core.RenderConfig) (string, error) {
	// Render should be a pure function - it should not modify internal state
	// All state updates should happen in UpdateRenderConfig
	return m.View(), nil
}

// InputComponentWrapper wraps InputModel to implement ComponentInterface properly
type InputComponentWrapper struct {
	model *InputModel
	bindings []core.ComponentBinding
	stateHelper *core.InputStateHelper
}

// inputValuerAdapter adapts InputModel to satisfy interface{GetValue() string}
type inputValuerAdapter struct {
	*InputModel
}

func (a *inputValuerAdapter) GetValue() string {
	return a.Model.Value()
}

// inputFocuserAdapter adapts InputModel to satisfy interface{Focused() bool}
type inputFocuserAdapter struct {
	*InputModel
}

func (a *inputFocuserAdapter) Focused() bool {
	return a.Model.Focused()
}

// NewInputComponentWrapper creates a wrapper that implements ComponentInterface
func NewInputComponentWrapper(inputModel *InputModel) *InputComponentWrapper {
	wrapper := &InputComponentWrapper{
		model: inputModel,
		bindings: inputModel.props.Bindings,
	}

	// 创建一个适配器来满足接口要求
	valuerAdapter := &inputValuerAdapter{inputModel}
	focuserAdapter := &inputFocuserAdapter{inputModel}
	wrapper.stateHelper = &core.InputStateHelper{
		Valuer:      valuerAdapter,
		Focuser:     focuserAdapter,
		ComponentID: inputModel.id,
	}
	return wrapper
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
	return w.model.id
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
			SourceID:  w.model.id,
			Timestamp: time.Now(),
		}
	}
}

// inputComponentWrapperAdapter adapts InputComponentWrapper to implement core.ComponentWrapper interface
type inputComponentWrapperAdapter struct {
	*InputComponentWrapper
}

func (a *inputComponentWrapperAdapter) GetModel() interface{} {
	return a.InputComponentWrapper.model
}

func (a *inputComponentWrapperAdapter) GetID() string {
	return a.InputComponentWrapper.model.id
}

func (a *inputComponentWrapperAdapter) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

func (a *inputComponentWrapperAdapter) ExecuteAction(action *core.Action) tea.Cmd {
	return a.InputComponentWrapper.ExecuteAction(action)
}


func (w *InputComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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

func (w *InputComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *InputComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// 使用通用绑定处理函数
	wrapper := &inputComponentWrapperAdapter{w}
	cmd, response, handled := core.HandleBinding(wrapper, keyMsg, binding)
	return cmd, response, handled
}

func (w *InputComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	
	// 如果是Enter键，处理后发布事件
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEnter {
		w.model.Model, cmd = w.model.Model.Update(msg)
		
		// 发布Enter按下事件
		enterCmd := core.PublishEvent(w.model.id, core.EventInputEnterPressed, map[string]interface{}{
			"value": w.model.Model.Value(),
		})
		
		// 如果原始命令存在，批处理两个命令
		if cmd != nil {
			return tea.Batch(enterCmd, cmd)
		}
		return enterCmd
	}
	
	w.model.Model, cmd = w.model.Model.Update(msg)
	return cmd
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
		w.model.Model.Blur()
		cmd := core.PublishEvent(w.model.id, core.EventEscapePressed, nil)
		return cmd, core.Ignored, true
	}

	// 其他按键不由这个函数处理
	return nil, core.Ignored, false
}

func (w *InputComponentWrapper) View() string {
	return w.model.View()
}

// GetValue returns the current value of the input component
func (w *InputComponentWrapper) GetValue() string {
	return w.model.Value()
}

// SetValue sets the value of the input component
func (w *InputComponentWrapper) SetValue(value string) {
	w.model.SetValue(value)
}

// SetFocus sets or removes focus from input component
func (m *InputModel) SetFocus(focus bool) {
	if focus {
		m.Model.Focus()
	} else {
		m.Model.Blur()
	}
}

func (w *InputComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
	// Note: We don't publish event here since it would require changing the interface.
	// Events for focus changes are published in the UpdateMsg method for ESC key.
}

// HasFocus returns whether the input component currently has focus
func (w *InputComponentWrapper) HasFocus() bool {
	return w.model.Model.Focused()
}

func (w *InputComponentWrapper) GetComponentType() string {
	return "input"
}

func (w *InputComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
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
	w.model.props = props

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
