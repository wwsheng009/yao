package runtime

// Measurable is an interface that allows components to report their intrinsic size.
//
// This is the key interface for the Measure phase. Components implement this
// to participate in layout calculations.
//
// The Measure method receives BoxConstraints from the parent and returns
// the component's preferred size.
//
// Note: The Measure phase should ONLY calculate size, never set position.
// Position calculation happens in the Layout phase.
type Measurable interface {
	// Measure returns the component's preferred size given the constraints.
	//
	// The constraints come from the parent container. The component should
	// return a Size that:
	//   - Is within the constraints (bc.MinWidth <= size.Width <= bc.MaxWidth)
	//   - Represents its ideal rendering size
	//   - Depends only on content and constraints, not on position
	//
	// If a component wants to fill all available space, it can return
	// bc.MaxWidth and bc.MaxHeight.
	//
	// This method must NOT modify any X/Y coordinates.
	Measure(c BoxConstraints) Size
}
