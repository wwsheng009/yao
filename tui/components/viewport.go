package components

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// ViewportProps defines the properties for the Viewport component
type ViewportProps struct {
	// Content is the content to display in viewport
	Content string `json:"content"`

	// Width specifies the viewport width (0 for auto)
	Width int `json:"width"`

	// Height specifies the viewport height (0 for auto)
	Height int `json:"height"`

	// ShowScrollbar determines if scrollbar is shown
	ShowScrollbar bool `json:"showScrollbar"`

	// ScrollbarStyle is the style for scrollbar
	ScrollbarStyle lipglossStyleWrapper `json:"scrollbarStyle"`

	// BorderStyle is the style for viewport border
	BorderStyle lipglossStyleWrapper `json:"borderStyle"`

	// Style is the general viewport style
	Style lipglossStyleWrapper `json:"style"`

	// EnableGlamour enables Markdown rendering with Glamour
	EnableGlamour bool `json:"enableGlamour"`

	// GlamourStyle sets the Glamour style for Markdown rendering
	GlamourStyle string `json:"glamourStyle"`

	// AutoScroll determines if viewport automatically scrolls to bottom
	AutoScroll bool `json:"autoScroll"`
	
	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// ViewportModel wraps the viewport.Model to handle TUI integration
type ViewportModel struct {
	viewport.Model
	props ViewportProps
}

// RenderViewport renders a viewport component
func RenderViewport(props ViewportProps, width int) string {
	vp := viewport.New(0, 0) // width and height will be set later

	// Set content
	content := props.Content

	// Apply Markdown rendering if enabled
	if props.EnableGlamour {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0), // Use viewport width for wrapping
		)
		if err == nil {
			rendered, err := renderer.Render(content)
			if err == nil {
				content = rendered
			}
		}
	}

	vp.SetContent(content)

	// Set dimensions
	viewWidth := props.Width
	if viewWidth <= 0 && width > 0 {
		viewWidth = width
	}

	viewHeight := props.Height
	if viewHeight <= 0 {
		// Estimate height based on content if not specified
		lineCount := strings.Count(content, "\n") + 1
		if lineCount < 10 {
			viewHeight = lineCount + 2 // Add some padding
		} else {
			viewHeight = 15 // Default height
		}
	}

	if viewWidth > 0 {
		vp.Width = viewWidth
	}
	if viewHeight > 0 {
		vp.Height = viewHeight
	}

	// Apply styles
	style := props.Style.GetStyle()
	if style.String() != lipgloss.NewStyle().String() {
		vp.Style = style
	}

	return vp.View()
}

// ParseViewportProps converts a generic props map to ViewportProps using JSON unmarshaling
func ParseViewportProps(props map[string]interface{}) ViewportProps {
	// Set defaults
	vp := ViewportProps{
		EnableGlamour: false,
		GlamourStyle:  "dark",
		AutoScroll:    false,
		ShowScrollbar: true,
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &vp)
	}

	return vp
}

// HandleViewportUpdate handles updates for viewport components
// This is used when the viewport is interactive (scrolling, etc.)
func HandleViewportUpdate(msg tea.Msg, viewportModel *ViewportModel) (ViewportModel, tea.Cmd) {
	if viewportModel == nil {
		return ViewportModel{}, nil
	}

	var cmd tea.Cmd
	viewportModel.Model, cmd = viewportModel.Model.Update(msg)
	return *viewportModel, cmd
}

// NewViewportModel creates a new ViewportModel from ViewportProps
func NewViewportModel(props ViewportProps, id string) ViewportModel {
	vp := viewport.New(0, 0) // width and height will be set later

	// Set content
	content := props.Content

	// Apply Markdown rendering if enabled
	if props.EnableGlamour {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0), // Use viewport width for wrapping
		)
		if err == nil {
			rendered, err := renderer.Render(content)
			if err == nil {
				content = rendered
			}
		}
	}

	vp.SetContent(content)

	// Set dimensions
	viewWidth := props.Width
	if viewWidth <= 0 {
		viewWidth = 80 // Default width
	}

	viewHeight := props.Height
	if viewHeight <= 0 {
		// Estimate height based on content if not specified
		lineCount := strings.Count(content, "\n") + 1
		if lineCount < 10 {
			viewHeight = lineCount + 2 // Add some padding
		} else {
			viewHeight = 15 // Default height
		}
	}

	vp.Width = viewWidth
	vp.Height = viewHeight

	// Apply styles
	style := props.Style.GetStyle()
	if style.String() != lipgloss.NewStyle().String() {
		vp.Style = style
	}

	return ViewportModel{
		Model: vp,
		props: props,
	}
}

