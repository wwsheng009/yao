package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/components"
)

// RenderLayout renders the entire layout tree recursively.
// It traverses the layout structure and renders each component,
// joining them according to the layout direction.
func (m *Model) RenderLayout() string {
	if m.Config == nil || m.Config.Layout.Direction == "" {
		return ""
	}

	return m.renderLayoutNode(&m.Config.Layout, m.Width, m.Height)
}

// renderLayoutNode renders a single layout node (can be nested).
func (m *Model) renderLayoutNode(layout *Layout, width, height int) string {
	if len(layout.Children) == 0 {
		return ""
	}

	var renderedChildren []string

	// Render each child component
	for _, child := range layout.Children {
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

// RenderComponent renders a single component based on its type.
// It routes to the appropriate component renderer.
func (m *Model) RenderComponent(comp *Component) string {
	if comp == nil || comp.Type == "" {
		return ""
	}

	// Apply state binding to props
	props := m.applyStateToProps(comp)

	// Route to component renderer
	switch comp.Type {
	case "header":
		return m.renderHeaderComponent(props)

	case "text":
		return m.renderTextComponent(props)

	case "layout":
		// Nested layout
		if nestedLayout, ok := props["layout"].(*Layout); ok {
			return m.renderLayoutNode(nestedLayout, m.Width, m.Height)
		}
		return ""

	default:
		// Unknown component type
		log.Warn("Unknown component type: %s", comp.Type)
		return m.renderUnknownComponent(comp.Type, props)
	}
}

// applyStateToProps processes component props and replaces {{}} expressions
// with actual values from the State.
func (m *Model) applyStateToProps(comp *Component) map[string]interface{} {
	if comp.Props == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	m.StateMu.RLock()
	defer m.StateMu.RUnlock()

	// Process each prop
	for key, value := range comp.Props {
		// Check if value is a string with {{}} expression
		if str, ok := value.(string); ok {
			result[key] = m.applyState(str)
		} else {
			result[key] = value
		}
	}

	// Handle bind attribute - bind entire state object
	if comp.Bind != "" {
		if bindValue, ok := m.State[comp.Bind]; ok {
			result["__bind_data"] = bindValue
		}
	}

	return result
}

// applyState replaces {{key}} expressions in a string with State values.
// Supports nested keys like {{user.name}}.
func (m *Model) applyState(text string) string {
	// Pattern to match {{key}} or {{key.nested}}
	re := regexp.MustCompile(`\{\{([a-zA-Z0-9_.]+)\}\}`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract key from {{key}}
		key := strings.Trim(match, "{}")
		key = strings.TrimSpace(key)

		// Look up value in State
		value := m.getStateValue(key)
		if value == nil {
			return "" // Return empty string if not found
		}

		// Convert value to string
		return fmt.Sprintf("%v", value)
	})
}

// getStateValue retrieves a value from State, supporting nested keys.
// Example: "user.name" looks up State["user"]["name"]
func (m *Model) getStateValue(key string) interface{} {
	if key == "" {
		return nil
	}

	// Split by dots for nested access
	parts := strings.Split(key, ".")
	
	var current interface{} = m.State

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
		} else {
			return nil
		}
	}

	return current
}

// renderHeaderComponent renders a header component.
func (m *Model) renderHeaderComponent(props map[string]interface{}) string {
	headerProps := components.ParseHeaderProps(props)
	return components.RenderHeader(headerProps, m.Width)
}

// renderTextComponent renders a text component.
func (m *Model) renderTextComponent(props map[string]interface{}) string {
	textProps := components.ParseTextProps(props)
	return components.RenderText(textProps, m.Width)
}

// renderUnknownComponent renders a placeholder for unknown component types.
func (m *Model) renderUnknownComponent(typeName string, props map[string]interface{}) string {
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

// getStringProp safely retrieves a string property with a default value.
func getStringProp(props map[string]interface{}, key, defaultValue string) string {
	if value, ok := props[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getIntProp safely retrieves an int property with a default value.
func getIntProp(props map[string]interface{}, key string, defaultValue int) int {
	if value, ok := props[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// getBoolProp safely retrieves a bool property with a default value.
func getBoolProp(props map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := props[key]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}
