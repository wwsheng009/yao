package layout

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
)

// Flex Flex 布局容器
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

// WithChildren 设置子组件 (V3: 使用 Node 类型)
func (f *Flex) WithChildren(children ...component.Node) *Flex {
	for _, child := range children {
		f.Add(child)
	}
	return f
}

// Render 渲染 Flex 容器 (V2 兼容)
func (f *Flex) Render(ctx *component.RenderContext) string {
	if !f.IsVisible() || f.ChildCount() == 0 {
		return ""
	}

	width, height := ctx.AvailableWidth, ctx.AvailableHeight

	// 计算子组件尺寸
	sizes := f.calculateSizes(width, height)

	// 渲染子组件
	var result []string
	currentPos := 0

	for i, child := range f.GetChildren() {
		childW, childH := sizes[i].width, sizes[i].height

		childCtx := ctx.WithOffset(currentPos, 0)
		childCtx.AvailableWidth = childW
		childCtx.AvailableHeight = childH

		// V2 兼容：如果 child 是 V2 Component，调用 Render()
		if v2Comp, ok := child.(component.Component); ok {
			content := v2Comp.Render(childCtx)
			contentLines := strings.Split(content, "\n")

			// 处理垂直布局
			if f.direction == Column {
				for _, line := range contentLines {
					result = append(result, line)
				}
				currentPos += childH
			} else {
				// 水平布局 - 需要合并行
				f.mergeHorizontal(&result, contentLines, childW, currentPos, height)
				currentPos += childW + f.gap
			}
		}
	}

	return strings.Join(result, "\n")
}

// Size 尺寸
type Size struct {
	width  int
	height int
}

// calculateSizes 计算子组件尺寸 (V2/V3 兼容)
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
			prefW := f.getPreferredWidth(child)
			prefSizes[i] = prefW
			prefTotal += prefW
		}

		// 分配空间
		if prefTotal <= availW {
			// 按首选尺寸分配
			for i := range children {
				sizes[i].width = prefSizes[i]
			}
		} else {
			// 平均分配
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

		// 高度使用可用高度
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

		// 计算每个子组件的首选高度
		var prefTotal int
		prefSizes := make([]int, count)
		for i, child := range children {
			prefH := f.getPreferredHeight(child)
			prefSizes[i] = prefH
			prefTotal += prefH
		}

		// 分配空间
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

		// 宽度使用可用宽度
		for i := range sizes {
			sizes[i].width = availableW
		}
	}

	return sizes
}

// getPreferredWidth 获取子组件首选宽度 (V2/V3 兼容)
func (f *Flex) getPreferredWidth(child component.Node) int {
	// 尝试 V2 Component 接口
	if v2Comp, ok := child.(component.Component); ok {
		w, _ := v2Comp.GetPreferredSize()
		return w
	}
	// 尝试 V3 Measurable 接口
	if measurable, ok := child.(component.Measurable); ok {
		w, _ := measurable.Measure(1000, 1000)
		return w
	}
	return 0
}

// getPreferredHeight 获取子组件首选高度 (V2/V3 兼容)
func (f *Flex) getPreferredHeight(child component.Node) int {
	// 尝试 V2 Component 接口
	if v2Comp, ok := child.(component.Component); ok {
		_, h := v2Comp.GetPreferredSize()
		return h
	}
	// 尝试 V3 Measurable 接口
	if measurable, ok := child.(component.Measurable); ok {
		_, h := measurable.Measure(1000, 1000)
		return h
	}
	return 0
}

// mergeHorizontal 合并水平布局的内容
func (f *Flex) mergeHorizontal(result *[]string, contentLines []string, width, offsetX, height int) {
	// 扩展结果到足够的行数
	for len(*result) < len(contentLines) {
		*result = append(*result, "")
	}

	for y, line := range contentLines {
		if y >= height {
			break
		}

		// 确保行足够长
		for len(*result) <= y {
			*result = append(*result, "")
		}

		// 在指定位置插入内容
		resultLine := (*result)[y]
		paddedLine := line
		lineLen := utf8RuneCount(line)
		if lineLen < width {
			paddedLine = line + strings.Repeat(" ", width-lineLen)
		}

		// 合并到结果行
		if offsetX+len(paddedLine) <= len(resultLine) {
			*result = append((*result)[:y], resultLine[:offsetX]+paddedLine+resultLine[offsetX+len(paddedLine):])
		} else {
			// 扩展行
			for offsetX+len(paddedLine) > len(resultLine) {
				resultLine += " "
			}
			if len(*result) > y {
				*result = append((*result)[:y], resultLine[:offsetX]+paddedLine)
			}
		}
	}
}

// GetPreferredSize 获取首选尺寸 (V2 兼容)
func (f *Flex) GetPreferredSize() (width, height int) {
	children := f.GetChildren()

	if f.direction == Row {
		totalW := 0
		maxH := 0
		for i, child := range children {
			w := f.getPreferredWidth(child)
			h := f.getPreferredHeight(child)
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

	// Column
	maxW := 0
	totalH := 0
	for i, child := range children {
		w := f.getPreferredWidth(child)
		h := f.getPreferredHeight(child)
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
