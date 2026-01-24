package components

import (
	"encoding/json"
	"fmt"

	help "github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// HelpItem represents a single help item
type HelpItem struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// HelpSection represents a section of help items
type HelpSection struct {
	Title string     `json:"title"`
	Items []HelpItem `json:"items"`
}

// HelpProps defines the properties for the Help component
type HelpProps struct {
	// KeyMap specifies the key bindings
	KeyMap map[string]interface{} `json:"keyMap"`

	// Sections specifies help sections with grouped items
	Sections []HelpSection `json:"sections"`

	// Width specifies the help width (0 for auto)
	Width int `json:"width"`

	// Height specifies the help height (0 for auto)
	Height int `json:"height"`

	// ShowAllKeys shows all keys or just common ones
	ShowAllKeys bool `json:"showAllKeys"`

	// Style specifies the help style: "full", "compact", "minimal", "sections"
	Style string `json:"style"`

	// Color specifies the text color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Border adds a border around the help
	Border bool `json:"border"`

	// BorderColor specifies the border color
	BorderColor string `json:"borderColor"`

	// Padding specifies the padding
	Padding []int `json:"padding"`

	// SectionTitleColor specifies the color for section titles
	SectionTitleColor string `json:"sectionTitleColor"`

	// SectionSeparator specifies the separator between sections
	SectionSeparator string `json:"sectionSeparator"`

	// ItemSeparator specifies the separator between items
	ItemSeparator string `json:"itemSeparator"`

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// HelpModel wraps the help.Model to handle TUI integration
type HelpModel struct {
	help.Model
	props HelpProps
	id    string // Unique identifier for this instance
}

// RenderHelp renders a help component
func RenderHelp(props HelpProps, width int) string {
	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Set width if specified
	if props.Width > 0 {
		style = style.Width(props.Width)
	} else if width > 0 {
		style = style.Width(width)
	}

	// Apply padding
	if len(props.Padding) > 0 {
		switch len(props.Padding) {
		case 1:
			style = style.Padding(props.Padding[0])
		case 2:
			style = style.Padding(props.Padding[0], props.Padding[1])
		case 4:
			style = style.Padding(props.Padding[0], props.Padding[1], props.Padding[2], props.Padding[3])
		}
	}

	// Add border if specified
	if props.Border {
		borderStyle := lipgloss.NewStyle()
		if props.BorderColor != "" {
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(props.BorderColor))
		}
		style = style.BorderStyle(lipgloss.NormalBorder()).Inherit(borderStyle)
	}

	// Create a simple help text
	helpText := "Use arrow keys to navigate, Enter to select"
	if props.KeyMap != nil && len(props.KeyMap) > 0 {
		helpText = "Available commands: "
		for key, desc := range props.KeyMap {
			helpText += key + ": " + desc.(string) + " "
		}
	}

	return style.Render(helpText)
}

// ParseHelpProps converts a generic props map to HelpProps using JSON unmarshaling
func ParseHelpProps(props map[string]interface{}) HelpProps {
	// Set defaults
	hp := HelpProps{
		Style:            "compact",
		ShowAllKeys:      false,
		SectionSeparator: "\n",
		ItemSeparator:    "\n",
	}

	// Handle key map
	if keyMap, ok := props["keyMap"].(map[string]interface{}); ok {
		hp.KeyMap = keyMap
	}

	// Handle sections
	if sections, ok := props["sections"].([]interface{}); ok {
		hp.Sections = parseHelpSections(sections)
	}

	// Handle padding
	if padding, ok := props["padding"].([]interface{}); ok {
		hp.Padding = make([]int, len(padding))
		for i, v := range padding {
			if intVal, ok := v.(int); ok {
				hp.Padding[i] = intVal
			} else if floatVal, ok := v.(float64); ok {
				hp.Padding[i] = int(floatVal)
			}
		}
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &hp)
	}

	return hp
}

// parseHelpSections parses sections from interface array
func parseHelpSections(sections []interface{}) []HelpSection {
	result := make([]HelpSection, 0, len(sections))
	for _, section := range sections {
		if sectionMap, ok := section.(map[string]interface{}); ok {
			hs := HelpSection{}

			// Parse title
			if title, ok := sectionMap["title"].(string); ok {
				hs.Title = title
			}

			// Parse items
			if items, ok := sectionMap["items"].([]interface{}); ok {
				hs.Items = parseHelpItems(items)
			}

			result = append(result, hs)
		}
	}
	return result
}

// parseHelpItems parses help items from interface array
func parseHelpItems(items []interface{}) []HelpItem {
	result := make([]HelpItem, 0, len(items))
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			hi := HelpItem{}
			if key, ok := itemMap["key"].(string); ok {
				hi.Key = key
			}
			if desc, ok := itemMap["description"].(string); ok {
				hi.Description = desc
			}
			result = append(result, hi)
		}
	}
	return result
}

