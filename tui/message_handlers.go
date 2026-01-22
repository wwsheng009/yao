package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// GetDefaultMessageHandlersFromCore returns default message handlers for TUI model
func GetDefaultMessageHandlersFromCore() map[string]core.MessageHandler {
	handlers := make(map[string]core.MessageHandler)

	// Register handler for tea.KeyMsg
	handlers["tea.KeyMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		return model.handleKeyPress(msg.(tea.KeyMsg))
	}

	// Register handler for tea.WindowSizeMsg
	handlers["tea.WindowSizeMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		sizeMsg := msg.(tea.WindowSizeMsg)
		model.Width = sizeMsg.Width
		model.Height = sizeMsg.Height
		model.Ready = true

		// Update layout engine with new window size
		if model.LayoutEngine != nil {
			model.LayoutEngine.UpdateWindowSize(sizeMsg.Width, sizeMsg.Height)
		}

		// Update all component instances with new window size
		// This ensures components receive the correct dimensions
		allComponents := model.ComponentInstanceRegistry.GetAll()
		for _, comp := range allComponents {
			updatedConfig := core.RenderConfig{
				Data:   comp.LastConfig.Data,
				Width:  sizeMsg.Width,
				Height: sizeMsg.Height,
			}
			updateComponentInstanceConfig(comp, updatedConfig, comp.ID)
		}

		// Broadcast window size to all components
		// This ensures that nested Bubble Tea models (like list, table) receive the message
		// and can update their internal dimensions/pagination accordingly.
		return model.dispatchMessageToAllComponents(msg)
	}

	// Register handler for core.TargetedMsg
	handlers["TargetedMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		targetedMsg := msg.(core.TargetedMsg)

		// Find target component and dispatch message
		updatedModel, cmd, _ := model.dispatchMessageToComponent(targetedMsg.TargetID, targetedMsg.InnerMsg)
		return updatedModel, cmd
	}

	// Register handler for core.ActionMsg
	handlers["ActionMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		actionMsg := msg.(core.ActionMsg)
		log.Trace("TUI Update: Received ActionMsg: %s from %s", actionMsg.Action, actionMsg.ID)

		// Handle specific system actions
		switch actionMsg.Action {
		case core.EventFocusNext:
			// Move focus to next input component via message-driven approach
			return model, model.focusNextInput()
		case core.EventFocusPrev:
			// Move focus to previous input component via message-driven approach
			return model, model.focusPrevInput()
		case core.EventFocusChanged:
			// DEPRECATED: Focus changes should be driven by FocusMsg only
			// Components no longer publish EventFocusChanged to avoid infinite loops
			// This event is ignored - all focus changes are handled via FocusMsg
			log.Trace("TUI: EventFocusChanged is deprecated and ignored, use FocusMsg instead")
			return model, nil
		default:
			// For other actions, publish to EventBus for component communication
			model.EventBus.Publish(actionMsg)
		}

		return model, nil
	}

	// Register handler for core.ProcessResultMsg
	handlers["ProcessResultMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		return model.handleProcessResult(msg.(core.ProcessResultMsg))
	}

	// Register handler for core.StateUpdateMsg
	handlers["StateUpdateMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		stateMsg := msg.(core.StateUpdateMsg)
		model.StateMu.Lock()
		model.State[stateMsg.Key] = stateMsg.Value
		model.StateMu.Unlock()
		return model, nil
	}

	// Register handler for core.StateBatchUpdateMsg
	handlers["StateBatchUpdateMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		batchMsg := msg.(core.StateBatchUpdateMsg)
		model.StateMu.Lock()
		for key, value := range batchMsg.Updates {
			model.State[key] = value
		}
		model.StateMu.Unlock()
		return model, nil
	}

	// Register handler for core.StreamChunkMsg
	handlers["StreamChunkMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		return model.handleStreamChunk(msg.(core.StreamChunkMsg))
	}

	// Register handler for core.StreamDoneMsg
	handlers["StreamDoneMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		return model.handleStreamDone(msg.(core.StreamDoneMsg))
	}

	// Register handler for core.ErrorMessage
	handlers["ErrorMessage"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		return model.handleError(msg.(core.ErrorMessage))
	}

	// Register handler for core.QuitMsg
	handlers["QuitMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		log.Trace("TUI Update: Received QuitMsg, quitting...")
		return m.(tea.Model), tea.Quit
	}

	// Register handler for core.RefreshMsg
	handlers["RefreshMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		return model, nil
	}

	// Register handler for FocusFirstComponentMsg
	handlers["FocusFirstComponentMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}

		// Only apply auto-focus once during initialization
		if model.AutoFocusApplied {
			return model, nil
		}

		// Get all focusable components
		focusableIDs := model.getFocusableComponentIDs()
		if len(focusableIDs) > 0 && model.Config.AutoFocus != nil && *model.Config.AutoFocus {
			// Set focus to the first focusable component via message-driven approach
			model.AutoFocusApplied = true
			log.Trace("TUI: Auto-focus to first focusable component: %s", focusableIDs[0])
			return model, model.setFocus(focusableIDs[0])
		}

		return model, nil
	}

	// Register handler for core.MenuActionTriggered
	handlers["MenuActionTriggered"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		menuMsg := msg.(core.MenuActionTriggered)

		// Execute the action if it's a process
		if processName, ok := menuMsg.Action["process"].(string); ok {
			action := &core.Action{
				Process: processName,
				Args:    []interface{}{},
				OnError: "__error",
			}
			return model, model.executeAction(action)
		}

		return model, nil
	}

	return handlers
}

