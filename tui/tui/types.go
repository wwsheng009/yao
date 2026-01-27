// Package tui provides Terminal User Interface engine for Yao.
//
// TUI engine allows developers to define terminal UIs through JSON/YAML
// configuration files (.tui.yao), supporting declarative UI layout,
// reactive state management, and JavaScript/TypeScript scripting.
package tui

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/tui/core"
	"github.com/yaoapp/yao/tui/tui/legacy/layout"
	tuiruntime "github.com/yaoapp/yao/tui/tui/runtime"
)

// MessageHandler defines a function that handles a specific message type
// and returns an updated model and command
type MessageHandler func(*Model, tea.Msg) (tea.Model, tea.Cmd)

// Bridge message types for internal routing
const (
	MsgTypeTargeted         = "TARGETED_MESSAGE"
	MsgTypeStateUpdate      = "STATE_UPDATE"
	MsgTypeStateBatchUpdate = "STATE_BATCH_UPDATE"
	MsgTypeProcessResult    = "PROCESS_RESULT"
	MsgTypeError            = "ERROR"
)

// Config represents the parsed .tui.yao configuration file.
// It defines the TUI's name, initial data state, layout structure,
// and event bindings.
type Config struct {
	// ID is the unique identifier for this TUI (derived from file path)
	ID string `json:"id,omitempty"`

	// Name is the human-readable name of the TUI
	Name string `json:"name,omitempty"`

	// Data contains the initial state data
	Data map[string]interface{} `json:"data,omitempty"`

	// OnLoad is the action to execute when TUI loads
	OnLoad *core.Action `json:"onLoad,omitempty"`

	// Layout defines the UI structure
	Layout Layout `json:"layout,omitempty"`

	// Bindings maps keyboard shortcuts to actions
	Bindings map[string]core.Action `json:"bindings,omitempty"`

	// LogLevel controls the verbosity of logging (default: "warn")
	// Options: "trace", "debug", "info", "warn", "error", "none"
	LogLevel string `json:"logLevel,omitempty"`

	// AutoFocus enables automatic focus to the first focusable component (default: true)
	// Use a pointer to distinguish between "not set" and "explicitly false"
	AutoFocus *bool `json:"autoFocus,omitempty"`

	// NavigationMode defines how Tab/ShiftTab keys are handled
	// "native": Tab always navigates between components (default)
	// "bindable": Tab can be bound to custom actions
	NavigationMode string `json:"navigationMode,omitempty"`

	// TabCycles defines whether Tab navigation cycles through components (default: true)
	TabCycles bool `json:"tabCycles,omitempty"`

	// UseRuntime enables the new Runtime layout engine (default: true)
	// When nil or true, the new Runtime engine is used for layout and rendering
	// Set to false to opt-out and use the legacy layout system
	UseRuntime *bool `json:"useRuntime,omitempty"`
}

// Layout describes the UI layout structure.
// It can be nested to create complex hierarchical layouts.
type Layout struct {
	// Direction specifies how children are arranged: "vertical" or "horizontal"
	Direction string `json:"direction,omitempty"`

	// Children contains the child components or sub-layouts
	Children []Component `json:"children,omitempty"`

	// Style is the optional style name to apply
	Style string `json:"style,omitempty"`

	// Padding specifies the padding [top, right, bottom, left]
	Padding []int `json:"padding,omitempty"`
}

// Component represents a UI component in the layout.
type Component struct {
	// ID is the unique identifier for this component instance
	ID string `json:"id,omitempty"`

	// Type specifies the component type (e.g., "header", "table", "form")
	Type string `json:"type,omitempty"`

	// Bind specifies the state key to bind data from
	Bind string `json:"bind,omitempty"`

	// Props contains component-specific properties
	Props map[string]interface{} `json:"props,omitempty"`

	// Style contains component-level style properties (position, etc.)
	Style map[string]interface{} `json:"style,omitempty"`

	// Actions maps event names to actions for this component
	Actions map[string]core.Action `json:"actions,omitempty"`

	// Height specifies the component height ("flex" for flexible, or number)
	Height interface{} `json:"height,omitempty"`

	// Width specifies the component width
	Width interface{} `json:"width,omitempty"`

	// Children contains child components (for nested layouts)
	// This is used in the old format where type="layout" can have children directly
	Children []Component `json:"children,omitempty"`

	// Direction specifies layout direction for layout components
	// This is used in the old format where type="layout" has direction property
	Direction string `json:"direction,omitempty"`
}

