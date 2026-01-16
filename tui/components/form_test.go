package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFormProps(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected FormProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: FormProps{
				SubmitLabel: "Submit",
				CancelLabel: "Cancel",
				AutoFocus:   true,
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"title":       "My Form",
				"description": "Form description",
				"fields": []interface{}{
					map[string]interface{}{
						"type":        "input",
						"name":        "username",
						"label":       "Username",
						"placeholder": "Enter username",
						"required":    true,
					},
				},
				"submitLabel": "Save",
				"cancelLabel": "Discard",
				"autoFocus":   false,
				"width":       60,
				"height":      20,
			},
			expected: FormProps{
				Title:       "My Form",
				Description: "Form description",
				Fields: []Field{
					{
						Type:        "input",
						Name:        "username",
						Label:       "Username",
						Placeholder: "Enter username",
						Required:    true,
					},
				},
				SubmitLabel: "Save",
				CancelLabel: "Discard",
				AutoFocus:   false,
				Width:       60,
				Height:      20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFormProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderForm(t *testing.T) {
	props := FormProps{
		Title:       "Test Form",
		Description: "A test form",
		Fields: []Field{
			{
				Type:        "input",
				Name:        "name",
				Label:       "Name",
				Placeholder: "Enter your name",
				Required:    true,
			},
		},
		SubmitLabel: "Submit",
		CancelLabel: "Cancel",
	}

	result := RenderForm(props, 80)
	assert.NotEmpty(t, result)
}