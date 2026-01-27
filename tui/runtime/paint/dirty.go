package paint

import "fmt"

// ==============================================================================
// Dirty Region Tracking (V3)
// ==============================================================================
// 脏区域跟踪，用于优化渲染（只重绘变化的部分）

// DirtyTracker 脏区域跟踪器
type DirtyTracker struct {
	// 脏单元格集合
	cells map[cellRef]struct{}

	// 脏矩形列表
	rects []Rect

	// 全局脏标记
	allDirty bool

	// 上次渲染的快照
	prevBuffer *Buffer
}

// cellRef 单元格引用
type cellRef struct {
	x int
	y int
}

// NewDirtyTracker 创建脏区域跟踪器
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		cells:     make(map[cellRef]struct{}),
		rects:     make([]Rect, 0),
		allDirty:  true, // 初始为全脏
		prevBuffer: nil,
	}
}

// MarkCell 标记单元格为脏
func (d *DirtyTracker) MarkCell(x, y int) {
	if d.allDirty {
		return
	}
	d.cells[cellRef{x: x, y: y}] = struct{}{}
}

// MarkRect 标记矩形区域为脏
func (d *DirtyTracker) MarkRect(rect Rect) {
	if d.allDirty {
		return
	}
	d.rects = append(d.rects, rect)
}

// MarkAll 标记全部为脏
func (d *DirtyTracker) MarkAll() {
	d.allDirty = true
	d.cells = make(map[cellRef]struct{})
	d.rects = make([]Rect, 0)
}

// IsAllDirty 检查是否全部为脏
func (d *DirtyTracker) IsAllDirty() bool {
	return d.allDirty
}

// GetDirtyRects 获取脏矩形列表
func (d *DirtyTracker) GetDirtyRects() []Rect {
	if d.allDirty {
		return []Rect{{X: 0, Y: 0, Width: d.prevBuffer.Width, Height: d.prevBuffer.Height}}
	}
	return d.rects
}

// GetDirtyCells 获取脏单元格集合
func (d *DirtyTracker) GetDirtyCells() map[cellRef]struct{} {
	if d.allDirty {
		return nil
	}
	return d.cells
}

// Clear 清除脏标记
func (d *DirtyTracker) Clear() {
	d.allDirty = false
	d.cells = make(map[cellRef]struct{})
	d.rects = make([]Rect, 0)
}

// IsDirtyCell 检查单元格是否为脏
func (d *DirtyTracker) IsDirtyCell(x, y int) bool {
	if d.allDirty {
		return true
	}
	_, ok := d.cells[cellRef{x: x, y: y}]
	if ok {
		return true
	}
	// 检查是否在脏矩形内
	for _, rect := range d.rects {
		if x >= rect.X && x < rect.X+rect.Width &&
			y >= rect.Y && y < rect.Y+rect.Height {
			return true
		}
	}
	return false
}

// MergeDirtyCells 合并脏单元格为矩形
// 优化：将相邻的脏单元格合并为更大的矩形以减少渲染调用
func (d *DirtyTracker) MergeDirtyCells() {
	if d.allDirty || len(d.cells) == 0 {
		return
	}

	// 简化实现：将每个单元格转换为1x1矩形
	// 实际可以实现更智能的合并算法
	for cell := range d.cells {
		d.rects = append(d.rects, Rect{
			X:      cell.x,
			Y:      cell.y,
			Width:  1,
			Height: 1,
		})
	}
	d.cells = make(map[cellRef]struct{})
}

// CompareBuffers 比较两个Buffer，找出差异区域
func (d *DirtyTracker) CompareBuffers(prev, curr *Buffer) {
	if prev == nil || prev.Width != curr.Width || prev.Height != curr.Height {
		d.MarkAll()
		return
	}

	d.prevBuffer = prev

	// 逐个单元格比较
	for y := 0; y < curr.Height; y++ {
		for x := 0; x < curr.Width; x++ {
			prevCell := prev.Cells[y][x]
			currCell := curr.Cells[y][x]
			if !cellsEqual(prevCell, currCell) {
				d.MarkCell(x, y)
			}
		}
	}

	// 合并为矩形
	d.MergeDirtyCells()
}

// cellsEqual 比较两个单元格是否相等
func cellsEqual(a, b Cell) bool {
	return a.Char == b.Char && a.Style == b.Style && a.Width == b.Width
}

// SetPreviousBuffer 设置上次渲染的Buffer
func (d *DirtyTracker) SetPreviousBuffer(buf *Buffer) {
	d.prevBuffer = buf
}

// GetPreviousBuffer 获取上次渲染的Buffer
func (d *DirtyTracker) GetPreviousBuffer() *Buffer {
	return d.prevBuffer
}

// OptimizeRects 优化矩形列表
// 合并重叠或相邻的矩形
func (d *DirtyTracker) OptimizeRects() {
	if len(d.rects) <= 1 {
		return
	}

	// 简化实现：去除完全包含的矩形
	optimized := make([]Rect, 0, len(d.rects))

	for i, r1 := range d.rects {
		contained := false
		for j, r2 := range d.rects {
			if i == j {
				continue
			}
			if rectContains(r2, r1) {
				contained = true
				break
			}
		}
		if !contained {
			optimized = append(optimized, r1)
		}
	}

	d.rects = optimized
}

// rectContains 检查r1是否包含r2
func rectContains(r1, r2 Rect) bool {
	return r2.X >= r1.X && r2.Y >= r1.Y &&
		r2.X+r2.Width <= r1.X+r1.Width &&
		r2.Y+r2.Height <= r1.Y+r1.Height
}

// String 返回脏区域的字符串表示
func (d *DirtyTracker) String() string {
	if d.allDirty {
		return "DirtyTracker{ALL}"
	}
	return "DirtyTracker{" +
		"cells:" + fmt.Sprint(len(d.cells)) +
		",rects:" + fmt.Sprint(len(d.rects)) +
		"}"
}
