package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/runtime"
	uicomponents "github.com/yaoapp/yao/tui/ui/components"
)

// TestPhase2_FlexJustifyCenter tests Justify: Center.
func TestPhase2_FlexJustifyCenter(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionRow).
		WithJustify(runtime.JustifyCenter).
		WithWidth(60)

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add small child
	textComp := uicomponents.NewTextComponent("X")
	textComp.SetContent("X")

	coreInstance := &core.ComponentInstance{
		ID:       "centerTest",
		Type:     "text",
		Instance: textComp,
	}

	childStyle := runtime.NewStyle().WithWidth(10)
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, childStyle)
	child.Component = coreInstance
	rootNode.AddChild(child)

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	childBox := findBoxByID2(result.Boxes, "child")
	assert.NotNil(t, childBox, "Child should be found")

	// Child should be centered in the 60-wide container
	// Container starts at 0, child is 10 wide
	// Centered position: (60 - 10) / 2 = 25
	expectedX := 25
	assert.Equal(t, expectedX, childBox.X, "Child should be centered")
}

// TestPhase2_FlexJustifySpaceBetween tests Justify: SpaceBetween.
func TestPhase2_FlexJustifySpaceBetween(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionRow).
		WithJustify(runtime.JustifySpaceBetween).
		WithWidth(60)

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add 3 children, each 10 wide, total 30
	// With space-between, gaps should be: (60 - 30) / 2 = 15
	for i := 0; i < 3; i++ {
		textComp := uicomponents.NewTextComponent(string(rune('A' + i)))
		textComp.SetContent(string(rune('A' + i)))

		coreInstance := &core.ComponentInstance{
			ID:       "space" + string(rune('1'+i)),
			Type:     "text",
			Instance: textComp,
		}

		childStyle := runtime.NewStyle().WithWidth(10)
		child := runtime.NewLayoutNode("child"+string(rune('1'+i)), runtime.NodeTypeText, childStyle)
		child.Component = coreInstance
		rootNode.AddChild(child)
	}

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	child1 := findBoxByID2(result.Boxes, "child1")
	child2 := findBoxByID2(result.Boxes, "child2")
	child3 := findBoxByID2(result.Boxes, "child3")

	assert.NotNil(t, child1)
	assert.NotNil(t, child2)
	assert.NotNil(t, child3)

	// Verify spacing
	gap1 := child2.X - (child1.X + child1.W)
	gap2 := child3.X - (child2.X + child2.W)

	// Gaps should be roughly equal and > 0
	t.Logf("Gaps: %d, %d", gap1, gap2)
	assert.True(t, gap1 > 0, "Should have gap between child1 and child2")
	assert.True(t, gap2 > 0, "Should have gap between child2 and child3")
	assert.Equal(t, gap1, gap2, "Gaps should be equal with SpaceBetween")
}

// TestPhase2_FlexAlignItemsCenter tests AlignItems: Center.
func TestPhase2_FlexAlignItemsCenter(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionRow).
		WithAlignItems(runtime.AlignCenter).
		WithHeight(20) // Container is 20 high

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add child that's only 5 high
	textComp := uicomponents.NewTextComponent("X")
	textComp.SetContent("X")

	coreInstance := &core.ComponentInstance{
		ID:       "alignTest",
		Type:     "text",
		Instance: textComp,
	}

	childStyle := runtime.NewStyle().WithHeight(5)
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, childStyle)
	child.Component = coreInstance
	child.MeasuredHeight = 5 // Simulate measurement
	rootNode.AddChild(child)

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	childBox := findBoxByID2(result.Boxes, "child")
	assert.NotNil(t, childBox, "Child should be found")

	// Child should be centered vertically
	// (20 - 5) / 2 = 7.5 -> 7 (int division)
	expectedY := 7
	assert.Equal(t, expectedY, childBox.Y, "Child should be vertically centered")
}

