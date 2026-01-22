package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/core"
)

// MockShrinkableComponent 模拟一个具有 shrink 属性的组件
type MockShrinkableComponent struct {
	preferredWidth  int
	preferredHeight int
	shrinkValue     float64
}

func (m *MockShrinkableComponent) View() string {
	return "mock"
}

func (m *MockShrinkableComponent) Init() core.Cmd {
	return nil
}

func (m *MockShrinkableComponent) UpdateMsg(msg core.Msg) (core.ComponentInterface, core.Cmd, core.Response) {
	return m, nil, core.Ignored
}

func (m *MockShrinkableComponent) GetID() string {
	return "mock-shrinkable"
}

func (m *MockShrinkableComponent) SetFocus(focus bool) {}

func (m *MockShrinkableComponent) GetFocus() bool {
	return false
}

func (m *MockShrinkableComponent) GetComponentType() string {
	return "mock-shrinkable"
}

func (m *MockShrinkableComponent) Render(config core.RenderConfig) (string, error) {
	return "mock", nil
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
	return m.preferredWidth, m.preferredHeight
}

func TestLayoutWithShrink(t *testing.T) {
	// 创建一个包含 shrink 组件的布局
	root := NewFlexContainer("root", DirectionRow)
	root.Style.Direction = DirectionRow
	
	// 创建两个组件，其中一个有 shrink 属性
	mockComponent1 := &MockShrinkableComponent{
		preferredWidth:  60,
		preferredHeight: 20,
		shrinkValue:     1.0,
	}
	
	mockComponent2 := &MockShrinkableComponent{
		preferredWidth:  60,
		preferredHeight: 20,
		shrinkValue:     1.0,
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
			Width:  &Size{Value: 60},
			Shrink: Grow{Value: 1.0}, // 设置 shrink 属性
		},
	}
	
	childNode2 := &LayoutNode{
		ID:        "child2",
		Type:      LayoutFlex,
		Component: componentInstance2,
		Style: &LayoutStyle{
			Width:  &Size{Value: 60},
			Shrink: Grow{Value: 1.0}, // 设置 shrink 属性
		},
	}
	
	root.Children = []*LayoutNode{childNode1, childNode2}
	childNode1.Parent = root
	childNode2.Parent = root

	engine := NewEngine(&LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 30}, // 总宽度100，但组件总宽度120，需要收缩
	})

	result := engine.Layout()

	// 验证布局结果
	assert.Len(t, result.Nodes, 3) // root + 2 children
	
	// 查找子节点
	var foundChild1, foundChild2 *LayoutNode
	for _, node := range result.Nodes {
		if node.ID == "child1" {
			foundChild1 = node
		} else if node.ID == "child2" {
			foundChild2 = node
		}
	}
	
	assert.NotNil(t, foundChild1)
	assert.NotNil(t, foundChild2)
	
	// 由于总空间是100，两个组件初始宽度各为60（共120），所以每个组件应该收缩10
	// 因为两个组件的shrink值相同，所以平均分配收缩量
	assert.InDelta(t, 40, foundChild1.Bound.Width, 2) // 大约40
	assert.InDelta(t, 40, foundChild2.Bound.Width, 2) // 大约40
	assert.Less(t, foundChild1.Bound.Width, 60)        // 确认确实收缩了
	assert.Less(t, foundChild2.Bound.Width, 60)        // 确认确实收缩了
}

func TestLayoutWithDifferentShrinkValues(t *testing.T) {
	// 测试不同 shrink 值的组件
	root := NewFlexContainer("root", DirectionRow)
	root.Style.Direction = DirectionRow
	
	// 创建两个组件，有不同的 shrink 值
	mockComponent1 := &MockShrinkableComponent{
		preferredWidth:  50,
		preferredHeight: 20,
		shrinkValue:     2.0, // 更大的 shrink 值，意味着更容易收缩
	}
	
	mockComponent2 := &MockShrinkableComponent{
		preferredWidth:  50,
		preferredHeight: 20,
		shrinkValue:     1.0, // 较小的 shrink 值，意味着更不容易收缩
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
			Width:  &Size{Value: 50},
			Shrink: Grow{Value: 2.0}, // 高 shrink 值
		},
	}
	
	childNode2 := &LayoutNode{
		ID:        "child2",
		Type:      LayoutFlex,
		Component: componentInstance2,
		Style: &LayoutStyle{
			Width:  &Size{Value: 50},
			Shrink: Grow{Value: 1.0}, // 低 shrink 值
		},
	}
	
	root.Children = []*LayoutNode{childNode1, childNode2}
	childNode1.Parent = root
	childNode2.Parent = root

	engine := NewEngine(&LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 30}, // 总宽度80，但组件总宽度100，需要收缩20
	})

	result := engine.Layout()

	// 验证布局结果
	assert.Len(t, result.Nodes, 3) // root + 2 children
	
	// 查找子节点
	var foundChild1, foundChild2 *LayoutNode
	for _, node := range result.Nodes {
		if node.ID == "child1" {
			foundChild1 = node
		} else if node.ID == "child2" {
			foundChild2 = node
		}
	}
	
	assert.NotNil(t, foundChild1)
	assert.NotNil(t, foundChild2)
	
	// 由于 child1 的 shrink 值是 child2 的两倍，所以 child1 应该收缩得更多
	// 总收缩量是20，按照 2:1 的比例分配，child1 收缩约 13.3，child2 收缩约 6.7
	// 所以 child1 最终宽度约为 36.7，child2 最终宽度约为 43.3
	assert.InDelta(t, 37, foundChild1.Bound.Width, 5) // 大约37
	assert.InDelta(t, 43, foundChild2.Bound.Width, 5) // 大约43
	assert.Less(t, foundChild1.Bound.Width, foundChild2.Bound.Width) // 验证高 shrink 值的组件收缩得更多
}