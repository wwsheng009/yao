package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
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

// PaginatorComponentWrapper wraps PaginatorModel to implement ComponentInterface properly
type PaginatorComponentWrapper struct {
	model *PaginatorModel
}

// NewPaginatorComponentWrapper creates a wrapper that implements ComponentInterface
func NewPaginatorComponentWrapper(paginatorModel *PaginatorModel) *PaginatorComponentWrapper {
	return &PaginatorComponentWrapper{
		model: paginatorModel,
	}
}

func (w *PaginatorComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *PaginatorComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case tea.KeyMsg:
		oldPage := w.model.Page
		var cmds []tea.Cmd

		switch msg.Type {
		case tea.KeyLeft:
			if w.model.Page > 0 {
				w.model.Page--
				// Publish page changed event
				cmds = append(cmds, core.PublishEvent(w.model.id, "PAGINATOR_PAGE_CHANGED", map[string]interface{}{
					"oldPage": oldPage + 1,
					"newPage": w.model.Page + 1,
				}))
			}
		case tea.KeyRight:
			if w.model.Page < w.model.TotalPages-1 {
				w.model.Page++
				// Publish page changed event
				cmds = append(cmds, core.PublishEvent(w.model.id, "PAGINATOR_PAGE_CHANGED", map[string]interface{}{
					"oldPage": oldPage + 1,
					"newPage": w.model.Page + 1,
				}))
			}
		}

		// For other key messages, update the model
		var cmd tea.Cmd
		w.model.Model, cmd = w.model.Model.Update(msg)

		// Check if page changed
		if w.model.Page != oldPage {
			cmds = append(cmds, cmd)
			if len(cmds) > 0 {
				return w, tea.Batch(cmds...), core.Handled
			}
		}
		return w, cmd, core.Handled
	}

	// For other messages, update using the underlying model
	oldPage := w.model.Page
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// Check if page changed
	if w.model.Page != oldPage {
		// Publish page changed event
		eventCmd := core.PublishEvent(w.model.id, "PAGINATOR_PAGE_CHANGED", map[string]interface{}{
			"oldPage": oldPage + 1,
			"newPage": w.model.Page + 1,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}
	return w, cmd, core.Handled
}

func (w *PaginatorComponentWrapper) View() string {
	return w.model.View()
}

func (w *PaginatorComponentWrapper) GetID() string {
	return w.model.id
}

func (w *PaginatorComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
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

func (m *PaginatorModel) GetComponentType() string {
	return "paginator"
}

func (w *PaginatorComponentWrapper) GetComponentType() string {
	return "paginator"
}

func (m *PaginatorModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("PaginatorModel: invalid data type")
	}

	// Parse paginator properties
	props := ParsePaginatorProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

func (w *PaginatorComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}
