package runtime

import (
	"sync"
)

// cellBufferPool is a pool of reusable CellBuffer instances.
// This reduces GC pressure by reusing buffers instead of allocating new ones.
var cellBufferPool = sync.Pool{
	New: func() interface{} {
		return &CellBuffer{}
	},
}

// maxPooledBufferSize is the maximum buffer size that will be pooled.
// Larger buffers are not pooled to avoid excessive memory usage.
const maxPooledBufferSize = 10000 // 100x100 cells

// NewCellBuffer creates a new CellBuffer, potentially from the pool.
func NewCellBuffer(width, height int) *CellBuffer {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	size := width * height

	// Only pool buffers of reasonable size
	if size <= maxPooledBufferSize {
		buf := cellBufferPool.Get().(*CellBuffer)
		buf.Reset(width, height)
		return buf
	}

	// For large buffers, create a new one
	return createCellBuffer(width, height)
}

// ReleaseCellBuffer returns a CellBuffer to the pool for reuse.
func ReleaseCellBuffer(buf *CellBuffer) {
	if buf == nil {
		return
	}

	size := buf.Width() * buf.Height()
	if size <= maxPooledBufferSize {
		cellBufferPool.Put(buf)
	}
	// Large buffers are not pooled, let GC handle them
}

// createCellBuffer creates a new CellBuffer without pooling.
func createCellBuffer(width, height int) *CellBuffer {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = Cell{
				Char:   ' ',
				Style:  CellStyle{},
				ZIndex: 0,
			}
		}
	}

	return &CellBuffer{
		cells:  cells,
		width:  width,
		height: height,
	}
}

// Reset resets the CellBuffer to the given dimensions.
// This is used by the pool to reuse buffers.
func (b *CellBuffer) Reset(width, height int) {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Check if we need to allocate new cells
	if b.cells == nil || len(b.cells) != height || (len(b.cells) > 0 && len(b.cells[0]) != width) {
		b.cells = make([][]Cell, height)
		for y := 0; y < height; y++ {
			b.cells[y] = make([]Cell, width)
		}
	}

	// Clear all cells
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			b.cells[y][x] = Cell{
				Char:   ' ',
				Style:  CellStyle{},
				ZIndex: 0,
			}
		}
	}

	b.width = width
	b.height = height
}

// nodePool is a pool of reusable LayoutNode instances.
var nodePool = sync.Pool{
	New: func() interface{} {
		return &LayoutNode{
			Style: NewStyle(),
		}
	},
}

// NewLayoutNodePooled creates a new LayoutNode, potentially from the pool.
func NewLayoutNodePooled(id string, nodeType NodeType, style Style) *LayoutNode {
	node := nodePool.Get().(*LayoutNode)
	node.ID = id
	node.Type = nodeType
	node.Style = style
	node.Parent = nil
	node.Children = nil
	node.X = 0
	node.Y = 0
	node.MeasuredWidth = 0
	node.MeasuredHeight = 0
	node.Component = nil
	// Note: Dirty field is managed by LayoutNode itself
	return node
}

// ReleaseLayoutNode returns a LayoutNode to the pool.
func ReleaseLayoutNode(node *LayoutNode) {
	if node == nil {
		return
	}
	// Clear references to avoid memory leaks
	node.Parent = nil
	node.Children = nil
	node.Component = nil
	nodePool.Put(node)
}

// boxPool is a pool of reusable LayoutBox instances.
var boxPool = sync.Pool{
	New: func() interface{} {
		return &LayoutBox{}
	},
}

// NewLayoutBoxPooled creates a new LayoutBox, potentially from the pool.
func NewLayoutBoxPooled() *LayoutBox {
	box := boxPool.Get().(*LayoutBox)
	box.NodeID = ""
	box.Node = nil
	box.X = 0
	box.Y = 0
	box.W = 0
	box.H = 0
	box.ZIndex = 0
	return box
}

// ReleaseLayoutBox returns a LayoutBox to the pool.
func ReleaseLayoutBox(box *LayoutBox) {
	if box == nil {
		return
	}
	box.Node = nil
	boxPool.Put(box)
}

// GetPoolStats returns statistics about the pool usage.
func GetPoolStats() map[string]interface{} {
	return map[string]interface{}{
		"max_pooled_buffer_size": maxPooledBufferSize,
		"buffer_pool_enabled":    true,
		"node_pool_enabled":      true,
		"box_pool_enabled":       true,
	}
}
