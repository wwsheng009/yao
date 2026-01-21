package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEngine(t *testing.T) {
	config := &LayoutConfig{
		Root: NewFlexContainer("root", DirectionColumn),
	}

	engine := NewEngine(config)

	assert.NotNil(t, engine)
	assert.Equal(t, config, engine.config)
	assert.Equal(t, 80, engine.window.Width)
	assert.Equal(t, 24, engine.window.Height)
}

func TestNewEngineWithWindowSize(t *testing.T) {
	config := &LayoutConfig{
		Root: NewFlexContainer("root", DirectionColumn),
		WindowSize: &WindowSize{
			Width:  100,
			Height: 50,
		},
	}

	engine := NewEngine(config)

	assert.Equal(t, 100, engine.window.Width)
	assert.Equal(t, 50, engine.window.Height)
}

func TestUpdateWindowSize(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	config := &LayoutConfig{
		Root: root,
	}

	engine := NewEngine(config)
	engine.UpdateWindowSize(100, 50)

	assert.Equal(t, 100, engine.window.Width)
	assert.Equal(t, 50, engine.window.Height)
	assert.True(t, root.Dirty)
}

func TestMarkDirty(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	child1 := NewFlexContainer("child1", DirectionRow)
	child2 := NewFlexContainer("child2", DirectionRow)

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{Root: root}
	_ = NewEngine(config)

	assert.False(t, child2.Dirty)
	assert.False(t, root.Dirty)
	assert.False(t, child1.Dirty)

	child2.Dirty = true
	assert.True(t, child2.Dirty)
	assert.False(t, root.Dirty)
	assert.False(t, child1.Dirty)
}

func TestLayoutEmpty(t *testing.T) {
	config := &LayoutConfig{Root: nil}
	engine := NewEngine(config)

	result := engine.Layout()

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Nodes))
	assert.Equal(t, 0, len(result.Dirties))
}

func TestLayoutSingleNode(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Dirty = false
	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Nodes))
	assert.Equal(t, "root", result.Nodes[0].ID)
	assert.Equal(t, 0, result.Nodes[0].Bound.X)
	assert.Equal(t, 0, result.Nodes[0].Bound.Y)
	assert.Equal(t, 80, result.Nodes[0].Bound.Width)
	assert.Equal(t, 24, result.Nodes[0].Bound.Height)
}

func TestLayoutFlexRowWithChildren(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionColumn

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(30))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(50))

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 3)

	rootNode := FindNodeByID(result.Nodes[0], "root")
	assert.NotNil(t, rootNode)
}

func TestLayoutFlexColumnWithChildren(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionColumn

	child1 := NewFlexContainer("child1", DirectionColumn)
	ApplyStyle(child1, WithHeight(10))

	child2 := NewFlexContainer("child2", DirectionColumn)
	ApplyStyle(child2, WithHeight(14))

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 3)
}

func TestLayoutGrid(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Type = LayoutGrid

	for i := 0; i < 4; i++ {
		child := NewFlexContainer("child", DirectionRow)
		child.ID = "child" + string(rune('0'+i))
		root.Children = append(root.Children, child)
		child.Parent = root
	}

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 5)
}

func TestLayoutAbsolute(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Type = LayoutAbsolute

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithPosition(PositionAbsolute), WithPositioning(10, 10, 0, 0))
	ApplyStyle(child1, WithWidth(20), WithHeight(10))

	root.Children = append(root.Children, child1)
	child1.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 2)
}

func TestLayoutWithPadding(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	ApplyStyle(root, WithPadding(5, 10, 5, 10))

	child := NewFlexContainer("child", DirectionRow)
	ApplyStyle(child, WithWidth(40))
	ApplyStyle(child, WithHeight(10))

	root.Children = append(root.Children, child)
	child.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 2)
}

func TestLayoutFlexGrowEqual(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionRow

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(40))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(40))

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 3)
}

func TestLayoutFlexAlignItemsCenter(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionRow
	root.Style.AlignItems = AlignCenter

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(30), WithHeight(10))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(50), WithHeight(20))

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 30},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 3)
}

func TestLayoutFlexJustifyCenter(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionRow
	root.Style.Justify = JustifyCenter

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(20))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(30))

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 3)
}

func TestLayoutNested(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)

	parent1 := NewFlexContainer("parent1", DirectionRow)
	ApplyStyle(parent1, WithHeight(10))

	parent2 := NewFlexContainer("parent2", DirectionRow)
	ApplyStyle(parent2, WithHeight(14))

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(20))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(30))

	parent1.Children = append(parent1.Children, child1, child2)
	child1.Parent = parent1
	child2.Parent = parent1

	root.Children = append(root.Children, parent1, parent2)
	parent1.Parent = root
	parent2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 5)
}

func TestLayoutGap(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionRow
	root.Style.Gap = 5

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(30))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(40))

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 3)
}

func TestDirtyFlags(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)

	child1 := NewFlexContainer("child1", DirectionRow)
	ApplyStyle(child1, WithWidth(30))

	child2 := NewFlexContainer("child2", DirectionRow)
	ApplyStyle(child2, WithWidth(40))
	child2.Dirty = true

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
	assert.Greater(t, len(result.Dirties), 0)
}

func TestCalculateMetrics(t *testing.T) {
	root := NewFlexContainer("test", DirectionColumn)
	ApplyStyle(root, WithPadding(5, 10, 5, 10))

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 80, Height: 40},
	}

	engine := NewEngine(config)
	engine.calculateMetrics(root, 80, 40)

	assert.NotNil(t, root.Metrics)
	assert.Equal(t, 60, root.Metrics.ContentWidth)
	assert.Equal(t, 30, root.Metrics.ContentHeight)
	assert.Equal(t, 20, root.Metrics.PaddingWidth)
	assert.Equal(t, 10, root.Metrics.PaddingHeight)
	assert.Equal(t, 80, root.Metrics.TotalWidth)
	assert.Equal(t, 40, root.Metrics.TotalHeight)
}

func TestMinSizeConstraint(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	root.Style.Direction = DirectionRow

	child := NewFlexContainer("child", DirectionRow)
	ApplyStyle(child, WithMinWidth(50), WithWidth(10))

	root.Children = append(root.Children, child)
	child.Parent = root

	config := &LayoutConfig{
		Root:       root,
		WindowSize: &WindowSize{Width: 100, Height: 24},
	}

	engine := NewEngine(config)
	result := engine.Layout()

	assert.NotNil(t, result)
}
