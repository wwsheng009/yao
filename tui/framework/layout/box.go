package layout

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Box 盒子容器
type Box struct {
	*component.BaseContainer

	border    BorderStyle
	borderFg  style.Color
	padding   BoxSpacing
	margin    BoxSpacing
	width     int
	height    int
	bgColor   style.Color
}

// BorderStyle 边框样式
type BorderStyle struct {
	Enabled bool
	Top     bool
	Bottom  bool
	Left    bool
	Right   bool
	Style   BorderType
}

// BorderType 边框类型
type BorderType int

const (
	BorderNormal BorderType = iota
	BorderRounded
	BorderDouble
	BorderThick
	BorderHidden
)

// BoxSpacing 间距
type BoxSpacing struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// 边框字符
var borderChars = map[BorderType]BorderEdges{
	BorderNormal: {
		TL: "┌", T: "─", TR: "┐",
		L:  "│", R:  "│",
		BL: "└", B: "─", BR: "┘",
	},
	BorderRounded: {
		TL: "╭", T: "─", TR: "╮",
		L:  "│", R:  "│",
		BL: "╰", B: "─", BR: "╯",
	},
	BorderDouble: {
		TL: "╔", T: "═", TR: "╗",
		L:  "║", R:  "║",
		BL: "╚", B: "═", BR: "╝",
	},
	BorderThick: {
		TL: "┏", T: "━", TR: "┓",
		L:  "┃", R:  "┃",
		BL: "┗", B: "━", BR: "┛",
	},
}

// BorderEdges 边框边缘字符
type BorderEdges struct {
	TL string // Top Left
	T  string // Top
	TR string // Top Right
	L  string // Left
	R  string // Right
	BL string // Bottom Left
	B  string // Bottom
	BR string // Bottom Right
}

// NewBox 创建盒子
func NewBox() *Box {
	b := &Box{
		BaseContainer: component.NewBaseContainer("box"),
		border:        BorderStyle{
			Enabled: true,
			Top:     true,
			Bottom:  true,
			Left:    true,
			Right:   true,
			Style:   BorderNormal,
		},
		borderFg: style.BrightBlack,
		padding: BoxSpacing{Top: 0, Right: 0, Bottom: 0, Left: 0},
		margin:  BoxSpacing{Top: 0, Right: 0, Bottom: 0, Left: 0},
	}
	return b
}

// WithChild 设置子组件
func (b *Box) WithChild(child component.Component) *Box {
	b.Add(child)
	return b
}

// WithChildren 设置子组件列表
func (b *Box) WithChildren(children ...component.Component) *Box {
	for _, child := range children {
		b.Add(child)
	}
	return b
}

// WithBorder 设置边框
func (b *Box) WithBorder(enabled bool) *Box {
	b.border.Enabled = enabled
	return b
}

// WithBorderStyle 设置边框类型
func (b *Box) WithBorderStyle(borderType BorderType) *Box {
	b.border.Style = borderType
	return b
}

// WithBorderColor 设置边框颜色
func (b *Box) WithBorderColor(color style.Color) *Box {
	b.borderFg = color
	return b
}

// WithPadding 设置内边距
func (b *Box) WithPadding(all int) *Box {
	b.padding = BoxSpacing{
		Top:    all,
		Right:  all,
		Bottom: all,
		Left:   all,
	}
	return b
}

// WithPaddingV 设置垂直内边距
func (b *Box) WithPaddingV(vertical int) *Box {
	b.padding.Top = vertical
	b.padding.Bottom = vertical
	return b
}

// WithPaddingH 设置水平内边距
func (b *Box) WithPaddingH(horizontal int) *Box {
	b.padding.Left = horizontal
	b.padding.Right = horizontal
	return b
}

// WithMargin 设置外边距
func (b *Box) WithMargin(all int) *Box {
	b.margin = BoxSpacing{
		Top:    all,
		Right:  all,
		Bottom: all,
		Left:   all,
	}
	return b
}

// WithWidth 设置宽度
func (b *Box) WithWidth(width int) *Box {
	b.width = width
	return b
}

// WithHeight 设置高度
func (b *Box) WithHeight(height int) *Box {
	b.height = height
	return b
}

// WithSize 设置尺寸
func (b *Box) WithSize(width, height int) *Box {
	b.width = width
	b.height = height
	return b
}

// WithBackground 设置背景色
func (b *Box) WithBackground(color style.Color) *Box {
	b.bgColor = color
	return b
}

// WithStyle 设置样式
func (b *Box) WithStyle(s style.Style) *Box {
	b.SetStyle(s)
	return b
}

