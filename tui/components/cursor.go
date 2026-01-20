package components

import (
	"encoding/json"
	"fmt"
	"time"

	cursor "github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// CursorProps defines the properties for the Cursor component
type CursorProps struct {
	// Style specifies the cursor style: "blink", "static", "hide"
	Style string `json:"style"`

	// Color specifies the cursor color
	Color string `json:"color"`

	// Background specifies the cursor background color
	Background string `json:"background"`

	// Blink makes the cursor blink
	Blink bool `json:"blink"`

	// BlinkSpeed specifies the blink speed in milliseconds
	BlinkSpeed int `json:"blinkSpeed"`

	// Width specifies the cursor width
	Width int `json:"width"`

	// Height specifies the cursor height
	Height int `json:"height"`

	// Position specifies the cursor position
	Position int `json:"position"`

	// Visible determines if the cursor is visible
	Visible bool `json:"visible"`

	// Char specifies the cursor character
	Char string `json:"char"`

	// Bindings define custom key bindings for the component (optional)
	Bindings []core.ComponentBinding `json:"bindings,omitempty"`
}

// CursorConfig is a simplified configuration for CursorHelper
type CursorConfig struct {
	Mode       cursor.Mode
	Char       string
	BlinkSpeed time.Duration
	Visible    bool
}

// CursorHelper is a utility helper for managing cursor behavior
// This is intended to be used internally by input components, not as a standalone component
type CursorHelper struct {
	config CursorConfig
	model  cursor.Model
}

// NewCursorHelper creates a new CursorHelper with the given configuration
func NewCursorHelper(config CursorConfig) *CursorHelper {
	c := cursor.New()
	if config.Mode != 0 {
		c.SetMode(config.Mode)
	}
	if config.Char != "" {
		c.SetChar(config.Char)
	}
	if config.BlinkSpeed > 0 {
		c.BlinkSpeed = config.BlinkSpeed
	}

	return &CursorHelper{
		config: config,
		model:  c,
	}
}

// GetModel returns the underlying cursor.Model
func (h *CursorHelper) GetModel() *cursor.Model {
	return &h.model
}

// GetChar returns the cursor character to display
func (h *CursorHelper) GetChar() string {
	if !h.config.Visible || h.config.Mode == cursor.CursorHide {
		return ""
	}
	if h.config.Char != "" {
		return h.config.Char
	}
	switch h.config.Mode {
	case cursor.CursorStatic:
		return "█"
	default:
		return "|"
	}
}

// SetMode sets the cursor mode
func (h *CursorHelper) SetMode(mode cursor.Mode) {
	h.model.SetMode(mode)
	h.config.Mode = mode
}

// SetChar sets the cursor character
func (h *CursorHelper) SetChar(char string) {
	h.model.SetChar(char)
	h.config.Char = char
}

// SetVisible sets cursor visibility
func (h *CursorHelper) SetVisible(visible bool) {
	h.config.Visible = visible
	if !visible {
		h.model.SetMode(cursor.CursorHide)
	} else if h.config.Mode != cursor.CursorHide {
		h.model.SetMode(h.config.Mode)
	}
}

// GetVisible returns current visibility state
func (h *CursorHelper) GetVisible() bool {
	return h.config.Visible
}

// GetMode returns current cursor mode
func (h *CursorHelper) GetMode() cursor.Mode {
	return h.config.Mode
}

// GetBlinkSpeed returns current blink speed
func (h *CursorHelper) GetBlinkSpeed() time.Duration {
	return h.config.BlinkSpeed
}

// GetOriginalChar returns the configured cursor character
func (h *CursorHelper) GetOriginalChar() string {
	return h.config.Char
}

// SetBlinkSpeed sets the blink speed
func (h *CursorHelper) SetBlinkSpeed(speed time.Duration) {
	h.config.BlinkSpeed = speed
	h.model.BlinkSpeed = speed
}

// Update updates the cursor model
func (h *CursorHelper) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	h.model, cmd = h.model.Update(msg)
	return cmd
}

// CursorModel wraps the cursor.Model to handle TUI integration
// DEPRECATED: Use CursorHelper instead for internal cursor management
type CursorModel struct {
	cursor.Model
	props CursorProps
	id    string
}

