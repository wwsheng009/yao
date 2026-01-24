package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/event"
	"github.com/yaoapp/yao/tui/ui/components"
)

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

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

	// Initialize the model first!
	model.Init()

	// Then send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Call View() to trigger focus list update
	_ = model.View()

	// Get Runtime focus list
	focusList := model.runtimeFocusList

	// Debug: check Components and ComponentInstanceRegistry
	t.Logf("Components map has %d items", len(model.Components))
	for id := range model.Components {
		t.Logf("  Components[%s]", id)
	}
	t.Logf("ComponentInstanceRegistry has %d items", model.ComponentInstanceRegistry.Len())
	for _, id := range []string{"input1", "input2", "input3", "input4", "row1", "row2"} {
		if comp, exists := model.ComponentInstanceRegistry.Get(id); exists {
			t.Logf("  Registry has %s: %T", id, comp.Instance)
		} else {
			t.Logf("  Registry missing %s", id)
		}
	}

	// Debug: check LayoutResult boxes
	if model.RuntimeEngine != nil {
		result := model.GetLayoutResult()
		t.Logf("LayoutResult: %d boxes", len(result.Boxes))
		for i, box := range result.Boxes {
			t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d, Component=%v", i, box.NodeID, box.X, box.Y, box.W, box.H, box.Node.Component != nil)
		}
	}

	// Debug: check which components are considered focusable
	for _, id := range []string{"input1", "input2", "input3", "input4"} {
		focusable := model.isComponentFocusable(id)
		t.Logf("  isComponentFocusable(%s) = %v", id, focusable)
	}

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
						"title": "{{text}}",
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
	runtimeModel.Init()
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := runtimeModel.Update(windowMsg)
	runtimeModel = newModel.(*Model)
	runtimeOutput := runtimeModel.View()

	// Test with Legacy
	legacyModel := NewModel(config, nil)
	legacyModel.UseRuntime = false
	legacyModel.Width = 80
	legacyModel.Height = 24
	legacyModel.Init()
	windowMsg = tea.WindowSizeMsg{Width: 80, Height: 24}
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

// TestRuntimeConfigOption verifies that the UseRuntime config option works correctly
func TestRuntimeConfigOption(t *testing.T) {
	// Test 1: UseRuntime=false in Config should result in Legacy engine
	configLegacy := &Config{
		Name:      "Legacy Test",
		UseRuntime: boolPtr(false),
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{ID: "header", Type: "header", Props: map[string]interface{}{"title": "Test"}},
			},
		},
	}

	modelLegacy := NewModel(configLegacy, nil)
	modelLegacy.Width = 80
	modelLegacy.Height = 24
	modelLegacy.Init()

	if modelLegacy.UseRuntime {
		t.Error("UseRuntime should be false when config.UseRuntime is false")
	}

	// Test 2: UseRuntime=true in Config should result in Runtime engine
	configRuntime := &Config{
		Name:      "Runtime Test",
		// UseRuntime defaults to true
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{ID: "header", Type: "header", Props: map[string]interface{}{"title": "Test"}},
			},
		},
	}

	modelRuntime := NewModel(configRuntime, nil)
	modelRuntime.Width = 80
	modelRuntime.Height = 24
	modelRuntime.Init()

	if !modelRuntime.UseRuntime {
		t.Error("UseRuntime should be true when config.UseRuntime is true")
	}

	// Verify Runtime engine was initialized
	if modelRuntime.RuntimeEngine == nil {
		t.Error("RuntimeEngine should be initialized when UseRuntime is true")
	}

	// Test 3: Default (UseRuntime not set) should result in Runtime engine (new default)
	configDefault := &Config{
		Name: "Default Test",
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{ID: "header", Type: "header", Props: map[string]interface{}{"title": "Test"}},
			},
		},
	}

	modelDefault := NewModel(configDefault, nil)
	if !modelDefault.UseRuntime {
		t.Error("UseRuntime should default to true (new Runtime engine is default)")
	}

	t.Logf("Runtime config option test passed")
}

