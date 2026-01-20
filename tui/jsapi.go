package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/gou/runtime/v8/bridge"
	"github.com/yaoapp/yao/tui/components"
	"github.com/yaoapp/yao/tui/core"
	"rogchap.com/v8go"
)

// NewTUIObject creates a JavaScript TUI object that provides access to the Go Model
// This allows JavaScript code to interact with the TUI state and functionality
func NewTUIObject(v8ctx *v8go.Context, model *Model) (*v8go.Value, error) {
	//创建并返回一个js对象实例
	jsObjTbl := v8go.NewObjectTemplate(v8ctx.Isolate())

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

	// Event bus methods
	jsObjTbl.Set("PublishEvent", tuiPublishEventMethod(v8ctx.Isolate(), model))
	jsObjTbl.Set("SubscribeToEvent", tuiSubscribeToEventMethod(v8ctx.Isolate(), model))

	// Focus management
	jsObjTbl.Set("SetFocus", tuiSetFocusMethod(v8ctx.Isolate(), model))

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

	return instance.Value, nil
}

// tuiGetStateMethod returns a function that gets a state value from the model
// Usage: tui.GetState(key) or tui.GetState() to get all state
func tuiGetStateMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

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
// Usage: tui.SetState(key, value[, targetID[, response]])
func tuiSetStateMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 2 {
			return bridge.JsException(v8ctx, "SetState requires key and value arguments")
		}

		// Get key and value
		key := args[0].String()
		value, err := bridge.GoValue(args[1], v8ctx)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert value: "+err.Error())
		}

		// Check for optional targetID
		targetID := ""
		if len(args) > 2 {
			targetID = args[2].String()
		}

		// Check for optional response type (default: Handled)
		response := core.Handled
		if len(args) > 3 {
			if respStr := args[3].String(); respStr == "ignored" {
				response = core.Ignored
			} else if respStr == "broadcast" {
				response = core.PassClick // Use PassClick for broadcast behavior
			}
		}

		// Create the state update message
		stateMsg := core.StateUpdateMsg{Key: key, Value: value}

		// Send message through appropriate channel
		if targetID != "" {
			// Targeted messaging
			targetedMsg := core.TargetedMsg{TargetID: targetID, InnerMsg: stateMsg}
			if model.Bridge != nil {
				model.Bridge.Send(targetedMsg)
			} else if model.Program != nil {
				model.Program.Send(targetedMsg)
			}
		} else {
			// Global messaging
			if model.Bridge != nil {
				model.Bridge.Send(stateMsg)
			} else if model.Program != nil {
				model.Program.Send(stateMsg)
			}
		}

		// Return response type as string for JS feedback
		var responseStr string
		switch response {
		case core.Ignored:
			responseStr = "ignored"
		case core.Handled:
			responseStr = "handled"
		case core.PassClick:
			responseStr = "broadcast"
		default:
			responseStr = "handled"
		}

		jsResponse, err := bridge.JsValue(v8ctx, responseStr)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert response: "+err.Error())
		}
		return jsResponse
	})
}

