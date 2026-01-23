package registry

import (
	"fmt"

	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/ui/components"
)

// Factory is a function that creates a new component instance.
type Factory func() runtime.Component

// Registry holds component factories for creating component instances.
// This is the native Runtime equivalent of the legacy component registry.
type Registry struct {
	factories map[string]Factory
}

// DefaultRegistry is the global default component registry.
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new component registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register registers a component factory for a given type.
func (r *Registry) Register(componentType string, factory Factory) {
	r.factories[componentType] = factory
}

// RegisterDefault registers a component in the default registry.
func RegisterDefault(componentType string, factory Factory) {
	DefaultRegistry.Register(componentType, factory)
}

// Create creates a new component instance of the given type.
// Returns an error if the type is not registered.
func (r *Registry) Create(componentType string) (runtime.Component, error) {
	factory, exists := r.factories[componentType]
	if !exists {
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}
	return factory(), nil
}

// MustCreate creates a component or panics if the type is unknown.
func (r *Registry) MustCreate(componentType string) runtime.Component {
	component, err := r.Create(componentType)
	if err != nil {
		panic(err)
	}
	return component
}

// CreateDefault creates a component using the default registry.
func CreateDefault(componentType string) (runtime.Component, error) {
	return DefaultRegistry.Create(componentType)
}

// MustCreateDefault creates a component using the default registry or panics.
func MustCreateDefault(componentType string) runtime.Component {
	return DefaultRegistry.MustCreate(componentType)
}

// IsRegistered checks if a component type is registered.
func (r *Registry) IsRegistered(componentType string) bool {
	_, exists := r.factories[componentType]
	return exists
}

// Types returns all registered component types.
func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// RegisterBuiltinComponents registers all built-in Runtime components.
func RegisterBuiltinComponents(r *Registry) {
	// Layout components
	r.Register("row", func() runtime.Component {
		return components.NewRow()
	})
	r.Register("column", func() runtime.Component {
		return components.NewColumn()
	})
	r.Register("flex", func() runtime.Component {
		return components.NewFlex()
	})

	// Basic components
	r.Register("text", func() runtime.Component {
		return components.NewTextComponent("")
	})
	r.Register("header", func() runtime.Component {
		return components.NewHeader("")
	})
	r.Register("footer", func() runtime.Component {
		return components.NewFooter("")
	})

	// Interactive components
	r.Register("input", func() runtime.Component {
		return components.NewInput()
	})
	r.Register("button", func() runtime.Component {
		return components.NewButton("")
	})
}

// init registers all built-in components in the default registry.
func init() {
	RegisterBuiltinComponents(DefaultRegistry)
}
