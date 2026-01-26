package event

import (
	"context"
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

// ==============================================================================
// Event Context (for Filter System)
// ==============================================================================

// Context 事件上下文，用于过滤器系统
type Context struct {
	// context Go 标准上下文
	context context.Context

	// source 事件源
	source string

	// metadata 元数据
	metadata map[string]interface{}
}

// NewContext 创建事件上下文
func NewContext() *Context {
	return &Context{
		context:  context.Background(),
		metadata: make(map[string]interface{}),
	}
}

// NewContextWithContext 创建带 Go 上下文的事件上下文
func NewContextWithContext(ctx context.Context) *Context {
	return &Context{
		context:  ctx,
		metadata: make(map[string]interface{}),
	}
}

// Context 返回 Go 标准上下文
func (c *Context) Context() context.Context {
	if c == nil {
		return context.Background()
	}
	return c.context
}

// WithContext 设置 Go 标准上下文
func (c *Context) WithContext(ctx context.Context) *Context {
	c.context = ctx
	return c
}

// Source 返回事件源
func (c *Context) Source() string {
	if c == nil {
		return ""
	}
	return c.source
}

// WithSource 设置事件源
func (c *Context) WithSource(source string) *Context {
	c.source = source
	return c
}

// Get 获取元数据
func (c *Context) Get(key string) interface{} {
	if c == nil || c.metadata == nil {
		return nil
	}
	return c.metadata[key]
}

// Set 设置元数据
func (c *Context) Set(key string, value interface{}) {
	if c.metadata == nil {
		c.metadata = make(map[string]interface{})
	}
	c.metadata[key] = value
}

// With 设置元数据（链式调用）
func (c *Context) With(key string, value interface{}) *Context {
	c.Set(key, value)
	return c
}

// Clone 克隆上下文
func (c *Context) Clone() *Context {
	if c == nil {
		return NewContext()
	}

	metadata := make(map[string]interface{}, len(c.metadata))
	for k, v := range c.metadata {
		metadata[k] = v
	}

	return &Context{
		context:  c.context,
		source:   c.source,
		metadata: metadata,
	}
}
