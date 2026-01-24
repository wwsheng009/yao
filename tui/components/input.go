package components

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/runtime"
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

	// CursorMode specifies the cursor mode: "blink", "static", "hide"
	CursorMode string `json:"cursorMode"`

	// CursorChar specifies the cursor character
	CursorChar string `json:"cursorChar"`

	// CursorBlinkSpeed specifies the cursor blink speed in milliseconds
	CursorBlinkSpeed int `json:"cursorBlinkSpeed"`

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
		// Do not call Focus() here, it should be handled by Init() method
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
	// Save current focus state
	wasFocused := input.Focused()

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

	// Handle disabled state without changing focus
	if props.Disabled {
		input.Blur() // Only blur if the component is currently focused and needs to be disabled
		// But we'll restore focus state later
	}

	// Restore the original focus state to avoid interfering with component state
	if wasFocused && !props.Disabled {
		input.Focus()
	} else if !wasFocused {
		input.Blur()
	}
}

// InputComponentWrapper directly implements ComponentInterface by wrapping textinput.Model
type InputComponentWrapper struct {
	model        textinput.Model
	cursorHelper *CursorHelper
	props        InputProps
	id           string
	bindings     []core.ComponentBinding
	stateHelper  *core.InputStateHelper
}

// NewInputComponentWrapper creates a wrapper that implements ComponentInterface
// This is the unified entry point that accepts props and id, creating the model internally
func NewInputComponentWrapper(props InputProps, id string) *InputComponentWrapper {
	// Directly create textinput.Model
	input := textinput.New()

	// Apply configuration directly to the native component
	applyTextInputConfig(&input, props)

	// Create cursor helper for managing cursor behavior
	blinkSpeed := 530 * time.Millisecond
	if props.CursorBlinkSpeed > 0 {
		blinkSpeed = time.Duration(props.CursorBlinkSpeed) * time.Millisecond
	}

	cursorConfig := CursorConfig{
		Mode:       ParseCursorMode(props.CursorMode),
		Char:       props.CursorChar,
		BlinkSpeed: blinkSpeed,
		Visible:    !props.Disabled,
	}

	// Set cursor mode and character on the input model
	if props.CursorMode != "" {
		input.Cursor.SetMode(ParseCursorMode(props.CursorMode))
	}
	if props.CursorChar != "" {
		input.Cursor.SetChar(props.CursorChar)
	}

	// Set blink speed if specified
	if props.CursorBlinkSpeed > 0 {
		input.Cursor.BlinkSpeed = blinkSpeed
	}

	// Create wrapper that directly implements all interfaces
	wrapper := &InputComponentWrapper{
		model:        input,
		cursorHelper: NewCursorHelper(cursorConfig),
		props:        props,
		id:           id,
		bindings:     props.Bindings,
	}

	// stateHelper uses wrapper itself as the implementation
	wrapper.stateHelper = &core.InputStateHelper{
		Valuer:      wrapper,
		Focuser:     wrapper,
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
	currentFocus := w.model.Focused()
	if focus != currentFocus {
		if focus {
			w.model.Focus()
			w.cursorHelper.SetVisible(true)
		} else {
			w.model.Blur()
			w.cursorHelper.SetVisible(false)
		}
	}
}

func (w *InputComponentWrapper) GetFocus() bool {
	return w.model.Focused()
}

// SetFocusWithCmd applies focus and returns the command that starts blinking
func (w *InputComponentWrapper) SetFocusWithCmd(focus bool) tea.Cmd {
	currentFocus := w.model.Focused()
	if focus != currentFocus {
		if focus {
			return w.model.Focus()
		}
		w.model.Blur()
		w.cursorHelper.SetVisible(false)
	}
	return nil
}

func (w *InputComponentWrapper) Init() tea.Cmd {
	// 不要在初始化时自动获取焦点
	// 焦点应该通过框架的焦点管理机制来控制
	// 只有当组件被明确设置焦点时才获取焦点
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
	// 非按键消息
	w.model, cmd = w.model.Update(msg)
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
	// 特殊处理 Enter 键
	if keyMsg.Type == tea.KeyEnter {
		// 更新模型状态
		w.model, _ = w.model.Update(keyMsg)

		// 发布 Enter 按下事件
		enterCmd := core.PublishEvent(w.id, core.EventInputEnterPressed, map[string]interface{}{
			"value": w.model.Value(),
		})

		// 返回已处理，阻止消息继续传递
		return enterCmd, core.Handled, true
	}

	// ESC 键：组件自己处理失去焦点
	if keyMsg.Type == tea.KeyEsc && w.model.Focused() {
		// 直接设置焦点状态
		w.SetFocus(false)
		// 发送 FocusLost 消息给外部框架，保持一致性
		cmd := func() tea.Msg {
			return core.TargetedMsg{
				TargetID: w.id,
				InnerMsg: core.FocusMsg{
					Type:   core.FocusLost,
					Reason: "USER_ESC",
					ToID:   "",
				},
			}
		}
		return cmd, core.Handled, true
	}

	// Tab 键：让框架处理焦点切换
	if keyMsg.Type == tea.KeyTab {
		// 返回 Ignored 让框架层处理 Tab 键的焦点切换
		return nil, core.Ignored, false
	}

	// 其他键：不特殊处理
	return nil, core.Ignored, false
}

func (w *InputComponentWrapper) View() string {
	return w.model.View()
}

// SetSize sets the allocated size for the input component.
// This is called by the Runtime before rendering.
func (w *InputComponentWrapper) SetSize(width, height int) {
	// Update the textinput model to use the allocated width
	if width > 0 {
		w.model.Width = width
	}
	// Height is typically 1 for input, but we can store it if needed
	w.props.Width = width
	w.props.Height = height
}

// SetValue sets the value of the input component
func (w *InputComponentWrapper) SetValue(value string) {
	w.model.SetValue(value)
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

	// Temporarily save focus state
	wasFocused := w.model.Focused()

	// Apply configuration to the model
	applyTextInputConfig(&w.model, props)

	// Update underlying model if value changed
	if props.Value != "" && w.model.Value() != props.Value {
		w.model.SetValue(props.Value)
	}

	// Restore the original focus state to avoid interfering with component state
	if wasFocused {
		w.model.Focus()
	} else {
		w.model.Blur()
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
	// This ensures the global state is kept in sync with the component state
	currentValue := w.GetValue()
	key := w.GetID()
	return map[string]interface{}{
		key: currentValue,
	}, true
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *InputComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}

// SetCursorMode sets the cursor mode for the input component
func (w *InputComponentWrapper) SetCursorMode(mode string) {
	w.props.CursorMode = mode
	cursorMode := ParseCursorMode(mode)
	w.model.Cursor.SetMode(cursorMode)
	w.cursorHelper.SetMode(cursorMode)
}

// SetCursorChar sets the cursor character for the input component
func (w *InputComponentWrapper) SetCursorChar(char string) {
	w.props.CursorChar = char
	w.model.Cursor.SetChar(char)
	w.cursorHelper.SetChar(char)
}

// GetCursorHelper returns the cursor helper for this input component
func (w *InputComponentWrapper) GetCursorHelper() *CursorHelper {
	return w.cursorHelper
}

// SetCursorBlinkSpeed sets the cursor blink speed in milliseconds
func (w *InputComponentWrapper) SetCursorBlinkSpeed(speedMs int) {
	w.props.CursorBlinkSpeed = speedMs
	if speedMs > 0 {
		w.model.Cursor.BlinkSpeed = time.Duration(speedMs) * time.Millisecond
		w.cursorHelper.SetBlinkSpeed(time.Duration(speedMs) * time.Millisecond)
	}
}

// Measure implements the runtime.Measurable interface for the Input component.
// Returns the preferred size of the input component given the constraints.
func (w *InputComponentWrapper) Measure(c runtime.BoxConstraints) runtime.Size {
	// Input components typically have a height of 1 (single line)
	height := 1

	// Apply height constraints
	if height < c.MinHeight {
		height = c.MinHeight
	}
	if c.MaxHeight > 0 && height > c.MaxHeight {
		height = c.MaxHeight
	}

	// Calculate width based on props
	width := w.props.Width

	// If width is not specified in props, calculate from content
	if width <= 0 {
		// Get the current value length
		valueLen := lipgloss.Width(w.model.Value())

		// Add prompt length
		promptLen := lipgloss.Width(w.model.Prompt)

		// Add placeholder length if value is empty
		placeholderLen := 0
		if w.model.Value() == "" {
			placeholderLen = lipgloss.Width(w.model.Placeholder)
		}

		// Width is the maximum of value+prompt and placeholder+prompt
		contentWidth := valueLen + promptLen
		if placeholderLen+promptLen > contentWidth {
			contentWidth = placeholderLen + promptLen
		}

		// Add some padding for visual comfort
		width = contentWidth + 2 // +2 for margins

		// Apply minimum width if specified
		if width < 10 {
			width = 10 // Minimum usable width for input
		}
	}

	// Apply width constraints
	if width < c.MinWidth {
		width = c.MinWidth
	}
	if c.MaxWidth > 0 && width > c.MaxWidth {
		width = c.MaxWidth
	}

	return runtime.Size{Width: width, Height: height}
}
