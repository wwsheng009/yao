package runtime

// ===========================================================================
// TUI Runtime - Core Type Definitions
// ===========================================================================
//
// This package provides an independent runtime implementation for tui/tui.
// It does NOT depend on tui/tui/runtime, providing a clean separation.
//
// Design Principles:
//   - Lightweight and focused on TUI (Bubble Tea) use cases
//   - Compatible with the tui/tui/runtime API where it makes sense
//   - Independent evolution from tui/tui/runtime
//

// ===========================================================================
// Node Types
// ===========================================================================

// NodeType represents the type of a layout node.
type NodeType string

const (
	NodeTypeRow    NodeType = "row"
	NodeTypeColumn NodeType = "column"
	NodeTypeFlex   NodeType = "flex"
	NodeTypeText   NodeType = "text"
	NodeTypeCustom NodeType = "custom"
)

// ===========================================================================
// Position
// ===========================================================================

// PositionType represents the positioning type.
type PositionType string

const (
	PositionRelative PositionType = "relative"
	PositionAbsolute PositionType = "absolute"
)

// Position represents the position configuration for a node.
type Position struct {
	Type   PositionType
	Top    *int
	Right  *int
	Bottom *int
	Left   *int
}

// NewPosition creates a new relative position (default).
func NewPosition() Position {
	return Position{Type: PositionRelative}
}

// NewAbsolutePosition creates a new absolute position with the given offsets.
func NewAbsolutePosition(top, right, bottom, left *int) Position {
	return Position{
		Type:   PositionAbsolute,
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

// ===========================================================================
// Direction
// ===========================================================================

// Direction represents the layout direction.
type Direction string

const (
	DirectionRow    Direction = "row"
	DirectionColumn Direction = "column"
)

// ===========================================================================
// Alignment
// ===========================================================================

// Align represents the cross-axis alignment.
type Align string

const (
	AlignStart   Align = "start"
	AlignCenter  Align = "center"
	AlignEnd     Align = "end"
	AlignStretch Align = "stretch"
)

// Justify represents the main-axis justification.
type Justify string

const (
	JustifyStart        Justify = "start"
	JustifyCenter       Justify = "center"
	JustifyEnd          Justify = "end"
	JustifySpaceBetween Justify = "space-between"
	JustifySpaceAround  Justify = "space-around"
	JustifySpaceEvenly  Justify = "space-evenly"
)

// ===========================================================================
// Overflow
// ===========================================================================

// Overflow represents how content overflow is handled.
type Overflow string

const (
	OverflowVisible Overflow = "visible"
	OverflowHidden  Overflow = "hidden"
	OverflowScroll  Overflow = "scroll"
)

// ===========================================================================
// Insets
// ===========================================================================

// Insets represents padding or border values.
type Insets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// NewInsets creates uniform insets from a single value.
func NewInsets(value int) Insets {
	return Insets{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}

// NewInsetsVertical creates vertical insets (top/bottom).
func NewInsetsVertical(value int) Insets {
	return Insets{Top: value, Bottom: value}
}

// NewInsetsHorizontal creates horizontal insets (left/right).
func NewInsetsHorizontal(value int) Insets {
	return Insets{Left: value, Right: value}
}

// ===========================================================================
// Size
// ===========================================================================

// Size represents a 2D size.
type Size struct {
	Width  int
	Height int
}

// ===========================================================================
// BoxConstraints
// ===========================================================================

// BoxConstraints represents the constraints for layout.
type BoxConstraints struct {
	MinWidth, MaxWidth  int
	MinHeight, MaxHeight int
}

// NewBoxConstraints creates a new box constraints.
func NewBoxConstraints(minW, maxW, minH, maxH int) BoxConstraints {
	return BoxConstraints{
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: minH,
		MaxHeight: maxH,
	}
}

// Loosen creates a looser constraint (min = 0).
func (bc BoxConstraints) Loosen() BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  bc.MaxWidth,
		MinHeight: 0,
		MaxHeight: bc.MaxHeight,
	}
}

// Tighten creates a tight constraint (min = max).
func (bc BoxConstraints) Tighten() BoxConstraints {
	w := bc.MaxWidth
	h := bc.MaxHeight
	return BoxConstraints{
		MinWidth:  w,
		MaxWidth:  w,
		MinHeight: h,
		MaxHeight: h,
	}
}

// ===========================================================================
// LayoutResult
// ===========================================================================

// LayoutBox represents a positioned node in the layout result.
type LayoutBox struct {
	NodeID string
	X, Y   int
	W, H   int
	ZIndex int
	// Node reference is omitted to avoid circular dependencies
	// Use NodeID to look up nodes if needed
}

// LayoutResult contains the result of a layout pass.
type LayoutResult struct {
	Boxes      []LayoutBox
	Dirty      bool
	RootWidth  int
	RootHeight int
}

// FindBoxByID finds a LayoutBox by node ID.
func (lr *LayoutResult) FindBoxByID(id string) *LayoutBox {
	for i := range lr.Boxes {
		if lr.Boxes[i].NodeID == id {
			return &lr.Boxes[i]
		}
	}
	return nil
}

// ===========================================================================
// Frame
// ===========================================================================

// Frame represents a rendered frame as a string.
// For TUI purposes, this is just the view string.
type Frame struct {
	content string
	width   int
	height  int
}

// NewFrame creates a new frame.
func NewFrame(content string, width, height int) Frame {
	return Frame{
		content: content,
		width:   width,
		height:  height,
	}
}

// String returns the frame content.
func (f Frame) String() string {
	return f.content
}

// Width returns the frame width.
func (f Frame) Width() int {
	return f.width
}

// Height returns the frame height.
func (f Frame) Height() int {
	return f.height
}
