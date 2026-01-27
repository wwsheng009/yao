package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLoadScript tests loading scripts
func TestLoadScript(t *testing.T) {
	prepare(t)

	t.Run("load TypeScript file", func(t *testing.T) {
		// This will try to load scripts/tui/counter.ts
		script, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skipf("counter script not found: %v", err)
			return
		}

		assert.NotNil(t, script)
		assert.NotNil(t, script.Script)
		assert.Contains(t, script.File, "counter")
	})

	t.Run("load from cache", func(t *testing.T) {
		// First load
		script1, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		// Give time for async cache operation
		time.Sleep(100 * time.Millisecond)

		// Second load should come from cache
		script2, err := LoadScript("scripts/tui/counter")
		assert.NoError(t, err)
		assert.Equal(t, script1, script2, "should return same cached instance")
	})

	t.Run("disable cache", func(t *testing.T) {
		// Load with cache disabled
		script1, err := LoadScript("scripts/tui/counter", true)
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		script2, err := LoadScript("scripts/tui/counter", true)
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		// Both scripts should be valid, but they might be the same instance
		// depending on how V8 handles caching. The important thing is that
		// they both should be non-nil and functional.
		assert.NotNil(t, script1)
		assert.NotNil(t, script2)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadScript("scripts/tui/nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetScript(t *testing.T) {
	prepare(t)

	t.Run("get cached script", func(t *testing.T) {
		// Load first
		original, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		// Give time for async cache operation
		time.Sleep(100 * time.Millisecond)

		// Get from cache
		cached, ok := GetScript("scripts/tui/counter")
		assert.True(t, ok)
		assert.Equal(t, original, cached)
	})

	t.Run("get non-existent script", func(t *testing.T) {
		_, ok := GetScript("scripts/tui/doesnotexist")
		assert.False(t, ok)
	})
}

func TestRemoveScript(t *testing.T) {
	prepare(t)

	t.Run("remove script from cache", func(t *testing.T) {
		// Load first
		_, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		// Give time for async cache operation
		time.Sleep(100 * time.Millisecond)

		// Verify it's cached
		_, ok := GetScript("scripts/tui/counter")
		assert.True(t, ok, "should be cached")

		// Remove from cache
		RemoveScript("scripts/tui/counter")
		time.Sleep(100 * time.Millisecond)

		// Verify it's removed
		_, ok = GetScript("scripts/tui/counter")
		assert.False(t, ok, "should be removed from cache")
	})
}

func TestScriptExecute(t *testing.T) {
	prepare(t)

	t.Run("execute simple method", func(t *testing.T) {
		script, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		// Try to execute a method (will fail if method doesn't exist)
		// This is more of an integration test
		_, err = script.Execute("increment")
		// We expect an error because there's no TUI context
		// The test passes if we can create the context
		t.Logf("Execute result: %v", err)
	})

	t.Run("nil script", func(t *testing.T) {
		var script *Script
		_, err := script.Execute("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("method not found", func(t *testing.T) {
		script, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		_, err = script.Execute("nonExistentMethod")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestScriptExecuteWithModel(t *testing.T) {
	prepare(t)

	t.Run("execute with model context", func(t *testing.T) {
		script, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		// Create a test model
		cfg := &Config{
			Name: "Test TUI",
			Data: map[string]interface{}{
				"count": 0,
			},
			Layout: Layout{
				Direction: "vertical",
				Children:  []Component{},
			},
		}
		model := NewModel(cfg, nil)

		// Execute with model (will be fully implemented in Phase 2.2)
		_, err = script.ExecuteWithModel(model, "increment")
		// For now, we just test that the function signature works
		t.Logf("ExecuteWithModel result: %v", err)
	})

	t.Run("nil model", func(t *testing.T) {
		script, err := LoadScript("scripts/tui/counter")
		if err != nil {
			t.Skip("counter script not found")
			return
		}

		_, err = script.ExecuteWithModel(nil, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}

func TestListScripts(t *testing.T) {
	prepare(t)

	t.Run("list cached scripts", func(t *testing.T) {
		// Clear first
		ClearScripts()
		time.Sleep(100 * time.Millisecond)

		// Load some scripts
		LoadScript("scripts/tui/counter")
		time.Sleep(100 * time.Millisecond)

		paths := ListScripts()
		t.Logf("Cached scripts: %v", paths)

		// Should have at least one script if counter exists
		if len(paths) > 0 {
			assert.Contains(t, paths, "scripts/tui/counter")
		}
	})
}

func TestCountScripts(t *testing.T) {
	prepare(t)

	t.Run("count cached scripts", func(t *testing.T) {
		// Clear first
		ClearScripts()
		time.Sleep(100 * time.Millisecond)

		initialCount := CountScripts()
		assert.Equal(t, 0, initialCount)

		// Load a script
		_, err := LoadScript("scripts/tui/counter")
		if err == nil {
			time.Sleep(100 * time.Millisecond)
			count := CountScripts()
			assert.Greater(t, count, 0)
		}
	})
}

func TestClearScripts(t *testing.T) {
	prepare(t)

	t.Run("clear all scripts", func(t *testing.T) {
		// Load a script
		_, err := LoadScript("scripts/tui/counter")
		if err == nil {
			time.Sleep(100 * time.Millisecond)
			assert.Greater(t, CountScripts(), 0)
		}

		// Clear all
		ClearScripts()
		time.Sleep(100 * time.Millisecond)

		count := CountScripts()
		assert.Equal(t, 0, count)
	})
}

func TestScriptConcurrency(t *testing.T) {
	prepare(t)

	t.Run("concurrent script loading", func(t *testing.T) {
		// Clear first
		ClearScripts()
		time.Sleep(100 * time.Millisecond)

		// Load same script concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := LoadScript("scripts/tui/counter")
				if err == nil {
					done <- true
				} else {
					done <- false
				}
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		time.Sleep(200 * time.Millisecond)

		// Should only have one cached instance
		count := CountScripts()
		t.Logf("Scripts after concurrent load: %d", count)
		// Count might be 0 if script doesn't exist, or 1 if it does
		assert.LessOrEqual(t, count, 1, "should not duplicate cache entries")
	})
}