// getMsgTypeName returns a string representation of the message type for routing
func getMsgTypeName(msg tea.Msg) string {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		return "tea.WindowSizeMsg"
	case tea.KeyMsg:
		return "tea.KeyMsg"
	case tea.MouseMsg:
		return "tea.MouseMsg"
	case core.TargetedMsg:
		return "TargetedMsg"
	case core.ActionMsg:
		return "ActionMsg"
	case core.ProcessResultMsg:
		return "ProcessResultMsg"
	case core.StateUpdateMsg:
		return "StateUpdateMsg"
	case core.StateBatchUpdateMsg:
		return "StateBatchUpdateMsg"
	case core.InputUpdateMsg:
		return "InputUpdateMsg"
	case core.StreamChunkMsg:
		return "StreamChunkMsg"
	case core.StreamDoneMsg:
		return "StreamDoneMsg"
	case core.ErrorMessage:
		return "ErrorMessage"
	case core.QuitMsg:
		return "QuitMsg"
	case core.RefreshMsg:
		return "RefreshMsg"
	case core.FocusFirstComponentMsg:
		return "FocusFirstComponentMsg"
	case core.LogMsg:
		return "LogMsg"
	case core.MenuActionTriggered:
		return "MenuActionTriggered"
	default:
		// For unknown message types, return the actual type name for better debugging
		// This helps identify messages from components that are not in the switch
		// For example: bubbletea cursor.BlinkMsg, etc.
		if msg == nil {
			return "nil"
		}
		// Use the full type name including package path for maximum clarity
		return fmt.Sprintf("%T", msg)
	}
}

