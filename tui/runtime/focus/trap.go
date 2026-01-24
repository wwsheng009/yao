package focus

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// TrapType represents the type of focus trap
type TrapType int

const (
	TrapModal TrapType = iota // Modal dialog
	TrapMenu                  // Menu/dropdown
	TrapPopover               // Popover/tooltip
)

// FocusTrap represents a focus trap that restricts focus navigation
// to a specific subtree of components (e.g., a modal dialog)
type FocusTrap struct {
	ID       string
	Type     TrapType
	RootNode *runtime.LayoutNode // Root of the trapped subtree
	Active   bool
}

// NewFocusTrap creates a new focus trap
func NewFocusTrap(id string, trapType TrapType, rootNode *runtime.LayoutNode) *FocusTrap {
	return &FocusTrap{
		ID:       id,
		Type:     trapType,
		RootNode: rootNode,
		Active:   false,
	}
}

// GetFocusableIDs returns all focusable component IDs within this trap
func (ft *FocusTrap) GetFocusableIDs() []string {
	if ft.RootNode == nil {
		return []string{}
	}

	return collectFocusableIDs(ft.RootNode)
}

// Contains checks if a component ID is within this trap
func (ft *FocusTrap) Contains(componentID string) bool {
	if ft.RootNode == nil {
		return false
	}

	return nodeContainsID(ft.RootNode, componentID)
}

// Activate marks this trap as active
func (ft *FocusTrap) Activate() {
	ft.Active = true
}

// Deactivate marks this trap as inactive
func (ft *FocusTrap) Deactivate() {
	ft.Active = false
}

// IsActive returns whether this trap is currently active
func (ft *FocusTrap) IsActive() bool {
	return ft.Active
}

// collectFocusableIDs recursively collects all focusable component IDs
func collectFocusableIDs(node *runtime.LayoutNode) []string {
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
		childIDs := collectFocusableIDs(child)
		ids = append(ids, childIDs...)
	}

	return ids
}

// nodeContainsID checks if a node or any of its descendants has the given ID
func nodeContainsID(node *runtime.LayoutNode, id string) bool {
	if node == nil {
		return false
	}

	if node.ID == id {
		return true
	}

	for _, child := range node.Children {
		if nodeContainsID(child, id) {
			return true
		}
	}

	return false
}

// TrapManager manages multiple focus traps
type TrapManager struct {
	trapStack []*FocusTrap // Stack of active traps (top is most recent)
}

// NewTrapManager creates a new trap manager
func NewTrapManager() *TrapManager {
	return &TrapManager{
		trapStack: make([]*FocusTrap, 0),
	}
}

// PushTrap adds a trap to the top of the stack and activates it
func (tm *TrapManager) PushTrap(trap *FocusTrap) {
	// Deactivate current top trap
	if len(tm.trapStack) > 0 {
		tm.trapStack[len(tm.trapStack)-1].Deactivate()
	}

	// Add and activate new trap
	trap.Activate()
	tm.trapStack = append(tm.trapStack, trap)
}

// PopTrap removes the top trap from the stack and returns it
func (tm *TrapManager) PopTrap() *FocusTrap {
	if len(tm.trapStack) == 0 {
		return nil
	}

	// Remove top trap
	trap := tm.trapStack[len(tm.trapStack)-1]
	trap.Deactivate()
	tm.trapStack = tm.trapStack[:len(tm.trapStack)-1]

	// Reactivate previous trap
	if len(tm.trapStack) > 0 {
		tm.trapStack[len(tm.trapStack)-1].Activate()
	}

	return trap
}

// GetActiveTrap returns the currently active trap (top of stack)
func (tm *TrapManager) GetActiveTrap() *FocusTrap {
	if len(tm.trapStack) == 0 {
		return nil
	}

	return tm.trapStack[len(tm.trapStack)-1]
}

// HasActiveTrap returns whether there's an active trap
func (tm *TrapManager) HasActiveTrap() bool {
	return len(tm.trapStack) > 0
}

// IsTrapActive checks if a specific trap is currently active
func (tm *TrapManager) IsTrapActive(trapID string) bool {
	active := tm.GetActiveTrap()
	return active != nil && active.ID == trapID
}

// RemoveTrap removes a specific trap from the stack
func (tm *TrapManager) RemoveTrap(trapID string) bool {
	for i, trap := range tm.trapStack {
		if trap.ID == trapID {
			// Remove this trap
			tm.trapStack = append(tm.trapStack[:i], tm.trapStack[i+1:]...)

			// If this was the top trap, activate the new top
			if i == len(tm.trapStack) && len(tm.trapStack) > 0 {
				tm.trapStack[len(tm.trapStack)-1].Activate()
			}

			return true
		}
	}

	return false
}

// Clear removes all traps
func (tm *TrapManager) Clear() {
	for _, trap := range tm.trapStack {
		trap.Deactivate()
	}
	tm.trapStack = make([]*FocusTrap, 0)
}

// GetTrapCount returns the number of active traps
func (tm *TrapManager) GetTrapCount() int {
	return len(tm.trapStack)
}

// ShouldTrapFocus checks if focus should be trapped (i.e., there's an active trap)
func (tm *TrapManager) ShouldTrapFocus() bool {
	return tm.HasActiveTrap()
}

// IsComponentTrapped checks if a component ID is within the active trap
func (tm *TrapManager) IsComponentTrapped(componentID string) bool {
	activeTrap := tm.GetActiveTrap()
	if activeTrap == nil {
		return false
	}

	return activeTrap.Contains(componentID)
}

// GetTrappableComponents returns focusable components within the active trap
func (tm *TrapManager) GetTrappableComponents() []string {
	activeTrap := tm.GetActiveTrap()
	if activeTrap == nil {
		return []string{}
	}

	return activeTrap.GetFocusableIDs()
}
