package runtime

// Measure phase: Calculate sizes without considering position.
//
// This is the first phase of the three-phase rendering model.
// It computes the ideal/intrinsic size of each node based on:
//   1. Constraints from parent
//   2. Component's Measurable interface
//   3. Layout algorithm for container nodes
//
// Key rules:
//   - ONLY calculate size, NEVER set X/Y coordinates
//   - Work bottom-up: measure children before parents
//   - Can cache results based on constraints (v1.1)
//
// This file implements the Measure phase for leaf nodes and containers.

import (
	"fmt"
)

// measure performs the measure phase on a node.
// It recursively measures all children and computes the node's size.
func measure(node *LayoutNode, c BoxConstraints) Size {
	if node == nil {
		return Size{Width: 0, Height: 0}
	}

	// Calculate inner constraints (subtract padding)
	innerC := c
	innerC.MaxWidth = max(0, c.MaxWidth-node.Style.Padding.Left-node.Style.Padding.Right)
	innerC.MaxHeight = max(0, c.MaxHeight-node.Style.Padding.Top-node.Style.Padding.Bottom)

	// Leaf node: use component's Measure interface
	if len(node.Children) == 0 {
		size := node.Measure(innerC)

		// Add padding back to total size
		size.Width += node.Style.Padding.Left + node.Style.Padding.Right
		size.Height += node.Style.Padding.Top + node.Style.Padding.Bottom

		// Constrain to parent's constraints
		size.Width, size.Height = c.Constrain(size.Width, size.Height)

		// Store in node
		node.MeasuredWidth = size.Width
		node.MeasuredHeight = size.Height

		return size
	}

	// Container node: measure children first, then compute container size
	return measureContainer(node, innerC, c)
}

// measureContainer measures a container node by measuring all children.
func measureContainer(node *LayoutNode, innerC, outerC BoxConstraints) Size {
	// Measure all children with inner constraints
	for _, child := range node.Children {
		childSize := measure(child, innerC)
		child.MeasuredWidth = childSize.Width
		child.MeasuredHeight = childSize.Height
	}

	// Compute container size based on layout algorithm
	var size Size

	switch node.Type {
	case NodeTypeFlex, NodeTypeRow, NodeTypeColumn:
		size = measureFlexContainer(node, innerC, outerC)
	default:
		size = Size{Width: 0, Height: 0}
	}

	// Add padding back to total size
	size.Width += node.Style.Padding.Left + node.Style.Padding.Right
	size.Height += node.Style.Padding.Top + node.Style.Padding.Bottom

	// Constrain to parent's constraints
	size.Width, size.Height = outerC.Constrain(size.Width, size.Height)

	// Store in node
	node.MeasuredWidth = size.Width
	node.MeasuredHeight = size.Height

	return size
}

// measureFlexContainer measures a flex container (row or column).
// This implements a simplified Flexbox algorithm (v1).
func measureFlexContainer(node *LayoutNode, innerC, outerC BoxConstraints) Size {
	if len(node.Children) == 0 {
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

	// Phase 1: Calculate fixed sizes and flex grow sum
	totalFixedSize := 0
	var growSum float64

	for _, child := range node.Children {
		// For each child, if it has explicit width/height, use that
		// Otherwise, use measured size from leaf node
		var childMainSize int

		if isRow {
			if child.Style.Width >= 0 {
				childMainSize = child.Style.Width
			} else {
				childMainSize = child.MeasuredWidth
			}
		} else {
			if child.Style.Height >= 0 {
				childMainSize = child.Style.Height
			} else {
				childMainSize = child.MeasuredHeight
			}
		}

		if child.Style.FlexGrow > 0 {
			growSum += child.Style.FlexGrow
		} else {
			totalFixedSize += childMainSize
		}
	}

	// Calculate remaining space
	remainingSpace := mainAxisMax - totalFixedSize
	totalGap := node.Style.Gap * (len(node.Children) - 1)
	if node.Style.Gap > 0 && len(node.Children) > 1 {
		remainingSpace -= totalGap
	}

	// Phase 2: Distribute remaining space based on flex-grow
	// Also update child sizes to match
	for i, child := range node.Children {
		var childMainSize int

		if child.Style.FlexGrow > 0 && remainingSpace > 0 {
			// Allocate proportional space based on flex-grow
			allocation := int(float64(remainingSpace) * child.Style.FlexGrow / growSum)
			childMainSize = allocation

			// Update child's measured size
			if isRow {
				child.MeasuredWidth = allocation
			} else {
				child.MeasuredHeight = allocation
			}
		} else {
			// Use fixed size
			if isRow {
				childMainSize = child.MeasuredWidth
			} else {
				childMainSize = child.MeasuredHeight
			}
		}

		if remainingSpace < 0 {
			// If space is negative, shrink children proportionally
			// v1: simplified, may add flex-shrink in v1.1
			childMainSize = max(0, childMainSize+remainingSpace/(len(node.Children)-i))
		}
	}

	// Phase 3: Calculate container size
	var containerMainSize, containerCrossSize int

	// Main axis: sum of children sizes + gaps
	for _, child := range node.Children {
		if isRow {
			containerMainSize += child.MeasuredWidth
		} else {
			containerMainSize += child.MeasuredHeight
		}
	}
	containerMainSize += totalGap

	// Cross axis: max of children sizes
	for _, child := range node.Children {
		var childCrossSize int
		if isRow {
			childCrossSize = child.MeasuredHeight
		} else {
			childCrossSize = child.MeasuredWidth
		}

		if childCrossSize > containerCrossSize {
			containerCrossSize = childCrossSize
		}
	}

	// Apply min/max constraints from node's explicit size
	if node.Style.Width >= 0 {
		if isRow {
			containerMainSize = max(containerMainSize, node.Style.Width)
		}
	}

	if node.Style.Height >= 0 {
		if !isRow {
			containerMainSize = max(containerMainSize, node.Style.Height)
		}
	}

	// Apply AlignItems for cross axis
	if containerCrossSize < crossAxisMax {
		if node.Style.AlignItems == AlignStretch {
			containerCrossSize = crossAxisMax
		}
	}

	// Return Size
	if isRow {
		return Size{
			Width:  min(containerMainSize, outerC.MaxWidth),
			Height: min(containerCrossSize, outerC.MaxHeight),
		}
	}
	return Size{
		Width:  min(containerCrossSize, outerC.MaxWidth),
		Height: min(containerMainSize, outerC.MaxHeight),
	}
}

// min and max helpers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Debug helper
func debugMeasure(node *LayoutNode, msg string) {
	if node == nil {
		return
	}
	fmt.Printf("[Measure] %s: Node %s, Size: %dx%d\n", msg, node.ID,
		node.MeasuredWidth, node.MeasuredHeight)
}

// MeasureAll performs a full measure pass on the tree
func MeasureAll(root *LayoutNode, c BoxConstraints) {
	if root == nil {
		return
	}
	measure(root, c)
}
