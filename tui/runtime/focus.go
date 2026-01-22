package runtime

// FocusManager manages keyboard focus navigation among focusable components.
type FocusManager struct {
	focusable      []*FocusableItem
	currentIndex   int
	onFocusChange  func(focused, previous *FocusableItem)
	enableWrap     bool // When true, navigation wraps around
	cyclic         bool // When true, Tab cycles back to first after last
}

// FocusableItem represents a component that can be focused.
type FocusableItem struct {
	ID       string
	Node     *LayoutNode
	Instance FocusableComponent
}

// NewFocusManager creates a new FocusManager.
func NewFocusManager() *FocusManager {
	return &FocusManager{
		focusable:     []*FocusableItem{},
		currentIndex:  -1, // -1 means no focus
		enableWrap:    true,
		cyclic:        false,
	}
}

// SetFocusChangeCallback sets a callback that is called when focus changes.
func (m *FocusManager) SetFocusChangeCallback(fn func(focused, previous *FocusableItem)) {
	m.onFocusChange = fn
}

// SetWrap enables or disables focus wrapping.
func (m *FocusManager) SetWrap(enabled bool) {
	m.enableWrap = enabled
}

// SetCyclic enables or disables cyclic focus navigation.
func (m *FocusManager) SetCyclic(cyclic bool) {
	m.cyclic = cyclic
}

// SetFocusable sets the list of focusable components.
func (m *FocusManager) SetFocusable(items []*FocusableItem) {
	m.focusable = items
	// Reset current index if out of bounds
	if m.currentIndex >= len(m.focusable) {
		m.currentIndex = -1
	}
}

// AddFocusable adds a focusable component to the manager.
func (m *FocusManager) AddFocusable(item *FocusableItem) {
	m.focusable = append(m.focusable, item)
}

// RemoveFocusable removes a focusable component by ID.
func (m *FocusManager) RemoveFocusable(id string) {
	for i, item := range m.focusable {
		if item.ID == id {
			// Remove the item
			m.focusable = append(m.focusable[:i], m.focusable[i+1:]...)
			// Adjust current index if needed
			if m.currentIndex > i {
				m.currentIndex--
			} else if m.currentIndex == i {
				m.currentIndex = -1
			}
			return
		}
	}
}

// FocusNext moves focus to the next focusable component.
// Returns the newly focused item, or nil if no focus change occurred.
func (m *FocusManager) FocusNext() *FocusableItem {
	if len(m.focusable) == 0 {
		return nil
	}

	previous := m.getCurrent()

	// Move to next
	m.currentIndex++
	if m.currentIndex >= len(m.focusable) {
		if m.enableWrap {
			m.currentIndex = 0
		} else {
			m.currentIndex = len(m.focusable) - 1
			return nil // No change
		}
	}

	focused := m.getCurrent()
	m.applyFocusChange(focused, previous)

	return focused
}

// FocusPrev moves focus to the previous focusable component.
// Returns the newly focused item, or nil if no focus change occurred.
func (m *FocusManager) FocusPrev() *FocusableItem {
	if len(m.focusable) == 0 {
		return nil
	}

	previous := m.getCurrent()

	// Move to previous
	m.currentIndex--
	if m.currentIndex < 0 {
		if m.enableWrap {
			m.currentIndex = len(m.focusable) - 1
		} else {
			m.currentIndex = 0
			return nil // No change
		}
	}

	focused := m.getCurrent()
	m.applyFocusChange(focused, previous)

	return focused
}

// Focus moves focus to a specific component by ID.
// Returns true if the component was found and focused.
func (m *FocusManager) Focus(id string) bool {
	for i, item := range m.focusable {
		if item.ID == id {
			previous := m.getCurrent()
			m.currentIndex = i
			focused := m.getCurrent()
			m.applyFocusChange(focused, previous)
			return true
		}
	}
	return false
}

// FocusAt moves focus to the component at the specified index.
// Returns true if the index was valid.
func (m *FocusManager) FocusAt(index int) bool {
	if index < 0 || index >= len(m.focusable) {
		return false
	}

	previous := m.getCurrent()
	m.currentIndex = index
	focused := m.getCurrent()
	m.applyFocusChange(focused, previous)

	return true
}

// GetCurrent returns the currently focused item, or nil if none.
func (m *FocusManager) GetCurrent() *FocusableItem {
	return m.getCurrent()
}

// GetCurrentIndex returns the current focus index.
// Returns -1 if no component is focused.
func (m *FocusManager) GetCurrentIndex() int {
	return m.currentIndex
}

// GetFocusable returns all focusable components.
func (m *FocusManager) GetFocusable() []*FocusableItem {
	return m.focusable
}

