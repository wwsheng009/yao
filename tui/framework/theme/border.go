package theme

// BorderType 边框类型
type BorderType int

const (
	BorderNormal BorderType = iota
	BorderRounded
	BorderDouble
	BorderThick
	BorderHidden
	BorderDashed
	BorderDotted
)

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

// BorderChars 边框字符映射
var BorderChars = map[BorderType]BorderEdges{
	BorderNormal: {
		TL: "┌", T: "─", TR: "┐",
		L: "│", R: "│",
		BL: "└", B: "─", BR: "┘",
	},
	BorderRounded: {
		TL: "╭", T: "─", TR: "╮",
		L: "│", R: "│",
		BL: "╰", B: "─", BR: "╯",
	},
	BorderDouble: {
		TL: "╔", T: "═", TR: "╗",
		L: "║", R: "║",
		BL: "╚", B: "═", BR: "╝",
	},
	BorderThick: {
		TL: "┏", T: "━", TR: "┓",
		L: "┃", R: "┃",
		BL: "┗", B: "━", BR: "┛",
	},
	BorderDashed: {
		TL: "┌", T: "┄", TR: "┐",
		L: "┆", R: "┆",
		BL: "└", B: "┄", BR: "┘",
	},
	BorderDotted: {
		TL: "┌", T: "┈", TR: "┐",
		L: "┊", R: "┊",
		BL: "└", B: "┈", BR: "┘",
	},
	BorderHidden: {
		TL: " ", T: " ", TR: " ",
		L: " ", R: " ",
		BL: " ", B: " ", BR: " ",
	},
}

// BorderStyle 边框样式
type BorderStyle struct {
	Enabled bool
	Top     bool
	Bottom  bool
	Left    bool
	Right   bool
	Style   BorderType
	FG      Color
	BG      Color
}

// NewBorder 创建默认边框
func NewBorder() BorderStyle {
	return BorderStyle{
		Enabled: true,
		Top:     true,
		Bottom:  true,
		Left:    true,
		Right:   true,
		Style:   BorderNormal,
	}
}

// All 启用所有边
func (b BorderStyle) All() BorderStyle {
	b.Top = true
	b.Bottom = true
	b.Left = true
	b.Right = true
	return b
}

// TopOnly 只启用顶部边
func (b BorderStyle) TopOnly() BorderStyle {
	b.Top = true
	b.Bottom = false
	b.Left = false
	b.Right = false
	return b
}

// BottomOnly 只启用底部边
func (b BorderStyle) BottomOnly() BorderStyle {
	b.Top = false
	b.Bottom = true
	b.Left = false
	b.Right = false
	return b
}

// Sides 只启用左右边
func (b BorderStyle) Sides() BorderStyle {
	b.Top = false
	b.Bottom = false
	b.Left = true
	b.Right = true
	return b
}

// WithStyle 设置边框类型
func (b BorderStyle) WithStyle(style BorderType) BorderStyle {
	b.Style = style
	return b
}

// WithColor 设置边框颜色
func (b BorderStyle) WithColor(fg Color) BorderStyle {
	b.FG = fg
	return b
}

// WithBackground 设置边框背景色
func (b BorderStyle) WithBackground(bg Color) BorderStyle {
	b.BG = bg
	return b
}

// Disable 禁用边框
func (b BorderStyle) Disable() BorderStyle {
	b.Enabled = false
	return b
}

// Enable 启用边框
func (b BorderStyle) Enable() BorderStyle {
	b.Enabled = true
	return b
}

// GetEdges 获取边框字符
func (b BorderStyle) GetEdges() BorderEdges {
	return BorderChars[b.Style]
}

