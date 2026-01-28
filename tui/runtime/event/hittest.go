package event

import (
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/layout"
)

// =============================================================================
// Hit Testing (Primary API - V3)
// =============================================================================
// 基于 layout.Node 接口的命中测试
// 这是推荐使用的 API，适用于新的 Runtime 核心

// HitTest 在 layout.Node 树中进行命中测试
func HitTest(root layout.Node, x, y int) *HitTestResult {
	if root == nil {
		return &HitTestResult{Found: false}
	}
	return hitTestNode(root, x, y)
}

// HitTestResult 命中测试结果
type HitTestResult struct {
	ComponentID string
	Node        layout.Node
	LocalX      int // 相对于组件的坐标
	LocalY      int
	Found       bool
}

// hitTestNode 递归查找节点
func hitTestNode(node layout.Node, x, y int) *HitTestResult {
	if node == nil {
		return &HitTestResult{Found: false}
	}

	nodeX, nodeY := node.GetPosition()
	nodeWidth, nodeHeight := node.GetSize()

	// 尚未布局的节点跳过
	if nodeWidth <= 0 || nodeHeight <= 0 {
		return &HitTestResult{Found: false}
	}

	// 检查点是否在节点内
	inBounds := x >= nodeX && x < nodeX+nodeWidth &&
		y >= nodeY && y < nodeY+nodeHeight

	if !inBounds {
		return &HitTestResult{Found: false}
	}

	// 先检查子节点（子节点在上层）
	children := node.Children()
	for i := len(children) - 1; i >= 0; i-- {
		childResult := hitTestNode(children[i], x, y)
		if childResult.Found {
			return childResult
		}
	}

	// 如果子节点都没命中，返回当前节点
	return &HitTestResult{
		ComponentID: node.ID(),
		Node:        node,
		LocalX:      x - nodeX,
		LocalY:      y - nodeY,
		Found:       true,
	}
}

// FindAllAt 在给定坐标点查找所有组件（包括被遮挡的）
func FindAllAt(root layout.Node, x, y int) []layout.Node {
	if root == nil {
		return nil
	}
	return findAllAtNode(root, x, y)
}

// findAllAtNode 递归查找所有包含该点的节点
func findAllAtNode(node layout.Node, x, y int) []layout.Node {
	var results []layout.Node

	if node == nil {
		return results
	}

	nodeX, nodeY := node.GetPosition()
	nodeWidth, nodeHeight := node.GetSize()

	if nodeWidth <= 0 || nodeHeight <= 0 {
		return results
	}

	inBounds := x >= nodeX && x < nodeX+nodeWidth &&
		y >= nodeY && y < nodeY+nodeHeight

	if !inBounds {
		return results
	}

	// 先收集子节点
	children := node.Children()
	for _, child := range children {
		childResults := findAllAtNode(child, x, y)
		results = append(results, childResults...)
	}

	// 添加当前节点
	results = append(results, node)

	return results
}

// HitTestRect 查找与矩形相交的所有组件
type Rect struct {
	X, Y, Width, Height int
}

// HitTestRect 查找与矩形相交的所有组件
func HitTestRect(root layout.Node, rect Rect) []layout.Node {
	if root == nil {
		return nil
	}
	return hitTestRectNode(root, rect)
}

// hitTestRectNode 递归查找与矩形相交的节点
func hitTestRectNode(node layout.Node, rect Rect) []layout.Node {
	var results []layout.Node

	if node == nil {
		return results
	}

	nodeX, nodeY := node.GetPosition()
	nodeWidth, nodeHeight := node.GetSize()

	if nodeWidth <= 0 || nodeHeight <= 0 {
		return results
	}

	// 检查矩形是否相交
	nodeRect := Rect{X: nodeX, Y: nodeY, Width: nodeWidth, Height: nodeHeight}
	if !rectsIntersect(nodeRect, rect) {
		return results
	}

	// 先收集子节点
	children := node.Children()
	for _, child := range children {
		childResults := hitTestRectNode(child, rect)
		results = append(results, childResults...)
	}

	// 添加当前节点
	results = append(results, node)

	return results
}

// rectsIntersect 检查两个矩形是否相交
func rectsIntersect(a, b Rect) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

// =============================================================================
// Legacy Hit Testing (Deprecated)
// =============================================================================
// 以下 API 使用已弃用的 runtime.LayoutNode
// 仅用于向后兼容，新代码请使用上面的 layout.Node 版本

// MouseEvent represents a mouse input event.
type MouseEvent struct {
	X     int
	Y     int
	Type  MouseEventType
	Data  interface{}
	Click MouseClickType
}

// MouseEventType is the type of mouse event.
type MouseEventType string

const (
	MousePress   MouseEventType = "press"
	MouseRelease MouseEventType = "release"
	MouseMove    MouseEventType = "move"
	MouseScroll  MouseEventType = "scroll"
)

// MouseClickType indicates which button was clicked.
type MouseClickType int

const (
	MouseLeft   MouseClickType = iota
	MouseMiddle
	MouseRight
)

