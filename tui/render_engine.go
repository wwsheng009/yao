package tui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// init initializes the random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// maxLayoutDepth is the maximum allowed depth for nested layouts to prevent stack overflow
const maxLayoutDepth = 50

// generateUniqueID generates a unique identifier for components
func generateUniqueID(prefix string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return prefix + "_" + string(b)
}

// RenderLayout renders the entire layout tree recursively.
// It traverses the layout structure and renders each component,
// joining them according to the layout direction.
func (m *Model) RenderLayout() string {
	if m.Config == nil || m.Config.Layout.Direction == "" {
		return ""
	}

	return m.renderLayoutNode(&m.Config.Layout, m.Width, m.Height, 0)
}

// renderLayoutNode renders a single layout node (can be nested).
// depth parameter tracks the current recursion depth to prevent stack overflow.
func (m *Model) renderLayoutNode(layout *Layout, width, height int, depth int) string {
	// Check maximum layout depth to prevent stack overflow
	if depth > maxLayoutDepth {
		errorMsg := fmt.Sprintf("Layout depth exceeded maximum limit: %d (max: %d)", depth, maxLayoutDepth)
		log.Error("Layout depth exceeded maximum limit: %d (max: %d)", depth, maxLayoutDepth)

		// Store error in state for potential display
		m.StateMu.Lock()
		m.State["__layout_depth_error"] = errorMsg
		m.StateMu.Unlock()

		return m.renderErrorComponent("layout", "root", fmt.Errorf("max layout depth exceeded"))
	}

	if len(layout.Children) == 0 {
		return ""
	}

	var renderedChildren []string

	// Render each child component
	for _, child := range layout.Children {
		// If child is a layout component, render it recursively with increased depth
		if child.Type == "layout" {
			if nestedLayout, ok := child.Props["layout"].(*Layout); ok {
				rendered := m.renderLayoutNode(nestedLayout, width, height, depth+1)
				if rendered != "" {
					renderedChildren = append(renderedChildren, rendered)
				}
				continue
			}
		}

		rendered := m.RenderComponent(&child)
		if rendered != "" {
			renderedChildren = append(renderedChildren, rendered)
		}
	}

	// Join components based on direction
	var result string
	if layout.Direction == "horizontal" {
		result = lipgloss.JoinHorizontal(lipgloss.Top, renderedChildren...)
	} else {
		// Default to vertical
		result = lipgloss.JoinVertical(lipgloss.Left, renderedChildren...)
	}

	// Apply padding if specified
	if len(layout.Padding) > 0 {
		result = applyPadding(result, layout.Padding)
	}

	return result
}

// RenderComponent renders a single component based on its type using new Render() method.
// It delegates rendering to the component's Render() method with new render configuration.
// Component instances must already be initialized via InitializeComponents().
func (m *Model) RenderComponent(comp *Component) string {
	if comp == nil || comp.Type == "" {
		return ""
	}

	// Validate component type
	if comp.Type == "" {
		log.Warn("Component type is empty for ID: %s", comp.ID)
		return m.renderUnknownComponent("empty_type")
	}

	// startTime := time.Now()
	// defer func() {
	// 	// duration := time.Since(startTime)
	// 	// log.Trace("RenderComponent: %s (type: %s) took %v", comp.ID, comp.Type, duration)
	// }()

	// Apply state binding to props
	props := m.resolveProps(comp)

	// Create render config
	// 渲染时width , Height会是屏幕的宽高。
	renderConfig := core.RenderConfig{
		Data:   props,
		Width:  m.Width,
		Height: m.Height,
	}

	// Get component instance from registry (should already exist)
	componentInstance, exists := m.ComponentInstanceRegistry.Get(comp.ID)
	if !exists {
		log.Error("RenderComponent: Component instance %s not found, was InitializeComponents() called?", comp.ID)
		return m.renderErrorComponent(comp.ID, comp.Type, fmt.Errorf("component instance not initialized"))
	}

	// Update render config if props changed
	if !isRenderConfigChanged(componentInstance.LastConfig, renderConfig) {
		// log.Trace("RenderComponent: config unchanged for %s, skipping update", comp.ID)
	} else {
		// Validate and update the config
		updateComponentInstanceConfig(componentInstance, renderConfig, comp.ID)
	}

	rendered, err := componentInstance.Instance.Render(renderConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("Component %s (%s) render failed: %v", comp.ID, comp.Type, err)
		log.Error("Component %s (%s) render failed: %v", comp.ID, comp.Type, err)

		// Store error in state for potential display
		m.StateMu.Lock()
		m.State["__error_"+comp.ID] = errorMsg
		m.StateMu.Unlock()

		// Render error component
		return m.renderErrorComponent(comp.ID, comp.Type, err)
	}
	return rendered
}

// renderErrorComponent renders an error display for a failed component
func (m *Model) renderErrorComponent(componentID string, componentType string, err error) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Background(lipgloss.Color("52")).  // Dark red
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

// applyPadding applies padding to a rendered string.
// padding format: [top, right, bottom, left]
func applyPadding(content string, padding []int) string {
	if len(padding) == 0 {
		return content
	}

	var top, right, bottom, left int

	switch len(padding) {
	case 1:
		// All sides
		top, right, bottom, left = padding[0], padding[0], padding[0], padding[0]
	case 2:
		// Vertical, Horizontal
		top, right, bottom, left = padding[0], padding[1], padding[0], padding[1]
	case 3:
		// Top, Horizontal, Bottom
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[1]
	case 4:
		// Top, Right, Bottom, Left
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[3]
	}

	style := lipgloss.NewStyle().
		Padding(top, right, bottom, left)

	return style.Render(content)
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
