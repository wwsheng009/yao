package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/gou/runtime/v8/bridge"
	"rogchap.com/v8go"
)

// NewTUIObject creates a JavaScript TUI object that provides access to the Go Model
// This allows JavaScript code to interact with the TUI state and functionality
func NewTUIObject(v8ctx *v8go.Context, model *Model) (*v8go.Value, error) {
	//创建并返回一个js对象实例
	jsObjTbl := v8go.NewObjectTemplate(v8ctx.Isolate())
	
	// Set internal field count to 1 to store the __go_id
	// Internal fields are not accessible from JavaScript, providing better security
	// jsObjTbl.SetInternalFieldCount(1)

	// Register manager in global bridge registry for efficient Go object retrieval
	// The goValueID will be stored in internal field (index 0) after instance creation
	// Internal fields are not accessible from JavaScript, providing better security
	// goValueID := bridge.RegisterGoObject(model)

	// releaseFunc := tuiReleaseMethod(v8ctx.Isolate(), goValueID)
	// jsObjTbl.Set("__release", releaseFunc)
	// jsObjTbl.Set("Release", releaseFunc)


	// Set primitive fields
	// 属性
	jsObjTbl.Set("id", model.Config.Name)
	jsObjTbl.Set("width", int32(0))  // Will be updated by window size events
	jsObjTbl.Set("height", int32(0)) // Will be updated by window size events

	// Set methods
	jsObjTbl.Set("GetState", tuiGetStateMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("SetState", tuiSetStateMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("UpdateState", tuiUpdateStateMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("ExecuteAction", tuiExecuteActionMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("Refresh", tuiRefreshMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("Quit", tuiQuitMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("Interrupt", tuiInterruptMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("Suspend", tuiSuspendMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("ClearScreen", tuiClearScreenMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("EnterAltScreen", tuiEnterAltScreenMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("ExitAltScreen", tuiExitAltScreenMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("ShowCursor", tuiShowCursorMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("HideCursor", tuiHideCursorMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("FocusNextInput", tuiFocusNextInputMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("SubmitForm", tuiSubmitFormMethod(v8ctx.Isolate(), model))
	
	instance, err := jsObjTbl.NewInstance(v8ctx)
	if err != nil {
		// Clean up: release from global registry if instance creation failed
		// bridge.ReleaseGoObject(goValueID)
		return nil, err
	}
	// obj, err := instance.Value.AsObject()
	// if err != nil {
	// 	// bridge.ReleaseGoObject(goValueID)
	// 	return nil, err
	// }
	// 实例化后再赋值，跟上面在模板上赋值有什么区别
	// jsObject.Set("dummy", int32(0))

	// Store the goValueID in internal field (index 0)
	// This is not accessible from JavaScript, providing better security
	// err = obj.SetInternalField(0, goValueID)
	// if err != nil {
	// 	bridge.ReleaseGoObject(goValueID)
	// 	return nil, err
	// }

	return instance.Value, nil
}

// tuiGetStateMethod returns a function that gets a state value from the model
// Usage: tui.GetState(key) or tui.GetState() to get all state
func tuiGetStateMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		// // Get the TUI object (this)
		// thisObj, err := info.This().AsObject()
		// if err != nil {
		// 	return bridge.JsException(v8ctx, "Invalid TUI object context")
		// }

		// // Get goValueID from internal field (index 0)
		// goValueID := thisObj.GetInternalField(0)
		// if goValueID == nil || !goValueID.IsString() {
		// 	return bridge.JsException(v8ctx, "Invalid internal field")
		// }

		// // Get the Go model from the bridge
		// goObj := bridge.GetGoObject(goValueID.String())
		// if goObj == nil {
		// 	return bridge.JsException(v8ctx, "Failed to get Go model")
		// }

		// model, ok := goObj.(*Model)
		// if !ok {
		// 	return bridge.JsException(v8ctx, "Invalid model type")
		// }

		// Acquire read lock for thread-safe access
		model.StateMu.RLock()
		defer model.StateMu.RUnlock()

		// If no arguments, return all state
		if len(args) == 0 {
			stateJS, err := bridge.JsValue(v8ctx, model.State)
			if err != nil {
				return bridge.JsException(v8ctx, "Failed to convert state to JS: "+err.Error())
			}
			return stateJS
		}

		// Get the key
		key := args[0].String()

		// Get the value from state
		value, exists := model.getStateValue(key)
		if !exists {
			return v8go.Undefined(iso)
		}

		// Convert to JavaScript value
		jsValue, err := bridge.JsValue(v8ctx, value)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert value to JS: "+err.Error())
		}

		return jsValue
	})
}

// tuiSetStateMethod returns a function that sets a state value in the model
// Usage: tui.SetState(key, value)
func tuiSetStateMethod(iso *v8go.Isolate,model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 2 {
			return bridge.JsException(v8ctx, "SetState requires key and value arguments")
		}

		// // Get the TUI object (this)
		// thisObj, err := info.This().AsObject()
		// if err != nil {
		// 	return bridge.JsException(v8ctx, "Invalid TUI object context")
		// }

		// // Get goValueID from internal field (index 0)
		// goValueID := thisObj.GetInternalField(0)
		// if goValueID == nil || !goValueID.IsString() {
		// 	return bridge.JsException(v8ctx, "Invalid internal field")
		// }

		// // Get the Go model from the bridge
		// goObj := bridge.GetGoObject(goValueID.String())
		// if goObj == nil {
		// 	return bridge.JsException(v8ctx, "Failed to get Go model")
		// }

		// model, ok := goObj.(*Model)
		// if !ok {
		// 	return bridge.JsException(v8ctx, "Invalid model type")
		// }

		// Get key and value
		key := args[0].String()
		value, err := bridge.GoValue(args[1], v8ctx)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert value: "+err.Error())
		}

		// Send update message to trigger UI refresh
		// This ensures thread safety and proper UI updates
		if model.Program != nil {
			model.Program.Send(StateUpdateMsg{Key: key, Value: value})
		}

		return v8go.Undefined(iso)
	})
}

// tuiUpdateStateMethod returns a function that updates multiple state values at once
// Usage: tui.UpdateState(newStateObject)
func tuiUpdateStateMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 1 {
			return bridge.JsException(v8ctx, "UpdateState requires a state object argument")
		}

		// // Get the TUI object (this)
		// thisObj, err := info.This().AsObject()
		// if err != nil {
		// 	return bridge.JsException(v8ctx, "Invalid TUI object context")
		// }

		// // Get goValueID from internal field (index 0)
		// goValueID := thisObj.GetInternalField(0)
		// if goValueID == nil || !goValueID.IsString() {
		// 	return bridge.JsException(v8ctx, "Invalid internal field")
		// }

		// // Get the Go model from the bridge
		// goObj := bridge.GetGoObject(goValueID.String())
		// if goObj == nil {
		// 	return bridge.JsException(v8ctx, "Failed to get Go model")
		// }

		// model, ok := goObj.(*Model)
		// if !ok {
		// 	return bridge.JsException(v8ctx, "Invalid model type")
		// }

		// Get the new state object
		newState, err := bridge.GoValue(args[0], v8ctx)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert state object: "+err.Error())
		}

		newStateMap, ok := newState.(map[string]interface{})
		if !ok {
			return bridge.JsException(v8ctx, "State must be an object")
		}

		// Send update message to trigger UI refresh
		// This ensures thread safety and proper UI updates
		if model.Program != nil {
			model.Program.Send(StateBatchUpdateMsg{Updates: newStateMap})
		}

		return v8go.Undefined(iso)
	})
}

// tuiExecuteActionMethod returns a function that executes an action
// Usage: tui.ExecuteAction(actionDefinition)
func tuiExecuteActionMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 1 {
			return bridge.JsException(v8ctx, "ExecuteAction requires an action argument")
		}

		// // Get the TUI object (this)
		// thisObj, err := info.This().AsObject()
		// if err != nil {
		// 	return bridge.JsException(v8ctx, "Invalid TUI object context")
		// }

		// // Get goValueID from internal field (index 0)
		// goValueID := thisObj.GetInternalField(0)
		// if goValueID == nil || !goValueID.IsString() {
		// 	return bridge.JsException(v8ctx, "Invalid internal field")
		// }

		// // Get the Go model from the bridge
		// goObj := bridge.GetGoObject(goValueID.String())
		// if goObj == nil {
		// 	return bridge.JsException(v8ctx, "Failed to get Go model")
		// }

		// model, ok := goObj.(*Model)
		// if !ok {
		// 	return bridge.JsException(v8ctx, "Invalid model type")
		// }

		// Get the action object
		actionJS, err := bridge.GoValue(args[0], v8ctx)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert action: "+err.Error())
		}

		// Convert to Action struct
		action := &Action{}
		actionMap, ok := actionJS.(map[string]interface{})
		if !ok {
			return bridge.JsException(v8ctx, "Action must be an object")
		}

		// Set fields from the action map
		if script, ok := actionMap["script"].(string); ok {
			action.Script = script
		}
		if method, ok := actionMap["method"].(string); ok {
			action.Method = method
		}
		if process, ok := actionMap["process"].(string); ok {
			action.Process = process
		}
		if scriptAction, ok := actionMap["script_action"].(string); ok {
			// Map script_action to Script field
			action.Script = scriptAction
		}

		// Execute the action asynchronously
		// Note: This is a simplified version, in practice you'd want to handle errors and return values
		go func() {
			_, err := ExecuteAction(model, action)
			if err != nil {
				model.Program.Send(ErrorMsg{Err: err})
			}
		}()

		return v8go.Undefined(iso)
	})
}

