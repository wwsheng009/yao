package tui

import (
	"fmt"

	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/log"
)

// ExecuteAction executes an action synchronously
// This is used by the JavaScript API to execute actions from scripts
func ExecuteAction(model *Model, action *Action) (interface{}, error) {
	if action == nil {
		return nil, fmt.Errorf("action is nil")
	}

	// Validate action
	if err := action.Validate(); err != nil {
		return nil, fmt.Errorf("action validation failed: %w", err)
	}

	// Check if it's a Process or Script action
	if action.Process != "" {
		return executeProcessAction(model, action)
	}

	if action.Script != "" {
		return executeScriptAction(model, action)
	}

	// Direct state update (if payload is present)
	if len(action.Payload) > 0 {
		model.UpdateState(action.Payload)
		return action.Payload, nil
	}

	return nil, fmt.Errorf("action has no process, script, or payload")
}

// executeProcessAction executes a Yao Process action
func executeProcessAction(model *Model, action *Action) (interface{}, error) {
	log.Trace("TUI ExecuteProcess: %s", action.Process)

	// Prepare arguments by evaluating any {{}} expressions against the model state
	args, err := prepareActionArguments(action.Args, model)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare arguments: %w", err)
	}

	// Create and execute the process
	p := process.New(action.Process, args...)
	result, err := p.Exec()
	if err != nil {
		return nil, fmt.Errorf("process execution failed: %w", err)
	}

	// Handle success - store result in state if OnSuccess is specified
	if action.OnSuccess != "" && result != nil {
		model.SetState(action.OnSuccess, result)
	}

	return result, nil
}

// executeScriptAction executes a script method action
func executeScriptAction(model *Model, action *Action) (interface{}, error) {
	log.Trace("TUI ExecuteScript: %s.%s", action.Script, action.Method)

	// Load the script
	script, err := LoadScript(action.Script)
	if err != nil {
		return nil, fmt.Errorf("failed to load script %s: %w", action.Script, err)
	}

	// Prepare arguments by evaluating any {{}} expressions against the model state
	args, err := prepareActionArguments(action.Args, model)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare arguments: %w", err)
	}

	// Execute the script method with the model context
	result, err := script.ExecuteWithModel(model, action.Method, args...)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	// Handle success - store result in state if OnSuccess is specified
	if action.OnSuccess != "" && result != nil {
		model.SetState(action.OnSuccess, result)
	}

	return result, nil
}

// prepareActionArguments evaluates {{}} expressions in action arguments
// against the model's state and returns the resolved values
func prepareActionArguments(rawArgs []interface{}, model *Model) ([]interface{}, error) {
	if len(rawArgs) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(rawArgs))
	for i, rawArg := range rawArgs {
		arg, err := evaluateExpressions(rawArg, model)
		if err != nil {
			return nil, fmt.Errorf("argument %d evaluation failed: %w", i, err)
		}
		args[i] = arg
	}

	return args, nil
}

// evaluateExpressions recursively evaluates {{}} expressions in the given value
// against the model's state and returns the resolved value
func evaluateExpressions(value interface{}, model *Model) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Check if it's an expression like {{key}}
		if len(v) >= 4 && v[0:2] == "{{" && v[len(v)-2:] == "}}" {
			// Extract the key
			key := v[2 : len(v)-2]
			key = trimWhitespace(key)

			// Get value from state
			model.StateMu.RLock()
			stateValue, exists := model.State[key]
			model.StateMu.RUnlock()

			if !exists {
				return nil, fmt.Errorf("state key '%s' not found", key)
			}

			return stateValue, nil
		}
		return v, nil

	case map[string]interface{}:
		// Recursively evaluate expressions in map values
		result := make(map[string]interface{})
		for k, val := range v {
			evaluated, err := evaluateExpressions(val, model)
			if err != nil {
				return nil, err
			}
			result[k] = evaluated
		}
		return result, nil

	case []interface{}:
		// Recursively evaluate expressions in slice elements
		result := make([]interface{}, len(v))
		for i, val := range v {
			evaluated, err := evaluateExpressions(val, model)
			if err != nil {
				return nil, err
			}
			result[i] = evaluated
		}
		return result, nil

	default:
		// For other types (numbers, booleans, etc.), return as-is
		return v, nil
	}
}

// trimWhitespace removes leading and trailing whitespace
func trimWhitespace(s string) string {
	// Simple whitespace trimming - in a real implementation you might want to use strings.TrimSpace
	// but we'll implement a basic version to avoid import issues
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}