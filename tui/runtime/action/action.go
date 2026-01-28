package action

import "time"

// ==============================================================================
// Action 系统 (V3)
// ==============================================================================
// Action 是语义化事件，不是原始按键
// Platform 只产生 RawInput，Runtime 负责转换 RawInput → Action
// Component 只处理 Action，不处理原始按键

// ActionType Action 类型
// 命名规范：动词 + 名词，表示语义化的用户意图
type ActionType string

// ==============================================================================
// 导航 Actions
// ==============================================================================
const (
	ActionNavigateFirst   ActionType = "navigate_first" // 跳到开头
	ActionNavigateLast    ActionType = "navigate_last"  // 跳到末尾
	ActionNavigateNext    ActionType = "navigate_next"   // 下一个
	ActionNavigatePrev    ActionType = "navigate_prev"   // 上一个
	ActionNavigateUp      ActionType = "navigate_up"     // 向上
	ActionNavigateDown    ActionType = "navigate_down"   // 向下
	ActionNavigateLeft    ActionType = "navigate_left"   // 向左
	ActionNavigateRight   ActionType = "navigate_right"  // 向右
	ActionNavigatePageUp   ActionType = "navigate_page_up"   // 上一页
	ActionNavigatePageDown ActionType = "navigate_page_down" // 下一页
)

// ==============================================================================
// 编辑 Actions
// ==============================================================================
const (
	ActionInputChar        ActionType = "input_char"         // 输入字符
	ActionInputText        ActionType = "input_text"         // 输入文本
	ActionDeleteChar       ActionType = "delete_char"        // 删除字符
	ActionDeleteWord       ActionType = "delete_word"        // 删除单词
	ActionDeleteLine       ActionType = "delete_line"        // 删除行
	ActionBackspace        ActionType = "backspace"          // 退格
	ActionCursorHome       ActionType = "cursor_home"        // 光标到行首
	ActionCursorEnd        ActionType = "cursor_end"         // 光标到行尾
	ActionCursorLeft       ActionType = "cursor_left"        // 光标左移
	ActionCursorRight      ActionType = "cursor_right"       // 光标右移
	ActionCursorWordLeft   ActionType = "cursor_word_left"   // 光标左移一词
	ActionCursorWordRight  ActionType = "cursor_word_right"  // 光标右移一词
	ActionSelectAll        ActionType = "select_all"         // 全选
	ActionSelectWord       ActionType = "select_word"        // 选择单词
	ActionSelectLine       ActionType = "select_line"        // 选择行
)

// ==============================================================================
// 表单 Actions
// ==============================================================================
const (
	ActionSubmit    ActionType = "submit"     // 提交表单
	ActionCancel    ActionType = "cancel"     // 取消
	ActionValidate  ActionType = "validate"   // 验证
	ActionReset     ActionType = "reset"      // 重置
	ActionClear     ActionType = "clear"      // 清空
)

// ==============================================================================
// 选择 Actions
// ==============================================================================
const (
	ActionSelectItem      ActionType = "select_item"       // 选择项
	ActionDeselectItem    ActionType = "deselect_item"     // 取消选择
	ActionToggleSelect    ActionType = "toggle_select"    // 切换选择
	ActionSelectRange     ActionType = "select_range"      // 范围选择
)

// ==============================================================================
// 鼠标 Actions
// ==============================================================================
const (
	ActionMouseClick      ActionType = "mouse_click"       // 鼠标点击
	ActionMouseDoubleClick ActionType = "mouse_double_click" // 鼠标双击
	ActionMousePress      ActionType = "mouse_press"       // 鼠标按下
	ActionMouseRelease    ActionType = "mouse_release"     // 鼠标释放
	ActionMouseMotion     ActionType = "mouse_motion"      // 鼠标移动
	ActionMouseWheel      ActionType = "mouse_wheel"       // 鼠标滚轮
)

// ==============================================================================
// 视图 Actions
// ==============================================================================
const (
	ActionScroll      ActionType = "scroll"       // 滚动
	ActionScrollUp    ActionType = "scroll_up"    // 向上滚动
	ActionScrollDown  ActionType = "scroll_down"  // 向下滚动
	ActionScrollLeft  ActionType = "scroll_left"  // 向左滚动
	ActionScrollRight ActionType = "scroll_right" // 向右滚动
	ActionZoomIn      ActionType = "zoom_in"      // 放大
	ActionZoomOut     ActionType = "zoom_out"     // 缩小
	ActionZoomReset   ActionType = "zoom_reset"   // 重置缩放
)

// ==============================================================================
// 窗口 Actions
// ==============================================================================
const (
	ActionQuit      ActionType = "quit"      // 退出
	ActionClose     ActionType = "close"     // 关闭窗口
	ActionMaximize  ActionType = "maximize"  // 最大化
	ActionMinimize  ActionType = "minimize"  // 最小化
	ActionFullscreen ActionType = "fullscreen" // 全屏
)

// ==============================================================================
// 系统 Actions
// ==============================================================================
const (
	ActionCopy     ActionType = "copy"     // 复制
	ActionCut      ActionType = "cut"      // 剪切
	ActionPaste    ActionType = "paste"    // 粘贴
	ActionUndo     ActionType = "undo"     // 撤销
	ActionRedo     ActionType = "redo"     // 重做
	ActionSearch   ActionType = "search"   // 搜索
	ActionHelp     ActionType = "help"     // 帮助
	ActionRefresh  ActionType = "refresh"  // 刷新
)

// ==============================================================================
// AI Actions (V3 新增)
// ==============================================================================
const (
	ActionAIInspect     ActionType = "ai_inspect"      // AI 检查 UI 状态
	ActionAIFind        ActionType = "ai_find"         // AI 查找组件
	ActionAIQuery       ActionType = "ai_query"        // AI 查询状态
	ActionAIDispatch    ActionType = "ai_dispatch"     // AI 分发 Action
	ActionAIWait        ActionType = "ai_wait"         // AI 等待状态
	ActionAIWatch       ActionType = "ai_watch"        // AI 监控状态
)

// ==============================================================================
// Action 结构体
// ==============================================================================

// Action 语义化 Action
// 所有状态变化必须能追溯到 Action
type Action struct {
	// Type Action 类型
	Type ActionType

	// Payload Action 携带的数据
	// 对于 ActionInputChar，Payload 是 rune
	// 对于 ActionInputText，Payload 是 string
	// 对于 ActionNavigate，Payload 可能是方向或步长
	Payload interface{}

	// Source Action 来源组件 ID
	Source string

	// Target Action 目标组件 ID
	Target string

	// Timestamp Action 时间戳
	Timestamp time.Time
}

// NewAction 创建 Action
func NewAction(typ ActionType) *Action {
	return &Action{
		Type:      typ,
		Timestamp: time.Now(),
	}
}

// WithPayload 设置 Payload
func (a *Action) WithPayload(payload interface{}) *Action {
	a.Payload = payload
	return a
}

// WithSource 设置 Source
func (a *Action) WithSource(source string) *Action {
	a.Source = source
	return a
}

// WithTarget 设置 Target
func (a *Action) WithTarget(target string) *Action {
	a.Target = target
	return a
}

// Clone 克隆 Action
func (a *Action) Clone() *Action {
	return &Action{
		Type:      a.Type,
		Payload:   a.Payload,
		Source:    a.Source,
		Target:    a.Target,
		Timestamp: a.Timestamp,
	}
}

// String 返回 Action 的字符串表示
func (a *Action) String() string {
	if a.Target != "" {
		return string(a.Type) + "{" + a.Target + "}"
	}
	return string(a.Type)
}