// tuiRefreshMethod returns a function that forces a UI refresh
// Usage: tui.Refresh()
func tuiRefreshMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// v8ctx := info.Context()

		// // Get the TUI object (this)
		// thisObj, err := info.This().AsObject()
		// if err != nil {
		// 	return bridge.JsException(v8ctx, "Invalid TUI object context")
		// }

		// // Get goValueID from internal field (index 0)
		// goValueID := thisObj.GetInternalField(0)
		// if goValueID == nil || !goValueID.IsString() {
		// 	return bridge.JsException(v8ctx, "Invalid internal field")
		// }

		// // Get the Go model from the bridge
		// goObj := bridge.GetGoObject(goValueID.String())
		// if goObj == nil {
		// 	return bridge.JsException(v8ctx, "Failed to get Go model")
		// }

		// model, ok := goObj.(*Model)
		// if !ok {
		// 	return bridge.JsException(v8ctx, "Invalid model type")
		// }

		// Send a refresh message to trigger UI update
		model.Program.Send(tea.WindowSizeMsg{Width: model.Width, Height: model.Height})

		return v8go.Undefined(iso)
	})
}

// tuiReleaseMethod returns a function that releases the Go object from the bridge registry
// This is called when the JavaScript object is garbage collected or explicitly released
func tuiReleaseMethod(iso *v8go.Isolate, goValueID string) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Get the TUI object (this)
		thisObj, err := info.This().AsObject()
		if err == nil {
			// Get goValueID from internal field (index 0)
			internalID := thisObj.GetInternalField(0)
			if internalID != nil && internalID.IsString() {
				// Release from global bridge registry
				bridge.ReleaseGoObject(internalID.String())
			}
		} else {
			// Fallback to the goValueID passed to the function template
			bridge.ReleaseGoObject(goValueID)
		}

		return v8go.Undefined(info.Context().Isolate())
	})
}

