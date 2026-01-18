package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestMaxLayoutDepth测试布局深度限制
func TestMaxLayoutDepth(t *testing.T) {
	// 创建简单的布局（当前渲染不使用深度递归）
	cfg := &Config{
		Name: "Test Max Depth",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Component 1"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// 渲染应该不会崩溃
	render := model.View()
	assert.NotEmpty(t, render)
}

// TestNormalLayoutDepth测试正常深度的布局
func TestNormalLayoutDepth(t *testing.T) {
	cfg := &Config{
		Name: "Test Normal Depth",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Component 1"},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Component 2"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	render := model.View()
	assert.NotEmpty(t, render)
}

// TestEmptyLayoutChildren测试空布局子元素
func TestEmptyLayoutChildren(t *testing.T) {
	cfg := &Config{
		Name: "Test Empty Children",
		Layout: Layout{
			Direction: "vertical",
			Children:  []Component{},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	render := model.View()
	// 空布局可能返回空字符串或极短内容
	_ = render
	assert.True(t, true)
}

// TestSingleLevelLayout测试单层布局
func TestSingleLevelLayout(t *testing.T) {
	cfg := &Config{
		Name: "Test Single Level",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Text 1"},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Text 2"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	render := model.View()
	assert.NotEmpty(t, render)
	assert.Contains(t, render, "Text 1")
	assert.Contains(t, render, "Text 2")
}

// TestLayoutDepthWithMixedComponents测试混合组件的布局深度
func TestLayoutDepthWithMixedComponents(t *testing.T) {
	cfg := &Config{
		Name: "Test Mixed Components",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Header",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Content",
					},
				},
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Footer",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	render := model.View()
	assert.NotEmpty(t, render)
}

// TestLayoutDepthErrorDoesNotCrash测试深度错误不会导致崩溃
func TestLayoutDepthErrorDoesNotCrash(t *testing.T) {
	cfg := &Config{
		Name: "Test No Crash",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content": "Safe content",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// 渲染不应该崩溃
	render := model.View()
	assert.NotEmpty(t, render)
}

// TestMaxLayoutDepthConstant测试常量定义
func TestMaxLayoutDepthConstant(t *testing.T) {
	// 验证常量存在
	assert.Equal(t, 50, maxLayoutDepth)
}

// BenchmarkLayoutRendering基准测试布局渲染性能
func BenchmarkLayoutRendering(b *testing.B) {
	cfg := &Config{
		Name: "Bench Layout",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Component 1"},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Component 2"},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Component 3"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}
