package components

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

func TestChatComponentWrapper_UpdateMsg_KeyEnter(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
		EnableMarkdown:   false,
	}
	wrapper := NewChatComponentWrapper(props, "test-chat")

	// Test Enter key with empty input
	msg := tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}}
	_, _, response := wrapper.UpdateMsg(msg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}

	// Set some input
	wrapper.SetValue("Hello world")
	if wrapper.GetValue() != "Hello world" {
		t.Errorf("Expected input value 'Hello world', got '%s'", wrapper.GetValue())
	}

	// Test Enter key with non-empty input
	msg = tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}}
	comp, cmd, response := wrapper.UpdateMsg(msg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	// Check that message was added
	messages := wrapper.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != "Hello world" {
		t.Errorf("Expected user message with 'Hello world', got role='%s', content='%s'",
			messages[0].Role, messages[0].Content)
	}

	// Check that input was cleared
	if wrapper.GetValue() != "" {
		t.Errorf("Expected input to be cleared, got '%s'", wrapper.GetValue())
	}

	// Verify wrapper is returned correctly
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}
}

func TestChatComponentWrapper_UpdateMsg_TargetedMsg(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
	}
	wrapper := NewChatComponentWrapper(props, "test-chat")

	// Test targeted message to this component
	innerMsg := tea.KeyMsg{Type: tea.KeyEnter}
	targetedMsg := core.TargetedMsg{
		TargetID: "test-chat",
		InnerMsg: innerMsg,
	}
	comp, _, response := wrapper.UpdateMsg(targetedMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}

	// Test targeted message to different component
	targetedMsg = core.TargetedMsg{
		TargetID: "other-component",
		InnerMsg: innerMsg,
	}
	comp, _, response = wrapper.UpdateMsg(targetedMsg)
	if response != core.Ignored {
		t.Errorf("Expected Ignored response, got %v", response)
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}
}

func TestChatComponentWrapper_AddMessage(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
	}
	chatModel := NewChatModel(props, "test-chat")

	// Add user message
	chatModel.AddMessage("user", "Hello")

	// Add assistant message
	chatModel.AddMessage("assistant", "Hi there!")

	messages := chatModel.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Verify first message
	if messages[0].Role != "user" || messages[0].Content != "Hello" {
		t.Errorf("First message incorrect: role='%s', content='%s'",
			messages[0].Role, messages[0].Content)
	}

	// Verify second message
	if messages[1].Role != "assistant" || messages[1].Content != "Hi there!" {
		t.Errorf("Second message incorrect: role='%s', content='%s'",
			messages[1].Role, messages[1].Content)
	}
}

func TestChatComponentWrapper_ClearMessages(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
	}
	chatModel := NewChatModel(props, "test-chat")

	// Add some messages
	chatModel.AddMessage("user", "Message 1")
	chatModel.AddMessage("assistant", "Message 2")

	// Clear messages
	chatModel.ClearMessages()

	messages := chatModel.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(messages))
	}
}

func TestChatComponentWrapper_SetFocus(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
	}
	wrapper := NewChatComponentWrapper(props, "test-chat")

	// Set focus
	wrapper.SetFocus(true)
	if !wrapper.TextInput.Focused() {
		t.Error("Expected TextInput to be focused")
	}

	// Remove focus
	wrapper.SetFocus(false)
	if wrapper.TextInput.Focused() {
		t.Error("Expected TextInput to be blurred")
	}
}

func TestChatComponentWrapper_ActionMsg(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
	}
	wrapper := NewChatComponentWrapper(props, "test-chat")

	// Test EventChatMessageReceived action
	actionMsg := core.ActionMsg{
		ID:     "test-chat",
		Action: core.EventChatMessageReceived,
		Data: map[string]interface{}{
			"role":    "assistant",
			"content": "Hello from assistant",
		},
	}
	_, _, response := wrapper.UpdateMsg(actionMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}

	// Verify message was added
	messages := wrapper.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != "assistant" || messages[0].Content != "Hello from assistant" {
		t.Errorf("Message incorrect: role='%s', content='%s'",
			messages[0].Role, messages[0].Content)
	}
}

func TestNewChatModel(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
		EnableMarkdown:   true,
		GlamourStyle:    "dark",
		InputHeight:     3,
		Width:          80,
		Height:         20,
	}
	chatModel := NewChatModel(props, "test-chat")

	// Verify properties
	if chatModel.props.InputPlaceholder != "Type a message..." {
		t.Errorf("Expected input placeholder 'Type a message...', got '%s'",
			chatModel.props.InputPlaceholder)
	}
	if chatModel.id != "test-chat" {
		t.Errorf("Expected id 'test-chat', got '%s'", chatModel.id)
	}

	// Verify input is focused
	if !chatModel.TextInput.Focused() {
		t.Error("Expected TextInput to be focused")
	}

	// Verify placeholder is set
	if chatModel.TextInput.Placeholder != "Type a message..." {
		t.Errorf("Expected placeholder 'Type a message...', got '%s'",
			chatModel.TextInput.Placeholder)
	}
}

func TestChatModel_GetID(t *testing.T) {
	props := ChatProps{ShowInput: true}
	chatModel := NewChatModel(props, "test-id-123")

	if chatModel.GetID() != "test-id-123" {
		t.Errorf("Expected id 'test-id-123', got '%s'", chatModel.GetID())
	}
}

func TestChatModel_Init(t *testing.T) {
	props := ChatProps{ShowInput: true}
	chatModel := NewChatModel(props, "test-chat")

	cmd := chatModel.Init()
	if cmd != nil {
		t.Error("Expected nil command from Init")
	}
}

func TestChatModel_View(t *testing.T) {
	props := ChatProps{
		InputPlaceholder: "Type a message...",
		ShowInput:       true,
	}
	chatModel := NewChatModel(props, "test-chat")
	chatModel.AddMessage("user", "Hello")

	view := chatModel.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestChatModel_MessageTimestamp(t *testing.T) {
	props := ChatProps{ShowInput: true}
	chatModel := NewChatModel(props, "test-chat")

	before := time.Now()
	chatModel.AddMessage("user", "Test message")
	after := time.Now()

	messages := chatModel.GetMessages()
	if len(messages) != 1 {
		t.Fatal("Expected 1 message")
	}

	if messages[0].Timestamp.Before(before) || messages[0].Timestamp.After(after) {
		t.Error("Message timestamp should be between before and after")
	}
}
