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
	"sync"
)

// measureCache implements a simple cache for measurement results.
// Cache key is: nodeID + constraints hash.
type measureCache struct {
	mu    sync.RWMutex
	cache map[string]Size
}

// global cache instance
var globalMeasureCache = &measureCache{
	cache: make(map[string]Size),
}

// cacheKey generates a cache key for a node and constraints.
func (c *measureCache) cacheKey(nodeID string, constraints BoxConstraints) string {
	// Simple cache key: nodeID + minW/maxW/minH/maxH
	return fmt.Sprintf("%s:%d,%d,%d,%d", nodeID,
		constraints.MinWidth, constraints.MaxWidth,
		constraints.MinHeight, constraints.MaxHeight)
}

// Get retrieves a cached measurement.
func (c *measureCache) Get(nodeID string, constraints BoxConstraints) (Size, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.cacheKey(nodeID, constraints)
	size, exists := c.cache[key]
	return size, exists
}

// Set stores a measurement in the cache.
func (c *measureCache) Set(nodeID string, constraints BoxConstraints, size Size) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.cacheKey(nodeID, constraints)
	c.cache[key] = size
}

// Invalidate clears the cache (e.g., after window resize).
func (c *measureCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]Size)
}

// InvalidateNode removes all cache entries for a specific node.
func (c *measureCache) InvalidateNode(nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all cache entries starting with nodeID
	for key := range c.cache {
		if len(key) > len(nodeID) && key[:len(nodeID)] == nodeID {
			delete(c.cache, key)
		}
	}
}

