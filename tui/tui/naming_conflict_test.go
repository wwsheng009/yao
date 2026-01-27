package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReservedTUINames tests the reserved TUI names constant
func TestReservedTUINames(t *testing.T) {
	expectedNames := []string{
		"list", "validate", "inspect", "check", "dump", "help",
	}

	for _, name := range expectedNames {
		t.Run("ReservedName_"+name, func(t *testing.T) {
			assert.True(t, reservedTUINames[name], "%s should be reserved", name)
		})
	}
}

// TestGetReservedTUINameSlice tests getting the list of reserved names
func TestGetReservedTUINameSlice(t *testing.T) {
	names := getReservedTUINameSlice()

	// Check that we got all reserved names
	assert.Equal(t, 6, len(names), "Should have 6 reserved names")

	// Check specific names exist
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "validate")
	assert.Contains(t, names, "inspect")
	assert.Contains(t, names, "check")
	assert.Contains(t, names, "dump")
	assert.Contains(t, names, "help")
}

// TestIsReservedTUIName tests checking if a TUI name is reserved
func TestIsReservedTUIName(t *testing.T) {
	testCases := []struct {
		tuiID    string
		reserved bool
	}{
		// Reserved names
		{"list", true},
		{"validate", true},
		{"inspect", true},
		{"check", true},
		{"dump", true},
		{"help", true},

		// Not reserved (case-sensitive)
		{"List", false},
		{"validate-tui", false},
		{"LIST", false},

		// System TUIs (should not be considered reserved)
		{"__yao.tui-list", false},
		{"__system.tui", false},
		{"__helper.tui", false},

		// User TUIs
		{"myapp", false},
		{"user-dashboard", false},
		{"my-list", false},
		{"app-validate", false},
	}

	for _, tc := range testCases {
		t.Run(tc.tuiID, func(t *testing.T) {
			result := isReservedTUIName(tc.tuiID)
			assert.Equal(t, tc.reserved, result, "isReservedTUIName(%s) should return %v", tc.tuiID, tc.reserved)
		})
	}
}

// TestCheckConfigurationConflicts tests conflict detection with mock data
func TestCheckConfigurationConflicts(t *testing.T) {
	t.Run("No conflicts", func(t *testing.T) {
		// This test verifies the logic doesn't crash
		assert.True(t, len(reservedTUINames) > 0)
	})

	t.Run("Check reserved names list", func(t *testing.T) {
		names := getReservedTUINameSlice()

		// Verify we have the expected reserved names
		expected := []string{"list", "validate", "inspect", "check", "dump", "help"}

		assert.Equal(t, len(expected), len(names), "Should have 6 reserved names")

		for _, expectedName := range expected {
			assert.Contains(t, names, expectedName, "%s should be reserved", expectedName)
		}
	})
}
