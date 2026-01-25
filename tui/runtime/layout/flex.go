package layout

// ==============================================================================
// Flexbox Layout Algorithm (V3)
// ==============================================================================
// 简化的 Flexbox 布局算法实现

// FlexDirection 弹性方向
type FlexDirection int

const (
	FlexRow FlexDirection = iota
	FlexColumn
	FlexRowReverse
	FlexColumnReverse
)

// MainAxisAlignment 主轴对齐
type MainAxisAlignment int

const (
	// MainStart 主轴起点对齐
	MainStart MainAxisAlignment = iota
	// MainEnd 主轴终点对齐
	MainEnd
	// Center 主轴居中对齐
	Center
	// SpaceBetween 两端对齐，间距平均分配
	SpaceBetween
	// SpaceAround 每个子元素两侧间距相等
	SpaceAround
	// SpaceEvenly 所有间距相等
	SpaceEvenly
)

// CrossAxisAlignment 交叉轴对齐
type CrossAxisAlignment int

const (
	// CrossStart 交叉轴起点对齐
	CrossStart CrossAxisAlignment = iota
	// CrossEnd 交叉轴终点对齐
	CrossEnd
	// CrossCenter 交叉轴居中对齐
	CrossCenter
	// Stretch 拉伸填满交叉轴
	Stretch
)

// FlexStyle 弹性布局样式
type FlexStyle struct {
	// Direction 弹性方向
	Direction FlexDirection

	// MainAxis 主轴对齐
	MainAxis MainAxisAlignment

	// CrossAxis 交叉轴对齐
	CrossAxis CrossAxisAlignment

	// Gap 主轴间距
	Gap int

	// CrossGap 交叉轴间距
	CrossGap int

	// Padding 内边距
	Padding Padding

	// FlexibleChildren 可伸缩子节点索引和配置
	FlexibleChildren map[int]*Flex
}

// Padding 内边距
type Padding struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Flex 弹性配置
type Flex struct {
	// Grow 放大比例（默认0，不放大）
	Grow int

	// Shrink 缩小比例（默认1，可以缩小）
	Shrink int

	// Basis 基础尺寸（默认auto，由内容决定）
	Basis int
}

// DefaultFlexStyle 默认弹性样式
func DefaultFlexStyle() *FlexStyle {
	return &FlexStyle{
		Direction:  FlexColumn,
		MainAxis:   MainStart,
		CrossAxis:  CrossStart,
		Gap:        0,
		CrossGap:   0,
		Padding:    Padding{},
		FlexibleChildren: make(map[int]*Flex),
	}
}

// FlexLayout 弹性布局节点
type FlexLayout struct {
	id       string
	children []Node
	style    *FlexStyle
	size     Size
	position Point
}

// NewFlexLayout 创建弹性布局
func NewFlexLayout(id string, children []Node) *FlexLayout {
	return &FlexLayout{
		id:       id,
		children: children,
		style:    DefaultFlexStyle(),
	}
}

// ID 返回节点ID
func (f *FlexLayout) ID() string {
	return f.id
}

// Type 返回节点类型
func (f *FlexLayout) Type() string {
	return "flex"
}

// Children 返回子节点
func (f *FlexLayout) Children() []Node {
	return f.children
}

// GetPosition 获取位置
func (f *FlexLayout) GetPosition() (int, int) {
	return f.position.X, f.position.Y
}

// SetPosition 设置位置
func (f *FlexLayout) SetPosition(x, y int) {
	f.position.X = x
	f.position.Y = y
}

// GetSize 获取尺寸
func (f *FlexLayout) GetSize() (int, int) {
	return f.size.Width, f.size.Height
}

// SetSize 设置尺寸
func (f *FlexLayout) SetSize(width, height int) {
	f.size.Width = width
	f.size.Height = height
}

// GetWidth 获取宽度
func (f *FlexLayout) GetWidth() int {
	return f.size.Width
}

// GetHeight 获取高度
func (f *FlexLayout) GetHeight() int {
	return f.size.Height
}

// SetDirection 设置弹性方向
func (f *FlexLayout) SetDirection(dir FlexDirection) {
	f.style.Direction = dir
}

// SetMainAxis 设置主轴对齐
func (f *FlexLayout) SetMainAxis(align MainAxisAlignment) {
	f.style.MainAxis = align
}

// SetCrossAxis 设置交叉轴对齐
func (f *FlexLayout) SetCrossAxis(align CrossAxisAlignment) {
	f.style.CrossAxis = align
}

// SetGap 设置主轴间距
func (f *FlexLayout) SetGap(gap int) {
	f.style.Gap = gap
}

// SetCrossGap 设置交叉轴间距
func (f *FlexLayout) SetCrossGap(gap int) {
	f.style.CrossGap = gap
}

// SetPadding 设置内边距
func (f *FlexLayout) SetPadding(left, right, top, bottom int) {
	f.style.Padding = Padding{
		Left:   left,
		Right:  right,
		Top:    top,
		Bottom: bottom,
	}
}