// KeyEvent represents a keyboard input event.
type KeyEvent struct {
	Key  rune
	Type KeyEventType
	Mod  KeyModifier
}

// KeyEventType is the type of keyboard event.
type KeyEventType string

const (
	KeyPress KeyEventType = "press"
	KeyRelease KeyEventType = "release"
)

// KeyModifier represents key modifiers (Ctrl, Alt, Shift).
type KeyModifier int

const (
	ModNone KeyModifier = iota
	ModShift
	ModCtrl
	ModAlt
)

// LegacyHitTestResult contains information about a hit test (deprecated).
// Use HitTestResult (based on layout.Node) for new code.
type LegacyHitTestResult struct {
	NodeID   string
	Found    bool
	X, Y     int   // Local coordinates within the node
	Width    int
	Height   int
	Node     *runtime.LayoutNode
}

// LegacyHitTest finds the node at a given screen position (deprecated).
// It searches through the layout boxes in reverse Z-Index order
// (top to bottom) to find the topmost node at the position.
//
// Deprecated: Use HitTest with layout.Node for new code.
func LegacyHitTest(x, y int, boxes []runtime.LayoutBox) *LegacyHitTestResult {
	// Search in reverse order (topmost first)
	for i := len(boxes) - 1; i >= 0; i-- {
		box := boxes[i]

		// Check if point is within the box
		if x >= box.X && x < box.X+box.W && y >= box.Y && y < box.Y+box.H {
			// Calculate local coordinates
			localX := x - box.X
			localY := y - box.Y

			return &LegacyHitTestResult{
				NodeID: box.NodeID,
				Found:  true,
				X:      localX,
				Y:      localY,
				Width:  box.W,
				Height: box.H,
				Node:   box.Node,
			}
		}
	}

	// No hit found
	return &LegacyHitTestResult{
		Found: false,
	}
}

// HitTestNode tests if a point is within a specific node's bounds.
func HitTestNode(x, y int, node *runtime.LayoutNode) bool {
	if node == nil {
		return false
	}

	return x >= node.X && x < node.X+node.MeasuredWidth &&
		y >= node.Y && y < node.Y+node.MeasuredHeight
}

// HitTestChildren tests if a point is within any of the node's children.
// Returns the first child that contains the point, or nil if none.
func HitTestChildren(x, y int, node *runtime.LayoutNode) *runtime.LayoutNode {
	if node == nil {
		return nil
	}

	for _, child := range node.Children {
		if HitTestNode(x, y, child) {
			return child
		}
	}

	return nil
}

// FindFocusableAt finds a focusable component at the given position.
// It traverses the node tree to find components that implement Focusable.
func FindFocusableAt(x, y int, node *runtime.LayoutNode) runtime.FocusableComponent {
	if node == nil {
		return nil
	}

	// Check if the node itself is at the position
	if !HitTestNode(x, y, node) {
		return nil
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				return focusable
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		if result := FindFocusableAt(x, y, child); result != nil {
			return result
		}
	}

	return nil
}

// FindFocusableComponent searches for a focusable component in a node tree.
// It performs a depth-first search to find the first focusable component.
func FindFocusableComponent(node *runtime.LayoutNode) runtime.FocusableComponent {
	if node == nil {
		return nil
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				return focusable
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		if result := FindFocusableComponent(child); result != nil {
			return result
		}
	}

	return nil
}

// CollectFocusableComponents collects all focusable components in a node tree.
func CollectFocusableComponents(node *runtime.LayoutNode) []runtime.FocusableComponent {
	var result []runtime.FocusableComponent

	collectFocusableRecursive(node, &result)

	return result
}

// collectFocusableRecursive is a helper for CollectFocusableComponents.
func collectFocusableRecursive(node *runtime.LayoutNode, result *[]runtime.FocusableComponent) {
	if node == nil {
		return
	}

	// Check if this node's component is focusable
	if node.Component != nil && node.Component.Instance != nil {
		if focusable, ok := node.Component.Instance.(runtime.FocusableComponent); ok {
			if focusable.IsFocusable() {
				*result = append(*result, focusable)
			}
		}
	}

	// Recursively check children
	for _, child := range node.Children {
		collectFocusableRecursive(child, result)
	}
}

// BoundingBox represents a rectangular bounding box.
type BoundingBox struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewBoundingBox creates a new BoundingBox.
func NewBoundingBox(x, y, width, height int) BoundingBox {
	return BoundingBox{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// Contains checks if a point is within the bounding box.
func (b BoundingBox) Contains(x, y int) bool {
	return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height
}

// Intersects checks if two bounding boxes intersect.
func (b BoundingBox) Intersects(other BoundingBox) bool {
	return b.X < other.X+other.Width &&
		b.X+b.Width > other.X &&
		b.Y < other.Y+other.Height &&
		b.Y+b.Height > other.Y
}

// ToRect converts the bounding box to a Rect.
func (b BoundingBox) ToRect() runtime.Rect {
	return runtime.Rect{
		X:      b.X,
		Y:      b.Y,
		Width:  b.Width,
		Height: b.Height,
	}
}
