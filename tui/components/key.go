package components

import (
	"encoding/json"
	"fmt"

	key "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// KeyBinding represents a single key binding with key and action
type KeyBinding struct {
	Key     string `json:"key"`
	Action  string `json:"action"`
	Enabled bool   `json:"enabled"`
}

// KeyProps defines the properties for the Key component
type KeyProps struct {
	// Keys specifies the key combinations (single key mode)
	Keys []string `json:"keys"`

	// Description specifies the key description (single key mode)
	Description string `json:"description"`

	// Bindings specifies multiple key bindings (batch mode)
	// When provided, this overrides Keys and Description
	Bindings []KeyBinding `json:"bindings"`

	// ShowLabels determines whether to show action labels (batch mode only)
	ShowLabels bool `json:"showLabels"`

	// Color specifies the text color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Bold makes the text bold
	Bold bool `json:"bold"`

	// Italic makes the text italic
	Italic bool `json:"italic"`

	// Underline underlines the text
	Underline bool `json:"underline"`

	// Width specifies the key width (0 for auto)
	Width int `json:"width"`

	// Height specifies the key height (0 for auto)
	Height int `json:"height"`

	// Enabled determines if the key binding is enabled (single key mode)
	Enabled bool `json:"enabled"`

	// Shortcut specifies the shortcut display (single key mode)
	Shortcut string `json:"shortcut"`

	// Spacing specifies spacing between bindings in batch mode (default: 2)
	Spacing int `json:"spacing"`
}

// KeyModel wraps the key.Binding to handle TUI integration
type KeyModel struct {
	key.Binding
	props KeyProps
	id    string // Unique identifier for this instance
}

// RenderKey renders a key component
func RenderKey(props KeyProps, width int) string {
	// Batch mode: render multiple bindings
	if len(props.Bindings) > 0 {
		return renderBatchBindings(props)
	}

	// Single key mode: render one binding
	return renderSingleKey(props, width)
}

// renderSingleKey renders a single key binding (original behavior)
func renderSingleKey(props KeyProps, width int) string {
	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}
	if props.Bold {
		style = style.Bold(true)
	}
	if props.Italic {
		style = style.Italic(true)
	}
	if props.Underline {
		style = style.Underline(true)
	}

	// Set width if specified
	if props.Width > 0 {
		style = style.Width(props.Width)
	} else if width > 0 {
		style = style.Width(width)
	}

	// Create key display
	display := props.Shortcut
	if display == "" && len(props.Keys) > 0 {
		display = props.Keys[0]
	}
	if props.Description != "" {
		display += " - " + props.Description
	}

	return style.Render(display)
}

// renderBatchBindings renders multiple key bindings in batch mode
func renderBatchBindings(props KeyProps) string {
	// Create base style
	baseStyle := lipgloss.NewStyle()
	if props.Color != "" {
		baseStyle = baseStyle.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		baseStyle = baseStyle.Background(lipgloss.Color(props.Background))
	}
	if props.Bold {
		baseStyle = baseStyle.Bold(true)
	}
	if props.Italic {
		baseStyle = baseStyle.Italic(true)
	}
	if props.Underline {
		baseStyle = baseStyle.Underline(true)
	}

	// Determine spacing
	spacing := props.Spacing
	if spacing <= 0 {
		spacing = 2 // Default spacing
	}

	// Render each binding
	var bindings []string
	for _, binding := range props.Bindings {
		display := binding.Key
		if props.ShowLabels && binding.Action != "" {
			display += " - " + binding.Action
		}
		bindings = append(bindings, display)
	}

	// Join with newlines
	return baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, bindings...))
}

// ParseKeyProps converts a generic props map to KeyProps using JSON unmarshaling
func ParseKeyProps(props map[string]interface{}) KeyProps {
	// Set defaults
	kp := KeyProps{
		Enabled:   true,
		Spacing:   2,
		ShowLabels: true,
	}

	// Handle keys (single key mode)
	if keys, ok := props["keys"].([]interface{}); ok {
		kp.Keys = make([]string, len(keys))
		for i, v := range keys {
			if str, ok := v.(string); ok {
				kp.Keys[i] = str
			}
		}
	}

	// Handle bindings (batch mode)
	if bindings, ok := props["bindings"].([]interface{}); ok {
		kp.Bindings = make([]KeyBinding, len(bindings))
		for i, v := range bindings {
			if bindingMap, ok := v.(map[string]interface{}); ok {
				// Marshal and unmarshal to populate KeyBinding struct
				if bytes, err := json.Marshal(bindingMap); err == nil {
					var kb KeyBinding
					if err := json.Unmarshal(bytes, &kb); err == nil {
						kp.Bindings[i] = kb
					}
				}
			}
		}
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &kp)
	}

	return kp
}

