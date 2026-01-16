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
	OnLoad *Action `json:"onLoad,omitempty"`

	// Layout defines the UI structure
	Layout Layout `json:"layout,omitempty"`

	// Bindings maps keyboard shortcuts to actions
	Bindings map[string]Action `json:"bindings,omitempty"`
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
	Actions map[string]Action `json:"actions,omitempty"`

	// Height specifies the component height ("flex" for flexible, or number)
	Height interface{} `json:"height,omitempty"`

	// Width specifies the component width
	Width interface{} `json:"width,omitempty"`
}

// Action defines an action to be executed in response to events.
// An action can either call a Yao Process or execute a script method.
type Action struct {
	// Process is the name of the Yao Process to execute
	Process string `json:"process,omitempty"`

	// Script is the path to the script file (e.g., "scripts/tui/handler")
	Script string `json:"script,omitempty"`

	// Method is the method name to call in the script
	Method string `json:"method,omitempty"`

	// Args contains the arguments to pass (supports {{}} expressions)
	Args []interface{} `json:"args,omitempty"`

	// OnSuccess specifies the state key to store the result
	OnSuccess string `json:"onSuccess,omitempty"`

	// OnError specifies the state key to store error information
	OnError string `json:"onError,omitempty"`

	// Payload contains data for direct state updates
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// ProcessResultMsg is sent when a Yao Process execution completes.
type ProcessResultMsg struct {
	// Target is the state key where the result should be stored
	Target string

	// Data is the result data from the Process
	Data interface{}
}

// StateUpdateMsg is sent when a single state key needs to be updated.
type StateUpdateMsg struct {
	// Key is the state key to update
	Key string

	// Value is the new value
	Value interface{}
}

// StateBatchUpdateMsg is sent when multiple state keys need to be updated.
type StateBatchUpdateMsg struct {
	// Updates contains the key-value pairs to update
	Updates map[string]interface{}
}

// StreamChunkMsg is sent when a chunk of streaming data is received (e.g., from AI).
type StreamChunkMsg struct {
	// ID identifies the stream
	ID string

	// Content is the chunk content
	Content string
}

// StreamDoneMsg is sent when a stream completes.
type StreamDoneMsg struct {
	// ID identifies the completed stream
	ID string
}

// ErrorMsg represents an error message with context.
type ErrorMsg struct {
	// Err is the underlying error
	Err error

	// Context describes where the error occurred
	Context string
}

// Error implements the error interface.
func (e ErrorMsg) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[TUI Error in %s] %v", e.Context, e.Err)
	}
	return fmt.Sprintf("[TUI Error] %v", e.Err)
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

	// Width is the terminal width
	Width int

	// Height is the terminal height
	Height int

	// Ready indicates whether the TUI is ready to render
	Ready bool

	// Program is a reference to the Bubble Tea program instance
	// Used for sending messages from external goroutines
	Program *tea.Program
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

// Validate validates the Action structure.
func (a *Action) Validate() error {
	// Must have either Process or Script
	if a.Process == "" && a.Script == "" {
		return fmt.Errorf("action must specify either 'process' or 'script'")
	}

	// If Script is specified, Method must also be specified
	if a.Script != "" && a.Method == "" {
		return fmt.Errorf("action with 'script' must also specify 'method'")
	}

	return nil
}
