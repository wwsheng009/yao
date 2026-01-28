package event

import (
	"time"

	"github.com/yaoapp/yao/tui/runtime/event"
)

// ==============================================================================
// Framework Event System - Built on Runtime Events
// ==============================================================================
// Framework events are built on top of runtime events.
// This provides framework-specific features while maintaining compatibility.

// EventType is an alias for runtime.EventType for convenience.
type EventType = event.EventType

// Event constants from runtime
const (
	EventKeyPress     = event.EventKeyPress
	EventKeyRelease   = event.EventKeyRelease
	EventKeyRepeat    = event.EventKeyRepeat
	EventMousePress   = event.EventMousePress
	EventMouseRelease = event.EventMouseRelease
	EventMouseMove    = event.EventMouseMove
	EventMouseWheel   = event.EventMouseWheel
	EventMouseEnter   = event.EventMouseEnter
	EventMouseLeave   = event.EventMouseLeave
	EventResize       = event.EventResize
	EventFocus        = event.EventFocus
	EventBlur         = event.EventBlur
	EventClick        = event.EventClick
	EventDoubleClick  = event.EventDoubleClick
	EventContextMenu  = event.EventContextMenu
	EventChange       = event.EventChange
	EventSubmit       = event.EventSubmit
	EventCancel       = event.EventCancel
	EventSelect       = event.EventSelect
	EventExpand       = event.EventExpand
	EventCollapse     = event.EventCollapse
	EventCustom       = event.EventCustom
	EventQuit         = event.EventQuit
)

// Event is the interface for all framework events.
// It embeds runtime.Event and adds framework-specific features.
type Event interface {
	// Type returns the event type.
	Type() EventType

	// Timestamp returns when the event occurred.
	Timestamp() time.Time

	// Target returns the framework component target.
	Target() Component

	// SetTarget sets the target component.
	SetTarget(target Component)

	// Source returns the event source (optional).
	Source() Component

	// SetSource sets the source component.
	SetSource(source Component)
}

// =============================================================================
// Framework Event Base
// =============================================================================

// BaseEvent provides a base implementation for framework events.
type BaseEvent struct {
	rtEvent  event.Event
	source   Component
	target   Component
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(eventType EventType) *BaseEvent {
	return &BaseEvent{
		rtEvent: event.NewBaseEvent(eventType),
	}
}

// Type returns the event type.
func (e *BaseEvent) Type() EventType {
	return e.rtEvent.Type()
}

// Timestamp returns the event timestamp.
func (e *BaseEvent) Timestamp() time.Time {
	return e.rtEvent.Timestamp()
}

// Target returns the target component.
func (e *BaseEvent) Target() Component {
	return e.target
}

// SetTarget sets the target component.
func (e *BaseEvent) SetTarget(target Component) {
	e.target = target
}

// Source returns the source component.
func (e *BaseEvent) Source() Component {
	return e.source
}

// SetSource sets the source component.
func (e *BaseEvent) SetSource(source Component) {
	e.source = source
}

// RuntimeEvent returns the underlying runtime event.
func (e *BaseEvent) RuntimeEvent() event.Event {
	return e.rtEvent
}

// =============================================================================
// Framework-Specific Event Interface
// =============================================================================

// Component is the interface for event handling components.
type Component interface {
	// HandleEvent processes an event.
	// Returns true if the event was handled, false to continue propagation.
	HandleEvent(Event) bool
}

// =============================================================================
// Event Handler Interface
// =============================================================================

// EventComponent is the interface for components that handle framework events.
// The Component interface already declares HandleEvent, this serves as a named
// marker for type assertions instead of using anonymous interfaces.
type EventComponent interface {
	Component
}

// Handler is the interface for event handlers.
type Handler interface {
	// HandleEvent handles an event.
	// Returns true if the event was handled, false to continue propagation.
	HandleEvent(e Event) bool
}

// HandlerFunc is a function adapter for Handler.
type HandlerFunc func(e Event) bool

// HandleEvent implements the Handler interface.
func (f HandlerFunc) HandleEvent(e Event) bool {
	return f(e)
}

// =============================================================================
// Event Dispatcher
// =============================================================================

// Dispatcher manages event handlers and dispatching.
type Dispatcher struct {
	handlers map[EventType][]Handler
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[EventType][]Handler),
	}
}

// On registers an event handler for the given event type.
func (d *Dispatcher) On(eventType EventType, handler Handler) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// OnFunc registers a function as an event handler.
func (d *Dispatcher) OnFunc(eventType EventType, handler func(Event) bool) {
	d.On(eventType, HandlerFunc(handler))
}

// Dispatch dispatches an event to registered handlers.
// Returns true if the event was handled by any handler.
func (d *Dispatcher) Dispatch(e Event) bool {
	handlers, ok := d.handlers[e.Type()]
	if !ok {
		return false
	}

	for _, handler := range handlers {
		if handler.HandleEvent(e) {
			return true
		}
	}
	return false
}

// RemoveAll removes all handlers for the given event type.
func (d *Dispatcher) RemoveAll(eventType EventType) {
	delete(d.handlers, eventType)
}

// Clear removes all event handlers.
func (d *Dispatcher) Clear() {
	d.handlers = make(map[EventType][]Handler)
}

// =============================================================================
// KeyEvent Specific
// =============================================================================

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	*BaseEvent
	Key      Key
	Special  SpecialKey
	Modifiers Modifier
}

// Key represents a keyboard key.
type Key struct {
	// Rune is the Unicode code point of the key (for printable keys).
	Rune rune

	// Name is the key name (for non-printable keys like "enter", "esc").
	Name string

	// Alt is true if Alt key was pressed.
	Alt bool

	// Ctrl is true if Ctrl key was pressed.
	Ctrl bool
}

// Modifier represents keyboard modifiers.
type Modifier int

const (
	ModNone Modifier = iota
	ModAlt
	ModCtrl
	ModShift
	ModMeta
)

// NewKeyEvent creates a new keyboard event.
func NewKeyEvent(key Key) *KeyEvent {
	ev := &KeyEvent{
		BaseEvent: NewBaseEvent(EventKeyPress),
		Key:       key,
	}
	if key.Ctrl {
		ev.Modifiers |= ModCtrl
	}
	if key.Alt {
		ev.Modifiers |= ModAlt
	}
	return ev
}

// =============================================================================
// MouseEvent Specific
// =============================================================================

// MouseEvent represents a mouse event.
type MouseEvent struct {
	*BaseEvent
	X, Y   int
	Button MouseButton
}

// MouseButton represents a mouse button.
type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
)

// NewMouseEvent creates a new mouse event.
func NewMouseEvent(eventType EventType, x, y int, button MouseButton) *MouseEvent {
	ev := &MouseEvent{
		BaseEvent: NewBaseEvent(eventType),
		X:         x,
		Y:         y,
		Button:    button,
	}
	return ev
}
