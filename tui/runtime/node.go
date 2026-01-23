package runtime

import "github.com/yaoapp/yao/tui/core"

// LayoutNode is the UI Intermediate Representation (IR) for the Yao TUI Runtime.
//
// It represents a single node in the layout tree, containing all information
// needed by Runtime to calculate layout and render.
//
// Key principles:
//   - DSL/Builder can only fill Type, Style, Props
//   - Runtime is the ONLY entity allowed to write X, Y, MeasuredWidth, MeasuredHeight
//   - Components cannot reverse-modify layout through LayoutNode
type LayoutNode struct {
	// Identification
	ID    string
	Type  NodeType
	Style Style
	Props map[string]interface{}

	// Component is optional. Leaf nodes (actual UI components) should have this.
	// Container nodes (flex, row, column) typically don't.
	Component *core.ComponentInstance

	// Tree structure
	Parent   *LayoutNode
	Children []*LayoutNode

	// Runtime fields (write-only for Runtime)
	// These MUST NOT be modified by anyone except the Runtime layout system

	// X, Y is the final position within the parent container
	X, Y int

	// MeasuredWidth, MeasuredHeight is the size calculated in the Measure phase
	// This represents the content size (before padding/margin if implemented)
	MeasuredWidth  int
	MeasuredHeight int

	// MeasuredContentWidth, MeasuredContentHeight is the size of the actual content
	// This is MeasuredWidth/Height minus padding (v1: simplified, may not use)
	MeasuredContentWidth  int
	MeasuredContentHeight int

	// Dirty flag indicates if the node needs re-layout
	dirty bool

	// cacheKey is used for measurement caching
	cacheKey string
}

// NewLayoutNode creates a new LayoutNode
func NewLayoutNode(id string, nodeType NodeType, style Style) *LayoutNode {
	return &LayoutNode{
		ID:    id,
		Type:  nodeType,
		Style: style,
		Props: make(map[string]interface{}),
		dirty: true,
	}
}

// AddChild adds a child node and sets the child's Parent reference
func (n *LayoutNode) AddChild(child *LayoutNode) {
	if child == nil {
		return
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}

// AddChildren adds multiple children
func (n *LayoutNode) AddChildren(children ...*LayoutNode) {
	for _, child := range children {
		n.AddChild(child)
	}
}

// MarkDirty marks this node and all descendants as dirty
// Dirty nodes will be recalculated in the next layout pass
func (n *LayoutNode) MarkDirty() {
	if n == nil {
		return
	}
	n.dirty = true
	for _, child := range n.Children {
		child.MarkDirty()
	}
}

// IsDirty returns true if the node needs re-layout
func (n *LayoutNode) IsDirty() bool {
	return n != nil && n.dirty
}

// ClearDirty clears the dirty flag
func (n *LayoutNode) ClearDirty() {
	if n == nil {
		return
	}
	n.dirty = false
}

// Measure attempts to measure this node using the Measurable interface
// If the component implements Measurable, it calls Measure with the constraints
// Returns the measured size, or (0, 0) if not measurable
//
// This method handles both:
//   - runtime.Measurable: New interface with BoxConstraints (preferred)
//   - core.Measurable: Legacy interface with maxWidth, maxHeight (for compatibility)
func (n *LayoutNode) Measure(c BoxConstraints) Size {
	if n.Component == nil || n.Component.Instance == nil {
		return Size{Width: 0, Height: 0}
	}

	// First, try the new runtime.Measurable interface
	if measurable, ok := n.Component.Instance.(Measurable); ok {
		return measurable.Measure(c)
	}

	// Fall back to the legacy core.Measurable interface for compatibility
	if coreMeasurable, ok := n.Component.Instance.(core.Measurable); ok {
		// Convert BoxConstraints to the legacy maxWidth, maxHeight format
		maxWidth := c.MaxWidth
		maxHeight := c.MaxHeight

		// If constraints are unbounded (MaxWidth < 0), use a reasonable default
		if maxWidth < 0 {
			maxWidth = 80 // Default terminal width
		}
		if maxHeight < 0 {
			maxHeight = 24 // Default terminal height
		}

		width, height := coreMeasurable.Measure(maxWidth, maxHeight)

		// Apply constraints
		if width < c.MinWidth {
			width = c.MinWidth
		}
		if height < c.MinHeight {
			height = c.MinHeight
		}
		// Apply max constraints (already constrained by component, but enforce)
		if width > maxWidth {
			width = maxWidth
		}
		if height > maxHeight {
			height = maxHeight
		}

		return Size{Width: width, Height: height}
	}

	return Size{Width: 0, Height: 0}
}

// GetBounds returns the node's bounding rectangle (X, Y, Width, Height)
func (n *LayoutNode) GetBounds() (int, int, int, int) {
	if n == nil {
		return 0, 0, 0, 0
	}
	return n.X, n.Y, n.MeasuredWidth, n.MeasuredHeight
}

// ContainsPoint checks if a point (x, y) is within this node's bounds
func (n *LayoutNode) ContainsPoint(x, y int) bool {
	if n == nil {
		return false
	}
	return x >= n.X && x < n.X+n.MeasuredWidth &&
		y >= n.Y && y < n.Y+n.MeasuredHeight
}

// GetInnerBounds returns the content area bounds (x + padding, width - padding)
// v1: simplified version
func (n *LayoutNode) GetInnerBounds() (int, int, int, int) {
	if n == nil {
		return 0, 0, 0, 0
	}
	x := n.X + n.Style.Padding.Left
	y := n.Y + n.Style.Padding.Top
	w := n.MeasuredWidth - n.Style.Padding.Left - n.Style.Padding.Right
	h := n.MeasuredHeight - n.Style.Padding.Top - n.Style.Padding.Bottom
	return x, y, w, h
}

// LayoutBox represents the final layout result for a node
// This is what gets passed to the Renderer
type LayoutBox struct {
	NodeID string
	X, Y   int
	W, H   int
	ZIndex int
	Node   *LayoutNode
}

// NewLayoutBox creates a LayoutBox from a LayoutNode
func NewLayoutBox(node *LayoutNode) LayoutBox {
	zIndex := 0
	if node != nil {
		zIndex = node.Style.ZIndex
	}
	return LayoutBox{
		NodeID: node.ID,
		X:      node.X,
		Y:      node.Y,
		W:      node.MeasuredWidth,
		H:      node.MeasuredHeight,
		ZIndex: zIndex,
		Node:   node,
	}
}

// LayoutResult contains all layout boxes from a layout pass
type LayoutResult struct {
	Boxes      []LayoutBox
	Dirty      bool
	RootWidth  int
	RootHeight int
}

// FindBoxByID finds a LayoutBox by node ID
func (lr *LayoutResult) FindBoxByID(id string) *LayoutBox {
	for i := range lr.Boxes {
		if lr.Boxes[i].NodeID == id {
			return &lr.Boxes[i]
		}
	}
	return nil
}
