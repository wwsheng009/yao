package components

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// MenuItem represents a single menu item
type MenuItem struct {
	// Title is the display text for the menu item
	Title string `json:"title"`

	// Description is the optional description for the menu item
	Description string `json:"description"`

	// Value is the value associated with the menu item (for selection)
	Value interface{} `json:"value"`

	// Disabled indicates if the menu item is disabled
	Disabled bool `json:"disabled"`

	// Action is the action to execute when the item is selected
	Action map[string]interface{} `json:"action,omitempty"`

	// Children contains submenu items (optional)
	Children []MenuItem `json:"children,omitempty"`

	// Selected indicates if the menu item is selected
	Selected bool `json:"selected"`
}

// Item implements the list.Item interface
func (i MenuItem) FilterValue() string {
	return i.Title
}

// HasSubmenu checks if the menu item has children/submenu
func (i MenuItem) HasSubmenu() bool {
	return len(i.Children) > 0
}

// MenuItemInterface implementation - allows MenuItem to be used as core.MenuItemInterface
func (i MenuItem) GetTitle() string {
	return i.Title
}

func (i MenuItem) GetDescription() string {
	return i.Description
}

func (i MenuItem) GetValue() interface{} {
	return i.Value
}

func (i MenuItem) GetAction() map[string]interface{} {
	return i.Action
}

func (i MenuItem) IsDisabled() bool {
	return i.Disabled
}

func (i MenuItem) IsSelected() bool {
	return i.Selected
}

func (i MenuItem) HasChildren() bool {
	return len(i.Children) > 0
}

// MenuProps defines the properties for the Menu component
type MenuProps struct {
	// Items contains the menu items
	Items []MenuItem `json:"items"`

	// Title is the optional title for the menu
	Title string `json:"title"`

	// Height specifies the menu height (0 for auto)
	Height int `json:"height"`

	// Width specifies the menu width (0 for auto)
	Width int `json:"width"`

	// Focused determines if the menu is focused (for selection)
	Focused bool `json:"focused"`

	// ShowStatusBar shows/hides the status bar
	ShowStatusBar bool `json:"showStatusBar"`

	// ShowFilter shows/hides the filter input
	ShowFilter bool `json:"showFilter"`

	// ActiveItemStyle is the style for active/selected items
	ActiveItemStyle lipglossStyleWrapper `json:"activeItemStyle"`

	// InactiveItemStyle is the style for inactive items
	InactiveItemStyle lipglossStyleWrapper `json:"inactiveItemStyle"`

	// SelectedItemStyle is the style for the currently selected item
	SelectedItemStyle lipglossStyleWrapper `json:"selectedItemStyle"`

	// DisabledItemStyle is the style for disabled items
	DisabledItemStyle lipglossStyleWrapper `json:"disabledItemStyle"`

	// TitleStyle is the style for the title
	TitleStyle lipglossStyleWrapper `json:"titleStyle"`

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// MenuModel wraps the list.Model to handle TUI integration
type MenuModel struct {
	list.Model
	props MenuProps
}

// itemDelegate implements the list.ItemDelegate interface
type itemDelegate struct {
	props MenuProps
}

// Render renders a single item in the list
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(MenuItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	if i.Disabled {
		style = d.props.DisabledItemStyle.GetStyle()
	} else if index == m.Index() {
		style = d.props.SelectedItemStyle.GetStyle()
	} else {
		style = d.props.InactiveItemStyle.GetStyle()
	}

	title := i.Title
	// Add submenu indicator if item has children
	if i.HasSubmenu() {
		title += " ▶" // Add submenu indicator
	}
	// Note: Description is intentionally not displayed to keep the menu compact

	fmt.Fprint(w, style.Render(title))
}

// Height returns the height of the item
func (d itemDelegate) Height() int {
	return 1 // Minimum height for compact menu
}

// Spacing returns the spacing between items
func (d itemDelegate) Spacing() int {
	return 0 // Reduce spacing to minimize distance between menu items
}

// Update handles update messages for the item
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	// Delegate all updates to the list model (maximum delegation)
	return nil
}

