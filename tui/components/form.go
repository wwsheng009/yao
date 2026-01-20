package components

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// Ensure lipgloss is used (referenced by lipglossStyleWrapper)
var _ = lipgloss.NewStyle()

// Field defines a form field
type Field struct {
	// Type is the field type (input, textarea, checkbox, etc.)
	Type string `json:"type"`

	// Name is the field name/key
	Name string `json:"name"`

	// Label is the field label
	Label string `json:"label"`

	// Placeholder is the placeholder text
	Placeholder string `json:"placeholder"`

	// Value is the field value
	Value string `json:"value"`

	// Required indicates if the field is required
	Required bool `json:"required"`

	// Validation is the validation rule
	Validation string `json:"validation"`

	// Options are the options for select/radio fields
	Options []string `json:"options"`
}

// FormProps defines the properties for the Form component
type FormProps struct {
	// Fields defines the form fields
	Fields []Field `json:"fields"`

	// Title is the form title
	Title string `json:"title"`

	// Description is the form description
	Description string `json:"description"`

	// SubmitLabel is the submit button label
	SubmitLabel string `json:"submitLabel"`

	// CancelLabel is the cancel button label
	CancelLabel string `json:"cancelLabel"`

	// AutoFocus determines if the first field should be auto-focused
	AutoFocus bool `json:"autoFocus"`

	// Width specifies the form width (0 for auto)
	Width int `json:"width"`

	// Height specifies the form height (0 for auto)
	Height int `json:"height"`

	// Style is the general form style
	Style lipglossStyleWrapper `json:"style"`

	// FieldStyle is the style for form fields
	FieldStyle lipglossStyleWrapper `json:"fieldStyle"`

	// LabelStyle is the style for labels
	LabelStyle lipglossStyleWrapper `json:"labelStyle"`

	// ErrorStyle is the style for error messages
	ErrorStyle lipglossStyleWrapper `json:"errorStyle"`

	// ButtonStyle is the style for buttons
	ButtonStyle lipglossStyleWrapper `json:"buttonStyle"`

	// CursorMode specifies the cursor mode for all input fields
	CursorMode string `json:"cursorMode"`

	// CursorChar specifies the cursor character for all input fields
	CursorChar string `json:"cursorChar"`

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// FormModel represents a form model for interactive forms
type FormModel struct {
	props      FormProps
	focusIndex int
	Values     map[string]string
	Errors     map[string]string
	Validated  bool
	focused    bool   // Whether the form has focus
	id         string // Unique identifier for this instance
	ID         string // For component interface
}

// FormStateHelper 表单组件状态捕获助手
type FormStateHelper struct {
	Valuer      interface{ GetValue() string }
	Focuser     interface{ Focused() bool }
	ComponentID string
}

func (h *FormStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"value":   h.Valuer.GetValue(),
		"focused": h.Focuser.Focused(),
	}
}

func (h *FormStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["value"] != new["value"] {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventInputValueChanged, map[string]interface{}{
			"oldValue": old["value"],
			"newValue": new["value"],
		}))
	}

	if old["focused"] != new["focused"] {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventInputFocusChanged, map[string]interface{}{
			"focused": new["focused"],
		}))
	}

	return cmds
}