// NewKeyModel creates a new KeyModel from KeyProps
func NewKeyModel(props KeyProps, id string) KeyModel {
	kb := key.NewBinding(
		key.WithKeys(props.Keys...),
		key.WithHelp(props.Shortcut, props.Description),
	)

	return KeyModel{
		Binding: kb,
		props:   props,
		id:      id,
	}
}

// HandleKeyUpdate handles updates for key components
func HandleKeyUpdate(msg tea.Msg, keyModel *KeyModel) (KeyModel, tea.Cmd) {
	if keyModel == nil {
		return KeyModel{}, nil
	}

	return *keyModel, nil
}

// Init initializes the key model
func (m *KeyModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the key
func (m *KeyModel) View() string {
	// Batch mode: render multiple bindings
	if len(m.props.Bindings) > 0 {
		return m.renderBatchBindings()
	}

	// Single key mode: render one binding
	return m.renderSingleKey()
}

// renderSingleKey renders a single key binding (original behavior)
func (m *KeyModel) renderSingleKey() string {
	// Apply styles
	style := lipgloss.NewStyle()
	if m.props.Color != "" {
		style = style.Foreground(lipgloss.Color(m.props.Color))
	}
	if m.props.Background != "" {
		style = style.Background(lipgloss.Color(m.props.Background))
	}
	if m.props.Bold {
		style = style.Bold(true)
	}
	if m.props.Italic {
		style = style.Italic(true)
	}
	if m.props.Underline {
		style = style.Underline(true)
	}

	// Create key display
	display := m.props.Shortcut
	if display == "" && len(m.props.Keys) > 0 {
		display = m.props.Keys[0]
	}
	if m.props.Description != "" {
		display += " - " + m.props.Description
	}

	return style.Render(display)
}

// renderBatchBindings renders multiple key bindings in batch mode
func (m *KeyModel) renderBatchBindings() string {
	// Create base style
	baseStyle := lipgloss.NewStyle()
	if m.props.Color != "" {
		baseStyle = baseStyle.Foreground(lipgloss.Color(m.props.Color))
	}
	if m.props.Background != "" {
		baseStyle = baseStyle.Background(lipgloss.Color(m.props.Background))
	}
	if m.props.Bold {
		baseStyle = baseStyle.Bold(true)
	}
	if m.props.Italic {
		baseStyle = baseStyle.Italic(true)
	}
	if m.props.Underline {
		baseStyle = baseStyle.Underline(true)
	}

	// Determine spacing
	spacing := m.props.Spacing
	if spacing <= 0 {
		spacing = 2 // Default spacing
	}

	// Render each binding
	var bindings []string
	for _, binding := range m.props.Bindings {
		display := binding.Key
		if m.props.ShowLabels && binding.Action != "" {
			display += " - " + binding.Action
		}
		bindings = append(bindings, display)
	}

	// Join with newlines (vertical layout)
	return baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, bindings...))
}

// GetID returns the unique identifier for this component instance
func (m *KeyModel) GetID() string {
	return m.id
}

// SetEnabled sets the key enabled state
func (m *KeyModel) SetEnabled(enabled bool) {
	m.props.Enabled = enabled
}

// KeyComponentWrapper wraps the native key.Binding directly
type KeyComponentWrapper struct {
	binding key.Binding
	props   KeyProps
	id      string
}

// NewKeyComponentWrapper creates a wrapper that implements ComponentInterface
func NewKeyComponentWrapper(props KeyProps, id string) *KeyComponentWrapper {
	kb := key.NewBinding(
		key.WithKeys(props.Keys...),
		key.WithHelp(props.Shortcut, props.Description),
	)

	return &KeyComponentWrapper{
		binding: kb,
		props:   props,
		id:      id,
	}
}