// Width returns the width of the item
func (d itemDelegate) Width() int {
	return 0
}

// RenderMenu renders a menu component
func RenderMenu(props MenuProps, width int) string {
	log.Trace("Menu Render: Rendering menu with %d items, title: %s", len(props.Items), props.Title)

	// Convert MenuItem slice to list.Item slice
	items := make([]list.Item, len(props.Items))
	for i, item := range props.Items {
		items[i] = item
	}

	// Since list.New requires a delegate but we're just rendering,
	// we'll create a simple static representation of the menu
	var result strings.Builder

	// Add title if provided
	if props.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("57")).
			Padding(0, 0).
			MarginBottom(0).
			Align(lipgloss.Center)
		result.WriteString(titleStyle.Render(props.Title))
		result.WriteString("\n")
	}

	// Render each menu item
	for _, item := range props.Items {
		var style lipgloss.Style
		if item.Disabled {
			style = props.DisabledItemStyle.GetStyle()
		} else {
			style = props.InactiveItemStyle.GetStyle()
		}

		title := item.Title
		// Add submenu indicator if item has children
		if item.HasSubmenu() {
			title += " ▶" // Add submenu indicator
		}
		// Note: Description is intentionally not displayed to keep the menu compact

		result.WriteString(style.Render(title))
		result.WriteString("\n")
	}

	log.Trace("Menu Render: Completed rendering menu, total length: %d", result.Len())
	return result.String()
}

// ParseMenuProps converts a generic props map to MenuProps using JSON unmarshaling
func ParseMenuProps(props map[string]interface{}) MenuProps {
	log.Trace("Menu ParseProps: Starting to parse menu props with %d keys", len(props))

	// Set defaults
	mp := MenuProps{
		ShowStatusBar: true,
		ShowFilter:    false,
		Focused:       true,
	}

	// Unmarshal properties first to get Items
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &mp)
	}

	// Handle Items separately as it needs special processing
	itemsValue := props["items"]

	// Helper function to process items array
	processItemsArray := func(itemsArray []interface{}) {
		log.Trace("Menu ParseProps: Processing items array with %d items", len(itemsArray))
		mp.Items = make([]MenuItem, 0, len(itemsArray))
		for _, itemIntf := range itemsArray {
			// Convert map to MenuItem
			if itemMap, ok := itemIntf.(map[string]interface{}); ok {
				itemBytes, _ := json.Marshal(itemMap)
				var menuItem MenuItem
				if err := json.Unmarshal(itemBytes, &menuItem); err == nil {
					mp.Items = append(mp.Items, menuItem)
				}
			}
		}
	}

	// Check if items is empty or nil
	if itemsValue == nil {
		log.Trace("Menu ParseProps: Items value is nil")
		return mp
	}

	// Case 1: items is already an array ([]interface{})
	if itemsArray, ok := itemsValue.([]interface{}); ok {
		log.Trace("Menu ParseProps: Items is array with %d items", len(itemsArray))
		processItemsArray(itemsArray)
		return mp
	}

	// Case 2: items is a map ({"menu": [...]} type)
	// Extract the first array value from the map
	if itemsMap, ok := itemsValue.(map[string]interface{}); ok {
		log.Trace("Menu ParseProps: Items is map, extracting arrays")
		for _, v := range itemsMap {
			if itemsArray, ok := v.([]interface{}); ok {
				processItemsArray(itemsArray)
				return mp
			}
		}
	}

	// Case 3: items is a string (template variable like "{{items}}" that was converted to string)
	if itemsStr, ok := itemsValue.(string); ok {
		log.Trace("Menu ParseProps: Items is string, attempting to parse as JSON")
		// Try to unmarshal as JSON array first
		var itemsArray []interface{}
		if err := json.Unmarshal([]byte(itemsStr), &itemsArray); err == nil {
			processItemsArray(itemsArray)
			return mp
		}
	}

	log.Trace("Menu ParseProps: Completed parsing, total items: %d", len(mp.Items))
	return mp
}

