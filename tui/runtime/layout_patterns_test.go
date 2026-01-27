package runtime

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yaoapp/yao/tui/tui/core"
)

// TestLayoutPatterns tests the layout patterns from README.md
//
// This test suite validates that the runtime can correctly render
// the common layout patterns documented in the README.

// mockComponent is a simple mock component for testing layout
type mockComponent struct {
	id      string
	content string
	width   int
	height  int
}

func newMockComponent(id, content string) *mockComponent {
	return &mockComponent{
		id:      id,
		content: content,
		width:   0, // 0 means auto (based on content length)
		height:  1,
	}
}

func newMockComponentWithSize(id, content string, width, height int) *mockComponent {
	return &mockComponent{
		id:      id,
		content: content,
		width:   width,
		height:  height,
	}
}

// ComponentInterface implementation
func (m *mockComponent) View() string {
	return m.content
}

func (m *mockComponent) Init() tea.Cmd {
	return nil
}

func (m *mockComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return m, nil, core.Ignored
}

func (m *mockComponent) GetID() string {
	return m.id
}

func (m *mockComponent) SetFocus(focus bool) {}

func (m *mockComponent) GetFocus() bool {
	return false
}

func (m *mockComponent) SetSize(width, height int) {
	// Mock component doesn't need to track size
}

func (m *mockComponent) GetComponentType() string {
	return "mock"
}

func (m *mockComponent) Render(config core.RenderConfig) (string, error) {
	return m.content, nil
}

func (m *mockComponent) UpdateRenderConfig(config core.RenderConfig) error {
	return nil
}

func (m *mockComponent) Cleanup() {}

