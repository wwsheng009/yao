package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/yao/tui/core"
)

func init() {
	// Register TUI processes
	process.Register("tui.load", ProcessLoad)
	process.Register("tui.get", ProcessGet)
	process.Register("tui.list", ProcessList)
	process.Register("tui.count", ProcessCount)
	process.Register("tui.reload", ProcessReload)
	process.Register("tui.message.send", ProcessMessageSend)
	process.Register("tui.message.targeted", ProcessMessageTargeted)
	process.Register("tui.quit", ProcessQuit)
	process.Register("tui.exit", ProcessExit)
	process.Register("tui.focus.next", ProcessFocusNext)
	process.Register("tui.focus.prev", ProcessFocusPrev)
	process.Register("tui.focus.set", ProcessFocusSet)
	process.Register("tui.form.submit", ProcessFormSubmit)
	process.Register("tui.submit", ProcessSubmit)
	process.Register("tui.refresh", ProcessRefresh)
	process.Register("tui.clear", ProcessClear)
	process.Register("tui.suspend", ProcessSuspend)
	process.Register("tui.input.escape", ProcessInputEscape)
	process.Register("tui.menu.select", ProcessMenuSelect)
	process.Register("tui.menu.navigate", ProcessMenuNavigate)
	process.Register("tui.state.update", ProcessStateUpdate)
	process.Register("tui.state.batch", ProcessStateBatch)
	process.Register("tui.event.publish", ProcessEventPublish)
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
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(tea.QuitMsg{})
	}

	return map[string]interface{}{
		"action":  "quit_sent",
		"modelID": modelID,
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

// ProcessExit exits the TUI application
// Usage: Process("tui.exit", modelID)
func ProcessExit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(tea.QuitMsg{})
	}

	return map[string]interface{}{
		"action":  "exit_sent",
		"modelID": modelID,
	}
}

// ProcessFocusNext focuses the next input component
// Usage: Process("tui.focus.next", modelID)
func ProcessFocusNext(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     "system",
			Action: core.EventFocusNext,
			Data:   nil,
		})
	}

	return map[string]interface{}{
		"action":  "focus_next_sent",
		"modelID": modelID,
	}
}

// ProcessFocusPrev focuses the previous input component
// Usage: Process("tui.focus.prev", modelID)
func ProcessFocusPrev(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     "system",
			Action: core.EventFocusPrev,
			Data:   nil,
		})
	}

	return map[string]interface{}{
		"action":  "focus_prev_sent",
		"modelID": modelID,
	}
}

// ProcessFormSubmit submits the current form
// Usage: Process("tui.form.submit", modelID)
func ProcessFormSubmit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     "form",
			Action: core.EventFormSubmitSuccess,
			Data:   map[string]interface{}{"timestamp": "now"},
		})
	}

	return map[string]interface{}{
		"action":  "form_submit_sent",
		"modelID": modelID,
	}
}

// ProcessSubmit handles general form/data submission
// Usage: Process("tui.submit", modelID)
func ProcessSubmit(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     "form",
			Action: core.EventFormSubmit,
			Data:   map[string]interface{}{"timestamp": "now"},
		})
	}

	return map[string]interface{}{
		"action":  "submit_sent",
		"modelID": modelID,
	}
}

// ProcessRefresh refreshes the TUI
// Usage: Process("tui.refresh", modelID)
func ProcessRefresh(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.RefreshMsg{})
	}

	return map[string]interface{}{
		"action":  "refresh_sent",
		"modelID": modelID,
	}
}

// ProcessClear clears the screen
// Usage: Process("tui.clear", modelID)
func ProcessClear(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(tea.ClearScreen())
	}

	return map[string]interface{}{
		"action":  "clear_sent",
		"modelID": modelID,
	}
}

// ProcessSuspend suspends the TUI application
// Usage: Process("tui.suspend", modelID)
func ProcessSuspend(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	modelID := process.ArgsString(0)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(tea.SuspendMsg{})
	}

	return map[string]interface{}{
		"action":  "suspend_sent",
		"modelID": modelID,
	}
}

// ProcessInputEscape handles escape from input component
// Usage: Process("tui.input.escape", modelID, inputID)
func ProcessInputEscape(process *process.Process) interface{} {
	process.ValidateArgNums(2)
	modelID := process.ArgsString(0)
	inputID := process.ArgsString(1)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     inputID,
			Action: core.EventEscapePressed,
			Data:   map[string]interface{}{"inputID": inputID},
		})
	}

	return map[string]interface{}{
		"action":  "input_escape_sent",
		"inputID": inputID,
		"modelID": modelID,
	}
}

// ProcessMenuSelect handles menu selection action
// Usage: Process("tui.menu.select", modelID, itemIndex)
func ProcessMenuSelect(process *process.Process) interface{} {
	process.ValidateArgNums(2)
	modelID := process.ArgsString(0)
	itemIndex := process.ArgsInt(1)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     "menu",
			Action: core.EventMenuItemSelected,
			Data:   map[string]interface{}{"itemIndex": itemIndex},
		})
	}

	return map[string]interface{}{
		"action":    "menu_select_sent",
		"itemIndex": itemIndex,
		"modelID":   modelID,
	}
}

