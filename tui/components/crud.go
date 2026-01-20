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

// UpdateMsg routes messages based on internal state
func (c *CRUDComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	if targetedMsg, ok := msg.(core.TargetedMsg); ok {
		if targetedMsg.TargetID != c.id {
			return c, nil, core.Ignored
		}
		return c.UpdateMsg(targetedMsg.InnerMsg)
	}

	if actionMsg, ok := msg.(core.ActionMsg); ok {
		switch actionMsg.Action {
		case core.EventRowSelected:
			if c.State == StateList {
				c.State = StateEditing
				if data, ok := actionMsg.Data.(map[string]interface{}); ok && c.Form != nil {
					log.Trace("CRUD %s: Loading row data into form from row index: %v", c.id, data["rowIndex"])
				}
				return c, nil, core.Handled
			}

		case core.EventNewItemRequested:
			if c.State == StateList {
				c.State = StateCreating
				log.Trace("CRUD %s: Clearing form for new item", c.id)
				return c, nil, core.Handled
			}

		case core.EventItemDeleted:
			if c.State == StateList {
				log.Trace("CRUD %s: Item deleted, refreshing table", c.id)
				return c, nil, core.Handled
			}

		case core.EventFormSubmitSuccess:
			if c.State == StateEditing || c.State == StateCreating {
				c.State = StateList
				log.Trace("CRUD %s: Form submitted successfully, returning to list", c.id)
				return c, nil, core.Handled
			}

		case core.EventFormCancel:
			if c.State == StateEditing || c.State == StateCreating {
				c.State = StateList
				log.Trace("CRUD %s: Form cancelled, returning to list", c.id)
				return c, nil, core.Handled
			}
		}
	}

	switch c.State {
	case StateEditing, StateCreating:
		if c.Form != nil {
			return c.Form.UpdateMsg(msg)
		}
		return c, nil, core.Ignored

	case StateList:
		if c.Table != nil {
			return c.Table.UpdateMsg(msg)
		}
		return c, nil, core.Ignored

	case StateFiltering:
		if c.Table != nil {
			return c.Table.UpdateMsg(msg)
		}
		return c, nil, core.Ignored

	case StateDeleting:
		if c.Table != nil {
			return c.Table.UpdateMsg(msg)
		}
		return c, nil, core.Ignored
	}

	return c, nil, core.Ignored
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