// RenderCursor renders a cursor component
// DEPRECATED: Use CursorHelper for dynamic cursor management
func RenderCursor(props CursorProps, width int) string {
	c := cursor.New()

	if props.Style != "" {
		c.SetMode(getCursorMode(props.Style))
	}

	if props.BlinkSpeed > 0 {
		c.BlinkSpeed = time.Duration(props.BlinkSpeed) * time.Millisecond
	}

	if props.Char != "" {
		c.SetChar(props.Char)
	}

	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	if props.Width > 0 {
		style = style.Width(props.Width)
	} else if width > 0 {
		style = style.Width(width)
	}

	cursorChar := getCursorChar(props.Style, props.Char)

	return style.Render(cursorChar)
}

// ParseCursorProps converts a generic props map to CursorProps using JSON unmarshaling
func ParseCursorProps(props map[string]interface{}) CursorProps {
	cp := CursorProps{
		Style:      "blink",
		Blink:      true,
		BlinkSpeed: 530,
		Visible:    true,
		Char:       "|",
	}

	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &cp)
	}

	return cp
}

// NewCursorModel creates a new CursorModel from CursorProps
// DEPRECATED: Use CursorHelper instead
func NewCursorModel(props CursorProps, id string) CursorModel {
	c := cursor.New()

	if props.Style != "" {
		c.SetMode(getCursorMode(props.Style))
	}

	if props.BlinkSpeed > 0 {
		c.BlinkSpeed = time.Duration(props.BlinkSpeed) * time.Millisecond
	}

	if props.Char != "" {
		c.SetChar(props.Char)
	}

	return CursorModel{
		Model: c,
		props: props,
		id:    id,
	}
}

// HandleCursorUpdate handles updates for cursor components
// DEPRECATED: Use CursorHelper.Update(msg) directly
func HandleCursorUpdate(msg tea.Msg, cursorModel *CursorModel) (CursorModel, tea.Cmd) {
	if cursorModel == nil {
		return CursorModel{}, nil
	}

	var cmd tea.Cmd
	cursorModel.Model, cmd = cursorModel.Model.Update(msg)
	return *cursorModel, cmd
}

func (m *CursorModel) View() string {
	if !m.props.Visible {
		return ""
	}

	style := lipgloss.NewStyle()
	if m.props.Color != "" {
		style = style.Foreground(lipgloss.Color(m.props.Color))
	}
	if m.props.Background != "" {
		style = style.Background(lipgloss.Color(m.props.Background))
	}

	cursorChar := getCursorChar(m.props.Style, m.props.Char)

	return style.Render(cursorChar)
}

func (m *CursorModel) GetID() string {
	return m.id
}

func (m *CursorModel) SetPosition(pos int) {
	m.props.Position = pos
}

func (m *CursorModel) SetVisible(visible bool) {
	m.props.Visible = visible
}

// CursorComponentWrapper wraps cursor functionality as a simplified component
// This uses CursorHelper internally for better cursor management
type CursorComponentWrapper struct {
	helper      *CursorHelper
	props       CursorProps
	id          string
	bindings    []core.ComponentBinding
	stateHelper *core.InputStateHelper
	focus       bool
}

// NewCursorComponentWrapper creates a wrapper that implements ComponentInterface
func NewCursorComponentWrapper(props CursorProps, id string) *CursorComponentWrapper {
	config := CursorConfig{
		Mode:       getCursorMode(props.Style),
		Char:       getCursorModeChar(props.Style, props.Char),
		BlinkSpeed: time.Duration(props.BlinkSpeed) * time.Millisecond,
		Visible:    props.Visible,
	}

	wrapper := &CursorComponentWrapper{
		helper:   NewCursorHelper(config),
		props:    props,
		id:       id,
		bindings: props.Bindings,
	}

	wrapper.stateHelper = &core.InputStateHelper{
		Valuer:      wrapper,
		Focuser:     wrapper,
		ComponentID: id,
	}

	return wrapper
}

func (w *CursorComponentWrapper) Init() tea.Cmd {
	if w.props.Blink && w.props.Style != "hide" {
		return w.helper.GetModel().BlinkCmd()
	}
	return nil
}

func (w *CursorComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	cmd, response := core.DefaultInteractiveUpdateMsg(
		w,
		msg,
		w.getBindings,
		w.handleBinding,
		w.delegateToBubbles,
	)

	return w, cmd, response
}

func (w *CursorComponentWrapper) getBindings() []core.ComponentBinding {
	return w.bindings
}

func (w *CursorComponentWrapper) handleBinding(keyMsg tea.KeyMsg, binding core.ComponentBinding) (tea.Cmd, core.Response, bool) {
	cmd, response, handled := core.HandleBinding(w, keyMsg, binding)
	return cmd, response, handled
}