// tuiQuitMethod returns a function that exits the TUI application
// Usage: tui.Quit()
func tuiQuitMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send a quit message to exit the application
		if model.Program != nil {
			model.Program.Send(tea.QuitMsg{})
		}

		return v8go.Undefined(iso)
	})
}

// tuiInterruptMethod returns a function that interrupts the TUI application
// Usage: tui.Interrupt()
func tuiInterruptMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send an interrupt message
		if model.Program != nil {
			model.Program.Send(tea.InterruptMsg{})
		}

		return v8go.Undefined(iso)
	})
}

// tuiSuspendMethod returns a function that suspends the TUI application
// Usage: tui.Suspend()
func tuiSuspendMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send a suspend message
		if model.Program != nil {
			model.Program.Send(tea.SuspendMsg{})
		}

		return v8go.Undefined(iso)
	})
}

// tuiClearScreenMethod returns a function that clears the screen
// Usage: tui.ClearScreen()
func tuiClearScreenMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send a clear screen message
		if model.Program != nil {
			model.Program.Send(tea.ClearScreen())
		}

		return v8go.Undefined(iso)
	})
}

// tuiEnterAltScreenMethod returns a function that enters alternate screen
// Usage: tui.EnterAltScreen()
func tuiEnterAltScreenMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send an enter alt screen message
		if model.Program != nil {
			model.Program.Send(tea.EnterAltScreen())
		}

		return v8go.Undefined(iso)
	})
}

