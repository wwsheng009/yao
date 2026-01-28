package focus

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime"
)

// Manager manages focus state across components
type Manager struct {
	mu                    sync.RWMutex
	focusableComponents   []string // Ordered list of focusable component IDs
	focusedIndex          int      // Index of currently focused component
	trapManager           *TrapManager // Manages focus traps (modals, etc.)
	rootNode              *runtime.LayoutNode
	geometricNavigator    *GeometricNavigator // Geometric-aware navigation helper
}

// NewManager creates a new focus manager
func NewManager(root *runtime.LayoutNode) *Manager {
	return &Manager{
		focusableComponents: make([]string, 0),
		focusedIndex:        -1,
		trapManager:         NewTrapManager(),
		rootNode:            root,
		geometricNavigator:  NewGeometricNavigator(root),
	}
}

// RefreshFocusables scans the component tree and updates the list of focusable components
func (fm *Manager) RefreshFocusables() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.focusableComponents = fm.collectFocusableIDs(fm.rootNode)
	fm.focusedIndex = -1 // Reset focus index
}

// collectFocusableIDs recursively collects all focusable component IDs
func (fm *Manager) collectFocusableIDs(node *runtime.LayoutNode) []string {
	if node == nil {
		return []string{}
	}

	var ids []string

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				ids = append(ids, node.ID)
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		childIDs := fm.collectFocusableIDs(child)
		ids = append(ids, childIDs...)
	}

	return ids
}

// FocusNext moves focus to the next focusable component
// Respects focus traps if active
func (fm *Manager) FocusNext() (string, bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Get available components (respecting traps)
	available := fm.getAvailableComponents()
	if len(available) == 0 {
		return "", false
	}

	// Find current focused index in available list
	currentIndex := -1
	if fm.focusedIndex >= 0 && fm.focusedIndex < len(fm.focusableComponents) {
		currentFocusedID := fm.focusableComponents[fm.focusedIndex]
		for i, id := range available {
			if id == currentFocusedID {
				currentIndex = i
				break
			}
		}
	}

	// Move to next component (with wraparound)
	nextIndex := (currentIndex + 1) % len(available)
	componentID := available[nextIndex]

	// Update focused index in the main list
	for i, id := range fm.focusableComponents {
		if id == componentID {
			fm.focusedIndex = i
			break
		}
	}

	// Apply focus to the component
	fm.setFocus(componentID)

	return componentID, true
}

// FocusPrev moves focus to the previous focusable component
// Respects focus traps if active
func (fm *Manager) FocusPrev() (string, bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Get available components (respecting traps)
	available := fm.getAvailableComponents()
	if len(available) == 0 {
		return "", false
	}

	// Find current focused index in available list
	currentIndex := -1
	if fm.focusedIndex >= 0 && fm.focusedIndex < len(fm.focusableComponents) {
		currentFocusedID := fm.focusableComponents[fm.focusedIndex]
		for i, id := range available {
			if id == currentFocusedID {
				currentIndex = i
				break
			}
		}
	}

	// Move to previous component (with wraparound)
	if currentIndex <= 0 {
		currentIndex = len(available)
	}
	prevIndex := currentIndex - 1
	componentID := available[prevIndex]

	// Update focused index in the main list
	for i, id := range fm.focusableComponents {
		if id == componentID {
			fm.focusedIndex = i
			break
		}
	}

	// Apply focus to the component
	fm.setFocus(componentID)

	return componentID, true
}

// getAvailableComponents returns the list of components that can receive focus
// If a focus trap is active, only returns components within that trap
func (fm *Manager) getAvailableComponents() []string {
	if fm.trapManager.HasActiveTrap() {
		return fm.trapManager.GetTrappableComponents()
	}

	return fm.focusableComponents
}

// FocusSpecific sets focus to a specific component by ID
func (fm *Manager) FocusSpecific(componentID string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Find the component in the focusable list
	for i, id := range fm.focusableComponents {
		if id == componentID {
			// Clear current focus
			fm.clearCurrentFocus()

			// Set new focus
			fm.focusedIndex = i
			fm.setFocus(componentID)
			return true
		}
	}

	return false
}

// FocusFirst focuses the first focusable component
// Respects focus traps if active
func (fm *Manager) FocusFirst() (string, bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Get available components (respecting traps)
	available := fm.getAvailableComponents()
	if len(available) == 0 {
		return "", false
	}

	componentID := available[0]

	// Update focused index in the main list
	for i, id := range fm.focusableComponents {
		if id == componentID {
			fm.focusedIndex = i
			break
		}
	}

	// Apply focus to the component
	fm.setFocus(componentID)

	return componentID, true
}

// GetFocused returns the ID of the currently focused component, if any
func (fm *Manager) GetFocused() (string, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.focusedIndex < 0 || fm.focusedIndex >= len(fm.focusableComponents) {
		return "", false
	}

	return fm.focusableComponents[fm.focusedIndex], true
}

