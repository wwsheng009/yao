package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

// FlexComponent implements a flexbox layout as a TUI component.
type FlexComponent struct {
	ID         string
	Direction  core.Direction
	AlignItems core.Align
	Justify    core.Justify
	Gap        int
	Padding    *core.Padding
	Children   []*core.ComponentInstance
	Bounds     core.Rect
	Style      string
}

// FlexConfig holds configuration for a flex component.
type FlexConfig struct {
	ID         string
	Direction  core.Direction
	AlignItems core.Align
	Justify    core.Justify
	Gap        int
	Padding    *core.Padding
	Children   []*core.ComponentInstance
}

// NewFlexComponent creates a new flexbox layout component.
func NewFlexComponent(config *FlexConfig) *FlexComponent {
	return &FlexComponent{
		ID:         config.ID,
		Direction:  config.Direction,
		AlignItems: config.AlignItems,
		Justify:    config.Justify,
		Gap:        config.Gap,
		Padding:    config.Padding,
		Children:   config.Children,
	}
}

// Init implements core.ComponentInterface.
func (c *FlexComponent) Init() tea.Cmd {
	return nil
}

// UpdateMsg implements core.ComponentInterface.
func (c *FlexComponent) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return c, nil, core.Handled
}

// View implements core.ComponentInterface.
func (c *FlexComponent) View() string {
	if len(c.Children) == 0 {
		return strings.Repeat(" ", c.Bounds.Width)
	}

	var lines []string
	for _, child := range c.Children {
		if child.Instance != nil {
			rendered := child.Instance.View()
			if rendered != "" {
				lines = append(lines, splitLines(rendered)...)
			}
		}
	}

	if c.Direction == core.DirectionRow {
		return joinRowLayout(c, lines)
	}
	return joinColumnLayout(c, lines)
}

// Render implements core.ComponentInterface.
func (c *FlexComponent) Render(config core.RenderConfig) (string, error) {
	c.Bounds = core.Rect{
		X:      0,
		Y:      0,
		Width:  config.Width,
		Height: config.Height,
	}

	if len(c.Children) == 0 {
		return strings.Repeat(" ", config.Width), nil
	}

	var renderedChildren []string
	for _, child := range c.Children {
		if child.Instance != nil {
			rendered, err := child.Instance.Render(core.RenderConfig{
				Width:  c.calculateChildWidth(child, config.Width),
				Height: config.Height,
			})
			if err != nil {
				return "", err
			}
			if rendered != "" {
				renderedChildren = append(renderedChildren, rendered)
			}
		}
	}

	if c.Direction == core.DirectionRow {
		return joinRowLayout(c, renderedChildren), nil
	}
	return joinColumnLayout(c, renderedChildren), nil
}

// GetID implements core.ComponentInterface.
func (c *FlexComponent) GetID() string {
	return c.ID
}

// SetFocus implements core.ComponentInterface.
func (c *FlexComponent) SetFocus(focus bool) {}

// GetFocus implements core.ComponentInterface.
func (c *FlexComponent) GetFocus() bool {
	return false
}

// GetComponentType implements core.ComponentInterface.
func (c *FlexComponent) GetComponentType() string {
	return "flex"
}

// UpdateRenderConfig implements core.ComponentInterface.
func (c *FlexComponent) UpdateRenderConfig(config core.RenderConfig) error {
	c.Bounds = core.Rect{
		X:      0,
		Y:      0,
		Width:  config.Width,
		Height: config.Height,
	}
	return nil
}

// Cleanup implements core.ComponentInterface.
func (c *FlexComponent) Cleanup() {}

// GetStateChanges implements core.ComponentInterface.
func (c *FlexComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

// GetSubscribedMessageTypes implements core.ComponentInterface.
func (c *FlexComponent) GetSubscribedMessageTypes() []string {
	return nil
}

// calculateChildWidth calculates the width for a child in row layout.
func (c *FlexComponent) calculateChildWidth(child *core.ComponentInstance, parentWidth int) int {
	return parentWidth / len(c.Children)
}

// joinRowLayout joins children horizontally with alignment.
func joinRowLayout(c *FlexComponent, children []string) string {
	if len(children) == 0 {
		return ""
	}

	if c.Direction == core.DirectionRow {
		return strings.Join(children, strings.Repeat(" ", c.Gap))
	}
	return strings.Join(children, "\n")
}

// joinColumnLayout joins children vertically with alignment.
func joinColumnLayout(c *FlexComponent, children []string) string {
	if len(children) == 0 {
		return ""
	}

	result := strings.Join(children, "")
	return result
}

// splitLines splits a string by newline.
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