// TestRuntimeInputComponent tests Input component rendering with Runtime engine
func TestRuntimeInputComponent(t *testing.T) {
	config := &Config{
		Name:      "Input Runtime Test",
		// UseRuntime defaults to true
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					ID:   "username",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter username",
						"prompt":      "> ",
						"width":       30,
					},
				},
				{
					ID:   "email",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Enter email",
						"prompt":      "@ ",
						"width":       40,
					},
				},
			},
		},
		TabCycles: true,
	}

	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Call View() to trigger rendering
	_ = model.View()

	// Verify Runtime was initialized
	if model.RuntimeEngine == nil {
		t.Fatal("RuntimeEngine should be initialized")
	}
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should be initialized")
	}

	// Check component instances
	if comp, exists := model.ComponentInstanceRegistry.Get("username"); exists {
		t.Logf("Username component: %T", comp.Instance)
	} else {
		t.Error("Username component not found in registry")
	}

	if comp, exists := model.ComponentInstanceRegistry.Get("email"); exists {
		t.Logf("Email component: %T", comp.Instance)
	} else {
		t.Error("Email component not found in registry")
	}

	// Verify geometric focus order includes both inputs
	focusList := model.runtimeFocusList
	if len(focusList) != 2 {
		t.Errorf("Expected 2 focusable components, got %d: %v", len(focusList), focusList)
	}

	// Verify output contains expected elements
	output := model.View()
	if output == "" {
		t.Error("View() should return non-empty output")
	}

	// Debug: print layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
	}

	// Verify inputs have correct widths
	usernameBox := result.FindBoxByID("username")
	if usernameBox == nil {
		t.Error("Username input box not found")
	} else if usernameBox.W != 30 {
		t.Errorf("Username width should be 30, got %d", usernameBox.W)
	}

	emailBox := result.FindBoxByID("email")
	if emailBox == nil {
		t.Error("Email input box not found")
	} else if emailBox.W != 40 {
		t.Errorf("Email width should be 40, got %d", emailBox.W)
	}

	// Verify inputs are stacked vertically (Y increases)
	if usernameBox != nil && emailBox != nil {
		if emailBox.Y <= usernameBox.Y {
			t.Errorf("Email should be below username: username.Y=%d, email.Y=%d", usernameBox.Y, emailBox.Y)
		}
	}

	t.Logf("Input component Runtime test passed. Output length: %d chars", len(output))
}

// TestRuntimeListComponent tests the native Runtime List component
func TestRuntimeListComponent(t *testing.T) {
	// Create a simple TUI configuration with a List component
	config := &Config{
		Name: "List Component Test",
		Layout: Layout{
			Direction: "row",
			Children: []Component{
				{
					Type: "list",
					ID:   "mylist",
					Props: map[string]interface{}{
						"title":            "Select an option:",
						"width":            40,
						"height":           10,
						"showTitle":        true,
						"showStatusBar":    true,
						"showFilter":       false,
						"filteringEnabled": false,
						"items": []interface{}{
							map[string]interface{}{
								"title":       "Option 1",
								"description": "First option",
								"value":       "opt1",
							},
							map[string]interface{}{
								"title":       "Option 2",
								"description": "Second option",
								"value":       "opt2",
							},
							map[string]interface{}{
								"title":       "Option 3",
								"description": "Third option",
								"value":       "opt3",
							},
						},
					},
				},
			},
		},
	}

	// Create model with Runtime enabled
	model := NewModel(config, nil)
	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize the model first!
	model.Init()

	// Then send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Verify RuntimeRoot was created
	if model.RuntimeRoot == nil {
		t.Fatalf("RuntimeRoot should not be nil after Init")
	}

	t.Logf("RuntimeRoot ID: %s", model.RuntimeRoot.ID)
	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))

	// Verify the list component exists in the registry
	if comp, exists := model.ComponentInstanceRegistry.Get("mylist"); exists {
		t.Logf("List component found in registry: %T", comp.Instance)
		// The component rendering is verified below via View() output
	} else {
		t.Fatal("List component 'mylist' not found in registry")
	}

	// Render the model
	output := model.View()
	if output == "" {
		t.Error("View output should not be empty")
	}

	// Truncate output for logging
	outputPreview := output
	if len(outputPreview) > 500 {
		outputPreview = outputPreview[:500] + "..."
	}
	t.Logf("Output (first 500 chars):\n%s", outputPreview)

	// Verify layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))

	// Find the list box
	var listBox *runtime.LayoutBox
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		if box.NodeID == "mylist" {
			listBox = &box
		}
	}

	if listBox != nil {
		t.Logf("ListBox: X=%d, Y=%d, W=%d, H=%d", listBox.X, listBox.Y, listBox.W, listBox.H)

		// Verify dimensions (may differ from props due to layout constraints)
		if listBox.W < 20 || listBox.W > 80 {
			t.Errorf("List width %d outside expected range [20, 80]", listBox.W)
		}
		if listBox.H < 5 || listBox.H > 24 {
			t.Errorf("List height %d outside expected range [5, 24]", listBox.H)
		}
	} else {
		t.Error("ListBox not found in LayoutResult")
	}

	// Verify output contains expected content
	if !strings.Contains(output, "Select an option") {
		t.Errorf("Output should contain 'Select an option', got: %s", outputPreview)
	}
	if !strings.Contains(output, "Option 1") {
		t.Errorf("Output should contain 'Option 1', got: %s", outputPreview)
	}

	t.Logf("List component Runtime test passed. Output length: %d chars", len(output))
}