func (w *CursorComponentWrapper) delegateToBubbles(msg tea.Msg) tea.Cmd {
	return w.helper.Update(msg)
}

func (w *CursorComponentWrapper) CaptureState() map[string]interface{} {
	return w.stateHelper.CaptureState()
}

func (w *CursorComponentWrapper) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return w.stateHelper.DetectStateChanges(old, new)
}

func (w *CursorComponentWrapper) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, core.Response, bool) {
	// ESC 和 Tab 现在由框架层统一处理，这里不处理
	// 如果有其他特殊的键处理需求，可以在这里添加
	return nil, core.Ignored, false
}

func (w *CursorComponentWrapper) View() string {
	if !w.props.Visible {
		return ""
	}

	style := lipgloss.NewStyle()
	if w.props.Color != "" {
		style = style.Foreground(lipgloss.Color(w.props.Color))
	}
	if w.props.Background != "" {
		style = style.Background(lipgloss.Color(w.props.Background))
	}

	cursorChar := getCursorChar(w.props.Style, w.props.Char)

	return style.Render(cursorChar)
}

func (w *CursorComponentWrapper) GetID() string {
	return w.id
}

func (w *CursorComponentWrapper) GetModel() interface{} {
	return w.helper
}

func (w *CursorComponentWrapper) SetFocus(focus bool) {
	w.focus = focus
}

func (w *CursorComponentWrapper) GetFocus() bool {
	return w.focus
}

func (w *CursorComponentWrapper) PublishEvent(sourceID, eventName string, payload map[string]interface{}) tea.Cmd {
	return core.PublishEvent(sourceID, eventName, payload)
}

func (w *CursorComponentWrapper) GetValue() string {
	return ""
}

func (w *CursorComponentWrapper) Focused() bool {
	return w.props.Visible
}

func (w *CursorComponentWrapper) GetComponentType() string {
	return "cursor"
}

func (w *CursorComponentWrapper) ExecuteAction(action *core.Action) tea.Cmd {
	return func() tea.Msg {
		return core.ExecuteActionMsg{
			Action:    action,
			SourceID:  w.id,
			Timestamp: time.Now(),
		}
	}
}

func (w *CursorComponentWrapper) Render(config core.RenderConfig) (string, error) {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("CursorComponentWrapper: invalid data type")
	}

	props := ParseCursorProps(propsMap)
	w.props = props

	if props.Style != "" {
		w.helper.SetMode(getCursorMode(props.Style))
	}

	if props.BlinkSpeed > 0 {
		w.helper.SetBlinkSpeed(time.Duration(props.BlinkSpeed) * time.Millisecond)
	}

	if props.Char != "" {
		w.helper.SetChar(props.Char)
	}

	return w.View(), nil
}

func (w *CursorComponentWrapper) SetBlinkSpeed(speed time.Duration) {
	w.helper.SetBlinkSpeed(speed)
}

func (w *CursorComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("CursorComponentWrapper: invalid data type")
	}

	props := ParseCursorProps(propsMap)
	w.props = props

	if props.Style != "" {
		w.helper.SetMode(getCursorMode(props.Style))
	}

	if props.BlinkSpeed > 0 {
		w.helper.SetBlinkSpeed(time.Duration(props.BlinkSpeed) * time.Millisecond)
	}

	if props.Char != "" {
		w.helper.SetChar(props.Char)
	}

	if props.Blink && props.Style != "hide" {
	}

	return nil
}

func (w *CursorComponentWrapper) Cleanup() {
}

func (w *CursorComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (w *CursorComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}

func getCursorMode(style string) cursor.Mode {
	switch style {
	case "static":
		return cursor.CursorStatic
	case "hide":
		return cursor.CursorHide
	case "block":
		return cursor.CursorStatic
	default:
		return cursor.CursorBlink
	}
}

// ParseCursorMode converts string to cursor.Mode
func ParseCursorMode(style string) cursor.Mode {
	return getCursorMode(style)
}

func getCursorChar(style, customChar string) string {
	switch style {
	case "static", "block":
		return customChar
	case "hide":
		return ""
	case "blink":
		return customChar
	default:
		return "|"
	}
}

func getCursorModeChar(style, customChar string) string {
	switch style {
	case "static", "block":
		if customChar != "" {
			return customChar
		}
		return "█"
	case "hide":
		return ""
	case "blink":
		if customChar != "" {
			return customChar
		}
		return "|"
	default:
		return "|"
	}
}
