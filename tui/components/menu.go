package components

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// When Enter is pressed, we could trigger an action
			log.Trace("Menu Delegate Update: Enter key pressed on item")
			// Get the selected item
			selectedItem := m.SelectedItem()
			if menuItem, ok := selectedItem.(MenuItem); ok && menuItem.Action != nil {
				log.Trace("Menu Delegate Update: Item has action, triggering action for: %s", menuItem.Title)
				// Here we could return a command to execute the action
				// For now, just return nil
			}
			return nil
		default:
			log.Trace("Menu Delegate Update: Other key pressed (%s), passing to list", msg.Type.String())
		}
	default:
		log.Trace("Menu Delegate Update: Non-key message received, passing to list")
	}
	// Let the list handle other messages like navigation
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
		// log.Trace("Menu Render: Processing item: %s (disabled: %t, hasChildren: %t)", item.Title, item.Disabled, item.HasSubmenu())

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
					// log.Trace("Menu ParseProps: Added item: %s (disabled: %t, hasChildren: %t)", menuItem.Title, menuItem.Disabled, menuItem.HasSubmenu())
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

// HandleMenuUpdate handles updates for menu components
// This is used when the menu is interactive (selection, scrolling, etc.)
func HandleMenuUpdate(msg tea.Msg, menuModel *MenuInteractiveModel) (MenuInteractiveModel, tea.Cmd) {
	if menuModel == nil {
		log.Trace("Menu Update: Received update with nil menu model")
		return MenuInteractiveModel{}, nil
	}

	log.Trace("Menu Update: Handling message type: %T, current level: %d, path: %v", msg, menuModel.CurrentLevel, menuModel.Path)

	var cmd tea.Cmd
	// Handle special menu-specific messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Trace("Menu Update: Key pressed: %s (type: %s)", msg.String(), msg.Type.String())
		switch msg.Type {
		case tea.KeyUp:
			log.Trace("Menu Update: Up arrow pressed, cursor before: %d", menuModel.Index())
			// Let the underlying list model handle the navigation
			updatedListModel, listCmd := menuModel.Model.Update(msg)
			menuModel.Model = updatedListModel
			log.Trace("Menu Update: Up arrow pressed, cursor after: %d", menuModel.Index())
			// Always return a refresh command to ensure UI updates after navigation
			refreshCmd := func() tea.Msg { return struct{}{} }
			if listCmd != nil {
				return *menuModel, tea.Batch(listCmd, refreshCmd)
			}
			return *menuModel, refreshCmd
		case tea.KeyDown:
			log.Trace("Menu Update: Down arrow pressed, cursor before: %d", menuModel.Index())
			// Let the underlying list model handle the navigation
			updatedListModel, listCmd := menuModel.Model.Update(msg)
			menuModel.Model = updatedListModel
			log.Trace("Menu Update: Down arrow pressed, cursor after: %d", menuModel.Index())
			// Always return a refresh command to ensure UI updates after navigation
			refreshCmd := func() tea.Msg { return struct{}{} }
			if listCmd != nil {
				return *menuModel, tea.Batch(listCmd, refreshCmd)
			}
			return *menuModel, refreshCmd
		default:
			// Handle string-based keys
			switch msg.String() {
			case "enter":
				log.Trace("Menu Update: Enter key pressed")
				// When Enter is pressed, check if the selected item has children
				selectedItem := menuModel.Model.SelectedItem()
				if menuItem, ok := selectedItem.(MenuItem); ok {
					log.Trace("Menu Update: Selected item: %s (has submenu: %t, action: %v)", menuItem.Title, menuItem.HasSubmenu(), menuItem.Action)
					if menuItem.HasSubmenu() {
						log.Trace("Menu Update: Navigating to submenu for item: %s", menuItem.Title)
						// Navigate to submenu
						menuModel.Path = append(menuModel.Path, menuItem.Title)
						menuModel.CurrentLevel++
						// Load submenu items
						submenuItems := make([]list.Item, len(menuItem.Children))
						for i, child := range menuItem.Children {
							submenuItems[i] = child
						}
						menuModel.SetItems(submenuItems)
						// Maintain the same width as the original menu
						if menuModel.props.Width > 0 {
							menuModel.SetWidth(menuModel.props.Width)
						}
						menuModel.Select(0) // Select first item in submenu
						log.Trace("Menu Update: Submenu loaded with %d items, now at level %d", len(submenuItems), menuModel.CurrentLevel)
						// Publish submenu entered event
						cmd = core.PublishEvent(menuModel.ID, core.EventMenuSubmenuEntered, map[string]interface{}{
							"item":        menuItem,
							"parentPath":  menuModel.Path[:len(menuModel.Path)-1],
							"currentPath": menuModel.Path,
							"level":       menuModel.CurrentLevel,
						})
					} else {
						// Leaf item selected
						var cmds []tea.Cmd
						// Publish item selected event
						cmds = append(cmds, core.PublishEvent(menuModel.ID, core.EventMenuItemSelected, map[string]interface{}{
							"item":   menuItem,
							"action": menuItem.Action,
							"path":   menuModel.Path,
							"level":  menuModel.CurrentLevel,
						}))
						if menuItem.Action != nil {
							log.Trace("Menu Update: Executing action for item: %s", menuItem.Title)
							// Execute the action for the item
							actionCmd := func() tea.Msg {
								// Use local MenuItem directly for message
								return core.MenuActionTriggered{Item: menuItem, Action: menuItem.Action}
							}
							cmds = append(cmds, actionCmd)
							// Also publish action triggered event
							cmds = append(cmds, core.PublishEvent(menuModel.ID, core.EventMenuActionTriggered, map[string]interface{}{
								"item":   menuItem,
								"action": menuItem.Action,
								"path":   menuModel.Path,
								"level":  menuModel.CurrentLevel,
							}))
						}
						if len(cmds) > 0 {
							cmd = tea.Batch(cmds...)
						}
					}
				} else {
					log.Trace("Menu Update: Selected item is not a MenuItem, skipping action")
				}
			case "right", "l":
				log.Trace("Menu Update: Right/l key pressed, attempting to enter submenu")
				// Navigate to submenu if available (alternative to Enter)
				selectedItem := menuModel.Model.SelectedItem()
				if menuItem, ok := selectedItem.(MenuItem); ok && menuItem.HasSubmenu() {
					log.Trace("Menu Update: Navigating to submenu for item: %s", menuItem.Title)
					// Navigate to submenu
					menuModel.Path = append(menuModel.Path, menuItem.Title)
					menuModel.CurrentLevel++
					// Load submenu items
					submenuItems := make([]list.Item, len(menuItem.Children))
					for i, child := range menuItem.Children {
						submenuItems[i] = child
					}
					menuModel.SetItems(submenuItems)
					// Maintain the same width as the original menu
					if menuModel.props.Width > 0 {
						menuModel.SetWidth(menuModel.props.Width)
					}
					menuModel.Select(0) // Select first item in submenu
					log.Trace("Menu Update: Submenu loaded with %d items, now at level %d", len(submenuItems), menuModel.CurrentLevel)
					// Publish submenu entered event
					cmd = core.PublishEvent(menuModel.ID, core.EventMenuSubmenuEntered, map[string]interface{}{
						"item":        menuItem,
						"parentPath":  menuModel.Path[:len(menuModel.Path)-1],
						"currentPath": menuModel.Path,
						"level":       menuModel.CurrentLevel,
					})
				}
			case "left", "h":
				log.Trace("Menu Update: Left/h key pressed, attempting to navigate back")
				// Navigate back to parent menu if at submenu level
				if menuModel.CurrentLevel > 0 && len(menuModel.Path) > 0 {
					log.Trace("Menu Update: Navigating back from level %d", menuModel.CurrentLevel)
					// Save current path before navigating back
					previousPath := make([]string, len(menuModel.Path))
					copy(previousPath, menuModel.Path)
					// Go back to parent menu
					menuModel.CurrentLevel--
					menuModel.Path = menuModel.Path[:len(menuModel.Path)-1]
					// Reload parent menu (this is simplified - in a real implementation you'd need to store parent menus)
					// For now, we'll reset to original items
					originalItems := make([]list.Item, len(menuModel.props.Items))
					for i, item := range menuModel.props.Items {
						originalItems[i] = item
					}
					menuModel.SetItems(originalItems)
					// Maintain the same width as the original menu
					if menuModel.props.Width > 0 {
						menuModel.SetWidth(menuModel.props.Width)
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
							// Add some padding to the calculated width
							menuModel.SetWidth(maxWidth + 4)
							log.Trace("Menu Update: Set calculated width to %d when returning to parent", maxWidth+4)
						}
					}
					menuModel.Select(0)
					log.Trace("Menu Update: Returned to parent menu at level %d", menuModel.CurrentLevel)
					// Publish submenu exited event
					cmd = core.PublishEvent(menuModel.ID, core.EventMenuSubmenuExited, map[string]interface{}{
						"previousPath": previousPath,
						"currentPath":  menuModel.Path,
						"level":        menuModel.CurrentLevel,
					})
				} else {
					log.Trace("Menu Update: Already at top level or no path to go back")
				}
			case "q", "ctrl+c", "esc":
				log.Trace("Menu Update: Exit key pressed (%s), initiating quit", msg.String())
				// Handle exit/quit
				cmd = tea.Quit
				fallthrough // Also pass to list model to handle quit behavior
			default:
				log.Trace("Menu Update: Other key pressed (%s), passing to list model", msg.String())
				// Pass other keys to the list model for default navigation and other functionality
				updatedListModel, listCmd := menuModel.Model.Update(msg)
				menuModel.Model = updatedListModel
				if cmd == nil {
					cmd = listCmd
				}
				return *menuModel, cmd
			}
		}
		// After handling our custom logic, still pass to the list model to ensure cursor updates
		updatedListModel, listCmd := menuModel.Model.Update(msg)
		menuModel.Model = updatedListModel
		if cmd == nil {
			cmd = listCmd
		}
		return *menuModel, cmd
	default:
		log.Trace("Menu Update: Non-key message received, passing to list model")
		// Pass non-key messages to the list model
		updatedListModel, listCmd := menuModel.Model.Update(msg)
		menuModel.Model = updatedListModel
		if cmd == nil {
			cmd = listCmd
		}
		return *menuModel, cmd
	}

	log.Trace("Menu Update: Completed update, returning cmd: %v", cmd != nil)
	return *menuModel, cmd
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


