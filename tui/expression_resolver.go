package tui

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/yaoapp/kun/log"
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
	expr.Function("NotNil", func(params ...interface{}) (interface{}, error) {
		if len(params) == 0 {
			return false, nil
		}
		return params[0] != nil, nil
	}),
	expr.AllowUndefinedVariables(),
}

// containsExpression checks if a string contains {{}} expressions
func containsExpression(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}

// evaluateExpression evaluates an expression and returns the result as interface{}
func (m *Model) evaluateExpression(expression string) (interface{}, error) {
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

	// Use cache to compile expression
	program, err := m.exprCache.GetOrCompile(processedExpression, func(exprStr string) (*vm.Program, error) {
		return expr.Compile(exprStr, append([]expr.Option{expr.Env(env)}, exprOptions...)...)
	})
	if err != nil {
		return nil, err
	}

	// Run expression
	res, err := vm.Run(program, env)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// resolveExpressionValue resolves an expression to its actual value without converting to string
func (m *Model) resolveExpressionValue(expression string) (interface{}, bool) {
	// Preprocess the expression to handle cases like features.0
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

	// Use cache to compile expression
	program, err := m.exprCache.GetOrCompile(processedExpression, func(exprStr string) (*vm.Program, error) {
		return expr.Compile(exprStr, append([]expr.Option{expr.Env(env)}, exprOptions...)...)
	})
	if err != nil {
		log.Warn("Expression compilation failed: %v, expression: %s", err, processedExpression)
		return nil, false
	}

	// Run expression
	res, err := vm.Run(program, env)
	if err != nil {
		log.Warn("Expression evaluation failed: %v, expression: %s", err, processedExpression)
		return nil, false
	}

	return res, true
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
		fullMatch := match[0]                     // {{ expression }}
		expression := strings.TrimSpace(match[1]) // expression without {{ }}

		// Skip empty expressions to avoid compilation errors
		if expression == "" {
			log.Warn("Skipping empty expression in: %s", text)
			continue
		}

		// Evaluate the expression
		res, err := m.evaluateExpression(expression)
		if err != nil {
			log.Warn("Expression evaluation failed: %v, expression: %s", err, expression)
			continue
		}

		// Convert evaluated result to string
		var replacement string
		if res == nil {
			replacement = ""
		} else {
			switch val := res.(type) {
			case string:
				replacement = val
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
				replacement = fmt.Sprintf("%v", val)
			default:
				// For complex data types like arrays and maps, convert to string representation
				replacement = fmt.Sprintf("%v", val)
				log.Trace("applyState: Converting complex data type to string: %T", val)
			}
		}

		// Replace in result
		result = strings.Replace(result, fullMatch, replacement, 1)
	}

	return result
}

// evaluateValue evaluates a value that might contain expressions
func (m *Model) evaluateValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// Check if the string contains {{}} expressions
		if containsExpression(v) {
			// Find all {{ expression }} patterns
			matches := stmtRe.FindAllStringSubmatchIndex(v, -1)
			if len(matches) == 0 {
				return v
			}

			// Process expressions in the string
			// If there's only one match and it's the entire string (like "{{variable}}"), return the evaluated value directly
			if len(matches) == 1 && matches[0][0] == 0 && matches[0][1] == len(v) {
				// This is a single expression like {{variable}}, return the evaluated value directly
				expression := strings.TrimSpace(v[matches[0][2]:matches[0][3]])
				// Skip empty expressions to avoid compilation errors
				if expression == "" {
					log.Warn("Skipping empty expression in: %s", v)
					return v
				}
				// Evaluate the expression
				res, err := m.evaluateExpression(expression)
				if err != nil {
					log.Warn("Expression evaluation failed: %v, expression: %s", err, expression)
					return v
				}
				// Return the evaluated value directly, preserving its type
				return res
			} else {
				// Multiple expressions or partial match, do string replacement using applyState
				return m.applyState(v)
			}
		}
		return v
	case map[string]interface{}:
		// Recursively evaluate map values
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = m.evaluateValue(val)
		}
		return result
	case []interface{}:
		// Recursively evaluate array values
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = m.evaluateValue(val)
		}
		return result
	default:
		// For other types, return as-is
		return v
	}
}

