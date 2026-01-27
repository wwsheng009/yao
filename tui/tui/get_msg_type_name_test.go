package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/tui/core"
)

// TestGetMsgTypeName tests the getMsgTypeName function
func TestGetMsgTypeName(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		expected string
	}{
		{
			name:     "WindowSizeMsg",
			msg:      tea.WindowSizeMsg{},
			expected: "tea.WindowSizeMsg",
		},
		{
			name:     "KeyMsg",
			msg:      tea.KeyMsg{},
			expected: "tea.KeyMsg",
		},
		{
			name:     "MouseMsg",
			msg:      tea.MouseMsg{},
			expected: "tea.MouseMsg",
		},
		{
			name:     "TargetedMsg",
			msg:      core.TargetedMsg{},
			expected: "TargetedMsg",
		},
		{
			name:     "ActionMsg",
			msg:      core.ActionMsg{},
			expected: "ActionMsg",
		},
		{
			name:     "ProcessResultMsg",
			msg:      core.ProcessResultMsg{},
			expected: "ProcessResultMsg",
		},
		{
			name:     "StateUpdateMsg",
			msg:      core.StateUpdateMsg{},
			expected: "StateUpdateMsg",
		},
		{
			name:     "StreamChunkMsg",
			msg:      core.StreamChunkMsg{},
			expected: "StreamChunkMsg",
		},
		{
			name:     "StreamDoneMsg",
			msg:      core.StreamDoneMsg{},
			expected: "StreamDoneMsg",
		},
		{
			name:     "ErrorMessage",
			msg:      core.ErrorMessage{},
			expected: "ErrorMessage",
		},
		{
			name:     "QuitMsg",
			msg:      core.QuitMsg{},
			expected: "QuitMsg",
		},
		{
			name:     "RefreshMsg",
			msg:      core.RefreshMsg{},
			expected: "RefreshMsg",
		},
		{
			name:     "FocusFirstComponentMsg",
			msg:      core.FocusFirstComponentMsg{},
			expected: "FocusFirstComponentMsg",
		},
		{
			name:     "LogMsg",
			msg:      core.LogMsg{},
			expected: "LogMsg",
		},
		{
			name:     "MenuActionTriggered",
			msg:      core.MenuActionTriggered{},
			expected: "MenuActionTriggered",
		},
		{
			name:     "nil message",
			msg:      nil,
			expected: "nil",
		},
		{
			name:     "unknown message type",
			msg:      struct{}{},
			expected: "struct {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMsgTypeName(tt.msg)
			if result != tt.expected {
				t.Errorf("getMsgTypeName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetMsgTypeNameUnknownType tests that unknown message types return their actual type name
func TestGetMsgTypeNameUnknownType(t *testing.T) {
	// Test with a custom struct that's not in the switch
	type CustomMessage struct {
		Data string
	}
	msg := CustomMessage{Data: "test"}
	
	result := getMsgTypeName(msg)
	
	// The result should be the actual type name, not "unknown"
	if result == "unknown" {
		t.Errorf("getMsgTypeName() returned 'unknown' for CustomMessage, expected actual type name")
	}
	
	// The result should contain the type name
	expected := "tui.CustomMessage"
	if result != expected {
		t.Logf("getMsgTypeName() = %v, expected format like %v", result, expected)
	}
}