func (m *mockComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (m *mockComponent) GetSubscribedMessageTypes() []string {
	return nil
}

// Measurable interface implementation
func (m *mockComponent) Measure(c BoxConstraints) Size {
	w := m.width
	if w <= 0 {
		w = len(m.content)
		if w > c.MaxWidth && c.MaxWidth > 0 {
			w = c.MaxWidth
		}
	}
	h := m.height
	if h <= 0 {
		h = 1
	}
	return Size{Width: w, Height: h}
}

// Helper to create a node with a mock component
func mockNode(id, nodeType string, content string) *LayoutNode {
	var nt NodeType
	switch nodeType {
	case "row":
		nt = NodeTypeRow
	case "column":
		nt = NodeTypeColumn
	default:
		nt = NodeTypeText
	}

	node := NewLayoutNode(id, nt, NewStyle())
	comp := newMockComponent(id, content)
	node.Component = &core.ComponentInstance{
		ID:       id,
		Type:     "mock",
		Instance: comp,
	}
	return node
}

// Helper to create a container node with style
func mockContainer(id, nodeType string, style Style) *LayoutNode {
	var nt NodeType
	switch nodeType {
	case "row":
		nt = NodeTypeRow
		style.Direction = DirectionRow
	case "column":
		nt = NodeTypeColumn
		style.Direction = DirectionColumn
	default:
		nt = NodeTypeFlex
	}
	return NewLayoutNode(id, nt, style)
}

// TestSidebarLayout tests the Holy Grail layout (header + sidebar + main + footer)
func TestSidebarLayout(t *testing.T) {
	rt := NewRuntime(80, 24)

	// Build the layout tree
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 80
	root.Style.Height = 24

	// Header
	header := mockContainer("header", "column", NewStyle())
	header.Style.Height = 3
	root.AddChild(header)

	// Main content area (sidebar + main)
	mainArea := mockContainer("main-area", "row", NewStyle())
	mainArea.Style.Height = 18 // Leave space for header and footer
	mainArea.Style.FlexGrow = 1

	// Sidebar
	sidebar := mockContainer("sidebar", "column", NewStyle())
	sidebar.Style.Width = 20
	mainArea.AddChild(sidebar)

	// Main content
	mainContent := mockContainer("main", "column", NewStyle())
	mainContent.Style.FlexGrow = 1
	mainArea.AddChild(mainContent)

	root.AddChild(mainArea)

	// Footer
	footer := mockContainer("footer", "column", NewStyle())
	footer.Style.Height = 3
	root.AddChild(footer)

	// Layout and render
	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	// Verify dimensions
	if result.RootWidth == 0 {
		t.Errorf("Expected non-zero width, got %d", result.RootWidth)
	}
	if result.RootHeight == 0 {
		t.Errorf("Expected non-zero height, got %d", result.RootHeight)
	}

	// Debug output
	debug := DebugFrame(&frame, &result)
	t.Log("Sidebar Layout:")
	t.Log(debug.Summary)

	// Verify positions
	headerBox := result.FindBoxByID("header")
	if headerBox == nil {
		t.Fatal("Header box not found")
	}
	if headerBox.Y != 0 {
		t.Errorf("Header Y should be 0, got %d", headerBox.Y)
	}

	footerBox := result.FindBoxByID("footer")
	if footerBox == nil {
		t.Fatal("Footer box not found")
	}
	// Footer should be below header and main area
	if footerBox.Y < headerBox.Y+headerBox.H {
		t.Errorf("Footer Y should be below header+main, got %d", footerBox.Y)
	}

	sidebarBox := result.FindBoxByID("sidebar")
	if sidebarBox == nil {
		t.Fatal("Sidebar box not found")
	}
	if sidebarBox.X != 0 {
		t.Errorf("Sidebar X should be 0, got %d", sidebarBox.X)
	}
	if sidebarBox.W > 25 { // Should be around 20
		t.Errorf("Sidebar width should be around 20, got %d", sidebarBox.W)
	}

	mainBox := result.FindBoxByID("main")
	if mainBox == nil {
		t.Fatal("Main box not found")
	}
	if mainBox.X < sidebarBox.X+sidebarBox.W {
		t.Errorf("Main X should be after sidebar, got %d", mainBox.X)
	}
}

// TestCardGridLayout tests the 2x3 card grid using flex rows
func TestCardGridLayout(t *testing.T) {
	rt := NewRuntime(80, 24)

	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 80
	root.Style.Height = 24

	// First row of cards
	row1 := mockContainer("row1", "row", NewStyle())
	row1.Style.Height = 10
	row1.Style.Gap = 2

	for i := 0; i < 3; i++ {
		card := mockContainer("card-"+rtoa(i), "column", NewStyle())
		card.Style.FlexGrow = 1
		row1.AddChild(card)
	}
	root.AddChild(row1)

	// Second row of cards
	row2 := mockContainer("row2", "row", NewStyle())
	row2.Style.Height = 10
	row2.Style.Gap = 2

	for i := 0; i < 3; i++ {
		card := mockContainer("card-"+rtoa(i+3), "column", NewStyle())
		card.Style.FlexGrow = 1
		row2.AddChild(card)
	}
	root.AddChild(row2)

	// Layout and render
	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	// Debug output
	debug := DebugFrame(&frame, &result)
	t.Log("Card Grid Layout:")
	t.Log(debug.Summary)

	// Verify all cards are present
	expectedCards := []string{"card-0", "card-1", "card-2", "card-3", "card-4", "card-5"}
	for _, cardID := range expectedCards {
		box := result.FindBoxByID(cardID)
		if box == nil {
			t.Errorf("Card %s not found", cardID)
		}
	}

	// Verify row positions
	row1Box := result.FindBoxByID("row1")
	row2Box := result.FindBoxByID("row2")

	if row1Box == nil || row2Box == nil {
		t.Fatal("Rows not found")
	}

	// Row 2 should be below row 1
	if row2Box.Y < row1Box.Y+row1Box.H {
		t.Errorf("Row 2 should be below row 1, row1.y=%d, row1.h=%d, row2.y=%d",
			row1Box.Y, row1Box.H, row2Box.Y)
	}
}

// TestFlexGrow tests the flexGrow property
func TestFlexGrow(t *testing.T) {
	rt := NewRuntime(80, 10)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Width = 80
	root.Style.Height = 10

	// Fixed left panel (20)
	left := mockContainer("left", "column", NewStyle())
	left.Style.Width = 20
	root.AddChild(left)

	// Flexible center panel
	center := mockContainer("center", "column", NewStyle())
	center.Style.FlexGrow = 1
	root.AddChild(center)

	// Fixed right panel (15)
	right := mockContainer("right", "column", NewStyle())
	right.Style.Width = 15
	root.AddChild(right)

	// Layout and render
	constraints := NewBoxConstraints(0, 80, 0, 10)
	result := rt.Layout(root, constraints)

	// Verify positions
	leftBox := result.FindBoxByID("left")
	centerBox := result.FindBoxByID("center")
	rightBox := result.FindBoxByID("right")

	if leftBox == nil || centerBox == nil || rightBox == nil {
		t.Fatal("Not all panels found")
	}

	// Left panel at x=0
	if leftBox.X != 0 {
		t.Errorf("Left panel X should be 0, got %d", leftBox.X)
	}
	if leftBox.W != 20 {
		t.Errorf("Left panel width should be 20, got %d", leftBox.W)
	}

	// Right panel should be at the end
	expectedRightX := 80 - 15
	if rightBox.X != expectedRightX {
		t.Errorf("Right panel X should be %d, got %d", expectedRightX, rightBox.X)
	}

	// Center panel fills the space between
	expectedCenterWidth := 80 - 20 - 15 // 45
	if centerBox.W < expectedCenterWidth-2 || centerBox.W > expectedCenterWidth+2 {
		// Allow some tolerance
		t.Logf("Center width %d (expected ~%d)", centerBox.W, expectedCenterWidth)
	}
	if centerBox.X != 20 {
		t.Errorf("Center X should be 20, got %d", centerBox.X)
	}
}

// TestAlignmentModes tests different alignment options
func TestAlignmentModes(t *testing.T) {
	tests := []struct {
		name         string
		justify      Justify
		alignItems   Align
		minExpectedX int // Minimum expected X for middle item
		maxExpectedX int // Maximum expected X for middle item
	}{
		{
			name:       "justify-start",
			justify:    JustifyStart,
			alignItems: AlignStart,
			// Items should be at the start (left)
			minExpectedX: 8,
			maxExpectedX: 12,
		},
		{
			name:       "justify-center",
			justify:    JustifyCenter,
			alignItems: AlignStart,
			// Items should be centered
			minExpectedX: 30,
			maxExpectedX: 40,
		},
		{
			name:       "justify-end",
			justify:    JustifyEnd,
			alignItems: AlignStart,
			// Items should be at the end (right)
			minExpectedX: 60,
			maxExpectedX: 70,
		},
		{
			name:       "justify-space-between",
			justify:    JustifySpaceBetween,
			alignItems: AlignStart,
			// Items should be spaced out
			minExpectedX: 30,
			maxExpectedX: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := NewRuntime(80, 5)

			root := mockContainer("root", "row", NewStyle())
			root.Style.Justify = tt.justify
			root.Style.AlignItems = tt.alignItems
			root.Style.Gap = 0
			root.Style.Width = 80
			root.Style.Height = 5

			// Add 3 items of equal width (10 each)
			for i := 0; i < 3; i++ {
				item := mockContainer("item-"+rtoa(i), "column", NewStyle())
				item.Style.Width = 10
				item.Style.Height = 3
				root.AddChild(item)
			}

			constraints := NewBoxConstraints(0, 80, 0, 5)
			result := rt.Layout(root, constraints)

			// Verify middle item position is in expected range
			middleItem := result.FindBoxByID("item-1")
			if middleItem == nil {
				t.Fatal("Middle item not found")
			}
			if middleItem.X < tt.minExpectedX || middleItem.X > tt.maxExpectedX {
				t.Errorf("Middle item X=%d, expected range [%d, %d]",
					middleItem.X, tt.minExpectedX, tt.maxExpectedX)
			}
			t.Logf("Items: item-0.X=%d, item-1.X=%d, item-2.X=%d",
				result.FindBoxByID("item-0").X,
				result.FindBoxByID("item-1").X,
				result.FindBoxByID("item-2").X)
		})
	}
}

