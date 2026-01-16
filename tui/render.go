package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/components"
)

// Regex for matching template expressions {{ }}
var stmtRe = regexp.MustCompile(`\{\{([\s\S]*?)\}\}`)

// Options for expr engine
var exprOptions = []expr.Option{
	expr.Function("len", func(params ...interface{}) (interface{}, error) {
		if len(params) == 0 {
			return 0, nil
		}
		v := params[0]
		switch val := v.(type) {
		case []interface{}:
			return len(val), nil
		case []map[string]interface{}:
			return len(val), nil
		case string:
			return len(val), nil
		case map[string]interface{}:
			return len(val), nil
		default:
			return 0, nil
		}
	}),
	expr.Function("index", func(params ...interface{}) (interface{}, error) {
		if len(params) < 2 {
			return nil, fmt.Errorf("index function requires 2 arguments")
		}
		container := params[0]
		key, ok := params[1].(string)
		if !ok {
			return nil, fmt.Errorf("index key must be a string")
		}
		
		// Helper function to get value from various map types
		getValue := func(container interface{}, key string) (interface{}, bool) {
			switch v := container.(type) {
			case map[string]interface{}:
				if val, exists := v[key]; exists {
					return val, true
				}
			case map[interface{}]interface{}:
				for k, v := range v {
					if kStr, ok := k.(string); ok && kStr == key {
						return v, true
					}
				}
			case *map[string]interface{}:
				if v != nil {
					if val, exists := (*v)[key]; exists {
						return val, true
					}
				}
			case *map[interface{}]interface{}:
				if v != nil {
					for k, v := range *v {
						if kStr, ok := k.(string); ok && kStr == key {
							return v, true
						}
					}
				}
			}
			return nil, false
		}
		
		if val, exists := getValue(container, key); exists {
			return val, nil
		}
		
		return nil, nil
	}),
	expr.Function("P_", func(params ...interface{}) (interface{}, error) {
		// Placeholder for process calls
		return nil, nil
	}),
	expr.Function("True", func(params ...interface{}) (interface{}, error) {
		if len(params) < 1 {
			return false, nil
		}
		if v, ok := params[0].(bool); ok {
			return v, nil
		}
		if v, ok := params[0].(string); ok {
			v = strings.ToLower(v)
			return v != "false" && v != "0", nil
		}
		if v, ok := params[0].(int); ok {
			return v != 0, nil
		}
		return false, nil
	}),
	expr.Function("False", func(params ...interface{}) (interface{}, error) {
		if len(params) < 1 {
			return true, nil
		}
		// Evaluate the truthiness of the first parameter using the same logic as True function
		v, ok := params[0].(bool)
		if ok {
			return !v, nil
		}
		str, ok := params[0].(string)
		if ok {
			str = strings.ToLower(str)
			return !(str != "false" && str != "0"), nil
		}
		if num, ok := params[0].(int); ok {
			return num == 0, nil
		}
		return true, nil
	}),
	expr.Function("Empty", func(params ...interface{}) (interface{}, error) {
		if len(params) < 1 {
			return true, nil
		}
		if params[0] == nil {
			return true, nil
		}
		if v, ok := params[0].(string); ok {
			return v == "", nil
		}
		if v, ok := params[0].(int); ok {
			return v == 0, nil
		}
		if v, ok := params[0].(bool); ok {
			return !v, nil
		}
		if v, ok := params[0].(map[string]interface{}); ok {
			return len(v) == 0, nil
		}
		if v, ok := params[0].([]interface{}); ok {
			return len(v) == 0, nil
		}
		return false, nil
	}),
	expr.AllowUndefinedVariables(),
}

// applyStateToProps processes component props and replaces {{}} expressions
// with actual values from the State.
func (m *Model) applyStateToProps(comp *Component) map[string]interface{} {
	if comp.Props == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

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
		m.StateMu.RLock()
		if bindValue, ok := m.State[comp.Bind]; ok {
			result["__bind_data"] = bindValue
		}
		m.StateMu.RUnlock()
	}

	return result
}

