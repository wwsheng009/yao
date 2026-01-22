package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/layout"
)

// RenderLayout renders the entire layout using the flexible layout engine.
func (m *Model) RenderLayout() string {
	if m.LayoutEngine == nil || m.LayoutRoot == nil {
		log.Error("RenderLayout: LayoutEngine or LayoutRoot is nil")
		return ""
	}

	result := m.LayoutEngine.Layout()
	if result == nil {
		log.Error("RenderLayout: Layout result is nil")
		return ""
	}

	log.Trace("RenderLayout: Got %d nodes from layout engine", len(result.Nodes))

	// Start from root and recursively render
	rendered := m.renderLayoutNode(m.LayoutRoot)
	log.Trace("RenderLayout: Rendered length: %d", len(rendered))
	return rendered
}

// renderLayoutNode recursively renders a layout node and all its children
func (m *Model) renderLayoutNode(node *layout.LayoutNode) string {
	if node == nil {
		return ""
	}

	log.Trace("renderLayoutNode: Rendering node %s with %d children", node.ID, len(node.Children))

	// Handle leaf nodes (actual components)
	if len(node.Children) == 0 && node.ID != "" {
		// This is a leaf - try to get the component from registry
		if compInstance, exists := m.ComponentInstanceRegistry.Get(node.ID); exists {
			log.Trace("renderLayoutNode: Found component instance for %s in registry", node.ID)
			config := core.RenderConfig{
				Width:  node.Bound.Width,
				Height: node.Bound.Height,
			}
			// Resolve props
			compConfig := m.findComponentConfig(node.ID)
			props := map[string]interface{}{}
			if compConfig != nil {
				props = m.resolveProps(compConfig)
			}
			config.Data = props

			// Ensure component is updated with latest config
			updateComponentInstanceConfig(compInstance, config, node.ID)

			rendered, err := compInstance.Instance.Render(config)
			if err != nil {
				log.Error("renderLayoutNode: Failed to render component %s: %v", node.ID, err)
				return m.renderErrorComponent(node.ID, compConfig.Type, err)
			}
			log.Trace("renderLayoutNode: Rendered component %s, length: %d", node.ID, len(rendered))
			return rendered
		} else {
			log.Trace("renderLayoutNode: No component instance found in registry for %s", node.ID)
		}
		return ""
	}

	var children []string

	for _, child := range node.Children {
		// Check if child has component bound
		if child.Component != nil && child.Component.Instance != nil {
			// Render actual component
			rendered := m.renderNodeWithBounds(child)
			log.Trace("renderLayoutNode: Rendered component %s (via node.Component), length: %d", child.ID, len(rendered))
			if rendered != "" {
				children = append(children, rendered)
			}
		} else {
			// Recursively render nested layout or look up in registry
			rendered := m.renderLayoutNode(child)
			log.Trace("renderLayoutNode: Rendered nested layout %s, length: %d", child.ID, len(rendered))
			if rendered != "" {
				children = append(children, rendered)
			}
		}
	}

	if len(children) == 0 {
		log.Trace("renderLayoutNode: No children rendered for node %s", node.ID)
		return ""
	}

	// Join children based on this node's direction
	result := ""
	if node.Style != nil && node.Style.Direction == layout.DirectionRow {
		result = lipgloss.JoinHorizontal(lipgloss.Top, children...)
	} else {
		result = lipgloss.JoinVertical(lipgloss.Left, children...)
	}
	log.Trace("renderLayoutNode: Node %s result length: %d", node.ID, len(result))
	return result
}

// renderNodeWithBounds renders a component with its calculated bounds.
func (m *Model) renderNodeWithBounds(node *layout.LayoutNode) string {
	if node == nil || node.Component == nil || node.Component.Instance == nil {
		return ""
	}

	// Resolve props for this component from original config
	compConfig := m.findComponentConfig(node.ID)
	props := map[string]interface{}{}
	if compConfig != nil {
		props = m.resolveProps(compConfig)
	}

	config := core.RenderConfig{
		Data:   props,
		Width:  node.Bound.Width,
		Height: node.Bound.Height,
	}

	// Update component configuration
	updateComponentInstanceConfig(node.Component, config, node.ID)

	rendered, err := node.Component.Instance.Render(config)
	if err != nil {
		log.Error("Component %s render failed: %v", node.ID, err)
		return m.renderErrorComponent(node.ID, node.Component.Type, err)
	}

	if node.Bound.Width > 0 || node.Bound.Height > 0 {
		style := lipgloss.NewStyle().
			Width(node.Bound.Width).
			Height(node.Bound.Height).
			MaxWidth(node.Bound.Width).  // ✅ 限制最大宽度
			MaxHeight(node.Bound.Height) // ✅ 限制最大高度

		rendered = style.Render(rendered)
	}

	return rendered
}