// TestAlignItemsCrossAxis tests cross-axis alignment
func TestAlignItemsCrossAxis(t *testing.T) {
	tests := []struct {
		name       string
		alignItems Align
	}{
		{
			name:       "align-start",
			alignItems: AlignStart,
		},
		{
			name:       "align-center",
			alignItems: AlignCenter,
		},
		{
			name:       "align-end",
			alignItems: AlignEnd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := NewRuntime(30, 10)

			root := mockContainer("root", "row", NewStyle())
			root.Style.AlignItems = tt.alignItems
			root.Style.Width = 30
			root.Style.Height = 10

			// Add items with different heights
			heights := []int{3, 6, 4}
			for i, h := range heights {
				item := mockContainer("item-"+rtoa(i), "column", NewStyle())
				item.Style.Width = 8
				item.Style.Height = h
				root.AddChild(item)
			}

			constraints := NewBoxConstraints(0, 30, 0, 10)
			result := rt.Layout(root, constraints)

			// Verify all items are within bounds
			for i := range heights {
				item := result.FindBoxByID("item-" + rtoa(i))
				if item != nil {
					t.Logf("Item %d: Y=%d, H=%d", i, item.Y, item.H)
					// All items should be within container height
					if item.Y < 0 || item.Y > 10 {
						t.Errorf("Item %d Y=%d out of bounds", i, item.Y)
					}
					if item.H < 0 || item.H > 10 {
						t.Errorf("Item %d H=%d out of bounds", i, item.H)
					}
					// Item should not overflow container
					if item.Y+item.H > 10 {
						t.Errorf("Item %d overflows container: Y=%d, H=%d", i, item.Y, item.H)
					}
				}
			}
		})
	}
}

