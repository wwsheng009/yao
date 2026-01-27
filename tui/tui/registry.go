package tui

import (
	"fmt"
	"sync"

	"github.com/yaoapp/yao/tui/tui/component"
	"github.com/yaoapp/yao/tui/tui/core"
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
	CRUDComponent       ComponentType = "crud"
)

// ComponentFactory creates a component instance
// Accepts RenderConfig for unified rendering approach
// The config parameter contains the initial properties for the component
type ComponentFactory func(config core.RenderConfig, id string) core.ComponentInterface

// ComponentRegistry holds all registered component factories
type ComponentRegistry struct {
	mutex          sync.RWMutex
	factories      map[ComponentType]ComponentFactory
	focusableTypes map[ComponentType]bool // Tracks which component types are focusable
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
		return component.NewHeaderComponent(config, id)
	}

	r.factories[TextComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewTextComponent(config, id)
	}

	r.factories[FooterComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewFooterComponent(config, id)
	}

	r.factories[InputComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewInputComponent(config, id)
	}

	r.factories[TextareaComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewTextareaComponent(config, id)
	}

	r.factories[MenuComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewMenuComponent(config, id)
	}

	r.factories[ChatComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewChatComponent(config, id)
	}

	r.factories[CursorComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewCursorComponent(config, id)
	}

	r.factories[FilePickerComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewFilePickerComponent(config, id)
	}

	r.factories[HelpComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewHelpComponent(config, id)
	}

	r.factories[KeyComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewKeyComponent(config, id)
	}

	r.factories[ListComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewListComponent(config, id)
	}

	r.factories[PaginatorComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewPaginatorComponent(config, id)
	}

	r.factories[ProgressComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewProgressComponent(config, id)
	}

	r.factories[SpinnerComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewSpinnerComponent(config, id)
	}

	r.factories[TimerComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewTimerComponent(config, id)
	}

	r.factories[StopwatchComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewStopwatchComponent(config, id)
	}

	r.factories[TableComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewTableComponent(config, id)
	}

	r.factories[FormComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewFormComponent(config, id)
	}

	r.factories[ViewportComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewViewportComponent(config, id)
	}

	r.factories[CRUDComponent] = func(config core.RenderConfig, id string) core.ComponentInterface {
		return component.NewCRUDComponent(config, id)
	}

	// Register focusable component types
	// Components can be focusable if they support user interaction
	r.focusableTypes[InputComponent] = true
	r.focusableTypes[TextareaComponent] = true
	r.focusableTypes[MenuComponent] = true
	r.focusableTypes[FormComponent] = true
	r.focusableTypes[TableComponent] = true
	r.focusableTypes[ChatComponent] = true
	r.focusableTypes[FilePickerComponent] = true
	r.focusableTypes[CRUDComponent] = true
	r.focusableTypes[ListComponent] = true
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		factories:      make(map[ComponentType]ComponentFactory),
		focusableTypes: make(map[ComponentType]bool),
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

// RegisterFocusableComponent marks a component type as focusable
// This allows components to declare their focus capability during registration
func (r *ComponentRegistry) RegisterFocusableComponent(componentType ComponentType) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.focusableTypes[componentType] = true
}

// UnregisterFocusableComponent removes focusable flag from a component type
func (r *ComponentRegistry) UnregisterFocusableComponent(componentType ComponentType) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.focusableTypes, componentType)
}

// IsFocusable checks if a component type is focusable
func (r *ComponentRegistry) IsFocusable(componentType ComponentType) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.focusableTypes[componentType]
}

// GetFocusableComponentTypes returns all focusable component types
func (r *ComponentRegistry) GetFocusableComponentTypes() []ComponentType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]ComponentType, 0, len(r.focusableTypes))
	for compType := range r.focusableTypes {
		types = append(types, compType)
	}
	return types
}