// MenuInteractiveModel wraps the list.Model to handle interactive menus
type MenuInteractiveModel struct {
	list.Model
	props MenuProps
	// CurrentLevel tracks the current menu level for nested menus
	CurrentLevel int
	// Path keeps track of the navigation path
	Path []string
	// ID is the unique identifier for this menu instance
	ID string
	// focused indicates if the menu has focus
	focused bool
}

// NewMenuInteractiveModel creates a new interactive menu model
func NewMenuInteractiveModel(props MenuProps) MenuInteractiveModel {
	log.Trace("Menu InteractiveModel: Creating new interactive menu model with %d items, title: %s", len(props.Items), props.Title)

	// Set default styles if not provided
	if props.SelectedItemStyle.Style == nil {
		defaultSelectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Bold(true).
			Padding(0, 0) // Reduced padding to minimize space
		props.SelectedItemStyle = lipglossStyleWrapper{Style: &defaultSelectedStyle}
	}
	if props.InactiveItemStyle.Style == nil {
		defaultInactiveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 0) // Reduced padding to minimize space
		props.InactiveItemStyle = lipglossStyleWrapper{Style: &defaultInactiveStyle}
	}
	if props.ActiveItemStyle.Style == nil {
		defaultActiveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Padding(0, 0) // Reduced padding to minimize space
		props.ActiveItemStyle = lipglossStyleWrapper{Style: &defaultActiveStyle}
	}
	if props.DisabledItemStyle.Style == nil {
		defaultDisabledStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		props.DisabledItemStyle = lipglossStyleWrapper{Style: &defaultDisabledStyle}
	}
	if props.TitleStyle.Style == nil {
		defaultTitleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("57")).
			Padding(0, 0).   // Reduced padding to minimize space
			MarginBottom(0). // Reduced margin to minimize space
			Align(lipgloss.Center)
		props.TitleStyle = lipglossStyleWrapper{Style: &defaultTitleStyle}
	}

	// Convert MenuItem slice to list.Item slice
	items := make([]list.Item, len(props.Items))
	for i, item := range props.Items {
		items[i] = item
	}

	// Create the list model with custom delegate
	delegate := itemDelegate{props: props}
	l := list.New(items, delegate, 0, 0)

	// Set title if provided
	if props.Title != "" {
		l.Title = props.Title
	} else {
		l.Title = ""
		l.DisableQuitKeybindings()
	}

	// Disable status bar to prevent showing item count
	l.SetShowStatusBar(false)
	// Disable filtering to keep the UI clean
	l.SetShowFilter(false)

	// Set custom delegate
	l.SetDelegate(delegate)

	// Set size if specified
	if props.Width > 0 {
		l.SetWidth(props.Width)
		log.Trace("Menu InteractiveModel: Set width to %d", props.Width)
	} else {
		// Calculate width based on the longest item if not explicitly set
		maxWidth := 0
		for _, item := range items {
			if menuItem, ok := item.(MenuItem); ok {
				itemWidth := lipgloss.Width(menuItem.Title)
				if itemWidth > maxWidth {
					maxWidth = itemWidth
				}
				// Account for submenu indicator if present
				if menuItem.HasSubmenu() {
					maxWidth += 2 // Account for " ▶" indicator
				}
			}
		}
		if maxWidth > 0 {
			// Add some padding to the calculated width
			l.SetWidth(maxWidth + 4)
			log.Trace("Menu InteractiveModel: Set calculated width to %d", maxWidth+4)
		}
	}

	if props.Height > 0 {
		l.SetHeight(props.Height)
		log.Trace("Menu InteractiveModel: Set height to %d", props.Height)
	}

	log.Trace("Menu InteractiveModel: Created model with %d items, current level: %d", len(items), 0)

	return MenuInteractiveModel{
		Model:        l,
		props:        props,
		CurrentLevel: 0,          // Start at top level
		Path:         []string{}, // Empty path initially
		ID:           "",         // Will be set by the parent
		focused:      false,
	}
}

