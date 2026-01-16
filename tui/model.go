package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
)

// NewModel creates a new Bubble Tea Model from a TUI configuration.
// It initializes the State with the data from Config and sets up
// the reactive environment.
func NewModel(cfg *Config, program *tea.Program) *Model {
	model := &Model{
		Config:  cfg,
		State:   make(map[string]interface{}),
		Program: program,
		Ready:   false,
	}

	// Copy initial data to State
	if cfg.Data != nil {
		for key, value := range cfg.Data {
			model.State[key] = value
		}
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

	case StreamChunkMsg:
		// Handle streaming chunk (e.g., from AI)
		return m.handleStreamChunk(msg)

	case StreamDoneMsg:
		// Handle stream completion
		return m.handleStreamDone(msg)

	case ErrorMsg:
		// Handle error
		return m.handleError(msg)

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
	key := msg.String()

	// Default quit key
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// Check for bound actions
	if m.Config.Bindings != nil {
		if action, ok := m.Config.Bindings[key]; ok {
			log.Trace("TUI KeyPress: %s -> action", key)
			return m, m.executeAction(&action)
		}
	}

	return m, nil
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

		// TODO: Integrate with gou.Process
		// result, err := process.New(action.Process, action.Args...).Exec()

		return ErrorMsg{
			Err:     nil,
			Context: "process execution not yet implemented",
		}
	}
}

// executeScriptAction creates a command to execute a script method.
func (m *Model) executeScriptAction(action *Action) tea.Cmd {
	return func() tea.Msg {
		// This will be implemented when we add script support
		log.Trace("TUI ExecuteScript: %s.%s", action.Script, action.Method)

		// TODO: Integrate with V8 runtime
		// script := LoadScript(action.Script)
		// result, err := script.Execute(action.Method, m, action.Args...)

		return ErrorMsg{
			Err:     nil,
			Context: "script execution not yet implemented",
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
