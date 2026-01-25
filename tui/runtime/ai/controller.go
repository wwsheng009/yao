package ai

import (
	"time"

	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// =============================================================================
// Controller Interface (V3)
// =============================================================================
// Controller AI 控制器接口
// AI 与人类用户平级，通过语义化接口与 TUI 交互

// Controller AI 控制器接口
type Controller interface {
	// === 感知能力 ===

	// Inspect 获取当前完整的 UI 状态快照
	Inspect() (*state.Snapshot, error)

	// Find 查找组件（类似 DOM 选择器）
	// 支持选择器：
	//   - "#id" - ID 选择器
	//   - ".Type" - 类型选择器
	//   - "[key=value]" - 属性选择器
	Find(selector string) ([]ComponentInfo, error)

	// Query 查询状态
	Query(query StateQuery) (map[string]interface{}, error)

	// WaitUntil 等待特定状态出现
	WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error

	// === 操作能力 ===

	// Dispatch 向指定组件发送语义指令
	Dispatch(a *action.Action) error

	// Click 点击组件
	Click(componentID string) error

	// Input 输入文本
	Input(componentID, text string) error

	// Navigate 焦点导航
	Navigate(direction Direction) error

	// === 高级能力 ===

	// Execute 执行复杂操作序列
	Execute(ops ...Operation) error

	// Watch 监控状态变化
	Watch(callback func(*state.Snapshot)) func()
}

// Direction 导航方向
type Direction int

const (
	DirectionUp Direction = iota
	DirectionDown
	DirectionLeft
	DirectionRight
	DirectionNext
	DirectionPrev
	DirectionFirst
	DirectionLast
)

// =============================================================================
// ComponentInfo
// =============================================================================

// ComponentInfo 组件信息
type ComponentInfo struct {
	// 基本信息
	ID string
	// Type 组件类型
	Type string

	// 状态
	Props map[string]interface{} // 静态属性
	State map[string]interface{} // 动态状态

	// 布局位置
	Rect state.Rect

	// 可见性
	Visible bool

	// 可交互性
	Disabled bool

	// 父子关系
	ParentID string
	Children []string
}

// =============================================================================
// StateQuery
// =============================================================================

// StateQuery 状态查询
type StateQuery struct {
	// 组件筛选
	ComponentID   string
	ComponentType string

	// 状态键筛选
	StateKey string

	// 值筛选
	Value interface{}
}

// =============================================================================
// Errors
// =============================================================================

// 组件未找到错误
type ComponentNotFoundError struct {
	ID string
}

func (e *ComponentNotFoundError) Error() string {
	return "component not found: " + e.ID
}

// 组件已禁用错误
type ComponentDisabledError struct {
	ID string
}

func (e *ComponentDisabledError) Error() string {
	return "component is disabled: " + e.ID
}

// 无效选择器错误
type InvalidSelectorError struct {
	Selector string
}

func (e *InvalidSelectorError) Error() string {
	return "invalid selector: " + e.Selector
}

// 超时错误
type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return "timeout waiting for condition"
}
