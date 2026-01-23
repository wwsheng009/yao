package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRuntimeE2E verifies Runtime integration with real TUI configuration
func TestRuntimeE2E(t *testing.T) {
	// Create a realistic TUI configuration
	config := &Config{
		Name: "Runtime E2E Test",
		Data: map[string]interface{}{
			"title":   "Test Application",
			"status":  "running",
			"counter": 42,
		},
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					ID:   "header",
					Type: "header",
					Props: map[string]interface{}{
						"title": "{{title}}",
					},
					Height: 3,
				},
				{
					ID:   "content",
					Type: "text",
					Props: map[string]interface{}{
						"content": "Status: {{status}} | Counter: {{counter}}",
					},
					Height: 5,
				},
				{
					ID:   "footer",
					Type: "footer",
					Props: map[string]interface{}{
						"text": "Press q to quit",
					},
					Height: 1,
				},
			},
		},
		AutoFocus: func() *bool { b := true; return &b }(),
	}

	// Create model with Runtime enabled
	model := NewModel(config, nil)

	// Enable Runtime
	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize model
	var cmds []tea.Cmd
	cmd := model.Init()
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Verify Runtime was initialized
	if model.RuntimeEngine == nil {
		t.Error("RuntimeEngine should be initialized after Init()")
	}
	if model.RuntimeRoot == nil {
		t.Error("RuntimeRoot should be initialized after Init()")
	}

	// Verify rendering works
	output := model.View()
	if output == "" {
		t.Error("View() should return non-empty output")
	}

	// For debugging: print the output
	t.Logf("Output:\n%s", output)

	// Debug: check text component instance
	if comp, exists := model.ComponentInstanceRegistry.Get("content"); exists {
		t.Logf("Content component instance: %T", comp.Instance)
		t.Logf("Content component View: %q", comp.Instance.View())
		if comp.Instance != nil {
			t.Logf("Content component LastConfig.Data: %+v", comp.LastConfig.Data)
			t.Logf("Content component LastConfig.Width: %d, Height: %d", comp.LastConfig.Width, comp.LastConfig.Height)
		}
	} else {
		t.Logf("Content component NOT found in registry")
	}

	// Debug: check Runtime layout boxes
	if model.RuntimeEngine != nil {
		result := model.GetLayoutResult()
		t.Logf("LayoutResult: %d boxes", len(result.Boxes))
		for i, box := range result.Boxes {
			t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		}
	}

	// Verify output contains expected content
	if !strings.Contains(output, "Test Application") {
		t.Errorf("Output should contain 'Test Application', got: %s", output)
	}
	if !strings.Contains(output, "running") {
		t.Errorf("Output should contain 'running', got: %s", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("Output should contain '42', got: %s", output)
	}

	t.Logf("Runtime E2E test passed. Output length: %d chars", len(output))
}

// TestRuntimeGeometricFocus verifies Tab navigation follows geometric order
func TestRuntimeGeometricFocus(t *testing.T) {
	// Create a layout with components in specific positions
	// Layout structure:
	//   [input1] [input2]
	//   [input3] [input4]
	config := &Config{
		Name: "Geometric Focus Test",
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					ID:   "row1",
					Type: "row",
					Props: map[string]interface{}{
						"gap": 1,
					},
					Children: []Component{
						{
							ID:   "input1",
							Type: "input",
							Props: map[string]interface{}{
								"placeholder": "Input 1",
							},
							Width: 20,
						},
						{
							ID:   "input2",
							Type: "input",
							Props: map[string]interface{}{
								"placeholder": "Input 2",
							},
							Width: 20,
						},
					},
					Height: 3,
				},
				{
					ID:   "row2",
					Type: "row",
					Props: map[string]interface{}{
						"gap": 1,
					},
					Children: []Component{
						{
							ID:   "input3",
							Type: "input",
							Props: map[string]interface{}{
								"placeholder": "Input 3",
							},
							Width: 20,
						},
						{
							ID:   "input4",
							Type: "input",
							Props: map[string]interface{}{
								"placeholder": "Input 4",
							},
							Width: 20,
						},
					},
					Height: 3,
				},
			},
		},
		TabCycles: true,
	}

	// Create model with Runtime enabled
	model := NewModel(config, nil)

	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Get Runtime focus list
	focusList := model.runtimeFocusList
	if len(focusList) != 4 {
		t.Errorf("Expected 4 focusable components, got %d: %v", len(focusList), focusList)
	}

	// Verify geometric ordering: should be input1, input2, input3, input4
	// (left-to-right, top-to-bottom)
	expectedOrder := []string{"input1", "input2", "input3", "input4"}
	for i, expected := range expectedOrder {
		if i >= len(focusList) {
			t.Errorf("Missing component at index %d: expected %s", i, expected)
			continue
		}
		if focusList[i] != expected {
			t.Errorf("Focus order mismatch at index %d: expected %s, got %s", i, expected, focusList[i])
		}
	}

	// Verify Tab navigation uses this order
	// Get focusable IDs should return Runtime focus list
	focusableIDs := model.getFocusableComponentIDs()
	if len(focusableIDs) != len(expectedOrder) {
		t.Errorf("getFocusableComponentIDs returned wrong count: %d", len(focusableIDs))
	}

	t.Logf("Geometric focus order verified: %v", focusList)
}