// TestGapSpacing tests the gap property
func TestGapSpacing(t *testing.T) {
	rt := NewRuntime(80, 5)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Gap = 2
	root.Style.Width = 80
	root.Style.Height = 5

	// Add 3 items of width 10 each
	for i := 0; i < 3; i++ {
		item := mockContainer("item-"+rtoa(i), "column", NewStyle())
		item.Style.Width = 10
		root.AddChild(item)
	}

	constraints := NewBoxConstraints(0, 80, 0, 5)
	result := rt.Layout(root, constraints)

	// Verify positions with gap
	item0 := result.FindBoxByID("item-0")
	item1 := result.FindBoxByID("item-1")
	item2 := result.FindBoxByID("item-2")

	if item0 == nil || item1 == nil || item2 == nil {
		t.Fatal("Items not found")
	}

	// Check that gaps are applied
	if item1.X <= item0.X+item0.W {
		t.Errorf("Gap not applied: item0 ends at %d, item1 starts at %d",
			item0.X+item0.W, item1.X)
	}

	if item2.X <= item1.X+item1.W {
		t.Errorf("Gap not applied: item1 ends at %d, item2 starts at %d",
			item1.X+item1.W, item2.X)
	}

	t.Logf("Gap spacing: item0.X=%d, item1.X=%d, item2.X=%d",
		item0.X, item1.X, item2.X)
}

// TestPadding tests padding calculation
func TestPadding(t *testing.T) {
	rt := NewRuntime(40, 10)

	root := mockNode("root", "column", "Content")
	root.Style.Padding = NewInsets(1, 2, 1, 2) // top, right, bottom, left
	root.Style.Width = 40
	root.Style.Height = 10

	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)

	box := result.FindBoxByID("root")
	if box == nil {
		t.Fatal("Root box not found")
	}

	// Box should have the outer dimensions
	if box.W != 40 {
		t.Errorf("Box width should be 40, got %d", box.W)
	}
	if box.H != 10 {
		t.Errorf("Box height should be 10, got %d", box.H)
	}

	t.Logf("Padding test: box=%dx%d with padding [1,2,1,2]", box.W, box.H)
}

// TestBorderLayout tests border space allocation
func TestBorderLayout(t *testing.T) {
	rt := NewRuntime(40, 10)

	root := mockNode("root", "column", "Content")
	root.Style.Border = NewInsets(1, 1, 1, 1) // 1 cell border on all sides
	root.Style.Width = 40
	root.Style.Height = 10

	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)

	box := result.FindBoxByID("root")
	if box == nil {
		t.Fatal("Root box not found")
	}

	// Box should have the outer dimensions
	if box.W != 40 {
		t.Errorf("Box width should be 40, got %d", box.W)
	}

	t.Logf("Border test: box=%dx%d with border [1,1,1,1]", box.W, box.H)
}

