package component

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTextareaInitReturnsCmd tests that TextareaComponentWrapper.Init returns nil (focus managed by framework)
func TestTextareaInitReturnsCmd(t *testing.T) {
	props := TextareaProps{
		Disabled:    false,
		Placeholder: "Enter text here",
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea")

	// Call Init and check if it returns nil
	// Focus should NOT be set automatically
	cmd := wrapper.Init()

	// Verify that the component is NOT focused after Init
	if wrapper.GetFocus() {
		t.Error("TextareaComponentWrapper should NOT be focused after Init - focus is managed by framework")
	}

	// Cmd should be nil
	if cmd != nil {
		t.Error("TextareaComponentWrapper.Init should return nil")
	}
}

// TestTextareaInitReturnsNilWhenDisabled tests that TextareaComponentWrapper.Init returns nil when disabled
func TestTextareaInitReturnsNilWhenDisabled(t *testing.T) {
	props := TextareaProps{
		Disabled:    true,
		Placeholder: "Enter text here",
	}
	wrapper := NewTextareaComponentWrapper(props, "test-textarea-disabled")

	cmd := wrapper.Init()

	// Should return nil when disabled
	if cmd != nil {
		t.Error("TextareaComponentWrapper.Init should return nil when disabled")
	}

	// Verify that the component is NOT focused when disabled
	if wrapper.GetFocus() {
		t.Error("TextareaComponentWrapper should NOT be focused after Init when disabled")
	}
}

// TestTextareaSetFocusWithCmd tests the SetFocusWithCmd method
func TestTextareaSetFocusWithCmd(t *testing.T) {
	props := TextareaProps{
		Disabled: false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-focus")

	// SetFocusWithCmd should focus the component
	cmd := wrapper.SetFocusWithCmd()

	// Verify focus is set
	if !wrapper.GetFocus() {
		t.Error("SetFocusWithCmd should set focus on the textarea")
	}

	// Cmd should be nil for textarea (no BlinkCmd)
	if cmd != nil {
		t.Error("SetFocusWithCmd should return nil for textarea (no BlinkCmd)")
	}
}

// TestTextareaInitFlow tests the complete Init flow
func TestTextareaInitFlow(t *testing.T) {
	testCases := []struct {
		name        string
		disabled    bool
		expectFocus bool
	}{
		{
			name:        "Enabled textarea should not be focused by default",
			disabled:    false,
			expectFocus: false, // Changed: No automatic focus
		},
		{
			name:        "Disabled textarea should not be focused",
			disabled:    true,
			expectFocus: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			props := TextareaProps{
				Disabled:    tc.disabled,
				Placeholder: "Test placeholder",
			}
			wrapper := NewTextareaComponentWrapper(props, "test-flow")

			cmd := wrapper.Init()

			if cmd != nil {
				t.Errorf("Init should return nil for textarea")
			}

			if wrapper.GetFocus() != tc.expectFocus {
				t.Errorf("Expected HasFocus to be %v, got %v", tc.expectFocus, wrapper.GetFocus())
			}
		})
	}
}

// TestTextareaInitWithDefaultValue tests that Init preserves the default value
func TestTextareaInitWithDefaultValue(t *testing.T) {
	props := TextareaProps{
		Disabled: false,
		Value:    "Default value",
	}
	wrapper := NewTextareaComponentWrapper(props, "test-default-value")

	wrapper.Init()

	// Check that the default value is preserved
	if wrapper.GetValue() != "Default value" {
		t.Errorf("Expected value to be 'Default value', got '%s'", wrapper.GetValue())
	}

	// Should not be focused by default
	if wrapper.GetFocus() {
		t.Error("Textarea should not be focused by default after Init")
	}
}

// TestTextareaInitAfterBlur tests that Init can re-focus a blurred textarea
func TestTextareaInitAfterBlur(t *testing.T) {
	props := TextareaProps{
		Disabled: false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-blur-focus")

	// First, blur the component
	wrapper.SetFocus(false)

	// Verify it's not focused (it wasn't focused to begin with)
	if wrapper.GetFocus() {
		t.Error("Textarea should not be focused by default after SetFocus(false)")
	}

	// Now call Init - still shouldn't be focused
	wrapper.Init()

	// Verify it's still not focused (init doesn't give focus anymore)
	if wrapper.GetFocus() {
		t.Error("Textarea should not be focused after Init - focus is managed by framework")
	}
}

// TestTextareaInitBatchWithOtherCommands tests that Init returns can be batched with other commands
func TestTextareaInitBatchWithOtherCommands(t *testing.T) {
	props := TextareaProps{
		Disabled: false,
	}
	wrapper := NewTextareaComponentWrapper(props, "test-batch")

	initCmd := wrapper.Init()

	// Create a mock command
	mockCmd := func() tea.Msg {
		return struct{}{} // empty message
	}

	// Batch the commands (should work even if initCmd is nil)
	batchedCmd := tea.Batch(initCmd, mockCmd)

	if batchedCmd == nil {
		t.Error("Batched command should not be nil")
	}
}