// HandleMenuUpdate handles updates for menu components (minimal wrapper, maximum delegation)
// This function only delegates all messages to list.Model, all other logic is in the wrapper
func HandleMenuUpdate(msg tea.Msg, menuModel *MenuInteractiveModel) (MenuInteractiveModel, tea.Cmd) {
	if menuModel == nil {
		log.Trace("Menu Update: Received update with nil menu model")
		return MenuInteractiveModel{}, nil
	}

	log.Trace("Menu Update: Handling message type: %T, current level: %d, path: %v", msg, menuModel.CurrentLevel, menuModel.Path)

	// Delegate all messages to list.Model (maximum delegation)
	updatedListModel, listCmd := menuModel.Model.Update(msg)
	menuModel.Model = updatedListModel

	return *menuModel, listCmd
}

// EnterSubmenu navigates to submenu for the selected menu item
func (m *MenuInteractiveModel) EnterSubmenu() tea.Cmd {
	selectedItem := m.Model.SelectedItem()
	if selectedItem == nil {
		return nil
	}

	menuItem, ok := selectedItem.(MenuItem)
	if !ok || !menuItem.HasSubmenu() {
		return nil
	}

	log.Trace("Menu: Entering submenu for item: %s", menuItem.Title)
	m.Path = append(m.Path, menuItem.Title)
	m.CurrentLevel++

	// Load submenu items
	submenuItems := make([]list.Item, len(menuItem.Children))
	for i, child := range menuItem.Children {
		submenuItems[i] = child
	}
	m.SetItems(submenuItems)

	// Maintain the same width as the original menu
	if m.props.Width > 0 {
		m.SetWidth(m.props.Width)
	}
	m.Select(0)

	log.Trace("Menu: Submenu loaded with %d items, now at level %d", len(submenuItems), m.CurrentLevel)

	// Publish submenu entered event
	return core.PublishEvent(m.ID, core.EventMenuSubmenuEntered, map[string]interface{}{
		"item":        menuItem,
		"parentPath":  m.Path[:len(m.Path)-1],
		"currentPath": m.Path,
		"level":       m.CurrentLevel,
	})
}

// ExitSubmenu navigates back to parent menu
func (m *MenuInteractiveModel) ExitSubmenu() tea.Cmd {
	if m.CurrentLevel <= 0 || len(m.Path) == 0 {
		return nil
	}

	log.Trace("Menu: Exiting submenu from level %d", m.CurrentLevel)

	// Save current path before navigating back
	previousPath := make([]string, len(m.Path))
	copy(previousPath, m.Path)

	// Go back to parent menu
	m.CurrentLevel--
	m.Path = m.Path[:len(m.Path)-1]

	// Reload parent menu
	originalItems := make([]list.Item, len(m.props.Items))
	for i, item := range m.props.Items {
		originalItems[i] = item
	}
	m.SetItems(originalItems)

	// Maintain the same width as the original menu
	if m.props.Width > 0 {
		m.SetWidth(m.props.Width)
	} else {
		// Calculate width based on the longest item if not explicitly set
		maxWidth := 0
		for _, item := range originalItems {
			if menuItem, ok := item.(MenuItem); ok {
				itemWidth := lipgloss.Width(menuItem.Title)
				if itemWidth > maxWidth {
					maxWidth = itemWidth
				}
				// Account for submenu indicator if present
				if menuItem.HasSubmenu() {
					maxWidth += 2 // Account for " ▶" indicator
				}
			}
		}
		if maxWidth > 0 {
			m.SetWidth(maxWidth + 4)
		}
	}

	m.Select(0)
	log.Trace("Menu: Returned to parent menu at level %d", m.CurrentLevel)

	// Publish submenu exited event
	return core.PublishEvent(m.ID, core.EventMenuSubmenuExited, map[string]interface{}{
		"previousPath": previousPath,
		"currentPath":  m.Path,
		"level":        m.CurrentLevel,
	})
}

