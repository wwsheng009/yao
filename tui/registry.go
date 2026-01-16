package tui

import (
	"fmt"
	"sync"

	"github.com/yaoapp/yao/tui/components"
)

// ComponentType represents the type of a component
type ComponentType string

const (
	// Built-in component types
	HeaderComponent    ComponentType = "header"
	TextComponent      ComponentType = "text"
	TableComponent     ComponentType = "table"
	FormComponent      ComponentType = "form"
	InputComponent     ComponentType = "input"
	ViewportComponent  ComponentType = "viewport"
	FooterComponent    ComponentType = "footer"
	ChatComponent      ComponentType = "chat"
	MenuComponent      ComponentType = "menu"
)

// ComponentRenderer is a function that renders a component
type ComponentRenderer func(map[string]interface{}, int) string

// ComponentRegistry holds all registered components
type ComponentRegistry struct {
	mutex      sync.RWMutex
	components map[ComponentType]ComponentRenderer
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

// RegisterBuiltInComponents registers all built-in components
func (r *ComponentRegistry) RegisterBuiltInComponents() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.components[HeaderComponent] = func(props map[string]interface{}, width int) string {
		headerProps := components.ParseHeaderProps(props)
		return components.RenderHeader(headerProps, width)
	}

	r.components[TextComponent] = func(props map[string]interface{}, width int) string {
		textProps := components.ParseTextProps(props)
		return components.RenderText(textProps, width)
	}

	r.components[TableComponent] = func(props map[string]interface{}, width int) string {
		tableProps := components.ParseTableProps(props)
		return components.RenderTable(tableProps, width)
	}

	r.components[FormComponent] = func(props map[string]interface{}, width int) string {
		formProps := components.ParseFormProps(props)
		return components.RenderForm(formProps, width)
	}

	r.components[InputComponent] = func(props map[string]interface{}, width int) string {
		inputProps := components.ParseInputProps(props)
		return components.RenderInput(inputProps, width)
	}

	r.components[ViewportComponent] = func(props map[string]interface{}, width int) string {
		viewportProps := components.ParseViewportProps(props)
		return components.RenderViewport(viewportProps, width)
	}

	r.components[FooterComponent] = func(props map[string]interface{}, width int) string {
		footerProps := components.ParseFooterProps(props)
		return components.RenderFooter(footerProps, width)
	}

	r.components[ChatComponent] = func(props map[string]interface{}, width int) string {
		chatProps := components.ParseChatProps(props)
		return components.RenderChat(chatProps, width)
	}

	r.components[MenuComponent] = func(props map[string]interface{}, width int) string {
		menuProps := components.ParseMenuProps(props)
		return components.RenderMenu(menuProps, width)
	}
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[ComponentType]ComponentRenderer),
	}
}

// RegisterComponent registers a new component renderer
func (r *ComponentRegistry) RegisterComponent(componentType ComponentType, renderer ComponentRenderer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if renderer == nil {
		return fmt.Errorf("renderer cannot be nil")
	}

	r.components[componentType] = renderer
	return nil
}

// GetComponent retrieves a component renderer
func (r *ComponentRegistry) GetComponent(componentType ComponentType) (ComponentRenderer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	renderer, exists := r.components[componentType]
	if !exists {
		return nil, fmt.Errorf("component type '%s' not registered", componentType)
	}

	return renderer, nil
}

// UnregisterComponent removes a component from the registry
func (r *ComponentRegistry) UnregisterComponent(componentType ComponentType) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.components, componentType)
}

// ListComponents returns all registered component types
func (r *ComponentRegistry) ListComponents() []ComponentType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]ComponentType, 0, len(r.components))
	for componentType := range r.components {
		types = append(types, componentType)
	}

	return types
}

// GetComponentRenderer returns the renderer function for a component type
func (r *ComponentRegistry) GetComponentRenderer(componentType ComponentType) (ComponentRenderer, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	renderer, exists := r.components[componentType]
	return renderer, exists
}