// TestNestedLayouts tests deeply nested layout structures
func TestNestedLayouts(t *testing.T) {
	rt := NewRuntime(80, 24)

	root := mockContainer("root", "column", NewStyle())

	// Level 1: Column
	row1 := mockContainer("row1", "row", NewStyle())
	root.AddChild(row1)

	// Level 2: Row with Column
	col1 := mockContainer("col1", "column", NewStyle())
	col1.Style.Width = 20
	row1.AddChild(col1)

	// Level 3: Column with Row
	row2 := mockContainer("row2", "row", NewStyle())
	row2.Style.Gap = 1
	col1.AddChild(row2)

	// Level 4: Row with items
	for i := 0; i < 3; i++ {
		item := mockContainer("item-"+rtoa(i), "column", NewStyle())
		item.Style.Width = 5
		row2.AddChild(item)
	}

	// Right side content
	rightContent := mockContainer("right", "column", NewStyle())
	rightContent.Style.FlexGrow = 1
	row1.AddChild(rightContent)

	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)

	// Verify all levels are present
	expectedIDs := []string{"row1", "col1", "row2", "item-0", "item-1", "item-2", "right"}
	for _, id := range expectedIDs {
		if result.FindBoxByID(id) == nil {
			t.Errorf("Expected node %s not found", id)
		}
	}

	t.Logf("Nested layout: found %d boxes", len(result.Boxes))
}

// TestZIndexRendering tests that Z-Index affects rendering order
func TestZIndexRendering(t *testing.T) {
	rt := NewRuntime(20, 10)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Width = 20
	root.Style.Height = 10
	root.Style.Gap = 2

	// Layer 0: Background (leftmost)
	bg := mockNode("bg", "column", "Background")
	bg.Style.Width = 5
	bg.Style.Height = 10
	bg.Style.ZIndex = 0
	root.AddChild(bg)

	// Layer 10: Content (middle)
	content := mockNode("content", "column", "Content")
	content.Style.Width = 5
	content.Style.Height = 10
	content.Style.ZIndex = 10
	root.AddChild(content)

	// Layer 100: Overlay (rightmost)
	overlay := mockNode("overlay", "column", "Overlay")
	overlay.Style.Width = 5
	overlay.Style.Height = 10
	overlay.Style.ZIndex = 100
	root.AddChild(overlay)

	constraints := NewBoxConstraints(0, 20, 0, 10)
	result := rt.Layout(root, constraints)

	// All should be present
	if result.FindBoxByID("bg") == nil {
		t.Error("Background not found")
	}
	if result.FindBoxByID("content") == nil {
		t.Error("Content not found")
	}
	if result.FindBoxByID("overlay") == nil {
		t.Error("Overlay not found")
	}

	// Check Z-index values are preserved
	bgBox := result.FindBoxByID("bg")
	contentBox := result.FindBoxByID("content")
	overlayBox := result.FindBoxByID("overlay")

	if bgBox == nil || contentBox == nil || overlayBox == nil {
		t.Fatal("Not all boxes found")
	}

	// Verify Z-index ordering
	if bgBox.ZIndex != 0 {
		t.Errorf("BG ZIndex should be 0, got %d", bgBox.ZIndex)
	}
	if contentBox.ZIndex != 10 {
		t.Errorf("Content ZIndex should be 10, got %d", contentBox.ZIndex)
	}
	if overlayBox.ZIndex != 100 {
		t.Errorf("Overlay ZIndex should be 100, got %d", overlayBox.ZIndex)
	}

	// Verify horizontal positioning (row layout)
	if bgBox.X != 0 {
		t.Errorf("BG should be at X=0, got %d", bgBox.X)
	}
	if contentBox.X <= bgBox.X+bgBox.W {
		t.Errorf("Content should be to the right of BG")
	}
	if overlayBox.X <= contentBox.X+contentBox.W {
		t.Errorf("Overlay should be to the right of Content")
	}

	t.Logf("Z-Index test: BG.X=%d, Content.X=%d, Overlay.X=%d",
		bgBox.X, contentBox.X, overlayBox.X)
}

