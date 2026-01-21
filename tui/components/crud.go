package components

import (
	"encoding/json"
	"fmt"
	"strings"

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

// 辅助函数
func getString(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getInt(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok { // JSON numbers are float64
			return int(f)
		}
		if i, ok := v.(int); ok {
			return i
		}
	}
	return defaultValue
}

func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getMap(m map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
	if v, ok := m[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return defaultValue
}

// ParseCRUDProps 从配置数据解析CRUD属性
func ParseCRUDProps(data map[string]interface{}) CRUDProps {
	props := CRUDProps{
		Title:      getString(data, "title", "CRUD"),
		Height:     getInt(data, "height", 0),
		Width:      getInt(data, "width", 0),
		Focused:    getBool(data, "focused", true),
		Bindings:   []core.ComponentBinding{},
		DataAPI:    getMap(data, "data_api", map[string]interface{}{}),
		FormFields: []Field{},
		TableProps: TableProps{},
		Actions:    map[string]*core.Action{},
	}
	
	// 解析绑定
	if bindingsData, ok := data["bindings"].([]interface{}); ok {
		for _, bindingData := range bindingsData {
			if bindingMap, ok := bindingData.(map[string]interface{}); ok {
				var binding core.ComponentBinding
				if bytes, err := json.Marshal(bindingMap); err == nil {
					if err := json.Unmarshal(bytes, &binding); err == nil {
						props.Bindings = append(props.Bindings, binding)
					}
				}
			}
		}
	}
	
	return props
}

// CRUDProps 定义CRUD组件的属性
type CRUDProps struct {
	Title      string                         `json:"title"`
	Height     int                           `json:"height"`
	Width      int                           `json:"width"`
	Focused    bool                          `json:"focused"`
	Bindings   []core.ComponentBinding       `json:"bindings"`  // 快捷键绑定
	DataAPI    map[string]interface{}        `json:"data_api"`  // 数据操作API配置
	FormFields []Field                       `json:"form_fields"` // 表单字段配置
	TableProps TableProps                    `json:"table_props"` // 表格属性
	Actions    map[string]*core.Action       `json:"actions"`     // 自定义操作
}

// CRUDComponent represents a CRUD component with state machine support
type CRUDComponent struct {
	State            CRUDState
	Table            core.ComponentInterface
	Form             core.ComponentInterface
	Data             interface{}              // 存储当前数据
	EventCache       map[string]interface{}   // 事件缓存
	DataAPI          map[string]interface{}   // 数据API配置
	Actions          map[string]*core.Action  // 自定义操作
	props            CRUDProps
	EventBus         *core.EventBus
	id               string
	unsubscribeFuncs []func()
	bindings         []core.ComponentBinding
	stateHelper      *CRUDStateHelper
}

// getDefaultCRUDBindings 获取默认的CRUD快捷键绑定
func getDefaultCRUDBindings() []core.ComponentBinding {
	return []core.ComponentBinding{
		{
			Key:         "enter",
			Event:       core.EventRowSelected,
			Description: "选择当前行进行编辑",
			Enabled:     true,
		},
		{
			Key:         "ctrl+n",
			Event:       core.EventNewItemRequested,
			Description: "新建记录",
			Enabled:     true,
		},
		{
			Key:         "ctrl+d",
			Event:       core.EventItemDeleted,
			Description: "删除当前记录",
			Enabled:     true,
		},
		{
			Key:         "ctrl+s",
			Event:       core.EventFormSubmit,
			Description: "保存/提交表单",
			Enabled:     true,
		},
		{
			Key:         "esc",
			Event:       core.EventFormCancel,
			Description: "取消操作并返回列表",
			Enabled:     true,
		},
	}
}

// NewCRUDComponent creates a new CRUD component with the given ID
func NewCRUDComponent(config core.RenderConfig, id string) *CRUDComponent {
	component := &CRUDComponent{
		State:            StateList,
		id:               id,
		EventBus:         core.NewEventBus(),
		unsubscribeFuncs: []func(){},
		bindings:         getDefaultCRUDBindings(), // 设置默认绑定
	}

	component.stateHelper = &CRUDStateHelper{
		StateHelper: component,
		ComponentID: id,
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

// LoadData 加载数据
func (c *CRUDComponent) LoadData() tea.Cmd {
	// 根据配置从 API 或其他数据源加载数据
	// 返回 tea.Cmd 来异步获取数据
	if api, exists := c.DataAPI["load"]; exists {
		// 执行加载操作
		return c.executeDataAPI(api, "load")
	}
	
	// 默认：使用配置的数据
	return nil
}

// SaveData 保存数据
func (c *CRUDComponent) SaveData(data interface{}) tea.Cmd {
	if api, exists := c.DataAPI["save"]; exists {
		return c.executeDataAPI(api, "save")
	}
	return nil
}

// DeleteData 删除数据
func (c *CRUDComponent) DeleteData(id interface{}) tea.Cmd {
	if api, exists := c.DataAPI["delete"]; exists {
		return c.executeDataAPI(api, "delete")
	}
	return nil
}

// executeDataAPI 执行数据API操作
func (c *CRUDComponent) executeDataAPI(api interface{}, operation string) tea.Cmd {
	// 实现 API 调用逻辑
	return func() tea.Msg {
		// 模拟 API 调用
		return core.ActionMsg{
			ID:     c.id,
			Action: fmt.Sprintf("DATA_%s_COMPLETED", strings.ToUpper(operation)),
			Data:   map[string]interface{}{"result": "success", "operation": operation},
		}
	}
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

func (c *CRUDComponent) GetState() CRUDState {
	return c.State
}



// UpdateMsg routes messages based on internal state
func (c *CRUDComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		c,                   // 实现了 InteractiveBehavior 接口的组件
		msg,                 // 接收的消息
		c.getBindings,       // 获取按键绑定的函数
		c.handleBinding,     // 处理按键绑定的函数
		c.delegateToBubbles, // 委托给原 bubbles 组件的函数
	)

	return c, cmd, response
}

// 实现 InteractiveBehavior 接口的方法
func (c *CRUDComponent) getBindings() []core.ComponentBinding {
	return c.bindings
}

func (c *CRUDComponent) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	cmd, response, handled := core.HandleBinding(c, keyMsg, binding)
	return cmd, response, handled
}

func (c *CRUDComponent) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// 委托给原始CRUD组件的UpdateMsg方法
	// 但先处理定向消息和动作消息
	switch msg := msg.(type) {
	case core.TargetedMsg:
		if msg.TargetID != c.id {
			return nil
		}
		// 递归调用，但需要防止无限递归
		// 直接处理内部消息
		return c.delegateToBubbles(msg.InnerMsg)

	case core.ActionMsg:
		switch msg.Action {
		case core.EventRowSelected:
			if c.State == StateList {
				c.State = StateEditing
				if data, ok := msg.Data.(map[string]interface{}); ok && c.Form != nil {
					log.Trace("CRUD %s: Loading row data into form from row index: %v", c.id, data["rowIndex"])
				}
				return core.PublishEvent(c.id, core.EventDataLoaded, map[string]interface{}{
					"transition": "StateList_to_StateEditing",
				})
			}

		case core.EventNewItemRequested:
			if c.State == StateList {
				c.State = StateCreating
				log.Trace("CRUD %s: Clearing form for new item", c.id)
				return core.PublishEvent(c.id, core.EventNewItemRequested, map[string]interface{}{
					"transition": "StateList_to_StateCreating",
				})
			}

		case core.EventItemDeleted:
			if c.State == StateList {
				log.Trace("CRUD %s: Item deleted, refreshing table", c.id)
				return core.PublishEvent(c.id, core.EventItemDeleted, map[string]interface{}{
					"state": "deleted",
				})
			}

		case core.EventFormSubmitSuccess:
			if c.State == StateEditing || c.State == StateCreating {
				c.State = StateList
				log.Trace("CRUD %s: Form submitted successfully, returning to list", c.id)
				return core.PublishEvent(c.id, core.EventFormSubmitSuccess, map[string]interface{}{
					"transition": "StateEditing_Creating_to_StateList",
				})
			}

		case core.EventFormCancel:
			if c.State == StateEditing || c.State == StateCreating {
				c.State = StateList
				log.Trace("CRUD %s: Form cancelled, returning to list", c.id)
				return core.PublishEvent(c.id, core.EventFormCancel, map[string]interface{}{
					"transition": "StateEditing_Creating_to_StateList",
				})
			}

		// 处理数据操作完成事件
		case "DATA_LOAD_COMPLETED":
			log.Trace("CRUD %s: Data load completed", c.id)
			return core.PublishEvent(c.id, core.EventDataLoaded, msg.Data)
			
		case "DATA_SAVE_COMPLETED":
			log.Trace("CRUD %s: Data save completed", c.id)
			c.State = StateList
			return core.PublishEvent(c.id, core.EventFormSubmitSuccess, msg.Data)
			
		case "DATA_DELETE_COMPLETED":
			log.Trace("CRUD %s: Data delete completed", c.id)
			return core.PublishEvent(c.id, core.EventItemDeleted, msg.Data)

		// 处理表单提交事件
		case core.EventFormSubmit:
			if (c.State == StateEditing || c.State == StateCreating) && c.Form != nil {
				// 获取表单数据并保存
				formData := c.getFormData()
				return c.SaveData(formData)
			}
		}
	}

	// 处理其他类型的事件
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理键盘消息
		matched, binding, handled := core.CheckComponentBindings(msg, c.bindings, c.id)
		if matched && handled {
			cmd, response, _ := c.handleBinding(msg, *binding)
			if response == core.Handled {
				return cmd
			}
		}
	}

	// 根据当前状态委托给相应的子组件
	switch c.State {
	case StateEditing, StateCreating:
		if c.Form != nil {
			_, cmd, _ = c.Form.UpdateMsg(msg)
		}

	case StateList, StateFiltering, StateDeleting:
		if c.Table != nil {
			_, cmd, _ = c.Table.UpdateMsg(msg)
		}
	}

	return cmd
}

// 实现 StateCapturable 接口
func (c *CRUDComponent) CaptureState() map[string]interface{} {
	return c.stateHelper.CaptureState()
}

func (c *CRUDComponent) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return c.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (c *CRUDComponent) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyEnter:
		// 在列表状态下按回车键，相当于选择当前行进行编辑
		if c.State == StateList && c.Table != nil {
			// 发布行选择事件，触发编辑模式
			return core.PublishEvent(c.id, core.EventRowSelected, map[string]interface{}{
				"tableID":  c.id,
				"rowIndex": -1, // 使用实际的索引位置，这里由Table组件提供
			}), core.Handled, true
		}
		// 在编辑/创建状态下按回车键，提交表单
		if (c.State == StateEditing || c.State == StateCreating) && c.Form != nil {
			return core.PublishEvent(c.id, core.EventFormSubmit, map[string]interface{}{
				"state": c.State,
			}), core.Handled, true
		}
		return nil, core.Handled, true

	case tea.KeyCtrlN:
		// Ctrl+N 新建记录
		if c.State == StateList {
			return core.PublishEvent(c.id, core.EventNewItemRequested, map[string]interface{}{
				"state": "creating",
			}), core.Handled, true
		}
		return nil, core.Handled, true

	case tea.KeyCtrlD:
		// Ctrl+D 删除记录
		if c.State == StateList && c.Table != nil {
			return core.PublishEvent(c.id, core.EventItemDeleted, map[string]interface{}{
				"state": "deleting",
			}), core.Handled, true
		}
		return nil, core.Handled, true

	case tea.KeyCtrlS:
		// Ctrl+S 保存/提交
		if (c.State == StateEditing || c.State == StateCreating) && c.Form != nil {
			return core.PublishEvent(c.id, core.EventFormSubmit, map[string]interface{}{
				"state": c.State,
			}), core.Handled, true
		}
		return nil, core.Handled, true

	case tea.KeyEsc:
		// ESC 键取消当前操作，返回列表状态
		if c.State == StateEditing || c.State == StateCreating || c.State == StateDeleting {
			c.State = StateList
			return core.PublishEvent(c.id, core.EventFormCancel, map[string]interface{}{
				"transition": "cancel_to_list",
			}), core.Handled, true
		}
		// 对于列表状态，让框架处理焦点切换
		return nil, core.Ignored, false
	}

	// 其他键：不特殊处理，让默认行为继续
	return nil, core.Ignored, false
}

