package display

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Text Component V3
// ==============================================================================
// V3 文本显示组件，使用 CellBuffer 绘制

// Text V3 文本显示组件
type Text struct {
	*component.BaseComponent
	*component.StateHolder

	content  string
	lines    []string
	align    component.TextAlign
	wrap     bool
	maxLines int
	style    style.Style

	// 主题样式支持
	styleID string // 用于从主题获取样式的组件ID
	state   string // 组件状态（如 "focus", "disabled" 等）
}

// NewText 创建 V3 文本组件
func NewText(content string) *Text {
	t := &Text{
		BaseComponent: component.NewBaseComponent("text"),
		StateHolder:     component.NewStateHolder(),
		content:         content,
		align:           component.AlignLeft,
		wrap:            false,
		maxLines:        0,
		style:           style.Style{},
	}
	t.updateLines()
	return t
}

// NewStyledText 创建带样式的 V3 文本组件
func NewStyledText(content string, s style.Style) *Text {
	t := NewText(content)
	t.style = s
	return t
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetContent 设置内容
func (t *Text) SetContent(content string) *Text {
	t.content = content
	t.updateLines()
	return t
}

// GetContent 获取内容
func (t *Text) GetContent() string {
	return t.content
}

// SetAlign 设置对齐方式
func (t *Text) SetAlign(align component.TextAlign) *Text {
	t.align = align
	return t
}

// SetWrap 设置自动换行
func (t *Text) SetWrap(wrap bool) *Text {
	t.wrap = wrap
	t.updateLines()
	return t
}

// SetMaxLines 设置最大行数
func (t *Text) SetMaxLines(max int) *Text {
	t.maxLines = max
	return t
}

// SetStyle 设置样式
func (t *Text) SetStyle(s style.Style) *Text {
	t.style = s
	return t
}

// GetStyle 获取样式
func (t *Text) GetStyle() style.Style {
	return t.style
}

// WithContent 链式设置内容
func (t *Text) WithContent(content string) *Text {
	return t.SetContent(content)
}

// WithAlign 链式设置对齐
func (t *Text) WithAlign(align component.TextAlign) *Text {
	return t.SetAlign(align)
}

// WithWrap 链式设置换行
func (t *Text) WithWrap(wrap bool) *Text {
	return t.SetWrap(wrap)
}

// WithStyle 链式设置样式
func (t *Text) WithStyle(s style.Style) *Text {
	return t.SetStyle(s)
}

// SetStyleID 设置主题样式ID
func (t *Text) SetStyleID(styleID string) *Text {
	t.styleID = styleID
	return t
}

// SetState 设置组件状态
func (t *Text) SetState(state string) *Text {
	t.state = state
	return t
}

// GetStyleID 获取主题样式ID
func (t *Text) GetStyleID() string {
	return t.styleID
}

// GetState 获取组件状态
func (t *Text) GetState() string {
	return t.state
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (t *Text) Measure(maxWidth, maxHeight int) (width, height int) {
	maxLineWidth := 0
	for _, line := range t.lines {
		lineWidth := textRuneCount(line)
		if lineWidth > maxLineWidth {
			maxLineWidth = lineWidth
		}
	}

	// 考虑换行
	if t.wrap && maxWidth > 0 && maxLineWidth > maxWidth {
		maxLineWidth = maxWidth
	}

	lineCount := len(t.lines)
	if t.maxLines > 0 && lineCount > t.maxLines {
		lineCount = t.maxLines
	}

	if maxHeight > 0 && lineCount > maxHeight {
		lineCount = maxHeight
	}

	// 确保至少返回 1 行高度（即使是空文本）
	if lineCount == 0 {
		lineCount = 1
	}

	return maxLineWidth, lineCount
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件到 CellBuffer
func (t *Text) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !t.IsVisible() {
		return
	}

	width := ctx.AvailableWidth
	if width <= 0 {
		return
	}

	height := ctx.AvailableHeight
	if height <= 0 {
		return
	}

	// 获取样式：如果设置了 styleID，从主题系统动态获取
	paintStyle := t.style
	if t.styleID != "" {
		paintStyle = style.GetStyle(t.styleID, t.state)
	}

	// 处理每一行
	y := 0
	for i, line := range t.lines {
		if t.maxLines > 0 && i >= t.maxLines {
			break
		}
		if y >= height {
			break
		}

		// 处理单行
		processedLine := t.processLine(line, width)

		// 处理对齐
		if textDisplayWidth(processedLine) < width {
			processedLine = t.alignLine(processedLine, width)
		}

		// 使用 PaintContext 的绘制方法（自动处理坐标偏移）
		// ctx.SetString 会自动处理宽字符和边界检查
		ctx.SetString(0, y, processedLine, paintStyle)

		y++
	}
}

// ============================================================================
// 内部方法
// ============================================================================

// updateLines 更新行
func (t *Text) updateLines() {
	t.lines = strings.Split(t.content, "\n")
}

// processLine 处理单行文本
func (t *Text) processLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	lineLen := textRuneCount(line)

	// 处理自动换行
	if t.wrap && lineLen > width {
		return t.wrapLine(line, width)
	}

	// 处理截断
	if lineLen > width {
		return line[:width]
	}

	return line
}

// wrapLine 换行处理（返回第一行）
func (t *Text) wrapLine(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	return string(runes[:width])
}

// alignLine 对齐行
func (t *Text) alignLine(line string, width int) string {
	lineLen := textDisplayWidth(line)
	if lineLen >= width {
		return line
	}

	padding := width - lineLen

	switch t.align {
	case component.AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)

	case component.AlignRight:
		return strings.Repeat(" ", padding) + line

	default: // component.AlignLeft
		return line + strings.Repeat(" ", padding)
	}
}

// textRuneCount 计算 rune 数量
func textRuneCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}

// textDisplayWidth 计算文本的显示宽度（考虑宽字符）
func textDisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeWidth(r)
	}
	return width
}

// runeWidth 返回字符的显示宽度 (1 或 2)
func runeWidth(r rune) int {
	// CJK 字符范围 (中文、日文、韩文等)
	if r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x20000 && r <= 0x2fffd) ||
		(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	// Emoji 和其他符号
	if r >= 0x1f300 && r <= 0x1f9f0 {
		return 2
	}
	return 1
}
