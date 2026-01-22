package layout

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// RenderContext provides all data and callbacks needed for rendering
// This interface allows the Renderer to be independent of Model implementation
type RenderContext interface {
	GetComponentInstance(id string) (*core.ComponentInstance, bool)
	ResolveProps(compID string) (map[string]interface{}, error)
	UpdateComponentConfig(instance *core.ComponentInstance, config core.RenderConfig, id string) bool
	RenderError(componentID, componentType string, err error) string
	RenderUnknown(typeName string) string
}

type Renderer struct {
	engine  *Engine
	context RenderContext
}

// NewRenderer creates a new Renderer with the given engine and render context
func NewRenderer(engine *Engine, context RenderContext) *Renderer {
	return &Renderer{
		engine:  engine,
		context: context,
	}
}

// Render renders the entire layout tree using the provided RenderContext
func (r *Renderer) Render() string {
	if r.engine.root == nil {
		log.Error("Renderer.Render: Layout root is nil")
		return ""
	}

	result := r.engine.Layout()
	if result == nil {
		log.Error("Renderer.Render: Layout result is nil")
		return ""
	}

	log.Trace("Renderer.Render: Got %d nodes from layout engine", len(result.Nodes))

	// Render the entire layout tree
	rendered := r.renderLayoutNode(r.engine.root)
	log.Trace("Renderer.Render: Rendered length: %d", len(rendered))
	return rendered
}

// renderLayoutNode recursively renders a layout node and all its children
func (r *Renderer) renderLayoutNode(node *LayoutNode) string {
	if node == nil {
		return ""
	}

	log.Trace("Renderer.renderLayoutNode: Rendering node %s with %d children", node.ID, len(node.Children))

	// Handle leaf nodes (actual components)
	if len(node.Children) == 0 && node.ID != "" {
		// This is a leaf - try to get the component from context
		if compInstance, exists := r.context.GetComponentInstance(node.ID); exists {
			log.Trace("Renderer.renderLayoutNode: Found component instance for %s", node.ID)
			config := core.RenderConfig{
				Width:  node.Bound.Width,
				Height: node.Bound.Height,
			}
			// Resolve props via context
			props, err := r.context.ResolveProps(node.ID)
			if err != nil {
				log.Error("Renderer.renderLayoutNode: Failed to resolve props for %s: %v", node.ID, err)
				props = make(map[string]interface{})
			}
			config.Data = props

			// Update component config via context
			r.context.UpdateComponentConfig(compInstance, config, node.ID)

			rendered, err := compInstance.Instance.Render(config)
			if err != nil {
				log.Error("Renderer.renderLayoutNode: Failed to render component %s: %v", node.ID, err)
				// Use context to render error
				return r.renderErrorComponent(node.ID, compInstance.Type, err)
			}
			log.Trace("Renderer.renderLayoutNode: Rendered component %s, length: %d", node.ID, len(rendered))
			return rendered
		} else {
			log.Trace("Renderer.renderLayoutNode: No component instance found for %s", node.ID)
		}
		return ""
	}

	var children []string

	for _, child := range node.Children {
		// Check if child has component bound
		if child.Component != nil && child.Component.Instance != nil {
			// Render actual component
			rendered := r.renderNodeWithBounds(child)
			log.Trace("Renderer.renderLayoutNode: Rendered component %s (via node.Component), length: %d", child.ID, len(rendered))
			if rendered != "" {
				children = append(children, rendered)
			}
		} else {
			// Recursively render nested layout
			rendered := r.renderLayoutNode(child)
			log.Trace("Renderer.renderLayoutNode: Rendered nested layout %s, length: %d", child.ID, len(rendered))
			if rendered != "" {
				children = append(children, rendered)
			}
		}
	}

	if len(children) == 0 {
		log.Trace("Renderer.renderLayoutNode: No children rendered for node %s", node.ID)
		return ""
	}

	// Join children based on this node's direction
	result := ""
	if node.Style != nil && node.Style.Direction == DirectionRow {
		result = lipgloss.JoinHorizontal(lipgloss.Top, children...)
	} else {
		result = lipgloss.JoinVertical(lipgloss.Left, children...)
	}

	// Apply container styles (Padding, Size)
	// This ensures the container node respects its calculated bounds and renders padding
	if node.Style != nil && (node.Bound.Width > 0 || node.Bound.Height > 0) {
		log.Trace("Renderer.renderLayoutNode: Applying container style for %s (Bound: %dx%d)", node.ID, node.Bound.Width, node.Bound.Height)
		style := lipgloss.NewStyle()

		// Apply Padding
		var pTop, pRight, pBottom, pLeft int
		if node.Style.Padding != nil {
			pTop = node.Style.Padding.Top
			pRight = node.Style.Padding.Right
			pBottom = node.Style.Padding.Bottom
			pLeft = node.Style.Padding.Left
			style = style.Padding(pTop, pRight, pBottom, pLeft)
		}

		// Apply Content Size (Bound - Padding)
		// Lipgloss Width/Height refers to the content area
		contentWidth := node.Bound.Width - pLeft - pRight
		contentHeight := node.Bound.Height - pTop - pBottom

		if contentWidth > 0 {
			style = style.Width(contentWidth).MaxWidth(contentWidth)
		}
		if contentHeight > 0 {
			style = style.Height(contentHeight).MaxHeight(contentHeight)
		}

		result = style.Render(result)

		// Force strict clipping for container as well, just in case
		if node.Bound.Height > 0 {
			lines := strings.Split(result, "\n")
			if len(lines) > node.Bound.Height {
				result = strings.Join(lines[:node.Bound.Height], "\n")
			}
		}
	}

	log.Trace("Renderer.renderLayoutNode: Node %s result length: %d", node.ID, len(result))
	return result
}