// GetModel returns the underlying model
func (c *CRUDComponent) GetModel() interface{} {
	return c
}

// PublishEvent creates and returns a command to publish an event
func (c *CRUDComponent) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (c *CRUDComponent) ExecuteAction(action *core.Action) tea.Cmd {
	// 创建执行动作的命令
	return func() tea.Msg {
		return core.ActionMsg{
			ID:     c.id,
			Action: "EXECUTE_ACTION", // 使用通用动作名
			Data: map[string]interface{}{
				"action": action,
			},
		}
	}
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

	// 解析 CRUD 特定配置
	props := ParseCRUDProps(dataMap)
	
	// 应用绑定配置
	if len(props.Bindings) > 0 {
		c.bindings = props.Bindings
	}
	
	// 应用数据API配置
	if props.DataAPI != nil {
		c.DataAPI = props.DataAPI
	}
	
	// 应用动作配置
	if props.Actions != nil {
		c.Actions = props.Actions
	}

	// 原有的表格数据处理逻辑...
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
		Bindings:   []core.ComponentBinding{}, // 可以从配置中继承绑定
	}

	c.Table = NewTableComponentWrapper(tableProps, c.id)

	return nil
}

func (c *CRUDComponent) Render(config core.RenderConfig) (string, error) {
	return c.View(), nil
}

// getFormData 获取表单当前数据
func (c *CRUDComponent) getFormData() interface{} {
	if c.Form != nil {
		if valuer, ok := c.Form.(interface{ GetValue() string }); ok {
			return valuer.GetValue()
		}
	}
	return nil
}

// setStateTransition 设置状态转换
func (c *CRUDComponent) setStateTransition(from, to CRUDState) {
	c.State = to
	log.Trace("CRUD %s: State transition %d -> %d", c.id, from, to)
}
