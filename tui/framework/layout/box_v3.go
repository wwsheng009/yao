package layout

import (
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Box Container V3
// ==============================================================================
// V3 盒子容器，使用 CellBuffer 绘制

// BoxV3 V3 盒子容器
type BoxV3 struct {
	*component.BaseComponentV3
	*component.StateHolder

	border     *BorderStyleV3
	padding    BoxPaddingV3
	margin     BoxPaddingV3
	background style.Color
	children   []component.Node
}

// BorderStyleV3 边框样式 V3（避免与 box.go 冲突）
type BorderStyleV3 struct {
	Enabled bool
	Top     bool
	Bottom  bool
	Left    bool
	Right   bool
	Type    string
	FgColor style.Color
}

// BoxPaddingV3 内边距 V3
type BoxPaddingV3 struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// NewBoxV3 创建 V3 盒子容器
func NewBoxV3() *BoxV3 {
	return &BoxV3{
		BaseComponentV3: component.NewBaseComponentV3("box"),
		StateHolder:     component.NewStateHolder(),
		border:         nil,
		padding:        BoxPaddingV3{Top: 0, Right: 0, Bottom: 0, Left: 0},
		margin:         BoxPaddingV3{Top: 0, Right: 0, Bottom: 0, Left: 0},
		background:     "",
		children:       make([]component.Node, 0),
	}
}

// ============================================================================
// 子节点管理
// ============================================================================

// Children 返回子节点
func (b *BoxV3) Children() []component.Node {
	return b.children
}

// AddNode 添加子节点
func (b *BoxV3) AddNode(child component.Node) {
	b.children = append(b.children, child)
}

// ClearChildren 清空所有子节点
func (b *BoxV3) ClearChildren() {
	b.children = make([]component.Node, 0)
}

// ChildCount 子节点数量
func (b *BoxV3) ChildCount() int {
	return len(b.children)
}

// ============================================================================
// 链式设置方法
// ============================================================================

// WithBorder 设置边框
func (b *BoxV3) WithBorder(enabled bool) *BoxV3 {
	if enabled && b.border == nil {
		b.border = &BorderStyleV3{
			Enabled: true,
			Top:     true,
			Bottom:  true,
			Left:    true,
			Right:   true,
			Type:    "normal",
			FgColor: style.BrightBlack,
		}
	} else if !enabled {
		b.border = nil
	}
	return b
}

// WithBorderType 设置边框类型 ("normal", "rounded", "double", "thick")
func (b *BoxV3) WithBorderType(borderType string) *BoxV3 {
	if b.border == nil {
		b.WithBorder(true)
	}
	b.border.Type = borderType
	return b
}

// WithBorderColor 设置边框颜色
func (b *BoxV3) WithBorderColor(color style.Color) *BoxV3 {
	if b.border == nil {
		b.WithBorder(true)
	}
	b.border.FgColor = color
	return b
}

// WithPadding 设置内边距（全部）
func (b *BoxV3) WithPadding(all int) *BoxV3 {
	b.padding = BoxPaddingV3{
		Top:    all,
		Right:  all,
		Bottom: all,
		Left:   all,
	}
	return b
}

// WithPaddingV 设置垂直内边距
func (b *BoxV3) WithPaddingV(vertical int) *BoxV3 {
	b.padding.Top = vertical
	b.padding.Bottom = vertical
	return b
}

// WithPaddingH 设置水平内边距
func (b *BoxV3) WithPaddingH(horizontal int) *BoxV3 {
	b.padding.Left = horizontal
	b.padding.Right = horizontal
	return b
}

// WithBackground 设置背景色
func (b *BoxV3) WithBackground(color style.Color) *BoxV3 {
	b.background = color
	return b
}

// WithChild 添加子组件
func (b *BoxV3) WithChild(child component.Node) *BoxV3 {
	b.AddNode(child)
	return b
}

// WithChildren 添加多个子组件
func (b *BoxV3) WithChildren(children ...component.Node) *BoxV3 {
	for _, child := range children {
		b.AddNode(child)
	}
	return b
}

// GetBorder 获取边框样式
func (b *BoxV3) GetBorder() *BorderStyleV3 {
	return b.border
}

// GetPadding 获取内边距
func (b *BoxV3) GetPadding() BoxPaddingV3 {
	return b.padding
}

// GetMargin 获取外边距
func (b *BoxV3) GetMargin() BoxPaddingV3 {
	return b.margin
}

// GetBackground 获取背景色
func (b *BoxV3) GetBackground() style.Color {
	return b.background
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (b *BoxV3) Measure(maxWidth, maxHeight int) (width, height int) {
	// 基础尺寸：边框 + 内边距 + 内容
	width = b.padding.Left + b.padding.Right
	height = b.padding.Top + b.padding.Bottom

	// 边框占用
	if b.border != nil && b.border.Enabled {
		if b.border.Left {
			width++
		}
		if b.border.Right {
			width++
		}
		if b.border.Top {
			height++
		}
		if b.border.Bottom {
			height++
		}
	}

	// 加上子组件尺寸
	for _, child := range b.children {
		if measurable, ok := child.(interface{ Measure(int, int) (int, int) }); ok {
			cw, ch := measurable.Measure(maxWidth-width, maxHeight-height)
			width += cw
			height += ch
		}
	}

	// 确保不超过最大值
	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}

	return width, height
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件到 CellBuffer
func (b *BoxV3) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !b.IsVisible() {
		return
	}

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	// 计算内容区域
	contentX := ctx.X + b.padding.Left
	contentY := ctx.Y + b.padding.Top
	contentW := width - b.padding.Left - b.padding.Right
	contentH := height - b.padding.Top - b.padding.Bottom

	// 绘制背景
	if b.background != "" {
		bgStyle := style.Style{}.Background(style.Color(b.background))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				buf.SetCell(ctx.X+x, ctx.Y+y, ' ', bgStyle)
			}
		}
	}

	// 绘制边框
	if b.border != nil && b.border.Enabled {
		b.paintBorder(ctx, buf, width, height)
		// 边框占用内容空间
		if b.border.Left {
			contentX++
			contentW--
		}
		if b.border.Right {
			contentW--
		}
		if b.border.Top {
			contentY++
			contentH--
		}
		if b.border.Bottom {
			contentH--
		}
	}

	// 绘制子组件
	if contentW > 0 && contentH > 0 {
		for _, child := range b.children {
			if paintable, ok := child.(interface {
				Paint(component.PaintContext, *paint.Buffer)
			}); ok {
				childCtx := component.PaintContext{
					AvailableWidth:  contentW,
					AvailableHeight: contentH,
					X:                contentX,
					Y:                contentY,
				}
				paintable.Paint(childCtx, buf)
			}
		}
	}
}

