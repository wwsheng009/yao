package tui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
	"github.com/yaoapp/yao/tui/core"
)

// cache stores loaded TUI configurations
var cache sync.Map

// Load scans the tuis/ directory and loads all .tui configuration files.
// It registers each TUI in the cache with an ID derived from its file path.
//
// Returns the number of loaded TUIs and any error encountered.
func Load(cfg config.Config) error {
	messages := []string{}

	// Ignore if the tuis directory does not exist
	exists, err := application.App.Exists("tuis")
	if err != nil {
		return err
	}
	if !exists {
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

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			messages = append(messages, fmt.Sprintf("invalid TUI configuration in %s: %v", file, err))
			return nil
		}

		// Store in cache
		Set(id, cfg)
		log.Trace("Loaded TUI: %s (%s)", id, cfg.Name)

		return nil
	}, exts...)

	if len(messages) > 0 {
		return fmt.Errorf("%s", strings.Join(messages, ";\n"))
	}

	return err
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
func loadFile(file string) (*Config, error) {
	// Read file content
	data, err := application.App.Read(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON (supports .json, .jsonc - JSON with comments will be handled by parser)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// If the config has data property, flatten it using kun/any but preserve original structure
	if cfg.Data != nil {
		wrappedRes := any.Of(cfg.Data)
		flattened := wrappedRes.Map().MapStrAny.Dot()
		
		// Merge flattened keys with original data to support both access patterns
		for k, v := range flattened {
			cfg.Data[k] = v
		}
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
	setMissingBinding(cfg.Bindings, "tab", core.Action{Process: "tui.focus.next"})
	setMissingBinding(cfg.Bindings, "shift+tab", core.Action{Process: "tui.focus.prev"})
	setMissingBinding(cfg.Bindings, "enter", core.Action{Process: "tui.form.submit"})
	setMissingBinding(cfg.Bindings, "ctrl+r", core.Action{Process: "tui.refresh"})
	setMissingBinding(cfg.Bindings, "ctrl+l", core.Action{Process: "tui.refresh"})
	setMissingBinding(cfg.Bindings, "ctrl+z", core.Action{Process: "tui.suspend"})
	
	// Add default submit bindings for input components
	setDefaultInputSubmitBindings(&cfg)

	return &cfg, nil
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
