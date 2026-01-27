package event

import (
	"time"

	"github.com/yaoapp/yao/tui/framework/platform"
	"github.com/yaoapp/yao/tui/runtime/event"
)

// Pump reads raw input from the platform and converts to events.
type Pump struct {
	input  platform.InputReader
	events chan Event
	quit   chan struct{}
	running bool
}

// NewPump creates a new event pump.
func NewPump(reader platform.InputReader) *Pump {
	return &Pump{
		input:  reader,
		events: make(chan Event, 100),
		quit:   make(chan struct{}),
		running: false,
	}
}

// Start starts the event pump.
func (p *Pump) Start() error {
	if p.running {
		return nil
	}

	// Create raw input channel
	rawInputs := make(chan platform.RawInput, 50)

	// Start platform input reader
	if err := p.input.Start(rawInputs); err != nil {
		return err
	}

	p.running = true

	// Start conversion loop
	go p.convertLoop(rawInputs)

	return nil
}

// convertLoop converts raw inputs to events.
func (p *Pump) convertLoop(rawInputs <-chan platform.RawInput) {
	for {
		select {
		case <-p.quit:
			return

		case raw, ok := <-rawInputs:
			if !ok {
				return
			}
			ev := p.convertToEvent(raw)
			if ev != nil {
				select {
				case p.events <- ev:
				case <-p.quit:
					return
				}
			}
		}
	}
}

// convertToEvent converts raw input to framework event.
func (p *Pump) convertToEvent(raw platform.RawInput) Event {
	switch raw.Type {
	case platform.InputKeyPress:
		return p.convertKeyEvent(raw)

	case platform.InputResize:
		return p.convertResizeEvent(raw)

	case platform.InputMouse:
		return p.convertMouseEvent(raw)

	default:
		return nil
	}
}

// convertKeyEvent converts keyboard raw input to KeyEvent.
func (p *Pump) convertKeyEvent(raw platform.RawInput) Event {
	baseEv := NewBaseEvent(event.EventKeyPress)

	// Create key event
	ev := &KeyEvent{
		BaseEvent: baseEv,
	}

	// Set special key
	if raw.Special != platform.KeyUnknown {
		ev.Special = SpecialKey(raw.Special)
		ev.Key.Name = ev.Special.String()
	} else {
		// Character key
		ev.Key.Rune = raw.Key
	}

	// Set modifiers
	if raw.Modifiers&platform.ModAlt != 0 {
		ev.Key.Alt = true
		ev.Modifiers |= ModAlt
	}
	if raw.Modifiers&platform.ModCtrl != 0 {
		ev.Key.Ctrl = true
		ev.Modifiers |= ModCtrl
	}
	if raw.Modifiers&platform.ModShift != 0 {
		ev.Modifiers |= ModShift
	}

	return ev
}

// convertResizeEvent converts resize raw input to ResizeEvent.
func (p *Pump) convertResizeEvent(raw platform.RawInput) Event {
	return &ResizeEvent{
		BaseEvent: NewBaseEvent(event.EventResize),
		OldWidth:  0,
		OldHeight: 0,
		NewWidth:  raw.Width,
		NewHeight: raw.Height,
	}
}

// convertMouseEvent converts mouse raw input to MouseEvent.
func (p *Pump) convertMouseEvent(raw platform.RawInput) Event {
	var eventType event.EventType

	switch raw.MouseAction {
	case platform.MousePress:
		eventType = event.EventMousePress
	case platform.MouseRelease:
		eventType = event.EventMouseRelease
	case platform.MouseMotion:
		eventType = event.EventMouseMove
	case platform.MouseWheelUp:
		eventType = event.EventMouseWheel
	case platform.MouseWheelDown:
		eventType = event.EventMouseWheel
	default:
		eventType = event.EventMousePress
	}

	return &MouseEvent{
		BaseEvent: NewBaseEvent(eventType),
		X:         raw.MouseX,
		Y:         raw.MouseY,
		Button:    MouseButton(raw.MouseButton),
	}
}

// Stop stops the event pump.
func (p *Pump) Stop() {
	if !p.running {
		return
	}

	p.running = false

	// Send quit signal
	close(p.quit)

	// Stop input reader
	if p.input != nil {
		p.input.Stop()
	}

	// Close events channel
	close(p.events)
}

// Events returns the event channel.
func (p *Pump) Events() <-chan Event {
	return p.events
}

// IsRunning checks if the pump is running.
func (p *Pump) IsRunning() bool {
	return p.running
}

// PumpWithTimeout gets an event with timeout.
func (p *Pump) PumpWithTimeout(timeout time.Duration) (Event, bool) {
	select {
	case ev := <-p.events:
		return ev, true
	case <-time.After(timeout):
		return nil, false
	}
}
