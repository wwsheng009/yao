package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFocusableComponentRegistration tests that focusable components are properly registered
func TestFocusableComponentRegistration(t *testing.T) {
	registry := GetGlobalRegistry()

	// Test that known focusable components are registered
	assert.True(t, registry.IsFocusable(InputComponent), "input component should be focusable")
	assert.True(t, registry.IsFocusable(TextareaComponent), "textarea component should be focusable")
	assert.True(t, registry.IsFocusable(MenuComponent), "menu component should be focusable")
	assert.True(t, registry.IsFocusable(FormComponent), "form component should be focusable")
	assert.True(t, registry.IsFocusable(TableComponent), "table component should be focusable")
	assert.True(t, registry.IsFocusable(ChatComponent), "chat component should be focusable")
	assert.True(t, registry.IsFocusable(FilePickerComponent), "filepicker component should be focusable")
	assert.True(t, registry.IsFocusable(CRUDComponent), "crud component should be focusable")

	// Test that non-focusable components are not registered
	assert.False(t, registry.IsFocusable(HeaderComponent), "header component should not be focusable")
	assert.False(t, registry.IsFocusable(TextComponent), "text component should not be focusable")
	assert.False(t, registry.IsFocusable(FooterComponent), "footer component should not be focusable")
}

// TestFocusableComponentDynamicRegistration tests dynamic registration of focusable components
func TestFocusableComponentDynamicRegistration(t *testing.T) {
	registry := NewComponentRegistry()

	// Create a custom component type
	customType := ComponentType("custom")
	assert.False(t, registry.IsFocusable(customType), "custom component should not be focusable initially")

	// Register as focusable
	registry.RegisterFocusableComponent(customType)
	assert.True(t, registry.IsFocusable(customType), "custom component should be focusable after registration")

	// Unregister
	registry.UnregisterFocusableComponent(customType)
	assert.False(t, registry.IsFocusable(customType), "custom component should not be focusable after unregistration")
}

// TestGetFocusableComponentTypes tests retrieval of all focusable component types
func TestGetFocusableComponentTypes(t *testing.T) {
	registry := GetGlobalRegistry()

	types := registry.GetFocusableComponentTypes()
	assert.Greater(t, len(types), 0, "should have at least one focusable component type")

	// Verify that all returned types are indeed focusable
	for _, compType := range types {
		assert.True(t, registry.IsFocusable(compType), "returned component type %s should be focusable", compType)
	}
}
