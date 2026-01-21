package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// CRUDState 定义CRUD组件的内部状态
type CRUDState int

const (
	StateList CRUDState = iota
	StateEditing
	StateCreating
	StateDeleting
	StateFiltering
)

// CRUDComponent represents a CRUD component with state machine support
type CRUDComponent struct {
	State            CRUDState
	Table            core.ComponentInterface
	Form             core.ComponentInterface
	Data             interface{}
	EventBus         *core.EventBus
	id               string
	unsubscribeFuncs []func()
}

// NewCRUDComponent creates a new CRUD component with the given ID
func NewCRUDComponent(config core.RenderConfig, id string) *CRUDComponent {
	component := &CRUDComponent{
		State:            StateList,
		id:               id,
		EventBus:         core.NewEventBus(),
		unsubscribeFuncs: []func(){},
	}

	if config.Data != nil {
		if err := component.UpdateRenderConfig(config); err != nil {
			log.Error("Failed to update CRUD component config: %v", err)
		}
	}

	return component
}

// Init initializes the CRUD component and sets up event subscriptions
func (c *CRUDComponent) Init() tea.Cmd {
	if c.EventBus != nil {
		unsubRowSelected := c.EventBus.Subscribe(core.EventRowSelected, func(msg core.ActionMsg) {
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if tableID, ok := data["tableID"].(string); ok && tableID == c.id {
					log.Trace("CRUD %s: Received ROW_SELECTED event, row index: %v", c.id, data["rowIndex"])
				}
			}
		})
		c.unsubscribeFuncs = append(c.unsubscribeFuncs, unsubRowSelected)

		unsubFormSubmit := c.EventBus.Subscribe(core.EventFormSubmitSuccess, func(msg core.ActionMsg) {
			log.Trace("CRUD %s: Received FORM_SUBMIT_SUCCESS event", c.id)
		})
		c.unsubscribeFuncs = append(c.unsubscribeFuncs, unsubFormSubmit)
	}

	return nil
}

// Cleanup unsubscribes from all events
func (c *CRUDComponent) Cleanup() {
	for _, unsub := range c.unsubscribeFuncs {
		unsub()
	}
	c.unsubscribeFuncs = []func(){}
}

// GetStateChanges returns the state changes from this component
func (c *CRUDComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (c *CRUDComponent) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
		"core.ActionMsg",
	}
}

// GetID returns the unique identifier for this component instance
func (c *CRUDComponent) GetID() string {
	return c.id
}

// View returns the string representation of the CRUD component
func (c *CRUDComponent) View() string {
	if c.Table != nil {
		return c.Table.View()
	}
	return "[CRUD Component]"
}

// CRUDStateHelper CRUD组件状态捕获助手
type CRUDStateHelper struct {
	StateHelper interface{ GetState() CRUDState }
	ComponentID string
}

func (h *CRUDStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"state": h.StateHelper.GetState(),
	}
}

func (h *CRUDStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	if old["state"] != new["state"] {
		// 状态改变事件，使用通用的焦点改变事件
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventFocusChanged, map[string]interface{}{
			"oldState": old["state"],
			"newState": new["state"],
		}))
	}

	return cmds
}

// CRUDComponentWrapper wraps CRUDComponent to implement ComponentInterface properly
type CRUDComponentWrapper struct {
	component   *CRUDComponent
	bindings    []core.ComponentBinding
	stateHelper *CRUDStateHelper
}

// NewCRUDComponentWrapper creates a wrapper that implements ComponentInterface
func NewCRUDComponentWrapper(config core.RenderConfig, id string) *CRUDComponentWrapper {
	component := NewCRUDComponent(config, id)

	wrapper := &CRUDComponentWrapper{
		component: component,
		bindings:  []core.ComponentBinding{}, // CRUD组件可能有自己的绑定
	}

	wrapper.stateHelper = &CRUDStateHelper{
		StateHelper: wrapper,
		ComponentID: id,
	}

	return wrapper
}

func (w *CRUDComponentWrapper) GetState() CRUDState {
	return w.component.State
}

func (w *CRUDComponentWrapper) Init() tea.Cmd {
	return w.component.Init()
}