// Render 渲染边框为字符串切片
func (b BorderStyle) Render(width, height int) []string {
	if !b.Enabled || width < 2 || height < 2 {
		return []string{}
	}

	edges := b.GetEdges()
	lines := make([]string, height)

	// 构建带颜色字符串的辅助函数
	makeLine := func(chars []rune) string {
		result := ""
		for _, ch := range chars {
			if !b.FG.IsNone() {
				result += b.FG.ToANSIString()
			}
			if !b.BG.IsNone() {
				result += b.BG.ToANSIBGString()
			}
			result += string(ch)
			if !b.FG.IsNone() || !b.BG.IsNone() {
				result += "\x1b[0m"
			}
		}
		return result
	}

	// 顶部边框
	if b.Top {
		topRunes := []rune(edges.TL)
		for i := 0; i < width-2; i++ {
			topRunes = append(topRunes, []rune(edges.T)...)
		}
		topRunes = append(topRunes, []rune(edges.TR)...)
		lines[0] = makeLine(topRunes)
	}

	// 中间行
	for y := 1; y < height-1; y++ {
		lineRunes := []rune{}
		if b.Left {
			lineRunes = append(lineRunes, []rune(edges.L)...)
		}
		for i := 0; i < width-2; i++ {
			lineRunes = append(lineRunes, ' ')
		}
		if b.Right {
			lineRunes = append(lineRunes, []rune(edges.R)...)
		}
		lines[y] = makeLine(lineRunes)
	}

	// 底部边框
	if b.Bottom && height > 1 {
		bottomRunes := []rune(edges.BL)
		for i := 0; i < width-2; i++ {
			bottomRunes = append(bottomRunes, []rune(edges.B)...)
		}
		bottomRunes = append(bottomRunes, []rune(edges.BR)...)
		lines[height-1] = makeLine(bottomRunes)
	}

	return lines
}

// GetBorderTop 返回顶部边框字符串
func (b BorderStyle) GetBorderTop(width int) string {
	if !b.Enabled || !b.Top || width < 2 {
		return ""
	}
	edges := b.GetEdges()
	result := edges.TL
	for i := 0; i < width-2; i++ {
		result += edges.T
	}
	result += edges.TR
	return result
}

// GetBorderBottom 返回底部边框字符串
func (b BorderStyle) GetBorderBottom(width int) string {
	if !b.Enabled || !b.Bottom || width < 2 {
		return ""
	}
	edges := b.GetEdges()
	result := edges.BL
	for i := 0; i < width-2; i++ {
		result += edges.B
	}
	result += edges.BR
	return result
}

// GetBorderLeft 返回左边框字符
func (b BorderStyle) GetBorderLeft() string {
	if !b.Enabled || !b.Left {
		return ""
	}
	return b.GetEdges().L
}

// GetBorderRight 返回右边框字符
func (b BorderStyle) GetBorderRight() string {
	if !b.Enabled || !b.Right {
		return ""
	}
	return b.GetEdges().R
}

// GetContentWidth 获取去除边框后的内容宽度
func (b BorderStyle) GetContentWidth(width int) int {
	if !b.Enabled {
		return width
	}
	contentWidth := width
	if b.Left {
		contentWidth--
	}
	if b.Right {
		contentWidth--
	}
	if contentWidth < 0 {
		contentWidth = 0
	}
	return contentWidth
}

// GetContentHeight 获取去除边框后的内容高度
func (b BorderStyle) GetContentHeight(height int) int {
	if !b.Enabled {
		return height
	}
	contentHeight := height
	if b.Top {
		contentHeight--
	}
	if b.Bottom {
		contentHeight--
	}
	if contentHeight < 0 {
		contentHeight = 0
	}
	return contentHeight
}

// BorderThickness 返回边框占用的空间
func (b BorderStyle) BorderThickness() (top, right, bottom, left int) {
	if !b.Enabled {
		return 0, 0, 0, 0
	}
	if b.Top {
		top = 1
	}
	if b.Bottom {
		bottom = 1
	}
	if b.Left {
		left = 1
	}
	if b.Right {
		right = 1
	}
	return
}
