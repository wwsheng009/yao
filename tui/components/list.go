package components

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// ListItem represents a single list item
type ListItem struct {
	// Title is the display text for the list item
	Title string `json:"title"`

	// Description is the optional description for the list item
	Description string `json:"description"`

	// Value is the value associated with the list item (for selection)
	Value interface{} `json:"value"`

	// Disabled indicates if the list item is disabled
	Disabled bool `json:"disabled"`

	// Selected indicates if the list item is selected
	Selected bool `json:"selected"`
}

// ListItemInterface implementation - allows ListItem to be used as list.Item
func (i ListItem) FilterValue() string {
	return i.Title
}

// ListProps defines the properties for the List component
type ListProps struct {
	// Items contains the list items
	Items []ListItem `json:"items"`

	// Title is the optional title for the list
	Title string `json:"title"`

	// Height specifies the list height (0 for auto)
	Height int `json:"height"`

	// Width specifies the list width (0 for auto)
	Width int `json:"width"`

	// Focused determines if the list is focused (for selection)
	Focused bool `json:"focused"`

	// ShowTitle shows/hides the list title
	ShowTitle bool `json:"showTitle"`

	// ShowStatusBar shows/hides the status bar
	ShowStatusBar bool `json:"showStatusBar"`

	// ShowFilter shows/hides the filter input
	ShowFilter bool `json:"showFilter"`

	// FilteringEnabled enables/disables filtering
	FilteringEnabled bool `json:"filteringEnabled"`

	// Color specifies the text color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`
	
	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// ListModel wraps the list.Model to handle TUI integration
type ListModel struct {
	list.Model
	props ListProps
	id    string // Unique identifier for this instance
}

// RenderList renders a list component
func RenderList(props ListProps, width int) string {
	// Convert items to list.Item interface
	items := make([]list.Item, len(props.Items))
	for i, item := range props.Items {
		items[i] = item
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	// Set title
	if props.Title != "" && props.ShowTitle {
		l.Title = props.Title
	}

	// Set dimensions
	if props.Width > 0 {
		l.SetWidth(props.Width)
	} else if width > 0 {
		l.SetWidth(width)
	}

	if props.Height > 0 {
		l.SetHeight(props.Height)
	}

	// Configure list options
	l.SetShowTitle(props.ShowTitle)
	l.SetShowStatusBar(props.ShowStatusBar)
	l.SetShowFilter(props.ShowFilter)
	l.SetFilteringEnabled(props.FilteringEnabled)

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to list
	l.Styles.Title = l.Styles.Title.Inherit(style)
	l.Styles.NoItems = l.Styles.NoItems.Inherit(style)

	return l.View()
}

// ParseListProps converts a generic props map to ListProps using JSON unmarshaling
func ParseListProps(props map[string]interface{}) ListProps {
	// Set defaults
	lp := ListProps{
		ShowTitle:        true,
		ShowStatusBar:    true,
		ShowFilter:       true,
		FilteringEnabled: true,
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &lp)
	}

	return lp
}

// NewListModel creates a new ListModel from ListProps
func NewListModel(props ListProps, id string) ListModel {
	// Convert items to list.Item interface
	items := make([]list.Item, len(props.Items))
	for i, item := range props.Items {
		items[i] = item
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	// Set title
	if props.Title != "" && props.ShowTitle {
		l.Title = props.Title
	}

	// Set dimensions
	if props.Width > 0 {
		l.SetWidth(props.Width)
	}

	if props.Height > 0 {
		l.SetHeight(props.Height)
	}

	// Configure list options
	l.SetShowTitle(props.ShowTitle)
	l.SetShowStatusBar(props.ShowStatusBar)
	l.SetShowFilter(props.ShowFilter)
	l.SetFilteringEnabled(props.FilteringEnabled)

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to list
	l.Styles.Title = l.Styles.Title.Inherit(style)
	l.Styles.NoItems = l.Styles.NoItems.Inherit(style)

	return ListModel{
		Model: l,
		props: props,
		id:    id,
	}
}

// HandleListUpdate handles updates for list components
func HandleListUpdate(msg tea.Msg, listModel *ListModel) (ListModel, tea.Cmd) {
	if listModel == nil {
		return ListModel{}, nil
	}

	var cmd tea.Cmd
	listModel.Model, cmd = listModel.Model.Update(msg)
	return *listModel, cmd
}

// Init initializes the list model
func (m *ListModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the list
func (m *ListModel) View() string {
	return m.Model.View()
}

// GetID returns the unique identifier for this component instance
func (m *ListModel) GetID() string {
	return m.id
}

// SetFocus sets or removes focus from list component
func (m *ListModel) SetFocus(focus bool) {
	m.props.Focused = focus
}

// ListComponentWrapper wraps ListModel to implement ComponentInterface properly
type ListComponentWrapper struct {
	model *ListModel
	bindings []core.ComponentBinding
	stateHelper *core.ListStateHelper
}

// listIndexerAdapter adapts ListModel to satisfy interface{Index() int}
type listIndexerAdapter struct {
	*ListModel
}

func (a *listIndexerAdapter) Index() int {
	return a.Model.Index()
}

// listSelectorAdapter adapts ListModel to satisfy interface{SelectedItem() interface{}}
type listSelectorAdapter struct {
	*ListModel
}

func (a *listSelectorAdapter) SelectedItem() interface{} {
	return a.Model.SelectedItem()
}

// NewListComponentWrapper creates a wrapper that implements ComponentInterface
func NewListComponentWrapper(listModel *ListModel) *ListComponentWrapper {
	wrapper := &ListComponentWrapper{
		model: listModel,
		bindings: listModel.props.Bindings,
	}

	// 创建适配器来满足接口要求
	indexerAdapter := &listIndexerAdapter{listModel}
	selectorAdapter := &listSelectorAdapter{listModel}
	wrapper.stateHelper = &core.ListStateHelper{
		Indexer:     indexerAdapter,
		Selector:    selectorAdapter,
		Focused:     listModel.props.Focused,
		ComponentID: listModel.id,
	}
	return wrapper
}

func (w *ListComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *ListComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,                           // 实现了 InteractiveBehavior 接口的组件
		msg,                         // 接收的消息
		w.getBindings,              // 获取按键绑定的函数
		w.handleBinding,            // 处理按键绑定的函数
		w.delegateToBubbles,        // 委托给原 bubbles 组件的函数
	)

	return w, cmd, response
}

// 实现 InteractiveBehavior 接口的方法

func (w *ListComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *ListComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// 创建适配器来实现 ComponentWrapper 接口
	wrapper := &listComponentWrapperAdapter{w}
	cmd, response, handled := core.HandleBinding(wrapper, keyMsg, binding)
	return cmd, response, handled
}

func (w *ListComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *ListComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *ListComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HasFocus 方法
func (w *ListComponentWrapper) HasFocus() bool {
	return w.model.props.Focused
}

// 实现 HandleSpecialKey 方法
func (w *ListComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyEnter:
		if selectedItem := w.model.Model.SelectedItem(); selectedItem != nil {
			item := selectedItem.(ListItem)
			// 发布菜单项选择事件
			cmd := core.PublishEvent(w.model.id, core.EventMenuItemSelected, map[string]interface{}{
				"item":  item,
				"index": w.model.Model.Index(),
				"title": item.Title,
				"value": item.Value,
			})
			return cmd, core.Handled, true
		}
	case tea.KeyTab:
		// 让 Tab 键冒泡以处理组件导航
		return nil, core.Ignored, true
	}
	
	// 其他按键不由这个函数处理
	return nil, core.Ignored, false
}

// listComponentWrapperAdapter 适配器实现 ComponentWrapper 接口
type listComponentWrapperAdapter struct {
	*ListComponentWrapper
}

func (a *listComponentWrapperAdapter) GetModel() interface{} {
	return a.ListComponentWrapper.model
}

func (a *listComponentWrapperAdapter) GetID() string {
	return a.ListComponentWrapper.model.id
}

func (a *listComponentWrapperAdapter) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

func (a *listComponentWrapperAdapter) ExecuteAction(action *core.Action) tea.Cmd {
	// 对于列表组件，返回一个创建 ExecuteActionMsg 的命令
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  a.ListComponentWrapper.model.id,
			Timestamp: time.Now(),
		}
	}
}