// renderNodeWithBounds renders a component with its calculated bounds.
func (r *Renderer) renderNodeWithBounds(node *LayoutNode) string {
	if node == nil || node.Component == nil || node.Component.Instance == nil {
		return ""
	}

	// Resolve props for this component from context
	props, err := r.context.ResolveProps(node.ID)
	if err != nil {
		log.Error("Renderer.renderNodeWithBounds: Failed to resolve props for %s: %v", node.ID, err)
		props = make(map[string]interface{})
	}

	config := core.RenderConfig{
		Data:   props,
		Width:  node.Bound.Width,
		Height: node.Bound.Height,
	}

	// Update component configuration via context
	r.context.UpdateComponentConfig(node.Component, config, node.ID)

	rendered, err := node.Component.Instance.Render(config)
	if err != nil {
		log.Error("Renderer.renderNodeWithBounds: Component %s render failed: %v", node.ID, err)
		return r.renderErrorComponent(node.ID, node.Component.Type, err)
	}

	if node.Bound.Width > 0 || node.Bound.Height > 0 {
		style := lipgloss.NewStyle().
			Width(node.Bound.Width).
			Height(node.Bound.Height).
			MaxWidth(node.Bound.Width).
			MaxHeight(node.Bound.Height)

		rendered = style.Render(rendered)

		// Force clip height if content exceeds bound height
		// Lipgloss MaxHeight doesn't strictly clip vertical content in all cases
		if node.Bound.Height > 0 {
			lines := strings.Split(rendered, "\n")
			if len(lines) > node.Bound.Height {
				rendered = strings.Join(lines[:node.Bound.Height], "\n")
			}
		}
	}

	return rendered
}

// renderErrorComponent renders an error display for a failed component
func (r *Renderer) renderErrorComponent(componentID string, componentType string, err error) string {
	// Use context callback to render error
	if r.context != nil {
		return r.context.RenderError(componentID, componentType, err)
	}

	// Fallback: render directly
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("52")).
		Padding(0, 2).
		Bold(true)

	errorMsg := fmt.Sprintf("[ERROR] %s (%s): %v", componentID, componentType, err)
	return style.Render(errorMsg)
}

// renderUnknownComponent renders a placeholder for unknown component types.
func (r *Renderer) renderUnknownComponent(typeName string) string {
	// Use context callback to render unknown component
	if r.context != nil {
		return r.context.RenderUnknown(typeName)
	}

	// Fallback: render directly
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	return style.Render(fmt.Sprintf("[Unknown component: %s]", typeName))
}
