package layout

// ==============================================================================
// Layout Types (V3)
// ==============================================================================
// 核心布局类型定义

// Node 布局节点接口
// 这是布局引擎操作的抽象节点
type Node interface {
	// ID 返回节点唯一标识
	ID() string

	// Type 返回节点类型
	Type() string

	// Children 返回子节点
	Children() []Node

	// GetPosition 获取位置
	GetPosition() (x, y int)

	// SetPosition 设置位置
	SetPosition(x, y int)

	// GetSize 获取尺寸
	GetSize() (width, height int)

	// SetSize 设置尺寸
	SetSize(width, height int)

	// GetWidth 获取宽度
	GetWidth() int

	// GetHeight 获取高度
	GetHeight() int
}

// Size 尺寸
type Size struct {
	Width  int
	Height int
}

// Point 位置
type Point struct {
	X int
	Y int
}

// Rect 矩形区域
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// LayoutBox 布局结果盒子
// 表示一个节点在布局后的最终位置和尺寸
type LayoutBox struct {
	// ID 节点ID
	ID string

	// X, Y 位置（相对于父节点）
	X int
	Y int

	// Width, Height 尺寸
	Width  int
	Height int

	// Baseline 基线（用于文本对齐）
	Baseline int

	// Children 子节点布局结果
	Children []*LayoutBox
}

// LayoutResult 布局结果
// 一次布局计算的完整结果
type LayoutResult struct {
	// Boxes 所有节点的布局结果
	Boxes []LayoutBox

	// Root 根节点
	Root *LayoutBox

	//ContentSize 内容尺寸
	ContentSize Size

	// Dirty 脏标记
	Dirty bool
}

// Constraints 布局约束
type Constraints struct {
	// MinWidth 最小宽度
	MinWidth int

	// MaxWidth 最大宽度
	MaxWidth int

	// MinHeight 最小高度
	MinHeight int

	// MaxHeight 最大高度
	MaxHeight int
}

// NewConstraints 创建约束
func NewConstraints(minWidth, maxWidth, minHeight, maxHeight int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// Tight 创建紧约束（固定尺寸）
func TightConstraints(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Loose 创建松约束（只有最小值）
func LooseConstraints(minWidth, minHeight int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  MaxInt,
		MinHeight: minHeight,
		MaxHeight: MaxInt,
	}
}

// Unbounded 创建无界约束
func UnboundedConstraints() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  MaxInt,
		MinHeight: 0,
		MaxHeight: MaxInt,
	}
}

// Width 创建宽度约束
func (c Constraints) Width(minWidth, maxWidth int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
}

// Height 创建高度约束
func (c Constraints) Height(minHeight, maxHeight int) Constraints {
	return Constraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// IsTight 检查是否为紧约束
func (c Constraints) IsTight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// IsBounded 检查是否有界
func (c Constraints) IsBounded() bool {
	return c.MaxWidth < MaxInt || c.MaxHeight < MaxInt
}

// Constrain 约束尺寸到范围内
func (c Constraints) Constrain(width, height int) (int, int) {
	if width < c.MinWidth {
		width = c.MinWidth
	}
	if width > c.MaxWidth {
		width = c.MaxWidth
	}
	if height < c.MinHeight {
		height = c.MinHeight
	}
	if height > c.MaxHeight {
		height = c.MaxHeight
	}
	return width, height
}

// ConstrainWidth 约束宽度
func (c Constraints) ConstrainWidth(width int) int {
	if width < c.MinWidth {
		return c.MinWidth
	}
	if width > c.MaxWidth {
		return c.MaxWidth
	}
	return width
}

// ConstrainHeight 约束高度
func (c Constraints) ConstrainHeight(height int) int {
	if height < c.MinHeight {
		return c.MinHeight
	}
	if height > c.MaxHeight {
		return c.MaxHeight
	}
	return height
}

// MaxInt 最大整数值（表示无界）
const MaxInt = 1 << 30

// Measurable 可测量节点
type Measurable interface {
	Node
	// Measure 测量节点在给定约束下的理想尺寸
	Measure(constraints Constraints) Size
}
