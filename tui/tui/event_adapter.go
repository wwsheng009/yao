package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// DispatchEventToRuntime dispatches a Bubble Tea message through the runtime system.
// This is a simplified version that doesn't depend on the old runtime/event package.
func (m *Model) DispatchEventToRuntime(msg tea.Msg) {
	if m.RuntimeEngine == nil {
		return
	}

	// Handle window size messages
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.updateRuntimeWindowSize(sizeMsg.Width, sizeMsg.Height)
		return
	}

	// For other messages, dispatch to focused component
	if m.CurrentFocus != "" {
		m.dispatchMessageToComponent(m.CurrentFocus, msg)
	}
}

// Legacy ConvertBubbleTeaMsg is kept for compatibility but no longer uses old runtime.
// The actual event handling is now done directly in the Model.
func ConvertBubbleTeaMsg(msg tea.Msg) interface{} {
	switch m := msg.(type) {
	case tea.MouseMsg:
		return MouseData{
			X: m.X,
			Y: m.Y,
		}
	case tea.KeyMsg:
		return KeyData{
			Key:  m.String(),
			Runes: m.Runes,
		}
	case tea.WindowSizeMsg:
		return SizeData{
			Width:  m.Width,
			Height: m.Height,
		}
	default:
		return msg
	}
}

// MouseData represents mouse position data.
type MouseData struct {
	X, Y int
}

// KeyData represents keyboard input data.
type KeyData struct {
	Key   string
	Runes []rune
}

// SizeData represents window size data.
type SizeData struct {
	Width, Height int
}
