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
	"github.com/yaoapp/yao/tui/core"
)

// ListItem represents a single list item
type ListItem struct {
	titleText string      // Internal storage for title
	descText  string      // Internal storage for description
	Value     interface{} `json:"value"`
	Disabled  bool        `json:"disabled"`
	Selected  bool        `json:"selected"`
}

// list.Item interface implementation

// Title returns the item's title (implements list.Item interface)
func (i ListItem) Title() string {
	return i.titleText
}

// Description returns the item's description (implements list.Item interface)
func (i ListItem) Description() string {
	return i.descText
}

// FilterValue returns the value used for filtering (implements list.Item interface)
func (i ListItem) FilterValue() string {
	return i.titleText
}

// ListItemDelegate is a compact delegate for rendering list items
type ListItemDelegate struct{}

func (d ListItemDelegate) Height() int  { return 1 }
func (d ListItemDelegate) Spacing() int { return 0 }
func (d ListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d ListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ListItem)
	if !ok {
		return
	}

	// Render the item
	// Simple rendering: title only, without description to save space
	str := fmt.Sprintf(" %s", i.Title())

	// Style differently if selected
	if index == m.Index() {
		// Selected item style
		color := lipgloss.Color("170") // Salmon pink
		fmt.Fprintf(w, "%s", lipgloss.NewStyle().Foreground(color).Render("> "+str))
	} else {
		// Normal item
		fmt.Fprintf(w, "  %s", str)
	}
}

// ListConfig defines the properties for the List component
type ListConfig struct {
	// Items contains the list items
	Items []ListItem `json:"items"`

	// Title is the optional title for the list
	Title string `json:"title"`

	// Height specifies the list height (0 for auto)
	Height int `json:"height"`

	// Width specifies the list width (0 for auto)
	Width int `json:"width"`

	// Focused determines if the list is focused (for selection)
	// This is managed internally via SetFocus(), not from config
	Focused bool `json:"-"`

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

// ListComponent represents a list component implementing ComponentInterface
type ListComponent struct {
	model       list.Model
	props       ListConfig
	id          string
	bindings    []core.ComponentBinding
	stateHelper *core.ListStateHelper
}

// NewListComponent creates a new list component with the given configuration
func NewListComponent(config core.RenderConfig, id string) *ListComponent {
	props := ListConfig{
		ShowTitle:        true,
		ShowStatusBar:    true,
		ShowFilter:       true,
		FilteringEnabled: true,
	}

	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseListPropsWithBinding(dataMap)
		}
	}

	items := make([]list.Item, len(props.Items))
	for i, item := range props.Items {
		items[i] = item
	}

	// Use a compact delegate similar to list-simple example
	delegate := &ListItemDelegate{}

	// Initialize with default non-zero dimensions to avoid panic/issues
	// These will be overwritten by updateListDimensions shortly after
	width := 20
	height := 10
	if config.Width > 0 {
		width = config.Width
	}
	if config.Height > 0 {
		height = config.Height
	}

	l := list.New(items, delegate, width, height)

	if props.Title != "" && props.ShowTitle {
		l.Title = props.Title
	}

	if props.Width > 0 {
		l.SetWidth(props.Width)
	}

	if props.Height > 0 {
		l.SetHeight(props.Height)
	}

	l.SetShowTitle(props.ShowTitle)
	l.SetShowStatusBar(props.ShowStatusBar)
	l.SetShowFilter(props.ShowFilter)
	l.SetFilteringEnabled(props.FilteringEnabled)

	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	l.Styles.Title = l.Styles.Title.Inherit(style)
	l.Styles.NoItems = l.Styles.NoItems.Inherit(style)

	component := &ListComponent{
		model:    l,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	// Initialize with correct dimensions from config if not specified in props
	component.updateListDimensions(config, props)

	component.stateHelper = &core.ListStateHelper{
		Indexer:     component,
		Selector:    component,
		Focused:     props.Focused,
		ComponentID: id,
	}

	return component
}

func (c *ListComponent) Init() tea.Cmd {
	return nil
}

func (c *ListComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	cmd, response := core.DefaultInteractiveUpdateMsg(
		c,
		msg,
		c.getBindings,
		c.handleBinding,
		c.delegateToBubbles,
	)

	// The bubbles list model might have been updated via delegateToBubbles
	// Return the current instance which contains the updated state
	return c, cmd, response
}

func (c *ListComponent) getBindings() []core.ComponentBinding {
	return c.bindings
}

func (c *ListComponent) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	cmd, response, handled := core.HandleBinding(c, keyMsg, binding)
	return cmd, response, handled
}

func (c *ListComponent) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.model, cmd = c.model.Update(msg)
	return cmd
}

func (c *ListComponent) CaptureState() map[string]interface{} {
	return c.stateHelper.CaptureState()
}

func (c *ListComponent) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return c.stateHelper.DetectStateChanges(old, new)
}

func (c *ListComponent) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	switch keyMsg.Type {
	case tea.KeyEnter:
		if selectedItem := c.model.SelectedItem(); selectedItem != nil {
			item := selectedItem.(ListItem)
			cmd := core.PublishEvent(c.id, core.EventMenuItemSelected, map[string]interface{}{
				"item":  item,
				"index": c.model.Index(),
				"title": item.Title,
				"value": item.Value,
			})
			return cmd, core.Handled, true
		}
	case tea.KeyEsc:
		// ESC 键：组件自己处理失去焦点
		// 通过发送 FocusMsg 来通知焦点变化
		if c.props.Focused {
			// 发送 FocusLost 消息给自己
			cmd := func() tea.Msg {
				return core.TargetedMsg{
					TargetID: c.id,
					InnerMsg: core.FocusMsg{
						Type:   core.FocusLost,
						Reason: "USER_ESC",
						ToID:   "",
					},
				}
			}
			return cmd, core.Handled, true
		}
	}
	return nil, core.Ignored, false
}

