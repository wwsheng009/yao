package components

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	model       viewport.Model
	props       ViewportProps
	id          string
	bindings    []core.ComponentBinding
	stateHelper *core.ViewportStateHelper
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

	wrapper := &ViewportComponentWrapper{
		model:    vp,
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	wrapper.stateHelper = &core.ViewportStateHelper{
		Scroller:    wrapper,
		ComponentID: id,
	}

	return wrapper
}

func (w *ViewportComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *ViewportComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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

func (w *ViewportComponentWrapper) View() string {
	return w.model.View()
}

func (w *ViewportComponentWrapper) GetID() string {
	return w.id
}

// GetOffset returns the current offset of the viewport
func (w *ViewportComponentWrapper) GetOffset() int {
	return w.model.YOffset
}

// GetModel returns the underlying model of the component
func (w *ViewportComponentWrapper) GetModel() interface{} {
	return w.model
}

// 实现 InteractiveBehavior 接口的方法

func (w *ViewportComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *ViewportComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	// ViewportComponentWrapper 已经实现了 ComponentWrapper 接口，可以直接传递
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *ViewportComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *ViewportComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *ViewportComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *ViewportComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// ESC 和 Tab 现在由框架层统一处理，这里不处理
	// 如果有其他特殊的键处理需求，可以在这里添加
	return nil, core.Ignored, false
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
// PublishEvent creates and returns a command to publish an event
func (w *ViewportComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

// ExecuteAction executes an action
func (w *ViewportComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// For viewport component, we return a command that creates an ExecuteActionMsg
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}

func (w *ViewportComponentWrapper) SetFocus(focus bool) {
	// Viewport doesn't have visual focus indicators like other components
	// Focus tracking is mainly for keyboard event routing
	// No action needed for viewport as it handles scroll keys globally
}

func (w *ViewportComponentWrapper) GetFocus() bool {
	// Viewport doesn't have a traditional focus state, return false
	return false
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

	// Update dimensions from render config (Constraints)
	if m.props.Width > 0 {
		m.Model.Width = m.props.Width
	} else if config.Width > 0 {
		m.Model.Width = config.Width
	}

	if m.props.Height > 0 {
		m.Model.Height = m.props.Height
	} else if config.Height > 0 {
		m.Model.Height = config.Height
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

	// Update dimensions from render config (Constraints)
	// Priority: Props (Fixed) > Config (Constraint)
	if w.props.Width > 0 {
		w.model.Width = w.props.Width
	} else if config.Width > 0 {
		w.model.Width = config.Width
	}

	if w.props.Height > 0 {
		w.model.Height = w.props.Height
	} else if config.Height > 0 {
		w.model.Height = config.Height
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

// SetSize 更新视口的实际显示尺寸
func (w *ViewportComponentWrapper) SetSize(width, height int) {
	// 直接设置底层 viewport.Model 的尺寸
	w.model.Width = width
	w.model.Height = height
}

// Measure returns the ideal size of the viewport
func (w *ViewportComponentWrapper) Measure(maxWidth, maxHeight int) (width, height int) {
	// Use content from props
	content := w.props.Content

	// Split into lines
	lines := strings.Split(content, "\n")

	maxWidthLine := 0
	for _, line := range lines {
		// lipgloss.Width handles ANSI codes and character width
		w := lipgloss.Width(line)
		if w > maxWidthLine {
			maxWidthLine = w
		}
	}

	// Add border padding (approximate)
	width = maxWidthLine + 2
	if width > maxWidth {
		width = maxWidth
	}

	// Height is line count + border padding
	height = len(lines) + 2
	if height > maxHeight {
		height = maxHeight
	}

	return width, height
}
