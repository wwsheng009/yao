package tui

import (
	"testing"

	"github.com/yaoapp/yao/tui/components"
)

// TestGlobalFocusExclusion verifies that only one component can have focus at any time
// This test ensures the global focus exclusion mechanism works correctly
func TestGlobalFocusExclusion(t *testing.T) {
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
					},
				},
				{
					ID:   "input3",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 3",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Set focus to input1
	model.setFocus("input1")

	// Verify only input1 has focus
	for id, comp := range model.Components {
		if id == "input1" {
			if !comp.Instance.GetFocus() {
				t.Errorf("Expected input1 to have focus")
			}
		} else {
			if comp.Instance.GetFocus() {
				t.Errorf("Component %s should not have focus (input1 is focused)", id)
			}
		}
	}

	// Switch focus to input2
	model.setFocus("input2")

	// Verify only input2 has focus
	for id, comp := range model.Components {
		if id == "input2" {
			if !comp.Instance.GetFocus() {
				t.Errorf("Expected input2 to have focus")
			}
		} else {
			if comp.Instance.GetFocus() {
				t.Errorf("Component %s should not have focus (input2 is focused)", id)
			}
		}
	}

	// Clear all focus
	model.clearFocus()

	// Verify no component has focus
	for id, comp := range model.Components {
		if comp.Instance.GetFocus() {
			t.Errorf("Component %s should not have focus after clearFocus", id)
		}
	}
}

// TestGlobalFocusExclusionAfterDirectSetFocus verifies that setFocus correctly
// handles the case where a component was directly given focus through non-standard paths
func TestGlobalFocusExclusionAfterDirectSetFocus(t *testing.T) {
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Directly give focus to input2 (simulating non-standard code path)
	comp1, _ := model.Components["input1"]
	comp1.Instance.SetFocus(true)

	// Use setFocus to focus input1
	// This should clear focus from input2 and set focus to input1
	model.setFocus("input1")

	// Verify only input1 has focus
	for id, comp := range model.Components {
		if id == "input1" {
			if !comp.Instance.GetFocus() {
				t.Errorf("Expected input1 to have focus")
			}
		} else {
			if comp.Instance.GetFocus() {
				t.Errorf("Component %s should not have focus after setFocus", id)
			}
		}
	}
}

// TestClearFocusRemovesAllFocus verifies that clearFocus removes focus from all components
func TestClearFocusRemovesAllFocus(t *testing.T) {
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Set focus to input1
	model.setFocus("input1")

	// Verify input1 has focus
	comp1, _ := model.Components["input1"]
	if !comp1.Instance.GetFocus() {
		t.Errorf("Expected input1 to have focus initially")
	}

	// Directly give focus to input2 (simulating non-standard code path)
	comp2, _ := model.Components["input2"]
	comp2.Instance.SetFocus(true)

	// Now both input1 and input2 have focus (inconsistent state)

	// Call clearFocus
	model.clearFocus()

	// Verify no component has focus
	for id, comp := range model.Components {
		if comp.Instance.GetFocus() {
			t.Errorf("Component %s should not have focus after clearFocus", id)
		}
	}
}

// TestValidateAndCorrectFocusState verifies the focus validation function
func TestValidateAndCorrectFocusState(t *testing.T) {
	cfg := &Config{
		Name: "test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// Set focus to input1
	model.setFocus("input1")

	// Simulate focus state corruption: give focus to input2 directly
	comp2, _ := model.Components["input2"]
	comp2.Instance.SetFocus(true)

	// Now both input1 and input2 have focus (inconsistent state)

	// Validate and correct focus state
	corrections := model.validateAndCorrectFocusState()

	if corrections == 0 {
		t.Errorf("Expected at least 1 correction, got 0")
	}

	// Verify only input1 has focus (matching CurrentFocus)
	for id, comp := range model.Components {
		if id == "input1" {
			if !comp.Instance.GetFocus() {
				t.Errorf("Expected input1 to have focus after validation")
			}
		} else {
			if comp.Instance.GetFocus() {
				t.Errorf("Component %s should not have focus after validation", id)
			}
		}
	}
}