// View returns the rendered view of the menu
func (m *MenuInteractiveModel) View() string {
	log.Trace("Menu View: Rendering menu view, current level: %d, path: %v", m.CurrentLevel, m.Path)

	// Build the menu view manually to have full control over the appearance
	var result strings.Builder

	// Add title if provided
	if m.props.Title != "" {
		titleStyle := m.props.TitleStyle.GetStyle()
		result.WriteString(titleStyle.Render(m.props.Title))
		result.WriteString("\n")
	}

	// Render each item with appropriate styling
	allItems := m.Items()
	for i, item := range allItems {
		if menuItem, ok := item.(MenuItem); ok {
			var style lipgloss.Style
			if menuItem.Disabled {
				style = m.props.DisabledItemStyle.GetStyle()
			} else if i == m.Index() { // Current selection
				style = m.props.SelectedItemStyle.GetStyle()
			} else {
				style = m.props.InactiveItemStyle.GetStyle()
			}

			title := menuItem.Title
			// Add submenu indicator if item has children
			if menuItem.HasSubmenu() {
				title += " ▶" // Add submenu indicator
			}
			// Note: Description is intentionally not displayed to keep the menu compact

			// Add indentation based on current level to show hierarchy
			indent := ""
			for level := 0; level < m.CurrentLevel; level++ {
				indent += "  " // Two spaces per level
			}
			title = indent + title

			result.WriteString(style.Render(title))
			result.WriteString("\n")
		}
	}

	output := result.String()
	log.Trace("Menu View: Completed rendering, output length: %d", len(output))
	return output
}

// GetSelectedItem returns the currently selected menu item
func (m *MenuInteractiveModel) GetSelectedItem() (MenuItem, bool) {
	log.Trace("Menu GetSelectedItem: Attempting to get selected item")
	item := m.Model.SelectedItem()
	if item == nil {
		log.Trace("Menu GetSelectedItem: No selected item (nil)")
		return MenuItem{}, false
	}

	menuItem, ok := item.(MenuItem)
	if ok {
		log.Trace("Menu GetSelectedItem: Got selected item: %s (disabled: %t, hasChildren: %t)", menuItem.Title, menuItem.Disabled, menuItem.HasSubmenu())
	} else {
		log.Trace("Menu GetSelectedItem: Selected item is not a MenuItem type")
	}
	return menuItem, ok
}

// Index returns the current cursor position
func (m *MenuInteractiveModel) Index() int {
	return m.Model.Index()
}

// SelectedItem returns the currently selected item
func (m *MenuInteractiveModel) SelectedItem() interface{} {
	return m.Model.SelectedItem()
}

// GetSelected returns the currently selected item and whether anything is selected
func (m *MenuInteractiveModel) GetSelected() (interface{}, bool) {
	item := m.SelectedItem()
	return item, item != nil
}

// Focused returns whether the menu is focused
func (m *MenuInteractiveModel) Focused() bool {
	return m.focused
}

// Init initializes the menu model
func (m *MenuInteractiveModel) Init() tea.Cmd {
	return nil
}

// GetID returns the unique identifier for this component instance
func (m *MenuInteractiveModel) GetID() string {
	return m.ID
}

// GetComponentType returns the component type
func (m *MenuInteractiveModel) GetComponentType() string {
	return "menu"
}

func (m *MenuInteractiveModel) Render(config core.RenderConfig) (string, error) {
	// Render should be a pure function - it should not modify internal state
	// All state updates should happen in UpdateRenderConfig
	return m.View(), nil
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (m *MenuInteractiveModel) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
	}
}

// MenuStateHelper 菜单组件状态捕获助手
type MenuStateHelper struct {
	Indexer     interface{ Index() int }
	Selector    interface{ SelectedItem() interface{} }
	Focuser     interface{ Focused() bool }
	ComponentID string
}

