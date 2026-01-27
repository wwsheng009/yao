package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yaoapp/gou/application"
	v8 "github.com/yaoapp/gou/runtime/v8"
	"github.com/yaoapp/kun/log"
	"rogchap.com/v8go"
)

// Script represents a TUI script (TypeScript/JavaScript)
type Script struct {
	*v8.Script
	File string // Original file path
}

// scripts stores loaded scripts by file path (without extension)
var scripts sync.Map // map[string]*Script

// scriptChannel handles concurrent script operations
const (
	saveScriptOp uint8 = iota
	removeScriptOp
)

type scriptOperation struct {
	file   string
	script *Script
	op     uint8
}

var scriptChannel = make(chan *scriptOperation, 10)

func init() {
	go scriptOperationHandler()
}

// scriptOperationHandler processes script operations sequentially
func scriptOperationHandler() {
	for {
		select {
		case operation := <-scriptChannel:
			switch operation.op {
			case saveScriptOp:
				scripts.Store(operation.file, operation.script)
				log.Trace("TUI script cached: %s", operation.file)
			case removeScriptOp:
				scripts.Delete(operation.file)
				log.Trace("TUI script removed from cache: %s", operation.file)
			}
		}
	}
}

// LoadScript loads a TUI script from the scripts/tui/ directory
// It supports both .ts and .js files, with .ts taking precedence
// Scripts are cached for performance
func LoadScript(file string, disableCache ...bool) (*Script, error) {
	// Normalize file path (remove .ts/.js extensions if present)
	base := strings.TrimSuffix(strings.TrimSuffix(file, ".ts"), ".js")

	// Check cache first
	if disableCache == nil || !disableCache[0] {
		if cached, ok := scripts.Load(base); ok {
			return cached.(*Script), nil
		}
	}

	// Try to find the script file
	// Priority: .ts > .js
	scriptPath := ""
	tsPath := base + ".ts"
	jsPath := base + ".js"

	if exists, _ := application.App.Exists(tsPath); exists {
		scriptPath = tsPath
	} else if exists, _ := application.App.Exists(jsPath); exists {
		scriptPath = jsPath
	} else {
		return nil, fmt.Errorf("script file not found: %s (.ts or .js)", base)
	}

	// Read script source
	source, err := application.App.Read(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script %s: %w", scriptPath, err)
	}

	// Compile script with V8
	v8script, err := v8.MakeScript(source, scriptPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to compile script %s: %w", scriptPath, err)
	}

	// Create TUI Script wrapper
	script := &Script{
		Script: v8script,
		File:   scriptPath,
	}

	// Cache the script
	scriptChannel <- &scriptOperation{
		file:   base,
		script: script,
		op:     saveScriptOp,
	}

	return script, nil
}

// GetScript retrieves a cached script
func GetScript(file string) (*Script, bool) {
	base := strings.TrimSuffix(strings.TrimSuffix(file, ".ts"), ".js")
	cached, ok := scripts.Load(base)
	if !ok {
		return nil, false
	}
	return cached.(*Script), true
}

// RemoveScript removes a script from cache
func RemoveScript(file string) {
	base := strings.TrimSuffix(strings.TrimSuffix(file, ".ts"), ".js")
	scriptChannel <- &scriptOperation{
		file: base,
		op:   removeScriptOp,
	}
}

// Execute executes a script method without Model context
func (s *Script) Execute(method string, args ...interface{}) (interface{}, error) {
	if s == nil || s.Script == nil {
		return nil, fmt.Errorf("script is nil")
	}

	// Create new context
	ctx, err := s.Script.NewContext("", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create V8 context: %w", err)
	}
	defer ctx.Close()

	// Check if method exists
	if !ctx.Global().Has(method) {
		return nil, fmt.Errorf("method %s not found in script %s", method, s.File)
	}

	// Call method
	result, err := ctx.Call(method, args...)
	if err != nil {
		return nil, fmt.Errorf("script execution failed (%s.%s): %w", s.File, method, err)
	}

	return result, nil
}

// ExecuteWithModel executes a script method with Model context injected
// The Model is available as 'ctx.tui' object in JavaScript, passed as the first argument
func (s *Script) ExecuteWithModel(model *Model, method string, args ...interface{}) (interface{}, error) {
	if s == nil || s.Script == nil {
		return nil, fmt.Errorf("script is nil")
	}

	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	// Create new context
	ctx, err := s.Script.NewContext("", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create V8 context: %w", err)
	}
	defer ctx.Close()
	// err = injectModelToContext(ctx.Context, model)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to injectModelToContext object: %w", err)
	// }
	// Create context object to pass as first argument
	ctxObj, err := NewContextObject(ctx.Context, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create context object: %w", err)
	}

	// Check if method exists
	if !ctx.Global().Has(method) {
		return nil, fmt.Errorf("method %s not found in script %s", method, s.File)
	}

	// Prepend the context object as the first argument
	callArgs := make([]interface{}, len(args)+1)
	// Convert to v8go.Valuer to ensure proper handling in bridge
	valuer := v8go.Valuer(ctxObj)
	callArgs[0] = valuer
	copy(callArgs[1:], args)

	// Call method with context as first argument
	result, err := ctx.Call(method, callArgs...)
	if err != nil {
		return nil, fmt.Errorf("script execution failed (%s.%s): %w", s.File, method, err)
	}

	return result, nil
}

// ListScripts returns all cached script paths
func ListScripts() []string {
	var paths []string
	scripts.Range(func(key, value interface{}) bool {
		paths = append(paths, key.(string))
		return true
	})
	return paths
}

// CountScripts returns the number of cached scripts
func CountScripts() int {
	count := 0
	scripts.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// ClearScripts removes all cached scripts
func ClearScripts() {
	scripts.Range(func(key, value interface{}) bool {
		scripts.Delete(key)
		return true
	})
	log.Trace("All TUI scripts cleared from cache")
}