// SetFlex 设置子节点的弹性配置
func (f *FlexLayout) SetFlex(index int, grow, shrink, basis int) {
	if f.style.FlexibleChildren == nil {
		f.style.FlexibleChildren = make(map[int]*Flex)
	}
	f.style.FlexibleChildren[index] = &Flex{
		Grow:   grow,
		Shrink: shrink,
		Basis:  basis,
	}
}

// Measure 测量节点尺寸
func (f *FlexLayout) Measure(constraints Constraints) Size {
	if len(f.children) == 0 {
		width := constraints.ConstrainWidth(f.style.Padding.Left + f.style.Padding.Right)
		height := constraints.ConstrainHeight(f.style.Padding.Top + f.style.Padding.Bottom)
		return Size{Width: width, Height: height}
	}

	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

	// Phase 1: 测量所有子节点
	childSizes := make([]Size, len(f.children))
	totalMainSize := 0
	maxCrossSize := 0

	for i, child := range f.children {
		childConstraints := f.childConstraints(constraints, i)
		if measurable, ok := child.(Measurable); ok {
			childSizes[i] = measurable.Measure(childConstraints)
		} else {
			// 默认尺寸
			childSizes[i] = Size{Width: childConstraints.MinWidth, Height: childConstraints.MinHeight}
		}

		if isRow {
			// 横向布局：宽度累加，高度取最大
			if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
				// 可伸缩节点，使用 basis
				basis := flex.Basis
				if basis == 0 { // auto
					basis = childSizes[i].Width
				}
				totalMainSize += basis
			} else {
				totalMainSize += childSizes[i].Width
			}
			if childSizes[i].Height > maxCrossSize {
				maxCrossSize = childSizes[i].Height
			}
		} else {
			// 纵向布局：高度累加，宽度取最大
			if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
				basis := flex.Basis
				if basis == 0 { // auto
					basis = childSizes[i].Height
				}
				totalMainSize += basis
			} else {
				totalMainSize += childSizes[i].Height
			}
			if childSizes[i].Width > maxCrossSize {
				maxCrossSize = childSizes[i].Width
			}
		}
	}

	// 添加间距
	gapCount := len(f.children) - 1
	if gapCount > 0 {
		totalMainSize += f.style.Gap * gapCount
	}

	// 计算总尺寸
	var width, height int
	if isRow {
		width = f.style.Padding.Left + totalMainSize + f.style.Padding.Right
		height = f.style.Padding.Top + maxCrossSize + f.style.Padding.Bottom
	} else {
		width = f.style.Padding.Left + maxCrossSize + f.style.Padding.Right
		height = f.style.Padding.Top + totalMainSize + f.style.Padding.Bottom
	}

	return Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

// childConstraints 计算子节点约束
func (f *FlexLayout) childConstraints(constraints Constraints, index int) Constraints {
	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

	// 减去内边距
	availableMain := constraints.MaxWidth - f.style.Padding.Left - f.style.Padding.Right
	availableCross := constraints.MaxHeight - f.style.Padding.Top - f.style.Padding.Bottom
	if !isRow {
		availableMain, availableCross = availableCross, availableMain
	}

	if isRow {
		return Constraints{
			MinWidth:  0,
			MaxWidth:  availableMain,
			MinHeight: constraints.MinHeight,
			MaxHeight: availableCross,
		}
	}
	return Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  availableCross,
		MinHeight: 0,
		MaxHeight: availableMain,
	}
}

