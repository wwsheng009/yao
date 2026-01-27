package component

import (
	"reflect"
	"sync"

	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/tui/core"
)

// Component represents a UI component in the layout.
type Component struct {
	// ID is the unique identifier for this component instance
	ID string `json:"id,omitempty"`

	// Type specifies the component type (e.g., "header", "table", "form")
	Type string `json:"type,omitempty"`

	// Bind specifies the state key to bind data from
	Bind string `json:"bind,omitempty"`

	// Props contains component-specific properties
	Props map[string]interface{} `json:"props,omitempty"`

	// Style contains component-level style properties (position, etc.)
	Style map[string]interface{} `json:"style,omitempty"`

	// Actions maps event names to actions for this component
	Actions map[string]core.Action `json:"actions,omitempty"`

	// Height specifies the component height ("flex" for flexible, or number)
	Height interface{} `json:"height,omitempty"`

	// Width specifies the component width
	Width interface{} `json:"width,omitempty"`

	// Children contains child components (for nested layouts)
	Children []Component `json:"children,omitempty"`

	// Direction specifies layout direction for layout components
	Direction string `json:"direction,omitempty"`
}

// GetID returns the component ID
func (c *Component) GetID() string {
	return c.ID
}

// GetType returns the component type
func (c *Component) GetType() string {
	return c.Type
}

// GetProps returns the component props
func (c *Component) GetProps() map[string]interface{} {
	return c.Props
}

// GetBind returns the component bind attribute
func (c *Component) GetBind() string {
	return c.Bind
}

// GetChildren returns the component children
func (c *Component) GetChildren() []Component {
	return c.Children
}

// GetDirection returns the component direction
func (c *Component) GetDirection() string {
	return c.Direction
}

// GetWidth returns the component width
func (c *Component) GetWidth() interface{} {
	return c.Width
}

// GetHeight returns the component height
func (c *Component) GetHeight() interface{} {
	return c.Height
}

// isRenderConfigChanged checks if two RenderConfig values are different
func isRenderConfigChanged(old, new core.RenderConfig) bool {
	if !reflect.DeepEqual(old.Data, new.Data) {
		return true
	}
	if old.Width != new.Width {
		return true
	}
	if old.Height != new.Height {
		return true
	}
	return false
}

// updateComponentInstanceConfig safely updates component config with validation
func updateComponentInstanceConfig(comp *core.ComponentInstance, renderConfig core.RenderConfig, componentID string) bool {
	updater, ok := comp.Instance.(interface{ UpdateRenderConfig(core.RenderConfig) error })
	if !ok {
		return false
	}

	if !isRenderConfigChanged(comp.LastConfig, renderConfig) {
		return false
	}

	if validator, ok := comp.Instance.(interface{ ValidateConfig(core.RenderConfig) error }); ok {
		if err := validator.ValidateConfig(renderConfig); err != nil {
			log.Warn("Config validation failed for %s: %v", componentID, err)
		}
	}

	if err := updater.UpdateRenderConfig(renderConfig); err != nil {
		log.Warn("Failed to update render config for component %s: %v", componentID, err)
		return false
	}

	comp.LastConfig = renderConfig
	return true
}

// InstanceRegistry manages component instances with lifecycle
type InstanceRegistry struct {
	components map[string]*core.ComponentInstance
	mu         sync.RWMutex
}

// NewInstanceRegistry creates a new InstanceRegistry
func NewInstanceRegistry() *InstanceRegistry {
	return &InstanceRegistry{
		components: make(map[string]*core.ComponentInstance),
	}
}

// GetOrCreate retrieves an existing component instance or creates a new one
func (r *InstanceRegistry) GetOrCreate(
	id string,
	componentType string,
	factory func(config core.RenderConfig, id string) core.ComponentInterface,
	renderConfig core.RenderConfig,
) (*core.ComponentInstance, bool) {
	r.mu.RLock()
	if comp, exists := r.components[id]; exists {
		if comp.Type == componentType {
			if !isRenderConfigChanged(comp.LastConfig, renderConfig) {
				r.mu.RUnlock()
				return comp, false
			}
			updateComponentInstanceConfig(comp, renderConfig, id)
			r.mu.RUnlock()
			return comp, false
		}
		log.Warn("Component type mismatch for %s: %s -> %s, will recreate", id, comp.Type, componentType)
		r.mu.RUnlock()
	} else {
		r.mu.RUnlock()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if comp, exists := r.components[id]; exists {
		updateComponentInstanceConfig(comp, renderConfig, id)
		return comp, false
	}

	instance := factory(renderConfig, id)
	comp := &core.ComponentInstance{
		ID:         id,
		Type:       componentType,
		Instance:   instance,
		LastConfig: renderConfig,
	}
	r.components[id] = comp
	return comp, true
}

// Get retrieves a component instance by ID (read-only)
func (r *InstanceRegistry) Get(id string) (*core.ComponentInstance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	comp, exists := r.components[id]
	return comp, exists
}

// Remove removes a component instance and calls its cleanup method if available
func (r *InstanceRegistry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comp, exists := r.components[id]; exists {
		if cleanup, ok := comp.Instance.(interface{ Cleanup() }); ok {
			cleanup.Cleanup()
		}
		delete(r.components, id)
	}
}

// Clear removes all component instances and calls their cleanup methods
func (r *InstanceRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, comp := range r.components {
		if cleanup, ok := comp.Instance.(interface{ Cleanup() }); ok {
			cleanup.Cleanup()
		}
	}
	r.components = make(map[string]*core.ComponentInstance)
}

// UpdateComponent updates the component instance for a given ID
func (r *InstanceRegistry) UpdateComponent(id string, instance core.ComponentInterface) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comp, exists := r.components[id]; exists {
		comp.Instance = instance
	}
}

// Len returns the number of registered components
func (r *InstanceRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.components)
}

// GetAll returns a copy of all registered components
func (r *InstanceRegistry) GetAll() map[string]*core.ComponentInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*core.ComponentInstance)
	for k, v := range r.components {
		result[k] = v
	}
	return result
}
