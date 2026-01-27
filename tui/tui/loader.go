package tui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/data"
	"github.com/yaoapp/yao/share"
	"github.com/yaoapp/yao/tui/tui/core"
	"github.com/yaoapp/yao/tui/tui/dsl"
)

// systemTUIs defines built-in TUI configurations
var systemTUIs = map[string]string{
	"__yao.tui-list": "yao/tuis/tui-list.tui.yao",
}

// cache stores loaded TUI configurations
var cache sync.Map

// reservedTUINames defines TUI names that are reserved for built-in subcommands
// These names conflict with Cobra subcommands and will be inaccessible via 'yao tui <name>'
var reservedTUINames = map[string]bool{
	"list":     true, // yao tui list
	"validate": true, // yao tui validate
	"inspect":  true, // yao tui inspect
	"check":    true, // yao tui check
	"dump":     true, // yao tui dump
	"help":     true, // yao tui help
}

// loadSystemTUIs loads built-in TUI configurations from source code
// Similar to loadSystemModels in model package
func loadSystemTUIs() error {
	messages := []string{}

	for id, tuiPath := range systemTUIs {
		content, err := data.Read(tuiPath)
		if err != nil {
			log.Error("Failed to read system TUI %s: %s", id, err.Error())
			messages = append(messages, err.Error())
			continue
		}

		// Parse TUI configuration
		var cfg Config
		if err := json.Unmarshal(content, &cfg); err != nil {
			log.Error("Failed to parse system TUI %s: %s", id, err.Error())
			messages = append(messages, err.Error())
			continue
		}

		// Assign ID
		cfg.ID = id

		// Run validation
		registry := GetGlobalRegistry()
		validator := NewConfigValidator(&cfg, registry)

		if !validator.Validate() {
			log.Warn("System TUI %s has validation issues:\n%s", id, validator.GetErrorSummary())
		}

		// Flatten data if present using unified flattening function
		if cfg.Data != nil {
			cfg.Data = FlattenData(cfg.Data)
		}

		// Assign component IDs
		counter := 0
		assignComponentIDs(&cfg.Layout, "comp", &counter)

		// Add default bindings
		if cfg.Bindings == nil {
			cfg.Bindings = make(map[string]core.Action)
		}
		setMissingBinding(cfg.Bindings, "q", core.Action{Process: "tui.quit"})
		setMissingBinding(cfg.Bindings, "ctrl+c", core.Action{Process: "tui.quit"})

		// Set default navigation mode
		if cfg.NavigationMode == "" {
			cfg.NavigationMode = "native"
		}

		if !cfg.TabCycles {
			cfg.TabCycles = true
		}

		// Add default submit bindings for input components
		setDefaultInputSubmitBindings(&cfg)

		// Store in cache
		Set(id, &cfg)
		log.Trace("Loaded system TUI: %s (%s)", id, cfg.Name)
	}

	if len(messages) > 0 {
		return fmt.Errorf("errors loading system TUIs: %s", strings.Join(messages, "; "))
	}

	return nil
}

// Load scans the tuis/ directory and loads all .tui configuration files.
// It registers each TUI in the cache with an ID derived from its file path.
//
// Returns the number of loaded TUIs and any error encountered.
func Load(cfg config.Config) error {
	messages := []string{}

	// First, load built-in system TUIs from source code
	if err := loadSystemTUIs(); err != nil {
		log.Warn("Failed to load some system TUIs: %v", err)
		// Continue loading filesystem TUIs even if system TUIs fail
	}

	// Load filesystem TUIs
	// Ignore if the tuis directory does not exist
	exists, err := application.App.Exists("tuis")
	if err != nil {
		return err
	}
	if !exists {
		// Even if tuis directory doesn't exist, system TUIs are already loaded
		return nil
	}
	exts := []string{"*.tui.yao", "*.tui.json", "*.tui.jsonc"}
	err = application.App.Walk("tuis", func(root, file string, isdir bool) error {
		if isdir {
			return nil
		}

		// Load the TUI configuration
		id := share.ID(root, file)
		cfg, err := loadFile(file)
		if err != nil {
			messages = append(messages, err.Error())
			return nil // Continue with other files
		}

		cfg.ID = id

		// Run comprehensive validation including rendering compatibility
		// The validator works with the component registry to ensure render compatibility
		registry := GetGlobalRegistry()
		validator := NewConfigValidator(cfg, registry)

		if !validator.Validate() {
			// Log validation summary but warn instead of error
			log.Warn("TUI configuration has validation issues in %s:\n%s", file, validator.GetErrorSummary())
			// Still load the TUI - validation is informational
			// Store in cache even with warnings
			Set(id, cfg)
			log.Trace("Loaded TUI with warnings: %s (%s)", id, cfg.Name)
			return nil
		}

		// Store in cache
		Set(id, cfg)
		log.Trace("Loaded TUI: %s (%s)", id, cfg.Name)

		return nil
	}, exts...)

	if len(messages) > 0 && strings.Contains(messages[0], "failed to load") {
		return fmt.Errorf("%s", strings.Join(messages, ";\n"))
	}

	// Check for naming conflicts after loading all TUIs
	checkConfigurationConflicts()

	return err
}

