package component

import (
	"encoding/json"
	"fmt"
	"time"

	timer "github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/tui/core"
)

// TimerProps defines the properties for the Timer component
type TimerProps struct {
	// Duration is the countdown duration
	Duration time.Duration `json:"duration"`

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

	// Width specifies the timer width (0 for auto)
	Width int `json:"width"`

	// Height specifies the timer height (0 for auto)
	Height int `json:"height"`

	// Running determines if the timer is running
	Running bool `json:"running"`

	// Repeat determines if the timer should repeat
	Repeat bool `json:"repeat"`

	// ShowElapsed shows elapsed time instead of remaining
	ShowElapsed bool `json:"showElapsed"`
}

// TimerModel wraps the timer.Model to handle TUI integration
type TimerModel struct {
	timer.Model
	props TimerProps
	id    string // Unique identifier for this instance
}

// RenderTimer renders a timer component
func RenderTimer(props TimerProps, width int) string {
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

	// Format the timer display
	format := props.Format
	if format == "" {
		format = "15:04:05"
	}

	var display string
	if props.ShowElapsed {
		// For elapsed time, show 00:00:00 as default
		display = "00:00:00"
	} else {
		// For countdown, show the duration
		display = props.Duration.String()
	}

	return style.Render(display)
}

// ParseTimerProps converts a generic props map to TimerProps using JSON unmarshaling
func ParseTimerProps(props map[string]interface{}) TimerProps {
	// Set defaults
	tp := TimerProps{
		Duration:    time.Minute, // 1 minute default
		Running:     true,
		Format:      "15:04:05",
		ShowElapsed: false,
	}

	// Handle duration separately as it might come as string
	if durationStr, ok := props["duration"].(string); ok {
		if dur, err := time.ParseDuration(durationStr); err == nil {
			tp.Duration = dur
		}
	} else if durationNum, ok := props["duration"].(float64); ok {
		tp.Duration = time.Duration(durationNum) * time.Second
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &tp)
	}

	return tp
}

// NewTimerModel creates a new TimerModel from TimerProps
func NewTimerModel(props TimerProps, id string) TimerModel {
	t := timer.NewWithInterval(props.Duration, time.Second)

	return TimerModel{
		Model: t,
		props: props,
		id:    id,
	}
}

// HandleTimerUpdate handles updates for timer components
func HandleTimerUpdate(msg tea.Msg, timerModel *TimerModel) (TimerModel, tea.Cmd) {
	if timerModel == nil {
		return TimerModel{}, nil
	}

	var cmd tea.Cmd
	timerModel.Model, cmd = timerModel.Model.Update(msg)
	return *timerModel, cmd
}

// Init initializes the timer model
func (m *TimerModel) Init() tea.Cmd {
	if m.props.Running {
		return m.Model.Init()
	}
	return nil
}

// View returns the string representation of the timer
func (m *TimerModel) View() string {
	// Format the timer display
	format := m.props.Format
	if format == "" {
		format = "15:04:05"
	}

	var display string
	if m.props.ShowElapsed {
		// For elapsed time, show 00:00:00 as default
		display = "00:00:00"
	} else {
		// For countdown, show the duration
		display = m.props.Duration.String()
	}

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
func (m *TimerModel) GetID() string {
	return m.id
}

// SetRunning sets the timer running state
func (m *TimerModel) SetRunning(running bool) {
	m.props.Running = running
}

// TimerComponentWrapper wraps the native timer.Model directly
type TimerComponentWrapper struct {
	model timer.Model
	props TimerProps
	id    string
	focus bool
}

// NewTimerComponentWrapper creates a wrapper that implements ComponentInterface
func NewTimerComponentWrapper(props TimerProps, id string) *TimerComponentWrapper {
	t := timer.NewWithInterval(props.Duration, time.Second)

	return &TimerComponentWrapper{
		model: t,
		props: props,
		id:    id,
	}
}

func (w *TimerComponentWrapper) Init() tea.Cmd {
	if w.props.Running {
		return w.model.Init()
	}
	return nil
}

func (w *TimerComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For timer, just update the model
	var cmd tea.Cmd
	w.model, cmd = w.model.Update(msg)

	// Publish timer tick event if running
	if w.props.Running {
		eventCmd := core.PublishEvent(w.id, "TIMER_TICK", map[string]interface{}{
			"running": w.props.Running,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}

	return w, cmd, core.Handled
}

func (w *TimerComponentWrapper) View() string {
	// Format the timer display
	format := w.props.Format
	if format == "" {
		format = "15:04:05"
	}

	var display string
	if w.props.ShowElapsed {
		// For elapsed time, show 00:00:00 as default
		display = "00:00:00"
	} else {
		// For countdown, show the duration
		display = w.props.Duration.String()
	}

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

	return style.Render(display)
}

func (w *TimerComponentWrapper) GetID() string {
	return w.id
}

func (w *TimerComponentWrapper) SetFocus(focus bool) {
	w.focus = focus
}

func (w *TimerComponentWrapper) GetFocus() bool {
	return w.focus
}

// SetSize sets the allocated size for the component.
func (w *TimerComponentWrapper) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

func (w *TimerComponentWrapper) GetComponentType() string {
	return "timer"
}

func (w *TimerComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TimerComponentWrapper: invalid data type")
	}

	// Parse timer properties
	props := ParseTimerProps(propsMap)

	// Update component properties
	w.props = props

	// Update timer with new duration if changed
	if props.Duration != w.props.Duration {
		w.model = timer.NewWithInterval(props.Duration, time.Second)
	}

	// Return rendered view
	return w.View(), nil
}

// UpdateRenderConfig 更新渲染配置
func (w *TimerComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("TimerComponentWrapper: invalid data type for UpdateRenderConfig")
	}

	// Parse timer properties
	props := ParseTimerProps(propsMap)

	// Update component properties
	w.props = props

	// Update timer with new duration if changed
	if props.Duration != w.props.Duration {
		w.model = timer.NewWithInterval(props.Duration, time.Second)
	}

	return nil
}

// Cleanup 清理资源
func (w *TimerComponentWrapper) Cleanup() {
	// 计时器组件通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (w *TimerComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// Timer component may have state
	return map[string]interface{}{
		w.GetID() + "_timeout": w.model.Timeout,
		w.GetID() + "_running": w.model.Running(),
	}, true
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *TimerComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}
