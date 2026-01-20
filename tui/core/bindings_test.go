package core

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// mockComponentWrapper implements ComponentWrapper interface for testing
type mockComponentWrapper struct {
	model   interface{}
	id      string
	events  []ActionMsg
	actions []ExecuteActionMsg
}

func (m *mockComponentWrapper) ExecuteAction(action *Action) tea.Cmd {
	return func() tea.Msg {
		return ExecuteActionMsg{
			Action:    action,
			SourceID:  m.id,
			Timestamp: time.Now(),
		}
	}
}

func (m *mockComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{
			ID:     sourceID,
			Action: eventName,
			Data:   payload,
		}
	}
}

func (m *mockComponentWrapper) GetModel() interface{} {
	return m.model
}

func (m *mockComponentWrapper) GetID() string {
	return m.id
}

func TestCheckComponentBindings_MatchKey(t *testing.T) {
	bindings := []ComponentBinding{
		{Key: "ctrl+c", Event: "copy", Enabled: true},
		{Key: "enter", UseDefault: true, Enabled: true},
		{Key: "f1", Event: "help", Enabled: false},
	}

	// Test matching enabled binding
	ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	matched, binding, handled := CheckComponentBindings(ctrlCMsg, bindings, "test")
	assert.True(t, matched)
	assert.True(t, handled)
	assert.Equal(t, "copy", binding.Event)

	// Test matching UseDefault binding
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	matched, binding, handled = CheckComponentBindings(enterMsg, bindings, "test")
	assert.True(t, matched)
	assert.False(t, handled) // UseDefault should not be handled
	assert.True(t, binding.UseDefault)

	// Test disabled binding (should not match)
	f1Msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}, Alt: false}
	matched, binding, handled = CheckComponentBindings(f1Msg, bindings, "test")
	assert.False(t, matched)
	assert.Nil(t, binding)
	assert.False(t, handled)
}

func TestCheckComponentBindings_NoMatch(t *testing.T) {
	bindings := []ComponentBinding{
		{Key: "ctrl+c", Event: "copy", Enabled: true},
	}

	// Test non-matching key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	matched, binding, handled := CheckComponentBindings(escMsg, bindings, "test")
	assert.False(t, matched)
	assert.Nil(t, binding)
	assert.False(t, handled)
}

func TestHandleBinding_ActionMode(t *testing.T) {
	action := &Action{
		Process: "test.process",
		Args:    []interface{}{"test_value"},
	}
	binding := ComponentBinding{
		Key:    "ctrl+s",
		Action: action,
	}

	wrapper := &mockComponentWrapper{
		id: "test-id",
	}

	cmd, response, handled := HandleBinding(wrapper, tea.KeyMsg{Type: tea.KeyCtrlS}, binding)
	assert.True(t, handled)
	assert.Equal(t, Handled, response)
	assert.NotNil(t, cmd)

	// Execute command and verify result
	msg := cmd()
	actionMsg, ok := msg.(ExecuteActionMsg)
	assert.True(t, ok)
	assert.Equal(t, "test.process", actionMsg.Action.Process)
	assert.Equal(t, "test-id", actionMsg.SourceID)
}

func TestHandleBinding_EventMode(t *testing.T) {
	binding := ComponentBinding{
		Key:   "f1",
		Event: "show_help",
	}

	wrapper := &mockComponentWrapper{
		id: "test-id",
	}

	cmd, response, handled := HandleBinding(wrapper, tea.KeyMsg{Type: tea.KeyF1}, binding)
	assert.True(t, handled)
	assert.Equal(t, Handled, response)
	assert.NotNil(t, cmd)

	// Execute command and verify result
	msg := cmd()
	eventMsg, ok := msg.(ActionMsg)
	assert.True(t, ok)
	assert.Equal(t, "show_help", eventMsg.Action)
	assert.Equal(t, "test-id", eventMsg.ID)
}

func TestHandleBinding_UseDefaultMode(t *testing.T) {
	// For UseDefault mode, the binding shouldn't have Action or Event
	// This mode is handled at the application level by skipping HandleBinding
	// when UseDefault is true. Here we simulate a binding that has neither
	// Action nor Event, which should return Ignored.
	
	binding := ComponentBinding{
		Key:        "tab",
		// No Action, no Event, UseDefault would be handled externally
	}

	wrapper := &mockComponentWrapper{
		id: "test-id",
	}

	cmd, response, handled := HandleBinding(wrapper, tea.KeyMsg{Type: tea.KeyTab}, binding)
	assert.Nil(t, cmd)
	assert.Equal(t, Ignored, response)
	assert.False(t, handled)
}

func TestCheckComponentBindings_KeyMatching(t *testing.T) {
	bindings := []ComponentBinding{
		{Key: "ctrl+c", Event: "copy", Enabled: true},
		{Key: "enter", Event: "submit", Enabled: true},
		{Key: "esc", Event: "cancel", Enabled: true},
		{Key: "a", Event: "action_a", Enabled: true},
	}

	tests := []struct {
		name            string
		keyMsg          tea.KeyMsg
		expectedEvent   string
		shouldMatch     bool
		shouldBeHandled bool
	}{
		{
			name:            "Ctrl+C matches",
			keyMsg:          tea.KeyMsg{Type: tea.KeyCtrlC},
			expectedEvent:   "copy",
			shouldMatch:     true,
			shouldBeHandled: true,
		},
		{
			name:            "Enter matches",
			keyMsg:          tea.KeyMsg{Type: tea.KeyEnter},
			expectedEvent:   "submit",
			shouldMatch:     true,
			shouldBeHandled: true,
		},
		{
			name:            "ESC matches",
			keyMsg:          tea.KeyMsg{Type: tea.KeyEsc},
			expectedEvent:   "cancel",
			shouldMatch:     true,
			shouldBeHandled: true,
		},
		{
			name:            "Letter 'a' matches",
			keyMsg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expectedEvent:   "action_a",
			shouldMatch:     true,
			shouldBeHandled: true,
		},
		{
			name:            "Non-bound key doesn't match",
			keyMsg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}},
			shouldMatch:     false,
			shouldBeHandled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, binding, handled := CheckComponentBindings(tt.keyMsg, bindings, "test")
			assert.Equal(t, tt.shouldMatch, matched)
			if tt.shouldMatch {
				assert.Equal(t, tt.expectedEvent, binding.Event)
				assert.Equal(t, tt.shouldBeHandled, handled)
			} else {
				assert.Nil(t, binding)
				assert.False(t, handled)
			}
		})
	}
}