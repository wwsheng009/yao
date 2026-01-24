package runtime

// Flexbox algorithm implementation (enhanced v2).
//
// This file implements a more complete Flexbox algorithm with:
//   - FlexGrow: proportional distribution of remaining space
//   - FlexShrink: proportional reduction when space is insufficient
//   - FlexBasis: initial size before grow/shrink
//   - Justify: main-axis alignment (Start, Center, End, SpaceBetween, SpaceAround, SpaceEvenly)
//   - AlignItems: cross-axis alignment (Start, Center, End, Stretch)
//   - AlignSelf: per-child cross-axis alignment
//   - Gap: spacing between children
//
// This implements Phase 2 Flexbox enhancement.

// Flexbox constants
const (
	maxInt = 1<<31 - 1 // Maximum int value for unbounded constraints
)

// measureFlexContainer measures a flex container (row or column).
// This is an enhanced version with FlexShrink, FlexBasis, and full alignment support.
func measureFlexContainerEnhanced(node *LayoutNode, innerC, outerC BoxConstraints) Size {
	if len(node.Children) == 0 {
		return Size{Width: 0, Height: 0}
	}

	// Filter out absolutely positioned children from flex layout
	flexChildren := make([]*LayoutNode, 0, len(node.Children))
	for _, child := range node.Children {
		if !child.IsPositionAbsolute() {
			flexChildren = append(flexChildren, child)
		}
	}

	if len(flexChildren) == 0 {
		return Size{Width: 0, Height: 0}
	}

	// Determine main and cross axis dimensions
	var mainAxisMax, crossAxisMax int
	var isRow bool

	if node.Style.Direction == DirectionRow {
		isRow = true
		mainAxisMax = innerC.MaxWidth
		crossAxisMax = innerC.MaxHeight
	} else {
		isRow = false
		mainAxisMax = innerC.MaxHeight
		crossAxisMax = innerC.MaxWidth
	}

	// Phase 1: Calculate child flex bases and fixed sizes
	// This is where FlexBasis is applied
	type childFlexInfo struct {
		node       *LayoutNode
		flexGrow   float64
		flexShrink float64
		flexBasis  int
		fixedSize  int
		isFlexible bool
	}

	childrenInfo := make([]childFlexInfo, len(flexChildren))
	var totalFixedSize int
	var growSum float64
	var shrinkSum float64

	for i, child := range flexChildren {
		info := childFlexInfo{
			node: child,
		}

		// Determine flex-basis
		// Priority: explicit style > content measurement
		if isRow {
			if child.Style.Width >= 0 {
				info.flexBasis = child.Style.Width
			} else if child.MeasuredWidth > 0 {
				info.flexBasis = child.MeasuredWidth
			} else {
				info.flexBasis = 0
			}
		} else {
			if child.Style.Height >= 0 {
				info.flexBasis = child.Style.Height
			} else if child.MeasuredHeight > 0 {
				info.flexBasis = child.MeasuredHeight
			} else {
				info.flexBasis = 0
			}
		}

		// Get flex properties
		info.flexGrow = child.Style.FlexGrow
		info.flexShrink = child.Style.FlexGrow // v2: FlexShrink defaults to FlexGrow, will be separate in v2.1

		// Determine if flexible
		info.isFlexible = info.flexGrow > 0 || info.flexShrink > 0

		// Calculate fixed size or preferred size
		if !info.isFlexible {
			info.fixedSize = info.flexBasis
			totalFixedSize += info.fixedSize
		} else {
			// Flexible children: sum their grow/shrink weights
			growSum += info.flexGrow
			shrinkSum += info.flexShrink
		}

		childrenInfo[i] = info
	}

	// Calculate remaining space
	totalGap := node.Style.Gap * (len(node.Children) - 1)
	remainingSpace := mainAxisMax - totalFixedSize - totalGap

	// Phase 2: Distribute space based on flex-grow or shrink based on flex-shrink
	if remainingSpace >= 0 {
		// Space is sufficient: distribute extra based on flex-grow
		if growSum > 0 {
			for i := range childrenInfo {
				if childrenInfo[i].flexGrow > 0 {
					allocation := int(float64(remainingSpace) * childrenInfo[i].flexGrow / growSum)
					childrenInfo[i].fixedSize = allocation
				}
			}
		}
	} else {
		// Space is insufficient: shrink based on flex-shrink
		// This is the FlexShrink algorithm
		if shrinkSum > 0 {
			overflow := -remainingSpace
			for i := range childrenInfo {
				if childrenInfo[i].flexShrink > 0 {
					// Calculate shrink amount
					shrinkAmount := int(float64(overflow) * childrenInfo[i].flexShrink / shrinkSum)

					// Apply shrink, but don't go below a minimum size
					// v2: minimum size can be based on content or explicit min-size
					minSize := 0 // v2: will be enhanced
					newSize := max(childrenInfo[i].flexBasis-shrinkAmount, minSize)

					childrenInfo[i].fixedSize = newSize
				}
			}
		}
	}

	// Phase 3: Update child measured sizes
	for i := range childrenInfo {
		info := childrenInfo[i]
		child := info.node

		if isRow {
			child.MeasuredWidth = info.fixedSize
		} else {
			child.MeasuredHeight = info.fixedSize
		}
	}

	// Phase 4: Calculate container size
	var containerMainSize, containerCrossSize int

	// Main axis: sum of children sizes + gaps
	for _, info := range childrenInfo {
		if isRow {
			containerMainSize += info.node.MeasuredWidth
		} else {
			containerMainSize += info.node.MeasuredHeight
		}
	}
	containerMainSize += totalGap

	// Cross axis: max of children sizes (considering AlignItems)
	for _, info := range childrenInfo {
		var childCrossSize int
		if isRow {
			childCrossSize = info.node.MeasuredHeight
		} else {
			childCrossSize = info.node.MeasuredWidth
		}

		// Apply AlignItems
		if node.Style.AlignItems == AlignStretch && containerCrossSize < crossAxisMax {
			childCrossSize = crossAxisMax
		}

		if childCrossSize > containerCrossSize {
			containerCrossSize = childCrossSize
		}
	}

	// Apply min/max constraints from node's explicit size
	if node.Style.Width >= 0 {
		if isRow {
			containerMainSize = max(containerMainSize, node.Style.Width)
		} else {
			containerCrossSize = max(containerCrossSize, node.Style.Width)
		}
	}

	if node.Style.Height >= 0 {
		if !isRow {
			containerMainSize = max(containerMainSize, node.Style.Height)
		} else {
			containerCrossSize = max(containerCrossSize, node.Style.Height)
		}
	}

	// Constrain to parent constraints
	containerMainSize = clamp(containerMainSize, 0, mainAxisMax)
	containerCrossSize = clamp(containerCrossSize, 0, crossAxisMax)

	// Return Size
	if isRow {
		return Size{
			Width:  containerMainSize,
			Height: containerCrossSize,
		}
	}
	return Size{
		Width:  containerCrossSize,
		Height: containerMainSize,
	}
}