// ClearFocus removes focus from the current component.
func (m *FocusManager) ClearFocus() {
	previous := m.getCurrent()
	m.currentIndex = -1
	m.applyFocusChange(nil, previous)
}

// HasFocus checks if a specific component ID currently has focus.
func (m *FocusManager) HasFocus(id string) bool {
	current := m.getCurrent()
	return current != nil && current.ID == id
}

// Count returns the number of focusable components.
func (m *FocusManager) Count() int {
	return len(m.focusable)
}

// IsEmpty returns true if there are no focusable components.
func (m *FocusManager) IsEmpty() bool {
	return len(m.focusable) == 0
}

// getCurrent returns the currently focused item, or nil.
func (m *FocusManager) getCurrent() *FocusableItem {
	if m.currentIndex < 0 || m.currentIndex >= len(m.focusable) {
		return nil
	}
	return m.focusable[m.currentIndex]
}

// applyFocusChange applies the focus change to the components.
func (m *FocusManager) applyFocusChange(focused, previous *FocusableItem) {
	// Remove focus from previous
	if previous != nil && previous.Instance != nil {
		previous.Instance.SetFocus(false)
	}

	// Add focus to new
	if focused != nil && focused.Instance != nil {
		focused.Instance.SetFocus(true)
	}

	// Notify callback
	if m.onFocusChange != nil {
		m.onFocusChange(focused, previous)
	}
}

// UpdateFromLayout updates the focusable list from a layout result.
func (m *FocusManager) UpdateFromLayout(result LayoutResult) {
	items := []*FocusableItem{}

	for _, box := range result.Boxes {
		if box.Node != nil && box.Node.Component != nil && box.Node.Component.Instance != nil {
			if focusable, ok := box.Node.Component.Instance.(FocusableComponent); ok {
				if focusable.IsFocusable() {
					items = append(items, &FocusableItem{
						ID:       box.NodeID,
						Node:     box.Node,
						Instance: focusable,
					})
				}
			}
		}
	}

	// Preserve current focus if possible
	currentID := ""
	if current := m.getCurrent(); current != nil {
		currentID = current.ID
	}

	m.focusable = items
	m.currentIndex = -1

	// Try to restore focus
	if currentID != "" {
		for i, item := range m.focusable {
			if item.ID == currentID {
				m.currentIndex = i
				break
			}
		}
	}
}

// FocusFirst focuses the first focusable component.
// Returns true if a component was focused.
func (m *FocusManager) FocusFirst() bool {
	if len(m.focusable) == 0 {
		return false
	}

	previous := m.getCurrent()
	m.currentIndex = 0
	focused := m.getCurrent()
	m.applyFocusChange(focused, previous)

	return true
}

// FocusLast focuses the last focusable component.
// Returns true if a component was focused.
func (m *FocusManager) FocusLast() bool {
	if len(m.focusable) == 0 {
		return false
	}

	previous := m.getCurrent()
	m.currentIndex = len(m.focusable) - 1
	focused := m.getCurrent()
	m.applyFocusChange(focused, previous)

	return true
}

// FocusNone removes focus from all components.
func (m *FocusManager) FocusNone() {
	m.ClearFocus()
}

// FindByID finds a focusable item by its node ID.
// Returns the item, or nil if not found.
func (m *FocusManager) FindByID(id string) *FocusableItem {
	for _, item := range m.focusable {
		if item.ID == id {
			return item
		}
	}
	return nil
}

// NextDirection moves focus in the specified direction.
// Direction: 1 = next, -1 = previous
func (m *FocusManager) NextDirection(direction int) *FocusableItem {
	if direction > 0 {
		return m.FocusNext()
	} else if direction < 0 {
		return m.FocusPrev()
	}
	return m.getCurrent()
}

// CollectFocusableFromNode collects all focusable components from a node tree.
// This is a helper function to populate the focus manager from a layout.
func CollectFocusableFromNode(root *LayoutNode) []*FocusableItem {
	return collectFocusableRecursive(root)
}

// collectFocusableRecursive recursively collects focusable components.
func collectFocusableRecursive(node *LayoutNode) []*FocusableItem {
	var items []*FocusableItem

	if node == nil {
		return items
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(FocusableComponent); ok {
			if focusable.IsFocusable() {
				items = append(items, &FocusableItem{
					ID:       node.ID,
					Node:     node,
					Instance: focusable,
				})
			}
		}
	}

	// Recursively collect from children
	for _, child := range node.Children {
		childItems := collectFocusableRecursive(child)
		items = append(items, childItems...)
	}

	return items
}
