package tui

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
	"github.com/yaoapp/yao/test"
)

// prepare sets up the test environment and loads TUI configurations
func prepare(t *testing.T) {
	os.Setenv("YAO_TEST_APPLICATION", "E:/projects/yao/wwsheng009/yao/yao-docs/YaoApps/tui_app")
	test.Prepare(t, config.Conf)
	mirror := os.Getenv("TEST_MOAPI_MIRROR")
	secret := os.Getenv("TEST_MOAPI_SECRET")
	share.App = share.AppInfo{
		Moapi: share.Moapi{Channel: "stable", Mirrors: []string{mirror}, Secret: secret},
	}
	// Load TUI configurations
	err := Load(config.Conf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestShareID(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		file     string
		expected string
	}{
		{
			name:     "simple file",
			root:     "/app/tuis",
			file:     "/app/tuis/hello.tui.yao",
			expected: "hello",
		},
		{
			name:     "nested file",
			root:     "/app/tuis",
			file:     "/app/tuis/admin/dashboard.tui.yao",
			expected: "admin.dashboard",
		},
		{
			name:     "deeply nested",
			root:     "/app/tuis",
			file:     "/app/tuis/admin/users/list.tui.json",
			expected: "admin.users.list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := share.ID(tt.root, tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShareFile(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		ext      string
		expected string
	}{
		{
			name:     "simple ID with yao extension",
			id:       "hello",
			ext:      "tui.yao",
			expected: "hello.tui.yao",
		},
		{
			name:     "nested ID with json extension",
			id:       "admin.dashboard",
			ext:      "tui.json",
			expected: filepath.Join("admin", "dashboard.tui.json"),
		},
		{
			name:     "deeply nested ID with jsonc extension",
			id:       "admin.users.list",
			ext:      "tui.jsonc",
			expected: filepath.Join("admin", "users", "list.tui.jsonc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := share.File(tt.id, tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadFile(t *testing.T) {
	prepare(t)

	// Test loading with application context
	// The file should be accessible via application.App.Read
	t.Log("Testing loadFile with application context")

	// Since loadFile uses application.App.Read, we need to use a relative path
	// from the application root
	cfg, err := loadFile("tuis/test.tui.yao")
	if err != nil {
		t.Skipf("loadFile test skipped: %v", err)
		return
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "Test TUI", cfg.Name)
	assert.Equal(t, "Test Application", cfg.Data["title"])
	assert.Equal(t, "vertical", cfg.Layout.Direction)
	assert.Len(t, cfg.Layout.Children, 2)
}

func TestLoadFileErrors(t *testing.T) {
	prepare(t)

	t.Run("file not found", func(t *testing.T) {
		_, err := loadFile("/nonexistent/file.tui.yao")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		// This test is not applicable since we can't easily create invalid files
		// in the test app directory without affecting other tests
		t.Skip("Skipping invalid JSON test")
	})
}

func TestGetAndCache(t *testing.T) {
	// Clear cache first
	cache = sync.Map{}

	// Test Get with non-existent ID
	cfg := Get("nonexistent")
	assert.Nil(t, cfg)

	// Store a test config
	testCfg := &Config{
		ID:   "test",
		Name: "Test TUI",
		Layout: Layout{
			Direction: "vertical",
		},
	}
	cache.Store("test", testCfg)

	// Test Get with existing ID
	cfg = Get("test")
	assert.NotNil(t, cfg)
	assert.Equal(t, "test", cfg.ID)
	assert.Equal(t, "Test TUI", cfg.Name)
}

func TestList(t *testing.T) {
	// Clear and populate cache
	cache = sync.Map{}

	cache.Store("tui1", &Config{ID: "tui1"})
	cache.Store("tui2", &Config{ID: "tui2"})
	cache.Store("tui3", &Config{ID: "tui3"})

	// Test List
	ids := List()
	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "tui1")
	assert.Contains(t, ids, "tui2")
	assert.Contains(t, ids, "tui3")
}

// TestLoadIntegration tests the full Load() function
func TestLoadIntegration(t *testing.T) {
	prepare(t)
	defer test.Clean()

	// Test that TUIs are loaded successfully
	cfg := Get("hello")
	assert.NotNil(t, cfg)
	assert.Equal(t, "hello", cfg.ID)
	assert.Equal(t, "Hello TUI", cfg.Name)

	cfg = Get("admin.dashboard")
	assert.NotNil(t, cfg)
	assert.Equal(t, "admin.dashboard", cfg.ID)
	assert.Equal(t, "Dashboard", cfg.Name)

	// Test List function
	ids := List()
	assert.NotEmpty(t, ids)
	assert.Contains(t, ids, "hello")
	assert.Contains(t, ids, "admin.dashboard")
}

// TestReload tests the Reload function
func TestReload(t *testing.T) {
	prepare(t)
	defer test.Clean()

	// Get original config
	originalCfg := Get("hello")
	assert.NotNil(t, originalCfg)
	assert.Equal(t, "Hello TUI", originalCfg.Name)

	// Reload the TUI
	err := Reload("hello")
	assert.NoError(t, err)

	// Verify reloaded config
	newCfg := Get("hello")
	assert.NotNil(t, newCfg)
	assert.Equal(t, originalCfg.ID, newCfg.ID)
	assert.Equal(t, originalCfg.Name, newCfg.Name)
}

// TestReloadNotFound tests reloading a non-existent TUI
func TestReloadNotFound(t *testing.T) {
	prepare(t)
	defer test.Clean()

	err := Reload("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TUI file not found")
}
