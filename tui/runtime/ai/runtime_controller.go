package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/focus"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// =============================================================================
// RuntimeController (V3)
// =============================================================================
// RuntimeController Runtime 实现的 AI 控制器

// RuntimeController Runtime AI 控制器
type RuntimeController struct {
	dispatcher *action.Dispatcher
	tracker    *state.Tracker
	focusMgr   *focus.Manager
}

// NewRuntimeController 创建 Runtime AI 控制器
func NewRuntimeController(
	dispatcher *action.Dispatcher,
	tracker *state.Tracker,
	focusMgr *focus.Manager,
) *RuntimeController {
	return &RuntimeController{
		dispatcher: dispatcher,
		tracker:    tracker,
		focusMgr:   focusMgr,
	}
}

// =============================================================================
// 感知能力实现
// =============================================================================

// Inspect 获取当前 UI 状态快照
func (c *RuntimeController) Inspect() (*state.Snapshot, error) {
	return c.tracker.Current(), nil
}

// Find 查找组件（类似 DOM 选择器）
func (c *RuntimeController) Find(selector string) ([]ComponentInfo, error) {
	snapshot := c.tracker.Current()

	// ID 选择器: #input-username
	if strings.HasPrefix(selector, "#") {
		id := selector[1:]
		comp, ok := snapshot.GetComponent(id)
		if !ok {
			return nil, &ComponentNotFoundError{ID: id}
		}
		return []ComponentInfo{c.toComponentInfo(comp)}, nil
	}

	// 类型选择器: .TextInput
	if strings.HasPrefix(selector, ".") {
		typ := selector[1:]
		return c.findByType(snapshot, typ)
	}

	// 属性选择器: [placeholder="Email"]
	if strings.HasPrefix(selector, "[") && strings.HasSuffix(selector, "]") {
		return c.findByAttribute(snapshot, selector[1:len(selector)-1])
	}

	// 通配符: 匹配所有
	if selector == "*" {
		return c.findAll(snapshot), nil
	}

	return nil, &InvalidSelectorError{Selector: selector}
}

// findByType 按类型查找
func (c *RuntimeController) findByType(snapshot *state.Snapshot, typ string) ([]ComponentInfo, error) {
	results := make([]ComponentInfo, 0)
	for _, comp := range snapshot.Components {
		if comp.Type == typ {
			results = append(results, c.toComponentInfo(comp))
		}
	}
	return results, nil
}

// findByAttribute 按属性查找
func (c *RuntimeController) findByAttribute(snapshot *state.Snapshot, attr string) ([]ComponentInfo, error) {
	// 解析属性选择器
	// 例如: placeholder="Email" → key=placeholder, value=Email
	parts := strings.SplitN(attr, "=", 2)
	if len(parts) != 2 {
		return nil, &InvalidSelectorError{Selector: "[" + attr + "]"}
	}
	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

	results := make([]ComponentInfo, 0)
	for _, comp := range snapshot.Components {
		if compVal, ok := comp.Props[key]; ok && fmt.Sprintf("%v", compVal) == value {
			results = append(results, c.toComponentInfo(comp))
		}
		if compVal, ok := comp.State[key]; ok && fmt.Sprintf("%v", compVal) == value {
			results = append(results, c.toComponentInfo(comp))
		}
	}
	return results, nil
}

// findAll 查找所有组件
func (c *RuntimeController) findAll(snapshot *state.Snapshot) []ComponentInfo {
	results := make([]ComponentInfo, 0, len(snapshot.Components))
	for _, comp := range snapshot.Components {
		results = append(results, c.toComponentInfo(comp))
	}
	return results
}

// toComponentInfo 转换为 ComponentInfo
func (c *RuntimeController) toComponentInfo(comp state.ComponentState) ComponentInfo {
	return ComponentInfo{
		ID:       comp.ID,
		Type:     comp.Type,
		Props:    comp.Props,
		State:    comp.State,
		Rect:     comp.Rect,
		Visible:  comp.Visible,
		Disabled: comp.Disabled,
	}
}

// Query 查询状态
func (c *RuntimeController) Query(query StateQuery) (map[string]interface{}, error) {
	snapshot := c.tracker.Current()

	if query.ComponentID != "" {
		comp, ok := snapshot.GetComponent(query.ComponentID)
		if !ok {
			return nil, &ComponentNotFoundError{ID: query.ComponentID}
		}

		if query.StateKey != "" {
			return map[string]interface{}{
				query.StateKey: comp.State[query.StateKey],
			}, nil
		}

		return comp.State, nil
	}

	if query.ComponentType != "" {
		result := make(map[string]interface{})
		for id, comp := range snapshot.Components {
			if comp.Type == query.ComponentType {
				result[id] = comp.State
			}
		}
		return result, nil
	}

	// 返回所有状态
	result := make(map[string]interface{})
	for id, comp := range snapshot.Components {
		result[id] = comp.State
	}
	return result, nil
}

// WaitUntil 等待条件满足
func (c *RuntimeController) WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		snapshot := c.tracker.Current()
		if condition(snapshot) {
			return nil
		}

		if time.Now().After(deadline) {
			return &TimeoutError{Timeout: timeout}
		}

		<-ticker.C
	}
}

// =============================================================================
// 操作能力实现
// =============================================================================

// Dispatch 分发 Action
func (c *RuntimeController) Dispatch(a *action.Action) error {
	handled := c.dispatcher.Dispatch(a)
	if !handled {
		return fmt.Errorf("action not handled: %s", a.Type)
	}
	return nil
}

