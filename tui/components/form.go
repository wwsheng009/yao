package components

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	props FormProps
	focusIndex int
	Values map[string]string
	Errors map[string]string
	Validated bool
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
