package core

import (
	tea "github.com/charmbracelet/bubbletea"
)

// StateCapturable 状态捕获接口（组件实现）
type StateCapturable interface {
	// CaptureState 捕获当前状态
	// 组件根据自身需要返回需要监控的状态字段
	CaptureState() map[string]interface{}

	// DetectStateChanges 检测并发布状态变化事件
	// 组件根据自身逻辑发布特定的事件
	DetectStateChanges(oldState, newState map[string]interface{}) []tea.Cmd
}

// InteractiveBehavior 交互组件通用行为
// 这个接口封装了所有交互组件的共同行为模式
type InteractiveBehavior interface {
	ComponentInterface

	// 必须实现的状态捕获
	StateCapturable

	// 可选实现的焦点检查（支持无焦点组件）
	HasFocus() bool

	// 可选实现的自定义按键处理（返回是否已处理）
	HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, Response, bool)
}

// HandleTargetedMsg 处理 Layer 1: 定向消息
// 返回: (是否需要递归, 递归命令, 响应状态)
func HandleTargetedMsg(msg tea.Msg, componentID string) (shouldRecurse bool, recurseCmd tea.Cmd, response Response) {
	if msg, ok := msg.(TargetedMsg); ok {
		if msg.TargetID == componentID {
			// 需要递归处理内部消息
			return true, nil, Handled
		}
		// 不是发给本组件的消息
		return false, nil, Ignored
	}
	// 不是定向消息，继续处理
	return false, nil, Handled
}

// CheckFocus 统一焦点检查
// 返回: (是否通过焦点检查, 响应状态)
func CheckFocus(focused bool) (passed bool, response Response) {
	if !focused {
		return false, Ignored
	}
	return true, Handled
}

// HandleStateChanges 统一状态变化处理
func HandleStateChanges(c StateCapturable, updateCmd tea.Cmd) tea.Cmd {
	if updateCmd == nil {
		return func() tea.Msg {
			oldState := c.CaptureState()
			newState := c.CaptureState()
			eventCmds := c.DetectStateChanges(oldState, newState)
			if len(eventCmds) == 0 {
				return nil
			}
			return tea.Batch(eventCmds...)()
		}
	}

	return func() tea.Msg {
		oldState := c.CaptureState()
		updateMsg := updateCmd()
		newState := c.CaptureState()

		eventCmds := c.DetectStateChanges(oldState, newState)
		if len(eventCmds) == 0 {
			return updateMsg
		}

		// 如果 updateMsg 为 nil，只执行 eventCmds
		if updateMsg == nil {
			return tea.Batch(eventCmds...)()
		}

		// 将 updateMsg 转换为返回该消息的命令，然后与 eventCmds 一起批量执行
		updateCmdWrapper := func() tea.Msg { return updateMsg }
		return tea.Batch(append([]tea.Cmd{updateCmdWrapper}, eventCmds...)...)()
	}
}

// DefaultInteractiveUpdateMsg 交互组件通用消息处理模板
// 组件可以组合使用这个模板，只需要实现 InteractiveBehavior 接口
func DefaultInteractiveUpdateMsg(
	w InteractiveBehavior,
	msg tea.Msg,
	getBindings func() []ComponentBinding,
	handleBinding func(keyMsg tea.KeyMsg, binding ComponentBinding) (tea.Cmd, Response, bool),
	delegateUpdate func(msg tea.Msg) tea.Cmd,
) (tea.Cmd, Response) {

	// ═════════════════════════════════════════════════
	// Layer 1: 定向消息处理
	// ═════════════════════════════════════════════════
	_, _, response := HandleTargetedMsg(msg, w.GetID())
	if response == Ignored {
		return nil, Ignored
	}

	// ═════════════════════════════════════════════════
	// Layer 2: 按键消息分层
	// ═════════════════════════════════════════════════
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Layer 0: 组件绑定检查（最高优先级）
		if getBindings != nil && handleBinding != nil {
			bindings := getBindings()
			if matched, binding, handled := CheckComponentBindings(keyMsg, bindings, w.GetID()); matched {
				if handled {
					cmd, resp, _ := handleBinding(keyMsg, *binding)
					return cmd, resp
				}
				// useDefault = true，继续执行默认处理
			}
		}

		// Layer 2.1: 焦点检查（交互组件必需）
		if !w.HasFocus() {
			return nil, Ignored
		}

		// Layer 2.2: 自定义特殊按键处理（可选）
		if cmd, response, handled := w.HandleSpecialKey(keyMsg); handled {
			return cmd, response
		}

		// Layer 2.3: 委托给原组件处理
		return HandleStateChanges(w, delegateUpdate(keyMsg)), Handled
	}

	// ═════════════════════════════════════════════════
	// Layer 3: 非按键消息处理
	// ═════════════════════════════════════════════════
	return HandleStateChanges(w, delegateUpdate(msg)), Handled
}

