package runtime

// RuntimeImpl is the default implementation of the Runtime interface.
//
// v1: Implementation focuses on basic layout and rendering.
//
// This implementation:
//   - Performs Measure phase using PerformMeasure
//   - Performs Layout phase using PerformLayout
//   - Performs Render phase to generate frames
type RuntimeImpl struct {
	width     int
	height    int
	lastFrame *Frame
}

// NewRuntime creates a new RuntimeImpl with the given dimensions.
func NewRuntime(width, height int) *RuntimeImpl {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return &RuntimeImpl{
		width:  width,
		height: height,
	}
}

// Layout performs a complete layout pass on the root node.
//
// This includes:
//  1. Measure phase: Calculate intrinsic sizes
//  2. Layout phase: Assign positions
//
// Returns a LayoutResult containing all positioned nodes.
func (r *RuntimeImpl) Layout(root *LayoutNode, c BoxConstraints) LayoutResult {
	if root == nil {
		return LayoutResult{}
	}

	// Phase 1: Measure (bottom-up)
	PerformMeasure(root, c)

	// Phase 2: Layout (top-down)
	r.layoutNode(root, c)
	root.ClearDirty()

	// Collect boxes
	boxes := r.collectBoxes(root)

	return LayoutResult{
		Boxes:      boxes,
		RootWidth:  root.MeasuredWidth,
		RootHeight: root.MeasuredHeight,
		Dirty:      false, // Will be enhanced in v1.1 with dirty tracking
	}
}

// layoutNode performs the layout phase for a single node and its children.
func (r *RuntimeImpl) layoutNode(node *LayoutNode, c BoxConstraints) {
	if node == nil {
		return
	}

	// Set initial position (will be adjusted by parents)
	// Root node starts at (0, 0)
	if node.Parent == nil {
		node.X = 0
		node.Y = 0
	}

	// Layout children based on direction
	switch node.Type {
	case NodeTypeFlex, NodeTypeRow, NodeTypeColumn:
		r.layoutFlexChildren(node)
	case NodeTypeText, NodeTypeCustom:
		// Leaf nodes: children already positioned by parent
	default:
		// Unknown type: just stack children vertically
		r.layoutDefault(node)
	}
}

// layoutFlexChildren layouts children in a flex layout.
func (r *RuntimeImpl) layoutFlexChildren(node *LayoutNode) {
	if len(node.Children) == 0 {
		return
	}

	innerX := node.X + node.Style.Padding.Left
	innerY := node.Y + node.Style.Padding.Top
	availableWidth := node.MeasuredWidth - node.Style.Padding.Left - node.Style.Padding.Right
	availableHeight := node.MeasuredHeight - node.Style.Padding.Top - node.Style.Padding.Bottom

	isRow := node.Style.Direction == DirectionRow

	if isRow {
		// Row layout: place children horizontally
		curX := innerX
		totalChildWidth := 0
		for _, child := range node.Children {
			totalChildWidth += child.MeasuredWidth
		}

		// Calculate starting position based on justify
		switch node.Style.Justify {
		case JustifyCenter:
			curX = innerX + (availableWidth-totalChildWidth)/2
		case JustifyEnd:
			curX = innerX + availableWidth - totalChildWidth
		case JustifyStart, JustifySpaceBetween, JustifySpaceAround, JustifySpaceEvenly:
			curX = innerX
		}

		for _, child := range node.Children {
			child.X = curX
			child.Y = innerY

			// Recursively layout child
			r.layoutNode(child, BoxConstraints{
				MinWidth:  child.MeasuredWidth,
				MaxWidth:  child.MeasuredWidth,
				MinHeight: child.MeasuredHeight,
				MaxHeight: child.MeasuredHeight,
			})

			curX += child.MeasuredWidth + node.Style.Gap
			r.layoutNode(child, BoxConstraints{}) // Ensure children are laid out
		}
	} else {
		// Column layout: place children vertically
		curY := innerY
		totalChildHeight := 0
		for _, child := range node.Children {
			totalChildHeight += child.MeasuredHeight
		}

		// Calculate starting position based on justify
		switch node.Style.Justify {
		case JustifyCenter:
			curY = innerY + (availableHeight-totalChildHeight)/2
		case JustifyEnd:
			curY = innerY + availableHeight - totalChildHeight
		case JustifyStart, JustifySpaceBetween, JustifySpaceAround, JustifySpaceEvenly:
			curY = innerY
		}

		for _, child := range node.Children {
			child.X = innerX
			child.Y = curY

			// Recursively layout child
			r.layoutNode(child, BoxConstraints{
				MinWidth:  child.MeasuredWidth,
				MaxWidth:  child.MeasuredWidth,
				MinHeight: child.MeasuredHeight,
				MaxHeight: child.MeasuredHeight,
			})

			curY += child.MeasuredHeight + node.Style.Gap
		}
	}
}

