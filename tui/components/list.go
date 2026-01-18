package components

import (
	"encoding/json"
	"fmt"

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
}

// NewListComponentWrapper creates a wrapper that implements ComponentInterface
func NewListComponentWrapper(listModel *ListModel) *ListComponentWrapper {
	return &ListComponentWrapper{
		model: listModel,
	}
}

func (w *ListComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *ListComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case tea.KeyMsg:
		oldIndex := w.model.Index()
		var cmds []tea.Cmd

		switch msg.Type {
		case tea.KeyEnter:
			if selectedItem := w.model.SelectedItem(); selectedItem != nil {
				item := selectedItem.(ListItem)
				// Publish item selected event
				cmds = append(cmds, core.PublishEvent(w.model.id, core.EventMenuItemSelected, map[string]interface{}{
					"item":  item,
					"index": w.model.Index(),
					"title": item.Title,
					"value": item.Value,
				}))
			}
		}

		// For other key messages, update the model
		var cmd tea.Cmd
		w.model.Model, cmd = w.model.Model.Update(msg)

		// Check if selection changed
		if w.model.Index() != oldIndex {
			cmds = append(cmds, cmd)
			if len(cmds) > 0 {
				return w, tea.Batch(cmds...), core.Handled
			}
		}
		return w, cmd, core.Handled
	}

	// For other messages, update using the underlying model
	oldIndex := w.model.Index()
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// Check if selection changed
	if w.model.Index() != oldIndex {
		// Publish selection changed event
		eventCmd := core.PublishEvent(w.model.id, "LIST_SELECTION_CHANGED", map[string]interface{}{
			"oldIndex": oldIndex,
			"newIndex": w.model.Index(),
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}
	return w, cmd, core.Handled
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

func (w *ListComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}
