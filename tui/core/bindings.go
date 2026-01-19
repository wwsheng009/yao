package core

import (
	"time"

	keypkg "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
)

// ComponentBinding 定义组件级别的按键绑定
// 支持三种配置模式：Action/Event/Default
type ComponentBinding struct {
	// Key 按键定义（必填）
	Key string `json:"key"`
	
	// 模式1: Action - 强大的回调支持 Process/Script/Payload（优先级最高）
	Action *Action `json:"action,omitempty"`
	
	// 模式2: Event - 简单事件发布（优先级其次）
	Event string `json:"event,omitempty"`
	
	// 模式3: UseDefault - 回退到原组件默认行为（优先级最低）
	UseDefault bool `json:"useDefault,omitempty"`
	
	// Optional fields
	Enabled     bool   `json:"enabled"`                 // 默认 true
	Description string `json:"description,omitempty"`   // 帮助信息显示
	Shortcut    string `json:"shortcut,omitempty"`      // 覆盖 Key 的显示文本
}

// ExecuteActionMsg 用于将组件绑定的 Action 发送到全局 Model 执行
type ExecuteActionMsg struct {
	Action    *Action
	SourceID  string
	Timestamp time.Time
}

// WithBindings 组件辅助接口（可选）
type WithBindings interface {
	SetBindings(bindings []ComponentBinding)
	GetBindings() []ComponentBinding
}

// CheckComponentBindings 快捷绑定匹配函数
// 返回: (是否匹配, 绑定配置, 是否已处理)
func CheckComponentBindings(
	keyMsg tea.KeyMsg,
	bindings []ComponentBinding,
	componentID string,
) (bool, *ComponentBinding, bool) {
	
	for _, binding := range bindings {
		if !binding.Enabled {
			continue
		}
		
		kb := keypkg.NewBinding(keypkg.WithKeys(binding.Key))
		if keypkg.Matches(keyMsg, kb) {
			if binding.UseDefault {
				// UseDefault 模式: 不拦截，让默认处理继续
				return true, &binding, false
			}
			// Action 或 Event 模式: 需要拦截并处理
			return true, &binding, true
		}
	}
	
	return false, nil, false
}

// ComponentWrapper 是包装器接口，用于访问组件模型
// 这个接口需要在实际组件包装器中实现
type ComponentWrapper interface {
	ExecuteAction(action *Action) tea.Cmd
	PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd
	GetModel() interface{}
	GetID() string
}

// HandleBinding 处理组件自定义按键绑定（三种模式智能分发）
// 这是组件 Wrapper 的方法，由各个组件实现
// 返回: (命令, 响应状态, 是否已处理)
func HandleBinding(
	wrapper ComponentWrapper,
	keyMsg tea.KeyMsg,
	binding ComponentBinding,
) (tea.Cmd, Response, bool) {
	
	// ═════════════════════════════════════════════════
	// 模式1: Action - Process/Script/Payload（最高优先级）
	// ═══════════════════════════════════════════━━━━━━━━━━
	if binding.Action != nil {
		log.Trace("Component[%s] Execute Action: %s", wrapper.GetID(), binding.Action.Process)
		return wrapper.ExecuteAction(binding.Action), Handled, true
	}
	
	// ═════════════════════════════════════════════════
	// 模式2: Event - 简单事件发布（次优先级）
	// ═══════════════════════════════════════════━━━━━━━━━━
	if binding.Event != "" {
		log.Trace("Component[%s] Publish Event: %s", wrapper.GetID(), binding.Event)
		
		// 收集组件上下文数据
		model := wrapper.GetModel()
		eventData := map[string]interface{}{
			"key":      binding.Key,
			"type":     keyMsg.Type.String(),
			"ctrl":     keyMsg.Type >= tea.KeyRunes && keyMsg.Type <= tea.KeySpace || keyMsg.Type == tea.KeyCtrlC || keyMsg.Type == tea.KeyCtrlD || keyMsg.Type == tea.KeyCtrlU || keyMsg.Type == tea.KeyCtrlW,
			"alt":      keyMsg.Alt,
		}
		
		// 尝试添加组件特定数据（需要类型断言）
		if valuer, ok := model.(interface{ GetValue() string }); ok {
			eventData["value"] = valuer.GetValue()
		}
		if selector, ok := model.(interface{ GetSelected() (interface{}, bool) }); ok {
			if item, found := selector.GetSelected(); found {
				eventData["selected"] = item
			}
		}
		if indexer, ok := model.(interface{ GetIndex() int }); ok {
			eventData["index"] = indexer.GetIndex()
		}
		
		eventCmd := wrapper.PublishEvent(wrapper.GetID(), binding.Event, eventData)
		return eventCmd, Handled, true
	}
	
	// 未配置任何处理（不应该到达这里）
	return nil, Ignored, false
}

// executeAction 执行绑定的 Action（Process/Script/Payload）
func executeAction(wrapper ComponentWrapper, action *Action) tea.Cmd {
	if action == nil {
		return nil
	}
	
	// 复制 action 以避免修改原配置
	actionCopy := *action
	
	// 智能参数注入：自动添加组件上下文
	if actionCopy.Args == nil {
		actionCopy.Args = []interface{}{}
	}
	
	// 构建上下文地图
	context := map[string]interface{}{
		"componentID": wrapper.GetID(),
		"timestamp":   time.Now(),
	}
	
	model := wrapper.GetModel()
	// 尝试添加组件特定数据
	if valuer, ok := model.(interface{ GetValue() string }); ok {
		context["value"] = valuer.GetValue()
	}
	if selector, ok := model.(interface{ GetSelected() (interface{}, bool) }); ok {
		if item, found := selector.GetSelected(); found {
			context["selected"] = item
		}
	}
	if indexer, ok := model.(interface{ GetIndex() int }); ok {
		context["index"] = indexer.GetIndex()
	}
	
	// 添加到参数列表（如果未配置 args，自动注入 context）
	if len(actionCopy.Args) == 0 {
		actionCopy.Args = []interface{}{context}
	}
	
	// 发送到全局 Model 执行
	return func() tea.Msg {
		return ExecuteActionMsg{
			Action:    &actionCopy,
			SourceID:  wrapper.GetID(),
			Timestamp: time.Now(),
		}
	}
}