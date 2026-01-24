package display

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
)

// List 列表组件
type List struct {
	*component.BaseComponent

	items     []ListItem
	cursor    int
	offset    int
	showCursor bool
}

// ListItem 列表项
type ListItem struct {
	Text      string
	Value     interface{}
	Secondary string
}

// NewList 创建列表
func NewList() *List {
	return &List{
		BaseComponent: component.NewBaseComponent("list"),
		items:         make([]ListItem, 0),
		cursor:        0,
		offset:        0,
		showCursor:    true,
	}
}

// NewListItems 创建带项目的列表
func NewListItems(items []string) *List {
	l := NewList()
	for _, item := range items {
		l.items = append(l.items, ListItem{Text: item})
	}
	return l
}

// SetItems 设置项目
func (l *List) SetItems(items []ListItem) {
	l.items = items
	if l.cursor >= len(items) {
		l.cursor = len(items) - 1
	}
}

// AddItem 添加项目
func (l *List) AddItem(item ListItem) {
	l.items = append(l.items, item)
}

// AddItemText 添加文本项目
func (l *List) AddItemText(text string) {
	l.items = append(l.items, ListItem{Text: text})
}

// ClearItems 清空项目
func (l *List) ClearItems() {
	l.items = l.items[:0]
	l.cursor = 0
	l.offset = 0
}

// SetCursor 设置光标位置
func (l *List) SetCursor(index int) {
	if index >= 0 && index < len(l.items) {
		l.cursor = index
		l.updateOffset()
	}
}

// GetCursor 获取光标位置
func (l *List) GetCursor() int {
	return l.cursor
}

// GetSelectedItem 获取选中项
func (l *List) GetSelectedItem() ListItem {
	if l.cursor >= 0 && l.cursor < len(l.items) {
		return l.items[l.cursor]
	}
	return ListItem{}
}

// GetSelectedText 获取选中项文本
func (l *List) GetSelectedText() string {
	item := l.GetSelectedItem()
	return item.Text
}

// CursorDown 光标下移
func (l *List) CursorDown() {
	if l.cursor < len(l.items)-1 {
		l.cursor++
		l.updateOffset()
	}
}

// CursorUp 光标上移
func (l *List) CursorUp() {
	if l.cursor > 0 {
		l.cursor--
		l.updateOffset()
	}
}

// PageDown 翻页下
func (l *List) PageDown() {
	l.cursor += 5
	if l.cursor >= len(l.items) {
		l.cursor = len(l.items) - 1
	}
	l.updateOffset()
}

// PageUp 翻页上
func (l *List) PageUp() {
	l.cursor -= 5
	if l.cursor < 0 {
		l.cursor = 0
	}
	l.updateOffset()
}

// Top 跳到顶部
func (l *List) Top() {
	l.cursor = 0
	l.offset = 0
}

// Bottom 跳到底部
func (l *List) Bottom() {
	l.cursor = len(l.items) - 1
	l.updateOffset()
}

// updateOffset 更新滚动偏移
func (l *List) updateOffset() {
	// 确保 cursor 在可见区域
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	// TODO: 需要知道可用高度来确定最大可见行数
}

// Render 渲染列表
func (l *List) Render(ctx *component.RenderContext) string {
	if !l.IsVisible() {
		return ""
	}

	width, height := ctx.AvailableWidth, ctx.AvailableHeight
	s := l.GetStyle()

	var result []string

	// 计算可见范围
	startIdx := l.offset
	maxVisible := height
	endIdx := startIdx + maxVisible
	if endIdx > len(l.items) {
		endIdx = len(l.items)
	}

	for i := startIdx; i < endIdx; i++ {
		if i >= len(l.items) {
			break
		}

		item := l.items[i]
		line := item.Text

		// 次要文本
		if item.Secondary != "" && width > len(line)+3 {
			line += " - " + item.Secondary
		}

		// 渲染光标
		prefix := " "
		if l.showCursor && i == l.cursor {
			prefix = "▶ "
			s = s.Reverse(true)
		}

		// 限制宽度
		if len(line) > width-3 {
			line = line[:width-3] + "..."
		}

		result = append(result, prefix+s.Apply(line))
	}

	// 填充剩余行
	for len(result) < height && len(result) < len(l.items) {
		result = append(result, strings.Repeat(" ", width))
	}

	return strings.Join(result, "\n")
}

// HandleEvent 处理事件
func (l *List) HandleEvent(ev component.Event) bool {
	switch e := ev.(type) {
	case *event.KeyEvent:
		return l.handleKey(e)
	}
	return false
}

// handleKey 处理键盘事件
func (l *List) handleKey(ev *event.KeyEvent) bool {
	switch ev.Special {
	case event.KeyUp, event.KeyK:
		l.CursorUp()
		return true
	case event.KeyDown, event.KeyJ:
		l.CursorDown()
		return true
	case event.KeyPageUp:
		l.PageUp()
		return true
	case event.KeyPageDown:
		l.PageDown()
		return true
	case event.KeyHome:
		l.Top()
		return true
	case event.KeyEnd:
		l.Bottom()
		return true
	}
	return false
}

// GetPreferredSize 获取首选尺寸
func (l *List) GetPreferredSize() (width, height int) {
	maxW := 0
	for _, item := range l.items {
		w := len(item.Text)
		if item.Secondary != "" {
			w += 3 + len(item.Secondary)
		}
		if w > maxW {
			maxW = w
		}
	}

	return maxW + 2, len(l.items)
}
