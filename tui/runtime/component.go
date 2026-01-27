package runtime

// Renderer is the interface that components must implement to be rendered.
//
// This is a minimal interface that components need to implement for rendering.
// It matches the Bubble Tea Model interface for View().
type Renderer interface {
	// View returns the string representation of the component.
	// This may include ANSI styling codes (e.g., from lipgloss).
	View() string
}

// Sizable is defined in interfaces.go to maintain consistency.
// This interface is kept as a reference - use runtime.Sizable from interfaces.go.

// ComponentRef is a lightweight reference to a component instance.
//
// This allows the runtime to hold component references without depending on
// Bubble Tea or any concrete component implementations.
//
// The Instance field holds any value that implements runtime.Measurable.
// This is typed as interface{} to avoid circular dependencies.
//
// Usage:
//   - DSL/Builder creates ComponentRef from actual component instances
//   - Runtime uses Measurable interface (via type assertion) for layout
//   - Renderer gets the actual Instance for rendering
type ComponentRef struct {
	// ID is the component's unique identifier
	ID string

	// Type is the component type name (e.g., "button", "text")
	Type string

	// Instance is the actual component instance.
	// For layout purposes, it should implement runtime.Measurable.
	Instance interface{}
}

// MeasurableLegacy is the legacy Measurable interface for backward compatibility.
//
// This interface uses the old (maxWidth, maxHeight) signature and is kept
// to support existing components that haven't migrated to the new BoxConstraints-based interface.
//
// New components should implement runtime.Measurable instead.
type MeasurableLegacy interface {
	// Measure returns the component's preferred size given max dimensions.
	Measure(maxWidth, maxHeight int) (width, height int)
}

// Measure attempts to measure the component using either the new or legacy interface.
//
// It tries:
//   1. runtime.Measurable (new, BoxConstraints-based) - preferred
//   2. MeasurableLegacy (old, maxWidth/maxHeight) - for compatibility
//
// Returns zero Size if the component doesn't implement either interface.
func (c *ComponentRef) Measure(bc BoxConstraints) Size {
	if c == nil || c.Instance == nil {
		return Size{}
	}

	// Try new runtime.Measurable interface first
	if measurable, ok := c.Instance.(Measurable); ok {
		return measurable.Measure(bc)
	}

	// Fall back to legacy MeasurableLegacy interface
	if legacy, ok := c.Instance.(MeasurableLegacy); ok {
		maxWidth := bc.MaxWidth
		maxHeight := bc.MaxHeight

		// If constraints are unbounded (MaxWidth < 0), use a reasonable default
		if maxWidth < 0 {
			maxWidth = 80
		}
		if maxHeight < 0 {
			maxHeight = 24
		}

		width, height := legacy.Measure(maxWidth, maxHeight)

		// Apply min constraints
		if width < bc.MinWidth {
			width = bc.MinWidth
		}
		if height < bc.MinHeight {
			height = bc.MinHeight
		}

		// Apply max constraints
		if width > maxWidth {
			width = maxWidth
		}
		if height > maxHeight {
			height = maxHeight
		}

		return Size{Width: width, Height: height}
	}

	return Size{}
}

// NewComponentRef creates a new ComponentRef from the given parameters.
func NewComponentRef(id, compType string, instance interface{}) *ComponentRef {
	return &ComponentRef{
		ID:       id,
		Type:     compType,
		Instance: instance,
	}
}

// NewComponentRefFromInstance creates a ComponentRef from a component instance.
// The instance should implement either runtime.Measurable or MeasurableLegacy.
func NewComponentRefFromInstance(id, compType string, instance interface{}) *ComponentRef {
	return &ComponentRef{
		ID:       id,
		Type:     compType,
		Instance: instance,
	}
}
