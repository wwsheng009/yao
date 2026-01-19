package examples

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

// 示例：如何使用统一消息处理工具重构 input 组件
//
// 使用统一消息处理工具重构后的 input 组件
func (w *InputComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 创建状态助手
	stateHelper := &core.InputStateHelper{
		Valuer:      w.model,
		Focuser:     w.model,
		ComponentID: w.model.id,
	}

	// 使用通用模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w, // 实现了 InteractiveBehavior 接口的组件
		msg,
		w.getBindings,           // 获取绑定（可选）
		w.handleBinding,         // 处理绑定（可选）
		w.delegateToBubbles,     // 委托给 bubbles 原组件
	)

	return w, cmd, response
}

func (w *InputComponentWrapper) getBindings() []core.ComponentBinding {
	return w.model.props.Bindings
}

func (w *InputComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// 复用已有的绑定处理逻辑
	cmd, response, handled := core.HandleBinding(keyMsg, binding, w.model.id)
	return cmd, response, handled
}

func (w *InputComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	w.model.Model, cmd := w.model.Model.Update(msg)
	return cmd
}

// 实现 InteractiveBehavior 接口的方法
func (w *InputComponentWrapper) HasFocus() bool {
	return w.model.Model.Focused()
}

func (w *InputComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// 只处理特定的特殊按键，否则返回 false 表示未处理
	switch keyMsg.Type {
	case tea.KeyEnter:
		// 发布输入提交事件
		cmd := core.PublishEvent(w.model.id, core.EventInputEnterPressed, map[string]interface{}{
			"value": w.model.Model.Value(),
		})
		return cmd, core.Handled, true
	case tea.KeyEscape:
		// 失焦处理
		w.model.Model.Blur()
		cmd := core.PublishEvent(w.model.id, core.EventEscapePressed, nil)
		return cmd, core.Handled, true
	}
	
	// 其他按键不由这个函数处理
	return nil, core.Ignored, false
}

// 实现 StateCapturable 接口
func (w *InputComponentWrapper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"value":   w.model.Model.Value(),
		"focused": w.model.Model.Focused(),
	}
}

func (w *InputComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["value"] != new["value"] {
		cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
			"oldValue": old["value"],
			"newValue": new["value"],
		}))
	}

	if old["focused"] != new["focused"] {
		cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputFocusChanged, map[string]interface{}{
			"focused": new["focused"],
		}))
	}

	return cmds
}

// 使用统一消息处理工具重构后的 list 组件示例
func (w *ListComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 创建状态助手
	stateHelper := &core.ListStateHelper{
		Indexer:     w.model,
		Selector:    w.model,
		Focused:     w.model.IsFocused(),
		ComponentID: w.model.id,
	}

	// 使用通用模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w, // 实现了 InteractiveBehavior 接口的组件
		msg,
		w.getBindings,           // 获取绑定（可选）
		w.handleBinding,         // 处理绑定（可选）
		w.delegateToBubbles,     // 委托给 bubbles 原组件
	)

	return w, cmd, response
}

// 重构前 vs 重构后对比
/*
重构前（原始 input 组件 UpdateMsg）：
- 70+ 行代码
- 手动实现所有层逻辑
- 重复的状态检测代码
- 分散的事件发布逻辑

重构后（使用统一消息处理工具）：
- 20-25 行核心代码
- 复用通用模板
- 标准化的状态检测
- 统一的事件发布模式
*/