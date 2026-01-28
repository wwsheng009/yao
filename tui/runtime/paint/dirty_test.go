package paint

import (
	"testing"

	"github.com/yaoapp/yao/tui/runtime/style"
)

// ==============================================================================
// 基础功能测试
// ==============================================================================

func TestNewDirtyTracker(t *testing.T) {
	tracker := NewDirtyTracker()
	if tracker == nil {
		t.Fatal("NewDirtyTracker returned nil")
	}
	if !tracker.IsAllDirty() {
		t.Error("New tracker should be all dirty initially")
	}
}

func TestMarkCell(t *testing.T) {
	tracker := NewDirtyTracker()
	tracker.MarkAll() // 先清空全脏标记
	tracker.Clear()

	tracker.MarkCell(5, 10)
	if !tracker.IsDirtyCell(5, 10) {
		t.Error("Cell (5, 10) should be dirty")
	}
	if tracker.IsDirtyCell(0, 0) {
		t.Error("Cell (0, 0) should not be dirty")
	}
}

func TestMarkRect(t *testing.T) {
	tracker := NewDirtyTracker()
	tracker.MarkAll()
	tracker.Clear()

	rect := Rect{X: 10, Y: 10, Width: 5, Height: 3}
	tracker.MarkRect(rect)

	// 检查矩形内的单元格
	if !tracker.IsDirtyCell(10, 10) {
		t.Error("Cell (10, 10) should be dirty")
	}
	if !tracker.IsDirtyCell(14, 12) {
		t.Error("Cell (14, 12) should be dirty")
	}
	// 检查矩形外的单元格
	if tracker.IsDirtyCell(9, 10) {
		t.Error("Cell (9, 10) should not be dirty")
	}
	if tracker.IsDirtyCell(10, 13) {
		t.Error("Cell (10, 13) should not be dirty")
	}
}

func TestMarkAll(t *testing.T) {
	tracker := NewDirtyTracker()
	tracker.Clear()
	tracker.MarkCell(1, 1)

	tracker.MarkAll()
	if !tracker.IsAllDirty() {
		t.Error("Should be all dirty after MarkAll")
	}
	// 全脏状态下，任何单元格都应该返回脏
	if !tracker.IsDirtyCell(999, 999) {
		t.Error("All cells should be dirty when allDirty is true")
	}
}

func TestClear(t *testing.T) {
	tracker := NewDirtyTracker()
	tracker.MarkCell(1, 1)
	tracker.MarkAll()
	tracker.Clear()

	if tracker.IsAllDirty() {
		t.Error("Should not be all dirty after Clear")
	}
	if tracker.IsDirtyCell(1, 1) {
		t.Error("Cell (1, 1) should not be dirty after Clear")
	}
	if len(tracker.GetDirtyCells()) != 0 {
		t.Error("No dirty cells after Clear")
	}
}

func TestGetDirtyRects(t *testing.T) {
	tracker := NewDirtyTracker()
	tracker.Clear()

	// 无脏区域时返回空列表
	rects := tracker.GetDirtyRects()
	if len(rects) != 0 {
		t.Errorf("Expected 0 rects, got %d", len(rects))
	}

	// 标记一些区域
	tracker.MarkRect(Rect{X: 0, Y: 0, Width: 10, Height: 10})
	tracker.MarkRect(Rect{X: 20, Y: 20, Width: 5, Height: 5})

	rects = tracker.GetDirtyRects()
	if len(rects) != 2 {
		t.Errorf("Expected 2 rects, got %d", len(rects))
	}
}

// ==============================================================================
// Diff 功能测试
// ==============================================================================

func TestDiff_NilBuffers(t *testing.T) {
	tracker := NewDirtyTracker()
	result := tracker.Diff(nil, nil)

	if result.HasChanges {
		t.Error("No changes when both buffers are nil")
	}
	if len(result.DirtyRegions) != 0 {
		t.Error("No dirty regions when both buffers are nil")
	}
}

