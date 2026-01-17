package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/components"
	"github.com/yaoapp/yao/tui/core"
)

// NewModel creates a new Bubble Tea Model from a TUI configuration.
// It initializes the State with the data from Config and sets up
// the reactive environment.
func NewModel(cfg *Config, program *tea.Program) *Model {
	model := &Model{
		Config:          cfg,
		State:           make(map[string]interface{}),
		Components:      make(map[string]*core.ComponentInstance),
		EventBus:        core.NewEventBus(),
		Program:         program,
		Ready:           false,
		MessageHandlers: GetDefaultMessageHandlersFromCore(),
	}

	// Initialize the Bridge after EventBus is created
	model.Bridge = NewBridge(model.EventBus)

	// Copy initial data to State
	if cfg.Data != nil {
		for key, value := range cfg.Data {
			model.State[key] = value
		}
	}

	// Register the model if it has an ID
	if cfg.ID != "" {
		RegisterModel(cfg.ID, model)
	}

	return model
}

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
		model.Ready = true // Mark model as ready after receiving window size

		// Broadcast window size to all components
		var cmds []tea.Cmd
		for id, comp := range model.Components {
			newComp, cmd, _ := comp.Instance.UpdateMsg(msg)
			model.Components[id].Instance = newComp
			cmds = append(cmds, cmd)
		}

		// Trigger auto-focus after window size is received and components are initialized
		cmds = append(cmds, func() tea.Msg {
			return core.FocusFirstComponentMsg{}
		})

		return model, tea.Batch(cmds...)
	}

	// Register handler for core.TargetedMsg
	handlers["TargetedMsg"] = func(m interface{}, msg tea.Msg) (tea.Model, tea.Cmd) {
		model, ok := m.(*Model)
		if !ok {
			return m.(tea.Model), nil
		}
		targetedMsg := msg.(core.TargetedMsg)

		// Find target component
		if comp, exists := model.Components[targetedMsg.TargetID]; exists {
			newComp, cmd, _ := comp.Instance.UpdateMsg(targetedMsg.InnerMsg)
			model.Components[targetedMsg.TargetID].Instance = newComp
			return model, cmd
		}

		// Target not found, ignore message
		return model, nil
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
			// Move focus to next input component
			model.focusNextInput()
		case core.EventFocusPrev:
			// Move focus to previous input component
			model.focusPrevInput()
		case core.EventFocusChanged:
			// Update focus based on data
			if data, ok := actionMsg.Data.(map[string]interface{}); ok {
				if focused, ok := data["focused"].(bool); ok && focused {
					// Set focus to the component that sent this message
					model.setFocus(actionMsg.ID)
				} else {
					// Clear focus if focused is false
					model.clearFocus()
				}
			}
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

		if !model.Ready {
			return model, nil
		}

		// Get all focusable components
		focusableIDs := model.getFocusableComponentIDs()
		if len(focusableIDs) > 0 {
			// Set focus to the first focusable component
			model.setFocus(focusableIDs[0])
			log.Trace("TUI: Auto-focus to first focusable component: %s", focusableIDs[0])
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

// Init initializes the Model and returns an initial command.
// This is called once when the program starts.
func (m *Model) Init() tea.Cmd {
	log.Trace("TUI Init: %s", m.Config.Name)

	// Build a list of commands to execute
	var cmds []tea.Cmd

	// Execute onLoad action if specified
	if m.Config.OnLoad != nil {
		cmds = append(cmds, m.executeAction(m.Config.OnLoad))
	}

	// Auto-focus to the first focusable component after initialization
	// This ensures that interactive components (like tables) can receive keyboard events
	cmds = append(cmds, func() tea.Msg {
		return core.FocusFirstComponentMsg{}
	})

	return tea.Batch(cmds...)
}

// Update handles incoming messages and updates the Model accordingly.
// This is the core of the Bubble Tea message loop.
// Implements a Windows-style message dispatching mechanism:
// 1. Capture phase: System-level interception
// 2. Dispatch phase: Route to focused component
// 3. Bubble phase: Global handlers
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Capture phase: System-level message interception
	// Priority 1: Critical system messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			// Always intercept Ctrl+C regardless of focus
			return m, tea.Quit
		}
		// For other keys, continue to dispatch phase
	case tea.WindowSizeMsg:
		// Window size changes are handled globally
		// but also need to be propagated to all components
		// Store dimensions and let the handler process it
	}

	// Dispatch phase: Route message to focused component
	// Priority 2: Targeted component handling
	msgType := getMsgTypeName(msg)
	log.Trace("TUI Update: Received message of type %s", msgType)

	// Check if we have a targeted message first
	if msgType == "TargetedMsg" {
		// This is already handled by the TargetedMsg handler
		if handler, exists := m.MessageHandlers[msgType]; exists {
			log.Trace("TUI Update: Using handler for message type %s", msgType)
			return handler(m, msg)
		}
	}

	// For regular messages, try to dispatch to focused component
	if m.CurrentFocus != "" {
		// Check if there's a registered component with this focus ID
		if comp, exists := m.Components[m.CurrentFocus]; exists {
			updatedComp, cmd, response := comp.Instance.UpdateMsg(msg)
			if response == core.Handled {
				log.Trace("TUI Update: Message handled by focused component %s", m.CurrentFocus)
				m.Components[m.CurrentFocus].Instance = updatedComp
				return m, cmd
			}
			// If not handled, continue to global handlers
		}
	}

	// Bubble phase: Global message handlers
	// Priority 3: Global handlers
	if handler, exists := m.MessageHandlers[msgType]; exists {
		log.Trace("TUI Update: Using handler for message type %s", msgType)
		return handler(m, msg)
	}

	log.Trace("TUI Update: No handler found for message type %s, using default behavior", msgType)
	// Fallback to default behavior for unhandled message types
	return m, nil
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
		return "unknown"
	}
}

