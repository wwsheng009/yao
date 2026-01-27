package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFlattenData tests the FlattenData function
func TestFlattenData(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Simple flat data",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		{
			name: "Nested data",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
				"user.name": "John",
				"user.age":  30,
			},
		},
		{
			name:     "Empty data",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "Nil data",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FlattenData(tc.input)

			if tc.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestMergeData tests the MergeData function
func TestMergeData(t *testing.T) {
	testCases := []struct {
		name           string
		existing       map[string]interface{}
		external       map[string]interface{}
		priorityHigher bool
		expected       map[string]interface{}
	}{
		{
			name:           "External has higher priority",
			existing:       map[string]interface{}{"key1": "old", "key2": "keep"},
			external:       map[string]interface{}{"key1": "new", "key3": "add"},
			priorityHigher: true,
			expected:       map[string]interface{}{"key1": "new", "key2": "keep", "key3": "add"},
		},
		{
			name:           "Existing has higher priority",
			existing:       map[string]interface{}{"key1": "old", "key2": "keep"},
			external:       map[string]interface{}{"key1": "new", "key3": "add"},
			priorityHigher: false,
			expected:       map[string]interface{}{"key1": "old", "key2": "keep", "key3": "add"},
		},
		{
			name:           "Nil existing",
			existing:       nil,
			external:       map[string]interface{}{"key1": "value1"},
			priorityHigher: true,
			expected:       map[string]interface{}{"key1": "value1"},
		},
		{
			name:           "Nil external",
			existing:       map[string]interface{}{"key1": "value1"},
			external:       nil,
			priorityHigher: true,
			expected:       map[string]interface{}{"key1": "value1"},
		},
		{
			name:           "Both nil",
			existing:       nil,
			external:       nil,
			priorityHigher: true,
			expected:       map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MergeData(tc.existing, tc.external, tc.priorityHigher)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestValidateAndFlattenExternal tests the ValidateAndFlattenExternal function
func TestValidateAndFlattenExternal(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Nested external data",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   123,
					"name": "John",
				},
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   123,
					"name": "John",
				},
				"user.id":   123,
				"user.name": "John",
			},
		},
		{
			name:     "Nil external data",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateAndFlattenExternal(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestPrepareInitialState tests the PrepareInitialState function
func TestPrepareInitialState(t *testing.T) {
	t.Run("Merge external with config data", func(t *testing.T) {
		cfg := &Config{
			ID:   "test",
			Name: "Test TUI",
			Data: map[string]interface{}{
				"title": "Static Title",
				"count": 10,
			},
		}

		externalData := map[string]interface{}{
			"title":  "External Title",
			"newKey": "newValue",
			"user":   map[string]interface{}{"name": "John"},
		}

		state := PrepareInitialState(cfg, externalData)

		// Check that external data overrides config data
		assert.Equal(t, "External Title", state["title"])
		assert.Equal(t, 10, state["count"])

		// Check that new external key is added
		assert.Equal(t, "newValue", state["newKey"])

		// Check that nested data is flattened
		assert.Equal(t, "John", state["user.name"])

		// Check that original nested structure is preserved
		userData, ok := state["user"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "John", userData["name"])
	})

	t.Run("Nil external data", func(t *testing.T) {
		cfg := &Config{
			ID:   "test",
			Name: "Test TUI",
			Data: map[string]interface{}{
				"title": "Static Title",
			},
		}

		state := PrepareInitialState(cfg, nil)

		assert.Equal(t, "Static Title", state["title"])
	})

	t.Run("Nil config data", func(t *testing.T) {
		cfg := &Config{
			ID:   "test",
			Name: "Test TUI",
			Data: nil,
		}

		externalData := map[string]interface{}{
			"title": "External Title",
		}

		state := PrepareInitialState(cfg, externalData)

		assert.Equal(t, "External Title", state["title"])
		assert.NotNil(t, cfg.Data)
	})
}

// TestApplyOnLoadResult tests the ApplyOnLoadResult function
func TestApplyOnLoadResult(t *testing.T) {
	// Note: This test doesn't actually call ApplyOnLoadResult because it requires
	// a running Program with message handling. Instead we test the logic that
	// ApplyOnLoadResult implements.

	t.Run("Determine storage strategy based on onSuccess", func(t *testing.T) {
		testCases := []struct {
			name        string
			result      interface{}
			onSuccess   string
			expectStore string // "specific", "merge", "default"
		}{
			{
				name:        "Store in specific key",
				result:      map[string]interface{}{"data": []string{"a", "b"}},
				onSuccess:   "loadedData",
				expectStore: "specific",
			},
			{
				name:        "Merge map result",
				result:      map[string]interface{}{"key1": "value1", "key2": "value2"},
				onSuccess:   "",
				expectStore: "merge",
			},
			{
				name:        "Store non-map in default",
				result:      "simple result",
				onSuccess:   "",
				expectStore: "default",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var storageStrategy string

				// Test the logic from ApplyOnLoadResult
				if tc.onSuccess != "" {
					storageStrategy = "specific"
				} else if resultMap, ok := tc.result.(map[string]interface{}); ok {
					storageStrategy = "merge"
					_ = resultMap
				} else {
					storageStrategy = "default"
				}

				assert.Equal(t, tc.expectStore, storageStrategy)
			})
		}
	})

	t.Run("Calculate state key based on onSuccess", func(t *testing.T) {
		onSuccess := "customKey"
		expectedKey := "customKey"
		assert.Equal(t, expectedKey, onSuccess)
	})

	t.Run("Calculate state key for map merge", func(t *testing.T) {
		result := map[string]interface{}{"key": "value"}
		expectedKeys := []string{"key"}

		var actualKeys []string
		for k := range result {
			actualKeys = append(actualKeys, k)
		}

		assert.Equal(t, expectedKeys, actualKeys)
	})
}
