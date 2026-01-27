package component

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/tui/core"
)

// StaticComponent 静态组件实现
// 用于无状态的简单组件，如header、text、footer等
type StaticComponent struct {
	ID         string
	Props      map[string]interface{}
	Width      int
	Type       string
	RenderFunc func(props map[string]interface{}, width int) string
}

// NewStaticComponent creates a new StaticComponent
func NewStaticComponent(renderFunc func(props map[string]interface{}, width int) string,
	props map[string]interface{}, id string, width int, compType string) *StaticComponent {
	return &StaticComponent{
		ID:         id,
		Props:      props,
		Width:      width,
		Type:       compType,
		RenderFunc: renderFunc,
	}
}

func (c *StaticComponent) Init() tea.Cmd {
	return nil
}

func (c *StaticComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	switch msg := msg.(type) {
	case core.TargetedMsg:
		if msg.TargetID == c.ID {
			return c, nil, core.Handled
		}
	}
	return c, nil, core.Ignored
}

func (c *StaticComponent) View() string {
	return c.RenderFunc(c.Props, c.Width)
}

func (c *StaticComponent) GetID() string {
	return c.ID
}

func (c *StaticComponent) SetFocus(focus bool) {
	// 静态组件不支持焦点
}

func (c *StaticComponent) GetFocus() bool {
	// 静态组件始终没有焦点
	return false
}

// SetSize sets the allocated size for the component.
func (c *StaticComponent) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

func (c *StaticComponent) GetComponentType() string {
	return c.Type
}

func (c *StaticComponent) Render(config core.RenderConfig) (string, error) {
	// 更新宽度配置
	if config.Width > 0 {
		c.Width = config.Width
	}

	// 更新属性
	if config.Data != nil {
		if props, ok := config.Data.(map[string]interface{}); ok {
			c.Props = props
		} else {
			return "", fmt.Errorf("invalid data type for StaticComponent: expected map[string]interface{}, got %T", config.Data)
		}
	}

	return c.View(), nil
}

// SetProps 设置组件属性
func (c *StaticComponent) SetProps(props map[string]interface{}) {
	c.Props = props
}

// SetWidth 设置组件宽度
func (c *StaticComponent) SetWidth(width int) {
	c.Width = width
}

// SetID 设置组件ID
func (c *StaticComponent) SetID(id string) {
	c.ID = id
}

// UpdateRenderConfig 更新渲染配置
func (c *StaticComponent) UpdateRenderConfig(config core.RenderConfig) error {
	// 更新宽度配置
	if config.Width > 0 {
		c.Width = config.Width
	}

	// 更新属性
	if config.Data != nil {
		if props, ok := config.Data.(map[string]interface{}); ok {
			c.Props = props
		} else {
			return fmt.Errorf("invalid data type for StaticComponent: expected map[string]interface{}, got %T", config.Data)
		}
	}

	return nil
}

// Cleanup 清理资源
func (c *StaticComponent) Cleanup() {
	// 静态组件通常不需要特殊清理操作
}

// GetStateChanges returns the state changes from this component
func (c *StaticComponent) GetStateChanges() (map[string]interface{}, bool) {
	// Static components don't have state changes
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (c *StaticComponent) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}