// NewHelpModel creates a new HelpModel from HelpProps
func NewHelpModel(props HelpProps, id string) HelpModel {
	return HelpModel{
		Model: help.New(),
		props: props,
		id:    id,
	}
}

// HandleHelpUpdate handles updates for help components
func HandleHelpUpdate(msg tea.Msg, helpModel *HelpModel) (HelpModel, tea.Cmd) {
	if helpModel == nil {
		return HelpModel{}, nil
	}

	var cmd tea.Cmd
	helpModel.Model, cmd = helpModel.Model.Update(msg)
	return *helpModel, cmd
}

// Init initializes the help model
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the help
func (m *HelpModel) View() string {
	// Apply styles
	style := lipgloss.NewStyle()
	if m.props.Color != "" {
		style = style.Foreground(lipgloss.Color(m.props.Color))
	}
	if m.props.Background != "" {
		style = style.Background(lipgloss.Color(m.props.Background))
	}

	// Set width if specified
	if m.props.Width > 0 {
		style = style.Width(m.props.Width)
	}

	// Apply padding
	if len(m.props.Padding) > 0 {
		switch len(m.props.Padding) {
		case 1:
			style = style.Padding(m.props.Padding[0])
		case 2:
			style = style.Padding(m.props.Padding[0], m.props.Padding[1])
		case 4:
			style = style.Padding(m.props.Padding[0], m.props.Padding[1], m.props.Padding[2], m.props.Padding[3])
		}
	}

	// Add border if specified
	if m.props.Border {
		borderStyle := lipgloss.NewStyle()
		if m.props.BorderColor != "" {
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(m.props.BorderColor))
		}
		style = style.BorderStyle(lipgloss.NormalBorder()).Inherit(borderStyle)
	}

	// Create help text based on style
	// Auto-detect: if sections are provided, use sections style
	effectiveStyle := m.props.Style
	if effectiveStyle == "compact" && len(m.props.Sections) > 0 {
		effectiveStyle = "sections"
	}

	var helpText string
	switch effectiveStyle {
	case "sections":
		helpText = m.renderSections()
	case "full":
		helpText = "Navigation: ↑↓←→ • Select: Enter • Quit: Ctrl+C or Esc"
	case "minimal":
		helpText = "↑↓ Enter Esc"
	default: // "compact"
		helpText = "↑↓: Navigate • Enter: Select • Esc: Back"
	}

	return style.Render(helpText)
}

// renderSections renders help in sections format
func (m *HelpModel) renderSections() string {
	if len(m.props.Sections) == 0 {
		return "No help sections available"
	}

	var result string
	separator := m.props.SectionSeparator
	itemSeparator := m.props.ItemSeparator

	for i, section := range m.props.Sections {
		if i > 0 {
			result += separator
		}

		// Render section title with style
		titleStyle := lipgloss.NewStyle()
		if m.props.SectionTitleColor != "" {
			titleStyle = titleStyle.Foreground(lipgloss.Color(m.props.SectionTitleColor))
		}
		titleStyle = titleStyle.Bold(true)
		result += titleStyle.Render(section.Title)
		result += "\n"

		// Render items
		for j, item := range section.Items {
			if j > 0 {
				result += itemSeparator
			}
			result += fmt.Sprintf("  %-20s %s", item.Key, item.Description)
		}
	}

	return result
}

// GetID returns the unique identifier for this component instance
func (m *HelpModel) GetID() string {
	return m.id
}

// SetKeyMap sets the key bindings
func (m *HelpModel) SetKeyMap(keyMap map[string]interface{}) {
	m.props.KeyMap = keyMap
}

// HelpComponentWrapper wraps the native help.Model directly
type HelpComponentWrapper struct {
	model       help.Model
	props       HelpProps
	id          string
	focus       bool
	bindings    []core.ComponentBinding
	stateHelper *HelpStateHelper
}

// NewHelpComponentWrapper creates a wrapper that implements ComponentInterface
func NewHelpComponentWrapper(props HelpProps, id string) *HelpComponentWrapper {
	wrapper := &HelpComponentWrapper{
		model: help.New(),
		props: props,
		id:    id,
		bindings: props.Bindings,
	}
	wrapper.stateHelper = &HelpStateHelper{
		ComponentID: id,
	}
	return wrapper
}

func (w *HelpComponentWrapper) Init() tea.Cmd {
	return nil
}

// HelpStateHelper 帮助组件状态捕获助手
type HelpStateHelper struct {
	ComponentID string
}

func (h *HelpStateHelper) CaptureState() map[string]interface{} {
	return map[string]interface{}{}
}

func (h *HelpStateHelper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return []tea.Cmd{}
}

