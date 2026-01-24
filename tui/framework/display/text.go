package display

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
)

// Text 文本显示组件
type Text struct {
	*component.BaseComponent

	content   string
	lines     []string
	align     TextAlignType
	wrap      bool
	maxLines  int
	truncate  string
	ellipsis  string
}

// TextAlignType 文本对齐
type TextAlignType int

const (
	AlignLeftType TextAlignType = iota
	AlignCenterType
	AlignRightType
)

// NewText 创建文本组件
func NewText(content string) *Text {
	t := &Text{
		BaseComponent: component.NewBaseComponent("text"),
		content:       content,
		align:         AlignLeftType,
		wrap:          false,
		maxLines:      0,
		truncate:      "",
		ellipsis:      "...",
	}
	t.updateLines()
	return t
}

// NewStyledText 创建带样式的文本组件
func NewStyledText(content string, s style.Style) *Text {
	t := NewText(content)
	t.SetStyle(s)
	return t
}

// SetContent 设置内容
func (t *Text) SetContent(content string) {
	t.content = content
	t.updateLines()
}

// GetContent 获取内容
func (t *Text) GetContent() string {
	return t.content
}

// Append 追加文本
func (t *Text) Append(text string) {
	t.content += text
	t.updateLines()
}

// Clear 清空内容
func (t *Text) Clear() {
	t.content = ""
	t.updateLines()
}

// SetAlign 设置对齐方式
func (t *Text) SetAlign(align TextAlignType) *Text {
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

// SetTruncate 设置截断模式 ("head", "tail", "middle", "")
func (t *Text) SetTruncate(mode string) *Text {
	t.truncate = mode
	return t
}

// SetEllipsis 设置省略符号
func (t *Text) SetEllipsis(ellipsis string) *Text {
	t.ellipsis = ellipsis
	return t
}

// WithAlign 链式设置对齐
func (t *Text) WithAlign(align TextAlignType) *Text {
	return t.SetAlign(align)
}

// WithWrap 链式设置换行
func (t *Text) WithWrap(wrap bool) *Text {
	return t.SetWrap(wrap)
}

// WithStyle 链式设置样式
func (t *Text) WithStyle(s style.Style) *Text {
	t.SetStyle(s)
	return t
}

// updateLines 更新行
func (t *Text) updateLines() {
	t.lines = strings.Split(t.content, "\n")
}

// Render 渲染组件
func (t *Text) Render(ctx *component.RenderContext) string {
	if !t.IsVisible() {
		return ""
	}

	width, height := ctx.AvailableWidth, ctx.AvailableHeight
	s := t.GetStyle()

	// 处理每一行
	var result []string
	for i, line := range t.lines {
		if t.maxLines > 0 && i >= t.maxLines {
			break
		}

		// 处理截断
		processedLine := t.processLine(line, width)

		// 处理对齐
		if runeCount(processedLine) < width {
			processedLine = t.alignLine(processedLine, width)
		}

		result = append(result, s.Apply(processedLine))

		if len(result) >= height {
			break
		}
	}

	return strings.Join(result, "\n")
}

// processLine 处理单行文本
func (t *Text) processLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	lineLen := runeCount(line)

	// 处理自动换行
	if t.wrap && lineLen > width {
		return t.wrapLine(line, width)
	}

	// 处理截断
	if lineLen > width {
		return t.truncateLine(line, width)
	}

	return line
}

// wrapLine 换行处理
func (t *Text) wrapLine(line string, width int) string {
	runes := []rune(line)
	var result []string

	for len(runes) > 0 {
		if len(runes) <= width {
			result = append(result, string(runes))
			break
		}

		// 在宽度处分割
		result = append(result, string(runes[:width]))
		runes = runes[width:]
	}

	return strings.Join(result, "\n")
}

// truncateLine 截断处理
func (t *Text) truncateLine(line string, width int) string {
	runes := []rune(line)
	ellipsis := []rune(t.ellipsis)

	switch t.truncate {
	case "head":
		if len(runes) <= width {
			return line
		}
		keep := width - len(ellipsis)
		if keep < 0 {
			keep = 0
		}
		return string(ellipsis) + string(runes[len(runes)-keep:])

	case "middle":
		if len(runes) <= width {
			return line
		}
		half := (width - len(ellipsis)) / 2
		if half < 0 {
			half = 0
		}
		return string(runes[:half]) + string(ellipsis) + string(runes[len(runes)-half:])

	case "tail":
		if len(runes) <= width {
			return line
		}
		keep := width - len(ellipsis)
		if keep < 0 {
			keep = 0
		}
		return string(runes[:keep]) + string(ellipsis)

	default:
		if len(runes) > width {
			return string(runes[:width])
		}
		return line
	}
}

// alignLine 对齐行
func (t *Text) alignLine(line string, width int) string {
	lineLen := runeCount(line)
	if lineLen >= width {
		return line
	}

	padding := width - lineLen

	switch t.align {
	case AlignCenterType:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)

	case AlignRightType:
		return strings.Repeat(" ", padding) + line

	default:
		return line + strings.Repeat(" ", padding)
	}
}

// GetPreferredSize 获取首选尺寸
func (t *Text) GetPreferredSize() (width, height int) {
	maxWidth := 0
	for _, line := range t.lines {
		lineWidth := runeCount(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	lineCount := len(t.lines)
	if t.maxLines > 0 && lineCount > t.maxLines {
		lineCount = t.maxLines
	}

	return maxWidth, lineCount
}

// HandleEvent 处理事件
func (t *Text) HandleEvent(ev component.Event) bool {
	return false
}

// runeCount 计算 rune 数量
func runeCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
