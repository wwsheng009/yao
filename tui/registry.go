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

// RegisterBuiltInComponents registers all built-in components
func (r *ComponentRegistry) RegisterBuiltInComponents() {
	// Register header component (assuming it exists in components package)
	r.RegisterComponent(HeaderComponent, func(props map[string]interface{}, width int) string {
		// Implementation would depend on actual header component
		title, _ := props["title"].(string)
		return title
	})

	// Register text component (assuming it exists in components package)
	r.RegisterComponent(TextComponent, func(props map[string]interface{}, width int) string {
		// Implementation would depend on actual text component
		content, _ := props["content"].(string)
		return content
	})

	// Register table component
	r.RegisterComponent(TableComponent, func(props map[string]interface{}, width int) string {
		tableProps := components.ParseTableProps(props)
		return components.RenderTable(tableProps, width)
	})

	// Register form component
	r.RegisterComponent(FormComponent, func(props map[string]interface{}, width int) string {
		formProps := components.ParseFormProps(props)
		return components.RenderForm(formProps, width)
	})

	// Register input component
	r.RegisterComponent(InputComponent, func(props map[string]interface{}, width int) string {
		inputProps := components.ParseInputProps(props)
		return components.RenderInput(inputProps, width)
	})

	// Register viewport component
	r.RegisterComponent(ViewportComponent, func(props map[string]interface{}, width int) string {
		viewportProps := components.ParseViewportProps(props)
		return components.RenderViewport(viewportProps, width)
	})

	// Register footer component
	r.RegisterComponent(FooterComponent, func(props map[string]interface{}, width int) string {
		footerProps := components.ParseFooterProps(props)
		return components.RenderFooter(footerProps, width)
	})

	// Register chat component
	r.RegisterComponent(ChatComponent, func(props map[string]interface{}, width int) string {
		chatProps := components.ParseChatProps(props)
		return components.RenderChat(chatProps, width)
	})
}