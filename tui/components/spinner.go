package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// SpinnerProps defines the properties for the Spinner component
type SpinnerProps struct {
	// Style is the spinner style
	Style string `json:"style"`

	// Color specifies the spinner color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// Speed specifies the animation speed (milliseconds per frame)
	Speed int `json:"speed"`

	// Frames specifies custom frames for the spinner
	Frames []string `json:"frames"`

	// Width specifies the spinner width (0 for auto)
	Width int `json:"width"`

	// Height specifies the spinner height (0 for auto)
	Height int `json:"height"`

	// Label specifies the label text to display with the spinner
	Label string `json:"label"`

	// LabelPosition specifies where to position the label relative to the spinner
	LabelPosition string `json:"labelPosition"` // "left", "right", "top", "bottom"

	// Running determines if the spinner is animating
	Running bool `json:"running"`
}

// SpinnerModel wraps the spinner.Model to handle TUI integration
type SpinnerModel struct {
	spinner.Model
	props SpinnerProps
	id    string // Unique identifier for this instance
}

// RenderSpinner renders a spinner component
func RenderSpinner(props SpinnerProps, width int) string {
	s := spinner.New()

	// Set spinner style
	if props.Style != "" {
		s.Spinner = getSpinnerStyle(props.Style)
	}

	// Set custom frames if provided
	if len(props.Frames) > 0 {
		s.Spinner.Frames = props.Frames
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to spinner
	s.Style = s.Style.Inherit(style)

	// Set width if specified
	if props.Width > 0 {
		style = style.Width(props.Width)
	} else if width > 0 {
		style = style.Width(width)
	}

	// Handle label
	view := s.View()
	if props.Label != "" {
		switch props.LabelPosition {
		case "left":
			view = props.Label + " " + view
		case "top":
			view = props.Label + "\n" + view
		case "bottom":
			view = view + "\n" + props.Label
		default: // "right" or default
			view = view + " " + props.Label
		}
	}

	return style.Render(view)
}

// ParseSpinnerProps converts a generic props map to SpinnerProps using JSON unmarshaling
func ParseSpinnerProps(props map[string]interface{}) SpinnerProps {
	// Set defaults
	sp := SpinnerProps{
		Style:         "dots",
		Speed:         100,
		Running:       true,
		LabelPosition: "right",
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &sp)
	}

	return sp
}

