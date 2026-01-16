package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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

	// Handle TUI-specific built-in processes
	switch action.Process {
	case "tui.quit", "tui.exit":
		result, err := ProcessQuitAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "tui.focus.next":
		result, err := ProcessFocusNextAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "tui.focus.prev":
		result, err := ProcessFocusPrevAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "tui.form.submit":
		result, err := ProcessFormSubmitAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "tui.refresh":
		result, err := ProcessRefreshAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "tui.clear":
		result, err := ProcessClearAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "tui.suspend":
		result, err := ProcessSuspendAction(model, action)
		if err != nil {
			return nil, err
		}
		return result, nil
	default:
		// Prepare arguments by evaluating any {{}} expressions against the model state
		preparedArgs, err := prepareActionArguments(action.Args, model)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare arguments: %w", err)
		}

		// Create and execute the process, passing the model ID as the first argument
		allArgs := append([]interface{}{model.Config.ID}, preparedArgs...)
		p := process.New(action.Process, allArgs...)
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
}

// ProcessQuitAction handles the quit action with model
// Usage: Called internally by executeProcessAction
func ProcessQuitAction(model *Model, action *Action) (interface{}, error) {
	// In the Bubble Tea framework, we typically send a tea.QuitMsg
	// rather than directly calling Quit() to allow graceful shutdown
	if model.Program != nil {
		// Check if we're in a test environment
		if _, isTest := os.LookupEnv("TESTING_TUI"); !isTest {
			// Send quit message to the program only in non-test environments
			model.Program.Send(QuitMsg{})
		}
	}
	return map[string]interface{}{"action": "quit"}, nil
}

// ProcessFocusNextAction handles focusing the next input
// Usage: Called internally by executeProcessAction
func ProcessFocusNextAction(model *Model, action *Action) (interface{}, error) {
	model.focusNextInput()
	// Update focus states in input models
	for id, inputModel := range model.InputModels {
		if id == model.CurrentFocus {
			inputModel.Model.Focus()
		} else {
			inputModel.Model.Blur()
		}
		model.InputModels[id] = inputModel
	}
	return map[string]interface{}{"action": "focus_next", "message": "Focus next input"}, nil
}

// ProcessFocusPrevAction handles focusing the previous input
// Usage: Called internally by executeProcessAction
func ProcessFocusPrevAction(model *Model, action *Action) (interface{}, error) {
	// For now, just return as not implemented; prev focus would need more complex logic
	return map[string]interface{}{"action": "focus_prev", "message": "Focus previous input"}, nil
}

// ProcessFormSubmitAction handles form submission
// Usage: Called internally by executeProcessAction
func ProcessFormSubmitAction(model *Model, action *Action) (interface{}, error) {
	// Collect all input values and update state
	model.StateMu.Lock()
	for id, inputModel := range model.InputModels {
		model.State[id] = inputModel.Value()
	}
	model.StateMu.Unlock()
	return map[string]interface{}{"action": "submit_form", "message": "Form submitted"}, nil
}

// ProcessSubmitAction handles general data submission
// Usage: Called internally by executeProcessAction
func ProcessSubmitAction(model *Model, action *Action) (interface{}, error) {
	// For general submission, we can collect input values similar to form submission
	// but may also include additional processing based on action parameters
	model.StateMu.Lock()
	for id, inputModel := range model.InputModels {
		model.State[id] = inputModel.Value()
	}
	model.StateMu.Unlock()

	// Here we could add additional processing specific to general submissions
	// For now, return a general submission result
	return map[string]interface{}{"action": "submit", "message": "Data submitted"}, nil
}

// ProcessRefreshAction handles refreshing the UI
// Usage: Called internally by executeProcessAction
func ProcessRefreshAction(model *Model, action *Action) (interface{}, error) {
	// Send refresh command
	if model.Program != nil {
		model.Program.Send(tea.WindowSizeMsg{Width: model.Width, Height: model.Height})
	}
	return map[string]interface{}{"action": "refresh", "message": "Refresh signal sent"}, nil
}

// ProcessClearAction handles clearing the screen
// Usage: Called internally by executeProcessAction
func ProcessClearAction(model *Model, action *Action) (interface{}, error) {
	if model.Program != nil {
		model.Program.Send(tea.ClearScreen())
	}
	return map[string]interface{}{"action": "clear", "message": "Clear screen signal sent"}, nil
}

// ProcessSuspendAction handles suspending the application
// Usage: Called internally by executeProcessAction
func ProcessSuspendAction(model *Model, action *Action) (interface{}, error) {
	if model.Program != nil {
		model.Program.Send(tea.SuspendMsg{})
	}
	return map[string]interface{}{"action": "suspend", "message": "Suspend signal sent"}, nil
}

// ProcessInputEscapeAction handles escaping from input component
// Usage: Called internally by executeProcessAction
func ProcessInputEscapeAction(model *Model, action *Action, inputID string) (interface{}, error) {
	// Blur the specified input component to exit edit mode
	if inputModel, exists := model.InputModels[inputID]; exists {
		inputModel.Blur()
		// Update the model's focus if this was the focused input
		if model.CurrentFocus == inputID {
			model.CurrentFocus = ""
		}
		model.InputModels[inputID] = inputModel
		return map[string]interface{}{
			"action": "input_escape", 
			"message": fmt.Sprintf("Input %s escaped edit mode", inputID),
			"inputID": inputID,
		}, nil
	}
	
	return nil, fmt.Errorf("input component %s not found", inputID)
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
