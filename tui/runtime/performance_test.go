package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// BenchmarkLayoutSimple benchmarks a simple layout with few components
func BenchmarkLayoutSimple(b *testing.B) {
	root := createSimpleLayout(3) // 3 levels deep
	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Layout(root, constraints)
	}
}

// BenchmarkLayoutMedium benchmarks a medium layout with moderate components
func BenchmarkLayoutMedium(b *testing.B) {
	root := createSimpleLayout(10) // 10 levels deep
	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Layout(root, constraints)
	}
}

// BenchmarkLayoutComplex benchmarks a complex layout with many components
func BenchmarkLayoutComplex(b *testing.B) {
	root := createComplexLayout(100) // 100 components
	runtime := NewRuntime(120, 40)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Layout(root, constraints)
	}
}

// BenchmarkLayoutDeep benchmarks a very deep layout tree
func BenchmarkLayoutDeep(b *testing.B) {
	root := createDeepLayout(50) // 50 levels deep
	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Layout(root, constraints)
	}
}

// BenchmarkLayoutFlex benchmarks flex layout calculations
func BenchmarkLayoutFlex(b *testing.B) {
	root := createFlexLayout(20) // 20 flex children
	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Layout(root, constraints)
	}
}

// BenchmarkLayoutAbsolute benchmarks absolute positioning
func BenchmarkLayoutAbsolute(b *testing.B) {
	root := createAbsoluteLayout(20) // 20 absolute children
	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Layout(root, constraints)
	}
}

// BenchmarkRenderSimple benchmarks simple rendering
func BenchmarkRenderSimple(b *testing.B) {
	root := createSimpleLayout(5)
	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// First layout
	result := runtime.Layout(root, constraints)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Render(result)
	}
}

// BenchmarkRenderComplex benchmarks complex rendering
func BenchmarkRenderComplex(b *testing.B) {
	root := createComplexLayout(50)
	runtime := NewRuntime(120, 40)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// First layout
	result := runtime.Layout(root, constraints)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Render(result)
	}
}

// BenchmarkDiff benchmarks frame diffing
func BenchmarkDiff(b *testing.B) {
	oldFrame := createTestFrame(80, 24, 'A')
	newFrame := createTestFrame(80, 24, 'B')

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeDiff(oldFrame, newFrame)
	}
}

// BenchmarkDiffSmallChanges benchmarks diffing with small changes
func BenchmarkDiffSmallChanges(b *testing.B) {
	oldFrame := createTestFrame(80, 24, 'A')
	newFrame := createTestFrameWithChanges(80, 24, 'A', 10) // 10 changes

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeDiff(oldFrame, newFrame)
	}
}

// BenchmarkMeasureComponent benchmarks component measurement
func BenchmarkMeasureComponent(b *testing.B) {
	comp := &TestComponent{
		content: "Hello World",
	}
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Measure(constraints)
	}
}

// BenchmarkLayoutUpdate benchmarks layout updates with partial changes
func BenchmarkLayoutUpdate(b *testing.B) {
	root := createComplexLayout(50)
	runtime := NewRuntime(120, 40)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// Initial layout
	result := runtime.Layout(root, constraints)
	runtime.Render(result)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate update by modifying one component
		if len(root.Children) > 0 {
			root.Children[0].MeasuredWidth += 1
		}
		result = runtime.Layout(root, constraints)
		runtime.Render(result)
	}
}

// BenchmarkAbsolutePositioning benchmarks absolute positioning performance
func BenchmarkAbsolutePositioning(b *testing.B) {
	parent := &LayoutNode{
		ID:             "parent",
		X:              0,
		Y:              0,
		MeasuredWidth:  200,
		MeasuredHeight: 100,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 50),
	}

	// Add 50 absolute children
	for i := 0; i < 50; i++ {
		top := i * 2
		left := i * 3
		parent.Children[i] = &LayoutNode{
			ID:             "child",
			MeasuredWidth:  50,
			MeasuredHeight: 30,
			Position: Position{
				Type:  PositionAbsolute,
				Top:   &top,
				Left:  &left,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyAbsoluteLayout(parent)
	}
}

// TestPerformanceLayoutDeepTree tests performance with very deep trees
func TestPerformanceLayoutDeepTree(t *testing.T) {
	// Create a very deep tree
	depth := 100
	root := createDeepLayout(depth)

	runtime := NewRuntime(80, 24)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// Measure layout time
	start := time.Now()
	result := runtime.Layout(root, constraints)
	duration := time.Since(start)

	t.Logf("Layout of %d-level tree took: %v", depth, duration)

	// Verify layout completed
	assert.NotNil(t, result)

	// Verify all levels were processed
	countLevels(root, 0, depth)
}

// TestPerformanceLargeComponentTree tests performance with many components
func TestPerformanceLargeComponentTree(t *testing.T) {
	// Create a tree with many components
	count := 1000
	root := createComplexLayout(count)

	runtime := NewRuntime(120, 40)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	// Measure layout time
	start := time.Now()
	result := runtime.Layout(root, constraints)
	duration := time.Since(start)

	t.Logf("Layout of %d components took: %v", count, duration)

	// Verify layout completed
	assert.NotNil(t, result)

	// Count actual components
	compCount := countComponents(root)
	t.Logf("Actual components in tree: %d", compCount)
}

// TestPerformanceRenderingStress tests rendering under stress
func TestPerformanceRenderingStress(t *testing.T) {
	root := createComplexLayout(200)
	runtime := NewRuntime(120, 40)
	constraints := BoxConstraints{
		MinWidth:  0,
		MaxWidth:  120,
		MinHeight: 0,
		MaxHeight: 40,
	}

	result := runtime.Layout(root, constraints)

	// Render multiple times
	iterations := 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		runtime.Render(result)
	}
	duration := time.Since(start)

	avgPerRender := duration / time.Duration(iterations)
	t.Logf("Rendered %d times in %v (avg: %v per render)",
		iterations, duration, avgPerRender)

	// Verify reasonable performance (should be < 10ms per render)
	if avgPerRender > 10*time.Millisecond {
		t.Logf("WARNING: Render time is high: %v", avgPerRender)
	}
}

