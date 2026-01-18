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
	"github.com/yaoapp/yao/tui/core"
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

	// Actions maps event names to actions for this component
	Actions map[string]core.Action `json:"actions,omitempty"`

	// Height specifies the component height ("flex" for flexible, or number)
	Height interface{} `json:"height,omitempty"`

	// Width specifies the component width
	Width interface{} `json:"width,omitempty"`
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

	// Program is a reference to the Bubble Tea program instance
	// Used for sending messages from external goroutines
	Program *tea.Program

	// MessageHandlers maps message types to their handlers
	MessageHandlers map[string]core.MessageHandler
}

// Validate validates the Config structure.
// Returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("config.name is required")
	}

	if c.Layout.Direction == "" {
		c.Layout.Direction = "vertical" // Default to vertical
	}

	if c.Layout.Direction != "vertical" && c.Layout.Direction != "horizontal" {
		return fmt.Errorf("invalid layout.direction: %s (must be 'vertical' or 'horizontal')", c.Layout.Direction)
	}

	// Validate components
	for i, comp := range c.Layout.Children {
		if comp.Type == "" {
			return fmt.Errorf("component at index %d missing 'type' field", i)
		}
	}

	return nil
}
