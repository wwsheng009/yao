package layout

import (
	"github.com/yaoapp/yao/tui/core"
)

type Builder struct {
	root  *LayoutNode
	stack []*LayoutNode
}

func NewBuilder() *Builder {
	root := &LayoutNode{
		Type: LayoutFlex,
		Style: &LayoutStyle{
			Direction:  DirectionColumn,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		},
	}
	return &Builder{
		root:  root,
		stack: []*LayoutNode{root},
	}
}

func (b *Builder) Root() *LayoutNode {
	return b.root
}

func (b *Builder) Current() *LayoutNode {
	if len(b.stack) == 0 {
		return nil
	}
	return b.stack[len(b.stack)-1]
}

func (b *Builder) PushContainer(config *ContainerConfig) *Builder {
	node := &LayoutNode{
		Type:   config.Type,
		ID:     config.ID,
		Style:  config.Style,
		Props:  config.Props,
		Parent: b.Current(),
		Dirty:  true,
	}

	parent := b.Current()
	if parent != nil {
		parent.Children = append(parent.Children, node)
	}

	b.stack = append(b.stack, node)
	return b
}

func (b *Builder) Pop() *Builder {
	if len(b.stack) > 1 {
		b.stack = b.stack[:len(b.stack)-1]
	}
	return b
}

func (b *Builder) AddComponent(component *core.ComponentInstance, config *ComponentConfig) *Builder {
	node := &LayoutNode{
		ID:        config.ID,
		Component: component,
		Style:     config.Style,
		Props:     config.Props,
		Parent:    b.Current(),
		Dirty:     true,
	}

	parent := b.Current()
	if parent != nil {
		parent.Children = append(parent.Children, node)
	}

	return b
}

func (b *Builder) AddNode(node *LayoutNode) *Builder {
	node.Parent = b.Current()
	parent := b.Current()
	if parent != nil {
		parent.Children = append(parent.Children, node)
	}
	return b
}

type ContainerConfig struct {
	ID    string
	Type  LayoutType
	Style *LayoutStyle
	Props map[string]interface{}
}

type ComponentConfig struct {
	ID    string
	Style *LayoutStyle
	Props map[string]interface{}
}

func NewFlexContainer(id string, direction Direction) *LayoutNode {
	node := &LayoutNode{
		ID:   id,
		Type: LayoutFlex,
		Style: &LayoutStyle{
			Direction:  direction,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		},
	}
	return node
}

func NewGridContainer(id string) *LayoutNode {
	node := &LayoutNode{
		ID:   id,
		Type: LayoutGrid,
		Style: &LayoutStyle{
			Direction:  DirectionColumn,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		},
	}
	return node
}

func NewAbsoluteContainer(id string) *LayoutNode {
	node := &LayoutNode{
		ID:   id,
		Type: LayoutAbsolute,
		Style: &LayoutStyle{
			Direction:  DirectionColumn,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		},
	}
	return node
}

func NewComponentNode(id string, component *core.ComponentInstance) *LayoutNode {
	node := &LayoutNode{
		ID:        id,
		Component: component,
		Style: &LayoutStyle{
			Direction:  DirectionColumn,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		},
	}
	return node
}

func WithFlexDirection(direction Direction) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Direction = direction
	}
}

func WithAlignItems(align Align) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.AlignItems = align
	}
}

func WithJustify(justify Justify) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Justify = justify
	}
}

func WithGap(gap int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Gap = gap
	}
}

func WithPadding(top, right, bottom, left int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Padding = NewPadding(top, right, bottom, left)
	}
}

func WithMargin(top, right, bottom, left int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Margin = NewMargin(top, right, bottom, left)
	}
}

func WithWidth(width interface{}) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Width = NewSize(width)
	}
}

func WithHeight(height interface{}) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Height = NewSize(height)
	}
}

func WithMinWidth(minWidth int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.MinWidth = minWidth
	}
}

func WithMinHeight(minHeight int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.MinHeight = minHeight
	}
}

func WithMaxWidth(maxWidth int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.MaxWidth = maxWidth
	}
}

func WithGrow(value float64) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Grow = NewGrow(value)
	}
}

func WithShrink(value float64) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Shrink = NewGrow(value)
	}
}

func WithMaxHeight(maxHeight int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.MaxHeight = maxHeight
	}
}

func WithPosition(position Position) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Position = position
	}
}

func WithPositioning(left, top, right, bottom int) func(*LayoutStyle) {
	return func(style *LayoutStyle) {
		style.Left = left
		style.Top = top
		style.Right = right
		style.Bottom = bottom
	}
}

func ApplyStyle(node *LayoutNode, modifiers ...func(*LayoutStyle)) {
	if node == nil {
		return
	}
	if node.Style == nil {
		node.Style = &LayoutStyle{
			Direction:  DirectionColumn,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		}
	}
	for _, modifier := range modifiers {
		modifier(node.Style)
	}
}
