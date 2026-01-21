package layout

import "github.com/yaoapp/yao/tui/core"

type LayoutType string

const (
	LayoutFlex     LayoutType = "flex"
	LayoutGrid     LayoutType = "grid"
	LayoutAbsolute LayoutType = "absolute"
)

type Direction string

const (
	DirectionRow    Direction = "row"
	DirectionColumn Direction = "column"
)

type Align string

const (
	AlignStart   Align = "start"
	AlignCenter  Align = "center"
	AlignEnd     Align = "end"
	AlignStretch Align = "stretch"
)

type Justify string

const (
	JustifyStart        Justify = "start"
	JustifyCenter       Justify = "center"
	JustifyEnd          Justify = "end"
	JustifySpaceBetween Justify = "space-between"
	JustifySpaceAround  Justify = "space-around"
	JustifySpaceEvenly  Justify = "space-evenly"
)

type Grow struct {
	Value float64
}

type Size struct {
	Value interface{}
	Min   int
	Max   int
	Unit  string
}

func NewSize(value interface{}) *Size {
	return &Size{
		Value: value,
		Min:   0,
		Max:   0,
		Unit:  "px",
	}
}

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type LayoutNode struct {
	ID        string
	Type      LayoutType
	Children  []*LayoutNode
	Component *core.ComponentInstance
	Style     *LayoutStyle
	Props     map[string]interface{}
	Bound     Rect
	Metrics   *LayoutMetrics
	Parent    *LayoutNode
	Dirty     bool
}

type LayoutStyle struct {
	Direction  Direction
	AlignItems Align
	Justify    Justify
	Wrap       bool
	Gap        int
	Padding    *Padding
	Margin     *Margin
	Width      *Size
	Height     *Size
	MinWidth   int
	MinHeight  int
	MaxWidth   int
	MaxHeight  int
	Position   Position
	Left       int
	Top        int
	Right      int
	Bottom     int
}

type Position string

const (
	PositionRelative Position = "relative"
	PositionAbsolute Position = "absolute"
)

type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

func NewPadding(top, right, bottom, left int) *Padding {
	return &Padding{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

type Margin struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

func NewMargin(top, right, bottom, left int) *Margin {
	return &Margin{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

type LayoutMetrics struct {
	ContentWidth  int
	ContentHeight int
	BorderWidth   int
	BorderHeight  int
	PaddingWidth  int
	PaddingHeight int
	MarginWidth   int
	MarginHeight  int
	TotalWidth    int
	TotalHeight   int
}

type WindowSize struct {
	Width  int
	Height int
}

type LayoutConfig struct {
	Root       *LayoutNode
	WindowSize *WindowSize
	Theme      map[string]interface{}
}

type FlexConfig struct {
	Direction  Direction
	AlignItems Align
	Justify    Justify
	Wrap       bool
	Gap        int
}

type GridConfig struct {
	Columns       int
	Rows          int
	ColumnGap     int
	RowGap        int
	AutoFit       bool
	AutoFill      bool
	MinColumnSize int
}

type LayoutResult struct {
	Nodes   []*LayoutNode
	Dirties []*LayoutNode
}