// HasFocus checks if a specific component has focus
func (fm *Manager) HasFocus(componentID string) bool {
	focusedID, hasFocus := fm.GetFocused()
	return hasFocus && focusedID == componentID
}

// PushFocusTrap adds a focus trap to the stack (e.g., for modals/dialogs)
// Components inside a focus trap can only cycle within that trap
func (fm *Manager) PushFocusTrap(trap *FocusTrap) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.trapManager.PushTrap(trap)
}

// PopFocusTrap removes and returns the most recent focus trap
func (fm *Manager) PopFocusTrap() *FocusTrap {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	return fm.trapManager.PopTrap()
}

// RemoveFocusTrap removes a specific trap by ID
func (fm *Manager) RemoveFocusTrap(trapID string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	return fm.trapManager.RemoveTrap(trapID)
}

// GetCurrentFocusTrap returns the current active focus trap, if any
func (fm *Manager) GetCurrentFocusTrap() *FocusTrap {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.trapManager.GetActiveTrap()
}

// IsFocusTrapActive checks if a specific focus trap is currently active
func (fm *Manager) IsFocusTrapActive(trapID string) bool {
	return fm.trapManager.IsTrapActive(trapID)
}

// HasActiveFocusTrap returns whether there's currently an active focus trap
func (fm *Manager) HasActiveFocusTrap() bool {
	return fm.trapManager.HasActiveTrap()
}

// ClearFocusTraps removes all focus traps
func (fm *Manager) ClearFocusTraps() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.trapManager.Clear()
}

// setFocus applies focus to a component by ID
func (fm *Manager) setFocus(componentID string) {
	node := fm.findNodeByID(fm.rootNode, componentID)
	if node != nil && node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			focusable.SetFocus(true)
		}
	}
}

// clearCurrentFocus removes focus from the currently focused component
func (fm *Manager) clearCurrentFocus() {
	if fm.focusedIndex >= 0 && fm.focusedIndex < len(fm.focusableComponents) {
		componentID := fm.focusableComponents[fm.focusedIndex]
		node := fm.findNodeByID(fm.rootNode, componentID)
		if node != nil && node.Component != nil && node.Component.Instance != nil {
			if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
				focusable.SetFocus(false)
			}
		}
	}
}

// findNodeByID recursively finds a node by ID in the tree
func (fm *Manager) findNodeByID(node *runtime.LayoutNode, id string) *runtime.LayoutNode {
	if node == nil {
		return nil
	}

	if node.ID == id {
		return node
	}

	for _, child := range node.Children {
		if found := fm.findNodeByID(child, id); found != nil {
			return found
		}
	}

	return nil
}

// GetFocusableComponents returns the list of all focusable component IDs
func (fm *Manager) GetFocusableComponents() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	result := make([]string, len(fm.focusableComponents))
	copy(result, fm.focusableComponents)
	return result
}

// GetFocusableCount returns the number of focusable components
func (fm *Manager) GetFocusableCount() int {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return len(fm.focusableComponents)
}

// Clear removes all focus state
func (fm *Manager) Clear() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.clearCurrentFocus()
	fm.focusableComponents = make([]string, 0)
	fm.focusedIndex = -1
	fm.trapManager.Clear()
}

// ========== Geometric Navigation ==========

// FocusDirection moves focus in the specified direction using geometric navigation
// Returns the component ID and whether the move was successful
func (fm *Manager) FocusDirection(direction NavigationDirection) (string, bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Get available components (respecting traps)
	available := fm.getAvailableComponents()
	if len(available) == 0 {
		return "", false
	}

	// Get current focused component ID
	currentID := ""
	if fm.focusedIndex >= 0 && fm.focusedIndex < len(fm.focusableComponents) {
		currentID = fm.focusableComponents[fm.focusedIndex]
	}

	// Use geometric navigator to find next component in direction
	nextID := fm.geometricNavigator.FindNearestInDirection(currentID, direction, available)
	if nextID == "" {
		return "", false
	}

	// Update focused index in the main list
	for i, id := range fm.focusableComponents {
		if id == nextID {
			fm.focusedIndex = i
			break
		}
	}

	// Apply focus to the component
	fm.setFocus(nextID)

	return nextID, true
}

// FocusUp moves focus to the component above the current one
func (fm *Manager) FocusUp() (string, bool) {
	return fm.FocusDirection(DirectionUp)
}

// FocusDown moves focus to the component below the current one
func (fm *Manager) FocusDown() (string, bool) {
	return fm.FocusDirection(DirectionDown)
}

// FocusLeft moves focus to the component to the left of the current one
func (fm *Manager) FocusLeft() (string, bool) {
	return fm.FocusDirection(DirectionLeft)
}

// FocusRight moves focus to the component to the right of the current one
func (fm *Manager) FocusRight() (string, bool) {
	return fm.FocusDirection(DirectionRight)
}