func (w *HelpComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
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
func (w *HelpComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *HelpComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

// 以下是 ComponentWrapper 接口的实现
func (w *HelpComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	// 帮助组件通常不需要执行动作
	return nil
}

func (w *HelpComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

func (w *HelpComponentWrapper) GetModel() interface{} {
	return w.model
}

func (w *HelpComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	// 对于帮助组件，只需更新模型
	w.model, cmd = w.model.Update(msg)
	return cmd
}

// 实现 StateCapturable 接口
func (w *HelpComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *HelpComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

// 实现 HandleSpecialKey 方法
func (w *HelpComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// 帮助组件可能有特殊键处理逻辑
	// 目前返回未处理，让框架继续处理
	return nil, core.Ignored, false
}

func (w *HelpComponentWrapper) View() string {
	// Apply styles
	style := lipgloss.NewStyle()
	if w.props.Color != "" {
		style = style.Foreground(lipgloss.Color(w.props.Color))
	}
	if w.props.Background != "" {
		style = style.Background(lipgloss.Color(w.props.Background))
	}

	// Set width if specified
	if w.props.Width > 0 {
		style = style.Width(w.props.Width)
	}

	// Apply padding
	if len(w.props.Padding) > 0 {
		switch len(w.props.Padding) {
		case 1:
			style = style.Padding(w.props.Padding[0])
		case 2:
			style = style.Padding(w.props.Padding[0], w.props.Padding[1])
		case 4:
			style = style.Padding(w.props.Padding[0], w.props.Padding[1], w.props.Padding[2], w.props.Padding[3])
		}
	}

	// Add border if specified
	if w.props.Border {
		borderStyle := lipgloss.NewStyle()
		if w.props.BorderColor != "" {
			borderStyle = borderStyle.BorderForeground(lipgloss.Color(w.props.BorderColor))
		}
		style = style.BorderStyle(lipgloss.NormalBorder()).Inherit(borderStyle)
	}

	// Create help text based on style
	// Auto-detect: if sections are provided, use sections style
	effectiveStyle := w.props.Style
	if effectiveStyle == "compact" && len(w.props.Sections) > 0 {
		effectiveStyle = "sections"
	}

	var helpText string
	switch effectiveStyle {
	case "sections":
		helpText = w.renderSections()
	case "full":
		helpText = "Navigation: ↑↓←→ • Select: Enter • Quit: Ctrl+C or Esc"
	case "minimal":
		helpText = "↑↓ Enter Esc"
	default: // "compact"
		helpText = "↑↓: Navigate • Enter: Select • Esc: Back"
	}

	return style.Render(helpText)
}

func (w *HelpComponentWrapper) GetID() string {
	return w.id
}

func (w *HelpComponentWrapper) SetFocus(focus bool) {
	w.focus = focus
}

func (w *HelpComponentWrapper) GetFocus() bool {
	return w.focus
}

// SetSize sets the allocated size for the component.
func (w *HelpComponentWrapper) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

func (w *HelpComponentWrapper) GetComponentType() string {
	return "help"
}

func (w *HelpComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("HelpComponentWrapper: invalid data type")
	}

	// Parse help properties
	props := ParseHelpProps(propsMap)

	// Update component properties
	w.props = props

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (w *HelpComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("HelpComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse help properties
	props := ParseHelpProps(propsMap)

	// Update component properties
	w.props = props

	return nil
}

func (w *HelpComponentWrapper) Cleanup() {
	// Help组件通常不需要清理资源
	// 这是一个空操作
}

// GetStateChanges returns the state changes from this component
func (w *HelpComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Static components don't have state changes
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *HelpComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}

// renderSections renders help in sections format
func (w *HelpComponentWrapper) renderSections() string {
	if len(w.props.Sections) == 0 {
		return "No help sections available"
	}

	var result string
	separator := w.props.SectionSeparator
	itemSeparator := w.props.ItemSeparator

	for i, section := range w.props.Sections {
		if i > 0 {
			result += separator
		}

		// Render section title with style
		titleStyle := lipgloss.NewStyle()
		if w.props.SectionTitleColor != "" {
			titleStyle = titleStyle.Foreground(lipgloss.Color(w.props.SectionTitleColor))
		}
		titleStyle = titleStyle.Bold(true)
		result += titleStyle.Render(section.Title)
		result += "\n"

		// Render items
		for j, item := range section.Items {
			if j > 0 {
				result += itemSeparator
			}
			result += fmt.Sprintf("  %-20s %s", item.Key, item.Description)
		}
	}

	return result
}

// SetKeyMap sets the key bindings
func (w *HelpComponentWrapper) SetKeyMap(keyMap map[string]interface{}) {
	w.props.KeyMap = keyMap
}