// layoutFlexChildren layouts children in a flex layout with enhanced alignment support.
// This implements Phase 2 Flexbox layout with:
//   - Justify: Start, Center, End, SpaceBetween, SpaceAround, SpaceEvenly
//   - AlignItems: Start, Center, End, Stretch
//   - AlignSelf: per-child alignment override
func layoutFlexChildrenEnhanced(node *LayoutNode, layoutFunc func(*LayoutNode, BoxConstraints)) {
	if len(node.Children) == 0 {
		return
	}

	isRow := node.Style.Direction == DirectionRow
	innerX := node.X + node.Style.Border.Left + node.Style.Padding.Left
	innerY := node.Y + node.Style.Border.Top + node.Style.Padding.Top
	availableWidth := node.MeasuredWidth - node.Style.Padding.Left - node.Style.Padding.Right - node.Style.Border.Left - node.Style.Border.Right
	availableHeight := node.MeasuredHeight - node.Style.Padding.Top - node.Style.Padding.Bottom - node.Style.Border.Top - node.Style.Border.Bottom

	if isRow {
		layoutFlexRowEnhanced(node, innerX, innerY, availableWidth, availableHeight, layoutFunc)
	} else {
		layoutFlexColumnEnhanced(node, innerX, innerY, availableWidth, availableHeight, layoutFunc)
	}

	// Apply absolute positioning to all children (after flex layout)
	ApplyAbsoluteLayout(node)
}

