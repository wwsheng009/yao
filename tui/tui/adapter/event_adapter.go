// Package adapter provides conversion between Bubble Tea messages and runtime events.
// This layer isolates the runtime package from direct Bubble Tea dependencies,
// maintaining the module boundary rule that runtime MUST NOT import Bubble Tea.
package adapter

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/runtime/event"
)

// ConvertBubbleTeaMsg converts a Bubble Tea message to a runtime Event.
// This bridges the legacy Bubble Tea event system with the new runtime event system.
//
// This function is the ONLY place where Bubble Tea message types are directly
// handled in the entire TUI system. All other code uses the runtime.Event type.
func ConvertBubbleTeaMsg(msg tea.Msg) *event.EventStruct {
	switch m := msg.(type) {
	case tea.MouseMsg:
		x, y := m.X, m.Y
		ev := &event.EventStruct{}
		ev.SetType(event.EventMousePress)
		ev.Mouse = &event.MouseEvent{
			X:     x,
			Y:     y,
			Type:  mouseActionToType(m),
			Click: mouseButtonToType(m),
		}
		return ev
	case tea.KeyMsg:
		// KeyMsg in newer Bubble Tea versions has Type and Runes fields
		var keyRune rune
		if len(m.Runes) > 0 {
			keyRune = m.Runes[0]
		} else {
			// For special keys, use a recognizable rune
			switch m.Type {
			case tea.KeyEnter:
				keyRune = '\r'
			case tea.KeyTab:
				keyRune = '\t'
			case tea.KeyEscape:
				keyRune = 27
			case tea.KeySpace:
				keyRune = ' '
			case tea.KeyBackspace:
				keyRune = 127
			case tea.KeyUp:
				keyRune = 16 + 1 // Use custom codes for arrow keys
			case tea.KeyDown:
				keyRune = 16 + 2
			case tea.KeyLeft:
				keyRune = 16 + 3
			case tea.KeyRight:
				keyRune = 16 + 4
			}
		}
		ev := &event.EventStruct{}
		ev.SetType(event.EventKeyPress)
		ev.Key = &event.KeyEvent{
			Key:  keyRune,
			Type: event.KeyPress,
			Mod:  keyModifierFromMsg(m),
		}
		return ev
	case tea.WindowSizeMsg:
		ev := &event.EventStruct{}
		ev.SetType(event.EventResize)
		ev.Resize = &event.ResizeEvent{
			Width:  m.Width,
			Height: m.Height,
		}
		return ev
	default:
		ev := &event.EventStruct{}
		ev.SetType(event.EventCustom)
		ev.Custom = msg
		return ev
	}
}

// mouseActionToType converts Bubble Tea mouse action to runtime event type.
func mouseActionToType(msg tea.MouseMsg) event.MouseEventType {
	// Bubble Tea v1.3+ uses MouseMsg struct with Type field
	// Type is an int16: 1=press, 2=release, 3=motion, 4=scroll
	switch msg.Type {
	case 1:
		return event.MousePress
	case 2:
		return event.MouseRelease
	case 3:
		return event.MouseMove
	case 4:
		return event.MouseScroll
	default:
		return event.MousePress
	}
}

// mouseButtonToType converts Bubble Tea mouse button to runtime click type.
func mouseButtonToType(msg tea.MouseMsg) event.MouseClickType {
	// Bubble Tea v1.3+ uses MouseMsg struct with Button field
	// Button: 1=left, 2=middle, 3=right, 4=release, etc.
	switch msg.Button {
	case 1:
		return event.MouseLeft
	case 2:
		return event.MouseMiddle
	case 3:
		return event.MouseRight
	default:
		return event.MouseLeft
	}
}

// keyModifierFromMsg extracts key modifiers from a Bubble Tea message.
func keyModifierFromMsg(msg tea.KeyMsg) event.KeyModifier {
	// Bubble Tea doesn't directly expose modifiers
	// We'd need to parse this from the key string
	return event.ModNone
}

// EventAdapter provides a bridge between Bubble Tea and runtime events.
// External code (like Bubble Tea programs) should use this adapter
// to convert Bubble Tea messages before passing them to the runtime.
type EventAdapter struct{}

// NewEventAdapter creates a new event adapter.
func NewEventAdapter() *EventAdapter {
	return &EventAdapter{}
}

// Convert converts a Bubble Tea message to a runtime event.
func (a *EventAdapter) Convert(msg tea.Msg) *event.EventStruct {
	return ConvertBubbleTeaMsg(msg)
}

// ConvertMouseMsg converts a Bubble Tea mouse message.
func (a *EventAdapter) ConvertMouseMsg(msg tea.MouseMsg) *event.EventStruct {
	ev := &event.EventStruct{}
	ev.SetType(event.EventMousePress)
	ev.Mouse = &event.MouseEvent{
		X:     msg.X,
		Y:     msg.Y,
		Type:  mouseActionToType(msg),
		Click: mouseButtonToType(msg),
	}
	return ev
}

// ConvertKeyMsg converts a Bubble Tea key message.
func (a *EventAdapter) ConvertKeyMsg(msg tea.KeyMsg) *event.EventStruct {
	var keyRune rune
	if len(msg.Runes) > 0 {
		keyRune = msg.Runes[0]
	} else {
		switch msg.Type {
		case tea.KeyEnter:
			keyRune = '\r'
		case tea.KeyTab:
			keyRune = '\t'
		case tea.KeyEscape:
			keyRune = 27
		case tea.KeySpace:
			keyRune = ' '
		case tea.KeyBackspace:
			keyRune = 127
		case tea.KeyUp:
			keyRune = 16 + 1
		case tea.KeyDown:
			keyRune = 16 + 2
		case tea.KeyLeft:
			keyRune = 16 + 3
		case tea.KeyRight:
			keyRune = 16 + 4
		default:
			keyRune = 0
		}
	}
	ev := &event.EventStruct{}
	ev.SetType(event.EventKeyPress)
	ev.Key = &event.KeyEvent{
		Key:  keyRune,
		Type: event.KeyPress,
		Mod:  event.ModNone,
	}
	return ev
}

// ConvertWindowSizeMsg converts a Bubble Tea window size message.
func (a *EventAdapter) ConvertWindowSizeMsg(msg tea.WindowSizeMsg) *event.EventStruct {
	ev := &event.EventStruct{}
	ev.SetType(event.EventResize)
	ev.Resize = &event.ResizeEvent{
		Width:  msg.Width,
		Height: msg.Height,
	}
	return ev
}