// RenderForm renders a form component
func RenderForm(props FormProps, width int) string {
	var sb strings.Builder

	// Add title if exists
	if props.Title != "" {
		titleStyle := props.Style.GetStyle().Copy().Bold(true)
		sb.WriteString(titleStyle.Render(props.Title))
		sb.WriteString("\n")
	}

	// Add description if exists
	if props.Description != "" {
		descStyle := props.Style.GetStyle()
		sb.WriteString(descStyle.Render(props.Description))
		sb.WriteString("\n")
	}

	// Render each field
	for i, field := range props.Fields {
		// Add spacing between fields
		if i > 0 {
			sb.WriteString("\n")
		}

		// Render label
		if field.Label != "" {
			labelStyle := props.LabelStyle.GetStyle()
			sb.WriteString(labelStyle.Render(field.Label))
			if field.Required {
				sb.WriteString(" *")
			}
			sb.WriteString(": ")
		}

		// Render field based on type
		switch field.Type {
		case "input", "text":
			// For static rendering, show placeholder or value
			fieldStyle := props.FieldStyle.GetStyle()
			placeholder := field.Placeholder
			if field.Value != "" {
				placeholder = field.Value
			}
			sb.WriteString(fieldStyle.Render(placeholder))
		case "textarea":
			// For static rendering, show placeholder or value
			fieldStyle := props.FieldStyle.GetStyle()
			placeholder := field.Placeholder
			if field.Value != "" {
				placeholder = field.Value
			}
			sb.WriteString(fieldStyle.Render(placeholder))
		case "checkbox":
			// For static rendering, show checkbox state
			fieldStyle := props.FieldStyle.GetStyle()
			if field.Value == "true" || field.Value == "1" {
				sb.WriteString(fieldStyle.Render("[x]"))
			} else {
				sb.WriteString(fieldStyle.Render("[ ]"))
			}
			sb.WriteString(" " + field.Label)
		case "select", "radio":
			// For static rendering, show selected value
			fieldStyle := props.FieldStyle.GetStyle()
			if field.Value != "" {
				sb.WriteString(fieldStyle.Render(field.Value))
			} else {
				sb.WriteString(fieldStyle.Render(field.Placeholder))
			}
		default:
			// Default to input-like field
			fieldStyle := props.FieldStyle.GetStyle()
			placeholder := field.Placeholder
			if field.Value != "" {
				placeholder = field.Value
			}
			sb.WriteString(fieldStyle.Render(placeholder))
		}
	}

	// Add submit/cancel buttons
	sb.WriteString("\n\n")
	buttonStyle := props.ButtonStyle.GetStyle()
	buttons := fmt.Sprintf(
		"[ %s ] [ %s ]",
		props.SubmitLabel,
		props.CancelLabel,
	)
	sb.WriteString(buttonStyle.Render(buttons))

	return sb.String()
}

// ParseFormProps converts a generic props map to FormProps using JSON unmarshaling
func ParseFormProps(props map[string]interface{}) FormProps {
	// Set defaults
	fp := FormProps{
		SubmitLabel: "Submit",
		CancelLabel: "Cancel",
		AutoFocus:   true,
	}

	// Handle Fields separately as it needs special processing
	if fields, ok := props["fields"].([]interface{}); ok {
		fp.Fields = make([]Field, 0, len(fields))
		for _, fieldIntf := range fields {
			if fieldMap, ok := fieldIntf.(map[string]interface{}); ok {
				// Create a temporary map without options for JSON unmarshal
				fieldCopy := make(map[string]interface{})
				for k, v := range fieldMap {
					if k != "options" {
						fieldCopy[k] = v
					}
				}

				field := Field{}

				// Unmarshal field properties
				if dataBytes, err := json.Marshal(fieldCopy); err == nil {
					_ = json.Unmarshal(dataBytes, &field)
				}

				// Handle options separately
				if options, ok := fieldMap["options"].([]interface{}); ok {
					field.Options = make([]string, len(options))
					for j, opt := range options {
						if optStr, ok := opt.(string); ok {
							field.Options[j] = optStr
						}
					}
				}

				fp.Fields = append(fp.Fields, field)
			}
		}
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &fp)
	}

	return fp
}

// HandleFormUpdate handles updates for form components
// This is used when the form is interactive
func HandleFormUpdate(msg tea.Msg, formModel *FormModel) (FormModel, tea.Cmd) {
	if formModel == nil {
		return FormModel{}, nil
	}

	// In a real implementation, this would handle form interactions
	// For now, just return the model unchanged
	return *formModel, nil
}

// NewFormModel creates a new FormModel from FormProps
func NewFormModel(props FormProps, id string) FormModel {
	return FormModel{
		props:      props,
		focusIndex: 0,
		Values:     make(map[string]string),
		Errors:     make(map[string]string),
		Validated:  false,
		focused:    false,
		id:         id,
		ID:         id,
	}
}