// checkConfigurationConflicts checks for TUI configuration names that conflict
// with built-in subcommands and emits warnings
func checkConfigurationConflicts() {
	conflicts := []string{}
	reservedNames := getReservedTUINameSlice()

	cache.Range(func(key, value interface{}) bool {
		tuiID, ok := key.(string)
		if !ok {
			return true
		}

		// Check if TUI name conflicts with reserved subcommand names
		// Exclude system TUIs (those starting with "__")
		if isReservedTUIName(tuiID) {
			conflicts = append(conflicts, tuiID)
		}

		return true
	})

	// Emit warnings for conflicts
	if len(conflicts) > 0 {
		log.Warn("=============================================================")
		log.Warn("TUI Naming Conflict Detected")
		log.Warn("=============================================================")
		for _, name := range conflicts {
			log.Warn("  TUI configuration '%s' conflicts with a built-in subcommand.", name)
			log.Warn("  It will be inaccessible via 'yao tui %s'.", name)
			log.Warn("  To access this TUI, use: yao run tui.%s", name)
			log.Warn("  Or rename the configuration file to avoid conflicts.")
		}
		log.Warn("")
		log.Warn("Reserved subcommand names: %v", reservedNames)
		log.Warn("")
		log.Warn("TUI naming guidelines:")
		log.Warn("  - Avoid using reserved names above")
		log.Warn("  - Use descriptive names: user-dashboard, data-editor, file-browser")
		log.Warn("  - Use prefixes to avoid conflicts: my-list, app-validate, tools/inspect")
		log.Warn("=============================================================")
	}
}

// isReservedTUIName checks if a TUI name is a reserved subcommand name
// Returns false for system TUIs (those starting with "__")
func isReservedTUIName(tuiID string) bool {
	// Exclude system TUIs (they start with "__")
	if strings.HasPrefix(tuiID, "__") {
		return false
	}

	// Check if the name is reserved
	return reservedTUINames[tuiID]
}

// getReservedTUINameSlice returns the list of reserved TUI names
func getReservedTUINameSlice() []string {
	names := make([]string, 0, len(reservedTUINames))
	for name := range reservedTUINames {
		names = append(names, name)
	}
	return names
}

// Set stores a TUI configuration in the cache with the given ID.
func Set(id string, cfg *Config) {
	cache.Store(id, cfg)
}

// Count returns the number of loaded TUI configurations.
func Count() int {
	count := 0
	cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Get retrieves a TUI configuration by its ID.
// Returns nil if the TUI is not found.
func Get(id string) *Config {
	value, ok := cache.Load(id)
	if !ok {
		return nil
	}

	cfg, ok := value.(*Config)
	if !ok {
		log.Error("Invalid TUI type in cache for ID: %s", id)
		return nil
	}

	return cfg
}

// Remove removes a TUI configuration from the cache by ID.
func Remove(id string) {
	cache.Delete(id)
}

// Reload reloads a specific TUI configuration from disk.
// This is useful during development for hot-reloading.
func Reload(id string) error {
	// Try different extensions
	extensions := []string{".tui.yao", ".tui.json", ".tui.jsonc"}
	var foundFile string
	for _, ext := range extensions {
		testFile := share.File(id, ext)
		exists, err := application.App.Exists(filepath.Join("tuis", testFile))
		if err == nil && exists {
			foundFile = filepath.Join("tuis", testFile)
			break
		}
	}

	if foundFile == "" {
		return fmt.Errorf("TUI file not found for ID: %s", id)
	}

	// Load the configuration
	cfg, err := loadFile(foundFile)
	if err != nil {
		return fmt.Errorf("failed to load TUI %s: %w", id, err)
	}

	cfg.ID = id

	// Validate
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration for TUI %s: %w", id, err)
	}

	// Update cache using Set
	Set(id, cfg)
	log.Info("Reloaded TUI: %s", id)

	return nil
}

