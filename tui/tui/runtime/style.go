package runtime

// ===========================================================================
// Style System
// ===========================================================================

// Style represents the styling configuration for a layout node.
type Style struct {
	// Layout properties
	Direction Direction

	// Size
	Width       *int
	Height      *int
	MinWidth    *int
	MinHeight   *int
	MaxWidth    *int
	MaxHeight   *int
	FlexGrow    float64
	FlexShrink  float64
	AspectRatio *float64

	// Spacing
	Padding Insets
	Border  Insets
	Margin  Insets
	Gap     int

	// Alignment
	AlignItems   Align
	AlignSelf    Align
	Justify      Justify
	AlignContent Align

	// Display
	Position Position
	ZIndex   int
	Overflow Overflow

	// Flags
	HasAbsolutePosition bool
}

// NewStyle creates a default style.
func NewStyle() Style {
	return Style{
		Direction:    DirectionColumn,
		AlignItems:   AlignStart,
		Justify:      JustifyStart,
		AlignContent: AlignStart,
		Position:     NewPosition(),
		ZIndex:       0,
		Overflow:     OverflowVisible,
	}
}

// WithDirection sets the direction.
func (s Style) WithDirection(d Direction) Style {
	s.Direction = d
	return s
}

// WithWidth sets the width.
func (s Style) WithWidth(w int) Style {
	s.Width = &w
	return s
}

// WithHeight sets the height.
func (s Style) WithHeight(h int) Style {
	s.Height = &h
	return s
}

// WithMinWidth sets the min width.
func (s Style) WithMinWidth(w int) Style {
	s.MinWidth = &w
	return s
}

// WithMinHeight sets the min height.
func (s Style) WithMinHeight(h int) Style {
	s.MinHeight = &h
	return s
}

// WithMaxWidth sets the max width.
func (s Style) WithMaxWidth(w int) Style {
	s.MaxWidth = &w
	return s
}

// WithMaxHeight sets the max height.
func (s Style) WithMaxHeight(h int) Style {
	s.MaxHeight = &h
	return s
}

// WithFlexGrow sets the flex grow factor.
func (s Style) WithFlexGrow(fg float64) Style {
	s.FlexGrow = fg
	return s
}

// WithFlexShrink sets the flex shrink factor.
func (s Style) WithFlexShrink(fs float64) Style {
	s.FlexShrink = fs
	return s
}

// WithAspectRatio sets the aspect ratio.
func (s Style) WithAspectRatio(ar float64) Style {
	s.AspectRatio = &ar
	return s
}

// WithPadding sets the padding.
func (s Style) WithPadding(p Insets) Style {
	s.Padding = p
	return s
}

// WithBorder sets the border.
func (s Style) WithBorder(b Insets) Style {
	s.Border = b
	return s
}

// WithMargin sets the margin.
func (s Style) WithMargin(m Insets) Style {
	s.Margin = m
	return s
}

// WithGap sets the gap.
func (s Style) WithGap(g int) Style {
	s.Gap = g
	return s
}

// WithAlignItems sets the align items.
func (s Style) WithAlignItems(a Align) Style {
	s.AlignItems = a
	return s
}

// WithAlignSelf sets the align self.
func (s Style) WithAlignSelf(a Align) Style {
	s.AlignSelf = a
	return s
}

// WithJustify sets the justify content.
func (s Style) WithJustify(j Justify) Style {
	s.Justify = j
	return s
}

// WithAlignContent sets the align content.
func (s Style) WithAlignContent(a Align) Style {
	s.AlignContent = a
	return s
}

// WithPosition sets the position.
func (s Style) WithPosition(p Position) Style {
	s.Position = p
	return s
}

// WithZIndex sets the z-index.
func (s Style) WithZIndex(z int) Style {
	s.ZIndex = z
	return s
}

// WithOverflow sets the overflow.
func (s Style) WithOverflow(o Overflow) Style {
	s.Overflow = o
	return s
}

// IsAbsolute returns true if position is absolute.
func (s Style) IsAbsolute() bool {
	return s.Position.Type == PositionAbsolute
}

// GetWidth returns the width or default value.
func (s Style) GetWidth(defaultValue int) int {
	if s.Width != nil {
		return *s.Width
	}
	return defaultValue
}

// GetHeight returns the height or default value.
func (s Style) GetHeight(defaultValue int) int {
	if s.Height != nil {
		return *s.Height
	}
	return defaultValue
}

// GetMinWidth returns the min width or 0.
func (s Style) GetMinWidth() int {
	if s.MinWidth != nil {
		return *s.MinWidth
	}
	return 0
}

// GetMinHeight returns the min height or 0.
func (s Style) GetMinHeight() int {
	if s.MinHeight != nil {
		return *s.MinHeight
	}
	return 0
}

// GetMaxWidth returns the max width or -1 (unbounded).
func (s Style) GetMaxWidth() int {
	if s.MaxWidth != nil {
		return *s.MaxWidth
	}
	return -1
}

// GetMaxHeight returns the max height or -1 (unbounded).
func (s Style) GetMaxHeight() int {
	if s.MaxHeight != nil {
		return *s.MaxHeight
	}
	return -1
}
