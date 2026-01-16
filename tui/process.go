package tui

import (
	"fmt"

	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
)

func init() {
	// Register TUI processes
	process.Register("tui.load", ProcessLoad)
	process.Register("tui.get", ProcessGet)
	process.Register("tui.list", ProcessList)
	process.Register("tui.count", ProcessCount)
	process.Register("tui.reload", ProcessReload)
	process.Register("tui.quit", ProcessQuit)
	process.Register("tui.exit", ProcessExit)
	process.Register("tui.focus.next", ProcessFocusNext)
	process.Register("tui.focus.prev", ProcessFocusPrev)
	process.Register("tui.form.submit", ProcessFormSubmit)
	process.Register("tui.submit", ProcessSubmit)
	process.Register("tui.refresh", ProcessRefresh)
	process.Register("tui.clear", ProcessClear)
	process.Register("tui.suspend", ProcessSuspend)
}

// ProcessLoad loads all TUI configurations
// Usage: Process("tui.load")
func ProcessLoad(process *process.Process) interface{} {
	// This will be called by engine.Load automatically
	// Just return success status
	return map[string]interface{}{
		"count": Count(),
	}
}

// ProcessGet gets a TUI configuration by ID
// Usage: Process("tui.get", "hello")
func ProcessGet(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	id := process.ArgsString(0)

	cfg := Get(id)
	if cfg == nil {
		exception.New("TUI not found: %s", 404, id).Throw()
	}

	return cfg
}

// ProcessList lists all loaded TUI IDs
// Usage: Process("tui.list")
func ProcessList(process *process.Process) interface{} {
	return List()
}

// ProcessCount returns the count of loaded TUIs
// Usage: Process("tui.count")
func ProcessCount(process *process.Process) interface{} {
	return Count()
}

// ProcessReload reloads a TUI configuration from disk
// Usage: Process("tui.reload", "hello")
func ProcessReload(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	id := process.ArgsString(0)

	err := Reload(id)
	if err != nil {
		exception.New("Failed to reload TUI: %s", 500, err.Error()).Throw()
	}

	return map[string]interface{}{
		"id":      id,
		"message": "TUI reloaded successfully",
	}
}

// ProcessQuit handles quit action
// Usage: Process("tui.quit", modelID)
func ProcessQuit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.quit"}
	result, err := ProcessQuitAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessExecute executes a TUI action
// This is called from within TUI when an action is triggered
// Usage: Process("tui.execute", tuiID, actionType, actionData)
func ProcessExecute(proc *process.Process) interface{} {
	proc.ValidateArgNums(3)

	tuiID := proc.ArgsString(0)
	actionType := proc.ArgsString(1)
	actionData := proc.Args[2]

	cfg := Get(tuiID)
	if cfg == nil {
		exception.New("TUI not found: %s", 404, tuiID).Throw()
	}

	// Execute based on action type
	switch actionType {
	case "process":
		// Execute a Yao Process
		if processName, ok := actionData.(string); ok {
			p := process.New(processName)
			result := p.Run()
			return result
		}
		exception.New("Invalid process action data", 400).Throw()

	case "script":
		// Execute a script method
		// This will be implemented in Phase 2
		exception.New("Script execution not yet implemented", 501).Throw()

	case "state":
		// Update state
		if updates, ok := actionData.(map[string]interface{}); ok {
			return map[string]interface{}{
				"action":  "state_update",
				"updates": updates,
			}
		}
		exception.New("Invalid state action data", 400).Throw()

	default:
		exception.New(fmt.Sprintf("Unknown action type: %s", actionType), 400).Throw()
	}

	return nil
}

// ProcessExit exits the TUI application
// Usage: Process("tui.exit", modelID)
func ProcessExit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.exit"}
	result, err := ProcessQuitAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessFocusNext focuses the next input component
// Usage: Process("tui.focus.next", modelID)
func ProcessFocusNext(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.focus.next"}
	result, err := ProcessFocusNextAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessFocusPrev focuses the previous input component
// Usage: Process("tui.focus.prev", modelID)
func ProcessFocusPrev(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.focus.prev"}
	result, err := ProcessFocusPrevAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessFormSubmit submits the current form
// Usage: Process("tui.form.submit", modelID)
func ProcessFormSubmit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.form.submit"}
	result, err := ProcessFormSubmitAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessSubmit handles general form/data submission
// Usage: Process("tui.submit", modelID)
func ProcessSubmit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.submit"}
	result, err := ProcessSubmitAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessRefresh refreshes the TUI
// Usage: Process("tui.refresh", modelID)
func ProcessRefresh(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.refresh"}
	result, err := ProcessRefreshAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessClear clears the screen
// Usage: Process("tui.clear", modelID)
func ProcessClear(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.clear"}
	result, err := ProcessClearAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}

// ProcessSuspend suspends the TUI application
// Usage: Process("tui.suspend", modelID)
func ProcessSuspend(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error": "Model not found",
			"modelID": modelID,
		}
	}

	action := &Action{Process: "tui.suspend"}
	result, err := ProcessSuspendAction(model, action)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
			"modelID": modelID,
		}
	}
	return result
}