// Render 渲染盒子
func (b *Box) Render(ctx *component.RenderContext) string {
	if !b.IsVisible() {
		return ""
	}

	// 计算可用尺寸（减去外边距）
	availW := ctx.AvailableWidth - b.margin.Left - b.margin.Right
	availH := ctx.AvailableHeight - b.margin.Top - b.margin.Bottom

	if availW < 2 || availH < 2 {
		return ""
	}

	// 计算边框占用
	borderW := 0
	borderH := 0
	if b.border.Enabled {
		if b.border.Left {
			borderW++
		}
		if b.border.Right {
			borderW++
		}
		if b.border.Top {
			borderH++
		}
		if b.border.Bottom {
			borderH++
		}
	}

	// 计算内容区域尺寸
	contentW := availW - borderW - b.padding.Left - b.padding.Right
	contentH := availH - borderH - b.padding.Top - b.padding.Bottom

	if contentW < 0 {
		contentW = 0
	}
	if contentH < 0 {
		contentH = 0
	}

	// 渲染边框
	lines := b.renderBorder(availW, availH)

	// 渲染子组件
	var contentLines []string
	if b.ChildCount() > 0 {
		childCtx := ctx.WithOffset(
			b.margin.Left+borderW+b.padding.Left,
			b.margin.Top+borderH+b.padding.Top,
		)
		childCtx.AvailableWidth = contentW
		childCtx.AvailableHeight = contentH

		child := b.GetChild(0)
		childContent := child.Render(childCtx)
		contentLines = strings.Split(childContent, "\n")
	}

	// 合并内容到边框
	result := b.mergeContent(lines, contentLines, contentW, contentH)

	// 添加上边距
	topMargin := strings.Repeat("\n", b.margin.Top)

	// 添加每行左边距
	leftMargin := strings.Repeat(" ", b.margin.Left)

	result = topMargin + result
	if b.margin.Left > 0 {
		resultLines := strings.Split(result, "\n")
		for i := range resultLines {
			if len(resultLines[i]) > 0 {
				resultLines[i] = leftMargin + resultLines[i]
			}
		}
		result = strings.Join(resultLines, "\n")
	}

	return result
}

// renderBorder 渲染边框
func (b *Box) renderBorder(width, height int) []string {
	if !b.border.Enabled {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = ""
		}
		return lines
	}

	edges := borderChars[b.border.Style]
	borderStyle := style.NewStyle().Foreground(b.borderFg)

	lines := make([]string, height)

	// 顶部边框
	if b.border.Top {
		top := edges.TL
		midWidth := width - 2
		if midWidth > 0 {
			top += strings.Repeat(edges.T, midWidth)
		}
		if width > 1 {
			top += edges.TR
		}
		lines[0] = borderStyle.Apply(top)
	}

	// 中间行
	for y := 1; y < height-1; y++ {
		line := ""
		if b.border.Left {
			line += edges.L
		}
		contentWidth := width
		if b.border.Left {
			contentWidth--
		}
		if b.border.Right {
			contentWidth--
		}
		if contentWidth > 0 {
			line += strings.Repeat(" ", contentWidth)
		}
		if b.border.Right {
			line += edges.R
		}
		lines[y] = borderStyle.Apply(line)
	}

	// 底部边框
	if b.border.Bottom && height > 1 {
		bottom := edges.BL
		midWidth := width - 2
		if midWidth > 0 {
			bottom += strings.Repeat(edges.B, midWidth)
		}
		if width > 1 {
			bottom += edges.BR
		}
		lines[height-1] = borderStyle.Apply(bottom)
	}

	return lines
}

// mergeContent 合并内容到边框
func (b *Box) mergeContent(borderLines []string, contentLines []string, contentW, contentH int) string {
	if len(borderLines) == 0 {
		if len(contentLines) > 0 {
			return strings.Join(contentLines, "\n")
		}
		return ""
	}

	// 计算内容起始位置
	startX := 0
	startY := 0
	if b.border.Left {
		startX = 1
	}
	if b.border.Top {
		startY = 1
	}

	// 将内容插入边框
	result := make([]string, len(borderLines))
	copy(result, borderLines)

	for y, line := range contentLines {
		lineY := startY + y
		if lineY >= len(result) {
			break
		}

		// 获取当前行（移除 ANSI 样式以便处理）
		currentLine := stripANSI(result[lineY])

		// 插入内容
		paddedLine := line
		if utf8RuneCount(line) < contentW {
			paddedLine = line + strings.Repeat(" ", contentW-utf8RuneCount(line))
		}

		// 组合行
		if startX < len(currentLine) {
			if startX+len(paddedLine) < len(currentLine) {
				result[lineY] = currentLine[:startX] + paddedLine + currentLine[startX+len(paddedLine):]
			} else {
				result[lineY] = currentLine[:startX] + paddedLine
			}
		} else {
			result[lineY] = currentLine + paddedLine
		}
	}

	return strings.Join(result, "\n")
}

// stripANSI 移除 ANSI 转义码
func stripANSI(s string) string {
	result := ""
	inEscape := false
	for _, c := range s {
		if c == '\x1b' {
			inEscape = true
		} else if inEscape {
			if c == 'm' {
				inEscape = false
			}
		} else {
			result += string(c)
		}
	}
	return result
}

// GetPreferredSize 获取首选尺寸
func (b *Box) GetPreferredSize() (width, height int) {
	if b.width > 0 {
		width = b.width
	}
	if b.height > 0 {
		height = b.height
	}

	// 考虑边框和内边距
	if b.border.Enabled {
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

	width += b.padding.Left + b.padding.Right + b.margin.Left + b.margin.Right
	height += b.padding.Top + b.padding.Bottom + b.margin.Top + b.margin.Bottom

	return width, height
}

// 创建常用布局组件
type (
	// Fg 创建前景色样式
	Fg = style.Color
	// Bg 创建背景色样式
	Bg = style.Color
)

// Fprintf 格式化文本
func Fprintf(format string, args ...interface{}) string {
	return format
}
