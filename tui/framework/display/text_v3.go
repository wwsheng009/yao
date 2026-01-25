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

// TextV3 V3 文本显示组件
type TextV3 struct {
	*component.BaseComponentV3
	*component.StateHolder

	content  string
	lines    []string
	align    component.TextAlign
	wrap     bool
	maxLines int
	style    style.Style
}

// NewTextV3 创建 V3 文本组件
func NewTextV3(content string) *TextV3 {
	t := &TextV3{
		BaseComponentV3: component.NewBaseComponentV3("text"),
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

// NewStyledTextV3 创建带样式的 V3 文本组件
func NewStyledTextV3(content string, s style.Style) *TextV3 {
	t := NewTextV3(content)
	t.style = s
	return t
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetContent 设置内容
func (t *TextV3) SetContent(content string) *TextV3 {
	t.content = content
	t.updateLines()
	return t
}

// GetContent 获取内容
func (t *TextV3) GetContent() string {
	return t.content
}

// SetAlign 设置对齐方式
func (t *TextV3) SetAlign(align component.TextAlign) *TextV3 {
	t.align = align
	return t
}

// SetWrap 设置自动换行
func (t *TextV3) SetWrap(wrap bool) *TextV3 {
	t.wrap = wrap
	t.updateLines()
	return t
}

// SetMaxLines 设置最大行数
func (t *TextV3) SetMaxLines(max int) *TextV3 {
	t.maxLines = max
	return t
}

// SetStyle 设置样式
func (t *TextV3) SetStyle(s style.Style) *TextV3 {
	t.style = s
	return t
}

// GetStyle 获取样式
func (t *TextV3) GetStyle() style.Style {
	return t.style
}

// WithContent 链式设置内容
func (t *TextV3) WithContent(content string) *TextV3 {
	return t.SetContent(content)
}

// WithAlign 链式设置对齐
func (t *TextV3) WithAlign(align component.TextAlign) *TextV3 {
	return t.SetAlign(align)
}

// WithWrap 链式设置换行
func (t *TextV3) WithWrap(wrap bool) *TextV3 {
	return t.SetWrap(wrap)
}

// WithStyle 链式设置样式
func (t *TextV3) WithStyle(s style.Style) *TextV3 {
	return t.SetStyle(s)
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (t *TextV3) Measure(maxWidth, maxHeight int) (width, height int) {
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
func (t *TextV3) Paint(ctx component.PaintContext, buf *paint.Buffer) {
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
func (t *TextV3) updateLines() {
	t.lines = strings.Split(t.content, "\n")
}

// processLine 处理单行文本
func (t *TextV3) processLine(line string, width int) string {
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
func (t *TextV3) wrapLine(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	return string(runes[:width])
}

// alignLine 对齐行
func (t *TextV3) alignLine(line string, width int) string {
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