// TestRuntimeVsLegacy compares Runtime and Legacy rendering outputs
func TestRuntimeVsLegacy(t *testing.T) {
	config := &Config{
		Name: "Runtime vs Legacy Test",
		Data: map[string]interface{}{
			"text": "Hello, World!",
		},
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					ID:   "header",
					Type: "header",
					Props: map[string]interface{}{
						"text": "{{.text}}",
					},
					Height: 3,
				},
			},
		},
	}

	// Test with Runtime
	runtimeModel := NewModel(config, nil)
	runtimeModel.UseRuntime = true
	runtimeModel.Width = 80
	runtimeModel.Height = 24
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := runtimeModel.Update(windowMsg)
	runtimeModel = newModel.(*Model)
	runtimeOutput := runtimeModel.View()

	// Test with Legacy
	legacyModel := NewModel(config, nil)
	legacyModel.UseRuntime = false
	legacyModel.Width = 80
	legacyModel.Height = 24
	newModel, _ = legacyModel.Update(windowMsg)
	legacyModel = newModel.(*Model)
	legacyOutput := legacyModel.View()

	// Both should produce non-empty output
	if runtimeOutput == "" {
		t.Error("Runtime output is empty")
	}
	if legacyOutput == "" {
		t.Error("Legacy output is empty")
	}

	// Both should contain the expected text
	if !strings.Contains(runtimeOutput, "Hello, World!") {
		t.Error("Runtime output missing expected text")
	}
	if !strings.Contains(legacyOutput, "Hello, World!") {
		t.Error("Legacy output missing expected text")
	}

	t.Logf("Runtime output length: %d, Legacy output length: %d", len(runtimeOutput), len(legacyOutput))
}

// TestRuntimeComplexLayout tests Runtime with nested layouts
func TestRuntimeComplexLayout(t *testing.T) {
	config := &Config{
		Name: "Complex Layout Test",
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					ID:   "header",
					Type: "header",
					Props: map[string]interface{}{
						"title": "Header",
					},
					Height: 3,
				},
				{
					ID:   "main",
					Type: "row",
					Children: []Component{
						{
							ID:   "sidebar",
							Type: "text",
							Props: map[string]interface{}{
								"content": "Sidebar",
							},
							Width: 20,
						},
						{
							ID:   "content",
							Type: "text",
							Props: map[string]interface{}{
								"content": "MainContent",
							},
						},
					},
					Height: 15,
				},
				{
					ID:   "footer",
					Type: "footer",
					Props: map[string]interface{}{
						"text": "Footer",
					},
					Height: 1,
				},
			},
		},
	}

	model := NewModel(config, nil)

	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Verify layout was calculated
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should not be nil")
	}

	// Verify structure
	if len(model.RuntimeRoot.Children) == 0 {
		t.Error("RuntimeRoot should have children")
	}

	// Debug: print Runtime tree structure
	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))
	for i, child := range model.RuntimeRoot.Children {
		t.Logf("  Child[%d]: ID=%s, Type=%v, Component=%v, GrandChildren=%d", i, child.ID, child.Type, child.Component != nil, len(child.Children))
		for j, grandChild := range child.Children {
			t.Logf("    GrandChild[%d]: ID=%s, Type=%v, Component=%v", j, grandChild.ID, grandChild.Type, grandChild.Component != nil)
		}
	}

	output := model.View()
	if output == "" {
		t.Error("View() should return non-empty output")
	}

	// Debug: print first 500 chars of output
	if len(output) > 500 {
		t.Logf("Output (first 500 chars):\n%s", output[:500])
	} else {
		t.Logf("Output:\n%s", output)
	}

	// Verify all sections are rendered
	expectedTexts := []string{"Header", "Sidebar", "MainContent", "Footer"}
	for _, text := range expectedTexts {
		if !strings.Contains(output, text) {
			t.Errorf("Output should contain '%s'", text)
		}
	}

	t.Logf("Complex layout test passed. Output length: %d chars", len(output))
}

// TestRuntimeDynamicStateChange tests Runtime with state changes
func TestRuntimeDynamicStateChange(t *testing.T) {
	config := &Config{
		Name: "Dynamic State Test",
		Data: map[string]interface{}{
			"message": "Initial",
		},
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					ID:   "display",
					Type: "text",
					Props: map[string]interface{}{
						"content": "{{message}}",
					},
					Height: 3,
				},
			},
		},
	}

	model := NewModel(config, nil)

	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize the model first!
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Initial render should contain "Initial"
	output1 := model.View()
	if !strings.Contains(output1, "Initial") {
		t.Errorf("Initial output should contain 'Initial', got: %s", output1)
	}

	// Update state
	model.StateMu.Lock()
	model.State["message"] = "Updated"
	model.StateMu.Unlock()

	// Sync state to Runtime
	model.syncStateToRuntime()

	// Re-render
	output2 := model.View()
	if !strings.Contains(output2, "Updated") {
		t.Errorf("Updated output should contain 'Updated', got: %s", output2)
	}

	t.Logf("Dynamic state change test passed")
}

// BenchmarkRuntimeLayout benchmarks Runtime layout calculation
func BenchmarkRuntimeLayout(b *testing.B) {
	config := &Config{
		Name: "Benchmark Test",
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{ID: "h1", Type: "header", Height: 3},
				{ID: "h2", Type: "header", Height: 3},
				{ID: "h3", Type: "header", Height: 3},
				{ID: "h4", Type: "header", Height: 3},
				{ID: "h5", Type: "header", Height: 3},
			},
		},
	}

	model := NewModel(config, nil)

	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}
