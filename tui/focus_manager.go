package tui

import (
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// focusNextInput finds the next input component and sets it as focused
func (m *Model) focusNextInput() {
	// Find all input component IDs from Components map
	inputIDs := []string{}
	for id, comp := range m.Components {
		if comp.Type == "input" {
			inputIDs = append(inputIDs, id)
		}
	}

	if len(inputIDs) == 0 {
		return
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

	// Use setFocus which handles focus change events
	m.setFocus(nextFocus)
}

// focusPrevInput focuses to previous input component
func (m *Model) focusPrevInput() {
	// Find all input component IDs from Components map
	inputIDs := []string{}
	for id, comp := range m.Components {
		if comp.Type == "input" {
			inputIDs = append(inputIDs, id)
		}
	}

	if len(inputIDs) == 0 {
		return
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

	// Use setFocus which handles focus change events
	m.setFocus(prevFocus)
}

// setFocus sets focus to a specific component
// This is the ONLY method that should be used to set focus
// It ensures GLOBAL FOCUS EXCLUSION: only one component can have focus at any time
func (m *Model) setFocus(componentID string) {
	if componentID == m.CurrentFocus {
		return // Already focused
	}

	// Step 1: Clear focus from ALL components to ensure global focus exclusion
	// This prevents any component from having focus if it shouldn't
	for id, comp := range m.Components {
		if id != componentID && comp.Instance.GetFocus() {
			// Only call SetFocus if the component actually has focus
			comp.Instance.SetFocus(false)
			comp.LastFocusState = false
		}
	}

	// Step 2: Set new focus to the target component
	m.CurrentFocus = componentID
	if comp, exists := m.Components[componentID]; exists {
		// Only call SetFocus if the component's actual focus state is false
		// Use GetFocus() to check the component's current state
		if !comp.Instance.GetFocus() {
			comp.Instance.SetFocus(true)
		}
		// Update LastFocusState to track this change
		comp.LastFocusState = true
	}

	// Publish focus changed event
	m.EventBus.Publish(core.ActionMsg{
		ID:     componentID,
		Action: core.EventFocusChanged,
		Data:   map[string]interface{}{"focused": true},
	})

	log.Trace("TUI Focus: Focus set to %s (global focus exclusion applied)", componentID)
}

// clearFocus clears focus from current component
// This is the ONLY method that should be used to clear focus
// It ensures GLOBAL FOCUS EXCLUSION: clears focus from ALL components
func (m *Model) clearFocus() {
	if m.CurrentFocus == "" {
		return
	}

	oldFocus := m.CurrentFocus

	// Clear focus from ALL components to ensure global focus exclusion
	// This prevents any component from retaining focus incorrectly
	for _, comp := range m.Components {
		if comp.Instance.GetFocus() {
			// Only call SetFocus if the component actually has focus
			comp.Instance.SetFocus(false)
			comp.LastFocusState = false
		}
	}

	m.CurrentFocus = ""

	// Publish focus changed event
	m.EventBus.Publish(core.ActionMsg{
		ID:     oldFocus,
		Action: core.EventFocusChanged,
		Data:   map[string]interface{}{"focused": false},
	})

	log.Trace("TUI Focus: Focus cleared from %s (global focus exclusion applied)", oldFocus)
}

// focusNextComponent moves focus to the next focusable component
func (m *Model) focusNextComponent() {
	// Get all focusable component IDs
	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		return
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

	m.setFocus(nextFocus)
}

// focusPrevComponent moves focus to the previous focusable component
func (m *Model) focusPrevComponent() {
	// Get all focusable component IDs
	focusableIDs := m.getFocusableComponentIDs()
	if len(focusableIDs) == 0 {
		return
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
		m.setFocus(prevFocus)
		log.Trace("Moved to previous component: %s (index %d)", prevFocus, currentIndex-1)
	} else if currentIndex == 0 {
		// At first component
		if tabCycles {
			// Wrap to last component
			prevFocus = focusableIDs[len(focusableIDs)-1]
			m.setFocus(prevFocus)
			log.Trace("Cycled to last component: %s (index %d)", prevFocus, len(focusableIDs)-1)
		} else {
			// Cycling disabled, don't move
			log.Trace("Already at first component, Tab cycling disabled, staying at %s", m.CurrentFocus)
		}
	} else {
		// No current focus, start from last
		prevFocus = focusableIDs[len(focusableIDs)-1]
		m.setFocus(prevFocus)
		log.Trace("No current focus, set to last component: %s", prevFocus)
	}
}

// getFocusableComponentIDs returns IDs of all focusable components
func (m *Model) getFocusableComponentIDs() []string {
	// Define which component types are focusable
	focusableTypes := map[string]bool{
		"input":    true,
		"textarea": true,
		"menu":     true,
		"form":     true,
		"table":    true,
		"crud":     true,
		"chat":     true,
	}

	ids := []string{}
	for id, comp := range m.Components {
		if focusableTypes[comp.Type] {
			ids = append(ids, id)
		}
	}
	return ids
}

// updateInputFocusStates updates the focus states of all components
// Only calls SetFocus when the focus state actually changes
// Uses GetFocus() to verify the component's actual focus state
// Ensures GLOBAL FOCUS EXCLUSION: only CurrentFocus component should have focus
func (m *Model) updateInputFocusStates() {
	for id, compInstance := range m.Components {
		shouldFocus := (id == m.CurrentFocus)
		// Check both LastFocusState and actual focus state using GetFocus()
		actualFocus := compInstance.Instance.GetFocus()
		// Only update if the tracked state differs from the target state
		// OR if the actual focus state differs from the target state
		if shouldFocus != compInstance.LastFocusState || actualFocus != shouldFocus {
			compInstance.Instance.SetFocus(shouldFocus)
			compInstance.LastFocusState = shouldFocus
		}
		// Sanity check: if component has focus but shouldn't, force clear it
		// This ensures global focus exclusion even if state got out of sync
		if !shouldFocus && compInstance.Instance.GetFocus() {
			compInstance.Instance.SetFocus(false)
			compInstance.LastFocusState = false
		}
	}
}

// validateAndCorrectFocusState validates the global focus state and corrects any inconsistencies
// This ensures that only CurrentFocus component has focus across all components
// Returns the number of components that had their focus corrected
func (m *Model) validateAndCorrectFocusState() int {
	corrections := 0

	for id, comp := range m.Components {
		shouldFocus := (id == m.CurrentFocus)
		actualFocus := comp.Instance.GetFocus()

		if shouldFocus != actualFocus {
			// Correct the focus state
			comp.Instance.SetFocus(shouldFocus)
			comp.LastFocusState = shouldFocus
			corrections++
			log.Trace("validateAndCorrectFocusState: Corrected focus for %s (shouldFocus=%v, actualFocus=%v)",
				id, shouldFocus, actualFocus)
		}
	}

	return corrections
}

// isMenuFocused checks if the current focus is on a menu component
func (m *Model) isMenuFocused() bool {
	if m.CurrentFocus == "" {
		return false
	}
	if comp, exists := m.Components[m.CurrentFocus]; exists {
		return comp.Type == "menu"
	}
	return false
}
