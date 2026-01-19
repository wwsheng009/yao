package components

import (
	"testing"
	"time"

	cursor "github.com/charmbracelet/bubbles/cursor"
)

func TestNewCursorHelper(t *testing.T) {
	config := CursorConfig{
		Mode:       cursor.CursorBlink,
		Char:       "|",
		BlinkSpeed: 530 * time.Millisecond,
		Visible:    true,
	}

	helper := NewCursorHelper(config)

	if helper == nil {
		t.Fatal("NewCursorHelper returned nil")
	}

	if helper.GetMode() != cursor.CursorBlink {
		t.Errorf("Expected mode %v, got %v", cursor.CursorBlink, helper.GetMode())
	}

	if helper.GetChar() != "|" {
		t.Errorf("Expected char '|', got '%s'", helper.GetChar())
	}

	if !helper.GetVisible() {
		t.Error("Expected visible to be true")
	}
}

func TestCursorHelperSetMode(t *testing.T) {
	config := CursorConfig{
		Mode:       cursor.CursorBlink,
		Char:       "|",
		BlinkSpeed: 530 * time.Millisecond,
		Visible:    true,
	}

	helper := NewCursorHelper(config)

	// Test setting to static mode
	helper.SetMode(cursor.CursorStatic)
	if helper.GetMode() != cursor.CursorStatic {
		t.Errorf("Expected mode %v, got %v", cursor.CursorStatic, helper.GetMode())
	}

	// Test setting to hide mode
	helper.SetMode(cursor.CursorHide)
	if helper.GetMode() != cursor.CursorHide {
		t.Errorf("Expected mode %v, got %v", cursor.CursorHide, helper.GetMode())
	}
}

func TestCursorHelperSetChar(t *testing.T) {
	config := CursorConfig{
		Mode:       cursor.CursorBlink,
		Char:       "|",
		BlinkSpeed: 530 * time.Millisecond,
		Visible:    true,
	}

	helper := NewCursorHelper(config)

	// Test setting custom char
	helper.SetChar("█")
	if helper.GetChar() != "█" {
		t.Errorf("Expected char '█', got '%s'", helper.GetChar())
	}
}

func TestCursorHelperSetVisible(t *testing.T) {
	config := CursorConfig{
		Mode:       cursor.CursorBlink,
		Char:       "|",
		BlinkSpeed: 530 * time.Millisecond,
		Visible:    true,
	}

	helper := NewCursorHelper(config)

	// Test setting to invisible
	helper.SetVisible(false)
	if helper.GetVisible() {
		t.Error("Expected visible to be false")
	}
	if helper.GetChar() != "" {
		t.Errorf("Expected empty char when invisible, got '%s'", helper.GetChar())
	}

	// Test setting to visible
	helper.SetVisible(true)
	if !helper.GetVisible() {
		t.Error("Expected visible to be true")
	}
}

func TestCursorHelperGetChar(t *testing.T) {
	tests := []struct {
		name     string
		config   CursorConfig
		expected string
	}{
		{
			name: "Blink mode with custom char",
			config: CursorConfig{
				Mode:    cursor.CursorBlink,
				Char:    "█",
				Visible: true,
			},
			expected: "█",
		},
		{
			name: "Blink mode with default char",
			config: CursorConfig{
				Mode:    cursor.CursorBlink,
				Char:    "",
				Visible: true,
			},
			expected: "|",
		},
		{
			name: "Static mode with default char",
			config: CursorConfig{
				Mode:    cursor.CursorStatic,
				Char:    "",
				Visible: true,
			},
			expected: "█",
		},
		{
			name: "Hide mode",
			config: CursorConfig{
				Mode:    cursor.CursorHide,
				Char:    "|",
				Visible: true,
			},
			expected: "",
		},
		{
			name: "Invisible",
			config: CursorConfig{
				Mode:    cursor.CursorBlink,
				Char:    "|",
				Visible: false,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewCursorHelper(tt.config)
			if helper.GetChar() != tt.expected {
				t.Errorf("Expected char '%s', got '%s'", tt.expected, helper.GetChar())
			}
		})
	}
}

func TestParseCursorMode(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		expected cursor.Mode
	}{
		{
			name:     "Blink mode",
			style:    "blink",
			expected: cursor.CursorBlink,
		},
		{
			name:     "Static mode",
			style:    "static",
			expected: cursor.CursorStatic,
		},
		{
			name:     "Hide mode",
			style:    "hide",
			expected: cursor.CursorHide,
		},
		{
			name:     "Block mode (maps to static)",
			style:    "block",
			expected: cursor.CursorStatic,
		},
		{
			name:     "Unknown mode (defaults to blink)",
			style:    "",
			expected: cursor.CursorBlink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := ParseCursorMode(tt.style)
			if mode != tt.expected {
				t.Errorf("Expected mode %v, got %v", tt.expected, mode)
			}
		})
	}
}

func TestInputComponentWrapperSetCursorMode(t *testing.T) {
	props := InputProps{
		Placeholder: "Enter text",
		CursorMode:  "blink",
		CursorChar:  "|",
	}

	wrapper := NewInputComponentWrapper(props, "test-input")

	// Test setting cursor mode
	wrapper.SetCursorMode("static")
	if cursorMode := ParseCursorMode("static"); wrapper.GetCursorHelper().GetMode() != cursorMode {
		t.Errorf("Expected cursor mode %v, got %v", cursorMode, wrapper.GetCursorHelper().GetMode())
	}
}

func TestInputComponentWrapperSetCursorChar(t *testing.T) {
	props := InputProps{
		Placeholder: "Enter text",
		CursorMode:  "blink",
		CursorChar:  "|",
	}

	wrapper := NewInputComponentWrapper(props, "test-input")

	// Test setting cursor char
	wrapper.SetCursorChar("█")
	if wrapper.GetCursorHelper().GetChar() != "█" {
		t.Errorf("Expected cursor char '█', got '%s'", wrapper.GetCursorHelper().GetChar())
	}
}
