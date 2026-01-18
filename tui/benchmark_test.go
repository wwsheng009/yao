package tui

import (
	"testing"

	"github.com/yaoapp/yao/tui/components"
	"github.com/yaoapp/yao/tui/core"
)

// BenchmarkComponentCreation benchmarks the creation performance of all component types
func BenchmarkComponentCreation(b *testing.B) {
	config := core.RenderConfig{
		Data: map[string]interface{}{
			"title": "Test",
			"value": 42,
		},
		Width:  80,
		Height: 24,
	}

	b.Run("TableComponent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = components.NewTableComponent(config, "test_table")
		}
	})

	b.Run("MenuComponent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = components.NewMenuComponent(config, "test_menu")
		}
	})

	b.Run("HeaderComponent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = components.NewHeaderComponent(config, "test_header")
		}
	})

	b.Run("TextComponent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = components.NewTextComponent(config, "test_text")
		}
	})
}

// BenchmarkConfigUpdate benchmarks the UpdateRenderConfig performance
func BenchmarkConfigUpdate(b *testing.B) {
	config := core.RenderConfig{
		Data: map[string]interface{}{
			"title": "Initial",
			"value": 42,
		},
		Width:  80,
		Height: 24,
	}

	updatedConfig := core.RenderConfig{
		Data: map[string]interface{}{
			"title": "Updated",
			"value": 84,
		},
		Width:  100,
		Height: 30,
	}

	// Create a component instance
	component := components.NewTableComponent(config, "benchmark_table")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := component.UpdateRenderConfig(updatedConfig)
		if err != nil {
			b.Fatalf("UpdateRenderConfig failed: %v", err)
		}
	}
}

// BenchmarkLargeDataTable benchmarks table rendering with large datasets
func BenchmarkLargeDataTable(b *testing.B) {
	// Create large dataset
	data := make([][]interface{}, 1000)
	for i := range data {
		row := make([]interface{}, 10)
		for j := range row {
			row[j] = i*10 + j
		}
		data[i] = row
	}

	config := core.RenderConfig{
		Data: map[string]interface{}{
			"data": data,
			"columns": []map[string]interface{}{
				{"key": "col0", "title": "Column 0"},
				{"key": "col1", "title": "Column 1"},
				{"key": "col2", "title": "Column 2"},
				{"key": "col3", "title": "Column 3"},
				{"key": "col4", "title": "Column 4"},
			},
		},
		Width:  120,
		Height: 40,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component := components.NewTableComponent(config, "large_table")
		_, err := component.Render(config)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

// BenchmarkExpressionEvaluation benchmarks the performance of expression evaluation with caching
func BenchmarkExpressionEvaluation(b *testing.B) {
	model := &Model{
		State:     make(map[string]interface{}),
		exprCache: NewExpressionCache(),
	}

	model.State["username"] = "testuser"
	model.State["count"] = 42

	b.Run("WithCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, success := model.resolveExpressionValue("username")
			if !success {
				b.Fatal("Expression resolution failed")
			}
		}
	})

	b.Run("WithComplexExpression", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, success := model.resolveExpressionValue("count > 20")
			if !success {
				b.Fatal("Expression resolution failed")
			}
		}
	})
}

// BenchmarkPropsResolution benchmarks the performance of props resolution with caching
func BenchmarkPropsResolution(b *testing.B) {
	model := &Model{
		State:      make(map[string]interface{}),
		propsCache: NewPropsCache(),
	}

	model.State["title"] = "Test Title"
	model.State["value"] = 100

	comp := &Component{
		ID:   "test_comp",
		Type: "header",
		Props: map[string]interface{}{
			"text": "{{title}} - {{value}}",
			"size": "large",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolved := model.resolveProps(comp)
		if resolved == nil {
			b.Fatal("Props resolution failed")
		}
	}
}
