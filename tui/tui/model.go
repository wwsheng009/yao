package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/tui/core"
	tuiruntime "github.com/yaoapp/yao/tui/tui/runtime"
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
		// Set default dimensions to avoid 0x0 initialization issues
		// These will be updated when the first WindowSizeMsg is received
		Width:  80,
		Height: 24,
	}

	// Initialize the Bridge after EventBus is created
	model.Bridge = NewBridge(model.EventBus)

	// Read UseRuntime from config (default to true for new Runtime engine)
	// nil = use default (true), false = legacy mode, true = new runtime
	// if cfg.UseRuntime == nil {
	// 	model.UseRuntime = true // Default to new Runtime
	// } else {
	// 	model.UseRuntime = *cfg.UseRuntime
	// }
	model.UseRuntime = true

	// Initialize the Layout Engine
	// Note: LayoutEngine and LayoutRoot will be initialized in InitializeComponents()
	// The Renderer will be initialized later with the LayoutEngine and this Model as context

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
// Init 初始化TUI模型，执行以下操作：
// 1. 收集并执行所有组件的初始化命令
// 2. 执行配置中指定的onLoad动作
// 3. 如果启用了AutoFocus，自动聚焦到第一个可聚焦组件
// 返回一个tea.Cmd命令，包含所有需要执行的初始化操作
func (m *Model) Init() tea.Cmd {
	log.Trace("TUI Init: %s", m.Config.Name)

	// Collect all component Init commands FIRST
	// This ensures Components map is populated before Runtime adapter uses it
	componentCmds := m.InitializeComponents()

	// Initialize Runtime engine if enabled
	// Must happen AFTER InitializeComponents() so adapter can populate Component field
	// if m.UseRuntime {
	// 	m.initializeRuntime()
	// }
	m.initializeRuntime()

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
//
// ARCHITECTURE: Dual-path message handling
// ========================================
// This method implements a dual-path architecture to properly handle both:
// 1. Geometry events (mouse hit testing, focus navigation) → Runtime event system
// 2. Component messages (key input, cursor blink) → Bubble Tea path with Cmd propagation
//
// The key insight is that Runtime event system cannot return tea.Cmd (module boundary),
// so we must route messages that require command propagation through the Bubble Tea path.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ========== 新增：处理文本选择键盘快捷键 ==========
	// 检查选择相关的键盘快捷键（在有选择时，Ctrl+C 复制而不是退出）
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.handleSelectionKeyMsg(keyMsg) {
			m.forceRender = true // 选择变化需要重新渲染
			return m, nil
		}
		// 没有选择时，Ctrl+C 才是退出
		if keyMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	// ============================================

	// Mark that a user interaction message was received
	// This ensures the UI refreshes after any user action (key press, etc.)
	// We exclude high-frequency messages like cursor.BlinkMsg
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		m.messageReceived = true
	}

	// Route message based on event classification
	switch ClassifyMessage(msg) {
	case GeometryEvent:
		return m.handleGeometryEvent(msg)
	case ComponentEvent:
		return m.handleComponentMessage(msg)
	case SystemEvent:
		return m.handleSystemMessage(msg)
	default:
		// Fallback to broadcast for unknown message types
		return m.dispatchMessageToAllComponents(msg)
	}
}

// handleGeometryEvent handles geometry events through Runtime event system.
// These events don't need to return tea.Cmd:
// - Mouse events: hit testing, focus on click
func (m *Model) handleGeometryEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if m.RuntimeEngine != nil {
			m.DispatchEventToRuntime(msg)
			// 鼠标事件可能导致焦点变化（点击聚焦组件）
			m.forceRender = true
		}
		return m, nil
	}

	return m, nil
}

// handleComponentMessage handles component messages through Bubble Tea path.
// CRITICAL: This path MUST preserve tea.Cmd for proper Bubble Tea integration.
// All messages from TEA components (cursor.BlinkMsg, KeyMsg input, etc.) go here.
func (m *Model) handleComponentMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	msgType := getMsgTypeName(msg)
	log.Trace("TUI Component Message: %s", msgType)

	// Check for targeted message handlers first
	if msgType == "TargetedMsg" {
		if handler, exists := m.MessageHandlers[msgType]; exists {
			return handler(m, msg)
		}
	}

	// Check for global message handlers
	if handler, exists := m.MessageHandlers[msgType]; exists {
		return handler(m, msg)
	}

	// Broadcast to all subscribed components
	// This collects tea.Cmd from all components and returns them to the main loop
	return m.dispatchMessageToAllComponents(msg)
}

