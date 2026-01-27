package component

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/tui/core"
)

// PaginatorProps defines the properties for the Paginator component
type PaginatorProps struct {
	// TotalPages is the total number of pages
	TotalPages int `json:"totalPages"`

	// CurrentPage is the current page (1-indexed)
	CurrentPage int `json:"currentPage"`

	// PageSize is the number of items per page
	PageSize int `json:"pageSize"`

	// TotalItems is the total number of items
	TotalItems int `json:"totalItems"`

	// Type specifies the paginator type: "dots" or "numbers"
	Type string `json:"type"`

	// Color specifies the active page color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// InactiveColor specifies the inactive page color
	InactiveColor string `json:"inactiveColor"`

	// Width specifies the paginator width (0 for auto)
	Width int `json:"width"`

	// Height specifies the paginator height (0 for auto)
	Height int `json:"height"`

	// ShowInfo shows/hides page info (e.g., "1/5")
	ShowInfo bool `json:"showInfo"`

	// Focused determines if the paginator is focused
	Focused bool `json:"focused"`

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// PaginatorModel wraps the paginator.Model to handle TUI integration
type PaginatorModel struct {
	paginator.Model
	props PaginatorProps
	id    string // Unique identifier for this instance
}

// RenderPaginator renders a paginator component
func RenderPaginator(props PaginatorProps, width int) string {
	p := paginator.New()

	// Set total pages
	if props.TotalPages > 0 {
		p.TotalPages = props.TotalPages
	} else if props.TotalItems > 0 && props.PageSize > 0 {
		p.TotalPages = (props.TotalItems + props.PageSize - 1) / props.PageSize
	}

	// Set current page
	if props.CurrentPage > 0 {
		p.Page = props.CurrentPage - 1 // Convert to 0-indexed
	}

	// Set paginator type
	if props.Type == "dots" {
		p.Type = paginator.Dots
	} else {
		p.Type = paginator.Arabic
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to paginator
	p.ActiveDot = style.Render("•")

	// Set inactive color
	if props.InactiveColor != "" {
		inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(props.InactiveColor))
		p.InactiveDot = inactiveStyle.Render("•")
	}

	// Set width if specified
	if props.Width > 0 {
		style = style.Width(props.Width)
	} else if width > 0 {
		style = style.Width(width)
	}

	// Build view
	view := p.View()

	// Add page info if requested
	if props.ShowInfo && p.TotalPages > 0 {
		info := lipgloss.NewStyle().Faint(true).Render(
			fmt.Sprintf(" (%d/%d)", p.Page+1, p.TotalPages),
		)
		view += info
	}

	return style.Render(view)
}

// ParsePaginatorProps converts a generic props map to PaginatorProps using JSON unmarshaling
func ParsePaginatorProps(props map[string]interface{}) PaginatorProps {
	// Set defaults
	pp := PaginatorProps{
		Type:        "dots",
		CurrentPage: 1,
		PageSize:    10,
		ShowInfo:    false,
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &pp)
	}

	return pp
}

// NewPaginatorModel creates a new PaginatorModel from PaginatorProps
func NewPaginatorModel(props PaginatorProps, id string) PaginatorModel {
	p := paginator.New()

	// Set total pages
	if props.TotalPages > 0 {
		p.TotalPages = props.TotalPages
	} else if props.TotalItems > 0 && props.PageSize > 0 {
		p.TotalPages = (props.TotalItems + props.PageSize - 1) / props.PageSize
	}

	// Set current page
	if props.CurrentPage > 0 {
		p.Page = props.CurrentPage - 1 // Convert to 0-indexed
	}

	// Set paginator type
	if props.Type == "dots" {
		p.Type = paginator.Dots
	} else {
		p.Type = paginator.Arabic
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to paginator
	p.ActiveDot = style.Render("•")

	// Set inactive color
	if props.InactiveColor != "" {
		inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(props.InactiveColor))
		p.InactiveDot = inactiveStyle.Render("•")
	}

	return PaginatorModel{
		Model: p,
		props: props,
		id:    id,
	}
}

// HandlePaginatorUpdate handles updates for paginator components
func HandlePaginatorUpdate(msg tea.Msg, paginatorModel *PaginatorModel) (PaginatorModel, tea.Cmd) {
	if paginatorModel == nil {
		return PaginatorModel{}, nil
	}

	var cmd tea.Cmd
	paginatorModel.Model, cmd = paginatorModel.Model.Update(msg)
	return *paginatorModel, cmd
}

