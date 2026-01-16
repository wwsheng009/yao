package components

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
func NewInputModel(props InputProps) InputModel {
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
