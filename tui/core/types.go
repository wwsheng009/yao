package core

import (
	"fmt"
	"reflect"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Response 消息处理结果
type Response int

const (
	Ignored   Response = iota // 组件不感兴趣，消息应继续传递
	Handled                   // 组件已处理并截获，消息停止分发
	PassClick                 // 专门用于鼠标事件：处理了点击，但允许透传
)

// MessageHandler defines a function that handles a specific message type
// and returns an updated model and command
type MessageHandler func(interface{}, tea.Msg) (tea.Model, tea.Cmd)

// ComponentInterface 增强型组件接口
type ComponentInterface interface {
	Init() tea.Cmd
	// 返回处理后的组件、命令以及处理状态
	UpdateMsg(msg tea.Msg) (ComponentInterface, tea.Cmd, Response)
	View() string
	GetID() string       // 返回组件的唯一标识符
	SetFocus(focus bool) // 设置/取消焦点
}

// TargetedMsg 用于定向消息投递
type TargetedMsg struct {
	TargetID string
	InnerMsg tea.Msg
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

	// Error contains any error from the process execution
	Error error `json:"error,omitempty"`
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

// InputUpdateMsg is sent to update an input component.
type InputUpdateMsg struct {
	// ID is the input component ID
	ID string
	// Value is the new value for the input
	Value string
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

// ErrorMessage represents an error message with context.
type ErrorMessage struct {
	// Err is the underlying error
	Err error

	// Context describes where the error occurred
	Context string

	// LogLevel is the log level for this error
	LogLevel string `json:"log_level,omitempty"`
}

// Error implements the error interface.
func (e ErrorMessage) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[TUI Error in %s] %v", e.Context, e.Err)
	}
	return fmt.Sprintf("[TUI Error] %v", e.Err)
}

// QuitMsg is sent to request the TUI to quit.
type QuitMsg struct{}

// RefreshMsg is sent to request a UI refresh
type RefreshMsg struct{}

// LogMsg is sent to log a message
type LogMsg struct {
	Level   string
	Message string
}

// ComponentInstance represents a runtime instance of a component with its own UID
type ComponentInstance struct {
	ID       string
	Type     string
	Instance ComponentInterface
}

// ActionMsg represents an internal action message for cross-component communication
type ActionMsg struct {
	ID     string      // Trigger ID
	Action string      // Action name like "SAVE_SUCCESS", "ROW_SELECTED"
	Data   interface{} // Associated data
}

// EventBus provides a simple event bus for component communication
type EventBus struct {
	Subscribers map[string][]func(ActionMsg)
	mu          sync.RWMutex
}

// NewEventBus creates a new EventBus instance
func NewEventBus() *EventBus {
	return &EventBus{
		Subscribers: make(map[string][]func(ActionMsg)),
	}
}

// Subscribe registers a callback for a specific action
// Returns an unsubscribe function that should be called to clean up
func (eb *EventBus) Subscribe(action string, callback func(ActionMsg)) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.Subscribers[action] == nil {
		eb.Subscribers[action] = []func(ActionMsg){}
	}

	// Add callback to list
	eb.Subscribers[action] = append(eb.Subscribers[action], callback)

	// Return unsubscribe function
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()

		callbacks := eb.Subscribers[action]
		for i, cb := range callbacks {
			// Compare function pointers (not perfect but works for most cases)
			if reflect.ValueOf(cb).Pointer() == reflect.ValueOf(callback).Pointer() {
				// Remove callback by slicing
				eb.Subscribers[action] = append(callbacks[:i], callbacks[i+1:]...)
				break
			}
		}

		// Clean up empty action list
		if len(eb.Subscribers[action]) == 0 {
			delete(eb.Subscribers, action)
		}
	}
}

// Publish sends an action message to all subscribers
func (eb *EventBus) Publish(msg ActionMsg) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if callbacks, exists := eb.Subscribers[msg.Action]; exists {
		for _, callback := range callbacks {
			callback(msg)
		}
	}
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

// GetDefaultMessageHandlers returns the default message handlers for the TUI model
func GetDefaultMessageHandlers() map[string]MessageHandler {
	handlers := make(map[string]MessageHandler)

	// This is a placeholder that will be implemented by the consumer
	// The actual implementation will depend on the specific model type

	return handlers
}

