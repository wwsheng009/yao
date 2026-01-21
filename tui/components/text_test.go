package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

func TestRenderText(t *testing.T) {
	props := TextProps{
		Content: "Hello World",
		Color:   "red",
		Bold:    true,
	}

	result := RenderText(props, 20, 5)
	if result == "" {
		t.Error("RenderText should return non-empty string")
	}
}

func TestParseTextProps(t *testing.T) {
	props := map[string]interface{}{
		"content": "Test content",
		"color":   "blue",
		"bold":    true,
	}

	textProps := ParseTextProps(props)
	if textProps.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", textProps.Content)
	}
	if textProps.Color != "blue" {
		t.Errorf("Expected color 'blue', got '%s'", textProps.Color)
	}
	if !textProps.Bold {
		t.Error("Expected bold to be true")
	}
}

func TestTextModel_UpdateMsg(t *testing.T) {
	model := &TextModel{
		Props: TextProps{Content: "Test"},
		Width: 20,
		id:    "text",
	}

	// Test targeted message
	targetedMsg := core.TargetedMsg{
		TargetID: "text",
		InnerMsg: nil,
	}

	updatedModel, cmd, response := model.UpdateMsg(targetedMsg)
	if response != core.Handled {
		t.Error("Expected TargetedMsg to be handled")
	}
	if cmd != nil {
		t.Error("Expected no command")
	}
	if updatedModel != model {
		t.Error("Expected same model")
	}

	// Test non-targeted message
	updatedModel, cmd, response = model.UpdateMsg(tea.KeyMsg{})
	if response != core.Ignored {
		t.Error("Expected non-targeted message to be ignored")
	}
}

func TestTextModel_View(t *testing.T) {
	model := &TextModel{
		Props: TextProps{Content: "Test View"},
		Width: 30,
	}

	result := model.View()
	if result == "" {
		t.Error("View should return non-empty string")
	}
}

func TestTextModel_GetID(t *testing.T) {
	model := &TextModel{id: "text"}
	if model.GetID() != "text" {
		t.Errorf("Expected ID 'text', got '%s'", model.GetID())
	}
}

func TestTextModel_SetFocus(t *testing.T) {
	model := &TextModel{}
	// Should not panic
	model.SetFocus(true)
	model.SetFocus(false)
}
