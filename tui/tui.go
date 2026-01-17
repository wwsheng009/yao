package tui

import (
	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/tui/core"
)

// Setup initializes the TUI module
// This should be called during engine startup
func Setup(cfg config.Config) error {
	log.Trace("Setting up TUI module...")

	// Load TUI configurations
	err := Load(cfg)
	if err != nil {
		log.Error("Failed to load TUI configurations: %v", err)
		return err
	}

	count := Count()
	if count > 0 {
		log.Info("TUI module initialized with %d configuration(s)", count)
	} else {
		log.Trace("TUI module initialized (no configurations found)")
	}

	return nil
}

// Unload clears all loaded TUI configurations
// Useful for cleanup or testing
func Unload() {
	cache.Range(func(key, value interface{}) bool {
		cache.Delete(key)
		return true
	})
	log.Trace("TUI module unloaded")
}

// Export exports TUI configurations for external use
// Returns a map of TUI ID to Config
func Export() map[string]*Config {
	result := make(map[string]*Config)
	cache.Range(func(key, value interface{}) bool {
		if id, ok := key.(string); ok {
			if cfg, ok := value.(*Config); ok {
				result[id] = cfg
			}
		}
		return true
	})
	return result
}

// Validate validates all loaded TUI configurations
// Returns a map of TUI ID to validation error (nil if valid)
func Validate() map[string]error {
	errors := make(map[string]error)
	cache.Range(func(key, value interface{}) bool {
		if id, ok := key.(string); ok {
			if cfg, ok := value.(*Config); ok {
				if err := cfg.Validate(); err != nil {
					errors[id] = err
				}
			}
		}
		return true
	})
	return errors
}

// GetInfo returns information about a TUI configuration
func GetInfo(id string) map[string]interface{} {
	cfg := Get(id)
	if cfg == nil {
		return nil
	}

	return map[string]interface{}{
		"id":              cfg.ID,
		"name":            cfg.Name,
		"hasOnLoad":       cfg.OnLoad != nil,
		"layoutDirection": cfg.Layout.Direction,
		"componentCount":  len(cfg.Layout.Children),
		"bindingCount":    len(cfg.Bindings),
	}
}

// GetStats returns statistics about loaded TUIs
func GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total":            Count(),
		"withOnLoad":       0,
		"withBindings":     0,
		"totalComponents":  0,
		"totalBindings":    0,
		"layoutDirections": make(map[string]int),
	}

	cache.Range(func(key, value interface{}) bool {
		if cfg, ok := value.(*Config); ok {
			if cfg.OnLoad != nil {
				stats["withOnLoad"] = stats["withOnLoad"].(int) + 1
			}
			if len(cfg.Bindings) > 0 {
				stats["withBindings"] = stats["withBindings"].(int) + 1
			}
			stats["totalComponents"] = stats["totalComponents"].(int) + len(cfg.Layout.Children)
			stats["totalBindings"] = stats["totalBindings"].(int) + len(cfg.Bindings)

			directions := stats["layoutDirections"].(map[string]int)
			directions[cfg.Layout.Direction]++
		}
		return true
	})

	return stats
}

// IsLoaded checks if a TUI with the given ID is loaded
func IsLoaded(id string) bool {
	_, ok := cache.Load(id)
	return ok
}

// GetOrCreate gets a TUI configuration or creates a default one
func GetOrCreate(id string) *Config {
	cfg := Get(id)
	if cfg != nil {
		return cfg
	}

	// Create a default configuration
	defaultCfg := &Config{
		ID:   id,
		Name: id,
		Data: make(map[string]interface{}),
		Layout: Layout{
			Direction: "vertical",
			Children:  []Component{},
		},
		Bindings: make(map[string]core.Action),
	}

	Set(id, defaultCfg)
	log.Warn("Created default TUI configuration for: %s", id)

	return defaultCfg
}

// WatchAndReload watches for file changes and reloads TUI configurations
// This is useful during development
func WatchAndReload(cfg config.Config) error {
	// This will be implemented in a future version
	// For now, just return nil
	log.Warn("TUI watch mode not yet implemented")
	return nil
}

// GetRoot returns the tuis directory path
func GetRoot() string {
	return application.App.Root() + "/tuis"
}