// List returns all loaded TUI IDs.
func List() []string {
	var ids []string
	cache.Range(func(key, value interface{}) bool {
		id, ok := key.(string)
		if ok {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}

// loadFile loads and parses a TUI configuration file.
// Supports both JSON (.json, .jsonc) and YAML (.yml, .yaml) formats.
func loadFile(file string) (*Config, error) {
	// Read file content
	data, err := application.App.Read(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse using DSL parser (auto-detects JSON vs YAML)
	dslCfg, err := dsl.ParseFile(data, file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Convert DSL Config to legacy Config format for backward compatibility
	cfg := convertDSLToLegacyConfig(dslCfg)

	// If the config has data property, flatten it using unified flattening function
	if cfg.Data != nil {
		cfg.Data = FlattenData(cfg.Data)
	}

	// Assign unique IDs to components that don't have an ID
	counter := 0
	assignComponentIDs(&cfg.Layout, "comp", &counter)

	// Add default bindings if none exist
	if cfg.Bindings == nil {
		cfg.Bindings = make(map[string]core.Action)
	}

	// Set default bindings if they don't exist
	setMissingBinding(cfg.Bindings, "q", core.Action{Process: "tui.quit"})
	setMissingBinding(cfg.Bindings, "ctrl+c", core.Action{Process: "tui.quit"})

	// Tab/ShiftTab are handled by native navigation in model.go
	// They are not bound to allow flexible behavior based on NavigationMode
	// In "native" mode: Tab/ShiftTab navigate between components
	// In "bindable" mode: Tab/ShiftTab can be bound to custom actions
	// setMissingBinding(cfg.Bindings, "tab", core.Action{Process: "tui.focus.next"})
	// setMissingBinding(cfg.Bindings, "shift+tab", core.Action{Process: "tui.focus.prev"})

	setMissingBinding(cfg.Bindings, "enter", core.Action{Process: "tui.form.submit"})
	setMissingBinding(cfg.Bindings, "ctrl+r", core.Action{Process: "tui.refresh"})
	setMissingBinding(cfg.Bindings, "ctrl+l", core.Action{Process: "tui.refresh"})
	setMissingBinding(cfg.Bindings, "ctrl+z", core.Action{Process: "tui.suspend"})

	// Set default navigation mode if not specified
	if cfg.NavigationMode == "" {
		cfg.NavigationMode = "native" // Default to native navigation
	}

	// Set default tab cycles if not specified (true for backward compatibility)
	if !cfg.TabCycles {
		cfg.TabCycles = true
	}

	// Add default submit bindings for input components
	setDefaultInputSubmitBindings(cfg)

	return cfg, nil
}

// convertDSLToLegacyConfig converts a DSL Config to the legacy Config format.
// This maintains backward compatibility with existing code.
func convertDSLToLegacyConfig(dslCfg *dsl.Config) *Config {
	cfg := &Config{
		ID:             dslCfg.ID,
		Name:           dslCfg.Name,
		Data:           dslCfg.Data,
		OnLoad:         dslCfg.OnLoad,
		Bindings:       dslCfg.Bindings,
		LogLevel:       dslCfg.LogLevel,
		AutoFocus:      dslCfg.AutoFocus,
		NavigationMode: dslCfg.NavigationMode,
		TabCycles:      dslCfg.TabCycles,
	}

	// Convert DSL Node tree to legacy Layout
	if dslCfg.Layout != nil {
		cfg.Layout = convertDSLNodeToLegacyLayout(dslCfg.Layout)
	}

	return cfg
}

// convertDSLNodeToLegacyLayout converts a DSL Node to a legacy Layout.
func convertDSLNodeToLegacyLayout(dslNode *dsl.Node) Layout {
	layout := Layout{
		Direction: dslNode.Direction,
		Style:     "", // Style name not used in DSL
		Padding:   dslNode.Padding,
	}

	// Convert children
	if len(dslNode.Children) > 0 {
		layout.Children = make([]Component, len(dslNode.Children))
		for i, child := range dslNode.Children {
			layout.Children[i] = convertDSLNodeToLegacyComponent(child)
		}
	}

	return layout
}

// convertDSLNodeToLegacyComponent converts a DSL Node to a legacy Component.
func convertDSLNodeToLegacyComponent(dslNode *dsl.Node) Component {
	comp := Component{
		ID:        dslNode.ID,
		Type:      dslNode.Type,
		Bind:      dslNode.Bind,
		Props:     dslNode.Props,
		Actions:   dslNode.Actions,
		Direction: dslNode.Direction,
	}

	// Convert size properties
	if dslNode.Width != nil {
		comp.Width = dslNode.Width
	}
	if dslNode.Height != nil {
		comp.Height = dslNode.Height
	}
	if dslNode.FlexGrow > 0 {
		if comp.Props == nil {
			comp.Props = make(map[string]interface{})
		}
		comp.Props["flexGrow"] = dslNode.FlexGrow
	}

	// Convert style properties to props for backward compatibility
	if dslNode.Style != nil {
		if comp.Props == nil {
			comp.Props = make(map[string]interface{})
		}
		if dslNode.Style.Width != nil {
			comp.Props["width"] = dslNode.Style.Width
		}
		if dslNode.Style.Height != nil {
			comp.Props["height"] = dslNode.Style.Height
		}
		if len(dslNode.Style.Padding) > 0 {
			comp.Props["padding"] = dslNode.Style.Padding
		}
		if dslNode.Style.Gap > 0 {
			comp.Props["gap"] = dslNode.Style.Gap
		}
	}

	// Convert children
	if len(dslNode.Children) > 0 {
		comp.Children = make([]Component, len(dslNode.Children))
		for i, child := range dslNode.Children {
			comp.Children[i] = convertDSLNodeToLegacyComponent(child)
		}
	}

	return comp
}

// setDefaultInputSubmitBindings adds default submit bindings for input components
func setDefaultInputSubmitBindings(cfg *Config) {
	// This function would analyze the layout and set up appropriate submit bindings
	// for input components based on their configuration and placement
	// For now, we ensure that enter key triggers form submission if it's not already bound differently

	// We already set the general enter binding above, so this is more about
	// specific input component actions
	walkComponents(&cfg.Layout, func(comp *Component) {
		// If it's an input component, we can set default actions
		if comp.Type == "input" {
			// Ensure submit action is available for input components
			if comp.Actions == nil {
				comp.Actions = make(map[string]core.Action)
			}
			// Add default submit action if not already defined
			setMissingAction(comp.Actions, "submit", core.Action{Process: "tui.form.submit"})
			setMissingAction(comp.Actions, "general_submit", core.Action{Process: "tui.submit"})
		}
	})
}

// walkComponents walks through all components in a layout recursively
func walkComponents(layout *Layout, fn func(*Component)) {
	for i := range layout.Children {
		comp := &layout.Children[i]
		fn(comp)
		// If it's a layout component, walk its children too
		if comp.Type == "layout" {
			if nestedLayout, ok := comp.Props["layout"].(*Layout); ok {
				walkComponents(nestedLayout, fn)
			}
		}
	}
}

// setMissingAction adds an action only if the key doesn't already exist
func setMissingAction(actions map[string]core.Action, key string, action core.Action) {
	if _, exists := actions[key]; !exists {
		actions[key] = action
	}
}

// setMissingBinding adds a binding only if the key doesn't already exist
func setMissingBinding(bindings map[string]core.Action, key string, action core.Action) {
	if _, exists := bindings[key]; !exists {
		bindings[key] = action
	}
}

// assignComponentIDs assigns unique IDs to components that don't have an ID
func assignComponentIDs(layout *Layout, prefix string, counter *int) {
	for i := range layout.Children {
		comp := &layout.Children[i]
		if comp.ID == "" {
			*counter++
			comp.ID = fmt.Sprintf("%s_%d", prefix, *counter)
		}
		// If it's a layout component, assign IDs to its children too
		if comp.Type == "layout" {
			if nestedLayout, ok := comp.Props["layout"].(*Layout); ok {
				assignComponentIDs(nestedLayout, comp.ID, counter)
			}
		}
	}
}