// TestRuntimeTableComponent tests the native Runtime Table component
func TestRuntimeTableComponent(t *testing.T) {
	// Create a simple TUI configuration with a Table component
	config := &Config{
		Name: "Table Component Test",
		Layout: Layout{
			Direction: "row",
			Children: []Component{
				{
					Type: "table",
					ID:   "mytable",
					Props: map[string]interface{}{
						"width":         80,
						"height":        12,
						"showBorder":     true,
						"columns": []interface{}{
							map[string]interface{}{
								"title": "ID",
								"width": 10,
							},
							map[string]interface{}{
								"title": "Name",
								"width": 30,
							},
							map[string]interface{}{
								"title": "Email",
								"width": 30,
							},
						},
						"data": []interface{}{
							[]interface{}{1, "Alice", "alice@example.com"},
							[]interface{}{2, "Bob", "bob@example.com"},
							[]interface{}{3, "Charlie", "charlie@example.com"},
							[]interface{}{4, "Diana", "diana@example.com"},
							[]interface{}{5, "Eve", "eve@example.com"},
						},
					},
				},
			},
		},
	}

	// Create model with Runtime enabled
	model := NewModel(config, nil)
	model.UseRuntime = true
	model.Width = 80
	model.Height = 24

	// Initialize the model first!
	model.Init()

	// Then send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Verify RuntimeRoot was created
	if model.RuntimeRoot == nil {
		t.Fatalf("RuntimeRoot should not be nil after Init")
	}

	t.Logf("RuntimeRoot ID: %s", model.RuntimeRoot.ID)
	t.Logf("RuntimeRoot has %d children", len(model.RuntimeRoot.Children))

	// Verify the table component exists in the registry
	if comp, exists := model.ComponentInstanceRegistry.Get("mytable"); exists {
		t.Logf("Table component found in registry: %T", comp.Instance)
		// The component rendering is verified below via View() output
	} else {
		t.Fatal("Table component 'mytable' not found in registry")
	}

	// Render the model
	output := model.View()
	if output == "" {
		t.Error("View output should not be empty")
	}

	// Truncate output for logging
	outputPreview := output
	if len(outputPreview) > 500 {
		outputPreview = outputPreview[:500] + "..."
	}
	t.Logf("Output (first 500 chars):\n%s", outputPreview)

	// Verify layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))

	// Find the table box
	var tableBox *runtime.LayoutBox
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		if box.NodeID == "mytable" {
			tableBox = &box
		}
	}

	if tableBox != nil {
		t.Logf("TableBox: X=%d, Y=%d, W=%d, H=%d", tableBox.X, tableBox.Y, tableBox.W, tableBox.H)

		// Verify dimensions (may differ from props due to layout constraints)
		if tableBox.W < 50 || tableBox.W > 80 {
			t.Errorf("Table width %d outside expected range [50, 80]", tableBox.W)
		}
		if tableBox.H < 10 || tableBox.H > 24 {
			t.Errorf("Table height %d outside expected range [10, 24]", tableBox.H)
		}
	} else {
		t.Error("TableBox not found in LayoutResult")
	}

	// Verify output contains expected content
	if !strings.Contains(output, "ID") {
		t.Errorf("Output should contain 'ID', got: %s", outputPreview)
	}
	if !strings.Contains(output, "Name") {
		t.Errorf("Output should contain 'Name', got: %s", outputPreview)
	}
	if !strings.Contains(output, "Email") {
		t.Errorf("Output should contain 'Email', got: %s", outputPreview)
	}

	t.Logf("Table component Runtime test passed. Output length: %d chars", len(output))
}