// tuiUpdateStateMethod returns a function that updates multiple state values at once
// Usage: tui.UpdateState(newStateObject[, targetID[, response]])
func tuiUpdateStateMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 1 {
			return bridge.JsException(v8ctx, "UpdateState requires a state object argument")
		}

		// Get the new state object
		newState, err := bridge.GoValue(args[0], v8ctx)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert state object: "+err.Error())
		}

		newStateMap, ok := newState.(map[string]interface{})
		if !ok {
			return bridge.JsException(v8ctx, "State must be an object")
		}

		// Check for optional targetID
		targetID := ""
		if len(args) > 1 {
			targetID = args[1].String()
		}

		// Check for optional response type (default: Handled)
		response := core.Handled
		if len(args) > 2 {
			if respStr := args[2].String(); respStr == "ignored" {
				response = core.Ignored
			} else if respStr == "broadcast" {
				response = core.PassClick // Use PassClick for broadcast behavior
			}
		}

		// Create the batch update message
		batchMsg := core.StateBatchUpdateMsg{Updates: newStateMap}

		// Send message through appropriate channel
		if targetID != "" {
			// Targeted messaging
			targetedMsg := core.TargetedMsg{TargetID: targetID, InnerMsg: batchMsg}
			if model.Bridge != nil {
				model.Bridge.Send(targetedMsg)
			} else if model.Program != nil {
				model.Program.Send(targetedMsg)
			}
		} else {
			// Global messaging
			if model.Bridge != nil {
				model.Bridge.Send(batchMsg)
			} else if model.Program != nil {
				model.Program.Send(batchMsg)
			}
		}

		// Return response type as string for JS feedback
		var responseStr string
		switch response {
		case core.Ignored:
			responseStr = "ignored"
		case core.Handled:
			responseStr = "handled"
		case core.PassClick:
			responseStr = "broadcast"
		default:
			responseStr = "handled"
		}

		jsResponse, err := bridge.JsValue(v8ctx, responseStr)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert response: "+err.Error())
		}
		return jsResponse
	})
}

// tuiExecuteActionMethod returns a function that executes an action
// Usage: tui.ExecuteAction(actionDefinition[, targetID[, response]])
func tuiExecuteActionMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 1 {
			return bridge.JsException(v8ctx, "ExecuteAction requires an action argument")
		}

		// Get the action object
		actionJS, err := bridge.GoValue(args[0], v8ctx)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert action: "+err.Error())
		}

		// Convert to Action struct
		action := &core.Action{}
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

		// Check for optional targetID
		targetID := ""
		if len(args) > 1 {
			targetID = args[1].String()
		}

		// Check for optional response type (default: Handled)
		response := core.Handled
		if len(args) > 2 {
			if respStr := args[2].String(); respStr == "ignored" {
				response = core.Ignored
			} else if respStr == "broadcast" {
				response = core.PassClick // Use PassClick for broadcast behavior
			}
		}

		// Execute the action asynchronously using the Bridge
		go func() {
			result, err := ExecuteAction(model, action)
			if err != nil {
				errorMsg := core.ErrorMessage{Err: err, Context: "JS API ExecuteAction error"}
				if targetID != "" {
					// Send targeted error message
					targetedMsg := core.TargetedMsg{TargetID: targetID, InnerMsg: errorMsg}
					if model.Bridge != nil {
						model.Bridge.Send(targetedMsg)
					} else if model.Program != nil {
						model.Program.Send(targetedMsg)
					}
				} else {
					// Send global error message
					if model.Bridge != nil {
						model.Bridge.Send(errorMsg)
					} else if model.Program != nil {
						model.Program.Send(errorMsg)
					}
				}
			} else if result != nil {
				// Send success result if available
				resultMsg := core.ProcessResultMsg{
					Target: action.OnSuccess,
					Data:   result,
				}
				if targetID != "" {
					targetedMsg := core.TargetedMsg{TargetID: targetID, InnerMsg: resultMsg}
					if model.Bridge != nil {
						model.Bridge.Send(targetedMsg)
					} else if model.Program != nil {
						model.Program.Send(targetedMsg)
					}
				} else {
					if model.Bridge != nil {
						model.Bridge.Send(resultMsg)
					} else if model.Program != nil {
						model.Program.Send(resultMsg)
					}
				}
			}
		}()

		// Return response type as string for JS feedback
		var responseStr string
		switch response {
		case core.Ignored:
			responseStr = "ignored"
		case core.Handled:
			responseStr = "handled"
		case core.PassClick:
			responseStr = "broadcast"
		default:
			responseStr = "handled"
		}

		jsResponse, err := bridge.JsValue(v8ctx, responseStr)
		if err != nil {
			return bridge.JsException(v8ctx, "Failed to convert response: "+err.Error())
		}
		return jsResponse
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
// Usage: tui.FocusNextInput([targetID])
func tuiFocusNextInputMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()

		// Check for optional targetID parameter
		targetID := ""
		if len(args) > 0 {
			targetID = args[0].String()
		}

		if targetID != "" {
			// Focus specific component
			if _, exists := model.Components[targetID]; exists {
				// Use the model's setFocus() method to ensure consistent focus management
				model.setFocus(targetID)
			}
		} else {
			// Original logic: cycle through inputs
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

			// Publish focus change event - components listen to this to update their state
			if model.EventBus != nil && model.CurrentFocus != "" {
				model.EventBus.Publish(core.ActionMsg{
					ID:     model.CurrentFocus,
					Action: core.EventFocusChanged,
					Data:   map[string]interface{}{"focused": true},
				})
			}
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

		for id, comp := range model.Components {
			if comp.Type == "input" {
				if inputWrapper, ok := comp.Instance.(*components.InputComponentWrapper); ok {
					model.State[id] = inputWrapper.GetValue()
				}
			}
		}

		return v8go.Undefined(iso)
	})
}

