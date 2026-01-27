package runtime

// =============================================================================
// Unified Component Interfaces
// =============================================================================
// These interfaces provide the foundation for components in the TUI framework.
// Both runtime and framework packages should use these interfaces.
//
// Design principles:
// - Minimal interfaces: each interface represents a single capability
// - Composable: components implement only the interfaces they need
// - No external dependencies: pure Go interfaces

// Node is the minimal interface that all components must implement.
// It provides identification and type information.
type Node interface {
	// ID returns the component's unique identifier.
	ID() string

	// Type returns the component's type name (e.g., "button", "text").
	Type() string
}

// Positionable is an interface for components that have position.
type Positionable interface {
	// GetPosition returns the component's position (x, y).
	GetPosition() (x, y int)

	// SetPosition sets the component's position.
	SetPosition(x, y int)
}

// Sizable is an interface for components that have size.
type Sizable interface {
	// GetSize returns the component's size (width, height).
	GetSize() (width, height int)

	// SetSize sets the component's size.
	SetSize(width, height int)
}

// Located combines Node, Positionable, and Sizable.
type Located interface {
	Node
	Positionable
	Sizable
}

// Measurable is an interface for components that can report their ideal size.
// This is already defined in measurable.go, but documented here for completeness.
//
// Measure returns the component's preferred size given constraints.
// The component should return a size within the constraints.

// Paintable is an interface for components that can render themselves.
// This is defined in the paint package to avoid circular dependencies.
//
// Paint(ctx PaintContext, buf *Buffer) renders the component.

// =============================================================================
// Tree Structure Interfaces
// =============================================================================

// Parent is an interface for components that can contain children.
type Parent interface {
	Node

	// Children returns the component's child nodes.
	Children() []Node

	// ChildCount returns the number of children.
	ChildCount() int
}

// Child is an interface for components that can be added to a parent.
type Child interface {
	Node
	// Can be added/removed from parent
}

// =============================================================================
// Lifecycle Interfaces
// =============================================================================

// Mountable is an interface for components that can be mounted to a parent.
type Mountable interface {
	Node

	// Mount attaches the component to a parent.
	Mount(parent Parent)

	// Unmount detaches the component from its parent.
	Unmount()

	// IsMounted returns true if the component is currently mounted.
	IsMounted() bool
}

// =============================================================================
// State Interfaces
// =============================================================================

// Visible is an interface for components that can be shown/hidden.
type Visible interface {
	// IsVisible returns true if the component should be rendered.
	IsVisible() bool

	// SetVisible sets the component's visibility.
	SetVisible(visible bool)
}

// Focusable is an interface for components that can receive keyboard focus.
type Focusable interface {
	Node

	// FocusID returns a unique identifier for focus management.
	FocusID() string

	// OnFocus is called when the component receives focus.
	OnFocus()

	// OnBlur is called when the component loses focus.
	OnBlur()

	// IsFocused returns true if the component currently has focus.
	IsFocused() bool
}

// =============================================================================
// Convenience Interface Combinations
// =============================================================================

// ComponentNode combines the most common component interfaces.
// Most static components (Text, Image) should implement this.
type ComponentNode interface {
	Node
	Positionable
	Sizable
	Measurable
}

// InteractiveComponent extends ComponentNode with interactivity.
// Interactive components (Button, Input) should implement this.
type InteractiveComponent interface {
	ComponentNode
	Focusable
}

// ContainerComponent combines Parent with layout capabilities.
// Container components (Box, Flex) should implement this.
type ContainerComponent interface {
	Parent
	Measurable
}

// =============================================================================
// Geometry Types (re-exported for convenience)
// =============================================================================
// Point is defined in node.go as part of LayoutNode
// Rect is defined in runtime.go

// Point represents a 2D point.
type Point struct {
	X int
	Y int
}
