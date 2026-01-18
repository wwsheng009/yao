package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
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
	factory, err := registry.GetComponentFactory(ComponentType("test"))
	assert.NoError(t, err)
	assert.NotNil(t, factory)

	// Test getting an unregistered component
	_, err = registry.GetComponentFactory(ComponentType("nonexistent"))
	assert.Error(t, err)

	// Test unregistering a component
	registry.UnregisterComponent(ComponentType("test"))
	_, err = registry.GetComponentFactory(ComponentType("test"))
	assert.Error(t, err)

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
	_, err = globalReg.GetComponentFactory(TableComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(FormComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(InputComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(ViewportComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(FooterComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(ChatComponent)
	assert.NoError(t, err)

	// Test new components are registered
	_, err = globalReg.GetComponentFactory(TimerComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(StopwatchComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(FilePickerComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(HelpComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(KeyComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(CursorComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(ListComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(PaginatorComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(ProgressComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(SpinnerComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(TextareaComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponentFactory(CRUDComponent)
	assert.NoError(t, err)
}
