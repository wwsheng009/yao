package runtime

// Core types for Yao TUI Runtime v1

// BoxConstraints defines the min/max constraints for layout
type BoxConstraints struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

// NewBoxConstraints creates a new BoxConstraints
func NewBoxConstraints(minWidth, maxWidth, minHeight, maxHeight int) BoxConstraints {
	return BoxConstraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// IsTight returns true if width and height are both fixed
func (bc BoxConstraints) IsTight() bool {
	return bc.MinWidth == bc.MaxWidth && bc.MinHeight == bc.MaxHeight
}

// Constrain clamps a width and height within the constraints
func (bc BoxConstraints) Constrain(width, height int) (int, int) {
	w := clamp(width, bc.MinWidth, bc.MaxWidth)
	h := clamp(height, bc.MinHeight, bc.MaxHeight)
	return w, h
}

// Loosen returns a new BoxConstraints with min values set to 0
func (bc BoxConstraints) Loosen() BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  bc.MaxWidth,
		MinHeight: 0,
		MaxHeight: bc.MaxHeight,
	}
}

// clamp clamps a value between min and max
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Size represents a 2D size
type Size struct {
	Width  int
	Height int
}

// ConstraintsAlias is an alias for backward compatibility with legacy code
// Legacy code uses Constraints with MaxW and MaxH fields
type Constraints = BoxConstraints

// NewConstraints creates a Constraints with only MaxW and MaxH set
func NewConstraints(maxW, maxH int) Constraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  maxW,
		MinHeight: 0,
		MaxHeight: maxH,
	}
}

// IsInfinite checks if constraint is infinite (-1 means infinite)
func (c Constraints) HasInfiniteWidth() bool {
	return c.MaxWidth < 0
}

func (c Constraints) HasInfiniteHeight() bool {
	return c.MaxHeight < 0
}