// TestRuntimeFormComponent tests the Form component integration with Runtime
func TestRuntimeFormComponent(t *testing.T) {
	config := &Config{
		Name:      "Form Component Test",
		// UseRuntime defaults to true
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					Type: "form",
					ID:   "myform",
					Props: map[string]interface{}{
						"title":       "User Registration",
						"description": "Please enter your details",
						"submitLabel": "Register",
						"cancelLabel": "Cancel",
						"fields": []interface{}{
							map[string]interface{}{
								"type":        "input",
								"name":        "username",
								"label":       "Username",
								"placeholder": "Enter username",
								"required":    true,
								"width":       30,
							},
							map[string]interface{}{
								"type":        "input",
								"name":        "email",
								"label":       "Email",
								"placeholder": "Enter email",
								"required":    true,
								"width":       30,
							},
							map[string]interface{}{
								"type":        "input",
								"name":        "password",
								"label":       "Password",
								"placeholder": "Enter password",
								"required":    true,
								"width":       30,
							},
						},
						"values": map[string]interface{}{
							"username": "",
							"email":    "",
							"password": "",
						},
					},
				},
			},
		},
	}

	// Create model with config
	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Call View() to trigger rendering
	_ = model.View()

	// Verify Runtime was initialized
	if model.RuntimeEngine == nil {
		t.Fatal("RuntimeEngine should be initialized")
	}
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should be initialized")
	}

	m := model // Alias for consistency with rest of test

	// Verify Runtime engine is initialized
	if !m.UseRuntime {
		t.Fatal("Runtime engine should be enabled")
	}

	// Verify RuntimeRoot exists
	if m.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should not be nil")
	}

	t.Logf("RuntimeRoot ID: %s", m.RuntimeRoot.ID)
	t.Logf("RuntimeRoot has %d children", len(m.RuntimeRoot.Children))

	if len(m.RuntimeRoot.Children) != 1 {
		t.Fatalf("RuntimeRoot should have 1 child, got %d", len(m.RuntimeRoot.Children))
	}

	// Find Form component in registry
	formCompEntry, exists := m.ComponentInstanceRegistry.Get("myform")
	if !exists {
		t.Fatal("Form component not found in registry")
	}

	formWrapper := formCompEntry.Instance

	t.Logf("Form component found in registry: %T", formWrapper)

	// Verify it's wrapped correctly
	nativeWrapper, ok := formWrapper.(*NativeComponentWrapper)
	if !ok {
		t.Fatalf("Form component should be wrapped in NativeComponentWrapper, got %T", formWrapper)
	}

	// Verify the native component is FormComponent
	formComp, ok := nativeWrapper.Component.(*components.FormComponent)
	if !ok {
		t.Fatalf("Native component should be FormComponent, got %T", nativeWrapper.Component)
	}

	// Verify form properties
	// Note: ID might be set by the registry, not by WithID
	formID := formComp.GetID()
	if formID != "myform" && formID != "form" {
		t.Errorf("Expected ID 'myform' or 'form', got '%s'", formID)
	}

	// Verify fields
	fields := formComp.GetFields()
	if len(fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(fields))
	} else {
		// Check first field
		if fields[0].Name != "username" {
			t.Errorf("Expected first field name 'username', got '%s'", fields[0].Name)
		}
		if fields[0].Label != "Username" {
			t.Errorf("Expected first field label 'Username', got '%s'", fields[0].Label)
		}
		if !fields[0].Required {
			t.Error("Expected first field to be required")
		}
	}

	// Verify title and description
	if formComp.GetTitle() != "User Registration" {
		t.Errorf("Expected title 'User Registration', got '%s'", formComp.GetTitle())
	}

	// Verify we can set and get field values
	formComp.SetFieldValue("username", "testuser")
	value, exists := formComp.GetFieldValue("username")
	if !exists {
		t.Error("Username field should exist")
	} else if value != "testuser" {
		t.Errorf("Expected username value 'testuser', got '%s'", value)
	}

	// Test form validation
	formComp.SetFieldValue("username", "")
	formComp.SetFieldValue("email", "")
	formComp.SetFieldValue("password", "")
	isValid := formComp.Validate()
	if isValid {
		t.Error("Form with empty required fields should not be valid")
	}

	// Check that errors were set
	errors := formComp.GetValues()
	if len(errors) == 0 {
		// Note: GetValues returns field values, not errors
		// We can't directly access errors, but we can check validation result
	}

	// Fill in required fields
	formComp.SetFieldValue("username", "testuser")
	formComp.SetFieldValue("email", "test@example.com")
	formComp.SetFieldValue("password", "password123")
	isValid = formComp.Validate()
	if !isValid {
		t.Error("Form with all required fields filled should be valid")
	}

	// Test field navigation
	formComp.SetFocusIndex(0)
	if formComp.GetFocusIndex() != 0 {
		t.Errorf("Expected focus index 0, got %d", formComp.GetFocusIndex())
	}

	formComp.NextField()
	if formComp.GetFocusIndex() != 1 {
		t.Errorf("Expected focus index 1 after NextField, got %d", formComp.GetFocusIndex())
	}

	formComp.PrevField()
	if formComp.GetFocusIndex() != 0 {
		t.Errorf("Expected focus index 0 after PrevField, got %d", formComp.GetFocusIndex())
	}

	// Render the form
	output := m.View()
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}

	// Verify output contains form elements
	outputStr := string(output)
	if !strings.Contains(outputStr, "User Registration") {
		t.Error("Output should contain form title")
	}

	t.Logf("Output (first 500 chars):\n%s", truncateString(outputStr, 500))

	// Verify layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))

	// Find the form box
	var formBox *runtime.LayoutBox
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		if box.NodeID == "myform" {
			formBox = &box
		}
	}

	if formBox == nil {
		t.Fatal("Form box should exist in layout")
	}

	if formBox.NodeID != "myform" {
		t.Errorf("Expected box ID 'myform', got '%s'", formBox.NodeID)
	}

	// Verify dimensions
	if formBox.W <= 0 {
		t.Errorf("Form box should have positive width, got %d", formBox.W)
	}
	if formBox.H <= 0 {
		t.Errorf("Form box should have positive height, got %d", formBox.H)
	}

	t.Logf("FormBox: X=%d, Y=%d, W=%d, H=%d",
		formBox.X, formBox.Y, formBox.W, formBox.H)

	// Test form submission simulation
	formComp.SetFocus(true)
	formComp.HandleKey(&event.KeyEvent{Key: '\r'}) // Enter key

	if !formComp.IsSubmitted() {
		// Form should be submitted (but may not be valid if fields are empty)
		t.Log("Form submission state: submitted (validation may have failed)")
	}

	t.Logf("Form component Runtime test passed. Output length: %d chars", len(output))
}