// Init initializes the viewport model
func (m *ViewportModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the viewport
func (m *ViewportModel) View() string {
	return m.Model.View()
}

// GetID returns the unique identifier for this component instance
func (m *ViewportModel) GetID() string {
	return "" // ViewportModel doesn't have an id field, return empty
}

// SetContent updates the viewport content
func (m *ViewportModel) SetContent(content string) {
	newContent := content

	// Apply Markdown rendering if enabled
	if m.props.EnableGlamour {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0),
		)
		if err == nil {
			rendered, err := renderer.Render(content)
			if err == nil {
				newContent = rendered
			}
		}
	}

	m.Model.SetContent(newContent)

	// Auto-scroll to bottom if enabled
	if m.props.AutoScroll {
		m.Model.GotoBottom()
	}
}

// GotoTop scrolls to the top of the viewport
func (m *ViewportModel) GotoTop() {
	m.Model.GotoTop()
}

// GotoBottom scrolls to the bottom of the viewport
func (m *ViewportModel) GotoBottom() {
	m.Model.GotoBottom()
}

// ViewportComponentWrapper wraps the native viewport.Model directly
type ViewportComponentWrapper struct {
	model viewport.Model
	props ViewportProps
	id    string
}

// NewViewportComponentWrapper creates a wrapper that implements ComponentInterface
func NewViewportComponentWrapper(props ViewportProps, id string) *ViewportComponentWrapper {
	vp := viewport.New(0, 0) // width and height will be set later

	// Set content
	content := props.Content

	// Apply Markdown rendering if enabled
	if props.EnableGlamour {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0), // Use viewport width for wrapping
		)
		if err == nil {
			rendered, err := renderer.Render(content)
			if err == nil {
				content = rendered
			}
		}
	}

	vp.SetContent(content)

	// Set dimensions
	viewWidth := props.Width
	if viewWidth <= 0 {
		viewWidth = 80 // Default width
	}

	viewHeight := props.Height
	if viewHeight <= 0 {
		// Estimate height based on content if not specified
		lineCount := strings.Count(content, "\n") + 1
		if lineCount < 10 {
			viewHeight = lineCount + 2 // Add some padding
		} else {
			viewHeight = 15 // Default height
		}
	}

	vp.Width = viewWidth
	vp.Height = viewHeight

	// Apply styles
	style := props.Style.GetStyle()
	if style.String() != lipgloss.NewStyle().String() {
		vp.Style = style
	}

	return &ViewportComponentWrapper{
		model: vp,
		props: props,
		id:    id,
	}
}

func (w *ViewportComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *ViewportComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case tea.KeyMsg:
		// Handle scrolling keys
		switch msg.Type {
		case tea.KeyUp, tea.KeyPgUp:
			// Handle scroll up
			var cmd tea.Cmd
			w.model, cmd = w.model.Update(msg)
			return w, cmd, core.Handled

		case tea.KeyDown, tea.KeyPgDown:
			// Handle scroll down
			var cmd tea.Cmd
			w.model, cmd = w.model.Update(msg)
			return w, cmd, core.Handled

		case tea.KeyHome:
			// Scroll to top
			w.model.GotoTop()
			return w, nil, core.Handled

		case tea.KeyEnd:
			// Scroll to bottom
			w.model.GotoBottom()
			return w, nil, core.Handled

		case tea.KeyEsc:
			// Pass ESC to parent for focus management
			return w, nil, core.Ignored

		default:
			// For other key messages, update viewport
			var cmd tea.Cmd
			w.model, cmd = w.model.Update(msg)
			return w, cmd, core.Handled
		}

	case core.ActionMsg:
		// Handle internal action messages
		switch msg.Action {
		case core.EventDataLoaded:
			// Update content when data is loaded
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if content, ok := data["content"].(string); ok {
					w.SetContent(content)
					return w, nil, core.Handled
				}
			}
		case core.EventDataRefreshed:
			// Refresh content
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if content, ok := data["content"].(string); ok {
					w.SetContent(content)
					return w, nil, core.Handled
				}
			}
		}
		return w, nil, core.Ignored

	case tea.WindowSizeMsg:
		// Handle window resize - update viewport dimensions
		// Use new dimensions if not explicitly set
		if w.props.Width <= 0 {
			w.model.Width = msg.Width
		}
		if w.props.Height <= 0 {
			w.model.Height = msg.Height
		}
		// Let viewport handle the resize
		var cmd tea.Cmd
		w.model, cmd = w.model.Update(msg)
		return w, cmd, core.Handled
	}

	// For other messages, update viewport
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)
	return w, cmd, core.Handled
}