// TestPhase2_PaddingIntegration tests that Padding works correctly with Flexbox.
func TestPhase2_PaddingIntegration(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	// Container with padding
	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionRow).
		WithPadding(runtime.NewInsets(5, 10, 5, 10)) // Top:5, Right:10, Bottom:5, Left:10

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add child
	textComp := uicomponents.NewTextComponent("Padding Test")
	textComp.SetContent("X")

	coreInstance := &core.ComponentInstance{
		ID:       "paddingTest",
		Type:     "text",
		Instance: textComp,
	}

	childStyle := runtime.NewStyle()
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, childStyle)
	child.Component = coreInstance
	rootNode.AddChild(child)

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	childBox := findBoxByID2(result.Boxes, "child")
	assert.NotNil(t, childBox, "Child should be found")

	// Child should be offset by left padding (10)
	assert.Equal(t, 10, childBox.X, "Child X should account for left padding")

	// Root size should include padding
	rootBox := findBoxByID2(result.Boxes, "root")
	assert.NotNil(t, rootBox, "Root should be found")
	t.Logf("Root size: %dx%d", rootBox.W, rootBox.H)
	assert.Greater(t, rootBox.W, childBox.W, "Root should be wider than child (includes padding)")
}

// TestPhase2_ColumnLayout tests Column layout with enhanced features.
func TestPhase2_ColumnLayout(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionColumn).
		WithAlignItems(runtime.AlignCenter).
		WithHeight(20)

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add 3 children
	for i := 0; i < 3; i++ {
		textComp := uicomponents.NewTextComponent(string(rune('A' + i)))
		textComp.SetContent(string(rune('A' + i)))

		coreInstance := &core.ComponentInstance{
			ID:       "col" + string(rune('1'+i)),
			Type:     "text",
			Instance: textComp,
		}

		childStyle := runtime.NewStyle()
		child := runtime.NewLayoutNode("child"+string(rune('1'+i)), runtime.NodeTypeText, childStyle)
		child.Component = coreInstance
		rootNode.AddChild(child)
	}

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	// Verify children are stacked vertically
	child1 := findBoxByID2(result.Boxes, "child1")
	child2 := findBoxByID2(result.Boxes, "child2")
	child3 := findBoxByID2(result.Boxes, "child3")

	assert.NotNil(t, child1)
	assert.NotNil(t, child2)
	assert.NotNil(t, child3)

	// Child Y should increase
	assert.True(t, child1.Y < child2.Y, "Child1 should be above Child2")
	assert.True(t, child2.Y < child3.Y, "Child2 should be above Child3")

	// Child X should be same (stacked vertically)
	assert.Equal(t, child1.X, child2.X, "Children should have same X in column layout")
	assert.Equal(t, child2.X, child3.X, "Children should have same X in column layout")

	t.Logf("Column layout: child1=(%d,%d), child2=(%d,%d), child3=(%d,%d)",
		child1.X, child1.Y, child2.X, child2.Y, child3.X, child3.Y)
}

// TestPhase2_ScrollableViewport tests basic viewport/scroll functionality.
func TestPhase2_ScrollableViewport(t *testing.T) {
	// Create a viewport
	viewport := runtime.NewViewport(20, 10) // 20x10 viewport

	// Content is larger than viewport
	viewport.SetContentSize(50, 30)

	// Test scrolling
	viewport.ScrollBy(5, 5)
	x, y := viewport.GetScrollOffset()
	assert.Equal(t, 5, x, "Should have scrolled X by 5")
	assert.Equal(t, 5, y, "Should have scrolled Y by 5")

	// Test visibility check
	vx, vy, vw, vh := viewport.GetVisibleRect()
	t.Logf("Visible rect: x=%d, y=%d, w=%d, h=%d", vx, vy, vw, vh)
	assert.Equal(t, 20, vw, "Visible rect width should match viewport")
	assert.Equal(t, 10, vh, "Visible rect height should match viewport")
}