func TestDiff_FirstBuffer(t *testing.T) {
	tracker := NewDirtyTracker()
	buf := NewBuffer(10, 5)
	baseStyle := style.Style{}

	// 填充一些内容
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			buf.SetCell(x, y, ' ', baseStyle)
		}
	}

	result := tracker.Diff(nil, buf)

	if !result.HasChanges {
		t.Error("Should have changes when first buffer is nil")
	}
	if len(result.DirtyRegions) != 1 {
		t.Errorf("Expected 1 dirty region, got %d", len(result.DirtyRegions))
	}
	if result.DirtyRegions[0].Width != 10 || result.DirtyRegions[0].Height != 5 {
		t.Error("Dirty region should cover entire buffer")
	}
}

func TestDiff_SameContent(t *testing.T) {
	tracker := NewDirtyTracker()

	prev := NewBuffer(10, 5)
	curr := NewBuffer(10, 5)
	baseStyle := style.Style{}

	// 填充相同内容
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			prev.SetCell(x, y, ' ', baseStyle)
			curr.SetCell(x, y, ' ', baseStyle)
		}
	}

	result := tracker.Diff(prev, curr)

	if result.HasChanges {
		t.Error("No changes when buffers have same content")
	}
	if result.ChangedCells != 0 {
		t.Errorf("Expected 0 changed cells, got %d", result.ChangedCells)
	}
}

func TestDiff_SingleCellChange(t *testing.T) {
	tracker := NewDirtyTracker()

	prev := NewBuffer(10, 5)
	curr := NewBuffer(10, 5)
	baseStyle := style.Style{}

	// 填充内容
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			prev.SetCell(x, y, ' ', baseStyle)
			curr.SetCell(x, y, ' ', baseStyle)
		}
	}

	// 修改一个单元格
	curr.SetCell(5, 3, 'X', baseStyle)

	result := tracker.Diff(prev, curr)

	if !result.HasChanges {
		t.Error("Should detect single cell change")
	}
	if result.ChangedCells != 1 {
		t.Errorf("Expected 1 changed cell, got %d", result.ChangedCells)
	}
	if len(result.DirtyRegions) != 1 {
		t.Errorf("Expected 1 dirty region, got %d", len(result.DirtyRegions))
	}
	// 验证区域是 1x1
	if result.DirtyRegions[0].Width != 1 || result.DirtyRegions[0].Height != 1 {
		t.Errorf("Expected 1x1 region, got %dx%d",
			result.DirtyRegions[0].Width, result.DirtyRegions[0].Height)
	}
}

func TestDiff_MultipleChanges(t *testing.T) {
	tracker := NewDirtyTracker()

	prev := NewBuffer(20, 10)
	curr := NewBuffer(20, 10)
	baseStyle := style.Style{}

	// 填充内容
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			prev.SetCell(x, y, ' ', baseStyle)
			curr.SetCell(x, y, ' ', baseStyle)
		}
	}

	// 修改多个单元格（两个分开的区域）
	curr.SetCell(5, 5, 'X', baseStyle)
	curr.SetCell(6, 5, 'X', baseStyle)
	curr.SetCell(15, 8, 'Y', baseStyle)

	result := tracker.Diff(prev, curr)

	if !result.HasChanges {
		t.Error("Should detect multiple changes")
	}
	if result.ChangedCells != 3 {
		t.Errorf("Expected 3 changed cells, got %d", result.ChangedCells)
	}
	// 应该有两个区域（两个X合并，Y单独一个）
	if len(result.DirtyRegions) != 2 {
		t.Errorf("Expected 2 dirty regions, got %d", len(result.DirtyRegions))
	}
}

func TestDiff_AdjacentChangesMerge(t *testing.T) {
	tracker := NewDirtyTracker()

	prev := NewBuffer(10, 10)
	curr := NewBuffer(10, 10)
	baseStyle := style.Style{}

	// 填充内容
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			prev.SetCell(x, y, ' ', baseStyle)
			curr.SetCell(x, y, ' ', baseStyle)
		}
	}

	// 修改相邻的单元格（应该合并为一个区域）
	for x := 3; x < 7; x++ {
		curr.SetCell(x, 5, 'X', baseStyle)
	}

	result := tracker.Diff(prev, curr)

	if !result.HasChanges {
		t.Error("Should detect changes")
	}
	// 相邻单元格应该合并为一个区域
	if len(result.DirtyRegions) != 1 {
		t.Errorf("Expected 1 merged dirty region, got %d", len(result.DirtyRegions))
	}
	region := result.DirtyRegions[0]
	if region.Width != 4 || region.Height != 1 {
		t.Errorf("Expected 4x1 region, got %dx%d", region.Width, region.Height)
	}
}

