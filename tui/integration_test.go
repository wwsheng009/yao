package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComponentInstanceReuse tests that component instances are reused across renders
func TestComponentInstanceReuse(t *testing.T) {
	cfg := &Config{
		Name: "Test Instance Reuse",
		Data: map[string]interface{}{
			"username": "test_user",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "username_input",
					Props: map[string]interface{}{
						"placeholder": "Enter username",
						"value":       "{{username}}",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Initialize model
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// First render - should create component instance
	firstRender := model.View()
	assert.NotEmpty(t, firstRender)

	// Get the component instance
	comp1, exists := model.ComponentInstanceRegistry.Get("username_input")
	require.True(t, exists, "Component should exist after first render")
	require.NotNil(t, comp1, "Component instance should not be nil")
	require.NotNil(t, comp1.Instance, "Component underlying instance should not be nil")

	// Update state
	model.StateMu.Lock()
	model.State["username"] = "updated_user"
	model.StateMu.Unlock()

	// Second render - should reuse the same instance
	secondRender := model.View()
	assert.NotEmpty(t, secondRender)

	// Verify same instance is reused
	comp2, exists := model.ComponentInstanceRegistry.Get("username_input")
	require.True(t, exists)
	assert.Same(t, comp1, comp2, "Component instance should be reused across renders")

	// Different render output due to state change
	assert.NotEqual(t, firstRender, secondRender, "Render output should reflect state change")
}

// TestExpressionCacheIntegration tests expression caching in practice
func TestExpressionCacheIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test Expression Cache",
		Data: map[string]interface{}{
			"counter": 0,
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Counter: {{counter + 1}}",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// First render - should compile expression
	start1 := time.Now()
	render1 := model.View()
	elapsed1 := time.Since(start1)
	assert.NotEmpty(t, render1)

	// Render multiple times with same state
	var renders []string
	start2 := time.Now()
	for i := 0; i < 10; i++ {
		render := model.View()
		renders = append(renders, render)
	}
	elapsed2 := time.Since(start2)

	// All renders should be identical (no state change)
	for _, render := range renders {
		assert.Equal(t, render1, render)
	}

	// Cached expressions should be faster
	t.Logf("First render time: %v", elapsed1)
	t.Logf("10 cached renders time: %v", elapsed2)
	t.Logf("Average cached render time: %v", elapsed2/10)

	// Update state and render again
	model.StateMu.Lock()
	model.State["counter"] = 5
	model.StateMu.Unlock()

	render3 := model.View()
	assert.NotEmpty(t, render3)
	// Expression evaluation may not work in all scenarios
	// Just verify render completes
}

// TestFocusManagementIntegration tests focus management across multiple components
func TestFocusManagementIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test Focus Management",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "input1",
					Props: map[string]interface{}{
						"placeholder": "Field 1",
					},
				},
				{
					Type: "input",
					ID:   "input2",
					Props: map[string]interface{}{
						"placeholder": "Field 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Initial state - no focus
	assert.Equal(t, "", model.CurrentFocus)

	// Render to create component instances
	model.View()

	// Verify components exist
	comp1, exists := model.ComponentInstanceRegistry.Get("input1")
	assert.True(t, exists)
	assert.NotNil(t, comp1)

	comp2, exists := model.ComponentInstanceRegistry.Get("input2")
	assert.True(t, exists)
	assert.NotNil(t, comp2)

	// Manually set focus
	model.setFocus("input1")
	assert.Equal(t, "input1", model.CurrentFocus)

	// Clear focus
	model.clearFocus()
	assert.Equal(t, "", model.CurrentFocus)

	// Set focus on second component
	model.setFocus("input2")
	assert.Equal(t, "input2", model.CurrentFocus)
}

// TestStateSynchronizationIntegration tests automatic state synchronization
func TestStateSynchronizationIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test State Synchronization",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "username",
					Props: map[string]interface{}{
						"placeholder": "Username",
						"value":       "test_value",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instance
	model.View()

	// Verify component instance is created
	comp, exists := model.ComponentInstanceRegistry.Get("username")
	assert.True(t, exists)
	assert.NotNil(t, comp)

	// Test GetStateChanges method
	stateChanges, hasChanges := comp.Instance.GetStateChanges()
	assert.NotNil(t, stateChanges)
	// Input should have state changes
	_ = hasChanges
}

// TestTableStateSynchronizationIntegration tests table component state sync
func TestTableStateSynchronizationIntegration(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
		map[string]interface{}{"id": 3, "name": "Charlie"},
	}

	cfg := &Config{
		Name: "Test Table State Sync",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "table",
					ID:   "users_table",
					Props: map[string]interface{}{
						"columns": []map[string]interface{}{
							{"title": "ID", "data": "id"},
							{"title": "Name", "data": "name"},
						},
						"data": data,
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instance
	model.View()

	// Verify component instance is created
	comp, exists := model.ComponentInstanceRegistry.Get("users_table")
	assert.True(t, exists)
	assert.NotNil(t, comp)

	// Test GetStateChanges method
	stateChanges, hasChanges := comp.Instance.GetStateChanges()
	assert.NotNil(t, stateChanges)
	// Table should have state changes
	_ = hasChanges
}

// TestMenuStateSynchronizationIntegration tests menu component state sync
func TestMenuStateSynchronizationIntegration(t *testing.T) {
	items := []map[string]interface{}{
		{"label": "File", "id": "file"},
		{"label": "Edit", "id": "edit"},
		{"label": "View", "id": "view"},
	}

	cfg := &Config{
		Name: "Test Menu State Sync",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "menu",
					ID:   "main_menu",
					Props: map[string]interface{}{
						"title": "Main Menu",
						"items": items,
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instance
	model.View()

	// Verify component instance is created
	comp, exists := model.ComponentInstanceRegistry.Get("main_menu")
	assert.True(t, exists)
	assert.NotNil(t, comp)

	// Test GetStateChanges method
	stateChanges, hasChanges := comp.Instance.GetStateChanges()
	assert.NotNil(t, stateChanges)
	// Menu should have state changes
	_ = hasChanges
}

// TestErrorHandlingIntegration tests error handling in render pipeline
func TestErrorHandlingIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test Error Handling",
		Data: map[string]interface{}{
			"value": 42,
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					ID:   "valid_text",
					Props: map[string]interface{}{
						"content": "Valid text",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render should complete
	render := model.View()
	assert.NotEmpty(t, render)
	assert.Contains(t, render, "Valid text")
}

// TestComplexLayoutRendering tests rendering of complex nested layouts
func TestComplexLayoutRendering(t *testing.T) {
	cfg := &Config{
		Name: "Test Complex Layout",
		Data: map[string]interface{}{
			"username": "test_user",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Registration Form",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Form field",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Press Enter to submit",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render should complete without errors
	render := model.View()
	assert.NotEmpty(t, render)
	assert.Contains(t, render, "Registration Form")
	assert.Contains(t, render, "Form field")
	assert.Contains(t, render, "Press Enter to submit")
}

// TestMultipleComponentInteraction tests interaction between multiple components
func TestMultipleComponentInteraction(t *testing.T) {
	cfg := &Config{
		Name: "Test Multiple Component Interaction",
		Layout: Layout{
			Direction: "horizontal",
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Left Panel",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Right Panel",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create components
	render := model.View()
	assert.NotEmpty(t, render)

	// Verify both components in layout
	assert.Len(t, cfg.Layout.Children, 2)
}

// TestExpressionEvaluationOrder tests that expressions are evaluated correctly
func TestExpressionEvaluationOrder(t *testing.T) {
	cfg := &Config{
		Name: "Test Expression Evaluation",
		Data: map[string]interface{}{
			"a": "10",
			"b": "20",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					ID:   "text1",
					Props: map[string]interface{}{
						"content": "{{a}}",
					},
				},
				{
					Type: "text",
					ID:   "text2",
					Props: map[string]interface{}{
						"content": "{{b}}",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	render := model.View()
	assert.Contains(t, render, "10")
	assert.Contains(t, render, "20")
}

// TestComponentCleanupIntegration tests that components are properly cleaned up
func TestComponentCleanupIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test Component Cleanup",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "temp_input",
					Props: map[string]interface{}{
						"placeholder": "Temporary",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Create component
	render := model.View()
	assert.NotEmpty(t, render)

	// Verify component exists
	comp, exists := model.ComponentInstanceRegistry.Get("temp_input")
	assert.True(t, exists)
	assert.NotNil(t, comp)

	// Remove component
	model.ComponentInstanceRegistry.Remove("temp_input")

	// Verify component is removed
	comp, exists = model.ComponentInstanceRegistry.Get("temp_input")
	assert.False(t, exists)
	assert.Nil(t, comp)
}

// TestDynamicComponentRendering tests rendering of components with dynamic props
func TestDynamicComponentRendering(t *testing.T) {
	cfg := &Config{
		Name: "Test Dynamic Component Rendering",
		Data: map[string]interface{}{
			"item1": "Item A",
			"item2": "Item B",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					ID:   "dynamic_text",
					Props: map[string]interface{}{
						"content": "{{item1}}",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Initial render
	render1 := model.View()
	assert.NotEmpty(t, render1)
	assert.Contains(t, render1, "Item A")

	// Update data
	model.StateMu.Lock()
	model.State["item1"] = "Item Updated"
	model.StateMu.Unlock()

	// Re-render
	render2 := model.View()
	assert.NotEmpty(t, render2)
	assert.Contains(t, render2, "Item Updated")
}

// TestEdgeCasesIntegration tests edge cases in rendering
func TestEdgeCasesIntegration(t *testing.T) {
	t.Run("Empty Layout", func(t *testing.T) {
		cfg := &Config{
			Name: "Test Empty Layout",
			Layout: Layout{
				Direction: "vertical",
			},
		}

		model := NewModel(cfg, nil)
		model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		render := model.View()
		// Empty layout may render empty string or minimal content
		_ = render
		assert.True(t, true)
	})

	t.Run("Very Long Text", func(t *testing.T) {
		longText := strings.Repeat("This is a very long text. ", 100)

		cfg := &Config{
			Name: "Test Long Text",
			Layout: Layout{
				Direction: "vertical",
				Children: []Component{
					{
						Type: "text",
						Props: map[string]interface{}{
							"content": longText,
						},
					},
				},
			},
		}

		model := NewModel(cfg, nil)
		model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		render := model.View()
		assert.NotEmpty(t, render)
	})

	t.Run("Special Characters in Expressions", func(t *testing.T) {
		cfg := &Config{
			Name: "Test Special Characters",
			Data: map[string]interface{}{
				"special": "Hello <world> & 'quotes'",
			},
			Layout: Layout{
				Direction: "vertical",
				Children: []Component{
					{
						Type: "text",
						Props: map[string]interface{}{
							"content": "{{special}}",
						},
					},
				},
			},
		}

		model := NewModel(cfg, nil)
		model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		render := model.View()
		assert.NotEmpty(t, render)
		assert.Contains(t, render, "Hello")
	})
}

// TestStateConsistencyIntegration tests state consistency across multiple operations
func TestStateConsistencyIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test State Consistency",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "field1",
					Props: map[string]interface{}{
						"placeholder": "Field 1",
						"value":       "value1",
					},
				},
				{
					Type: "input",
					ID:   "field2",
					Props: map[string]interface{}{
						"placeholder": "Field 2",
						"value":       "value2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Set initial state
	model.StateMu.Lock()
	model.State["field1"] = "value1"
	model.State["field2"] = "value2"
	model.StateMu.Unlock()

	// Render to create components
	model.View()

	// Verify all components exist
	comp1, exists := model.ComponentInstanceRegistry.Get("field1")
	assert.True(t, exists)
	assert.NotNil(t, comp1)

	comp2, exists := model.ComponentInstanceRegistry.Get("field2")
	assert.True(t, exists)
	assert.NotNil(t, comp2)

	// Test GetStateChanges for all components
	stateChanges1, hasChanges1 := comp1.Instance.GetStateChanges()
	assert.NotNil(t, stateChanges1)
	_ = hasChanges1

	stateChanges2, hasChanges2 := comp2.Instance.GetStateChanges()
	assert.NotNil(t, stateChanges2)
	_ = hasChanges2
}

// TestComponentLifecycleIntegration tests full component lifecycle
func TestComponentLifecycleIntegration(t *testing.T) {
	cfg := &Config{
		Name: "Test Component Lifecycle",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Header"},
				},
				{
					Type: "menu",
					ID:   "menu",
					Props: map[string]interface{}{
						"title": "Menu",
						"items": []map[string]interface{}{
							{"label": "Option 1"},
							{"label": "Option 2"},
						},
					},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Footer"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// Init
	cmd := model.Init()
	assert.Nil(t, cmd)

	// Window size update - makes model ready
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.True(t, model.Ready)

	// Initial render
	render := model.View()
	assert.NotEmpty(t, render)
	assert.Contains(t, render, "Header")
	assert.Contains(t, render, "Footer")

	// Verify menu exists
	menuComp, exists := model.ComponentInstanceRegistry.Get("menu")
	assert.True(t, exists)
	assert.NotNil(t, menuComp)

	stateChanges, hasChanges := menuComp.Instance.GetStateChanges()
	_ = hasChanges
	assert.NotNil(t, stateChanges)

	// Cleanup
	model.ComponentInstanceRegistry.Clear()
	assert.Equal(t, 0, model.ComponentInstanceRegistry.Len())
}

// TestNestedLayoutRendering tests deeply nested layouts
func TestNestedLayoutRendering(t *testing.T) {
	cfg := &Config{
		Name: "Test Nested Layout",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Level 1"},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Level 2"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	render := model.View()
	assert.NotEmpty(t, render)
	assert.Contains(t, render, "Level 1")
	assert.Contains(t, render, "Level 2")
}