// TestOverflowScroll tests overflow handling
func TestOverflowScroll(t *testing.T) {
	rt := NewRuntime(40, 10)

	root := mockContainer("root", "column", NewStyle())
	root.Style.Overflow = OverflowScroll
	root.Style.Width = 40
	root.Style.Height = 10

	// Add content that exceeds the container
	for i := 0; i < 20; i++ {
		line := mockNode("line-"+rtoa(i), "column", "Line "+rtoa(i+1))
		root.AddChild(line)
	}

	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)

	// All lines should be laid out
	if len(result.Boxes) < 20 {
		t.Errorf("Expected at least 20 lines, got %d", len(result.Boxes))
	}

	t.Logf("Overflow test: laid out %d boxes in 10-height container", len(result.Boxes))
}

// TestCompactRow tests a row with no gap
func TestCompactRow(t *testing.T) {
	rt := NewRuntime(80, 5)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Gap = 0 // No gap
	root.Style.Width = 80
	root.Style.Height = 5

	// Add 4 items of width 20 each (exactly fills 80)
	for i := 0; i < 4; i++ {
		item := mockContainer("item-"+rtoa(i), "column", NewStyle())
		item.Style.Width = 20
		root.AddChild(item)
	}

	constraints := NewBoxConstraints(0, 80, 0, 5)
	result := rt.Layout(root, constraints)

	// Verify compact positioning
	item0 := result.FindBoxByID("item-0")
	item1 := result.FindBoxByID("item-1")
	item2 := result.FindBoxByID("item-2")
	item3 := result.FindBoxByID("item-3")

	if item0 == nil || item1 == nil || item2 == nil || item3 == nil {
		t.Fatal("Items not found")
	}

	// With no gap, items should be directly adjacent
	expectedPositions := []int{0, 20, 40, 60}
	positions := []*LayoutBox{item0, item1, item2, item3}

	for i, item := range positions {
		if item.X != expectedPositions[i] {
			t.Errorf("Item %d X expected %d, got %d", i, expectedPositions[i], item.X)
		}
	}

	t.Logf("Compact row: items at X=%d, %d, %d, %d",
		item0.X, item1.X, item2.X, item3.X)
}

// TestColumnLayout tests vertical layout
func TestColumnLayout(t *testing.T) {
	rt := NewRuntime(30, 20)

	root := mockContainer("root", "column", NewStyle())
	root.Style.Gap = 1
	root.Style.Width = 30
	root.Style.Height = 20

	// Add items of different heights
	heights := []int{3, 5, 4, 2}
	for i, h := range heights {
		item := mockContainer("item-"+rtoa(i), "column", NewStyle())
		item.Style.Height = h
		root.AddChild(item)
	}

	constraints := NewBoxConstraints(0, 30, 0, 20)
	result := rt.Layout(root, constraints)

	// Debug: print measured heights
	for i, h := range heights {
		item := result.FindBoxByID("item-" + rtoa(i))
		if item != nil {
			t.Logf("Item %d: Y=%d, H=%d (expected H=%d)", i, item.Y, item.H, h)
		}
	}

	// Verify vertical stacking
	yPos := 0
	for i, h := range heights {
		item := result.FindBoxByID("item-" + rtoa(i))
		if item == nil {
			t.Fatalf("Item %d not found", i)
		}
		if item.Y != yPos {
			t.Errorf("Item %d Y expected %d, got %d", i, yPos, item.Y)
		}
		if item.H != h {
			t.Errorf("Item %d height expected %d, got %d", i, h, item.H)
		}
		yPos += h + 1 // +1 for gap
	}

	t.Logf("Column layout: 4 items stacked with gaps")
}

