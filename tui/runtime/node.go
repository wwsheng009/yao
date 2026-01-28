package runtime

import "github.com/yaoapp/yao/tui/runtime/priority"

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

	// Position defines absolute/relative positioning
	Position Position

	// Component is optional. Leaf nodes (actual UI components) should have this.
	// Container nodes (flex, row, column) typically don't.
	Component *ComponentRef

	// Tree structure
	Parent   *LayoutNode
	Children []*LayoutNode

	// Runtime fields (write-only for Runtime)
	// These MUST NOT be modified by anyone except the Runtime layout system

	// X, Y is the final position within the parent container
	X, Y int

	// AbsoluteX, AbsoluteY is the computed absolute position (for absolutely positioned elements)
	// For relative elements, these are the same as X, Y
	AbsoluteX, AbsoluteY int

	// MeasuredWidth, MeasuredHeight is the size calculated in the Measure phase
	// This represents the content size (before padding/margin if implemented)
	MeasuredWidth  int
	MeasuredHeight int

	// MeasuredContentWidth, MeasuredContentHeight is the size of the actual content
	// This is MeasuredWidth/Height minus padding (v1: simplified, may not use)
	MeasuredContentWidth  int
	MeasuredContentHeight int

	// Dirty flags - separated for optimization
	// layoutDirty indicates the node needs measure/layout phase (size changed)
	layoutDirty bool
	// paintDirty indicates the node needs render phase only (content changed, size same)
	paintDirty bool

	// Priority level for time-sliced rendering
	priorityLevel priority.DirtyLevel

	// cacheKey is used for measurement caching
	cacheKey string
}

// NewLayoutNode creates a new LayoutNode
func NewLayoutNode(id string, nodeType NodeType, style Style) *LayoutNode {
	return &LayoutNode{
		ID:            id,
		Type:          nodeType,
		Style:         style,
		Props:         make(map[string]interface{}),
		layoutDirty:   true,
		priorityLevel: priority.DirtyNormal, // Default priority
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

// MarkDirty marks this node and all descendants as both layout and paint dirty
// This is the "conservative" default when unsure what changed
func (n *LayoutNode) MarkDirty() {
	if n == nil {
		return
	}
	n.layoutDirty = true
	n.paintDirty = true
	for _, child := range n.Children {
		child.MarkDirty()
	}
}

// MarkLayoutDirty marks this node as needing layout (measure/layout phase)
// Layout dirtiness propagates to ancestors since size changes may affect parent layout
func (n *LayoutNode) MarkLayoutDirty() {
	if n == nil || n.layoutDirty {
		return
	}
	n.layoutDirty = true
	// Size change may affect parent layout
	if n.Parent != nil {
		n.Parent.MarkLayoutDirty()
	}
}

// MarkPaintDirty marks this node as needing repaint only
// Paint dirtiness does NOT propagate to ancestors (content change doesn't affect layout)
func (n *LayoutNode) MarkPaintDirty() {
	if n == nil {
		return
	}
	n.paintDirty = true
}

// IsLayoutDirty returns true if node needs layout (measure/layout phase)
func (n *LayoutNode) IsLayoutDirty() bool {
	return n != nil && n.layoutDirty
}

// IsPaintDirty returns true if node needs paint (render phase only)
func (n *LayoutNode) IsPaintDirty() bool {
	return n != nil && n.paintDirty
}

// IsDirty returns true if the node needs either layout or paint
func (n *LayoutNode) IsDirty() bool {
	return n != nil && (n.layoutDirty || n.paintDirty)
}

// ClearLayoutDirty clears the layout dirty flag
func (n *LayoutNode) ClearLayoutDirty() {
	if n != nil {
		n.layoutDirty = false
	}
}

// ClearPaintDirty clears the paint dirty flag
func (n *LayoutNode) ClearPaintDirty() {
	if n != nil {
		n.paintDirty = false
	}
}

// ClearDirty clears both layout and paint dirty flags
func (n *LayoutNode) ClearDirty() {
	if n != nil {
		n.layoutDirty = false
		n.paintDirty = false
	}
}

// SetPriority sets the node's priority level
func (n *LayoutNode) SetPriority(level priority.DirtyLevel) {
	if n != nil {
		n.priorityLevel = level
	}
}

// GetPriority returns the node's priority level
func (n *LayoutNode) GetPriority() priority.DirtyLevel {
	if n == nil {
		return priority.DirtyNormal
	}
	return n.priorityLevel
}

// Measure attempts to measure this node's component.
// Returns the measured size, or (0, 0) if the component is nil or not measurable.
func (n *LayoutNode) Measure(c BoxConstraints) Size {
	if n.Component == nil {
		return Size{}
	}
	return n.Component.Measure(c)
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
