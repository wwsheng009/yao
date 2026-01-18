package tui

import (
	"reflect"
	"sync"

	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// isRenderConfigChanged checks if two RenderConfig values are different
func isRenderConfigChanged(old, new core.RenderConfig) bool {
	// Compare Data field using reflection
	if !reflect.DeepEqual(old.Data, new.Data) {
		return true
	}
	
	// Compare Width
	if old.Width != new.Width {
		return true
	}
	
	// Compare Height
	if old.Height != new.Height {
		return true
	}
	
	// No significant changes detected
	return false
}

// ComponentInstanceRegistry manages component instances with lifecycle
type ComponentInstanceRegistry struct {
	components map[string]*core.ComponentInstance
	mu         sync.RWMutex
}

// NewComponentInstanceRegistry creates a new ComponentInstanceRegistry
func NewComponentInstanceRegistry() *ComponentInstanceRegistry {
	return &ComponentInstanceRegistry{
		components: make(map[string]*core.ComponentInstance),
	}
}

// GetOrCreate retrieves an existing component instance or creates a new one
// Returns (componentInstance, isNew) where isNew is true if a new instance was created
func (r *ComponentInstanceRegistry) GetOrCreate(
	id string,
	componentType string,
	factory func(config core.RenderConfig, id string) core.ComponentInterface,
	renderConfig core.RenderConfig,
) (*core.ComponentInstance, bool) {
	// Try read lock first
	r.mu.RLock()
	if comp, exists := r.components[id]; exists {
		// Check if the existing component is of the same type
		if comp.Type == componentType {
			log.Trace("GetOrCreate: component=%s, type=%s, updating existing", id, componentType)
			// Update existing instance's render config only if config changed
			if updater, ok := comp.Instance.(interface{ UpdateRenderConfig(core.RenderConfig) error }); ok {
				// Check if config has changed before updating
				if !isRenderConfigChanged(comp.LastConfig, renderConfig) {
					log.Trace("GetOrCreate: config unchanged for %s, skipping update", id)
					r.mu.RUnlock()
					return comp, false
				}

				// Validate config before updating
				if validator, ok := comp.Instance.(interface{ ValidateConfig(core.RenderConfig) error }); ok {
					if err := validator.ValidateConfig(renderConfig); err != nil {
						log.Warn("Config validation failed for %s: %v", id, err)
					}
				}
				if err := updater.UpdateRenderConfig(renderConfig); err != nil {
					log.Warn("Failed to update render config for component %s: %v", id, err)
				}
				// Store the new config for future comparison
				comp.LastConfig = renderConfig
			}
			r.mu.RUnlock()
			return comp, false // false means existing instance
		}
		log.Warn("Component type mismatch for %s: %s -> %s, will recreate", id, comp.Type, componentType)
		r.mu.RUnlock()
	} else {
		log.Trace("GetOrCreate: component=%s not found, will create new", id)
		r.mu.RUnlock()
	}

	// Acquire write lock to create new instance
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check locking
	if comp, exists := r.components[id]; exists {
		if updater, ok := comp.Instance.(interface{ UpdateRenderConfig(core.RenderConfig) error }); ok {
			// Check if config has changed before updating
			if isRenderConfigChanged(comp.LastConfig, renderConfig) {
				// Validate config before updating
				if validator, ok := comp.Instance.(interface{ ValidateConfig(core.RenderConfig) error }); ok {
					if err := validator.ValidateConfig(renderConfig); err != nil {
						log.Warn("Config validation failed for %s: %v", id, err)
					}
				}
				if err := updater.UpdateRenderConfig(renderConfig); err != nil {
					log.Warn("Failed to update render config for component %s: %v", id, err)
				}
				// Store the new config for future comparison
				comp.LastConfig = renderConfig
			}
		}
		return comp, false
	}

	// Create new instance
	instance := factory(renderConfig, id)
	comp := &core.ComponentInstance{
		ID:        id,
		Type:      componentType,
		Instance:  instance,
		LastConfig: renderConfig,
	}
	r.components[id] = comp
	return comp, true // true means newly created
}

// Get retrieves a component instance by ID (read-only)
func (r *ComponentInstanceRegistry) Get(id string) (*core.ComponentInstance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	comp, exists := r.components[id]
	return comp, exists
}

// Remove removes a component instance and calls its cleanup method if available
func (r *ComponentInstanceRegistry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comp, exists := r.components[id]; exists {
		// Call cleanup method if available
		if cleanup, ok := comp.Instance.(interface{ Cleanup() }); ok {
			cleanup.Cleanup()
		}
		delete(r.components, id)
	}
}

// Clear removes all component instances and calls their cleanup methods
func (r *ComponentInstanceRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Cleanup all components
	for _, comp := range r.components {
		if cleanup, ok := comp.Instance.(interface{ Cleanup() }); ok {
			cleanup.Cleanup()
		}
	}
	r.components = make(map[string]*core.ComponentInstance)
}

// UpdateComponent updates the component instance for a given ID
func (r *ComponentInstanceRegistry) UpdateComponent(id string, instance core.ComponentInterface) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comp, exists := r.components[id]; exists {
		comp.Instance = instance
	}
}

// Len returns the number of registered components
func (r *ComponentInstanceRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.components)
}

// GetAll returns a copy of all registered components
func (r *ComponentInstanceRegistry) GetAll() map[string]*core.ComponentInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*core.ComponentInstance)
	for k, v := range r.components {
		result[k] = v
	}
	return result
}