func (w *CRUDComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,                   // 实现了 InteractiveBehavior 接口的组件
		msg,                 // 接收的消息
		w.getBindings,       // 获取按键绑定的函数
		w.handleBinding,     // 处理按键绑定的函数
		w.delegateToBubbles, // 委托给原 bubbles 组件的函数
	)

	return w, cmd, response
}

// 实现 InteractiveBehavior 接口的方法
func (w *CRUDComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *CRUDComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *CRUDComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// 委托给原始CRUD组件的UpdateMsg方法
	// 但先处理定向消息和动作消息
	switch msg := msg.(type) {
	case core.TargetedMsg:
		if msg.TargetID != w.component.id {
			return nil
		}
		// 递归调用，但需要防止无限递归
		// 直接处理内部消息
		return w.delegateToBubbles(msg.InnerMsg)

	case core.ActionMsg:
		switch msg.Action {
		case core.EventRowSelected:
			if w.component.State == StateList {
				w.component.State = StateEditing
				if data, ok := msg.Data.(map[string]interface{}); ok && w.component.Form != nil {
					log.Trace("CRUD %s: Loading row data into form from row index: %v", w.component.id, data["rowIndex"])
				}
				return core.PublishEvent(w.component.id, core.EventDataLoaded, map[string]interface{}{
					"transition": "StateList_to_StateEditing",
				})
			}

		case core.EventNewItemRequested:
			if w.component.State == StateList {
				w.component.State = StateCreating
				log.Trace("CRUD %s: Clearing form for new item", w.component.id)
				return core.PublishEvent(w.component.id, core.EventNewItemRequested, map[string]interface{}{
					"transition": "StateList_to_StateCreating",
				})
			}

		case core.EventItemDeleted:
			if w.component.State == StateList {
				log.Trace("CRUD %s: Item deleted, refreshing table", w.component.id)
				return core.PublishEvent(w.component.id, core.EventItemDeleted, map[string]interface{}{
					"state": "deleted",
				})
			}

		case core.EventFormSubmitSuccess:
			if w.component.State == StateEditing || w.component.State == StateCreating {
				w.component.State = StateList
				log.Trace("CRUD %s: Form submitted successfully, returning to list", w.component.id)
				return core.PublishEvent(w.component.id, core.EventFormSubmitSuccess, map[string]interface{}{
					"transition": "StateEditing_Creating_to_StateList",
				})
			}

		case core.EventFormCancel:
			if w.component.State == StateEditing || w.component.State == StateCreating {
				w.component.State = StateList
				log.Trace("CRUD %s: Form cancelled, returning to list", w.component.id)
				return core.PublishEvent(w.component.id, core.EventFormCancel, map[string]interface{}{
					"transition": "StateEditing_Creating_to_StateList",
				})
			}
		}
	}

	// 根据当前状态委托给相应的子组件
	switch w.component.State {
	case StateEditing, StateCreating:
		if w.component.Form != nil {
			_, cmd, _ = w.component.Form.UpdateMsg(msg)
		}

	case StateList, StateFiltering, StateDeleting:
		if w.component.Table != nil {
			_, cmd, _ = w.component.Table.UpdateMsg(msg)
		}
	}

	return cmd
}

// 实现 StateCapturable 接口
func (w *CRUDComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *CRUDComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *CRUDComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// CRUD组件可能有特殊键处理逻辑
	// 目前返回未处理，让框架继续处理
	return nil, core.Ignored, false
}
// GetModel returns the underlying model
func (w *CRUDComponentWrapper) GetModel() interface{} {
	return w.component
}

// GetID returns the component ID
func (w *CRUDComponentWrapper) GetID() string {
	return w.component.GetID()
}

// View returns the view of the component
func (w *CRUDComponentWrapper) View() string {
	return w.component.View()
}

// PublishEvent creates and returns a command to publish an event
func (w *CRUDComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *CRUDComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// 创建执行动作的命令
	return func() tea.Msg {
		return core.ActionMsg{
			ID:     w.component.id,
			Action: "EXECUTE_ACTION", // 使用通用动作名
			Data: map[string]interface{}{
				"action": action,
			},
		}
	}
}

// SetFocus sets or removes focus from the CRUD component
func (w *CRUDComponentWrapper) SetFocus(focus bool) {
	w.component.SetFocus(focus)
}

// GetFocus returns whether the CRUD component has focus
func (w *CRUDComponentWrapper) GetFocus() bool {
	return w.component.GetFocus()
}