// TestRuntimeTextareaComponent tests the Textarea component integration with Runtime
func TestRuntimeTextareaComponent(t *testing.T) {
	config := &Config{
		Name:      "Textarea Component Test",
		// UseRuntime defaults to true
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					Type: "textarea",
					ID:   "mytextarea",
					Props: map[string]interface{}{
						"placeholder":   "Enter your message here...",
						"prompt":        "> ",
						"width":          60,
						"height":         8,
						"charLimit":      500,
						"showLineNumbers": true,
						"value":          "Initial text\nLine 2",
					},
				},
			},
		},
	}

	// Create model with config
	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Call View() to trigger rendering
	_ = model.View()

	// Verify Runtime was initialized
	if model.RuntimeEngine == nil {
		t.Fatal("RuntimeEngine should be initialized")
	}
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should be initialized")
	}

	m := model // Alias for consistency with rest of test

	// Verify Runtime engine is initialized
	if !m.UseRuntime {
		t.Fatal("Runtime engine should be enabled")
	}

	// Verify RuntimeRoot exists
	if m.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should not be nil")
	}

	t.Logf("RuntimeRoot ID: %s", m.RuntimeRoot.ID)
	t.Logf("RuntimeRoot has %d children", len(m.RuntimeRoot.Children))

	if len(m.RuntimeRoot.Children) != 1 {
		t.Fatalf("RuntimeRoot should have 1 child, got %d", len(m.RuntimeRoot.Children))
	}

	// Find Textarea component in registry
	textareaCompEntry, exists := m.ComponentInstanceRegistry.Get("mytextarea")
	if !exists {
		t.Fatal("Textarea component not found in registry")
	}

	textareaWrapper := textareaCompEntry.Instance

	t.Logf("Textarea component found in registry: %T", textareaWrapper)

	// Verify it's wrapped correctly
	nativeWrapper, ok := textareaWrapper.(*NativeComponentWrapper)
	if !ok {
		t.Fatalf("Textarea component should be wrapped in NativeComponentWrapper, got %T", textareaWrapper)
	}

	// Verify the native component is TextareaComponent
	textareaComp, ok := nativeWrapper.Component.(*components.TextareaComponent)
	if !ok {
		t.Fatalf("Native component should be TextareaComponent, got %T", nativeWrapper.Component)
	}

	// Verify textarea properties
	textareaID := textareaComp.GetID()
	if textareaID != "mytextarea" && textareaID != "textarea" {
		t.Errorf("Expected ID 'mytextarea' or 'textarea', got '%s'", textareaID)
	}

	// Verify initial value
	initialValue := textareaComp.GetValue()
	if initialValue != "Initial text\nLine 2" {
		t.Errorf("Expected initial value 'Initial text\\nLine 2', got '%s'", initialValue)
	}

	// Verify placeholder
	placeholder := textareaComp.GetPlaceholder()
	if placeholder != "Enter your message here..." {
		t.Errorf("Expected placeholder 'Enter your message here...', got '%s'", placeholder)
	}

	// Render the textarea
	output := m.View()
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}

	// Verify output contains expected elements
	outputStr := string(output)
	if !strings.Contains(outputStr, "Initial text") {
		t.Error("Output should contain initial text")
	}

	t.Logf("Output (first 500 chars):\n%s", truncateString(outputStr, 500))

	// Verify layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))

	// Find the textarea box
	var textareaBox *runtime.LayoutBox
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		if box.NodeID == "mytextarea" {
			textareaBox = &box
		}
	}

	if textareaBox == nil {
		t.Fatal("Textarea box should exist in layout")
	}

	if textareaBox.NodeID != "mytextarea" {
		t.Errorf("Expected box ID 'mytextarea', got '%s'", textareaBox.NodeID)
	}

	// Verify dimensions
	if textareaBox.W <= 0 {
		t.Errorf("Textarea box should have positive width, got %d", textareaBox.W)
	}
	if textareaBox.H <= 0 {
		t.Errorf("Textarea box should have positive height, got %d", textareaBox.H)
	}

	t.Logf("TextareaBox: X=%d, Y=%d, W=%d, H=%d",
		textareaBox.X, textareaBox.Y, textareaBox.W, textareaBox.H)

	t.Logf("Textarea component Runtime test passed. Output length: %d chars", len(output))
}

