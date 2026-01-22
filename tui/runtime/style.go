package runtime

// Style represents declarative layout intent (v1 simplified)
//
// v1 supports:
//   - Width/Height with -1 for auto
//   - FlexGrow for proportion distribution
//   - Direction: Row/Column
//   - Padding
//   - ZIndex
//   - Overflow: Visible/Hidden/Scroll
//
// v1 explicitly does NOT support:
//   - Percentages, Grid, Wrap, CSS Selectors, Animations, Rich Text
type Style struct {
	// Width and Height. -1 means auto (determined by content)
	Width  int
	Height int

	// FlexGrow determines how much the node should grow relative to siblings.
	// 0 means don't grow. Values > 0 are proportional.
	FlexGrow float64

	// Direction: Row (horizontal) or Column (vertical)
	Direction Direction

	// Padding: internal spacing
	Padding Insets

	// ZIndex determines rendering order. Higher values render on top.
	ZIndex int

	// Overflow determines how content that exceeds bounds is handled
	Overflow Overflow

	// AlignItems for cross-axis alignment (flex children only)
	AlignItems Align

	// Justify for main-axis alignment
	Justify Justify

	// Gap between children (flex only)
	Gap int
}

// NewStyle creates a default Style
func NewStyle() Style {
	return Style{
		Width:      -1,
		Height:     -1,
		FlexGrow:   0,
		Direction:  DirectionRow,
		ZIndex:     0,
		Overflow:   OverflowVisible,
		AlignItems: AlignStart,
		Justify:    JustifyStart,
		Gap:        0,
	}
}

// WithWidth sets Width
func (s Style) WithWidth(width int) Style {
	s.Width = width
	return s
}

// WithHeight sets Height
func (s Style) WithHeight(height int) Style {
	s.Height = height
	return s
}

// WithFlexGrow sets FlexGrow
func (s Style) WithFlexGrow(grow float64) Style {
	s.FlexGrow = grow
	return s
}

// WithDirection sets Direction
func (s Style) WithDirection(direction Direction) Style {
	s.Direction = direction
	return s
}

// WithPadding sets Padding
func (s Style) WithPadding(padding Insets) Style {
	s.Padding = padding
	return s
}

// WithZIndex sets ZIndex
func (s Style) WithZIndex(zIndex int) Style {
	s.ZIndex = zIndex
	return s
}

// WithOverflow sets Overflow
func (s Style) WithOverflow(overflow Overflow) Style {
	s.Overflow = overflow
	return s
}

// WithAlignItems sets AlignItems
func (s Style) WithAlignItems(align Align) Style {
	s.AlignItems = align
	return s
}

// WithJustify sets Justify
func (s Style) WithJustify(justify Justify) Style {
	s.Justify = justify
	return s
}

// WithGap sets Gap
func (s Style) WithGap(gap int) Style {
	s.Gap = gap
	return s
}
