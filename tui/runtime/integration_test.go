package runtime_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/runtime"
	uicomponents "github.com/yaoapp/yao/tui/ui/components"
)

// MockTextComponent is a mock text component for testing
type MockTextComponent struct {
	content string
	id      string
	width   int
	height  int
}

func NewMockTextComponent(id, content string) *MockTextComponent {
	return &MockTextComponent{
		id:      id,
		content: content,
	}
}

func (m *MockTextComponent) GetID() string {
	return m.id
}

func (m *MockTextComponent) SetFocus(focus bool) {}

func (m *MockTextComponent) GetFocus() bool {
	return false
}

func (m *MockTextComponent) GetComponentType() string {
	return "mock-text"
}

func (m *MockTextComponent) View() string {
	return m.content
}

func (m *MockTextComponent) Init() tea.Cmd {
	return nil
}

func (m *MockTextComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return m, nil, core.Ignored
}

func (m *MockTextComponent) Render(config core.RenderConfig) (string, error) {
	return m.content, nil
}

func (m *MockTextComponent) UpdateRenderConfig(config core.RenderConfig) error {
	return nil
}

func (m *MockTextComponent) Cleanup() {
	// Cleanup resources
}

func (m *MockTextComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (m *MockTextComponent) GetSubscribedMessageTypes() []string {
	return nil
}

// Measurable interface implementation
func (m *MockTextComponent) Measure(c runtime.BoxConstraints) runtime.Size {
	if m.content == "" {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Simple mock implementation: calculate lines
	width := c.MaxWidth
	if width < 0 {
		width = len(m.content)
	}

	// Simple character count (v1: simplified wrapping)
	lineWidth := min(width, len(m.content))
	lines := (len(m.content) + lineWidth - 1) / lineWidth

	m.width = lineWidth
	m.height = lines

	return runtime.Size{
		Width:  m.width,
		Height: m.height,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestRuntime_TextRendering is an end-to-end test of the Runtime system.
// It verifies that a simple Text component can be:
//  1. Measured
//  2. Laid out
//  3. Rendered
func TestRuntime_TextRendering(t *testing.T) {
	// Setup
	runtimeImpl := runtime.NewRuntime(80, 24)

	// Create a Text component
	textComp := &uicomponents.TextComponent{}
	textComp.SetContent("Hello, Runtime!")

	// Wrap in a ComponentInstance
	coreInstance := &core.ComponentInstance{
		ID:       "text1",
		Type:     "text",
		Instance: textComp,
	}

	// Create a LayoutNode
	style := runtime.NewStyle()
	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeText, style)
	rootNode.Component = coreInstance

	// Perform Layout phase
	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	// Verify Measure phase worked
	assert.NotZero(t, result.RootWidth, "Root width should be non-zero")
	assert.NotZero(t, result.RootHeight, "Root height should be non-zero")
	assert.NotZero(t, rootNode.MeasuredWidth, "Node measured width should be non-zero")
	assert.NotZero(t, rootNode.MeasuredHeight, "Node measured height should be non-zero")

	// Verify Layout phase worked
	assert.GreaterOrEqual(t, len(result.Boxes), 1, "Should have at least one layout box")
	assert.Equal(t, rootNode.ID, result.Boxes[0].NodeID, "First box should be the root node")

	// Verify Render phase works
	frame := runtimeImpl.Render(result)
	assert.NotNil(t, frame, "Frame should not be nil")
	assert.NotNil(t, frame.Buffer, "Frame buffer should not be nil")
	output := frame.String()
	assert.NotEmpty(t, output, "Rendered output should not be empty")

	t.Logf("Rendered output:\n%s", output)
}

// TestRuntime_BoxConstraints verifies box constraints functionality.
func TestRuntime_BoxConstraints(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(40, 10)

	// Text with constraint
	textComp := &uicomponents.TextComponent{}
	textComp.SetContent("This is a longer text that should wrap")

	coreInstance := &core.ComponentInstance{
		ID:       "text2",
		Type:     "text",
		Instance: textComp,
	}

	style := runtime.NewStyle()
	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeText, style)
	rootNode.Component = coreInstance

	// Measure with width constraint
	constraints := runtime.NewBoxConstraints(10, 20, 0, 100)
	_ = runtimeImpl.Layout(rootNode, constraints) // result is not used but layout needed

	assert.GreaterOrEqual(t, rootNode.MeasuredWidth, constraints.MinWidth)
	assert.LessOrEqual(t, rootNode.MeasuredWidth, constraints.MaxWidth)

	t.Logf("Text size with constraint: %dx%d", rootNode.MeasuredWidth, rootNode.MeasuredHeight)
}

// TestRuntime_FlexLayout tests basic flex layout with multiple children.
func TestRuntime_FlexLayout(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	// Create a flex container
	flexStyle := runtime.NewStyle().
		WithDirection(runtime.DirectionRow).
		WithFlexGrow(1.0)

	rootNode := runtime.NewLayoutNode("root", runtime.NodeTypeFlex, flexStyle)

	// Add three text components
	for i := 0; i < 3; i++ {
		textComp := &uicomponents.TextComponent{}
		textComp.SetContent(string(rune('A' + i)))

		coreInstance := &core.ComponentInstance{
			ID:       "text" + string(rune('1'+i)),
			Type:     "text",
			Instance: textComp,
		}

		childStyle := runtime.NewStyle().WithFlexGrow(1.0)
		child := runtime.NewLayoutNode("child"+string(rune('1'+i)), runtime.NodeTypeText, childStyle)
		child.Component = coreInstance
		rootNode.AddChild(child)
	}

	// Perform layout
	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	// Verify all children were measured and positioned
	assert.Equal(t, 4, len(result.Boxes), "Should have root and 3 children")

	// Verify children are positioned horizontally (Row layout)
	child1Pos := childPosition(result.Boxes, "child1")
	child2Pos := childPosition(result.Boxes, "child2")
	child3Pos := childPosition(result.Boxes, "child3")

	assert.NotNil(t, child1Pos, "child1 should be found")
	assert.NotNil(t, child2Pos, "child2 should be found")
	assert.NotNil(t, child3Pos, "child3 should be found")

	// Children should be positioned in order
	assert.LessOrEqual(t, child1Pos.X, child2Pos.X, "child1 should be left of child2")
	assert.LessOrEqual(t, child2Pos.X, child3Pos.X, "child2 should be left of child3")

	t.Logf("Child positions: child1=(%d,%d, %dx%d), child2=(%d,%d, %dx%d), child3=(%d,%d, %dx%d)",
		child1Pos.X, child1Pos.Y, child1Pos.W, child1Pos.H,
		child2Pos.X, child2Pos.Y, child2Pos.W, child2Pos.H,
		child3Pos.X, child3Pos.Y, child3Pos.W, child3Pos.H)

	// Render
	frame := runtimeImpl.Render(result)
	assert.NotEmpty(t, frame.String(), "Rendered output should not be empty")
}

// childPosition finds a LayoutBox by node ID.
func childPosition(boxes []runtime.LayoutBox, id string) *runtime.LayoutBox {
	for _, box := range boxes {
		if box.NodeID == id {
			return &box
		}
	}
	return nil
}

// TestRuntime_MeasureOnly tests that Measure phase works independently.
func TestRuntime_MeasureOnly(t *testing.T) {
	textComp := &uicomponents.TextComponent{}
	textComp.SetContent("Measure test")

	coreInstance := &core.ComponentInstance{
		ID:       "measureTest",
		Type:     "text",
		Instance: textComp,
	}

	style := runtime.NewStyle()
	node := runtime.NewLayoutNode("measureNode", runtime.NodeTypeText, style)
	node.Component = coreInstance

	// Perform measure only
	constraints := runtime.NewConstraints(30, 10)
	runtime.PerformMeasure(node, constraints)

	// Verify measured size
	assert.Greater(t, node.MeasuredWidth, 0, "Measured width should be > 0")
	assert.Greater(t, node.MeasuredHeight, 0, "Measured height should be > 0")
}

// TestRuntime_EmptyNode tests handling of empty nodes.
func TestRuntime_EmptyNode(t *testing.T) {
	runtimeImpl := runtime.NewRuntime(80, 24)

	// Empty root node
	rootNode := runtime.NewLayoutNode("emptyRoot", runtime.NodeTypeFlex, runtime.NewStyle())

	constraints := runtime.NewConstraints(80, 24)
	result := runtimeImpl.Layout(rootNode, constraints)

	// Empty nodes should still have boxes (just with zero size)
	assert.Equal(t, 1, len(result.Boxes), "Should have root box")
	assert.Equal(t, 0, result.RootWidth, "Empty root should have zero width")
	assert.Equal(t, 0, result.RootHeight, "Empty root should have zero height")
}
