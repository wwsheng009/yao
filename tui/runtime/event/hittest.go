package event

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// MouseEvent represents a mouse input event.
type MouseEvent struct {
	X     int
	Y     int
	Type  MouseEventType
	Data  interface{}
	Click MouseClickType
}

// MouseEventType is the type of mouse event.
type MouseEventType string

const (
	MousePress   MouseEventType = "press"
	MouseRelease MouseEventType = "release"
	MouseMove    MouseEventType = "move"
	MouseScroll  MouseEventType = "scroll"
)

// MouseClickType indicates which button was clicked.
type MouseClickType int

const (
	MouseLeft   MouseClickType = iota
	MouseMiddle
	MouseRight
)

// KeyEvent represents a keyboard input event.
type KeyEvent struct {
	Key  rune
	Type KeyEventType
	Mod  KeyModifier
}

// KeyEventType is the type of keyboard event.
type KeyEventType string

const (
	KeyPress KeyEventType = "press"
	KeyRelease KeyEventType = "release"
)

// KeyModifier represents key modifiers (Ctrl, Alt, Shift).
type KeyModifier int

const (
	ModNone KeyModifier = iota
	ModShift
	ModCtrl
	ModAlt
)

// HitTestResult contains information about a hit test.
type HitTestResult struct {
	NodeID   string
	Found    bool
	X, Y     int   // Local coordinates within the node
	Width    int
	Height   int
	Node     *runtime.LayoutNode
}

// HitTest finds the node at a given screen position.
// It searches through the layout boxes in reverse Z-Index order
// (top to bottom) to find the topmost node at the position.
func HitTest(x, y int, boxes []runtime.LayoutBox) *HitTestResult {
	// Search in reverse order (topmost first)
	for i := len(boxes) - 1; i >= 0; i-- {
		box := boxes[i]

		// Check if point is within the box
		if x >= box.X && x < box.X+box.W && y >= box.Y && y < box.Y+box.H {
			// Calculate local coordinates
			localX := x - box.X
			localY := y - box.Y

			return &HitTestResult{
				NodeID: box.NodeID,
				Found:  true,
				X:      localX,
				Y:      localY,
				Width:  box.W,
				Height: box.H,
				Node:   box.Node,
			}
		}
	}

	// No hit found
	return &HitTestResult{
		Found: false,
	}
}

// HitTestNode tests if a point is within a specific node's bounds.
func HitTestNode(x, y int, node *runtime.LayoutNode) bool {
	if node == nil {
		return false
	}

	return x >= node.X && x < node.X+node.MeasuredWidth &&
		y >= node.Y && y < node.Y+node.MeasuredHeight
}

// HitTestChildren tests if a point is within any of the node's children.
// Returns the first child that contains the point, or nil if none.
func HitTestChildren(x, y int, node *runtime.LayoutNode) *runtime.LayoutNode {
	if node == nil {
		return nil
	}

	for _, child := range node.Children {
		if HitTestNode(x, y, child) {
			return child
		}
	}

	return nil
}

// FindFocusableAt finds a focusable component at the given position.
// It traverses the node tree to find components that implement Focusable.
func FindFocusableAt(x, y int, node *runtime.LayoutNode) runtime.FocusableComponent {
	if node == nil {
		return nil
	}

	// Check if the node itself is at the position
	if !HitTestNode(x, y, node) {
		return nil
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				return focusable
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		if result := FindFocusableAt(x, y, child); result != nil {
			return result
		}
	}

	return nil
}

// FindFocusableComponent searches for a focusable component in a node tree.
// It performs a depth-first search to find the first focusable component.
func FindFocusableComponent(node *runtime.LayoutNode) runtime.FocusableComponent {
	if node == nil {
		return nil
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				return focusable
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		if result := FindFocusableComponent(child); result != nil {
			return result
		}
	}

	return nil
}

// CollectFocusableComponents collects all focusable components in a node tree.
func CollectFocusableComponents(node *runtime.LayoutNode) []runtime.FocusableComponent {
	var result []runtime.FocusableComponent

	collectFocusableRecursive(node, &result)

	return result
}

// collectFocusableRecursive is a helper for CollectFocusableComponents.
func collectFocusableRecursive(node *runtime.LayoutNode, result *[]runtime.FocusableComponent) {
	if node == nil {
		return
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				*result = append(*result, focusable)
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		collectFocusableRecursive(child, result)
	}
}

// BoundingBox represents a rectangular bounding box.
type BoundingBox struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewBoundingBox creates a new BoundingBox.
func NewBoundingBox(x, y, width, height int) BoundingBox {
	return BoundingBox{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// Contains checks if a point is within the bounding box.
func (b BoundingBox) Contains(x, y int) bool {
	return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height
}

// Intersects checks if two bounding boxes intersect.
func (b BoundingBox) Intersects(other BoundingBox) bool {
	return b.X < other.X+other.Width &&
		b.X+b.Width > other.X &&
		b.Y < other.Y+other.Height &&
		b.Y+b.Height > other.Y
}

// ToRect converts the bounding box to a Rect.
func (b BoundingBox) ToRect() runtime.Rect {
	return runtime.Rect{
		X:      b.X,
		Y:      b.Y,
		Width:  b.Width,
		Height: b.Height,
	}
}