// tuiPublishEventMethod returns a function that publishes an event to the event bus
// Usage: tui.PublishEvent(componentID, action, data)
func tuiPublishEventMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 2 {
			return bridge.JsException(v8ctx, "PublishEvent requires componentID and action arguments")
		}

		componentID := args[0].String()
		action := args[1].String()

		// Optional data parameter
		var data interface{}
		if len(args) > 2 {
			var err error
			data, err = bridge.GoValue(args[2], v8ctx)
			if err != nil {
				return bridge.JsException(v8ctx, "Failed to convert event data: "+err.Error())
			}
		}

		// Create and publish the event
		eventMsg := core.ActionMsg{
			ID:     componentID,
			Action: action,
			Data:   data,
		}

		if model.EventBus != nil {
			model.EventBus.Publish(eventMsg)
		}

		// Also send through the bridge for async handling
		if model.Bridge != nil {
			model.Bridge.Send(eventMsg)
		}

		return v8go.Undefined(iso)
	})
}

// tuiSubscribeToEventMethod returns a function that subscribes to events on the event bus
// Usage: tui.SubscribeToEvent(action, callbackFunction)
func tuiSubscribeToEventMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		v8ctx := info.Context()
		args := info.Args()

		if len(args) < 2 {
			return bridge.JsException(v8ctx, "SubscribeToEvent requires action and callback arguments")
		}

		action := args[0].String()
		callbackFunc := args[1]

		if !callbackFunc.IsFunction() {
			return bridge.JsException(v8ctx, "Second argument must be a function")
		}

		if model.EventBus == nil {
			return bridge.JsException(v8ctx, "EventBus not available")
		}

		// Create a callback that will invoke the JS function
		jsCallback := func(msg core.ActionMsg) {
			// Convert ActionMsg to JS object
			jsData := map[string]interface{}{
				"id":     msg.ID,
				"action": msg.Action,
				"data":   msg.Data,
			}

			jsValue, err := bridge.JsValue(v8ctx, jsData)
			if err != nil {
				// Log error but don't crash
				return
			}

			// Call the JS callback function
			if jsFunc, err := callbackFunc.AsFunction(); err == nil {
				_, callErr := jsFunc.Call(v8ctx.Global(), jsValue)
				if callErr != nil {
					// Log error but don't crash
					return
				}
			}
		}

		// Subscribe to the event (unsubscribe function not used in this simple implementation)
		model.EventBus.Subscribe(action, jsCallback)

		return v8go.Undefined(iso)
	})
}

// tuiSetFocusMethod returns a function that sets focus to a specific component
// Usage: tui.SetFocus(componentID)
func tuiSetFocusMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()

		if len(args) < 1 {
			return bridge.JsException(info.Context(), "SetFocus requires componentID argument")
		}

		targetID := args[0].String()

		// Use the model's setFocus() method to ensure consistent focus management
		if _, exists := model.Components[targetID]; exists {
			model.setFocus(targetID)
		} else {
			// If component doesn't exist, clear all focus
			model.clearFocus()
		}

		return v8go.Undefined(iso)
	})
}