// handleKeyPress processes keyboard input and executes bound actions.
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Capture phase: Global system keys (Priority 1)
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Dispatch phase: Route to focused component (Priority 2)
	componentHandled := false
	if m.CurrentFocus != "" {
		log.Trace("TUI: Dispatching key to focused component: %s", m.CurrentFocus)
		updatedModel, cmd, handled := m.dispatchMessageToComponent(m.CurrentFocus, msg)
		log.Trace("TUI: Component %s returned: handled=%v", m.CurrentFocus, handled)
		if handled {
			// Component handled the message
			componentHandled = true

			// For ESC key, return the component's command which sends FocusMsg
			// Don't interfere with component's internal focus management
			if msg.Type == tea.KeyEsc {
				log.Trace("TUI: ESC key handled by component, executing its command")
				return updatedModel, cmd
			}

			return updatedModel, cmd
		}
		m = updatedModel.(*Model)
	}

	// ESC key handling (Priority 3): No component handled it, use default behavior
	if msg.Type == tea.KeyEsc && !componentHandled {
		log.Trace("TUI: ESC key not handled by component, checking bindings")
		// Check for ESC bindings first
		if m.Config.Bindings != nil {
			key := msg.String()
			if action, ok := m.Config.Bindings[key]; ok {
				actionDesc := action.Process
				if actionDesc == "" && action.Script != "" {
					actionDesc = action.Script + "." + action.Method
				}
				log.Trace("TUI: ESC key has binding, executing action: %s", actionDesc)
				return m.executeBoundAction(&action, key)
			}
		}
		// Default: use clearFocus() which sends FocusMsg via tea.Cmd
		log.Trace("TUI: Using default clearFocus for ESC")
		return m, m.clearFocus()
	}

	// Native navigation keys (Priority 4): Tab/ShiftTab handling
	// Allow Tab/Shift+Tab to work even when no component has focus
	if msg.Type == tea.KeyTab || msg.Type == tea.KeyShiftTab {
		log.Trace("TUI: Navigation key detected (Tab/Shift+Tab), CurrentFocus=%q", m.CurrentFocus)
		return m.handleNativeNavigation(msg)
	}

	// Global bindings (Priority 5): Handle bound actions
	// Execute when:
	// - No component has focus, OR
	// - Component ignored the message (componentHandled == false)
	if !componentHandled {
		log.Trace("TUI: Component did not handle message, checking global bindings, CurrentFocus=%q", m.CurrentFocus)
		return m.handleBoundActions(msg)
	}

	log.Trace("TUI: Message handled by component, no further processing needed")
	return m, nil
}

// handleBoundActions handles bound actions for keys
func (m *Model) handleBoundActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Trace("TUI: handleBoundActions: checking bindings for key=%q", msg.String())
	// Build key string for matching
	key := msg.String()

	// Also try single rune if available (for single character keys like 'a', '+', etc.)
	if len(msg.Runes) == 1 {
		char := string(msg.Runes[0])
		log.Trace("TUI: Single rune key: %q, checking bindings", char)
		// Check for bound actions with the character
		if m.Config.Bindings != nil {
			if action, ok := m.Config.Bindings[char]; ok {
				actionDesc := action.Process
				if actionDesc == "" && action.Script != "" {
					actionDesc = action.Script + "." + action.Method
				}
				log.Trace("TUI: Found binding for single rune: %s -> %s", char, actionDesc)
				return m.executeBoundAction(&action, char)
			}
		} else {
			log.Trace("TUI: No bindings configured")
		}
	}

	// Check for bound actions with full key string
	if m.Config.Bindings != nil {
		if action, ok := m.Config.Bindings[key]; ok {
			actionDesc := action.Process
			if actionDesc == "" && action.Script != "" {
				actionDesc = action.Script + "." + action.Method
			}
			log.Trace("TUI: Found binding for key: %s -> %s", key, actionDesc)
			return m.executeBoundAction(&action, key)
		}
	}

	log.Trace("TUI: No binding found for key: %q", key)
	return m, nil
}

// executeBoundAction executes an action bound to a key
func (m *Model) executeBoundAction(action *core.Action, key string) (tea.Model, tea.Cmd) {
	// If the action is a quit action, execute it
	if action.Process == "tui.quit" || action.Process == "tui.exit" {
		log.Trace("TUI KeyPress: %s -> quit action", key)
		return m, m.executeAction(action)
	} else {
		log.Trace("TUI KeyPress: %s -> action", key)
		return m, m.executeAction(action)
	}
}

