package layout_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/layout"
)

// MockShrinkableComponent 模拟一个组件
type MockShrinkableComponent struct {
	id     string
	width  int
	height int
	typ    string
}

// 确保实现了接口
var _ core.ComponentInterface = (*MockShrinkableComponent)(nil)
var _ core.Measurable = (*MockShrinkableComponent)(nil)

func NewMockShrinkableComponent(id string, width, height int) *core.ComponentInstance {
	mock := &MockShrinkableComponent{
		id:     id,
		width:  width,
		height: height,
		typ:    "mock-shrinkable",
	}
	return &core.ComponentInstance{
		ID:       id,
		Type:     "mock-shrinkable",
		Instance: mock,
	}
}

func (m *MockShrinkableComponent) View() string {
	return "mock"
}

func (m *MockShrinkableComponent) Init() tea.Cmd {
	return nil
}

func (m *MockShrinkableComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return m, nil, core.Ignored
}

func (m *MockShrinkableComponent) GetID() string {
	return m.id
}

func (m *MockShrinkableComponent) SetFocus(focus bool) {}

func (m *MockShrinkableComponent) GetFocus() bool {
	return false
}

func (m *MockShrinkableComponent) GetComponentType() string {
	return m.typ
}

func (m *MockShrinkableComponent) Render(config core.RenderConfig) (string, error) {
	return strings.Repeat("x", m.width), nil
}

func (m *MockShrinkableComponent) UpdateRenderConfig(config core.RenderConfig) error {
	return nil
}

func (m *MockShrinkableComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (m *MockShrinkableComponent) GetSubscribedMessageTypes() []string {
	return nil
}

func (m *MockShrinkableComponent) Cleanup() {}

// 实现 Measurable 接口
func (m *MockShrinkableComponent) Measure(maxWidth, maxHeight int) (int, int) {
	return m.width, m.height
}

func TestLayoutWithShrink(t *testing.T) {
	// 创建一个包含 flex 组件的布局
	// 注意：当前引擎中，shrink 是通过 "flex" 属性隐式启用的

	childNode1 := &layout.LayoutNode{
		ID:        "child1",
		Type:      layout.LayoutFlex,
		Component: NewMockShrinkableComponent("child1", 60, 20),
		Style: &layout.LayoutStyle{
			Width: layout.NewSize("flex"), // 启用 flex/shrink
		},
	}

	childNode2 := &layout.LayoutNode{
		ID:        "child2",
		Type:      layout.LayoutFlex,
		Component: NewMockShrinkableComponent("child2", 60, 20),
		Style: &layout.LayoutStyle{
			Width: layout.NewSize("flex"), // 启用 flex/shrink
		},
	}

	root := &layout.LayoutNode{
		ID:   "root",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Direction: layout.DirectionRow,
		},
		Children: []*layout.LayoutNode{childNode1, childNode2},
	}

	// 容器宽度 100
	engine := layout.NewEngine(&layout.LayoutConfig{
		Root:       root,
		WindowSize: &layout.WindowSize{Width: 100, Height: 30},
	})

	result := engine.Layout()

	// 验证布局结果
	// 两个 flex item 应该平分空间 (50 each)

	// 查找子节点
	var foundChild1, foundChild2 *layout.LayoutNode
	for _, node := range result.Nodes {
		if node.ID == "child1" {
			foundChild1 = node
		} else if node.ID == "child2" {
			foundChild2 = node
		}
	}

	assert.NotNil(t, foundChild1)
	assert.NotNil(t, foundChild2)

	assert.Equal(t, 50, foundChild1.Bound.Width)
	assert.Equal(t, 50, foundChild2.Bound.Width)
}

// TestFixedElementShrink 测试固定宽度的元素在设置 Shrink > 0 时是否收缩
func TestFixedElementShrink(t *testing.T) {
	childNode1 := &layout.LayoutNode{
		ID:   "child1",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Width:  layout.NewSize(50),
			Shrink: layout.NewGrow(1),
		},
		Component: NewMockShrinkableComponent("child1", 50, 20),
	}

	childNode2 := &layout.LayoutNode{
		ID:   "child2",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Width:  layout.NewSize(50),
			Shrink: layout.NewGrow(1),
		},
		Component: NewMockShrinkableComponent("child2", 50, 20),
	}

	root := &layout.LayoutNode{
		ID:   "root",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Direction: layout.DirectionRow,
		},
		Children: []*layout.LayoutNode{childNode1, childNode2},
	}

	// 容器宽 80，内容总宽 100，溢出 20
	engine := layout.NewEngine(&layout.LayoutConfig{
		Root:       root,
		WindowSize: &layout.WindowSize{Width: 80, Height: 30},
	})

	result := engine.Layout()

	var foundChild1, foundChild2 *layout.LayoutNode
	for _, node := range result.Nodes {
		if node.ID == "child1" {
			foundChild1 = node
		} else if node.ID == "child2" {
			foundChild2 = node
		}
	}

	assert.NotNil(t, foundChild1)
	assert.NotNil(t, foundChild2)

	// 期望每个收缩 10
	assert.Equal(t, 40, foundChild1.Bound.Width)
	assert.Equal(t, 40, foundChild2.Bound.Width)
}

// TestFixedElementDifferentShrink 测试不同 shrink 值导致的收缩量不同
func TestFixedElementDifferentShrink(t *testing.T) {
	childNode1 := &layout.LayoutNode{
		ID:   "child1",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Width:  layout.NewSize(50),
			Shrink: layout.NewGrow(2), // 更容易收缩
		},
		Component: NewMockShrinkableComponent("child1", 50, 20),
	}

	childNode2 := &layout.LayoutNode{
		ID:   "child2",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Width:  layout.NewSize(50),
			Shrink: layout.NewGrow(1),
		},
		Component: NewMockShrinkableComponent("child2", 50, 20),
	}

	root := &layout.LayoutNode{
		ID:   "root",
		Type: layout.LayoutFlex,
		Style: &layout.LayoutStyle{
			Direction: layout.DirectionRow,
		},
		Children: []*layout.LayoutNode{childNode1, childNode2},
	}

	// 容器宽 80，内容总宽 100，溢出 20
	engine := layout.NewEngine(&layout.LayoutConfig{
		Root:       root,
		WindowSize: &layout.WindowSize{Width: 80, Height: 30},
	})

	result := engine.Layout()

	var foundChild1, foundChild2 *layout.LayoutNode
	for _, node := range result.Nodes {
		if node.ID == "child1" {
			foundChild1 = node
		} else if node.ID == "child2" {
			foundChild2 = node
		}
	}

	assert.NotNil(t, foundChild1)
	assert.NotNil(t, foundChild2)

	// Child1 收缩 20 * (2/3) ≈ 13.33 -> 36.67
	// Child2 收缩 20 * (1/3) ≈ 6.67 -> 43.33

	// 使用 InDelta 允许一定的整数舍入误差
	assert.InDelta(t, 37, foundChild1.Bound.Width, 1)
	assert.InDelta(t, 43, foundChild2.Bound.Width, 1)
	assert.Less(t, foundChild1.Bound.Width, foundChild2.Bound.Width)
}