// Click 点击组件
func (c *RuntimeController) Click(componentID string) error {
	// 检查组件是否存在且可交互
	snapshot := c.tracker.Current()
	comp, ok := snapshot.GetComponent(componentID)
	if !ok {
		return &ComponentNotFoundError{ID: componentID}
	}
	if comp.Disabled {
		return &ComponentDisabledError{ID: componentID}
	}

	// 发送点击 Action（使用 submit 作为通用的点击/激活 action）
	return c.Dispatch(action.NewAction(action.ActionSubmit).
		WithTarget(componentID))
}

// Input 输入文本
func (c *RuntimeController) Input(componentID, text string) error {
	// 检查组件是否存在
	snapshot := c.tracker.Current()
	_, ok := snapshot.GetComponent(componentID)
	if !ok {
		return &ComponentNotFoundError{ID: componentID}
	}

	// 发送输入 Action
	return c.Dispatch(action.NewAction(action.ActionInputText).
		WithTarget(componentID).
		WithPayload(text))
}

// Navigate 焦点导航
func (c *RuntimeController) Navigate(direction Direction) error {
	if c.focusMgr == nil {
		return fmt.Errorf("focus manager not available")
	}

	var result string
	var ok bool

	switch direction {
	case DirectionUp:
		// 向上导航（使用前一个）
		result, ok = c.focusMgr.FocusPrev()
	case DirectionDown:
		// 向下导航（使用下一个）
		result, ok = c.focusMgr.FocusNext()
	case DirectionLeft:
		result, ok = c.focusMgr.FocusPrev()
	case DirectionRight:
		result, ok = c.focusMgr.FocusNext()
	case DirectionNext:
		result, ok = c.focusMgr.FocusNext()
	case DirectionPrev:
		result, ok = c.focusMgr.FocusPrev()
	case DirectionFirst:
		result, ok = c.focusMgr.FocusFirst()
	case DirectionLast:
		// 获取所有可聚焦组件，跳到最后一个
		comps := c.focusMgr.GetFocusableComponents()
		if len(comps) == 0 {
			return fmt.Errorf("no focusable components")
		}
		lastID := comps[len(comps)-1]
		ok = c.focusMgr.FocusSpecific(lastID)
		if ok {
			result = lastID
		}
	default:
		return fmt.Errorf("invalid direction: %v", direction)
	}

	if !ok || result == "" {
		return fmt.Errorf("navigation failed: no component in direction %v", direction)
	}

	return nil
}

// =============================================================================
// 高级能力实现
// =============================================================================

// Execute 执行操作序列
func (c *RuntimeController) Execute(ops ...Operation) error {
	for _, op := range ops {
		if err := op.Execute(c); err != nil {
			return err
		}
	}
	return nil
}

// Watch 监控状态变化
func (c *RuntimeController) Watch(callback func(*state.Snapshot)) func() {
	return c.tracker.Subscribe(func(old, new *state.Snapshot) {
		callback(new)
	})
}

// =============================================================================
// Helper Methods
// =============================================================================

// GetState 获取组件状态的快捷方法
func (c *RuntimeController) GetState(componentID, stateKey string) (interface{}, error) {
	result, err := c.Query(StateQuery{
		ComponentID: componentID,
		StateKey:   stateKey,
	})
	if err != nil {
		return nil, err
	}
	return result[stateKey], nil
}

// SetValue 设置组件状态的快捷方法
func (c *RuntimeController) SetValue(componentID string, stateKey string, value interface{}) error {
	c.tracker.SetComponentState(componentID, map[string]interface{}{
		stateKey: value,
	})
	return nil
}

// IsVisible 检查组件是否可见
func (c *RuntimeController) IsVisible(componentID string) (bool, error) {
	snapshot := c.tracker.Current()
	comp, ok := snapshot.GetComponent(componentID)
	if !ok {
		return false, &ComponentNotFoundError{ID: componentID}
	}
	return comp.Visible, nil
}

// IsDisabled 检查组件是否禁用
func (c *RuntimeController) IsDisabled(componentID string) (bool, error) {
	snapshot := c.tracker.Current()
	comp, ok := snapshot.GetComponent(componentID)
	if !ok {
		return false, &ComponentNotFoundError{ID: componentID}
	}
	return comp.Disabled, nil
}

// GetFocused 获取当前焦点组件 ID
func (c *RuntimeController) GetFocused() (string, error) {
	if c.focusMgr == nil {
		return "", fmt.Errorf("focus manager not available")
	}
	id, ok := c.focusMgr.GetFocused()
	if !ok {
		return "", fmt.Errorf("no focused component")
	}
	return id, nil
}

// WaitForVisible 等待组件可见
func (c *RuntimeController) WaitForVisible(componentID string, timeout time.Duration) error {
	return c.WaitUntil(func(s *state.Snapshot) bool {
		comp, ok := s.GetComponent(componentID)
		return ok && comp.Visible
	}, timeout)
}

// WaitForValue 等待组件状态值
func (c *RuntimeController) WaitForValue(componentID, stateKey string, expected interface{}, timeout time.Duration) error {
	return c.WaitUntil(func(s *state.Snapshot) bool {
		comp, ok := s.GetComponent(componentID)
		if !ok {
			return false
		}
		value, ok := comp.State[stateKey]
		if !ok {
			return false
		}
		return value == expected
	}, timeout)
}