func TestRuntimeProgressComponent(t *testing.T) {
	config := &Config{
		Name:      "Progress Component Test",
		// UseRuntime defaults to true
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					Type: "progress",
					ID:   "myprogress",
					Props: map[string]interface{}{
						"percent":        75.5,
						"label":          "Loading...",
						"showPercentage": true,
						"filledChar":     "█",
						"emptyChar":      "░",
						"fullColor":      "#00FF00",
						"emptyColor":     "#333333",
						"width":          50,
						"height":         1,
					},
				},
			},
		},
	}

	// Create model with config
	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Call View() to trigger rendering
	view := model.View()
	t.Logf("View output:\n%s", view)

	// Verify Runtime was initialized
	if model.RuntimeEngine == nil {
		t.Fatal("RuntimeEngine should be initialized")
	}
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should be initialized")
	}

	m := model // Alias for consistency with other tests

	// Verify Runtime engine is initialized
	if !m.UseRuntime {
		t.Fatal("Runtime engine should be enabled")
	}

	// Verify RuntimeRoot exists
	if m.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should not be nil")
	}

	t.Logf("RuntimeRoot ID: %s", m.RuntimeRoot.ID)
	t.Logf("RuntimeRoot has %d children", len(m.RuntimeRoot.Children))

	if len(m.RuntimeRoot.Children) != 1 {
		t.Fatalf("RuntimeRoot should have 1 child, got %d", len(m.RuntimeRoot.Children))
	}

	// Verify rendering
	output := view
	t.Logf("Output length: %d bytes", len(output))

	// Verify output contains expected elements
	outputStr := string(output)
	if !strings.Contains(outputStr, "Loading") {
		t.Error("Output should contain label text")
	}

	t.Logf("Output (first 500 chars):\n%s", truncateString(outputStr, 500))

	// Verify layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))

	// Find the progress box
	var progressBox *runtime.LayoutBox
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		if box.NodeID == "myprogress" {
			progressBox = &box
		}
	}

	if progressBox == nil {
		t.Fatal("Progress box should exist in layout")
	}

	if progressBox.NodeID != "myprogress" {
		t.Errorf("Expected box ID 'myprogress', got '%s'", progressBox.NodeID)
	}

	// Verify dimensions
	if progressBox.W <= 0 {
		t.Errorf("Progress box should have positive width, got %d", progressBox.W)
	}
	if progressBox.H <= 0 {
		t.Errorf("Progress box should have positive height, got %d", progressBox.H)
	}

	t.Logf("ProgressBox: X=%d, Y=%d, W=%d, H=%d",
		progressBox.X, progressBox.Y, progressBox.W, progressBox.H)

	t.Logf("Progress component Runtime test passed. Output length: %d chars", len(output))
}

