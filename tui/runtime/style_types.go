package runtime

// Layout style and type definitions

// Direction represents layout direction for flexbox
type Direction string

const (
	DirectionRow    Direction = "row"
	DirectionColumn Direction = "column"
)

// Align specifies alignment for cross-axis
type Align string

const (
	AlignStart   Align = "start"
	AlignCenter  Align = "center"
	AlignEnd     Align = "end"
	AlignStretch Align = "stretch"
)

// Justify specifies justification for main-axis
type Justify string

const (
	JustifyStart        Justify = "start"
	JustifyCenter       Justify = "center"
	JustifyEnd          Justify = "end"
	JustifySpaceBetween Justify = "space-between"
	JustifySpaceAround  Justify = "space-around"
	JustifySpaceEvenly  Justify = "space-evenly"
)

// NodeType represents the type of a layout node
type NodeType string

const (
	NodeTypeText   NodeType = "text"
	NodeTypeFlex   NodeType = "flex"
	NodeTypeRow    NodeType = "row"
	NodeTypeColumn NodeType = "column"
	NodeTypeCustom NodeType = "custom"
)

// Overflow represents overflow behavior
type Overflow string

const (
	OverflowVisible Overflow = "visible"
	OverflowHidden  Overflow = "hidden"
	OverflowScroll  Overflow = "scroll"
)

// Insets represents box insets (padding/margin)
type Insets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// NewInsets creates a new Insets
func NewInsets(top, right, bottom, left int) Insets {
	return Insets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}