func (w *KeyComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *KeyComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For key, check if this key was pressed
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, w.binding) {
			// Publish key pressed event
			eventCmd := core.PublishEvent(w.id, "KEY_PRESSED", map[string]interface{}{
				"keys":        w.props.Keys,
				"description": w.props.Description,
				"enabled":     w.props.Enabled,
			})
			return w, eventCmd, core.Handled
		}
	}

	return w, nil, core.Ignored
}

func (w *KeyComponentWrapper) View() string {
	// Batch mode: render multiple bindings
	if len(w.props.Bindings) > 0 {
		return w.renderBatchBindings()
	}

	// Single key mode: render one binding
	return w.renderSingleKey()
}

// renderSingleKey renders a single key binding (original behavior)
func (w *KeyComponentWrapper) renderSingleKey() string {
	// Apply styles
	style := lipgloss.NewStyle()
	if w.props.Color != "" {
		style = style.Foreground(lipgloss.Color(w.props.Color))
	}
	if w.props.Background != "" {
		style = style.Background(lipgloss.Color(w.props.Background))
	}
	if w.props.Bold {
		style = style.Bold(true)
	}
	if w.props.Italic {
		style = style.Italic(true)
	}
	if w.props.Underline {
		style = style.Underline(true)
	}

	// Create key display
	display := w.props.Shortcut
	if display == "" && len(w.props.Keys) > 0 {
		display = w.props.Keys[0]
	}
	if w.props.Description != "" {
		display += " - " + w.props.Description
	}

	return style.Render(display)
}

// renderBatchBindings renders multiple key bindings in batch mode
func (w *KeyComponentWrapper) renderBatchBindings() string {
	// Create base style
	baseStyle := lipgloss.NewStyle()
	if w.props.Color != "" {
		baseStyle = baseStyle.Foreground(lipgloss.Color(w.props.Color))
	}
	if w.props.Background != "" {
		baseStyle = baseStyle.Background(lipgloss.Color(w.props.Background))
	}
	if w.props.Bold {
		baseStyle = baseStyle.Bold(true)
	}
	if w.props.Italic {
		baseStyle = baseStyle.Italic(true)
	}
	if w.props.Underline {
		baseStyle = baseStyle.Underline(true)
	}

	// Determine spacing
	spacing := w.props.Spacing
	if spacing <= 0 {
		spacing = 2 // Default spacing
	}

	// Render each binding
	var bindings []string
	for _, binding := range w.props.Bindings {
		display := binding.Key
		if w.props.ShowLabels && binding.Action != "" {
			display += " - " + binding.Action
		}
		bindings = append(bindings, display)
	}

	// Join with newlines (vertical layout)
	return baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, bindings...))
}

func (w *KeyComponentWrapper) GetID() string {
	return w.id
}

func (w *KeyComponentWrapper) SetFocus(focus bool) {
	// Key doesn't have focus concept
}

func (w *KeyComponentWrapper) GetComponentType() string {
	return "key"
}

func (w *KeyComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("KeyComponentWrapper: invalid data type")
	}

	// Parse key properties
	props := ParseKeyProps(propsMap)

	// Update component properties
	w.props = props

	// Update key binding if needed
	w.binding = key.NewBinding(
		key.WithKeys(props.Keys...),
		key.WithHelp(props.Shortcut, props.Description),
	)

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (w *KeyComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("KeyComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse key properties
	props := ParseKeyProps(propsMap)

	// Update component properties
	w.props = props

	// Update key binding if needed
	w.binding = key.NewBinding(
		key.WithKeys(props.Keys...),
		key.WithHelp(props.Shortcut, props.Description),
	)

	return nil
}

func (w *KeyComponentWrapper) Cleanup() {
	// Key组件通常不需要清理资源
	// 这是一个空操作
}

// GetStateChanges returns the state changes from this component
func (w *KeyComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Static components don't have state changes
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *KeyComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"tea.KeyMsg",
		"core.TargetedMsg",
	}
}

// SetEnabled sets the key enabled state
func (w *KeyComponentWrapper) SetEnabled(enabled bool) {
	w.props.Enabled = enabled
}
