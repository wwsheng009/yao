package components

import (
	"testing"
	"time"
)

func TestInputCursorBlinkSpeed(t *testing.T) {
	tests := []struct {
		name          string
		props         InputProps
		expectedSpeed time.Duration
	}{
		{
			name: "Default blink speed",
			props: InputProps{
				Placeholder: "Test",
				CursorMode:  "blink",
			},
			expectedSpeed: 530 * time.Millisecond,
		},
		{
			name: "Fast blink speed",
			props: InputProps{
				Placeholder:      "Test",
				CursorMode:       "blink",
				CursorBlinkSpeed: 200,
			},
			expectedSpeed: 200 * time.Millisecond,
		},
		{
			name: "Slow blink speed",
			props: InputProps{
				Placeholder:      "Test",
				CursorMode:       "blink",
				CursorBlinkSpeed: 1000,
			},
			expectedSpeed: 1000 * time.Millisecond,
		},
		{
			name: "Zero blink speed (should use default)",
			props: InputProps{
				Placeholder:      "Test",
				CursorMode:       "blink",
				CursorBlinkSpeed: 0,
			},
			expectedSpeed: 530 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := NewInputComponentWrapper(tt.props, "test-input")

			// Check that cursor helper has the correct blink speed
			if wrapper.cursorHelper.GetBlinkSpeed() != tt.expectedSpeed {
				t.Errorf("Expected blink speed %v, got %v", tt.expectedSpeed, wrapper.cursorHelper.GetBlinkSpeed())
			}

			// Check that the underlying textinput model has the correct blink speed
			if wrapper.model.Cursor.BlinkSpeed != tt.expectedSpeed {
				t.Errorf("Expected model blink speed %v, got %v", tt.expectedSpeed, wrapper.model.Cursor.BlinkSpeed)
			}
		})
	}
}
