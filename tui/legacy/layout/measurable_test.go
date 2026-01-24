package layout

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
)

// MockMeasurableComponent 实现 Measurable 接口用于测试
type MockMeasurableComponent struct {
	preferredWidth  int
	preferredHeight int
}

func (m *MockMeasurableComponent) View() string {
	return "mock"
}

func (m *MockMeasurableComponent) Init() tea.Cmd {
	return nil
}

func (m *MockMeasurableComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return m, nil, core.Ignored
}

func (m *MockMeasurableComponent) GetID() string {
	return "mock"
}

func (m *MockMeasurableComponent) SetFocus(focus bool) {}

func (m *MockMeasurableComponent) SetSize(width, height int) {
	// 空实现
}

func (m *MockMeasurableComponent) GetFocus() bool {
	return false
}

func (m *MockMeasurableComponent) GetComponentType() string {
	return "mock"
}

func (m *MockMeasurableComponent) Render(config core.RenderConfig) (string, error) {
	return "mock", nil
}

func (m *MockMeasurableComponent) UpdateRenderConfig(config core.RenderConfig) error {
	return nil
}

func (m *MockMeasurableComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (m *MockMeasurableComponent) GetSubscribedMessageTypes() []string {
	return nil
}

func (m *MockMeasurableComponent) Cleanup() {}

// 实现 Measurable 接口
func (m *MockMeasurableComponent) Measure(maxWidth, maxHeight int) (int, int) {
	w := m.preferredWidth
	if w > maxWidth && maxWidth > 0 {
		w = maxWidth
	}
	h := m.preferredHeight
	if h > maxHeight && maxHeight > 0 {
		h = maxHeight
	}
	return w, h
}

func TestMeasurableInterface(t *testing.T) {
	tests := []struct {
		name            string
		maxWidth        int
		maxHeight       int
		preferredWidth  int
		preferredHeight int
		expectedWidth   int
		expectedHeight  int
	}{
		{
			name:            "理想尺寸小于约束",
			maxWidth:        100,
			maxHeight:       50,
			preferredWidth:  80,
			preferredHeight: 30,
			expectedWidth:   80,
			expectedHeight:  30,
		},
		{
			name:            "理想尺寸大于约束",
			maxWidth:        50,
			maxHeight:       20,
			preferredWidth:  80,
			preferredHeight: 30,
			expectedWidth:   50, // 限制在 maxWidth
			expectedHeight:  20, // 限制在 maxHeight
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := &MockMeasurableComponent{
				preferredWidth:  tt.preferredWidth,
				preferredHeight: tt.preferredHeight,
			}

			width, height := component.Measure(tt.maxWidth, tt.maxHeight)

			if width != tt.expectedWidth {
				t.Errorf("Measure() width = %v, want %v", width, tt.expectedWidth)
			}
			if height != tt.expectedHeight {
				t.Errorf("Measure() height = %v, want %v", height, tt.expectedHeight)
			}
		})
	}
}

func TestLayoutWithMeasurableComponent(t *testing.T) {
	// 创建一个包含 Measurable 组件的简单布局
	root := NewFlexContainer("root", DirectionColumn)

	mockComponent := &MockMeasurableComponent{
		preferredWidth:  50,
		preferredHeight: 20,
	}

	componentInstance := &core.ComponentInstance{
		Type:     "mock",
		Instance: mockComponent,
	}

	childNode := &LayoutNode{
		ID:        "child",
		Type:      LayoutFlex,
		Component: componentInstance,
		Style: &LayoutStyle{
			Width:  NewSize(50),
			Height: NewSize(20),
		},
	}

	root.Children = []*LayoutNode{childNode}
	childNode.Parent = root

	engine := NewEngine(&LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 50},
	})

	result := engine.Layout()

	// 验证布局结果
	assert.Len(t, result.Nodes, 2) // root + child

	child := result.Nodes[1]
	assert.Equal(t, 50, child.Bound.Width)
	assert.Equal(t, 20, child.Bound.Height)
}

func TestLayoutWithFlexibleMeasurableComponent(t *testing.T) {
	// 创建一个包含灵活尺寸组件的布局
	root := NewFlexContainer("root", DirectionRow)
	root.Style.Direction = DirectionRow

	mockComponent1 := &MockMeasurableComponent{
		preferredWidth:  30,
		preferredHeight: 20,
	}

	mockComponent2 := &MockMeasurableComponent{
		preferredWidth:  40,
		preferredHeight: 20,
	}

	componentInstance1 := &core.ComponentInstance{
		Type:     "mock1",
		Instance: mockComponent1,
	}

	componentInstance2 := &core.ComponentInstance{
		Type:     "mock2",
		Instance: mockComponent2,
	}

	childNode1 := &LayoutNode{
		ID:        "child1",
		Type:      LayoutFlex,
		Component: componentInstance1,
		Style: &LayoutStyle{
			Width: NewSize("flex"), // flex: 1
		},
	}

	childNode2 := &LayoutNode{
		ID:        "child2",
		Type:      LayoutFlex,
		Component: componentInstance2,
		Style: &LayoutStyle{
			Width: NewSize(30),
		},
	}

	root.Children = []*LayoutNode{childNode1, childNode2}
	childNode1.Parent = root
	childNode2.Parent = root

	engine := NewEngine(&LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 30},
	})

	result := engine.Layout()

	// 验证布局结果
	assert.Len(t, result.Nodes, 3) // root + 2 children

	// 查找子节点
	var child1, child2 *LayoutNode
	for _, node := range result.Nodes {
		if node.ID == "child1" {
			child1 = node
		} else if node.ID == "child2" {
			child2 = node
		}
	}

	assert.NotNil(t, child1)
	assert.NotNil(t, child2)

	// child2 有固定宽度 30，所以 child1 应该获得剩余空间
	assert.Equal(t, 30, child2.Bound.Width)
	assert.InDelta(t, 70, child1.Bound.Width, 5) // 允许少量误差
}