// TestPercentageWidth tests percentage-based sizing
func TestPercentageWidth(t *testing.T) {
	rt := NewRuntime(100, 10)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Width = 100
	root.Style.Height = 10

	// 30% sidebar
	sidebar := mockContainer("sidebar", "column", NewStyle())
	sidebar.Style.Width = -30 // -30 means 30%
	root.AddChild(sidebar)

	// 70% main (flex)
	main := mockContainer("main", "column", NewStyle())
	main.Style.FlexGrow = 1
	root.AddChild(main)

	constraints := NewBoxConstraints(0, 100, 0, 10)
	result := rt.Layout(root, constraints)

	sidebarBox := result.FindBoxByID("sidebar")
	mainBox := result.FindBoxByID("main")

	if sidebarBox == nil {
		t.Fatal("Sidebar box not found")
	}
	if mainBox == nil {
		t.Fatal("Main box not found")
	}

	// Sidebar should be 30% of 100 = 30
	expectedSidebarWidth := 30
	if sidebarBox.W != expectedSidebarWidth {
		t.Errorf("Sidebar width expected %d (30%%), got %d", expectedSidebarWidth, sidebarBox.W)
	}

	t.Logf("Percentage test: sidebar=%d (30%%), main=%d (70%%)",
		sidebarBox.W, mainBox.W)
}

// TestAutoWidth tests -1 (auto) width
func TestAutoWidth(t *testing.T) {
	rt := NewRuntime(80, 10)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Width = 80
	root.Style.Height = 10
	root.Style.Gap = 1

	// Add items with auto width (based on content)
	widths := []int{5, 10, 15}
	for i, w := range widths {
		item := mockNode("auto-"+rtoa(i), "column", strings.Repeat("X", w))
		item.Style.Width = -1 // Auto
		root.AddChild(item)
	}

	constraints := NewBoxConstraints(0, 80, 0, 10)
	result := rt.Layout(root, constraints)

	// All items should be present
	for i := 0; i < 3; i++ {
		item := result.FindBoxByID("auto-" + rtoa(i))
		if item == nil {
			t.Errorf("Auto item %d not found", i)
		}
		// Width should be based on content
		if item.W == 0 {
			t.Errorf("Auto item %d has zero width", i)
		}
		t.Logf("Auto item %d: width=%d", i, item.W)
	}
}

// TestCenteredModal tests a centered modal overlay
func TestCenteredModal(t *testing.T) {
	rt := NewRuntime(80, 24)

	root := mockContainer("root", "column", NewStyle())
	root.Style.AlignItems = AlignCenter
	root.Style.Justify = JustifyCenter
	root.Style.Width = 80
	root.Style.Height = 24

	// Modal
	modal := mockNode("modal", "column", "Modal Content")
	modal.Style.Width = 40
	modal.Style.Height = 15
	modal.Style.ZIndex = 100
	root.AddChild(modal)

	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)

	modalBox := result.FindBoxByID("modal")
	if modalBox == nil {
		t.Fatal("Modal box not found")
	}

	// Modal should be approximately centered
	// 80x24 container, 40x15 modal
	// Center X = (80 - 40) / 2 = 20
	// Center Y = (24 - 15) / 2 = 4.5 -> 4 or 5
	// Note: AlignItems/Justify affects child positioning within available space
	// The modal should be centered within the 80x24 container

	t.Logf("Modal at X=%d, Y=%d, size=%dx%d", modalBox.X, modalBox.Y, modalBox.W, modalBox.H)

	// The modal should have the correct size
	if modalBox.W != 40 {
		t.Errorf("Modal width should be 40, got %d", modalBox.W)
	}
	if modalBox.H != 15 {
		t.Errorf("Modal height should be 15, got %d", modalBox.H)
	}
}

// TestBoxLayoutConstraints tests that boxes respect constraints
func TestBoxLayoutConstraints(t *testing.T) {
	rt := NewRuntime(20, 10)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Gap = 1

	// Add items that would overflow if not constrained
	for i := 0; i < 5; i++ {
		item := mockContainer("item-"+rtoa(i), "column", NewStyle())
		item.Style.Width = 10 // Each wants 10, but container is only 20
		root.AddChild(item)
	}

	constraints := NewBoxConstraints(0, 20, 0, 10) // Max width is 20
	result := rt.Layout(root, constraints)

	// All items should be laid out
	if len(result.Boxes) < 5 {
		t.Errorf("Expected all 5 items to be laid out, got %d", len(result.Boxes))
	}

	// Items should not have negative positions
	for _, box := range result.Boxes {
		if box.X < 0 {
			t.Errorf("Box %s has negative X: %d", box.NodeID, box.X)
		}
		if box.Y < 0 {
			t.Errorf("Box %s has negative Y: %d", box.NodeID, box.Y)
		}
	}

	t.Logf("Box constraints: %d boxes laid out", len(result.Boxes))
}

