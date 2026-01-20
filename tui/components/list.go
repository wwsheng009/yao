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

func (m *ListModel) GetFocus() bool {
	return m.props.Focused
}

// ListComponentWrapper wraps the native list.Model to implement ComponentInterface properly
type ListComponentWrapper struct {
	model       list.Model // Directly use the native model
	props       ListProps  // Component properties
	id          string     // Component ID
	bindings    []core.ComponentBinding
	stateHelper *core.ListStateHelper
}

// NewListComponentWrapper creates a wrapper that implements ComponentInterface
// This is the unified entry point that accepts props and id, creating the model internally
func NewListComponentWrapper(props ListProps, id string) *ListComponentWrapper {
	// Convert items to list.Item interface
	items := make([]list.Item, len(props.Items))
	for i, item := range props.Items {
		items[i] = item
	}

	// Create the native list model
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

	// Create wrapper that directly implements all interfaces
	wrapper := &ListComponentWrapper{
		model:    l,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	// stateHelper uses wrapper itself as the implementation
	wrapper.stateHelper = &core.ListStateHelper{
		Indexer:     wrapper, // wrapper implements Index() method
		Selector:    wrapper, // wrapper implements SelectedItem() method
		Focused:     props.Focused,
		ComponentID: id,
	}

	return wrapper
}

func (w *ListComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *ListComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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

func (w *ListComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *ListComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// ListComponentWrapper 已经实现了 ComponentWrapper 接口，可以直接传递
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *ListComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *ListComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *ListComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *ListComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyEnter:
		if selectedItem := w.model.SelectedItem(); selectedItem != nil {
			item := selectedItem.(ListItem)
			// 发布菜单项选择事件
			cmd := core.PublishEvent(w.id, core.EventMenuItemSelected, map[string]interface{}{
				"item":  item,
				"index": w.model.Index(),
				"title": item.Title,
				"value": item.Value,
			})
			return cmd, core.Handled, true
		}
	}

	// ESC 和 Tab 现在由框架层统一处理，这里不处理
	// 如果有其他特殊的键处理需求，可以在这里添加
	return nil, core.Ignored, false
}

// Index returns the current cursor position
func (w *ListComponentWrapper) Index() int {
	return w.model.Index()
}

// SelectedItem returns the currently selected item
func (w *ListComponentWrapper) SelectedItem() interface{} {
	return w.model.SelectedItem()
}

func (w *ListComponentWrapper) GetModel() interface{} {
	return w.model
}

func (w *ListComponentWrapper) GetID() string {
	return w.id
}

func (w *ListComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

func (w *ListComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// 对于列表组件，返回一个创建 ExecuteActionMsg 的命令
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}

func (w *ListComponentWrapper) View() string {
	return w.model.View()
}

func (w *ListComponentWrapper) SetFocus(focus bool) {
	// Update the focused property
	w.props.Focused = focus
}

func (w *ListComponentWrapper) GetFocus() bool {
	return w.props.Focused
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
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ListComponentWrapper: invalid data type")
	}

	// Parse list properties
	props := ParseListProps(propsMap)

	// Update component properties
	w.props = props

	// Update list items if provided
	if props.Items != nil {
		listItems := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			listItems[i] = item
		}
		w.model.SetItems(listItems)
	}

	// Return the view
	return w.View(), nil
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
	w.props = props

	// Update list items if provided
	if props.Items != nil {
		listItems := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			listItems[i] = item
		}
		w.model.SetItems(listItems)
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
	selectedItem := w.model.SelectedItem()
	if selectedItem == nil {
		return map[string]interface{}{
			w.GetID() + "_selected_index": -1,
			w.GetID() + "_selected_item":  nil,
		}, false
	}

	return map[string]interface{}{
		w.GetID() + "_selected_index": w.model.Index(),
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
