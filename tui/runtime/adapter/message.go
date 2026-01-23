package adapter

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/event"
)

// MessageConverter converts Bubble Tea messages to Runtime events.
// This bridges the gap between the Bubble Tea framework and the Runtime engine.
type MessageConverter struct {
	// lastMousePosition tracks mouse position for click detection
	lastMousePress *event.MouseEvent
}

// NewMessageConverter creates a new message converter.
func NewMessageConverter() *MessageConverter {
	return &MessageConverter{}
}

// Convert converts a Bubble Tea message to a Runtime event.
// Returns (event, ok) where ok is true if the message was converted.
func (c *MessageConverter) Convert(msg tea.Msg) (event.Event, bool) {
	if msg == nil {
		return event.Event{}, false
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		return c.convertKeyMsg(m), true

	case tea.MouseMsg:
		return c.convertMouseMsg(m), true

	case tea.WindowSizeMsg:
		return c.convertWindowSizeMsg(m), true

	default:
		// Check for specific message types
		return c.convertCustomMsg(msg)
	}
}

// convertKeyMsg converts a Bubble Tea KeyMsg to a Runtime KeyEvent.
func (c *MessageConverter) convertKeyMsg(msg tea.KeyMsg) event.Event {
	// Check for special keys first
	switch msg.Type {
	case tea.KeyTab:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  '\t',
				Type: event.KeyPress,
			},
		}

	case tea.KeyShiftTab:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  '\t',
				Type: event.KeyPress,
				Mod:  event.ModShift,
			},
		}

	case tea.KeyEnter:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  '\r',
				Type: event.KeyPress,
			},
		}

	case tea.KeyBackspace:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  '\b',
				Type: event.KeyPress,
			},
		}

	case tea.KeyDelete:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  '\x7f',
				Type: event.KeyPress,
			},
		}

	case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
		// Arrow keys - map to special runes for navigation
		var key rune
		switch msg.Type {
		case tea.KeyUp:
			key = '↑'
		case tea.KeyDown:
			key = '↓'
		case tea.KeyLeft:
			key = '←'
		case tea.KeyRight:
			key = '→'
		}
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  key,
				Type: event.KeyPress,
			},
		}

	case tea.KeyEscape:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key:  '\x1b',
				Type: event.KeyPress,
			},
		}

	case tea.KeyRunes:
		// Regular character input
		if len(msg.Runes) > 0 {
			key := msg.Runes[0]

			// Check for Alt modifier
			var mod event.KeyModifier
			if msg.Alt {
				mod |= event.ModAlt
			}

			return event.Event{
				Type: event.EventTypeKey,
				Key: &event.KeyEvent{
					Key:  key,
					Type: event.KeyPress,
					Mod:  mod,
				},
			}
		}
	}

	// Default: return as Tab for unknown keys (safe default for navigation)
	return event.Event{
		Type: event.EventTypeKey,
		Key: &event.KeyEvent{
			Key:  '\t',
			Type: event.KeyPress,
		},
	}
}

// convertMouseMsg converts a Bubble Tea MouseMsg to a Runtime MouseEvent.
func (c *MessageConverter) convertMouseMsg(msg tea.MouseMsg) event.Event {
	x, y := msg.X, msg.Y

	// Determine mouse event type
	var eventType event.MouseEventType
	var click event.MouseClickType

	// Use the new API (Action + Button) instead of deprecated Type
	switch msg.Action {
	case tea.MouseActionPress:
		eventType = event.MousePress
		switch msg.Button {
		case tea.MouseButtonLeft:
			click = event.MouseLeft
		case tea.MouseButtonRight:
			click = event.MouseRight
		case tea.MouseButtonMiddle:
			click = event.MouseMiddle
		}
		// Track press for click detection
		c.lastMousePress = &event.MouseEvent{
			X:     x,
			Y:     y,
			Type:  eventType,
			Click: click,
		}

	case tea.MouseActionRelease:
		eventType = event.MouseRelease
		// Check if this is a release after a press (click)
		if c.lastMousePress != nil {
			// Calculate if this was a click (press and release in same area)
			dx := x - c.lastMousePress.X
			dy := y - c.lastMousePress.Y
			if dx*dx+dy*dy <= 25 { // Within 5 pixels
				// This is a click, use the press type
				click = c.lastMousePress.Click
			}
			c.lastMousePress = nil
		}

	case tea.MouseActionMotion:
		eventType = event.MouseMove

	default:
		// Unknown action, default to press
		eventType = event.MousePress
	}

	// Handle wheel events specially
	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
		eventType = event.MouseScroll
	}

	return event.Event{
		Type: event.EventTypeMouse,
		Mouse: &event.MouseEvent{
			X:     x,
			Y:     y,
			Type:  eventType,
			Click: click,
		},
	}
}

// convertWindowSizeMsg converts a Bubble Tea WindowSizeMsg to a Runtime ResizeEvent.
func (c *MessageConverter) convertWindowSizeMsg(msg tea.WindowSizeMsg) event.Event {
	return event.Event{
		Type: event.EventTypeResize,
		Resize: &event.ResizeEvent{
			Width:  msg.Width,
			Height: msg.Height,
		},
	}
}

// convertCustomMsg attempts to convert custom message types.
func (c *MessageConverter) convertCustomMsg(msg tea.Msg) (event.Event, bool) {
	// Check for focus messages
	switch msg.(type) {
	case interface{ FocusNext() bool }:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key: '\t',
				Type: event.KeyPress,
			},
		}, true

	case interface{ FocusPrev() bool }:
		return event.Event{
			Type: event.EventTypeKey,
			Key: &event.KeyEvent{
				Key: '\t',
				Type: event.KeyPress,
				Mod: event.ModShift,
			},
		}, true
	}

	return event.Event{}, false
}

// ToRuntimeEvent converts a Bubble Tea message to a runtime.Event for the RuntimeImpl.Dispatch method.
// This is a compatibility layer for the existing Runtime interface.
func ToRuntimeEvent(msg tea.Msg) runtime.Event {
	var evType string
	var x, y int
	var data interface{}

	switch m := msg.(type) {
	case tea.KeyMsg:
		evType = "key"
		runes := m.Runes
		if len(runes) > 0 {
			data = runes[0]
		} else if m.Type == tea.KeyTab {
			data = '\t'
		} else if m.Type == tea.KeyEnter {
			data = '\r'
		}

	case tea.MouseMsg:
		evType = "mouse"
		x, y = m.X, m.Y
		// Use String() method for representation
		data = m.String()

	case tea.WindowSizeMsg:
		evType = "resize"
		data = struct{ W, H int }{W: m.Width, H: m.Height}
	}

	return runtime.Event{
		X:    x,
		Y:    y,
		Type: evType,
		Data: data,
	}
}
