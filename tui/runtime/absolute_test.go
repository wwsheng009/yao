package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAbsolutePositioning_Basic tests basic absolute positioning
func TestAbsolutePositioning_Basic(t *testing.T) {
	parent := &LayoutNode{
		ID:             "parent",
		X:              10,
		Y:              10,
		MeasuredWidth:  200,
		MeasuredHeight: 100,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Add absolutely positioned child
	top := 5
	left := 15
	child := &LayoutNode{
		ID:             "child",
		MeasuredWidth:  50,
		MeasuredHeight: 30,
		Position: Position{
			Type:  PositionAbsolute,
			Top:   &top,
			Left:  &left,
		},
	}
	parent.AddChild(child)

	// Apply absolute layout
	ApplyAbsoluteLayout(parent)

	// Verify absolute position
	assert.Equal(t, 10+15, child.AbsoluteX, "AbsoluteX should be parent.X + left")
	assert.Equal(t, 10+5, child.AbsoluteY, "AbsoluteY should be parent.Y + top")
}

// TestAbsolutePositioning_RightBottom tests right and bottom positioning
func TestAbsolutePositioning_RightBottom(t *testing.T) {
	parent := &LayoutNode{
		ID:             "parent",
		X:              0,
		Y:              0,
		MeasuredWidth:  200,
		MeasuredHeight: 100,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Add child with right and bottom offsets
	right := 20
	bottom := 10
	child := &LayoutNode{
		ID:             "child",
		MeasuredWidth:  50,
		MeasuredHeight: 30,
		Position: Position{
			Type:   PositionAbsolute,
			Right:  &right,
			Bottom: &bottom,
		},
	}
	parent.AddChild(child)

	ApplyAbsoluteLayout(parent)

	// AbsoluteX = parent.X + parent.Width - right - child.Width
	// AbsoluteY = parent.Y + parent.Height - bottom - child.Height
	expectedX := 0 + 200 - 20 - 50  // 130
	expectedY := 0 + 100 - 10 - 30  // 60

	assert.Equal(t, expectedX, child.AbsoluteX, "AbsoluteX calculated from right")
	assert.Equal(t, expectedY, child.AbsoluteY, "AbsoluteY calculated from bottom")
}

// TestAbsolutePositioning_Combination tests combination of top/left/right/bottom
func TestAbsolutePositioning_Combination(t *testing.T) {
	parent := &LayoutNode{
		ID:             "parent",
		X:              50,
		Y:              50,
		MeasuredWidth:  300,
		MeasuredHeight: 200,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Test with both left and right (right should override)
	top := 10
	left := 20
	right := 30
	child1 := &LayoutNode{
		ID:             "child1",
		MeasuredWidth:  60,
		MeasuredHeight: 40,
		Position: Position{
			Type:   PositionAbsolute,
			Top:    &top,
			Left:   &left,
			Right:  &right,
		},
	}
	parent.AddChild(child1)

	ApplyAbsoluteLayout(parent)

	// Right overrides left
	expectedX1 := 50 + 300 - 30 - 60  // 260
	assert.Equal(t, expectedX1, child1.AbsoluteX, "Right should override left")
	assert.Equal(t, 50+10, child1.AbsoluteY, "Top should be used")

	// Test with both top and bottom (bottom should override)
	bottomVal := 15
	child2 := &LayoutNode{
		ID:             "child2",
		MeasuredWidth:  80,
		MeasuredHeight: 50,
		Position: Position{
			Type:   PositionAbsolute,
			Top:    &top,
			Bottom: &bottomVal,
		},
	}
	parent.AddChild(child2)

	ApplyAbsoluteLayout(parent)

	// Bottom overrides top
	expectedY2 := 50 + 200 - 15 - 50  // 185
	assert.Equal(t, expectedY2, child2.AbsoluteY, "Bottom should override top")
}

// TestAbsolutePositioning_RelativeChildren tests relative positioning for non-absolute children
func TestAbsolutePositioning_RelativeChildren(t *testing.T) {
	parent := &LayoutNode{
		ID:             "parent",
		X:              0,
		Y:              0,
		MeasuredWidth:  200,
		MeasuredHeight: 100,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Add relative child (default)
	relativeChild := &LayoutNode{
		ID:             "relative-child",
		X:              20,
		Y:              30,
		MeasuredWidth:  50,
		MeasuredHeight: 40,
		Position:       NewPosition(), // Default: relative
	}
	parent.AddChild(relativeChild)

	ApplyAbsoluteLayout(parent)

	// For relative children, AbsoluteX/AbsoluteY should equal X/Y
	assert.Equal(t, 20, relativeChild.AbsoluteX, "Relative child: AbsoluteX = X")
	assert.Equal(t, 30, relativeChild.AbsoluteY, "Relative child: AbsoluteY = Y")
}

// TestAbsolutePositioning_Mixed tests mixed absolute and relative children
func TestAbsolutePositioning_Mixed(t *testing.T) {
	parent := &LayoutNode{
		ID:             "parent",
		X:              10,
		Y:              10,
		MeasuredWidth:  200,
		MeasuredHeight: 100,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Add relative child
	relativeChild := &LayoutNode{
		ID:             "relative",
		X:              0,
		Y:              0,
		MeasuredWidth:  80,
		MeasuredHeight: 60,
		Position:       NewPosition(),
	}
	parent.AddChild(relativeChild)

	// Add absolute child
	top := 5
	left := 10
	absoluteChild := &LayoutNode{
		ID:             "absolute",
		MeasuredWidth:  40,
		MeasuredHeight: 30,
		Position: Position{
			Type:  PositionAbsolute,
			Top:   &top,
			Left:  &left,
		},
	}
	parent.AddChild(absoluteChild)

	ApplyAbsoluteLayout(parent)

	// Relative child should keep its X, Y
	assert.Equal(t, 0, relativeChild.AbsoluteX)
	assert.Equal(t, 0, relativeChild.AbsoluteY)

	// Absolute child should be positioned relative to parent
	assert.Equal(t, 10+10, absoluteChild.AbsoluteX)
	assert.Equal(t, 10+5, absoluteChild.AbsoluteY)
}

// TestAbsolutePositioning_Nested tests nested absolute positioning
func TestAbsolutePositioning_Nested(t *testing.T) {
	grandParent := &LayoutNode{
		ID:             "grandparent",
		X:              5,
		Y:              5,
		MeasuredWidth:  400,
		MeasuredHeight: 300,
		Position:       NewPosition(),
		Children:       make([]*LayoutNode, 0),
	}

	// Parent is absolutely positioned in grandparent
	parentTop := 10
	parentLeft := 20
	parent := &LayoutNode{
		ID:             "parent",
		MeasuredWidth:  200,
		MeasuredHeight: 150,
		Position: Position{
			Type:  PositionAbsolute,
			Top:   &parentTop,
			Left:  &parentLeft,
		},
		Children: make([]*LayoutNode, 0),
	}
	grandParent.AddChild(parent)

	// Child is absolutely positioned in parent
	childTop := 15
	childLeft := 25
	child := &LayoutNode{
		ID:             "child",
		MeasuredWidth:  50,
		MeasuredHeight: 40,
		Position: Position{
			Type:  PositionAbsolute,
			Top:   &childTop,
			Left:  &childLeft,
		},
	}
	parent.AddChild(child)

	ApplyAbsoluteLayout(grandParent)

	// Parent absolute position
	assert.Equal(t, 5+20, parent.AbsoluteX)
	assert.Equal(t, 5+10, parent.AbsoluteY)

	// Apply absolute layout to parent to position child
	ApplyAbsoluteLayout(parent)

	// Child absolute position (relative to parent's X/Y, not parent's AbsoluteX/AbsoluteY)
	// Since parent is absolutely positioned, parent.X = 0 and parent.Y = 0
	// So child is positioned at (0+25, 0+15) = (25, 15)
	assert.Equal(t, 0+25, child.AbsoluteX)
	assert.Equal(t, 0+15, child.AbsoluteY)
}

// TestIsPositionAbsolute tests the IsPositionAbsolute helper
func TestIsPositionAbsolute(t *testing.T) {
	tests := []struct {
		name     string
		node     *LayoutNode
		expected bool
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: false,
		},
		{
			name: "relative position",
			node: &LayoutNode{
				Position: NewPosition(),
			},
			expected: false,
		},
		{
			name: "absolute position",
			node: &LayoutNode{
				Position: Position{
					Type: PositionAbsolute,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.IsPositionAbsolute()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetAbsolutePosition tests the GetAbsolutePosition helper
func TestGetAbsolutePosition(t *testing.T) {
	top := 10
	left := 20
	node := &LayoutNode{
		X:              5,
		Y:              15,
		MeasuredWidth:  50,
		MeasuredHeight: 30,
		Position: Position{
			Type:  PositionAbsolute,
			Top:   &top,
			Left:  &left,
		},
	}

	// Set absolute position manually (simulating layout)
	node.AbsoluteX = 100
	node.AbsoluteY = 200

	x, y := node.GetAbsolutePosition()
	assert.Equal(t, 100, x)
	assert.Equal(t, 200, y)
}

// TestAbsolutePositioning_EdgeCases tests edge cases
func TestAbsolutePositioning_EdgeCases(t *testing.T) {
	t.Run("nil offsets", func(t *testing.T) {
		parent := &LayoutNode{
			ID:             "parent",
			X:              10,
			Y:              10,
			MeasuredWidth:  100,
			MeasuredHeight: 100,
			Position:       NewPosition(),
			Children:       make([]*LayoutNode, 0),
		}

		// Absolute child with nil offsets (should default to 0)
		child := &LayoutNode{
			ID:             "child",
			MeasuredWidth:  50,
			MeasuredHeight: 30,
			Position: Position{
				Type: PositionAbsolute,
				// Top, Left, Right, Bottom all nil
			},
		}
		parent.AddChild(child)

		ApplyAbsoluteLayout(parent)

		// Should position at parent origin
		assert.Equal(t, 10, child.AbsoluteX)
		assert.Equal(t, 10, child.AbsoluteY)
	})

	t.Run("zero offsets", func(t *testing.T) {
		parent := &LayoutNode{
			ID:             "parent",
			X:              10,
			Y:              10,
			MeasuredWidth:  100,
			MeasuredHeight: 100,
			Position:       NewPosition(),
			Children:       make([]*LayoutNode, 0),
		}

		top := 0
		left := 0
		child := &LayoutNode{
			ID:             "child",
			MeasuredWidth:  50,
			MeasuredHeight: 30,
			Position: Position{
				Type:  PositionAbsolute,
				Top:   &top,
				Left:  &left,
			},
		}
		parent.AddChild(child)

		ApplyAbsoluteLayout(parent)

		assert.Equal(t, 10, child.AbsoluteX)
		assert.Equal(t, 10, child.AbsoluteY)
	})

	t.Run("negative offsets", func(t *testing.T) {
		parent := &LayoutNode{
			ID:             "parent",
			X:              50,
			Y:              50,
			MeasuredWidth:  100,
			MeasuredHeight: 100,
			Position:       NewPosition(),
			Children:       make([]*LayoutNode, 0),
		}

		top := -10
		left := -20
		child := &LayoutNode{
			ID:             "child",
			MeasuredWidth:  50,
			MeasuredHeight: 30,
			Position: Position{
				Type:  PositionAbsolute,
				Top:   &top,
				Left:  &left,
			},
		}
		parent.AddChild(child)

		ApplyAbsoluteLayout(parent)

		// Negative offsets should work
		assert.Equal(t, 50-20, child.AbsoluteX)
		assert.Equal(t, 50-10, child.AbsoluteY)
	})
}
