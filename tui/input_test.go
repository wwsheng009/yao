package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/components"
)

func TestInputComponent(t *testing.T) {
	cfg := &Config{
		Name: "Input Test",
		Data: map[string]interface{}{
			"username": "",
			"email":    "",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "username-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter username",
						"prompt":      "> ",
					},
				},
				{
					ID:   "email-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter email",
						"prompt":      "> ",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Test that input models are created when components are rendered
	view1 := model.RenderComponent(&cfg.Layout.Children[0])
	view2 := model.RenderComponent(&cfg.Layout.Children[1])
	assert.NotEmpty(t, view1)
	assert.NotEmpty(t, view2)

	// Check that input models were created
	assert.Contains(t, model.InputModels, "username-input")
	assert.Contains(t, model.InputModels, "email-input")

	// Test input value update
	usernameInput := model.InputModels["username-input"]
	usernameInput.Model.SetValue("testuser")
	
	// Simulate updating state with input value
	model.StateMu.Lock()
	model.State["username-input"] = usernameInput.Value()
	model.StateMu.Unlock()

	// Verify state was updated
	value, exists := model.getStateValue("username-input")
	assert.True(t, exists)
	assert.Equal(t, "testuser", value)
}

func TestInputNavigation(t *testing.T) {
	cfg := &Config{
		Name: "Input Navigation Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "first-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "First input",
						"prompt":      "> ",
					},
				},
				{
					ID:   "second-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Second input",
						"prompt":      "> ",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Manually render the components to initialize input models
	for i := range cfg.Layout.Children {
		model.RenderComponent(&cfg.Layout.Children[i])
	}

	// Initially, first input should be focused
	assert.Equal(t, "first-input", model.CurrentFocus)

	// Simulate tabbing to next input
	model.focusNextInput()
	assert.Equal(t, "second-input", model.CurrentFocus)

	// Tab again should wrap to first
	model.focusNextInput()
	assert.Equal(t, "first-input", model.CurrentFocus)
}

func TestHandleInputUpdate(t *testing.T) {
	props := components.InputProps{
		Placeholder: "Test input",
		Prompt:      "> ",
	}
	
	inputModel := components.NewInputModel(props)
	
	// Test typing 'hello'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updatedModel, _ := components.HandleInputUpdate(msg, &inputModel)
	
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedModel, _ = components.HandleInputUpdate(msg, &updatedModel)
	
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updatedModel, _ = components.HandleInputUpdate(msg, &updatedModel)
	
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updatedModel, _ = components.HandleInputUpdate(msg, &updatedModel)
	
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	updatedModel, _ = components.HandleInputUpdate(msg, &updatedModel)

	assert.Equal(t, "hello", updatedModel.Value())
}