// ProcessMenuNavigate handles menu navigation
// Usage: Process("tui.menu.navigate", modelID, direction)
func ProcessMenuNavigate(process *process.Process) interface{} {
	process.ValidateArgNums(2)
	modelID := process.ArgsString(0)
	direction := process.ArgsString(1) // "up" or "down"
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     "menu",
			Action: core.EventMenuNavigate,
			Data:   map[string]interface{}{"direction": direction},
		})
	}

	return map[string]interface{}{
		"action":    "menu_navigate_sent",
		"direction": direction,
		"modelID":   modelID,
	}
}

// ProcessMessageSend sends a custom action message through the message system
// Usage: Process("tui.message.send", modelID, actionName, data)
func ProcessMessageSend(process *process.Process) interface{} {
	process.ValidateArgNums(3)
	modelID := process.ArgsString(0)
	actionName := process.ArgsString(1)
	data := process.Args[2]
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	actionMsg := core.ActionMsg{
		ID:     "process",
		Action: actionName,
		Data:   data,
	}

	if model.Bridge != nil {
		model.Bridge.Send(actionMsg)
	} else if model.Program != nil {
		model.Program.Send(actionMsg)
	}

	return map[string]interface{}{
		"action":     "message_sent",
		"actionName": actionName,
		"modelID":    modelID,
	}
}

// ProcessMessageTargeted sends a targeted message to a specific component
// Usage: Process("tui.message.targeted", modelID, targetID, messageType, messageData)
func ProcessMessageTargeted(process *process.Process) interface{} {
	process.ValidateArgNums(4)
	modelID := process.ArgsString(0)
	targetID := process.ArgsString(1)
	messageType := process.ArgsString(2)
	messageData := process.Args[3]
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	var innerMsg tea.Msg
	switch messageType {
	case "state_update":
		if dataMap, ok := messageData.(map[string]interface{}); ok {
			for k, v := range dataMap {
				innerMsg = core.StateUpdateMsg{Key: k, Value: v}
				break
			}
		}
	case "state_batch":
		if dataMap, ok := messageData.(map[string]interface{}); ok {
			innerMsg = core.StateBatchUpdateMsg{Updates: dataMap}
		}
	case "action":
		innerMsg = core.ActionMsg{
			ID:     targetID,
			Action: messageData.(string),
			Data:   nil,
		}
	case "refresh":
		innerMsg = core.RefreshMsg{}
	case "custom":
		innerMsg = messageData.(tea.Msg)
	}

	if innerMsg != nil {
		targetedMsg := core.TargetedMsg{TargetID: targetID, InnerMsg: innerMsg}
		if model.Bridge != nil {
			model.Bridge.Send(targetedMsg)
		} else if model.Program != nil {
			model.Program.Send(targetedMsg)
		}
	}

	return map[string]interface{}{
		"action":   "targeted_sent",
		"targetID": targetID,
		"modelID":  modelID,
		"msgType":  messageType,
	}
}

// ProcessFocusSet sets focus to a specific component
// Usage: Process("tui.focus.set", modelID, componentID)
func ProcessFocusSet(process *process.Process) interface{} {
	process.ValidateArgNums(2)
	modelID := process.ArgsString(0)
	componentID := process.ArgsString(1)
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.ActionMsg{
			ID:     componentID,
			Action: core.EventFocusChanged,
			Data:   map[string]interface{}{"focused": true},
		})
	}

	return map[string]interface{}{
		"action":      "focus_set",
		"componentID": componentID,
		"modelID":     modelID,
	}
}

// ProcessStateUpdate updates a single state value
// Usage: Process("tui.state.update", modelID, key, value)
func ProcessStateUpdate(process *process.Process) interface{} {
	process.ValidateArgNums(3)
	modelID := process.ArgsString(0)
	key := process.ArgsString(1)
	value := process.Args[2]
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if model.Program != nil {
		model.Program.Send(core.StateUpdateMsg{Key: key, Value: value})
	}

	return map[string]interface{}{
		"action":  "state_updated",
		"key":     key,
		"modelID": modelID,
	}
}

// ProcessStateBatch performs a batch state update
// Usage: Process("tui.state.batch", modelID, updatesObject)
func ProcessStateBatch(process *process.Process) interface{} {
	process.ValidateArgNums(2)
	modelID := process.ArgsString(0)
	updates := process.Args[1]
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	if updatesMap, ok := updates.(map[string]interface{}); ok {
		if model.Program != nil {
			model.Program.Send(core.StateBatchUpdateMsg{Updates: updatesMap})
		}
	}

	return map[string]interface{}{
		"action":  "state_batch_updated",
		"modelID": modelID,
	}
}

// ProcessEventPublish publishes an event through the event bus
// Usage: Process("tui.event.publish", modelID, eventName, sourceID, data)
func ProcessEventPublish(process *process.Process) interface{} {
	process.ValidateArgNums(4)
	modelID := process.ArgsString(0)
	eventName := process.ArgsString(1)
	sourceID := process.ArgsString(2)
	data := process.Args[3]
	model := GetModel(modelID)
	if model == nil {
		return map[string]interface{}{
			"error":   "Model not found",
			"modelID": modelID,
		}
	}

	actionMsg := core.ActionMsg{
		ID:     sourceID,
		Action: eventName,
		Data:   data,
	}

	if model.EventBus != nil {
		model.EventBus.Publish(actionMsg)
	}

	return map[string]interface{}{
		"action":    "event_published",
		"eventName": eventName,
		"modelID":   modelID,
	}
}
