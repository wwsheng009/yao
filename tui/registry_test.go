package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponentRegistry(t *testing.T) {
	// Create a new registry
	registry := NewComponentRegistry()

	// Test registering a component
	err := registry.RegisterComponent("test", func(props map[string]interface{}, width int) string {
		return "test component"
	})
	assert.NoError(t, err)

	// Test getting a registered component
	renderer, err := registry.GetComponent("test")
	assert.NoError(t, err)
	assert.NotNil(t, renderer)

	// Test getting the rendered output
	result := renderer(nil, 0)
	assert.Equal(t, "test component", result)

	// Test getting an unregistered component
	_, err = registry.GetComponent("nonexistent")
	assert.Error(t, err)

	// Test unregistering a component
	registry.UnregisterComponent("test")
	_, err = registry.GetComponent("test")
	assert.Error(t, err)

	// Test listing components
	err = registry.RegisterComponent("test1", func(props map[string]interface{}, width int) string {
		return "test1"
	})
	assert.NoError(t, err)

	err = registry.RegisterComponent("test2", func(props map[string]interface{}, width int) string {
		return "test2"
	})
	assert.NoError(t, err)

	components := registry.ListComponents()
	assert.Len(t, components, 2)

	// Test that global registry works
	globalReg := GetGlobalRegistry()
	assert.NotNil(t, globalReg)

	// Ensure it has built-in components
	_, err = globalReg.GetComponent(TableComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponent(FormComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponent(InputComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponent(ViewportComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponent(FooterComponent)
	assert.NoError(t, err)

	_, err = globalReg.GetComponent(ChatComponent)
	assert.NoError(t, err)
}