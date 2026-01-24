package runtime

// ApplyAbsoluteLayout applies absolute positioning to children of a parent node.
// This function calculates the absolute position (AbsoluteX, AbsoluteY) for each
// child node based on its Position properties.
//
// For absolutely positioned children:
//   - top/offsets are relative to the parent's position
//   - right/bottom are calculated as: parent.right - offset - child.width
//   - These children are removed from normal flex flow
//
// For relatively positioned children:
//   - AbsoluteX/AbsoluteY are set to X/Y (normal flow position)
func ApplyAbsoluteLayout(parent *LayoutNode) {
	if parent == nil {
		return
	}

	parentWidth := parent.MeasuredWidth
	parentHeight := parent.MeasuredHeight

	for _, child := range parent.Children {
		if child.Position.Type == PositionAbsolute {
			// Calculate absolute position relative to parent
			child.AbsoluteX = parent.X
			child.AbsoluteY = parent.Y

			// Apply left offset
			if child.Position.Left != nil {
				child.AbsoluteX = parent.X + *child.Position.Left
			}

			// Apply top offset
			if child.Position.Top != nil {
				child.AbsoluteY = parent.Y + *child.Position.Top
			}

			// Apply right offset (overrides left if both specified)
			if child.Position.Right != nil {
				// X = ParentX + ParentWidth - Right - ChildWidth
				child.AbsoluteX = parent.X + parentWidth - *child.Position.Right - child.MeasuredWidth
			}

			// Apply bottom offset (overrides top if both specified)
			if child.Position.Bottom != nil {
				// Y = ParentY + ParentHeight - Bottom - ChildHeight
				child.AbsoluteY = parent.Y + parentHeight - *child.Position.Bottom - child.MeasuredHeight
			}
		} else {
			// For relative positioning, AbsoluteX/Y = X/Y
			child.AbsoluteX = child.X
			child.AbsoluteY = child.Y

			// Recursively apply absolute layout to children
			ApplyAbsoluteLayout(child)
		}
	}
}

// IsPositionAbsolute returns true if the node has absolute positioning
func (n *LayoutNode) IsPositionAbsolute() bool {
	return n != nil && n.Position.Type == PositionAbsolute
}

// GetAbsolutePosition returns the absolute position (X, Y) of the node
// For relative nodes, this returns X, Y
// For absolute nodes, this returns AbsoluteX, AbsoluteY
func (n *LayoutNode) GetAbsolutePosition() (int, int) {
	if n == nil {
		return 0, 0
	}
	if n.Position.Type == PositionAbsolute {
		return n.AbsoluteX, n.AbsoluteY
	}
	return n.X, n.Y
}
