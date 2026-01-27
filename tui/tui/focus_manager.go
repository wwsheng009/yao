package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/tui/core"
)

// focusNextInput finds the next input component and sets it as focused
// Returns tea.Cmd to send focus messages via message-driven approach
func (m *Model) focusNextInput() tea.Cmd {
	// Find all input component IDs from Components map
	inputIDs := []string{}
	for id, comp := range m.Components {
		if comp.Type == string(InputComponent) {
			inputIDs = append(inputIDs, id)
		}
	}

	if len(inputIDs) == 0 {
		return nil
	}

	// Find current position
	currentIndex := -1
	for i, id := range inputIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Determine next focus ID
	var nextFocus string
	if currentIndex >= 0 && currentIndex < len(inputIDs)-1 {
		nextFocus = inputIDs[currentIndex+1]
	} else {
		nextFocus = inputIDs[0] // Wrap to first
	}

	// Use setFocus which handles focus change via TargetedMsg
	return m.setFocus(nextFocus)
}

// focusPrevInput focuses to previous input component
// Returns tea.Cmd to send focus messages via message-driven approach
func (m *Model) focusPrevInput() tea.Cmd {
	// Find all input component IDs from Components map
	inputIDs := []string{}
	for id, comp := range m.Components {
		if comp.Type == string(InputComponent) {
			inputIDs = append(inputIDs, id)
		}
	}

	if len(inputIDs) == 0 {
		return nil
	}

	// Find current position
	currentIndex := -1
	for i, id := range inputIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Determine previous focus ID
	var prevFocus string
	if currentIndex > 0 {
		prevFocus = inputIDs[currentIndex-1]
	} else if currentIndex == 0 {
		prevFocus = inputIDs[len(inputIDs)-1] // Wrap to last
	} else {
		// No current focus, start from last
		prevFocus = inputIDs[len(inputIDs)-1]
	}

	// Use setFocus which handles focus change via TargetedMsg
	return m.setFocus(prevFocus)
}

// setFocus sets focus to a specific component
// This sends TargetedMsg to inform components about focus changes via tea message mechanism.
// Old component receives FocusMsg with Type=FocusLost
// New component receives FocusMsg with Type=FocusGranted
func (m *Model) setFocus(componentID string) tea.Cmd {
	if componentID == m.CurrentFocus {
		return nil // Already focused
	}

	oldFocus := m.CurrentFocus
	m.CurrentFocus = componentID

	// Create commands to send focus messages
	var cmds []tea.Cmd = make([]tea.Cmd, 0, 2)

	// Send focus lost message to old component (if any)
	if oldFocus != "" {
		cmds = append(cmds, func() tea.Msg {
			return core.TargetedMsg{
				TargetID: oldFocus,
				InnerMsg: core.FocusMsg{
					Type:   core.FocusLost,
					Reason: "TAB_NAVIGATE",
					ToID:   componentID,
				},
			}
		})
	}

	// Send focus granted message to new component (if any)
	if componentID != "" {
		cmds = append(cmds, func() tea.Msg {
			return core.TargetedMsg{
				TargetID: componentID,
				InnerMsg: core.FocusMsg{
					Type:   core.FocusGranted,
					Reason: "TAB_NAVIGATE",
					FromID: oldFocus,
				},
			}
		})
	}

	log.Trace("TUI Focus: Focus changed from %s to %s via tea Cmd", oldFocus, componentID)

	return tea.Batch(cmds...)
}

// clearFocus clears focus from current component
// This returns a tea.Cmd that sends FocusMsg with FocusLost to the focused component.
func (m *Model) clearFocus() tea.Cmd {
	if m.CurrentFocus == "" {
		return nil
	}

	oldFocus := m.CurrentFocus

	// Send focus lost message to the component via tea.Cmd
	cmd := func() tea.Msg {
		return core.TargetedMsg{
			TargetID: oldFocus,
			InnerMsg: core.FocusMsg{
				Type:   core.FocusLost,
				Reason: "USER_ESC",
				ToID:   "",
			},
		}
	}

	m.CurrentFocus = ""

	log.Trace("TUI Focus: Focus cleared from %s via tea Cmd", oldFocus)

	return cmd
}

// focusNextComponent moves focus to the next focusable component
func (m *Model) focusNextComponent() tea.Cmd {
	// Get all focusable component IDs
	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		return nil
	}

	// Find current position
	currentIndex := -1
	for i, id := range focusableIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Move to next component, wrap around if needed
	var nextFocus string
	if currentIndex >= 0 && currentIndex < len(focusableIDs)-1 {
		nextFocus = focusableIDs[currentIndex+1]
	} else {
		nextFocus = focusableIDs[0] // Wrap to first
	}

	return m.setFocus(nextFocus)
}

// focusPrevComponent moves focus to the previous focusable component
func (m *Model) focusPrevComponent() tea.Cmd {
	// Get all focusable component IDs
	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		return nil
	}

	// Check if Tab cycling is enabled
	tabCycles := m.Config.TabCycles
	if !tabCycles {
		// Default to true for backward compatibility
		tabCycles = true
	}

	// Find current position
	currentIndex := -1
	for i, id := range focusableIDs {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// Move to previous component
	var prevFocus string
	if currentIndex > 0 {
		prevFocus = focusableIDs[currentIndex-1]
		log.Trace("Moved to previous component: %s (index %d)", prevFocus, currentIndex-1)
		return m.setFocus(prevFocus)
	} else if currentIndex == 0 {
		// At first component
		if tabCycles {
			// Wrap to last component
			prevFocus = focusableIDs[len(focusableIDs)-1]
			log.Trace("Cycled to last component: %s (index %d)", prevFocus, len(focusableIDs)-1)
			return m.setFocus(prevFocus)
		} else {
			// Cycling disabled, don't move
			log.Trace("Already at first component, Tab cycling disabled, staying at %s", m.CurrentFocus)
			return nil
		}
	} else {
		// No current focus, start from last
		prevFocus = focusableIDs[len(focusableIDs)-1]
		log.Trace("No current focus, set to last component: %s", prevFocus)
		return m.setFocus(prevFocus)
	}
}

// getFocusableComponentIDs returns IDs of all focusable components
// Uses the global component registry to determine which component types are focusable
// When UseRuntime is true, returns the geometrically-ordered focus list from Runtime
func (m *Model) getFocusableComponentIDs() []string {
	// Use Runtime focus list if enabled (geometric ordering)
	if len(m.runtimeFocusList) > 0 {
		return m.runtimeFocusList
	}

	// Legacy mode: use registry order
	registry := GetGlobalRegistry()
	ids := []string{}
	for id, comp := range m.Components {
		// Check if component type is registered as focusable
		if registry.IsFocusable(ComponentType(comp.Type)) {
			ids = append(ids, id)
		}
	}
	return ids
}

// isMenuFocused checks if the current focus is on a menu component
func (m *Model) isMenuFocused() bool {
	if m.CurrentFocus == "" {
		return false
	}
	if comp, exists := m.Components[m.CurrentFocus]; exists {
		return comp.Type == string(MenuComponent)
	}
	return false
}
