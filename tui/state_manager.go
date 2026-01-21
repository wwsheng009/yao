package tui

import (
	"encoding/json"

	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/kun/log"
)

// FlattenData flattens nested data structures using kun/any library.
// This converts nested objects to dot-notation keys.
// Example: {"user": {"name": "John"}} → {"user.name": "John", "user": {"name": "John"}}
// This supports both flattened and nested access patterns.
func FlattenData(data map[string]interface{}) map[string]interface{} {
	if data == nil || len(data) == 0 {
		return data
	}

	wrappedRes := any.Of(data)
	flattened := wrappedRes.Map().MapStrAny.Dot()

	// Merge flattened keys with original data to support both access patterns
	for k, v := range flattened {
		data[k] = v
	}

	return data
}

// MergeData merges external data into existing data with configurable priority.
// priorityHigher means external data overrides existing data.
//
// Use cases:
// - Merge external command-line arguments into TUI config data
// - Merge OnLoad process results into state
// - Merge default data with user-provided data
//
// Parameters:
//   - existing: the destination data map (will be modified)
//   - external: the source data map to merge from
//   - priorityHigher: if true, external data overrides existing data
//
// Returns:
//   - The merged data map
func MergeData(existing, external map[string]interface{}, priorityHigher bool) map[string]interface{} {
	if existing == nil {
		existing = make(map[string]interface{})
	}

	if external == nil || len(external) == 0 {
		return existing
	}

	for k, v := range external {
		if priorityHigher {
			// External data takes priority
			existing[k] = v
		} else {
			// Existing data takes priority (external only adds new keys)
			if _, exists := existing[k]; !exists {
				existing[k] = v
			}
		}
	}

	return existing
}

// PrepareInitialState prepares the initial state for a TUI configuration.
// This is the ONLY entry point for preparing TUI state data.
// All data sources should be merged and flattened through this function.
//
// Data flow (in order of priority, highest to lowest):
//  1. External parameters (command-line args with :: prefix)
//  2. Static configuration data (from .tui.yao file)
//  3. OnLoad process/script results (executed during Init)
//
// Parameters:
//   - cfg: the TUI configuration
//   - externalData: external data from command-line arguments (optional)
//
// Returns:
//   - The prepared initial state (flattened and ready for model.State)
//
// Usage:
//
//	// In cmd/tui/tui.go, before creating model:
//	externalData = parseExternalData(args)
//	PrepareInitialState(cfg, externalData)
//	model := NewModel(cfg, nil)
func PrepareInitialState(cfg *Config, externalData map[string]interface{}) map[string]interface{} {
	// Ensure cfg.Data exists
	if cfg.Data == nil {
		cfg.Data = make(map[string]interface{})
	}

	// Merge external data into config data (external has higher priority)
	// This ensures external command-line arguments override static config
	if externalData != nil && len(externalData) > 0 {
		cfg.Data = MergeData(cfg.Data, externalData, true) // priorityHigher = true
		log.Trace("Merged %d external keys into config data", len(externalData))
	}

	// Flatten the merged data to support dot-notation access
	// This happens ONLY here, ensuring all data is consistently flattened
	cfg.Data = FlattenData(cfg.Data)

	log.Trace("Prepared initial state with %d keys (flattened)", len(cfg.Data))

	// Return the prepared data (which will be copied to model.State in NewModel)
	return cfg.Data
}

// ValidateAndFlattenExternal validates and flattens external data before merging.
// This ensures external data follows the same format as internal data.
//
// Parameters:
//   - externalData: raw external data from command-line or other sources
//
// Returns:
//   - Validated and flattened external data
//   - Error if validation fails
func ValidateAndFlattenExternal(externalData map[string]interface{}) (map[string]interface{}, error) {
	if externalData == nil || len(externalData) == 0 {
		return externalData, nil
	}

	// Flatten the external data before merging
	// This ensures consistency with internal data format
	flattened := FlattenData(externalData)

	log.Trace("Validated and flattened %d external data keys", len(flattened))

	return flattened, nil
}

// LoadTUIDefaults loads default data for system TUIs.
// This provides a way to set default values for built-in TUIs.
//
// Parameters:
//   - tuiID: the TUI identifier
//
// Returns:
//   - Default data for the TUI, or nil if no defaults defined
func LoadTUIDefaults(tuiID string) map[string]interface{} {
	// Load default data from data/tui directory if it exists
	defaultsPath := "data/tui/" + tuiID + ".json"

	exists, err := application.App.Exists(defaultsPath)
	if err != nil || !exists {
		return nil
	}

	bytes, err := application.App.Read(defaultsPath)
	if err != nil {
		log.Warn("Failed to read TUI defaults from %s: %v", defaultsPath, err)
		return nil
	}

	var defaults map[string]interface{}
	if err := json.Unmarshal(bytes, &defaults); err != nil {
		log.Warn("Failed to parse TUI defaults from %s: %v", defaultsPath, err)
		return nil
	}

	log.Trace("Loaded %d default keys for TUI %s", len(defaults), tuiID)

	// Flatten defaults to support dot-notation
	return FlattenData(defaults)
}

// ApplyOnLoadResult applies OnLoad process/script result to TUI state.
// This is used by the Init() method to apply OnLoad results.
//
// Parameters:
//   - model: the TUI model to update
//   - result: the result from OnLoad process/script
//   - onSuccess: the state key to store the result (optional)
func ApplyOnLoadResult(model *Model, result interface{}, onSuccess string) {
	if result == nil {
		return
	}

	// If onSuccess is specified, store the result in the specified state key
	if onSuccess != "" {
		model.SetState(onSuccess, result)
		log.Trace("Stored OnLoad result in state key: %s", onSuccess)
		return
	}

	// If onSuccess is not specified and result is a map, merge into state
	if resultMap, ok := result.(map[string]interface{}); ok {
		model.UpdateState(resultMap)
		log.Trace("Merged OnLoad result with %d keys into state", len(resultMap))
		return
	}

	// Otherwise, store in default __onLoadResult key
	model.SetState("__onLoadResult", result)
	log.Trace("Stored OnLoad result in __onLoadResult")
}
