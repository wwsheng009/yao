package component

// Node is the base interface for all components in the framework.
// It provides the fundamental identity and type information required
// by the runtime and framework systems.
type Node interface {
	// ID returns the unique identifier of the component.
	// This ID is used for focus management, event routing, and state persistence.
	ID() string

	// Type returns the component type identifier (e.g., "button", "input", "box").
	// This can be used for styling, serialization, and debugging.
	Type() string
}
