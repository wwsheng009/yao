package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

// ProgressProps defines the properties for the Progress component
type ProgressProps struct {
	// Percent is the progress percentage (0-100)
	Percent float64 `json:"percent"`

	// Width specifies the progress bar width (0 for auto)
	Width int `json:"width"`

	// Color specifies the progress bar color
	Color string `json:"color"`

	// Background specifies the background color
	Background string `json:"background"`

	// EmptyColor specifies the empty bar color
	EmptyColor string `json:"emptyColor"`

	// ShowPercentage shows/hides the percentage text
	ShowPercentage bool `json:"showPercentage"`

	// FilledChar specifies the character used for filled portion
	FilledChar string `json:"filledChar"`

	// EmptyChar specifies the character used for empty portion
	EmptyChar string `json:"emptyChar"`

	// Label specifies the label text
	Label string `json:"label"`

	// Animated determines if the progress bar is animated
	Animated bool `json:"animated"`
}

// ProgressModel wraps the progress.Model to handle TUI integration
type ProgressModel struct {
	progress.Model
	props ProgressProps
	id    string // Unique identifier for this instance
}

// RenderProgress renders a progress bar component
func RenderProgress(props ProgressProps, width int) string {
	p := progress.New()

	// Set width
	if props.Width > 0 {
		p.Width = props.Width
	} else if width > 0 {
		p.Width = width
	}

	// Set percentage
	if props.Percent < 0 {
		props.Percent = 0
	} else if props.Percent > 100 {
		props.Percent = 100
	}

	// Apply colors
	if props.Color != "" {
		p.FullColor = props.Color
	}

	if props.EmptyColor != "" {
		p.EmptyColor = props.EmptyColor
	}

	// Set custom characters
	if props.FilledChar != "" && len(props.FilledChar) > 0 {
		p.Full = rune(props.FilledChar[0])
	}

	if props.EmptyChar != "" && len(props.EmptyChar) > 0 {
		p.Empty = rune(props.EmptyChar[0])
	}

	// Show percentage
	p.ShowPercentage = props.ShowPercentage

	// Build view
	view := p.ViewAs(props.Percent / 100)

	// Add label if provided
	if props.Label != "" {
		view = props.Label + " " + view
	}

	return view
}

// ParseProgressProps converts a generic props map to ProgressProps using JSON unmarshaling
func ParseProgressProps(props map[string]interface{}) ProgressProps {
	// Set defaults
	pp := ProgressProps{
		Percent:        0,
		ShowPercentage: false,
		FilledChar:     "█",
		EmptyChar:      "░",
		Animated:       false,
	}

	// Unmarshal properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &pp)
	}

	return pp
}

// NewProgressModel creates a new ProgressModel from ProgressProps
func NewProgressModel(props ProgressProps, id string) ProgressModel {
	p := progress.New()

	// Set width
	if props.Width > 0 {
		p.Width = props.Width
	}

	// Apply colors
	if props.Color != "" {
		p.FullColor = props.Color
	}

	if props.EmptyColor != "" {
		p.EmptyColor = props.EmptyColor
	}

	// Set custom characters
	if props.FilledChar != "" && len(props.FilledChar) > 0 {
		p.Full = rune(props.FilledChar[0])
	}

	if props.EmptyChar != "" && len(props.EmptyChar) > 0 {
		p.Empty = rune(props.EmptyChar[0])
	}

	// Show percentage
	p.ShowPercentage = props.ShowPercentage

	return ProgressModel{
		Model: p,
		props: props,
		id:    id,
	}
}

// HandleProgressUpdate handles updates for progress components
func HandleProgressUpdate(msg tea.Msg, progressModel *ProgressModel) (ProgressModel, tea.Cmd) {
	if progressModel == nil {
		return ProgressModel{}, nil
	}

	var cmd tea.Cmd
	updatedModel, cmd := progressModel.Model.Update(msg)
	progressModel.Model = updatedModel.(progress.Model)
	return *progressModel, cmd
}

