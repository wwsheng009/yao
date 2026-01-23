package runtime

import (
	"github.com/yaoapp/yao/tui/runtime/render"
)

// NOTE: Component rendering with lipgloss styling is now handled by the render module
// to maintain module boundary rules (runtime MUST NOT import lipgloss directly).
//
// The render module provides RenderNode() and RenderNodeWithStyle() functions
// for component rendering.

// RuntimeImpl is the default implementation of the Runtime interface.
//
// v1: Implementation focuses on basic layout and rendering.
//
// This implementation:
//   - Performs Measure phase using PerformMeasure
//   - Performs Layout phase using PerformLayout
//   - Performs Render phase to generate frames
//   - Manages event dispatch and focus navigation (Phase 3)
//   - Supports diff rendering for performance optimization (Phase 4)
type RuntimeImpl struct {
	width       int
	height      int
	lastFrame   *Frame
	lastResult  LayoutResult // Cached for event dispatch
	dirtyRegions []Rect      // Track dirty regions for optimization
	forceFullRender bool     // Force full render on next frame
	focusMgr    *FocusManager
	lastRoot    *LayoutNode  // Cached for focus updates
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
		width:    width,
		height:   height,
		focusMgr: NewFocusManager(),
	}
}

// Layout performs a complete layout pass on the root node.
//
// This includes:
//  1. Measure phase: Calculate intrinsic sizes
//  2. Layout phase: Assign positions
//  3. Focus management: Update focusable components
//
// Returns a LayoutResult containing all positioned nodes.
func (r *RuntimeImpl) Layout(root *LayoutNode, c BoxConstraints) LayoutResult {
	if root == nil {
		return LayoutResult{}
	}

	// Cache root for focus updates
	r.lastRoot = root

	// Phase 1: Measure (bottom-up)
	PerformMeasure(root, c)

	// Phase 2: Layout (top-down)
	r.layoutNode(root, c)
	root.ClearDirty()

	// Collect boxes
	boxes := r.collectBoxes(root)

	result := LayoutResult{
		Boxes:      boxes,
		RootWidth:  root.MeasuredWidth,
		RootHeight: root.MeasuredHeight,
		Dirty:      false,
	}

	// Cache result for event dispatch
	r.lastResult = result

	// Update focus manager with new layout
	r.updateFocusManager(root)

	return result
}