// tuiExitAltScreenMethod returns a function that exits alternate screen
// Usage: tui.ExitAltScreen()
func tuiExitAltScreenMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send an exit alt screen message
		if model.Program != nil {
			model.Program.Send(tea.ExitAltScreen())
		}

		return v8go.Undefined(iso)
	})
}

// tuiShowCursorMethod returns a function that shows the cursor
// Usage: tui.ShowCursor()
func tuiShowCursorMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send a show cursor message
		if model.Program != nil {
			model.Program.Send(tea.ShowCursor())
		}

		return v8go.Undefined(iso)
	})
}

// tuiHideCursorMethod returns a function that hides the cursor
// Usage: tui.HideCursor()
func tuiHideCursorMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Send a hide cursor message
		if model.Program != nil {
			model.Program.Send(tea.HideCursor())
		}

		return v8go.Undefined(iso)
	})
}

// NewContextObject creates a JavaScript context object that wraps the TUI object
// This allows JavaScript code to interact with the TUI through ctx.tui.xxx
func NewContextObject(v8ctx *v8go.Context, model *Model) (*v8go.Value, error) {
	
	// Create context object template
	ctxObject := v8go.NewObjectTemplate(v8ctx.Isolate())

	// Create instance
	instance, err := ctxObject.NewInstance(v8ctx)
	if err != nil {
		return nil, err
	}
	ctxObj, err := instance.Value.AsObject()
	if err != nil {
		return nil, err
	}

	// Create the TUI object
	tuiValue, err := NewTUIObject(v8ctx, model)
	if err != nil {
		return nil, err
	}
	// tuiObj, err := tuiValue.AsObject()
	// if err != nil {
	// 	return nil, err
	// }
	// Set the tui object as a property in the template
	ctxObj.Set("tui", tuiValue)


	return instance.Value, nil
}

// injectModelToContext injects the TUI model into the V8 context as a global object
// This makes the TUI object available to JavaScript code as 'tui'
func injectModelToContext(v8ctx *v8go.Context, model *Model) error {
	// Create the context object that wraps the TUI object
	ctxObj, err := NewTUIObject(v8ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create context object: %w", err)
	}


	// Set it as a global variable 'ctx'
	global := v8ctx.Global()
	err = global.Set("xui", ctxObj)
	if err != nil {
		return fmt.Errorf("failed to set ctx global: %w", err)
	}

	return nil
}

// tuiFocusNextInputMethod returns a function that focuses the next input component
// Usage: tui.FocusNextInput()
func tuiFocusNextInputMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Find the next input component ID after the current focus
		currentFocus := model.CurrentFocus
		
		// Find all input component IDs in the layout
		inputIDs := []string{}
		for _, comp := range model.Config.Layout.Children {
			if comp.Type == "input" && comp.ID != "" {
				inputIDs = append(inputIDs, comp.ID)
			}
		}
		
		// Find current position and move to next
		currentIndex := -1
		for i, id := range inputIDs {
			if id == currentFocus {
				currentIndex = i
				break
			}
		}
		
		// Move to next input, wrap around if needed
		if currentIndex >= 0 && currentIndex < len(inputIDs)-1 {
				model.CurrentFocus = inputIDs[currentIndex+1]
		} else if len(inputIDs) > 0 {
				model.CurrentFocus = inputIDs[0] // Wrap to first
			}
		
		// Update focus states in input models
		for id, inputModel := range model.InputModels {
			if id == model.CurrentFocus {
				inputModel.Model.Focus()
			} else {
				inputModel.Model.Blur()
			}
			model.InputModels[id] = inputModel
		}
		
		return v8go.Undefined(iso)
	})
}

// tuiSubmitFormMethod returns a function that submits the form
// Usage: tui.SubmitForm()
func tuiSubmitFormMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// Collect all input values and update state
		model.StateMu.Lock()
		defer model.StateMu.Unlock()
		
		for id, inputModel := range model.InputModels {
			model.State[id] = inputModel.Value()
		}
		
		return v8go.Undefined(iso)
	})
}