func TestDiff_SizeChange(t *testing.T) {
	tracker := NewDirtyTracker()

	prev := NewBuffer(10, 10)
	curr := NewBuffer(20, 15)

	result := tracker.Diff(prev, curr)

	if !result.HasChanges {
		t.Error("Should detect size change")
	}
	// 尺寸变化应该返回全屏
	if len(result.DirtyRegions) != 1 {
		t.Errorf("Expected 1 dirty region for size change, got %d", len(result.DirtyRegions))
	}
	if result.DirtyRegions[0].Width != 20 || result.DirtyRegions[0].Height != 15 {
		t.Error("Dirty region should cover new buffer size")
	}
}

func TestDiff_StyleChange(t *testing.T) {
	tracker := NewDirtyTracker()

	prev := NewBuffer(10, 5)
	curr := NewBuffer(10, 5)

	style1 := style.Style{}
	style2 := style.Style{}.Bold(true)

	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			prev.SetCell(x, y, 'A', style1)
			curr.SetCell(x, y, 'A', style2)
		}
	}

	result := tracker.Diff(prev, curr)

	if !result.HasChanges {
		t.Error("Should detect style changes")
	}
	if result.ChangedCells != 50 { // 10x5 = 50
		t.Errorf("Expected 50 changed cells, got %d", result.ChangedCells)
	}
}

// ==============================================================================
// 区域合并测试
// ==============================================================================

func TestMergeTwoRects(t *testing.T) {
	tracker := NewDirtyTracker()

	r1 := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	r2 := Rect{X: 10, Y: 0, Width: 5, Height: 10}

	merged := tracker.mergeTwoRects(r1, r2)

	if merged.X != 0 || merged.Y != 0 {
		t.Errorf("Expected origin at (0,0), got (%d,%d)", merged.X, merged.Y)
	}
	if merged.Width != 15 || merged.Height != 10 {
		t.Errorf("Expected size 15x10, got %dx%d", merged.Width, merged.Height)
	}
}

func TestShouldMerge_Overlap(t *testing.T) {
	tracker := NewDirtyTracker()

	r1 := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	r2 := Rect{X: 5, Y: 5, Width: 10, Height: 10} // 重叠

	if !tracker.shouldMerge(r1, r2) {
		t.Error("Overlapping rects should merge")
	}
}

func TestShouldMerge_Adjacent(t *testing.T) {
	tracker := NewDirtyTracker()

	r1 := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	r2 := Rect{X: 10, Y: 0, Width: 5, Height: 10} // 相邻

	if !tracker.shouldMerge(r1, r2) {
		t.Error("Adjacent rects should merge")
	}

	r3 := Rect{X: 0, Y: 10, Width: 10, Height: 5} // 下方相邻
	if !tracker.shouldMerge(r1, r3) {
		t.Error("Adjacent rects should merge")
	}
}

func TestShouldMerge_Separated(t *testing.T) {
	tracker := NewDirtyTracker()

	r1 := Rect{X: 0, Y: 0, Width: 5, Height: 5}
	r2 := Rect{X: 20, Y: 20, Width: 5, Height: 5} // 分开

	if tracker.shouldMerge(r1, r2) {
		t.Error("Separated rects should not merge")
	}
}

// ==============================================================================
// Flood Fill 测试
// ==============================================================================

func TestExtractRegion_SingleCell(t *testing.T) {
	tracker := NewDirtyTracker()
	grid := newDirtyGrid(10, 10)
	grid.Mark(5, 5)

	visited := make([][]bool, 10)
	for i := range visited {
		visited[i] = make([]bool, 10)
	}

	region := tracker.extractRegion(5, 5, grid, visited)

	if region.X != 5 || region.Y != 5 {
		t.Errorf("Expected region at (5,5), got (%d,%d)", region.X, region.Y)
	}
	if region.Width != 1 || region.Height != 1 {
		t.Errorf("Expected 1x1 region, got %dx%d", region.Width, region.Height)
	}
}

