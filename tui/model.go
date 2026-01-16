package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/components"
)

// NewModel creates a new Bubble Tea Model from a TUI configuration.
// It initializes the State with the data from Config and sets up
// the reactive environment.
func NewModel(cfg *Config, program *tea.Program) *Model {
	model := &Model{
		Config:      cfg,
		State:       make(map[string]interface{}),
		InputModels: make(map[string]*components.InputModel),
		MenuModels:  make(map[string]*components.MenuInteractiveModel),
		Program:     program,
		Ready:       false,
	}

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

	// Execute onLoad action if specified
	if m.Config.OnLoad != nil {
		return m.executeAction(m.Config.OnLoad)
	}

	return nil
}

// Update handles incoming messages and updates the Model accordingly.
// This is the core of the Bubble Tea message loop.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update terminal dimensions
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		log.Trace("TUI WindowSize: %dx%d", m.Width, m.Height)
		return m, nil

	case tea.KeyMsg:
		// Handle keyboard input
		return m.handleKeyPress(msg)

	case ProcessResultMsg:
		// Handle Process execution result
		return m.handleProcessResult(msg)

	case StateUpdateMsg:
		// Handle single state update
		m.StateMu.Lock()
		m.State[msg.Key] = msg.Value
		m.StateMu.Unlock()
		log.Trace("TUI StateUpdate: %s = %v", msg.Key, msg.Value)
		return m, nil

	case StateBatchUpdateMsg:
		// Handle batch state updates
		m.StateMu.Lock()
		for key, value := range msg.Updates {
			m.State[key] = value
		}
		m.StateMu.Unlock()
		log.Trace("TUI StateBatchUpdate: %d keys", len(msg.Updates))
		return m, nil

	case InputUpdateMsg:
		// Handle input component updates
		m.StateMu.Lock()
		m.State[msg.ID] = msg.Value
		m.StateMu.Unlock()
		log.Trace("TUI InputUpdate: %s = %s", msg.ID, msg.Value)
		return m, nil

	case StreamChunkMsg:
		// Handle streaming chunk (e.g., from AI)
		return m.handleStreamChunk(msg)

	case StreamDoneMsg:
		// Handle stream completion
		return m.handleStreamDone(msg)

	case ErrorMsg:
		// Handle error
		return m.handleError(msg)

	case QuitMsg:
		// Handle quit request
		return m, tea.Quit

	default:
		return m, nil
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
	// Default quit key
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Handle input component if there's a focused input
	if m.CurrentFocus != "" {
		if inputModel, exists := m.InputModels[m.CurrentFocus]; exists {
			// Check if the key pressed is Escape to blur the input
			if msg.Type == tea.KeyEsc {
				// Blur the input and clear focus
				inputModel.Blur()
				m.InputModels[m.CurrentFocus] = inputModel
				m.CurrentFocus = ""
				log.Trace("Input blurred via ESC, focus cleared")
				// Now that focus is cleared, we should continue to check for key bindings
				// But since we return here, we'll fall through to general key handling
				return m, nil
			}
			
			updatedInputModel, cmd := components.HandleInputUpdate(msg, inputModel)
			m.InputModels[m.CurrentFocus] = &updatedInputModel
			
			// Update the state with the current input value
			m.StateMu.Lock()
			m.State[m.CurrentFocus] = updatedInputModel.Value()
			m.StateMu.Unlock()
			
			// If Enter is pressed, submit the form
			if msg.Type == tea.KeyEnter {
				log.Trace("Input submitted: %s = %s", m.CurrentFocus, updatedInputModel.Value())
				// Submit form action if bound
				if m.Config.Bindings != nil {
					if action, ok := m.Config.Bindings["Enter"]; ok {
						log.Trace("TUI KeyPress: Enter -> action")
						return m, m.executeAction(&action)
					}
				}
			}
			
			// If Tab is pressed, move to next input
			if msg.Type == tea.KeyTab {
				log.Trace("Tab pressed, moving to next input")
				// Call focus next input
				m.focusNextInput()
				// Update focus states in input models
				for id, inputModel := range m.InputModels {
					if id == m.CurrentFocus {
						inputModel.Model.Focus()
					} else {
						inputModel.Model.Blur()
					}
					m.InputModels[id] = inputModel
				}
			}
			return m, cmd
		}
	}
	
	// If no specific component has focus, handle tab navigation between components
	if msg.Type == tea.KeyTab {
		log.Trace("Tab pressed, cycling focus between components, current focus: %s", m.CurrentFocus)
		// Cycle focus between available components
		// First, cycle through inputs if available
		if len(m.InputModels) > 0 {
			if m.CurrentFocus == "" {
				// If no focus, start with first input
				for id := range m.InputModels {
					m.CurrentFocus = id
					m.InputModels[id].Model.Focus()
					log.Trace("Focused input: %s", id)
					break
				}
			} else {
				// Cycle to next input
				m.focusNextInput()
				for id, inputModel := range m.InputModels {
					if id == m.CurrentFocus {
						inputModel.Model.Focus()
					} else {
						inputModel.Model.Blur()
					}
					m.InputModels[id] = inputModel
				}
			}
		} else if len(m.MenuModels) > 0 {
			// If no inputs but there are menus, cycle to menu
			for id := range m.MenuModels {
				m.CurrentFocus = "menu:" + id
				log.Trace("Focused menu: %s", id)
				break
			}
		}
		return m, nil
	}

	// Handle menu component if there's a focused menu
	log.Trace("TUI KeyPress: Checking for menu models, found %d, current focus: %s", len(m.MenuModels), m.CurrentFocus)
	for menuID, menuModel := range m.MenuModels {
		log.Trace("TUI KeyPress: Checking menu model with ID: %s, current focus: %s", menuID, m.CurrentFocus)
		// If we have a focused menu, handle its navigation
		if m.CurrentFocus == "menu:"+menuID {
			log.Trace("TUI KeyPress: Handling navigation for menu: %s", menuID)
			// Handle menu navigation keys - pass all key events to the menu handler
			updatedMenuModel, cmd := components.HandleMenuUpdate(msg, menuModel)
			m.MenuModels[menuID] = &updatedMenuModel
			
			// Update state with selected item
			if selectedItem, ok := updatedMenuModel.GetSelectedItem(); ok {
				m.StateMu.Lock()
				// Check if the selected item has actually changed
				oldSelectedItem, existed := m.State[menuID+"_selected"]
				m.State[menuID+"_selected"] = selectedItem
				m.StateMu.Unlock()
				log.Trace("TUI KeyPress: Updated selected item for %s: %s", menuID, selectedItem.Title)
				
				// If the selected item has changed, send a refresh command to update UI
				if !existed {
					// If there was no previous selection, this is a change
					log.Trace("TUI KeyPress: First selection for %s, sending refresh command", menuID)
					// Refresh the UI to show the selected item
					return m, tea.Batch(cmd, func() tea.Msg { 
						// Create a custom message to trigger refresh
						return struct{}{} // Generic struct as refresh signal
					})
				} else if oldMenuItem, ok := oldSelectedItem.(components.MenuItem); ok {
					// Compare the titles to determine if selection changed
					if oldMenuItem.Title != selectedItem.Title {
						log.Trace("TUI KeyPress: Selection changed for %s (%s -> %s), sending refresh command", menuID, oldMenuItem.Title, selectedItem.Title)
						// Refresh the UI to reflect the new selection
						return m, tea.Batch(cmd, func() tea.Msg { 
							// Create a custom message to trigger refresh
							return struct{}{} // Generic struct as refresh signal
						})
					}
				}
			}
			
			// If the command is to trigger a menu action, handle it
			if cmd != nil {
				// Create a temporary command to check its type
				tempCmd := cmd
				// Execute temp command to get the message
				testMsg := tempCmd()
				if menuActionMsg, ok := testMsg.(components.MenuActionTriggered); ok {
					log.Trace("TUI KeyPress: Menu action triggered for %s: %s", menuActionMsg.Item.Title, menuActionMsg.Action)
					// Execute the action associated with the selected menu item
					action := &Action{}
					// Convert map to Action struct
					if process, ok := menuActionMsg.Action["process"].(string); ok {
						action.Process = process
					}
					if script, ok := menuActionMsg.Action["script"].(string); ok {
						action.Script = script
					}
					if method, ok := menuActionMsg.Action["method"].(string); ok {
						action.Method = method
					}
					if args, ok := menuActionMsg.Action["args"].([]interface{}); ok {
						action.Args = args
					}
					
					log.Trace("TUI KeyPress: Executing action for %s: %s", menuActionMsg.Item.Title, action.Process)
					return m, m.executeAction(action)
				}
			}
			
			return m, cmd
		} else {
			log.Trace("TUI KeyPress: Menu %s is not focused (focus: %s)", menuID, m.CurrentFocus)
		}
	}
	log.Trace("TUI KeyPress: Finished checking menu models, no focused menu processed")


	// Build key string for matching
	key := msg.String()
	
	// Also try single rune if available (for single character keys like 'a', '+', etc.)
	if len(msg.Runes) == 1 {
		char := string(msg.Runes[0])
		// Check for bound actions with the character
		if m.Config.Bindings != nil {
			if action, ok := m.Config.Bindings[char]; ok {
				// If the action is a quit action, execute it
				if action.Process == "tui.quit" || action.Process == "tui.exit" {
					log.Trace("TUI KeyPress: %s -> quit action", char)
					return m, m.executeAction(&action)
				} else {
					log.Trace("TUI KeyPress: %s -> action", char)
					return m, m.executeAction(&action)
				}
			}
		}
	}
	
	// Check for bound actions with full key string
	if m.Config.Bindings != nil {
		if action, ok := m.Config.Bindings[key]; ok {
			// If the action is a quit action, execute it
			if action.Process == "tui.quit" || action.Process == "tui.exit" {
				log.Trace("TUI KeyPress: %s -> quit action", key)
				return m, m.executeAction(&action)
			} else {
				log.Trace("TUI KeyPress: %s -> action", key)
				return m, m.executeAction(&action)
			}
		}
	}

	return m, nil
}