// Bridge provides external message bridge for async operations
type Bridge struct {
	MessageChan chan tea.Msg
	EventBus    *core.EventBus
}

// NewBridge creates a new Bridge instance
func NewBridge(eventBus *core.EventBus) *Bridge {
	bridge := &Bridge{
		MessageChan: make(chan tea.Msg, 100), // Buffered channel to prevent blocking
		EventBus:    eventBus,
	}

	// Start a goroutine to process messages from the channel
	go func() {
		for msg := range bridge.MessageChan {
			// Handle different message types
			switch v := msg.(type) {
			case core.ActionMsg:
				// Forward ActionMsg to EventBus for component communication
				bridge.EventBus.Publish(v)

			case core.TargetedMsg:
				// TargetedMsg should be routed through message loop
				// It will be handled by the TargetedMsg handler in Model.Update
				// We can't directly forward it here without access to the Model
				// So we'll publish it as a special action
				bridge.EventBus.Publish(core.ActionMsg{
					ID:     "bridge",
					Action: MsgTypeTargeted,
					Data:   v,
				})

			case core.StateUpdateMsg:
				// Handle state updates directly
				bridge.EventBus.Publish(core.ActionMsg{
					ID:     "bridge",
					Action: MsgTypeStateUpdate,
					Data:   v,
				})

			case core.StateBatchUpdateMsg:
				// Handle batch state updates
				bridge.EventBus.Publish(core.ActionMsg{
					ID:     "bridge",
					Action: MsgTypeStateBatchUpdate,
					Data:   v,
				})

			case core.ProcessResultMsg:
				// Handle process results
				bridge.EventBus.Publish(core.ActionMsg{
					ID:     "bridge",
					Action: MsgTypeProcessResult,
					Data:   v,
				})

			case core.ErrorMessage:
				// Handle errors
				bridge.EventBus.Publish(core.ActionMsg{
					ID:     "bridge",
					Action: MsgTypeError,
					Data:   v,
				})

			default:
				// Log that message type is not supported for forwarding
				// In production, you might want to log this or handle it differently
				// For now, we'll just ignore it
			}
		}
	}()

	return bridge
}

// Send sends a message through the bridge
func (b *Bridge) Send(msg tea.Msg) {
	select {
	case b.MessageChan <- msg:
	default:
		// Channel is full, drop the message to prevent blocking
		// In production, you might want to log this
	}
}

