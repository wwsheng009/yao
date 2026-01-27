package focus

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/core"
	"github.com/yaoapp/yao/tui/runtime"
)

// MockFocusableComponent is a test component that can be focused
type MockFocusableComponent struct {
	ID      string
	Focused bool
}

func (m *MockFocusableComponent) View() string {
	return "mock"
}

func (m *MockFocusableComponent) Init() tea.Cmd {
	return nil
}

func (m *MockFocusableComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return m, nil, core.Handled
}

func (m *MockFocusableComponent) GetID() string {
	return m.ID
}

func (m *MockFocusableComponent) SetFocus(focused bool) {
	m.Focused = focused
}

func (m *MockFocusableComponent) GetFocus() bool {
	return m.Focused
}

func (m *MockFocusableComponent) GetComponentType() string {
	return "mock-focusable"
}

func (m *MockFocusableComponent) Render(config core.RenderConfig) (string, error) {
	return "mock", nil
}

func (m *MockFocusableComponent) UpdateRenderConfig(config core.RenderConfig) error {
	return nil
}

func (m *MockFocusableComponent) Cleanup() {}

func (m *MockFocusableComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (m *MockFocusableComponent) GetSubscribedMessageTypes() []string {
	return nil
}

func (m *MockFocusableComponent) IsFocusable() bool {
	return true
}

func (m *MockFocusableComponent) IsFocused() bool {
	return m.Focused
}

func (m *MockFocusableComponent) SetSize(width, height int) {
	// Mock implementation - does nothing
}

// createTestTree creates a simple test component tree
func createTestTree() *runtime.LayoutNode {
	// Create focusable components
	comp1 := &MockFocusableComponent{ID: "comp1"}
	comp2 := &MockFocusableComponent{ID: "comp2"}
	comp3 := &MockFocusableComponent{ID: "comp3"}

	// Create nodes
	node1 := &runtime.LayoutNode{
		ID:   "comp1",
		Type: runtime.NodeTypeFlex,
		Component: &core.ComponentInstance{
			ID:       "comp1",
			Instance: comp1,
		},
		MeasuredWidth:  10,
		MeasuredHeight: 5,
	}

	node2 := &runtime.LayoutNode{
		ID:   "comp2",
		Type: runtime.NodeTypeFlex,
		Component: &core.ComponentInstance{
			ID:       "comp2",
			Instance: comp2,
		},
		MeasuredWidth:  10,
		MeasuredHeight: 5,
	}

	node3 := &runtime.LayoutNode{
		ID:   "comp3",
		Type: runtime.NodeTypeFlex,
		Component: &core.ComponentInstance{
			ID:       "comp3",
			Instance: comp3,
		},
		MeasuredWidth:  10,
		MeasuredHeight: 5,
	}

	// Create parent
	root := &runtime.LayoutNode{
		ID:             "root",
		Type:           runtime.NodeTypeFlex,
		Children:       []*runtime.LayoutNode{node1, node2, node3},
		MeasuredWidth:  30,
		MeasuredHeight: 15,
	}

	node1.Parent = root
	node2.Parent = root
	node3.Parent = root

	return root
}

func TestNewManager(t *testing.T) {
	t.Run("Creates new manager", func(t *testing.T) {
		root := &runtime.LayoutNode{ID: "root"}
		fm := NewManager(root)

		assert.NotNil(t, fm)
		assert.NotNil(t, fm.trapManager)
		assert.Equal(t, 0, fm.GetFocusableCount())
	})
}

func TestRefreshFocusables(t *testing.T) {
	t.Run("Collects focusable components", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)

		fm.RefreshFocusables()

		assert.Equal(t, 3, fm.GetFocusableCount())
		focusables := fm.GetFocusableComponents()
		assert.Contains(t, focusables, "comp1")
		assert.Contains(t, focusables, "comp2")
		assert.Contains(t, focusables, "comp3")
	})
}

func TestFocusNext(t *testing.T) {
	t.Run("Moves focus to next component", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		// Focus first
		fm.FocusFirst()
		focusedID, _ := fm.GetFocused()
		assert.Equal(t, "comp1", focusedID)

		// Move to next
		nextID, ok := fm.FocusNext()
		assert.True(t, ok)
		assert.Equal(t, "comp2", nextID)

		focusedID, _ = fm.GetFocused()
		assert.Equal(t, "comp2", focusedID)
	})

	t.Run("Wraps around to first", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		// Focus last
		fm.FocusSpecific("comp3")
		focusedID, _ := fm.GetFocused()
		assert.Equal(t, "comp3", focusedID)

		// Move to next should wrap to first
		nextID, ok := fm.FocusNext()
		assert.True(t, ok)
		assert.Equal(t, "comp1", nextID)
	})
}

