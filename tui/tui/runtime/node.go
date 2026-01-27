package runtime

// ===========================================================================
// Layout Node
// ===========================================================================

// ComponentRef is a lightweight reference to a component.
// The Instance field holds the actual component (typically implementing
// core.ComponentInterface or similar).
type ComponentRef struct {
	ID       string
	Type     string
	Instance interface{}
}

// LayoutNode represents a node in the layout tree.
type LayoutNode struct {
	// Identification
	ID    string
	Type  NodeType
	Style Style
	Props map[string]interface{}

	// Position
	Position Position

	// Component (optional - for leaf nodes)
	Component *ComponentRef

	// Tree structure
	Parent   *LayoutNode
	Children []*LayoutNode

	// Layout output (computed by runtime)
	X, Y           int
	AbsoluteX, AbsoluteY int
	MeasuredWidth, MeasuredHeight int

	// Dirty flag
	Dirty bool
}

// NewLayoutNode creates a new LayoutNode.
func NewLayoutNode(id string, nodeType NodeType, style Style) *LayoutNode {
	return &LayoutNode{
		ID:      id,
		Type:    nodeType,
		Style:   style,
		Props:   make(map[string]interface{}),
		Dirty:   true,
	}
}

// AddChild adds a child node.
func (n *LayoutNode) AddChild(child *LayoutNode) {
	if child == nil {
		return
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}

// AddChildren adds multiple children.
func (n *LayoutNode) AddChildren(children ...*LayoutNode) {
	for _, child := range children {
		n.AddChild(child)
	}
}

// MarkDirty marks this node and all descendants as dirty.
func (n *LayoutNode) MarkDirty() {
	if n == nil {
		return
	}
	n.Dirty = true
	for _, child := range n.Children {
		child.MarkDirty()
	}
}

// IsDirty returns true if the node needs re-layout.
func (n *LayoutNode) IsDirty() bool {
	return n != nil && n.Dirty
}

// ClearDirty clears the dirty flag.
func (n *LayoutNode) ClearDirty() {
	if n == nil {
		return
	}
	n.Dirty = false
}

// GetBounds returns the node's bounding rectangle.
func (n *LayoutNode) GetBounds() (x, y, w, h int) {
	if n == nil {
		return 0, 0, 0, 0
	}
	return n.X, n.Y, n.MeasuredWidth, n.MeasuredHeight
}

// ContainsPoint checks if a point is within this node's bounds.
func (n *LayoutNode) ContainsPoint(x, y int) bool {
	if n == nil {
		return false
	}
	return x >= n.X && x < n.X+n.MeasuredWidth &&
		y >= n.Y && y < n.Y+n.MeasuredHeight
}

// GetInnerBounds returns the content area bounds (x + padding, width - padding).
func (n *LayoutNode) GetInnerBounds() (x, y, w, h int) {
	if n == nil {
		return 0, 0, 0, 0
	}
	x = n.X + n.Style.Padding.Left
	y = n.Y + n.Style.Padding.Top
	w = n.MeasuredWidth - n.Style.Padding.Left - n.Style.Padding.Right
	h = n.MeasuredHeight - n.Style.Padding.Top - n.Style.Padding.Bottom
	return
}

// FindByID finds a descendant node by ID.
func (n *LayoutNode) FindByID(id string) *LayoutNode {
	if n == nil {
		return nil
	}
	if n.ID == id {
		return n
	}
	for _, child := range n.Children {
		if found := child.FindByID(id); found != nil {
			return found
		}
	}
	return nil
}

// NewComponentRef creates a new ComponentRef.
func NewComponentRef(id, compType string, instance interface{}) *ComponentRef {
	return &ComponentRef{
		ID:       id,
		Type:     compType,
		Instance: instance,
	}
}