func (w *ListComponentWrapper) View() string {
	return w.model.View()
}

func (w *ListComponentWrapper) GetID() string {
	return w.model.id
}

func (w *ListComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
}

// GetSelectedItem returns the currently selected item
func (w *ListComponentWrapper) GetSelectedItem() ListItem {
	if selected := w.model.SelectedItem(); selected != nil {
		return selected.(ListItem)
	}
	return ListItem{}
}

// SetItems sets the list items
func (w *ListComponentWrapper) SetItems(items []ListItem) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	w.model.SetItems(listItems)
}

func (m *ListModel) GetComponentType() string {
	return "list"
}

func (w *ListComponentWrapper) GetComponentType() string {
	return "list"
}

func (m *ListModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ListModel: invalid data type")
	}

	// Parse list properties
	props := ParseListProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (m *ListModel) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ListModel: invalid data type for UpdateRenderConfig")
	}

	// Parse list properties
	props := ParseListProps(propsMap)

	// Update component properties
	m.props = props

	// Update list items if provided
	if props.Items != nil {
		listItems := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			listItems[i] = item
		}
		m.Model.SetItems(listItems)
	}

	return nil
}

// Cleanup 清理资源
func (m *ListModel) Cleanup() {
	// ListModel 通常不需要特殊清理操作
}

func (w *ListComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *ListComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ListComponentWrapper: invalid data type")
	}

	// Parse list properties
	props := ParseListProps(propsMap)

	// Update component properties
	w.model.props = props

	// Update list items if provided
	if props.Items != nil {
		listItems := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			listItems[i] = item
		}
		w.model.Model.SetItems(listItems)
	}

	return nil
}

// Cleanup cleans up resources used by the list component
func (w *ListComponentWrapper) Cleanup() {
	// List components typically don't need cleanup
	// This is a no-op for list components
}

// GetStateChanges returns the state changes from this component
func (w *ListComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	selectedItem := w.model.Model.SelectedItem()
	if selectedItem == nil {
		return map[string]interface{}{
			w.GetID() + "_selected_index": -1,
			w.GetID() + "_selected_item":  nil,
		}, false
	}

	return map[string]interface{}{
		w.GetID() + "_selected_index": w.model.Model.Index(),
		w.GetID() + "_selected_item":  selectedItem,
	}, true
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *ListComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}
