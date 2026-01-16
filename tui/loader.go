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
	"github.com/yaoapp/yao/share"
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

	return &cfg, nil
}