// handleNativeNavigation handles Tab/ShiftTab navigation based on NavigationMode
func (m *Model) handleNativeNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Trace("TUI: handleNativeNavigation: key=%v, CurrentFocus=%q", msg.Type, m.CurrentFocus)
	key := msg.String()
	navigationMode := m.Config.NavigationMode
	if navigationMode == "" {
		navigationMode = "native" // Default to native
	}
	log.Trace("TUI: NavigationMode=%s", navigationMode)

	switch msg.Type {
	case tea.KeyTab:
		// Check if there's a binding for Tab (only in bindable mode)
		if navigationMode == "bindable" && m.Config.Bindings != nil {
			if action, ok := m.Config.Bindings[key]; ok {
				return m.executeBoundAction(&action, key)
			}
		}
		// Default: navigate to next component
		return m.handleTabNavigation()

	case tea.KeyShiftTab:
		// Check if there's a binding for Shift+Tab (only in bindable mode)
		if navigationMode == "bindable" && m.Config.Bindings != nil {
			if action, ok := m.Config.Bindings[key]; ok {
				return m.executeBoundAction(&action, key)
			}
		}
		// Default: navigate to previous component
		return m.handleShiftTabNavigation()
	}

	return m, nil
}

// handleTabNavigation handles tab navigation between components
func (m *Model) handleTabNavigation() (tea.Model, tea.Cmd) {
	log.Trace("Tab pressed, cycling focus between components, current focus: %s", m.CurrentFocus)

	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		log.Trace("No focusable components found")
		return m, nil
	}

	if m.CurrentFocus == "" {
		return m, m.setFocus(focusableIDs[0])
	}

	currentIndex := -1
	for i, id := range focusableIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return m, m.setFocus(focusableIDs[0])
	}

	// Check if Tab cycling is enabled
	tabCycles := m.Config.TabCycles
	if !tabCycles {
		// Default to true for backward compatibility
		tabCycles = true
	}

	var nextFocus string
	if tabCycles {
		// Cycling mode: wrap around
		nextIndex := (currentIndex + 1) % len(focusableIDs)
		nextFocus = focusableIDs[nextIndex]
	} else {
		// Non-cycling mode: stop at last component
		if currentIndex < len(focusableIDs)-1 {
			nextFocus = focusableIDs[currentIndex+1]
		} else {
			// Already at last component, don't cycle
			log.Trace("Already at last focusable component, Tab cycling disabled")
			return m, nil
		}
	}

	log.Trace("Focused to next component: %s (index %d, cycles=%v)", nextFocus, currentIndex+1, tabCycles)

	return m, m.setFocus(nextFocus)
}

// handleShiftTabNavigation handles Shift+Tab to focus previous component
func (m *Model) handleShiftTabNavigation() (tea.Model, tea.Cmd) {
	log.Trace("Shift+Tab pressed, moving to previous component, current focus: %s", m.CurrentFocus)
	// focusPrevComponent internally calls setFocus, so we need to get the command from it
	return m, m.focusPrevComponent()
}

// handleProcessResult processes the result from a Yao Process execution.
func (m *Model) handleProcessResult(msg core.ProcessResultMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		// Handle error case
		log.Error("TUI ProcessResult Error: %v", msg.Error)
		if msg.Target != "" {
			m.StateMu.Lock()
			m.State[msg.Target] = msg.Error.Error()
			m.StateMu.Unlock()
		} else {
			// Store error in default error field
			m.StateMu.Lock()
			m.State["__error"] = msg.Error.Error()
			m.StateMu.Unlock()
		}
		// Trigger refresh to display error in UI
		return m, func() tea.Msg { return core.RefreshMsg{} }
	} else {
		// Handle success case
		if msg.Target != "" {
			m.StateMu.Lock()
			m.State[msg.Target] = msg.Data
			m.StateMu.Unlock()
			log.Trace("TUI ProcessResult: %s = %v", msg.Target, msg.Data)

			// Invalidate props cache when state changes
			if m.propsCache != nil {
				m.propsCache.Clear()
				log.Trace("Process result updated state, cleared props cache")
			}

			// Trigger refresh to display new data in UI
			return m, func() tea.Msg { return core.RefreshMsg{} }
		}
	}
	return m, nil
}