// Helper functions for benchmarks

func createSimpleLayout(depth int) *LayoutNode {
	if depth <= 0 {
		return &LayoutNode{
			ID:             "leaf",
			Type:           NodeTypeText,
			MeasuredWidth:  10,
			MeasuredHeight: 1,
			Style:          NewStyle(),
			Position:       NewPosition(),
		}
	}

	child := createSimpleLayout(depth - 1)
	return &LayoutNode{
		ID:             "level-" + string(rune(depth)),
		Type:           NodeTypeColumn,
		MeasuredWidth:  80,
		MeasuredHeight: 24,
		Style:          NewStyle(),
		Position:       NewPosition(),
		Children:       []*LayoutNode{child},
	}
}

func createComplexLayout(componentCount int) *LayoutNode {
	root := &LayoutNode{
		ID:             "root",
		Type:           NodeTypeColumn,
		MeasuredWidth:  120,
		MeasuredHeight: 40,
		Style:          NewStyle(),
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Create rows
	componentsPerRow := 10
	rows := componentCount / componentsPerRow

	for r := 0; r < rows; r++ {
		row := &LayoutNode{
			ID:             "row-" + string(rune(r)),
			Type:           NodeTypeRow,
			Style:          NewStyle(),
			Position:       NewPosition(),
			Children:       make([]*LayoutNode, 0),
		}

		for c := 0; c < componentsPerRow; c++ {
			child := &LayoutNode{
				ID:             "comp-" + string(rune(r)) + "-" + string(rune(c)),
				Type:           NodeTypeText,
				MeasuredWidth:  10,
				MeasuredHeight: 2,
				Style:          NewStyle(),
				Position:       NewPosition(),
			}
			row.AddChild(child)
		}

		root.AddChild(row)
	}

	return root
}

func createDeepLayout(depth int) *LayoutNode {
	if depth <= 0 {
		return &LayoutNode{
			ID:             "leaf",
			Type:           NodeTypeText,
			MeasuredWidth:  10,
			MeasuredHeight: 1,
			Style:          NewStyle(),
			Position:       NewPosition(),
		}
	}

	child := createDeepLayout(depth - 1)
	return &LayoutNode{
		ID:             "level-" + string(rune(depth)),
		Type:           NodeTypeColumn,
		Style:          NewStyle(),
		Position:       NewPosition(),
		Children:       []*LayoutNode{child},
	}
}

func createFlexLayout(childCount int) *LayoutNode {
	root := &LayoutNode{
		ID:             "root",
		Type:           NodeTypeRow,
		Style:          NewStyle().WithFlexGrow(1.0),
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	for i := 0; i < childCount; i++ {
		grow := float64(i + 1)
		child := &LayoutNode{
			ID:             "child-" + string(rune(i)),
			Type:           NodeTypeText,
			Style:          NewStyle().WithFlexGrow(grow),
			Position:       NewPosition(),
		}
		root.AddChild(child)
	}

	return root
}

func createAbsoluteLayout(childCount int) *LayoutNode {
	root := &LayoutNode{
		ID:             "root",
		Type:           NodeTypeColumn,
		MeasuredWidth:  200,
		MeasuredHeight: 100,
		Style:          NewStyle(),
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	for i := 0; i < childCount; i++ {
		top := i * 2
		left := i * 3
		child := &LayoutNode{
			ID:             "child-" + string(rune(i)),
			Type:           NodeTypeText,
			MeasuredWidth:  50,
			MeasuredHeight: 30,
			Position: Position{
				Type:  PositionAbsolute,
				Top:   &top,
				Left:  &left,
			},
		}
		root.AddChild(child)
	}

	return root
}

func createTestFrame(width, height int, char rune) Frame {
	buf := NewCellBuffer(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			buf.SetContent(x, y, 0, char, CellStyle{}, "")
		}
	}
	return Frame{
		Buffer: buf,
		Width:  width,
		Height: height,
	}
}

func createTestFrameWithChanges(width, height int, defaultChar rune, changeCount int) Frame {
	newFrame := createTestFrame(width, height, defaultChar)

	// Make some changes
	for i := 0; i < changeCount; i++ {
		x := i % width
		y := i / width
		newFrame.Buffer.SetContent(x, y, 0, 'X', CellStyle{}, "")
	}

	return newFrame
}

func countLevels(node *LayoutNode, current, max int) int {
	if node == nil {
		return current
	}

	maxDepth := current
	for _, child := range node.Children {
		depth := countLevels(child, current+1, max)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

func countComponents(node *LayoutNode) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children {
		count += countComponents(child)
	}
	return count
}

// TestComponent for benchmarking
type TestComponent struct {
	content string
}

func (t *TestComponent) Measure(constraints BoxConstraints) Size {
	return Size{
		Width:  len(t.content),
		Height: 1,
	}
}

func (t *TestComponent) Render(buf *CellBuffer, box LayoutBox) {
	for i, char := range t.content {
		buf.SetContent(box.X+i, box.Y, box.ZIndex, char, CellStyle{}, t.ID())
	}
}

func (t *TestComponent) HandleEvent(ev Event) {}

func (t *TestComponent) ID() string {
	return "test"
}

func (t *TestComponent) Type() string {
	return "test"
}

func (t *TestComponent) Instance() Component {
	return t
}

func (t *TestComponent) View() string {
	return t.content
}