// layoutDefault is a fallback layout that stacks children vertically.
func (r *RuntimeImpl) layoutDefault(node *LayoutNode) {
	curY := node.Y + node.Style.Padding.Top
	for _, child := range node.Children {
		child.X = node.X + node.Style.Padding.Left
		child.Y = curY

		r.layoutNode(child, BoxConstraints{
			MinWidth:  child.MeasuredWidth,
			MaxWidth:  child.MeasuredWidth,
			MinHeight: child.MeasuredHeight,
			MaxHeight: child.MeasuredHeight,
		})

		curY += child.MeasuredHeight
	}
}

// collectBoxes collects all LayoutBoxes from a node tree.
func (r *RuntimeImpl) collectBoxes(root *LayoutNode) []LayoutBox {
	var boxes []LayoutBox
	r.collectBoxesRecursive(root, &boxes)
	return boxes
}

// collectBoxesRecursive is a helper for collectBoxes.
func (r *RuntimeImpl) collectBoxesRecursive(node *LayoutNode, boxes *[]LayoutBox) {
	if node == nil {
		return
	}

	// Add current node's box
	*boxes = append(*boxes, NewLayoutBox(node))

	// Recursively collect children
	for _, child := range node.Children {
		r.collectBoxesRecursive(child, boxes)
	}
}

// Render generates a Frame from a LayoutResult.
//
// This creates a CellBuffer and renders all nodes in Z-Index order.
// v1: Simple implementation, will be enhanced in render module.
func (r *RuntimeImpl) Render(result LayoutResult) Frame {
	// Sort boxes by Z-Index
	sortedBoxes := make([]LayoutBox, len(result.Boxes))
	copy(sortedBoxes, result.Boxes)

	// Simple sort by Z-Index
	for i := 0; i < len(sortedBoxes); i++ {
		for j := i + 1; j < len(sortedBoxes); j++ {
			if sortedBoxes[i].ZIndex > sortedBoxes[j].ZIndex {
				sortedBoxes[i], sortedBoxes[j] = sortedBoxes[j], sortedBoxes[i]
			}
		}
	}

	// Create buffer
	buf := NewCellBuffer(r.width, r.height)

	// Render each node
	for _, box := range sortedBoxes {
		if box.Node != nil && box.Node.Component != nil {
			r.renderComponent(buf, box)
		}
	}

	frame := Frame{
		Buffer: buf,
		Width:  r.width,
		Height: r.height,
		Dirty:  false, // Will be enhanced in v1.1
	}

	r.lastFrame = &frame
	return frame
}

// renderComponent renders a component to the CellBuffer.
// v1: Very basic implementation, just renders Text components.
func (r *RuntimeImpl) renderComponent(buf *CellBuffer, box LayoutBox) {
	if box.Node == nil || box.Node.Component == nil || box.Node.Component.Instance == nil {
		return
	}

	// Try to render using core.ComponentInterface.View
	text := box.Node.Component.Instance.View()
	if text == "" {
		return
	}

	style := CellStyle{} // v1: simplified, will use lipgloss.Style in v1.1

	// Render text to buffer
	y := box.Y
	lines := splitLines(text)
	for _, line := range lines {
		if y >= box.Y+box.H {
			break
		}
		for x, r := range line {
			if x >= box.W {
				break
			}
			buf.SetContent(box.X+x, y, box.ZIndex, r, style, box.NodeID)
		}
		y++
	}
}

// UpdateDimensions updates the runtime dimensions.
func (r *RuntimeImpl) UpdateDimensions(width, height int) {
	if width > 0 {
		r.width = width
	}
	if height > 0 {
		r.height = height
	}
}

// Dispatch is a placeholder for event handling (Phase 3).
func (r *RuntimeImpl) Dispatch(ev Event) {
	// v1: Placeholder - will be implemented in Phase 3
}

// FocusNext is a placeholder for focus management (Phase 3).
func (r *RuntimeImpl) FocusNext() {
	// v1: Placeholder - will be implemented in Phase 3
}

// GetWidth returns the runtime width.
func (r *RuntimeImpl) GetWidth() int {
	return r.width
}

// GetHeight returns the runtime height.
func (r *RuntimeImpl) GetHeight() int {
	return r.height
}

// splitLines splits a string into lines.
func splitLines(text string) []string {
	if text == "" {
		return []string{}
	}

	lines := []string{}
	currentLine := ""

	for _, r := range text {
		if r == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(r)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