// MenuComponentWrapper wraps MenuInteractiveModel to implement ComponentInterface properly
type MenuComponentWrapper struct {
	model *MenuInteractiveModel
}

// NewMenuComponentWrapper creates a wrapper that implements ComponentInterface
func NewMenuComponentWrapper(menuModel *MenuInteractiveModel) *MenuComponentWrapper {
	return &MenuComponentWrapper{
		model: menuModel,
	}
}

func (w *MenuComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *MenuComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle key press events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Blur menu when ESC is pressed - but list.Model doesn't have Blur
			// Return Handled to indicate we processed it
			return w, nil, core.Handled
		case tea.KeyEnter:
			// Let Enter bubble to handleKeyPress for action execution
			// We'll handle action selection here but return Ignored
			// to allow of parent to process action
			selectedItem := w.model.Model.SelectedItem()
			if selectedItem != nil {
				if menuItem, ok := selectedItem.(MenuItem); ok && menuItem.Action != nil {
					log.Trace("MenuComponentWrapper: Selected item has action: %s", menuItem.Title)
					// The action will be triggered by parent component
				}
			}
			return w, nil, core.Ignored
		}
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.ID {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For other messages, update using the underlying model
	var cmd tea.Cmd
	updatedModel, menuCmd := HandleMenuUpdate(msg, w.model)
	w.model = &updatedModel
	if cmd == nil {
		cmd = menuCmd
	}
	return w, cmd, core.Handled
}

func (w *MenuComponentWrapper) View() string {
	return w.model.View()
}

func (w *MenuComponentWrapper) GetID() string {
	return w.model.ID
}

// SetFocus sets or removes focus from the menu component
func (m *MenuInteractiveModel) SetFocus(focus bool) {
	m.focused = focus
	// Note: list.Model doesn't have Focus/Blur methods
	// but we can update visual indicators based on focus state
	// For now, just track the focus state internally
}

func (w *MenuComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
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

// GetStateChanges returns the state changes from this component
func (w *MenuComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	selectedItem, ok := w.model.GetSelectedItem()
	if !ok {
		return map[string]interface{}{
			w.GetID() + "_selected_index": -1,
			w.GetID() + "_selected_item": nil,
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