// handleStreamChunk handles a streaming data chunk.
func (m *Model) handleStreamChunk(msg core.StreamChunkMsg) (tea.Model, tea.Cmd) {
	m.StateMu.Lock()
	defer m.StateMu.Unlock()

	// Append chunk to the stream buffer
	key := "stream_" + msg.ID
	current, ok := m.State[key]
	if !ok {
		current = ""
	}

	if str, ok := current.(string); ok {
		m.State[key] = str + msg.Content
	} else {
		m.State[key] = msg.Content
	}

	return m, nil
}

// handleStreamDone handles stream completion.
func (m *Model) handleStreamDone(msg core.StreamDoneMsg) (tea.Model, tea.Cmd) {
	log.Trace("TUI StreamDone: %s", msg.ID)
	// Mark stream as complete in state
	m.StateMu.Lock()
	m.State["stream_"+msg.ID+"_done"] = true
	m.StateMu.Unlock()
	return m, nil
}

// handleError handles error messages.
func (m *Model) handleError(msg core.ErrorMessage) (tea.Model, tea.Cmd) {
	if msg.LogLevel == "warn" {
		log.Warn("TUI Warning: %v", msg)
	} else {
		log.Error("TUI Error: %v", msg)
	}

	// Store error in state
	m.StateMu.Lock()
	m.State["__error"] = msg.Error()
	m.StateMu.Unlock()

	return m, nil
}

// unwrapTargetedMsg checks if the message is a TargetedMsg and returns the inner message if the target matches
// Returns (inner_message, is_targeted_msg, should_process)
func (m *Model) unwrapTargetedMsg(msg tea.Msg, targetID string) (tea.Msg, bool, bool) {
	if targetedMsg, isTargeted := msg.(core.TargetedMsg); isTargeted {
		// If target ID is specified and doesn't match, don't process
		if targetID != "" && targetedMsg.TargetID != targetID {
			return nil, true, false
		}
		return targetedMsg.InnerMsg, true, true
	}
	return msg, false, true
}

// handleMenuSelectionChange handles changes in menu selection
func (m *Model) handleMenuSelectionChange(menuID string, selectedItem interface{}) {
	m.StateMu.Lock()
	oldSelectedItem, existed := m.State[menuID+"_selected"]
	m.State[menuID+"_selected"] = selectedItem
	m.StateMu.Unlock()
	log.Trace("TUI KeyPress: Updated selected item for %s: %v", menuID, selectedItem)

	// Invalidate props cache when selection changes
	if m.propsCache != nil {
		m.propsCache.Clear()
		log.Trace("Menu selection changed, cleared props cache")
	}

	// If the selected item has changed, send a refresh command to update UI
	if !existed {
		// If there was no previous selection, this is a change
		log.Trace("TUI KeyPress: First selection for %s, sending refresh command", menuID)
		// Refresh the UI to show the selected item
		if m.Program != nil {
			m.Program.Send(core.RefreshMsg{})
		}
	} else {
		// Compare the items to determine if selection changed
		if oldSelectedItem != selectedItem {
			log.Trace("TUI KeyPress: Selection changed for %s (%v -> %v), sending refresh command", menuID, oldSelectedItem, selectedItem)
			// Refresh the UI to reflect the new selection
			if m.Program != nil {
				m.Program.Send(core.RefreshMsg{})
			}
		}
	}
}

// convertMenuActionToAction converts a menu action map to an Action struct
func (m *Model) convertMenuActionToAction(menuAction map[string]interface{}) *core.Action {
	action := &core.Action{}
	// Convert map to Action struct
	if process, ok := menuAction["process"].(string); ok {
		action.Process = process
	}
	if script, ok := menuAction["script"].(string); ok {
		action.Script = script
	}
	if method, ok := menuAction["method"].(string); ok {
		action.Method = method
	}
	if args, ok := menuAction["args"].([]interface{}); ok {
		action.Args = args
	}
	return action
}
