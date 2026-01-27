package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/tui/core"
)

// executeAction creates a command to execute an action.
// This returns a tea.Cmd that will be executed asynchronously.
func (m *Model) executeAction(action *core.Action) tea.Cmd {
	if action == nil {
		return nil
	}

	// Validate action
	if err := action.Validate(); err != nil {
		return func() tea.Msg {
			return core.ProcessResultMsg{
				Data:   nil,
				Target: action.OnError,
				Error:  err,
			}
		}
	}

	// Check if it's a Process or Script action
	if action.Process != "" {
		return m.executeProcessAction(action)
	}

	if action.Script != "" {
		return m.executeScriptAction(action)
	}

	// Direct state update (if payload is present)
	if len(action.Payload) > 0 {
		return func() tea.Msg {
			return core.StateBatchUpdateMsg{
				Updates: action.Payload,
			}
		}
	}

	// If no process, script, or payload, return a success message
	return func() tea.Msg {
		return core.ProcessResultMsg{
			Data:   nil,
			Target: "",
			Error:  nil,
		}
	}
}

// executeProcessAction creates a command to execute a Yao Process.
func (m *Model) executeProcessAction(action *core.Action) tea.Cmd {
	return func() tea.Msg {
		// This will be implemented when we integrate with Yao's Process system
		// For now, return a placeholder
		log.Trace("TUI ExecuteProcess: %s", action.Process)

		result, err := executeProcessAction(m, action)
		if err != nil {
			return core.ProcessResultMsg{
				Data:   nil,
				Target: action.OnError,
				Error:  err,
			}
		}

		return core.ProcessResultMsg{
			Data:   result,
			Target: action.OnSuccess,
			Error:  nil,
		}
	}
}

// executeScriptAction creates a command to execute a script method.
func (m *Model) executeScriptAction(action *core.Action) tea.Cmd {
	return func() tea.Msg {
		// This will be implemented when we add script support
		log.Trace("TUI ExecuteScript: %s.%s", action.Script, action.Method)

		result, err := executeScriptAction(m, action)
		if err != nil {
			log.Error("TUI ExecuteScript error: %v", err)
			return core.ProcessResultMsg{
				Data:   nil,
				Target: action.OnError,
				Error:  err,
			}
		}

		return core.ProcessResultMsg{
			Data:   result,
			Target: action.OnSuccess,
			Error:  nil,
		}
	}
}

// getStateValue safely gets a state value.
// This is used internally by the JavaScript API.
func (m *Model) getStateValue(key string) (interface{}, bool) {
	m.StateMu.RLock()
	defer m.StateMu.RUnlock()

	// Handle nested keys separated by dots (e.g., "user.name")
	keys := strings.Split(key, ".")
	currentValue, exists := m.State[keys[0]]
	if !exists {
		return nil, false
	}

	// Navigate through nested maps
	for i := 1; i < len(keys); i++ {
		if currentMap, ok := currentValue.(map[string]interface{}); ok {
			currentValue, exists = currentMap[keys[i]]
			if !exists {
				return nil, false
			}
		} else {
			// If intermediate value is not a map, return not found
			return nil, false
		}
	}

	return currentValue, true
}

// setStateValue safely sets a state value.
// This is used internally by the JavaScript API.
func (m *Model) setStateValue(key string, value interface{}) {
	m.StateMu.Lock()
	m.State[key] = value
	m.StateMu.Unlock()
	// CRITICAL: Mark for re-render when state changes
	m.forceRender = true
}