// NewSpinnerModel creates a new SpinnerModel from SpinnerProps
func NewSpinnerModel(props SpinnerProps, id string) SpinnerModel {
	s := spinner.New()

	// Set spinner style
	if props.Style != "" {
		s.Spinner = getSpinnerStyle(props.Style)
	}

	// Set custom frames if provided
	if len(props.Frames) > 0 {
		s.Spinner.Frames = props.Frames
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to spinner
	s.Style = s.Style.Inherit(style)

	return SpinnerModel{
		Model: s,
		props: props,
		id:    id,
	}
}

// HandleSpinnerUpdate handles updates for spinner components
func HandleSpinnerUpdate(msg tea.Msg, spinnerModel *SpinnerModel) (SpinnerModel, tea.Cmd) {
	if spinnerModel == nil {
		return SpinnerModel{}, nil
	}

	var cmd tea.Cmd
	spinnerModel.Model, cmd = spinnerModel.Model.Update(msg)
	return *spinnerModel, cmd
}

// Init initializes the spinner model
func (m *SpinnerModel) Init() tea.Cmd {
	if m.props.Running {
		return func() tea.Msg {
			return m.Model.Tick()
		}
	}
	return nil
}

// View returns the string representation of the spinner
func (m *SpinnerModel) View() string {
	view := m.Model.View()

	// Handle label
	if m.props.Label != "" {
		switch m.props.LabelPosition {
		case "left":
			view = m.props.Label + " " + view
		case "top":
			view = m.props.Label + "\n" + view
		case "bottom":
			view = view + "\n" + m.props.Label
		default: // "right" or default
			view = view + " " + m.props.Label
		}
	}

	return view
}

// GetID returns the unique identifier for this component instance
func (m *SpinnerModel) GetID() string {
	return m.id
}

// SetRunning sets the spinner running state
func (m *SpinnerModel) SetRunning(running bool) {
	m.props.Running = running
}

// SpinnerComponentWrapper wraps the native spinner.Model directly
type SpinnerComponentWrapper struct {
	model spinner.Model
	props SpinnerProps
	id    string
	focus bool
}

// NewSpinnerComponentWrapper creates a wrapper that implements ComponentInterface
func NewSpinnerComponentWrapper(props SpinnerProps, id string) *SpinnerComponentWrapper {
	s := spinner.New()

	// Set spinner style
	if props.Style != "" {
		s.Spinner = getSpinnerStyle(props.Style)
	}

	// Set custom frames if provided
	if len(props.Frames) > 0 {
		s.Spinner.Frames = props.Frames
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to spinner
	s.Style = s.Style.Inherit(style)

	return &SpinnerComponentWrapper{
		model: s,
		props: props,
		id:    id,
	}
}

func (w *SpinnerComponentWrapper) Init() tea.Cmd {
	if w.props.Running {
		return func() tea.Msg {
			return w.model.Tick()
		}
	}
	return nil
}

func (w *SpinnerComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For spinner, just update the model
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)

	// Publish spinner tick event if running
	if w.props.Running {
		eventCmd := core.PublishEvent(w.id, "SPINNER_TICK", map[string]interface{}{
			"running": w.props.Running,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}

	return w, cmd, core.Handled
}

func (w *SpinnerComponentWrapper) View() string {
	view := w.model.View()

	// Handle label
	if w.props.Label != "" {
		switch w.props.LabelPosition {
		case "left":
			view = w.props.Label + " " + view
		case "top":
			view = w.props.Label + "\n" + view
		case "bottom":
			view = view + "\n" + w.props.Label
		default: // "right" or default
			view = view + " " + w.props.Label
		}
	}

	return view
}

func (w *SpinnerComponentWrapper) GetID() string {
	return w.id
}

func (w *SpinnerComponentWrapper) SetFocus(focus bool) {
	w.focus = focus
}

func (w *SpinnerComponentWrapper) GetFocus() bool {
	return w.focus
}

// SetSize sets the allocated size for the component.
func (w *SpinnerComponentWrapper) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

// getSpinnerStyle returns the spinner style based on the style name
func getSpinnerStyle(styleName string) spinner.Spinner {
	switch styleName {
	case "dots":
		return spinner.Dot
	case "line":
		return spinner.Line
	case "minidot":
		return spinner.MiniDot
	case "jump":
		return spinner.Jump
	case "pulse":
		return spinner.Pulse
	case "points":
		return spinner.Points
	case "globe":
		return spinner.Globe
	case "moon":
		return spinner.Moon
	case "monkey":
		return spinner.Monkey
	default:
		return spinner.Dot
	}
}

func (w *SpinnerComponentWrapper) GetComponentType() string {
	return "spinner"
}

func (w *SpinnerComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("SpinnerComponentWrapper: invalid data type")
	}

	// Parse spinner properties
	props := ParseSpinnerProps(propsMap)

	// Update component properties
	w.props = props

	// Update spinner style if changed
	if props.Style != "" {
		w.model.Spinner = getSpinnerStyle(props.Style)
	}

	// Update custom frames if provided
	if len(props.Frames) > 0 {
		w.model.Spinner.Frames = props.Frames
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to spinner
	w.model.Style = w.model.Style.Inherit(style)

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *SpinnerComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("SpinnerComponentWrapper: invalid data type")
	}

	// Parse spinner properties
	props := ParseSpinnerProps(propsMap)

	// Update component properties
	w.props = props

	// Update spinner style if changed
	if props.Style != "" {
		w.model.Spinner = getSpinnerStyle(props.Style)
	}

	// Update custom frames if provided
	if len(props.Frames) > 0 {
		w.model.Spinner.Frames = props.Frames
	}

	// Apply styles
	style := lipgloss.NewStyle()
	if props.Color != "" {
		style = style.Foreground(lipgloss.Color(props.Color))
	}
	if props.Background != "" {
		style = style.Background(lipgloss.Color(props.Background))
	}

	// Apply style to spinner
	w.model.Style = w.model.Style.Inherit(style)

	return nil
}

// Cleanup cleans up resources used by the spinner component
func (w *SpinnerComponentWrapper) Cleanup() {
	// Spinner components typically don't need cleanup
	// This is a no-op for spinner components
}

// GetStateChanges returns the state changes from this component
func (w *SpinnerComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Spinner component doesn't have meaningful state
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *SpinnerComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}