// bindData recursively processes data and replaces {{}} expressions with actual values from the State.
// This is similar to the gou/helper/bind.go implementation but adapted for TUI usage.
func (m *Model) bindData(v interface{}) interface{} {
	value := reflect.ValueOf(v)
	if value.IsValid() && value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	if !value.IsValid() {
		return v
	}

	valueKind := value.Kind()
	switch valueKind {
	case reflect.Slice, reflect.Array: // Slice || Array
		val := make([]interface{}, value.Len())
		for i := 0; i < value.Len(); i++ {
			val[i] = m.bindData(value.Index(i).Interface())
		}
		return val
	case reflect.Map: // Map
		// Create a new map to hold the processed values
		val := make(map[string]interface{})
		for _, key := range value.MapKeys() {
			k := fmt.Sprintf("%v", key.Interface())
			val[k] = m.bindData(value.MapIndex(key).Interface())
		}
		return val
	case reflect.String: // String with {{}} expressions
		input := value.String()
		// Find all {{ expression }} patterns
		matches := stmtRe.FindAllStringSubmatchIndex(input, -1)
		if len(matches) == 0 {
			return input
		}
		// Process expressions in the string
		// If there's only one match and it's entire string (like "{{variable}}"), return the evaluated value directly
		if len(matches) == 1 && matches[0][0] == 0 && matches[0][1] == len(input) {
			// This is a single expression like {{variable}}, return the evaluated value directly
			expression := strings.TrimSpace(input[matches[0][2]:matches[0][3]])
			// Skip empty expressions to avoid compilation errors
			if expression == "" {
				log.Warn("Skipping empty expression in: %s", input)
				return input
			}
			// Evaluate the expression
			res, err := m.evaluateExpression(expression)
			if err != nil {
				log.Warn("Expression evaluation failed: %v, expression: %s", err, expression)
				return input
			}
			// Return the evaluated value directly, preserving its type
			return res
		} else {
			// Multiple expressions or partial match, do string replacement
			result := input
			for _, match := range matches {
				fullMatchStart, fullMatchEnd := match[0], match[1]
				exprStart, exprEnd := match[2], match[3]
				expression := strings.TrimSpace(input[exprStart:exprEnd])

				// Skip empty expressions to avoid compilation errors
				if expression == "" {
					log.Warn("Skipping empty expression in: %s", input)
					continue
				}

				// Evaluate the expression
				res, err := m.evaluateExpression(expression)
				if err != nil {
					log.Warn("Expression evaluation failed: %v, expression: %s", err, expression)
					continue
				}

				// Convert evaluated result to string for replacement
				var replacement string
				if res == nil {
					replacement = ""
				} else {
					switch val := res.(type) {
					case string:
						replacement = val
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
						replacement = fmt.Sprintf("%v", val)
					default:
						replacement = fmt.Sprintf("%v", val)
					}
				}

				// Replace in result
				original := input[fullMatchStart:fullMatchEnd]
				result = strings.Replace(result, original, replacement, 1)
			}
			return result
		}
	default:
		return v
	}
}

// preprocessExpression handles special cases for expressions like features.0
// that need to be converted to index($, "features.0") when direct access fails
func preprocessExpression(expr string, state map[string]interface{}) string {
	// This is a simple preprocessing to detect patterns like "identifier.number" or "identifier.identifier"
	// where identifier might not exist as an object but combined key exists in the flattened state

	// First, let's check if the expression is a simple identifier like "features.0"
	// that could potentially be a flattened key
	if strings.Contains(expr, ".") {
		// Check if the whole expression as a key exists in the state
		if _, exists := state[expr]; exists {
			// If key exists, wrap it in index function for safe access
			return fmt.Sprintf(`index($, "%s")`, expr)
		}
	}

	return expr
}
