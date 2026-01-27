// Package adapter provides conversion between Bubble Tea messages and simplified events.
// This package is independent and does not depend on the runtime package.
package adapter

import (
	tea "github.com/charmbracelet/bubbletea"
)

// EventData represents a simple event that can be passed around.
type EventData struct {
	Type string
	Data interface{}
}

// EventTypes
const (
	EventTypeMouse   = "mouse"
	EventTypeKey     = "key"
	EventTypeResize  = "resize"
	EventTypeCustom  = "custom"
)

// ConvertBubbleTeaMsg converts a Bubble Tea message to a simple EventData.
// This is a lightweight adapter that doesn't depend on runtime/event types.
func ConvertBubbleTeaMsg(msg tea.Msg) *EventData {
	switch m := msg.(type) {
	case tea.MouseMsg:
		return &EventData{
			Type: EventTypeMouse,
			Data: MouseData{
				X:      m.X,
				Y:      m.Y,
				Action: int(m.Type),
				Button: int(m.Button),
			},
		}
	case tea.KeyMsg:
		runes := []rune{}
		if len(m.Runes) > 0 {
			runes = m.Runes
		}
		return &EventData{
			Type: EventTypeKey,
			Data: KeyData{
				Key:   m.String(),
				Type:  m.Type,
				Runes: runes,
			},
		}
	case tea.WindowSizeMsg:
		return &EventData{
			Type: EventTypeResize,
			Data: SizeData{
				Width:  m.Width,
				Height: m.Height,
			},
		}
	default:
		return &EventData{
			Type: EventTypeCustom,
			Data: msg,
		}
	}
}

// MouseData represents mouse event data.
type MouseData struct {
	X, Y   int
	Action int
	Button int
}

// KeyData represents keyboard event data.
type KeyData struct {
	Key   string
	Type  tea.KeyType
	Runes []rune
}

// SizeData represents window size data.
type SizeData struct {
	Width, Height int
}

// EventAdapter provides a bridge between Bubble Tea and simple events.
type EventAdapter struct{}

// NewEventAdapter creates a new event adapter.
func NewEventAdapter() *EventAdapter {
	return &EventAdapter{}
}

// Convert converts a Bubble Tea message to an EventData.
func (a *EventAdapter) Convert(msg tea.Msg) *EventData {
	return ConvertBubbleTeaMsg(msg)
}