func (c *ListComponent) Index() int {
	return c.model.Index()
}

func (c *ListComponent) SelectedItem() interface{} {
	return c.model.SelectedItem()
}

func (c *ListComponent) GetModel() interface{} {
	return c.model
}

func (c *ListComponent) GetID() string {
	return c.id
}

func (c *ListComponent) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

func (c *ListComponent) ExecuteAction(action *core.Action) tea.Cmd {
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  c.id,
			Timestamp: time.Now(),
		}
	}
}

func (c *ListComponent) View() string {
	return c.model.View()
}

func (c *ListComponent) SetFocus(focus bool) {
	c.props.Focused = focus
}

func (c *ListComponent) GetFocus() bool {
	return c.props.Focused
}

// SetSize sets the allocated size for the component.
func (c *ListComponent) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

func (c *ListComponent) GetComponentType() string {
	return "list"
}

func (c *ListComponent) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ListComponent: invalid data type")
	}

	props := ParseListPropsWithBinding(propsMap)

	// Preserve the focused state - don't overwrite it with render config
	oldFocused := c.props.Focused
	c.props = props
	c.props.Focused = oldFocused

	// Update items
	if props.Items != nil {
		items := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			items[i] = item
		}
		c.model.SetItems(items)
	}

	// Update dimensions from render config
	c.updateListDimensions(config, props)

	return nil
}

func (c *ListComponent) Render(config core.RenderConfig) (string, error) {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ListComponent: invalid data type")
	}

	props := ParseListPropsWithBinding(propsMap)

	// Preserve the focused state - don't overwrite it with render config
	oldFocused := c.props.Focused
	c.props = props
	c.props.Focused = oldFocused

	// Update items
	if props.Items != nil {
		items := make([]list.Item, len(props.Items))
		for i, item := range props.Items {
			items[i] = item
		}
		c.model.SetItems(items)
	}

	// Update dimensions from render config
	c.updateListDimensions(config, props)

	return c.View(), nil
}

// updateListDimensions updates the list model's width and height based on config and props
func (c *ListComponent) updateListDimensions(config core.RenderConfig, props ListConfig) {
	// Props-specified width/height take precedence (for fixed sizing)
	if props.Width > 0 {
		c.model.SetWidth(props.Width)
	} else if config.Width > 0 {
		// Otherwise use render config width (from window size)
		c.model.SetWidth(config.Width)
	}

	if props.Height > 0 {
		c.model.SetHeight(props.Height)
	} else if config.Height > 0 {
		// Otherwise use render config height (from window size)
		c.model.SetHeight(config.Height)
	}
}

func (c *ListComponent) Cleanup() {}

func (c *ListComponent) GetStateChanges() (map[string]interface{}, bool) {
	selectedItem := c.model.SelectedItem()
	if selectedItem == nil {
		return map[string]interface{}{
			c.GetID() + "_selected_index": -1,
			c.GetID() + "_selected_item":  nil,
		}, false
	}

	return map[string]interface{}{
		c.GetID() + "_selected_index": c.model.Index(),
		c.GetID() + "_selected_item":  selectedItem,
	}, true
}

func (c *ListComponent) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}

// ParseListPropsWithBinding converts a props map to ListProps with __bind_data support
func ParseListPropsWithBinding(props map[string]interface{}) ListConfig {
	lp := ListConfig{
		ShowTitle:        true,
		ShowStatusBar:    true,
		ShowFilter:       true,
		FilteringEnabled: true,
	}

	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &lp)
	}

	var items []interface{}
	var itemTemplate string
	if template, ok := props["itemTemplate"].(string); ok {
		itemTemplate = template
	}

	if bindData, ok := props["__bind_data"].([]interface{}); ok {
		items = bindData
		lp.Items = []ListItem{}
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				listItem := ListItem{
					titleText: fmt.Sprintf("%v", itemMap),
					Value:     item,
				}

				// Check for explicit title field first
				if title, ok := itemMap["title"].(string); ok && title != "" {
					listItem.titleText = title
				} else if itemTemplate != "" {
					// Use itemTemplate to generate title
					listItem.titleText = applyTemplate(itemTemplate, itemMap)
				} else {
					// Try to find a reasonable fallback
					if name, ok := itemMap["name"].(string); ok && name != "" {
						listItem.titleText = name
					} else if id, ok := itemMap["id"]; ok {
						listItem.titleText = fmt.Sprintf("Item %v", id)
					}
				}

				// Handle description
				if desc, ok := itemMap["description"].(string); ok {
					listItem.descText = desc
				}

				// Handle value
				if val, ok := itemMap["value"]; ok {
					listItem.Value = val
				}
				lp.Items = append(lp.Items, listItem)
			} else {
				lp.Items = append(lp.Items, ListItem{
					titleText: fmt.Sprintf("%v", item),
					Value:     item,
				})
			}
		}
	}

	return lp
}

// applyTemplate applies a simple template string to item data
// Supports placeholders like {{id}}, {{name}}, {{price}}
func applyTemplate(template string, data map[string]interface{}) string {
	result := template

	// Simple replacement for {{field}} patterns
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		result = fmt.Sprintf("%s", strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value)))
	}

	return result
}

// ParseListProps converts a generic props map to ListProps (legacy, kept for compatibility)
func ParseListProps(props map[string]interface{}) ListConfig {
	return ParseListPropsWithBinding(props)
}
