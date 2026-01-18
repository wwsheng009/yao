package components

import (
	"encoding/json"
	"fmt"

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
}

// NewInputComponentWrapper creates a wrapper that implements ComponentInterface
func NewInputComponentWrapper(inputModel *InputModel) *InputComponentWrapper {
	return &InputComponentWrapper{
		model: inputModel,
	}
}

func (w *InputComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *InputComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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

func (w *InputComponentWrapper) View() string {
	return w.model.View()
}

func (w *InputComponentWrapper) GetID() string {
	return w.model.id
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