// handleSystemMessage handles system messages that affect the entire UI.
func (m *Model) handleSystemMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		if m.LayoutEngine != nil {
			m.LayoutEngine.UpdateWindowSize(msg.Width, msg.Height)
		}
		if m.UseRuntime {
			m.updateRuntimeWindowSize(msg.Width, msg.Height)
		}
		// 窗口大小变化必须重新渲染
		m.forceRender = true
		// Also notify components of window size changes
		return m.dispatchMessageToAllComponents(msg)
	}
	return m, nil
}

// handleSelectionKeyMsg handles keyboard shortcuts for text selection.
// Returns true if the key was handled by the selection system.
// Supports: Ctrl+C (copy), Ctrl+A (select all), Escape (clear), Ctrl+X (cut).
// Note: Selection functionality is simplified in the new runtime.
func (m *Model) handleSelectionKeyMsg(msg tea.KeyMsg) bool {
	// Selection features are not fully implemented in the new runtime yet
	// Return false to let the default handlers process the keys
	return false
}

// View renders the current state of the Model to a string.
// This is called after every Update.
//
// RENDER CACHE: 为了防止闪烁，此方法实现通用的渲染缓存机制。
// 只有当状态真正变化时才重新渲染，否则返回缓存的输出。
// 这解决了 cursor.BlinkMsg 等高频消息导致的闪烁问题。
func (m *Model) View() string {
	if !m.Ready {
		return "Initializing..."
	}

	// 检查是否需要重新渲染
	if !m.needsRender() {
		// 返回缓存的渲染结果
		return m.lastRenderedOutput
	}

	// 执行实际渲染
	var output string
	// output = m.renderWithRuntime()
	// if m.UseRuntime {
		// output = m.renderWithRuntime()
	// } else {
		output = m.renderLayout()
	// }

	// 更新缓存并重置 forceRender 标志
	m.lastRenderedOutput = output
	m.forceRender = false
	m.messageReceived = false

	// 清除 Runtime 引擎的 dirty 标志
	if m.RuntimeEngine != nil {
		m.RuntimeEngine.ClearDirty()
	}

	return output
}

// needsRender 检查是否需要重新渲染
// 返回 true 表示需要重新渲染，false 表示可以返回缓存
func (m *Model) needsRender() bool {
	// 如果是首次渲染或被强制渲染
	if m.lastRenderedOutput == "" || m.forceRender {
		return true
	}

	// 检查 Runtime 引擎是否被标记为 dirty
	if m.RuntimeEngine != nil && m.RuntimeEngine.IsDirty() {
		return true
	}

	// 检查是否有消息被处理 - 这确保任何用户交互都会触发刷新
	if m.messageReceived {
		return true
	}

	// 检查是否有焦点组件 - 焦点组件需要渲染以显示光标闪烁
	// 这是必要的，因为 cursor.BlinkMsg 不会设置 messageReceived 标志
	if m.CurrentFocus != "" {
		return true
	}

	// 检查是否有组件状态变化
	if m.hasComponentChanges() {
		return true
	}

	return false
}