// measure performs the measure phase on a node.
// It recursively measures all children and computes the node's size.
func measure(node *LayoutNode, c BoxConstraints) Size {
	if node == nil {
		return Size{Width: 0, Height: 0}
	}

	// Check cache first (for non-dirty nodes)
	if !node.layoutDirty && node.ID != "" {
		if cached, ok := globalMeasureCache.Get(node.ID, c); ok {
			// Use cached measurement
			node.MeasuredWidth = cached.Width
			node.MeasuredHeight = cached.Height
			return cached
		}
	}

	// Calculate inner constraints (subtract padding and border)
	innerC := c
	innerC.MaxWidth = max(0, c.MaxWidth-node.Style.Padding.Left-node.Style.Padding.Right-node.Style.Border.Left-node.Style.Border.Right)
	innerC.MaxHeight = max(0, c.MaxHeight-node.Style.Padding.Top-node.Style.Padding.Bottom-node.Style.Border.Top-node.Style.Border.Bottom)

	// Leaf node: use component's Measure interface
	if len(node.Children) == 0 {
		// Resolve percentages against parent constraints
		explicitWidth := node.Style.Width
		explicitHeight := node.Style.Height

		// Resolve width percentage
		if IsPercent(node.Style.Width) {
			explicitWidth, _ = ResolvePercent(node.Style.Width, c.MaxWidth)
		}
		// Resolve height percentage
		if IsPercent(node.Style.Height) {
			explicitHeight, _ = ResolvePercent(node.Style.Height, c.MaxHeight)
		}

		// If both width and height are explicitly set (including resolved percentages)
		if explicitWidth >= 0 && explicitHeight >= 0 {
			node.MeasuredWidth = explicitWidth
			node.MeasuredHeight = explicitHeight
			return Size{Width: explicitWidth, Height: explicitHeight}
		}

		size := node.Measure(innerC)

		// Apply explicit width or height if set
		if explicitWidth >= 0 {
			size.Width = explicitWidth
		}
		if explicitHeight >= 0 {
			size.Height = explicitHeight
		}

		// Add padding and border back to total size
		size.Width += node.Style.Padding.Left + node.Style.Padding.Right + node.Style.Border.Left + node.Style.Border.Right
		size.Height += node.Style.Padding.Top + node.Style.Padding.Bottom + node.Style.Border.Top + node.Style.Border.Bottom

		// Constrain to parent's constraints
		size.Width, size.Height = c.Constrain(size.Width, size.Height)

		// Store in node
		node.MeasuredWidth = size.Width
		node.MeasuredHeight = size.Height

		// Cache the result for non-dirty nodes with ID
		if !node.layoutDirty && node.ID != "" {
			globalMeasureCache.Set(node.ID, c, size)
		}

		return size
	}

	// Container node: measure children first, then compute container size
	size := measureContainer(node, innerC, c)

	// Cache the result for non-dirty nodes with ID
	if !node.layoutDirty && node.ID != "" {
		globalMeasureCache.Set(node.ID, c, size)
	}

	return size
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

	// Add padding and border back to total size
	size.Width += node.Style.Padding.Left + node.Style.Padding.Right + node.Style.Border.Left + node.Style.Border.Right
	size.Height += node.Style.Padding.Top + node.Style.Padding.Bottom + node.Style.Border.Top + node.Style.Border.Bottom

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
	// Also update child measured sizes to match explicit dimensions
	totalFixedSize := 0
	var growSum float64

	for _, child := range node.Children {
		// For each child, if it has explicit width/height, use that
		// Otherwise, use measured size from leaf node
		var childMainSize int

		if isRow {
			if child.Style.Width >= 0 {
				childMainSize = child.Style.Width
				child.MeasuredWidth = child.Style.Width
			} else {
				childMainSize = child.MeasuredWidth
			}
		} else {
			if child.Style.Height >= 0 {
				childMainSize = child.Style.Height
				child.MeasuredHeight = child.Style.Height
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

	// CRITICAL: For flex layouts with alignment (center, end, space-between, etc.),
	// the container must measure to MAX available size, not just fit content.
	// This allows the layout phase to distribute children within available space.
	useMaxForMainAxis := false
	if node.Style.Justify != JustifyStart {
		// Center, end, space-between, space-around need available space
		useMaxForMainAxis = true
	} else if mainAxisMax > 0 {
		// Even with justify=start, if there are flex-grow children, use max size
		// to allow children to expand
		for _, child := range node.Children {
			if child.Style.FlexGrow > 0 {
				useMaxForMainAxis = true
				break
			}
		}
	}

	if useMaxForMainAxis && mainAxisMax > 0 {
		// Use available space for alignment
		containerMainSize = mainAxisMax
	} else {
		// Main axis: sum of children sizes + gaps
		for _, child := range node.Children {
			if isRow {
				containerMainSize += child.MeasuredWidth
			} else {
				containerMainSize += child.MeasuredHeight
			}
		}
		containerMainSize += totalGap
	}

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

	// Apply explicit size constraints
	// When a node has explicit width/height, use that instead of calculated size
	// Resolve percentages against parent constraints
	if node.Style.Width >= 0 {
		containerMainSize = node.Style.Width
	} else if IsPercent(node.Style.Width) {
		containerMainSize, _ = ResolvePercent(node.Style.Width, outerC.MaxWidth)
	}

	if node.Style.Height >= 0 {
		if isRow {
			containerCrossSize = node.Style.Height
		} else {
			containerMainSize = node.Style.Height
		}
	} else if IsPercent(node.Style.Height) {
		if isRow {
			containerCrossSize, _ = ResolvePercent(node.Style.Height, outerC.MaxHeight)
		} else {
			containerMainSize, _ = ResolvePercent(node.Style.Height, outerC.MaxHeight)
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

// InvalidateMeasureCache clears the entire measurement cache.
// This should be called when the window is resized or when
// major layout changes occur.
func InvalidateMeasureCache() {
	globalMeasureCache.Invalidate()
}

// InvalidateMeasureCacheForNode clears cached measurements for a specific node.
// This should be called when a node's content changes but the overall
// layout structure remains the same.
func InvalidateMeasureCacheForNode(nodeID string) {
	globalMeasureCache.InvalidateNode(nodeID)
}

