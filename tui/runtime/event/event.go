package event

import (
	"time"
)

// Node is a minimal interface for event targets.
// Any type can be a node in the event tree.
type Node interface{}

// Event is the interface for all events in the V3 Event System.
// Events follow a three-phase propagation model: Capture → Target → Bubble.
type Event interface {
	// Type returns the event type
	Type() EventType

	// Phase returns the current propagation phase
	Phase() EventPhase

	// Timestamp returns when the event was created
	Timestamp() time.Time

	// Target returns the original target of the event
	Target() Node

	// CurrentTarget returns the current node in the propagation path
	CurrentTarget() Node

	// SetPhase sets the current event phase (used internally during propagation)
	SetPhase(phase EventPhase)

	// PreventDefault marks the event as having its default action prevented
	PreventDefault()

	// IsDefaultPrevented returns true if the default action has been prevented
	IsDefaultPrevented() bool

	// StopPropagation stops event propagation in the bubbling phase
	StopPropagation()

	// IsPropagationStopped returns true if propagation has been stopped
	IsPropagationStopped() bool

	// StopImmediatePropagation stops all further event propagation
	StopImmediatePropagation()

	// IsImmediatePropagationStopped returns true if immediate propagation is stopped
	IsImmediatePropagationStopped() bool
}

// BaseEvent provides a default implementation of the Event interface.
// Specific event types can embed this struct to get common event behavior.
type BaseEvent struct {
	eventType  EventType
	phase       EventPhase
	timestamp   time.Time
	target      Node
	current     Node
	prevented   bool
	stopped     bool
	stoppedImm  bool
}

// NewBaseEvent creates a new BaseEvent with the given type.
func NewBaseEvent(eventType EventType) *BaseEvent {
	return &BaseEvent{
		eventType: eventType,
		timestamp: time.Now(),
		phase:     PhaseNone,
	}
}

// Type returns the event type.
func (e *BaseEvent) Type() EventType {
	return e.eventType
}

// Phase returns the current event phase.
func (e *BaseEvent) Phase() EventPhase {
	return e.phase
}

// Timestamp returns when the event was created.
func (e *BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// Target returns the original target of the event.
func (e *BaseEvent) Target() Node {
	return e.target
}

// CurrentTarget returns the current node in the propagation path.
func (e *BaseEvent) CurrentTarget() Node {
	return e.current
}

// SetTarget sets the target node for this event.
func (e *BaseEvent) SetTarget(target Node) {
	e.target = target
}

// SetCurrentTarget sets the current target node during propagation.
func (e *BaseEvent) SetCurrentTarget(current Node) {
	e.current = current
}

// SetPhase sets the current event phase.
func (e *BaseEvent) SetPhase(phase EventPhase) {
	e.phase = phase
}

// PreventDefault marks the event as having its default action prevented.
func (e *BaseEvent) PreventDefault() {
	e.prevented = true
}

// IsDefaultPrevented returns true if the default action has been prevented.
func (e *BaseEvent) IsDefaultPrevented() bool {
	return e.prevented
}

// StopPropagation stops event propagation in the bubbling phase.
func (e *BaseEvent) StopPropagation() {
	e.stopped = true
}

// IsPropagationStopped returns true if propagation has been stopped.
func (e *BaseEvent) IsPropagationStopped() bool {
	return e.stopped
}

// StopImmediatePropagation stops all further event propagation.
func (e *BaseEvent) StopImmediatePropagation() {
	e.stoppedImm = true
	e.stopped = true
}

// IsImmediatePropagationStopped returns true if immediate propagation is stopped.
func (e *BaseEvent) IsImmediatePropagationStopped() bool {
	return e.stoppedImm
}

// Reset resets the event state for reuse in event pools.
func (e *BaseEvent) Reset() {
	e.phase = PhaseNone
	e.target = nil
	e.current = nil
	e.prevented = false
	e.stopped = false
	e.stoppedImm = false
	e.timestamp = time.Time{}
}