// paintBorder 绘制边框
func (b *BoxV3) paintBorder(ctx component.PaintContext, buf *paint.Buffer, width, height int) {
	if b.border == nil || !b.border.Enabled {
		return
	}

	edges := b.getBorderEdges()
	borderStyle := style.Style{}.Foreground(b.border.FgColor)

	// 顶部边框
	if b.border.Top && height > 0 {
		if width >= 2 {
			buf.SetCell(ctx.X, ctx.Y, []rune(edges.TL)[0], borderStyle)
			buf.SetCell(ctx.X+width-1, ctx.Y, []rune(edges.TR)[0], borderStyle)
		}
		for x := 1; x < width-1; x++ {
			buf.SetCell(ctx.X+x, ctx.Y, []rune(edges.T)[0], borderStyle)
		}
	}

	// 中间行
	for y := 1; y < height-1; y++ {
		// 左边框
		if b.border.Left {
			buf.SetCell(ctx.X, ctx.Y+y, []rune(edges.L)[0], borderStyle)
		}
		// 右边框
		if b.border.Right {
			buf.SetCell(ctx.X+width-1, ctx.Y+y, []rune(edges.R)[0], borderStyle)
		}
	}

	// 底部边框
	if b.border.Bottom && height > 1 {
		if width >= 2 {
			buf.SetCell(ctx.X, ctx.Y+height-1, []rune(edges.BL)[0], borderStyle)
			buf.SetCell(ctx.X+width-1, ctx.Y+height-1, []rune(edges.BR)[0], borderStyle)
		}
		for x := 1; x < width-1; x++ {
			buf.SetCell(ctx.X+x, ctx.Y+height-1, []rune(edges.B)[0], borderStyle)
		}
	}
}

// getBorderEdges 获取边框字符
func (b *BoxV3) getBorderEdges() BorderEdgesV3 {
	switch b.border.Type {
	case "rounded":
		return BorderEdgesV3{
			TL: "╭", T: "─", TR: "╮",
			L:  "│", R:  "│",
			BL: "╰", B: "─", BR: "╯",
		}
	case "double":
		return BorderEdgesV3{
			TL: "╔", T: "═", TR: "╗",
			L:  "║", R:  "║",
			BL: "╚", B: "═", BR: "╝",
		}
	case "thick":
		return BorderEdgesV3{
			TL: "┏", T: "━", TR: "┓",
			L:  "┃", R:  "┃",
			BL: "┗", B: "━", BR: "┛",
		}
	default: // "normal"
		return BorderEdgesV3{
			TL: "┌", T: "─", TR: "┐",
			L:  "│", R:  "│",
			BL: "└", B: "─", BR: "┘",
		}
	}
}

// BorderEdgesV3 边框边缘字符 V3
type BorderEdgesV3 struct {
	TL string // Top Left
	T  string // Top
	TR string // Top Right
	L  string // Left
	R  string // Right
	BL string // Bottom Left
	B  string // Bottom
	BR string // Bottom Right
}