// updateFocusManager updates the focus manager with focusable components from the layout.
func (r *RuntimeImpl) updateFocusManager(root *LayoutNode) {
	// Collect focusable components from the layout
	focusable := CollectFocusableFromNode(root)
	if len(focusable) > 0 {
		r.focusMgr.SetFocusable(focusable)
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
// Uses the enhanced flex algorithm with full justify and align support.
func (r *RuntimeImpl) layoutFlexChildren(node *LayoutNode) {
	if len(node.Children) == 0 {
		return
	}

	// Use the enhanced flex layout algorithm
	layoutFlexChildrenEnhanced(node, r.layoutNode)
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
// v1: Enhanced with diff rendering for performance optimization.
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

	// Optimize: If forceFullRender is set or no last frame, render everything
	// Otherwise, only render dirty regions
	if r.forceFullRender || r.lastFrame == nil {
		// Full render
		for _, box := range sortedBoxes {
			if box.Node != nil && box.Node.Component != nil {
				r.renderComponent(buf, box)
			}
		}
		r.forceFullRender = false
		r.dirtyRegions = nil
	} else {
		// Partial render based on dirty regions
		if len(r.dirtyRegions) > 0 {
			for _, region := range r.dirtyRegions {
				// Clear dirty region
				for y := region.Y; y < region.Y+region.Height && y < r.height; y++ {
					for x := region.X; x < region.X+region.Width && x < r.width; x++ {
						buf.SetContentRuntime(x, y, 0, ' ', false, false, false, "")
					}
				}
			}
		}
		// Render all components (they will check dirty state internally)
		for _, box := range sortedBoxes {
			if box.Node != nil && box.Node.Component != nil {
				// Check if component intersects any dirty region or if no regions tracked
				shouldRender := len(r.dirtyRegions) == 0 || r.intersectsDirtyRegion(box)
				if shouldRender {
					r.renderComponent(buf, box)
				}
			}
		}
	}

	frame := Frame{
		Buffer: buf,
		Width:  r.width,
		Height: r.height,
		Dirty:  len(r.dirtyRegions) > 0,
	}

	// Compute diff for next frame
	r.computeDiff(&frame)

	r.lastFrame = &frame
	r.lastResult = result
	return frame
}

// renderComponent renders a component to the CellBuffer.
// Uses the render module for styling to maintain module boundary rules.
func (r *RuntimeImpl) renderComponent(buf *CellBuffer, box LayoutBox) {
	if box.Node == nil || box.Node.Component == nil || box.Node.Component.Instance == nil {
		return
	}

	// Get component view text
	text := box.Node.Component.Instance.View()
	if text == "" {
		return
	}

	// Wrap the buffer with an adapter to avoid circular dependency
	adapter := render.NewCellBufferAdapter(buf)

	// Render text to buffer using the render module
	// This handles multi-line text, wrapping, and styling
	render.RenderNode(adapter, text, box.X, box.Y, box.ZIndex, box.W)
}

// intersectsDirtyRegion checks if a layout box intersects any dirty region.
func (r *RuntimeImpl) intersectsDirtyRegion(box LayoutBox) bool {
	for _, region := range r.dirtyRegions {
		// Check if box intersects region
		if !(box.X+box.W <= region.X || box.X >= region.X+region.Width ||
			box.Y+box.H <= region.Y || box.Y >= region.Y+region.Height) {
			return true
		}
	}
	return false
}

// computeDiff computes the difference between the new frame and the last frame.
// This stores dirty regions for the next render cycle.
func (r *RuntimeImpl) computeDiff(newFrame *Frame) {
	if r.lastFrame == nil || r.lastFrame.Buffer == nil {
		// No previous frame, mark entire frame as dirty
		r.dirtyRegions = []Rect{{
			X:      0,
			Y:      0,
			Width:  r.width,
			Height: r.height,
		}}
		return
	}

	// Convert frames to render.Frame format for diff computation
	oldRenderFrame := render.Frame{
		Buffer: &cellBufferAdapter{r.lastFrame.Buffer},
		Width:  r.lastFrame.Width,
		Height: r.lastFrame.Height,
	}
	newRenderFrame := render.Frame{
		Buffer: &cellBufferAdapter{newFrame.Buffer},
		Width:  newFrame.Width,
		Height: newFrame.Height,
	}

	// Compute diff
	diffResult := render.ComputeDiff(oldRenderFrame, newRenderFrame)

	// Convert render.Rect to runtime.Rect and store
	if diffResult.HasChanges {
		r.dirtyRegions = make([]Rect, len(diffResult.DirtyRegions))
		for i, region := range diffResult.DirtyRegions {
			r.dirtyRegions[i] = Rect{
				X:      region.X,
				Y:      region.Y,
				Width:  region.Width,
				Height: region.Height,
			}
		}
	} else {
		r.dirtyRegions = nil
	}
}

// MarkDirty marks a specific region as dirty, forcing a re-render on the next frame.
func (r *RuntimeImpl) MarkDirty(x, y, width, height int) {
	r.dirtyRegions = append(r.dirtyRegions, Rect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	})
}

// MarkFullRender forces a full render on the next frame.
func (r *RuntimeImpl) MarkFullRender() {
	r.forceFullRender = true
}

// GetDirtyRegions returns the current dirty regions.
func (r *RuntimeImpl) GetDirtyRegions() []Rect {
	return r.dirtyRegions
}

// cellBufferAdapter adapts runtime.CellBuffer to render.FrameBuffer interface.
type cellBufferAdapter struct {
	buffer *CellBuffer
}

func (a *cellBufferAdapter) GetContent(x, y int) render.Cell {
	if a.buffer == nil {
		return render.Cell{}
	}
	cell := a.buffer.GetContent(x, y)
	return render.Cell{
		Char: cell.Char,
		Style: render.CellStyle{
			Bold:      cell.Style.Bold,
			Underline: cell.Style.Underline,
			Italic:    cell.Style.Italic,
		},
	}
}

func (a *cellBufferAdapter) Width() int {
	if a.buffer == nil {
		return 0
	}
	return a.buffer.Width()
}

func (a *cellBufferAdapter) Height() int {
	if a.buffer == nil {
		return 0
	}
	return a.buffer.Height()
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

// DispatchEvent handles an input event using the event dispatch system.
// This is the new method that uses the event package for proper event handling.
// The event package is imported indirectly through the event dispatch functions.
func (r *RuntimeImpl) DispatchEvent(ev Event) {
	// The event package has DispatchEvent which requires boxes
	// We need to get the boxes from the last layout result
	if r.lastResult.Boxes == nil {
		return
	}

	// Import event package functions
	// For now, handle Tab key directly for focus navigation
	if ev.Type == "key" {
		if key, ok := ev.Data.(rune); ok {
			if key == '\t' {
				r.focusMgr.FocusNext()
				return
			}
		}
	}
}

// Dispatch handles an input event (keyboard, mouse, etc.).
// Phase 3: Implementation uses the event dispatch system.
func (r *RuntimeImpl) Dispatch(ev Event) {
	r.DispatchEvent(ev)
}

// FocusNext moves focus to the next focusable component.
// Phase 3: Implementation uses the focus manager.
func (r *RuntimeImpl) FocusNext() {
	r.focusMgr.FocusNext()
}

// FocusPrev moves focus to the previous focusable component.
func (r *RuntimeImpl) FocusPrev() {
	r.focusMgr.FocusPrev()
}

// GetFocusManager returns the runtime's focus manager.
// This allows external code to query or manipulate focus state.
func (r *RuntimeImpl) GetFocusManager() *FocusManager {
	return r.focusMgr
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
