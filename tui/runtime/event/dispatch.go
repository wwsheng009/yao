package event

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// NOTE: Bubble Tea message conversion has been moved to the adapter layer
// to maintain module boundary rules (runtime MUST NOT import Bubble Tea).
//
// For converting Bubble Tea messages to runtime events, use:
//   github.com/yaoapp/yao/tui/tea/adapter.ConvertBubbleTeaMsg(msg tea.Msg)

// EventPhase represents the phase of event propagation
type EventPhase int

const (
	PhaseNone EventPhase = iota
	PhaseCapturing // Event is propagating down from root to target
	PhaseAtTarget  // Event is at the target node
	PhaseBubbling  // Event is propagating up from target to root
)

// Event is the unified event type that can represent different input types.
type Event struct {
	Type   EventType
	Mouse  *MouseEvent
	Key    *KeyEvent
	Resize *ResizeEvent
	Custom interface{}

	// Propagation control
	Phase          EventPhase
	CurrentTarget  *runtime.LayoutNode
	Target         *runtime.LayoutNode // Original target from hit test
	StoppedPropagation bool
	StoppedImmediatePropagation bool
}

// EventType is the type of event.
type EventType string

const (
	EventTypeMouse  EventType = "mouse"
	EventTypeKey    EventType = "key"
	EventTypeResize EventType = "resize"
	EventTypeCustom EventType = "custom"
)

// StopPropagation stops the event from propagating further in the bubbling phase
func (e *Event) StopPropagation() {
	e.StoppedPropagation = true
}

// StopImmediatePropagation stops the event from propagating to any other listeners
func (e *Event) StopImmediatePropagation() {
	e.StoppedImmediatePropagation = true
	e.StoppedPropagation = true
}

// PreventDefault marks the event as prevented (for potential default action handling)
// In v1, this is informational only
func (e *Event) PreventDefault() {
	// Future: implement default action prevention
}

// ResizeEvent represents a terminal resize event.
type ResizeEvent struct {
	Width  int
	Height int
}

