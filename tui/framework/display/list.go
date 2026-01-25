package display

import (
	"fmt"
	"sync"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// List Component V3 (with Virtual Scrolling)
// ==============================================================================
// V3 列表组件，支持虚拟滚动、多选、键盘导航

// ListItem 列表项
type ListItem struct {
	Index    int
	Data     interface{}
	Selected bool
	Focused  bool
}

// List V3 列表组件
type List struct {
	*component.BaseComponent
	*component.StateHolder

	mu sync.RWMutex

	// 数据源
	dataSource component.DataSource

	// 视口状态
	offset   int // 滚动偏移（数据源索引）
	cursor   int // 当前光标位置（数据源索引）
	height   int // 可见高度

	// 选中状态（支持多选）
	selected map[int]bool

	// 样式
	normalStyle    style.Style
	selectedStyle  style.Style
	focusedStyle   style.Style
	cursorStyle    style.Style

	// 回调函数
	onSelect func(item interface{})
	onChange func()

	// 是否显示光标
	showCursor bool
}

// NewList 创建 V3 列表组件
func NewList(dataSource component.DataSource) *List {
	return &List{
		BaseComponent: component.NewBaseComponent("list"),
		StateHolder:     component.NewStateHolder(),
		dataSource:      dataSource,
		offset:          0,
		cursor:          0,
		height:          10,
		selected:        make(map[int]bool),
		normalStyle:     style.Style{},
		selectedStyle:   style.Style{}.Foreground(style.White),
		focusedStyle:    style.Style{}.Background(style.Blue),
		cursorStyle:     style.Style{}.Reverse(true),
		showCursor:      true,
	}
}

// NewListItems 从切片创建列表
func NewListItems(items []interface{}) *List {
	return NewList(component.NewSimpleDataSource(items))
}

// NewListStrings 从字符串切片创建列表
func NewListStrings(items []string) *List {
	return NewList(component.NewStringDataSource(items))
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetDataSource 设置数据源
func (l *List) SetDataSource(ds component.DataSource) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.dataSource = ds
	l.resetCursor()
	return l
}

// SetHeight 设置可见高度
func (l *List) SetHeight(h int) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.height = h
	return l
}

// SetShowCursor 设置是否显示光标
func (l *List) SetShowCursor(show bool) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showCursor = show
	return l
}

// SetNormalStyle 设置普通样式
func (l *List) SetNormalStyle(s style.Style) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.normalStyle = s
	return l
}

// SetSelectedStyle 设置选中样式
func (l *List) SetSelectedStyle(s style.Style) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.selectedStyle = s
	return l
}

// SetFocusedStyle 设置焦点样式
func (l *List) SetFocusedStyle(s style.Style) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.focusedStyle = s
	return l
}

// SetCursorStyle 设置光标样式
func (l *List) SetCursorStyle(s style.Style) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cursorStyle = s
	return l
}

// SetOnSelect 设置选中回调
func (l *List) SetOnSelect(fn func(interface{})) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onSelect = fn
	return l
}

// SetOnChange 设置变化回调
func (l *List) SetOnChange(fn func()) *List {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onChange = fn
	return l
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
func (l *List) Measure(maxWidth, maxHeight int) (width, height int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 计算最大文本宽度
	maxTextWidth := 0
	for i := 0; i < l.dataSource.Count(); i++ {
		text := l.formatItem(l.dataSource.Get(i))
		textWidth := len([]rune(text))
		if textWidth > maxTextWidth {
			maxTextWidth = textWidth
		}
	}

	width = maxTextWidth
	height = l.dataSource.Count()

	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}
	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}

	return width, height
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件到 CellBuffer
func (l *List) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !l.IsVisible() {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	// 更新可见高度
	l.height = height

	// 计算可见范围
	start, end := l.getVisibleRange()

	// 绘制可见项
	y := ctx.Y
	for i := start; i < end && y < ctx.Y+height; i++ {
		item := l.dataSource.Get(i)
		text := l.formatItem(item)

		// 确定样式
		drawStyle := l.normalStyle
		isSelected := l.selected[i]
		isCursor := l.showCursor && i == l.cursor

		if isSelected {
			drawStyle = l.selectedStyle
		}

		if l.IsFocused() && isCursor {
			drawStyle = l.cursorStyle
		} else if l.IsFocused() && isSelected {
			drawStyle = drawStyle.Background(style.Blue)
		}

		// 绘制单行
		l.paintLine(buf, ctx.X, y, text, width, drawStyle)
		y++
	}

	// 填充剩余空间
	for y < ctx.Y+height {
		l.paintLine(buf, ctx.X, y, "", width, style.Style{})
		y++
	}

	// 绘制滚动条提示
	if l.dataSource.Count() > height {
		l.paintScrollbar(ctx, buf, width)
	}
}

// paintLine 绘制单行文本
func (l *List) paintLine(buf *paint.Buffer, x, y int, text string, width int, s style.Style) {
	runes := []rune(text)
	for i := 0; i < width; i++ {
		if i < len(runes) {
			buf.SetCell(x+i, y, runes[i], s)
		} else {
			buf.SetCell(x+i, y, ' ', s)
		}
	}
}

