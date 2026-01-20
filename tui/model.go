package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// NewModel creates a new Bubble Tea Model from a TUI configuration.
// It initializes the State with the data from Config and sets up
// the reactive environment.
func NewModel(cfg *Config, program *tea.Program) *Model {
	model := &Model{
		Config:                     cfg,
		State:                      make(map[string]interface{}),
		Components:                 make(map[string]*core.ComponentInstance),
		ComponentInstanceRegistry:  NewComponentInstanceRegistry(),
		EventBus:                   core.NewEventBus(),
		Program:                    program,
		Ready:                      false,
		MessageHandlers:            GetDefaultMessageHandlersFromCore(),
		MessageSubscriptionManager: NewMessageSubscriptionManager(),
		exprCache:                  NewExpressionCache(),
		logLevel:                   cfg.LogLevel,
		propsCache:                 NewPropsCache(),
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

// Init initializes the Model and returns an initial command.
// This is called once when the program starts.
func (m *Model) Init() tea.Cmd {
	log.Trace("TUI Init: %s", m.Config.Name)

	// Collect all component Init commands
	componentCmds := m.InitializeComponents()

	// Build a list of commands to execute
	var cmds []tea.Cmd

	// Add component Init commands first
	cmds = append(cmds, componentCmds...)

	// Execute onLoad action if specified
	if m.Config.OnLoad != nil {
		cmds = append(cmds, m.executeAction(m.Config.OnLoad))
	}

	// Auto-focus to the first focusable component after initialization
	// This ensures that interactive components (like tables) can receive keyboard events
	// Only auto-focus if AutoFocus is enabled (default: true for backward compatibility)
	if m.Config.AutoFocus != nil && *m.Config.AutoFocus {
		focusableIDs := m.getFocusableComponentIDs()
		if len(focusableIDs) > 0 {
			cmds = append(cmds, func() tea.Msg {
				return core.FocusFirstComponentMsg{}
			})
		}
	}

	if len(cmds) == 0 {
		return nil
	}

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
	case tea.MouseMsg:
		return m, nil
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

	// Bubble phase: Global message handlers
	// Priority 3: Global handlers
	if handler, exists := m.MessageHandlers[msgType]; exists {
		log.Trace("TUI Update: Using handler for message type %s", msgType)
		return handler(m, msg)
	}

	log.Trace("TUI Update: No handler found for message type %s, broadcasting to all components", msgType)
	// Fallback: Broadcast unhandled messages to all components
	// This ensures messages like cursor.BlinkMsg reach components that need them
	return m.dispatchMessageToAllComponents(msg)
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

// renderLayout renders the TUI layout using the render module.
// The actual rendering logic is in render_engine.go
func (m *Model) renderLayout() string {
	return m.RenderLayout()
}

// syncInputComponentState synchronizes the state of an input component
func (m *Model) syncInputComponentState(id string, wrapper interface{}) {
	// This is a placeholder - actual implementation depends on component type
	// Will be handled by component-specific logic
}

// dispatchMessageToComponent dispatches a message to a specific component
// Returns (updatedModel, cmd, handled) where handled indicates if the component processed the message
func (m *Model) dispatchMessageToComponent(componentID string, msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	comp, exists := m.Components[componentID]
	if !exists {
		return m, nil, false
	}

	updatedComp, cmd, response := comp.Instance.UpdateMsg(msg)
	m.Components[componentID].Instance = updatedComp

	// Unified state synchronization using GetStateChanges()
	stateChanges, hasChanges := updatedComp.GetStateChanges()
	if hasChanges {
		m.StateMu.Lock()
		for key, value := range stateChanges {
			m.State[key] = value
		}
		m.StateMu.Unlock()

		// Invalidate props cache for components that reference state
		// This ensures expressions are re-evaluated when state changes
		if m.propsCache != nil {
			m.propsCache.Clear()
			log.Trace("State changes detected, cleared props cache")
		}
	}

	// Check if component lost focus after processing message
	// This handles ESC key to clear focus from components
	keyMsg, isKeyMsg := msg.(tea.KeyMsg)
	isESC := isKeyMsg && keyMsg.Type == tea.KeyEsc
	componentWasFocused := m.CurrentFocus == componentID

	// For ESC key, always check focus state regardless of response
	// For other keys, only check if the component handled the message
	shouldCheckFocus := (response == core.Handled && componentWasFocused) || (isESC && componentWasFocused)

	if shouldCheckFocus {
		// Check if this component lost focus
		if !comp.Instance.GetFocus() {
			m.clearFocus()
		}
	}

	return m, cmd, response == core.Handled
}

// dispatchMessageToAllComponents dispatches a message to all components
// Uses subscription manager to efficiently route messages
// Returns (updatedModel, cmd) where cmd is a batch of all component commands
func (m *Model) dispatchMessageToAllComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Get message type
	msgType := GetMessageTypeString(msg)

	// Get subscribers for this message type
	subscribers := m.MessageSubscriptionManager.GetSubscribers(msgType)

	if len(subscribers) > 0 {
		// Only dispatch to subscribed components
		for _, id := range subscribers {
			_, cmd, _ := m.dispatchMessageToComponent(id, msg)
			cmds = append(cmds, cmd)
		}
	} else {
		// No subscribers, fall back to dispatching to all components
		// This happens when components don't implement subscription
		for id := range m.Components {
			_, cmd, _ := m.dispatchMessageToComponent(id, msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}
