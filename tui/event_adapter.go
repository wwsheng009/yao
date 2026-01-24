package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/event"
)

// ConvertBubbleTeaMsg converts a Bubble Tea message to a Runtime event.
// This adapter maintains module boundary rules by sitting in the tui package
// (which can import both Bubble Tea and runtime).
func ConvertBubbleTeaMsg(msg tea.Msg) event.Event {
	switch m := msg.(type) {
	case tea.MouseMsg:
		return convertMouseMsg(m)
	case tea.KeyMsg:
		return convertKeyMsg(m)
	case tea.WindowSizeMsg:
		return convertWindowSizeMsg(m)
	default:
		// For custom messages, wrap as-is
		return event.Event{
			Type:   event.EventTypeCustom,
			Custom: msg,
		}
	}
}

// convertMouseMsg converts a Bubble Tea MouseMsg to a Runtime MouseEvent.
func convertMouseMsg(msg tea.MouseMsg) event.Event {
	// Determine mouse event type
	var evType event.MouseEventType
	var click event.MouseClickType

	// Bubble Tea MouseMsg contains: X, Y, Action, and Button
	switch msg.Action {
	case tea.MouseActionPress:
		evType = event.MousePress
		// Map button type
		switch msg.Button {
		case tea.MouseButtonLeft:
			click = event.MouseLeft
		case tea.MouseButtonMiddle:
			click = event.MouseMiddle
		case tea.MouseButtonRight:
			click = event.MouseRight
		default:
			click = event.MouseLeft
		}
	case tea.MouseActionRelease:
		evType = event.MouseRelease
		switch msg.Button {
		case tea.MouseButtonLeft:
			click = event.MouseLeft
		case tea.MouseButtonMiddle:
			click = event.MouseMiddle
		case tea.MouseButtonRight:
			click = event.MouseRight
		default:
			click = event.MouseLeft
		}
	case tea.MouseActionMotion:
		evType = event.MouseMove
		click = event.MouseLeft
	default:
		// Check for scroll wheel
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
			evType = event.MouseScroll
		} else {
			evType = event.MousePress
			click = event.MouseLeft
		}
	}

	return event.Event{
		Type: event.EventTypeMouse,
		Mouse: &event.MouseEvent{
			X:     msg.X,
			Y:     msg.Y,
			Type:  evType,
			Click: click,
		},
	}
}

// convertKeyMsg converts a Bubble Tea KeyMsg to a Runtime KeyEvent.
func convertKeyMsg(msg tea.KeyMsg) event.Event {
	// Extract rune from key
	var key rune
	if len(msg.Runes) > 0 {
		key = msg.Runes[0]
	}

	// Determine key modifier
	var mod event.KeyModifier
	if msg.Alt {
		mod = event.ModAlt
	} else if msg.Type == tea.KeyShiftLeft || msg.Type == tea.KeyShiftRight {
		mod = event.ModShift
	} else if msg.String() == "ctrl+c" || msg.String() == "ctrl+v" ||
		msg.String() == "ctrl+up" || msg.String() == "ctrl+down" ||
		msg.String() == "ctrl+left" || msg.String() == "ctrl+right" {
		// Check string representation for ctrl combinations
		mod = event.ModCtrl
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

// convertWindowSizeMsg converts a Bubble Tea WindowSizeMsg to a Runtime ResizeEvent.
func convertWindowSizeMsg(msg tea.WindowSizeMsg) event.Event {
	return event.Event{
		Type:   event.EventTypeResize,
		Resize: &event.ResizeEvent{
			Width:  msg.Width,
			Height: msg.Height,
		},
	}
}

// DispatchEventToRuntime dispatches a Bubble Tea message through the Runtime event system.
// It converts the message to a Runtime event, performs hit testing, and dispatches
// to the appropriate component.
func (m *Model) DispatchEventToRuntime(msg tea.Msg) event.EventResult {
	// Convert to Runtime event
	ev := ConvertBubbleTeaMsg(msg)

	// Get the layout boxes from Runtime engine
	// The boxes are the output of the layout phase and contain position information
	var boxes []runtime.LayoutBox
	if m.RuntimeEngine != nil {
		boxes = m.RuntimeEngine.GetBoxes()
	}

	// Dispatch the event
	result := event.DispatchEvent(ev, boxes)

	// Handle focus changes if needed
	if result.FocusChange != event.FocusChangeNone {
		m.handleFocusChange(result.FocusChange)
	}

	// Handle focus target (e.g., mouse click on a focusable component)
	if result.FocusTarget != "" {
		m.handleFocusTarget(result.FocusTarget)
	}

	return result
}

// handleFocusChange handles focus change requests from event dispatch.
func (m *Model) handleFocusChange(change event.FocusChangeType) {
	if m.RuntimeEngine == nil {
		return
	}

	focusMgr := m.RuntimeEngine.GetFocusManager()
	if focusMgr == nil {
		return
	}

	switch change {
	case event.FocusChangeNext:
		focused := focusMgr.FocusNext()
		if focused != nil {
			log.Trace("Focus moved to next: %s", focused.ID)
			m.forceRender = true // 焦点变化需要重新渲染
		}
	case event.FocusChangePrev:
		focused := focusMgr.FocusPrev()
		if focused != nil {
			log.Trace("Focus moved to previous: %s", focused.ID)
			m.forceRender = true // 焦点变化需要重新渲染
		}
	case event.FocusChangeSpecific:
		// Focus would be set to a specific component
		// This requires additional context about which component to focus
		log.Trace("Focus change to specific component requested")
		m.forceRender = true // 焦点变化需要重新渲染
	}
}

// handleFocusTarget handles focusing a specific component by ID.
func (m *Model) handleFocusTarget(targetID string) {
	if m.RuntimeEngine == nil {
		return
	}

	focusMgr := m.RuntimeEngine.GetFocusManager()
	if focusMgr == nil {
		return
	}

	// Try to focus the specific component
	if focusMgr.Focus(targetID) {
		log.Trace("Focus set to component: %s", targetID)
		m.forceRender = true // 焦点变化需要重新渲染
	}
}