// EventResult contains the result of event processing.
type EventResult struct {
	Handled      bool
	Updated      bool
	FocusChange  FocusChangeType
	FocusTarget  string // Node ID to focus (for mouse clicks)
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
// It implements the complete event propagation model:
//   1. Capturing phase: Event propagates from root to target
//   2. At-target phase: Event is delivered to the target component
//   3. Bubbling phase: Event propagates from target back to root
//
// Returns an EventResult indicating if the event was handled.
func DispatchEvent(ev Event, boxes []runtime.LayoutBox) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	// Initialize event propagation fields
	ev.Target = nil
	ev.CurrentTarget = nil
	ev.Phase = PhaseNone
	ev.StoppedPropagation = false
	ev.StoppedImmediatePropagation = false

	switch ev.Type {
	case EventTypeMouse:
		if ev.Mouse != nil {
			return dispatchMouseEventWithPropagation(ev, ev.Mouse, boxes)
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

// dispatchMouseEventWithPropagation handles mouse events with full propagation support
func dispatchMouseEventWithPropagation(ev Event, mouseEv *MouseEvent, boxes []runtime.LayoutBox) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	// Find the target using hit testing
	hitResult := HitTest(mouseEv.X, mouseEv.Y, boxes)
	if !hitResult.Found {
		return result
	}

	// Set the target
	ev.Target = hitResult.Node
	if ev.Target == nil {
		return result
	}

	// Build the propagation path (from root to target)
	path := buildPropagationPath(ev.Target)
	if len(path) == 0 {
		return result
	}

	// Phase 1: Capturing (root -> target)
	ev.Phase = PhaseCapturing
	for i := 0; i < len(path)-1; i++ {
		if ev.StoppedImmediatePropagation {
			break
		}

		node := path[i]
		ev.CurrentTarget = node

		// Try to handle event during capture phase
		handled := dispatchToNode(ev, mouseEv, node, mouseEv.X, mouseEv.Y)
		if handled {
			result.Handled = true
			result.Updated = true
		}
	}

	// Phase 2: At Target
	if !ev.StoppedImmediatePropagation {
		ev.Phase = PhaseAtTarget
		ev.CurrentTarget = ev.Target

		localX := hitResult.X
		localY := hitResult.Y

		handled := dispatchToNode(ev, mouseEv, ev.Target, localX, localY)
		if handled {
			result.Handled = true
			result.Updated = true

			// If this was a click and the component is focusable, focus it
			if mouseEv.Type == MousePress && mouseEv.Click == MouseLeft {
				if focusable, ok := ev.Target.Component.Instance.(runtime.FocusableComponent); ok {
					if focusable.IsFocusable() {
						result.FocusTarget = hitResult.NodeID
					}
				}
			}
		}
	}

	// Phase 3: Bubbling (target -> root)
	if !ev.StoppedPropagation && !ev.StoppedImmediatePropagation {
		ev.Phase = PhaseBubbling

		// Traverse in reverse (excluding target which we already did)
		for i := len(path) - 2; i >= 0; i-- {
			if ev.StoppedImmediatePropagation {
				break
			}

			node := path[i]
			ev.CurrentTarget = node

			// Calculate local coordinates for this ancestor
			localX := mouseEv.X - node.AbsoluteX
			localY := mouseEv.Y - node.AbsoluteY

			handled := dispatchToNode(ev, mouseEv, node, localX, localY)
			if handled {
				result.Handled = true
				result.Updated = true
			}
		}
	}

	// Clean up
	ev.Phase = PhaseNone
	ev.CurrentTarget = nil

	return result
}

// buildPropagationPath builds the path from root to target for event propagation
func buildPropagationPath(target *runtime.LayoutNode) []*runtime.LayoutNode {
	var path []*runtime.LayoutNode

	// Walk up from target to root
	current := target
	for current != nil {
		path = append([]*runtime.LayoutNode{current}, path...) // Prepend
		current = current.Parent
	}

	return path
}

// dispatchToNode attempts to dispatch an event to a specific node
func dispatchToNode(ev Event, mouseEv *MouseEvent, node *runtime.LayoutNode, localX, localY int) bool {
	if node == nil || node.Component == nil || node.Component.Instance == nil {
		return false
	}

	// Check if component implements mouse event handler
	if mouseHandler, ok := node.Component.Instance.(MouseEventHandler); ok {
		return mouseHandler.HandleMouse(mouseEv, localX, localY)
	}

	// Check for delegated event handlers (future enhancement)
	// if delegated := node.Component.GetDelegatedHandler(ev.Type); delegated != nil {
	//     return delegated(ev, localX, localY)
	// }

	return false
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

// EventPriority defines the priority of event handlers
type EventPriority int

const (
	PriorityLow EventPriority = iota
	PriorityDefault
	PriorityHigh
)

// EventHandlerFunc is a function that can handle events
type EventHandlerFunc func(ev Event) EventResult

// DelegatedEventHandler represents a delegated event handler with metadata
type DelegatedEventHandler struct {
	Handler  EventHandlerFunc
	Priority EventPriority
	Phase    EventPhase // Which phase to handle (0 = any phase)
	Once     bool       // Only handle once then remove
}

// EventDelegator manages event delegation for a component
type EventDelegator struct {
	handlers map[EventType][]*DelegatedEventHandler
}

// NewEventDelegator creates a new event delegator
func NewEventDelegator() *EventDelegator {
	return &EventDelegator{
		handlers: make(map[EventType][]*DelegatedEventHandler),
	}
}

// On adds an event handler with default priority
func (ed *EventDelegator) On(eventType EventType, handler EventHandlerFunc) {
	ed.AddHandler(eventType, handler, PriorityDefault, PhaseNone, false)
}

// OnWithPriority adds an event handler with specified priority
func (ed *EventDelegator) OnWithPriority(eventType EventType, handler EventHandlerFunc, priority EventPriority) {
	ed.AddHandler(eventType, handler, priority, PhaseNone, false)
}

// OnDuringPhase adds an event handler for a specific phase
func (ed *EventDelegator) OnDuringPhase(eventType EventType, handler EventHandlerFunc, phase EventPhase) {
	ed.AddHandler(eventType, handler, PriorityDefault, phase, false)
}

// Once adds an event handler that only executes once
func (ed *EventDelegator) Once(eventType EventType, handler EventHandlerFunc) {
	ed.AddHandler(eventType, handler, PriorityDefault, PhaseNone, true)
}

// AddHandler adds an event handler with full configuration
func (ed *EventDelegator) AddHandler(eventType EventType, handler EventHandlerFunc, priority EventPriority, phase EventPhase, once bool) {
	if ed.handlers == nil {
		ed.handlers = make(map[EventType][]*DelegatedEventHandler)
	}

	delegated := &DelegatedEventHandler{
		Handler:  handler,
		Priority: priority,
		Phase:    phase,
		Once:     once,
	}

	ed.handlers[eventType] = append(ed.handlers[eventType], delegated)

	// Sort handlers by priority (high to low)
	ed.sortHandlers(eventType)
}

// sortHandlers sorts handlers for an event type by priority (high to low)
func (ed *EventDelegator) sortHandlers(eventType EventType) {
	handlers := ed.handlers[eventType]

	// Simple insertion sort (small arrays expected)
	// Sort in descending order (high priority first)
	for i := 1; i < len(handlers); i++ {
		for j := i; j > 0 && handlers[j-1].Priority < handlers[j].Priority; j-- {
			handlers[j-1], handlers[j] = handlers[j], handlers[j-1]
		}
	}
}

// HandleEvent attempts to handle an event using delegated handlers
func (ed *EventDelegator) HandleEvent(ev Event) EventResult {
	result := EventResult{
		Handled:     false,
		Updated:     false,
		FocusChange: FocusChangeNone,
	}

	if ed.handlers == nil {
		return result
	}

	handlers := ed.handlers[ev.Type]
	if len(handlers) == 0 {
		return result
	}

	// Filter by phase if specified
	var toRemove []int
	for i, handler := range handlers {
		// Check phase filter
		if handler.Phase != PhaseNone && handler.Phase != ev.Phase {
			continue
		}

		// Execute handler
		handlerResult := handler.Handler(ev)
		if handlerResult.Handled {
			result.Handled = true
		}
		if handlerResult.Updated {
			result.Updated = true
		}
		if handlerResult.FocusChange != FocusChangeNone {
			result.FocusChange = handlerResult.FocusChange
		}

		// Mark one-time handlers for removal
		if handler.Once {
			toRemove = append(toRemove, i)
		}

		// Stop propagation if handler stopped it
		if ev.StoppedImmediatePropagation {
			break
		}
	}

	// Remove one-time handlers (in reverse order to maintain indices)
	for i := len(toRemove) - 1; i >= 0; i-- {
		idx := toRemove[i]
		ed.handlers[ev.Type] = append(handlers[:idx], handlers[idx+1:]...)
	}

	return result
}

// RemoveAll removes all handlers for an event type
func (ed *EventDelegator) RemoveAll(eventType EventType) {
	if ed.handlers != nil {
		delete(ed.handlers, eventType)
	}
}

// Clear removes all handlers
func (ed *EventDelegator) Clear() {
	ed.handlers = make(map[EventType][]*DelegatedEventHandler)
}

// ComponentTarget wraps a LayoutNode to implement EventTarget.
type ComponentTarget struct {
	Node        *runtime.LayoutNode
	Delegator   *EventDelegator
}

// NewComponentTarget creates a new ComponentTarget.
func NewComponentTarget(node *runtime.LayoutNode) EventTarget {
	return &ComponentTarget{
		Node:      node,
		Delegator: NewEventDelegator(),
	}
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

	// Try delegated handlers first
	if t.Delegator != nil {
		delegatedResult := t.Delegator.HandleEvent(ev)
		if delegatedResult.Handled {
			result.Handled = true
		}
		if delegatedResult.Updated {
			result.Updated = true
		}
		if delegatedResult.FocusChange != FocusChangeNone {
			result.FocusChange = delegatedResult.FocusChange
		}

		// If delegated handler stopped propagation, return early
		if ev.StoppedImmediatePropagation {
			return result
		}
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

// On adds a delegated event handler to this target
func (t *ComponentTarget) On(eventType EventType, handler EventHandlerFunc) {
	if t.Delegator != nil {
		t.Delegator.On(eventType, handler)
	}
}

// Once adds a one-time event handler to this target
func (t *ComponentTarget) Once(eventType EventType, handler EventHandlerFunc) {
	if t.Delegator != nil {
		t.Delegator.Once(eventType, handler)
	}
}
