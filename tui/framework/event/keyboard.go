package event

// ==============================================================================
// Special Key Definitions
// ==============================================================================
// KeyEvent and Key are defined in event.go to avoid duplication.
// This file contains only the SpecialKey enum for non-printable keys.

// SpecialKey represents special keyboard keys.
type SpecialKey int

const (
	KeyUnknown SpecialKey = iota

	// Control keys
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert

	// Cursor keys
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	// Function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// Combo keys
	KeySpace

	// Vim-style navigation keys
	KeyK // vim up
	KeyJ // vim down
	KeyH // vim left
	KeyL // vim right
)

// String returns the string representation of a special key.
func (k SpecialKey) String() string {
	names := map[SpecialKey]string{
		KeyEscape:    "escape",
		KeyEnter:     "enter",
		KeyTab:       "tab",
		KeyBackspace: "backspace",
		KeyDelete:    "delete",
		KeyInsert:    "insert",
		KeyUp:        "up",
		KeyDown:      "down",
		KeyLeft:      "left",
		KeyRight:     "right",
		KeyHome:      "home",
		KeyEnd:       "end",
		KeyPageUp:    "pageup",
		KeyPageDown:  "pagedown",
		KeyF1:        "f1",
		KeyF2:        "f2",
		KeyF3:        "f3",
		KeyF4:        "f4",
		KeyF5:        "f5",
		KeyF6:        "f6",
		KeyF7:        "f7",
		KeyF8:        "f8",
		KeyF9:        "f9",
		KeyF10:       "f10",
		KeyF11:       "f11",
		KeyF12:       "f12",
		KeySpace:     "space",
		KeyK:         "k",
		KeyJ:         "j",
		KeyH:         "h",
		KeyL:         "l",
	}
	if name, ok := names[k]; ok {
		return name
	}
	return ""
}

// ==============================================================================
// Key Creation Helpers
// ============================================================================

// NewKeyEventFromRune creates a keyboard event from a rune.
func NewKeyEventFromRune(r rune) *KeyEvent {
	return &KeyEvent{
		BaseEvent: NewBaseEvent(EventKeyPress),
		Key: Key{
			Rune: r,
		},
	}
}

// NewSpecialKeyEvent creates a keyboard event from a special key.
func NewSpecialKeyEvent(special SpecialKey, modifiers ...Modifier) *KeyEvent {
	ev := &KeyEvent{
		BaseEvent: NewBaseEvent(EventKeyPress),
		Special:   special,
		Key: Key{
			Name: special.String(),
		},
	}
	for _, m := range modifiers {
		if m == ModAlt {
			ev.Key.Alt = true
			ev.Modifiers |= ModAlt
		}
		if m == ModCtrl {
			ev.Key.Ctrl = true
			ev.Modifiers |= ModCtrl
		}
	}
	return ev
}

// NewMouseEventFromButton creates a mouse event from button type.
func NewMouseEventFromButton(eventType EventType, x, y int, button MouseButton) *MouseEvent {
	return &MouseEvent{
		BaseEvent: NewBaseEvent(eventType),
		X:         x,
		Y:         y,
		Button:    button,
	}
}
