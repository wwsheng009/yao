package component

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// ==============================================================================
// BaseContainer Implementation (V3)
// ==============================================================================
// BaseContainer provides a default implementation of the Container interface.
// Container and Layout interfaces are defined in capabilities.go.

// BaseContainer is the default container implementation.
type BaseContainer struct {
	*BaseComponent
	children []Node
	layout   Layout
}

// NewBaseContainer creates a new BaseContainer.
func NewBaseContainer(typ string) *BaseContainer {
	return &BaseContainer{
		BaseComponent: NewBaseComponent(typ),
		children:      make([]Node, 0),
	}
}

// ============================================================================
// Container Interface Implementation
// ============================================================================

// Add adds a child component.
func (c *BaseContainer) Add(child Node) {
	c.children = append(c.children, child)

	// Try MountableWithContext first (if component supports it and we have context)
	if mountable, ok := child.(MountableWithContext); ok {
		ctx := c.GetComponentContext()
		if ctx != nil {
			// Has context: use MountWithContext
			mountable.MountWithContext(c, ctx)
		} else {
			// No context: fall back to regular Mount
			if m, ok := child.(Mountable); ok {
				m.Mount(c)
			}
		}
	} else if mountable, ok := child.(Mountable); ok {
		mountable.Mount(c)
	}
}

// Remove removes a child component.
func (c *BaseContainer) Remove(child Node) {
	for i, ch := range c.children {
		if ch == child {
			c.children = append(c.children[:i], c.children[i+1:]...)
			if mountable, ok := child.(Mountable); ok {
				mountable.Unmount()
			}
			break
		}
	}
}

// RemoveAt removes the child at the given index.
func (c *BaseContainer) RemoveAt(index int) {
	if index >= 0 && index < len(c.children) {
		child := c.children[index]
		c.children = append(c.children[:index], c.children[index+1:]...)
		if mountable, ok := child.(Mountable); ok {
			mountable.Unmount()
		}
	}
}

// GetChildren returns the list of children.
func (c *BaseContainer) GetChildren() []Node {
	return c.children
}

// GetChild returns the child at the given index.
func (c *BaseContainer) GetChild(index int) Node {
	if index >= 0 && index < len(c.children) {
		return c.children[index]
	}
	return nil
}

// ChildCount returns the number of children.
func (c *BaseContainer) ChildCount() int {
	return len(c.children)
}

// SetLayout sets the layout algorithm.
func (c *BaseContainer) SetLayout(layout Layout) {
	c.layout = layout
}

// GetLayout returns the current layout algorithm.
func (c *BaseContainer) GetLayout() Layout {
	return c.layout
}

// ============================================================================
// Measurable Interface Implementation
// ============================================================================

// Measure calculates the ideal size for this container using runtime.Measurable.
func (c *BaseContainer) Measure(maxWidth, maxHeight int) (width, height int) {
	if c.layout != nil {
		return c.layout.Measure(c, maxWidth, maxHeight)
	}
	// Default: calculate total size of all children
	// Convert to BoxConstraints for runtime.Measurable
	bc := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: 0,
		MaxHeight: maxHeight,
	}
	for _, child := range c.children {
		if measurable, ok := child.(runtime.Measurable); ok {
			size := measurable.Measure(bc)
			if size.Width > width {
				width = size.Width
			}
			height += size.Height
		}
	}
	return width, height
}