func (h *MenuStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{
		"index":    h.Indexer.Index(),
		"selected": h.Selector.SelectedItem(),
		"focused":  h.Focuser.Focused(),
	}
}

func (h *MenuStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	var cmds []tea.Cmd

	// 检测索引变化
	if old["index"] != new["index"] {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventMenuItemSelected, map[string]interface{}{
			"oldIndex": old["index"],
			"newIndex": new["index"],
		}))
	}

	// 检测焦点变化
	if old["focused"] != new["focused"] {
		cmds = append(cmds, core.PublishEvent(h.ComponentID, core.EventFocusChanged, map[string]interface{}{
			"focused": new["focused"],
		}))
	}

	return cmds
}

// MenuComponentWrapper wraps MenuInteractiveModel to implement ComponentInterface properly
type MenuComponentWrapper struct {
	model       MenuInteractiveModel
	bindings    []core.ComponentBinding
	stateHelper *MenuStateHelper
}



// NewMenuComponentWrapper creates a wrapper that implements ComponentInterface
func NewMenuComponentWrapper(props MenuProps, id string) *MenuComponentWrapper {
	// Create menu model with props
	menuModel := NewMenuInteractiveModel(props)
	menuModel.ID = id

	wrapper := &MenuComponentWrapper{
		model:    menuModel,
		bindings: props.Bindings,
	}
	
	// MenuInteractiveModel 已经实现了所需接口，直接使用
	wrapper.stateHelper = &MenuStateHelper{
		Indexer:     wrapper,  // wrapper自己实现Indexer接口
		Selector:    wrapper, // wrapper自己实现Selector接口
		Focuser:     wrapper, // wrapper自己实现Focuser接口
		ComponentID: id,
	}
	return wrapper
}

func (w *MenuComponentWrapper) Init() tea.Cmd {
	return nil
}

// GetModel returns the underlying model
func (w *MenuComponentWrapper) GetModel() interface{} {
	return w.model
}

// GetID returns the component ID
func (w *MenuComponentWrapper) GetID() string {
	return w.model.ID
}

// View returns the view of the component
func (w *MenuComponentWrapper) View() string {
	return w.model.View()
}

// PublishEvent creates and returns a command to publish an event
func (w *MenuComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *MenuComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For menu component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.model.ID,
			Timestamp: time.Now(),
		}
	}
}

func (w *MenuComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// 使用通用消息处理模板（最小化封装，最大化委托）
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,                // 实现了 InteractiveBehavior 接口的组件
		msg,              // 接收的消息
		w.getBindings,     // 获取按键绑定的函数
		w.handleBinding,   // 处理按键绑定的函数
		w.delegateToBubbles, // 委托给原 bubbles 组件的函数
	)

	return w, cmd, response
}

// 实现 InteractiveBehavior 接口的方法

func (w *MenuComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *MenuComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// MenuComponentWrapper 已经实现了 core.ComponentWrapper 接口，无需适配器
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *MenuComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	// 由于HandleMenuUpdate函数期望*MenuInteractiveModel，我们需要创建一个临时指针
	tempModel := w.model
	updatedModel, menuCmd := HandleMenuUpdate(msg, &tempModel)
	w.model = updatedModel
	if cmd == nil {
		cmd = menuCmd
	}
	return cmd
}