// GetComponentType returns the component type
func (w *CRUDComponentWrapper) GetComponentType() string {
	return "crud"
}

// Render renders the component
func (w *CRUDComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.component.View(), nil
}

// UpdateRenderConfig updates the render configuration
func (w *CRUDComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	return w.component.UpdateRenderConfig(config)
}

// Cleanup cleans up resources
func (w *CRUDComponentWrapper) Cleanup() {
	w.component.Cleanup()
}

// GetStateChanges returns state changes
func (w *CRUDComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	return w.component.GetStateChanges()
}

// GetSubscribedMessageTypes returns subscribed message types
func (w *CRUDComponentWrapper) GetSubscribedMessageTypes() []string {
	return w.component.GetSubscribedMessageTypes()
}

// UpdateMsg routes messages based on internal state
func (c *CRUDComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Create a temporary wrapper to use the standard message handling
	wrapper := &CRUDComponentWrapper{
		component:   c,
		bindings:    []core.ComponentBinding{},
	}

	wrapper.stateHelper = &CRUDStateHelper{
		StateHelper: wrapper,
		ComponentID: c.id,
	}

	// Use the wrapper's UpdateMsg implementation for consistency
	_, cmd, response := wrapper.UpdateMsg(msg)
	// Note: In this case, the original component's state is modified directly in delegateToBubbles

	return c, cmd, response
}

// SetFocus sets or removes focus from the CRUD component
func (c *CRUDComponent) SetFocus(focus bool) {
	switch c.State {
	case StateList:
		if c.Table != nil {
			c.Table.SetFocus(focus)
		}
	case StateEditing, StateCreating:
		if c.Form != nil {
			c.Form.SetFocus(focus)
		}
	case StateFiltering, StateDeleting:
		if c.Table != nil {
			c.Table.SetFocus(focus)
		}
	}
}

func (c *CRUDComponent) GetFocus() bool {
	switch c.State {
	case StateList:
		if c.Table != nil {
			if focusGetter, ok := c.Table.(interface{ GetFocus() bool }); ok {
				return focusGetter.GetFocus()
			}
			return true
		}
	case StateEditing, StateCreating:
		if c.Form != nil {
			if focusGetter, ok := c.Form.(interface{ GetFocus() bool }); ok {
				return focusGetter.GetFocus()
			}
			return true
		}
	case StateFiltering, StateDeleting:
		if c.Table != nil {
			if focusGetter, ok := c.Table.(interface{ GetFocus() bool }); ok {
				return focusGetter.GetFocus()
			}
			return true
		}
	}
	return false
}

func (c *CRUDComponent) GetComponentType() string {
	return "crud"
}

func (c *CRUDComponent) UpdateRenderConfig(config core.RenderConfig) error {
	data := config.Data
	if data == nil {
		return nil
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	var items []interface{}
	var tableHeight, tableWidth int

	if h, ok := dataMap["height"].(int); ok {
		tableHeight = h
	}
	if w, ok := dataMap["width"].(int); ok {
		tableWidth = w
	}

	if bindData, ok := dataMap["__bind_data"].([]interface{}); ok {
		items = bindData
	}

	if len(items) == 0 {
		return nil
	}

	var columns []Column
	if firstItem, ok := items[0].(map[string]interface{}); ok {
		i := 0
		for key, val := range firstItem {
			width := 15
			if i == 0 {
				width = 8
			}
			_ = val
			columns = append(columns, Column{
				Key:   key,
				Title: key,
				Width: width,
			})
			i++
		}
	}

	if len(columns) == 0 {
		return nil
	}

	convertedData := make([][]interface{}, 0, len(items))
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			row := make([]interface{}, len(columns))
			for i, col := range columns {
				row[i] = itemMap[col.Key]
			}
			convertedData = append(convertedData, row)
		}
	}

	tableProps := TableProps{
		Columns:    columns,
		Data:       convertedData,
		Focused:    true,
		Height:     tableHeight,
		Width:      tableWidth,
		ShowBorder: true,
		Bindings:   []core.ComponentBinding{},
	}

	c.Table = NewTableComponentWrapper(tableProps, c.id)

	return nil
}

func (c *CRUDComponent) Render(config core.RenderConfig) (string, error) {
	return c.View(), nil
}