func TestFocusPrev(t *testing.T) {
	t.Run("Moves focus to previous component", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		// Focus last
		fm.FocusSpecific("comp3")
		focusedID, _ := fm.GetFocused()
		assert.Equal(t, "comp3", focusedID)

		// Move to previous
		prevID, ok := fm.FocusPrev()
		assert.True(t, ok)
		assert.Equal(t, "comp2", prevID)

		focusedID, _ = fm.GetFocused()
		assert.Equal(t, "comp2", focusedID)
	})

	t.Run("Wraps around to last", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		// Focus first
		fm.FocusFirst()
		focusedID, _ := fm.GetFocused()
		assert.Equal(t, "comp1", focusedID)

		// Move to previous should wrap to last
		prevID, ok := fm.FocusPrev()
		assert.True(t, ok)
		assert.Equal(t, "comp3", prevID)
	})
}

func TestFocusSpecific(t *testing.T) {
	t.Run("Focuses specific component", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		ok := fm.FocusSpecific("comp2")
		assert.True(t, ok)

		focusedID, hasFocus := fm.GetFocused()
		assert.True(t, hasFocus)
		assert.Equal(t, "comp2", focusedID)
	})

	t.Run("Returns false for non-existent component", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		ok := fm.FocusSpecific("nonexistent")
		assert.False(t, ok)
	})
}

func TestFocusFirst(t *testing.T) {
	t.Run("Focuses first component", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		firstID, ok := fm.FocusFirst()
		assert.True(t, ok)
		assert.Equal(t, "comp1", firstID)

		focusedID, hasFocus := fm.GetFocused()
		assert.True(t, hasFocus)
		assert.Equal(t, "comp1", focusedID)
	})

	t.Run("Returns false when no focusable components", func(t *testing.T) {
		root := &runtime.LayoutNode{ID: "root"}
		fm := NewManager(root)
		fm.RefreshFocusables()

		_, ok := fm.FocusFirst()
		assert.False(t, ok)
	})
}

func TestHasFocus(t *testing.T) {
	t.Run("Returns true when component has focus", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		fm.FocusSpecific("comp2")
		assert.True(t, fm.HasFocus("comp2"))
		assert.False(t, fm.HasFocus("comp1"))
	})
}

func TestFocusTrap(t *testing.T) {
	t.Run("FocusTrap creation", func(t *testing.T) {
		modalNode := &runtime.LayoutNode{
			ID:             "modal",
			Type:           runtime.NodeTypeFlex,
			MeasuredWidth:  20,
			MeasuredHeight: 10,
		}

		trap := NewFocusTrap("modal-trap", TrapModal, modalNode)

		assert.Equal(t, "modal-trap", trap.ID)
		assert.Equal(t, TrapModal, trap.Type)
		assert.Equal(t, modalNode, trap.RootNode)
		assert.False(t, trap.IsActive())
	})

	t.Run("FocusTrap activation", func(t *testing.T) {
		modalNode := &runtime.LayoutNode{
			ID:             "modal",
			Type:           runtime.NodeTypeFlex,
			MeasuredWidth:  20,
			MeasuredHeight: 10,
		}

		trap := NewFocusTrap("modal-trap", TrapModal, modalNode)

		trap.Activate()
		assert.True(t, trap.IsActive())

		trap.Deactivate()
		assert.False(t, trap.IsActive())
	})
}

func TestTrapManager(t *testing.T) {
	t.Run("Push and pop traps", func(t *testing.T) {
		tm := NewTrapManager()

		modal1Node := &runtime.LayoutNode{ID: "modal1"}
		modal2Node := &runtime.LayoutNode{ID: "modal2"}

		trap1 := NewFocusTrap("trap1", TrapModal, modal1Node)
		trap2 := NewFocusTrap("trap2", TrapModal, modal2Node)

		// Push first trap
		tm.PushTrap(trap1)
		assert.Equal(t, 1, tm.GetTrapCount())
		assert.True(t, tm.IsTrapActive("trap1"))
		assert.True(t, trap1.IsActive())

		// Push second trap
		tm.PushTrap(trap2)
		assert.Equal(t, 2, tm.GetTrapCount())
		assert.True(t, tm.IsTrapActive("trap2"))
		assert.False(t, trap1.IsActive()) // First trap is deactivated
		assert.True(t, trap2.IsActive())

		// Pop second trap
		popped := tm.PopTrap()
		assert.Equal(t, trap2, popped)
		assert.Equal(t, 1, tm.GetTrapCount())
		assert.True(t, tm.IsTrapActive("trap1"))
		assert.True(t, trap1.IsActive()) // Reactivated
	})

	t.Run("Remove specific trap", func(t *testing.T) {
		tm := NewTrapManager()

		modal1Node := &runtime.LayoutNode{ID: "modal1"}
		modal2Node := &runtime.LayoutNode{ID: "modal2"}

		trap1 := NewFocusTrap("trap1", TrapModal, modal1Node)
		trap2 := NewFocusTrap("trap2", TrapModal, modal2Node)

		tm.PushTrap(trap1)
		tm.PushTrap(trap2)

		// Remove first trap (not top)
		removed := tm.RemoveTrap("trap1")
		assert.True(t, removed)
		assert.Equal(t, 1, tm.GetTrapCount())
		assert.True(t, tm.IsTrapActive("trap2"))
	})

	t.Run("Clear all traps", func(t *testing.T) {
		tm := NewTrapManager()

		modal1Node := &runtime.LayoutNode{ID: "modal1"}
		modal2Node := &runtime.LayoutNode{ID: "modal2"}

		trap1 := NewFocusTrap("trap1", TrapModal, modal1Node)
		trap2 := NewFocusTrap("trap2", TrapModal, modal2Node)

		tm.PushTrap(trap1)
		tm.PushTrap(trap2)

		tm.Clear()
		assert.Equal(t, 0, tm.GetTrapCount())
		assert.False(t, tm.HasActiveTrap())
		assert.False(t, trap1.IsActive())
		assert.False(t, trap2.IsActive())
	})
}

