package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

type Renderer struct {
	engine *Engine
}

func NewRenderer(engine *Engine) *Renderer {
	return &Renderer{
		engine: engine,
	}
}

func (r *Renderer) Render() string {
	if r.engine.root == nil {
		return ""
	}

	// For Flex and Grid layouts, use the recursive line-based renderer
	// which preserves ANSI escape codes correctly.
	if r.engine.root.Type != LayoutAbsolute {
		// Ensure layout is calculated before rendering
		r.engine.Layout()
		return r.RenderNode(r.engine.root)
	}

	// For Absolute layouts, use the buffer approach
	result := r.engine.Layout()
	if len(result.Nodes) == 0 {
		return ""
	}

	return r.renderToBuffer(result)
}

func (r *Renderer) renderToBuffer(result *LayoutResult) string {
	if r.engine.root == nil {
		return ""
	}

	bound := r.engine.root.Bound
	if bound.Width == 0 || bound.Height == 0 {
		return ""
	}

	buffer := make([][]rune, bound.Height)
	for i := range buffer {
		buffer[i] = make([]rune, bound.Width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}

	for _, node := range result.Nodes {
		if node.Component != nil && node.Component.Instance != nil {
			config := core.RenderConfig{
				Width:  node.Bound.Width,
				Height: node.Bound.Height,
				Data:   node.Props,
			}
			content, err := node.Component.Instance.Render(config)
			if err != nil {
				content = err.Error()
			}

			lines := strings.Split(content, "\n")
			for i, line := range lines {
				targetY := node.Bound.Y + i
				if targetY >= 0 && targetY < bound.Height {
					runes := []rune(line)
					for j, ch := range runes {
						targetX := node.Bound.X + j
						if targetX >= 0 && targetX < bound.Width {
							buffer[targetY][targetX] = ch
						}
					}
				}
			}
		}
	}

	var builder strings.Builder
	for _, row := range buffer {
		builder.WriteString(string(row) + "\n")
	}

	if builder.Len() > 0 {
		return builder.String()[:builder.Len()-1]
	}
	return ""
}

func (r *Renderer) RenderNode(node *LayoutNode) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder

	style := r.createStyle(node)
	containerWidth := r.getWidth(node)
	containerHeight := r.getHeight(node)

	lines := r.renderNodeInternal(node, containerWidth, containerHeight)

	for i, line := range lines {
		styled := style.Width(containerWidth).Render(line)
		builder.WriteString(styled)
		if i < len(lines)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func (r *Renderer) renderNodeInternal(node *LayoutNode, width, height int) []string {
	if node == nil {
		return []string{}
	}

	if node.Component != nil && node.Component.Instance != nil {
		config := core.RenderConfig{
			Width:  width,
			Height: height,
			Data:   node.Props,
		}
		content, err := node.Component.Instance.Render(config)
		if err != nil {
			return []string{err.Error()}
		}
		return strings.Split(content, "\n")
	}

	var lines []string
	if r.isRow(node) {
		lines = r.renderRow(node, width, height)
	} else {
		lines = r.renderColumn(node, width, height)
	}

	return lines
}

func (r *Renderer) renderRow(node *LayoutNode, width, height int) []string {
	if len(node.Children) == 0 {
		line := strings.Repeat(" ", width)
		return []string{line}
	}

	result := r.engine.Layout()
	lines := make([]string, height)

	for i := range lines {
		var builder strings.Builder
		for _, child := range result.Nodes {
			if child.Parent == node {
				childLines := r.renderNodeInternal(child, child.Bound.Width, child.Bound.Height)
				if len(childLines) > i {
					builder.WriteString(childLines[i])
				} else {
					if len(childLines) > 0 {
						builder.WriteString(strings.Repeat(" ", child.Bound.Width))
					}
				}
			}
		}
		lines[i] = builder.String()
	}

	return lines
}

func (r *Renderer) renderColumn(node *LayoutNode, width, height int) []string {
	if len(node.Children) == 0 {
		line := strings.Repeat(" ", width)
		return []string{line}
	}

	var allLines []string

	result := r.engine.Layout()
	for _, child := range result.Nodes {
		if child.Parent == node {
			childLines := r.renderNodeInternal(child, child.Bound.Width, child.Bound.Height)
			allLines = append(allLines, childLines...)
		}
	}

	return allLines
}

func (r *Renderer) createStyle(node *LayoutNode) lipgloss.Style {
	var style lipgloss.Style

	if node.Style != nil {
		if node.Style.Padding != nil {
			style = style.
				Padding(node.Style.Padding.Top, node.Style.Padding.Right,
					node.Style.Padding.Bottom, node.Style.Padding.Left)
		}

		if node.Style.Margin != nil {
			style = style.
				Margin(node.Style.Margin.Top, node.Style.Margin.Right,
					node.Style.Margin.Bottom, node.Style.Margin.Left)
		}
	}

	return style
}

func (r *Renderer) getWidth(node *LayoutNode) int {
	if node.Style != nil && node.Style.Width != nil {
		switch v := node.Style.Width.Value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return node.Bound.Width
}

func (r *Renderer) getHeight(node *LayoutNode) int {
	if node.Style != nil && node.Style.Height != nil {
		switch v := node.Style.Height.Value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return node.Bound.Height
}

func (r *Renderer) isRow(node *LayoutNode) bool {
	if node.Style != nil {
		return node.Style.Direction == DirectionRow
	}
	return false
}