// Model implements the Bubble Tea tea.Model interface.
// It manages the reactive state and handles the message loop.
type Model struct {
	// Config is the parsed .tui.yao configuration
	Config *Config

	// State holds the reactive data, protected by StateMu
	State map[string]interface{}

	// StateMu protects State for concurrent access
	StateMu sync.RWMutex

	// Components holds the runtime instances of components (for backward compatibility)
	Components map[string]*core.ComponentInstance

	// ComponentInstanceRegistry manages component lifecycle and prevents recreation
	ComponentInstanceRegistry *ComponentInstanceRegistry

	// EventBus provides cross-component communication
	EventBus *core.EventBus

	// Bridge provides external message bridge for async operations
	Bridge *Bridge

	// CurrentFocus holds the ID of the currently focused input component
	CurrentFocus string

	// Width is the terminal width
	Width int

	// Height is the terminal height
	Height int

	// Ready indicates whether the TUI is ready to render
	Ready bool

	// AutoFocusApplied indicates whether auto-focus has already been applied
	AutoFocusApplied bool

	// Program is a reference to the Bubble Tea program instance
	// Used for sending messages from external goroutines
	Program *tea.Program

	// MessageHandlers maps message types to their handlers
	MessageHandlers map[string]core.MessageHandler

	// MessageSubscriptionManager manages component message subscriptions
	MessageSubscriptionManager *MessageSubscriptionManager

	// exprCache caches compiled expressions for performance
	exprCache *ExpressionCache

	// logLevel controls the verbosity of logging for this TUI instance
	logLevel string

	// propsCache caches resolved props for performance
	propsCache *PropsCache

	// LayoutEngine is the flexible layout engine for calculating component bounds
	LayoutEngine *layout.Engine

	// LayoutRoot is the root node of the layout tree
	LayoutRoot *layout.LayoutNode

	// Renderer is the layout renderer that renders the layout tree
	Renderer *layout.Renderer

	// ========== Runtime Integration (v1) ==========
	// RuntimeEngine is the new runtime-based layout engine
	RuntimeEngine tuiruntime.Runtime

	// RuntimeRoot is the root node of the runtime layout tree
	RuntimeRoot *tuiruntime.LayoutNode

	// UseRuntime enables the new runtime engine (default: false for compatibility)
	// When true, RuntimeEngine is used instead of LayoutEngine
	UseRuntime bool

	// runtimeFocusList holds focusable component IDs sorted by geometric position
	// This is used when UseRuntime is true for geometry-based focus navigation
	runtimeFocusList []string

	// ========== Render Cache ==========
	// lastRenderedOutput caches the last rendered output string to avoid unnecessary re-renders
	// This prevents flickering when cursor.BlinkMsg triggers View() but nothing changed
	lastRenderedOutput string

	// forceRender forces a full re-render on next View() call
	forceRender bool

	// messageReceived indicates that a message was processed since last render
	// This ensures the UI refreshes after any user interaction
	messageReceived bool

	// ========== Text Selection Support ==========
	// mouseButtonDown tracks which mouse button is currently down (for drag selection)
	mouseButtonDown int

	// lastMouseX, lastMouseY track the last mouse position for drag detection
	lastMouseX int
	lastMouseY int

	// mouseDragStart tracks where a drag selection started
	mouseDragStartX int
	mouseDragStartY int

	// clickCount tracks consecutive clicks for double/triple click detection
	clickCount int

	// lastClickTime tracks the time of the last click for double-click detection
	lastClickTime int64

	// lastClickX, lastClickY track the position of the last click
	lastClickX int
	lastClickY int
}

// Validate validates the Config structure.
// Returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("config.name is required")
	}

	// Normalize direction to support both old and new formats
	if c.Layout.Direction == "" {
		c.Layout.Direction = "vertical" // Default to vertical
	}

	// Support both old (vertical/horizontal) and new (row/column) format
	validDirections := map[string]bool{
		"vertical":   true,
		"horizontal": true,
		"column":     true,
		"row":        true,
	}
	if !validDirections[c.Layout.Direction] {
		return fmt.Errorf("invalid layout.direction: %s (must be one of: vertical, horizontal, column, row)", c.Layout.Direction)
	}

	// Normalize new format to old format for compatibility
	if c.Layout.Direction == "column" {
		c.Layout.Direction = "vertical"
	} else if c.Layout.Direction == "row" {
		c.Layout.Direction = "horizontal"
	}

	// Validate log level
	if c.LogLevel == "" {
		c.LogLevel = "warn" // Default to warn
	} else {
		validLevels := map[string]bool{
			"trace": true,
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
			"none":  true,
		}
		if !validLevels[c.LogLevel] {
			return fmt.Errorf("invalid log level: %s (must be one of: trace, debug, info, warn, error, none)", c.LogLevel)
		}
	}

	// Set default for AutoFocus
	if c.AutoFocus == nil {
		defaultAutoFocus := true
		c.AutoFocus = &defaultAutoFocus
	}

	// Validate components (be lenient - only check for type)
	for i, comp := range c.Layout.Children {
		// Allow extra fields like "width", "height", "align", etc. for future compatibility
		// Only validate that required fields are present
		if comp.Type == "" {
			return fmt.Errorf("component at index %d missing 'type' field", i)
		}
	}

	return nil
}

