package layout

import (
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// Flex Flex 布局容器 (V3)
type Flex struct {
	*component.BaseContainer

	direction Direction
	gap       int
}

// Direction 布局方向
type Direction int

const (
	Row Direction = iota
	Column
)

// NewFlex 创建 Flex 容器
func NewFlex(dir Direction) *Flex {
	return &Flex{
		BaseContainer: component.NewBaseContainer("flex"),
		direction:     dir,
		gap:           0,
	}
}

// NewRow 创建水平布局
func NewRow() *Flex {
	return NewFlex(Row)
}

// NewColumn 创建垂直布局
func NewColumn() *Flex {
	return NewFlex(Column)
}

// WithDirection 设置方向
func (f *Flex) WithDirection(dir Direction) *Flex {
	f.direction = dir
	return f
}

// WithGap 设置间距
func (f *Flex) WithGap(gap int) *Flex {
	f.gap = gap
	return f
}

// WithChildren 设置子组件
func (f *Flex) WithChildren(children ...component.Node) *Flex {
	for _, child := range children {
		f.Add(child)
	}
	return f
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (f *Flex) Measure(maxWidth, maxHeight int) (width, height int) {
	children := f.GetChildren()
	if len(children) == 0 {
		return 0, 0
	}

	if f.direction == Row {
		// 水平布局：宽度累加，高度取最大
		totalW := 0
		maxH := 0
		for i, child := range children {
			w, h := f.measureChild(child, maxWidth, maxHeight)
			totalW += w
			if h > maxH {
				maxH = h
			}
			if i < len(children)-1 {
				totalW += f.gap
			}
		}
		return totalW, maxH
	}

	// 垂直布局：宽度取最大，高度累加
	maxW := 0
	totalH := 0
	for i, child := range children {
		w, h := f.measureChild(child, maxWidth, maxHeight)
		if w > maxW {
			maxW = w
		}
		totalH += h
		if i < len(children)-1 {
			totalH += f.gap
		}
	}
	return maxW, totalH
}

// measureChild 测量子组件尺寸
func (f *Flex) measureChild(child component.Node, maxWidth, maxHeight int) (width, height int) {
	if measurable, ok := child.(component.Measurable); ok {
		return measurable.Measure(maxWidth, maxHeight)
	}
	return 0, 0
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制到缓冲区
func (f *Flex) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !f.IsVisible() || f.ChildCount() == 0 {
		return
	}

	children := f.GetChildren()
	availableWidth := ctx.AvailableWidth
	availableHeight := ctx.AvailableHeight

	// 计算子组件尺寸和位置
	sizes := f.calculateSizes(availableWidth, availableHeight)

	currentPos := 0
	for i, child := range children {
		childW, childH := sizes[i].width, sizes[i].height

		childCtx := component.PaintContext{
			AvailableWidth:  childW,
			AvailableHeight: childH,
			X:               ctx.X + currentPos,
			Y:               ctx.Y,
		}

		// 绘制可绘制的子组件
		if paintable, ok := child.(component.Paintable); ok {
			paintable.Paint(childCtx, buf)
		}

		// 更新位置
		if f.direction == Row {
			currentPos += childW + f.gap
		} else {
			currentPos += childH + f.gap
		}
	}
}

// ============================================================================
// Layout 接口实现
// ============================================================================

// Measure 布局测量
func (f *Flex) LayoutMeasure(container component.Container, availableWidth, availableHeight int) (width, height int) {
	return f.Measure(availableWidth, availableHeight)
}

// Layout 布局计算
func (f *Flex) LayoutLayout(container component.Container, x, y, width, height int) {
	// 布局计算在 Paint 中完成
}

// Invalidate 使布局失效
func (f *Flex) Invalidate() {
	// 无状态，无需失效
}

// ============================================================================
// 内部方法
// ============================================================================

// Size 尺寸
type Size struct {
	width  int
	height int
}

// calculateSizes 计算子组件尺寸
func (f *Flex) calculateSizes(availableW, availableH int) []Size {
	children := f.GetChildren()
	count := len(children)
	sizes := make([]Size, count)

	if f.direction == Row {
		// 水平布局
		totalGap := f.gap * (count - 1)
		availW := availableW - totalGap
		if availW < 0 {
			availW = 0
		}

		// 计算每个子组件的首选宽度
		var prefTotal int
		prefSizes := make([]int, count)
		for i, child := range children {
			prefW, _ := f.measureChild(child, 1000, 1000)
			prefSizes[i] = prefW
			prefTotal += prefW
		}

		// 分配空间
		if prefTotal <= availW {
			for i := range children {
				sizes[i].width = prefSizes[i]
			}
		} else {
			if count > 0 {
				avgW := availW / count
				for i := range sizes {
					sizes[i].width = avgW
					if i == count-1 {
						sizes[i].width = availW - avgW*(count-1)
					}
				}
			}
		}

		for i := range sizes {
			sizes[i].height = availableH
		}

	} else {
		// 垂直布局
		totalGap := f.gap * (count - 1)
		availH := availableH - totalGap
		if availH < 0 {
			availH = 0
		}

		var prefTotal int
		prefSizes := make([]int, count)
		for i, child := range children {
			_, prefH := f.measureChild(child, 1000, 1000)
			prefSizes[i] = prefH
			prefTotal += prefH
		}

		if prefTotal <= availH {
			for i := range children {
				sizes[i].height = prefSizes[i]
			}
		} else {
			if count > 0 {
				avgH := availH / count
				for i := range sizes {
					sizes[i].height = avgH
					if i == count-1 {
						sizes[i].height = availH - avgH*(count-1)
					}
				}
			}
		}

		for i := range sizes {
			sizes[i].width = availableW
		}
	}

	return sizes
}