// Init initializes the form model
func (m *FormModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the form
func (m *FormModel) View() string {
	return RenderForm(m.props, 80)
}

// GetID returns the unique identifier for this component instance
func (m *FormModel) GetID() string {
	return m.id
}

// UpdateMsg implements ComponentInterface for form component
func (m *FormModel) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle key press events for navigation and submission
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If form doesn't have focus, ignore all key messages
		// This allows global bindings (like 'q' for quit) and Tab navigation to work
		if !m.focused {
			return m, nil, core.Ignored
		}

		switch msg.Type {
		case tea.KeyEnter:
			// Form submission - bubble to parent to handle
			return m, nil, core.Ignored
		case tea.KeyEsc:
			// Cancel form - bubble to parent
			return m, nil, core.Ignored
		case tea.KeyTab:
			// Let Tab bubble to handleKeyPress for component navigation
			return m, nil, core.Ignored
		case tea.KeyUp:
			// Navigate to previous field
			if m.focusIndex > 0 {
				m.focusIndex--
			}
			return m, nil, core.Handled
		case tea.KeyDown:
			// Navigate to next field
			if m.focusIndex < len(m.props.Fields)-1 {
				m.focusIndex++
			}
			return m, nil, core.Handled
		}
	}

	return m, nil, core.Ignored
}

// GetModel returns the underlying model
func (w *FormComponentWrapper) GetModel() interface{} {
	return w.model
}

// GetID returns the component ID
func (w *FormComponentWrapper) GetID() string {
	return w.id
}

// PublishEvent creates and returns a command to publish an event
func (w *FormComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *FormComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For form component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.model.id,
			Timestamp: time.Now(),
		}
	}
}

// FormComponentWrapper wraps FormModel to implement ComponentInterface properly
type FormComponentWrapper struct {
	model       FormModel
	inputFields map[string]*InputComponentWrapper
	props       FormProps
	bindings    []core.ComponentBinding
	stateHelper *FormStateHelper
	id          string
}

// NewFormComponentWrapper creates a wrapper that implements ComponentInterface
func NewFormComponentWrapper(props FormProps, id string) *FormComponentWrapper {
	formModel := NewFormModel(props, id)
	formModel.ID = id

	wrapper := &FormComponentWrapper{
		model:       formModel,
		inputFields: make(map[string]*InputComponentWrapper),
		props:       props,
		bindings:    props.Bindings,
		id:          id,
	}

	wrapper.stateHelper = &FormStateHelper{
		Valuer:      wrapper,
		Focuser:     wrapper,
		ComponentID: id,
	}

	return wrapper
}

func (w *FormComponentWrapper) Init() tea.Cmd {
	// 不要在初始化时自动获取焦点
	// 焦点应该通过框架的焦点管理机制来控制
	// 只有当组件被明确设置焦点时才获取焦点
	// 注意：我们仍需要调用子字段的Init方法来初始化它们的状态，
	// 但不应该让它们在初始化时获取焦点
	for _, field := range w.inputFields {
		if field != nil {
			// 调用子字段的Init方法进行初始化，但忽略返回的命令
			field.Init()
		}
	}

	return nil
}

func (w *FormComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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
func (w *FormComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *FormComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// FormComponentWrapper 已经实现了 ComponentWrapper 接口，可以直接传递
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *FormComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle key press events for navigation and submission
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If form doesn't have focus, ignore all key messages
		// This allows global bindings (like 'q' for quit) and Tab navigation to work
		if !w.model.focused {
			return nil
		}

		switch msg.Type {
		case tea.KeyEnter:
			// Form submission - bubble to parent to handle
			return nil
		case tea.KeyEsc:
			// Cancel form - bubble to parent
			return nil
		case tea.KeyTab:
			// Let Tab bubble to handleKeyPress for component navigation
			return nil
		case tea.KeyUp:
			// Navigate to previous field
			if w.model.focusIndex > 0 {
				w.model.focusIndex--
			}
			return nil
		case tea.KeyDown:
			// Navigate to next field
			if w.model.focusIndex < len(w.model.props.Fields)-1 {
				w.model.focusIndex++
			}
			return nil
		}
	}

	return cmd
}