// isMenuFocus checks if the current focus is on a menu component
func (m *Model) isMenuFocus() bool {
	if m.CurrentFocus == "" {
		return false
	}
	return strings.HasPrefix(m.CurrentFocus, "menu:")
}

// handleProcessResult processes the result from a Yao Process execution.
func (m *Model) handleProcessResult(msg ProcessResultMsg) (tea.Model, tea.Cmd) {
	if msg.Target != "" {
		m.StateMu.Lock()
		m.State[msg.Target] = msg.Data
		m.StateMu.Unlock()
		log.Trace("TUI ProcessResult: %s = %v", msg.Target, msg.Data)
	}
	return m, nil
}

// handleStreamChunk handles a streaming data chunk.
func (m *Model) handleStreamChunk(msg StreamChunkMsg) (tea.Model, tea.Cmd) {
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
func (m *Model) handleStreamDone(msg StreamDoneMsg) (tea.Model, tea.Cmd) {
	log.Trace("TUI StreamDone: %s", msg.ID)
	// Mark stream as complete in state
	m.StateMu.Lock()
	m.State["stream_"+msg.ID+"_done"] = true
	m.StateMu.Unlock()
	return m, nil
}

// handleError handles error messages.
func (m *Model) handleError(msg ErrorMsg) (tea.Model, tea.Cmd) {
	log.Error("TUI Error: %v", msg)

	// Store error in state
	m.StateMu.Lock()
	m.State["__error"] = msg.Error()
	m.StateMu.Unlock()

	return m, nil
}

// executeAction creates a command to execute an action.
// This returns a tea.Cmd that will be executed asynchronously.
func (m *Model) executeAction(action *Action) tea.Cmd {
	if action == nil {
		return nil
	}

	// Validate action
	if err := action.Validate(); err != nil {
		return func() tea.Msg {
			return ErrorMsg{
				Err:     err,
				Context: "action validation",
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
			return StateBatchUpdateMsg{
				Updates: action.Payload,
			}
		}
	}

	return nil
}

// executeProcessAction creates a command to execute a Yao Process.
func (m *Model) executeProcessAction(action *Action) tea.Cmd {
	return func() tea.Msg {
		// This will be implemented when we integrate with Yao's Process system
		// For now, return a placeholder
		log.Trace("TUI ExecuteProcess: %s", action.Process)

		result, err := executeProcessAction(m, action)
		if err != nil {
			return ErrorMsg{
				Err:     err,
				Context: "process execution failed",
			}
		}
		
		return ProcessResultMsg{
			Data:    result,
			Target:  action.OnSuccess,
		}
	}
}

// executeScriptAction creates a command to execute a script method.
func (m *Model) executeScriptAction(action *Action) tea.Cmd {
	return func() tea.Msg {
		// This will be implemented when we add script support
		log.Trace("TUI ExecuteScript: %s.%s", action.Script, action.Method)

		result, err := executeScriptAction(m, action)
		if err != nil {
			log.Error("TUI ExecuteScript error: %v", err)
			return ErrorMsg{
				Err:     err,
				Context: "script execution failed",
			}
		}
		
		return ProcessResultMsg{
			Data:    result,
			Target:  action.OnSuccess,
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
		m.Program.Send(StateUpdateMsg{
			Key:   key,
			Value: value,
		})
	}
}

// UpdateState safely updates multiple state values at once.
func (m *Model) UpdateState(updates map[string]interface{}) {
	if m.Program != nil {
		m.Program.Send(StateBatchUpdateMsg{
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
func (m *Model) focusNextInput() {
	// Find all input component IDs in the layout
	inputIDs := []string{}
	for _, comp := range m.Config.Layout.Children {
		if comp.Type == "input" && comp.ID != "" {
			inputIDs = append(inputIDs, comp.ID)
		}
	}
	
	// Find current position and move to next
	currentIndex := -1
	for i, id := range inputIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}
	
	// Move to next input, wrap around if needed
	if currentIndex >= 0 && currentIndex < len(inputIDs)-1 {
		m.CurrentFocus = inputIDs[currentIndex+1]
	} else if len(inputIDs) > 0 {
		m.CurrentFocus = inputIDs[0] // Wrap to first
	}
}

