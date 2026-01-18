package tui

import (
	"fmt"
	"sync"

	"github.com/yaoapp/yao/tui/components"
	"github.com/yaoapp/yao/tui/core"
)

// ComponentType represents the type of a component
type ComponentType string

const (
	// Built-in component types
	HeaderComponent     ComponentType = "header"
	TextComponent       ComponentType = "text"
	TableComponent      ComponentType = "table"
	FormComponent       ComponentType = "form"
	InputComponent      ComponentType = "input"
	ViewportComponent   ComponentType = "viewport"
	FooterComponent     ComponentType = "footer"
	ChatComponent       ComponentType = "chat"
	MenuComponent       ComponentType = "menu"
	TimerComponent      ComponentType = "timer"
	StopwatchComponent  ComponentType = "stopwatch"
	FilePickerComponent ComponentType = "filepicker"
	HelpComponent       ComponentType = "help"
	KeyComponent        ComponentType = "key"
	CursorComponent     ComponentType = "cursor"
	ListComponent       ComponentType = "list"
	PaginatorComponent  ComponentType = "paginator"
	ProgressComponent   ComponentType = "progress"
	SpinnerComponent    ComponentType = "spinner"
	TextareaComponent   ComponentType = "textarea"
	CRUDComponent      ComponentType = "crud"
)

// ComponentFactory creates a component instance
// Accepts RenderConfig for unified rendering approach
// The config parameter contains the initial properties for the component
type ComponentFactory func(config core.RenderConfig, id string) core.ComponentInterface

// ComponentRegistry holds all registered component factories
type ComponentRegistry struct {
	mutex     sync.RWMutex
	factories map[ComponentType]ComponentFactory
}

// Global registry instance
var globalRegistry *ComponentRegistry
var registryOnce sync.Once

// GetGlobalRegistry returns the global component registry
func GetGlobalRegistry() *ComponentRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewComponentRegistry()
		// Register built-in components
		globalRegistry.RegisterBuiltInComponents()
	})
	return globalRegistry
}

// RegisterBuiltInComponents registers all built-in component factories.
// Uses unified signature: func(config core.RenderConfig, id string) ComponentInterface
// Props are parsed during the component's Render() method call.
func (r *ComponentRegistry) RegisterBuiltInComponents() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// All components use the unified factory signature
	// Components will parse props in their Render() methods

	r.factories[HeaderComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewHeaderComponent(config, id)
	}

	r.factories[TextComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewTextComponent(config, id)
	}

	r.factories[FooterComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewFooterComponent(config, id)
	}

	r.factories[InputComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewInputComponent(config, id)
	}

	r.factories[TextareaComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewTextareaComponent(config, id)
	}

	r.factories[MenuComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewMenuComponent(config, id)
	}

	r.factories[ChatComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewChatComponent(config, id)
	}

	r.factories[CursorComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewCursorComponent(config, id)
	}

	r.factories[FilePickerComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewFilePickerComponent(config, id)
	}

	r.factories[HelpComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewHelpComponent(config, id)
	}

	r.factories[KeyComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewKeyComponent(config, id)
	}

	r.factories[ListComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewListComponent(config, id)
	}

	r.factories[PaginatorComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewPaginatorComponent(config, id)
	}

	r.factories[ProgressComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewProgressComponent(config, id)
	}

	r.factories[SpinnerComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewSpinnerComponent(config, id)
	}

	r.factories[TimerComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewTimerComponent(config, id)
	}

	r.factories[StopwatchComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewStopwatchComponent(config, id)
	}

	r.factories[TableComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewTableComponent(config, id)
	}

	r.factories[FormComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewFormComponent(config, id)
	}

	r.factories[ViewportComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewViewportComponent(config, id)
	}

	r.factories[CRUDComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return components.NewCRUDComponentWrapper(config, id)
	}
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		factories: make(map[ComponentType]ComponentFactory),
	}
}

// RegisterComponent registers a new component factory
func (r *ComponentRegistry) RegisterComponent(componentType ComponentType, factory ComponentFactory) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.factories[componentType] = factory
	return nil
}

// GetComponent retrieves a component factory
func (r *ComponentRegistry) GetComponent(componentType ComponentType) (ComponentFactory, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.factories[componentType]
	if !exists {
		return nil, fmt.Errorf("component type '%s' not registered", componentType)
	}

	return factory, nil
}

// UnregisterComponent removes a component from the registry
func (r *ComponentRegistry) UnregisterComponent(componentType ComponentType) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.factories, componentType)
}

// ListComponents returns all registered component types
func (r *ComponentRegistry) ListComponents() []ComponentType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]ComponentType, 0, len(r.factories))
	for componentType := range r.factories {
		types = append(types, componentType)
	}

	return types
}

// GetComponentFactory returns the factory function for a component type
func (r *ComponentRegistry) GetComponentFactory(componentType ComponentType) (ComponentFactory, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.factories[componentType]
	return factory, exists
}