func (w *ViewportComponentWrapper) View() string {
	return w.model.View()
}

func (w *ViewportComponentWrapper) GetID() string {
	return w.id
}

// SetContent updates the viewport content through the wrapper
func (w *ViewportComponentWrapper) SetContent(content string) {
	newContent := content

	// Apply Markdown rendering if enabled
	if w.props.EnableGlamour {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0),
		)
		if err == nil {
			rendered, err := renderer.Render(content)
			if err == nil {
				newContent = rendered
			}
		}
	}

	w.model.SetContent(newContent)

	// Auto-scroll to bottom if enabled
	if w.props.AutoScroll {
		w.model.GotoBottom()
	}
}

// GotoTop scrolls to the top of the viewport through the wrapper
func (w *ViewportComponentWrapper) GotoTop() {
	w.model.GotoTop()
}

// GotoBottom scrolls to the bottom of the viewport through the wrapper
func (w *ViewportComponentWrapper) GotoBottom() {
	w.model.GotoBottom()
}

// SetFocus sets or removes focus from the viewport component
// Viewport doesn't have a traditional focus state, but we can track it
func (w *ViewportComponentWrapper) SetFocus(focus bool) {
	// Viewport doesn't have visual focus indicators like other components
	// Focus tracking is mainly for keyboard event routing
	// No action needed for viewport as it handles scroll keys globally
}

func (m *ViewportModel) GetComponentType() string {
	return "viewport"
}

func (w *ViewportComponentWrapper) GetComponentType() string {
	return "viewport"
}

func (m *ViewportModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ViewportModel: invalid data type")
	}

	// Parse viewport properties
	props := ParseViewportProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (m *ViewportModel) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ViewportModel: invalid data type for UpdateRenderConfig")
	}

	// Parse viewport properties
	props := ParseViewportProps(propsMap)

	// Update component properties
	m.props = props

	// Update content if provided
	if content, exists := propsMap["content"]; exists {
		if contentStr, ok := content.(string); ok {
			m.SetContent(contentStr)
		}
	}

	return nil
}

// Cleanup 清理资源
func (m *ViewportModel) Cleanup() {
	// ViewportModel 通常不需要特殊清理操作
}

func (w *ViewportComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ViewportComponentWrapper: invalid data type")
	}

	// Parse viewport properties
	props := ParseViewportProps(propsMap)

	// Update component properties
	w.props = props

	// Update content if provided
	if content, exists := propsMap["content"]; exists {
		if contentStr, ok := content.(string); ok {
			w.SetContent(contentStr)
		}
	}

	// Update dimensions if provided
	if width, exists := propsMap["width"]; exists {
		if widthInt, ok := width.(int); ok && widthInt > 0 {
			w.model.Width = widthInt
		}
	}
	if height, exists := propsMap["height"]; exists {
		if heightInt, ok := height.(int); ok && heightInt > 0 {
			w.model.Height = heightInt
		}
	}

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (w *ViewportComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ViewportComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse viewport properties
	props := ParseViewportProps(propsMap)

	// Update component properties
	w.props = props

	// Update content if provided
	if content, exists := propsMap["content"]; exists {
		if contentStr, ok := content.(string); ok {
			w.model.SetContent(contentStr)
		}
	}

	return nil
}

// Cleanup 清理资源
func (w *ViewportComponentWrapper) Cleanup() {
	// 视口组件通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (w *ViewportComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Viewport component may have state (scroll position, etc.)
	return map[string]interface{}{
		w.GetID() + "_offset": w.model.YOffset,
	}, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *ViewportComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
		"core.ActionMsg",
		"tea.WindowSizeMsg",
	}
}

