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

// ProcessQuit is a placeholder for quit action
// This will be handled by the TUI model's key bindings
// Usage: Process("tui.quit")
func ProcessQuit(process *process.Process) interface{} {
	return map[string]interface{}{
		"action":  "quit",
		"message": "Quit signal sent",
	}
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