// InputStateHelper 输入组件状态捕获助手
type InputStateHelper struct {
	Valuer      interface{ GetValue() string }
	Focuser     interface{ Focused() bool }
	ComponentID string
}

func (h *InputStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"value":   h.Valuer.GetValue(),
		"focused": h.Focuser.Focused(),
	}
}

func (h *InputStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["value"] != new["value"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, "INPUT_VALUE_CHANGED", map[string]interface{}{
			"oldValue": old["value"],
			"newValue": new["value"],
		}))
	}

	if old["focused"] != new["focused"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, "INPUT_FOCUS_CHANGED", map[string]interface{}{
			"focused": new["focused"],
		}))
	}

	return cmds
}

// ListStateHelper 列表组件状态捕获助手
type ListStateHelper struct {
	Indexer     interface{ Index() int }
	Selector    interface{ SelectedItem() interface{} }
	Focused     bool
	ComponentID string
}

func (h *ListStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"index":    h.Indexer.Index(),
		"selected": h.Selector.SelectedItem(),
		"focused":  h.Focused,
	}
}

func (h *ListStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["index"] != new["index"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, "LIST_SELECTION_CHANGED", map[string]interface{}{
			"oldIndex": old["index"],
			"newIndex": new["index"],
		}))
	}

	return cmds
}

// ChatStateHelper 聊天组件状态捕获助手
type ChatStateHelper struct {
	InputValuer interface{ GetValue() string }
	Focuser     interface{ Focused() bool }
	ComponentID string
}

func (h *ChatStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"input_value": h.InputValuer.GetValue(),
		"focused":     h.Focuser.Focused(),
	}
}

func (h *ChatStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["input_value"] != new["input_value"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, EventInputValueChanged, map[string]interface{}{
			"oldValue": old["input_value"],
			"newValue": new["input_value"],
		}))
	}

	if old["focused"] != new["focused"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, EventInputFocusChanged, map[string]interface{}{
			"focused": new["focused"],
		}))
	}

	return cmds
}

// ViewportStateHelper 视口组件状态捕获助手
type ViewportStateHelper struct {
	Scroller    interface{ GetOffset() int }
	ComponentID string
}

func (h *ViewportStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"offset": h.Scroller.GetOffset(),
	}
}

func (h *ViewportStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["offset"] != new["offset"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, "VIEWPORT_SCROLLED", map[string]interface{}{
			"oldOffset": old["offset"],
			"newOffset": new["offset"],
		}))
	}

	return cmds
}

// PaginatorStateHelper 分页器组件状态捕获助手
type PaginatorStateHelper struct {
	Pager       interface{ GetCurrentPage() int }
	ComponentID string
}

func (h *PaginatorStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"page": h.Pager.GetCurrentPage(),
	}
}

func (h *PaginatorStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["page"] != new["page"] {
		cmds = append(cmds, PublishEvent(h.ComponentID, "PAGINATOR_PAGE_CHANGED", map[string]interface{}{
			"oldPage": old["page"],
			"newPage": new["page"],
		}))
	}

	return cmds
}