// 实现 StateCapturable 接口
func (w *MenuComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *MenuComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HasFocus 方法
func (w *MenuComponentWrapper) HasFocus() bool {
	return w.model.focused
}

// 实现 HandleSpecialKey 方法（仅处理Menu特定逻辑）
func (w *MenuComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyTab:
		// 让Tab键冒泡以处理组件导航
		return nil, core.Ignored, true

	case tea.KeyEnter:
		// 处理回车键：检查是否需要进入子菜单或选择叶子项
		selectedItem := w.model.Model.SelectedItem()
		if selectedItem != nil {
			if menuItem, ok := selectedItem.(MenuItem); ok {
				if menuItem.HasSubmenu() {
					// 进入子菜单
					return w.model.EnterSubmenu(), core.Handled, true
				} else {
					// 叶子项选择
					log.Trace("MenuComponentWrapper: Selected leaf item: %s", menuItem.Title)
					var cmds []tea.Cmd

					// 发布选择事件
					cmds = append(cmds, core.PublishEvent(w.model.ID, core.EventMenuItemSelected, map[string]interface{}{
						"item":   menuItem,
						"action": menuItem.Action,
						"path":   w.model.Path,
						"level":  w.model.CurrentLevel,
					}))

					// 如果有action，执行它
					if menuItem.Action != nil {
						cmds = append(cmds, core.PublishEvent(w.model.ID, core.EventMenuActionTriggered, map[string]interface{}{
							"item":   menuItem,
							"action": menuItem.Action,
							"path":   w.model.Path,
							"level":  w.model.CurrentLevel,
						}))
					}

					if len(cmds) > 0 {
						return tea.Batch(cmds...), core.Handled, true
					}
				}
			}
		}
		return nil, core.Handled, true

	case tea.KeyEscape:
		// 处理ESC键：返回父菜单或失焦
		if w.model.CurrentLevel > 0 && len(w.model.Path) > 0 {
			// 返回父菜单
			return w.model.ExitSubmenu(), core.Handled, true
		}
		// 已经在顶层，失焦
		w.model.focused = false
		cmd := core.PublishEvent(w.model.ID, core.EventEscapePressed, nil)
		return cmd, core.Handled, true
	}

	// 其他按键不由这个函数处理，委托给原组件
	return nil, core.Handled, false
}

// SetFocus sets or removes focus from the menu component
func (m *MenuInteractiveModel) SetFocus(focus bool) {
	m.focused = focus
	// Note: list.Model doesn't have Focus/Blur methods
	// but we can update visual indicators based on focus state
	// For now, just track the focus state internally
}

func (m *MenuInteractiveModel) GetFocus() bool {
	return m.focused
}

func (w *MenuComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
}

func (w *MenuComponentWrapper) GetFocus() bool {
	return w.model.focused
}

// Index returns the current cursor position
func (w *MenuComponentWrapper) Index() int {
	return w.model.Index()
}

// SelectedItem returns the currently selected item
func (w *MenuComponentWrapper) SelectedItem() interface{} {
	return w.model.SelectedItem()
}

// Focused returns whether the menu is focused
func (w *MenuComponentWrapper) Focused() bool {
	return w.model.Focused()
}

func (w *MenuComponentWrapper) GetComponentType() string {
	return "menu"
}

func (w *MenuComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *MenuComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("MenuComponentWrapper: invalid data type")
	}

	// Parse menu properties
	props := ParseMenuProps(propsMap)

	// Update component properties
	w.model.props = props
	w.bindings = props.Bindings

	// Update menu items if provided
	if props.Items != nil {
		menuItems := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			menuItems[i] = MenuItem{
				Title:       item.Title,
				Description: item.Description,
				Value:       item.Value,
				Action:      item.Action,
				Disabled:    item.Disabled,
			}
		}
		w.model.Model.SetItems(menuItems)
	}

	return nil
}

// Cleanup cleans up resources used by the menu component
func (w *MenuComponentWrapper) Cleanup() {
	// Menu components typically don't need cleanup
	// This is a no-op for menu components
}

// GetStateChanges returns state changes from this component
func (w *MenuComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	selectedItem, ok := w.model.GetSelectedItem()
	if !ok {
		return map[string]interface{}{
			w.GetID() + "_selected_index": -1,
			w.GetID() + "_selected_item":  nil,
		}, false
	}

	return map[string]interface{}{
		w.GetID() + "_selected_index": w.model.Index(),
		w.GetID() + "_selected_item": map[string]interface{}{
			"title":      selectedItem.Title,
			"value":      selectedItem.Value,
			"disabled":   selectedItem.Disabled,
			"hasSubmenu": selectedItem.HasSubmenu(),
		},
	}, true
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *MenuComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}