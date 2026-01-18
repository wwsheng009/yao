package components

import (
	"encoding/json"
	"fmt"
	"strings"

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
		switch msg.Type {
		case tea.KeyEnter:
			// Form submission - bubble to parent to handle
			return m, nil, core.Ignored
		case tea.KeyEsc:
			// Cancel form - bubble to parent
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

// FormComponentWrapper wraps FormModel to implement ComponentInterface properly
type FormComponentWrapper struct {
	model *FormModel
}

// NewFormComponentWrapper creates a wrapper that implements ComponentInterface
func NewFormComponentWrapper(formModel *FormModel) *FormComponentWrapper {
	return &FormComponentWrapper{
		model: formModel,
	}
}

func (w *FormComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *FormComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle key press events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Publish FORM_CANCEL event when ESC is pressed
			cancelCmd := core.PublishEvent(
				w.model.id,
				core.EventFormCancel,
				map[string]interface{}{
					"formID": w.model.id,
					"reason": "user_pressed_esc",
				},
			)
			return w, cancelCmd, core.Handled
		case tea.KeyEnter:
			// Let Enter bubble to handleKeyPress for form submission
			return w, nil, core.Ignored
		case tea.KeyTab:
			// Navigate to next field
			if w.model.props.Fields != nil && len(w.model.props.Fields) > 0 {
				w.model.focusIndex = (w.model.focusIndex + 1) % len(w.model.props.Fields)
			}
			return w, nil, core.Handled
		}
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	case core.ActionMsg:
		// Handle internal action messages
		if msg.Action == core.EventFormSubmitSuccess {
			// Reset focus and values after successful submission
			w.model.Values = make(map[string]string)
			w.model.focusIndex = 0
			return w, nil, core.Handled
		}
		if msg.Action == core.EventFormCancel {
			// Clear form on cancel
			w.model.Values = make(map[string]string)
			return w, nil, core.Handled
		}
	}

	// Default: ignore message
	return w, nil, core.Ignored
}

func (w *FormComponentWrapper) View() string {
	return w.model.View()
}

func (w *FormComponentWrapper) GetID() string {
	return w.model.id
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

func (w *FormComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
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
	return w.model.Render(config)
}

func (w *FormComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	// 委托给底层的 FormModel
	return w.model.UpdateRenderConfig(config)
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