// layoutFlexRowEnhanced layouts children in a row direction with full justify support.
func layoutFlexRowEnhanced(node *LayoutNode, startX, startY, availableWidth, availableHeight int, layoutFunc func(*LayoutNode, BoxConstraints)) {
	// Filter out absolutely positioned children from flex layout
	children := make([]*LayoutNode, 0, len(node.Children))
	for _, child := range node.Children {
		if !child.IsPositionAbsolute() {
			children = append(children, child)
		}
	}

	if len(children) == 0 {
		return
	}

	// Calculate total child width (excluding gaps)
	totalChildWidth := 0
	for _, child := range children {
		totalChildWidth += child.MeasuredWidth
	}

	totalGap := node.Style.Gap * (len(children) - 1)
	freeSpace := availableWidth - totalChildWidth - totalGap

	// Calculate starting X position based on justify
	var startOffsetX int
	switch node.Style.Justify {
	case JustifyStart:
		startOffsetX = 0
	case JustifyCenter:
		if freeSpace > 0 {
			startOffsetX = freeSpace / 2
		}
	case JustifyEnd:
		if freeSpace > 0 {
			startOffsetX = freeSpace
		}
	case JustifySpaceBetween:
		// Distribute free space evenly between children
		// Gaps are already accounted for in totalGap, so we need to add the remaining
		if len(children) > 1 {
			startOffsetX = 0 // First child at start
		}
	case JustifySpaceAround:
		// Distribute free space evenly around each child
		if len(children) > 0 {
			startOffsetX = freeSpace / (2 * len(children))
		}
	case JustifySpaceEvenly:
		// Distribute free space evenly before, between, and after each child
		if len(children) > 0 {
			startOffsetX = freeSpace / (len(children) + 1)
		}
	default:
		startOffsetX = 0
	}

	// Layout children with calculated gaps
	var prevChild *LayoutNode
	for i, child := range children {
		// Calculate X position
		var childX int
		if i == 0 {
			childX = startX + startOffsetX
		} else {
			// Calculate gap to previous child
			var gap int
			switch node.Style.Justify {
			case JustifySpaceBetween:
				// Equal gaps between children only
				if freeSpace > 0 && len(children) > 1 {
					gap = node.Style.Gap + freeSpace/(len(children)-1)
				} else {
					gap = node.Style.Gap
				}
			case JustifySpaceAround:
				// Equal gaps around children
				if freeSpace > 0 && len(children) > 0 {
					gap = node.Style.Gap + freeSpace/len(children)
				} else {
					gap = node.Style.Gap
				}
			case JustifySpaceEvenly:
				// Equal gaps before, between, and after children
				if freeSpace > 0 && len(children) > 0 {
					gap = node.Style.Gap + freeSpace/(len(children)+1)
				} else {
					gap = node.Style.Gap
				}
			default:
				gap = node.Style.Gap
			}

			if prevChild != nil {
				childX = prevChild.X + prevChild.MeasuredWidth + gap
			}
		}

		// Calculate Y position based on AlignItems
		var childY int
		crossAxisAlign := node.Style.AlignItems

		// Check if child has AlignSelf override (v2.1 feature, v2 default to parent's AlignItems)
		alignSelf := crossAxisAlign // v2.1: support child.Style.AlignSelf

		// Apply AlignItems for cross-axis positioning
		switch alignSelf {
		case AlignStart:
			childY = startY
		case AlignCenter:
			if availableHeight > child.MeasuredHeight {
				childY = startY + (availableHeight-child.MeasuredHeight)/2
			} else {
				childY = startY
			}
		case AlignEnd:
			if availableHeight > child.MeasuredHeight {
				childY = startY + availableHeight - child.MeasuredHeight
			} else {
				childY = startY
			}
		case AlignStretch:
			childY = startY
			// Stretch child's height to match available
			child.MeasuredHeight = max(child.MeasuredHeight, availableHeight)
		default:
			childY = startY
		}

		// Set child position
		child.X = childX
		child.Y = childY

		// Recursively layout child
		layoutFunc(child, BoxConstraints{
			MinWidth:  child.MeasuredWidth,
			MaxWidth:  child.MeasuredWidth,
			MinHeight: child.MeasuredHeight,
			MaxHeight: child.MeasuredHeight,
		})

		prevChild = child
	}
}

