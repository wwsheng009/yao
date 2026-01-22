package event

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// NOTE: Bubble Tea message conversion has been moved to the adapter layer
// to maintain module boundary rules (runtime MUST NOT import Bubble Tea).
//
// For converting Bubble Tea messages to runtime events, use:
//   github.com/yaoapp/yao/tui/tea/adapter.ConvertBubbleTeaMsg(msg tea.Msg)

// Event is the unified event type that can represent different input types.
type Event struct {
	Type   EventType
	Mouse  *MouseEvent
	Key    *KeyEvent
	Resize *ResizeEvent
	Custom interface{}
}

// EventType is the type of event.
type EventType string

const (
	EventTypeMouse  EventType = "mouse"
	EventTypeKey    EventType = "key"
	EventTypeResize EventType = "resize"
	EventTypeCustom EventType = "custom"
)

// ResizeEvent represents a terminal resize event.
type ResizeEvent struct {
	Width  int
	Height int
}

// EventResult contains the result of event processing.
type EventResult struct {
	Handled     bool
	Updated     bool
	FocusChange FocusChangeType
}

// FocusChangeType indicates how focus changed.
type FocusChangeType int

const (
	FocusChangeNone   FocusChangeType = iota
	FocusChangeNext
	FocusChangePrev
	FocusChangeSpecific
)

// DispatchEvent routes an event to the appropriate component based on hit testing.
// It performs a hit test to find the target component, then attempts to deliver the event.
// Returns an EventResult indicating if the event was handled.
func DispatchEvent(ev Event, boxes []runtime.LayoutBox) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	switch ev.Type {
	case EventTypeMouse:
		if ev.Mouse != nil {
			return dispatchMouseEvent(ev.Mouse, boxes)
		}
	case EventTypeKey:
		if ev.Key != nil {
			return dispatchKeyEvent(ev.Key, boxes)
		}
	case EventTypeResize:
		if ev.Resize != nil {
			return dispatchResizeEvent(ev.Resize, boxes)
		}
	}

	return result
}

// dispatchMouseEvent handles mouse events using hit testing.
func dispatchMouseEvent(ev *MouseEvent, boxes []runtime.LayoutBox) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	// Find the target using hit testing
	hitResult := HitTest(ev.X, ev.Y, boxes)
	if !hitResult.Found {
		return result
	}

	// Try to handle the event through the component
	node := hitResult.Node
	if node != nil && node.Component != nil && node.Component.Instance != nil {
		// Check if component handles mouse events
		if mouseHandler, ok := node.Component.Instance.(MouseEventHandler); ok {
			handled := mouseHandler.HandleMouse(ev, hitResult.X, hitResult.Y)
			if handled {
				result.Handled = true
				result.Updated = true

				// If this was a click and the component is focusable, focus it
				if ev.Type == MousePress && ev.Click == MouseLeft {
					if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
						if focusable.IsFocusable() {
							// Focus would be set here, but we defer to the focus manager
						}
					}
				}
			}
		}
	}

	return result
}

// dispatchKeyEvent handles keyboard events.
func dispatchKeyEvent(ev *KeyEvent, boxes []runtime.LayoutBox) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	// Check for navigation keys
	switch ev.Key {
	case '\t': // Tab key
		result.FocusChange = FocusChangeNext
		result.Handled = true
		return result
	case 20, 23: // Shift+Tab
		result.FocusChange = FocusChangePrev
		result.Handled = true
		return result
	}

	// For other keys, we'd need to know the currently focused component
	// This is handled by the FocusManager
	return result
}

// dispatchResizeEvent handles resize events.
func dispatchResizeEvent(ev *ResizeEvent, boxes []runtime.LayoutBox) EventResult {
	result := EventResult{
		Handled:     true,
		Updated:     true,
		FocusChange: FocusChangeNone,
	}

	// Resize events are typically handled by the runtime itself
	// to trigger a re-layout
	return result
}

// MouseEventHandler is an interface for components that handle mouse events.
type MouseEventHandler interface {
	// HandleMouse handles a mouse event at the given local coordinates.
	// Returns true if the event was handled.
	HandleMouse(ev *MouseEvent, localX, localY int) bool
}

// KeyEventHandler is an interface for components that handle keyboard events.
type KeyEventHandler interface {
	// HandleKey handles a keyboard event.
	// Returns true if the event was handled.
	HandleKey(ev *KeyEvent) bool
}

// EventHandler combines both mouse and keyboard event handling.
type EventHandler interface {
	MouseEventHandler
	KeyEventHandler
}

// EventTarget represents something that can receive events.
type EventTarget interface {
	// SendEvent sends an event to this target.
	SendEvent(ev Event) EventResult
}

// ComponentTarget wraps a LayoutNode to implement EventTarget.
type ComponentTarget struct {
	Node *runtime.LayoutNode
}

// NewComponentTarget creates a new ComponentTarget.
func NewComponentTarget(node *runtime.LayoutNode) EventTarget {
	return &ComponentTarget{Node: node}
}

// SendEvent sends an event to the wrapped component.
func (t *ComponentTarget) SendEvent(ev Event) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	if t.Node == nil || t.Node.Component == nil {
		return result
	}

	instance := t.Node.Component.Instance
	if instance == nil {
		return result
	}

	switch ev.Type {
	case EventTypeMouse:
		if ev.Mouse != nil {
			if handler, ok := instance.(MouseEventHandler); ok {
				// Calculate local coordinates (assuming node is at origin for this target)
				if handler.HandleMouse(ev.Mouse, 0, 0) {
					result.Handled = true
					result.Updated = true
				}
			}
		}
	case EventTypeKey:
		if ev.Key != nil {
			if handler, ok := instance.(KeyEventHandler); ok {
				if handler.HandleKey(ev.Key) {
					result.Handled = true
					result.Updated = true
				}
			}
		}
	}

	return result
}