// Init initializes the progress model
func (m *ProgressModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the progress bar
func (m *ProgressModel) View() string {
	view := m.Model.ViewAs(m.props.Percent / 100)

	// Add label if provided
	if m.props.Label != "" {
		view = m.props.Label + " " + view
	}

	return view
}

// GetID returns the unique identifier for this component instance
func (m *ProgressModel) GetID() string {
	return m.id
}

// SetPercent sets the progress percentage
func (m *ProgressModel) SetPercent(percent float64) {
	if percent < 0 {
		m.props.Percent = 0
	} else if percent > 100 {
		m.props.Percent = 100
	} else {
		m.props.Percent = percent
	}
}

// ProgressComponentWrapper wraps the native progress.Model directly
type ProgressComponentWrapper struct {
	model progress.Model
	props ProgressProps
	id    string
}

// NewProgressComponentWrapper creates a wrapper that implements ComponentInterface
func NewProgressComponentWrapper(props ProgressProps, id string) *ProgressComponentWrapper {
	p := progress.New()

	// Set width
	if props.Width > 0 {
		p.Width = props.Width
	}

	// Apply colors
	if props.Color != "" {
		p.FullColor = props.Color
	}

	if props.EmptyColor != "" {
		p.EmptyColor = props.EmptyColor
	}

	// Set custom characters
	if props.FilledChar != "" && len(props.FilledChar) > 0 {
		p.Full = rune(props.FilledChar[0])
	}

	if props.EmptyChar != "" && len(props.EmptyChar) > 0 {
		p.Empty = rune(props.EmptyChar[0])
	}

	// Show percentage
	p.ShowPercentage = props.ShowPercentage

	return &ProgressComponentWrapper{
		model: p,
		props: props,
		id:    id,
	}
}

func (w *ProgressComponentWrapper) Init() tea.Cmd {
	return nil
}

func (w *ProgressComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case core.StateUpdateMsg:
		// Check if this is a progress update
		if msg.Key == "percent" || msg.Key == "progress" {
			if percent, ok := msg.Value.(float64); ok {
				w.SetPercent(percent)
				return w, nil, core.Handled
			}
		}
	}

	// For progress bar, just update the model
	var cmd tea.Cmd
	updatedModel, cmd := w.model.Update(msg)
	w.model = updatedModel.(progress.Model)
	return w, cmd, core.Handled
}

func (w *ProgressComponentWrapper) View() string {
	view := w.model.ViewAs(w.props.Percent / 100)

	// Add label if provided
	if w.props.Label != "" {
		view = w.props.Label + " " + view
	}

	return view
}

func (w *ProgressComponentWrapper) GetID() string {
	return w.id
}

func (w *ProgressComponentWrapper) SetFocus(focus bool) {
	// Progress doesn't have focus concept
}

// GetPercent returns the current progress percentage
func (w *ProgressComponentWrapper) GetPercent() float64 {
	return w.props.Percent
}

// SetPercent sets the progress percentage
func (w *ProgressComponentWrapper) SetPercent(percent float64) {
	if percent < 0 {
		w.props.Percent = 0
	} else if percent > 100 {
		w.props.Percent = 100
	} else {
		w.props.Percent = percent
	}
}

func (w *ProgressComponentWrapper) GetComponentType() string {
	return "progress"
}

func (w *ProgressComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ProgressComponentWrapper: invalid data type")
	}

	// Parse progress properties
	props := ParseProgressProps(propsMap)

	// Update component properties
	w.props = props

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig updates the render configuration without recreating the instance
func (w *ProgressComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ProgressComponentWrapper: invalid data type")
	}

	// Parse progress properties
	props := ParseProgressProps(propsMap)

	// Update component properties
	w.props = props

	// Update width if provided
	if props.Width > 0 {
		w.model.Width = props.Width
	}

	// Update colors if provided
	if props.Color != "" {
		w.model.FullColor = props.Color
	}
	if props.EmptyColor != "" {
		w.model.EmptyColor = props.EmptyColor
	}

	return nil
}

// Cleanup cleans up resources used by the progress component
func (w *ProgressComponentWrapper) Cleanup() {
	// Progress components typically don't need cleanup
	// This is a no-op for progress components
}

// GetStateChanges returns the state changes from this component
func (w *ProgressComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Progress component may have state
	return map[string]interface{}{
		w.GetID() + "_percent": w.props.Percent,
		w.GetID() + "_value":   w.View(),
	}, true
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *ProgressComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
		"core.StateUpdateMsg",
	}
}