func TestExtractRegion_ConnectedCells(t *testing.T) {
	tracker := NewDirtyTracker()
	grid := newDirtyGrid(10, 10)

	// 标记一个 3x2 的连通区域
	for x := 3; x < 6; x++ {
		for y := 4; y < 6; y++ {
			grid.Mark(x, y)
		}
	}

	visited := make([][]bool, 10)
	for i := range visited {
		visited[i] = make([]bool, 10)
	}

	region := tracker.extractRegion(3, 4, grid, visited)

	// 应该提取整个连通区域
	if region.X != 3 || region.Y != 4 {
		t.Errorf("Expected region origin at (3,4), got (%d,%d)", region.X, region.Y)
	}
	if region.Width != 3 || region.Height != 2 {
		t.Errorf("Expected 3x2 region, got %dx%d", region.Width, region.Height)
	}
}

// ==============================================================================
// 边界情况测试
// ==============================================================================

func TestDiff_EmptyBuffers(t *testing.T) {
	tracker := NewDirtyTracker()

	// 两个空 buffer
	prev := NewBuffer(0, 0)
	curr := NewBuffer(0, 0)

	result := tracker.Diff(prev, curr)

	if result.HasChanges {
		t.Error("Empty buffers should have no changes")
	}
}

func TestGetDirtyRects_AllDirty(t *testing.T) {
	tracker := NewDirtyTracker()
	tracker.SetPreviousBuffer(NewBuffer(100, 50))

	rects := tracker.GetDirtyRects()

	if len(rects) != 1 {
		t.Errorf("Expected 1 rect when all dirty, got %d", len(rects))
	}
	if rects[0].X != 0 || rects[0].Y != 0 {
		t.Error("All dirty rect should start at (0,0)")
	}
	if rects[0].Width != 100 || rects[0].Height != 50 {
		t.Errorf("All dirty rect should be 100x50, got %dx%d",
			rects[0].Width, rects[0].Height)
	}
}

func TestString(t *testing.T) {
	tracker := NewDirtyTracker()

	str := tracker.String()
	if str != "DirtyTracker{ALL}" {
		t.Errorf("Expected 'DirtyTracker{ALL}', got '%s'", str)
	}

	tracker.Clear()
	tracker.MarkCell(1, 1)

	str = tracker.String()
	if str == "DirtyTracker{ALL}" {
		t.Error("Should not be ALL after clearing and marking single cell")
	}
}

// ==============================================================================
// 性能测试
// ==============================================================================

func BenchmarkDiff_LargeBuffer(b *testing.B) {
	tracker := NewDirtyTracker()
	prev := NewBuffer(80, 24)
	curr := NewBuffer(80, 24)
	baseStyle := style.Style{}

	// 填充内容
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			prev.SetCell(x, y, ' ', baseStyle)
			curr.SetCell(x, y, ' ', baseStyle)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Diff(prev, curr)
	}
}

func BenchmarkDiff_WithChanges(b *testing.B) {
	tracker := NewDirtyTracker()
	prev := NewBuffer(80, 24)
	curr := NewBuffer(80, 24)
	baseStyle := style.Style{}

	// 填充内容
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			prev.SetCell(x, y, ' ', baseStyle)
			curr.SetCell(x, y, ' ', baseStyle)
		}
	}

	// 修改 10% 的单元格
	for i := 0; i < 192; i++ {
		x := i % 80
		y := i / 80
		curr.SetCell(x, y, 'X', baseStyle)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Diff(prev, curr)
	}
}

// ==============================================================================
// 辅助函数
// ==============================================================================

func TestCellsEqual(t *testing.T) {
	style1 := style.Style{}
	style2 := style.Style{}.Bold(true)

	cell1 := Cell{Char: 'A', Style: style1, Width: 1}
	cell2 := Cell{Char: 'A', Style: style1, Width: 1}
	cell3 := Cell{Char: 'A', Style: style2, Width: 1}
	cell4 := Cell{Char: 'B', Style: style1, Width: 1}

	if !cellsEqual(cell1, cell2) {
		t.Error("Identical cells should be equal")
	}
	if cellsEqual(cell1, cell3) {
		t.Error("Cells with different styles should not be equal")
	}
	if cellsEqual(cell1, cell4) {
		t.Error("Cells with different chars should not be equal")
	}
}
