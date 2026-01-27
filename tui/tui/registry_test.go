package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/core"
)

func TestComponentRegistry(t *testing.T) {
	// Create a new registry
	registry := NewComponentRegistry()

	// Test registering a component
	err := registry.RegisterComponent(ComponentType("test"), func(config core.RenderConfig, id string) core.ComponentInterface {
		return nil
	})
	assert.NoError(t, err)

	// Test getting a registered component
	factory, exists := registry.GetComponentFactory(ComponentType("test"))
	assert.True(t, exists)
	assert.NotNil(t, factory)

	// Test getting an unregistered component
	_, exists = registry.GetComponentFactory(ComponentType("nonexistent"))
	assert.False(t, exists)

	// Test unregistering a component
	registry.UnregisterComponent(ComponentType("test"))
	_, exists = registry.GetComponentFactory(ComponentType("test"))
	assert.False(t, exists)

	// Test listing components
	err = registry.RegisterComponent(ComponentType("test1"), func(config core.RenderConfig, id string) core.ComponentInterface {
		return nil
	})
	assert.NoError(t, err)

	err = registry.RegisterComponent(ComponentType("test2"), func(config core.RenderConfig, id string) core.ComponentInterface {
		return nil
	})
	assert.NoError(t, err)

	components := registry.ListComponents()
	assert.Len(t, components, 2)

	// Test that global registry works
	globalReg := GetGlobalRegistry()
	assert.NotNil(t, globalReg)

	// Ensure it has built-in components
	_, exists = globalReg.GetComponentFactory(TableComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(FormComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(InputComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(ViewportComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(FooterComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(ChatComponent)
	assert.True(t, exists)

	// Test new components are registered
	_, exists = globalReg.GetComponentFactory(TimerComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(StopwatchComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(FilePickerComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(HelpComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(KeyComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(CursorComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(ListComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(PaginatorComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(ProgressComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(SpinnerComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(TextareaComponent)
	assert.True(t, exists)

	_, exists = globalReg.GetComponentFactory(CRUDComponent)
	assert.True(t, exists)
}