// View renders the current state of the Model to a string.
// This is called after every Update.
func (m *Model) View() string {
	if !m.Ready {
		return "Initializing..."
	}

	// Render the layout
	return m.renderLayout()
}

// handleKeyPress processes keyboard input and executes bound actions.
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Capture phase: Global system keys
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Dispatch phase: Route to focused component
	if m.CurrentFocus != "" {
		// Check if there's a registered component with this focus ID
		if comp, exists := m.Components[m.CurrentFocus]; exists {
			updatedComp, cmd, response := comp.Instance.UpdateMsg(msg)
			if response == core.Handled {
				// Update the component instance
				m.Components[m.CurrentFocus].Instance = updatedComp
				// Sync input state for input components
				if inputWrapper, ok := updatedComp.(*components.InputComponentWrapper); ok {
					m.syncInputComponentState(m.CurrentFocus, inputWrapper)
				}
				return m, cmd
			}
		}

		// All components should be in the new system (Components)
		// No legacy handling needed
	}

	// Bubble phase: Handle general navigation and bound actions
	// Handle ESC key to clear focus
	if msg.Type == tea.KeyEsc && m.CurrentFocus != "" {
		m.clearFocus()
		return m, nil
	}

	// Handle bound actions for keys
	return m.handleBoundActions(msg)
}

// handleTabNavigation handles tab navigation between components
func (m *Model) handleTabNavigation() (tea.Model, tea.Cmd) {
	log.Trace("Tab pressed, cycling focus between components, current focus: %s", m.CurrentFocus)

	// First, check if there are any input components
	hasInputs := false
	for _, comp := range m.Components {
		if comp.Type == "input" {
			hasInputs = true
			break
		}
	}

	if hasInputs {
		m.focusNextInput()
	}
	// Note: Menu components are handled through their own UpdateMsg method
	return m, nil
}

