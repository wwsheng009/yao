package event

import "fmt"

// EventType represents the type of an event.
type EventType int

const (
	// System events (1000+ range)
	EventInit EventType = 1000 + iota
	EventTick
	EventResize
	EventSignal
	EventQuit

	// Keyboard events (raw, Platform layer, 2000+ range)
	EventKeyPress
	EventKeyRelease
	EventKeyRepeat

	// Mouse events (raw, Platform layer, 3000+ range)
	EventMousePress
	EventMouseRelease
	EventMouseMove
	EventMouseWheel
	EventMouseEnter
	EventMouseLeave

	// Action events (semantic, Framework layer, 4000+ range)
	// These are the result of RawInput → Action transformation
	EventAction

	// Component events (5000+ range)
	EventClick
	EventDoubleClick
	EventContextMenu
	EventChange
	EventFocus
	EventBlur
	EventSubmit
	EventCancel
	EventSelect
	EventExpand
	EventCollapse

	// Custom events (user-defined, 10000+ range)
	EventCustom EventType = 10000
)

// String returns the string representation of the event type.
func (t EventType) String() string {
	switch t {
	case EventInit:
		return "Init"
	case EventTick:
		return "Tick"
	case EventResize:
		return "Resize"
	case EventSignal:
		return "Signal"
	case EventQuit:
		return "Quit"
	case EventKeyPress:
		return "KeyPress"
	case EventKeyRelease:
		return "KeyRelease"
	case EventKeyRepeat:
		return "KeyRepeat"
	case EventMousePress:
		return "MousePress"
	case EventMouseRelease:
		return "MouseRelease"
	case EventMouseMove:
		return "MouseMove"
	case EventMouseWheel:
		return "MouseWheel"
	case EventMouseEnter:
		return "MouseEnter"
	case EventMouseLeave:
		return "MouseLeave"
	case EventAction:
		return "Action"
	case EventClick:
		return "Click"
	case EventDoubleClick:
		return "DoubleClick"
	case EventContextMenu:
		return "ContextMenu"
	case EventChange:
		return "Change"
	case EventFocus:
		return "Focus"
	case EventBlur:
		return "Blur"
	case EventSubmit:
		return "Submit"
	case EventCancel:
		return "Cancel"
	case EventSelect:
		return "Select"
	case EventExpand:
		return "Expand"
	case EventCollapse:
		return "Collapse"
	default:
		if t >= EventCustom {
			return fmt.Sprintf("Custom(%d)", t)
		}
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// IsSystem returns true if this is a system event.
func (t EventType) IsSystem() bool {
	return t >= EventInit && t <= EventQuit
}

// IsKeyboard returns true if this is a keyboard event.
func (t EventType) IsKeyboard() bool {
	return t >= EventKeyPress && t <= EventKeyRepeat
}

// IsMouse returns true if this is a mouse event.
func (t EventType) IsMouse() bool {
	return t >= EventMousePress && t <= EventMouseLeave
}

// IsAction returns true if this is an action event.
func (t EventType) IsAction() bool {
	return t == EventAction
}

// IsComponent returns true if this is a component event.
func (t EventType) IsComponent() bool {
	return t >= EventClick && t <= EventCollapse
}

// IsCustom returns true if this is a custom user-defined event.
func (t EventType) IsCustom() bool {
	return t >= EventCustom
}

// IsInput returns true if this is an input event (keyboard or mouse).
func (t EventType) IsInput() bool {
	return t.IsKeyboard() || t.IsMouse()
}
