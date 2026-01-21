package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSize(t *testing.T) {
	size := NewSize(100)
	assert.NotNil(t, size)
	assert.Equal(t, 100, size.Value)
	assert.Equal(t, 0, size.Min)
	assert.Equal(t, 0, size.Max)
	assert.Equal(t, "px", size.Unit)
}

func TestNewPadding(t *testing.T) {
	padding := NewPadding(1, 2, 3, 4)
	assert.NotNil(t, padding)
	assert.Equal(t, 1, padding.Top)
	assert.Equal(t, 2, padding.Right)
	assert.Equal(t, 3, padding.Bottom)
	assert.Equal(t, 4, padding.Left)
}

func TestNewMargin(t *testing.T) {
	margin := NewMargin(5, 6, 7, 8)
	assert.NotNil(t, margin)
	assert.Equal(t, 5, margin.Top)
	assert.Equal(t, 6, margin.Right)
	assert.Equal(t, 7, margin.Bottom)
	assert.Equal(t, 8, margin.Left)
}

func TestFlexContainer(t *testing.T) {
	container := NewFlexContainer("test-flex", DirectionRow)
	assert.NotNil(t, container)
	assert.Equal(t, "test-flex", container.ID)
	assert.Equal(t, LayoutFlex, container.Type)
	assert.Equal(t, DirectionRow, container.Style.Direction)
	assert.False(t, container.Dirty)
}

func TestGridContainer(t *testing.T) {
	container := NewGridContainer("test-grid")
	assert.NotNil(t, container)
	assert.Equal(t, "test-grid", container.ID)
	assert.Equal(t, LayoutGrid, container.Type)
	assert.False(t, container.Dirty)
}

func TestAbsoluteContainer(t *testing.T) {
	container := NewAbsoluteContainer("test-absolute")
	assert.NotNil(t, container)
	assert.Equal(t, "test-absolute", container.ID)
	assert.Equal(t, LayoutAbsolute, container.Type)
	assert.False(t, container.Dirty)
}

func TestApplyStyle(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)

	ApplyStyle(node,
		WithFlexDirection(DirectionRow),
		WithAlignItems(AlignCenter),
		WithJustify(JustifyCenter),
		WithGap(10),
	)

	assert.Equal(t, DirectionRow, node.Style.Direction)
	assert.Equal(t, AlignCenter, node.Style.AlignItems)
	assert.Equal(t, JustifyCenter, node.Style.Justify)
	assert.Equal(t, 10, node.Style.Gap)
}

func TestApplyStyleWithPadding(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)

	ApplyStyle(node,
		WithPadding(5, 10, 5, 10),
	)

	assert.NotNil(t, node.Style.Padding)
	assert.Equal(t, 5, node.Style.Padding.Top)
	assert.Equal(t, 10, node.Style.Padding.Right)
	assert.Equal(t, 5, node.Style.Padding.Bottom)
	assert.Equal(t, 10, node.Style.Padding.Left)
}

func TestApplyStyleWithMargin(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)

	ApplyStyle(node,
		WithMargin(2, 4, 2, 4),
	)

	assert.NotNil(t, node.Style.Margin)
	assert.Equal(t, 2, node.Style.Margin.Top)
	assert.Equal(t, 4, node.Style.Margin.Right)
	assert.Equal(t, 2, node.Style.Margin.Bottom)
	assert.Equal(t, 4, node.Style.Margin.Left)
}

func TestApplyStyleWithWidthHeight(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)

	ApplyStyle(node,
		WithWidth(100),
		WithHeight(50),
	)

	assert.NotNil(t, node.Style.Width)
	assert.Equal(t, 100, node.Style.Width.Value)
	assert.NotNil(t, node.Style.Height)
	assert.Equal(t, 50, node.Style.Height.Value)
}

