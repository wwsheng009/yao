package runtime

// Style represents declarative layout intent (v1 simplified)
//
// v1 supports:
//   - Width/Height with -1 for auto
//   - Width/Height with negative values -2 to -101 for percentages (2% to 100%)
//   - FlexGrow for proportion distribution
//   - Direction: Row/Column
//   - Padding
//   - Border (physical spacing)
//   - ZIndex
//   - Overflow: Visible/Hidden/Scroll
//
// v1 explicitly does NOT support:
//   - Grid, Wrap, CSS Selectors, Animations, Rich Text

// Percentage encoding: -2 to -101 represents 2% to 101%
const (
	AutoSize   = -1 // Size determined by content
	MinPercent = -2 // 2%
	MaxPercent = -101 // 101% (allows >100%)
)
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

	// Border: physical border spacing (v1.1)
	// Border takes up physical space around the content
	Border Insets

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
		Padding:    Insets{0, 0, 0, 0},
		Border:     Insets{0, 0, 0, 0},
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

// WithBorder sets Border
func (s Style) WithBorder(border Insets) Style {
	s.Border = border
	return s
}

// WithBorderWidth sets uniform border width on all sides
func (s Style) WithBorderWidth(width int) Style {
	s.Border = Insets{Top: width, Right: width, Bottom: width, Left: width}
	return s
}

// WithWidthPercent sets width as a percentage of parent width
// percent: 50 means 50%, 100 means 100%
func (s Style) WithWidthPercent(percent int) Style {
	// Encode as negative value: -50 means 50%
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	s.Width = -percent
	return s
}

// WithHeightPercent sets height as a percentage of parent height
// percent: 50 means 50%, 100 means 100%
func (s Style) WithHeightPercent(percent int) Style {
	// Encode as negative value: -50 means 50%
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	s.Height = -percent
	return s
}

// IsPercent checks if a size value is a percentage
func IsPercent(size int) bool {
	return size <= MinPercent && size >= MaxPercent
}

// ResolvePercent resolves a percentage size to actual value
// Returns the resolved size and whether it was a percentage
func ResolvePercent(size int, parentSize int) (int, bool) {
	if IsPercent(size) {
		percent := -size
		return parentSize * percent / 100, true
	}
	return size, false
}
