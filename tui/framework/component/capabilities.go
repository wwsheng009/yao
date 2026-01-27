package component

import (
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Capability Interfaces (V3) - Built on Runtime Interfaces
// ==============================================================================
// Framework component interfaces are built on top of runtime interfaces.
// This ensures consistency and allows components to work with both
// framework and runtime systems.
//
// Framework adds framework-specific capabilities (like Mountable with context)
// on top of the minimal runtime interfaces.

// Node is an alias for runtime.Node for convenience.
// All components must implement this minimal interface.
type Node = runtime.Node

// Positionable is an alias for runtime.Positionable.
type Positionable = runtime.Positionable

// Sizable is an alias for runtime.Sizable.
type Sizable = runtime.Sizable

// Located is an alias for runtime.Located.
type Located = runtime.Located

// Measurable is an alias for runtime.Measurable.
type Measurable = runtime.Measurable

// ComponentNode is an alias for runtime.ComponentNode.
type ComponentNode = runtime.ComponentNode

// InteractiveComponent is an alias for runtime.InteractiveComponent.
type InteractiveComponent = runtime.InteractiveComponent

// Focusable is an alias for runtime.Focusable.
type Focusable = runtime.Focusable

// Visible is an alias for runtime.Visible.
type Visible = runtime.Visible

// =============================================================================
// Framework-Specific Interfaces
// =============================================================================

// Paintable extends runtime with framework-specific painting.
// Components implement this to draw into the Buffer.
type Paintable interface {
	Node

	// Paint renders the component to the buffer.
	// ctx contains the painting context (position, available size, etc.)
	// buf is the virtual canvas to draw into.
	Paint(ctx PaintContext, buf *paint.Buffer)
}

// PaintContext is the painting context for framework components.
// It extends the runtime paint context with framework-specific features.
type PaintContext struct {
	// Available size for the component
	AvailableWidth  int
	AvailableHeight int

	// Component position (relative to parent)
	X int
	Y int

	// Scroll offset
	OffsetX int
	OffsetY int

	// Z-index for layering
	ZIndex int

	// Clip region (optional)
	ClipRect *runtime.Rect
}

// NewPaintContext creates a new PaintContext with the given dimensions.
func NewPaintContext(x, y, width, height int) PaintContext {
	return PaintContext{
		X:               x,
		Y:               y,
		AvailableWidth:  width,
		AvailableHeight: height,
	}
}

// =============================================================================
// Mountable with Context
// =============================================================================

// Mountable extends runtime with framework-specific mounting.
// Framework components can receive a ComponentContext when mounted.
type Mountable interface {
	Node

	// Mount attaches the component to a parent.
	Mount(parent Container)

	// Unmount detaches the component from its parent.
	Unmount()

	// IsMounted returns true if the component is currently mounted.
	IsMounted() bool
}

// MountableWithContext is an extended Mountable that receives context.
// Components implement this to access runtime resources (like dirty marking).
type MountableWithContext interface {
	Node

	// MountWithContext attaches the component and receives component context.
	MountWithContext(parent Container, ctx *ComponentContext)
}

// =============================================================================
// Container Interface
// =============================================================================

// Container is a component that can hold child components.
type Container interface {
	Node
	Mountable

	// Child management
	Add(child Node)
	Remove(child Node)
	RemoveAt(index int)
	GetChildren() []Node
	GetChild(index int) Node
	ChildCount() int

	// Layout
	SetLayout(layout Layout)
	GetLayout() Layout
}

// Layout is the interface for layout algorithms.
type Layout interface {
	// Measure calculates the ideal size for the container.
	Measure(container Container, availableWidth, availableHeight int) (width, height int)

	// Layout positions children within the container.
	Layout(container Container, x, y, width, height int)

	// Invalidate is called when the layout needs to be recalculated.
	Invalidate()
}

// =============================================================================
// Action Interface
// =============================================================================

// ActionTarget handles semantic actions (not raw key events).
// Components implement this to respond to high-level actions.
type ActionTarget interface {
	Node

	// HandleAction processes a semantic action.
	// Returns true if the action was handled, false to continue propagation.
	HandleAction(a Action) bool
}

// Action represents a semantic action (like "confirm", "cancel").
type Action interface {
	// Type returns the action type.
	Type() ActionType

	// Payload returns the action data.
	Payload() interface{}

	// Source returns where the action originated.
	Source() string

	// Target returns the intended target of the action.
	Target() string
}

// ActionType is the type of action.
type ActionType string

// =============================================================================
// Framework Component Interface Combinations
// =============================================================================

// FrameworkComponent combines the most common framework interfaces.
// Most components should implement this.
type FrameworkComponent interface {
	Node
	Mountable
	Measurable
	Paintable
}

// FrameworkInteractiveComponent extends FrameworkComponent with interactivity.
// Interactive components (Button, Input) should implement this.
type FrameworkInteractiveComponent interface {
	FrameworkComponent
	Focusable
	ActionTarget
}

// FrameworkContainerComponent combines Container with layout capabilities.
// Container components (Box, Flex) should implement this.
type FrameworkContainerComponent interface {
	Container
	Measurable
	Paintable
}

// =============================================================================
// Other Capability Interfaces
// =============================================================================

// Scrollable is for components that support scrolling.
type Scrollable interface {
	Node

	// ScrollTo moves to an absolute position.
	ScrollTo(x, y int)

	// ScrollBy moves relative to current position.
	ScrollBy(dx, dy int)

	// GetScrollPosition returns the current scroll position.
	GetScrollPosition() (x, y int)
}

// Validatable is for components that can validate their state.
type Validatable interface {
	Node

	// Validate checks the component state.
	Validate() error

	// IsValid returns true if the component is in a valid state.
	IsValid() bool
}

// =============================================================================
// Common Types
// =============================================================================

// TextAlign for text alignment.
type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)