// hasComponentChanges 检查是否有组件状态变化
func (m *Model) hasComponentChanges() bool {
	// 检查当前聚焦的组件是否需要更新（用于光标闪烁等）
	if m.CurrentFocus != "" {
		if comp, exists := m.Components[m.CurrentFocus]; exists {
			if instance, ok := comp.Instance.(interface{ NeedsRender() bool }); ok {
				if instance.NeedsRender() {
					return true
				}
			}
		}
	}
	return false
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

// renderLayout renders the TUI layout using the Renderer.
// The Renderer handles all layout tree traversal and component rendering.
func (m *Model) renderLayout() string {
	if m.Renderer == nil {
		log.Error("Model.renderLayout: Renderer is not initialized")
		return ""
	}
	return m.Renderer.Render()
}

// syncInputComponentState synchronizes the state of an input component
func (m *Model) syncInputComponentState(id string, wrapper interface{}) {
	// This is a placeholder - actual implementation depends on component type
	// Will be handled by component-specific logic
}

// dispatchMessageToComponent dispatches a message to a specific component
// Returns (updatedModel, cmd, handled) where handled indicates if the component processed the message
// Supports both legacy mode (m.Components) and runtime mode (m.RuntimeRoot.Children)
func (m *Model) dispatchMessageToComponent(componentID string, msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	// First, try legacy mode (m.Components)
	comp, exists := m.Components[componentID]
	if !exists {
		// If not found in Components and runtime is enabled, try runtime nodes
		// if m.RuntimeRoot != nil {
		// 	node := m.findRuntimeNodeByID(m.RuntimeRoot, componentID)
		// 	if node != nil && node.Component != nil {
		// 		// Found in runtime tree - use runtime component
		// 		return m.dispatchToRuntimeComponent(node, msg)
		// 	}
		// }
		return m, nil, false
	}

	updatedComp, cmd, response := comp.Instance.UpdateMsg(msg)
	m.Components[componentID].Instance = updatedComp

	// Unified state synchronization using GetStateChanges()
	stateChanges, hasChanges := updatedComp.GetStateChanges()
	if hasChanges {
		var actualChanges bool
		m.StateMu.Lock()
		for key, value := range stateChanges {
			// Only record actual changes
			oldValue, exists := m.State[key]
			if !exists || !valuesEqual(oldValue, value) {
				m.State[key] = value
				actualChanges = true
			}
		}
		m.StateMu.Unlock()

		// Only invalidate props cache if there were actual changes
		if actualChanges && m.propsCache != nil {
			m.propsCache.Clear()
			log.Trace("State changes detected, cleared props cache")
		}

		// CRITICAL: Mark for re-render when state changes
		// This ensures the UI updates to reflect the new state
		m.forceRender = true
	}

	// NOTE: Removed focus state check after message processing
	// Components should manage their own focus state autonomously.
	// If a component wants to lose focus (e.g., on ESC), it should do so
	// internally and not rely on the Model to clear CurrentFocus.
	// The Model should only track routing information, not manage component state.

	return m, cmd, response == core.Handled
}

// dispatchToRuntimeComponent dispatches a message to a runtime component
// Returns (updatedModel, cmd, handled) where handled indicates if the component processed the message
func (m *Model) dispatchToRuntimeComponent(node *tuiruntime.LayoutNode, msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if node == nil || node.Component == nil || node.Component.Instance == nil {
		return m, nil, false
	}

	// Type assert to core.ComponentInterface
	compIntf, ok := node.Component.Instance.(interface {
		UpdateMsg(tea.Msg) (interface{}, tea.Cmd, core.Response)
		GetStateChanges() (map[string]interface{}, bool)
	})
	if !ok {
		return m, nil, false
	}

	updatedComp, cmd, response := compIntf.UpdateMsg(msg)
	node.Component.Instance = updatedComp

	// Unified state synchronization using GetStateChanges()
	stateChanges, hasChanges := compIntf.GetStateChanges()
	if hasChanges {
		var actualChanges bool
		m.StateMu.Lock()
		for key, value := range stateChanges {
			// Only record actual changes
			oldValue, exists := m.State[key]
			if !exists || !valuesEqual(oldValue, value) {
				m.State[key] = value
				actualChanges = true
			}
		}
		m.StateMu.Unlock()

		// Only invalidate props cache if there were actual changes
		if actualChanges && m.propsCache != nil {
			m.propsCache.Clear()
			log.Trace("State changes detected, cleared props cache")
		}

		// CRITICAL: Mark for re-render when state changes
		// This ensures the UI updates to reflect the new state
		m.forceRender = true
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

// ============================================================================
// RenderContext Interface Implementation
// Model implements the RenderContext interface to work with the Renderer
// ============================================================================

// GetComponentInstance retrieves a component instance from the registry
func (m *Model) GetComponentInstance(id string) (*core.ComponentInstance, bool) {
	if m.ComponentInstanceRegistry != nil {
		return m.ComponentInstanceRegistry.Get(id)
	}
	return nil, false
}

// ResolveProps resolves component properties using the Model's props resolution logic
func (m *Model) ResolveProps(compID string) (map[string]interface{}, error) {
	compConfig := m.findComponentConfig(compID)
	if compConfig == nil {
		return make(map[string]interface{}), nil
	}
	return m.resolveProps(compConfig), nil
}

// UpdateComponentConfig updates a component's render configuration
func (m *Model) UpdateComponentConfig(instance *core.ComponentInstance, config core.RenderConfig, id string) bool {
	return updateComponentInstanceConfig(instance, config, id)
}

// RenderError renders an error display for a failed component
func (m *Model) RenderError(componentID, componentType string, err error) string {
	return m.renderErrorComponent(componentID, componentType, err)
}

// RenderUnknown renders a placeholder for unknown component types
func (m *Model) RenderUnknown(typeName string) string {
	return m.renderUnknownComponent(typeName)
}