// EvaluateExpressions evaluates {{}} expressions in the given value
// against the model's state and returns the resolved value
func EvaluateExpressions(value interface{}, getStateFunc func(string) (interface{}, bool)) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Check if it's an expression like {{key}}
		if len(v) >= 4 && v[0:2] == "{{" && v[len(v)-2:] == "}}" {
			// Extract the key
			key := v[2 : len(v)-2]
			key = TrimWhitespace(key)

			// Get value from state
			stateValue, exists := getStateFunc(key)

			if !exists {
				return nil, fmt.Errorf("state key '%s' not found", key)
			}

			return stateValue, nil
		}
		return v, nil

	case map[string]interface{}:
		// Recursively evaluate expressions in map values
		result := make(map[string]interface{})
		for k, val := range v {
			evaluated, err := EvaluateExpressions(val, getStateFunc)
			if err != nil {
				return nil, err
			}
			result[k] = evaluated
		}
		return result, nil

	case []interface{}:
		// Recursively evaluate expressions in slice elements
		result := make([]interface{}, len(v))
		for i, val := range v {
			evaluated, err := EvaluateExpressions(val, getStateFunc)
			if err != nil {
				return nil, err
			}
			result[i] = evaluated
		}
		return result, nil

	default:
		// For other types (numbers, booleans, etc.), return as-is
		return v, nil
	}
}

// TrimWhitespace removes leading and trailing whitespace
func TrimWhitespace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// MenuItemInterface defines the interface for menu items
// This allows different implementations (e.g., components.MenuItem) to be used
type MenuItemInterface interface {
	GetTitle() string
	GetDescription() string
	GetValue() interface{}
	GetAction() map[string]interface{}
	IsDisabled() bool
	IsSelected() bool
	HasChildren() bool
}

// MenuActionTriggered is sent when a menu item's action is triggered
type MenuActionTriggered struct {
	Item   MenuItemInterface      `json:"item"`
	Action map[string]interface{} `json:"action"`
}

// Standard event actions for component communication
const (
	// Form events
	EventFormSubmitSuccess   = "FORM_SUBMIT_SUCCESS"
	EventFormCancel          = "FORM_CANCEL"
	EventFormValidationError = "FORM_VALIDATION_ERROR"

	// Table events
	EventRowSelected      = "ROW_SELECTED"
	EventRowDoubleClicked = "ROW_DOUBLE_CLICKED"
	EventCellEdited       = "CELL_EDITED"

	// CRUD events
	EventNewItemRequested = "NEW_ITEM_REQUESTED"
	EventItemDeleted      = "ITEM_DELETED"
	EventItemSaved        = "ITEM_SAVED"

	// Navigation events
	EventFocusChanged  = "FOCUS_CHANGED"
	EventTabPressed    = "TAB_PRESSED"
	EventEscapePressed = "ESCAPE_PRESSED"

	// Menu events
	EventMenuItemSelected    = "MENU_ITEM_SELECTED"
	EventMenuActionTriggered = "MENU_ACTION_TRIGGERED"
	EventMenuSubmenuEntered  = "MENU_SUBMENU_ENTERED"
	EventMenuSubmenuExited   = "MENU_SUBMENU_EXITED"

	// Input events
	EventInputValueChanged = "INPUT_VALUE_CHANGED"
	EventInputFocusChanged = "INPUT_FOCUS_CHANGED"
	EventInputEnterPressed = "INPUT_ENTER_PRESSED"

	// Chat events
	EventChatMessageSent     = "CHAT_MESSAGE_SENT"
	EventChatMessageReceived = "CHAT_MESSAGE_RECEIVED"
	EventChatTypingStarted   = "CHAT_TYPING_STARTED"
	EventChatTypingStopped   = "CHAT_TYPING_STOPPED"

	// Data events
	EventDataLoaded    = "DATA_LOADED"
	EventDataRefreshed = "DATA_REFRESHED"
	EventDataFiltered  = "DATA_FILTERED"

	// UI events
	EventUIResized      = "UI_RESIZED"
	EventUIThemeChanged = "UI_THEME_CHANGED"
)

// PublishEvent creates a tea.Cmd that publishes an action message
// Components can return this command from their UpdateMsg method to publish events
func PublishEvent(componentID string, action string, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{
			ID:     componentID,
			Action: action,
			Data:   data,
		}
	}
}

// SubscribeToEvent subscribes a callback to an event on the event bus
// Returns a function to unsubscribe (call to clean up)
func SubscribeToEvent(bus *EventBus, action string, callback func(ActionMsg)) func() {
	// Subscribe and return the unsubscribe function directly
	return bus.Subscribe(action, callback)
}
