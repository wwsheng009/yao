package components

import (
	"encoding/json"
	"fmt"

	key "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// KeyProps defines the properties for the Key component
type KeyProps struct {
	// Keys specifies the key combinations
	Keys []string `json:"keys"`

	// Description specifies the key description
	Description string `json:"description"`

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

	// Enabled determines if the key binding is enabled
	Enabled bool `json:"enabled"`

	// Shortcut specifies the shortcut display
	Shortcut string `json:"shortcut"`
}

// KeyModel wraps the key.Binding to handle TUI integration
type KeyModel struct {
	key.Binding
	props KeyProps
	id    string // Unique identifier for this instance
}

// RenderKey renders a key component
func RenderKey(props KeyProps, width int) string {
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

// ParseKeyProps converts a generic props map to KeyProps using JSON unmarshaling
func ParseKeyProps(props map[string]interface{}) KeyProps {
	// Set defaults
	kp := KeyProps{
		Enabled: true,
	}

	// Handle keys
	if keys, ok := props["keys"].([]interface{}); ok {
		kp.Keys = make([]string, len(keys))
		for i, v := range keys {
			if str, ok := v.(string); ok {
				kp.Keys[i] = str
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

// GetID returns the unique identifier for this component instance
func (m *KeyModel) GetID() string {
	return m.id
}

// SetEnabled sets the key enabled state
func (m *KeyModel) SetEnabled(enabled bool) {
	m.props.Enabled = enabled
}

// KeyComponentWrapper wraps KeyModel to implement ComponentInterface properly
type KeyComponentWrapper struct {
	model *KeyModel
}

// NewKeyComponentWrapper creates a wrapper that implements ComponentInterface
func NewKeyComponentWrapper(keyModel *KeyModel) *KeyComponentWrapper {
	return &KeyComponentWrapper{
		model: keyModel,
	}
}

func (w *KeyComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *KeyComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For key, check if this key was pressed
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, w.model.Binding) {
			// Publish key pressed event
			eventCmd := core.PublishEvent(w.model.id, "KEY_PRESSED", map[string]interface{}{
				"keys":        w.model.props.Keys,
				"description": w.model.props.Description,
				"enabled":     w.model.props.Enabled,
			})
			return w, eventCmd, core.Handled
		}
	}

	return w, nil, core.Ignored
}

func (w *KeyComponentWrapper) View() string {
	return w.model.View()
}

func (w *KeyComponentWrapper) GetID() string {
	return w.model.id
}

func (w *KeyComponentWrapper) SetFocus(focus bool) {
	// Key doesn't have focus concept
}

func (w *KeyComponentWrapper) GetComponentType() string {
	return "key"
}

func (m *KeyModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("KeyModel: invalid data type")
	}

	// Parse key properties
	props := ParseKeyProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

func (w *KeyComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}