func TestFocusManagerWithTraps(t *testing.T) {
	t.Run("Focus navigation respects traps", func(t *testing.T) {
		// Create main tree
		root := createTestTree()

		// Create modal with its own focusable components
		modalComp1 := &MockFocusableComponent{ID: "modal-comp1"}
		modalComp2 := &MockFocusableComponent{ID: "modal-comp2"}

		modalNode1 := &runtime.LayoutNode{
			ID:   "modal-comp1",
			Type: runtime.NodeTypeFlex,
			Component: &core.ComponentInstance{
				ID:       "modal-comp1",
				Instance: modalComp1,
			},
			MeasuredWidth:  10,
			MeasuredHeight: 5,
		}

		modalNode2 := &runtime.LayoutNode{
			ID:   "modal-comp2",
			Type: runtime.NodeTypeFlex,
			Component: &core.ComponentInstance{
				ID:       "modal-comp2",
				Instance: modalComp2,
			},
			MeasuredWidth:  10,
			MeasuredHeight: 5,
		}

		modalRoot := &runtime.LayoutNode{
			ID:             "modal",
			Type:           runtime.NodeTypeFlex,
			Children:       []*runtime.LayoutNode{modalNode1, modalNode2},
			MeasuredWidth:  20,
			MeasuredHeight: 10,
		}

		modalNode1.Parent = modalRoot
		modalNode2.Parent = modalRoot

		// Add modal to the root tree
		root.Children = append(root.Children, modalRoot)
		modalRoot.Parent = root

		fm := NewManager(root)
		fm.RefreshFocusables()

		// Verify we can focus main components initially
		fm.FocusFirst()
		focusedID, _ := fm.GetFocused()
		assert.Contains(t, []string{"comp1", "comp2", "comp3", "modal-comp1", "modal-comp2"}, focusedID)

		// Add focus trap for modal
		trap := NewFocusTrap("modal-trap", TrapModal, modalRoot)
		fm.PushFocusTrap(trap)

		// Now should only focus within modal when trap is active
		fm.FocusFirst()
		focusedID, _ = fm.GetFocused()

		// The focused ID should be one of the modal components
		isModalComponent := (focusedID == "modal-comp1" || focusedID == "modal-comp2")
		assert.True(t, isModalComponent, "Focus should be trapped in modal, got: %s", focusedID)

		// Navigate within trap - should stay in modal
		nextID, _ := fm.FocusNext()
		isNextModalComponent := (nextID == "modal-comp1" || nextID == "modal-comp2")
		assert.True(t, isNextModalComponent, "Next focus should still be in modal, got: %s", nextID)
	})

	t.Run("Popping trap restores normal navigation", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		// Create modal
		modalRoot := &runtime.LayoutNode{
			ID:             "modal",
			Type:           runtime.NodeTypeFlex,
			MeasuredWidth:  20,
			MeasuredHeight: 10,
		}

		trap := NewFocusTrap("modal-trap", TrapModal, modalRoot)
		fm.PushFocusTrap(trap)

		assert.True(t, fm.HasActiveFocusTrap())

		// Pop trap
		fm.PopFocusTrap()
		assert.False(t, fm.HasActiveFocusTrap())
	})
}

func TestClear(t *testing.T) {
	t.Run("Clears all focus state", func(t *testing.T) {
		root := createTestTree()
		fm := NewManager(root)
		fm.RefreshFocusables()

		// Set focus
		fm.FocusFirst()
		assert.True(t, fm.HasFocus("comp1"))

		// Add trap
		modalRoot := &runtime.LayoutNode{
			ID:             "modal",
			Type:           runtime.NodeTypeFlex,
			MeasuredWidth:  20,
			MeasuredHeight: 10,
		}
		trap := NewFocusTrap("modal-trap", TrapModal, modalRoot)
		fm.PushFocusTrap(trap)

		// Clear
		fm.Clear()

		// Verify everything is cleared
		_, hasFocus := fm.GetFocused()
		assert.False(t, hasFocus)
		assert.Equal(t, 0, fm.GetFocusableCount())
		assert.False(t, fm.HasActiveFocusTrap())
	})
}