// handleMenuSelectionChange handles changes in menu selection
func (m *Model) handleMenuSelectionChange(menuID string, selectedItem interface{}) {
	m.StateMu.Lock()
	oldSelectedItem, existed := m.State[menuID+"_selected"]
	m.State[menuID+"_selected"] = selectedItem
	m.StateMu.Unlock()
	log.Trace("TUI KeyPress: Updated selected item for %s: %v", menuID, selectedItem)

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

// updateInputFocusStates updates the focus states of all components
func (m *Model) updateInputFocusStates() {
	for id, compInstance := range m.Components {
		if id == m.CurrentFocus {
			compInstance.Instance.SetFocus(true)
		} else {
			compInstance.Instance.SetFocus(false)
		}
	}
}

// isMenuFocused checks if the current focus is on a menu component
func (m *Model) isMenuFocused() bool {
	if m.CurrentFocus == "" {
		return false
	}
	if comp, exists := m.Components[m.CurrentFocus]; exists {
		return comp.Type == "menu"
	}
	return false
}

// handleBoundActions handles bound actions for keys
func (m *Model) handleBoundActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Build key string for matching
	key := msg.String()

	// Also try single rune if available (for single character keys like 'a', '+', etc.)
	if len(msg.Runes) == 1 {
		char := string(msg.Runes[0])
		// Check for bound actions with the character
		if m.Config.Bindings != nil {
			if action, ok := m.Config.Bindings[char]; ok {
				return m.executeBoundAction(&action, char)
			}
		}
	}

	// Check for bound actions with full key string
	if m.Config.Bindings != nil {
		if action, ok := m.Config.Bindings[key]; ok {
			return m.executeBoundAction(&action, key)
		}
	}

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
	} else {
		// Handle success case
		if msg.Target != "" {
			m.StateMu.Lock()
			m.State[msg.Target] = msg.Data
			m.StateMu.Unlock()
			log.Trace("TUI ProcessResult: %s = %v", msg.Target, msg.Data)
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
	if action.Payload != nil && len(action.Payload) > 0 {
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

// renderLayout renders the TUI layout using the render module.
// The actual rendering logic is in render.go
func (m *Model) renderLayout() string {
	return m.RenderLayout()
}

// GetState safely retrieves a state value.
// This is thread-safe and can be called from external goroutines.
func (m *Model) GetState(key string) (interface{}, bool) {
	m.StateMu.RLock()
	defer m.StateMu.RUnlock()
	value, ok := m.State[key]
	return value, ok
}

// SetState safely sets a state value.
// This is thread-safe and can be called from external goroutines.
// It sends a message to the Model's update loop.
func (m *Model) SetState(key string, value interface{}) {
	if m.Program != nil {
		m.Program.Send(core.StateUpdateMsg{
			Key:   key,
			Value: value,
		})
	}
}

// UpdateState safely updates multiple state values at once.
func (m *Model) UpdateState(updates map[string]interface{}) {
	if m.Program != nil {
		m.Program.Send(core.StateBatchUpdateMsg{
			Updates: updates,
		})
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
	defer m.StateMu.Unlock()
	m.State[key] = value
}

// focusNextInput finds the next input component and sets it as focused
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

func (m *Model) focusNextInput() {
	// Find all input component IDs from Components map
	inputIDs := []string{}
	for id, comp := range m.Components {
		if comp.Type == "input" {
			inputIDs = append(inputIDs, id)
		}
	}

	if len(inputIDs) == 0 {
		return
	}

	// Find current position
	currentIndex := -1
	for i, id := range inputIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Determine next focus ID
	var nextFocus string
	if currentIndex >= 0 && currentIndex < len(inputIDs)-1 {
		nextFocus = inputIDs[currentIndex+1]
	} else {
		nextFocus = inputIDs[0] // Wrap to first
	}

	// Use setFocus which handles focus change events
	m.setFocus(nextFocus)
}

func (m *Model) focusPrevInput() {
	// Find all input component IDs from Components map
	inputIDs := []string{}
	for id, comp := range m.Components {
		if comp.Type == "input" {
			inputIDs = append(inputIDs, id)
		}
	}

	if len(inputIDs) == 0 {
		return
	}

	// Find current position
	currentIndex := -1
	for i, id := range inputIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Determine previous focus ID
	var prevFocus string
	if currentIndex > 0 {
		prevFocus = inputIDs[currentIndex-1]
	} else if currentIndex == 0 {
		prevFocus = inputIDs[len(inputIDs)-1] // Wrap to last
	} else {
		// No current focus, start from last
		prevFocus = inputIDs[len(inputIDs)-1]
	}

	// Use setFocus which handles focus change events
	m.setFocus(prevFocus)
}

// setFocus sets focus to a specific component
func (m *Model) setFocus(componentID string) {
	if componentID == m.CurrentFocus {
		return // Already focused
	}

	// Clear focus from current component
	m.clearFocus()

	// Set new focus
	m.CurrentFocus = componentID
	if comp, exists := m.Components[componentID]; exists {
		comp.Instance.SetFocus(true)
	}

	// Publish focus changed event
	m.EventBus.Publish(core.ActionMsg{
		ID:     componentID,
		Action: core.EventFocusChanged,
		Data:   map[string]interface{}{"focused": true},
	})

	log.Trace("TUI Focus: Focus set to %s", componentID)
}

// clearFocus clears focus from current component
func (m *Model) clearFocus() {
	if m.CurrentFocus == "" {
		return
	}

	// Clear focus from component
	if comp, exists := m.Components[m.CurrentFocus]; exists {
		comp.Instance.SetFocus(false)
	}

	oldFocus := m.CurrentFocus
	m.CurrentFocus = ""

	// Publish focus changed event
	m.EventBus.Publish(core.ActionMsg{
		ID:     oldFocus,
		Action: core.EventFocusChanged,
		Data:   map[string]interface{}{"focused": false},
	})

	log.Trace("TUI Focus: Focus cleared from %s", oldFocus)
}

// focusNextComponent moves focus to the next focusable component
func (m *Model) focusNextComponent() {
	// Get all focusable component IDs
	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		return
	}

	// Find current position
	currentIndex := -1
	for i, id := range focusableIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Move to next component, wrap around if needed
	var nextFocus string
	if currentIndex >= 0 && currentIndex < len(focusableIDs)-1 {
		nextFocus = focusableIDs[currentIndex+1]
	} else {
		nextFocus = focusableIDs[0] // Wrap to first
	}

	m.setFocus(nextFocus)
}

// focusPrevComponent moves focus to the previous focusable component
func (m *Model) focusPrevComponent() {
	// Get all focusable component IDs
	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		return
	}

	// Find current position
	currentIndex := -1
	for i, id := range focusableIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Move to previous component, wrap around if needed
	var prevFocus string
	if currentIndex > 0 {
		prevFocus = focusableIDs[currentIndex-1]
	} else if currentIndex == 0 {
		prevFocus = focusableIDs[len(focusableIDs)-1] // Wrap to last
	} else {
		// No current focus, start from last
		prevFocus = focusableIDs[len(focusableIDs)-1]
	}

	m.setFocus(prevFocus)
}

// getFocusableComponentIDs returns IDs of all focusable components
func (m *Model) getFocusableComponentIDs() []string {
	// Define which component types are focusable
	focusableTypes := map[string]bool{
		"input": true,
		"menu":  true,
		"form":  true,
		"table": true,
		"crud":  true,
		"chat":  true,
	}

	ids := []string{}
	for id, comp := range m.Components {
		if focusableTypes[comp.Type] {
			ids = append(ids, id)
		}
	}
	return ids
}

// syncInputComponentState synchronizes the state of an input component
func (m *Model) syncInputComponentState(id string, wrapper *components.InputComponentWrapper) {
	// Update state with current value from component
	m.StateMu.Lock()
	m.State[id] = wrapper.GetValue()
	m.StateMu.Unlock()
}