// TestPhase2_NodeInViewport tests node visibility in viewport.
func TestPhase2_NodeInViewport(t *testing.T) {
	// Viewport: (0,0) to (20,10)
	viewportX, viewportY, viewportW, viewportH := 0, 0, 20, 10

	// Test cases
	tests := []struct {
		name     string
		node     struct{ x, y, w, h int }
		expected bool
	}{
		{
			name:     "Fully visible",
			node:     struct{ x, y, w, h int }{5, 5, 10, 5},
			expected: true,
		},
		{
			name:     "Partially visible (left edge)",
			node:     struct{ x, y, w, h int }{-5, 5, 10, 5},
			expected: true,
		},
		{
			name:     "Outside viewport (left)",
			node:     struct{ x, y, w, h int }{-20, 5, 10, 5},
			expected: false,
		},
		{
			name:     "Outside viewport (below)",
			node:     struct{ x, y, w, h int }{5, 15, 10, 5},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runtime.IsNodeInViewport(
				tt.node.x, tt.node.y, tt.node.w, tt.node.h,
				viewportX, viewportY, viewportW, viewportH,
			)
			assert.Equal(t, tt.expected, result, "Visibility check should match")
		})
	}
}

// TestPhase2_GapSupport tests Gap property in Flex layouts.
func TestPhase2_GapSupport(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionRow).
		WithGap(10) // 10 character gap between children

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add 2 children, each 10 wide
	for i := 0; i < 2; i++ {
		textComp := uicomponents.NewTextComponent(string(rune('A' + i)))
		textComp.SetContent(string(rune('A' + i)))

		coreInstance := &core.ComponentInstance{
			ID:       "gap" + string(rune('1'+i)),
			Type:     "text",
			Instance: textComp,
		}

		childStyle := runtime.NewStyle().WithWidth(10)
		child := runtime.NewLayoutNode("child"+string(rune('1'+i)), runtime.NodeTypeText, childStyle)
		child.Component = coreInstance
		rootNode.AddChild(child)
	}

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	child1 := findBoxByID2(result.Boxes, "child1")
	child2 := findBoxByID2(result.Boxes, "child2")

	assert.NotNil(t, child1)
	assert.NotNil(t, child2)

	// Gap should be 10 between children
	gap := child2.X - (child1.X + child1.W)
	assert.Equal(t, 10, gap, "Gap between children should be 10")
}

// TestPhase2_ZIndexRendering tests that Z-Index is respected during rendering.
func TestPhase2_ZIndexRendering(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	flexStyle := runtime.NewStyle()

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add 2 overlapping children with different Z-Index
	for i := 0; i < 2; i++ {
		textComp := uicomponents.NewTextComponent(string(rune('A' + i)))
		textComp.SetContent(string(rune('A' + i)))

		coreInstance := &core.ComponentInstance{
			ID:       "z" + string(rune('1'+i)),
			Type:     "text",
			Instance: textComp,
		}

		childStyle := runtime.NewStyle()
		child := runtime.NewLayoutNode("child"+string(rune('1'+i)), runtime.NodeTypeText, childStyle)
		child.Component = coreInstance
		child.X = 5 // Same position
		child.Y = 5 // Same position

		// Second child has higher Z-Index
		if i == 1 {
			child.Style = runtime.NewStyle().WithZIndex(10)
		}

		rootNode.AddChild(child)
	}

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	// Render the frame
	frame := runtimeImpl.Render(result)
	assert.NotNil(t, frame, "Frame should not be nil")
	assert.NotNil(t, frame.Buffer, "Frame buffer should not be nil")

	// Verify both boxes exist
	assert.Equal(t, 3, len(result.Boxes), "Should have root and 2 children")

	// Verify Z-Index values
	child1Box := findBoxByID2(result.Boxes, "child1")
	child2Box := findBoxByID2(result.Boxes, "child2")
	assert.NotNil(t, child1Box)
	assert.NotNil(t, child2Box)
	assert.Less(t, child1Box.ZIndex, child2Box.ZIndex, "Child2 should have higher Z-Index")
}

// Helper function to find a LayoutBox by node ID
func findBoxByID2(boxes []runtime.LayoutBox, id string) *runtime.LayoutBox {
	for _, box := range boxes {
		if box.NodeID == id {
			return &box
		}
	}
	return nil
}