// paintScrollbar 绘制滚动条
func (l *List) paintScrollbar(ctx component.PaintContext, buf *paint.Buffer, width int) {
	total := l.dataSource.Count()
	if total <= l.height {
		return
	}

	// 计算滚动条位置
	barHeight := max(1, l.height*l.height/total)
	barPos := l.offset * (l.height - barHeight) / (total - l.height)

	x := ctx.X + width - 1

	for i := 0; i < l.height; i++ {
		y := ctx.Y + i
		if i >= barPos && i < barPos+barHeight {
			buf.SetCell(x, y, '█', style.Style{}.Foreground(style.BrightBlack))
		} else {
			buf.SetCell(x, y, '│', style.Style{}.Foreground(style.BrightBlack))
		}
	}
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction 处理语义化 Action
func (l *List) HandleAction(a action.Action) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch a.Type {
	case action.ActionNavigateUp:
		l.moveCursor(-1)
		return true

	case action.ActionNavigateDown:
		l.moveCursor(1)
		return true

	case action.ActionNavigatePageUp:
		l.moveCursor(-l.height)
		return true

	case action.ActionNavigatePageDown:
		l.moveCursor(l.height)
		return true

	case action.ActionNavigateFirst:
		l.setCursor(0)
		return true

	case action.ActionNavigateLast:
		l.setCursor(l.dataSource.Count() - 1)
		return true

	case action.ActionSelectItem:
		l.toggleSelect(l.cursor)
		return true

	case action.ActionSubmit:
		if l.onSelect != nil && l.cursor >= 0 && l.cursor < l.dataSource.Count() {
			item := l.dataSource.Get(l.cursor)
			l.onSelect(item)
		}
		return true

	case action.ActionNavigateNext:
		// Tab 跳到下一个控件，不做处理
		return false
	}

	return false
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (l *List) FocusID() string {
	return l.ID()
}

// OnFocus 获得焦点时调用
func (l *List) OnFocus() {
	// 可以在这里添加获得焦点时的逻辑
}

// OnBlur 失去焦点时调用
func (l *List) OnBlur() {
	// 可以在这里添加失去焦点时的逻辑
}

// ============================================================================
// 公共方法
// ============================================================================

// Select 选中指定项
func (l *List) Select(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index >= 0 && index < l.dataSource.Count() {
		l.selected[index] = true
		l.fireChange()
	}
}

// Deselect 取消选中指定项
func (l *List) Deselect(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.selected, index)
	l.fireChange()
}

// ToggleSelect 切换选中状态
func (l *List) ToggleSelect(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.toggleSelect(index)
}

// IsSelected 检查是否选中
func (l *List) IsSelected(index int) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.selected[index]
}

// GetSelected 获取所有选中项的索引
func (l *List) GetSelected() []int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]int, 0, len(l.selected))
	for idx := range l.selected {
		result = append(result, idx)
	}
	return result
}

// GetSelectedItems 获取所有选中项
func (l *List) GetSelectedItems() []interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]interface{}, 0, len(l.selected))
	for idx := range l.selected {
		result = append(result, l.dataSource.Get(idx))
	}
	return result
}

// ClearSelection 清空所有选中
func (l *List) ClearSelection() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.selected = make(map[int]bool)
	l.fireChange()
}

// SetCursor 设置光标位置
func (l *List) SetCursor(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.setCursor(index)
}

// GetCursor 获取光标位置
func (l *List) GetCursor() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.cursor
}

// ScrollTo 滚动到指定项
func (l *List) ScrollTo(index int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index >= 0 && index < l.dataSource.Count() {
		l.setCursor(index)
		l.ensureVisible(index)
	}
}

// GetOffset 获取滚动偏移
func (l *List) GetOffset() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.offset
}

// GetItemCount 获取项目总数
func (l *List) GetItemCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.dataSource.Count()
}

// ============================================================================
// 内部方法
// ============================================================================

// getVisibleRange 获取可见范围
func (l *List) getVisibleRange() (start, end int) {
	start = l.offset
	end = min(l.offset+l.height, l.dataSource.Count())
	return
}

// moveCursor 移动光标
func (l *List) moveCursor(delta int) {
	newCursor := l.cursor + delta

	// 限制范围
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor >= l.dataSource.Count() {
		newCursor = l.dataSource.Count() - 1
	}

	l.setCursor(newCursor)
	l.ensureVisible(newCursor)
}

// setCursor 设置光标位置
func (l *List) setCursor(index int) {
	if index < 0 {
		index = 0
	}
	if index >= l.dataSource.Count() {
		index = l.dataSource.Count() - 1
	}

	if l.cursor != index {
		l.cursor = index
	}
}

// ensureVisible 确保指定项可见
func (l *List) ensureVisible(index int) {
	if index < l.offset {
		l.offset = index
	} else if index >= l.offset+l.height {
		l.offset = index - l.height + 1
	}

	// 限制偏移范围
	maxOffset := max(0, l.dataSource.Count()-l.height)
	if l.offset > maxOffset {
		l.offset = maxOffset
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

// toggleSelect 切换选中状态
func (l *List) toggleSelect(index int) {
	if index < 0 || index >= l.dataSource.Count() {
		return
	}

	if l.selected[index] {
		delete(l.selected, index)
	} else {
		l.selected[index] = true
	}
	l.fireChange()
}

// resetCursor 重置光标
func (l *List) resetCursor() {
	l.cursor = 0
	l.offset = 0
	l.selected = make(map[int]bool)
}

// fireChange 触发变化事件
func (l *List) fireChange() {
	if l.onChange != nil {
		l.onChange()
	}
}

// formatItem 格式化项为文本
func (l *List) formatItem(item interface{}) string {
	if item == nil {
		return ""
	}
	switch v := item.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", item)
	}
}

// resetCursor 重置光标
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
