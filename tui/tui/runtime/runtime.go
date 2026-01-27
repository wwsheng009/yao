package runtime

import (
	"github.com/yaoapp/kun/log"
)

// ===========================================================================
// Runtime Engine
// ===========================================================================

// Runtime is the interface for the layout runtime engine.
type Runtime interface {
	// Layout performs layout on the root node with constraints.
	Layout(root *LayoutNode, constraints BoxConstraints) LayoutResult

	// Render renders the layout result to a frame.
	Render(result LayoutResult) Frame

	// UpdateDimensions updates the window dimensions.
	UpdateDimensions(width, height int)

	// GetWidth returns the current width.
	GetWidth() int

	// GetHeight returns the current height.
	GetHeight() int

	// MarkDirtyGlobal marks the entire layout as dirty.
	MarkDirtyGlobal()

	// ClearDirty clears the dirty flag.
	ClearDirty()

	// IsDirty returns true if the layout is dirty.
	IsDirty() bool

	// GetSelection returns the selection manager (for text selection).
	GetSelection() interface{}

	// CopySelection copies the selected text to clipboard.
	CopySelection() string

	// GetBoxes returns the current layout boxes.
	GetBoxes() []LayoutBox
}

// RuntimeImpl is the default implementation of Runtime.
type RuntimeImpl struct {
	width  int
	height int
	dirty  bool
	boxes  []LayoutBox
}

// NewRuntime creates a new Runtime.
func NewRuntime(width, height int) Runtime {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return &RuntimeImpl{
		width:  width,
		height: height,
		dirty:  true,
	}
}

// Layout performs layout on the root node.
func (r *RuntimeImpl) Layout(root *LayoutNode, constraints BoxConstraints) LayoutResult {
	if root == nil {
		return LayoutResult{}
	}

	log.Trace("TUI Runtime: Starting layout, constraints: %dx%d to %dx%d",
		constraints.MinWidth, constraints.MaxWidth,
		constraints.MinHeight, constraints.MaxHeight)

	result := LayoutResult{
		RootWidth:  r.width,
		RootHeight: r.height,
		Boxes:      make([]LayoutBox, 0),
	}

	// Perform layout using flex algorithm
	r.layoutNode(root, constraints, 0, 0, &result)

	// Collect all layout boxes
	r.collectBoxes(root, &result)

	// Store boxes for later access
	r.boxes = result.Boxes

	result.Dirty = r.dirty
	r.dirty = false

	return result
}

// layoutNode performs layout on a single node.
func (r *RuntimeImpl) layoutNode(node *LayoutNode, constraints BoxConstraints, x, y int, result *LayoutResult) {
	if node == nil {
		return
	}

	// Calculate available size
	availableWidth := r.width
	if constraints.MaxWidth > 0 {
		availableWidth = constraints.MaxWidth
	}
	availableHeight := r.height
	if constraints.MaxHeight > 0 {
		availableHeight = constraints.MaxHeight
	}

	// Apply absolute positioning
	if node.Style.IsAbsolute() {
		absX := x
		absY := y
		if node.Position.Left != nil {
			absX = *node.Position.Left
		}
		if node.Position.Top != nil {
			absY = *node.Position.Top
		}
		node.AbsoluteX = absX
		node.AbsoluteY = absY
		node.X = absX
		node.Y = absY
	} else {
		node.X = x
		node.Y = y
		node.AbsoluteX = x
		node.AbsoluteY = y
	}

	// Calculate node size
	width := availableWidth
	height := availableHeight

	if node.Style.Width != nil {
		width = *node.Style.Width
	}
	if node.Style.Height != nil {
		height = *node.Style.Height
	}

	node.MeasuredWidth = width
	node.MeasuredHeight = height
	node.Dirty = false

	// Layout children
	if len(node.Children) > 0 {
		childX := x
		childY := y

		for _, child := range node.Children {
			childConstraints := constraints
			r.layoutNode(child, childConstraints, childX, childY, result)

			if node.Style.Direction == DirectionRow {
				childX += child.MeasuredWidth + node.Style.Gap
			} else {
				childY += child.MeasuredHeight + node.Style.Gap
			}
		}
	}
}

// collectBoxes collects all layout boxes from the tree.
func (r *RuntimeImpl) collectBoxes(node *LayoutNode, result *LayoutResult) {
	if node == nil {
		return
	}

	result.Boxes = append(result.Boxes, LayoutBox{
		NodeID: node.ID,
		X:      node.X,
		Y:      node.Y,
		W:      node.MeasuredWidth,
		H:      node.MeasuredHeight,
		ZIndex: node.Style.ZIndex,
	})

	for _, child := range node.Children {
		r.collectBoxes(child, result)
	}
}

// Render renders the layout result to a frame.
func (r *RuntimeImpl) Render(result LayoutResult) Frame {
	// For now, return an empty frame
	// The actual rendering is done by the Model's renderWithRuntime method
	// which calls component View() methods
	return Frame{
		content: "",
		width:   result.RootWidth,
		height:  result.RootHeight,
	}
}

// UpdateDimensions updates the window dimensions.
func (r *RuntimeImpl) UpdateDimensions(width, height int) {
	if width > 0 {
		r.width = width
	}
	if height > 0 {
		r.height = height
	}
	r.dirty = true
}

// GetWidth returns the current width.
func (r *RuntimeImpl) GetWidth() int {
	return r.width
}

// GetHeight returns the current height.
func (r *RuntimeImpl) GetHeight() int {
	return r.height
}

// MarkDirtyGlobal marks the entire layout as dirty.
func (r *RuntimeImpl) MarkDirtyGlobal() {
	r.dirty = true
}

// ClearDirty clears the dirty flag.
func (r *RuntimeImpl) ClearDirty() {
	r.dirty = false
}

// IsDirty returns true if the layout is dirty.
func (r *RuntimeImpl) IsDirty() bool {
	return r.dirty
}

// GetSelection returns the selection manager (placeholder).
func (r *RuntimeImpl) GetSelection() interface{} {
	return nil
}

// CopySelection copies the selected text to clipboard (placeholder).
func (r *RuntimeImpl) CopySelection() string {
	return ""
}

// GetBoxes returns the current layout boxes.
func (r *RuntimeImpl) GetBoxes() []LayoutBox {
	return r.boxes
}
