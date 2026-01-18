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

// CRUDComponent 表示一个具有状态机的CRUD组件
type CRUDComponent struct {
	State            CRUDState
	Table            core.ComponentInterface // 内嵌表格组件
	Form             core.ComponentInterface // 内嵌表单组件
	Data             interface{}             // 当前数据
	EventBus         *core.EventBus          // 用于内部组件通信（避免循环依赖）
	id               string                  // Unique identifier for this component instance
	unsubscribeFuncs []func()                // Store unsubscribe functions for cleanup
}

// NewCRUDComponent creates a new CRUD component with the given ID and event bus
func NewCRUDComponent(id string, eventBus *core.EventBus) *CRUDComponent {
	return &CRUDComponent{
		State:            StateList,
		id:               id,
		EventBus:         eventBus,
		unsubscribeFuncs: []func(){},
	}
}

// Init initializes the CRUD component and sets up event subscriptions
func (c *CRUDComponent) Init() tea.Cmd {
	// Subscribe to table events if we have an event bus
	if c.EventBus != nil {
		// Subscribe to row selected events from table
		unsubRowSelected := c.EventBus.Subscribe(core.EventRowSelected, func(msg core.ActionMsg) {
			// Check if this event came from our own table
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if tableID, ok := data["tableID"].(string); ok && tableID == c.id {
					// This is a row selection from our own table
					log.Trace("CRUD %s: Received ROW_SELECTED event, row index: %v", c.id, data["rowIndex"])

					// In a real implementation, we would trigger a state transition here
					// For now, we just log the event
				}
			}
		})
		c.unsubscribeFuncs = append(c.unsubscribeFuncs, unsubRowSelected)

		// Subscribe to form submit success events
		unsubFormSubmit := c.EventBus.Subscribe(core.EventFormSubmitSuccess, func(msg core.ActionMsg) {
			// Handle form submission success
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
	// CRUD component state is managed by the wrapper
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
	// For now, return a placeholder view
	if c.Table != nil {
		return c.Table.View()
	}
	return "[CRUD Component]"
}

// UpdateMsg 实现 ComponentInterface 的 UpdateMsg 方法
// 根据内部状态将消息路由到相应的子组件
func (c *CRUDComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	if targetedMsg, ok := msg.(core.TargetedMsg); ok {
		if targetedMsg.TargetID != c.id {
			return c, nil, core.Ignored
		}
		return c.UpdateMsg(targetedMsg.InnerMsg)
	}

	// Handle internal action messages that trigger state transitions
	if actionMsg, ok := msg.(core.ActionMsg); ok {
		switch actionMsg.Action {
		case core.EventRowSelected:
			// Transition from List to Editing state
			if c.State == StateList {
				c.State = StateEditing
				// Load row data into form if available
				if data, ok := actionMsg.Data.(map[string]interface{}); ok && c.Form != nil {
					// Populate form with row data
					log.Trace("CRUD %s: Loading row data into form from row index: %v", c.id, data["rowIndex"])
				}
				return c, nil, core.Handled
			}

		case core.EventNewItemRequested:
			// Transition from List to Creating state
			if c.State == StateList {
				c.State = StateCreating
				// Clear form for new item
				log.Trace("CRUD %s: Clearing form for new item", c.id)
				return c, nil, core.Handled
			}

		case core.EventItemDeleted:
			// Stay in List state, refresh table
			if c.State == StateList {
				log.Trace("CRUD %s: Item deleted, refreshing table", c.id)
				return c, nil, core.Handled
			}

		case core.EventFormSubmitSuccess:
			// Transition back to List state after successful save
			if c.State == StateEditing || c.State == StateCreating {
				c.State = StateList
				log.Trace("CRUD %s: Form submitted successfully, returning to list", c.id)
				return c, nil, core.Handled
			}

		case core.EventFormCancel:
			// Cancel editing/creating, return to List state
			if c.State == StateEditing || c.State == StateCreating {
				c.State = StateList
				log.Trace("CRUD %s: Form cancelled, returning to list", c.id)
				return c, nil, core.Handled
			}
		}
	}

	// Route message based on current state
	var comp core.ComponentInterface
	var cmd tea.Cmd
	var response core.Response

	switch c.State {
	case StateEditing, StateCreating:
		// In editing/creating state, send message to form component
		if c.Form != nil {
			comp, cmd, response = c.Form.UpdateMsg(msg)
			return c, cmd, response
		}
		return c, nil, core.Ignored

	case StateList:
		// In list state, send message to table component
		if c.Table != nil {
			comp, cmd, response = c.Table.UpdateMsg(msg)
			return c, cmd, response
		}
		return c, nil, core.Ignored

	case StateFiltering:
		// In filtering state, send message to filter component (if any)
		if c.Table != nil {
			comp, cmd, response = c.Table.UpdateMsg(msg)
			return comp, cmd, response
		}
		return c, nil, core.Ignored

	case StateDeleting:
		// In deleting state, confirm and delete
		if c.Table != nil {
			comp, cmd, response = c.Table.UpdateMsg(msg)
			return comp, cmd, response
		}
		return c, nil, core.Ignored
	}

	// Default: ignore message
	return c, nil, core.Ignored
}

// SetFocus sets or removes focus from the CRUD component
func (c *CRUDComponent) SetFocus(focus bool) {
	// Delegate focus to the current active sub-component
	switch c.State {
	case StateList:
		if c.Table != nil {
			c.Table.SetFocus(focus)
		}
	case StateEditing, StateCreating:
		if c.Form != nil {
			c.Form.SetFocus(focus)
		}
	case StateFiltering:
		if c.Table != nil {
			c.Table.SetFocus(focus)
		}
	case StateDeleting:
		if c.Table != nil {
			c.Table.SetFocus(focus)
		}
	}
}

func (c *CRUDComponent) GetComponentType() string {
	return "crud"
}

func (c *CRUDComponent) UpdateRenderConfig(config core.RenderConfig) error {
	// CRUD组件目前不需要更新渲染配置
	// 未来可以在这里更新子组件的配置
	return nil
}

func (c *CRUDComponent) Render(config core.RenderConfig) (string, error) {
	return c.View(), nil
}

// CRUDComponentWrapper wraps CRUDComponent for unified factory interface
type CRUDComponentWrapper struct {
	component *CRUDComponent
}

// NewCRUDComponentWrapper creates a wrapper around CRUD component
func NewCRUDComponentWrapper(id string) *CRUDComponentWrapper {
	// Create a new event bus for this CRUD component instance
	eventBus := core.NewEventBus()
	return &CRUDComponentWrapper{
		component: NewCRUDComponent(id, eventBus),
	}
}

func (w *CRUDComponentWrapper) Init() tea.Cmd {
	return w.component.Init()
}

func (w *CRUDComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return w.component.UpdateMsg(msg)
}

func (w *CRUDComponentWrapper) View() string {
	return w.component.View()
}

func (w *CRUDComponentWrapper) GetID() string {
	return w.component.GetID()
}

func (w *CRUDComponentWrapper) SetFocus(focus bool) {
	w.component.SetFocus(focus)
}

func (w *CRUDComponentWrapper) GetComponentType() string {
	return w.component.GetComponentType()
}

func (w *CRUDComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.component.Render(config)
}

func (w *CRUDComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	return w.component.UpdateRenderConfig(config)
}

func (w *CRUDComponentWrapper) Cleanup() {
	w.component.Cleanup()
}

// GetStateChanges returns the state changes from this component
func (w *CRUDComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// CRUD component wraps a table, so delegate to the table component
	return w.component.GetStateChanges()
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *CRUDComponentWrapper) GetSubscribedMessageTypes() []string {
	return w.component.GetSubscribedMessageTypes()
}