// applyState replaces {{key}} expressions in a string with State values.
// Uses expr-lang for powerful expression evaluation.
func (m *Model) applyState(text string) string {
	// Find all {{ expression }} patterns
	matches := stmtRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return text
	}

	result := text

	// Process each match
	for _, match := range matches {
		fullMatch := match[0]   // {{ expression }}
		expression := strings.TrimSpace(match[1])  // expression without {{ }}

		// Skip empty expressions to avoid compilation errors
		if expression == "" {
			log.Warn("Skipping empty expression in: %s", text)
			continue
		}

		// Preprocess the expression to handle cases like features.0
		// When expr tries to access features.0, it looks for features[0] or features.0
		// But in our flattened data, the key is "features.0"
		// So we transform expressions like "features.0" to "index($, \"features.0\")" when needed
		processedExpression := preprocessExpression(expression, m.State)

		// Evaluate the expression with current state
		m.StateMu.RLock()
		env := make(map[string]interface{})
		for k, v := range m.State {
			env[k] = v
		}
		// Add special $ variable to reference the entire state object
		env["$"] = m.State
		m.StateMu.RUnlock()

		// Compile expression with custom functions
		program, err := expr.Compile(processedExpression, append([]expr.Option{expr.Env(env)}, exprOptions...)...)

		if err != nil {
			log.Warn("Expression compilation failed: %v, expression: %s", err, processedExpression)
			continue
		}

		// Run expression
		res, err := vm.Run(program, env)
		if err != nil {
			log.Warn("Expression evaluation failed: %v, expression: %s", err, processedExpression)
			continue
		}

		// Convert evaluated result to string
		var replacement string
		if res == nil {
			replacement = ""
		} else {
			switch v := res.(type) {
			case string:
				replacement = v
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
				replacement = fmt.Sprintf("%v", v)
			default:
				replacement = fmt.Sprintf("%v", v)
			}
		}

		// Replace in result
		result = strings.Replace(result, fullMatch, replacement, 1)
	}

	return result
}

// preprocessExpression handles special cases for expressions like features.0
// that need to be converted to index($, "features.0") when the direct access fails
func preprocessExpression(expr string, state map[string]interface{}) string {
	// This is a simple preprocessing to detect patterns like "identifier.number" or "identifier.identifier"
	// where the identifier might not exist as an object but the combined key exists in the flattened state
	
	// First, let's check if the expression is a simple identifier like "features.0"
	// that could potentially be a flattened key
	if strings.Contains(expr, ".") {
		// Check if the whole expression as a key exists in the state
		if _, exists := state[expr]; exists {
			// If the key exists, wrap it in index function for safe access
			return fmt.Sprintf(`index($, "%s")`, expr)
		}
	}
	
	return expr
}

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
	case "input":
		// Special handling for input components
		if comp.ID != "" {
			// Create or update input model if it doesn't exist
			if _, exists := m.InputModels[comp.ID]; !exists {
				inputProps := components.ParseInputProps(props)
				inputModel := components.NewInputModel(inputProps)
				m.InputModels[comp.ID] = &inputModel
				
				// Set this input as focused if it's the first one or no other input is focused
				if m.CurrentFocus == "" {
					m.CurrentFocus = comp.ID
					inputModel.Model.Focus()
				} else if m.CurrentFocus == comp.ID {
					inputModel.Model.Focus()
				} else {
					inputModel.Model.Blur()
				}
				m.InputModels[comp.ID] = &inputModel
			}
			
			// Return the input view
			inputModel := m.InputModels[comp.ID]
			return inputModel.View()
		}
		return ""

	default:
		// Use component registry for other component types
		registry := GetGlobalRegistry()
		renderer, err := registry.GetComponent(ComponentType(comp.Type))
		if err != nil {
			log.Warn("Unknown component type: %s", comp.Type)
			return m.renderUnknownComponent(comp.Type, props)
		}
		
		// Apply state binding and render using the registered renderer
		boundProps := m.applyStateToProps(comp)
		return renderer(boundProps, m.Width)
	}
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