// logLevelPriority returns the priority of a log level (higher = more important)
func logLevelPriority(level string) int {
	levels := map[string]int{
		"none":  0,
		"error": 1,
		"warn":  2,
		"info":  3,
		"debug": 4,
		"trace": 5,
	}
	return levels[level]
}

// shouldLog checks if a message at the given level should be logged based on the configured level
func (m *Model) shouldLog(level string) bool {
	// If log level is "none", never log
	if m.logLevel == "none" {
		return false
	}
	// Always log errors and warnings unless log level is "none"
	if level == "error" || level == "warn" {
		return true
	}
	// For other levels, check if the message level is >= configured level
	return logLevelPriority(level) <= logLevelPriority(m.logLevel)
}

// Trace logs a trace-level message if the log level allows it
// This is a convenience wrapper to avoid repeated shouldLog checks
func (m *Model) Trace(format string, args ...interface{}) {
	if m.shouldLog("trace") {
		log.Trace(format, args...)
	}
}

// Debug logs a debug-level message if the log level allows it
func (m *Model) Debug(format string, args ...interface{}) {
	if m.shouldLog("debug") {
		log.Debug(format, args...)
	}
}

// Info logs an info-level message if the log level allows it
func (m *Model) Info(format string, args ...interface{}) {
	if m.shouldLog("info") {
		log.Info(format, args...)
	}
}

// Warn logs a warning-level message if the log level allows it
func (m *Model) Warn(format string, args ...interface{}) {
	if m.shouldLog("warn") {
		log.Warn(format, args...)
	}
}

// Error logs an error-level message if the log level allows it
func (m *Model) Error(format string, args ...interface{}) {
	if m.shouldLog("error") {
		log.Error(format, args...)
	}
}

// PropsCache caches resolved props for performance
type PropsCache struct {
	cache map[string]cachedProps
	mu    sync.RWMutex
}

type cachedProps struct {
	originalProps map[string]interface{}
	resolvedProps map[string]interface{}
	state         map[string]interface{}
	time          int64
}

func NewPropsCache() *PropsCache {
	return &PropsCache{
		cache: make(map[string]cachedProps),
	}
}

// GetOrResolve returns cached props if valid, otherwise resolves and caches them
func (pc *PropsCache) GetOrResolve(
	compID string,
	compProps map[string]interface{},
	currentState map[string]interface{},
	resolver func() (map[string]interface{}, error),
) (map[string]interface{}, error) {
	pc.mu.RLock()
	cached, exists := pc.cache[compID]
	pc.mu.RUnlock()

	if exists && mapsEqual(cached.originalProps, compProps) && mapsEqual(cached.state, currentState) {
		return cached.resolvedProps, nil
	}

	resolvedProps, err := resolver()
	if err != nil {
		return nil, err
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache[compID] = cachedProps{
		originalProps: cloneMap(compProps),
		resolvedProps: resolvedProps,
		state:         cloneMap(currentState),
	}

	return resolvedProps, nil
}

// Invalidate removes a cached entry
func (pc *PropsCache) Invalidate(compID string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	delete(pc.cache, compID)
}

// Clear removes all cached entries
func (pc *PropsCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache = make(map[string]cachedProps)
}

// mapsEqual compares two maps for equality
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if bv, exists := b[k]; !exists {
			return false
		} else if !valuesEqual(v, bv) {
			return false
		}
	}
	return true
}

// valuesEqual compares two values for equality (handles nested maps and slices)
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle primitive types
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case int, int8, int16, int32, int64:
		switch bv := b.(type) {
		case int, int8, int16, int32, int64:
			return fmt.Sprintf("%v", av) == fmt.Sprintf("%v", bv)
		default:
			return false
		}
	case float32, float64:
		switch bv := b.(type) {
		case float32, float64:
			return fmt.Sprintf("%v", av) == fmt.Sprintf("%v", bv)
		default:
			return false
		}
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return mapsEqual(av, bv)
	default:
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}
}

// cloneMap creates a shallow copy of a map
func cloneMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