func TestApplyStyleWithMinMax(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)

	ApplyStyle(node,
		WithMinWidth(50),
		WithMinHeight(30),
		WithMaxWidth(200),
		WithMaxHeight(150),
	)

	assert.Equal(t, 50, node.Style.MinWidth)
	assert.Equal(t, 30, node.Style.MinHeight)
	assert.Equal(t, 200, node.Style.MaxWidth)
	assert.Equal(t, 150, node.Style.MaxHeight)
}

func TestApplyStyleWithPositioning(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)

	ApplyStyle(node,
		WithPosition(PositionAbsolute),
		WithPositioning(10, 20, 30, 40),
	)

	assert.Equal(t, PositionAbsolute, node.Style.Position)
	assert.Equal(t, 10, node.Style.Left)
	assert.Equal(t, 20, node.Style.Top)
	assert.Equal(t, 30, node.Style.Right)
	assert.Equal(t, 40, node.Style.Bottom)
}

func TestBuilder(t *testing.T) {
	builder := NewBuilder()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.Root())
	assert.Equal(t, builder.Root(), builder.Current())

	builder.PushContainer(&ContainerConfig{
		ID:   "container1",
		Type: LayoutFlex,
	})

	assert.NotEqual(t, builder.Root(), builder.Current())
	assert.Equal(t, "container1", builder.Current().ID)

	child := NewFlexContainer("child1", DirectionRow)
	builder.AddNode(child)

	assert.Equal(t, 1, len(builder.Current().Children))

	builder.Pop()

	assert.Equal(t, builder.Root(), builder.Current())
}

func TestFindNodeByID(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	child1 := NewFlexContainer("child1", DirectionRow)
	child2 := NewFlexContainer("child2", DirectionRow)

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	grandchild := NewFlexContainer("grandchild", DirectionColumn)
	child1.Children = append(child1.Children, grandchild)
	grandchild.Parent = child1

	found := FindNodeByID(root, "grandchild")
	assert.NotNil(t, found)
	assert.Equal(t, "grandchild", found.ID)

	notFound := FindNodeByID(root, "nonexistent")
	assert.Nil(t, notFound)
}

func TestGetNodePath(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	child := NewFlexContainer("child", DirectionRow)
	grandchild := NewFlexContainer("grandchild", DirectionColumn)

	root.Children = append(root.Children, child)
	child.Parent = root
	child.Children = append(child.Children, grandchild)
	grandchild.Parent = child

	path := GetNodePath(root, "grandchild")
	assert.NotNil(t, path)
	assert.Equal(t, 3, len(path))
	assert.Equal(t, "root", path[0].ID)
	assert.Equal(t, "child", path[1].ID)
	assert.Equal(t, "grandchild", path[2].ID)

	nonExistentPath := GetNodePath(root, "nonexistent")
	assert.Equal(t, 0, len(nonExistentPath))
}

func TestValidateLayoutTree(t *testing.T) {
	root := NewFlexContainer("root", DirectionColumn)
	child1 := NewFlexContainer("child1", DirectionRow)
	child2 := NewFlexContainer("child2", DirectionRow)

	root.Children = append(root.Children, child1, child2)
	child1.Parent = root
	child2.Parent = root

	err := ValidateLayoutTree(root, nil)
	assert.NoError(t, err)

	child1.Parent = child2
	err = ValidateLayoutTree(root, nil)
	assert.Error(t, err)
}

func TestMetrics(t *testing.T) {
	node := NewFlexContainer("test", DirectionColumn)
	ApplyStyle(node,
		WithPadding(2, 3, 2, 3),
	)

	metrics := &LayoutMetrics{
		ContentWidth:  80,
		ContentHeight: 20,
		PaddingWidth:  6,
		PaddingHeight: 4,
		TotalWidth:    86,
		TotalHeight:   24,
	}

	assert.Equal(t, 80, metrics.ContentWidth)
	assert.Equal(t, 20, metrics.ContentHeight)
	assert.Equal(t, 6, metrics.PaddingWidth)
	assert.Equal(t, 4, metrics.PaddingHeight)
	assert.Equal(t, 86, metrics.TotalWidth)
	assert.Equal(t, 24, metrics.TotalHeight)
}
