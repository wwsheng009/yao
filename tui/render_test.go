package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
)

func TestApplyState(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Hello World",
			"count": 42,
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple substitution",
			input:    "Title: {{title}}",
			expected: "Title: Hello World",
		},
		{
			name:     "number substitution",
			input:    "Count: {{count}}",
			expected: "Count: 42",
		},
		{
			name:     "multiple substitutions",
			input:    "{{title}} - Count: {{count}}",
			expected: "Hello World - Count: 42",
		},
		{
			name:     "non-existent key",
			input:    "Value: {{nonexistent}}",
			expected: "Value: ",
		},
		{
			name:     "no substitution",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "empty template",
			input:    "{{}}",
			expected: "{{}}", // 空的 {{}} 不会被替换
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.applyState(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStateValue(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Hello",
			"user": map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{
			name:     "simple key",
			key:      "title",
			expected: "Hello",
		},
		{
			name:     "nested key",
			key:      "user.name",
			expected: "Alice",
		},
		{
			name:     "nested number",
			key:      "user.age",
			expected: 30,
		},
		{
			name:     "non-existent key",
			key:      "missing",
			expected: nil,
		},
		{
			name:     "non-existent nested",
			key:      "user.missing",
			expected: nil,
		},
		{
			name:     "empty key",
			key:      "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := model.getStateValue(tt.key)
			var result interface{}
			if exists {
				result = value
			} else {
				result = nil
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyStateToProps(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Dynamic Title",
			"users": []string{"Alice", "Bob"},
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	t.Run("substitute string props", func(t *testing.T) {
		comp := &Component{
			Type: "header",
			Props: map[string]interface{}{
				"title": "Welcome: {{title}}",
				"count": 10,
			},
		}

		props := model.applyStateToProps(comp)
		assert.Equal(t, "Welcome: Dynamic Title", props["title"])
		assert.Equal(t, 10, props["count"])
	})

	t.Run("bind attribute", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Bind: "users",
			Props: map[string]interface{}{
				"style": "list",
			},
		}

		props := model.applyStateToProps(comp)
		assert.Equal(t, "list", props["style"])
		assert.Equal(t, []string{"Alice", "Bob"}, props["__bind_data"])
	})

	t.Run("nil props", func(t *testing.T) {
		comp := &Component{
			Type: "text",
		}

		props := model.applyStateToProps(comp)
		assert.NotNil(t, props)
		assert.Empty(t, props)
	})
}

func TestRenderHeaderComponent(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80

	// Test using the new renderer system
	comp := &Component{
		Type: "header",
		Props: map[string]interface{}{
			"title": "Test Header",
		},
	}

	result := model.RenderComponent(comp)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Test Header")
}

func TestRenderTextComponent(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80

	t.Run("with content", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Props: map[string]interface{}{
				"content": "Hello World",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Hello World")
	})

	t.Run("with bind data", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Props: map[string]interface{}{
				"__bind_data": "Bound content",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Bound content")
	})

	t.Run("with alignment", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			Props: map[string]interface{}{
				"content": "Centered",
				"align":   "center",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
	})
}

func TestRenderComponent(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Dynamic",
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80
	model.Height = 24

	t.Run("header component", func(t *testing.T) {
		comp := &Component{
			Type: "header",
			Props: map[string]interface{}{
				"title": "{{title}}",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Dynamic")
	})

	t.Run("text component", func(t *testing.T) {
		comp := &Component{
			Type: "text",
			ID:   "test-text-component",
			Props: map[string]interface{}{
				"content": "Some text",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Some text")
	})

	t.Run("unknown component", func(t *testing.T) {
		comp := &Component{
			Type: "unknown",
			Props: map[string]interface{}{
				"foo": "bar",
			},
		}

		result := model.RenderComponent(comp)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Unknown component")
	})

	t.Run("nil component", func(t *testing.T) {
		result := model.RenderComponent(nil)
		assert.Empty(t, result)
	})

	t.Run("empty type", func(t *testing.T) {
		comp := &Component{
			Props: map[string]interface{}{
				"foo": "bar",
			},
		}

		result := model.RenderComponent(comp)
		assert.Empty(t, result)
	})
}

func TestRenderLayout(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Test App",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					ID:   "test-layout-header",
					Props: map[string]interface{}{
						"title": "{{title}}",
					},
				},
				{
					Type: "text",
					ID:   "test-layout-text",
					Props: map[string]interface{}{
						"content": "Welcome to the app",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Width = 80
	model.Height = 24
	model.Ready = true

	// Initialize components
	model.InitializeComponents()

	result := model.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Test App")
	assert.Contains(t, result, "Welcome to the app")
}

func TestApplyPadding(t *testing.T) {
	tests := []struct {
		name    string
		padding []int
		content string
	}{
		{
			name:    "no padding",
			padding: []int{},
			content: "test",
		},
		{
			name:    "all sides equal",
			padding: []int{1},
			content: "test",
		},
		{
			name:    "vertical and horizontal",
			padding: []int{1, 2},
			content: "test",
		},
		{
			name:    "top, horizontal, bottom",
			padding: []int{1, 2, 3},
			content: "test",
		},
		{
			name:    "all sides different",
			padding: []int{1, 2, 3, 4},
			content: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyPadding(tt.content, tt.padding)
			assert.NotEmpty(t, result)
		})
	}
}

func TestGetPropHelpers(t *testing.T) {
	props := map[string]interface{}{
		"string": "value",
		"int":    42,
		"float":  3.14,
		"bool":   true,
	}

	t.Run("getStringProp", func(t *testing.T) {
		assert.Equal(t, "value", getStringProp(props, "string", "default"))
		assert.Equal(t, "default", getStringProp(props, "missing", "default"))
		assert.Equal(t, "default", getStringProp(props, "int", "default"))
	})

	t.Run("getIntProp", func(t *testing.T) {
		assert.Equal(t, 42, getIntProp(props, "int", 0))
		assert.Equal(t, 3, getIntProp(props, "float", 0))
		assert.Equal(t, 99, getIntProp(props, "missing", 99))
	})

	t.Run("getBoolProp", func(t *testing.T) {
		assert.Equal(t, true, getBoolProp(props, "bool", false))
		assert.Equal(t, false, getBoolProp(props, "missing", false))
		assert.Equal(t, false, getBoolProp(props, "string", false))
	})
}

func TestWindowSizeMsgHandling(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(*Model)
	assert.NotNil(t, m)
	assert.Equal(t, 80, m.Width)
	assert.Equal(t, 24, m.Height)
	assert.True(t, m.Ready)
}

func TestLayoutEngineUpdateWindowSize(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	if model.LayoutEngine != nil {
		model.LayoutEngine.UpdateWindowSize(100, 30)
		layoutResult := model.LayoutEngine.Layout()
		assert.NotNil(t, layoutResult)
	}
}

func TestFlexLayoutAdaptation(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "horizontal",
			Children: []Component{
				{
					Type: "text",
					ID:   "flex1",
					Props: map[string]interface{}{
						"content": "Left",
					},
				},
				{
					Type: "text",
					ID:   "flex2",
					Props: map[string]interface{}{
						"content": "Right",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	result := m.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Left")
	assert.Contains(t, result, "Right")

	msg = tea.WindowSizeMsg{Width: 120, Height: 30}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)
}

func TestVerticalLayoutAdaptation(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					ID:   "header1",
					Props: map[string]interface{}{
						"title": "Title 1",
					},
				},
				{
					Type: "text",
					ID:   "text1",
					Props: map[string]interface{}{
						"content": "Content",
					},
				},
				{
					Type: "header",
					ID:   "header2",
					Props: map[string]interface{}{
						"title": "Title 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	result := m.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Title 1")
	assert.Contains(t, result, "Content")
	assert.Contains(t, result, "Title 2")

	msg = tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)
}

func TestHorizontalLayoutAdaptation(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "horizontal",
			Children: []Component{
				{
					Type: "text",
					ID:   "item1",
					Props: map[string]interface{}{
						"content": "Item 1",
					},
				},
				{
					Type: "text",
					ID:   "item2",
					Props: map[string]interface{}{
						"content": "Item 2",
					},
				},
				{
					Type: "text",
					ID:   "item3",
					Props: map[string]interface{}{
						"content": "Item 3",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	result := m.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Item 1")
	assert.Contains(t, result, "Item 2")
	assert.Contains(t, result, "Item 3")

	msg = tea.WindowSizeMsg{Width: 40, Height: 24}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)
}

func TestMultiLevelLayoutAdaptation(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					ID:   "main-header",
					Props: map[string]interface{}{
						"title": "Main Header",
					},
				},
				{
					Type: "text",
					ID:   "text1",
					Props: map[string]interface{}{
						"content": "Text Panel 1",
					},
				},
				{
					Type: "text",
					ID:   "text2",
					Props: map[string]interface{}{
						"content": "Text Panel 2",
					},
				},
				{
					Type: "text",
					ID:   "text3",
					Props: map[string]interface{}{
						"content": "Text Panel 3",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	result := m.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Main Header")
	assert.Contains(t, result, "Text Panel 1")
	assert.Contains(t, result, "Text Panel 2")
	assert.Contains(t, result, "Text Panel 3")

	msg = tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)
}

func TestDifferentWindowSizeRendering(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					ID:   "content",
					Props: map[string]interface{}{
						"content": "This is a long text that should adapt to different window sizes",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	testCases := []struct {
		width  int
		height int
	}{
		{60, 20},
		{80, 24},
		{100, 30},
		{120, 40},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("width_%d_height_%d", tc.width, tc.height), func(t *testing.T) {
			msg := tea.WindowSizeMsg{Width: tc.width, Height: tc.height}
			updatedModel, _ := model.Update(msg)
			m := updatedModel.(*Model)

			result := m.RenderLayout()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "This is a long text")
		})
	}
}

func TestLayoutWithPaddingAdaptation(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
			Padding:   []int{1, 2, 1, 2},
			Children: []Component{
				{
					Type: "text",
					ID:   "content",
					Props: map[string]interface{}{
						"content": "Content with padding",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	result := m.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Content with padding")

	msg = tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)
}

func TestComplexLayoutAdaptation(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					ID:   "header",
					Props: map[string]interface{}{
						"title": "Header",
					},
				},
				{
					Type: "text",
					ID:   "sidebar",
					Props: map[string]interface{}{
						"content": "Sidebar",
					},
				},
				{
					Type: "text",
					ID:   "text1",
					Props: map[string]interface{}{
						"content": "Main Text 1",
					},
				},
				{
					Type: "text",
					ID:   "text2",
					Props: map[string]interface{}{
						"content": "Main Text 2",
					},
				},
				{
					Type: "header",
					ID:   "footer",
					Props: map[string]interface{}{
						"title": "Footer",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	result := m.RenderLayout()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Header")
	assert.Contains(t, result, "Sidebar")
	assert.Contains(t, result, "Main Text 1")
	assert.Contains(t, result, "Main Text 2")
	assert.Contains(t, result, "Footer")

	msg = tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)
}

func TestTUIListWithFlexLayout(t *testing.T) {
	items := make([]interface{}, 0, 10)
	for i := 1; i <= 10; i++ {
		items = append(items, map[string]interface{}{
			"id":      fmt.Sprintf("tui-%d", i),
			"name":    fmt.Sprintf("TUI Configuration %d", i),
			"title":   fmt.Sprintf("tui-%d - TUI Configuration %d", i, i),
			"command": fmt.Sprintf("yao tui tui-%d", i),
		})
	}

	cfg := &Config{
		Name: "TUI List",
		Data: map[string]interface{}{
			"title":       "Available TUI Configurations",
			"description": "Select a TUI to run",
			"items":       items,
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					Props: map[string]interface{}{
						"title": "{{title}}",
						"align": "center",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "{{description}}",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Total items: {{len(items)}}",
					},
				},
				{
					Type:   "list",
					ID:     "tuiList",
					Bind:   "items",
					Height: "flex",
					Width:  "flex",
					Props: map[string]interface{}{
						"showTitle":        true,
						"showStatusBar":    true,
						"showFilter":       true,
						"filteringEnabled": true,
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Press Enter to show command for selected TUI, Esc or q to quit",
					},
				},
			},
		},
		Bindings: map[string]core.Action{
			"q":      {Process: "tui.quit"},
			"ctrl+c": {Process: "tui.quit"},
		},
	}

	model := NewModel(cfg, nil)
	model.InitializeComponents()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(*Model)

	if m.LayoutRoot != nil {
		fmt.Printf("\n=== Layout tree info ===\n")
		fmt.Printf("Root ID: %s\n", m.LayoutRoot.ID)
		fmt.Printf("Root children count: %d\n", len(m.LayoutRoot.Children))
		for i, child := range m.LayoutRoot.Children {
			fmt.Printf("  Child %d: ID=%s, Type=%s, Bound=%+v\n", i, child.ID, child.ComponentType, child.Bound)
			if len(child.Children) > 0 {
				fmt.Printf("    Grandchildren count: %d\n", len(child.Children))
				for j, gc := range child.Children {
					fmt.Printf("      Grandchild %d: ID=%s, Type=%s, Bound=%+v\n", j, gc.ID, gc.ComponentType, gc.Bound)
					if gc.Component != nil {
						fmt.Printf("        Component instance exists: %v\n", gc.Component.Instance != nil)
						if gc.Component.Instance != nil && gc.Component.Instance.GetComponentType() == "list" {
							fmt.Printf("        List component props: %v\n", gc.Component.LastConfig.Data)
						}
					}
				}
			}
		}
		fmt.Printf("=== End ===\n\n")
	}

	allComponents := model.ComponentInstanceRegistry.GetAll()
	fmt.Printf("\n=== Component registry ===\n")
	fmt.Printf("Total components: %d\n", len(allComponents))
	for id, comp := range allComponents {
		fmt.Printf("  Component ID: %s, Type: %s, Width: %d, Height: %d\n", id, comp.Type, comp.LastConfig.Width, comp.LastConfig.Height)
		if id == "comp_header_0" {
			fmt.Printf("    Header props: %v\n", comp.LastConfig.Data)
		}
		if id == "comp_text_1" {
			fmt.Printf("    Text props: %v\n", comp.LastConfig.Data)
		}
		if id == "comp_text_2" {
			fmt.Printf("    Text2 props: %v\n", comp.LastConfig.Data)
		}
		if id == "comp_text_4" {
			fmt.Printf("    Text4 props: %v\n", comp.LastConfig.Data)
		}
	}
	fmt.Printf("=== End ===\n\n")

	result := m.RenderLayout()
	fmt.Printf("\n=== Rendered output (80x24) ===\n%s\n=== End ===\n", result)

	lines := countLines(result)
	fmt.Printf("\n=== Layout info ===\n")
	fmt.Printf("Window size: %dx%d\n", 80, 24)
	fmt.Printf("Rendered lines: %d\n", lines)
	fmt.Printf("Lines match window height: %v\n", lines <= 24)
	fmt.Printf("=== End ===\n\n")

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Available TUI Configurations", "Header should be rendered")
	assert.Contains(t, result, "Select a TUI to run", "Description text should be rendered")
	assert.Contains(t, result, "Total items: 10", "Count text should be rendered")
	assert.True(t, lines <= 24, fmt.Sprintf("Rendered output (%d lines) should fit within window height of 24", lines))

	msg = tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)

	lines = countLines(result)
	assert.True(t, lines <= 40, fmt.Sprintf("Rendered output (%d lines) should fit within window height of 40", lines))

	msg = tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	result = m.RenderLayout()
	assert.NotEmpty(t, result)

	lines = countLines(result)
	assert.True(t, lines <= 30, fmt.Sprintf("Rendered output (%d lines) should fit within window height of 30", lines))
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	count := 0
	for _, r := range s {
		if r == '\n' {
			count++
		}
	}
	if len(s) > 0 && s[len(s)-1] != '\n' {
		count++
	}
	return count
}
