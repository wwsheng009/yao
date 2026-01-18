package components

import (
	"encoding/json"
	"fmt"
	"time"

	timer "github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
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

// TimerComponentWrapper wraps TimerModel to implement ComponentInterface properly
type TimerComponentWrapper struct {
	model *TimerModel
}

// NewTimerComponentWrapper creates a wrapper that implements ComponentInterface
func NewTimerComponentWrapper(timerModel *TimerModel) *TimerComponentWrapper {
	return &TimerComponentWrapper{
		model: timerModel,
	}
}

func (w *TimerComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *TimerComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For timer, just update the model
	var cmd tea.Cmd
	w.model.Model, cmd = w.model.Model.Update(msg)

	// Publish timer tick event if running
	if w.model.props.Running {
		eventCmd := core.PublishEvent(w.model.id, "TIMER_TICK", map[string]interface{}{
			"running": w.model.props.Running,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}

	return w, cmd, core.Handled
}

func (w *TimerComponentWrapper) View() string {
	return w.model.View()
}

func (w *TimerComponentWrapper) GetID() string {
	return w.model.id
}

func (w *TimerComponentWrapper) SetFocus(focus bool) {
	// Timer doesn't have focus concept
}

func (w *TimerComponentWrapper) GetComponentType() string {
	return "timer"
}

func (m *TimerModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("TimerModel: invalid data type")
	}

	// Parse timer properties
	props := ParseTimerProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

func (w *TimerComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}