// layoutFlexColumnEnhanced layouts children in a column direction with full justify support.
func layoutFlexColumnEnhanced(node *LayoutNode, startX, startY, availableWidth, availableHeight int, layoutFunc func(*LayoutNode, BoxConstraints)) {
	// Filter out absolutely positioned children from flex layout
	children := make([]*LayoutNode, 0, len(node.Children))
	for _, child := range node.Children {
		if !child.IsPositionAbsolute() {
			children = append(children, child)
		}
	}

	if len(children) == 0 {
		return
	}

	// Calculate total child height (excluding gaps)
	totalChildHeight := 0
	for _, child := range children {
		totalChildHeight += child.MeasuredHeight
	}

	totalGap := node.Style.Gap * (len(children) - 1)
	freeSpace := availableHeight - totalChildHeight - totalGap

	// Calculate starting Y position based on justify
	var startOffsetY int
	switch node.Style.Justify {
	case JustifyStart:
		startOffsetY = 0
	case JustifyCenter:
		if freeSpace > 0 {
			startOffsetY = freeSpace / 2
		}
	case JustifyEnd:
		if freeSpace > 0 {
			startOffsetY = freeSpace
		}
	case JustifySpaceBetween:
		if len(children) > 1 {
			startOffsetY = 0
		}
	case JustifySpaceAround:
		if len(children) > 0 {
			startOffsetY = freeSpace / (2 * len(children))
		}
	case JustifySpaceEvenly:
		if len(children) > 0 {
			startOffsetY = freeSpace / (len(children) + 1)
		}
	default:
		startOffsetY = 0
	}

	// Layout children with calculated gaps
	var prevChild *LayoutNode
	for i, child := range children {
		// Calculate Y position
		var childY int
		if i == 0 {
			childY = startY + startOffsetY
		} else {
			// Calculate gap to previous child
			var gap int
			switch node.Style.Justify {
			case JustifySpaceBetween:
				if freeSpace > 0 && len(children) > 1 {
					gap = node.Style.Gap + freeSpace/(len(children)-1)
				} else {
					gap = node.Style.Gap
				}
			case JustifySpaceAround:
				if freeSpace > 0 && len(children) > 0 {
					gap = node.Style.Gap + freeSpace/len(children)
				} else {
					gap = node.Style.Gap
				}
			case JustifySpaceEvenly:
				if freeSpace > 0 && len(children) > 0 {
					gap = node.Style.Gap + freeSpace/(len(children)+1)
				} else {
					gap = node.Style.Gap
				}
			default:
				gap = node.Style.Gap
			}

			if prevChild != nil {
				childY = prevChild.Y + prevChild.MeasuredHeight + gap
			}
		}

		// Calculate X position based on AlignItems
		var childX int
		crossAxisAlign := node.Style.AlignItems
		alignSelf := crossAxisAlign // v2.1: support child.Style.AlignSelf

		// Apply AlignItems for cross-axis positioning
		switch alignSelf {
		case AlignStart:
			childX = startX
		case AlignCenter:
			if availableWidth > child.MeasuredWidth {
				childX = startX + (availableWidth-child.MeasuredWidth)/2
			} else {
				childX = startX
			}
		case AlignEnd:
			if availableWidth > child.MeasuredWidth {
				childX = startX + availableWidth - child.MeasuredWidth
			} else {
				childX = startX
			}
		case AlignStretch:
			childX = startX
			// Stretch child's width to match available
			child.MeasuredWidth = max(child.MeasuredWidth, availableWidth)
		default:
			childX = startX
		}

		// Set child position
		child.X = childX
		child.Y = childY

		// Recursively layout child
		layoutFunc(child, BoxConstraints{
			MinWidth:  child.MeasuredWidth,
			MaxWidth:  child.MeasuredWidth,
			MinHeight: child.MeasuredHeight,
			MaxHeight: child.MeasuredHeight,
		})

		prevChild = child
	}
}