// RenderComponent renders a single component based on its type.
func (m *Model) RenderComponent(comp *Component) string {
	if comp == nil || comp.Type == "" {
		return ""
	}

	props := m.resolveProps(comp)

	renderConfig := core.RenderConfig{
		Data:   props,
		Width:  m.Width,
		Height: m.Height,
	}

	componentInstance, exists := m.ComponentInstanceRegistry.Get(comp.ID)
	if !exists {
		// If component instance doesn't exist in registry, try to create it on-demand
		// This is mainly for testing scenarios where direct rendering is needed
		registry := GetGlobalRegistry()
		factory, factoryExists := registry.GetComponentFactory(ComponentType(comp.Type))
		if !factoryExists {
			log.Error("RenderComponent: Unknown component type %s for %s", comp.Type, comp.ID)
			return m.renderUnknownComponent(comp.Type)
		}

		// Create component instance on-demand
		instance := factory(renderConfig, comp.ID)
		componentInstance = &core.ComponentInstance{
			ID:         comp.ID,
			Type:       comp.Type,
			Instance:   instance,
			LastConfig: renderConfig,
		}

		// Store in registry for future use
		m.ComponentInstanceRegistry.UpdateComponent(comp.ID, componentInstance.Instance)
		if m.Components == nil {
			m.Components = make(map[string]*core.ComponentInstance)
		}
		m.Components[comp.ID] = componentInstance

		log.Trace("RenderComponent: Created component instance on-demand for %s (type: %s)", comp.ID, comp.Type)
	}

	if !isRenderConfigChanged(componentInstance.LastConfig, renderConfig) {
	} else {
		updateComponentInstanceConfig(componentInstance, renderConfig, comp.ID)
	}

	rendered, err := componentInstance.Instance.Render(renderConfig)
	if err != nil {
		log.Error("Component %s (%s) render failed: %v", comp.ID, comp.Type, err)

		m.StateMu.Lock()
		m.State["__error_"+comp.ID] = fmt.Sprintf("Component %s (%s) render failed: %v", comp.ID, comp.Type, err)
		m.StateMu.Unlock()

		return m.renderErrorComponent(comp.ID, comp.Type, err)
	}
	return rendered
}

// renderErrorComponent renders an error display for a failed component
func (m *Model) renderErrorComponent(componentID string, componentType string, err error) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("52")).
		Padding(0, 2).
		Bold(true)

	errorMsg := fmt.Sprintf("[ERROR] %s (%s): %v", componentID, componentType, err)
	return style.Render(errorMsg)
}

// renderUnknownComponent renders a placeholder for unknown component types.
func (m *Model) renderUnknownComponent(typeName string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)

	return style.Render(fmt.Sprintf("[Unknown component: %s]", typeName))
}

// isInteractiveComponent 判断组件是否是交互式的
func isInteractiveComponent(componentType string) bool {
	switch componentType {
	case "input", "textarea", "menu", "table", "form", "viewport", "chat",
		"list", "paginator", "filepicker", "cursor", "crud":
		return true
	default:
		return false
	}
}

// applyPadding applies padding to a rendered string.
// padding format: [top, right, bottom, left]
func applyPadding(content string, padding []int) string {
	if len(padding) == 0 {
		return content
	}

	var top, right, bottom, left int

	switch len(padding) {
	case 1:
		top, right, bottom, left = padding[0], padding[0], padding[0], padding[0]
	case 2:
		top, right, bottom, left = padding[0], padding[1], padding[0], padding[1]
	case 3:
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[1]
	case 4:
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[3]
	}

	style := lipgloss.NewStyle().
		Padding(top, right, bottom, left)

	return style.Render(content)
}