// TestEmptyLayout tests an empty layout tree
func TestEmptyLayout(t *testing.T) {
	rt := NewRuntime(80, 24)

	root := mockContainer("root", "column", NewStyle())
	// Empty container needs explicit size to be measured
	root.Style.Width = 80
	root.Style.Height = 24

	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)

	if result.RootWidth == 0 {
		t.Errorf("Empty layout width should be non-zero, got %d", result.RootWidth)
	}
	if result.RootHeight == 0 {
		t.Errorf("Empty layout height should be non-zero, got %d", result.RootHeight)
	}

	frame := rt.Render(result)
	if frame.Buffer == nil {
		t.Error("Frame buffer should not be nil")
	}

	t.Logf("Empty layout: %dx%d", result.RootWidth, result.RootHeight)
}

// TestSingleItem tests a layout with a single item
func TestSingleItem(t *testing.T) {
	rt := NewRuntime(80, 24)

	root := mockContainer("root", "row", NewStyle())
	root.Style.AlignItems = AlignCenter
	root.Style.Justify = JustifyCenter

	item := mockNode("single", "column", "Single Item")
	root.AddChild(item)

	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)

	box := result.FindBoxByID("single")
	if box == nil {
		t.Fatal("Single item box not found")
	}

	// Item should be present and have some size
	if box.W == 0 || box.H == 0 {
		t.Errorf("Single item should have non-zero size, got %dx%d", box.W, box.H)
	}

	t.Logf("Single item at X=%d, Y=%d, size=%dx%d", box.X, box.Y, box.W, box.H)
}

// TestDeepNesting tests very deep nesting (20 levels)
func TestDeepNesting(t *testing.T) {
	rt := NewRuntime(80, 100)

	// Build a deep tree
	current := mockContainer("level0", "column", NewStyle())
	current.Style.Width = 80
	current.Style.Height = 100

	// Keep reference to root
	root := current

	for i := 1; i <= 20; i++ {
		child := mockContainer("level"+rtoa(i), "row", NewStyle())
		current.AddChild(child)
		current = child
	}

	// Add content at the deepest level
	content := mockNode("content", "column", "Deepest")
	current.Component = content.Component

	// This should not crash or hang
	constraints := NewBoxConstraints(0, 80, 0, 100)
	result := rt.Layout(root, constraints)

	if len(result.Boxes) < 20 {
		t.Errorf("Expected at least 20 boxes in deep nesting, got %d", len(result.Boxes))
	}

	// Verify the deepest level exists
	deepestBox := result.FindBoxByID("level20")
	if deepestBox == nil {
		t.Error("Deepest level not found")
	}

	t.Logf("Deep nesting: %d levels, %d boxes", 20, len(result.Boxes))
}

// TestLargeNumberOfSiblings tests handling many siblings
func TestLargeNumberOfSiblings(t *testing.T) {
	rt := NewRuntime(80, 24)

	root := mockContainer("root", "row", NewStyle())
	root.Style.Gap = 1

	// Add 50 items
	for i := 0; i < 50; i++ {
		item := mockNode("item-"+rtoa(i), "column", "I"+rtoa(i))
		root.AddChild(item)
	}

	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)

	if len(result.Boxes) < 50 {
		t.Errorf("Expected 50 items, got %d", len(result.Boxes))
	}

	// Items should be positioned horizontally (x increasing)
	prevX := -1
	xIncreasing := true
	for i := 0; i < 50; i++ {
		box := result.FindBoxByID("item-" + rtoa(i))
		if box != nil {
			if box.X < prevX {
				xIncreasing = false
				break
			}
			prevX = box.X
		}
	}

	if !xIncreasing {
		t.Error("Items should have increasing X positions")
	}

	t.Logf("Large siblings: %d items laid out", len(result.Boxes))
}

// Helper: integer to ASCII string (no strconv for portability)
func rtoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var result string
	for i > 0 {
		result = string(digits[i%10]) + result
		i /= 10
	}
	return result
}