// 实现 StateCapturable 接口
func (w *FormComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *FormComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *FormComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// ESC 和 Tab 现在由框架层统一处理，这里不处理
	// 如果有其他特殊的键处理需求，可以在这里添加
	return nil, core.Ignored, false
}

func (w *FormComponentWrapper) View() string {
	return RenderForm(w.props, w.props.Width)
}

// SetFocus sets or removes focus from form component
func (m *FormModel) SetFocus(focus bool) {
	m.focused = focus
	// When form gets focus, ensure focusIndex is valid
	if focus && len(m.props.Fields) > 0 {
		if m.focusIndex < 0 || m.focusIndex >= len(m.props.Fields) {
			m.focusIndex = 0
		}
	}
}

func (m *FormModel) GetFocus() bool {
	return m.focused
}

func (w *FormComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
}

func (w *FormComponentWrapper) GetFocus() bool {
	return w.model.focused
}

// GetValue returns the current value of the form component
func (w *FormComponentWrapper) GetValue() string {
	return w.model.GetValue()
}

// Focused returns whether the form is focused
func (w *FormComponentWrapper) Focused() bool {
	return w.model.Focused()
}

func (m *FormModel) GetComponentType() string {
	return "form"
}

func (m *FormModel) UpdateRenderConfig(config core.RenderConfig) error {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("FormModel: invalid data type")
	}

	// Parse form properties
	props := ParseFormProps(propsMap)

	// Update component properties
	m.props = props

	return nil
}

func (m *FormModel) Cleanup() {
	// Form模型通常不需要清理资源
	// 这是一个空操作
}

// GetStateChanges returns the state changes from this component
func (m *FormModel) GetStateChanges() (map[string]interface{}, bool) {
	// Form component doesn't have state changes at the model level
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (m *FormModel) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
	}
}

func (w *FormComponentWrapper) GetComponentType() string {
	return "form"
}

func (m *FormModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("FormModel: invalid data type")
	}

	// Parse form properties
	props := ParseFormProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

func (w *FormComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("FormComponentWrapper: invalid data type")
	}

	// Parse form properties
	props := ParseFormProps(propsMap)

	// Update component properties
	w.props = props
	w.model.props = props

	// Return rendered view
	return w.View(), nil
}

func (w *FormComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("FormComponentWrapper: invalid data type")
	}

	// Parse form properties
	props := ParseFormProps(propsMap)

	// Update component properties
	w.props = props
	w.model.props = props

	return nil
}

func (w *FormComponentWrapper) Cleanup() {
	// Form组件通常不需要清理资源
	// 这是一个空操作
}

// GetStateChanges returns the state changes from this component
func (w *FormComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Form component collects values from child components
	// For now, return nil as form values are collected differently
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *FormComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}

// GetValue returns the current value of the form component
func (m *FormModel) GetValue() string {
	// Return a string representation of the form values
	var sb strings.Builder
	for key, value := range m.Values {
		sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	return sb.String()
}

// Focused returns whether the form is focused
func (m *FormModel) Focused() bool {
	return m.focused
}

// SetCursorMode sets the cursor mode for all input fields in the form
func (w *FormComponentWrapper) SetCursorMode(mode string) {
	w.props.CursorMode = mode
	for _, field := range w.inputFields {
		if field != nil {
			field.SetCursorMode(mode)
		}
	}
}

// SetCursorChar sets the cursor character for all input fields in the form
func (w *FormComponentWrapper) SetCursorChar(char string) {
	w.props.CursorChar = char
	for _, field := range w.inputFields {
		if field != nil {
			field.SetCursorChar(char)
		}
	}
}

// RegisterInputField registers an input field with the form wrapper
func (w *FormComponentWrapper) RegisterInputField(name string, field *InputComponentWrapper) {
	w.inputFields[name] = field

	if w.props.CursorMode != "" {
		field.SetCursorMode(w.props.CursorMode)
	}
	if w.props.CursorChar != "" {
		field.SetCursorChar(w.props.CursorChar)
	}
}

// GetInputField returns an input field by name
func (w *FormComponentWrapper) GetInputField(name string) (*InputComponentWrapper, bool) {
	field, ok := w.inputFields[name]
	return field, ok
}