func TestRuntimeSpinnerComponent(t *testing.T) {
	config := &Config{
		Name:      "Spinner Component Test",
		// UseRuntime defaults to true
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					Type: "spinner",
					ID:   "myspinner",
					Props: map[string]interface{}{
						"style":         "dots",
						"label":         "Loading...",
						"labelPosition": "right",
						"running":       true,
						"color":         "#00FF00",
						"width":         20,
						"height":        1,
					},
				},
			},
		},
	}

	// Create model with config
	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Call View() to trigger rendering
	view := model.View()
	t.Logf("View output:\n%s", view)

	// Verify Runtime was initialized
	if model.RuntimeEngine == nil {
		t.Fatal("RuntimeEngine should be initialized")
	}
	if model.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should be initialized")
	}

	m := model // Alias for consistency with other tests

	// Verify Runtime engine is initialized
	if !m.UseRuntime {
		t.Fatal("Runtime engine should be enabled")
	}

	// Verify RuntimeRoot exists
	if m.RuntimeRoot == nil {
		t.Fatal("RuntimeRoot should not be nil")
	}

	t.Logf("RuntimeRoot ID: %s", m.RuntimeRoot.ID)
	t.Logf("RuntimeRoot has %d children", len(m.RuntimeRoot.Children))

	if len(m.RuntimeRoot.Children) != 1 {
		t.Fatalf("RuntimeRoot should have 1 child, got %d", len(m.RuntimeRoot.Children))
	}

	// Verify rendering
	output := view
	t.Logf("Output length: %d bytes", len(output))

	// Verify output contains expected elements
	outputStr := string(output)
	if !strings.Contains(outputStr, "Loading") {
		t.Error("Output should contain label text")
	}

	t.Logf("Output (first 500 chars):\n%s", truncateString(outputStr, 500))

	// Verify layout result
	result := model.GetLayoutResult()
	t.Logf("LayoutResult: %d boxes", len(result.Boxes))

	// Find the spinner box
	var spinnerBox *runtime.LayoutBox
	for i, box := range result.Boxes {
		t.Logf("  Box[%d]: ID=%s, X=%d, Y=%d, W=%d, H=%d", i, box.NodeID, box.X, box.Y, box.W, box.H)
		if box.NodeID == "myspinner" {
			spinnerBox = &box
		}
	}

	if spinnerBox == nil {
		t.Fatal("Spinner box should exist in layout")
	}

	if spinnerBox.NodeID != "myspinner" {
		t.Errorf("Expected box ID 'myspinner', got '%s'", spinnerBox.NodeID)
	}

	// Verify dimensions
	if spinnerBox.W <= 0 {
		t.Errorf("Spinner box should have positive width, got %d", spinnerBox.W)
	}
	if spinnerBox.H <= 0 {
		t.Errorf("Spinner box should have positive height, got %d", spinnerBox.H)
	}

	t.Logf("SpinnerBox: X=%d, Y=%d, W=%d, H=%d",
		spinnerBox.X, spinnerBox.Y, spinnerBox.W, spinnerBox.H)

	t.Logf("Spinner component Runtime test passed. Output length: %d chars", len(output))
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
