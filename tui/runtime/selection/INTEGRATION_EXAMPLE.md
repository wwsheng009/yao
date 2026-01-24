// 此文件展示如何在 runtime_impl.go 中集成文本选择功能
// 注意：这是一个文档文件，不是实际的源代码
//
// ========== 在 tui/runtime/runtime_impl.go 中添加的代码 ==========

package selection

// 本文件仅用于文档目的

/*
以下是需要在 tui/runtime/runtime_impl.go 中修改的内容：

1. 添加导入：
    sel "github.com/yaoapp/yao/tui/runtime/selection"

2. 修改 RuntimeImpl 结构体，添加字段：
    selectionAdapter *sel.RuntimeAdapter
    selectionEnabled bool

3. 修改 NewRuntime 函数：
    selectionEnabled: true,
    selectionAdapter: sel.NewRuntimeAdapter(),

4. 修改 Render 方法（在返回 frame 之前）：
    if r.selectionEnabled {
        r.selectionAdapter.OnRender(&frame)
    }

5. 添加新的方法到 RuntimeImpl：

func (r *RuntimeImpl) EnableSelection(enabled bool) {
    r.selectionEnabled = enabled
    r.selectionAdapter.SetEnabled(enabled)
}

func (r *RuntimeImpl) IsSelectionEnabled() bool {
    return r.selectionEnabled && r.selectionAdapter.IsEnabled()
}

func (r *RuntimeImpl) HandleSelectionEvent(ev interface{}) bool {
    if !r.selectionEnabled {
        return false
    }
    return r.selectionAdapter.OnEvent(ev)
}

func (r *RuntimeImpl) CopySelection() (string, error) {
    return r.selectionAdapter.Copy()
}

func (r *RuntimeImpl) GetSelectedText() string {
    return r.selectionAdapter.GetSelectedText()
}

func (r *RuntimeImpl) ClearSelection() {
    r.selectionAdapter.ClearSelection()
}

func (r *RuntimeImpl) IsSelectionActive() bool {
    return r.selectionAdapter.IsSelectionActive()
}

func (r *RuntimeImpl) SelectAll() {
    r.selectionAdapter.SelectAll()
}

func (r *RuntimeImpl) IsSelected(x, y int) bool {
    return r.selectionAdapter.IsSelected(x, y)
}

func (r *RuntimeImpl) GetSelectionAdapter() *sel.RuntimeAdapter {
    return r.selectionAdapter
}

func (r *RuntimeImpl) IsClipboardSupported() bool {
    return r.selectionAdapter.IsClipboardSupported()
}
*/