// LayoutChildren 布局子节点
func (f *FlexLayout) LayoutChildren(width, height int) []LayoutBox {
	if len(f.children) == 0 {
		return nil
	}

	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse
	isReverse := f.style.Direction == FlexRowReverse || f.style.Direction == FlexColumnReverse

	// 可用空间（减去内边距）
	availableWidth := width - f.style.Padding.Left - f.style.Padding.Right
	availableHeight := height - f.style.Padding.Top - f.style.Padding.Bottom

	// Phase 1: 测量所有子节点
	childSizes := make([]Size, len(f.children))
	fixedTotal := 0    // 固定尺寸总和
	flexGrowTotal := 0 // flex-grow 总和

	for i, child := range f.children {
		constraints := Constraints{}
		if isRow {
			constraints = Constraints{
				MinWidth:  0,
				MaxWidth:  availableWidth,
				MinHeight: 0,
				MaxHeight: availableHeight,
			}
		} else {
			constraints = Constraints{
				MinWidth:  0,
				MaxWidth:  availableWidth,
				MinHeight: 0,
				MaxHeight: availableHeight,
			}
		}

		if measurable, ok := child.(Measurable); ok {
			childSizes[i] = measurable.Measure(constraints)
		} else {
			childSizes[i] = Size{Width: 0, Height: 0}
		}

		if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
			flexGrowTotal += flex.Grow
			// 使用 basis 作为基础尺寸
			if flex.Basis > 0 {
				fixedTotal += flex.Basis
			} else {
				if isRow {
					fixedTotal += childSizes[i].Width
				} else {
					fixedTotal += childSizes[i].Height
				}
			}
		} else {
			// 固定尺寸节点
			if isRow {
				fixedTotal += childSizes[i].Width
			} else {
				fixedTotal += childSizes[i].Height
			}
		}
	}

	// Phase 2: 计算剩余空间
	gapCount := len(f.children) - 1
	totalGap := 0
	if gapCount > 0 {
		totalGap = f.style.Gap * gapCount
	}

	remainingSpace := 0
	if isRow {
		remainingSpace = availableWidth - fixedTotal - totalGap
	} else {
		remainingSpace = availableHeight - fixedTotal - totalGap
	}

	// Phase 3: 分配剩余空间给可伸缩节点
	finalSizes := make([]Size, len(f.children))
	flexIndex := 0
	for i := range f.children {
		if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
			// 按比例分配剩余空间
			extra := 0
			if flexGrowTotal > 0 {
				extra = (remainingSpace * flex.Grow) / flexGrowTotal
			}
			if isRow {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width + extra,
					Height: childSizes[i].Height,
				}
			} else {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width,
					Height: childSizes[i].Height + extra,
				}
			}
			flexIndex++
		} else {
			finalSizes[i] = childSizes[i]
		}
	}

	// Phase 4: 计算位置
	boxes := make([]LayoutBox, len(f.children))

	// 主轴起始位置
	mainPos := 0
	switch f.style.MainAxis {
	case MainStart:
		mainPos = 0
	case MainEnd:
		if isRow {
			mainPos = availableWidth - fixedTotal - totalGap
		} else {
			mainPos = availableHeight - fixedTotal - totalGap
		}
	case Center:
		if isRow {
			mainPos = (availableWidth - fixedTotal - totalGap) / 2
		} else {
			mainPos = (availableHeight - fixedTotal - totalGap) / 2
		}
	case SpaceBetween:
		mainPos = 0
	case SpaceAround, SpaceEvenly:
		// 需要额外计算间距
		mainPos = 0
	}

	// 交叉轴起始位置
	crossPos := 0
	switch f.style.CrossAxis {
	case CrossStart:
		crossPos = 0
	case CrossEnd:
		if isRow {
			crossPos = availableHeight - f.getMaxCrossSize(finalSizes)
		} else {
			crossPos = availableWidth - f.getMaxCrossSize(finalSizes)
		}
	case CrossCenter:
		if isRow {
			crossPos = (availableHeight - f.getMaxCrossSize(finalSizes)) / 2
		} else {
			crossPos = (availableWidth - f.getMaxCrossSize(finalSizes)) / 2
		}
	case Stretch:
		crossPos = 0
	}

	// SpaceBetween/Around/Evenly 的额外间距
	extraGap := 0
	if (f.style.MainAxis == SpaceBetween || f.style.MainAxis == SpaceAround || f.style.MainAxis == SpaceEvenly) && len(f.children) > 1 {
		switch f.style.MainAxis {
		case SpaceBetween:
			extraGap = remainingSpace / gapCount
		case SpaceAround:
			extraGap = remainingSpace / len(f.children)
		case SpaceEvenly:
			extraGap = remainingSpace / (len(f.children) + 1)
		}
	}

	if f.style.MainAxis == SpaceEvenly {
		mainPos += extraGap
	} else if f.style.MainAxis == SpaceAround {
		mainPos += extraGap / 2
	}

	// 布局每个子节点
	for i, child := range f.children {
		var x, y int

		if isReverse {
			idx := len(f.children) - 1 - i
			if isRow {
				x = f.style.Padding.Left + mainPos
				y = f.style.Padding.Top + crossPos
			} else {
				x = f.style.Padding.Left + crossPos
				y = f.style.Padding.Top + mainPos
			}
			mainPos += finalSizes[idx].Width + f.style.Gap
			if extraGap > 0 && i < len(f.children)-1 {
				mainPos += extraGap
			}
		} else {
			if isRow {
				x = f.style.Padding.Left + mainPos
				y = f.style.Padding.Top + crossPos
			} else {
				x = f.style.Padding.Left + crossPos
				y = f.style.Padding.Top + mainPos
			}
			mainPos += finalSizes[i].Width + f.style.Gap
			if extraGap > 0 && i < len(f.children)-1 {
				mainPos += extraGap
			}
		}

		// Stretch 处理
		if f.style.CrossAxis == Stretch {
			if isRow {
				finalSizes[i].Height = availableHeight
			} else {
				finalSizes[i].Width = availableWidth
			}
		}

		boxes[i] = LayoutBox{
			ID:     child.ID(),
			X:      x,
			Y:      y,
			Width:  finalSizes[i].Width,
			Height: finalSizes[i].Height,
		}

		// 设置子节点位置和尺寸
		child.SetPosition(x, y)
		child.SetSize(finalSizes[i].Width, finalSizes[i].Height)
	}

	return boxes
}

// getMaxCrossSize 获取最大交叉轴尺寸
func (f *FlexLayout) getMaxCrossSize(sizes []Size) int {
	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse
	maxSize := 0
	for _, size := range sizes {
		if isRow {
			if size.Height > maxSize {
				maxSize = size.Height
			}
		} else {
			if size.Width > maxSize {
				maxSize = size.Width
			}
		}
	}
	return maxSize
}
