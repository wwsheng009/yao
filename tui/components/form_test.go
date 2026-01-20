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

// TestFormPropsWithMultipleFields tests form with multiple field types
func TestFormPropsWithMultipleFields(t *testing.T) {
	props := map[string]interface{}{
		"title":       "User Registration",
		"description": "Create a new account",
		"fields": []interface{}{
			map[string]interface{}{
				"type":        "input",
				"name":        "username",
				"label":       "Username",
				"placeholder": "Enter username",
				"required":    true,
			},
			map[string]interface{}{
				"type":        "input",
				"name":        "email",
				"label":       "Email",
				"placeholder": "Enter email",
				"required":    true,
			},
			map[string]interface{}{
				"type":        "textarea",
				"name":        "bio",
				"label":       "Bio",
				"placeholder": "Tell us about yourself",
				"required":    false,
			},
		},
		"submitLabel": "Create Account",
		"cancelLabel": "Cancel",
	}

	result := ParseFormProps(props)
	assert.Equal(t, "User Registration", result.Title)
	assert.Equal(t, "Create a new account", result.Description)
	assert.Len(t, result.Fields, 3)
	assert.Equal(t, "input", result.Fields[0].Type)
	assert.Equal(t, "textarea", result.Fields[2].Type)
	assert.True(t, result.Fields[0].Required)
	assert.False(t, result.Fields[2].Required)
}

// TestFormPropsWithValidationRules tests form with validation rules
func TestFormPropsWithValidationRules(t *testing.T) {
	props := map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{
				"type":       "input",
				"name":       "password",
				"label":      "Password",
				"required":   true,
				"validation": map[string]interface{}{"minLength": 8},
			},
			map[string]interface{}{
				"type":       "input",
				"name":       "age",
				"label":      "Age",
				"required":   true,
				"validation": map[string]interface{}{"min": 18, "max": 120},
			},
		},
	}

	result := ParseFormProps(props)
	assert.Len(t, result.Fields, 2)
	assert.Equal(t, "password", result.Fields[0].Name)
	assert.Equal(t, "age", result.Fields[1].Name)
}

// TestFormRenderWithOptions tests form rendering with different options
func TestFormRenderWithOptions(t *testing.T) {
	props := FormProps{
		Title:       "Settings Form",
		Description: "Configure your preferences",
		Fields: []Field{
			{
				Type:        "input",
				Name:        "theme",
				Label:       "Theme",
				Placeholder: "light or dark",
				Required:    true,
			},
		},
		SubmitLabel: "Save",
		CancelLabel: "Discard",
		Width:       80,
		Height:      20,
	}

	result := RenderForm(props, 80)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Settings Form")
	assert.Contains(t, result, "Save")
	assert.Contains(t, result, "Discard")
}

// TestFormWithEmptyFields tests form with no fields
func TestFormWithEmptyFields(t *testing.T) {
	props := FormProps{
		Title:       "Empty Form",
		Description: "A form with no fields",
		Fields:      []Field{},
		SubmitLabel: "Submit",
		CancelLabel: "Cancel",
	}

	result := RenderForm(props, 50)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Empty Form")
}

// TestFormWithDefaultValues tests form with default field values
func TestFormWithDefaultValues(t *testing.T) {
	props := map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{
				"type":     "input",
				"name":     "username",
				"label":    "Username",
				"value":    "john_doe",
				"required": true,
			},
			map[string]interface{}{
				"type":     "input",
				"name":     "email",
				"label":    "Email",
				"value":    "john@example.com",
				"required": true,
			},
		},
	}

	result := ParseFormProps(props)
	assert.Len(t, result.Fields, 2)
	assert.Equal(t, "john_doe", result.Fields[0].Value)
	assert.Equal(t, "john@example.com", result.Fields[1].Value)
}
