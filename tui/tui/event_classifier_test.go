package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestClassifyMessage tests the event classification logic
func TestClassifyMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		expected EventClass
	}{
		// Geometry events
		{"Mouse press", tea.MouseMsg{Action: tea.MouseActionPress}, GeometryEvent},
		{"Mouse release", tea.MouseMsg{Action: tea.MouseActionRelease}, GeometryEvent},
		{"Mouse motion", tea.MouseMsg{Action: tea.MouseActionMotion}, GeometryEvent},

		// Component events
		{"Key input", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, ComponentEvent},
		{"Key enter", tea.KeyMsg{Type: tea.KeyEnter}, ComponentEvent},
		{"Key backspace", tea.KeyMsg{Type: tea.KeyBackspace}, ComponentEvent},
		{"Key space", tea.KeyMsg{Type: tea.KeySpace}, ComponentEvent},
		{"Custom message", struct{}{}, ComponentEvent},

		// System events
		{"Window size", tea.WindowSizeMsg{Width: 80, Height: 24}, SystemEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyMessage(tt.msg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsNavigationKey tests navigation key detection
func TestIsNavigationKey(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.KeyMsg
		expected bool
	}{
		{"Tab key", tea.KeyMsg{Type: tea.KeyTab}, true},
		{"Shift+Tab", tea.KeyMsg{Type: tea.KeyShiftTab}, true},
		{"Tab as rune", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\t'}}, true},
		{"Regular key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, false},
		{"Enter key", tea.KeyMsg{Type: tea.KeyEnter}, false},
		{"Escape", tea.KeyMsg{Type: tea.KeyEsc}, false},
		{"Ctrl+C", tea.KeyMsg{Type: tea.KeyCtrlC}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNavigationKey(tt.msg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldDispatchToRuntime tests Runtime dispatch decision
func TestShouldDispatchToRuntime(t *testing.T) {
	assert.True(t, ShouldDispatchToRuntime(tea.MouseMsg{}))
	assert.False(t, ShouldDispatchToRuntime(tea.KeyMsg{Type: tea.KeyEnter}))
	assert.False(t, ShouldDispatchToRuntime(tea.WindowSizeMsg{}))
}

// TestShouldPreserveCommands tests command preservation decision
func TestShouldPreserveCommands(t *testing.T) {
	// Component messages should preserve commands
	assert.True(t, ShouldPreserveCommands(tea.KeyMsg{Type: tea.KeyEnter}))
	assert.True(t, ShouldPreserveCommands(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}))

	// Geometry events don't preserve commands
	assert.False(t, ShouldPreserveCommands(tea.MouseMsg{}))

	// System events
	assert.False(t, ShouldPreserveCommands(tea.WindowSizeMsg{}))
}
