package event

// CaptureHandler is the interface for components that handle events during the capture phase.
// The capture phase occurs first, propagating from the root down to the target.
// This is useful for global shortcuts, interception, and ancestor-level event handling.
type CaptureHandler interface {
	// HandleCapture handles an event during the capture phase.
	// Returning true stops further propagation through the capture chain.
	HandleCapture(ev EventStruct) bool
}

// TargetHandler is the interface for components that handle events at the target phase.
// The target phase occurs when the event reaches its target component.
// This is where most component-specific event handling happens.
type TargetHandler interface {
	// HandleTarget handles an event at the target phase.
	// Returning true indicates the event was handled.
	HandleTarget(ev EventStruct) bool
}

// BubbleHandler is the interface for components that handle events during the bubble phase.
// The bubble phase occurs last, propagating from the target back up to the root.
// This is useful for ancestor components to handle events that weren't handled by descendants.
type BubbleHandler interface {
	// HandleBubble handles an event during the bubble phase.
	// Returning true stops further propagation through the bubble chain.
	HandleBubble(ev EventStruct) bool
}

// PhaseEventHandler combines all three phase handler interfaces.
// Components can implement this single interface instead of implementing each phase separately.
type PhaseEventHandler interface {
	CaptureHandler
	TargetHandler
	BubbleHandler
}

// HandlerFunc is a function adapter that allows plain functions to act as event handlers.
type HandlerFunc func(ev EventStruct) bool

// HandleCapture implements CaptureHandler for HandlerFunc.
func (f HandlerFunc) HandleCapture(ev EventStruct) bool {
	return f(ev)
}

// HandleTarget implements TargetHandler for HandlerFunc.
func (f HandlerFunc) HandleTarget(ev EventStruct) bool {
	return f(ev)
}

// HandleBubble implements BubbleHandler for HandlerFunc.
func (f HandlerFunc) HandleBubble(ev EventStruct) bool {
	return f(ev)
}

// CaptureHandlerFunc is a function adapter specifically for capture phase handling.
type CaptureHandlerFunc func(ev EventStruct) bool

// HandleCapture implements CaptureHandler for CaptureHandlerFunc.
func (f CaptureHandlerFunc) HandleCapture(ev EventStruct) bool {
	return f(ev)
}

// TargetHandlerFunc is a function adapter specifically for target phase handling.
type TargetHandlerFunc func(ev EventStruct) bool

// HandleTarget implements TargetHandler for TargetHandlerFunc.
func (f TargetHandlerFunc) HandleTarget(ev EventStruct) bool {
	return f(ev)
}

// BubbleHandlerFunc is a function adapter specifically for bubble phase handling.
type BubbleHandlerFunc func(ev EventStruct) bool

// HandleBubble implements BubbleHandler for BubbleHandlerFunc.
func (f BubbleHandlerFunc) HandleBubble(ev EventStruct) bool {
	return f(ev)
}
