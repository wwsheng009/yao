package selection

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// RuntimeAdapter 将文本选择功能集成到 Runtime 系统。
// 它作为 Runtime 和 Selection 模块之间的桥梁。
type RuntimeAdapter struct {
	textSelection *TextSelection
	enabled       bool
}

// NewRuntimeAdapter 创建一个新的 Runtime 适配器。
func NewRuntimeAdapter() *RuntimeAdapter {
	return &RuntimeAdapter{
		enabled: true,
	}
}

// OnRender 在每次渲染后调用，应用选择高亮。
// 这应该在 Runtime.Render() 返回之前调用。
func (a *RuntimeAdapter) OnRender(frame *runtime.Frame) {
	if !a.enabled || frame == nil || frame.Buffer == nil {
		return
	}

	// 初始化或更新选择系统
	if a.textSelection == nil {
		a.textSelection = NewTextSelection(frame.Buffer)
		a.textSelection.SetEnabled(a.enabled)
	} else {
		a.textSelection.UpdateBuffer(frame.Buffer)
	}

	// 应用选择高亮
	a.textSelection.ApplySelection()
}

// OnEvent 在事件分发前调用，处理选择相关的事件。
// 返回 true 表示事件被选择系统处理，不需要继续传播。
func (a *RuntimeAdapter) OnEvent(ev interface{}) bool {
	if !a.enabled || a.textSelection == nil {
		return false
	}
	return a.textSelection.HandleEvent(ev)
}

// GetSelection 返回文本选择系统实例。
func (a *RuntimeAdapter) GetSelection() *TextSelection {
	return a.textSelection
}

// SetEnabled 启用或禁用文本选择。
func (a *RuntimeAdapter) SetEnabled(enabled bool) {
	a.enabled = enabled
	if a.textSelection != nil {
		a.textSelection.SetEnabled(enabled)
	}
}

// IsEnabled 返回文本选择是否启用。
func (a *RuntimeAdapter) IsEnabled() bool {
	return a.enabled
}

// Copy 复制选中的文本到剪贴板。
func (a *RuntimeAdapter) Copy() (string, error) {
	if a.textSelection == nil {
		return "", nil
	}
	return a.textSelection.Copy()
}

// GetSelectedText 返回选中的文本。
func (a *RuntimeAdapter) GetSelectedText() string {
	if a.textSelection == nil {
		return ""
	}
	return a.textSelection.GetSelectedText()
}

// ClearSelection 清除当前选择。
func (a *RuntimeAdapter) ClearSelection() {
	if a.textSelection != nil {
		a.textSelection.Clear()
	}
}

// IsSelectionActive 返回是否有活动选择。
func (a *RuntimeAdapter) IsSelectionActive() bool {
	if a.textSelection == nil {
		return false
	}
	return a.textSelection.IsActive()
}

// SelectAll 选择全部文本。
func (a *RuntimeAdapter) SelectAll() {
	if a.textSelection != nil {
		a.textSelection.SelectAll()
	}
}

// IsSelected 返回指定位置是否被选中。
func (a *RuntimeAdapter) IsSelected(x, y int) bool {
	if a.textSelection == nil {
		return false
	}
	return a.textSelection.IsSelected(x, y)
}

// SetHighlightStyle 设置选择高亮样式。
func (a *RuntimeAdapter) SetHighlightStyle(style CellStyle) {
	if a.textSelection != nil {
		a.textSelection.SetHighlightStyle(style)
	}
}

// SetSelectionMode 设置选择模式。
func (a *RuntimeAdapter) SetSelectionMode(mode SelectionMode) {
	if a.textSelection != nil {
		a.textSelection.SetSelectionMode(mode)
	}
}

// IsClipboardSupported 返回剪贴板是否支持。
func (a *RuntimeAdapter) IsClipboardSupported() bool {
	if a.textSelection == nil {
		return false
	}
	return a.textSelection.IsClipboardSupported()
}

// ExtendSelection 扩展选择到指定位置。
func (a *RuntimeAdapter) ExtendSelection(x, y int) {
	if a.textSelection != nil {
		a.textSelection.ExtendSelection(x, y)
	}
}

// MoveSelectionStart 移动选择起始位置。
func (a *RuntimeAdapter) MoveSelectionStart(dx, dy int) {
	if a.textSelection != nil {
		a.textSelection.MoveSelectionStart(dx, dy)
	}
}

// MoveSelectionEnd 移动选择结束位置。
func (a *RuntimeAdapter) MoveSelectionEnd(dx, dy int) {
	if a.textSelection != nil {
		a.textSelection.MoveSelectionEnd(dx, dy)
	}
}

// IsDragging 返回是否正在拖动选择。
func (a *RuntimeAdapter) IsDragging() bool {
	if a.textSelection == nil {
		return false
	}
	return a.textSelection.IsDragging()
}

// GetSelectionRegion 返回选择区域。
func (a *RuntimeAdapter) GetSelectionRegion() SelectionRegion {
	if a.textSelection == nil {
		return SelectionRegion{}
	}
	return a.textSelection.GetSelectionRegion()
}

// GetSelectedCells 返回所有选中的单元格。
func (a *RuntimeAdapter) GetSelectedCells() []struct{ X, Y int } {
	if a.textSelection == nil {
		return nil
	}
	return a.textSelection.GetSelectedCells()
}

// SelectWord 选择指定位置的单词。
func (a *RuntimeAdapter) SelectWord(x, y int) {
	if a.textSelection != nil {
		a.textSelection.SelectWord(x, y)
	}
}

// SelectLine 选择指定行。
func (a *RuntimeAdapter) SelectLine(y int) {
	if a.textSelection != nil {
		a.textSelection.SelectLine(y)
	}
}
