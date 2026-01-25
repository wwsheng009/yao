package display

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
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
		if textRuneCount(processedLine) < width {
			processedLine = t.alignLine(processedLine, width)
		}

		// 应用样式并绘制
		for x, char := range processedLine {
			buf.SetCell(ctx.X+x, ctx.Y+y, char, t.style)
		}

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
	lineLen := textRuneCount(line)
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
