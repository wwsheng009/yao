package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaoapp/yao/tui/runtime"
)

func TestBoxConstraints(t *testing.T) {
	// Test creating constraints
	c := runtime.NewBoxConstraints(10, 100, 20, 200)

	assert.Equal(t, 10, c.MinWidth)
	assert.Equal(t, 100, c.MaxWidth)
	assert.Equal(t, 20, c.MinHeight)
	assert.Equal(t, 200, c.MaxHeight)

	// Test IsTight
	tight := runtime.NewBoxConstraints(50, 50, 50, 50)
	assert.True(t, tight.IsTight())

	loose := runtime.NewBoxConstraints(0, 100, 0, 100)
	assert.False(t, loose.IsTight())

	// Test Constrain
	width, height := c.Constrain(150, 250)
	assert.Equal(t, 100, width)  // MaxWidth
	assert.Equal(t, 200, height) // MaxHeight

	width, height = c.Constrain(5, 15)
	assert.Equal(t, 10, width)  // MinWidth
	assert.Equal(t, 20, height) // MinHeight

	// Test Loosen
	looseC := loose.Loosen()
	assert.Equal(t, 0, looseC.MinWidth)
	assert.Equal(t, 0, looseC.MinHeight)
	assert.Equal(t, 100, looseC.MaxWidth)
	assert.Equal(t, 100, looseC.MaxHeight)
}

func TestConstraintsAlias(t *testing.T) {
	// Test that Constraints type alias works
	c := runtime.NewConstraints(80, 24)

	assert.Equal(t, 0, c.MinWidth)
	assert.Equal(t, 80, c.MaxWidth)
	assert.Equal(t, 0, c.MinHeight)
	assert.Equal(t, 24, c.MaxHeight)
}

func TestStyle(t *testing.T) {
	// Test default style
	style := runtime.NewStyle()
	assert.Equal(t, -1, style.Width)
	assert.Equal(t, -1, style.Height)
	assert.Equal(t, float64(0), style.FlexGrow)
	assert.Equal(t, runtime.DirectionRow, style.Direction)
	assert.Equal(t, runtime.AlignStart, style.AlignItems)
	assert.Equal(t, runtime.JustifyStart, style.Justify)
	assert.Equal(t, 0, style.Gap)

	// Test builder pattern
	style = style.WithWidth(100).
		WithHeight(50).
		WithFlexGrow(1.5).
		WithDirection(runtime.DirectionColumn).
		WithAlignItems(runtime.AlignCenter)

	assert.Equal(t, 100, style.Width)
	assert.Equal(t, 50, style.Height)
	assert.Equal(t, float64(1.5), style.FlexGrow)
	assert.Equal(t, runtime.DirectionColumn, style.Direction)
	assert.Equal(t, runtime.AlignCenter, style.AlignItems)
}

func TestInsets(t *testing.T) {
	insets := runtime.NewInsets(1, 2, 3, 4)
	assert.Equal(t, 1, insets.Top)
	assert.Equal(t, 2, insets.Right)
	assert.Equal(t, 3, insets.Bottom)
	assert.Equal(t, 4, insets.Left)
}

func TestCellBuffer(t *testing.T) {
	// Test creating buffer
	buf := runtime.NewCellBuffer(10, 5)

	assert.Equal(t, 10, buf.Width())
	assert.Equal(t, 5, buf.Height())

	// Test default cell
	cell := buf.GetContent(5, 3)
	assert.Equal(t, ' ', cell.Char)
	assert.Equal(t, 0, cell.ZIndex)

	// Test setting content
	style := runtime.CellStyle{Bold: true}
	buf.SetContent(5, 3, 10, 'A', style, "test-node")

	cell = buf.GetContent(5, 3)
	assert.Equal(t, 'A', cell.Char)
	assert.Equal(t, 10, cell.ZIndex)
	assert.Equal(t, "test-node", cell.NodeID)
	assert.True(t, cell.Style.Bold)

	// Test Z-Index (overwrites lower Z-Index)
	buf.SetContent(5, 3, 5, 'B', runtime.CellStyle{}, "low-node")
	cell = buf.GetContent(5, 3)
	assert.Equal(t, 'A', cell.Char) // Higher Z-Index wins

	// Test Clear
	buf.Clear()
	cell = buf.GetContent(5, 3)
	assert.Equal(t, ' ', cell.Char)
	assert.Equal(t, 0, cell.ZIndex)

	// Test String output
	buf.SetContent(0, 0, 0, 'H', runtime.CellStyle{}, "")
	buf.SetContent(1, 0, 0, 'i', runtime.CellStyle{}, "")
	buf.SetContent(2, 0, 0, '!', runtime.CellStyle{}, "")

	buf.SetContent(0, 1, 0, 'B', runtime.CellStyle{}, "")
	buf.SetContent(1, 1, 0, 'y', runtime.CellStyle{}, "")
	buf.SetContent(2, 1, 0, 'e', runtime.CellStyle{}, "")

	str := buf.String()
	assert.Contains(t, str, "Hi!")
	assert.Contains(t, str, "Bye")
}

func TestLayoutNode(t *testing.T) {
	// Test creating node
	style := runtime.NewStyle().WithWidth(100).WithHeight(50)
	node := runtime.NewLayoutNode("test-node", runtime.NodeTypeText, style)

	assert.Equal(t, "test-node", node.ID)
	assert.Equal(t, runtime.NodeTypeText, node.Type)
	assert.Equal(t, 100, node.Style.Width)
	assert.Equal(t, 50, node.Style.Height)
	assert.True(t, node.IsDirty())

	// Test AddChild
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, runtime.NewStyle())
	node.AddChild(child)

	assert.Equal(t, 1, len(node.Children))
	assert.Equal(t, node, child.Parent)

	// Test MarkDirty
	child.ClearDirty()
	assert.False(t, child.IsDirty())
	node.MarkDirty()
	assert.True(t, child.IsDirty())

	// Test ContainsPoint
	node.X = 10
	node.Y = 20
	node.MeasuredWidth = 30
	node.MeasuredHeight = 40

	assert.True(t, node.ContainsPoint(15, 30))
	assert.True(t, node.ContainsPoint(10, 20))
	assert.False(t, node.ContainsPoint(5, 30))
	assert.False(t, node.ContainsPoint(50, 70))
}

func TestSize(t *testing.T) {
	size := runtime.Size{Width: 100, Height: 50}
	assert.Equal(t, 100, size.Width)
	assert.Equal(t, 50, size.Height)
}
