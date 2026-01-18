package components

import (
	"encoding/json"
	"fmt"
	"time"

	stopwatch "github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// StopwatchProps defines the properties for the Stopwatch component
type StopwatchProps struct {
	// Format specifies the time format string
	Format string `json:"format"`

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

	// Width specifies the stopwatch width (0 for auto)
	Width int `json:"width"`

	// Height specifies the stopwatch height (0 for auto)
	Height int `json:"height"`

	// Running determines if the stopwatch is running
	Running bool `json:"running"`

	// Interval specifies the update interval
	Interval time.Duration `json:"interval"`

	// ShowMilliseconds shows milliseconds in display
	ShowMilliseconds bool `json:"showMilliseconds"`
}

// StopwatchModel wraps the stopwatch.Model to handle TUI integration
type StopwatchModel struct {
	stopwatch.Model
	props StopwatchProps
	id    string // Unique identifier for this instance
}

// RenderStopwatch renders a stopwatch component
func RenderStopwatch(props StopwatchProps, width int) string {
	sw := stopwatch.NewWithInterval(props.Interval)

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

	// Format the stopwatch display
	format := props.Format
	if format == "" {
		if props.ShowMilliseconds {
			format = "15:04:05.000"
		} else {
			format = "15:04:05"
		}
	}

	// Create a mock time for display
	mockTime := time.Unix(int64(sw.Elapsed().Seconds()), 0).UTC()
	display := mockTime.Format(format)

	return style.Render(display)
}

// ParseStopwatchProps converts a generic props map to StopwatchProps using JSON unmarshaling
func ParseStopwatchProps(props map[string]interface{}) StopwatchProps {
	// Set defaults
	sp := StopwatchProps{
		Running:          true,
		Interval:         time.Second,
		Format:           "15:04:05",
		ShowMilliseconds: false,
	}

	// Handle interval separately as it might come as string
	if intervalStr, ok := props["interval"].(string); ok {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			sp.Interval = interval
		}
	} else if intervalNum, ok := props["interval"].(float64); ok {
		sp.Interval = time.Duration(intervalNum) * time.Second
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &sp)
	}

	return sp
}

// NewStopwatchModel creates a new StopwatchModel from StopwatchProps
func NewStopwatchModel(props StopwatchProps, id string) StopwatchModel {
	sw := stopwatch.NewWithInterval(props.Interval)

	return StopwatchModel{
		Model: sw,
		props: props,
		id:    id,
	}
}

// HandleStopwatchUpdate handles updates for stopwatch components
func HandleStopwatchUpdate(msg tea.Msg, stopwatchModel *StopwatchModel) (StopwatchModel, tea.Cmd) {
	if stopwatchModel == nil {
		return StopwatchModel{}, nil
	}

	var cmd tea.Cmd
	stopwatchModel.Model, cmd = stopwatchModel.Model.Update(msg)
	return *stopwatchModel, cmd
}

// Init initializes the stopwatch model
func (m *StopwatchModel) Init() tea.Cmd {
	if m.props.Running {
		return m.Model.Init()
	}
	return nil
}

// View returns the string representation of the stopwatch
func (m *StopwatchModel) View() string {
	// Format the stopwatch display
	format := m.props.Format
	if format == "" {
		if m.props.ShowMilliseconds {
			format = "15:04:05.000"
		} else {
			format = "15:04:05"
		}
	}

	// Convert elapsed time to time.Time for formatting
	elapsed := m.Elapsed()
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	milliseconds := int(elapsed.Milliseconds()) % 1000

	// Create a mock time for display
	mockTime := time.Date(0, 1, 1, hours, minutes, seconds, milliseconds*int(time.Millisecond), time.UTC)
	display := mockTime.Format(format)

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

	return style.Render(display)
}

// GetID returns the unique identifier for this component instance
func (m *StopwatchModel) GetID() string {
	return m.id
}

// SetRunning sets the stopwatch running state
func (m *StopwatchModel) SetRunning(running bool) {
	m.props.Running = running
}

// StopwatchComponentWrapper wraps StopwatchModel to implement ComponentInterface properly
type StopwatchComponentWrapper struct {
	model *StopwatchModel
}

// NewStopwatchComponentWrapper creates a wrapper that implements ComponentInterface
func NewStopwatchComponentWrapper(stopwatchModel *StopwatchModel) *StopwatchComponentWrapper {
	return &StopwatchComponentWrapper{
		model: stopwatchModel,
	}
}

func (w *StopwatchComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *StopwatchComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For stopwatch, just update the model
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// Publish stopwatch tick event if running
	if w.model.props.Running {
		eventCmd := core.PublishEvent(w.model.id, "STOPWATCH_TICK", map[string]interface{}{
			"elapsed": w.model.Elapsed().String(),
			"running": w.model.props.Running,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}

	return w, cmd, core.Handled
}

func (w *StopwatchComponentWrapper) View() string {
	return w.model.View()
}

func (w *StopwatchComponentWrapper) GetID() string {
	return w.model.id
}

func (w *StopwatchComponentWrapper) SetFocus(focus bool) {
	// Stopwatch doesn't have focus concept
}

func (w *StopwatchComponentWrapper) GetComponentType() string {
	return "stopwatch"
}

func (m *StopwatchModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("StopwatchModel: invalid data type")
	}

	// Parse stopwatch properties
	props := ParseStopwatchProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

func (w *StopwatchComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

// UpdateRenderConfig 更新渲染配置
func (w *StopwatchComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("StopwatchComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse stopwatch properties
	props := ParseStopwatchProps(propsMap)

	// Update component properties
	w.model.props = props

	return nil
}

// Cleanup 清理资源
func (w *StopwatchComponentWrapper) Cleanup() {
	// 停止秒表并清理相关资源
	w.model.Model.Stop()
}