// Init initializes the paginator model
func (m *PaginatorModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the paginator
func (m *PaginatorModel) View() string {
	view := m.Model.View()

	// Add page info if requested
	if m.props.ShowInfo && m.TotalPages > 0 {
		info := lipgloss.NewStyle().Faint(true).Render(
			fmt.Sprintf(" (%d/%d)", m.Page+1, m.TotalPages),
		)
		view += info
	}

	return view
}

// GetID returns the unique identifier for this component instance
func (m *PaginatorModel) GetID() string {
	return m.id
}

// SetFocus sets or removes focus from paginator component
func (m *PaginatorModel) SetFocus(focus bool) {
	m.props.Focused = focus
}

func (m *PaginatorModel) GetFocus() bool {
	return m.props.Focused
}

// PaginatorComponentWrapper wraps the native paginator.Model directly
type PaginatorComponentWrapper struct {
	model       paginator.Model
	props       PaginatorProps
	id          string
	bindings    []core.ComponentBinding
	stateHelper *core.PaginatorStateHelper
}

// NewPaginatorComponentWrapper creates a wrapper that implements ComponentInterface
func NewPaginatorComponentWrapper(props PaginatorProps, id string) *PaginatorComponentWrapper {
	p := paginator.New()

	// Set total pages
	if props.TotalPages > 0 {
		p.TotalPages = props.TotalPages
	} else if props.TotalItems > 0 && props.PageSize > 0 {
		p.TotalPages = (props.TotalItems + props.PageSize - 1) / props.PageSize
	}

	// Set current page
	if props.CurrentPage > 0 {
		p.Page = props.CurrentPage - 1 // Convert to 0-indexed
	}

	// Set paginator type
	if props.Type == "dots" {
		p.Type = paginator.Dots
	} else {
		p.Type = paginator.Arabic
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to paginator
	p.ActiveDot = style.Render("•")

	// Set inactive color
	if props.InactiveColor != "" {
		inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(props.InactiveColor))
		p.InactiveDot = inactiveStyle.Render("•")
	}

	wrapper := &PaginatorComponentWrapper{
		model:    p,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	wrapper.stateHelper = &core.PaginatorStateHelper{
		Pager:       wrapper,
		ComponentID: id,
	}

	return wrapper
}

func (w *PaginatorComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *PaginatorComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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

func (w *PaginatorComponentWrapper) View() string {
	view := w.model.View()

	// Add page info if requested
	if w.props.ShowInfo && w.model.TotalPages > 0 {
		info := lipgloss.NewStyle().Faint(true).Render(
			fmt.Sprintf(" (%d/%d)", w.model.Page+1, w.model.TotalPages),
		)
		view += info
	}

	return view
}

func (w *PaginatorComponentWrapper) GetID() string {
	return w.id
}

func (w *PaginatorComponentWrapper) SetFocus(focus bool) {
	// Paginator doesn't have focus concept in bubbletea paginator
	// But we can update our local property
	w.props.Focused = focus
}

func (w *PaginatorComponentWrapper) GetFocus() bool {
	return w.props.Focused
}

// SetSize sets the allocated size for the component.
func (w *PaginatorComponentWrapper) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

// GetCurrentPage returns the current page (1-indexed)
func (w *PaginatorComponentWrapper) GetCurrentPage() int {
	return w.model.Page + 1
}

// SetCurrentPage sets the current page (1-indexed)
func (w *PaginatorComponentWrapper) SetCurrentPage(page int) {
	if page > 0 && page <= w.model.TotalPages {
		w.model.Page = page - 1
	}
}

// 实现 InteractiveBehavior 接口的方法

func (w *PaginatorComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *PaginatorComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// PaginatorComponentWrapper 已经实现了 ComponentWrapper 接口，可以直接传递
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *PaginatorComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *PaginatorComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *PaginatorComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *PaginatorComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// ESC 和 Tab 现在由框架层统一处理，这里不处理
	// 如果有其他特殊的键处理需求，可以在这里添加
	return nil, core.Ignored, false
}

// PublishEvent creates and returns a command to publish an event
func (w *PaginatorComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *PaginatorComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For paginator component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}

// GetModel returns the underlying model of the component
func (w *PaginatorComponentWrapper) GetModel() interface{} {
	return w.model
}

func (m *PaginatorModel) GetComponentType() string {
	return "paginator"
}

func (w *PaginatorComponentWrapper) GetComponentType() string {
	return "paginator"
}

func (w *PaginatorComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("PaginatorComponentWrapper: invalid data type")
	}

	// Parse paginator properties
	props := ParsePaginatorProps(propsMap)

	// Update component properties
	w.props = props

	// Update pagination settings
	if props.TotalPages > 0 {
		w.model.TotalPages = props.TotalPages
	} else if props.TotalItems > 0 && props.PageSize > 0 {
		w.model.TotalPages = (props.TotalItems + props.PageSize - 1) / props.PageSize
	}

	// Set current page
	if props.CurrentPage > 0 {
		w.model.Page = props.CurrentPage - 1 // Convert to 0-indexed
	}

	// Set paginator type
	if props.Type == "dots" {
		w.model.Type = paginator.Dots
	} else {
		w.model.Type = paginator.Arabic
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to paginator
	w.model.ActiveDot = style.Render("•")

	// Set inactive color
	if props.InactiveColor != "" {
		inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(props.InactiveColor))
		w.model.InactiveDot = inactiveStyle.Render("•")
	}

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *PaginatorComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("PaginatorComponentWrapper: invalid data type")
	}

	// Parse paginator properties
	props := ParsePaginatorProps(propsMap)

	// Update component properties
	w.props = props

	// Update pagination settings
	if props.TotalPages > 0 {
		w.model.TotalPages = props.TotalPages
	} else if props.TotalItems > 0 && props.PageSize > 0 {
		w.model.TotalPages = (props.TotalItems + props.PageSize - 1) / props.PageSize
	}

	// Set current page
	if props.CurrentPage > 0 {
		w.model.Page = props.CurrentPage - 1 // Convert to 0-indexed
	}

	// Set paginator type
	if props.Type == "dots" {
		w.model.Type = paginator.Dots
	} else {
		w.model.Type = paginator.Arabic
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to paginator
	w.model.ActiveDot = style.Render("•")

	// Set inactive color
	if props.InactiveColor != "" {
		inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(props.InactiveColor))
		w.model.InactiveDot = inactiveStyle.Render("•")
	}

	return nil
}

// Cleanup cleans up resources used by the paginator component
func (w *PaginatorComponentWrapper) Cleanup() {
	// Paginator components typically don't need cleanup
	// This is a no-op for paginator components
}

// GetStateChanges returns the state changes from this component
func (w *PaginatorComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	page := w.model.Page
	perPage := w.model.PerPage
	totalPages := w.model.TotalPages

	return map[string]interface{}{
		w.GetID() + "_page":        page,
		w.GetID() + "_per_page":    perPage,
		w.GetID() + "_total_pages": totalPages,
	}, totalPages > 0
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *PaginatorComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}
